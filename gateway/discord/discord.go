// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package discord

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/gowarcraft3/network"
)

// Errors
var (
	ErrSayBufferFull = errors.New("gw-discord: Say buffer full")
	ErrInvalidGuild  = errors.New("gw-discord: Invalid guild ID")
)

// Config stores the configuration of a Discord session
type Config struct {
	gateway.Config
	AuthToken  string
	Channels   map[string]*ChannelConfig
	Presence   string
	AccessDM   gateway.AccessLevel
	AccessUser map[string]gateway.AccessLevel
}

// Gateway manages a Discord connection
type Gateway struct {
	gateway.Common
	network.EventEmitter
	*discordgo.Session

	chatmut sync.Mutex
	users   map[string]struct{}
	guilds  map[string][]string

	// Set once before Run(), read-only after that
	*Config
	Channels map[string]*Channel
}

// New initializes a new Gateway struct
func New(conf *Config) (*Gateway, error) {
	s, err := discordgo.New("Bot " + conf.AuthToken)
	if err != nil {
		return nil, err
	}

	s.SyncEvents = true
	s.State.TrackEmojis = false
	s.State.TrackVoice = false
	s.State.MaxMessageCount = 0

	var d = Gateway{
		Session:  s,
		Config:   conf,
		Channels: make(map[string]*Channel),

		users:  make(map[string]struct{}),
		guilds: make(map[string][]string),
	}

	for id, c := range d.Config.Channels {
		ch, err := NewChannel(d.Session, id, c)
		if err != nil {
			return nil, err
		}
		d.Channels[id] = ch
	}

	d.InitDefaultHandlers()

	return &d, nil
}

// Channel residing in
func (d *Gateway) Channel() *gateway.Channel {
	return nil
}

// ChannelUsers online
func (d *Gateway) ChannelUsers() []gateway.User {
	return nil
}

// User by ID
func (d *Gateway) User(uid string) (*gateway.User, error) {
	uid = strings.TrimSuffix(strings.TrimPrefix(uid, "<@"), ">")
	if err := validateUID(uid); err != nil {
		return nil, err
	}

	u, err := d.Session.User(uid)
	if err != nil {
		return nil, err
	}

	var res = gateway.User{
		ID:        u.ID,
		Name:      u.Username,
		AvatarURL: u.AvatarURL(""),
		Access:    d.AccessDM,
	}

	if access := d.AccessUser[u.ID]; access != gateway.AccessDefault {
		res.Access = access
	}

	return &res, nil
}

// Users with non-default access level
func (d *Gateway) Users() map[string]gateway.AccessLevel {
	return d.AccessUser
}

func validateUID(uid string) error {
	if _, err := strconv.ParseUint(uid, 10, 64); err != nil {
		return gateway.ErrNoUser
	}
	return nil
}

// SetUserAccess overrides accesslevel for a specific user
func (d *Gateway) SetUserAccess(uid string, a gateway.AccessLevel) (*gateway.AccessLevel, error) {
	if err := validateUID(uid); err != nil {
		return nil, err
	}

	var o = d.AccessUser[uid]
	if a != gateway.AccessDefault {
		if d.AccessUser == nil {
			d.AccessUser = make(map[string]gateway.AccessLevel)
		}

		d.AccessUser[uid] = a
	} else {
		delete(d.AccessUser, uid)
	}

	d.Fire(&gateway.ConfigUpdate{})
	return &o, nil
}

// Say sends a chat message
func (d *Gateway) Say(s string) error {
	return nil
}

func sayPrivate(d *discordgo.Session, uid string, s string) error {
	if err := validateUID(uid); err != nil {
		return err
	}

	ch, err := d.UserChannelCreate(uid)
	if err != nil {
		return err
	}
	if len(s) > 2000 {
		s = s[:1997] + "..."
	}

	_, err = d.ChannelMessageSend(ch.ID, s)
	return err
}

// SayPrivate sends a private chat message to uid
func (d *Gateway) SayPrivate(uid string, s string) error {
	return sayPrivate(d.Session, uid, s)
}

// Kick user from channel
func (d *Gateway) Kick(uid string) error {
	return gateway.ErrNoChannel
}

// Ban user from channel
func (d *Gateway) Ban(uid string) error {
	return gateway.ErrNoChannel
}

// Unban user from channel
func (d *Gateway) Unban(uid string) error {
	return gateway.ErrNoChannel
}

// Ping user to calculate RTT in milliseconds
func (d *Gateway) Ping(uid string) (time.Duration, error) {
	return 0, gateway.ErrNotImplemented
}

// Run reads packets and emits an event for each received packet
func (d *Gateway) Run(ctx context.Context) error {
	var err error
	for i := 1; i < 60 && ctx.Err() == nil; i++ {
		err = d.Session.Open()
		if err == nil {
			break
		}

		d.Fire(&network.AsyncError{Src: "Run[Open]", Err: err})

		select {
		case <-time.After(2 * time.Minute):
		case <-ctx.Done():
		}
	}

	if err != nil {
		return err
	}

	<-ctx.Done()
	d.Close()

	return ctx.Err()
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (d *Gateway) InitDefaultHandlers() {
	d.AddHandler(func(s *discordgo.Session, i interface{}) {
		// Forward all discordgo events
		d.Fire(i)
	})

	d.AddHandler(d.onConnect)
	d.AddHandler(d.onDisconnect)
	d.AddHandler(d.onReady)

	d.AddHandler(d.onGuildCreate)
	d.AddHandler(d.onGuildUpdate)
	d.AddHandler(d.onGuildMemberUpdate)
	d.AddHandler(d.onGuildMemberRemove)
	d.AddHandler(d.onPresenceUpdate)

	d.AddHandler(d.onMessageCreate)
}

func (d *Gateway) onConnect(s *discordgo.Session, msg *discordgo.Connect) {
	d.Fire(&gateway.Connected{})
}

func (d *Gateway) onDisconnect(s *discordgo.Session, msg *discordgo.Disconnect) {
	d.Fire(&gateway.Disconnected{})
	d.clear()
}

func (d *Gateway) onReady(s *discordgo.Session, r *discordgo.Ready) {
	if d.Presence != "" {
		go func() {
			if err := s.UpdateStatus(0, d.Presence); err != nil {
				d.Fire(&network.AsyncError{Src: "onReady[UpdateStatus]", Err: err})
			}
		}()
	}

	d.clear()

	for _, g := range r.Guilds {
		d.onGuildCreate(s, &discordgo.GuildCreate{Guild: g})
	}
}

func (d *Gateway) clear() {
	d.Fire(&gateway.Clear{})
	for _, c := range d.Channels {
		c.Fire(&gateway.Clear{})
	}

	d.chatmut.Lock()
	for u := range d.users {
		delete(d.users, u)
	}
	d.chatmut.Unlock()
}

func (d *Gateway) updatePresence(guildID string, presence *discordgo.Presence) {
	if presence.User.Bot {
		return
	}

	d.chatmut.Lock()
	var _, online = d.users[presence.User.ID]

	if (presence.Status != discordgo.StatusOffline) == online {
		d.chatmut.Unlock()
		return
	}

	var track = false
	var channels = d.guilds[guildID]
	for _, cid := range channels {
		perm, err := d.UserChannelPermissions(presence.User.ID, cid)
		if err != nil && !online {
			d.Fire(&network.AsyncError{Src: "updatePresence[permissions]", Err: err})
			continue
		}

		// Check if user is allowed to read channel
		if perm&discordgo.PermissionReadMessages == 0 && !online {
			continue
		}

		evUser, err := d.Channels[cid].User(presence.User.ID)
		if err != nil {
			if !online {
				d.Fire(&network.AsyncError{Src: "updatePresence[user]", Err: err})
				continue
			}

			// Assume user left the guild
			evUser = &gateway.User{
				ID:     presence.User.ID,
				Name:   presence.User.Username,
				Access: gateway.AccessMin,
			}
		}
		if evUser == nil {
			continue
		}

		track = true

		if presence.Status != discordgo.StatusOffline {
			d.Channels[cid].Fire(&gateway.Join{User: *evUser})
		} else {
			d.Channels[cid].Fire(&gateway.Leave{User: *evUser})
		}
	}

	if presence.Status == discordgo.StatusOffline {
		delete(d.users, presence.User.ID)
	} else if track {
		d.users[presence.User.ID] = struct{}{}
	}
	d.chatmut.Unlock()
}

func (d *Gateway) onGuildCreate(s *discordgo.Session, msg *discordgo.GuildCreate) {
	// Clear guild channels
	delete(d.guilds, msg.Guild.ID)

	for _, ch := range msg.Channels {
		c := d.Channels[ch.ID]
		if c == nil {
			continue
		}

		if c.guildID != ch.GuildID {
			c.guildID = ch.GuildID
			d.Fire(&network.AsyncError{Src: "onGuildCreate", Err: ErrInvalidGuild})
		}
		d.guilds[ch.GuildID] = append(d.guilds[ch.GuildID], ch.ID)

		c.Fire(c.Channel())
	}

	for _, p := range msg.Presences {
		d.updatePresence(msg.Guild.ID, p)
	}
}

func (d *Gateway) onGuildUpdate(s *discordgo.Session, msg *discordgo.GuildUpdate) {
	for _, p := range msg.Presences {
		d.updatePresence(msg.Guild.ID, p)
	}
}

func (d *Gateway) onGuildMemberUpdate(s *discordgo.Session, msg *discordgo.GuildMemberUpdate) {
	if _, online := d.users[msg.User.ID]; !online {
		return
	}

	var channels = d.guilds[msg.GuildID]
	for _, cid := range channels {
		if u, err := d.Channels[cid].User(msg.User.ID); err == nil && u != nil {
			d.Channels[cid].Fire(u)
		}
	}
}

func (d *Gateway) onGuildMemberRemove(s *discordgo.Session, msg *discordgo.GuildMemberRemove) {
	d.updatePresence(msg.GuildID, &discordgo.Presence{
		User:   msg.User,
		Status: discordgo.StatusOffline,
	})
}

func (d *Gateway) onPresenceUpdate(s *discordgo.Session, msg *discordgo.PresenceUpdate) {
	d.updatePresence(msg.GuildID, &msg.Presence)
}

var patternEmotiji = regexp.MustCompile("<a?:([^:]*):[^>]*>")

func replaceContentReferences(s *discordgo.Session, msg *discordgo.Message) string {
	var res = msg.Content

	// Replace usernames, channels, roles
	if c, err := msg.ContentWithMoreMentionsReplaced(s); err == nil {
		res = c
	}

	// Replace emojis
	res = patternEmotiji.ReplaceAllString(res, ":$1:")

	return res
}

func (d *Gateway) onMessageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {
	if msg.Content == "" || msg.Author.Bot {
		return
	}

	var c = d.Channels[msg.ChannelID]
	if c == nil {
		channel, err := s.State.Channel(msg.ChannelID)
		if err != nil {
			d.Fire(&network.AsyncError{Src: "onMessageCreate[Channel]", Err: err})
			return
		}

		// Check if private message
		if channel.Type == discordgo.ChannelTypeDM {
			var u = gateway.User{
				ID:        msg.Author.ID,
				Name:      msg.Author.Username,
				AvatarURL: msg.Author.AvatarURL(""),
				Access:    d.AccessDM,
			}

			if access := d.AccessUser[msg.Author.ID]; access != gateway.AccessDefault {
				u.Access = access
			}

			var chat = gateway.PrivateChat{
				User:    u,
				Content: replaceContentReferences(s, msg.Message),
			}

			d.Fire(&chat)

			if chat.User.Access >= d.Commands.Access {
				if t := d.FindTrigger(msg.Message.Content); t != nil {
					t.User = chat.User
					t.Resp = func(s string) error { _, err = d.ChannelMessageSend(msg.ChannelID, s); return err }
					d.Fire(t, &chat)
				}
			}
		}

		return
	}

	evUser, err := c.User(msg.Author.ID)
	if err != nil {
		d.Fire(&network.AsyncError{Src: "onMessageCreate[user]", Err: err})
		return
	}
	if evUser == nil {
		return
	}

	var chat = gateway.Chat{
		User:    *evUser,
		Content: replaceContentReferences(s, msg.Message),
	}

	c.Fire(&chat)

	if chat.User.Access < c.Commands.Access {
		return
	}

	if t := c.FindTrigger(msg.Message.Content); t != nil {
		t.User = chat.User
		t.Resp = c.Responder(c, chat.User.ID, false)
		c.Fire(t, &chat)
	}
}

// Relay placeholder to implement Gateway interface
// Events should instead be relayed directly to a Channel
func (d *Gateway) Relay(ev *network.Event, from gateway.Gateway) error {
	return nil
}

// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package discord

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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
)

// RelayJoinMode enum
type RelayJoinMode int32

// RelayJoins
const (
	RelayJoinsSay = 1 << iota
	RelayJoinsList

	RelayJoinsBoth = RelayJoinsSay | RelayJoinsList
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

// ChannelConfig stores the configuration of a single Discord channel
type ChannelConfig struct {
	gateway.Config
	BufSize        uint8
	Webhook        string
	RelayJoins     RelayJoinMode
	AccessMentions gateway.AccessLevel
	AccessTalk     gateway.AccessLevel
	AccessRole     map[string]gateway.AccessLevel
	AccessUser     map[string]gateway.AccessLevel
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

// Channel manages a Discord channel
type Channel struct {
	gateway.Common
	network.EventEmitter

	wg      *sync.WaitGroup
	id      string
	session *discordgo.Session

	smut  sync.Mutex
	saych chan string
	saywh chan *discordgo.WebhookParams

	omut   sync.Mutex
	ochan  chan struct{}
	online []online

	*ChannelConfig
}

type online struct {
	Name  string
	Since time.Time
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

	var wg sync.WaitGroup
	wg.Add(1)
	d.Once(&gateway.Connected{}, func(ev *network.Event) {
		go func() {
			time.Sleep(time.Second)
			wg.Done()
		}()
	})

	for id, c := range d.Config.Channels {
		ch, err := d.Session.Channel(id)
		if err != nil {
			return nil, err
		}
		d.guilds[ch.GuildID] = append(d.guilds[ch.GuildID], id)

		d.Channels[id] = &Channel{
			ChannelConfig: c,

			wg:      &wg,
			id:      id,
			session: s,
		}
	}

	d.InitDefaultHandlers()

	return &d, nil
}

// Channel currently being monitoring
func (d *Gateway) Channel() *gateway.Channel {
	return nil
}

// Say sends a chat message
func (d *Gateway) Say(s string) error {
	return gateway.ErrNoChannel
}

func sayPrivate(d *discordgo.Session, uid string, s string) error {
	ch, err := d.UserChannelCreate(uid)
	if err != nil {
		return err
	}

	_, err = d.ChannelMessageSend(ch.ID, s)
	return err
}

// SayPrivate sends a private chat message to uid
func (d *Gateway) SayPrivate(uid string, s string) error {
	return sayPrivate(d.Session, uid, s)
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

func (d *Gateway) user(chanID string, userID string) (*gateway.User, error) {
	var c = d.Channels[chanID]
	if c == nil {
		return nil, nil
	}

	channel, err := d.State.Channel(chanID)
	if err != nil {
		return nil, err
	}

	member, err := d.State.Member(channel.GuildID, userID)
	if err != nil {
		return nil, err
	}

	if member.User.Bot {
		return nil, nil
	}

	var res = gateway.User{
		ID:        member.User.ID,
		Name:      member.User.Username,
		AvatarURL: member.User.AvatarURL(""),
		Access:    c.AccessTalk,
	}

	if member.Nick != "" {
		res.Name = member.Nick
	}

	if c.AccessRole != nil {
		for _, rid := range member.Roles {
			r, err := d.State.Role(channel.GuildID, rid)
			if err != nil {
				continue
			}
			if access, ok := c.AccessRole[strings.ToLower(r.Name)]; ok {
				res.Access = access
			}
		}
	}

	if access, ok := d.AccessUser[strings.ToLower(member.User.String())]; ok {
		res.Access = access
	}
	if access, ok := c.AccessUser[strings.ToLower(member.User.String())]; ok {
		res.Access = access
	}

	return &res, nil
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (d *Gateway) InitDefaultHandlers() {
	d.AddHandler(d.onConnect)
	d.AddHandler(d.onDisconnect)

	d.AddHandler(d.onPresenceUpdate)
	d.AddHandler(d.onGuildCreate)
	d.AddHandler(d.onGuildUpdate)

	d.AddHandler(d.onMessageCreate)
}

func (d *Gateway) onConnect(s *discordgo.Session, msg *discordgo.Connect) {
	if d.Presence != "" {
		go func() {
			time.Sleep(time.Second)
			if err := s.UpdateStatus(0, d.Presence); err != nil {
				d.Fire(&network.AsyncError{Src: "onConnect[UpdateStatus]", Err: err})
			}
		}()
	}
	d.Fire(&gateway.Connected{})
}

func (d *Gateway) onDisconnect(s *discordgo.Session, msg *discordgo.Disconnect) {
	d.Fire(&gateway.Disconnected{})
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
		perm, err := d.State.UserChannelPermissions(presence.User.ID, cid)
		if err != nil {
			d.Fire(&network.AsyncError{Src: "updatePresence[permissions]", Err: err})
			continue
		}

		// Check if user is allowed to read channel
		if perm&discordgo.PermissionReadMessages == 0 {
			continue
		}

		evUser, err := d.user(cid, presence.User.ID)
		if err != nil {
			d.Fire(&network.AsyncError{Src: "updatePresence[user]", Err: err})
			continue
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

func (d *Gateway) onPresenceUpdate(s *discordgo.Session, msg *discordgo.PresenceUpdate) {
	d.updatePresence(msg.GuildID, &msg.Presence)
}

func (d *Gateway) onGuildCreate(s *discordgo.Session, msg *discordgo.GuildCreate) {
	for _, p := range msg.Presences {
		d.updatePresence(msg.Guild.ID, p)
	}
}

func (d *Gateway) onGuildUpdate(s *discordgo.Session, msg *discordgo.GuildUpdate) {
	for _, p := range msg.Presences {
		d.updatePresence(msg.Guild.ID, p)
	}
}

var patternEmotiji = regexp.MustCompile("<a?:([^:]*):[^>]>")

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

			if access, ok := d.AccessUser[strings.ToLower(msg.Author.String())]; ok {
				u.Access = access
			}

			var chat = gateway.PrivateChat{
				User:    u,
				Content: replaceContentReferences(s, msg.Message),
			}

			d.Fire(&chat)

			if chat.User.Access >= d.Commands.Access {
				if r, cmd, arg := d.FindTrigger(chat.Content); r {
					c.Fire(&gateway.Trigger{
						User: chat.User,
						Cmd:  cmd,
						Arg:  arg,
						Resp: func(s string) error { _, err = d.ChannelMessageSend(msg.ChannelID, s); return err },
					}, chat)
				}
			}
		}

		return
	}

	evUser, err := d.user(msg.ChannelID, msg.Author.ID)
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

	if r, cmd, arg := c.FindTrigger(chat.Content); r {
		c.Fire(&gateway.Trigger{
			User: chat.User,
			Cmd:  cmd,
			Arg:  arg,
			Resp: c.Say,
		}, chat)
	}
}

// Relay placeholder to implement Gateway interface
// Events should instead be relayed directly to a Channel
func (d *Gateway) Relay(ev *network.Event, from gateway.Gateway) error {
	return nil
}

// Channel currently being monitoring
func (c *Channel) Channel() *gateway.Channel {
	var name = c.id
	if ch, err := c.session.State.Channel(c.id); err == nil {
		if g, err := c.session.State.Guild(ch.GuildID); err == nil {
			name = fmt.Sprintf("[%s]%s", g.Name, ch.Name)
		} else {
			name = ch.Name
		}
	}
	return &gateway.Channel{ID: c.id, Name: name}
}

func (c *Channel) say(s string) error {
	if c.session == nil {
		return nil
	}

	c.smut.Lock()
	if c.saych == nil {
		c.saych = make(chan string, c.BufSize)

		go func() {
			for s := range c.saych {
				_, err := c.session.ChannelMessageSend(c.id, s)
				if err != nil {
					c.Fire(&network.AsyncError{Src: "Say", Err: err})
				}
			}
		}()
	}
	c.smut.Unlock()

	select {
	case c.saych <- s:
		return nil
	default:
		return ErrSayBufferFull
	}
}

// Say sends a chat message
func (c *Channel) Say(s string) error {
	if err := c.say(s); err != nil {
		return err
	}
	c.Fire(&gateway.Say{Content: s})
	return nil
}

// SayPrivate sends a private chat message to uid
func (c *Channel) SayPrivate(uid string, s string) error {
	return sayPrivate(c.session, uid, s)
}

// WebhookOrSay sends a chat message preferably via webhook
func (c *Channel) WebhookOrSay(p *discordgo.WebhookParams) error {
	if c.Webhook == "" {
		var s = p.Content
		if p.Username != "" {
			s = fmt.Sprintf("**<%s>** %s", p.Username, p.Content)
		}
		return c.say(s)
	}

	c.smut.Lock()
	if c.saywh == nil {
		c.saywh = make(chan *discordgo.WebhookParams, c.BufSize)

		go func() {
			for p := range c.saywh {
				_, err := c.session.RequestWithBucketID("POST", c.Webhook, p, discordgo.EndpointWebhookToken("", ""))
				if err != nil {
					c.Fire(&network.AsyncError{Src: "WebhookOrSay", Err: err})
				}
			}
		}()
	}
	c.smut.Unlock()

	select {
	case c.saywh <- p:
		return nil
	default:
		return ErrSayBufferFull
	}
}

func (c *Channel) filter(s string, r gateway.AccessLevel) string {
	if r < c.AccessMentions {
		s = strings.Replace(s, "@", "@\u200B", -1)
	}
	return s
}

// Run placeholder to implement Gateway interface
func (c *Channel) Run(ctx context.Context) error {
	var done = make(chan struct{})

	go func() {
		c.wg.Wait()
		c.Fire(c.Channel())
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	return nil
}

// FindTrigger checks if s starts with trigger, returns cmd and args if true
func (c *Channel) FindTrigger(s string) (bool, string, []string) {
	if r, cmd, arg := c.Config.FindTrigger(s); r {
		return r, cmd, arg
	}
	if r, cmd, arg := gateway.FindTrigger(fmt.Sprintf("@%s ", c.session.State.User.Username), s); r {
		return r, cmd, arg
	}
	return false, "", nil
}

func (c *Channel) updateOnline() {
	c.omut.Lock()
	if c.ochan == nil {
		c.ochan = make(chan struct{}, 1)

		var t = time.NewTicker(time.Minute)
		go func() {
			var msg = ""

			if messages, err := c.session.ChannelMessagesPinned(c.id); err == nil {
				for _, m := range messages {
					if m.Author.ID != c.session.State.User.ID || !strings.HasPrefix(m.Content, "💬 **Online**") {
						continue
					}
					msg = m.ID
					break
				}
			}

			var last = ""
			for {
				select {
				case <-c.ochan:
				case <-t.C:
				}

				if c.RelayJoins&RelayJoinsList == 0 {
					continue
				}

				c.omut.Lock()
				var content = fmt.Sprintf("💬 **Online**: %d users", len(c.online))
				for i := len(c.online) - 1; i >= 0; i-- {
					content += fmt.Sprintf("\n`%s` *%s*", c.online[i].Name, time.Now().Sub(c.online[i].Since).Round(time.Second).String())
				}
				c.omut.Unlock()

				if content == last {
					continue
				}

				last = content

				if msg == "" {
					m, err := c.session.ChannelMessageSend(c.id, content)
					if err != nil {
						c.Fire(&network.AsyncError{Src: "updateOnline[Send]", Err: err})
					}
					msg = m.ID
					continue
				}

				if _, err := c.session.ChannelMessageEdit(c.id, msg, content); err != nil {
					c.Fire(&network.AsyncError{Src: "updateOnline[Update]", Err: err})
				}
			}
		}()
	}
	c.omut.Unlock()

	select {
	case c.ochan <- struct{}{}:
	default:
	}
}

// Relay dumps the event content in channel
func (c *Channel) Relay(ev *network.Event, from gateway.Gateway) error {
	switch msg := ev.Arg.(type) {
	case *gateway.Connected:
		return c.say(fmt.Sprintf("*Established connection to %s*", from.ID()))
	case *gateway.Disconnected:
		return c.say(fmt.Sprintf("*Connection to %s closed*", from.ID()))
	case *network.AsyncError:
		return c.say(fmt.Sprintf("❗ **%s** ERROR: %s", from.Discriminator(), msg.Error()))
	case *gateway.SystemMessage:
		return c.say(fmt.Sprintf("📢 **%s** %s", from.Discriminator(), msg.Content))
	case *gateway.Channel:
		return c.say(fmt.Sprintf("*Joined channel %s@%s*", msg.Name, from.Discriminator()))
	case *gateway.Join:
		if c.RelayJoins&RelayJoinsList != 0 {
			c.omut.Lock()
			c.online = append(c.online, online{
				Name:  fmt.Sprintf("%s@%s", msg.User.Name, from.Discriminator()),
				Since: time.Now(),
			})
			c.omut.Unlock()
			c.updateOnline()
		}

		if c.RelayJoins == 0 || c.RelayJoins&RelayJoinsSay != 0 {
			return c.say(fmt.Sprintf("➡️ **%s@%s** has joined the channel", msg.User.Name, from.Discriminator()))
		}

		return nil

	case *gateway.Leave:
		if c.RelayJoins&RelayJoinsList != 0 {
			var name = fmt.Sprintf("%s@%s", msg.User.Name, from.Discriminator())
			c.omut.Lock()
			for i := len(c.online) - 1; i >= 0; i-- {
				if c.online[i].Name == name {
					c.online = append(c.online[:i], c.online[i+1:]...)
					break
				}
			}
			c.omut.Unlock()
			c.updateOnline()
		}

		if c.RelayJoins == 0 || c.RelayJoins&RelayJoinsSay != 0 {
			return c.say(fmt.Sprintf("⬅️ **%s@%s** has left the channel", msg.User.Name, from.Discriminator()))
		}

		return nil

	case *gateway.PrivateChat:
		return c.WebhookOrSay(&discordgo.WebhookParams{
			Content:   c.filter(msg.Content, msg.User.Access),
			Username:  fmt.Sprintf("%s@%s (Direct Message)", msg.User.Name, from.Discriminator()),
			AvatarURL: msg.User.AvatarURL,
		})
	case *gateway.Chat:
		return c.WebhookOrSay(&discordgo.WebhookParams{
			Content:   c.filter(msg.Content, msg.User.Access),
			Username:  fmt.Sprintf("%s@%s", msg.User.Name, from.Discriminator()),
			AvatarURL: msg.User.AvatarURL,
		})
	case *gateway.Say:
		var p = &discordgo.WebhookParams{
			Content:  c.filter(msg.Content, 0),
			Username: from.Discriminator(),
		}
		if c.session != nil {
			p.AvatarURL = c.session.State.User.AvatarURL("")
		}
		return c.WebhookOrSay(p)

	default:
		return gateway.ErrUnknownEvent
	}
}

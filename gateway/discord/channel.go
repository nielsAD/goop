// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package discord

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/gowarcraft3/network"
)

// ChannelConfig stores the configuration of a single Discord channel
type ChannelConfig struct {
	gateway.Config
	BufSize        uint8
	Webhook        string
	OnlineListID   string
	RelayJoins     RelayJoinMode
	AccessMentions gateway.AccessLevel
	AccessTalk     gateway.AccessLevel
	AccessRole     map[string]gateway.AccessLevel
	AccessUser     map[string]gateway.AccessLevel
}

// Channel manages a Discord channel
type Channel struct {
	gateway.Common
	network.EventEmitter

	chanID  string
	guildID string
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
	gateway.User
	Gateway string
	Discr   string
	Since   time.Time
}

// NewChannel initializes a new Channel struct
func NewChannel(s *discordgo.Session, id string, conf *ChannelConfig) (*Channel, error) {
	var c = Channel{
		ChannelConfig: conf,

		chanID:  id,
		session: s,
	}

	if ch, err := s.Channel(id); err == nil {
		c.chanID = ch.ID
		c.guildID = ch.GuildID
	}

	return &c, nil
}

// Channel residing in
func (c *Channel) Channel() *gateway.Channel {
	var name = c.chanID
	if ch, err := c.session.State.Channel(c.chanID); err == nil {
		if g, err := c.session.State.Guild(ch.GuildID); err == nil {
			name = fmt.Sprintf("[%s]%s", g.Name, ch.Name)
		} else {
			name = ch.Name
		}
	}

	return &gateway.Channel{ID: c.chanID, Name: name}
}

// ChannelUsers online
func (c *Channel) ChannelUsers() []gateway.User {
	g, err := c.session.State.Guild(c.guildID)
	if err != nil {
		return nil
	}

	c.session.State.RLock()

	var res = make([]gateway.User, 0, len(g.Presences))
	for _, p := range g.Presences {
		perm, err := c.session.State.UserChannelPermissions(p.User.ID, c.chanID)
		if err != nil || perm&discordgo.PermissionReadMessages == 0 {
			continue
		}

		u, err := c.User(p.User.ID)
		if err != nil || u == nil {
			continue
		}
		res = append(res, *u)
	}

	c.session.State.RUnlock()

	return res
}

// User by ID
func (c *Channel) User(uid string) (*gateway.User, error) {
	uid = strings.TrimSuffix(strings.TrimPrefix(uid, "<@"), ">")

	member, err := c.session.State.Member(c.guildID, uid)
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
			r, err := c.session.State.Role(c.guildID, rid)
			if err != nil {
				continue
			}
			if access := c.AccessRole[strings.ToLower(r.Name)]; access != gateway.AccessDefault {
				res.Access = access
			}
		}
	}

	if access := c.AccessUser[member.User.ID]; access != gateway.AccessDefault {
		res.Access = access
	}

	return &res, nil
}

// Users with non-default access level
func (c *Channel) Users() map[string]gateway.AccessLevel {
	return c.AccessUser
}

// SetUserAccess overrides accesslevel for a specific user
func (c *Channel) SetUserAccess(uid string, a gateway.AccessLevel) (*gateway.AccessLevel, error) {
	if err := validateUID(uid); err != nil {
		return nil, err
	}

	var o = c.AccessUser[uid]
	if a != gateway.AccessDefault {
		if c.AccessUser == nil {
			c.AccessUser = make(map[string]gateway.AccessLevel)
		}

		c.AccessUser[uid] = a
	} else {
		delete(c.AccessUser, uid)
	}

	c.Fire(&gateway.ConfigUpdate{})

	if p, err := c.session.State.Presence(c.guildID, uid); err == nil && p.Status != discordgo.StatusOffline {
		if u, err := c.User(uid); err == nil && u != nil {
			c.Fire(u)
		}
	}

	return &o, nil
}

func (c *Channel) say(s string) error {
	c.smut.Lock()
	if c.saych == nil {
		c.saych = make(chan string, c.BufSize)

		go func() {
			for s := range c.saych {
				if len(s) > 2000 {
					s = s[:1997] + "..."
				}
				_, err := c.session.ChannelMessageSend(c.chanID, s)
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

// Kick user from channel
func (c *Channel) Kick(uid string) error {
	return gateway.ErrNotImplemented
}

// Ban user from channel
func (c *Channel) Ban(uid string) error {
	return gateway.ErrNotImplemented
}

// Unban user from channel
func (c *Channel) Unban(uid string) error {
	return gateway.ErrNotImplemented
}

// Ping user to calculate RTT in milliseconds
func (c *Channel) Ping(uid string) (time.Duration, error) {
	return 0, gateway.ErrNotImplemented
}

// WebhookOrSay sends a chat message preferably via webhook
func (c *Channel) WebhookOrSay(p *discordgo.WebhookParams) error {
	if len(p.Content) > 2000 {
		p.Content = p.Content[:1997] + "..."
	}

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

var mentionPat = regexp.MustCompile(`\B(@\S+)`)
var channelPat = regexp.MustCompile(`\B(#\S+)`)

func (c *Channel) parse(s string, l gateway.AccessLevel) string {
	if l < c.AccessMentions {
		// Add zero width space to prevent mentions
		s = strings.Replace(s, "@", "@\u200B", -1)
	}

	g, err := c.session.State.Guild(c.guildID)
	if err != nil {
		return s
	}

	c.session.RLock()

	s = channelPat.ReplaceAllStringFunc(s, func(m string) string {
		// Strip #
		var m1 = m[1:]

		for _, ch := range g.Channels {
			if strings.EqualFold(m1, ch.Name) {
				return ch.Mention()
			}
		}

		return m
	})

	s = mentionPat.ReplaceAllStringFunc(s, func(m string) string {
		if strings.EqualFold(m, "@everyone") || strings.EqualFold(m, "@here") {
			return m
		}

		// Strip @
		var m1 = m[1:]

		for _, r := range g.Roles {
			if r.Mentionable && strings.EqualFold(m1, r.Name) {
				return r.Mention()
			}
		}

		// TODO: Consider adding cache, but how often do we get here realistically?
		for _, u := range g.Members {
			if u.Nick != "" && strings.EqualFold(m1, u.Nick) {
				return u.Mention()
			}
			if strings.EqualFold(m1, u.User.Username) || strings.EqualFold(m1, u.User.String()) {
				return u.User.Mention()
			}
		}

		return m
	})

	c.session.RUnlock()

	return s
}

// Run placeholder to implement Gateway interface
func (c *Channel) Run(ctx context.Context) error {
	return nil
}

// FindTrigger checks if s starts with trigger, return Trigger{} if true
func (c *Channel) FindTrigger(s string) *gateway.Trigger {
	if t := c.Config.FindTrigger(s); t != nil {
		return t
	}
	if t := gateway.FindTrigger(fmt.Sprintf("<@%s> ", c.session.State.User.ID), s); t != nil {
		return t
	}
	return nil
}

func fmtDuration(d time.Duration) string {
	var s = (int64)(d.Round(time.Second) / time.Second)
	var m = s / 60
	var h = m / 60

	if h > 0 {
		m -= h * 60
		return fmt.Sprintf("%dh%dm", h, m)
	} else if m > 0 {
		s -= m * 60
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

var plural = map[bool]string{
	false: "",
	true:  "s",
}

func (c *Channel) updateOnline() {
	c.omut.Lock()
	if c.ochan == nil {
		c.ochan = make(chan struct{}, 1)

		var t = time.NewTicker(time.Minute)
		go func() {
			for c.session.State.User == nil {
				time.Sleep(time.Second)
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

				var now = time.Now()

				c.omut.Lock()
				var content = fmt.Sprintf("__💬 **Online**: %d user%s__\n\n", len(c.online), plural[len(c.online) != 1])
				for _, o := range c.online {
					var a = ""
					switch {
					case o.User.Access >= gateway.AccessOperator:
						a = "⚔️"
					case o.User.Access >= gateway.AccessWhitelist:
						a = "⭐"
					case o.User.Access >= gateway.AccessVoice:
						a = "🔈"
					case o.User.Access < gateway.AccessVoice:
						a = "🔇"
					case o.User.Access <= gateway.AccessBan:
						a = "💩"
					}

					var s = fmt.Sprintf("%s `%-15s@ %-9s\u200B` *%s*\n", a, o.User.Name, o.Discr, fmtDuration(now.Sub(o.Since)))
					if len(content)+len(s) >= 2000 {
						break
					}

					content += s
				}
				c.omut.Unlock()

				content += "\u200B"
				if content == last {
					continue
				}

				last = content

				if c.OnlineListID == "" {
					if m, err := c.session.ChannelMessageSend(c.chanID, content); err != nil {
						c.Fire(&network.AsyncError{Src: "updateOnline[Send]", Err: err})
					} else {
						c.OnlineListID = m.ID
					}
					continue
				}

				if _, err := c.session.ChannelMessageEdit(c.chanID, c.OnlineListID, content); err != nil {
					c.Fire(&network.AsyncError{Src: "updateOnline[Update]", Err: err})

					switch err := err.(type) {
					case *discordgo.RESTError:
						if err.Message != nil && err.Message.Code == discordgo.ErrCodeUnknownMessage {
							// Message has been deleted
							c.OnlineListID = ""
						}
					}
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

func (c *Channel) sortOnline() {
	c.omut.Lock()
	sort.Slice(c.online, func(i, j int) bool {
		if c.online[i].User.Access == c.online[j].User.Access {
			return c.online[i].Since.Before(c.online[j].Since)
		}
		return c.online[i].User.Access > c.online[j].User.Access
	})
	c.omut.Unlock()
}

func (c *Channel) clearOnline(gw string) {
	c.omut.Lock()
	var n = make([]online, 0, len(c.online))
	for _, o := range c.online {
		if o.Gateway == gw {
			continue
		}
		n = append(n, o)
	}
	c.online = n
	c.omut.Unlock()
}

// Relay dumps the event content in channel
func (c *Channel) Relay(ev *network.Event, from gateway.Gateway) error {
	switch msg := ev.Arg.(type) {
	case *gateway.Connected:
		if strings.HasPrefix(c.ID(), from.ID()+gateway.Delimiter) {
			return nil
		}
		return c.say(fmt.Sprintf("🔗 *Established connection to `%s`*", from.ID()))
	case *gateway.Disconnected:
		if strings.HasPrefix(c.ID(), from.ID()+gateway.Delimiter) {
			return nil
		}
		return c.say(fmt.Sprintf("🔗 *Connection to `%s` closed*", from.ID()))
	case *network.AsyncError:
		return c.say(fmt.Sprintf("❗ **%s** `ERROR` %s", from.Discriminator(), msg.Error()))
	case *gateway.SystemMessage:
		return c.say(fmt.Sprintf("📢 **%s** `%s` %s", from.Discriminator(), msg.Type, msg.Content))
	case *gateway.Channel:
		return c.say(fmt.Sprintf("💬 *Joined channel `%s@%s`*", msg.Name, from.Discriminator()))

	case *gateway.Clear:
		if c.RelayJoins&RelayJoinsList != 0 {
			c.clearOnline(from.ID())
			c.updateOnline()
		}
		return nil
	case *gateway.Join:
		if c.RelayJoins&RelayJoinsList != 0 {
			c.omut.Lock()
			c.online = append(c.online, online{
				Gateway: from.ID(),
				Discr:   from.Discriminator(),
				User:    msg.User,
				Since:   time.Now(),
			})
			c.omut.Unlock()
			c.sortOnline()
			c.updateOnline()
		}

		if c.RelayJoins == 0 || c.RelayJoins&RelayJoinsSay != 0 {
			return c.say(fmt.Sprintf("➡️ **%s@%s** has joined the channel", msg.User.Name, from.Discriminator()))
		}

		return nil
	case *gateway.User:
		if c.RelayJoins&RelayJoinsList != 0 {
			c.omut.Lock()
			for i := range c.online {
				if c.online[i].Gateway != from.ID() || c.online[i].User.ID != msg.ID {
					continue
				}
				c.online[i].User = *msg
				break
			}
			c.omut.Unlock()
			c.sortOnline()
			c.updateOnline()
		}
		return nil
	case *gateway.Leave:
		if c.RelayJoins&RelayJoinsList != 0 {
			c.omut.Lock()
			for i := len(c.online) - 1; i >= 0; i-- {
				if c.online[i].Gateway != from.ID() || c.online[i].User.ID != msg.User.ID {
					continue
				}
				c.online = append(c.online[:i], c.online[i+1:]...)
				break
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
			Content:   c.parse(msg.Content, msg.User.Access),
			Username:  fmt.Sprintf("%s@%s (Direct Message)", msg.User.Name, from.Discriminator()),
			AvatarURL: msg.User.AvatarURL,
		})
	case *gateway.Chat:
		return c.WebhookOrSay(&discordgo.WebhookParams{
			Content:   c.parse(msg.Content, msg.User.Access),
			Username:  fmt.Sprintf("%s@%s", msg.User.Name, from.Discriminator()),
			AvatarURL: msg.User.AvatarURL,
		})
	case *gateway.Say:
		var p = &discordgo.WebhookParams{
			Content:  c.parse(msg.Content, gateway.AccessDefault),
			Username: from.Discriminator(),
		}
		if c.session.State.User != nil {
			p.AvatarURL = c.session.State.User.AvatarURL("")
		}
		return c.WebhookOrSay(p)

	default:
		return gateway.ErrUnknownEvent
	}
}

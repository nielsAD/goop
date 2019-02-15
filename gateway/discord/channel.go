// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package discord

import (
	"context"
	"fmt"
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
	Gateway string
	Name    string
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

// Channel currently being monitoring\
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

		var online = false
		if p, err := c.session.State.Presence(c.guildID, uid); err == nil && p.Status != discordgo.StatusOffline {
			online = true
		}

		if online {
			if u, err := c.User(uid); err == nil && u != nil {
				c.Fire(&gateway.Leave{User: *u})
			}
		}

		c.AccessUser[uid] = a

		if online {
			if u, err := c.User(uid); err == nil && u != nil {
				c.Fire(&gateway.Join{User: *u})
			}
		}
	} else {
		delete(c.AccessUser, uid)
	}

	c.Fire(&gateway.ConfigUpdate{})
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
	return nil
}

// FindTrigger checks if s starts with trigger, returns cmd and args if true
func (c *Channel) FindTrigger(s string) (bool, string, []string) {
	if r, cmd, arg := c.Config.FindTrigger(s); r {
		return r, cmd, arg
	}
	if r, cmd, arg := gateway.FindTrigger(fmt.Sprintf("<@%s> ", c.session.State.User.ID), s); r {
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

			for c.session.State.User == nil {
				time.Sleep(time.Second)
			}

			if messages, err := c.session.ChannelMessagesPinned(c.chanID); err == nil {
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
					var s = fmt.Sprintf("\n`%s` *%s*", c.online[i].Name, time.Now().Sub(c.online[i].Since).Round(time.Second).String())
					if len(content)+len(s) > 2000 {
						break
					}

					content += s
				}
				c.omut.Unlock()

				if content == last {
					continue
				}

				last = content

				if msg == "" {
					if m, err := c.session.ChannelMessageSend(c.chanID, content); err != nil {
						c.Fire(&network.AsyncError{Src: "updateOnline[Send]", Err: err})
					} else {
						msg = m.ID
					}
					continue
				}

				if _, err := c.session.ChannelMessageEdit(c.chanID, msg, content); err != nil {
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
				Name:    fmt.Sprintf("%s@%s", msg.User.Name, from.Discriminator()),
				Since:   time.Now(),
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
				if c.online[i].Gateway != from.ID() || c.online[i].Name != name {
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
			Content:  c.filter(msg.Content, gateway.AccessDefault),
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

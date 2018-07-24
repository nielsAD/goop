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

// Config stores the configuration of a Discord session
type Config struct {
	AuthToken       string
	Channels        map[string]*ChannelConfig
	Presence        string
	AccessDM        gateway.AccessLevel
	AccessNoChannel gateway.AccessLevel
	AccessUser      map[string]gateway.AccessLevel
}

// ChannelConfig stores the configuration of a single Discord channel
type ChannelConfig struct {
	CommandTrigger string
	BufSize        uint8
	Webhook        string
	AccessMentions gateway.AccessLevel
	AccessTalk     gateway.AccessLevel
	AccessRole     map[string]gateway.AccessLevel
	AccessUser     map[string]gateway.AccessLevel
}

// Gateway manages a Discord connection
type Gateway struct {
	network.EventEmitter
	*discordgo.Session

	// Set once before Run(), read-only after that
	*Config
	Channels map[string]*Channel
}

// Channel manages a Discord channel
type Channel struct {
	network.EventEmitter

	wg      *sync.WaitGroup
	id      string
	session *discordgo.Session

	smut  sync.Mutex
	say   chan string
	saywh chan *discordgo.WebhookParams

	*ChannelConfig
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
	}

	var wg sync.WaitGroup
	wg.Add(1)
	d.Once(gateway.Connected{}, func(ev *network.Event) {
		go func() {
			time.Sleep(time.Second)
			wg.Done()
		}()
	})

	for id, c := range d.Config.Channels {
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
	d.AddHandler(d.onConnect)
	d.AddHandler(d.onDisconnect)
	d.AddHandler(d.onPresenceUpdate)
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
	d.Fire(gateway.Connected{})
}

func (d *Gateway) onDisconnect(s *discordgo.Session, msg *discordgo.Disconnect) {
	d.Fire(gateway.Disconnected{})
}

func (d *Gateway) onPresenceUpdate(s *discordgo.Session, msg *discordgo.PresenceUpdate) {
	old, _ := d.Session.State.Presence(msg.GuildID, msg.User.ID)
	if old == nil || msg.Presence.Status != old.Status {
		fmt.Println(msg)
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

	var chat = gateway.Chat{
		User: gateway.User{
			ID:        msg.Author.ID,
			Name:      msg.Author.Username,
			Access:    d.AccessNoChannel,
			AvatarURL: msg.Author.AvatarURL(""),
		},
		Channel: gateway.Channel{
			ID:   msg.ChannelID,
			Name: msg.ChannelID,
		},
		Content: replaceContentReferences(s, msg.Message),
	}

	var channel = d.Channels[msg.ChannelID]
	if channel != nil {
		chat.User.Access = channel.AccessTalk
	}

	if ch, err := s.State.Channel(msg.ChannelID); err == nil {
		if member, err := s.State.Member(ch.GuildID, msg.Author.ID); err == nil {
			if member.Nick != "" {
				chat.User.Name = member.Nick
			}

			if channel != nil && channel.AccessRole != nil {
				for _, rid := range member.Roles {
					r, err := s.State.Role(ch.GuildID, rid)
					if err != nil {
						continue
					}
					var access, ok = channel.AccessRole[strings.ToLower(r.Name)]
					if ok {
						chat.User.Access = access
					}
				}
			}
		}

		if channel == nil && ch.Type == discordgo.ChannelTypeDM {
			chat.User.Access = d.AccessDM

			var access, ok = d.AccessUser[msg.Author.ID]
			if ok {
				chat.User.Access = access
			}

			d.Fire(&gateway.PrivateChat{
				User:    chat.User,
				Content: chat.Content,
			})
			return
		}

		if g, err := s.State.Guild(ch.GuildID); err == nil {
			chat.Channel.Name = fmt.Sprintf("%s.%s", g.Name, ch.Name)
		} else {
			chat.Channel.Name = ch.Name
		}
	}

	var access, ok = d.AccessUser[msg.Author.ID]
	if ok {
		chat.User.Access = access
	}

	if channel != nil {
		var access, ok = channel.AccessUser[msg.Author.ID]
		if ok {
			chat.User.Access = access
		}

		channel.Fire(&chat)
	} else {
		d.Fire(&chat)
	}
}

// Relay placeholder to implement Realm interface
// Events should instead be relayed directly to a Channel
func (d *Gateway) Relay(ev *network.Event, sender string) {
}

// Run placeholder to implement Realm interface
func (c *Channel) Run(ctx context.Context) error {
	var done = make(chan struct{})

	go func() {
		c.wg.Wait()
		var name = c.id
		if ch, err := c.session.State.Channel(c.id); err == nil {
			if g, err := c.session.State.Guild(ch.GuildID); err == nil {
				name = fmt.Sprintf("%s.%s", g.Name, ch.Name)
			} else {
				name = ch.Name
			}
		}
		c.Fire(&gateway.Channel{ID: c.id, Name: name})
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	return nil
}

// Say sends a chat message
func (c *Channel) Say(s string) error {
	c.smut.Lock()
	if c.say == nil {
		c.say = make(chan string, c.BufSize)

		go func() {
			for s := range c.say {
				_, err := c.session.ChannelMessageSend(c.id, s)
				if err != nil {
					c.Fire(&network.AsyncError{Src: "Say", Err: err})
				}
			}
		}()
	}
	c.smut.Unlock()

	select {
	case c.say <- s:
		return nil
	default:
		return ErrSayBufferFull
	}
}

// WebhookOrSay sends a chat message preferably via webhook
func (c *Channel) WebhookOrSay(p *discordgo.WebhookParams) error {
	if c.Webhook == "" {
		var s = p.Content
		if p.Username != "" {
			s = fmt.Sprintf("**<%s>** %s", p.Username, p.Content)
		}
		return c.Say(s)
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
		s = strings.Replace(s, "@", "@"+string('\u200B'), -1)
	}
	return s
}

// Relay dumps the event content in channel
func (c *Channel) Relay(ev *network.Event, sender string) {
	var err error

	sender = strings.SplitN(sender, gateway.Delimiter, 2)[0]

	switch msg := ev.Arg.(type) {
	case gateway.Connected:
		err = c.Say(fmt.Sprintf("*Established connection to %s*", sender))
	case gateway.Disconnected:
		err = c.Say(fmt.Sprintf("*Connection to %s closed*", sender))
	case *gateway.Channel:
		err = c.Say(fmt.Sprintf("*Joined %s on %s*", msg.Name, sender))
	case *gateway.SystemMessage:
		err = c.Say(fmt.Sprintf("📢 **%s** %s", sender, msg.Content))
	case *gateway.Join:
		err = c.Say(fmt.Sprintf("➡️ **%s@%s** has joined the channel", msg.User.Name, sender))
	case *gateway.Leave:
		err = c.Say(fmt.Sprintf("⬅️ **%s@%s** has left the channel", msg.User.Name, sender))
	case *gateway.Chat:
		err = c.WebhookOrSay(&discordgo.WebhookParams{
			Content:   c.filter(msg.Content, msg.User.Access),
			Username:  fmt.Sprintf("%s@%s", msg.User.Name, sender),
			AvatarURL: msg.User.AvatarURL,
		})
	case *gateway.PrivateChat:
		err = c.WebhookOrSay(&discordgo.WebhookParams{
			Content:   c.filter(msg.Content, msg.User.Access),
			Username:  fmt.Sprintf("%s@%s (Direct Message)", msg.User.Name, sender),
			AvatarURL: msg.User.AvatarURL,
		})
	default:
		err = gateway.ErrUnknownEvent
	}

	if err != nil && !network.IsConnClosedError(err) {
		c.Fire(&network.AsyncError{Src: "Relay", Err: err})
	}
}

// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package capi

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/chat"
	pcapi "github.com/nielsAD/gowarcraft3/protocol/capi"
)

// Errors
var (
	ErrSayBufferFull = errors.New("gw-capi: Say buffer full")
)

// Config stores the configuration of a single CAPI connection
type Config struct {
	GatewayConfig
	chat.Config
}

// GatewayConfig stores the config additions of capi.Gateway over chat.Bot
type GatewayConfig struct {
	gateway.Config

	ReconnectDelay   time.Duration
	BufSize          uint8
	AvatarDefaultURL string

	AccessWhisper  gateway.AccessLevel
	AccessTalk     gateway.AccessLevel
	AccessOperator gateway.AccessLevel
	AccessUser     map[string]gateway.AccessLevel
}

// Gateway manages a CAPI connection
type Gateway struct {
	gateway.Common
	*chat.Bot

	chatmut sync.Mutex
	users   map[string]int64
	name    string

	smut  sync.Mutex
	saych chan string

	// Set once before Run(), read-only after that
	*GatewayConfig
}

// New initializes a new Gateway struct
func New(conf *Config) (*Gateway, error) {
	c, err := chat.NewBot(&conf.Config)
	if err != nil {
		return nil, err
	}

	var b = Gateway{
		Bot:           c,
		GatewayConfig: &conf.GatewayConfig,
	}

	b.InitDefaultHandlers()

	return &b, nil
}

// Operator in chat
func (b *Gateway) Operator() bool {
	if u, ok := b.Bot.User(1); ok {
		return u.Operator()
	}
	return false
}

// Channel residing in
func (b *Gateway) Channel() *gateway.Channel {
	var name = b.Bot.Channel()
	if name == "" {
		return nil
	}
	return &gateway.Channel{
		ID:   name,
		Name: name,
	}
}

// ChannelUsers online
func (b *Gateway) ChannelUsers() []gateway.User {
	var users = b.Bot.Users()

	var res = make([]gateway.User, 0, len(users))
	for k, u := range users {
		if k == 1 {
			continue
		}
		res = append(res, b.userFromCapi(&u))
	}

	return res
}

// User by ID
func (b *Gateway) User(uid string) (*gateway.User, error) {
	if id, ok := b.uid(uid); ok {
		var res = b.userFromID(id)
		return &res, nil
	}

	var s = strings.ToLower(uid)
	if access := b.AccessUser[s]; access != gateway.AccessDefault {
		return &gateway.User{
			ID:        s,
			Name:      uid,
			Access:    access,
			AvatarURL: b.AvatarDefaultURL,
		}, nil
	}

	return nil, gateway.ErrNoUser
}

// Users with non-default access level
func (b *Gateway) Users() map[string]gateway.AccessLevel {
	return b.AccessUser
}

// SetUserAccess overrides accesslevel for a specific user
func (b *Gateway) SetUserAccess(uid string, a gateway.AccessLevel) (*gateway.AccessLevel, error) {
	uid = strings.ToLower(uid)
	if uid == "" {
		return nil, gateway.ErrNoUser
	}

	var o = b.AccessUser[uid]
	if a != gateway.AccessDefault {
		if b.AccessUser == nil {
			b.AccessUser = make(map[string]gateway.AccessLevel)
		}

		b.AccessUser[uid] = a
	} else {
		delete(b.AccessUser, uid)
	}

	b.Fire(&gateway.ConfigUpdate{})

	if id, inchat := b.users[uid]; inchat {
		var u = b.userFromID(id)
		b.Fire(&u)
	}

	return &o, nil
}

func (b *Gateway) say(s string) error {
	b.smut.Lock()
	if b.saych == nil {
		b.saych = make(chan string, b.BufSize)

		go func() {
			for s := range b.saych {
				err := b.Bot.SendMessage(s)
				if err != nil && !network.IsCloseError(err) {
					b.Fire(&network.AsyncError{Src: "Say", Err: err})
				}
			}
		}()
	}
	b.smut.Unlock()

	select {
	case b.saych <- s:
		return nil
	default:
		return ErrSayBufferFull
	}
}

// Say sends a chat message
func (b *Gateway) Say(s string) error {
	if err := b.say(s); err != nil {
		return err
	}
	b.Fire(&gateway.Say{Content: s})
	return nil
}

// SayPrivate sends a private chat message to uid
func (b *Gateway) SayPrivate(uid string, s string) error {
	id, ok := b.uid(uid)
	if !ok {
		return gateway.ErrNoUser
	}
	go func() {
		err := b.Bot.SendWhisper(id, s)
		if err != nil {
			b.Fire(&network.AsyncError{Src: "SayPrivate", Err: err})
		}
	}()
	return nil
}

// Kick user from channel
func (b *Gateway) Kick(uid string) error {
	if !b.Operator() {
		return gateway.ErrNoPermission
	}
	id, ok := b.uid(uid)
	if !ok {
		return gateway.ErrNoUser
	}
	go func() {
		err := b.Bot.KickUser(id)
		if err != nil {
			b.Fire(&network.AsyncError{Src: "Kick", Err: err})
		}
	}()
	return nil
}

// Ban user from channel
func (b *Gateway) Ban(uid string) error {
	if !b.Operator() {
		return gateway.ErrNoPermission
	}
	id, ok := b.uid(uid)
	if !ok {
		return gateway.ErrNoUser
	}
	go func() {
		err := b.Bot.BanUser(id)
		if err != nil {
			b.Fire(&network.AsyncError{Src: "Ban", Err: err})
		}
	}()
	return nil
}

// Unban user from channel
func (b *Gateway) Unban(uid string) error {
	if !b.Operator() {
		return gateway.ErrNoPermission
	}
	go func() {
		err := b.Bot.UnbanUser(uid)
		if err != nil {
			b.Fire(&network.AsyncError{Src: "Unban", Err: err})
		}
	}()
	return nil
}

// Ping user to calculate RTT in milliseconds
func (b *Gateway) Ping(uid string) (time.Duration, error) {
	return 0, gateway.ErrNotImplemented
}

// Run reads packets and emits an event for each received packet
func (b *Gateway) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		b.Bot.Close()
	}()

	var backoff = b.ReconnectDelay
	for ctx.Err() == nil {
		if backoff < 10*time.Second {
			backoff = 10 * time.Second
		} else if backoff > 4*time.Hour {
			backoff = 4 * time.Hour
		}

		var err = b.Bot.Connect()
		if err != nil {
			var reconnect bool
			switch err {
			case chat.ErrUnexpectedPacket, websocket.ErrBadHandshake:
				reconnect = true
			default:
				reconnect = network.IsTemporary(err) || network.IsCloseError(err) || network.IsUnexpectedCloseError(err)
			}

			if reconnect && ctx.Err() == nil {
				b.Fire(&network.AsyncError{Src: "Run[Connect]", Err: err})

				select {
				case <-time.After(backoff):
				case <-ctx.Done():
				}

				backoff = time.Duration(float64(backoff) * 2)
				continue
			}

			return err
		}

		b.Fire(&gateway.Connected{})

		backoff = b.ReconnectDelay
		if err := b.Bot.Run(); err != nil && ctx.Err() == nil {
			b.Fire(&network.AsyncError{Src: "Run[Bot]", Err: err})
		}

		b.Fire(&gateway.Disconnected{})
		b.Fire(&gateway.Clear{})
		b.SetConn(nil)
	}

	return ctx.Err()
}

func (b *Gateway) userFromCapi(u *chat.User) gateway.User {
	var res = gateway.User{
		ID:        strings.ToLower(u.Username),
		Name:      u.Username,
		Access:    b.AccessTalk,
		AvatarURL: b.AvatarDefaultURL,
	}

	if b.AccessOperator != gateway.AccessDefault && u.Operator() {
		res.Access = b.AccessOperator
	}

	if access := b.AccessUser[res.ID]; access != gateway.AccessDefault {
		res.Access = access
	}

	return res
}

func (b *Gateway) userFromID(uid int64) gateway.User {
	if u, ok := b.Bot.User(uid); ok {
		return b.userFromCapi(u)
	}

	return gateway.User{
		Access:    b.AccessTalk,
		AvatarURL: b.AvatarDefaultURL,
	}
}

func (b *Gateway) uid(uid string) (int64, bool) {
	b.chatmut.Lock()
	var res, ok = b.users[uid]
	b.chatmut.Unlock()
	return res, ok
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (b *Gateway) InitDefaultHandlers() {
	b.On(&pcapi.ConnectEvent{}, b.onConnectEvent)
	b.On(&pcapi.MessageEvent{}, b.onMessageEvent)
	b.On(&chat.UserJoined{}, b.onUserJoined)
	b.On(&chat.UserUpdate{}, b.onUserUpdate)
	b.On(&chat.UserLeft{}, b.onUserLeft)
}

func (b *Gateway) onConnectEvent(ev *network.Event) {
	var pkt = ev.Arg.(*pcapi.ConnectEvent)

	b.chatmut.Lock()
	b.users = nil
	b.chatmut.Unlock()

	b.Fire(&gateway.Clear{})
	b.Fire(&gateway.Channel{ID: pkt.Channel, Name: pkt.Channel})
}

func (b *Gateway) onUserJoined(ev *network.Event) {
	var u = ev.Arg.(*chat.UserJoined)
	if u.UserID == 1 {
		b.name = u.Username
		return
	}

	b.chatmut.Lock()
	if b.users == nil {
		b.users = make(map[string]int64)
	}
	b.users[strings.ToLower(u.Username)] = u.UserID
	b.chatmut.Unlock()

	b.Fire(&gateway.Join{User: b.userFromCapi(&u.User)})
}

func (b *Gateway) onUserUpdate(ev *network.Event) {
	var u = ev.Arg.(*chat.UserUpdate)
	if u.UserID == 1 {
		return
	}

	var user = b.userFromCapi(&u.User)
	b.Fire(&user)
}

func (b *Gateway) onUserLeft(ev *network.Event) {
	var u = ev.Arg.(*chat.UserLeft)
	if u.UserID == 1 {
		return
	}

	b.chatmut.Lock()
	var id = strings.ToLower(u.Username)
	var leave = b.users[id] == u.UserID
	if leave {
		delete(b.users, id)
	}
	b.chatmut.Unlock()

	if leave {
		b.Fire(&gateway.Leave{User: b.userFromCapi(&u.User)})
	}
}

// FindTrigger checks if s starts with trigger, return Trigger{} if true
func (b *Gateway) FindTrigger(s string) *gateway.Trigger {
	if t := b.GatewayConfig.FindTrigger(s); t != nil {
		return t
	}

	idx := strings.IndexAny(s, ",:")
	if idx <= 0 || idx+2 >= len(s) || s[idx+1] != ' ' {
		return nil
	}

	pat := s[:idx]
	if !strings.EqualFold(pat, "goop") || strings.EqualFold(pat, "all") || (strings.EqualFold(pat, "ops") && b.Operator()) {
		if m, _ := filepath.Match(pat, b.name); !m || len(b.name) == 0 {
			return nil
		}
	}

	return gateway.ExtractTrigger(s[idx+2:])
}

var banPat = regexp.MustCompile(`^([^ ]+) was banned by ([^ ]+).*\.$`)
var unbanPat = regexp.MustCompile(`^([^ ]+) was unbanned by ([^ ]+).*\.$`)

func (b *Gateway) onMessageEvent(ev *network.Event) {
	var pkt = ev.Arg.(*pcapi.MessageEvent)
	if pkt.Message == "" {
		return
	}

	switch pkt.Type {
	case pcapi.MessageEmote, pcapi.MessageChannel, pcapi.MessageWhisper:
		var u = b.userFromID(pkt.UserID)

		var ev interface{}

		switch pkt.Type {
		case pcapi.MessageWhisper:
			ev = &gateway.PrivateChat{
				User:    u,
				Content: pkt.Message,
			}
		case pcapi.MessageEmote:
			ev = &gateway.Chat{
				User:    u,
				Content: fmt.Sprintf("%s %s", u.Name, pkt.Message),
			}
		default:
			ev = &gateway.Chat{
				User:    u,
				Content: pkt.Message,
			}
		}

		b.Fire(ev)

		if u.Access < b.Commands.Access {
			return
		}

		if t := b.FindTrigger(pkt.Message); t != nil {
			t.User = u
			t.Resp = b.Responder(b, u.ID, pkt.Type == pcapi.MessageWhisper)
			b.Fire(t, ev)
		}
	case pcapi.MessageServerInfo:
		if pkt.Message == "No one hears you." {
			return
		}

		// Persist bans/unbans
		if m := banPat.FindStringSubmatch(pkt.Message); m != nil {
			var u = strings.ToLower(m[1])
			if access := b.AccessUser[u]; access > gateway.AccessBan && access < gateway.AccessWhitelist {
				b.SetUserAccess(u, gateway.AccessBan)
			}
		} else if m := unbanPat.FindStringSubmatch(pkt.Message); m != nil {
			var u = strings.ToLower(m[1])
			if access := b.AccessUser[u]; access < gateway.AccessDefault {
				b.SetUserAccess(u, gateway.AccessDefault)
			}
		}

		fallthrough
	default:
		b.Fire(&gateway.SystemMessage{Type: strings.ToUpper(pkt.Type.String()), Content: pkt.Message})
	}
}

// Relay dumps the event content in current channel
func (b *Gateway) Relay(ev *network.Event, from gateway.Gateway) error {
	switch msg := ev.Arg.(type) {
	case *gateway.Clear:
		return nil
	case *gateway.User:
		return nil
	case *gateway.Connected:
		return b.say(fmt.Sprintf("Established connection to %s", from.ID()))
	case *gateway.Disconnected:
		return b.say(fmt.Sprintf("Connection to %s closed", from.ID()))
	case *network.AsyncError:
		return b.say(fmt.Sprintf("[%s] [ERROR] %s", from.Discriminator(), msg.Error()))
	case *gateway.SystemMessage:
		return b.say(fmt.Sprintf("[%s] [%s] %s", from.Discriminator(), msg.Type, msg.Content))
	case *gateway.Channel:
		return b.say(fmt.Sprintf("Joined channel %s@%s", msg.Name, from.Discriminator()))
	case *gateway.Join:
		return b.say(fmt.Sprintf("%s@%s has joined the channel", msg.User.Name, from.Discriminator()))
	case *gateway.Leave:
		return b.say(fmt.Sprintf("%s@%s has left the channel", msg.User.Name, from.Discriminator()))
	case *gateway.PrivateChat:
		return b.say(fmt.Sprintf("[DM] <%s@%s> %s", msg.User.Name, from.Discriminator(), msg.Content))
	case *gateway.Chat:
		return b.say(fmt.Sprintf("<%s@%s> %s", msg.User.Name, from.Discriminator(), msg.Content))
	case *gateway.Say:
		return b.say(fmt.Sprintf("<%s> %s", from.Discriminator(), msg.Content))
	default:
		return gateway.ErrUnknownEvent
	}
}

// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package bnet

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/bnet"
	"github.com/nielsAD/gowarcraft3/protocol/bncs"
	"github.com/nielsAD/gowarcraft3/protocol/w3gs"
)

// Errors
var (
	ErrSayBufferFull = errors.New("gw-bnet: Say buffer full")
	ErrSayCommand    = errors.New("gw-bnet: Say prevented execution of command")
)

// Config stores the configuration of a single BNet server
type Config struct {
	GatewayConfig
	bnet.Config
}

// GatewayConfig stores the config additions of bnet.Gateway over bnet.Client
type GatewayConfig struct {
	gateway.Config

	ReconnectDelay   time.Duration
	HomeChannel      string
	BufSize          uint8
	AvatarIconURL    string
	AvatarDefaultURL string

	AccessWhisper    gateway.AccessLevel
	AccessTalk       gateway.AccessLevel
	AccessNoWarcraft gateway.AccessLevel
	AccessOperator   gateway.AccessLevel
	AccessLevel      map[int]gateway.AccessLevel
	AccessClanTag    map[string]gateway.AccessLevel
	AccessUser       map[string]gateway.AccessLevel
}

// Gateway manages a BNet connection
type Gateway struct {
	gateway.Common
	*bnet.Client

	smut  sync.Mutex
	saych chan string

	// Set once before Run(), read-only after that
	*GatewayConfig
}

// New initializes a new Gateway struct
func New(conf *Config) (*Gateway, error) {
	c, err := bnet.NewClient(&conf.Config)
	if err != nil {
		return nil, err
	}

	var b = Gateway{
		Client:        c,
		GatewayConfig: &conf.GatewayConfig,
	}

	b.InitDefaultHandlers()

	return &b, nil
}

// Operator in chat
func (b *Gateway) Operator() bool {
	if u, ok := b.Client.User(b.UniqueName); ok {
		return u.Operator()
	}
	return false
}

// Channel residing in
func (b *Gateway) Channel() *gateway.Channel {
	var name = b.Client.Channel()
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
	var users = b.Client.Users()

	var res = make([]gateway.User, 0, len(users))
	for _, u := range users {
		res = append(res, b.user(&u))
	}

	return res
}

// User by ID
func (b *Gateway) User(uid string) (*gateway.User, error) {
	if u, ok := b.Client.User(uid); ok {
		var res = b.user(u)
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

		if u, ok := b.Client.User(uid); ok {
			b.Fire(&gateway.Leave{User: b.user(u)})
		}

		b.AccessUser[uid] = a

		if u, ok := b.Client.User(uid); ok {
			b.Fire(&gateway.Join{User: b.user(u)})
		}
	} else {
		delete(b.AccessUser, uid)
	}

	b.Fire(&gateway.ConfigUpdate{})
	return &o, nil
}

func (b *Gateway) say(s string) error {
	b.smut.Lock()
	if b.saych == nil {
		b.saych = make(chan string, b.BufSize)

		go func() {
			for s := range b.saych {
				err := b.Client.Say(s)
				if err != nil {
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
	if strings.HasPrefix(s, "/") {
		return ErrSayCommand
	}
	if err := b.say(s); err != nil {
		return err
	}
	b.Fire(&gateway.Say{Content: s})
	return nil
}

// SayPrivate sends a private chat message to uid
func (b *Gateway) SayPrivate(uid string, s string) error {
	return b.say(fmt.Sprintf("/w %s %s", uid, s))
}

// Kick user from channel
func (b *Gateway) Kick(uid string) error {
	if !b.Operator() {
		return gateway.ErrNoPermission
	}
	return b.say(fmt.Sprintf("/kick %s", uid))
}

// Ban user from channel
func (b *Gateway) Ban(uid string) error {
	if !b.Operator() {
		return gateway.ErrNoPermission
	}
	return b.say(fmt.Sprintf("/ban %s", uid))
}

// Unban user from channel
func (b *Gateway) Unban(uid string) error {
	if !b.Operator() {
		return gateway.ErrNoPermission
	}
	return b.say(fmt.Sprintf("/unban %s", uid))
}

// Ping user to calculate RTT in milliseconds
func (b *Gateway) Ping(uid string) (time.Duration, error) {
	u, ok := b.Client.User(uid)
	if !ok {
		return 0, gateway.ErrNoUser
	}
	return time.Duration(u.Ping) * time.Millisecond, nil
}

// Run reads packets and emits an event for each received packet
func (b *Gateway) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		b.Client.Close()
	}()

	var backoff = b.ReconnectDelay
	for ctx.Err() == nil {
		if backoff < 10*time.Second {
			backoff = 10 * time.Second
		} else if backoff > 4*time.Hour {
			backoff = 4 * time.Hour
		}

		var err = b.Client.Logon()
		if err != nil {
			var reconnect bool
			switch err {
			case bnet.ErrCDKeyInUse, bnet.ErrUnexpectedPacket:
				reconnect = true
			default:
				reconnect = network.IsConnClosedError(err) || os.IsTimeout(err)
			}

			if reconnect && ctx.Err() == nil {
				b.Fire(&network.AsyncError{Src: "Run[Logon]", Err: err})

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

		var channel = b.Client.Channel()
		if channel == "" {
			channel = b.HomeChannel
		}
		if channel != "" {
			b.say("/join " + channel)
		}

		backoff = b.ReconnectDelay
		if err := b.Client.Run(); err != nil && ctx.Err() == nil {
			b.Fire(&network.AsyncError{Src: "Run[Client]", Err: err})
		}

		b.Fire(&gateway.Disconnected{})
		b.Fire(&gateway.Clear{})
	}

	return ctx.Err()
}

func (b *Gateway) user(u *bnet.User) gateway.User {
	var res = gateway.User{
		ID:        strings.ToLower(u.Name),
		Name:      u.Name,
		Access:    b.AccessTalk,
		AvatarURL: b.AvatarDefaultURL,
	}

	var prod, icon, lvl, tag = u.Stat()
	if prod != 0 {
		switch prod {
		case w3gs.ProductDemo, w3gs.ProductROC, w3gs.ProductTFT:
			if b.AvatarIconURL != "" {
				res.AvatarURL = strings.Replace(b.AvatarIconURL, "${ICON}", icon.String(), -1)
			}

			var max = 0
			for l, a := range b.AccessLevel {
				if l >= max && lvl >= l {
					max = l
					res.Access = a
				}
			}

			if access := b.AccessClanTag[tag.String()]; access != gateway.AccessDefault {
				res.Access = access
			}
		default:
			if b.AccessNoWarcraft != gateway.AccessDefault {
				res.Access = b.AccessNoWarcraft
			}
		}
	}

	if b.AccessOperator != gateway.AccessDefault && u.Operator() {
		res.Access = b.AccessOperator
	}

	if access := b.AccessUser[res.ID]; access != gateway.AccessDefault {
		res.Access = access
	}

	return res
}

// InitDefaultHandlers adds the default callbacks for relevant packets
func (b *Gateway) InitDefaultHandlers() {
	b.On(&bnet.UserJoined{}, b.onUserJoined)
	b.On(&bnet.UserLeft{}, b.onUserLeft)
	b.On(&bnet.Chat{}, b.onChat)
	b.On(&bnet.Whisper{}, b.onWhisper)
	b.On(&bnet.JoinError{}, b.onJoinError)
	b.On(&bnet.Channel{}, b.onChannel)
	b.On(&bnet.SystemMessage{}, b.onSystemMessage)
	b.On(&bncs.FloodDetected{}, b.onFloodDetected)
}

func (b *Gateway) onUserJoined(ev *network.Event) {
	var user = ev.Arg.(*bnet.UserJoined)
	if user.Name == b.UniqueName {
		return
	}

	b.Fire(&gateway.Join{User: b.user(&user.User)})
}

func (b *Gateway) onUserLeft(ev *network.Event) {
	var user = ev.Arg.(*bnet.UserLeft)
	if user.Name == b.UniqueName {
		return
	}

	b.Fire(&gateway.Leave{User: b.user(&user.User)})
}

func extractCmdAndArgs(s string) (bool, string, []string) {
	if len(s) < 1 || s[0] == ' ' {
		return false, "", nil
	}
	f := strings.Fields(s)
	return true, f[0], f[1:]
}

// FindTrigger checks if s starts with trigger, returns cmd and args if true
func (b *Gateway) FindTrigger(s string) (bool, string, []string) {
	if r, cmd, arg := b.GatewayConfig.FindTrigger(s); r {
		return r, cmd, arg
	}

	idx := strings.IndexAny(s, ",:")
	if idx <= 0 || idx+2 >= len(s) || s[idx+1] != ' ' {
		return false, "", nil
	}

	pat := s[:idx]
	if !strings.EqualFold(pat, "goop") || strings.EqualFold(pat, "all") || (strings.EqualFold(pat, "ops") && b.Operator()) {
		if m, _ := filepath.Match(pat, b.UniqueName); !m {
			return false, "", nil
		}
	}

	return extractCmdAndArgs(s[idx+2:])
}

func (b *Gateway) onChat(ev *network.Event) {
	var msg = ev.Arg.(*bnet.Chat)
	if msg.Content == "" {
		return
	}

	var chat = gateway.Chat{
		User:    b.user(&msg.User),
		Content: msg.Content,
	}

	switch msg.Type {
	case bncs.ChatEmote:
		chat.Content = fmt.Sprintf("%s %s", msg.User.Name, msg.Content)
	default:
		chat.Content = msg.Content
	}

	b.Fire(&chat)

	if chat.User.Access < b.Commands.Access {
		return
	}

	if r, cmd, arg := b.FindTrigger(chat.Content); r {
		b.Fire(&gateway.Trigger{
			User: chat.User,
			Cmd:  cmd,
			Arg:  arg,
			Resp: b.Responder(b, chat.User.ID, false),
		}, &chat)
	}
}

func (b *Gateway) onWhisper(ev *network.Event) {
	var msg = ev.Arg.(*bnet.Whisper)
	if msg.Content == "" {
		return
	}

	if msg.Username[:1] == "#" {
		b.Fire(&gateway.SystemMessage{Type: msg.Username, Content: msg.Content})
		return
	}

	var chat = gateway.PrivateChat{
		User: gateway.User{
			ID:        strings.ToLower(msg.Username),
			Name:      msg.Username,
			Access:    b.AccessWhisper,
			AvatarURL: b.AvatarDefaultURL,
		},
		Content: msg.Content,
	}

	if access := b.AccessUser[chat.ID]; access != gateway.AccessDefault {
		chat.User.Access = access
	}

	b.Fire(&chat)

	if chat.User.Access < b.Commands.Access {
		return
	}

	if r, cmd, arg := b.FindTrigger(chat.Content); r {
		b.Fire(&gateway.Trigger{
			User: chat.User,
			Cmd:  cmd,
			Arg:  arg,
			Resp: b.Responder(b, chat.User.ID, true),
		}, &chat)
	}
}

func (b *Gateway) onJoinError(ev *network.Event) {
	var err = ev.Arg.(*bnet.JoinError)
	b.Fire(&gateway.SystemMessage{Type: "JOINERROR", Content: fmt.Sprintf("Could not join %s (%s)", err.Channel, err.Error.String())})
}

func (b *Gateway) onChannel(ev *network.Event) {
	var c = ev.Arg.(*bnet.Channel)
	b.Fire(&gateway.Clear{})
	b.Fire(&gateway.Channel{ID: c.Name, Name: c.Name})
}

var banPat = regexp.MustCompile("^([^ ]+) was banned by ([^ ]+).*\\.$")
var unbanPat = regexp.MustCompile("^([^ ]+) was unbanned by ([^ ]+).*\\.$")

func (b *Gateway) onSystemMessage(ev *network.Event) {
	var msg = ev.Arg.(*bnet.SystemMessage)

	if msg.Type == bncs.ChatInfo {
		if msg.Content == "No one hears you." {
			return
		}

		// Persist bans/unbans
		if m := banPat.FindStringSubmatch(msg.Content); m != nil {
			var u = strings.ToLower(m[1])
			if access := b.AccessUser[u]; access > gateway.AccessBan && access < gateway.AccessWhitelist {
				b.SetUserAccess(u, gateway.AccessBan)
			}
		} else if m := unbanPat.FindStringSubmatch(msg.Content); m != nil {
			var u = strings.ToLower(m[1])
			if access := b.AccessUser[u]; access < gateway.AccessDefault {
				b.SetUserAccess(u, gateway.AccessDefault)
			}
		}
	}

	b.Fire(&gateway.SystemMessage{Type: strings.ToUpper(msg.Type.String()), Content: msg.Content})
}

func (b *Gateway) onFloodDetected(ev *network.Event) {
	b.Fire(&gateway.SystemMessage{Type: "FLOOD", Content: "Flood detected"})
}

// Relay dumps the event content in current channel
func (b *Gateway) Relay(ev *network.Event, from gateway.Gateway) error {
	switch msg := ev.Arg.(type) {
	case *gateway.Clear:
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

// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package stdio

import (
	"bufio"
	"context"
	"io"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/gowarcraft3/network"
)

// Config stores the gateway configuration
type Config struct {
	gateway.Config
	Read      bool
	Access    gateway.AccessLevel
	AvatarURL string
}

// Gateway relays between stdin/stdout
type Gateway struct {
	gateway.Common
	network.EventEmitter

	*Config
	In  io.ReadCloser
	Out *log.Logger
}

// New initializes a new Gateway struct
func New(in io.ReadCloser, out *log.Logger, conf *Config) *Gateway {
	return &Gateway{
		Config: conf,
		In:     in,
		Out:    out,
	}
}

const stdioUID = "stdio"

func (o *Gateway) read() error {
	var r = bufio.NewReader(o.In)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		if line == "" {
			continue
		}

		var chat = gateway.PrivateChat{
			User: gateway.User{
				ID:        stdioUID,
				Name:      "stdin",
				Access:    o.Access,
				AvatarURL: o.AvatarURL,
			},
			Content: line,
		}

		o.Fire(&chat)

		if chat.User.Access < o.Commands.Access {
			continue
		}

		if t := o.Config.FindTrigger(chat.Content); t != nil {
			t.User = chat.User
			t.Resp = o.Responder(o, stdioUID, true)
			o.Fire(t, &chat)
		}
	}
}

// Channel residing in
func (o *Gateway) Channel() *gateway.Channel {
	return nil
}

// ChannelUsers online
func (o *Gateway) ChannelUsers() []gateway.User {
	return nil
}

// User by ID
func (o *Gateway) User(uid string) (*gateway.User, error) {
	return nil, gateway.ErrNoUser
}

// Users with non-default access level
func (o *Gateway) Users() map[string]gateway.AccessLevel {
	return nil
}

// SetUserAccess overrides accesslevel for a specific user
func (o *Gateway) SetUserAccess(uid string, a gateway.AccessLevel) (*gateway.AccessLevel, error) {
	return nil, gateway.ErrNotImplemented
}

// Say sends a chat message
func (o *Gateway) Say(s string) error {
	o.Out.Printf("[%s][SAY] %s\n", o.ID(), s)
	o.Fire(&gateway.Say{Content: s})
	return nil
}

// SayPrivate sends a private chat message to uid
func (o *Gateway) SayPrivate(uid string, s string) error {
	if uid != stdioUID {
		return gateway.ErrNoChannel
	}

	o.Out.Println(color.GreenString("[%s][SAYP] %s", o.ID(), s))
	return nil
}

// Kick user from channel
func (o *Gateway) Kick(uid string) error {
	return gateway.ErrNotImplemented
}

// Ban user from channel
func (o *Gateway) Ban(uid string) error {
	return gateway.ErrNotImplemented
}

// Unban user from channel
func (o *Gateway) Unban(uid string) error {
	return gateway.ErrNotImplemented
}

// Ping user to calculate RTT in milliseconds
func (o *Gateway) Ping(uid string) (time.Duration, error) {
	return 0, gateway.ErrNotImplemented
}

// Run reads packets and emits an event for each received packet
func (o *Gateway) Run(ctx context.Context) error {
	if !o.Read {
		return nil
	}

	defer o.In.Close()

	var res = make(chan error)
	go func() {
		res <- o.read()
	}()

	select {
	case err := <-res:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Relay dumps the event content to stdout
func (o *Gateway) Relay(ev *network.Event, from gateway.Gateway) error {
	switch msg := ev.Arg.(type) {
	case *gateway.Connected:
		o.Out.Println(color.MagentaString("[%s] Established connection", from.ID()))
	case *gateway.Disconnected:
		o.Out.Println(color.MagentaString("[%s] Connection closed", from.ID()))
	case *gateway.Clear:
		o.Out.Println(color.MagentaString("[%s] Cleared", from.ID()))
	case *network.AsyncError:
		o.Out.Println(color.RedString("[%s][ERR] %s", from.ID(), msg.Error()))
	case *gateway.SystemMessage:
		o.Out.Println(color.CyanString("[%s][%s] %s", from.ID(), msg.Type, msg.Content))
	case *gateway.Channel:
		o.Out.Println(color.MagentaString("[%s] Joined %s@%s", from.ID(), msg.Name, from.Discriminator()))
	case *gateway.Join:
		o.Out.Println(color.YellowString("[%s][CHAT] %s@%s has joined the channel", from.ID(), msg.User.Name, from.Discriminator()))
	case *gateway.Leave:
		o.Out.Println(color.YellowString("[%s][CHAT] %s@%s has left the channel", from.ID(), msg.User.Name, from.Discriminator()))
	case *gateway.User:
		o.Out.Println(color.YellowString("[%s][CHAT] %s@%s updated", from.ID(), msg.Name, from.Discriminator()))
	case *gateway.PrivateChat:
		o.Out.Println(color.GreenString("[%s][PRIV] <%s@%s> %s", from.ID(), msg.User.Name, from.Discriminator(), msg.Content))
	case *gateway.Chat:
		o.Out.Printf("[%s][CHAT] <%s@%s> %s\n", from.ID(), msg.User.Name, from.Discriminator(), msg.Content)
	case *gateway.Say:
		o.Out.Printf("[%s][CHAT] <%s> %s\n", from.ID(), from.Discriminator(), msg.Content)
	default:
		return gateway.ErrUnknownEvent
	}

	return nil
}

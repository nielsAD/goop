// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package stdio

import (
	"bufio"
	"context"
	"log"
	"strings"
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
	In  *bufio.Reader
	Out *log.Logger
}

// New initializes a new Gateway struct
func New(in *bufio.Reader, out *log.Logger, conf *Config) *Gateway {
	return &Gateway{
		Config: conf,
		In:     in,
		Out:    out,
	}
}

func (o *Gateway) read() error {
	for {
		line, err := o.In.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		if line == "" {
			continue
		}

		var chat = gateway.PrivateChat{
			User: gateway.User{
				ID:        "stdio",
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

		if r, cmd, arg := o.Config.FindTrigger(chat.Content); r {
			o.Fire(&gateway.Trigger{
				User: chat.User,
				Cmd:  cmd,
				Arg:  arg,
				Resp: o.Say,
			}, chat)
		}
	}
}

// Channel currently being monitoring
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

// AddUser overrides accesslevel for a specific user
func (o *Gateway) AddUser(uid string, a gateway.AccessLevel) (*gateway.AccessLevel, error) {
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
	if uid != "stdio" {
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
	return gateway.ErrNoChannel
}

// Run reads packets and emits an event for each received packet
func (o *Gateway) Run(ctx context.Context) error {
	if !o.Read {
		return nil
	}

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

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
	Read           bool
	Access         gateway.AccessLevel
	CommandTrigger string
	AvatarURL      string
}

// Gateway relays between stdin/stdout
type Gateway struct {
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

		o.Fire(&gateway.PrivateChat{
			User: gateway.User{
				ID:        "stdio",
				Name:      "stdin",
				Access:    o.Access,
				AvatarURL: o.AvatarURL,
			},
			Content: line,
		})
	}
}

// Discriminator unique among gateways
func (o *Gateway) Discriminator() string {
	return "stdio"
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
func (o *Gateway) Relay(ev *network.Event) {
	var sender = ev.Opt[1].(string)

	switch msg := ev.Arg.(type) {
	case *gateway.Connected:
		o.Out.Println(color.MagentaString("Established connection to %s", sender))
	case *gateway.Disconnected:
		o.Out.Println(color.MagentaString("Connection to %s closed", sender))
	case *gateway.Channel:
		o.Out.Println(color.MagentaString("Joined %s on %s", msg.Name, sender))
	case *gateway.SystemMessage:
		o.Out.Println(color.CyanString("[SYSTEM][%s] %s", sender, msg.Content))
	case *gateway.Join:
		o.Out.Println(color.YellowString("[CHAT][%s#%s] %s has joined the channel", sender, msg.Channel.Name, msg.User.Name))
	case *gateway.Leave:
		o.Out.Println(color.YellowString("[CHAT][%s#%s] %s has left the channel", sender, msg.Channel.Name, msg.User.Name))
	case *gateway.Chat:
		o.Out.Printf("[CHAT][%s#%s] <%s> %s\n", sender, msg.Channel.Name, msg.User.Name, msg.Content)
	case *gateway.PrivateChat:
		o.Out.Println(color.GreenString("[PRIVATE][%s] <%s> %s", sender, msg.User.Name, msg.Content))
	default:
		o.Fire(&network.AsyncError{Src: "Relay", Err: gateway.ErrUnknownEvent})
	}
}

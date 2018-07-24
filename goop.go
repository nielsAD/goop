// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/bnet"
	"github.com/nielsAD/goop/gateway/discord"
	"github.com/nielsAD/goop/gateway/stdio"
	"github.com/nielsAD/gowarcraft3/network"
)

// Errors
var (
	ErrUnkownRealm      = errors.New("goop: Unknown realm")
	ErrUnknownConfigKey = errors.New("goop: Unknown config key")
	ErrInvalidType      = errors.New("goop: Type mismatch")
)

// Goop main
type Goop struct {
	network.EventEmitter

	// Read-only
	Gateways map[string]gateway.Gateway
	Config   Config
}

// New initializes a Goop struct
func New(conf *Config) (*Goop, error) {
	var res = Goop{
		Config: *conf,
		Gateways: map[string]gateway.Gateway{
			"STDIO": stdio.New(bufio.NewReader(os.Stdin), logOut, &conf.StdIO),
		},
	}

	var gateways = []string{"STDIO"}

	for k, g := range res.Config.BNet.Gateways {
		gw, err := bnet.New(g)
		if err != nil {
			return nil, err
		}

		res.Gateways[k] = gw
		gateways = append(gateways, k)
	}

	for k, g := range res.Config.Discord.Gateways {
		gw, err := discord.New(g)
		if err != nil {
			return nil, err
		}

		res.Gateways[k] = gw
		gateways = append(gateways, k)

		for cid, c := range gw.Channels {
			var idx = k + gateway.Delimiter + cid
			res.Gateways[idx] = c
			gateways = append(gateways, idx)
		}
	}

	for i := 0; i < len(res.Config.Relay); i++ {
		var r = res.Config.Relay[i]
		if r.In == nil {
			r.In = gateways
		}
		if r.Out == nil {
			r.Out = gateways
		}
		for _, in := range r.In {
			var r1 = res.Gateways[in]
			if r1 == nil {
				return nil, ErrUnkownRealm
			}

			for _, out := range r.Out {
				var r2 = res.Gateways[out]
				if r2 == nil {
					return nil, ErrUnkownRealm
				}
				if r1 == r2 {
					continue
				}

				var sender = in
				var handler = func(ev *network.Event) { r2.Relay(ev, sender) }

				if r.Log {
					r1.On(gateway.Connected{}, handler)
					r1.On(gateway.Disconnected{}, handler)
					r1.On(&gateway.Channel{}, handler)
				}
				if r.System {
					r1.On(&gateway.SystemMessage{}, handler)
				}

				if r.Joins {
					r1.On(&gateway.Join{}, func(ev *network.Event) {
						var user = ev.Arg.(*gateway.Join)
						if user.Access < r.JoinAccess {
							return
						}
						r2.Relay(ev, sender)
					})
					r1.On(&gateway.Leave{}, func(ev *network.Event) {
						var user = ev.Arg.(*gateway.Leave)
						if user.Access < r.JoinAccess {
							return
						}
						r2.Relay(ev, sender)
					})
				}

				if r.Chat {
					r1.On(&gateway.Chat{}, func(ev *network.Event) {
						var msg = ev.Arg.(*gateway.Chat)
						if msg.User.Access < r.ChatAccess {
							return
						}
						r2.Relay(ev, sender)
					})
				}

				if r.PrivateChat {
					r1.On(&gateway.PrivateChat{}, func(ev *network.Event) {
						var msg = ev.Arg.(*gateway.PrivateChat)
						if msg.User.Access < r.PrivateChatAccess {
							return
						}
						r2.Relay(ev, sender)
					})
				}
			}
		}
	}

	return &res, nil
}

// Run connects to each realm and returns when all connections have ended
func (g *Goop) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := range g.Gateways {
		wg.Add(1)

		var k = i
		var r = g.Gateways[k]
		go func() {
			if err := r.Run(ctx); err != nil && err != context.Canceled {
				g.Fire(&network.AsyncError{Src: fmt.Sprintf("Run[gw:%s]", k), Err: err})
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

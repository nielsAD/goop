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
	ErrUnkownRealm    = errors.New("goop: Unknown realm")
	ErrDuplicateRealm = errors.New("goop: Duplicate realm")
)

// Goop main
type Goop struct {
	network.EventEmitter

	// Read-only
	Gateways map[string]gateway.Gateway
	Relay    map[string]map[string]*Relay
	Config   Config
}

// New initializes a Goop struct
func New(conf *Config) (*Goop, error) {
	var res = Goop{
		Config:   *conf,
		Relay:    map[string]map[string]*Relay{},
		Gateways: map[string]gateway.Gateway{},
	}

	if err := res.add("std"+gateway.Delimiter+"io", stdio.New(bufio.NewReader(os.Stdin), logOut, &conf.StdIO)); err != nil {
		return nil, err
	}

	for k, g := range res.Config.BNet.Gateways {
		gw, err := bnet.New(g)
		if err != nil {
			return nil, err
		}

		k = "bnet" + gateway.Delimiter + k
		res.add(k, gw)
	}

	for k, g := range res.Config.Discord.Gateways {
		gw, err := discord.New(g)
		if err != nil {
			return nil, err
		}

		k = "discord" + gateway.Delimiter + k
		res.add(k, gw)

		for cid, c := range gw.Channels {
			res.add(k+gateway.Delimiter+cid, c)
		}
	}

	for g1, r := range res.Config.Relay.To {
		if res.Gateways[g1] == nil {
			return nil, ErrUnkownRealm
		}
		for g2 := range r.From {
			if res.Gateways[g2] == nil {
				return nil, ErrUnkownRealm
			}
		}
	}

	res.InitDefaultHandlers()

	return &res, nil
}

func (g *Goop) newRelay(to, from string) *Relay {
	if g.Config.Relay.To[to] == nil {
		g.Config.Relay.To[to] = &RelayToConfig{
			Default: g.Config.Relay.Default,
		}
	}
	if g.Config.Relay.To[to].From == nil {
		g.Config.Relay.To[to].From = make(map[string]*RelayConfig)
	}
	if g.Config.Relay.To[to].From[from] == nil {
		var cfg = g.Config.Relay.To[to].Default
		g.Config.Relay.To[to].From[from] = &cfg
	}
	return NewRelay(g.Gateways[from], g.Gateways[to], g.Config.Relay.To[to].From[from])
}

func (g *Goop) add(id string, gw gateway.Gateway) error {
	if g.Gateways[id] != nil {
		return ErrDuplicateRealm
	}

	g.Gateways[id] = gw
	g.Relay[id] = make(map[string]*Relay)

	// These handlers are called after relay handlers
	gw.On(&gateway.Chat{}, checkTriggerChat)
	gw.On(&gateway.PrivateChat{}, checkTriggerPrivateChat)

	for wid := range g.Gateways {
		if id == wid {
			continue
		}

		g.Relay[id][wid] = g.newRelay(id, wid)
		g.Relay[wid][id] = g.newRelay(wid, id)
	}

	gw.On(nil, func(ev *network.Event) {
		// Add sender and sender_id info to all events
		ev.Opt = append([]network.EventArg{gw, id}, ev.Opt...)

		// Fire on main object (called before relay handlers)
		if g.Fire(ev.Arg, ev.Opt...) {
			ev.PreventNext()
		}
	})

	return nil
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

// InitDefaultHandlers adds the default callbacks for relevant packets
func (g *Goop) InitDefaultHandlers() {
	g.On(&gateway.Chat{}, g.onChat)
	g.On(&gateway.PrivateChat{}, g.onPrivateChat)
	g.On(&gateway.Command{}, g.onCommand)
	g.On(&gateway.Join{}, g.onJoin)
	g.On(&gateway.Leave{}, g.onLeave)
}

func checkTrigger(ev *network.Event, s string, u *gateway.User) {
	if s == "?trigger" {
		if f, ok := ev.Opt[0].(network.Firer); ok {
			f.Fire(&gateway.Command{
				User: *u,
				Cmd:  "trigger",
			}, ev.Arg)
		}
	}
}

func checkTriggerChat(ev *network.Event) {
	var msg = ev.Arg.(*gateway.Chat)
	checkTrigger(ev, msg.Content, &msg.User)
}

func checkTriggerPrivateChat(ev *network.Event) {
	var msg = ev.Arg.(*gateway.PrivateChat)
	checkTrigger(ev, msg.Content, &msg.User)
}

func (g *Goop) onChat(ev *network.Event) {
	//var msg = ev.Arg.(*gateway.Chat)
}

func (g *Goop) onPrivateChat(ev *network.Event) {
	//var msg = ev.Arg.(*gateway.PrivateChat)
}

func (g *Goop) onCommand(ev *network.Event) {
	//var cmd = ev.Arg.(*gateway.Command)
}

func (g *Goop) onJoin(ev *network.Event) {
	//var user = ev.Arg.(*gateway.Join)
}

func (g *Goop) onLeave(ev *network.Event) {
	//var user = ev.Arg.(*gateway.Leave)
}

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
	return NewRelay(from, g.Gateways[from], g.Gateways[to], g.Config.Relay.To[to].From[from])
}

func (g *Goop) add(id string, gw gateway.Gateway) error {
	if g.Gateways[id] != nil {
		return ErrDuplicateRealm
	}

	g.Gateways[id] = gw
	g.Relay[id] = make(map[string]*Relay)

	for wid := range g.Gateways {
		if id == wid {
			continue
		}

		g.Relay[id][wid] = g.newRelay(id, wid)
		g.Relay[wid][id] = g.newRelay(wid, id)
	}

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

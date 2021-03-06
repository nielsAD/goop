// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package goop

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/gowarcraft3/network"
)

// Errors
var (
	ErrDuplicateGateway = errors.New("goop: Duplicate gateway")
	ErrDuplicateCommand = errors.New("goop: Duplicate command")
)

// Config interface
type Config interface {
	GetRelay(to, from string) *RelayConfig

	Map() map[string]interface{}
	FlatMap() map[string]interface{}
	Get(key string) (interface{}, error)
	Set(key string, val interface{}) error
	Unset(key string) (err error)
	GetString(key string) (string, error)
	SetString(key string, val string) error
}

// Command interface
type Command interface {
	CanExecute(t *gateway.Trigger) bool
	Execute(t *gateway.Trigger, gw gateway.Gateway, g *Goop) error
}

// Goop main
type Goop struct {
	network.EventEmitter

	// Read-only
	Commands map[string]Command
	Gateways map[string]gateway.Gateway
	Relay    map[string]map[string]*Relay
	Config   Config
}

// New initializes a Goop struct
func New(conf Config) *Goop {
	var res = &Goop{
		Commands: map[string]Command{},
		Gateways: map[string]gateway.Gateway{},
		Relay:    map[string]map[string]*Relay{},
		Config:   conf,
	}

	return res
}

// AddGateway to goop
func (g *Goop) AddGateway(id string, gw gateway.Gateway) error {
	if g.Gateways[id] != nil {
		return ErrDuplicateGateway
	}

	gw.SetID(id)

	g.Gateways[id] = gw
	g.Relay[id] = make(map[string]*Relay)

	// These handlers are called after relay handlers
	gw.On(&gateway.Chat{}, checkTriggerChat)
	gw.On(&gateway.PrivateChat{}, checkTriggerPrivateChat)

	gw.On(&gateway.Trigger{}, g.execTrigger)
	gw.On(&gateway.Chat{}, g.autoKickChat)
	gw.On(&gateway.Join{}, g.autoKickJoin)

	for wid := range g.Gateways {
		g.Relay[id][wid] = NewRelay(g.Gateways[wid], g.Gateways[id], g.Config.GetRelay(id, wid))
		if id == wid {
			continue
		}
		g.Relay[wid][id] = NewRelay(g.Gateways[id], g.Gateways[wid], g.Config.GetRelay(wid, id))
	}

	gw.On(nil, func(ev *network.Event) {
		// Add sender to all events
		ev.Opt = append([]network.EventArg{gw}, ev.Opt...)

		// Fire on main object (called before relay handlers)
		if g.Fire(ev.Arg, ev.Opt...) {
			ev.PreventNext()
		}
	})

	g.Fire(&NewGateway{
		Gateway: gw,
	})

	return nil
}

// AddCommand to goop
func (g *Goop) AddCommand(name string, c Command) error {
	name = strings.ToLower(name)
	if g.Commands[name] != nil {
		return ErrDuplicateCommand
	}

	g.Commands[name] = c

	g.Fire(&NewCommand{
		Name:    name,
		Command: c,
	})

	return nil
}

// Run connects to each gateway and returns when all connections have ended
func (g *Goop) Run(ctx context.Context) {
	g.Fire(Start{})

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
	g.Fire(Stop{})
}

func checkTriggerChat(ev *network.Event) {
	var msg = ev.Arg.(*gateway.Chat)
	if msg.User.Access < gateway.AccessVoice || !strings.EqualFold(msg.Content, "?trigger") {
		return
	}
	gw, ok := ev.Opt[0].(gateway.Gateway)
	if !ok || gw.Trigger() == "?" {
		return
	}

	gw.Fire(&gateway.Trigger{
		User: msg.User,
		Cmd:  "trigger",
		Resp: gw.Responder(gw, msg.User.ID, false),
	}, ev.Arg)
}

func checkTriggerPrivateChat(ev *network.Event) {
	var msg = ev.Arg.(*gateway.PrivateChat)
	if msg.User.Access < gateway.AccessVoice || !strings.EqualFold(msg.Content, "?trigger") {
		return
	}
	gw, ok := ev.Opt[0].(gateway.Gateway)
	if !ok || gw.Trigger() == "?" {
		return
	}

	gw.Fire(&gateway.Trigger{
		User: msg.User,
		Cmd:  "trigger",
		Resp: gw.Responder(gw, msg.User.ID, true),
	}, ev.Arg)
}

func (g *Goop) execTrigger(ev *network.Event) {
	var t = *ev.Arg.(*gateway.Trigger)

	t.Cmd = strings.ToLower(t.Cmd)
	if c, ok := g.Commands[strings.ToLower(t.Cmd)]; ok {
		gw, ok := ev.Opt[0].(gateway.Gateway)
		if !ok || !c.CanExecute(&t) {
			return
		}
		go func() {
			if err := c.Execute(&t, gw, g); err != nil {
				g.Fire(&network.AsyncError{Src: "execTrigger", Err: err})
			}
		}()
		return
	}

	var s = strings.Split(t.Cmd, gateway.Delimiter)
	if len(s) < 2 {
		return
	}

	t.Cmd = s[len(s)-1]
	c, ok := g.Commands[t.Cmd]
	if !ok || !c.CanExecute(&t) {
		return
	}

	var p = strings.ToLower(fmt.Sprintf("*%s%s%s*", gateway.Delimiter, strings.Join(s[:len(s)-1], gateway.Delimiter), gateway.Delimiter))
	for k := range g.Gateways {
		var gw = g.Gateways[k]
		if ok, err := filepath.Match(p, gateway.Delimiter+strings.ToLower(k)+gateway.Delimiter); err != nil || !ok {
			continue
		}
		if gw.Channel() == nil {
			continue
		}

		var trig = t
		trig.Resp = func(s string) error { return t.Resp(fmt.Sprintf("[%s] %s", gw.Discriminator(), s)) }

		go func() {
			if err := c.Execute(&trig, gw, g); err != nil {
				g.Fire(&network.AsyncError{Src: "execTrigger", Err: err})
			}
		}()
	}
}

func (g *Goop) autoKick(gw gateway.Gateway, u *gateway.User) bool {
	var err error
	if u.Access <= gateway.AccessBan {
		err = gw.Ban(u.ID)
	} else {
		err = gw.Kick(u.ID)
	}

	switch err {
	case nil:
		return true
	case gateway.ErrNotImplemented, gateway.ErrNoPermission:
		// ignore
	default:
		g.Fire(&network.AsyncError{Src: "autoKick", Err: err})
	}

	return false
}

func (g *Goop) autoKickChat(ev *network.Event) {
	var msg = ev.Arg.(*gateway.Chat)
	if msg.User.Access > gateway.AccessKick {
		return
	}
	gw, ok := ev.Opt[0].(gateway.Gateway)
	if !ok {
		return
	}

	if g.autoKick(gw, &msg.User) {
		ev.PreventNext()
	}
}

func (g *Goop) autoKickJoin(ev *network.Event) {
	var user = ev.Arg.(*gateway.Join)
	if user.Access > gateway.AccessKick {
		return
	}
	gw, ok := ev.Opt[0].(gateway.Gateway)
	if !ok {
		return
	}

	if g.autoKick(gw, &user.User) {
		ev.PreventNext()
	}
}

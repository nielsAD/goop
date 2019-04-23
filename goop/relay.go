// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package goop

import (
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/gowarcraft3/network"
)

// RelayConfig stores the configuration of a gateway relay
type RelayConfig struct {
	Log         bool
	System      bool
	Channel     bool
	Joins       bool
	Chat        bool
	PrivateChat bool
	Say         bool

	JoinAccess        gateway.AccessLevel
	ChatAccess        gateway.AccessLevel
	PrivateChatAccess gateway.AccessLevel
}

// Relay manages a relay between two gateways
type Relay struct {
	From gateway.Gateway
	To   gateway.Gateway

	*RelayConfig
}

// NewRelay initializes a new GatRelayeway struct
func NewRelay(from, to gateway.Gateway, conf *RelayConfig) *Relay {
	var r = Relay{
		From:        from,
		To:          to,
		RelayConfig: conf,
	}
	r.InitDefaultHandlers()
	return &r
}

// InitDefaultHandlers adds the default callbacks for relevant events
func (r *Relay) InitDefaultHandlers() {
	r.From.On(&network.AsyncError{}, r.onLog)
	r.From.On(&gateway.Connected{}, r.onLog)
	r.From.On(&gateway.Disconnected{}, r.onLog)
	r.From.On(&gateway.Channel{}, r.onChannel)
	r.From.On(&gateway.SystemMessage{}, r.onSystemMessage)
	r.From.On(&gateway.Clear{}, r.onClear)
	r.From.On(&gateway.Join{}, r.onJoin)
	r.From.On(&gateway.User{}, r.onUser)
	r.From.On(&gateway.Leave{}, r.onLeave)
	r.From.On(&gateway.Chat{}, r.onChat)
	r.From.On(&gateway.PrivateChat{}, r.onPrivateChat)
	r.From.On(&gateway.Say{}, r.onSay)
}

func (r *Relay) relay(ev *network.Event) {
	err := r.To.Relay(ev, r.From)
	if err == nil || network.IsCloseError(err) {
		return
	}

	r.To.Fire(&network.AsyncError{Src: "Relay", Err: err}, ev)
}

func (r *Relay) onLog(ev *network.Event) {
	if !r.Log {
		return
	}
	r.relay(ev)
}

func (r *Relay) onClear(ev *network.Event) {
	if !r.Log && !r.Joins {
		return
	}
	r.relay(ev)
}

func (r *Relay) onSystemMessage(ev *network.Event) {
	if !r.System {
		return
	}
	r.relay(ev)
}

func (r *Relay) onChannel(ev *network.Event) {
	if !r.Channel {
		return
	}
	r.relay(ev)
}

func (r *Relay) onJoin(ev *network.Event) {
	var user = ev.Arg.(*gateway.Join)
	if !r.Joins || user.Access < r.JoinAccess {
		return
	}
	r.relay(ev)
}

func (r *Relay) onUser(ev *network.Event) {
	var user = ev.Arg.(*gateway.User)
	if !r.Joins || user.Access < r.JoinAccess {
		return
	}
	r.relay(ev)
}

func (r *Relay) onLeave(ev *network.Event) {
	var user = ev.Arg.(*gateway.Leave)
	if !r.Joins || user.Access < r.JoinAccess {
		return
	}
	r.relay(ev)
}

func (r *Relay) onChat(ev *network.Event) {
	var msg = ev.Arg.(*gateway.Chat)
	if !r.Chat || msg.User.Access < r.ChatAccess {
		return
	}
	r.relay(ev)
}

func (r *Relay) onPrivateChat(ev *network.Event) {
	var msg = ev.Arg.(*gateway.PrivateChat)
	if !r.PrivateChat || msg.User.Access < r.PrivateChatAccess {
		return
	}
	r.relay(ev)
}

func (r *Relay) onSay(ev *network.Event) {
	if !r.Say {
		return
	}
	r.relay(ev)
}

// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Kick user
type Kick struct{ Cmd }

// Execute command
func (c *Kick) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 arguments: [user]")
	}
	var u = gateway.FindUser(gw, t.Arg[0])
	switch len(u) {
	case 0:
		return t.Resp(MsgNoUserFound)
	case 1:
		return gw.Kick(u[0].ID)
	default:
		return t.Resp(MsgMoreUserFound)
	}
}

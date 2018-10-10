// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"strings"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Say echoes input
type Say struct{ Cmd }

// Execute command
func (c *Say) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	return gw.Say(strings.Join(t.Arg, " "))
}

// SayPrivate forwards input to user in private
type SayPrivate struct{ Cmd }

// Execute command
func (c *SayPrivate) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 2 {
		return t.Resp("Expected 2 arguments: [user] [message]")
	}
	var u = gateway.FindUser(gw, t.Arg[0])
	switch len(u) {
	case 0:
		return t.Resp(MsgNoUserFound)
	case 1:
		return gw.SayPrivate(u[0].ID, strings.Join(t.Arg[1:], " "))
	default:
		return t.Resp(MsgMoreUserFound)
	}
}

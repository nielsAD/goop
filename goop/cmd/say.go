// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"strings"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Echo input
type Echo struct{ Cmd }

// Execute command
func (c *Echo) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	return t.Resp(strings.Join(t.Arg, " "))
}

// Say input in channel
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
		u = []*gateway.User{&gateway.User{ID: t.Arg[0]}}
		fallthrough
	case 1:
		if err := gw.SayPrivate(u[0].ID, strings.Join(t.Arg[1:], " ")); err != nil && err != gateway.ErrNotImplemented {
			t.Resp(MsgInternalError)
			return err
		}
		return nil
	default:
		return t.Resp(MsgMoreUserFound)
	}
}

// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"fmt"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Ban user
type Ban struct{ Cmd }

// Execute command
func (c *Ban) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [user]")
	}
	var u = gateway.FindUser(gw, t.Arg[0])
	switch len(u) {
	case 0:
		u = []*gateway.User{&gateway.User{ID: t.Arg[0]}}
		fallthrough
	case 1:
		err := gw.Ban(u[0].ID)
		switch err {
		case nil:
			return t.Resp(fmt.Sprintf(MsgBannedUser, u[0].ID))
		case gateway.ErrNotImplemented:
			return nil
		case gateway.ErrNoPermission:
			return t.Resp(MsgNoPermission)
		default:
			return err
		}
	default:
		return t.Resp(MsgMoreUserFound)
	}
}

// Unban user
type Unban struct{ Cmd }

// Execute command
func (c *Unban) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [user]")
	}
	var u = gateway.FindUser(gw, t.Arg[0])
	switch len(u) {
	case 0:
		u = []*gateway.User{&gateway.User{ID: t.Arg[0]}}
		fallthrough
	case 1:
		err := gw.Ban(u[0].ID)
		switch err {
		case nil:
			return t.Resp(fmt.Sprintf(MsgUnbannedUser, u[0].ID))
		case gateway.ErrNotImplemented:
			return nil
		case gateway.ErrNoPermission:
			return t.Resp(MsgNoPermission)
		default:
			return err
		}
	default:
		return t.Resp(MsgMoreUserFound)
	}
}

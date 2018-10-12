// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
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
	var users = gateway.FindUser(gw, t.Arg[0])
	if len(users) == 0 {
		users = []*gateway.User{&gateway.User{ID: t.Arg[0]}}
	}

	for _, u := range users {
		if u.Access.HasAccess(gateway.AccessWhitelist) && !t.User.HasAccess(gateway.AccessAdmin) {
			continue
		}
		err := gw.Ban(u.ID)
		switch err {
		case nil:
			//nothing
		case gateway.ErrNotImplemented:
			return nil
		case gateway.ErrNoPermission:
			return t.Resp(MsgNoPermission)
		default:
			return err
		}
	}

	return nil
}

// Unban user
type Unban struct{ Cmd }

// Execute command
func (c *Unban) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [user]")
	}
	var users = gateway.FindUser(gw, t.Arg[0])
	if len(users) == 0 {
		users = []*gateway.User{&gateway.User{ID: t.Arg[0]}}
	}

	for _, u := range users {
		if u.Access.HasAccess(gateway.AccessBlacklist) && !t.User.HasAccess(gateway.AccessAdmin) {
			continue
		}
		err := gw.Unban(u.ID)
		switch err {
		case nil:
			//nothing
		case gateway.ErrNotImplemented:
			return nil
		case gateway.ErrNoPermission:
			return t.Resp(MsgNoPermission)
		default:
			return err
		}
	}

	return nil
}

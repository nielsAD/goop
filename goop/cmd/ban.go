// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Ban user
type Ban struct {
	Cmd
	AccessProtect  gateway.AccessLevel
	AccessOverride gateway.AccessLevel
}

// Execute command
func (c *Ban) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [user]")
	}
	var users = gateway.FindUserInChannel(gw, t.Arg[0])
	if len(users) == 0 {
		users = []*gateway.User{&gateway.User{ID: t.Arg[0]}}
	}

	for _, u := range users {
		if u.Access >= t.User.Access || (u.Access >= c.AccessProtect && t.User.Access < c.AccessOverride) {
			continue
		}
		err := gw.Ban(u.ID)
		switch err {
		case nil:
			gw.AddUser(u.ID, gateway.AccessBan)
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
type Unban struct {
	Cmd
	AccessProtect  gateway.AccessLevel
	AccessOverride gateway.AccessLevel
}

// Execute command
func (c *Unban) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [user]")
	}
	var users = gateway.FindUserInChannel(gw, t.Arg[0])
	if len(users) == 0 {
		users = []*gateway.User{&gateway.User{ID: t.Arg[0]}}
	}

	for _, u := range users {
		if u.Access < c.AccessProtect && t.User.Access < c.AccessOverride {
			continue
		}
		err := gw.Unban(u.ID)
		switch err {
		case nil:
			if u.Access < gateway.AccessDefault {
				gw.AddUser(u.ID, gateway.AccessDefault)
			}
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

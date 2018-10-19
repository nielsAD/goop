// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"fmt"
	"strings"

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
		users = []*gateway.User{&gateway.User{ID: t.Arg[0], Name: t.Arg[0]}}
	}

	var l = make([]string, 0)
	for _, u := range users {
		if u.ID == t.User.ID || u.Access >= t.User.Access || (u.Access >= c.AccessProtect && t.User.Access < c.AccessOverride) {
			continue
		}
		err := gw.Ban(u.ID)
		switch err {
		case nil:
			if _, err := gw.SetUserAccess(u.ID, gateway.AccessBan); err != nil && err != gateway.ErrNotImplemented {
				t.Resp(MsgInternalError)
				return err
			}
			l = append(l, fmt.Sprintf("`%s`", u.Name))
		case gateway.ErrNotImplemented, gateway.ErrNoChannel:
			return nil
		case gateway.ErrNoPermission:
			return t.Resp(MsgNoPermission)
		default:
			t.Resp(MsgInternalError)
			return err
		}
	}

	if len(l) == 0 {
		return t.Resp(MsgNoChanges)
	}
	return t.Resp(fmt.Sprintf("Banned [%s]", strings.Join(l, ", ")))
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
	var users = gateway.FindUser(gw, t.Arg[0])
	if len(users) == 0 {
		users = []*gateway.User{&gateway.User{ID: t.Arg[0], Name: t.Arg[0]}}
	}

	var l = make([]string, 0)
	for _, u := range users {
		if u.ID == t.User.ID || u.Access < c.AccessProtect && t.User.Access < c.AccessOverride {
			continue
		}
		err := gw.Unban(u.ID)
		switch err {
		case nil:
			if u.Access < gateway.AccessDefault {
				if _, err := gw.SetUserAccess(u.ID, gateway.AccessDefault); err != nil && err != gateway.ErrNotImplemented {
					t.Resp(MsgInternalError)
					return err
				}
			}
			l = append(l, fmt.Sprintf("`%s`", u.Name))
		case gateway.ErrNotImplemented, gateway.ErrNoChannel:
			return nil
		case gateway.ErrNoPermission:
			return t.Resp(MsgNoPermission)
		default:
			t.Resp(MsgInternalError)
			return err
		}
	}

	if len(l) == 0 {
		return t.Resp(MsgNoChanges)
	}
	return t.Resp(fmt.Sprintf("Unbanned [%s]", strings.Join(l, ", ")))
}

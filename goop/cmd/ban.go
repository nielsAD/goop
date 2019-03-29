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

	var p = 0
	var l = []string{}

	for _, u := range users {
		if u.ID == t.User.ID || u.Access >= t.User.Access || (u.Access >= c.AccessProtect && t.User.Access < c.AccessOverride) {
			p++
			continue
		}

		_, err := gw.SetUserAccess(u.ID, gateway.AccessBan)
		switch err {
		case nil, gateway.ErrNotImplemented:
			// no error
		case gateway.ErrNoUser:
			t.Resp(MsgNoUserFound)
			return err
		default:
			t.Resp(MsgInternalError)
			return err
		}

		err = gw.Ban(u.ID)
		switch err {
		case nil, gateway.ErrNoUser:
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

	switch len(l) {
	case 0:
		if p == 0 {
			return t.Resp(MsgNoUserFound)
		}
		return t.Resp(MsgNoPermission)
	case 1:
		return t.Resp(fmt.Sprintf("Banned %s", l[0]))
	default:
		return t.Resp(fmt.Sprintf("Banned [%s]", strings.Join(l, ", ")))
	}
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

	var p = 0
	var l = []string{}

	for _, u := range users {
		if u.ID == t.User.ID || (u.Access <= c.AccessProtect && t.User.Access < c.AccessOverride) {
			p++
			continue
		}

		if u.Access < gateway.AccessDefault {
			_, err := gw.SetUserAccess(u.ID, gateway.AccessDefault)
			switch err {
			case nil, gateway.ErrNotImplemented:
				// no error
			case gateway.ErrNoUser:
				t.Resp(MsgNoUserFound)
				return err
			default:
				t.Resp(MsgInternalError)
				return err
			}
		}

		err := gw.Unban(u.ID)
		switch err {
		case nil, gateway.ErrNoUser:
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

	switch len(l) {
	case 0:
		if p == 0 {
			return t.Resp(MsgNoUserFound)
		}
		return t.Resp(MsgNoPermission)
	case 1:
		return t.Resp(fmt.Sprintf("Unbanned %s", l[0]))
	default:
		return t.Resp(fmt.Sprintf("Unbanned [%s]", strings.Join(l, ", ")))
	}
}

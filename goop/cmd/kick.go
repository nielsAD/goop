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

// Kick user
type Kick struct {
	Cmd
	AccessProtect  gateway.AccessLevel
	AccessOverride gateway.AccessLevel
}

// Execute command
func (c *Kick) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [user]")
	}
	var users = gateway.FindUserInChannel(gw, t.Arg[0])

	var p = 0
	var l = []string{}

	for _, u := range users {
		if u.ID == t.User.ID || u.Access >= t.User.Access || (u.Access >= c.AccessProtect && t.User.Access < c.AccessOverride) {
			p++
			continue
		}
		err := gw.Kick(u.ID)
		switch err {
		case nil:
			l = append(l, fmt.Sprintf("`%s`", u.Name))
		case gateway.ErrNoUser:
			// ignore
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
		return t.Resp(fmt.Sprintf("Kicked %s", l[0]))
	default:
		return t.Resp(fmt.Sprintf("Kicked [%s]", strings.Join(l, ", ")))
	}

}

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

// Set accesslevel for user
type Set struct {
	Cmd
	DefaultAccess gateway.AccessLevel
}

// Execute command
func (c *Set) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [user]")
	}
	var users = gateway.FindUser(gw, t.Arg[0])
	if len(users) == 0 {
		users = []*gateway.User{&gateway.User{ID: t.Arg[0], Name: t.Arg[0]}}
	}

	var access = c.DefaultAccess
	if len(t.Arg) > 1 {
		if err := access.UnmarshalText([]byte(t.Arg[1])); err != nil {
			return t.Resp("Expected 2 arguments: [user] [access]")
		}
	}

	if access >= t.User.Access {
		return t.Resp("You cannot grant this access level")
	}

	var l = make([]string, 0)
	for _, u := range users {
		if u.ID == t.User.ID || u.Access == access || u.Access >= t.User.Access {
			continue
		}
		prev, err := gw.SetUserAccess(u.ID, access)
		switch err {
		case nil:
			var action = "Promoted"
			if access < *prev {
				action = "Demoted"
			}
			l = append(l, fmt.Sprintf("%s `%s` from <%s> to <%s>", action, u.Name, prev.String(), access.String()))
		case gateway.ErrNotImplemented:
			return nil
		default:
			t.Resp(MsgInternalError)
			return err
		}
	}

	if len(l) == 0 {
		return t.Resp(MsgNoChanges)
	}
	return t.Resp(strings.Join(l, "\n"))
}

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

// List users for a given access level
type List struct{ Cmd }

// Execute command
func (c *List) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [access]")
	}

	var a1 = gateway.AccessDefault
	var a2 = gateway.AccessDefault
	if len(t.Arg) == 1 {
		if err := a1.UnmarshalText([]byte(t.Arg[0])); err != nil {
			return t.Resp("Expected 1 argument: [access]")
		}
		a2 = a1
	} else {
		var err = a1.UnmarshalText([]byte(t.Arg[0]))
		if err == nil {
			err = a2.UnmarshalText([]byte(t.Arg[1]))
		}
		if err != nil {
			return t.Resp("Expected 2 arguments: [from_access] [to_access]")
		}
		if a2 < a1 {
			a1, a2 = a2, a1
		}
	}

	var users = gw.Users()

	var l = []string{}
	for uid, a := range users {
		if a < a1 || a > a2 {
			continue
		}
		var u, err = gw.User(uid)
		if err != nil || u == nil {
			continue
		}

		l = append(l, fmt.Sprintf("`%s`", u.Name))
	}

	if len(l) == 0 {
		return t.Resp("No users found with given access")
	}
	if a1 == a2 {
		return t.Resp(fmt.Sprintf("Users with %s access: [%s]", a1.String(), strings.Join(l, ", ")))
	}
	return t.Resp(fmt.Sprintf("Users with  %s <= access <= %s: [%s]", a1.String(), a2.String(), strings.Join(l, ", ")))
}

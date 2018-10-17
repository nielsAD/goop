// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"fmt"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

func userToString(u *gateway.User, gw gateway.Gateway) string {
	return fmt.Sprintf("NAME=`%s@%s` ID=`%s@%s` ACCESS=`%s`", u.Name, gw.Discriminator(), u.ID, gw.Discriminator(), u.Access.String())
}

// Whois displays user info
type Whois struct{ Cmd }

// Execute command
func (c *Whois) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [user]")
	}
	var u = gateway.FindUser(gw, t.Arg[0])
	switch len(u) {
	case 0:
		return t.Resp(MsgNoUserFound)
	case 1:
		return t.Resp(userToString(u[0], gw))
	default:
		return t.Resp(MsgMoreUserFound)
	}
}

// Whoami displays user info
type Whoami struct{ Cmd }

// Execute command
func (c *Whoami) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	return t.Resp(userToString(&t.User, gw))
}

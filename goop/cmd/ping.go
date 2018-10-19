// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"fmt"
	"time"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Ping user
type Ping struct{ Cmd }

// Execute command
func (c *Ping) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	if len(t.Arg) < 1 {
		return t.Resp("Expected 1 argument: [user]")
	}
	var u = gateway.FindUser(gw, t.Arg[0])
	switch len(u) {
	case 0:
		return t.Resp(MsgNoUserFound)
	case 1:
		d, err := gw.Ping(u[0].ID)
		switch err {
		case nil:
			return t.Resp(fmt.Sprintf("Ping to `%s` is %dms", u[0].Name, d.Nanoseconds()/int64(time.Millisecond)))
		case gateway.ErrNotImplemented:
			return nil
		default:
			return t.Resp(MsgInternalError)
		}
	default:
		return t.Resp(MsgMoreUserFound)
	}
}

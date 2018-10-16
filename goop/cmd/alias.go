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

// Alias abbreviates a command
type Alias struct {
	Cmd
	Exe            string
	Arg            []string
	ArgExpected    int
	WithPriviledge gateway.AccessLevel
}

// Execute command
func (c *Alias) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	var trig = *t
	trig.Cmd = c.Exe

	if len(t.Arg) < c.ArgExpected {
		return t.Resp(fmt.Sprintf("Expected %d arguments", c.ArgExpected))
	}

	if c.Arg != nil {
		var r = []string{
			"%CMD%", t.Cmd,
			"%NARGS%", fmt.Sprintf("%d", len(t.Arg)),
			"%ARGS%", strings.Join(t.Arg, " "),
			"%UID%", t.User.ID,
			"%USTR%", t.User.Name,
			"%ULVL%", t.User.Access.String(),
			"%GWID%", gw.ID(),
			"%GWDIS%", gw.Discriminator(),
			"%GWDEL%", gateway.Delimiter,
		}
		for i, a := range t.Arg {
			r = append(r,
				fmt.Sprintf("%%ARG%d%%", i+1), a,
				fmt.Sprintf("%%..ARG%d%%", i+1), strings.Join(t.Arg[:i], " "),
				fmt.Sprintf("%%ARG%d..%%", i+1), strings.Join(t.Arg[i:], " "),
			)
		}
		trig.Arg = make([]string, len(c.Arg))
		for i, a := range c.Arg {
			trig.Arg[i] = strings.NewReplacer(r...).Replace(a)
		}
	}

	if c.WithPriviledge != gateway.AccessDefault {
		trig.User.Access = c.WithPriviledge
	}

	gw.Fire(&trig)
	return nil
}

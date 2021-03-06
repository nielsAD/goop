// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Trigger outputs the command trigger for gateway
type Trigger struct{ Cmd }

// Execute command
func (c *Trigger) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	return t.Resp(gw.Trigger())
}

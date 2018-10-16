// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"time"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Time prints current time
type Time struct {
	Cmd
	Format string
}

// Execute command
func (c *Time) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	return t.Resp(time.Now().Format(c.Format))
}

// Uptime prints time since start
type Uptime struct{ Cmd }

var ts = time.Now()

// Execute command
func (c *Uptime) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	return t.Resp("Uptime: " + time.Now().Sub(ts).Round(time.Second).String())
}

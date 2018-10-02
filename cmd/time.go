// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"time"

	"github.com/nielsAD/goop/gateway"
)

// Time print current time
type Time struct {
	Cmd
	Format string
}

// Execute command
func (c *Time) Execute(t *gateway.Trigger, gw gateway.Gateway) error {
	return t.Resp(time.Now().Format(c.Format))
}

// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"strings"

	"github.com/nielsAD/goop/gateway"
)

// Say echoes input
type Say struct{ Cmd }

// Execute command
func (c *Say) Execute(t *gateway.Trigger, gw gateway.Gateway) error {
	return t.Resp(strings.Join(t.Arg, " "))
}

// SayPrivate forwards input to user in private
type SayPrivate struct{ Cmd }

// Execute command
func (c *SayPrivate) Execute(t *gateway.Trigger, gw gateway.Gateway) error {
	if len(t.Arg) < 2 {
		return t.Resp("Expected 2 arguments: [user] [message]")
	}
	return gw.SayPrivate(t.Arg[0], strings.Join(t.Arg[1:], " "))
}

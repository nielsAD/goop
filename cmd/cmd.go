// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"github.com/nielsAD/goop/gateway"
)

// Command interface
type Command interface {
	CanExecute(t *gateway.Trigger) bool
	Execute(t *gateway.Trigger, gw gateway.Gateway) error
}

// Cmd is command base struct that implements Command.CanExecute
type Cmd struct {
	Disabled   bool
	Priviledge gateway.AccessLevel
}

// CanExecute returns true if t.Access >= c.Access
func (c *Cmd) CanExecute(t *gateway.Trigger) bool {
	return !c.Disabled && t.User.Access >= c.Priviledge
}

// Commands listing
type Commands struct {
	Trigger    Trigger
	Say        Say
	SayPrivate SayPrivate
	Time       Time
}

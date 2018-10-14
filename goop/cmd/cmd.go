// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"reflect"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Messages
const (
	MsgNoUserFound   = "No user found with that name"
	MsgMoreUserFound = "Found more than one user with that name"
	MsgNoPermission  = "No permission to perform action"
)

// Cmd is command base struct that implements Command.CanExecute
type Cmd struct {
	Disabled   bool
	Priviledge gateway.AccessLevel
}

// CanExecute returns true if t.Access >= c.Access
func (c *Cmd) CanExecute(t *gateway.Trigger) bool {
	return !c.Disabled && t.User.HasAccess(c.Priviledge)
}

// Commands listing
type Commands struct {
	Trigger    Trigger
	Whoami     Whoami
	Whois      Whois
	Settings   Settings
	Say        Say
	SayPrivate SayPrivate
	Kick       Kick
	Ban        Ban
	Unban      Unban
	Where      Where
	Time       Time
	Uptime     Uptime
	Flip       Flip
	Roll       Roll
}

// AddTo goop
func (c *Commands) AddTo(g *goop.Goop) error {
	var v = reflect.ValueOf(c).Elem()
	for i := 0; i < v.NumField(); i++ {
		var f = v.Type().Field(i)
		if err := g.AddCommand(f.Name, v.Field(i).Addr().Interface().(goop.Command)); err != nil {
			return err
		}
	}
	return nil
}

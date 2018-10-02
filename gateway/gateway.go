// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package gateway

import (
	"context"
	"errors"
	"strings"

	"github.com/nielsAD/gowarcraft3/network"
)

// Errors
var (
	ErrUnknownEvent = errors.New("gw: Unknown event")
	ErrNoChannel    = errors.New("gw: No channel")
)

// Delimiter between main/sub gateway name in ID (i.e. discord:{CHANNELID})
const Delimiter = ":"

// Gateway interface
type Gateway interface {
	network.Emitter
	network.Listener
	ID() string
	SetID(id string)
	Discriminator() string
	Channel() *Channel
	Trigger() string
	Say(s string) error
	SayPrivate(uid string, s string) error
	Run(ctx context.Context) error
	Relay(ev *network.Event, from Gateway) error
}

// User struct
type User struct {
	ID        string
	Name      string
	Access    AccessLevel
	AvatarURL string
}

// Channel struct
type Channel struct {
	ID   string
	Name string
}

// Common gateway struct
type Common struct {
	id string
}

// ID returns the unique gateway identifier
func (c *Common) ID() string {
	return c.id
}

// SetID set the unique gateway identifier
func (c *Common) SetID(id string) {
	c.id = id
}

// Discriminator tag
func (c *Common) Discriminator() string {
	var s = strings.SplitN(c.id, Delimiter, 3)
	if len(s) < 2 {
		return c.id
	}
	return s[1]
}

// TriggerConfig for commands
type TriggerConfig struct {
	Trigger string
	Access  AccessLevel
}

// Config common struct
type Config struct {
	Commands TriggerConfig
}

// Trigger for commands
func (c *Config) Trigger() string {
	return c.Commands.Trigger
}

// FindTrigger checks if s starts with trigger t, returns cmd and args if true
func FindTrigger(t, s string) (bool, string, []string) {
	if len(t) == 0 || !strings.HasPrefix(s, t) {
		return false, "", nil
	}

	s = s[len(t):]
	if len(s) < 1 || s[0] == ' ' {
		return false, "", nil
	}

	f := strings.Fields(s)
	return true, f[0], f[1:]
}

// FindTrigger checks if s starts with trigger, returns cmd and args if true
func (c *Config) FindTrigger(s string) (bool, string, []string) {
	return FindTrigger(c.Commands.Trigger, s)
}

// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package gateway

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/nielsAD/gowarcraft3/network"
)

// Errors
var (
	ErrUnknownEvent   = errors.New("gw: Unknown event")
	ErrUnknownAccess  = errors.New("gw: Unknown access level")
	ErrNoChannel      = errors.New("gw: No channel")
	ErrNoUser         = errors.New("gw: No user")
	ErrNoPermission   = errors.New("gw: No permission")
	ErrNotImplemented = errors.New("gw: Not implemented")
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
	Users() []User
	User(uid string) (*User, error)
	Trigger() string
	Say(s string) error
	SayPrivate(uid string, s string) error
	Kick(uid string) error
	Ban(uid string) error
	Unban(uid string) error
	Responder(gw Gateway, uid string, forcePrivate bool) Responder
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

// HasAccess to o
func (u *User) HasAccess(o AccessLevel) bool {
	return u.Access.HasAccess(o)
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
	Trigger        string
	Access         AccessLevel
	RespondPrivate bool
}

// Config common struct
type Config struct {
	Commands TriggerConfig
}

// Trigger for commands
func (c *Config) Trigger() string {
	return c.Commands.Trigger
}

// Responder for trigger
func (c *Config) Responder(gw Gateway, uid string, forcePrivate bool) Responder {
	if !c.Commands.RespondPrivate && !forcePrivate {
		return gw.Say
	}
	return func(s string) error { return gw.SayPrivate(uid, s) }
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

// FindUser finds user(s) by pattern
func FindUser(gw Gateway, pat string) []*User {
	if u, err := gw.User(pat); err == nil {
		return []*User{u}
	}
	pat = strings.ToLower(pat)

	var res = make([]*User, 0)

	var users = gw.Users()
	for i := range users {
		if m, err := filepath.Match(pat, strings.ToLower(users[i].Name)); err == nil && m {
			res = append(res, &users[i])
		}
	}

	return res
}

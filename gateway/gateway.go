// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package gateway

import (
	"context"
	"errors"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

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
	ChannelUsers() []User
	User(uid string) (*User, error)
	Users() map[string]AccessLevel
	SetUserAccess(uid string, a AccessLevel) (*AccessLevel, error)
	Trigger() string
	Say(s string) error
	SayPrivate(uid string, s string) error
	Kick(uid string) error
	Ban(uid string) error
	Unban(uid string) error
	Ping(uid string) (time.Duration, error)
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

var argPat = regexp.MustCompile(`("(?:\\.|[^\"])*")|('(?:\\.|[^\'])*')|(\S+)(\s*)`)

// ExtractTrigger from s
func ExtractTrigger(s string) *Trigger {
	if r, _ := utf8.DecodeRuneInString(s); len(s) < 1 || unicode.IsSpace(r) {
		return nil
	}

	var m = argPat.FindAllStringSubmatch(s, -1)
	var r = make([]string, len(m))
	var a = make([]string, len(m))
	for i, g := range m {
		r[i] = g[0]
		if g[1] != "" {
			if q, err := strconv.Unquote(g[1]); err == nil {
				a[i] = q
			} else {
				a[i] = g[1][1 : len(g[1])-1]
			}
		} else if g[2] != "" {
			a[i] = g[2][1 : len(g[2])-1]
		} else {
			a[i] = g[3]
		}
	}

	return &Trigger{
		Cmd: a[0],
		Raw: r[1:],
		Arg: a[1:],
	}
}

// FindTrigger checks if s starts with trigger t, return Trigger{} if true
func FindTrigger(t, s string) *Trigger {
	if len(t) == 0 || !strings.HasPrefix(s, t) {
		return nil
	}

	return ExtractTrigger(s[len(t):])
}

// FindTrigger checks if s starts with trigger, return Trigger{} if true
func (c *Config) FindTrigger(s string) *Trigger {
	return FindTrigger(c.Commands.Trigger, s)
}

// FindUserInChannel finds user(s) by pattern
func FindUserInChannel(gw Gateway, pat string) []*User {
	if u, err := gw.User(pat); err == nil {
		if u == nil {
			return nil
		}
		return []*User{u}
	}
	pat = strings.ToLower(pat)

	var res = make([]*User, 0)

	var users = gw.ChannelUsers()
	for i := range users {
		if m, err := filepath.Match(pat, strings.ToLower(users[i].Name)); err == nil && m {
			res = append(res, &users[i])
		}
	}

	return res
}

// FindUser finds user(s) by pattern
func FindUser(gw Gateway, pat string) []*User {
	if u, err := gw.User(pat); err == nil {
		if u == nil {
			return nil
		}
		return []*User{u}
	}
	pat = strings.ToLower(pat)

	var res = make([]*User, 0)

	var users = gw.Users()
	for k := range users {
		u, err := gw.User(k)
		if err != nil || u == nil {
			continue
		}
		if m, err := filepath.Match(pat, strings.ToLower(u.Name)); err == nil && m {
			res = append(res, u)
		}
	}

	var cusers = gw.ChannelUsers()
	for i := range cusers {
		if m, err := filepath.Match(pat, strings.ToLower(cusers[i].Name)); err == nil && m && users[cusers[i].ID] == AccessDefault {
			res = append(res, &cusers[i])
		}
	}

	return res
}

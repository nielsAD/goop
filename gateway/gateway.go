// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package gateway

import (
	"context"
	"errors"

	"github.com/nielsAD/gowarcraft3/network"
)

// Errors
var (
	ErrUnknownEvent = errors.New("gw: Unknown event")
)

// Delimiter between main/sub gateway name (i.e. discord:{CHANNELID})
const Delimiter = ":"

// AccessLevel for user
type AccessLevel int32

// Access constants
const (
	AccessOwner     AccessLevel = 1000
	AccessAdmin     AccessLevel = 300
	AccessOperator  AccessLevel = 200
	AccessWhitelist AccessLevel = 100
	AccessVoice     AccessLevel = 1
	AccessDefault   AccessLevel = 0
	AccessIgnore    AccessLevel = -1
	AccessBlacklist AccessLevel = -100
	AccessKick      AccessLevel = -200
	AccessBan       AccessLevel = -300
)

// Gateway interface
type Gateway interface {
	network.Emitter
	Run(ctx context.Context) error
	Relay(ev *network.Event)
}

// Connected event
type Connected struct{}

// Disconnected event
type Disconnected struct{}

// SystemMessage event
type SystemMessage struct {
	Content string
}

// User base
type User struct {
	ID        string
	Name      string
	Access    AccessLevel
	AvatarURL string
}

// Channel event
type Channel struct {
	ID   string
	Name string
}

// Chat event
type Chat struct {
	User
	Channel
	Content string
}

// PrivateChat event
type PrivateChat struct {
	User
	Content string
}

// Join event
type Join struct {
	User
	Channel
}

// Leave event
type Leave struct {
	User
	Channel
}

// Events types
var Events = []interface{}{
	&Connected{},
	&Disconnected{},
	&SystemMessage{},
	&Channel{},
	&Chat{},
	&PrivateChat{},
	&Join{},
	&Leave{},
}

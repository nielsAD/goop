// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package gateway

import (
	"github.com/nielsAD/gowarcraft3/network"
)

// Connected event
type Connected struct{}

// Disconnected event
type Disconnected struct{}

// SystemMessage event
type SystemMessage struct {
	Content string
}

// Chat event
type Chat struct {
	User
	Content string
}

// PrivateChat event
type PrivateChat struct {
	User
	Content string
}

// Say event
type Say struct {
	Content string
}

// Join event
type Join struct {
	User
}

// Leave event
type Leave struct {
	User
}

// RelayEvents types
var RelayEvents = []interface{}{
	&network.AsyncError{},
	&Connected{},
	&Disconnected{},
	&SystemMessage{},
	&Channel{},
	&Chat{},
	&PrivateChat{},
	&Say{},
	&Join{},
	&Leave{},
}

// Responder callback
type Responder func(s string) error

// Trigger event
type Trigger struct {
	User
	Cmd  string
	Arg  []string
	Resp Responder
}

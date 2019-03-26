// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package goop

import (
	"github.com/nielsAD/goop/gateway"
)

// Start event
type Start struct{}

// Stop event
type Stop struct{}

// NewGateway event
type NewGateway struct {
	gateway.Gateway
}

// NewCommand event
type NewCommand struct {
	Name    string
	Command Command
}

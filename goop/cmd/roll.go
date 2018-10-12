// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Roll a dice
type Roll struct{ Cmd }

// Execute command
func (c *Roll) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	var max = int64(100)
	if len(t.Arg) > 0 {
		v, err := strconv.ParseInt(t.Arg[0], 0, 32)
		if err != nil {
			return t.Resp(err.Error())
		}
		max = v
	}
	if max <= 0 {
		return t.Resp("Pick a positive number larger than 0")
	}
	return t.Resp(fmt.Sprintf("%d", rand.Int63n(max)))
}

// Flip a coin
type Flip struct{ Cmd }

// Execute command
func (c *Flip) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	var coin = "Heads"
	if rand.Float64() < 0.5 {
		coin = "Tails"
	}
	return t.Resp(coin)
}

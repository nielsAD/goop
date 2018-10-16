// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Where prints connected gateways
type Where struct{ Cmd }

// Execute command
func (c *Where) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	var channels = make([]string, 0)
	for _, gw := range g.Gateways {
		var c = gw.Channel()
		if c == nil {
			continue
		}
		channels = append(channels, fmt.Sprintf("`%s@%s`", c.Name, gw.Discriminator()))
	}
	sort.Strings(channels)
	return t.Resp(fmt.Sprintf("Present in channels: [%s]", strings.Join(channels, ", ")))
}

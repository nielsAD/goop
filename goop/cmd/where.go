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
	var channels = []string{}
	for _, gw := range g.Gateways {
		var c = gw.Channel()
		if c == nil {
			continue
		}
		channels = append(channels, fmt.Sprintf("%s@%s", c.Name, gw.Discriminator()))
	}
	sort.Strings(channels)
	return t.Resp(fmt.Sprintf("Present in channels: [%s]", strings.Join(channels, ", ")))
}

// Who prints who is in channel
type Who struct{ Cmd }

var plural = map[bool]string{
	false: "",
	true:  "s",
}

// Execute command
func (c *Who) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	var total = 0
	var online = []string{}
	for _, gw := range g.Gateways {
		var users = gw.ChannelUsers()
		if users == nil {
			continue
		}

		total += len(users)

		if len(users) >= 10 {
			online = append(online, fmt.Sprintf("<%d in %s>", len(users), gw.Discriminator()))
		} else {
			for _, u := range users {
				if u.Access >= gateway.AccessOperator {
					online = append(online, fmt.Sprintf("[%s@%s]", u.Name, gw.Discriminator()))
				} else {
					online = append(online, fmt.Sprintf("%s@%s", u.Name, gw.Discriminator()))
				}
			}
		}
	}
	sort.Strings(online)
	return t.Resp(fmt.Sprintf("%d user%s online: %s", total, plural[total != 1], strings.Join(online, " ")))
}

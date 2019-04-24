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

type whoList struct {
	name   []string
	access []gateway.AccessLevel
}

func (l *whoList) Append(name string, access gateway.AccessLevel) {
	l.name = append(l.name, name)
	l.access = append(l.access, access)
}

func (l *whoList) String() string {
	sort.Sort(l)
	return strings.Join(l.name, " ")
}

func (l *whoList) Len() int {
	return len(l.name)
}

func (l *whoList) Swap(i, j int) {
	l.name[i], l.name[j] = l.name[j], l.name[i]
	l.access[i], l.access[j] = l.access[j], l.access[i]
}

func (l *whoList) Less(i, j int) bool {
	if l.access[i] == l.access[j] {
		return strings.ToLower(l.name[i]) < strings.ToLower(l.name[j])
	}
	return l.access[i] > l.access[j]
}

// Execute command
func (c *Who) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	var total = 0
	var list whoList
	for _, gw := range g.Gateways {
		var users = gw.ChannelUsers()
		if users == nil {
			continue
		}

		total += len(users)

		if len(users) >= 10 {
			list.Append(fmt.Sprintf("<%d in %s>", len(users), gw.Discriminator()), gateway.AccessMax)
		} else {
			for _, u := range users {
				if u.Access >= gateway.AccessOperator {
					list.Append(fmt.Sprintf("[%s@%s]", u.Name, gw.Discriminator()), u.Access)
				} else if u.Access > gateway.AccessIgnore {
					list.Append(fmt.Sprintf("%s@%s", u.Name, gw.Discriminator()), u.Access)
				}
			}
		}
	}

	return t.Resp(fmt.Sprintf("%d user%s online: %s", total, plural[total != 1], list.String()))
}

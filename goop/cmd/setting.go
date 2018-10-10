// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Settings management
type Settings struct{ Cmd }

func fixsep(in string) string {
	if os.PathSeparator == '/' {
		return in
	}
	return strings.Replace(in, "/", string(os.PathSeparator), -1)
}

// Execute command
func (c *Settings) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	// Always respond in private
	var resp = gw.Responder(gw, t.User.ID, true)

	if len(t.Arg) < 2 {
		return resp("Expected 2 arguments: get|set|unset [setting]")
	}

	var m = g.Config.FlatMap()
	var q = strings.ToLower(fixsep(t.Arg[1]))
	var s = make([]string, 0)
	for k := range m {
		if m, err := filepath.Match(q, strings.ToLower(k)); err != nil || !m {
			continue
		}
		s = append(s, k)
	}
	sort.Strings(s)

	switch strings.ToLower(t.Arg[0]) {
	case "get":
		if len(s) == 0 {
			return resp("No matching settings found")
		}

		var lines = make([]string, 0)
		for _, k := range s {
			lines = append(lines, fmt.Sprintf("%s = %v", k, m[k]))
		}

		return resp(strings.Join(lines, "\n"))
	case "set":
		if len(t.Arg) < 3 {
			return resp("Expected 2 arguments: set [setting] [value]")
		}
		if len(s) == 0 {
			s = append(s, t.Arg[1])
		}
		var val = strings.Join(t.Arg[2:], " ")

		var lines = make([]string, 0)
		for _, k := range s {
			if err := g.Config.SetString(k, val); err == nil {
				if fmt.Sprintf("%v", m[k]) == val {
					continue
				}
				lines = append(lines, fmt.Sprintf("Changed %s from %v to %s", k, m[k], val))
			}
		}

		if len(lines) == 0 {
			return resp("No settings changed")
		}
		return resp(strings.Join(lines, "\n"))
	case "unset":
		var lines = make([]string, 0)
		for _, k := range s {
			if err := g.Config.Unset(k); err == nil {
				lines = append(lines, fmt.Sprintf("Unset %s = %v", k, m[k]))
			}
		}

		if len(lines) == 0 {
			return resp("No matching settings found")
		}
		return resp(strings.Join(lines, "\n"))
	default:
		return resp("Expected action to be one of get|set|unset")
	}
}

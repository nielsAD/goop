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

func findKeys(m map[string]interface{}, pat ...string) []string {
	for i := range pat {
		pat[i] = strings.ToLower(pat[i])
	}
	var s = make([]string, 0)

outer:
	for k := range m {
		for _, p := range pat {
			if !strings.Contains(k, p) {
				continue outer
			}
		}
		s = append(s, k)
	}

	sort.Strings(s)
	return s
}

func matchKeys(m map[string]interface{}, pat string) []string {
	var q = strings.ToLower(fixsep(pat))
	var s = make([]string, 0)
	for k := range m {
		if m, err := filepath.Match(q, k); err != nil || !m {
			continue
		}
		s = append(s, k)
	}
	sort.Strings(s)
	return s
}

// Execute command
func (c *Settings) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	// Always respond in private
	var resp = gw.Responder(gw, t.User.ID, true)

	if len(t.Arg) < 2 {
		return resp("Expected 2 arguments: find|get|set|unset [setting]")
	}

	var m = g.Config.FlatMap()
	var l = make([]string, 0)

	switch strings.ToLower(t.Arg[0]) {
	case "find", "f":
		var k = findKeys(m, t.Arg[1:]...)
		for _, v := range k {
			l = append(l, fmt.Sprintf("%s = %v", v, m[v]))
		}
	case "get", "g":
		var k = matchKeys(m, t.Arg[1])
		for _, v := range k {
			l = append(l, fmt.Sprintf("%s = %v", v, m[v]))
		}
	case "unset", "u", "us":
		var k = matchKeys(m, t.Arg[1])
		for _, v := range k {
			err := g.Config.Unset(v)
			if err != nil {
				if len(k) == 1 {
					return resp(err.Error())
				}
				continue
			}
			l = append(l, fmt.Sprintf("Unset %s = %v", v, m[v]))
		}
	case "set", "s":
		if len(t.Arg) < 3 {
			return resp("Expected 2 arguments: set [setting] [value]")
		}
		var k = matchKeys(m, t.Arg[1])
		var s = strings.Join(t.Arg[2:], " ")
		if len(k) == 0 {
			k = []string{t.Arg[1]}
		}
		for _, v := range k {
			err := g.Config.SetString(v, s)
			if err != nil {
				if len(k) == 1 {
					return resp(err.Error())
				}
				continue
			}
			if fmt.Sprintf("%v", m[v]) == s {
				continue
			}
			l = append(l, fmt.Sprintf("Changed %s from %v to %s", v, m[v], s))
		}
	default:
		return resp("Expected action to be one of find|get|set|unset")
	}

	if len(l) == 0 {
		return resp("No matching settings found")
	}
	return resp(strings.Join(l, "\n"))
}

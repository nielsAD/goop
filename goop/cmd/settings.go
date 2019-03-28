// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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

var ignorePat = regexp.MustCompile("^.*/_.*$")

func ignore(key string, val interface{}) bool {
	if ignorePat.MatchString(key) {
		return true
	}

	var v = reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return v.Len() != 0
	default:
		return false
	}
}

func findKeys(m map[string]interface{}, pat ...string) []string {
	for i := range pat {
		pat[i] = strings.ToLower(pat[i])
	}
	var s = []string{}

outer:
	for k := range m {
		if ignore(k, m[k]) {
			continue
		}
		var l = strings.ToLower(k)
		for _, p := range pat {
			if !strings.Contains(l, p) {
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
	var s = []string{}
	for k := range m {
		if ignore(k, m[k]) {
			continue
		}
		var l = strings.ToLower(k)
		if m, err := filepath.Match(q, l); err != nil || !m {
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
	var l = []string{}
	var u = false

	switch strings.ToLower(t.Arg[0]) {
	case "find", "f":
		var k = findKeys(m, t.Arg[1:]...)
		for _, v := range k {
			l = append(l, fmt.Sprintf("%s = %v", strings.ToLower(v), m[v]))
		}
	case "get", "g":
		var k = matchKeys(m, t.Arg[1])
		for _, v := range k {
			l = append(l, fmt.Sprintf("%s = %v", strings.ToLower(v), m[v]))
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
			l = append(l, fmt.Sprintf("Unset %s = %v", strings.ToLower(v), m[v]))
		}
		u = true
	case "set", "s":
		if len(t.Arg) < 3 {
			return resp("Expected 3 arguments: set [setting] [value]")
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
			l = append(l, fmt.Sprintf("Changed %s from %v to %s", strings.ToLower(v), m[v], s))
		}
		u = true
	default:
		return resp("Expected action to be one of find|get|set|unset")
	}

	if len(l) == 0 {
		return resp("No matching settings found")
	}

	if u {
		g.Fire(&gateway.ConfigUpdate{})
	}

	return resp(strings.Join(l, "\n"))
}

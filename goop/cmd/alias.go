// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/goop"
)

// Alias abbreviates a command
type Alias struct {
	Cmd
	Exe            string
	Arg            []string
	ArgExpected    int
	WithPriviledge gateway.AccessLevel
}

// Replacer callback
type Replacer func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string

// Placeholders map
var Placeholders = map[string]Replacer{
	"%CMD%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return t.Cmd
	},
	"%RAW%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return strings.Join(t.Raw, "")
	},
	"%ARGS%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return strings.Join(t.Arg, " ")
	},
	"%NARGS%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return fmt.Sprintf("%d", len(t.Arg))
	},
	"%UID%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return t.User.ID
	},
	"%USTR%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return t.User.Name
	},
	"%ULVL%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return t.User.Access.String()
	},
	"%GWID%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return gw.ID()
	},
	"%GWDIS%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return gw.Discriminator()
	},
	"%GWDEL%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		return gateway.Delimiter
	},
	"%RARG000%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		idx, err := strconv.Atoi(replacersInt.FindString(m))
		if err != nil || idx <= 0 || len(t.Raw) <= idx {
			return m
		}
		return t.Raw[idx-1]
	},
	"%..RARG000%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		idx, err := strconv.Atoi(replacersInt.FindString(m))
		if err != nil || idx <= 0 || len(t.Raw) <= idx {
			return m
		}
		return strings.Join(t.Raw[:idx-1], "")
	},
	"%RARG000..%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		idx, err := strconv.Atoi(replacersInt.FindString(m))
		if err != nil || idx <= 0 || len(t.Raw) <= idx {
			return m
		}
		return strings.Join(t.Raw[idx-1:], "")
	},
	"%ARG000%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		idx, err := strconv.Atoi(replacersInt.FindString(m))
		if err != nil || idx <= 0 || len(t.Arg) <= idx {
			return m
		}
		return t.Arg[idx-1]
	},
	"%..ARG000%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		idx, err := strconv.Atoi(replacersInt.FindString(m))
		if err != nil || idx <= 0 || len(t.Arg) <= idx {
			return m
		}
		return strings.Join(t.Arg[:idx-1], " ")
	},
	"%ARG000..%": func(m string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
		idx, err := strconv.Atoi(replacersInt.FindString(m))
		if err != nil || idx <= 0 || len(t.Arg) <= idx {
			return m
		}
		return strings.Join(t.Arg[idx-1:], " ")
	},
}

var replacersInt = regexp.MustCompile("\\d+")
var replacersPat = regexp.MustCompile((func() string {
	var s = []string{}
	for k := range Placeholders {
		s = append(s, strings.ReplaceAll(regexp.QuoteMeta(k), "000", "\\d+"))
	}
	return strings.Join(s, "|")
}()))

// Replace all placeholders
func Replace(s string, t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) string {
	return replacersPat.ReplaceAllStringFunc(s, func(s string) string {
		var idx = replacersInt.ReplaceAllString(s, "000")
		return Placeholders[idx](s, t, gw, g)
	})
}

// Execute command
func (c *Alias) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	var trig = *t
	trig.Cmd = c.Exe

	if len(t.Arg) < c.ArgExpected {
		return t.Resp(fmt.Sprintf("Expected %d arguments", c.ArgExpected))
	}

	if c.Arg != nil {
		trig.Raw = make([]string, len(c.Arg))
		trig.Arg = make([]string, len(c.Arg))
		for i, a := range c.Arg {
			var s = Replace(a, t, gw, g)
			trig.Raw[i] = s + " "
			trig.Arg[i] = s
		}
	}

	if c.WithPriviledge != gateway.AccessDefault {
		trig.User.Access = c.WithPriviledge
	}

	gw.Fire(&trig)
	return nil
}

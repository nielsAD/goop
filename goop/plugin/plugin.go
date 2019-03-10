// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package plugin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nielsAD/goop/gateway"

	luar "github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"
)

// Config stores the configuration of a single plugin
type Config struct {
	Path    string
	Options map[string]interface{}
}

// Globals stores shared variables
type Globals map[string]interface{}

// Plugin loads and executes a lua script
type Plugin struct {
	l *lua.LState

	// Set once before Run(), read-only after that
	*Config
}

func luaTypeOf(i interface{}) string {
	return reflect.TypeOf(i).String()
}

func luaInspect(i interface{}) string {
	return fmt.Sprintf("%+v", i)
}

// Load a lua plugin
func Load(conf *Config, g Globals) (*Plugin, error) {
	if conf.Options == nil {
		conf.Options = make(map[string]interface{})
	}

	var p = Plugin{
		Config: conf,
		l: lua.NewState(lua.Options{
			IncludeGoStackTrace: true,
		}),
	}

	// Import functions
	p.l.SetGlobal("gotypeof", luar.New(p.l, luaTypeOf))
	p.l.SetGlobal("inspect", luar.New(p.l, luaInspect))

	// Import script variables
	p.l.SetGlobal("globals", luar.New(p.l, g))
	p.l.SetGlobal("options", luar.New(p.l, conf.Options))

	// Import event constants
	var events = p.l.NewTable()
	for _, e := range gateway.RelayEvents {
		p.l.SetTable(events, lua.LString(luaTypeOf(e)[1:]), luar.New(p.l, e))
	}
	p.l.SetGlobal("events", events)

	// Import access level constants
	var access = p.l.NewTable()
	for i, str := range gateway.ConStrings {
		if len(str) == 0 {
			str = "Default"
		}
		str = strings.Title(str)
		p.l.SetTable(access, lua.LString(str), lua.LNumber(gateway.ConLevels[i]))
		p.l.SetTable(access, lua.LNumber(gateway.ConLevels[i]), lua.LString(str))
	}
	p.l.SetGlobal("access", access)

	if err := p.l.DoFile(p.Path); err != nil {
		return nil, err
	}

	return &p, nil
}

// Close the lua context
func (p *Plugin) Close() {
	p.l.Close()
}

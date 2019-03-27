// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package plugin

import (
	luar "github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"
)

// Config stores the configuration of a single plugin, maps to lua.LTable
type Config map[interface{}]interface{}

// Globals stores shared variables
type Globals map[string]interface{}

// Plugin loads and executes a lua script
type Plugin struct {
	*lua.LState

	timers Timers

	// Set once before Run(), read-only after that
	Config
}

// NewState prepares a new Lua environment
func NewState() *lua.LState {
	var ls = lua.NewState(lua.Options{
		SkipOpenLibs:        true,
		IncludeGoStackTrace: true,
	})
	importModules(ls)
	importGlobal(ls)
	importPreload(ls)
	return ls
}

// Load a lua plugin
func Load(path string, conf Config, g Globals) (*Plugin, error) {
	var p = Plugin{
		Config: conf,
		LState: NewState(),
	}

	p.timers.ImportTo(p.LState)

	p.SetGlobal("globals", g)
	p.SetGlobal("options", conf)

	for k, v := range g {
		p.SetGlobal(k, v)
	}

	if err := p.DoFile(path); err != nil {
		return nil, err
	}

	return &p, nil
}

// SetGlobal variable
func (p *Plugin) SetGlobal(name string, val interface{}) {
	p.LState.SetGlobal(name, luar.New(p.LState, val))
}

// Close plugin
func (p *Plugin) Close() {
	p.timers.Stop()
	p.LState.Close()
}

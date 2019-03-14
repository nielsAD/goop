// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package plugin

import (
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
	*lua.LState

	// Set once before Run(), read-only after that
	*Config
}

// Load a lua plugin
func Load(conf *Config, g Globals) (*Plugin, error) {
	if conf.Options == nil {
		conf.Options = make(map[string]interface{})
	}

	var p = Plugin{
		Config: conf,
		LState: lua.NewState(lua.Options{
			SkipOpenLibs:        true,
			IncludeGoStackTrace: true,
		}),
	}

	p.SetGlobal("globals", g)
	p.SetGlobal("options", conf.Options)

	for k, v := range g {
		p.SetGlobal(k, v)
	}

	importModules(p.LState)
	importGlobal(p.LState)
	importPreload(p.LState)

	if err := p.DoFile(p.Path); err != nil {
		return nil, err
	}

	return &p, nil
}

// SetGlobal variable
func (p *Plugin) SetGlobal(name string, val interface{}) {
	p.LState.SetGlobal(name, luar.New(p.LState, val))
}

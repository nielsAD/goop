// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package plugin

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

// Config stores the configuration of a single plugin
type Config struct {
	Path          string
	CallStackSize int
	RegistrySize  int
}

// Globals stores shared variables
type Globals map[string]interface{}

// Plugin loads and executes a lua script
type Plugin struct {
	*lua.LState

	main   *lua.LFunction
	timers Timers

	// Set once before Run(), read-only after that
	*Config
}

// NewState prepares a new Lua environment
func NewState(callStackSize int, registrySize int) *lua.LState {
	var ls = lua.NewState(lua.Options{
		CallStackSize:       callStackSize,
		RegistrySize:        registrySize,
		SkipOpenLibs:        true,
		IncludeGoStackTrace: true,
	})
	importModules(ls)
	importGlobal(ls)
	importPreload(ls)
	return ls
}

// Load a lua plugin
func Load(conf *Config) (*Plugin, error) {
	var p = Plugin{
		Config: conf,
		LState: NewState(conf.CallStackSize, conf.RegistrySize),
	}

	fun, err := p.LoadFile(p.Path)
	if err != nil {
		return nil, err
	}

	p.timers.ImportTo(p.LState)
	p.main = fun

	return &p, nil
}

// SetGlobal variable
func (p *Plugin) SetGlobal(name string, val interface{}) {
	p.LState.SetGlobal(name, luar.New(p.LState, val))
}

// Run plugin
func (p *Plugin) Run() error {
	p.Push(p.main)
	return p.PCall(0, lua.MultRet, nil)
}

// Close plugin
func (p *Plugin) Close() {
	p.timers.Stop()
	p.LState.Close()
}

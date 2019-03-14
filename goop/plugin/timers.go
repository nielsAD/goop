// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package plugin

import (
	"sync"
	"time"

	luar "github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"
)

// Timers registers all active timers
type Timers struct {
	stop   bool
	mutex  sync.Mutex
	timers map[*time.Timer]struct{}
}

// ImportTo Lua state
func (t *Timers) ImportTo(ls *lua.LState) {
	ls.SetGlobal("setTimeout", luar.New(ls, t.AfterFunc))
}

func (t *Timers) add(m *time.Timer) {
	t.mutex.Lock()
	if !t.stop {
		if t.timers == nil {
			t.timers = make(map[*time.Timer]struct{})
		}
		t.timers[m] = struct{}{}
	}
	t.mutex.Unlock()
}

func (t *Timers) del(m *time.Timer) {
	t.mutex.Lock()
	delete(t.timers, m)
	t.mutex.Unlock()
}

// Stop all timers
func (t *Timers) Stop() {
	t.mutex.Lock()
	t.stop = true
	for m := range t.timers {
		m.Stop()
	}
	t.mutex.Unlock()
}

// AfterFunc waits for the duration to elapse and then calls f in its own goroutine
// see time.AfterFunc
func (t *Timers) AfterFunc(ms int, f func()) func() bool {
	var m *time.Timer
	m = time.AfterFunc((time.Duration)(ms)*time.Millisecond, func() {
		f()
		t.del(m)
	})

	t.add(m)
	return func() bool {
		t.del(m)
		return m.Stop()
	}
}

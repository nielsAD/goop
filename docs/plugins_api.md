Writing Plugins
===============

Plugins are written in a subset of [Lua 5.1](https://www.lua.org/manual/5.1/manual.html), a light-weight scripting language. [Gopher-lua](https://github.com/yuin/gopher-lua) is used as compiler, but the Lua `os` and `io` default libraries are not imported. Instead, a large part of the Go standard library is available for import.

The plugin API is almost a direct mapping of the Go API, this means that source code [documentation](https://godoc.org/github.com/nielsAD/goop) is a very good starting point for script documentation. Typically, a plugin will subscribe to one or more events and execute a callback function whenever such an event occurs.

?> **TIP:** Enable the [debug](plugins.md#debug) plugin to find out what events are fired.


Global Imports
--------------

|      Global     | Description |
|-----------------|-------------|
|goop             | Global [Goop](https://godoc.org/github.com/nielsAD/goop/goop#Goop) instance. |
|log              | Global [Logger](https://golang.org/pkg/log/) instance. |
|access           | Table with all [access levels](access.md) (e.g. `access.Voice` or `access.Admin`). |
|events           | Table with all events (e.g. `events.Chat`, `events.Join`). Use with `goop:On()`. |
|gotypeof(x)      | Returns string with go type of `x`. |
|inspect(x)       | Returns string representation of value of `x`. |
|topic(s)         | Create event topic with string literal `s`. Use with `goop:On()` and `goop:Fire()`. |
|command(f)       | Create command with callback `f`. Use with `goop:AddCommand()`. |
|command_alias(t) | Create command alias from table `t`. Use with `goop:AddCommand()`. |
|interface()      | Create `interface{}` instance. |
|setTimeout(ms, f)| Call `f` after waiting for `ms` milliseconds. |

_Example:_
```lua
-- Listen for Chat events and log event information
goop:On(events.Chat, function(ev)
    log:Println(inspect(ev))
end)
```


Module Imports
--------------

| Lua Module   | Go Package |
|--------------|------------|
| go.errors    | [errors](https://golang.org/pkg/errors) |
| go.reflect   | [reflect](https://golang.org/pkg/reflect) |
| go.io        | [io](https://golang.org/pkg/io) |
| go.context   | [context](https://golang.org/pkg/context) |
| go.sync      | [sync](https://golang.org/pkg/sync) |
| go.time      | [time](https://golang.org/pkg/time) |
| go.bytes     | [bytes](https://golang.org/pkg/bytes) |
| go.strings   | [strings](https://golang.org/pkg/strings) |
| go.strconv   | [strconv](https://golang.org/pkg/strconv) |
| go.fmt       | [fmt](https://golang.org/pkg/fmt) |
| go.regexp    | [regexp](https://golang.org/pkg/regexp) |
| go.sort      | [sort](https://golang.org/pkg/sort) |
| go.net       | [net](https://golang.org/pkg/net) |
| go.url       | [net/url](https://golang.org/pkg/net/url) |
| go.http      | [net/http](https://golang.org/pkg/net/http) |
| go.json      | [encoding/json](https://golang.org/pkg/encoding/json) |
| go.color     | [fatih/color](https://godoc.org/github.com/fatih/color) |
| go.websocket | [gorilla/websocket](https://github.com/gorilla/websocket) |

_Example:_
```lua
local ioutil  = require("go.io")
local strings = require("go.strings")
local http    = require("go.http")
local url     = require("go.url")

-- Add .weather command that prints current weather on execution
goop:AddCommand("weather", command(function(trig)
    local resp, err_get = http.Get("https://wttr.in/?format=3")
    if err_get ~= nil then
        return err_get
    end

    local body, err_read = ioutil.ReadAll(resp.Body)
    resp.Body:Close()

    if err_read ~= nil then
        return err_read
    end

    if resp.StatusCode == http.StatusOK then
        return trig.Resp(strings.TrimSpace(body))
    else
        return trig.Resp("Could not get weather for location (" .. resp.Status .. ")")
    end
end))
```


Options
-------

The global `options` table can be used to read and write persistent configuration settings. Use `defoptions` to define a set of default options to fall back on when no user configuration is available for a certain setting.

Users can modify these settings through the regular interfaces (configuration [file](plugins.md) or the [.settings](commands_builtin.md#settings) command).  
Scripts can use `options` table for persistent storage, fieldnames starting with underscore ("_") are treated as private.


_Example:_
```lua
-- Define default configuration
defoptions({
    AccessTrigger = access.Voice, -- Access level required to trigger command
})

-- Flip command, randomly flips a coin and prints the result
goop:AddCommand("flip", command(function(trig, _)

    -- Access current configuration through options table
    if trig.User.Access < options.AccessTrigger then
        return nil
    end

    local coin = "Heads"
    if math.random() < 0.5 then
        coin = "Tails"
    end

    return trig.Resp(coin)
end))
```
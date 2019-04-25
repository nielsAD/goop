-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Check for available updates

defoptions({
    Interval = "24h", -- Check interval
})

local errors = require("go.errors")
local ioutil = require("go.io")
local json   = require("go.json")
local time   = require("go.time")
local http   = require("go.http")

local API_URL = "https://api.github.com/repos/nielsAD/goop/releases/latest"
local WEB_URL = "https://github.com/nielsAD/goop/releases/latest"

local function check()
    local now      = time.Now()
    local interval = time.ParseDuration(options.Interval)
    if interval <= 0 or now:Sub(time.Unix(options._last_check or 0, 0)) < interval then
        return false
    end

    -- persist in config
    options._last_check = now:Unix()

    local resp, get_err = http.Get(API_URL)
    if get_err ~= nil then
        goop:Fire(events.async_error(get_err))
        return false
    end

    local body, read_err = ioutil.ReadAll(resp.Body)
    resp.Body:Close()

    if read_err ~= nil or resp.StatusCode ~= http.StatusOK then
        goop:Fire(events.async_error(read_err or errors.New(resp.Status)))
        return false
    end

    local json_obj = interface()
    local json_err = json.Unmarshal(body, json_obj)
    if json_err ~= nil then
        goop:Fire(events.async_error(json_err))
        return false
    end

    -- dereference *interface{} to interface{}
    json_obj = -json_obj

    if json_obj.tag_name == globals.BUILD_VERSION then
        return false
    end

    return json_obj.tag_name
end

local available = false

goop:On(events.Join, function(ev)
    local user = ev.Arg
    local gw   = ev.Opt[1]

    if user.Access < access.Owner then
        return
    end

    available = available or check()
    if not available then
        return
    end

    gw:SayPrivate(user.ID, "Goop " .. available .. " available, download update at " .. WEB_URL)
end)

goop:On(events.Start, function()
    available = check()
    if not available then
        return
    end

    log:Printf("[UPDATE] Goop %s available, download update at %s\n", available, WEB_URL)
end)
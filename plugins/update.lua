-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Greet users when they join the channel
--
-- Options:
--   Interval:  Check interval

local errors = require("go.errors")
local ioutil = require("go.io")
local json   = require("go.json")
local time   = require("go.time")
local http   = require("go.http")

local API_URL = "https://api.github.com/repos/nielsAD/goop/releases/latest"
local WEB_URL = "https://github.com/nielsAD/goop/releases/latest"

local last_check = {}

local function check()
    local now      = time.Now()
    local interval = time.ParseDuration(options["Interval"] or "24h")
    if interval <= 0 or now:Sub(last_check) < interval then
        return false
    end
    last_check = now

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
        goop:Fire(events.async_error(read_err or errors.New(resp.Status)))
        return false
    end

    -- dereference *interface{} to interface{}
    json_obj = -json_obj

    if json_obj.tag_name == globals["version"] then
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

    gw:SayPrivate(user.ID, "Goop " .. available .. " available, be sure to update! " .. WEB_URL)
end)

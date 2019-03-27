-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Add weather command that prints weather on location

options._default = {
    AccessTrigger   = access.Whitelist, -- Min access level
    DefaultLocation = "",               -- Default location
}

local ioutil  = require("go.io")
local strings = require("go.strings")
local http    = require("go.http")
local url     = require("go.url")

goop:AddCommand("weather", command(function(trig)
    if trig.User.Access < options.AccessTrigger then
        return nil
    end

    local loc = options.DefaultLocation
    if #trig.Arg > 0 then
        loc = strings.Join(trig.Arg, " ")
    end

    local resp, err_get = http.Get("https://wttr.in/" .. url.PathEscape(loc) .. "?format=3")
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

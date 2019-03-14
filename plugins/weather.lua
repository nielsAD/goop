-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Add weather command that prints weather on location
--
-- Options:
--   AccessTrigger:    Access level required to trigger command
--   DefaultLocation:  Default location

local ioutil  = require("go.io")
local strings = require("go.strings")
local http    = require("go.http")
local url     = require("go.url")

goop:AddCommand("weather", command(function(trig, _)
    local lvl = options["AccessTrigger"] or access.Whitelist
    if trig.User.Access < lvl then
        return nil
    end

    local loc = options["DefaultLocation"] or ""
    if #trig.Arg > 0 then
        loc = strings.Join(trig.Arg, " ")
    end

    local resp, err = http.Get("https://wttr.in/" .. url.PathEscape(loc) .. "?format=3")
    if err ~= nil then
        return err
    end

    local res
    if resp.StatusCode == http.StatusOK then
        local body, r = ioutil.ReadAll(resp.Body)
        if r == nil then
            res = trig.Resp(strings.TrimSpace(body))
        else
            res = r
        end
    end

    resp.Body:Close()
    return res
end))

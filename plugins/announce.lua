-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Add announce command that can repeat a message at specified interval

options._default = {
    AccessTrigger = access.Operator, -- Access level required to trigger command
}

local time    = require("go.time")
local strings = require("go.strings")

local cancel

goop:AddCommand("announce", command(function(trig, gw)
    if trig.User.Access < options.AccessTrigger then
        return nil
    end

    if cancel then
        cancel()
    end

    if #trig.Arg == 0 then
        return nil
    elseif #trig.Arg < 2 then
        return trig.Resp("Expected 2 arguments: [duration] [message]")
    end

    local dur, dur_err = time.ParseDuration(trig.Arg[1])
    if dur_err ~= nil or dur <= 0 then
        return trig.Resp((dur_err and inspect(dur_err)) or "Invalid duration")
    end

    local msg = strings.Join(strings.SSlice(trig.Arg, 1, #trig.Arg), " ")

    local function announce()
        local chan = gw:ChannelUsers()
        if chan == nil or #chan > 0 then
            local say_err = gw:Say(msg)
            if say_err ~= nil then
                goop:Fire(events.async_error(say_err))
            end
        end

        cancel = setTimeout(dur / time.Millisecond, announce)
    end

    announce()
    return nil
end))

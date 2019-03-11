-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Prevent relay of commands

local function starts_with(str, start)
    return str:sub(1, #start) == start
end

local function filter(ev)
    local msg = ev.Arg
    local gw  = ev.Opt[1]

    if starts_with(msg.Content, gw:Trigger()) then
        ev:PreventNext()
    end
end

goop:On(events.Chat, filter)
goop:On(events.PrivateChat, filter)

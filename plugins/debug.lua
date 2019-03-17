-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Print all events

local color = require("go.color")

print("ACCESS", inspect(access))
print("EVENTS", inspect(events))

local blacklist = {
    "*capi.Packet", "*capi.Response",
    "*discordgo.Event", "*discordgo.PresenceUpdate", "*discordgo.TypingStart"
}

goop:On(nil, function(ev)
    local arg = ev.Arg
    local typ = gotypeof(arg)

    local gw = ""
    if ev.Opt and #ev.Opt > 0 then
        gw = ev.Opt[1]:ID()
    end

    for _, bl in ipairs(blacklist) do
        if typ:find(bl) then
            return
        end
    end

    log:Println(color.Blue("EVENT %s %s %+v", gw, typ, arg))
end)

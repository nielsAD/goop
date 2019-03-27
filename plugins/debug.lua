-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Print all events

options._default ={
    Blacklist = {
        ["*capi.Packet"]              = true,
        ["*capi.Response"]            = true,
        ["*discordgo.Event"]          = true,
        ["*discordgo.PresenceUpdate"] = true,
        ["*discordgo.TypingStart"]    = true,
    },
}

local color = require("go.color")

goop:On(nil, function(ev)
    local arg = ev.Arg
    local typ = gotypeof(arg)

    local gw = ""
    if ev.Opt and #ev.Opt > 0 then
        gw = ev.Opt[1]:ID()
    end

    if options.Blacklist and options.Blacklist[typ] then
        return
    end

    log:Println(color.Blue("EVENT %s %s %+v", gw, typ, arg))
end)

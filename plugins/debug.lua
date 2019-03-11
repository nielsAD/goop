-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Print all events

print("ACCESS", inspect(access))
print("EVENTS", inspect(events))

blacklist = { "*discordgo.Event", "*discordgo.PresenceUpdate", "*capi.Packet" }

goop:On(nil, function(ev)
    local arg = ev.Arg
    local typ = gotypeof(arg)
    local gw  = ev.Opt[1]

    for _, bl in ipairs(blacklist) do
        if typ:find(bl) then
            return
        end
    end

    print("EVENT", gw:ID(), typ, inspect(arg))
end)

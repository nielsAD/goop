-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Echo everything said in channel or private message
-- 
-- Options:
--   Access:  Min access level

goop:On(events.Chat, function(ev)
    local msg = ev.Arg
    local gw  = ev.Opt[1]

    local lvl = options["Access"] or access.Voice
    if msg.User.Access < lvl then
        return
    end

    gw:Say(msg.Content)
end)

goop:On(events.PrivateChat, function(ev)
    local msg = ev.Arg
    local gw  = ev.Opt[1]

    local lvl = options["Access"] or access.Voice
    if msg.User.Access < lvl then
        return
    end

    gw:SayPrivate(msg.User.ID, msg.Content)
end)
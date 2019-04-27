-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Echo everything said in channel or private message

defoptions({
    Access = access.Voice, -- Min access level
})

goop:On(events.Chat, function(ev)
    local msg = ev.Arg
    if msg.User.Access < options.Access then
        return
    end

    local gw = ev.Opt[1]
    gw:Say(msg.Content)
end)

goop:On(events.PrivateChat, function(ev)
    local msg = ev.Arg
    if msg.User.Access < options.Access then
        return
    end

    local gw  = ev.Opt[1]
    gw:SayPrivate(msg.User.ID, msg.Content)
end)
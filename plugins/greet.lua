-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Greet users when they join the channel

options._default = {
    AccessMin   = access.Voice,               -- Min access level
    AccessMax   = access.Owner,               -- Max access level
    Gateways    = {"^bnet:.*$", "^capi:.*$"}, -- Pattern for gateway IDs
    Public      = false,                      -- Send public instead of private message

    -- Greeting template. The following placeholders are available
    --   * #i:     User ID
    --   * #n:     User name
    --   * #a:     Access level
    --   * #g:     Gateway ID
    --   * #d:     Gateway discriminator
    Message = "Welcome to ##c, #n! Your access level is <#a>.",
}

goop:On(events.Join, function(ev)
    local user = ev.Arg
    local gw   = ev.Opt[1]

    if user.Access < options.AccessMin or user.Access > options.AccessMax then
        return
    end

    local found   = false
    for _, t in options.Gateways() do
        if gw:ID():find(t) then
            found = true
            break
        end
    end
    if not found then
        return
    end

    local msg = options.Message
    msg = msg:gsub("#i", user.ID)
    msg = msg:gsub("#n", user.Name)
    msg = msg:gsub("#a", (access[tonumber(user.Access)] or tostring(user.Access)):lower())
    msg = msg:gsub("#g", gw:ID())
    msg = msg:gsub("#d", gw:Discriminator())

    local chan = gw:Channel()
    if chan then
        msg = msg:gsub("#c", chan.Name)
    end

    if options.Public then
        gw:Say(msg)
    else
        gw:SayPrivate(user.ID, msg)
    end
end)

-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Greet users when they join the channel
--
-- Options:
--   AccessMin:  Min access level
--   AccessMax:  Max access level
--   Gateway:    Pattern for gateway ID
--   Public:     Send public instead of private message
--   Message:    Greeting template. The following placeholders are available:
--     * #i:     User ID
--     * #n:     User name
--     * #a:     Access level
--     * #g:     Gateway ID
--     * #d:     Gateway discriminator

goop = globals.goop

goop:On(events["gateway.Join"], function(ev)
    local user = ev.Arg
    local gw   = ev.Opt[1]

    local min_lvl = options["AccessMin"] or access.Default
    local max_lvl = options["AccessMax"] or access.Owner
    if user.Access < min_lvl or user.Access > max_lvl then
        return
    end

    local target = options["Gateway"] or "^capi:.*$"
    if not gw:ID():find(target) then
        return
    end

    local msg = options["Message"] or "Welcome to ##c, #n! Your access level is `#a`."
    msg = msg:gsub("#i", user.ID)
    msg = msg:gsub("#n", user.Name)
    msg = msg:gsub("#a", (access[tonumber(user.Access)] or tostring(user.Access)):lower())
    msg = msg:gsub("#g", gw:ID())
    msg = msg:gsub("#d", gw:Discriminator())

    local chan = gw:Channel()
    if chan then
        msg = msg:gsub("#c", chan.Name)
    end

    if options["Public"] then
        gw:Say(msg)
    else
        gw:SayPrivate(user.ID, msg)
    end
end)

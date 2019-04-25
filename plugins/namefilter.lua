-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Ban Battle.net users with specified patterns in their name

defoptions({
    Patterns      = {"|[cnr]"},       -- Patterns to match
    AccessProtect = access.Whitelist, -- Protected access level
    Kick          = false,            -- Kick instead of ban
})

goop:On(events.Join, function(ev)
    local user = ev.Arg
    local gw   = ev.Opt[1]

    if user.Access >= options.AccessProtect then
        return
    end

    if not gw:ID():find("^bnet:.*$") and not gw:ID():find("^capi:.*$") then
        return
    end

    local found = false
    for _, p in options.Patterns() do
        if user.Name:find(p) then
            found = true
            break
        end
    end
    if not found then
        return
    end

    if options.Kick then
        user.Access = access.Kick
    else
        user.Access = access.Ban
    end
end)

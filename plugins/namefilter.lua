-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Ban Battle.net users with specified patterns in their name
--
-- Options:
--   Patterns:       Patterns to match
--   AccessProtect:  Whitelist access level
--   Kick:           Kick instead of ban

goop:On(events.Join, function(ev)
    local user = ev.Arg
    local gw   = ev.Opt[1]

    local lvl = options["AccessProtect"] or access.Whitelist
    if user.Access >= lvl then
        return
    end

    if not gw:ID():find("^bnet:.*$") and not gw:ID():find("^capi:.*$") then
        return
    end

    local patterns = options["Patterns"] or {"@", "|[cnr]"}
    local found    = false
    for _, p in ipairs(patterns) do
        if user.Name:find(p) then
            found = true
            break
        end
    end
    if not found then
        return
    end

    if options["Kick"] then
        user.Access = access.Kick
    else
        user.Access = access.Ban
    end
end)

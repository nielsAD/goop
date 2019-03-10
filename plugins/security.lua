-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0

-- Ban users with '@', '|r', or '|n' in their name
--
-- Options:
--   AccessProtect: Whitelist access level
--   Kick:          Kick instead of ban

goop = globals.goop

goop:On(events["gateway.Join"], function(ev)
    local user = ev.Arg
    local gw   = ev.Opt[1]

    local lvl = options["AccessProtect"] or access.Whitelist
    if user.Access >= max_lvl then
        return
    end

    if not gw:ID():find("^bnet:.*$") and not gw:ID():find("^capi:.*$") then
        return
    end

    if not user.Name.find("@") and not user.Name.find("|[rn]") then
        return
    end

    if options["Kick"] then
        user.Access = access.Kick
    else
        user.Access = access.Ban
    end
end)

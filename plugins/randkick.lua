-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Add randkick command that randomly kicks a user from the channel when triggered
--
-- Options:
--   AccessTrigger:  Access level required to trigger command
--   AccessProtect:  Whitelist access level

goop:AddCommand("randkick", command(function(trig, gw)
    local lvl = options["AccessTrigger"] or access.Operator
    if trig.User.Access < lvl then
        return nil
    end

    local chan = gw:ChannelUsers()
    if chan == nil then
        return nil
    end

    local users   = {}
    local max_lvl = options["AccessProtect"] or access.Whitelist
    for _, u in chan() do
        if u.Access < max_lvl then
            table.insert(users, u)
        end
    end

    if #users == 0 then
        return nil
    end

    return gw:Kick(users[math.random(#users)].ID)
end))

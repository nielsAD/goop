-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Add randkick command that randomly kicks a user from the channel when triggered

options._default = {
    AccessTrigger = access.Operator,  -- Access level required to trigger command
    AccessProtect = access.Whitelist, -- Whitelist access level
}

goop:AddCommand("randkick", command(function(trig, gw)
    if trig.User.Access < options.AccessTrigger then
        return nil
    end

    local chan = gw:ChannelUsers()
    if chan == nil then
        return nil
    end

    local users   = {}
    for _, u in chan() do
        if u.Access < options.AccessProtect then
            table.insert(users, u)
        end
    end

    if #users == 0 then
        return nil
    end

    return gw:Kick(users[math.random(#users)].ID)
end))

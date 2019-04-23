-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Add roll/flip commands that can be used to roll a dice/flip a coin

defoptions({
    AccessTrigger = access.Voice, -- Access level required to trigger command
})

goop:AddCommand("roll", command(function(trig, _)
    if trig.User.Access < options.AccessTrigger then
        return nil
    end

    local max = 100
    if #trig.Arg > 0 then
        max = tonumber(trig.Arg[1])
        if max == nil then
            return trig.Resp("Cannot convert " .. trig.Arg[1] .. " to number")
        end
    end
    if max <= 0 then
        return trig.Resp("Pick a number above 0")
    end
    return trig.Resp(tostring(math.random(max)))

end))

goop:AddCommand("flip", command(function(trig, _)
    if trig.User.Access < options.AccessTrigger then
        return nil
    end

    local coin = "Heads"
    if math.random() < 0.5 then
        coin = "Tails"
    end
    return trig.Resp(coin)
end))

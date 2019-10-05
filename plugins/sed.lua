-- Author:  Niels A.D.
-- Project: goop (https://github.com/nielsAD/goop)
-- License: Mozilla Public License, v2.0
--
-- Find and execute sed substitute patterns in chat messages.

defoptions({
    Access      = access.Voice, -- Min access level
    History     = 10,           -- Number of messages to search back
    SkipTrigger = true,         -- Ignore messages starting with trigger
})

local regexp = require("go.regexp")
local sync   = require("go.sync")

-- sed substitute pattern, lazy match and dot matches line breaks
local sedpat, _ = regexp.Compile("(?Us)"..
    "^\\s*s/"..                                              -- ^s/
    "((?:[^\\[/\\\\]|\\\\.|\\[(?:[^\\]\\\\]|\\\\.)*\\])+)".. -- (escaped char | char group | char)+
    "/"..                                                    -- /
    "((?:[^/\\\\]|\\\\.)*)"..                                -- (escaped char | char)*
    "/([0-9A-Za-z]*)\\s*$"                                   -- /flag*$
)

local msgTab = {}
local msgMut = sync.NewMutex()

local function insertMsg(m)
    msgMut:Lock()
    table.insert(msgTab, m)
    if #msgTab > options.History then
        table.remove(msgTab, 1)
    end
    msgMut:Unlock()
end

local function findMsg(regex)
    local res = nil

    msgMut:Lock()
    for i = #msgTab, 1, -1 do
        if regex:MatchString(msgTab[i]) then
            res = msgTab[i]
            break
        end
    end
    msgMut:Unlock()

    return res
end

local function substitute(msg)
    local sed = sedpat:FindAllStringSubmatch(msg, 1)
    if not sed then
        insertMsg(msg)
        return nil
    end

    local search  = sed[1][2]
    local replace = sed[1][3]
    local flags   = sed[1][4]

    local all = false
    local mod = ""
    for i = 1, #flags do
        local f = flags:sub(i,i)
        if f == "g" then
            all = true
        elseif f == "i" or f == "I" then -- case insensitive
            mod = mod .. "i"
        elseif f == "m" or f == "M" then -- multi-line mode
            mod = mod .. "m"
        elseif f == "u" or f == "U" then -- lazy
            mod = mod .. "U"
        elseif f == "s" or f == "S" then -- dot matches line break
            mod = mod .. "s"
        end
    end

    if not all then
        search  = search .. "(?P<leftover>.*)$"
    end
    if #mod > 0 then
        search = "(?" .. mod .. ")" .. search
    end

    local regex, err = regexp.Compile(search)
    if err ~= nil then
        return nil
    end

    local m = findMsg(regex)
    if not m then
        return nil
    end

    replace = replace:gsub("%$", "$$")
    replace = replace:gsub("([^\\])\\([0-9]+)", "%1$%2"):gsub("^\\([0-9]+)", "$%1")
    replace = replace:gsub("\\\\", "\\")
    if not all then
        replace = replace .. "$leftover"
    end

    return regex:ReplaceAllString(m, replace)
end

local function starts_with(str, start)
    return str:sub(1, #start) == start
end

goop:On(events.Chat, function(ev)
    local msg = ev.Arg
    if msg.User.Access < options.Access then
        return
    end
    local gw = ev.Opt[1]
    if options.SkipTrigger and starts_with(msg.Content, gw:Trigger()) then
        return
    end

    local sub = substitute(msg.Content)
    if not sub then
        return
    end

    gw:Say(sub)
    ev:PreventNext()
end)

Commands
========

Trigger
-------

Commands can be invoked by starting a chat message with the predefined trigger (`.` by default). As a special case, `?trigger` will query the current trigger. Alternatively, mentioning the bot's name will also act as a trigger.

```
<niels>  ?trigger
<goop>   .
<niels>  .say hello
<goop>   hello
<niels>  goop: say world
<goop>   world
```

To trigger a command on another gateway, prefix the command with the gateway name or gateway type.

```
<niels>           .capi:say hello
<goop@northrend>  hello
<goop@azeroth>    hello
<niels>           .northrend:say world
<goop@northrend>  world
```


Arguments
---------

All text after the command trigger is passed on as command argument. Whitespace separates different arguments, unless they are wrapped in quotes.

```properties
# incorrect; 3 arguments
.whois username with space

# correct; 1 argument
.whois "username with space"
```

To find out how many arguments a command expects, look up its [syntax](commands_builtin.md).

?> **TIP:** Most commands accept a [glob pattern](https://en.wikipedia.org/wiki/Glob_(programming)#Syntax) as input. This can be useful to target several users at once.  
**For example:** Executing [.kick](commands_builtin.md#kick) `4k*` will kick all users from channel that have a name starting with 4k.


Access level
------------
Goop uses a hierarchical access model. Access levels determine what commands can be accessed, whether or not messages will be relayed to other realms, and whether or not the user will be banned upon joining the channel. 

|Role        | Level|Description|
|------------|-----:|-----------|
|`owner`     | 1000 | Bot owner, has access to everything. |
|`admin`     | 300  | Administrator, has access to everything except settings. |
|`operator`  | 200  | Channel operator, can kick/ban. |
|`whitelist` | 100  | Trusted user, only kickable/bannable by admins. |
|`voice`     | 1    | Chat will be relayed between gateways. |
|`ignore`    | -1   | Ignore user, do not relay chat and do not process commands. |
|`kick`      | -100 | Auto kick. |
|`ban`       | -200 | Auto ban. |
|`blacklist` | -300 | Auto ban, only unbannable by admins. |

An access level can be assigned to a particular user (with the [.set](commands_builtin.md#set) command) or to a particular group (such as users with a certain role on Discord or users from a certain clan on Battle.net).


Config
------

_Default config:_
```toml
[Default.Commands]
  Access         = "voice"
  RespondPrivate = false
  Trigger        = "."

# For each built-in command:
[Commands]
  [Commands.Who]
    Disabled   = false
    Priviledge = ""
  [Commands.Whoami]
    Disabled   = false
    Priviledge = ""
  [Commands.Whois]
    Disabled   = false
    Priviledge = "admin"

# [...]
```
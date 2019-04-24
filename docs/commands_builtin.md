Built-in Commands
=================

| Command                 | Arguments        | Access    | BNet  | CAPI  |Discord|
|-------------------------|------------------|-----------|:-----:|:-----:|:-----:|
|[settings](#settings)    |action, key, value|`owner`    |&check;|&check;|&check;|
|[sayprivate](#sayprivate)|username, message |`admin`    |&check;|&check;|&check;|
|[whois](#whois)          |username          |`admin`    |&check;|&check;|&check;|
|[set](#set)              |username, access  |`admin`    |&check;|&check;|&check;|
|[unset](#unset)          |username          |`admin`    |&check;|&check;|&check;|
|[list](#list)            |access            |`operator` |&check;|&check;|&check;|
|[ban](#ban)              |username          |`operator` |&check;|&check;|&cross;|
|[unban](#unban)          |username          |`operator` |&check;|&check;|&cross;|
|[kick](#kick)            |username          |`operator` |&check;|&check;|&cross;|
|[echo](#echo)            |message           |`whitelist`|&check;|&check;|&check;|
|[say](#say)              |message           |`whitelist`|&check;|&check;|&check;|
|[whisper](#whisper)      |message           |`whitelist`|&check;|&check;|&check;|
|[ping](#ping)            |username          |`whitelist`|&check;|&cross;|&cross;|
|[pingme](#pingme)        |                  |           |&check;|&cross;|&cross;|
|[whoami](#whoami)        |                  |           |&check;|&check;|&check;|
|[where](#where)          |                  |           |&check;|&check;|&check;|
|[who](#who)              |                  |           |&check;|&check;|&check;|
|[time](#time)            |                  |           |&check;|&check;|&check;|
|[uptime](#uptime)        |                  |           |&check;|&check;|&check;|

<br>
<hr>

## Settings
|||
|--------------------:|-|
| Access              |[`owner`](commands.md#access-level)|
| Syntax              |`.settings [action] [key] [new_value]`|
|_<sub>[action]</sub>_|Find, get, set, or unset.|
|_<sub>[key]</sub>_   |Key for setting (accepts [glob pattern](commands.md#arguments)).|
|_<sub>[value]</sub>_ |New value for setting (only used when _[action]_ is _set_).|

Examine and update configuration at runtime. Refer to [Getting Started](config.md#at-runtime) for more elaborate examples.

_Example:_
```properties
.settings find commands trigger
.settings get default/commands/trigger
.settings set default/commands/trigger newtrigger
.settings unset default/commands/trigger
```


## SayPrivate
|||
|----------------------:|-|
| Access                |[`admin`](commands.md#access-level)|
| Syntax                |`.sayprivate [username] [message...]`|
|_<sub>[username]</sub>_|Target user (accepts [glob pattern](commands.md#arguments)).|
|_<sub>[message]</sub>_ |Message to send.|

Whisper `[message]` to `[username]`.

_Example:_
```properties
.sayprivate moon Hi there!
```


## Whois
|||
|----------------------:|-|
| Access                |[`admin`](commands.md#access-level)|
| Syntax                |`.whois [username]`|
|_<sub>[username]</sub>_|Target user (accepts [glob pattern](commands.md#arguments)).|

Display ID and Access Level for `[username]`.

_Example:_
```properties
.whois happy
```


## Set
|||
|----------------------:|-|
| Access                |[`admin`](commands.md#access-level)|
| Syntax                |`.set [username] [access]`|
|_<sub>[username]</sub>_|Target user (accepts [glob pattern](commands.md#arguments)).|
|_<sub>[access]</sub>_  |[Access level](commands.md#access-level).|

Change access level for `[username]` to `[level]`.

_Example:_
```properties
.set niels admin+1
.set grubby admin
.set tod 100
```

_Aliases:_
* `.whitelist [username]` is an alias for `.set [username] whitelist`
* `.blacklist [username]` is an alias for `.set [username] blacklist`
* `.ignore [username]` is an alias for `.set [username] ignore`
* `.squelch [username]` is an alias for `.set [username] ignore`


## Unset
|||
|----------------------:|-|
| Access                |[`admin`](commands.md#access-level)|
| Syntax                |`.unset [username]`|
|_<sub>[username]</sub>_|Target user (accepts [glob pattern](commands.md#arguments)).|

Revert access level for `[username]` to default.

_Example:_
```properties
.unset grubby
```

_Aliases:_
* `.unwhitelist` is an alias for `.unset`
* `.unblacklist` is an alias for `.unset`
* `.unignore` is an alias for `.unset`
* `.unsquelch` is an alias for `.unset`


## List
|||
|------------------------:|-|
| Access                  |[`operator`](commands.md#access-level)|
| Syntax                  |`.list [min-access] [max-access]`|
|_<sub>[min-access]</sub>_|[Access level](commands.md#access-level).|
|_<sub>[max-access]</sub>_|[Access level](commands.md#access-level) (optional).|

List all users with `[min-access]` &le; access level &le; `[max-access]`.

_Example:_
```properties
.list admin
.list voice whitelist
```

_Aliases:_
* `.l` is an alias for `*:list` (list on all gateways)
* `.banlist` is an alias for `.list min ban`


## Ban
|||
|----------------------:|-|
| Access                |[`operator`](commands.md#access-level)|
| Syntax                |`.ban [username]`|
|_<sub>[username]</sub>_|Target user (accepts [glob pattern](commands.md#arguments)).|

Ban `[username]` from channel.

_Example:_
```properties
.ban grubby
.ban *niels*
```

_Aliases:_
* `.b` is an alias for `.capi:ban` (ban on all [CAPI](bnet.md#capi) gateways)


## Unban
|||
|----------------------:|-|
| Access                |[`operator`](commands.md#access-level)|
| Syntax                |`.unban [username]`|
|_<sub>[username]</sub>_|Target user (accepts [glob pattern](commands.md#arguments)).|

Unban `[username]`.

_Example:_
```properties
.unban grubby
```


## Kick
|||
|----------------------:|-|
| Access                |[`operator`](commands.md#access-level)|
| Syntax                |`.kick [username]`|
|_<sub>[username]</sub>_|Target user (accepts [glob pattern](commands.md#arguments)).|

Kick `[username]` from channel.

_Example:_
```properties
.kick moon
.kick *niels*
```

_Aliases:_
* `.k` is an alias for `.capi:kick` (whisper on all [CAPI](bnet.md#capi) gateways)


## Echo
|||
|---------------------:|-|
| Access               |[`whitelist`](commands.md#access-level)|
| Syntax               |`.echo [message...]`|
|_<sub>[message]</sub>_|Message to send.|

Echo `[message]`.

_Example:_
```properties
.echo Hello, world!
```


## Say
|||
|---------------------:|-|
| Access               |[`whitelist`](commands.md#access-level)|
| Syntax               |`.say [message...]`|
|_<sub>[message]</sub>_|Message to send.|

Say `[message]` in channel.

_Example:_
```properties
.say I come from the darkness of the pit.
```

_Aliases:_
* `.s` is an alias for `.say` (say on all gateways)


## Whisper
|||
|----------------------:|-|
| Access                |[`whitelist`](commands.md#access-level)|
| Syntax                |`.whisper [username] [message...]`|
|_<sub>[username]</sub>_|Target user (accepts [glob pattern](commands.md#arguments)).|
|_<sub>[message]</sub>_ |Message to send.|

Whisper `[message]` to `[username]`, with username of the sender prepended.

_Example:_
```properties
.whisper I come from the darkness of the pit.
```

_Aliases:_
* `.w` is an alias for `.capi:whisper` (whisper on all [CAPI](bnet.md#capi) gateways)


## Ping
|||
|----------------------:|-|
| Access                |[`whitelist`](commands.md#access-level)|
| Syntax                |`.ping [username]`|
|_<sub>[username]</sub>_|Target user (accepts [glob pattern](commands.md#arguments)).|

Display ping to `[username]` in milliseconds.

_Example:_
```properties
.ping moon
```


## PingMe
|||
|----------------------:|-|
| Access                |[Default (0)](commands.md#access-level)|
| Syntax                |`.pingme`|

Display [.ping](#ping) info for invoking user.

_Example:_
```properties
.pingme
```


## Whoami
|||
|----------------------:|-|
| Access                |[Default (0)](commands.md#access-level)|
| Syntax                |`.whoami`|

Display [.whois](#whois)  info for invoking user.

_Example:_
```properties
.whoami
```


## Where
|||
|----------------------:|-|
| Access                |[Default (0)](commands.md#access-level)|
| Syntax                |`.where`|

List connected gateways.

_Example:_
```properties
.where
```


## Who
|||
|----------------------:|-|
| Access                |[Default (0)](commands.md#access-level)|
| Syntax                |`.who`|

List users online on other gateways.

_Example:_
```properties
.who
```


## Time
|||
|----------------------:|-|
| Access                |[Default (0)](commands.md#access-level)|
| Syntax                |`.time`|

Print current time.

_Example:_
```properties
.time
```


## Uptime
|||
|----------------------:|-|
| Access                |[Default (0)](commands.md#access-level)|
| Syntax                |`.uptime`|

Print running time.

_Example:_
```properties
.uptime
```

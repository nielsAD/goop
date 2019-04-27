Plugins
=======

Goop can be customized and extended with Lua plugins. [Lua](https://www.lua.org/manual/5.1/manual.html) is a powerful, light-weight scripting language.  
To load a plugin, add a configuration section with its name. Goop ships with numerous plugins by default.

_Default config:_

```toml
[Plugins.Default]
  Path = ""
  CallStackSize = 0
  RegistrySize = 0
  Options = {}
```

_Example:_
```toml
# Load plugins/announce.lua
[Plugins.announce]

# Load plugins/update.lua with specified options
[Plugins.update]
  Options = { Interval = "30m" }
```


## announce
Add `.announce` command that can repeat a message at specified interval.

_Options:_

|    Option   | Default  | Description |
|-------------|----------|-------------|
|AccessTrigger|`operator`|Access level required to trigger command.|


## cmdfilter 
Prevent relay of commands.


## copycat
Echo everything said in channel or private message.

_Options:_

|    Option   | Default  | Description |
|-------------|----------|-------------|
|Access       |`voice`   |Min access level.|


## debug
Print all events.


## greet
Greet users when they join the channel.

_Options:_

|    Option   | Default  | Description |
|-------------|----------|-------------|
|AccessMin    |`voice`                                           |Min access level.|
|AccessMax    |`owner`                                           |Max access level.|
|Gateways     |`["^bnet:.*$", "^capi:.*$"]`                      |Gateway IDs (accepts [Lua pattern](https://www.lua.org/manual/5.1/manual.html#6.4.1)).|
|Public       |`false`                                           |Public instead of private message.|
|Message      |`"Welcome to ##c, #n! Your access level is <#a>."`|Greeting template.|

_Message placeholders:_

| Placeholder | Description |
|-------------|-------------|
| _#i_        | User ID |
| _#n_        | User name |
| _#a_        | Access level |
| _#g_        | Gateway ID |
| _#d_        | Gateway discriminator |


## namefilter
Ban Battle.net users with specified patterns in their name.

_Options:_

|    Option   |   Default   | Description |
|-------------|-------------|-------------|
|Patterns     |`["\|[cnr]"]`|Patterns to match (accepts [Lua pattern](https://www.lua.org/manual/5.1/manual.html#6.4.1)).|
|AccessProtect|`whitelist`  |Protected access level.|
|Kick         |`false`      |Kick instead of ban.|


## randkick
Add `.randkick` command that randomly kicks a user from the channel.

_Options:_

|    Option   |  Default  | Description |
|-------------|-----------|-------------|
|AccessTrigger|`operator` |Access level required to trigger command.|
|AccessProtect|`whitelist`|Protected access level.|


## roll
Add `.roll` and `.flip` commands that can be used to roll a dice, flip a coin.

_Options:_

|    Option   | Default  | Description |
|-------------|----------|-------------|
|AccessTrigger|`voice`   |Access level required to trigger command.|


## sed
Find and execute sed substitute patterns in chat messages.

_Options:_

|    Option   | Default  | Description |
|-------------|----------|-------------|
|Access       |`voice`   |Min access level.|
|History      |`10`      |Number of messages to search back.|
|SkipTrigger  |`true`    |Ignore messages starting with trigger.|


## update
Check for available updates.

_Options:_

|    Option   | Default  | Description |
|-------------|----------|-------------|
|Interval     |`"24h"`   |Check interval.|


## weather
Add weather command that prints weather on location.

_Options:_

|     Option    |  Default  | Description |
|---------------|-----------|-------------|
|AccessTrigger  |`whitelist`|Access level required to trigger command.|
|DefaultLocation|`""`       |Default location.|

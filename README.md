Goop
===========
[![Build Status](https://travis-ci.org/nielsAD/goop.svg?branch=master)](https://travis-ci.org/nielsAD/goop)
[![Build status](https://ci.appveyor.com/api/projects/status/vfft3xr00pk2vpnu/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/goop)
[![GoDoc](https://godoc.org/github.com/nielsAD/goop?status.svg)](https://godoc.org/github.com/nielsAD/goop)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

Goop (GO OPerator) is a BNCS Channel Operator.

Features:

* Battle.net & Discord integration
* Relay chat between gateways
* Channel moderation
  * Ban/kick commands
  * Persist bans
  * Whitelist/blacklist
* Fully configurable
* Cross platform (Windows, Linux, OSX)


Installation
------------

Official binaries are [available](https://github.com/nielsAD/goop/releases/latest). Simply download and run.

Configuration
-------------
Configuration is stored in [TOML](https://github.com/toml-lang/toml/blob/master/versions/en/toml-v0.4.0.md) files. By default, Goop tries to load `config.toml` in the working directory. See [`config.toml.example`](/config.toml.example) for a minimal example. Running `./goop -makeconf` will generate a fresh configuration file containing all default values.


### Gateways
Goop can connect to one or multiple gateways. Add a configuration section for each gateway.

##### Battle.net (API)
As of 2018, classic Battle.net provides an official chatbot API. Use this gateway if you want to connect to official Blizzard servers. Register an API key by executing the `/register-bot` command in your preferred channel (limited to Op and Clan channels). The bot will be limited to the registered channel.

```toml
# Chatbot API connection
[Capi.Gateways."{capi_name}"]
APIKey = "{APIKey}"
```

##### Battle.net (CD-Keys)
This connection method has been superseded by the official chatbot API, but can still be used to connect to non-official Battle.net servers.

```toml
# Battle.net account (defunct for official servers)
[BNet.Gateways."{bnet_name}"]
ServerAddr = "europe.battle.net"
CDKeys     = ["{ROC-CDKEY}", "{TFT-CDKEY}"]
Username   = "{username}"
Password   = "{password}"
```

##### Discord
Get an [authorization token and channel ID](https://github.com/Chikachi/DiscordIntegration/wiki/How-to-get-a-token-and-channel-ID-for-Discord) for Discord.

```toml
# Discord connection
[Discord.Gateways."{discord_name}"]
AuthToken = "{AuthorizationToken}"

    # Channel configuration
    [Discord.Gateways.bridge.Channels.{channel_id}]
```

### Relay
By default, only chat events are relayed between gateways. The `Relay` configuration section can be used to change this.

```toml
[Relay]
    # Relay everything everywhere by default
    [Relay.Default]
    Log         = true
    System      = true
    Channel     = true
    Joins       = true
    Chat        = true
    PrivateChat = true
    Say         = true

    # For BNet -> Discord, only relay chat + whispers + system messages
    [Relay.To."discord:{discord_name}:{channel_id}".From."capi:{capi_name}"]
    Joins  = true
    Chat   = true
    System = true

    # For Discord -> BNet, only relay chat
    [Relay.To."capi:{capi_name}".From."discord:{discord_name}:{channel_id}"]
    Chat = true
```

### Access level

Access levels determine what commands can be accessed, whether or not messages will be relayed to other realms, and whether or not the user will be banned upon joining the channel. 

|Role        | Level|Description|
|------------|-----:|-----------|
|`owner`     | 1000 | Bot owner, has access to all commands and can appoint admins. |
|`admin`     | 300  | Administrator, has access to everything except settings and can appoint operators. |
|`operator`  | 200  | Channel operator, can kick/ban and appoint whitelisted users. |
|`whitelist` | 100  | Trusted user, only kickable/bannable by admins.  |
|`voice`     | 1    | Chat will be relayed between gateways. |
|`ignore`    | -1   | Ignore user, do not relay chat and do not process commands. |
|`kick`      | -100 | Auto kick. |
|`ban`       | -200 | Auto ban. |
|`blacklist` | -300 | Auto ban, only unbannable by admins. |

An access level can be assigned to a particular user (with the `.set {user} {level}` command) or to a particular group (such as users with a certain role on Discord or users from a certain clan on Battle.net). The configuration file can be used to assign access levels as well.

```toml
[Capi]
    [Capi.Gateways.northrend]
    APIKey = "{APIKey}"

    # Access level based on user name
    AccessUser = { niels = "owner", pepernooooooot = "admin" }

    # Give user access level "operator" to channel moderators
    AccessOperator = "operator"

[Discord]
    [Discord.Gateways.discord]
    AuthToken = "{AuthorizationToken}"

        [Discord.Gateways.discord.Channels.{channel_id}]
        # Access level based on user ID
        AccessUser = { 193313355237294080 = "owner" }

        # Access level based on role
        AccessRole = { chief = "admin", shaman = "operator" }
```

Commands
--------

|Command                        |Alias                    |Access     |Description|
|-------------------------------|-------------------------|-----------|-----------|
|`.settings [action] [key]`     |                         |`owner`    |Manage configuration. Action must be one of `find`, `get`, `set`, `unset`.|
|`.sayprivate [user] [message]` |`.whisper`, `.w`         |`admin`    |Whisper `[message]` to `[user]`.|
|`.whois [user]`                |                         |`admin`    |Display ID and Access Level for `[user]`.|
|`.set [user] [level]`          |                         |`operator` |Change access level for `[user]` to `[level]`.|
|`.unset [user]`                |`.unignore`, `.unsquelch`|`operator` |Revert access level for `[user]` to default.|
|`.list [access]`               |`.l`                     |`operator` |List all users with access level `[access]`.|
|`.ban [user]`                  |`.b`                     |`operator` |Ban `[user]` from channel.|
|`.unban [user]`                |                         |`operator` |Unban `[user]`.|
|`.kick [user]`                 |`.k`                     |`operator` |Kick `[user]` from channel.|
|`.ignore [user]`               |`.squelch`, `.i`         |`operator` |Change access level for `[user]` to ignore.|
|`.say [message]`               |`.s`                     |`whitelist`|Echo `[message]`.|
|`.ping [user]`                 |`.pingme`, `.p`          |`whitelist`|Display ping to `[user]` in milliseconds.|
|`.whoami`                      |                         |           |Display `.whois` info for invoking user.|
|`.where`                       |                         |           |List connected gateways.|
|`.time`                        |                         |           |Print current time.|
|`.uptime`                      |                         |           |Print running time.|
|`.flip`                        |                         |           |Flip a coin.|
|`.roll [n]`                    |                         |           |Roll a dice with `[n]` sides.|

#### Glob

Most commands accept a [glob pattern](https://en.wikipedia.org/wiki/Glob_(programming)#Syntax) as input. This can be useful to target several users at once. For example, `.kick 4k*` will kick all users from channel that have a name starting with 4k (`4k`, `4k.grubby`, but not `niels.4k`).

#### Trigger

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


Build
-----

```bash
# Linux dependencies
apt-get install --no-install-recommends -y build-essential cmake git golang-go libgmp-dev libbz2-dev zlib1g-dev

# OSX dependencies
brew install cmake git go gmp bzip2 zlib

# Windows dependencies (use MSYS2 -- https://www.msys2.org/)
pacman --needed --noconfirm -S git mingw-w64-x86_64-toolchain mingw-w64-x86_64-go mingw-w64-x86_64-cmake

# Download vendor submodules
git submodule update --init --recursive

# Run tests
make test

# Build release files in ./bin/
make release
```
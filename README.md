Goop
===========
[![Build Status](https://travis-ci.org/nielsAD/goop.svg?branch=master)](https://travis-ci.org/nielsAD/goop)
[![Build status](https://ci.appveyor.com/api/projects/status/vfft3xr00pk2vpnu/branch/master?svg=true)](https://ci.appveyor.com/project/nielsAD/goop)
[![GoDoc](https://godoc.org/github.com/nielsAD/goop?status.svg)](https://godoc.org/github.com/nielsAD/goop)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

Goop (GO OPerator) is a BNCS Channel Operator.

Installation
------------

Official binaries are [available](https://github.com/nielsAD/goop/releases/latest). Simply download and run.

Configuration
-------------
Configuration is stored in [TOML](https://github.com/toml-lang/toml/blob/master/versions/en/toml-v0.4.0.md) files. By default, Goop tries to load `config.toml` in the working directory. See [`config.toml.example`](/config.toml.example) for a minimal example. Running `./goop -makeconf` will generate a fresh configuration file containing all default values.


### Gateways
Goop can connect to one or multiple gateways. For each gateway, add a section to the configuration.

##### Battle.net (API)
As of 2018, classic Battle.net provides an official chatbot API. Use this gateway if you want to connect to official Blizzard servers. Register an API key by executing the `/register-bot` command in your preferred channel (limited to Op and Clan channels). The bot will be limited to the registered channel.

```toml
[Capi.Gateways."{capi_name}"]
APIKey = "{APIKey}"
```

##### Battle.net (CD-Keys)
This connection method has been superseded by the official chatbot API, but can still be used to connect to non-official Battle.net servers.

```toml
[BNet.Gateways."{bnet_name}"]
ServerAddr = "europe.battle.net"
CDKeys     = ["{ROC-CDKEY}", "{TFT-CDKEY}"]
Username   = "{username}"
Password   = "{password}"
```

##### Discord
Get an [authorization token and channel ID](https://github.com/Chikachi/DiscordIntegration/wiki/How-to-get-a-token-and-channel-ID-for-Discord) for Discord.

```toml
[Discord.Gateways."{discord_name}"]
AuthToken = "{AuthorizationToken}"

    [Discord.Gateways.bridge.Channels.{ChannelID}]
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
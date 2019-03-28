Battle.net
==========
Goop can connect to one or multiple gateways. Add a configuration section to [`config.toml`](config.md) for each gateway.

CAPI
----

As of 2018, classic Battle.net provides an official Chatbot API (CAPI). Use this gateway if you want to connect to official Blizzard servers. Register an API key by executing the `/register-bot` command in your preferred channel (limited to Op and Clan channels). The bot will be restricted to the registered channel.

_Default config:_
```toml
[Capi.Default]
  APIKey = ""
  AccessOperator = ""
  AccessTalk = "voice"
  AccessWhisper = "ignore"
  AvatarDefaultURL = ""
  BufSize = 16
  Endpoint = "wss://connect-bot.classic.blizzard.com/v1/rpc/chat"
  RPCTimeout = 0
  ReconnectDelay = "30s"
  AccessUser = {}
```

_Example:_
```toml
[Capi.Gateways.Northrend]
  APIKey     = "00000000000000000000000000000000000000000000000000000000"
  AccessUser = { niels = "owner", grubby = "admin" }
```


CD-Keys
-------

!> **NOTE:** This connection method has been superseded by the official chatbot API, but can still be used to connect to unofficial Battle.net servers. THIS WILL NOT WORK ON OFFICIAL SERVERS!

_Default config:_
```toml
[BNet.Default]
  AccessNoWarcraft = ""
  AccessOperator = ""
  AccessTalk = "voice"
  AccessWhisper = "ignore"
  AvatarDefaultURL = ""
  AvatarIconURL = ""
  BinPath = ""
  BufSize = 16
  CDKeyOwner = ""
  CDKeys = []
  ExeHash = 0
  ExeInfo = ""
  ExeVersion = 0
  GamePort = 0
  HomeChannel = ""
  KeepAliveInterval = 0
  Password = ""
  ReconnectDelay = "30s"
  SHA1Auth = false
  ServerAddr = ""
  Username = ""
  VerifySignature = false
  AccessClanTag = {}
  AccessLevel = {}
  AccessUser = {}
```

_Example:_
```toml
[BNet.Gateways.Northrend]
  ServerAddr = "europe.battle.net"
  Username   = "chatbot"
  Password   = "hunter2"
  AccessUser = { niels = "owner", grubby = "admin" }
```

_Example (PvPGN):_
```toml
[BNet.Gateways.Rubattle]
  ServerAddr = "rubattle.net"
  CDKeys     = ["FFFFFFFFFFFFFFFFFFFFFFFFFF", "FFFFFFFFFFFFFFFFFFFFFFFFFF"]
  ExeVersion = 0x011b01ad
  ExeHash    = 0xaaaba048
  SHA1Auth   = true
  Username   = "chatbot"
  Password   = "hunter2"
  AccessUser = { niels = "owner", grubby = "admin" }
```
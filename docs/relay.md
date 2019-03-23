Relay
=====

Naming
------

Each configuration section follows a similar structure.

_Example config:_
```toml
[Capi.Gateways."{capi_name_1}"]
  APIKey = "{APIKey1}"

[Capi.Gateways."{capi_name_2}"]
  APIKey = "{APIKey2}"

[Discord.Gateways."{discord_name}"]
  AuthToken = "{AuthorizationToken}"
```

Config
------

By default, only chat events are relayed between gateways. The `Relay` configuration section can be used to change this.

_Example config:_
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

Relay
=====

By default, only chat events are relayed between gateways. The `Relay` configuration section can be used to change this.


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

In this example, curly braces mark *gateway identifiers*. These identifiers are used in the relay configuration section to refer to gateways.


Config
------

_Default config:_
```toml
[Relay]
  # Relay to other gateways
  [Relay.Default]
    Log         = false
    System      = false
    Channel     = false
    Joins       = false
    Say         = true
    Chat        = true
    PrivateChat = true
    JoinAccess        = ""
    ChatAccess        = "voice"
    PrivateChatAccess = "voice"

  # Relay to sender
  [Relay.DefaultSelf]
    Log         = false
    System      = false
    Channel     = false
    Joins       = false
    Say         = false
    Chat        = false
    PrivateChat = true
    JoinAccess        = ""
    ChatAccess        = ""
    PrivateChatAccess = "voice"
```

_Example:_
```toml
[Relay]
  # Relay chat + joins from all other gateways to Discord
  [Relay.To."discord:{discord_name}:{channel_id}".Default]
    Joins      = true
    Say        = true
    Chat       = true
    JoinAccess = "min"
    ChatAccess = "voice"

  # Only relay joins from Capi to Discord
  [Relay.To."discord:{discord_name}:{channel_id}".From."capi:{capi_name}"]
    Joins = true
```

### Precedence
To find the relay configuration between two specific gateways, Goop searches in the following order and uses the first found section:

1. `[Relay.To."A".From."B"]`
2. `[Relay.To."A".Default]`
3. `[Relay.Default]`

!> Note that the relay subsections do not merge in `Default` records! Contrary to gateway configuration sections, the `Default` record is only used when a subsection between two gateways is not defined, it is not used as fallback for individual undefined fields.
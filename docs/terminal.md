Terminal
========

Usage
-----

`./goop [options] [config_file...]`

|    Flag   |  Type  | Description |
|-----------|--------|-------------|
|`-makeconf`|`string`|Generate a configuration file|

Configuration is primarily done using [config files](config.md). By default, Goop tries to load `config.toml` in the working directory.  Multiple config files can be passed as arguments, they will be merged and loaded into a single application instance.


StdIO
-----

StdIO is a predefined gateway that connects to [standard streams](https://en.wikipedia.org/wiki/Standard_streams) (stdin, stdout, stderr). It acts as main gateway and prints all events to the terminal (including errors and log messages). Any messages typed in the terminal will be forwarded as chat.

_Default config:_
```toml
[StdIO]
  Access = "owner+1"
  AvatarURL = ""
  Read = true

# Everything is relayed to terminal by default
[Relay.To."std:io".Default]
  Log         = true
  System      = true
  Channel     = true
  Joins       = true
  Say         = true
  Chat        = true
  PrivateChat = true
```


Log Format
----------

Configure what to prepend to each line of text.

_Default config:_
```toml
[Log]
  Date = false
  Microseconds = false
  Time = true
  UTC = false
```

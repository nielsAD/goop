Custom Commands
===============

Alias
-----

Aliases can be added to the `Commands.Alias` configuration section. An alias defines a shortcut for another (often longer) command.

_Example:_
```toml
# Alias .hello with ".echo Hello, world!"
[Commands.Alias.hello]
  Exe = "echo"
  Arg = ["Hello, world!"]
```

### Placeholders

A command alias will forward any arguments passed to it, but can apply mutations before doing so.  
The following placeholders can be used in `Arg` strings.

| Placeholder | Description |
|:-----------:|-------------|
|`%CMD%`      | Command name |
|`%RAW%`      | Raw arguments |
|`%ARGS%`     | Arguments (with processed string quotes) |
|`%NARGS%`    | Number of arguments |
|`%UID%`      | Ivoking user ID |
|`%USTR%`     | Ivoking user name |
|`%ULVL%`     | Ivoking user access level |
|`%GWID%`     | Target gateway ID |
|`%GWDIS%`    | Target gateway discriminator |
|`%GWDEL%`    | Target gateway delimiter |
|`%RARG123%`  | Raw argument at index 123 (i.e. `raw[123]`) |
|`%..RARG123%`| Raw arguments until index 123 (i.e. `raw[:123]`) |
|`%RARG123..%`| Raw arguments from index 123 onward (i.e. `raw[123:]`) |
|`%ARG123%`   | Argument at index 123 (i.e. `arg[123]`) | 
|`%..ARG123%` | Arguments until index 123 (i.e. `arg[:123]`) |
|`%ARG123..%` | Arguments from index 123 onward (i.e. `arg[123:]`) |

<br>

_Example:_
```toml
# Alias .whisper with ".sayprivate [username] <sender_name> message..."
[Commands.Alias.whisper]
  Priviledge  = "whitelist"
  Exe         = "sayprivate"
  ArgExpected = 2
  Arg         = ["%ARG1%", "<%USTR%> %RARG2..%"]

  # Execute .sayprivate as admin, but require only whitelist to execute .whisper
  WithPriviledge = "admin"
```


Plugins
-------

Plugins allow elaborate logic for custom commands in cases where aliases are insufficient.  
Refer to the [plugin API](plugins_api.md) for more information on how to write plugins.

_Example (flip a coin):_
```lua
-- Flip command, randomly flips a coin and prints the result
goop:AddCommand("flip", command(function(trigger, _)
    local coin = "Heads"
    if math.random() < 0.5 then
        coin = "Tails"
    end
    return trigger.Resp(coin)
end))
```
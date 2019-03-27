Configure
=========

File
----

Configuration files are stored in [TOML](https://github.com/toml-lang/toml/blob/master/versions/en/toml-v0.4.0.md) format. By default, Goop tries to load `config.toml` in the working directory.  
Edit the file in your text editor of choice (even notepad will do).

A minimal example:

```toml
# config.toml
# LINES STARTING WITH # ARE IGNORED, THEY ARE USED FOR COMMENTS

[Capi.Gateways.BNet]
    # Register your bot on Battle.net with /register-bot and
    # replace the zeroes with the API key you receive per email.
    APIKey = "00000000000000000000000000000000000000000000000000000000"

    # Assign roles to users. Pick one of:
    # owner,admin,operator,whitelist,voice,ignore,kick,ban,blacklist
    AccessUser = { niels = "owner", grubby = "admin" }

# Load plugins
[Plugins.announce]
[Plugins.update]
[Plugins.weather]
```

?> **TIP:** To reload the configuration while running, restart Goop with the `.restart` command.

?> **TIP:** Running `goop -makeconf` will generate a fresh configuration file containing all default values.

At Runtime
----------

Configuration files are static and only loaded during start-up. To make changes while Goop is running, use the `.settings` command. These changes are saved on disk and will be loaded in future runs as well.

!> **NOTE:** It is preferable to make changes to `config.toml` directly rather than using the `.settings` command. Use the command to experiment with settings and make temporary changes.

!> **NOTE:** Not all changes can be applied at runtime, some require a restart (i.e. connection settings).

Examples usage:

##### Find all settings with `capi` and `apikey` in their name
  * _command_
  ```
  .settings find capi apikey
  ```
  * _response_
  ```
  > capi/default/apikey = 
  > capi/gateways/bnet/apikey = 0000000000
  ```

#####  Get the current value of `capi/gateways/bnet/accesstalk`
  * _command_ 
  ```
  .settings get capi/gateways/bnet/accesstalk
  ```
  * _response_
  ```
  > capi/gateways/bnet/accesstalk = voice
  ```

#####  Change value of `capi/gateways/bnet/accesstalk` to `ignore`
  * _command_
  ```
  .settings set capi/gateways/bnet/accesstalk ignore
  ```
  * _response_
  ```
  > Changed capi/gateways/bnet/accesstalk from voice to ignore
  ```

#####  Revert the value of `capi/gateways/bnet/accesstalk` to its default
  * _command_
  ```
  .settings unset capi/gateways/bnet/accesstalk
  ```
  * _response_
  ```
  > Unset capi/gateways/bnet/accesstalk = ignore
  ```

<br>

?> **TIP:** Use `*` as a wildcard and get/set multiple settings at once!  
**For example**: `.settings set capi/gateways/*/accesstalk ignore`

Structure
---------

The configuration structure directly correlates with the `Config` struct in [`config.go`](https://github.com/nielsAD/goop/blob/master/config.go).  
Examining the source code is the best way to find out exactly how settings are used.

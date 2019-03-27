// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/bnet"
	"github.com/nielsAD/goop/gateway/capi"
	"github.com/nielsAD/goop/gateway/discord"
	"github.com/nielsAD/goop/gateway/stdio"
	"github.com/nielsAD/goop/goop"
	"github.com/nielsAD/goop/goop/cmd"
	"github.com/nielsAD/goop/goop/plugin"
	"github.com/nielsAD/gowarcraft3/network/chat"
	pcapi "github.com/nielsAD/gowarcraft3/protocol/capi"
)

// DefaultConfig values used as fallback
func DefaultConfig() *Config {
	return &Config{
		Config: "./config.persist.toml",
		Log: LogConfig{
			Time: true,
		},
		Commands: CommandsConfig{
			Commands: cmd.Commands{
				Settings: cmd.Settings{
					Cmd: cmd.Cmd{Priviledge: gateway.AccessOwner},
				},
				Whois: cmd.Whois{
					Cmd: cmd.Cmd{Priviledge: gateway.AccessAdmin},
				},
				SayPrivate: cmd.SayPrivate{
					Cmd: cmd.Cmd{Priviledge: gateway.AccessAdmin},
				},
				Echo: cmd.Echo{
					Cmd: cmd.Cmd{Priviledge: gateway.AccessWhitelist},
				},
				Say: cmd.Say{
					Cmd: cmd.Cmd{Priviledge: gateway.AccessWhitelist},
				},
				List: cmd.List{
					Cmd: cmd.Cmd{Priviledge: gateway.AccessOperator},
				},
				Set: cmd.Set{
					Cmd:           cmd.Cmd{Priviledge: gateway.AccessOperator},
					DefaultAccess: gateway.AccessWhitelist,
				},
				Kick: cmd.Kick{
					Cmd:            cmd.Cmd{Priviledge: gateway.AccessOperator},
					AccessProtect:  gateway.AccessWhitelist,
					AccessOverride: gateway.AccessAdmin,
				},
				Ban: cmd.Ban{
					Cmd:            cmd.Cmd{Priviledge: gateway.AccessOperator},
					AccessProtect:  gateway.AccessWhitelist,
					AccessOverride: gateway.AccessAdmin,
				},
				Unban: cmd.Unban{
					Cmd:            cmd.Cmd{Priviledge: gateway.AccessOperator},
					AccessProtect:  gateway.AccessBlacklist,
					AccessOverride: gateway.AccessAdmin,
				},
				Ping: cmd.Ping{
					Cmd: cmd.Cmd{Priviledge: gateway.AccessWhitelist},
				},
				Time: cmd.Time{
					Format: "15:04:05 MST",
				},
			},
			Alias: map[string]*cmd.Alias{
				"whisper": &cmd.Alias{
					Cmd:            cmd.Cmd{Priviledge: gateway.AccessWhitelist},
					Exe:            "sayprivate",
					Arg:            []string{"%ARG1%", "<%USTR%> %ARG2..%"},
					ArgExpected:    2,
					WithPriviledge: gateway.AccessAdmin,
				},
				"pingme": &cmd.Alias{
					Cmd:            cmd.Cmd{Priviledge: gateway.AccessVoice},
					Exe:            "ping",
					Arg:            []string{"%USTR%"},
					WithPriviledge: gateway.AccessWhitelist,
				},
				"unset": &cmd.Alias{
					Exe:         "set",
					Arg:         []string{"%ARG1%", gateway.AccessDefault.String()},
					ArgExpected: 1,
				},
				"ignore": &cmd.Alias{
					Exe:         "set",
					Arg:         []string{"%ARG1%", gateway.AccessIgnore.String()},
					ArgExpected: 1,
				},
				"banlist": &cmd.Alias{
					Exe: "list",
					Arg: []string{gateway.AccessMin.String(), gateway.AccessBan.String()},
				},
				"unignore":  &cmd.Alias{Exe: "unset"},
				"squelch":   &cmd.Alias{Exe: "ignore"},
				"unsquelch": &cmd.Alias{Exe: "unignore"},
				"s":         &cmd.Alias{Exe: "say"},
				"p":         &cmd.Alias{Exe: "pingme"},
				"l":         &cmd.Alias{Exe: "*" + gateway.Delimiter + "list"},
				"i":         &cmd.Alias{Exe: "*" + gateway.Delimiter + "ignore"},
				"w":         &cmd.Alias{Exe: "capi" + gateway.Delimiter + "whisper"},
				"k":         &cmd.Alias{Exe: "capi" + gateway.Delimiter + "kick"},
				"b":         &cmd.Alias{Exe: "capi" + gateway.Delimiter + "ban"},
			},
		},
		Default: gateway.Config{
			Commands: gateway.TriggerConfig{
				Access:  gateway.AccessVoice,
				Trigger: ".",
			},
		},
		StdIO: stdio.Config{
			Read:   true,
			Access: gateway.AccessOwner + 1,
		},
		Capi: CapiConfigWithDefault{
			Default: capi.Config{
				Config: chat.Config{
					Endpoint: pcapi.Endpoint,
				},
				GatewayConfig: capi.GatewayConfig{
					BufSize:        16,
					ReconnectDelay: 30 * time.Second,
					AccessWhisper:  gateway.AccessIgnore,
					AccessTalk:     gateway.AccessVoice,
				},
			},
		},
		BNet: BNetConfigWithDefault{
			Default: bnet.Config{
				GatewayConfig: bnet.GatewayConfig{
					BufSize:        16,
					ReconnectDelay: 30 * time.Second,
					AccessWhisper:  gateway.AccessIgnore,
					AccessTalk:     gateway.AccessVoice,
				},
			},
		},
		Discord: DiscordConfigWithDefault{
			Default: discord.Config{
				Presence: "Battle.net",
				AccessDM: gateway.AccessIgnore,
			},
			ChannelDefault: discord.ChannelConfig{
				BufSize:        64,
				AccessMentions: gateway.AccessWhitelist,
				AccessTalk:     gateway.AccessVoice,
			},
		},
		Relay: RelayConfigWithDefault{
			Default: goop.RelayConfig{
				Say:               true,
				Chat:              true,
				PrivateChat:       true,
				ChatAccess:        gateway.AccessVoice,
				PrivateChatAccess: gateway.AccessVoice,
			},
			DefaultSelf: goop.RelayConfig{
				PrivateChat:       true,
				PrivateChatAccess: gateway.AccessVoice,
			},
			To: map[string]*RelayToConfig{
				"std" + gateway.Delimiter + "io": &RelayToConfig{
					Default: goop.RelayConfig{
						Log:         true,
						System:      true,
						Channel:     true,
						Joins:       true,
						Chat:        true,
						PrivateChat: true,
						Say:         true,
					},
				},
			},
		},
		Plugins: map[string]plugin.Config{},
	}
}

// Config struct maps the layout of main configuration file
type Config struct {
	Hash     string
	Config   string
	Log      LogConfig
	Commands CommandsConfig
	Plugins  PluginsConfig
	Default  gateway.Config
	StdIO    stdio.Config
	Capi     CapiConfigWithDefault
	BNet     BNetConfigWithDefault
	Discord  DiscordConfigWithDefault
	Relay    RelayConfigWithDefault
}

// LogConfig struct maps the layout of the Log configuration section
type LogConfig struct {
	Date         bool
	Time         bool
	Microseconds bool
	UTC          bool
}

// CommandsConfig struct maps the layout of the Commands configuration section
type CommandsConfig struct {
	cmd.Commands
	Alias map[string]*cmd.Alias
}

// PluginsConfig struct maps the layout of the Plugins configuration section
type PluginsConfig map[string]plugin.Config

// CapiConfigWithDefault struct maps the layout of the Capi configuration section
type CapiConfigWithDefault struct {
	Default  capi.Config
	Gateways map[string]*capi.Config
}

// BNetConfigWithDefault struct maps the layout of the BNet configuration section
type BNetConfigWithDefault struct {
	Default  bnet.Config
	Gateways map[string]*bnet.Config
}

// DiscordConfigWithDefault struct maps the layout of the Discord configuration section
type DiscordConfigWithDefault struct {
	Default        discord.Config
	ChannelDefault discord.ChannelConfig
	Gateways       map[string]*discord.Config
}

// RelayConfigWithDefault struct maps the layout of the Relay configuration section
type RelayConfigWithDefault struct {
	Default     goop.RelayConfig
	DefaultSelf goop.RelayConfig
	To          map[string]*RelayToConfig
}

// RelayToConfig struct maps the layout of the inner part of the Relay matrix
type RelayToConfig struct {
	Default goop.RelayConfig
	From    map[string]*goop.RelayConfig
}

// Decode configuration file
func Decode(v interface{}, files ...string) ([]string, error) {
	var m = make(map[string]interface{})
	var u = make([]string, 0)
	for _, f := range files {
		if _, err := toml.DecodeFile(f, &m); err != nil {
			return nil, err
		}
		undec, err := Merge(v, m, &MergeOptions{Overwrite: true})
		if err != nil {
			return nil, err
		}
		u = append(u, undec...)
	}
	return u, nil
}

// Load from c.Config file
func (c *Config) Load() (*Config, error) {
	res, err := c.Copy()
	if err != nil {
		return nil, err
	}
	if _, err := Decode(res, c.Config); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := res.MergeDefaults(); err != nil {
		return nil, err
	}
	return res, nil
}

// Save configuration to c.Config file
func (c *Config) Save(def *Config) error {
	cp, err := def.Copy()
	if err != nil {
		return err
	}
	if err := cp.MergeDefaults(); err != nil {
		return err
	}

	m := c.Map()
	DeleteEqual(m, cp.Map())

	delete(m, "Hash")
	var sha = sha256.Sum256([]byte(fmt.Sprintf("%+v", m)))
	var str = base64.StdEncoding.EncodeToString(sha[:])
	if c.Hash == str {
		return nil
	}
	c.Hash = str
	m["Hash"] = str

	file, err := os.Create(def.Config)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "# Generated at %v\n", time.Now().Format(time.RFC1123))
	return toml.NewEncoder(file).Encode(m)
}

// Copy config
func (c *Config) Copy() (*Config, error) {
	var conf Config
	if _, err := Merge(&conf, c, &MergeOptions{Overwrite: true}); err != nil {
		return nil, err
	}
	return &conf, nil
}

// GetRelay config between to and from
func (c *Config) GetRelay(to, from string) *goop.RelayConfig {
	if c.Relay.To[to] == nil {
		c.Relay.To[to] = &RelayToConfig{
			Default: c.Relay.Default,
		}
	}
	if c.Relay.To[to].From == nil {
		c.Relay.To[to].From = make(map[string]*goop.RelayConfig)
	}
	if c.Relay.To[to].From[from] == nil {
		var cfg = c.Relay.To[to].Default
		if to == from {
			cfg = c.Relay.DefaultSelf
		}
		c.Relay.To[to].From[from] = &cfg
	}
	return c.Relay.To[to].From[from]
}

// MergeDefaults applies default configuration for unset fields
func (c *Config) MergeDefaults() error {
	if _, err := Merge(&c.StdIO.Config, c.Default, &MergeOptions{}); err != nil {
		return err
	}
	if _, err := Merge(&c.Capi.Default.GatewayConfig.Config, c.Default, &MergeOptions{}); err != nil {
		return err
	}
	if _, err := Merge(&c.BNet.Default.GatewayConfig.Config, c.Default, &MergeOptions{}); err != nil {
		return err
	}
	if _, err := Merge(&c.Discord.Default.Config, c.Default, &MergeOptions{}); err != nil {
		return err
	}
	if _, err := Merge(&c.Discord.ChannelDefault.Config, c.Default, &MergeOptions{}); err != nil {
		return err
	}

	for _, r := range c.Capi.Gateways {
		if _, err := Merge(r, c.Capi.Default, &MergeOptions{}); err != nil {
			return err
		}
	}

	for _, r := range c.BNet.Gateways {
		if _, err := Merge(r, c.BNet.Default, &MergeOptions{}); err != nil {
			return err
		}
	}

	for _, g := range c.Discord.Gateways {
		if _, err := Merge(g, c.Discord.Default, &MergeOptions{}); err != nil {
			return err
		}
		for _, n := range g.Channels {
			if _, err := Merge(n, c.Discord.ChannelDefault, &MergeOptions{}); err != nil {
				return err
			}
		}
	}

	for k, p := range c.Plugins {
		if p == nil {
			c.Plugins[k] = make(plugin.Config)
			continue
		}

		var d, ok = p["_default"]
		if !ok {
			continue
		}
		if _, err := Merge(p, d, &MergeOptions{}); err != nil {
			return err
		}
	}

	return nil
}

// type alias for easy type casts
type mi = map[string]interface{}

// Map converts Config to a map[string]interface{} representation
func (c *Config) Map() map[string]interface{} {
	var m = Map(c).(mi)

	var d = m["Default"].(mi)
	DeleteEqual(m["StdIO"].(mi), d)

	var cn = m["Capi"].(mi)["Default"].(mi)
	for _, g := range m["Capi"].(mi)["Gateways"].(mi) {
		DeleteEqual(g.(mi), cn)
	}
	DeleteEqual(cn, d)

	var bn = m["BNet"].(mi)["Default"].(mi)
	for _, g := range m["BNet"].(mi)["Gateways"].(mi) {
		DeleteEqual(g.(mi), bn)
	}
	DeleteEqual(bn, d)

	var dd = m["Discord"].(mi)["Default"].(mi)
	var dc = m["Discord"].(mi)["ChannelDefault"].(mi)
	for _, g := range m["Discord"].(mi)["Gateways"].(mi) {
		for _, c := range g.(mi)["Channels"].(mi) {
			DeleteEqual(c.(mi), dc)
		}
		DeleteEqual(g.(mi), dd)
	}
	DeleteEqual(dc, d)
	DeleteEqual(dd, d)

	var g1d = m["Relay"].(mi)["Default"].(mi)
	var g1s = m["Relay"].(mi)["DefaultSelf"].(mi)
	var gto = m["Relay"].(mi)["To"].(mi)
	for k1, g1 := range gto {
		var g2d = g1.(mi)["Default"].(mi)
		var gfr = g1.(mi)["From"].(mi)
		for k2, g2 := range gfr {
			if k1 == k2 {
				if reflect.DeepEqual(g1s, g2.(mi)) {
					delete(gfr, k2)
				}
				continue
			}
			if reflect.DeepEqual(g2d, g2.(mi)) {
				delete(gfr, k2)
			}
		}
		if reflect.DeepEqual(g1d, g2d) {
			delete(g1.(mi), "Default")
		}
		if len(gfr) == 0 {
			delete(g1.(mi), "From")
		}
		if len(g1.(mi)) == 0 {
			delete(gto, k1)
		}
	}

	return m
}

// FlatMap list all the (nested) config keys
func (c *Config) FlatMap() map[string]interface{} {
	return FlatMap(&c)
}

// Get config value via flat index string
func (c *Config) Get(key string) (interface{}, error) {
	return Get(&c, key)
}

// Set config value via flat index string
func (c *Config) Set(key string, val interface{}) error {
	var conf Config
	if _, err := Merge(&conf, c.Map(), &MergeOptions{Overwrite: true}); err != nil {
		return err
	}
	if err := Set(&conf, key, val); err != nil {
		return err
	}
	if err := conf.MergeDefaults(); err != nil {
		return err
	}
	_, err := Merge(c, conf, &MergeOptions{Overwrite: true, Delete: true})
	return err
}

// Unset config value via flat index string
func (c *Config) Unset(key string) error {
	var conf Config
	if _, err := Merge(&conf, c.Map(), &MergeOptions{Overwrite: true}); err != nil {
		return err
	}
	if err := Unset(&conf, key); err != nil {
		return err
	}
	if err := conf.MergeDefaults(); err != nil {
		return err
	}
	_, err := Merge(c, conf, &MergeOptions{Overwrite: true, Delete: true})
	return err
}

// GetString config value via flat index string
func (c *Config) GetString(key string) (string, error) {
	return GetString(&c, key)
}

// SetString config value via flat index string
func (c *Config) SetString(key string, val string) error {
	var conf Config
	if _, err := Merge(&conf, c.Map(), &MergeOptions{Overwrite: true}); err != nil {
		return err
	}
	if err := SetString(&conf, key, val); err != nil {
		return err
	}
	if err := conf.MergeDefaults(); err != nil {
		return err
	}
	_, err := Merge(c, conf, &MergeOptions{Overwrite: true, Delete: true})
	return err
}

// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/bnet"
	"github.com/nielsAD/goop/gateway/discord"
	"github.com/nielsAD/goop/gateway/stdio"
)

// DefaultConfig values used as fallback
var DefaultConfig = Config{
	Config: "./config.toml.persist",
	Log: LogConfig{
		Time: true,
	},
	StdIO: stdio.Config{
		Read:           true,
		CommandTrigger: "/",
		Access:         gateway.AccessOwner,
	},
	BNet: BNetConfigWithDefault{
		Default: bnet.Config{
			GatewayConfig: bnet.GatewayConfig{
				BufSize:        16,
				ReconnectDelay: 30 * time.Second,
				CommandTrigger: ".",
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
			CommandTrigger: ".",
			AccessMentions: gateway.AccessWhitelist,
			AccessTalk:     gateway.AccessVoice,
		},
	},
	Relay: RelayConfigWithDefault{
		Default: RelayConfig{
			Chat:              true,
			PrivateChat:       true,
			ChatAccess:        gateway.AccessVoice,
			PrivateChatAccess: gateway.AccessWhitelist,
		},
		To: map[string]*RelayToConfig{
			"std" + gateway.Delimiter + "io": &RelayToConfig{
				Default: RelayConfig{
					Log:         true,
					System:      true,
					Joins:       true,
					Chat:        true,
					PrivateChat: true,
				},
			},
		},
	},
}

// Config struct maps the layout of main configuration file
type Config struct {
	Config  string
	Log     LogConfig
	StdIO   stdio.Config
	BNet    BNetConfigWithDefault
	Discord DiscordConfigWithDefault
	Relay   RelayConfigWithDefault
}

// LogConfig struct maps the layout of the Log configuration section
type LogConfig struct {
	Date         bool
	Time         bool
	Microseconds bool
	UTC          bool
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
	Default RelayConfig
	To      map[string]*RelayToConfig
}

// RelayToConfig struct maps the layout of the inner part of the Relay matrix
type RelayToConfig struct {
	Default RelayConfig
	From    map[string]*RelayConfig
}

// LoadConfig from DefaultConfig.Config file
func LoadConfig() (*Config, error) {
	var conf = DefaultConfig
	if _, err := toml.DecodeFile(DefaultConfig.Config, &conf); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := conf.MergeDefaults(); err != nil {
		return nil, err
	}
	return &conf, nil
}

// Save configuration to DefaultConfig.Config file
func (c *Config) Save() error {
	file, err := os.Create(DefaultConfig.Config)
	if err != nil {
		return err
	}
	defer file.Close()

	var m = c.Map()
	DeleteEqual(m, DefaultConfig.Map())

	fmt.Fprintf(file, "# Generated at %v\n", time.Now().Format(time.RFC1123))
	return toml.NewEncoder(file).Encode(m)
}

// MergeDefaults applies default configuration for unset fields
func (c *Config) MergeDefaults() error {
	for _, r := range c.BNet.Gateways {
		if err := mergo.Merge(r, c.BNet.Default); err != nil {
			return err
		}
	}

	for _, g := range c.Discord.Gateways {
		if err := mergo.Merge(g, c.Discord.Default); err != nil {
			return err
		}
		for _, n := range g.Channels {
			if err := mergo.Merge(n, c.Discord.ChannelDefault); err != nil {
				return err
			}
		}
	}

	return nil
}

// type alias for easy type casts
type mi = map[string]interface{}

// Map converts Config to a map[string]interface{} representation
func (c *Config) Map() map[string]interface{} {
	var m = Map(c).(mi)

	var bn = m["BNet"].(mi)["Default"].(mi)
	for _, g := range m["BNet"].(mi)["Gateways"].(mi) {
		DeleteEqual(g.(mi), bn)
	}

	var dd = m["Discord"].(mi)["Default"].(mi)
	var dc = m["Discord"].(mi)["ChannelDefault"].(mi)
	for _, g := range m["Discord"].(mi)["Gateways"].(mi) {
		for _, c := range g.(mi)["Channels"].(mi) {
			DeleteEqual(c.(mi), dc)
		}
		DeleteEqual(g.(mi), dd)
	}

	var g1d = m["Relay"].(mi)["Default"].(mi)
	var gto = m["Relay"].(mi)["To"].(mi)
	for k1, g1 := range gto {
		var g2d = g1.(mi)["Default"].(mi)
		var gfr = g1.(mi)["From"].(mi)
		for k2, g2 := range gfr {
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
	return Set(&c, key, val)
}

// Unset config value via flat index string
func (c *Config) Unset(key string) (err error) {
	return Unset(&c, key)
}

// GetString config value via flat index string
func (c *Config) GetString(key string) (string, error) {
	return GetString(&c, key)
}

// SetString config value via flat index string
func (c *Config) SetString(key string, val string) error {
	return SetString(&c, key, val)
}

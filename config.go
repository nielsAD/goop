// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package main

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/imdario/mergo"

	bnetc "github.com/nielsAD/gowarcraft3/network/bnet"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/bnet"
	"github.com/nielsAD/goop/gateway/discord"
	"github.com/nielsAD/goop/gateway/stdio"
)

// DefaultConfig values used as fallback
var DefaultConfig = Config{
	StdIO: stdio.Config{
		Read:           true,
		Access:         gateway.AccessOwner,
		CommandTrigger: "/",
	},
	BNet: BNetConfigWithDefault{
		Default: bnet.Config{
			GatewayConfig: bnet.GatewayConfig{
				ReconnectDelay: 30 * time.Second,
				CommandTrigger: "!",
				BufSize:        16,
			},
			Config: bnetc.Config{
				BinPath: bnetc.DefaultConfig.BinPath,
			},
		},
	},
	Discord: DiscordConfigWithDefault{
		Default: DefaultDiscordConfig{
			Config: discord.Config{
				Presence:        "Battle.net",
				AccessNoChannel: gateway.AccessIgnore,
				AccessDM:        gateway.AccessWhitelist,
			},
			ChannelConfig: discord.ChannelConfig{
				CommandTrigger: "!",
				BufSize:        32,
				AccessMentions: gateway.AccessWhitelist,
			},
		},
	},
}

// Config struct maps the layout of main configuration file
type Config struct {
	StdIO   stdio.Config
	BNet    BNetConfigWithDefault
	Discord DiscordConfigWithDefault
	Relay   []Relay
}

// BNetConfigWithDefault struct maps the layout of the BNet configuration section
type BNetConfigWithDefault struct {
	Default  bnet.Config
	Gateways map[string]*bnet.Config
}

// DiscordConfigWithDefault struct maps the layout of the Discord configuration section
type DiscordConfigWithDefault struct {
	Default  DefaultDiscordConfig
	Gateways map[string]*discord.Config
}

// DefaultDiscordConfig struct maps the layout of the Discord.Default configuration section
type DefaultDiscordConfig struct {
	discord.Config
	discord.ChannelConfig
}

// Relay struct maps the layout of Relay configuration section
type Relay struct {
	In  []string
	Out []string

	Log         bool
	System      bool
	Joins       bool
	Chat        bool
	PrivateChat bool

	JoinAccess        gateway.AccessLevel
	ChatAccess        gateway.AccessLevel
	PrivateChatAccess gateway.AccessLevel
}

func deleteDefaults(def map[string]interface{}, dst map[string]interface{}) {
	for k := range def {
		if reflect.DeepEqual(def[k], dst[k]) {
			delete(dst, k)
			continue
		}
		var v, ok = dst[k].(map[string]interface{})
		if ok {
			deleteDefaults(def[k].(map[string]interface{}), v)
		}
	}
}

func imap(val interface{}) interface{} {
	var v = reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return imap(v.Elem().Interface())
	case reflect.Map:
		var m = make(map[string]interface{})
		for _, key := range v.MapKeys() {
			m[fmt.Sprintf("%v", key.Interface())] = imap(v.MapIndex(key).Interface())
		}
		return m
	case reflect.Slice, reflect.Array:
		var r = make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			r[i] = imap(v.Index(i).Interface())
		}
		return r
	case reflect.Struct:
		var m = make(map[string]interface{})
		for i := 0; i < v.NumField(); i++ {
			var f = v.Type().Field(i)
			if f.Name == "" {
				continue
			}

			var x = imap(v.Field(i).Interface())
			if xx, ok := x.(map[string]interface{}); f.Anonymous && ok {
				for k, v := range xx {
					m[k] = v
				}
			} else {
				m[f.Name] = x
			}
		}
		return m
	default:
		return v.Interface()
	}
}

func flatten(prf string, val reflect.Value, dst map[string]reflect.Value) {
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			dst[strings.ToLower(prf)] = val
		} else {
			flatten(prf, val.Elem(), dst)
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			var pre string
			if prf == "" {
				pre = fmt.Sprintf("%v", key.Interface())
			} else {
				pre = fmt.Sprintf("%s.%v", prf, key.Interface())
			}
			flatten(pre, val.MapIndex(key), dst)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			var pre string
			if prf == "" {
				pre = fmt.Sprintf("%d", i)
			} else {
				pre = fmt.Sprintf("%s.%d", prf, i)
			}
			flatten(pre, val.Index(i), dst)
		}
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			var f = val.Type().Field(i)
			if f.Name == "" {
				continue
			}

			var pre = f.Name
			if f.Anonymous {
				pre = prf
			} else if prf != "" {
				pre = fmt.Sprintf("%s.%v", prf, f.Name)
			}
			flatten(pre, val.Field(i), dst)
		}
	default:
		dst[strings.ToLower(prf)] = val
	}
}

// MergeDefaults applies default configuration for unset fields
func (c *Config) MergeDefaults() error {
	for _, r := range c.BNet.Gateways {
		if err := mergo.Merge(r, c.BNet.Default); err != nil {
			return err
		}
	}

	for _, g := range c.Discord.Gateways {
		if err := mergo.Merge(g, c.Discord.Default.Config); err != nil {
			return err
		}
		for _, n := range g.Channels {
			if err := mergo.Merge(n, c.Discord.Default.ChannelConfig); err != nil {
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
	var m = imap(c).(mi)

	var bn = m["BNet"].(mi)["Default"].(mi)
	for _, g := range m["BNet"].(mi)["Gateways"].(mi) {
		deleteDefaults(bn, g.(mi))
	}

	var dc = m["Discord"].(mi)["Default"].(mi)
	for _, g := range m["Discord"].(mi)["Gateways"].(mi) {
		for _, c := range g.(mi)["Channels"].(mi) {
			deleteDefaults(dc, c.(mi))
		}
		deleteDefaults(dc, g.(mi))
	}

	return m
}

// Flat list all the (nested) config keys
func (c *Config) Flat() map[string]reflect.Value {
	var f = make(map[string]reflect.Value)
	flatten("", reflect.ValueOf(&c), f)
	return f
}

// Get config value via flat index string
func (c *Config) Get(key string) (interface{}, error) {
	var f, ok = c.Flat()[strings.ToLower(key)]
	if !ok {
		return nil, ErrUnknownConfigKey
	}
	return f.Interface(), nil
}

// Set config value via flat index string
func (c *Config) Set(key string, val interface{}) error {
	var dst, ok = c.Flat()[strings.ToLower(key)]
	if !ok || !dst.CanSet() {
		return ErrUnknownConfigKey
	}

	var src = reflect.ValueOf(val)
	if !src.Type().AssignableTo(dst.Type()) {
		if !src.Type().ConvertibleTo(dst.Type()) {
			return ErrInvalidType
		}
		src = src.Convert(dst.Type())
	}

	dst.Set(src)
	return nil
}

// GetString config value via flat index string
func (c *Config) GetString(key string) (string, error) {
	val, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", val), nil
}

// SetString config value via flat index string
func (c *Config) SetString(key string, val string) error {
	var dst, ok = c.Flat()[strings.ToLower(key)]
	if !ok || !dst.CanSet() {
		return ErrUnknownConfigKey
	}

	if i, ok := dst.Interface().(encoding.TextUnmarshaler); ok {
		return i.UnmarshalText([]byte(val))
	}

	switch dst.Kind() {
	case reflect.String:
		dst.SetString(val)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		dst.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		dst.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		dst.SetUint(n)
		return nil
	default:
		return ErrInvalidType
	}
}

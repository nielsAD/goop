// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package main

import (
	"encoding"
	"errors"
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

// Errors
var (
	ErrUnknownConfigKey = errors.New("goop: Unknown config key")
	ErrTypeMismatch     = errors.New("goop: Type mismatch")
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
		Default: discord.Config{
			Presence:   "Battle.net",
			AccessTalk: gateway.AccessIgnore,
			AccessDM:   gateway.AccessWhitelist,
		},
		ChannelDefault: discord.ChannelConfig{
			CommandTrigger: "!",
			BufSize:        64,
			AccessMentions: gateway.AccessWhitelist,
		},
	},
	Relay: RelayConfigWithDefault{
		Default: RelayConfig{
			Chat:              true,
			PrivateChat:       true,
			PrivateChatAccess: gateway.AccessWhitelist,
		},
		To: map[string]*RelayToConfig{
			"std" + gateway.Delimiter + "io": &RelayToConfig{
				Default: RelayConfig{
					Log:               true,
					System:            true,
					Joins:             true,
					Chat:              true,
					PrivateChat:       true,
					PrivateChatAccess: gateway.AccessDefault,
				},
			},
		},
	},
}

// Config struct maps the layout of main configuration file
type Config struct {
	StdIO   stdio.Config
	BNet    BNetConfigWithDefault
	Discord DiscordConfigWithDefault
	Relay   RelayConfigWithDefault
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

func deleteEmpty(dst map[string]interface{}) {
	var empty = map[string]interface{}{}
	for k := range dst {
		var v, ok = dst[k].(map[string]interface{})
		if !ok {
			continue
		}
		deleteEmpty(v)
		if reflect.DeepEqual(empty, dst[k]) {
			delete(dst, k)
		}
	}
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
		if !val.IsNil() {
			flatten(prf, val.Elem(), dst)
		}
		dst[strings.ToLower(prf)] = val
	case reflect.Map:
		dst[strings.ToLower(prf)] = val
		for _, key := range val.MapKeys() {
			var pre string
			if prf == "" {
				pre = fmt.Sprintf("%v", key.Interface())
			} else {
				pre = fmt.Sprintf("%s.%v", prf, key.Interface())
			}
			flatten(pre, val.MapIndex(key), dst)
		}
	case reflect.Slice:
		dst[strings.ToLower(prf)] = val
		fallthrough
	case reflect.Array:
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

func assign(dst, src reflect.Value) error {
	if !src.Type().AssignableTo(dst.Type()) {
		if src.Type().ConvertibleTo(dst.Type()) {
			src = src.Convert(dst.Type())
		} else if dst.Kind() == reflect.Ptr {
			if dst.IsNil() {
				dst.Set(reflect.New(dst.Type().Elem()))
			}
			return assign(dst.Elem(), src)
		} else {
			return ErrTypeMismatch
		}
	}

	dst.Set(src)
	return nil
}

func assignString(dst reflect.Value, src string) error {
	if i, ok := dst.Interface().(encoding.TextUnmarshaler); ok {
		return i.UnmarshalText([]byte(src))
	}

	switch dst.Kind() {
	case reflect.Ptr:
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return assignString(dst.Elem(), src)
	case reflect.String:
		dst.SetString(src)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(src)
		if err != nil {
			return err
		}
		dst.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(src, 10, 64)
		if err != nil {
			return err
		}
		dst.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(src, 10, 64)
		if err != nil {
			return err
		}
		dst.SetUint(n)
		return nil
	default:
		return ErrTypeMismatch
	}
}

// Parent of key
func Parent(key string) (string, string) {
	if strings.HasSuffix(key, "[]") {
		return key[0 : len(key)-2], "[]"
	}

	var idx = strings.LastIndexByte(key, '.')
	if idx == -1 {
		return "", ""
	}

	return key[0:idx], key[idx+1 : len(key)]
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

	for _, g1 := range c.Relay.To {
		if err := mergo.Merge(&g1.Default, c.Relay.Default); err != nil {
			return err
		}

		for _, g2 := range g1.From {
			if err := mergo.Merge(g2, g1.Default); err != nil {
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

	var dd = m["Discord"].(mi)["Default"].(mi)
	var dc = m["Discord"].(mi)["ChannelDefault"].(mi)
	for _, g := range m["Discord"].(mi)["Gateways"].(mi) {
		for _, c := range g.(mi)["Channels"].(mi) {
			deleteDefaults(dc, c.(mi))
		}
		deleteDefaults(dd, g.(mi))
	}

	var g1d = m["Relay"].(mi)["Default"].(mi)
	var gto = m["Relay"].(mi)["To"].(mi)
	for _, g1 := range gto {
		var g2d = g1.(mi)["Default"].(mi)
		var gfr = g1.(mi)["From"].(mi)
		for _, g2 := range gfr {
			deleteDefaults(g2d, g2.(mi))
		}
		deleteDefaults(g1d, g2d)
	}
	deleteEmpty(m["Relay"].(mi))

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
	var flat = c.Flat()
	var dst, ok = flat[strings.ToLower(key)]
	if ok && dst.CanSet() {
		return assign(dst, reflect.ValueOf(val))
	}

	parent, key := Parent(key)
	if parent == "" || key == "" {
		return ErrUnknownConfigKey
	}

	dst, ok = flat[strings.ToLower(parent)]
	if !ok {
		return ErrUnknownConfigKey
	}

	switch dst.Kind() {
	case reflect.Map:
		if dst.IsNil() {
			dst.Set(reflect.MakeMap(dst.Type()))
		}

		var idx = reflect.New(dst.Type().Key()).Elem()
		if err := assignString(idx, key); err != nil {
			return err
		}

		var tmp = reflect.New(dst.Type().Elem()).Elem()
		if err := assign(tmp, reflect.ValueOf(val)); err != nil {
			return err
		}

		dst.SetMapIndex(idx, tmp)
		return nil
	case reflect.Slice:
		if key != "[]" {
			return ErrUnknownConfigKey
		}

		var tmp = reflect.New(dst.Type().Elem()).Elem()
		if err := assign(tmp, reflect.ValueOf(val)); err != nil {
			return err
		}

		dst.Set(reflect.Append(dst, tmp))
		return nil
	default:
		return ErrUnknownConfigKey
	}
}

// Unset config value via flat index string
func (c *Config) Unset(key string) (err error) {
	var flat = c.Flat()
	var dst, ok = flat[strings.ToLower(key)]
	if !ok {
		return ErrUnknownConfigKey
	}
	if dst.CanSet() {
		err = assign(dst, reflect.Zero(dst.Type()))
	}

	parent, key := Parent(key)
	dst = flat[strings.ToLower(parent)]

	switch dst.Kind() {
	case reflect.Map:
		var idx = reflect.New(dst.Type().Key()).Elem()
		if err := assignString(idx, key); err != nil {
			return err
		}

		dst.SetMapIndex(idx, reflect.Value{})
		return nil
	case reflect.Slice:
		idx, err := strconv.Atoi(key)
		if err != nil {
			return err
		}

		var len = dst.Len()
		dst.Set(reflect.AppendSlice(dst.Slice(0, idx), dst.Slice(idx+1, len)))
		return nil
	default:
		return err
	}
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
	var flat = c.Flat()
	var dst, ok = flat[strings.ToLower(key)]
	if ok && dst.CanSet() {
		return assignString(dst, val)
	}

	parent, key := Parent(key)
	if parent == "" || key == "" {
		return ErrUnknownConfigKey
	}

	dst, ok = flat[strings.ToLower(parent)]
	if !ok {
		return ErrUnknownConfigKey
	}

	switch dst.Kind() {
	case reflect.Map:
		if dst.IsNil() {
			dst.Set(reflect.MakeMap(dst.Type()))
		}

		var idx = reflect.New(dst.Type().Key()).Elem()
		if err := assignString(idx, key); err != nil {
			return err
		}

		var tmp = reflect.New(dst.Type().Elem()).Elem()
		if err := assignString(tmp, val); err != nil {
			return err
		}

		dst.SetMapIndex(idx, tmp)
		return nil
	case reflect.Slice:
		if key != "[]" {
			return ErrUnknownConfigKey
		}

		var tmp = reflect.New(dst.Type().Elem()).Elem()
		if err := assignString(tmp, val); err != nil {
			return err
		}

		dst.Set(reflect.Append(dst, tmp))
		return nil
	default:
		return ErrUnknownConfigKey
	}

}

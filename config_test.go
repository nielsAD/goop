// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package main

import (
	"fmt"
	"reflect"
	"testing"

	bnetc "github.com/nielsAD/gowarcraft3/network/bnet"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/bnet"
	"github.com/nielsAD/goop/gateway/discord"
)

func TestMergeDefaults(t *testing.T) {
	var cfg = Config{
		BNet: BNetConfigWithDefault{
			Default: bnet.Config{
				Config: bnetc.Config{
					Username: "foo",
				},
			},
			Gateways: map[string]*bnet.Config{
				"g1": &bnet.Config{},
				"g2": &bnet.Config{
					Config: bnetc.Config{
						Username: "bar",
					},
				},
			},
		},
		Discord: DiscordConfigWithDefault{
			Default: discord.Config{
				Presence: "foo",
			},
			ChannelDefault: discord.ChannelConfig{
				Webhook: "foo",
			},
			Gateways: map[string]*discord.Config{
				"g1": &discord.Config{},
				"g2": &discord.Config{
					Presence: "bar",
					Channels: map[string]*discord.ChannelConfig{
						"c1": &discord.ChannelConfig{},
						"c2": &discord.ChannelConfig{
							Webhook: "bar",
						},
					},
				},
			},
		},
	}

	if err := cfg.MergeDefaults(); err != nil {
		t.Fatal(err)
	}

	if cfg.BNet.Gateways["g1"].Username != "foo" {
		t.Fatal("Expected username to be foo")
	}
	if cfg.BNet.Gateways["g2"].Username != "bar" {
		t.Fatal("Expected username to be bar")
	}

	if cfg.Discord.Gateways["g1"].Presence != "foo" {
		t.Fatal("Expected presence to be foo")
	}
	if cfg.Discord.Gateways["g2"].Presence != "bar" {
		t.Fatal("Expected presence to be bar")
	}

	if cfg.Discord.Gateways["g2"].Channels["c1"].Webhook != "foo" {
		t.Fatal("Expected webhook to be foo")
	}
	if cfg.Discord.Gateways["g2"].Channels["c2"].Webhook != "bar" {
		t.Fatal("Expected webhook to be bar")
	}
}

func notnil(obj map[string]interface{}, key ...string) bool {
	o, ok := obj[key[0]]
	if !ok || len(key) <= 1 {
		return ok && o != nil
	}
	i, ok := o.(map[string]interface{})
	if !ok {
		return false
	}
	return notnil(i, key[1:]...)
}

func TestMap(t *testing.T) {
	var cfg = Config{
		BNet: BNetConfigWithDefault{
			Default: bnet.Config{
				Config: bnetc.Config{
					Username: "foo",
				},
			},
			Gateways: map[string]*bnet.Config{
				"gw": &bnet.Config{
					Config: bnetc.Config{
						Password: "bar",
					},
				},
			},
		},
	}

	if err := cfg.MergeDefaults(); err != nil {
		t.Fatal(err)
	}

	var m = cfg.Map()
	if !notnil(m, "BNet", "Default", "Username") {
		t.Fatal("BNet.Default.Username does not exist")
	}
	if !notnil(m, "BNet", "Default", "Password") {
		t.Fatal("BNet.Default.Password does not exist")
	}
	if !notnil(m, "BNet", "Gateways", "gw") {
		t.Fatal("BNet.Gateways.gw does not exist")
	}
	if notnil(m, "BNet", "Gateways", "gw", "Username") {
		t.Fatal("BNet.Gateways.gw.Username exists")
	}
	if !notnil(m, "BNet", "Gateways", "gw", "Password") {
		t.Fatal("BNet.Gateways.gw.Password does not exist")
	}
}

func TestFlatMap(t *testing.T) {
	var cfg = Config{
		BNet: BNetConfigWithDefault{
			Default: bnet.Config{
				Config: bnetc.Config{
					Username: "foo",
				},
			},
			Gateways: map[string]*bnet.Config{
				"gw": &bnet.Config{
					Config: bnetc.Config{
						Password: "bar",
					},
				},
			},
		},
	}

	if err := cfg.MergeDefaults(); err != nil {
		t.Fatal(err)
	}

	var m = cfg.FlatMap()
	if m["BNet/Default/Username"] != "foo" {
		t.Fatal("BNet/Default/Username != foo")
	}
	if _, ok := m["BNet/Default/Password"]; !ok {
		t.Fatal("BNet/Default/Password does not exist")
	}
	if m["BNet/Gateways/gw/Username"] != "foo" {
		t.Fatal("BNet/Gateways/gw/Username != foo")
	}
	if m["BNet/Gateways/gw/Password"] != "bar" {
		t.Fatal("BNet/Gateways/gw/Password != bar")
	}
}

func TestGet(t *testing.T) {
	var dc = DefaultConfig()
	if v, _ := dc.Get("StdIO/Access"); v != dc.StdIO.Access {
		t.Fatal("Access different from expected value")
	}
	if v, _ := dc.Get("BNet/Default/BinPath"); v != dc.BNet.Default.BinPath {
		t.Fatal("BinPath different from expected value")
	}

	if v, _ := dc.GetString("StdIO/Access"); v != fmt.Sprintf("%v", dc.StdIO.Access) {
		t.Fatal("Access different from expected value")
	}
	if v, _ := dc.GetString("BNet/Default/BinPath"); v != dc.BNet.Default.BinPath {
		t.Fatal("BinPath different from expected value")
	}
}

func TestError(t *testing.T) {
	var cfg Config
	if _, err := cfg.Get("Foo"); err != ErrUnknownKey {
		t.Fatal("Expected ErrUnknownKey, got", err)
	}
	if err := cfg.Set("Foo", "Bar"); err != ErrUnknownKey {
		t.Fatal("Expected ErrUnknownKey, got", err)
	}
	if err := cfg.SetString("Foo", "Bar"); err != ErrUnknownKey {
		t.Fatal("Expected ErrUnknownKey, got", err)
	}

	if _, err := cfg.Get("BNet/Foo"); err != ErrUnknownKey {
		t.Fatal("Expected ErrUnknownKey, got", err)
	}
	if err := cfg.Set("BNet/Foo", "bar"); err != ErrUnknownKey {
		t.Fatal("Expected ErrUnknownKey, got", err)
	}
	if err := cfg.SetString("BNet/Foo", "bar"); err != ErrUnknownKey {
		t.Fatal("Expected ErrUnknownKey, got", err)
	}

	if _, err := cfg.Get("BNet/Default/CDKeys/99"); err != ErrUnknownKey {
		t.Fatal("Expected ErrUnknownKey, got", err)
	}
	if err := cfg.Set("BNet/Default/CDKeys/99", "xxx"); err != ErrUnknownKey {
		t.Fatal("Expected ErrUnknownKey, got", err)
	}
	if err := cfg.SetString("BNet/Default/CDKeys/99", "xxx"); err != ErrUnknownKey {
		t.Fatal("Expected ErrUnknownKey, got", err)
	}

	if err := cfg.Set("BNet/Default/CDKeys", 123); err != ErrTypeMismatch {
		t.Fatal("Expected ErrTypeMismatch, got", err)
	}
	if err := cfg.SetString("BNet/Default/CDKeys", "123"); err != ErrTypeMismatch {
		t.Fatal("Expected ErrTypeMismatch, got", err)
	}
}

func TestSet(t *testing.T) {
	var cfg Config

	if err := cfg.Set("bNeT/DEFAULT/username", "foo"); err != nil {
		t.Fatal(err)
	}
	if cfg.BNet.Default.Username != "foo" {
		t.Fatal("Expected username to be foo")
	}

	cfg.Unset("bNeT/DEFAULT/username")
	if cfg.BNet.Default.Username != "" {
		t.Fatal("Expected username to be unset")
	}

	if err := cfg.Set("BnEt/default/ACCESSTALK", gateway.AccessWhitelist); err != nil {
		t.Fatal(err)
	}
	if cfg.BNet.Default.AccessTalk != gateway.AccessWhitelist {
		t.Fatal("Expected accesstalk to be wl")
	}

	cfg.Unset("BnEt/default/ACCESSTALK")
	if cfg.BNet.Default.AccessTalk != gateway.AccessDefault {
		t.Fatal("Expected accesstalk to be unset")
	}

	if err := cfg.Set("BNET/default/AccessOperator", gateway.AccessOperator); err != nil {
		t.Fatal(err)
	}
	if cfg.BNet.Default.AccessOperator != gateway.AccessOperator {
		t.Fatal("Expected accessoperator to be op")
	}

	cfg.Unset("BNET/default/AccessOperator")
	if cfg.BNet.Default.AccessOperator != gateway.AccessDefault {
		t.Fatal("Expected accessoperator to be unset")
	}

	if err := cfg.Set("bnet/default/accessuser/niels", gateway.AccessVoice); err != nil {
		t.Fatal(err)
	}
	if cfg.BNet.Default.AccessUser["niels"] != gateway.AccessVoice {
		t.Fatal("Expected accessuser[niels] to be v")
	}

	cfg.Unset("bnet/default/accessuser/niels")
	if _, ok := cfg.BNet.Default.AccessUser["niels"]; ok {
		t.Fatal("Expected accessuser[niels] to be unset")
	}

	cfg.Set("bnet/default/cdkeys[]", "111")
	cfg.Set("bnet/default/cdkeys[]", "333")
	cfg.Set("bnet/default/cdkeys[]", "555")
	cfg.Set("bnet/default/cdkeys[]", "777")
	cfg.Set("bnet/default/cdkeys[]", "999")
	cfg.Set("bnet/default/cdkeys/1", "xxx")
	if !reflect.DeepEqual(cfg.BNet.Default.CDKeys, []string{"111", "xxx", "555", "777", "999"}) {
		t.Fatal("CDKeys(5) mismatch")
	}

	cfg.Unset("bnet/default/cdkeys/4")
	cfg.Unset("bnet/default/cdkeys/2")
	cfg.Unset("bnet/default/cdkeys/0")
	if !reflect.DeepEqual(cfg.BNet.Default.CDKeys, []string{"xxx", "777"}) {
		t.Fatal("CDKeys(2) mismatch")
	}

	cfg.Unset("bnet/default/cdkeys")
	if len(cfg.BNet.Default.CDKeys) != 0 {
		t.Fatal("Expected cdkeys to be unset")
	}
}

func TestSetString(t *testing.T) {
	var cfg Config

	if err := cfg.SetString("bNeT/DEFAULT/username", "foo"); err != nil {
		t.Fatal(err)
	}
	if cfg.BNet.Default.Username != "foo" {
		t.Fatal("Expected username to be foo")
	}

	if err := cfg.SetString("BnEt/default/ACCESSTALK", gateway.AccessWhitelist.String()); err != nil {
		t.Fatal(err)
	}
	if cfg.BNet.Default.AccessTalk != gateway.AccessWhitelist {
		t.Fatal("Expected accesstalk to be wl")
	}

	if err := cfg.SetString("BNET/default/AccessOperator", gateway.AccessOperator.String()); err != nil {
		t.Fatal(err)
	}
	if cfg.BNet.Default.AccessOperator != gateway.AccessOperator {
		t.Fatal("Expected accessoperator to be op")
	}

	if err := cfg.SetString("bnet/default/accessuser/niels", gateway.AccessVoice.String()); err != nil {
		t.Fatal(err)
	}
	if cfg.BNet.Default.AccessUser["niels"] != gateway.AccessVoice {
		t.Fatal("Expected accessuser[niels] to be v")
	}

	cfg.SetString("bnet/default/cdkeys[]", "111")
	cfg.SetString("bnet/default/cdkeys[]", "333")
	cfg.SetString("bnet/default/cdkeys[]", "555")
	cfg.SetString("bnet/default/cdkeys[]", "777")
	cfg.SetString("bnet/default/cdkeys[]", "999")
	cfg.SetString("bnet/default/cdkeys/1", "xxx")
	if !reflect.DeepEqual(cfg.BNet.Default.CDKeys, []string{"111", "xxx", "555", "777", "999"}) {
		t.Fatal("CDKeys(5) mismatch")
	}
}

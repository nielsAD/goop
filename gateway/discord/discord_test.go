// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package discord_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/discord"
	"github.com/nielsAD/gowarcraft3/network"
)

func Test(t *testing.T) {
	d, err := discord.New(&discord.Config{})
	if err != nil {
		t.Fatal(err)
	}
	gw := gateway.Gateway(d)
	gw.On(&network.AsyncError{}, func(ev *network.Event) {
		err := ev.Arg.(*network.AsyncError)
		if !websocket.IsCloseError(err.Err, 4004) {
			t.Fatal(err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if err := gw.Run(ctx); !websocket.IsCloseError(err, 4004) {
		t.Fatal(err)
	}
	cancel()

	for _, e := range gateway.RelayEvents {
		if gw.Relay(&network.Event{Arg: e, Opt: []network.EventArg{gw, "discord" + gateway.Delimiter + "test"}}, gw) == gateway.ErrUnknownEvent {
			t.Fatal(reflect.TypeOf(e))
		}
	}
}

func TestChannel(t *testing.T) {
	s, err := discordgo.New("")
	if err != nil {
		t.Fatal(err)
	}

	c, err := discord.NewChannel(s, &discord.ChannelConfig{})
	if err != nil {
		t.Fatal(err)
	}

	gw := gateway.Gateway(c)
	gw.On(&network.AsyncError{}, func(ev *network.Event) {
		err := ev.Arg.(*network.AsyncError)
		if err, ok := err.Err.(*discordgo.RESTError); ok && err.Response.StatusCode == http.StatusMethodNotAllowed {
			return
		}

		t.Fatal(err)
	})

	for _, e := range gateway.RelayEvents {
		if gw.Relay(&network.Event{Arg: e, Opt: []network.EventArg{gw, "discord_chan" + gateway.Delimiter + "test"}}, gw) == gateway.ErrUnknownEvent {
			t.Fatal(reflect.TypeOf(e))
		}
	}
}

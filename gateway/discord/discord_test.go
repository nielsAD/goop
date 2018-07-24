// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package discord_test

import (
	"context"
	"testing"
	"time"

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
		if err.Err == gateway.ErrUnknownEvent {
			t.Fatal(err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if err := gw.Run(ctx); !websocket.IsCloseError(err, 4004) {
		t.Fatal(err)
	}
	cancel()

	for _, e := range gateway.Events {
		gw.Relay(&network.Event{Arg: e}, "Test")
	}
}

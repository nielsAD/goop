// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package bnet_test

import (
	"context"
	"testing"
	"time"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/bnet"
	"github.com/nielsAD/gowarcraft3/network"
)

func Test(t *testing.T) {
	b, err := bnet.New(&bnet.Config{})
	if err != nil {
		t.Fatal(err)
	}
	gw := gateway.Gateway(b)
	gw.On(&network.AsyncError{}, func(ev *network.Event) {
		err := ev.Arg.(*network.AsyncError)
		if err.Err == gateway.ErrUnknownEvent {
			t.Fatal(err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if err := gw.Run(ctx); !network.IsConnRefusedError(err) {
		t.Fatal(err)
	}
	cancel()

	for _, e := range gateway.Events {
		gw.Relay(&network.Event{Arg: e}, "Test")
	}
}

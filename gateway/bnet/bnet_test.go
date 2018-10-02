// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package bnet_test

import (
	"context"
	"reflect"
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
		if !network.IsConnClosedError(err) {
			t.Fatal(err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if err := gw.Run(ctx); !network.IsConnRefusedError(err) {
		t.Fatal(err)
	}
	cancel()

	for _, e := range gateway.RelayEvents {
		if gw.Relay(&network.Event{Arg: e, Opt: []network.EventArg{gw, "bnet" + gateway.Delimiter + "test"}}, gw) == gateway.ErrUnknownEvent {
			t.Fatal(reflect.TypeOf(e))
		}
	}
}

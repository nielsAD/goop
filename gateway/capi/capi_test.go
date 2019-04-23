// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package capi_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/capi"
	"github.com/nielsAD/gowarcraft3/network"
	"github.com/nielsAD/gowarcraft3/network/chat"
)

func Test(t *testing.T) {
	b, err := capi.New(&capi.Config{Config: chat.Config{Endpoint: "wss://0.0.0.0"}})
	if err != nil {
		t.Fatal(err)
	}

	gw := gateway.Gateway(b)
	gw.On(&network.AsyncError{}, func(ev *network.Event) {
		t.Fatal(ev.Arg.(*network.AsyncError))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if err := gw.Run(ctx); !network.IsRefusedError(err) {
		t.Fatal(err)
	}
	cancel()

	for _, e := range gateway.RelayEvents {
		if gw.Relay(&network.Event{Arg: e, Opt: []network.EventArg{gw, "capi" + gateway.Delimiter + "test"}}, gw) == gateway.ErrUnknownEvent {
			t.Fatal(reflect.TypeOf(e))
		}
	}
}

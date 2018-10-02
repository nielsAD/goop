// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

package stdio_test

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/stdio"
	"github.com/nielsAD/gowarcraft3/network"
)

func Test(t *testing.T) {
	gw := gateway.Gateway(stdio.New(bufio.NewReader(os.Stdin), log.New(os.Stdout, "", 0), &stdio.Config{Read: true}))
	gw.On(&network.AsyncError{}, func(ev *network.Event) {
		err := ev.Arg.(*network.AsyncError)
		t.Fatal(err)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if err := gw.Run(ctx); err != io.EOF {
		t.Fatal(err)
	}
	cancel()

	for _, e := range gateway.RelayEvents {
		if gw.Relay(&network.Event{Arg: e, Opt: []network.EventArg{gw, "std" + gateway.Delimiter + "test"}}, gw) == gateway.ErrUnknownEvent {
			t.Fatal(reflect.TypeOf(e))
		}
	}
}

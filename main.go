// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

// Goop (GO OPerator) is a BNet Channel Operator.
package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	"github.com/nielsAD/goop/gateway"
	"github.com/nielsAD/goop/gateway/bnet"
	"github.com/nielsAD/goop/gateway/capi"
	"github.com/nielsAD/goop/gateway/discord"
	"github.com/nielsAD/goop/gateway/stdio"
	"github.com/nielsAD/goop/goop"
	"github.com/nielsAD/goop/goop/cmd"
	"github.com/nielsAD/gowarcraft3/network"
)

var (
	makeconf = flag.Bool("makeconf", false, "Generate a configuration file")
)

var logOut = log.New(color.Output, "", 0)
var logErr = log.New(color.Error, "", 0)

// New initializes a Goop struct
func New(conf *Config) (*goop.Goop, error) {
	var res = goop.New(conf)

	if err := conf.Commands.AddTo(res); err != nil {
		return nil, err
	}
	for c, a := range conf.Commands.Alias {
		if err := res.AddCommand(c, a); err != nil {
			return nil, err
		}
	}

	if err := res.AddGateway("std"+gateway.Delimiter+"io", stdio.New(bufio.NewReader(os.Stdin), logOut, &conf.StdIO)); err != nil {
		return nil, err
	}

	for k, g := range conf.Capi.Gateways {
		gw, err := capi.New(g)
		if err != nil {
			return nil, err
		}

		if err := res.AddGateway("capi"+gateway.Delimiter+k, gw); err != nil {
			return nil, err
		}
	}

	for k, g := range conf.BNet.Gateways {
		gw, err := bnet.New(g)
		if err != nil {
			return nil, err
		}

		if err := res.AddGateway("bnet"+gateway.Delimiter+k, gw); err != nil {
			return nil, err
		}
	}

	for k, g := range conf.Discord.Gateways {
		gw, err := discord.New(g)
		if err != nil {
			return nil, err
		}

		k = "discord" + gateway.Delimiter + k
		if err := res.AddGateway(k, gw); err != nil {
			return nil, err
		}

		for cid, c := range gw.Channels {
			if err := res.AddGateway(k+gateway.Delimiter+cid, c); err != nil {
				return nil, err
			}
		}
	}

	for g1, r := range conf.Relay.To {
		if res.Gateways[g1] == nil {
			logErr.Println(color.RedString("[ERROR] Unused relay configuration '%s'", g1))
			continue
		}
		for g2 := range r.From {
			if res.Gateways[g2] == nil {
				logErr.Println(color.RedString("[ERROR]Unused relay configuration '%s.%s'", g1, g2))
			}
		}
	}

	return res, nil
}

// Quit program
type Quit struct {
	cmd.Cmd
	cancel context.CancelFunc
}

// Execute command
func (c *Quit) Execute(t *gateway.Trigger, gw gateway.Gateway, g *goop.Goop) error {
	c.cancel()
	return nil
}

func main() {
	runtime.GOMAXPROCS(1)
	flag.Parse()

	var args = flag.Args()
	if len(args) == 0 {
		args = []string{"./config.toml"}
	}

start:
	undecoded, err := Decode(&DefaultConfig, args...)
	if err != nil {
		logErr.Fatalf("Error reading default configuration: %v\n", err)
	}
	if len(undecoded) > 0 {
		logErr.Printf("Undecoded configuration keys: [%s]\n", strings.Join(undecoded, ", "))
	}

	conf, err := Load()
	if err != nil {
		logErr.Fatal("Error reading persistent configuration: ", err)
	}

	var flags = 0
	if conf.Log.Date {
		flags |= log.Ldate
	}
	if conf.Log.Time {
		flags |= log.Ltime
		if conf.Log.Microseconds {
			flags |= log.Lmicroseconds
		}
	}
	if conf.Log.UTC {
		flags |= log.LUTC
	}
	logOut.SetFlags(flags)
	logErr.SetFlags(flags)

	if *makeconf {
		var m = conf.Map()
		if err := toml.NewEncoder(os.Stdout).Encode(m); err != nil {
			logErr.Fatal("Configuration encoding error: ", err)
		}
		return
	}

	g, err := New(conf)
	if err != nil {
		logErr.Fatal("Initialization error: ", err)
	}

	g.On(&gateway.ConfigUpdate{}, func(ev *network.Event) {
		if _, ok := ev.Opt[0].(gateway.Gateway); ok {
			// Assume this error has already been handled
			return
		}
		if err := conf.MergeDefaults(); err != nil {
			g.Fire(&network.AsyncError{Src: "ConfigUpdate", Err: err})
		}
	})

	g.On(&network.AsyncError{}, func(ev *network.Event) {
		var err = ev.Arg.(*network.AsyncError)
		if len(ev.Opt) > 0 {
			if _, ok := ev.Opt[0].(gateway.Gateway); ok {
				// Assume this error has already been handled
				return
			}
		}

		logErr.Println(color.RedString("[ERROR] %s", err.Error()))
	})

	var ctx, cancel = context.WithCancel(context.Background())
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sig
		cancel()
	}()

	g.AddCommand("quit", &Quit{
		Cmd:    cmd.Cmd{Priviledge: gateway.AccessOwner},
		cancel: cancel,
	})

	var restart = false
	g.AddCommand("restart", &Quit{
		Cmd:    cmd.Cmd{Priviledge: gateway.AccessOwner},
		cancel: func() { restart = true; cancel() },
	})

	var done = make(chan struct{})
	go func() {
		for ctx.Err() == nil {
			select {
			case <-time.After(time.Minute * 3):
			case <-ctx.Done():
			}
			if err := conf.Save(); err != nil {
				logErr.Println(color.RedString("[ERROR][CONFIG] %s", err.Error()))
			}
		}
		done <- struct{}{}
	}()

	logOut.Println(color.MagentaString("Starting goop.."))
	g.Run(ctx)
	cancel()

	<-done

	if restart {
		goto start
	}
}

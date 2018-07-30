// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

// Goop (GO OPerator) is a BNet Channel Operator.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"

	"github.com/nielsAD/gowarcraft3/network"
)

var (
	makeconf = flag.Bool("makeconf", false, "Generate a configuration file")
)

var logOut = log.New(color.Output, "", 0)
var logErr = log.New(color.Error, "", 0)

func main() {
	flag.Parse()

	var conf = DefaultConfig
	for _, f := range flag.Args() {
		md, err := toml.DecodeFile(f, &conf)
		if err != nil {
			logErr.Fatal("Error reading configuration: ", err)
		}
		uk := md.Undecoded()
		if len(uk) > 0 {
			logErr.Printf("Undecoded configuration keys: %v\n", uk)
		}
	}

	if err := conf.MergeDefaults(); err != nil {
		logErr.Fatal("Merging defaults error: ", err)
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

	g, err := New(&conf)
	if err != nil {
		logErr.Fatal("Initialization error: ", err)
	}

	if *makeconf {
		var m = conf.Map()
		if err := toml.NewEncoder(os.Stdout).Encode(m); err != nil {
			logErr.Fatal("Configuration encoding error: ", err)
		}
		return
	}

	g.On(&network.AsyncError{}, func(ev *network.Event) {
		var err = ev.Arg.(*network.AsyncError)
		logErr.Println(color.RedString("[ERROR] %s", err.Error()))
	})

	for i, g := range g.Gateways {
		var k = i
		g.On(&network.AsyncError{}, func(ev *network.Event) {
			var err = ev.Arg.(*network.AsyncError)
			logErr.Println(color.RedString("[ERROR][%s] %s", k, err.Error()))
		})
	}

	var ctx, cancel = context.WithCancel(context.Background())
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sig
		cancel()
	}()

	logOut.Println(color.MagentaString("Starting goop.."))
	g.Run(ctx)
}

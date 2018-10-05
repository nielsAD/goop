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

	for _, f := range flag.Args() {
		md, err := toml.DecodeFile(f, &DefaultConfig)
		if err != nil {
			logErr.Fatalf("Error reading default configuration (%v): %v\n", f, err)
		}
		uk := md.Undecoded()
		if len(uk) > 0 {
			logErr.Printf("Undecoded configuration keys in %v: %v\n", f, uk)
		}
	}

	conf, err := LoadConfig()
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

	g.On(&network.AsyncError{}, func(ev *network.Event) {
		var err = ev.Arg.(*network.AsyncError)
		if len(ev.Opt) > 1 {
			return
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

	logOut.Println(color.MagentaString("Starting goop.."))
	g.Run(ctx)

	if err := conf.Save(); err != nil {
		logErr.Println(color.RedString("[ERROR][CONFIG] %s", err.Error()))
	}
}

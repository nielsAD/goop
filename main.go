// Author:  Niels A.D.
// Project: goop (https://github.com/nielsAD/goop)
// License: Mozilla Public License, v2.0

// Goop (GO OPerator) is a classic Battle.net chat bot.
//
// Main package basically loads/updates configuration files and creates a Goop instance.
// Configuration structure works with "Default" structs that are recursively merged into
// the main struct; all zero values are overwritten by defaults.
//
// Order of precedence: config.persist.toml > config.toml > DefaultConfig()
//
// config.toml stores user configuration and is not modified by the application.
// config.persist.toml persists all runtime changes and is managed by the application.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
	"github.com/nielsAD/goop/goop/plugin"
	"github.com/nielsAD/gowarcraft3/network"
)

var (
	makeconf = flag.Bool("makeconf", false, "Generate a configuration file")
)

var logOut = log.New(color.Output, "", 0)
var logErr = log.New(color.Error, "", 0)

const intro = `
   _____                      
  / ____|                     
 | |  __   ___    ___   _ __  
 | | |_ | / _ \  / _ \ | '_ \ 
 | |__| || (_) || (_) || |_) |
  \_____| \___/  \___/ | .__/ 
                       | |    
                       |_|    
`

// Load configuration
func Load(files ...string) (*Config, *Config, error) {

	// User configuration (config.toml)
	var def = DefaultConfig()
	if undecoded, err := Decode(def, files...); err != nil {
		return nil, nil, err
	} else if len(undecoded) > 0 {
		logErr.Printf("Undecoded configuration keys: [%s]\n", strings.Join(undecoded, ", "))
	}

	// Set plugin.Path if empty
	for name, p := range def.Plugins {
		if p.Path != "" {
			continue
		}
		if filepath.Ext(name) == "" {
			name += ".lua"
		}
		if !filepath.IsAbs(name) {
			name = filepath.Join("plugins", name)
		}
		p.Path = name
	}

	// Set discord.Channel.ChannelID if empty
	for _, d := range def.Discord.Gateways {
		for id, c := range d.Channels {
			if c.ChannelID == "" {
				c.ChannelID = id
			}
		}
	}

	// Persistent configuration, i.e. runtime changes (config.persist.toml)
	conf, err := def.Load()
	if err != nil {
		return nil, nil, err
	}

	return def, conf, nil
}

// TODO: Persistent config maybe become out-of-sync with user config after edit, resulting
// in incomplete structs after merge. To minimize the impact of this phenomenon, we check
// if essential fields are set before loading a specific section. Can this be done better?

// New initializes a Goop struct
func New(stdin io.ReadCloser, def *Config, conf *Config) (*goop.Goop, error) {
	var res = goop.New(conf)

	// Global map, shared between plugins
	var globals = map[string]interface{}{
		"BUILD_VERSION": BuildTag,
		"BUILD_COMMIT":  BuildCommit,
		"BUILD_DATE":    BuildDate,
		"GOOS":          runtime.GOOS,
		"GOARCH":        runtime.GOARCH,
		"GOVERSION":     runtime.Version(),
	}

	for k, c := range conf.Plugins {
		if c.Path == "" {
			continue
		}
		p, err := plugin.Load(&c.Config)
		if err != nil {
			return nil, err
		}

		p.SetGlobal("goop", res)
		p.SetGlobal("log", logOut)

		p.SetGlobal("globals", globals)
		p.SetGlobal("options", c.Options)
		p.SetGlobal("defoptions", func(o PluginOptions) {
			def.Plugins[k].DefaultOptions = o
			if err := def.MergeDefaults(); err != nil {
				res.Fire(&network.AsyncError{Src: "defoptions[MergeDefaults(def)]", Err: err})
			}

			conf.Plugins[k].DefaultOptions = o
			if err := conf.MergeDefaults(); err != nil {
				res.Fire(&network.AsyncError{Src: "defoptions[MergeDefaults(conf)]", Err: err})
			}
		})

		if err := p.Run(); err != nil {
			return nil, err
		}

		res.Once(goop.Stop{}, func(_ *network.Event) {
			p.Close()
		})
	}

	if err := conf.Commands.AddTo(res); err != nil {
		return nil, err
	}
	for c, a := range conf.Commands.Alias {
		if err := res.AddCommand(c, a); err != nil {
			return nil, err
		}
	}

	if err := res.AddGateway("std"+gateway.Delimiter+"io", stdio.New(stdin, logOut, &conf.StdIO)); err != nil {
		return nil, err
	}

	for k, g := range conf.Capi.Gateways {
		if g.APIKey == "" {
			logErr.Println(color.RedString("[ERROR] Unused capi configuration '%s'", k))
			continue
		}

		gw, err := capi.New(g)
		if err != nil {
			return nil, err
		}

		if err := res.AddGateway("capi"+gateway.Delimiter+k, gw); err != nil {
			return nil, err
		}
	}

	for k, g := range conf.BNet.Gateways {
		if g.Password == "" {
			logErr.Println(color.RedString("[ERROR] Unused bnet configuration '%s'", k))
			continue
		}

		gw, err := bnet.New(g)
		if err != nil {
			return nil, err
		}

		if err := res.AddGateway("bnet"+gateway.Delimiter+k, gw); err != nil {
			return nil, err
		}
	}

	for k, g := range conf.Discord.Gateways {
		if g.AuthToken == "" {
			logErr.Println(color.RedString("[ERROR] Unused discord configuration '%s'", k))
			continue
		}

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
				logErr.Println(color.RedString("[ERROR] Unused relay configuration '%s.%s'", g1, g2))
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

	var sig = make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	// Prevent closing stdin before restart
	var pw io.Writer = os.Stdout
	go func() {
		var r = bufio.NewReader(os.Stdin)
		for {
			line, err := r.ReadBytes('\n')
			if err != nil {
				break
			}
			pw.Write(line)
		}
	}()

start:
	// User configuration (config.toml)
	def, conf, err := Load(args...)
	if err != nil {
		logErr.Fatalf("Error loading configuration: %v\n", err)
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

	pr, pw := io.Pipe()
	g, err := New(pr, def, conf)
	if err != nil {
		logErr.Fatal("Initialization error: ", err)
	}

	g.On(&gateway.ConfigUpdate{}, func(ev *network.Event) {
		if len(ev.Opt) == 0 {
			// Assume this has already been handled
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

	var restart bool
	var ctx, cancel = context.WithCancel(context.Background())
	go func() {
		select {
		case <-ctx.Done():
		case <-sig:
			restart = false
		}
		cancel()
	}()

	g.AddCommand("version", &cmd.Alias{
		Cmd: cmd.Cmd{Priviledge: gateway.AccessOwner},
		Exe: "echo",
		Arg: []string{fmt.Sprintf("goop %s (%s), %s", BuildTag, BuildCommit[:10], BuildDate.Format("02 January 2006"))},
	})

	g.AddCommand("quit", &Quit{
		Cmd:    cmd.Cmd{Priviledge: gateway.AccessOwner},
		cancel: cancel,
	})

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
			if err := conf.Save(def); err != nil {
				logErr.Println(color.RedString("[ERROR][CONFIG] %s", err.Error()))
			}
		}
		done <- struct{}{}
	}()

	logOut.Println(color.MagentaString(intro))
	logOut.Println(color.MagentaString("Starting goop %s..", BuildTag))
	g.Run(ctx)
	cancel()

	<-done

	if restart {
		goto start
	}
}

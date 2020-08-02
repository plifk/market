package main

import (
	"context"
	_ "expvar"
	"log"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/henvic/clino"
	"github.com/plifk/market/internal/cli"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	state := &cli.State{}
	p := clino.Program{
		Root: &cli.RootCommand{
			State: state,
		},
		GlobalFlags: state.Flags,
	}
	if err := p.Run(context.Background(), os.Args[1:]...); err != nil {
		log.Printf("%+v\n", err)
		os.Exit(clino.ExitCode(err))
	}
}

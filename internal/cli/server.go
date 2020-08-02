package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/henvic/ctxsignal"
	"github.com/plifk/market"
)

type serveCommand struct {
	s *State

	ierror chan struct{} // inspection server failure
	ec     chan error    // market server error
}

func (c *serveCommand) Name() string {
	return "serve"
}

func (c *serveCommand) Short() string {
	return "run the market server"
}

func (c *serveCommand) Run(ctx context.Context, args ...string) (err error) {
	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()
	var system market.System
	if err := system.Load(c.s.ConfigPath); err != nil {
		return err
	}
	modules := system.Modules
	c.startInspectionServer(ctx, cancel, modules.Settings.HTTPInspectionAddress)
	c.startMarketServer(ctx, &system)
	select {
	case err := <-c.ec:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case _, ie := <-c.ierror: // we may reach this code both on inspection server failure and on regular graceful shutdown
		es := <-c.ec // wait for application server to shutdown
		if es != nil && es != http.ErrServerClosed {
			return es
		}
		if ie {
			return errors.New("application server shutdown gracefully after inspection server failure")
		}
		return nil
	}
}

func (c *serveCommand) startMarketServer(ctx context.Context, system *market.System) {
	c.ec = make(chan error, 1)
	go func() {
		defer close(c.ec)
		if err := system.ListenAndServe(ctx); err == http.ErrServerClosed {
			c.ec <- nil
		} else {
			c.ec <- fmt.Errorf("market server error: %w", err)
		}
	}()
}

func (c *serveCommand) startInspectionServer(ctx context.Context, cancel context.CancelFunc, addr string) {
	if addr != "" {
		c.ierror = make(chan struct{}, 1)
		go func() {
			defer close(c.ierror)
			if f := inspectionServer(ctx, addr); f {
				cancel()
				c.ierror <- struct{}{}
			}
		}()
	}
}

func inspectionServer(ctx context.Context, addr string) (failure bool) {
	// let expvar and pprof be exposed here indirectly through http.DefaultServeMux
	// server uses http.DefaultServeMux when the Handler is not set.
	log.Printf("exposing expvar and pprof probes on address %q", addr)
	server := &http.Server{Addr: addr}
	f := make(chan struct{}, 1)
	go func() {
		defer close(f)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("inspection server terminated with an error: %v", err)
			f <- struct{}{}
		}
	}()

	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("inspection server returned error on shutdown: %v", err)
		} else {
			log.Println("graceful shutdown of inspection probes")
		}
		return false // ignore error on shutdown
	case _, failure := <-f: // the inspection server probably terminated abnormally
		return failure
	}
}

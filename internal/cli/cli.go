package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/henvic/clino"
	"github.com/plifk/market/internal/config"
)

// State of the application holds flags that persists between the
// root command and its offspring, plus any other value you would like it to.
type State struct {
	ConfigPath string
	Verbose    bool
}

// Flags implements the clino.FlagSet.
func (s *State) Flags(flags *flag.FlagSet) {
	flags.StringVar(&s.ConfigPath, "config", "market.ini", "configuration path")
	flags.BoolVar(&s.Verbose, "verbose", false, "show more information")
}

// RootCommand for the market server.
type RootCommand struct {
	State *State
}

// Name of the program.
func (c *RootCommand) Name() string {
	return "market"
}

// Commands available to the CLI user.
func (c *RootCommand) Commands() []clino.Command {
	return []clino.Command{
		&serveCommand{
			s: c.State,
		},
		&checkConfigCommand{
			s: c.State,
		},
		&usersCommand{
			s: c.State,
		},
		&tasksCommand{
			s: c.State,
		},
	}
}

type checkConfigCommand struct {
	s *State
}

func (c *checkConfigCommand) Name() string {
	return "check-config"
}

func (c *checkConfigCommand) Short() string {
	return "check and print configuration"
}

func (c *checkConfigCommand) Run(ctx context.Context, args ...string) error {
	settings, err := config.ReadFile(c.s.ConfigPath)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		return fmt.Errorf("cannot represent config with JSON: %w", err)
	}
	fmt.Printf("%s\n", b)
	return nil
}

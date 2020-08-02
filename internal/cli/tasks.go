package cli

import (
	"context"
	"fmt"

	"github.com/henvic/clino"
	"github.com/plifk/market"
)

type tasksCommand struct {
	s *State
}

func (c *tasksCommand) Name() string {
	return "tasks"
}

func (c *tasksCommand) Short() string {
	return "manage tasks"
}

func (c *tasksCommand) Commands() []clino.Command {
	return []clino.Command{
		&cleanupSessionsCommand{s: c.s},
	}
}

type cleanupSessionsCommand struct {
	s *State
}

func (c *cleanupSessionsCommand) Name() string {
	return "cleanup-sessions"
}

func (c *cleanupSessionsCommand) Short() string {
	return "clean up old sessions"
}

func (c *cleanupSessionsCommand) Long() string {
	return `Sessions are marked with the active status on the database for longer than they should.
This status is used mostly to help partitioning the number of active sessions.
This command marks old sessions that already expired as expired.`
}

func (c *cleanupSessionsCommand) Run(ctx context.Context, args ...string) (err error) {
	var system market.System
	if err := system.Load(c.s.ConfigPath); err != nil {
		return err
	}

	modules := system.Modules
	marked, err := modules.Sessions.CloseExpired(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%d expired sessions had their status field changed from \"active\" to \"expired\".\n", marked)
	return nil
}

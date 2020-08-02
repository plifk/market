package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/henvic/clino"
	"github.com/plifk/market"
	"github.com/plifk/market/internal/color"
	"github.com/plifk/market/internal/services"
	"golang.org/x/crypto/ssh/terminal"
)

type usersCommand struct {
	s *State
}

func (c *usersCommand) Name() string {
	return "users"
}

func (c *usersCommand) Short() string {
	return "manage users"
}

func (c *usersCommand) Commands() []clino.Command {
	return []clino.Command{
		&newAdminCommand{s: c.s},
		&setPasswordCommand{s: c.s},
	}
}

type setPasswordCommand struct {
	s *State
}

func (c *setPasswordCommand) Name() string {
	return "set-password"
}

func (c *setPasswordCommand) Short() string {
	return "set password for user"
}

func (c *setPasswordCommand) Run(ctx context.Context, args ...string) (err error) {
	var system market.System
	if err := system.Load(c.s.ConfigPath); err != nil {
		return err
	}

	var s scanner
	fmt.Println(color.Format(color.FgHiCyan, "Set a new password"))
	fmt.Print("User ID or email address: ")
	find, err := s.prompt()
	if err != nil {
		return err
	}

	accounts := system.Modules.Accounts
	var u *services.User
	if strings.Contains(find, "@") {
		u, err = accounts.GetUserByEmail(ctx, find)
	} else {
		u, err = accounts.GetUserByID(ctx, find)
	}

	fmt.Printf(`Name: %v
User ID: %v
Email: %v
Account created: %v
`,
		color.Escape(u.Name),
		color.Escape(u.UserID),
		color.Escape(u.Email),
		u.CreatedAt.Format(time.UnixDate))

	fmt.Print(color.Format(color.FgHiRed, "change password: yes/no? "))
	switch y, err := s.prompt(); {
	case err != nil:
		return err
	case y == "n" || y == "no":
		return nil
	case y == "y" || y == "yes":
		break
	default:
		return errors.New("invalid option")
	}

	fmt.Printf("Password for %v: ", color.Escape(u.Name))
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	fmt.Println("█")

	if err := accounts.SetCredentials(ctx, services.SetPasswordParams{
		UserID:   u.UserID,
		Password: string(password),
	}); err != nil {
		return err
	}
	fmt.Println("Password changed.")
	return nil
}

type newAdminCommand struct {
	s *State
}

func (c *newAdminCommand) Name() string {
	return "new-admin"
}

func (c *newAdminCommand) Short() string {
	return "create admin user"
}

func (c *newAdminCommand) Run(ctx context.Context, args ...string) (err error) {
	var system market.System
	if err := system.Load(c.s.ConfigPath); err != nil {
		return err
	}

	fmt.Printf("Create an admin user (bypassing email address verification).\n%s\n\n",
		color.Format(color.FgHiRed, "DANGER: this tool bypasses security checks. Only use it as a last resource."))

	var p services.NewAdminParams
	var s scanner
	fmt.Print("Name: ")
	if p.Name, err = s.prompt(); err != nil {
		return err
	}
	fmt.Print("Email: ")
	if p.Email, err = s.prompt(); err != nil {
		return err
	}
	fmt.Print("Password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	fmt.Println("█")
	p.Password = string(password)

	accounts := system.Modules.Accounts
	id, err := accounts.NewAdmin(ctx, p)
	if err != nil {
		return err
	}

	fmt.Printf("Admin account created (id: %s)\n", id)
	return nil
}

type scanner struct {
	r *bufio.Scanner
}

func (s *scanner) prompt() (string, error) {
	if s.r == nil {
		s.r = bufio.NewScanner(os.Stdin)
	}
	if s.r.Scan() {
		return s.r.Text(), nil
	}
	return "", s.r.Err()
}

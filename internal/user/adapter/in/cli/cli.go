// Package cli provides the user module's command-line adapter.
package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/rzfhlv/go-std/internal/user/app"
)

// CLI adapts command-line invocations to the user use cases.
type CLI struct {
	users app.UserUseCases
	out   io.Writer
}

// New constructs a CLI whose output is written to out.
func New(users app.UserUseCases, out io.Writer) *CLI {
	return &CLI{users: users, out: out}
}

// Run parses args (excluding the program name) and executes a subcommand.
func (c *CLI) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		c.usage()
		return fmt.Errorf("no command provided")
	}

	command, rest := args[0], args[1:]
	switch command {
	case "create":
		return c.create(ctx, rest)
	case "get":
		return c.get(ctx, rest)
	case "list":
		return c.list(ctx)
	case "count":
		return c.count(ctx)
	case "rename":
		return c.rename(ctx, rest)
	case "delete":
		return c.delete(ctx, rest)
	default:
		c.usage()
		return fmt.Errorf("unknown command %q", command)
	}
}

func (c *CLI) create(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(c.out)
	id := fs.String("id", "", "user id")
	name := fs.String("name", "", "display name")
	email := fs.String("email", "", "email address")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	out, err := c.users.CreateUser(ctx, app.CreateUserInput{ID: *id, Name: *name, Email: *email})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.out, "created user %s (%s)\n", out.ID, out.Email)
	return err
}

func (c *CLI) get(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: get <id>")
	}
	out, err := c.users.GetUser(ctx, args[0])
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.out, "%s\t%s\t%s\n", out.ID, out.Name, out.Email)
	return err
}

func (c *CLI) list(ctx context.Context) error {
	users, err := c.users.ListUsers(ctx)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(c.out, 0, 2, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tNAME\tEMAIL"); err != nil {
		return err
	}
	for _, user := range users {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", user.ID, user.Name, user.Email); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (c *CLI) count(ctx context.Context) error {
	n, err := c.users.CountUsers(ctx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.out, "%d\n", n)
	return err
}

func (c *CLI) rename(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("rename", flag.ContinueOnError)
	fs.SetOutput(c.out)
	id := fs.String("id", "", "user id")
	name := fs.String("name", "", "new display name")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	if *id == "" {
		return fmt.Errorf("usage: rename -id <id> -name <name>")
	}

	out, err := c.users.RenameUser(ctx, *id, *name)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.out, "renamed user %s to %s\n", out.ID, out.Name)
	return err
}

func (c *CLI) delete(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(c.out)
	id := fs.String("id", "", "user id")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	if *id == "" {
		return fmt.Errorf("usage: delete -id <id>")
	}

	if err := c.users.DeleteUser(ctx, *id); err != nil {
		return err
	}
	_, err := fmt.Fprintf(c.out, "deleted user %s\n", *id)
	return err
}

func (c *CLI) usage() {
	_, _ = fmt.Fprint(c.out, `usage: user <command> [flags]

commands:
  create -id <id> -name <name> -email <email>   create a user
  get <id>                                       show a user
  list                                           list all users
  count                                          count users
  rename -id <id> -name <name>                   rename a user
  delete -id <id>                                delete a user
`)
}

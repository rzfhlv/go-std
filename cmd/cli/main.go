// Command cli runs the application's command-line interface, dispatching to the
// selected module by name.
//
// The in-memory repositories are process-local, so each invocation starts empty.
// Swap a module constructor for a persistent adapter to keep data across runs.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rzfhlv/go-std/internal/order"
	"github.com/rzfhlv/go-std/internal/user"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		usage()
		return fmt.Errorf("no module provided")
	}

	module, rest := args[0], args[1:]
	switch module {
	case "user":
		return user.NewInMemory().RunCLI(ctx, os.Stdout, rest)
	case "order":
		return order.NewInMemory(os.Stdout).RunCLI(ctx, os.Stdout, rest)
	default:
		usage()
		return fmt.Errorf("unknown module %q", module)
	}
}

func usage() {
	fmt.Fprint(os.Stdout, `usage: cli <module> <command> [flags]

modules:
  user    manage users
  order   manage orders
`)
}

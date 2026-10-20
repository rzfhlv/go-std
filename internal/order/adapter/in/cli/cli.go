// Package cli provides the order module's command-line adapter.
package cli

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/rzfhlv/go-std/internal/order/app"
)

// CLI adapts command-line invocations to the order use cases.
type CLI struct {
	orders app.OrderUseCases
	out    io.Writer
}

// New constructs a CLI whose output is written to out.
func New(orders app.OrderUseCases, out io.Writer) *CLI {
	return &CLI{orders: orders, out: out}
}

// Run parses args (excluding the program name) and executes a subcommand.
func (c *CLI) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		c.usage()
		return fmt.Errorf("no command provided")
	}

	command, rest := args[0], args[1:]
	switch command {
	case "place":
		return c.place(ctx, rest)
	case "add":
		return c.add(ctx, rest)
	case "confirm":
		return c.confirm(ctx, rest)
	case "cancel":
		return c.cancel(ctx, rest)
	case "get":
		return c.get(ctx, rest)
	case "list":
		return c.list(ctx)
	case "export":
		return c.export(ctx, rest)
	default:
		c.usage()
		return fmt.Errorf("unknown command %q", command)
	}
}

func (c *CLI) place(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("place", flag.ContinueOnError)
	fs.SetOutput(c.out)
	id := fs.String("id", "", "order id")
	customer := fs.String("customer", "", "customer id")
	product := fs.String("product", "", "product id")
	name := fs.String("name", "", "product name")
	price := fs.Int64("price", 0, "unit price in minor units")
	currency := fs.String("currency", "USD", "currency code")
	quantity := fs.Int("qty", 1, "quantity")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	out, err := c.orders.PlaceOrder(ctx, app.PlaceOrderInput{
		OrderID:    *id,
		CustomerID: *customer,
		Lines: []app.LineInput{{
			ProductID: *product,
			Name:      *name,
			UnitPrice: *price,
			Currency:  *currency,
			Quantity:  *quantity,
		}},
	})
	if err != nil {
		return err
	}
	return c.printOrder(out)
}

func (c *CLI) add(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(c.out)
	id := fs.String("id", "", "order id")
	product := fs.String("product", "", "product id")
	name := fs.String("name", "", "product name")
	price := fs.Int64("price", 0, "unit price in minor units")
	currency := fs.String("currency", "USD", "currency code")
	quantity := fs.Int("qty", 1, "quantity")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	out, err := c.orders.AddLine(ctx, *id, app.LineInput{
		ProductID: *product,
		Name:      *name,
		UnitPrice: *price,
		Currency:  *currency,
		Quantity:  *quantity,
	})
	if err != nil {
		return err
	}
	return c.printOrder(out)
}

func (c *CLI) confirm(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: confirm <id>")
	}
	out, err := c.orders.ConfirmOrder(ctx, args[0])
	if err != nil {
		return err
	}
	return c.printOrder(out)
}

func (c *CLI) cancel(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("cancel", flag.ContinueOnError)
	fs.SetOutput(c.out)
	reason := fs.String("reason", "", "cancellation reason")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: cancel <id> [-reason <reason>]")
	}

	out, err := c.orders.CancelOrder(ctx, fs.Arg(0), *reason)
	if err != nil {
		return err
	}
	return c.printOrder(out)
}

func (c *CLI) get(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: get <id>")
	}
	out, err := c.orders.GetOrder(ctx, args[0])
	if err != nil {
		return err
	}
	return c.printOrder(out)
}

func (c *CLI) list(ctx context.Context) error {
	orders, err := c.orders.ListOrders(ctx)
	if err != nil {
		return err
	}

	tw := tabwriter.NewWriter(c.out, 0, 2, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tCUSTOMER\tSTATUS\tSUBTOTAL\tDISCOUNT\tTOTAL\tCURRENCY\tEMPTY"); err != nil {
		return err
	}
	for _, order := range orders {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%d\t%d\t%s\t%s\n",
			order.ID, order.CustomerID, order.Status, order.Subtotal, order.Discount, order.Total, order.Currency, emptyMarker(order)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// emptyMarker reports whether an order should be flagged as empty in list output.
func emptyMarker(order app.OrderOutput) string {
	if len(order.Lines) == 0 {
		return "yes"
	}
	return "no"
}

// export writes every order as CSV to the CLI output.
func (c *CLI) export(ctx context.Context, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: export")
	}
	orders, err := c.orders.ListOrders(ctx)
	if err != nil {
		return err
	}
	return writeCSV(c.out, orders)
}

// writeCSV writes orders as CSV: a header row followed by one row per order.
func writeCSV(w io.Writer, orders []app.OrderOutput) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "customer", "status", "subtotal", "discount", "total", "currency"}); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}
	for _, order := range orders {
		record := []string{
			order.ID,
			order.CustomerID,
			order.Status,
			fmt.Sprintf("%d", order.Subtotal),
			fmt.Sprintf("%d", order.Discount),
			fmt.Sprintf("%d", order.Total),
			order.Currency,
		}
		if err := cw.Write(record); err != nil {
			return fmt.Errorf("write csv record for order %s: %w", order.ID, err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("flush csv: %w", err)
	}
	return nil
}

func (c *CLI) printOrder(order app.OrderOutput) error {
	if _, err := fmt.Fprintf(c.out, "order %s [%s] customer=%s subtotal=%d discount=%d total=%d %s\n",
		order.ID, order.Status, order.CustomerID, order.Subtotal, order.Discount, order.Total, order.Currency); err != nil {
		return err
	}
	for _, line := range order.Lines {
		if _, err := fmt.Fprintf(c.out, "  %s\t%s\tunit=%d\tqty=%d\tsubtotal=%d\n",
			line.ProductID, line.Name, line.UnitPrice, line.Quantity, line.Subtotal); err != nil {
			return err
		}
	}
	return nil
}

func (c *CLI) usage() {
	_, _ = fmt.Fprint(c.out, `usage: order <command> [flags]

commands:
  place   -id <id> -customer <id> -product <p> -name <n> -price <c> -currency <cur> -qty <n>   place an order
  add     -id <id> -product <p> -name <n> -price <c> -currency <cur> -qty <n>                    add a line
  confirm <id>                                                                                  confirm an order
  cancel  <id> [-reason <reason>]                                                               cancel an order
  get     <id>                                                                                  show an order
  list                                                                                          list all orders
  export                                                                                        export all orders as CSV
`)
}

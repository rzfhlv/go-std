// Package order is the public surface of the order module (bounded context).
//
// Everything the module needs internally lives under internal/ and cannot be
// imported by other modules or by cmd — Go's internal/ rule enforces this. Code
// outside the module interacts with it only through Module.
package order

import (
	"context"
	"io"
	"net/http"

	"github.com/rzfhlv/go-std/internal/order/adapter/in/cli"
	"github.com/rzfhlv/go-std/internal/order/adapter/in/grpc"
	"github.com/rzfhlv/go-std/internal/order/adapter/in/rest"
	"github.com/rzfhlv/go-std/internal/order/adapter/out/eventlog"
	"github.com/rzfhlv/go-std/internal/order/adapter/out/memory"
	"github.com/rzfhlv/go-std/internal/order/adapter/out/systemclock"
	"github.com/rzfhlv/go-std/internal/order/app"
)

// Module is the order bounded context wired to its use cases and adapters.
type Module struct {
	useCases app.OrderUseCases
}

// NewInMemory builds the order module backed by an in-memory repository, writing
// domain events to out and reading the time from the system clock. It is
// intended for local runs and tests; add a constructor per real infrastructure
// adapter (for example NewPostgres) as those adapters are built, so cmd never
// needs to import the module's internals.
func NewInMemory(out io.Writer) *Module {
	service := app.NewOrderService(
		memory.NewOrderRepository(),
		eventlog.New(out),
		systemclock.New(),
	)
	return &Module{useCases: service}
}

// RegisterREST mounts the module's REST routes onto mux.
func (m *Module) RegisterREST(mux *http.ServeMux) {
	rest.NewServer(m.useCases).Register(mux)
}

// RunCLI runs the module's command-line adapter with the given arguments.
func (m *Module) RunCLI(ctx context.Context, out io.Writer, args []string) error {
	return cli.New(m.useCases, out).Run(ctx, args)
}

// GRPCService returns the module's gRPC adapter. Register it on a *grpc.Server
// in cmd/grpc once the protobuf definitions and google.golang.org/grpc exist.
func (m *Module) GRPCService() *grpc.OrderServer {
	return grpc.NewOrderServer(m.useCases)
}

// Command grpc runs the application's gRPC server.
//
// The gRPC transport is not implemented yet. To finish it: add a .proto service
// definition, generate Go code, add google.golang.org/grpc, implement the
// generated server in each module's adapter/in/grpc package, then register them
// on a *grpc.Server here using the module facades
// (user.NewInMemory().GRPCService(), order.NewInMemory(os.Stdout).GRPCService()).
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "grpc transport not implemented yet: see cmd/grpc and internal/*/adapter/in/grpc")
	os.Exit(1)
}

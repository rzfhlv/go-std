// Package grpc is the order module's inbound gRPC adapter.
//
// It is a documented placeholder. A full implementation needs a .proto service
// definition, generated Go code, and the google.golang.org/grpc dependency. The
// generated server struct should embed the generated Unimplemented server and
// delegate to the app.OrderUseCases interface, mapping protobuf messages to and
// from the app DTOs at this boundary exactly as the REST and CLI adapters do.
package grpc

import "github.com/rzfhlv/go-std/internal/order/app"

// OrderServer will implement the generated gRPC order service. Register it on a
// *grpc.Server in cmd/grpc once the protobuf code exists.
type OrderServer struct {
	orders app.OrderUseCases
}

// NewOrderServer wires a gRPC order server to the order use cases.
func NewOrderServer(orders app.OrderUseCases) *OrderServer {
	return &OrderServer{orders: orders}
}

// Orders returns the use cases backing this server, for use by generated handler
// methods.
func (s *OrderServer) Orders() app.OrderUseCases {
	return s.orders
}

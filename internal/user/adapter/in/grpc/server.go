// Package grpc is the user module's inbound gRPC adapter.
//
// It is a documented placeholder. A full implementation needs a .proto service
// definition, generated Go code, and the google.golang.org/grpc dependency. The
// generated server struct should embed the generated Unimplemented server and
// delegate to the app.UserUseCases interface, mapping protobuf messages to and
// from the app DTOs at this boundary exactly as the REST and CLI adapters do.
package grpc

import "github.com/rzfhlv/go-std/internal/user/app"

// UserServer will implement the generated gRPC user service. Register it on a
// *grpc.Server in cmd/grpc once the protobuf code exists.
type UserServer struct {
	users app.UserUseCases
}

// NewUserServer wires a gRPC user server to the user use cases.
func NewUserServer(users app.UserUseCases) *UserServer {
	return &UserServer{users: users}
}

// Users returns the use cases backing this server, for use by generated handler
// methods.
func (s *UserServer) Users() app.UserUseCases {
	return s.users
}

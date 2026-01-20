// Package userservice provides a gRPC client for external user service.
package userservice

import (
	"context"

	proto "microservice-template/protocols/userservice"
)

// IUserServiceClient defines the interface for user service client operations.
// This interface works with proto types only and has no dependencies on internal packages.
type IUserServiceClient interface {
	// UserByEmail retrieves a user by email address.
	UserByEmail(ctx context.Context, email string) (*proto.User, error)

	// UserByUUID retrieves a user by UUID bytes.
	UserByUUID(ctx context.Context, uuidBytes []byte) (*proto.User, error)

	// CreateUser creates a new user.
	CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.User, error)

	// Close closes the gRPC connection.
	Close() error
}

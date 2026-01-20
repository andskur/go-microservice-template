// Package mock provides mock implementation of grpcclient.IClient for testing.
package mock

import (
	"context"

	"github.com/gofrs/uuid"

	"microservice-template/internal/grpcclient"
	"microservice-template/internal/models"
)

// GRPCClient is a mock implementation of grpcclient.IClient for testing.
type GRPCClient struct {
	GetUserByEmailFunc func(ctx context.Context, email string) (*models.User, error)
	GetUserByUUIDFunc  func(ctx context.Context, userUUID uuid.UUID) (*models.User, error)
	CreateUserFunc     func(ctx context.Context, user *models.User) (*models.User, error)
}

// Ensure GRPCClient implements grpcclient.IClient interface.
var _ grpcclient.IClient = (*GRPCClient)(nil)

// Name returns the module identifier.
func (m *GRPCClient) Name() string {
	return "mock-grpc-client"
}

// Init is a no-op for mock.
func (m *GRPCClient) Init(ctx context.Context) error {
	if ctx != nil {
		return ctx.Err()
	}
	return nil
}

// Start is a no-op for mock.
func (m *GRPCClient) Start(ctx context.Context) error {
	if ctx != nil {
		return ctx.Err()
	}
	return nil
}

// Stop is a no-op for mock.
func (m *GRPCClient) Stop(ctx context.Context) error {
	if ctx != nil {
		return ctx.Err()
	}
	return nil
}

// HealthCheck is a no-op for mock.
func (m *GRPCClient) HealthCheck(ctx context.Context) error {
	if ctx != nil {
		return ctx.Err()
	}
	return nil
}

// GetUserByEmail calls the mock function if set.
func (m *GRPCClient) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

// GetUserByUUID calls the mock function if set.
func (m *GRPCClient) GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*models.User, error) {
	if m.GetUserByUUIDFunc != nil {
		return m.GetUserByUUIDFunc(ctx, userUUID)
	}
	return nil, nil
}

// CreateUser calls the mock function if set.
func (m *GRPCClient) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	return nil, nil
}

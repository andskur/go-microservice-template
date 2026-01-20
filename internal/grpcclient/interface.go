package grpcclient

import (
	"context"

	"github.com/gofrs/uuid"

	"microservice-template/internal/models"
)

// IClient defines the interface for gRPC client operations.
// This interface allows for easy mocking in tests.
type IClient interface {
	// Module lifecycle methods
	Name() string
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	HealthCheck(ctx context.Context) error

	// User service methods
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
}

// Ensure Module implements IClient interface.
var _ IClient = (*Module)(nil)

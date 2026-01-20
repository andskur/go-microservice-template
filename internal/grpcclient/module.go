// Package grpcclient provides a module wrapper for external gRPC client connections.
package grpcclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/keepalive"

	"microservice-template/config"
	"microservice-template/internal/models"
	"microservice-template/pkg/logger"
	"microservice-template/pkg/userservice"
	proto "microservice-template/protocols/userservice"
)

// Module implements module.Module interface for gRPC client.
type Module struct {
	config *config.GRPCClientConfig
	client userservice.IUserServiceClient
}

// NewModule creates a new gRPC client module.
func NewModule(cfg *config.GRPCClientConfig) *Module {
	return &Module{
		config: cfg,
	}
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "grpc-client"
}

// Init initializes the gRPC client connection.
func (m *Module) Init(ctx context.Context) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	logger.Log().Infof("initializing %s module", m.Name())

	// Parse timeout
	timeout, err := time.ParseDuration(m.config.Timeout)
	if err != nil {
		return fmt.Errorf("parse timeout: %w", err)
	}

	// Parse keep-alive settings
	kaTime, err := time.ParseDuration(m.config.KeepAlive.Time)
	if err != nil {
		return fmt.Errorf("parse keep-alive time: %w", err)
	}

	kaTimeout, err := time.ParseDuration(m.config.KeepAlive.Timeout)
	if err != nil {
		return fmt.Errorf("parse keep-alive timeout: %w", err)
	}

	kacp := keepalive.ClientParameters{
		Time:                kaTime,
		Timeout:             kaTimeout,
		PermitWithoutStream: m.config.KeepAlive.PermitWithoutStream,
	}

	// Create client
	client, err := userservice.New(m.config.Address, timeout, kacp)
	if err != nil {
		return fmt.Errorf("create user service client: %w", err)
	}

	m.client = client

	logger.Log().Infof("%s module initialized (address: %s)", m.Name(), m.config.Address)
	return nil
}

// Start begins module operation (no-op for client).
func (m *Module) Start(ctx context.Context) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	logger.Log().Infof("starting %s module", m.Name())
	// gRPC client doesn't need background workers
	return nil
}

// Stop gracefully shuts down the gRPC client.
func (m *Module) Stop(ctx context.Context) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	logger.Log().Infof("stopping %s module", m.Name())

	if m.client != nil {
		if err := m.client.Close(); err != nil {
			logger.Log().Errorf("close grpc client: %v", err)
			return fmt.Errorf("close client: %w", err)
		}
	}

	logger.Log().Infof("%s module stopped", m.Name())
	return nil
}

// HealthCheck performs a simple connection state check.
func (m *Module) HealthCheck(ctx context.Context) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	// Simple check: if client is initialized, consider healthy
	// Actual RPC calls happen on-demand with their own timeouts
	if m.client == nil {
		return fmt.Errorf("client not initialized")
	}
	return nil
}

// GetUserByEmail retrieves a user from external service by email.
// Converts proto User to domain User.
func (m *Module) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	pbUser, err := m.client.UserByEmail(ctx, email)
	if err != nil {
		return nil, mapError(err)
	}

	return UserFromProto(pbUser)
}

// GetUserByUUID retrieves a user from external service by UUID.
// Converts proto User to domain User.
func (m *Module) GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*models.User, error) {
	pbUser, err := m.client.UserByUUID(ctx, userUUID.Bytes())
	if err != nil {
		return nil, mapError(err)
	}

	return UserFromProto(pbUser)
}

// CreateUser creates a user in external service.
// Converts domain User to proto and back.
func (m *Module) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	// Convert domain user to proto create request
	req := &proto.CreateUserRequest{
		Email: user.Email,
		Name:  user.Name,
	}

	pbUser, err := m.client.CreateUser(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}

	return UserFromProto(pbUser)
}

// mapError maps gRPC client errors to standardized errors.
// Returns error with appropriate categorization for HTTP layer.
func mapError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for gRPC status codes in error message
	if strings.Contains(errStr, "not found") || strings.Contains(errStr, "NotFound") {
		return fmt.Errorf("user not found: %w", err)
	}

	if strings.Contains(errStr, "invalid argument") || strings.Contains(errStr, "InvalidArgument") {
		return fmt.Errorf("invalid input: %w", err)
	}

	if strings.Contains(errStr, "unavailable") || strings.Contains(errStr, "Unavailable") {
		return fmt.Errorf("service unavailable: %w", err)
	}

	if strings.Contains(errStr, "deadline") || strings.Contains(errStr, "DeadlineExceeded") {
		return fmt.Errorf("request timeout: %w", err)
	}

	return err
}

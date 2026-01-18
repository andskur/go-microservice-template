package service

import (
	"context"
	"fmt"

	"microservice-template/internal/repository"
	"microservice-template/pkg/logger"
)

// IUsersService defines the business logic interface for user operations.
// This service orchestrates repository operations and implements business rules.
type IUsersService interface {
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user interface{}) error

	// GetUserByEmail retrieves a user by email address.
	GetUserByEmail(ctx context.Context, email string) (interface{}, error)
}

// UsersService implements IUsersService interface.
type UsersService struct {
	repository repository.IRepository

	// TODO: Add more dependencies as optional when implementing
	// Example:
	// sessions ISessions  // Session management (optional)
	// cache    ICache     // Caching layer (optional)
	// logger   ILogger    // Structured logging (optional)
	//
	// All dependencies should be optional and service should
	// gracefully handle their absence. Check for nil before using.
}

// NewUsersService creates a new users service instance.
// Dependencies:
//   - repository: Required. Handles data persistence.
//
// Future dependencies (add as needed):
//   - sessions: Optional. Session management for auth.
//   - cache: Optional. Caching layer for performance.
//   - events: Optional. Event publishing for async operations.
func NewUsersService(repository repository.IRepository) IUsersService {
	return &UsersService{
		repository: repository,
	}
}

// CreateUser creates a new user in the system.
// TODO: Add validation, business rules, etc.
func (s *UsersService) CreateUser(ctx context.Context, user interface{}) error {
	logger.Log().Info("creating user")

	// TODO: Validate user data
	// TODO: Check for duplicates
	// TODO: Apply business rules

	if err := s.repository.CreateUser(user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	// TODO: Publish user created event (when event system is added)
	// TODO: Send welcome email (when notification system is added)

	return nil
}

// GetUserByEmail retrieves a user by email address.
// TODO: Implement caching when cache module is added.
func (s *UsersService) GetUserByEmail(ctx context.Context, email string) (interface{}, error) {
	logger.Log().Infof("getting user by email: %s", email)

	// TODO: Check cache first (when cache module is added)

	// Create a temporary user object with email
	// In real implementation, this will be *models.User
	user := map[string]interface{}{
		"email": email,
	}

	if err := s.repository.UserBy(user, repository.Email); err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	// TODO: Store in cache (when cache module is added)

	return user, nil
}

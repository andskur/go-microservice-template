// Package auth provides authentication helpers for the HTTP module.
package auth

import (
	"strings"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/gofrs/uuid"

	"microservice-template/internal/http/models"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
)

// Auth handles JWT token validation and user authentication.
type Auth struct {
	service service.IService
	admins  map[string]struct{}
	mocked  bool
}

// NewAuth creates a new Auth instance.
func NewAuth(svc service.IService, mocked bool, adminEmails []string) *Auth {
	adminsMap := make(map[string]struct{})
	for _, email := range adminEmails {
		adminsMap[email] = struct{}{}
	}

	return &Auth{
		service: svc,
		mocked:  mocked,
		admins:  adminsMap,
	}
}

// CheckAuth validates JWT token and returns user principal.
func (a *Auth) CheckAuth(token string) (*models.User, error) {
	if a.mocked {
		logger.Log().Info("using mock authentication (bypassing gatekeeper)")
		return a.mockUser(), nil
	}

	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return nil, errors.New(401, "unauthorized access or invalid credentials")
	}

	logger.Log().Info("checking authorization for token")

	// TODO: Integrate with gatekeeper service for JWT token validation.
	// For local development, set http.mock_auth=true to bypass gatekeeper validation.

	return nil, errors.New(401, "unauthorized access or invalid credentials")
}

// IsAdmin checks if email belongs to admin user.
func (a *Auth) IsAdmin(email string) bool {
	_, ok := a.admins[email]
	return ok
}

// mockUser returns a mock user for testing without gatekeeper.
func (a *Auth) mockUser() *models.User {
	email := strfmt.Email("test@example.com")
	name := "Test User"
	status := "active"

	return &models.User{
		UUID:   strfmt.UUID(uuid.Must(uuid.FromString("FA734DC4-22E6-41C5-A913-30C302C1CA68")).String()),
		Email:  &email,
		Name:   &name,
		Status: &status,
	}
}

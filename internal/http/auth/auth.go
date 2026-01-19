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

// Auth handles JWT token validation and user authentication
type Auth struct {
	service service.IService
	mocked  bool
	admins  map[string]struct{}
}

// NewAuth creates a new Auth instance
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

// CheckAuth validates JWT token and returns user principal
func (a *Auth) CheckAuth(token string) (*models.User, error) {
	if a.mocked {
		logger.Log().Info("using mock authentication (bypassing gatekeeper)")
		return a.mockUser(), nil
	}

	token = strings.TrimPrefix(token, "Bearer ")

	logger.Log().Info("checking authorization for token")

	// TODO: Integrate with gatekeeper service for JWT token validation
	//
	// Gatekeeper is a microservice that handles JWT validation and user authentication.
	// Repository: https://github.com/andskur/gatekeeper
	//
	// Integration steps:
	// 1. Add gatekeeper gRPC client dependency to go.mod:
	//    go get github.com/andskur/gatekeeper (or your gatekeeper client package)
	//
	// 2. Initialize gatekeeper gRPC client in NewAuth() or module initialization:
	//    gatekeeperConn, err := grpc.Dial(
	//        config.HTTP.Gatekeeper.Address,
	//        grpc.WithInsecure(), // Or with TLS creds for production
	//        grpc.WithTimeout(timeout),
	//    )
	//    if err != nil {
	//        return nil, fmt.Errorf("connect to gatekeeper: %w", err)
	//    }
	//    gatekeeperClient := gatekeeper.NewGatekeeperClient(gatekeeperConn)
	//
	// 3. Call gatekeeper ValidateToken in CheckAuth:
	//    timeout, _ := time.ParseDuration(config.HTTP.Gatekeeper.Timeout)
	//    ctx, cancel := context.WithTimeout(context.Background(), timeout)
	//    defer cancel()
	//
	//    resp, err := a.gatekeeperClient.ValidateToken(ctx, &gatekeeper.ValidateTokenRequest{
	//        Token: token,
	//    })
	//    if err != nil {
	//        logger.Log().Errorf("gatekeeper validation failed: %v", err)
	//        return nil, errors.New(401, "unauthorized: invalid or expired token")
	//    }
	//
	// 4. Convert gatekeeper response to API model:
	//    user := &models.User{
	//        UUID:   strfmt.UUID(resp.User.UUID),
	//        Email:  &resp.User.Email,
	//        Name:   &resp.User.Name,
	//        Status: &resp.User.Status,
	//    }
	//    return user, nil
	//
	// 5. Handle gatekeeper errors appropriately:
	//    - Network errors: log and return 503 Service Unavailable
	//    - Invalid token: return 401 Unauthorized
	//    - Expired token: return 401 Unauthorized with specific message
	//
	// 6. Add connection pooling and retry logic for production reliability
	//
	// 7. Consider caching valid tokens (with TTL) to reduce gatekeeper load:
	//    - Use in-memory cache or Redis
	//    - Cache key: hash of token
	//    - Cache value: validated user data
	//    - TTL: shorter than token expiration
	//
	// For local development, set http.mock_auth=true to bypass gatekeeper validation.

	return nil, errors.New(401, "unauthorized access or invalid credentials")
}

// IsAdmin checks if email belongs to admin user
func (a *Auth) IsAdmin(email string) bool {
	_, ok := a.admins[email]
	return ok
}

// mockUser returns a mock user for testing without gatekeeper
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

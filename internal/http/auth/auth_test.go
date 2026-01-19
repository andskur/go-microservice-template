package auth

import (
	"context"
	"testing"

	"microservice-template/internal/models"
)

// mockService is a mock implementation of service.IService for testing
type mockService struct{}

func (m *mockService) CreateUser(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func TestNewAuth(t *testing.T) {
	tests := []struct {
		name        string
		mocked      bool
		adminEmails []string
	}{
		{
			name:        "create auth with no admins",
			mocked:      false,
			adminEmails: []string{},
		},
		{
			name:        "create auth with admins",
			mocked:      false,
			adminEmails: []string{"admin1@example.com", "admin2@example.com"},
		},
		{
			name:        "create auth in mock mode",
			mocked:      true,
			adminEmails: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{}
			auth := NewAuth(svc, tt.mocked, tt.adminEmails)

			if auth == nil {
				t.Fatal("NewAuth returned nil")
			}

			if auth.mocked != tt.mocked {
				t.Errorf("expected mocked=%v, got %v", tt.mocked, auth.mocked)
			}

			if len(auth.admins) != len(tt.adminEmails) {
				t.Errorf("expected %d admins, got %d", len(tt.adminEmails), len(auth.admins))
			}

			for _, email := range tt.adminEmails {
				if _, ok := auth.admins[email]; !ok {
					t.Errorf("admin email %s not found in admins map", email)
				}
			}
		})
	}
}

func TestAuth_CheckAuth_MockMode(t *testing.T) {
	svc := &mockService{}
	auth := NewAuth(svc, true, []string{})

	user, err := auth.CheckAuth("Bearer test-token")
	if err != nil {
		t.Fatalf("CheckAuth failed in mock mode: %v", err)
	}

	if user == nil {
		t.Fatal("CheckAuth returned nil user in mock mode")
	}

	if user.UUID == "" {
		t.Error("mock user has empty UUID")
	}

	if user.Email == nil || *user.Email == "" {
		t.Error("mock user has empty email")
	}

	if user.Name == nil || *user.Name == "" {
		t.Error("mock user has empty name")
	}

	if user.Status == nil || *user.Status == "" {
		t.Error("mock user has empty status")
	}
}

func TestAuth_CheckAuth_NonMockMode(t *testing.T) {
	svc := &mockService{}
	auth := NewAuth(svc, false, []string{})

	// Non-mock mode should return unauthorized (since gatekeeper is not integrated)
	user, err := auth.CheckAuth("Bearer valid-token")
	if err == nil {
		t.Error("expected error in non-mock mode, got nil")
	}

	if user != nil {
		t.Errorf("expected nil user on error, got %v", user)
	}
}

func TestAuth_CheckAuth_TokenPrefixStripping(t *testing.T) {
	svc := &mockService{}
	auth := NewAuth(svc, true, []string{})

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "token with Bearer prefix",
			token: "Bearer test-token",
		},
		{
			name:  "token without Bearer prefix",
			token: "test-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := auth.CheckAuth(tt.token)
			if err != nil {
				t.Fatalf("CheckAuth failed: %v", err)
			}

			if user == nil {
				t.Fatal("CheckAuth returned nil user")
			}
		})
	}
}

func TestAuth_IsAdmin(t *testing.T) {
	svc := &mockService{}
	adminEmails := []string{"admin1@example.com", "admin2@example.com"}
	auth := NewAuth(svc, false, adminEmails)

	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "admin email 1",
			email:    "admin1@example.com",
			expected: true,
		},
		{
			name:     "admin email 2",
			email:    "admin2@example.com",
			expected: true,
		},
		{
			name:     "non-admin email",
			email:    "user@example.com",
			expected: false,
		},
		{
			name:     "empty email",
			email:    "",
			expected: false,
		},
		{
			name:     "case sensitive check",
			email:    "ADMIN1@EXAMPLE.COM",
			expected: false, // emails are case-sensitive in our implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.IsAdmin(tt.email)
			if result != tt.expected {
				t.Errorf("IsAdmin(%s) = %v, expected %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestAuth_MockUser(t *testing.T) {
	svc := &mockService{}
	auth := NewAuth(svc, true, []string{})

	user := auth.mockUser()

	if user == nil {
		t.Fatal("mockUser returned nil")
	}

	// Check UUID format
	if user.UUID == "" {
		t.Error("mock user UUID is empty")
	}

	// Expected UUID (lowercase from gofrs/uuid)
	expectedUUID := "fa734dc4-22e6-41c5-a913-30c302c1ca68"
	if string(user.UUID) != expectedUUID {
		t.Errorf("expected UUID %s, got %s", expectedUUID, user.UUID)
	}

	// Check email
	if user.Email == nil {
		t.Fatal("mock user email is nil")
	}
	expectedEmail := "test@example.com"
	if string(*user.Email) != expectedEmail {
		t.Errorf("expected email %s, got %s", expectedEmail, *user.Email)
	}

	// Check name
	if user.Name == nil {
		t.Fatal("mock user name is nil")
	}
	if *user.Name != "Test User" {
		t.Errorf("expected name 'Test User', got '%s'", *user.Name)
	}

	// Check status
	if user.Status == nil {
		t.Fatal("mock user status is nil")
	}
	if *user.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", *user.Status)
	}
}

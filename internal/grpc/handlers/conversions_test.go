package handlers

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"microservice-template/internal/models"
	proto "microservice-template/protocols/userservice"
)

func TestUserToProto(t *testing.T) {
	t.Parallel()

	testUUID := uuid.Must(uuid.NewV4())
	createdAt := time.Now().UTC().Truncate(time.Second)
	updatedAt := createdAt.Add(time.Hour)

	tests := []struct {
		user     *models.User
		expected *proto.User
		name     string
	}{
		{
			name: "all fields populated",
			user: &models.User{
				UUID:      testUUID,
				Email:     "test@example.com",
				Name:      "Test User",
				Status:    models.UserActive,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			expected: &proto.User{
				Uuid:      testUUID.Bytes(),
				Email:     "test@example.com",
				Name:      "Test User",
				Status:    proto.UserStatus_USER_STATUS_ACTIVE,
				CreatedAt: createdAt.Unix(),
				UpdatedAt: updatedAt.Unix(),
			},
		},
		{
			name: "deleted user",
			user: &models.User{
				UUID:      testUUID,
				Email:     "deleted@example.com",
				Name:      "Deleted User",
				Status:    models.UserDeleted,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			expected: &proto.User{
				Uuid:      testUUID.Bytes(),
				Email:     "deleted@example.com",
				Name:      "Deleted User",
				Status:    proto.UserStatus_USER_STATUS_DELETED,
				CreatedAt: createdAt.Unix(),
				UpdatedAt: updatedAt.Unix(),
			},
		},
		{
			name:     "nil user",
			user:     nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := UserToProto(tt.user)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if result.Email != tt.expected.Email {
				t.Errorf("email: expected %q, got %q", tt.expected.Email, result.Email)
			}

			if result.Name != tt.expected.Name {
				t.Errorf("name: expected %q, got %q", tt.expected.Name, result.Name)
			}

			if result.Status != tt.expected.Status {
				t.Errorf("status: expected %v, got %v", tt.expected.Status, result.Status)
			}

			if result.CreatedAt != tt.expected.CreatedAt {
				t.Errorf("createdAt: expected %d, got %d", tt.expected.CreatedAt, result.CreatedAt)
			}

			if result.UpdatedAt != tt.expected.UpdatedAt {
				t.Errorf("updatedAt: expected %d, got %d", tt.expected.UpdatedAt, result.UpdatedAt)
			}
		})
	}
}

func TestUserStatusToProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   models.UserStatus
		expected proto.UserStatus
	}{
		{
			name:     "active status",
			status:   models.UserActive,
			expected: proto.UserStatus_USER_STATUS_ACTIVE,
		},
		{
			name:     "deleted status",
			status:   models.UserDeleted,
			expected: proto.UserStatus_USER_STATUS_DELETED,
		},
		{
			name:     "invalid status",
			status:   models.UserStatus(999),
			expected: proto.UserStatus_USER_STATUS_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := UserStatusToProto(tt.status)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUserStatusFromProto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expected    models.UserStatus
		status      proto.UserStatus
		expectError bool
	}{
		{
			name:        "active status",
			status:      proto.UserStatus_USER_STATUS_ACTIVE,
			expected:    models.UserActive,
			expectError: false,
		},
		{
			name:        "deleted status",
			status:      proto.UserStatus_USER_STATUS_DELETED,
			expected:    models.UserDeleted,
			expectError: false,
		},
		{
			name:        "unspecified defaults to active",
			status:      proto.UserStatus_USER_STATUS_UNSPECIFIED,
			expected:    models.UserActive,
			expectError: false,
		},
		{
			name:        "unknown status",
			status:      proto.UserStatus(999),
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := UserStatusFromProto(tt.status)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCreateUserRequestToModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		request     *proto.CreateUserRequest
		expected    *models.User
		name        string
		expectError bool
	}{
		{
			name: "valid request with active status",
			request: &proto.CreateUserRequest{
				Email:  "new@example.com",
				Name:   "New User",
				Status: proto.UserStatus_USER_STATUS_ACTIVE,
			},
			expected: &models.User{
				Email:  "new@example.com",
				Name:   "New User",
				Status: models.UserActive,
			},
			expectError: false,
		},
		{
			name: "valid request with deleted status",
			request: &proto.CreateUserRequest{
				Email:  "deleted@example.com",
				Name:   "Deleted User",
				Status: proto.UserStatus_USER_STATUS_DELETED,
			},
			expected: &models.User{
				Email:  "deleted@example.com",
				Name:   "Deleted User",
				Status: models.UserDeleted,
			},
			expectError: false,
		},
		{
			name: "unspecified status defaults to active",
			request: &proto.CreateUserRequest{
				Email:  "test@example.com",
				Name:   "Test User",
				Status: proto.UserStatus_USER_STATUS_UNSPECIFIED,
			},
			expected: &models.User{
				Email:  "test@example.com",
				Name:   "Test User",
				Status: models.UserActive,
			},
			expectError: false,
		},
		{
			name:        "nil request",
			request:     nil,
			expected:    nil,
			expectError: true,
		},
		{
			name: "invalid status",
			request: &proto.CreateUserRequest{
				Email:  "test@example.com",
				Name:   "Test User",
				Status: proto.UserStatus(999),
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := CreateUserRequestToModel(tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if result.Email != tt.expected.Email {
				t.Errorf("email: expected %q, got %q", tt.expected.Email, result.Email)
			}

			if result.Name != tt.expected.Name {
				t.Errorf("name: expected %q, got %q", tt.expected.Name, result.Name)
			}

			if result.Status != tt.expected.Status {
				t.Errorf("status: expected %v, got %v", tt.expected.Status, result.Status)
			}
		})
	}
}

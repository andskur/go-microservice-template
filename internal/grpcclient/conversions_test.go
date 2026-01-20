package grpcclient

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"

	"microservice-template/internal/models"
	proto "microservice-template/protocols/userservice"
)

func TestUserToProto(t *testing.T) {
	now := time.Now()
	userUUID := uuid.Must(uuid.NewV4())

	user := &models.User{
		UUID:      userUUID,
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    models.UserActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	pb := UserToProto(user)

	if pb == nil {
		t.Fatal("expected proto user, got nil")
	}

	if string(pb.Uuid) != string(userUUID.Bytes()) {
		t.Errorf("uuid mismatch")
	}

	if pb.Email != user.Email {
		t.Errorf("email mismatch: got %s, want %s", pb.Email, user.Email)
	}

	if pb.Status != proto.UserStatus_USER_STATUS_ACTIVE {
		t.Errorf("status mismatch: got %v, want %v", pb.Status, proto.UserStatus_USER_STATUS_ACTIVE)
	}
}

func TestUserFromProto(t *testing.T) {
	userUUID := uuid.Must(uuid.NewV4())
	now := time.Now()

	pb := &proto.User{
		Uuid:      userUUID.Bytes(),
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    proto.UserStatus_USER_STATUS_ACTIVE,
		CreatedAt: now.Unix(),
		UpdatedAt: now.Unix(),
	}

	user, err := UserFromProto(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.UUID != userUUID {
		t.Errorf("uuid mismatch")
	}

	if user.Status != models.UserActive {
		t.Errorf("status mismatch: got %v, want %v", user.Status, models.UserActive)
	}
}

func TestUserStatusConversion(t *testing.T) {
	tests := []struct {
		name         string
		domainStatus models.UserStatus
		protoStatus  proto.UserStatus
	}{
		{"active", models.UserActive, proto.UserStatus_USER_STATUS_ACTIVE},
		{"deleted", models.UserDeleted, proto.UserStatus_USER_STATUS_DELETED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Domain -> Proto
			pb := UserStatusToProto(tt.domainStatus)
			if pb != tt.protoStatus {
				t.Errorf("UserStatusToProto: got %v, want %v", pb, tt.protoStatus)
			}

			// Proto -> Domain
			domain, err := UserStatusFromProto(tt.protoStatus)
			if err != nil {
				t.Fatalf("UserStatusFromProto: unexpected error: %v", err)
			}
			if domain != tt.domainStatus {
				t.Errorf("UserStatusFromProto: got %v, want %v", domain, tt.domainStatus)
			}
		})
	}
}

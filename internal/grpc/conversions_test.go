package grpc

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"microservice-template/internal/models"
	userProto "microservice-template/protocols/user"
)

func TestUserToProto_NilUser(t *testing.T) {
	t.Parallel()

	if got := userToProto(nil); got != nil {
		t.Fatalf("expected nil proto, got %#v", got)
	}
}

func TestUserFromProto_NilProto(t *testing.T) {
	t.Parallel()

	if got := userFromProto(nil); got != nil {
		t.Fatalf("expected nil user, got %#v", got)
	}
}

func TestUserStatusToProto_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   models.UserStatus
		out  userProto.UserStatus
	}{
		{name: "active", in: models.UserActive, out: userProto.UserStatus_USER_STATUS_ACTIVE},
		{name: "deleted", in: models.UserDeleted, out: userProto.UserStatus_USER_STATUS_DELETED},
		{name: "unsupported", in: models.UserStatus(99), out: userProto.UserStatus_USER_STATUS_UNSPECIFIED},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := userStatusToProto(tc.in); got != tc.out {
				t.Fatalf("expected %v, got %v", tc.out, got)
			}
		})
	}
}

func TestUserStatusFromProto_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   userProto.UserStatus
		out  models.UserStatus
	}{
		{name: "active", in: userProto.UserStatus_USER_STATUS_ACTIVE, out: models.UserActive},
		{name: "deleted", in: userProto.UserStatus_USER_STATUS_DELETED, out: models.UserDeleted},
		{name: "unspecified", in: userProto.UserStatus_USER_STATUS_UNSPECIFIED, out: models.UserActive},
		{name: "unknown", in: userProto.UserStatus(99), out: models.UserActive},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := userStatusFromProto(tc.in); got != tc.out {
				t.Fatalf("expected %v, got %v", tc.out, got)
			}
		})
	}
}

func TestUserToProto_And_FromProto_RoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	id := uuid.Must(uuid.NewV4())
	user := &models.User{
		UUID:      id,
		Email:     "test@example.com",
		Name:      "Test User",
		Status:    models.UserDeleted,
		CreatedAt: now,
		UpdatedAt: now.Add(time.Minute),
	}

	pb := userToProto(user)
	if pb == nil {
		t.Fatalf("expected proto, got nil")
	}

	if pb.Email != user.Email || pb.Name != user.Name {
		t.Fatalf("proto fields mismatch")
	}

	if pb.Status != userProto.UserStatus_USER_STATUS_DELETED {
		t.Fatalf("expected status deleted, got %v", pb.Status)
	}

	if pb.CreatedAt.AsTime() != user.CreatedAt || pb.UpdatedAt.AsTime() != user.UpdatedAt {
		t.Fatalf("timestamp mismatch")
	}

	if got := userFromProto(pb); got != nil {
		if got.UUID != user.UUID {
			t.Fatalf("uuid mismatch")
		}
		if got.Email != user.Email || got.Name != user.Name {
			t.Fatalf("fields mismatch")
		}
		if got.Status != user.Status {
			t.Fatalf("status mismatch: expected %v, got %v", user.Status, got.Status)
		}
		if !got.CreatedAt.Equal(user.CreatedAt) || !got.UpdatedAt.Equal(user.UpdatedAt) {
			t.Fatalf("time mismatch")
		}
	} else {
		t.Fatalf("expected user, got nil")
	}
}

func TestUserFromProto_InvalidUUID(t *testing.T) {
	t.Parallel()

	pb := &userProto.User{
		Uuid:      []byte{0x01, 0x02}, // invalid length
		Email:     "x@y.z",
		Name:      "X",
		Status:    userProto.UserStatus_USER_STATUS_ACTIVE,
		CreatedAt: timestamppb.New(time.Unix(0, 0)),
		UpdatedAt: timestamppb.New(time.Unix(0, 0)),
	}

	got := userFromProto(pb)
	if got == nil {
		t.Fatalf("expected user, got nil")
	}

	if got.UUID != uuid.Nil {
		t.Fatalf("expected nil UUID, got %v", got.UUID)
	}
}

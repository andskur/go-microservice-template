package grpc

import (
	"context"
	"errors"
	"testing"

	"microservice-template/internal/models"
	userProto "microservice-template/protocols/user"
)

type mockService struct {
	getUserByEmailFunc func(context.Context, string) (*models.User, error)
	createUserFunc     func(context.Context, *models.User) error
}

func (m *mockService) CreateUser(ctx context.Context, user *models.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user)
	}
	return errors.New("not implemented")
}

func (m *mockService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("not implemented")
}

func TestNewUserHandlers_AllowsNilService(t *testing.T) {
	t.Parallel()

	h := NewUserHandlers(nil)
	if h == nil {
		t.Fatalf("expected handler, got nil")
	}
}

func TestUserHandlers_GetUser_Success(t *testing.T) {
	t.Parallel()

	h := NewUserHandlers(&mockService{
		getUserByEmailFunc: func(_ context.Context, email string) (*models.User, error) {
			return &models.User{
				Email:  email,
				Name:   "Tester",
				Status: models.UserActive,
			}, nil
		},
	})

	resp, err := h.GetUser(context.Background(), &userProto.GetUserRequest{Email: "test@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Email != "test@example.com" || resp.Name != "Tester" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestUserHandlers_GetUser_EmptyEmail(t *testing.T) {
	t.Parallel()

	h := NewUserHandlers(&mockService{})

	_, err := h.GetUser(context.Background(), &userProto.GetUserRequest{Email: ""})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestUserHandlers_GetUser_ServiceError(t *testing.T) {
	t.Parallel()

	h := NewUserHandlers(&mockService{
		getUserByEmailFunc: func(_ context.Context, _ string) (*models.User, error) {
			return nil, errors.New("boom")
		},
	})

	_, err := h.GetUser(context.Background(), &userProto.GetUserRequest{Email: "x@y.z"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestUserHandlers_CreateUser_Success(t *testing.T) {
	t.Parallel()

	var created models.User

	h := NewUserHandlers(&mockService{
		createUserFunc: func(_ context.Context, user *models.User) error {
			created = *user
			return nil
		},
	})

	resp, err := h.CreateUser(context.Background(), &userProto.CreateUserRequest{
		Email: "test@example.com",
		Name:  "Tester",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.Email != "test@example.com" || created.Name != "Tester" || created.Status != models.UserActive {
		t.Fatalf("unexpected created user: %+v", created)
	}

	if resp.Email != created.Email || resp.Name != created.Name {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestUserHandlers_CreateUser_EmptyEmail(t *testing.T) {
	t.Parallel()

	h := NewUserHandlers(&mockService{})

	_, err := h.CreateUser(context.Background(), &userProto.CreateUserRequest{Email: "", Name: "Tester"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestUserHandlers_CreateUser_EmptyName(t *testing.T) {
	t.Parallel()

	h := NewUserHandlers(&mockService{})

	_, err := h.CreateUser(context.Background(), &userProto.CreateUserRequest{Email: "x@y.z", Name: ""})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestUserHandlers_CreateUser_ServiceError(t *testing.T) {
	t.Parallel()

	h := NewUserHandlers(&mockService{
		createUserFunc: func(_ context.Context, _ *models.User) error {
			return errors.New("boom")
		},
	})

	_, err := h.CreateUser(context.Background(), &userProto.CreateUserRequest{Email: "x@y.z", Name: "Tester"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

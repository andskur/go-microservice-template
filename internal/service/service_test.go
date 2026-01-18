package service

import (
	"context"
	"testing"

	"microservice-template/internal/repository"
)

// mockRepository is a simple in-memory repository for testing purposes.
type mockRepository struct{}

func (m *mockRepository) CreateUser(model interface{}) error {
	return nil
}

func (m *mockRepository) UserBy(model interface{}, getter repository.UserGetter) error {
	return nil
}

func TestService_CreateUser(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)

	ctx := context.Background()
	user := map[string]interface{}{
		"email": "test@example.com",
	}

	if err := svc.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
}

func TestService_GetUserByEmail(t *testing.T) {
	repo := &mockRepository{}
	svc := NewService(repo)

	ctx := context.Background()

	user, err := svc.GetUserByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail returned error: %v", err)
	}

	if user == nil {
		t.Fatal("expected non-nil user")
	}
}

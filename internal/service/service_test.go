package service

import (
	"context"
	"strings"
	"testing"

	"microservice-template/internal/repository"
)

// serviceMockRepository is a simple in-memory repository for testing purposes.
type serviceMockRepository struct{}

func (m *serviceMockRepository) CreateUser(model interface{}) error {
	return nil
}

func (m *serviceMockRepository) UserBy(model interface{}, getter repository.UserGetter) error {
	return nil
}

func TestService_CreateUser(t *testing.T) {
	repo := &serviceMockRepository{}
	svc := NewService(repo)

	ctx := context.Background()
	user := map[string]interface{}{
		"email": "test@example.com",
	}

	if err := svc.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
}

func TestService_CreateUser_NoRepository(t *testing.T) {
	svc := NewService(nil)

	ctx := context.Background()
	user := map[string]interface{}{
		"email": "test@example.com",
	}

	err := svc.CreateUser(ctx, user)
	if err == nil {
		t.Fatal("expected error when repository is nil, got nil")
	}

	expected := "repository not available"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %v", expected, err)
	}
}

func TestService_GetUserByEmail(t *testing.T) {
	repo := &serviceMockRepository{}
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

func TestService_GetUserByEmail_NoRepository(t *testing.T) {
	svc := NewService(nil)

	ctx := context.Background()

	user, err := svc.GetUserByEmail(ctx, "test@example.com")
	if err == nil {
		t.Fatal("expected error when repository is nil, got nil")
	}

	if user != nil {
		t.Fatalf("expected nil user on error, got %v", user)
	}

	expected := "repository not available"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error containing %q, got %v", expected, err)
	}
}

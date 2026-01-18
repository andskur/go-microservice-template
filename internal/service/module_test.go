package service

import (
	"context"
	"testing"

	"microservice-template/internal/repository"
)

// moduleMockRepository is a simple in-memory repository for testing purposes.
type moduleMockRepository struct{}

func (m *moduleMockRepository) CreateUser(model interface{}) error {
	return nil
}

func (m *moduleMockRepository) UserBy(model interface{}, getter repository.UserGetter) error {
	return nil
}

func TestModule_Lifecycle_WithRepository(t *testing.T) {
	ctx := context.Background()

	repo := &moduleMockRepository{}
	mod := NewModule(repo)

	if err := mod.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if mod.Service() == nil {
		t.Fatal("Service() returned nil")
	}

	if err := mod.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := mod.HealthCheck(ctx); err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	if err := mod.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestModule_Lifecycle_WithoutRepository(t *testing.T) {
	ctx := context.Background()

	mod := NewModule(nil)

	if err := mod.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if mod.Service() == nil {
		t.Fatal("Service() returned nil without repository")
	}
}

func TestModule_Name(t *testing.T) {
	mod := NewModule(nil)
	if got := mod.Name(); got != "service" {
		t.Errorf("Name() = %q, want %q", got, "service")
	}
}

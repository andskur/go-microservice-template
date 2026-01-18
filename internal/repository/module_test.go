package repository

import (
	"context"
	"testing"

	"microservice-template/config"
	"microservice-template/internal/models"
)

// moduleMockRepository is a simple in-memory repository for testing purposes.
type moduleMockRepository struct{}

func (m *moduleMockRepository) CreateUser(user *models.User) error                { return nil }
func (m *moduleMockRepository) UserBy(user *models.User, getter UserGetter) error { return nil }

func TestRepositoryModule_Lifecycle(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Enabled: true,
		Driver:  "postgres",
		Host:    "localhost",
		Port:    5432,
	}

	mod := NewModule(cfg)
	ctx := context.Background()

	// Test Name
	if mod.Name() != "repository" {
		t.Errorf("expected name 'repository', got '%s'", mod.Name())
	}

	// Test Init
	if err := mod.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test Start
	if err := mod.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Test Repository accessor
	repo := mod.Repository()
	if repo == nil {
		t.Error("Repository() should return non-nil repository")
	}

	// Test HealthCheck
	if err := mod.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	// Test Stop
	if err := mod.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestRepositoryModule_Repository(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Enabled: true,
		Driver:  "postgres",
	}

	mod := NewModule(cfg)
	if err := mod.Init(context.Background()); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	repo := mod.Repository()

	// Repository should be accessible
	if repo == nil {
		t.Fatal("Repository should not be nil")
	}

	// Should return PostgresRepository (stub)
	if _, ok := repo.(*PostgresRepository); !ok {
		t.Error("Repository should be PostgresRepository instance")
	}
}

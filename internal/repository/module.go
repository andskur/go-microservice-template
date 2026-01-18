package repository

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v10"

	"microservice-template/config"
	"microservice-template/pkg/logger"
)

// Module implements module.Module interface for repository layer.
// It manages database connection lifecycle and provides repository instances.
type Module struct {
	config *config.DatabaseConfig
	db     *pg.DB
	repo   IRepository
}

// NewModule creates a new repository module instance.
func NewModule(cfg *config.DatabaseConfig) *Module {
	return &Module{
		config: cfg,
	}
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "repository"
}

// Init initializes the repository module and establishes database connection.
func (m *Module) Init(_ context.Context) error {
	logger.Log().Infof("initializing %s module with driver: %s", m.Name(), m.config.Driver)

	db := pg.Connect(&pg.Options{
		Addr:         fmt.Sprintf("%s:%d", m.config.Host, m.config.Port),
		User:         m.config.User,
		Password:     m.config.Password,
		Database:     m.config.Name,
		PoolSize:     m.config.MaxOpenConns,
		MinIdleConns: m.config.MaxIdleConns,
	})

	m.db = db
	m.repo = NewPostgresRepository(db)

	logger.Log().Infof("%s module initialized successfully", m.Name())
	return nil
}

// Start begins module operation (no-op for repository).
func (m *Module) Start(_ context.Context) error {
	logger.Log().Infof("starting %s module", m.Name())
	return nil
}

// Stop gracefully shuts down the module and closes database connection.
func (m *Module) Stop(_ context.Context) error {
	logger.Log().Infof("stopping %s module", m.Name())

	if m.db != nil {
		if err := m.db.Close(); err != nil {
			return fmt.Errorf("close database connection: %w", err)
		}
		logger.Log().Info("database connection closed")
	}

	return nil
}

// HealthCheck verifies database connectivity.
func (m *Module) HealthCheck(ctx context.Context) error {
	if m.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if _, err := m.db.WithContext(ctx).Exec("SELECT 1"); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Repository returns the repository instance.
// This is used by other parts of the application (e.g., Service layer).
func (m *Module) Repository() IRepository {
	return m.repo
}

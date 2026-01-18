package repository

import (
	"context"

	"microservice-template/config"
	"microservice-template/pkg/logger"
)

// Module implements module.Module interface for repository layer.
// It manages database connection lifecycle and provides repository instances.
type Module struct {
	config *config.DatabaseConfig
	db     interface{} // TODO: Change to *pg.DB when go-pg is added
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
//
// TODO: Add database connection when go-pg is added:
//
//	db := pg.Connect(&pg.Options{
//	    Addr:     fmt.Sprintf("%s:%d", m.config.Host, m.config.Port),
//	    User:     m.config.User,
//	    Password: m.config.Password,
//	    Database: m.config.Name,
//	})
//	m.db = db
func (m *Module) Init(_ context.Context) error {
	logger.Log().Infof("initializing %s module with driver: %s", m.Name(), m.config.Driver)

	// TODO: Connect to database
	// For now, create PostgreSQL repository with nil db
	m.db = nil
	m.repo = NewPostgresRepository(m.db)

	logger.Log().Infof("%s module initialized (database connection pending)", m.Name())
	return nil
}

// Start begins module operation (no-op for repository).
func (m *Module) Start(_ context.Context) error {
	logger.Log().Infof("starting %s module", m.Name())
	// Repository is passive, nothing to start
	return nil
}

// Stop gracefully shuts down the module and closes database connection.
//
// TODO: Close database connection when go-pg is added:
//
//	if m.db != nil {
//	    return m.db.(*pg.DB).Close()
//	}
func (m *Module) Stop(_ context.Context) error {
	logger.Log().Infof("stopping %s module", m.Name())

	// TODO: Close database connection
	return nil
}

// HealthCheck verifies database connectivity.
//
// TODO: Implement database ping when go-pg is added:
//
//	if m.db != nil {
//	    _, err := m.db.(*pg.DB).Exec("SELECT 1")
//	    return err
//	}
func (m *Module) HealthCheck(_ context.Context) error {
	// TODO: Implement database ping
	return nil
}

// Repository returns the repository instance.
// This is used by other parts of the application (e.g., Service layer).
func (m *Module) Repository() IRepository {
	return m.repo
}

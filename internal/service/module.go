// Package service contains business logic layer implementations.
package service

import (
	"context"

	"microservice-template/internal/repository"
	"microservice-template/pkg/logger"
)

// RepositoryProvider provides access to a repository instance.
// This allows the repository to be retrieved after it's been initialized.
type RepositoryProvider interface {
	Repository() repository.IRepository
}

// Module implements module.Module interface for the service layer.
// It wires the business logic service with optional dependencies such as repository.
type Module struct {
	repoProvider RepositoryProvider
	service      IService
}

// NewModule creates a new service module instance.
// repoProvider can be nil when the database module is not enabled.
func NewModule(repoProvider RepositoryProvider) *Module {
	return &Module{
		repoProvider: repoProvider,
	}
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "service"
}

// Init initializes the service module.
func (m *Module) Init(_ context.Context) error {
	logger.Log().Infof("initializing %s module", m.Name())

	// Retrieve repository from provider (after repository module has been initialized)
	var repo repository.IRepository
	if m.repoProvider != nil {
		repo = m.repoProvider.Repository()
	}

	m.service = NewService(repo)

	if repo == nil {
		logger.Log().Warn("service initialized without repository; database operations will be unavailable")
	} else {
		logger.Log().Info("service initialized with repository")
	}

	return nil
}

// Start begins module operation (no-op for service).
func (m *Module) Start(_ context.Context) error {
	logger.Log().Infof("starting %s module", m.Name())
	return nil
}

// Stop gracefully shuts down the module (no-op currently).
func (m *Module) Stop(_ context.Context) error {
	logger.Log().Infof("stopping %s module", m.Name())
	return nil
}

// HealthCheck verifies service health.
func (m *Module) HealthCheck(_ context.Context) error {
	return nil
}

// Service returns the business logic service instance.
func (m *Module) Service() IService {
	return m.service
}

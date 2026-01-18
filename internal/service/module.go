package service

import (
	"context"

	"microservice-template/internal/repository"
	"microservice-template/pkg/logger"
)

// Module implements module.Module interface for the service layer.
// It wires the business logic service with optional dependencies such as repository.
type Module struct {
	repository repository.IRepository
	service    IService
}

// NewModule creates a new service module instance.
// repository can be nil when the database module is not enabled.
func NewModule(repo repository.IRepository) *Module {
	return &Module{
		repository: repo,
	}
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "service"
}

// Init initializes the service module.
func (m *Module) Init(_ context.Context) error {
	logger.Log().Infof("initializing %s module", m.Name())

	m.service = NewService(m.repository)

	if m.repository == nil {
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

package grpc

import (
	"context"
	"fmt"

	"microservice-template/config"
	"microservice-template/internal/grpc/handlers"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
	proto "microservice-template/protocols/userservice"
)

// Module implements module.Module interface for gRPC server.
type Module struct {
	config  *config.GRPCConfig
	service service.IService
	server  *Server
}

// NewModule creates a new gRPC module instance.
func NewModule(cfg *config.GRPCConfig, svc service.IService) *Module {
	return &Module{
		config:  cfg,
		service: svc,
	}
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "grpc"
}

// Init initializes the gRPC module.
func (m *Module) Init(_ context.Context) error {
	logger.Log().Infof("initializing %s module on %s:%d", m.Name(), m.config.Host, m.config.Port)

	server, err := NewServer(m.config)
	if err != nil {
		return fmt.Errorf("create grpc server: %w", err)
	}
	m.server = server

	if err := m.server.RegisterHealthService(); err != nil {
		return fmt.Errorf("register health service: %w", err)
	}

	if err := m.registerHandlers(); err != nil {
		return fmt.Errorf("register handlers: %w", err)
	}

	logger.Log().Infof("%s module initialized successfully", m.Name())
	return nil
}

// Start begins gRPC server operation (non-blocking).
func (m *Module) Start(_ context.Context) error {
	m.server.MarkRunning()

	go func() {
		if err := m.server.Serve(); err != nil {
			logger.Log().Errorf("grpc server error: %v", err)
		}
	}()

	logger.Log().Infof("grpc server listening on %s:%d", m.config.Host, m.config.Port)
	return nil
}

// Stop gracefully shuts down the gRPC server.
func (m *Module) Stop(_ context.Context) error {
	logger.Log().Infof("stopping %s module", m.Name())

	if m.server != nil {
		m.server.GracefulStop()
		logger.Log().Info("grpc server stopped gracefully")
	}

	return nil
}

// HealthCheck verifies gRPC server health.
func (m *Module) HealthCheck(_ context.Context) error {
	if m.server == nil {
		return fmt.Errorf("grpc server not initialized")
	}

	if !m.server.IsRunning() {
		return fmt.Errorf("grpc server not running")
	}

	return nil
}

// registerHandlers registers all gRPC service handlers with the server.
func (m *Module) registerHandlers() error {
	// Create and register UserService handler
	// logger.Log().Logger accesses the embedded *logrus.Logger
	userServiceHandler := handlers.NewUserServiceHandler(m.service, logger.Log().Logger)
	proto.RegisterUserServiceServer(m.server.Server(), userServiceHandler)

	logger.Log().Info("grpc handlers registered successfully")
	return nil
}

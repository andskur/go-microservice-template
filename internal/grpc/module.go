package grpc

import (
	"context"
	"fmt"

	proto "microservice-template/api/proto/userservice"
	"microservice-template/config"
	"microservice-template/internal/grpc/handlers"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
)

// Module implements module.Module interface for gRPC server.
type Module struct {
	config  *config.GRPCConfig
	service service.IService
	server  *Server
	failed  chan error
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
	m.failed = make(chan error, 1)

	go func() {
		defer close(m.failed)
		// Serve returns nil after GracefulStop/Stop; anything else is fatal.
		if err := m.server.Serve(); err != nil {
			m.failed <- err
		}
	}()

	logger.Log().Infof("grpc server listening on %s:%d", m.config.Host, m.config.Port)
	return nil
}

// Failed reports fatal errors of the running server (see module.Failer).
func (m *Module) Failed() <-chan error {
	return m.failed
}

// Stop gracefully shuts down the gRPC server, forcing it closed when ctx expires.
func (m *Module) Stop(ctx context.Context) error {
	logger.Log().Infof("stopping %s module", m.Name())

	if m.server == nil {
		return nil
	}

	done := make(chan struct{})
	go func() {
		m.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Log().Info("grpc server stopped gracefully")
	case <-ctx.Done():
		m.server.Server().Stop()
		<-done
		logger.Log().Warn("grpc server stop timed out; connections closed forcefully")
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

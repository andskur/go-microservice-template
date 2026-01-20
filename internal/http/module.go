// Package http implements the HTTP transport module.
package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-openapi/loads"
	"github.com/justinas/alice"

	"microservice-template/config"
	"microservice-template/internal/http/auth"
	"microservice-template/internal/http/handlers"
	"microservice-template/internal/http/middlewares"
	httpserver "microservice-template/internal/http/server"
	"microservice-template/internal/http/server/operations"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
)

// Module implements module.Module interface for the HTTP server.
type Module struct {
	config  *config.HTTPConfig
	service service.IService
	server  *httpserver.Server
	api     *operations.MicroserviceTemplateAPIAPI
	handler *http.Handler
	auth    *auth.Auth
}

// NewModule creates a new HTTP module instance.
func NewModule(cfg *config.HTTPConfig, svc service.IService) *Module {
	return &Module{
		config:  cfg,
		service: svc,
	}
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "http"
}

// Init initializes the HTTP module.
func (m *Module) Init(_ context.Context) error {
	logger.Log().Infof("initializing %s module", m.Name())

	// Initialize auth
	m.auth = auth.NewAuth(m.service, m.config.MockAuth, m.config.AdminEmails)

	// Initialize API
	if err := m.initAPI(); err != nil {
		return fmt.Errorf("init API: %w", err)
	}

	// Initialize server
	if err := m.initServer(); err != nil {
		return fmt.Errorf("init server: %w", err)
	}

	logger.Log().Infof("HTTP server configured on %s", m.config.Address)
	return nil
}

// Start begins module operation.
func (m *Module) Start(_ context.Context) error {
	logger.Log().Infof("starting %s module", m.Name())

	go func() {
		logger.Log().Infof("HTTP server listening on %s", m.config.Address)
		if err := m.server.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log().Errorf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the module.
func (m *Module) Stop(_ context.Context) error {
	logger.Log().Infof("stopping %s module", m.Name())

	if m.server != nil {
		if err := m.server.Shutdown(); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
	}

	logger.Log().Info("HTTP module stopped successfully")
	return nil
}

// HealthCheck returns module health status.
func (m *Module) HealthCheck(_ context.Context) error {
	// HTTP module is healthy if server is running
	// Could add a ping to actual server if needed
	return nil
}

// initAPI initializes the API and wires handlers.
func (m *Module) initAPI() error {
	// Load swagger spec
	swaggerSpec, err := loads.Analyzed(httpserver.SwaggerJSON, "")
	if err != nil {
		return fmt.Errorf("load swagger spec: %w", err)
	}

	// Create API instance
	api := operations.NewMicroserviceTemplateAPIAPI(swaggerSpec)

	// Configure logger
	api.Logger = logger.Log().Infof

	// Configure auth
	api.JwtAuth = m.auth.CheckAuth

	// Register handlers
	api.UsersGetUserByEmailHandler = handlers.NewGetUserByEmail(m.service)
	api.HealthGetHealthHandler = handlers.NewHealth()

	// TODO: Add more handlers as you expand the API
	// api.UsersCreateUserHandler = handlers.NewCreateUser(m.service)
	// api.UsersUpdateUserHandler = handlers.NewUpdateUser(m.service)
	// api.UsersDeleteUserHandler = handlers.NewDeleteUser(m.service)
	// api.UsersListUsersHandler = handlers.NewListUsers(m.service)

	// Build middleware chain
	handler := alice.New(
		middlewares.Recovery(),
		middlewares.Logger(),
		middlewares.Cors(m.config.CORS),
		middlewares.RateLimit(m.config.RateLimit),
	).Then(api.Serve(nil))

	m.api = api
	m.handler = &handler

	return nil
}

// initServer initializes the HTTP server.
func (m *Module) initServer() error {
	// Create server instance
	m.server = httpserver.NewServer(m.api)

	// Parse host and port
	host, portStr, err := net.SplitHostPort(m.config.Address)
	if err != nil {
		return fmt.Errorf("parse address: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("parse port: %w", err)
	}

	m.server.Host = host
	m.server.Port = port

	// Parse and set timeouts
	timeout, err := time.ParseDuration(m.config.Timeout)
	if err != nil {
		return fmt.Errorf("parse timeout: %w", err)
	}
	m.server.ReadTimeout = timeout
	m.server.WriteTimeout = timeout

	// Set handler with middleware
	m.server.SetHandler(*m.handler)

	return nil
}

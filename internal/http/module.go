// Package http implements the HTTP transport module.
package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-openapi/loads"
	"github.com/justinas/alice"

	"microservice-template/config"
	"microservice-template/internal/grpcclient"
	"microservice-template/internal/http/auth"
	"microservice-template/internal/http/handlers"
	"microservice-template/internal/http/middlewares"
	httpserver "microservice-template/internal/http/server"
	"microservice-template/internal/http/server/operations"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
)

// HealthAliasPath is the unversioned health endpoint, served in addition to
// <basePath>/health from the swagger spec (e.g. /api/v1/health).
const HealthAliasPath = "/health"

// Module implements module.Module interface for the HTTP server.
//
// The go-swagger generated code provides the API (routing, validation,
// handlers); the HTTP server itself is a plain net/http.Server so that the
// application controls listening, shutdown and signal handling.
type Module struct {
	config      *config.HTTPConfig
	service     service.IService
	grpcClient  grpcclient.IClient
	healthCheck handlers.HealthChecker
	api         *operations.MicroserviceTemplateAPIAPI
	handler     http.Handler
	auth        *auth.Auth
	server      *http.Server
	listener    net.Listener
	failed      chan error
	addr        atomic.Pointer[net.Addr] // set once listening; read concurrently
	serving     atomic.Bool
}

// NewModule creates a new HTTP module instance.
func NewModule(cfg *config.HTTPConfig, svc service.IService, grpcClient grpcclient.IClient) *Module {
	return &Module{
		config:     cfg,
		service:    svc,
		grpcClient: grpcClient,
	}
}

// SetHealthChecker sets the function used by the health endpoints to report
// the health of the application (typically module.Manager.HealthCheckAll).
// Must be called before Init.
func (m *Module) SetHealthChecker(fn handlers.HealthChecker) {
	m.healthCheck = fn
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

	logger.Log().Infof("HTTP server configured on %s", m.server.Addr)
	return nil
}

// Start binds the listen address and serves in the background. A bind error
// fails Start; a later serve error is reported through Failed.
func (m *Module) Start(ctx context.Context) error {
	logger.Log().Infof("starting %s module", m.Name())

	if m.server == nil {
		return errors.New("http module not initialized")
	}

	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", m.server.Addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", m.server.Addr, err)
	}
	m.listener = listener
	m.failed = make(chan error, 1)
	addr := listener.Addr()
	m.addr.Store(&addr)
	m.serving.Store(true)

	go func() {
		defer close(m.failed)
		defer m.serving.Store(false)

		logger.Log().Infof("HTTP server listening on %s", listener.Addr())
		if err := m.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			m.failed <- fmt.Errorf("http server: %w", err)
		}
	}()

	return nil
}

// Failed reports fatal errors of the running server (see module.Failer).
func (m *Module) Failed() <-chan error {
	return m.failed
}

// Addr returns the address the server is listening on (nil before Start).
func (m *Module) Addr() net.Addr {
	if addr := m.addr.Load(); addr != nil {
		return *addr
	}
	return nil
}

// Stop gracefully shuts down the module.
func (m *Module) Stop(ctx context.Context) error {
	logger.Log().Infof("stopping %s module", m.Name())

	if m.server != nil && m.listener != nil {
		if err := m.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
	}

	logger.Log().Info("HTTP module stopped successfully")
	return nil
}

// HealthCheck returns module health status.
func (m *Module) HealthCheck(_ context.Context) error {
	if !m.serving.Load() {
		return errors.New("http server not serving")
	}
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

	// Register handlers - pass grpcClient for external service access
	api.UsersGetUserByEmailHandler = handlers.NewGetUserByEmail(m.service, m.grpcClient)
	api.HealthGetHealthHandler = handlers.NewHealth(m.healthCheck)

	// TODO: Add more handlers as you expand the API
	// api.UsersCreateUserHandler = handlers.NewCreateUser(m.service)
	// api.UsersUpdateUserHandler = handlers.NewUpdateUser(m.service)
	// api.UsersDeleteUserHandler = handlers.NewDeleteUser(m.service)
	// api.UsersListUsersHandler = handlers.NewListUsers(m.service)

	versionedHealth := path.Join("/", swaggerSpec.BasePath(), "health")

	rateLimit, err := middlewares.RateLimit(m.config.RateLimit, HealthAliasPath, versionedHealth)
	if err != nil {
		return fmt.Errorf("rate limit: %w", err)
	}

	// Build middleware chain
	m.handler = alice.New(
		middlewares.Recovery(),
		middlewares.Logger(),
		middlewares.Cors(m.config.CORS),
		rateLimit,
		middlewares.PathAlias(HealthAliasPath, versionedHealth),
	).Then(api.Serve(nil))
	m.api = api

	return nil
}

// initServer initializes the HTTP server.
func (m *Module) initServer() error {
	if m.config.Host == "" {
		return fmt.Errorf("http host is required")
	}

	if m.config.Port < 0 || m.config.Port > 65535 {
		return fmt.Errorf("http port must be between 0 and 65535 (0 = random free port)")
	}

	timeout, err := time.ParseDuration(m.config.Timeout)
	if err != nil {
		return fmt.Errorf("parse timeout: %w", err)
	}

	m.server = &http.Server{
		Addr:              net.JoinHostPort(m.config.Host, strconv.Itoa(m.config.Port)),
		Handler:           m.handler,
		ReadHeaderTimeout: timeout,
		ReadTimeout:       timeout,
		WriteTimeout:      timeout,
	}

	return nil
}

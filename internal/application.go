// Package internal contains the core application wiring.
package internal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"microservice-template/config"
	"microservice-template/internal/module"
	"microservice-template/internal/repository"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
	"microservice-template/pkg/version"
)

// App is the main microservice application instance.
type App struct {
	config  *config.Scheme
	version *version.Version
	modules *module.Manager

	// Services (exposed to transports like HTTP/gRPC)
	// Will be nil if dependent modules (e.g., repository) are not enabled.
	// Example: userService depends on repository module.
	userService service.IUsersService
}

// NewApplication creates a new App instance.
func NewApplication() (app *App, err error) {
	ver, err := version.NewVersion()
	if err != nil {
		return nil, fmt.Errorf("init app version: %w", err)
	}

	return &App{
		config:  &config.Scheme{},
		version: ver,
		modules: module.NewManager(),
	}, nil
}

// Init initializes the application and all registered modules.
func (app *App) Init() error {
	ctx := context.Background()

	// Register modules based on configuration
	if err := app.registerModules(); err != nil {
		return fmt.Errorf("register modules: %w", err)
	}

	// Initialize all registered modules
	if err := app.modules.InitAll(ctx); err != nil {
		return fmt.Errorf("init modules: %w", err)
	}

	// At this point, services may still be nil if dependencies are disabled.
	// Transports (HTTP/gRPC) should check for nil before using services.

	return nil
}

// registerModules registers enabled modules based on configuration.
// Modules are registered in dependency order:
// 1. Infrastructure (database, cache, queue).
// 2. Business logic (repositories, services).
// 3. Transport (http, grpc).
func (app *App) registerModules() error {
	// Repository module (database-backed) and dependent service
	if app.config.Database != nil && app.config.Database.Enabled {
		logger.Log().Info("database enabled, registering repository module")

		repoModule := repository.NewModule(app.config.Database)
		app.modules.Register(repoModule)

		// Service layer depends on repository
		app.userService = service.NewUsersService(repoModule.Repository())
		logger.Log().Info("users service initialized with repository")
	} else {
		logger.Log().Info("database not enabled, repository module and user service not registered")
		app.userService = nil
	}

	logger.Log().Infof("registered %d modules", app.modules.Count())
	return nil
}

// Serve starts all modules and waits for shutdown signal.
func (app *App) Serve() error {
	ctx := context.Background()

	// Start all modules
	if err := app.modules.StartAll(ctx); err != nil {
		return fmt.Errorf("start modules: %w", err)
	}

	logger.Log().Info("application is running, press Ctrl+C to stop")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit
	logger.Log().Info("shutdown signal received, stopping gracefully...")

	return nil
}

// Stop gracefully shuts down all modules.
func (app *App) Stop() error {
	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return app.modules.StopAll(ctx)
}

// Config returns the application configuration.
func (app *App) Config() *config.Scheme {
	return app.config
}

// Version returns the application version string.
func (app *App) Version() string {
	return app.version.String()
}

// Modules returns the module manager (useful for health checks).
func (app *App) Modules() *module.Manager {
	return app.modules
}

// UserService returns the user service instance.
// Returns nil if the repository module is not enabled.
func (app *App) UserService() service.IUsersService {
	return app.userService
}

// CreateAddr creates an address string from host and port.
func CreateAddr(host string, port int) string {
	return fmt.Sprintf("%s:%v", host, port)
}

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
	grpcmod "microservice-template/internal/grpc"
	grpcclientmod "microservice-template/internal/grpcclient"
	httpmod "microservice-template/internal/http"
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
	svc service.IService
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

	// Capture service instance from service module (always registered)
	for _, mod := range app.modules.List() {
		if svcMod, ok := mod.(interface{ Service() service.IService }); ok {
			app.svc = svcMod.Service()
			break
		}
	}

	return nil
}

// registerModules registers enabled modules based on configuration.
// Modules are registered in dependency order:
// 1. Infrastructure (database, cache, queue).
// 2. Business logic (repositories, services).
// 3. Transport (http, grpc).
func (app *App) registerModules() error {
	// 1. Infrastructure: Repository (database-backed) is optional
	var repoModule *repository.Module
	if app.config.Database != nil && app.config.Database.Enabled {
		logger.Log().Info("database enabled, registering repository module")

		repoModule = repository.NewModule(app.config.Database)
		app.modules.Register(repoModule)
	} else {
		logger.Log().Info("database not enabled, repository module not registered")
	}

	// 2. Infrastructure: gRPC Client for external services (optional)
	var grpcClientModule *grpcclientmod.Module
	if app.config.GRPCClient != nil && app.config.GRPCClient.Enabled {
		logger.Log().Info("grpc_client enabled, registering grpc client module")

		grpcClientModule = grpcclientmod.NewModule(app.config.GRPCClient)
		app.modules.Register(grpcClientModule)
	} else {
		logger.Log().Info("grpc_client not enabled, grpc client module not registered")
	}

	// 3. Business logic: Service module is always registered; repository may be nil
	logger.Log().Info("registering service module")

	var repo repository.IRepository
	if repoModule != nil {
		repo = repoModule.Repository()
	}

	// Service module does NOT depend on grpcClient
	// Service handles local business logic only
	svcModule := service.NewModule(repo)
	app.modules.Register(svcModule)

	// Capture service instance for downstream transports
	for _, mod := range app.modules.List() {
		if svcMod, ok := mod.(interface{ Service() service.IService }); ok {
			app.svc = svcMod.Service()
			break
		}
	}

	// 4. Transport: HTTP module (optional) - receives both service AND grpcClient
	if app.config.HTTP != nil && app.config.HTTP.Enabled {
		logger.Log().Info("http enabled, registering http module")

		// Pass grpcClient to HTTP module (can be nil)
		httpModule := httpmod.NewModule(app.config.HTTP, app.svc, grpcClientModule)
		app.modules.Register(httpModule)
	} else {
		logger.Log().Info("http not enabled, http module not registered")
	}

	// 5. Transport: gRPC server module (optional)
	if app.config.GRPC != nil && app.config.GRPC.Enabled {
		logger.Log().Info("grpc enabled, registering grpc module")

		grpcModule := grpcmod.NewModule(app.config.GRPC, app.svc)
		app.modules.Register(grpcModule)
	} else {
		logger.Log().Info("grpc not enabled, grpc module not registered")
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

// Service returns the service instance.
// Service is always registered; methods may fail if dependencies are unavailable.
func (app *App) Service() service.IService {
	return app.svc
}

// CreateAddr creates an address string from host and port.
func CreateAddr(host string, port int) string {
	return fmt.Sprintf("%s:%v", host, port)
}

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
	wsmod "microservice-template/internal/websocket"
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
	// Register and initialize modules based on configuration
	// Note: registerModules handles both registration and initialization
	// in the correct order to ensure dependencies are available
	if err := app.registerModules(); err != nil {
		return fmt.Errorf("register modules: %w", err)
	}

	return nil
}

// registerModules registers enabled modules based on configuration.
// Modules are registered in dependency order:
// 1. Infrastructure (database, cache, queue).
// 2. Business logic (repositories, services).
// 3. Transport (http, grpc).
//
//nolint:gocyclo // should decompose later
func (app *App) registerModules() error {
	// 1. Infrastructure: Repository (database-backed) is optional
	var repoModule *repository.Module
	if app.config.Database != nil && app.config.Database.Enabled {
		logger.Log().Info("database enabled, registering repository module")

		repoModule = repository.NewModule(app.config.Database)
		app.modules.Register(repoModule)
	}

	// 2. Infrastructure: gRPC Client for external services (optional)
	var grpcClientModule *grpcclientmod.Module
	if app.config.GRPCClient != nil && app.config.GRPCClient.Enabled {
		logger.Log().Info("grpc_client enabled, registering grpc client module")

		grpcClientModule = grpcclientmod.NewModule(app.config.GRPCClient)
		app.modules.Register(grpcClientModule)
	}

	// 3. Business logic: Service module is always registered; repository may be nil
	logger.Log().Info("registering service module")

	// Pass repository module as provider; service will retrieve repository during Init
	// (after repository module has been initialized).
	// Explicitly pass nil to avoid typed nil interface gotcha.
	var repoProvider service.RepositoryProvider
	if repoModule != nil {
		repoProvider = repoModule
	}
	svcModule := service.NewModule(repoProvider)
	app.modules.Register(svcModule)

	// Initialize infrastructure and business logic modules first
	// so we can retrieve the service instance for transport modules
	ctx := context.Background()
	if repoModule != nil {
		if err := repoModule.Init(ctx); err != nil {
			return fmt.Errorf("init repository module: %w", err)
		}
	}
	if grpcClientModule != nil {
		if err := grpcClientModule.Init(ctx); err != nil {
			return fmt.Errorf("init grpc client module: %w", err)
		}
	}
	if err := svcModule.Init(ctx); err != nil {
		return fmt.Errorf("init service module: %w", err)
	}

	// Capture service instance after initialization
	app.svc = svcModule.Service()

	logger.Log().Info("infrastructure modules initialized successfully")

	// 4. Transport: HTTP module (optional) - receives both service AND grpcClient
	if app.config.HTTP != nil && app.config.HTTP.Enabled {
		logger.Log().Info("http enabled, registering http module")

		// Pass grpcClient to HTTP module (can be nil)
		httpModule := httpmod.NewModule(app.config.HTTP, app.svc, grpcClientModule)
		app.modules.Register(httpModule)

		// Initialize HTTP module
		if err := httpModule.Init(ctx); err != nil {
			return fmt.Errorf("init http module: %w", err)
		}
	}

	// 5. Transport: gRPC server module (optional)
	if app.config.GRPC != nil && app.config.GRPC.Enabled {
		logger.Log().Info("grpc enabled, registering grpc module")

		grpcModule := grpcmod.NewModule(app.config.GRPC, app.svc)
		app.modules.Register(grpcModule)

		// Initialize gRPC module
		if err := grpcModule.Init(ctx); err != nil {
			return fmt.Errorf("init grpc module: %w", err)
		}
	}

	// 6. Transport: WebSocket server module (optional)
	if app.config.WebSocket != nil && app.config.WebSocket.Enabled {
		logger.Log().Info("websocket enabled, registering websocket module")

		wsModule := wsmod.NewModule(app.config.WebSocket, app.svc)
		app.modules.Register(wsModule)

		// Initialize WebSocket module
		if err := wsModule.Init(ctx); err != nil {
			return fmt.Errorf("init websocket module: %w", err)
		}
	}

	logger.Log().Infof("registered and initialized %d modules", app.modules.Count())
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

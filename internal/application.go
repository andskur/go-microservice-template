package internal

import (
	"os"
	"os/signal"
	"syscall"

	"microservice-template/config"
)

// App is main microservice application instance that
// have all necessary dependencies inside structure
type App struct {
	// application configuration
	config *config.Scheme

	// TODO add all needed dependencies
}

// NewApplication create new App instance
func NewApplication() (app *App, err error) {
	return &App{
		config: &config.Scheme{},
	}, nil
}

// Init initialize application and all necessary instances
func (app *App) Init() error {
	// TODO add dependencies initialisations

	return nil
}

// Serve start serving Application service
func (app *App) Serve() error {
	// TODO add all runners that needed in separate goroutines

	// Gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit

	return nil
}

// Stop shutdown the application
func (app *App) Stop() error {
	// TODO shutdown all dependencies that need to be stopped

	return nil
}

// Config return App config Scheme
func (app *App) Config() *config.Scheme {
	return app.config
}

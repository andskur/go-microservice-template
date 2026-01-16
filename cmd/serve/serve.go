// Package serve defines the CLI serve command.
package serve

import (
	"fmt"

	"github.com/spf13/cobra"

	"microservice-template/internal"
	"microservice-template/pkg/logger"
)

// Cmd returns the "serve" command of the application.
// This command is responsible for initializing and running the service.
func Cmd(app *internal.App) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run Application",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := app.Init(); err != nil {
				return fmt.Errorf("application initialisation: %w", err)
			}

			return app.Serve()
		},
		PreRun: func(_ *cobra.Command, _ []string) {
			logger.Log().Info(app.Version())
		},
		PostRun: func(_ *cobra.Command, _ []string) {
			if err := app.Stop(); err != nil {
				logger.Log().Errorf("cant stop app")
			}
			logger.Log().Info("App stopped")
		},
	}
}

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
		RunE: func(_ *cobra.Command, _ []string) (err error) {
			// Stop runs on every exit path (cobra skips PostRun when RunE fails),
			// so modules initialized before a failure are shut down too.
			defer func() {
				if stopErr := app.Stop(); stopErr != nil {
					logger.Log().Errorf("stop app: %v", stopErr)
					if err == nil {
						err = fmt.Errorf("stop app: %w", stopErr)
					}
				}
				logger.Log().Info("App stopped")
			}()

			if err := app.Init(); err != nil {
				return fmt.Errorf("application initialisation: %w", err)
			}

			return app.Serve()
		},
		PreRun: func(_ *cobra.Command, _ []string) {
			logger.Log().Info(app.Version())
		},
	}
}

package serve

import (
	"fmt"

	"github.com/Misnaged/annales/logger"
	"github.com/spf13/cobra"

	"microservice-template/internal"
)

// Cmd returns the "serve" command of the application.
// This command is responsible for initializing and
func Cmd(app *internal.App) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run Application",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Init(); err != nil {
				return fmt.Errorf("application initialisation: %w", err)
			}

			return app.Serve()
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			logger.Log().Info(app.Version())
		},
	}
}

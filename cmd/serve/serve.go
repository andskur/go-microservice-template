package serve

import (
	"fmt"
	"github.com/spf13/cobra"
	"microservice-template/internal"
)

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
	}
}

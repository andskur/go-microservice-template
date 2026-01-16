// Package root defines the root CLI command.
package root

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"microservice-template/config"
	"microservice-template/internal"
)

// Cmd returns the root command for the application.
func Cmd(app *internal.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "microservice",
		Short:            "Service Template",
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return initializeConfig(cmd, app.Config())
		},
	}

	cmd.SetVersionTemplate(app.Version())

	return cmd
}

// initializeConfig reads in config file and sets configuration via environment variables.
func initializeConfig(cmd *cobra.Command, cfg *config.Scheme) error {
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			return fmt.Errorf("read config file: %w", err)
		}
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)

	bindFlags(cmd)

	return viper.Unmarshal(cfg)
}

// bindFlags binds flags to the command.
func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

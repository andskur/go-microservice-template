package root

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"microservice-template/config"
	"microservice-template/internal"
)

func Cmd(app *internal.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "microservice",
		Short:            "Service Template",
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd, app.Config())
		},
	}

	return cmd
}

func initializeConfig(cmd *cobra.Command, cfg *config.Scheme) error {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config file: %w", err)
		}
	}

	// set config via env vars
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)

	bindFlags(cmd)

	return viper.Unmarshal(cfg)
}

func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

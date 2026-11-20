// Package root defines the root CLI command.
package root

import (
	"errors"
	"fmt"
	"os"
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
		Use:              "microservice-template",
		Short:            "Service Template",
		TraverseChildren: true,
		// Runtime errors are logged by main; do not print usage for them.
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			configFile, err := cmd.Flags().GetString(configFlag)
			if err != nil {
				return fmt.Errorf("read --%s flag: %w", configFlag, err)
			}
			setConfigFile(configFile, cmd.Root().Name())
			return initializeConfig(cmd, app.Config())
		},
	}

	cmd.PersistentFlags().String(configFlag, os.Getenv(configFileEnv),
		"path to a config file (yaml, json, toml, ...); env: "+configFileEnv+
			". Without it, an optional config.{yaml,json,toml} is searched in ., ./config and /etc/<name>")

	cmd.Version = app.Version()
	cmd.SetVersionTemplate("{{printf \"%s\" .Version}}\n")

	return cmd
}

const (
	configFlag    = "config"
	configFileEnv = "CONFIG_FILE"
)

// setConfigFile tells viper where to find the config file.
// An explicitly given file must exist; otherwise "config.<ext>" is looked up in
// the working directory, ./config and /etc/<appName> and is optional.
func setConfigFile(path, appName string) {
	if path != "" {
		viper.SetConfigFile(path)
		return
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/" + appName)
}

// initializeConfig reads in config file and sets configuration via environment variables.
// Env and flags are bound after config load so CLI flags override env, which override config file.
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

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("bind flags: %w", err)
	}

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

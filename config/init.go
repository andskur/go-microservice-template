// Package config defines application configuration defaults and schema.
package config

import (
	"github.com/spf13/viper"
)

// init initialize default config params.
//
//nolint:gochecknoinits // configuration defaults are registered at package load.
func init() {
	// environment - could be "local", "prod", "dev"
	viper.SetDefault("env", "prod")

	// TODO add default values for all configuration fields
}

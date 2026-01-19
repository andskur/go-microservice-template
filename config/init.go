// Package config defines application configuration defaults and schema.
package config

import (
	"github.com/spf13/viper"
)

// init initialize default config params.
//
//nolint:gochecknoinits // configuration defaults are registered at package load.
func init() {
	setDefaults()
}

// setDefaults exposes default registration for testing.
// Keep defaults centralized here so tests can reset viper and reapply them.
func setDefaults() {
	// Core application defaults
	viper.SetDefault("env", "prod")

	// Database/Repository module defaults
	viper.SetDefault("database.enabled", false)
	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", 300)

	// gRPC module defaults
	viper.SetDefault("grpc.enabled", false)
	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("grpc.port", 9090)
	viper.SetDefault("grpc.timeout", "30s")
	viper.SetDefault("grpc.max_send_msg_size", 60*1024*1024)
	viper.SetDefault("grpc.max_recv_msg_size", 60*1024*1024)
	viper.SetDefault("grpc.num_stream_workers", 0)

	// TODO: Add default values for additional module configurations as you implement them
	//
	// Example: HTTP module defaults
	// viper.SetDefault("http.enabled", false)
	// viper.SetDefault("http.host", "0.0.0.0")
	// viper.SetDefault("http.port", 8080)
	// viper.SetDefault("http.read_timeout", 30)
	// viper.SetDefault("http.write_timeout", 30)
	// viper.SetDefault("http.shutdown_timeout", 10)
}

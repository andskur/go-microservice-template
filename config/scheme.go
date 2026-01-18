// Package config defines application configuration defaults and schema.
package config

// Scheme represents the application configuration scheme.
type Scheme struct {
	// Database configuration for repository module (optional; nil if disabled).
	Database *DatabaseConfig `mapstructure:"database"`

	// Env is the application environment (e.g. prod, dev, local).
	Env string `mapstructure:"env"`

	// TODO: Add more module configurations here as you implement them
	// Example: HTTP server module configuration
	// HTTP *HTTPConfig `mapstructure:"http"`
	// Example: gRPC server module configuration
	// GRPC *GRPCConfig `mapstructure:"grpc"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"` // postgres, mysql, sqlite
	Host            string `mapstructure:"host"`
	Name            string `mapstructure:"name"` // database name
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	SSLMode         string `mapstructure:"ssl_mode"` // disable, require, verify-full
	Port            int    `mapstructure:"port"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // seconds
	Enabled         bool   `mapstructure:"enabled"`
}

// Example module configuration structs (uncomment and modify as needed):
//
// type HTTPConfig struct {
// 	Host            string `mapstructure:"host"`
// 	Port            int    `mapstructure:"port"`
// 	ReadTimeout     int    `mapstructure:"read_timeout"`      // seconds
// 	WriteTimeout    int    `mapstructure:"write_timeout"`     // seconds
// 	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`  // seconds
// 	Enabled         bool   `mapstructure:"enabled"`
// }
//
// type GRPCConfig struct {
// 	Host    string `mapstructure:"host"`
// 	Port    int    `mapstructure:"port"`
// 	Enabled bool   `mapstructure:"enabled"`
// }

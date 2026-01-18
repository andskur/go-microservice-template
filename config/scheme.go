// Package config defines application configuration defaults and schema.
package config

// Scheme represents the application configuration scheme.
type Scheme struct {
	// Env is the application environment (e.g. prod, dev, local).
	Env string `mapstructure:"env"`

	// TODO: Add module configurations here as you implement them
	//
	// Example: Database module configuration
	// Database *DatabaseConfig `mapstructure:"database"`
	//
	// Example: HTTP server module configuration
	// HTTP *HTTPConfig `mapstructure:"http"`
	//
	// Example: gRPC server module configuration
	// GRPC *GRPCConfig `mapstructure:"grpc"`
}

// Example module configuration structs (uncomment and modify as needed):
//
// type DatabaseConfig struct {
// 	Driver          string `mapstructure:"driver"`           // postgres, mysql, sqlite
// 	Host            string `mapstructure:"host"`
// 	Name            string `mapstructure:"name"`             // database name
// 	User            string `mapstructure:"user"`
// 	Password        string `mapstructure:"password"`
// 	SSLMode         string `mapstructure:"ssl_mode"`         // disable, require, verify-full
// 	Port            int    `mapstructure:"port"`
// 	MaxOpenConns    int    `mapstructure:"max_open_conns"`
// 	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
// 	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // seconds
// 	Enabled         bool   `mapstructure:"enabled"`
// }
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

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	ReadTimeout     int    `mapstructure:"read_timeout"`     // seconds
	WriteTimeout    int    `mapstructure:"write_timeout"`    // seconds
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"` // seconds
	Enabled         bool   `mapstructure:"enabled"`
}

// GRPCConfig holds gRPC server settings.
type GRPCConfig struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	Enabled bool   `mapstructure:"enabled"`
}

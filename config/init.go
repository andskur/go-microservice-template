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
	viper.SetDefault("database.user", "dev")
	viper.SetDefault("database.password", "dev")
	viper.SetDefault("database.name", "microservices_dev")

	// gRPC module defaults
	viper.SetDefault("grpc.enabled", false)
	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("grpc.port", 9090)
	viper.SetDefault("grpc.timeout", "30s")
	viper.SetDefault("grpc.max_send_msg_size", 60*1024*1024)
	viper.SetDefault("grpc.max_recv_msg_size", 60*1024*1024)
	viper.SetDefault("grpc.num_stream_workers", 0)

	// gRPC Client module defaults
	viper.SetDefault("grpc_client.enabled", false)
	viper.SetDefault("grpc_client.address", "localhost:9090")
	viper.SetDefault("grpc_client.timeout", "30s")
	viper.SetDefault("grpc_client.keep_alive.time", "10s")
	viper.SetDefault("grpc_client.keep_alive.timeout", "1s")
	viper.SetDefault("grpc_client.keep_alive.permit_without_stream", true)

	// HTTP module defaults
	viper.SetDefault("http.enabled", false)
	viper.SetDefault("http.host", "0.0.0.0")
	viper.SetDefault("http.port", 8080)
	viper.SetDefault("http.timeout", "30s")
	viper.SetDefault("http.swagger_spec", "./api/swagger.yaml")
	viper.SetDefault("http.mock_auth", false)
	viper.SetDefault("http.admin_emails", []string{})

	// CORS defaults
	viper.SetDefault("http.cors.enabled", true)
	viper.SetDefault("http.cors.allowed_origins", []string{"*"})
	viper.SetDefault("http.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"})
	viper.SetDefault("http.cors.allowed_headers", []string{"*"})
	viper.SetDefault("http.cors.max_age", 3600)

	// Rate limit defaults
	viper.SetDefault("http.rate_limit.enabled", false)
	viper.SetDefault("http.rate_limit.requests_per_sec", 100.0)
	viper.SetDefault("http.rate_limit.burst", 20)

	// Gatekeeper defaults (for future use)
	viper.SetDefault("http.gatekeeper.address", "localhost:9091")
	viper.SetDefault("http.gatekeeper.timeout", "5s")

	// WebSocket module defaults
	viper.SetDefault("websocket.enabled", false)
	viper.SetDefault("websocket.host", "0.0.0.0")
	viper.SetDefault("websocket.port", 8081)
	viper.SetDefault("websocket.timeout", "30s")
	viper.SetDefault("websocket.read_buffer_size", 1024)
	viper.SetDefault("websocket.write_buffer_size", 1024)
	viper.SetDefault("websocket.max_message_size", 512000) // 500KB
	viper.SetDefault("websocket.ping_interval", "54s")
	viper.SetDefault("websocket.pong_wait", "60s")
	viper.SetDefault("websocket.write_wait", "10s")

	// WebSocket connection limits
	viper.SetDefault("websocket.limits.max_connections", 0)          // 0 = unlimited
	viper.SetDefault("websocket.limits.max_connections_per_room", 0) // 0 = unlimited
}

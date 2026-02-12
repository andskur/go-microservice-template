// Package config defines application configuration defaults and schema.
package config

// GRPCConfig holds gRPC server settings.
type GRPCConfig struct {
	Host             string `mapstructure:"host"`
	Timeout          string `mapstructure:"timeout"`
	MaxSendMsgSize   int    `mapstructure:"max_send_msg_size"`
	MaxRecvMsgSize   int    `mapstructure:"max_recv_msg_size"`
	Port             int    `mapstructure:"port"`
	NumStreamWorkers uint32 `mapstructure:"num_stream_workers"`
	Enabled          bool   `mapstructure:"enabled"`
}

// GRPCClientConfig holds gRPC client settings for connecting to external services.
type GRPCClientConfig struct {
	KeepAlive *KeepAliveConfig `mapstructure:"keep_alive"` // Keep-alive settings
	Address   string           `mapstructure:"address"`    // External service address (e.g., "user-service:9090")
	Timeout   string           `mapstructure:"timeout"`    // Request timeout (e.g., "30s")
	Enabled   bool             `mapstructure:"enabled"`    // Enable gRPC client module
}

// KeepAliveConfig holds gRPC keep-alive settings.
type KeepAliveConfig struct {
	Time                string `mapstructure:"time"`                  // Send pings interval (e.g., "10s")
	Timeout             string `mapstructure:"timeout"`               // Ping ack timeout (e.g., "1s")
	PermitWithoutStream bool   `mapstructure:"permit_without_stream"` // Send pings even without active streams
}

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	// Pointers first to reduce padding.
	CORS       *CORSConfig       `mapstructure:"cors"`       // CORS settings
	RateLimit  *RateLimitConfig  `mapstructure:"rate_limit"` // Rate limiting
	Gatekeeper *GatekeeperConfig `mapstructure:"gatekeeper"` // Gatekeeper configuration (JWT validation service)

	Host        string   `mapstructure:"host"`         // Server host (e.g., "0.0.0.0" or "localhost")
	Timeout     string   `mapstructure:"timeout"`      // Request timeout (e.g., "30s")
	SwaggerSpec string   `mapstructure:"swagger_spec"` // Path to swagger.yaml
	AdminEmails []string `mapstructure:"admin_emails"` // Admin user emails for role checking
	Port        int      `mapstructure:"port"`         // Server port (e.g., 8080)

	Enabled  bool `mapstructure:"enabled"`   // Enable HTTP module
	MockAuth bool `mapstructure:"mock_auth"` // Enable mock auth for testing (bypasses gatekeeper)
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"` // ["*"] or ["https://myapp.com"]
	AllowedMethods []string `mapstructure:"allowed_methods"` // ["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"]
	AllowedHeaders []string `mapstructure:"allowed_headers"` // ["*"] or specific headers
	MaxAge         int      `mapstructure:"max_age"`         // Preflight cache duration in seconds
	Enabled        bool     `mapstructure:"enabled"`         // Enable CORS middleware
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	Enabled        bool    `mapstructure:"enabled"`          // Enable rate limiting middleware
	RequestsPerSec float64 `mapstructure:"requests_per_sec"` // Requests per second (e.g., 100.0)
	Burst          int     `mapstructure:"burst"`            // Burst size (e.g., 20)
}

// GatekeeperConfig holds gatekeeper service settings.
type GatekeeperConfig struct {
	Address string `mapstructure:"address"` // gRPC address (e.g., "localhost:9091")
	Timeout string `mapstructure:"timeout"` // Request timeout (e.g., "5s")
}

// WebSocketConfig holds WebSocket server settings.
type WebSocketConfig struct {
	Limits          *WSLimitsConfig `mapstructure:"limits"`            // Connection limits
	Host            string          `mapstructure:"host"`              // Server host (e.g., "0.0.0.0")
	Timeout         string          `mapstructure:"timeout"`           // Connection timeout (e.g., "30s")
	PingInterval    string          `mapstructure:"ping_interval"`     // Ping keepalive interval (e.g., "54s")
	PongWait        string          `mapstructure:"pong_wait"`         // Pong response timeout (e.g., "60s")
	WriteWait       string          `mapstructure:"write_wait"`        // Write deadline (e.g., "10s")
	MaxMessageSize  int64           `mapstructure:"max_message_size"`  // Max message size in bytes
	Port            int             `mapstructure:"port"`              // Server port (e.g., 8081)
	ReadBufferSize  int             `mapstructure:"read_buffer_size"`  // Read buffer size in bytes
	WriteBufferSize int             `mapstructure:"write_buffer_size"` // Write buffer size in bytes
	Enabled         bool            `mapstructure:"enabled"`           // Enable WebSocket module
}

// WSLimitsConfig holds WebSocket connection limit settings.
type WSLimitsConfig struct {
	MaxConnections        int `mapstructure:"max_connections"`          // Global max connections (0 = unlimited)
	MaxConnectionsPerRoom int `mapstructure:"max_connections_per_room"` // Per-room max connections (0 = unlimited)
}

// Scheme represents the application configuration scheme.
type Scheme struct {
	// Database configuration for repository module (optional; nil if disabled).
	Database *DatabaseConfig `mapstructure:"database"`

	// GRPC configuration for gRPC module (optional; nil if disabled).
	GRPC *GRPCConfig `mapstructure:"grpc"`

	// GRPCClient configuration for gRPC client module (optional; nil if disabled).
	GRPCClient *GRPCClientConfig `mapstructure:"grpc_client"`

	// HTTP configuration for HTTP module (optional; nil if disabled).
	HTTP *HTTPConfig `mapstructure:"http"`

	// WebSocket configuration for WebSocket module (optional; nil if disabled).
	WebSocket *WebSocketConfig `mapstructure:"websocket"`

	// Env is the application environment (e.g. prod, dev, local).
	Env string `mapstructure:"env"`
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

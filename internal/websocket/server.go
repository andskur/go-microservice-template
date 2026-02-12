package websocket

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"microservice-template/config"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
)

// Server is the WebSocket HTTP server.
type Server struct {
	config   *config.WebSocketConfig
	hub      *Hub
	handlers *MessageHandlers
	server   *http.Server
	upgrader *websocket.Upgrader
	listener net.Listener
	running  bool

	// Parsed durations from config.
	pongWait     time.Duration
	pingInterval time.Duration
	writeWait    time.Duration
}

// NewServer creates a new WebSocket server.
func NewServer(cfg *config.WebSocketConfig, svc service.IService) (*Server, error) {
	// Parse durations
	pongWait, err := time.ParseDuration(cfg.PongWait)
	if err != nil {
		return nil, fmt.Errorf("parse pong_wait: %w", err)
	}

	pingInterval, err := time.ParseDuration(cfg.PingInterval)
	if err != nil {
		return nil, fmt.Errorf("parse ping_interval: %w", err)
	}

	writeWait, err := time.ParseDuration(cfg.WriteWait)
	if err != nil {
		return nil, fmt.Errorf("parse write_wait: %w", err)
	}

	// Create hub
	hub := NewHub(cfg.Limits)

	// Create handlers with service access
	handlers := NewMessageHandlers(hub, svc)

	// Create upgrader
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
		CheckOrigin: func(_ *http.Request) bool {
			// Allow all origins for now
			// TODO: Add origin checking based on config
			return true
		},
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	s := &Server{
		config:       cfg,
		hub:          hub,
		handlers:     handlers,
		upgrader:     upgrader,
		pongWait:     pongWait,
		pingInterval: pingInterval,
		writeWait:    writeWait,
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	return s, nil
}

// handleWebSocket handles WebSocket upgrade requests.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Log().Errorf("websocket upgrade failed: %v", err)
		return
	}

	// Create client configuration
	clientCfg := &ClientConfig{
		PongWait:     s.pongWait,
		PingInterval: s.pingInterval,
		WriteWait:    s.writeWait,
		MaxMsgSize:   s.config.MaxMessageSize,
	}

	// Create new client
	client := NewClient(s.hub, conn, s.handlers, clientCfg)

	// Register with hub
	s.hub.Register(client)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}

// handleHealth handles health check requests.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	if !s.hub.IsRunning() {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"unhealthy","error":"hub not running"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"healthy","clients":%d,"rooms":%d}`,
		s.hub.ClientCount(), s.hub.RoomCount())
}

// Start starts the WebSocket server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Create listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	s.listener = listener

	// Start hub
	go s.hub.Run()

	// Start HTTP server
	s.running = true
	go func() {
		logger.Log().Infof("WebSocket server listening on %s", addr)
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Log().Errorf("WebSocket server error: %v", err)
		}
		s.running = false
	}()

	return nil
}

// Stop gracefully stops the WebSocket server.
func (s *Server) Stop(ctx context.Context) error {
	// Stop hub first (closes all client connections)
	s.hub.Stop()

	// Shutdown HTTP server
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	s.running = false
	logger.Log().Info("WebSocket server stopped")
	return nil
}

// IsRunning returns true if the server is running.
func (s *Server) IsRunning() bool {
	return s.running
}

// Hub returns the server's hub.
func (s *Server) Hub() *Hub {
	return s.hub
}

// Handlers returns the server's message handlers.
func (s *Server) Handlers() *MessageHandlers {
	return s.handlers
}

// ClientCount returns the number of connected clients.
func (s *Server) ClientCount() int {
	return s.hub.ClientCount()
}

// RoomCount returns the number of active rooms.
func (s *Server) RoomCount() int {
	return s.hub.RoomCount()
}

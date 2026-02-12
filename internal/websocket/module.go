package websocket

import (
	"context"
	"fmt"

	"microservice-template/config"
	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
)

// Module implements the module.Module interface for WebSocket server.
type Module struct {
	config  *config.WebSocketConfig
	service service.IService
	server  *Server
}

// NewModule creates a new WebSocket module.
func NewModule(cfg *config.WebSocketConfig, svc service.IService) *Module {
	return &Module{
		config:  cfg,
		service: svc,
	}
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "websocket"
}

// Init initializes the WebSocket server.
func (m *Module) Init(_ context.Context) error {
	logger.Log().Infof("initializing %s module", m.Name())

	// Create server
	server, err := NewServer(m.config, m.service)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}
	m.server = server

	logger.Log().Infof("%s module initialized (port: %d)", m.Name(), m.config.Port)
	return nil
}

// Start begins WebSocket server operation.
func (m *Module) Start(_ context.Context) error {
	logger.Log().Infof("starting %s module", m.Name())

	if err := m.server.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	logger.Log().Infof("%s module started on %s:%d", m.Name(), m.config.Host, m.config.Port)
	return nil
}

// Stop gracefully shuts down the WebSocket server.
func (m *Module) Stop(ctx context.Context) error {
	logger.Log().Infof("stopping %s module", m.Name())

	if m.server != nil {
		if err := m.server.Stop(ctx); err != nil {
			return fmt.Errorf("stop server: %w", err)
		}
	}

	logger.Log().Infof("%s module stopped", m.Name())
	return nil
}

// HealthCheck returns the module health status.
func (m *Module) HealthCheck(_ context.Context) error {
	if m.server == nil {
		return ErrServerNotRunning
	}

	if !m.server.IsRunning() {
		return ErrServerNotRunning
	}

	if !m.server.Hub().IsRunning() {
		return ErrHubNotRunning
	}

	return nil
}

// Server returns the WebSocket server instance.
func (m *Module) Server() *Server {
	return m.server
}

// Hub returns the WebSocket hub instance.
func (m *Module) Hub() *Hub {
	if m.server == nil {
		return nil
	}
	return m.server.Hub()
}

// ClientCount returns the number of connected clients.
func (m *Module) ClientCount() int {
	if m.server == nil {
		return 0
	}
	return m.server.ClientCount()
}

// RoomCount returns the number of active rooms.
func (m *Module) RoomCount() int {
	if m.server == nil {
		return 0
	}
	return m.server.RoomCount()
}

// Broadcast sends a message to all connected clients.
func (m *Module) Broadcast(msg *OutgoingMessage) {
	if m.server != nil && m.server.Hub() != nil {
		m.server.Hub().Broadcast(msg, nil)
	}
}

// RoomBroadcast sends a message to all clients in a room.
func (m *Module) RoomBroadcast(room string, msg *OutgoingMessage) {
	if m.server != nil && m.server.Hub() != nil {
		m.server.Hub().RoomBroadcast(room, msg, nil)
	}
}

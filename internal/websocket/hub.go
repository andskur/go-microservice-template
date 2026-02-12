package websocket

import (
	"sync"
	"sync/atomic"

	"microservice-template/config"
	"microservice-template/pkg/logger"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Room manager.
	rooms *RoomManager

	// Connection limits.
	limits *config.WSLimitsConfig

	// Inbound messages for broadcast to all clients.
	broadcast chan *BroadcastMessage

	// Inbound messages for room broadcast.
	roomcast chan *RoomMessage

	// Register requests from clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Done channel for graceful shutdown.
	done chan struct{}

	// Mutex for clients map.
	mu sync.RWMutex

	// Running state.
	running int32
}

// NewHub creates a new Hub.
func NewHub(limits *config.WSLimitsConfig) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      NewRoomManager(),
		broadcast:  make(chan *BroadcastMessage, 256),
		roomcast:   make(chan *RoomMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		limits:     limits,
		done:       make(chan struct{}),
	}
}

// Run starts the hub's main event loop.
// This should be called in a goroutine.
func (h *Hub) Run() {
	atomic.StoreInt32(&h.running, 1)
	defer atomic.StoreInt32(&h.running, 0)

	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)

		case message := <-h.roomcast:
			h.handleRoomcast(message)

		case <-h.done:
			h.handleShutdown()
			return
		}
	}
}

// handleRegister handles a client registration.
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check connection limits
	if h.limits != nil && h.limits.MaxConnections > 0 {
		if len(h.clients) >= h.limits.MaxConnections {
			logger.Log().Warnf("max connections reached (%d), rejecting client %s",
				h.limits.MaxConnections, client.ID())
			client.SendMessage(NewErrorMessage(ErrMsgMaxConnections))
			client.Close()
			return
		}
	}

	h.clients[client] = true
	logger.Log().Infof("client %s registered, total clients: %d", client.ID(), len(h.clients))

	// Send welcome message
	client.SendMessage(&OutgoingMessage{
		Type: MsgTypeConnected,
		Data: &ConnectedData{
			ClientID: client.ID(),
			Message:  "Connected to WebSocket server",
		},
	})
}

// handleUnregister handles a client unregistration.
func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		// Remove client from all rooms
		rooms := client.LeaveAllRooms()
		for _, roomName := range rooms {
			room := h.rooms.Get(roomName)
			if room != nil {
				room.RemoveClient(client)
				// Clean up empty rooms
				h.rooms.RemoveIfEmpty(roomName)
			}
		}

		delete(h.clients, client)
		logger.Log().Infof("client %s unregistered, total clients: %d", client.ID(), len(h.clients))
	}
}

// handleBroadcast handles a broadcast message to all clients.
func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := msg.Message.ToJSON()
	if err != nil {
		logger.Log().Errorf("failed to serialize broadcast message: %v", err)
		return
	}

	for client := range h.clients {
		// Skip the sender if specified
		if msg.Sender != nil && client == msg.Sender {
			continue
		}
		client.Send(data)
	}
}

// handleRoomcast handles a message to a specific room.
func (h *Hub) handleRoomcast(msg *RoomMessage) {
	room := h.rooms.Get(msg.Room)
	if room == nil {
		return
	}

	room.Broadcast(msg.Message, msg.Sender)
}

// handleShutdown handles graceful shutdown.
func (h *Hub) handleShutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close all client connections
	for client := range h.clients {
		client.SendMessage(NewErrorMessage(ErrMsgConnectionClosed))
		client.Close()
		delete(h.clients, client)
	}

	logger.Log().Info("hub shutdown complete")
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg *OutgoingMessage, sender *Client) {
	h.broadcast <- &BroadcastMessage{
		Message: msg,
		Sender:  sender,
	}
}

// RoomBroadcast sends a message to all clients in a room.
func (h *Hub) RoomBroadcast(room string, msg *OutgoingMessage, sender *Client) {
	h.roomcast <- &RoomMessage{
		Room:    room,
		Message: msg,
		Sender:  sender,
	}
}

// JoinRoom adds a client to a room.
func (h *Hub) JoinRoom(client *Client, roomName string) error {
	if roomName == "" {
		return ErrEmptyRoom
	}

	// Check per-room connection limit
	if h.limits != nil && h.limits.MaxConnectionsPerRoom > 0 {
		room := h.rooms.Get(roomName)
		if room != nil && room.ClientCount() >= h.limits.MaxConnectionsPerRoom {
			return ErrRoomFull
		}
	}

	room := h.rooms.GetOrCreate(roomName)
	room.AddClient(client)
	client.JoinRoom(roomName)

	logger.Log().Infof("client %s joined room %s (total in room: %d)",
		client.ID(), roomName, room.ClientCount())

	return nil
}

// LeaveRoom removes a client from a room.
func (h *Hub) LeaveRoom(client *Client, roomName string) error {
	if roomName == "" {
		return ErrEmptyRoom
	}

	room := h.rooms.Get(roomName)
	if room == nil {
		return nil // Room doesn't exist, nothing to do
	}

	room.RemoveClient(client)
	client.LeaveRoom(roomName)

	// Clean up empty rooms
	h.rooms.RemoveIfEmpty(roomName)

	logger.Log().Infof("client %s left room %s", client.ID(), roomName)

	return nil
}

// Stop signals the hub to stop.
func (h *Hub) Stop() {
	close(h.done)
}

// IsRunning returns true if the hub is running.
func (h *Hub) IsRunning() bool {
	return atomic.LoadInt32(&h.running) == 1
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// RoomCount returns the number of active rooms.
func (h *Hub) RoomCount() int {
	return h.rooms.Count()
}

// RoomClientCount returns the number of clients in a specific room.
func (h *Hub) RoomClientCount(roomName string) int {
	room := h.rooms.Get(roomName)
	if room == nil {
		return 0
	}
	return room.ClientCount()
}

// GetRoomInfo returns information about a room.
func (h *Hub) GetRoomInfo(roomName string) *RoomInfoData {
	room := h.rooms.Get(roomName)
	if room == nil {
		return nil
	}

	return &RoomInfoData{
		Room:        roomName,
		ClientCount: room.ClientCount(),
		Clients:     room.ClientIDs(),
	}
}

// GetAllRooms returns information about all rooms.
func (h *Hub) GetAllRooms() []*RoomInfoData {
	rooms := h.rooms.All()
	result := make([]*RoomInfoData, 0, len(rooms))

	for name, room := range rooms {
		result = append(result, &RoomInfoData{
			Room:        name,
			ClientCount: room.ClientCount(),
		})
	}

	return result
}

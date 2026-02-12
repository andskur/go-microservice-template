package websocket

import (
	"sync"
)

// Room represents a pub/sub room/channel.
//
//nolint:govet // fieldalignment: minor optimization not worth restructuring
type Room struct {
	name    string
	clients map[*Client]bool
	mu      sync.RWMutex
}

// NewRoom creates a new room with the given name.
func NewRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[*Client]bool),
	}
}

// Name returns the room name.
func (r *Room) Name() string {
	return r.name
}

// AddClient adds a client to the room.
func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client] = true
}

// RemoveClient removes a client from the room.
func (r *Room) RemoveClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, client)
}

// HasClient checks if a client is in the room.
func (r *Room) HasClient(client *Client) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.clients[client]
}

// ClientCount returns the number of clients in the room.
func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// Clients returns a slice of all clients in the room.
func (r *Room) Clients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*Client, 0, len(r.clients))
	for client := range r.clients {
		clients = append(clients, client)
	}
	return clients
}

// ClientIDs returns a slice of all client IDs in the room.
func (r *Room) ClientIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.clients))
	for client := range r.clients {
		ids = append(ids, client.ID())
	}
	return ids
}

// Broadcast sends a message to all clients in the room except the sender.
func (r *Room) Broadcast(msg *OutgoingMessage, sender *Client) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := msg.ToJSON()
	if err != nil {
		return
	}

	for client := range r.clients {
		// Skip the sender
		if sender != nil && client == sender {
			continue
		}
		client.Send(data)
	}
}

// IsEmpty returns true if the room has no clients.
func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients) == 0
}

// RoomManager manages all rooms.
type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewRoomManager creates a new room manager.
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

// GetOrCreate gets an existing room or creates a new one.
func (rm *RoomManager) GetOrCreate(name string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[name]; exists {
		return room
	}

	room := NewRoom(name)
	rm.rooms[name] = room
	return room
}

// Get returns a room by name, or nil if it doesn't exist.
func (rm *RoomManager) Get(name string) *Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[name]
}

// Remove removes a room if it exists.
func (rm *RoomManager) Remove(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.rooms, name)
}

// RemoveIfEmpty removes a room if it exists and is empty.
func (rm *RoomManager) RemoveIfEmpty(name string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[name]
	if !exists {
		return false
	}

	if room.IsEmpty() {
		delete(rm.rooms, name)
		return true
	}
	return false
}

// Count returns the total number of rooms.
func (rm *RoomManager) Count() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.rooms)
}

// Names returns a list of all room names.
func (rm *RoomManager) Names() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	names := make([]string, 0, len(rm.rooms))
	for name := range rm.rooms {
		names = append(names, name)
	}
	return names
}

// All returns a map of all rooms.
func (rm *RoomManager) All() map[string]*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy to prevent external modification
	rooms := make(map[string]*Room, len(rm.rooms))
	for name, room := range rm.rooms {
		rooms[name] = room
	}
	return rooms
}

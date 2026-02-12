package websocket

import (
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"

	"microservice-template/pkg/logger"
)

// Client represents a single WebSocket connection.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	rooms    map[string]bool
	handlers *MessageHandlers
	id       string

	// Configuration.
	pongWait     time.Duration
	pingInterval time.Duration
	writeWait    time.Duration
	maxMsgSize   int64
	mu           sync.RWMutex
	closed       bool
}

// ClientConfig holds configuration for a client connection.
type ClientConfig struct {
	PongWait     time.Duration
	PingInterval time.Duration
	WriteWait    time.Duration
	MaxMsgSize   int64
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *websocket.Conn, handlers *MessageHandlers, cfg *ClientConfig) *Client {
	clientID := uuid.Must(uuid.NewV4()).String()

	return &Client{
		id:           clientID,
		hub:          hub,
		conn:         conn,
		send:         make(chan []byte, 256),
		rooms:        make(map[string]bool),
		handlers:     handlers,
		pongWait:     cfg.PongWait,
		pingInterval: cfg.PingInterval,
		writeWait:    cfg.WriteWait,
		maxMsgSize:   cfg.MaxMsgSize,
	}
}

// ID returns the client's unique identifier.
func (c *Client) ID() string {
	return c.id
}

// Send queues a message to be sent to the client.
func (c *Client) Send(data []byte) bool {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return false
	}
	c.mu.RUnlock()

	select {
	case c.send <- data:
		return true
	default:
		// Channel full, client is slow
		logger.Log().Warnf("client %s send buffer full, dropping message", c.id)
		return false
	}
}

// SendMessage sends an OutgoingMessage to the client.
func (c *Client) SendMessage(msg *OutgoingMessage) bool {
	data, err := msg.ToJSON()
	if err != nil {
		logger.Log().Errorf("client %s: failed to serialize message: %v", c.id, err)
		return false
	}
	return c.Send(data)
}

// Rooms returns a copy of the rooms the client is subscribed to.
func (c *Client) Rooms() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rooms := make([]string, 0, len(c.rooms))
	for room := range c.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// IsInRoom checks if the client is subscribed to a room.
func (c *Client) IsInRoom(room string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rooms[room]
}

// JoinRoom adds the client to a room.
func (c *Client) JoinRoom(room string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rooms[room] = true
}

// LeaveRoom removes the client from a room.
func (c *Client) LeaveRoom(room string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.rooms, room)
}

// LeaveAllRooms removes the client from all rooms.
func (c *Client) LeaveAllRooms() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	rooms := make([]string, 0, len(c.rooms))
	for room := range c.rooms {
		rooms = append(rooms, room)
	}
	c.rooms = make(map[string]bool)
	return rooms
}

// Close closes the client connection.
func (c *Client) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()

	close(c.send)
	c.conn.Close() //nolint:errcheck // Best effort close
}

// IsClosed returns true if the client connection is closed.
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// ReadPump pumps messages from the WebSocket connection to the hub.
// This method runs in a dedicated goroutine per client.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.Close()
	}()

	c.conn.SetReadLimit(c.maxMsgSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(c.pongWait)); err != nil {
		logger.Log().Errorf("client %s: set read deadline: %v", c.id, err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Log().Warnf("client %s: unexpected close: %v", c.id, err)
			}
			break
		}

		// Parse and handle the incoming message
		incoming, err := ParseIncomingMessage(message)
		if err != nil {
			c.SendMessage(NewErrorMessage(ErrMsgInvalidJSON))
			continue
		}

		// Process the message through handlers
		if c.handlers != nil {
			c.handlers.HandleMessage(c, incoming)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
// This method runs in a dedicated goroutine per client.
//
//nolint:gocognit // WebSocket write pump requires handling multiple cases
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.pingInterval)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
				logger.Log().Errorf("client %s: set write deadline: %v", c.id, err)
				return
			}
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{}) //nolint:errcheck // Best effort
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err = w.Write(message); err != nil {
				logger.Log().Errorf("client %s: write message: %v", c.id, err)
				return
			}

			// Add queued messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err = w.Write([]byte{'\n'}); err != nil {
					break
				}
				if _, err = w.Write(<-c.send); err != nil {
					break
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)); err != nil {
				logger.Log().Errorf("client %s: set ping deadline: %v", c.id, err)
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

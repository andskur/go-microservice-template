// Package websocket provides a WebSocket server module with pub/sub and room support.
package websocket

import (
	"encoding/json"
	"time"
)

// Message types for WebSocket communication.
const (
	// Client to server message types.
	MsgTypeSubscribe   = "subscribe"   // Subscribe to a room
	MsgTypeUnsubscribe = "unsubscribe" // Unsubscribe from a room
	MsgTypePublish     = "publish"     // Publish message to a room
	MsgTypeBroadcast   = "broadcast"   // Broadcast to all connected clients

	// Server to client message types.
	MsgTypeMessage      = "message"      // Incoming message (from room or broadcast)
	MsgTypeError        = "error"        // Error response
	MsgTypeSubscribed   = "subscribed"   // Subscription confirmation
	MsgTypeUnsubscribed = "unsubscribed" // Unsubscription confirmation
	MsgTypeConnected    = "connected"    // Connection established
	MsgTypeRoomInfo     = "room_info"    // Room information response

	// Bidirectional.
	MsgTypePing = "ping" // Ping request
	MsgTypePong = "pong" // Pong response
)

// IncomingMessage represents a message received from a client.
type IncomingMessage struct {
	Type string          `json:"type"`           // Message type (subscribe, publish, etc.)
	Room string          `json:"room,omitempty"` // Target room (for subscribe/unsubscribe/publish)
	Data json.RawMessage `json:"data,omitempty"` // Message payload
}

// OutgoingMessage represents a message sent to a client.
type OutgoingMessage struct {
	Data      interface{} `json:"data,omitempty"`      // Message payload
	Type      string      `json:"type"`                // Message type
	Room      string      `json:"room,omitempty"`      // Source room (if applicable)
	Error     string      `json:"error,omitempty"`     // Error message (if type is "error")
	ClientID  string      `json:"client_id,omitempty"` // Sender's client ID (for messages)
	Timestamp int64       `json:"timestamp"`           // Unix timestamp in milliseconds
}

// RoomMessage represents a message to be sent to a specific room.
type RoomMessage struct {
	Message *OutgoingMessage // Message to send
	Sender  *Client          // Original sender (excluded from broadcast)
	Room    string           // Target room
}

// BroadcastMessage represents a message to be broadcast to all clients.
type BroadcastMessage struct {
	Message *OutgoingMessage // Message to send
	Sender  *Client          // Original sender (excluded from broadcast)
}

// NewOutgoingMessage creates a new outgoing message with timestamp.
func NewOutgoingMessage(msgType string, data interface{}) *OutgoingMessage {
	return &OutgoingMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewErrorMessage creates an error message.
func NewErrorMessage(errMsg string) *OutgoingMessage {
	return &OutgoingMessage{
		Type:      MsgTypeError,
		Error:     errMsg,
		Timestamp: time.Now().UnixMilli(),
	}
}

// NewRoomMessage creates a new room message.
func NewRoomMessage(msgType, room string, data interface{}) *OutgoingMessage {
	return &OutgoingMessage{
		Type:      msgType,
		Room:      room,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

// ToJSON serializes the outgoing message to JSON bytes.
func (m *OutgoingMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// ParseIncomingMessage parses raw JSON bytes into an IncomingMessage.
func ParseIncomingMessage(data []byte) (*IncomingMessage, error) {
	var msg IncomingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// ConnectedData represents the data sent on successful connection.
type ConnectedData struct {
	ClientID string `json:"client_id"` // Assigned client ID
	Message  string `json:"message"`   // Welcome message
}

// SubscribedData represents the data sent on successful subscription.
type SubscribedData struct {
	Room        string `json:"room"`         // Room name
	ClientCount int    `json:"client_count"` // Number of clients in the room
}

// UnsubscribedData represents the data sent on successful unsubscription.
type UnsubscribedData struct {
	Room string `json:"room"` // Room name
}

// RoomInfoData represents room information.
//
//nolint:govet // fieldalignment: minor optimization not worth restructuring JSON field order
type RoomInfoData struct {
	Room        string   `json:"room"`         // Room name
	ClientCount int      `json:"client_count"` // Number of clients in the room
	Clients     []string `json:"clients"`      // List of client IDs (optional)
}

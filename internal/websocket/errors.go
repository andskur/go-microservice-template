package websocket

import "errors"

// WebSocket-specific errors.
var (
	// ErrMaxConnectionsReached is returned when the maximum number of global connections is reached.
	ErrMaxConnectionsReached = errors.New("maximum number of connections reached")

	// ErrRoomFull is returned when a room has reached its maximum capacity.
	ErrRoomFull = errors.New("room has reached maximum capacity")

	// ErrNotSubscribed is returned when trying to publish to a room the client is not subscribed to.
	ErrNotSubscribed = errors.New("not subscribed to room")

	// ErrInvalidMessage is returned when a message cannot be parsed.
	ErrInvalidMessage = errors.New("invalid message format")

	// ErrUnknownMessageType is returned when the message type is not recognized.
	ErrUnknownMessageType = errors.New("unknown message type")

	// ErrEmptyRoom is returned when a room name is empty.
	ErrEmptyRoom = errors.New("room name cannot be empty")

	// ErrHubNotRunning is returned when the hub is not running.
	ErrHubNotRunning = errors.New("hub is not running")

	// ErrClientNotFound is returned when the client is not found.
	ErrClientNotFound = errors.New("client not found")

	// ErrServerNotRunning is returned when the server is not running.
	ErrServerNotRunning = errors.New("server is not running")
)

// Error messages for client responses.
const (
	ErrMsgInvalidJSON       = "invalid JSON format"
	ErrMsgUnknownType       = "unknown message type"
	ErrMsgEmptyRoom         = "room name is required"
	ErrMsgNotSubscribed     = "not subscribed to this room"
	ErrMsgRoomFull          = "room is at maximum capacity"
	ErrMsgMaxConnections    = "server is at maximum capacity"
	ErrMsgInternalError     = "internal server error"
	ErrMsgConnectionClosed  = "connection closed"
	ErrMsgMessageTooLarge   = "message exceeds maximum size"
	ErrMsgRateLimitExceeded = "rate limit exceeded"
)

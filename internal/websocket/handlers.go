package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"microservice-template/internal/service"
	"microservice-template/pkg/logger"
)

// MessageHandlers handles incoming WebSocket messages.
type MessageHandlers struct {
	hub     *Hub
	service service.IService
}

// NewMessageHandlers creates a new MessageHandlers instance.
func NewMessageHandlers(hub *Hub, svc service.IService) *MessageHandlers {
	return &MessageHandlers{
		hub:     hub,
		service: svc,
	}
}

// HandleMessage routes an incoming message to the appropriate handler.
func (h *MessageHandlers) HandleMessage(client *Client, msg *IncomingMessage) {
	switch msg.Type {
	case MsgTypeSubscribe:
		h.handleSubscribe(client, msg)
	case MsgTypeUnsubscribe:
		h.handleUnsubscribe(client, msg)
	case MsgTypePublish:
		h.handlePublish(client, msg)
	case MsgTypeBroadcast:
		h.handleBroadcast(client, msg)
	case MsgTypePing:
		h.handlePing(client)
	default:
		client.SendMessage(NewErrorMessage(ErrMsgUnknownType))
	}
}

// handleSubscribe handles a subscribe request.
func (h *MessageHandlers) handleSubscribe(client *Client, msg *IncomingMessage) {
	if msg.Room == "" {
		client.SendMessage(NewErrorMessage(ErrMsgEmptyRoom))
		return
	}

	// Check if already subscribed
	if client.IsInRoom(msg.Room) {
		client.SendMessage(NewRoomMessage(MsgTypeSubscribed, msg.Room, &SubscribedData{
			Room:        msg.Room,
			ClientCount: h.hub.RoomClientCount(msg.Room),
		}))
		return
	}

	// Join the room
	if err := h.hub.JoinRoom(client, msg.Room); err != nil {
		if errors.Is(err, ErrRoomFull) {
			client.SendMessage(NewErrorMessage(ErrMsgRoomFull))
		} else {
			client.SendMessage(NewErrorMessage(ErrMsgInternalError))
		}
		return
	}

	// Send confirmation
	client.SendMessage(NewRoomMessage(MsgTypeSubscribed, msg.Room, &SubscribedData{
		Room:        msg.Room,
		ClientCount: h.hub.RoomClientCount(msg.Room),
	}))
}

// handleUnsubscribe handles an unsubscribe request.
func (h *MessageHandlers) handleUnsubscribe(client *Client, msg *IncomingMessage) {
	if msg.Room == "" {
		client.SendMessage(NewErrorMessage(ErrMsgEmptyRoom))
		return
	}

	// Check if subscribed
	if !client.IsInRoom(msg.Room) {
		client.SendMessage(NewErrorMessage(ErrMsgNotSubscribed))
		return
	}

	// Leave the room
	if err := h.hub.LeaveRoom(client, msg.Room); err != nil {
		client.SendMessage(NewErrorMessage(ErrMsgInternalError))
		return
	}

	// Send confirmation
	client.SendMessage(NewRoomMessage(MsgTypeUnsubscribed, msg.Room, &UnsubscribedData{
		Room: msg.Room,
	}))
}

// handlePublish handles a publish request to a room.
func (h *MessageHandlers) handlePublish(client *Client, msg *IncomingMessage) {
	if msg.Room == "" {
		client.SendMessage(NewErrorMessage(ErrMsgEmptyRoom))
		return
	}

	// Check if client is subscribed to the room
	if !client.IsInRoom(msg.Room) {
		client.SendMessage(NewErrorMessage(ErrMsgNotSubscribed))
		return
	}

	// Parse the data if needed for business logic
	// For now, we just relay the message to the room

	// Create outgoing message
	outMsg := &OutgoingMessage{
		Type:      MsgTypeMessage,
		Room:      msg.Room,
		Data:      msg.Data,
		ClientID:  client.ID(),
		Timestamp: time.Now().UnixMilli(),
	}

	// Broadcast to the room (excluding sender)
	h.hub.RoomBroadcast(msg.Room, outMsg, client)

	logger.Log().Debugf("client %s published to room %s", client.ID(), msg.Room)
}

// handleBroadcast handles a broadcast request to all clients.
func (h *MessageHandlers) handleBroadcast(client *Client, msg *IncomingMessage) {
	// Create outgoing message
	outMsg := &OutgoingMessage{
		Type:      MsgTypeMessage,
		Data:      msg.Data,
		ClientID:  client.ID(),
		Timestamp: time.Now().UnixMilli(),
	}

	// Broadcast to all clients (excluding sender)
	h.hub.Broadcast(outMsg, client)

	logger.Log().Debugf("client %s broadcast to all", client.ID())
}

// handlePing handles a ping request.
func (h *MessageHandlers) handlePing(client *Client) {
	client.SendMessage(NewOutgoingMessage(MsgTypePong, nil))
}

// Service returns the service layer for custom handlers.
func (h *MessageHandlers) Service() service.IService {
	return h.service
}

// Hub returns the hub for custom handlers.
func (h *MessageHandlers) Hub() *Hub {
	return h.hub
}

// CustomHandler is a function type for custom message handlers.
type CustomHandler func(ctx context.Context, client *Client, data json.RawMessage) error

// customHandlers stores registered custom handlers.
var customHandlers = make(map[string]CustomHandler)

// RegisterCustomHandler registers a custom message handler.
func RegisterCustomHandler(msgType string, handler CustomHandler) {
	customHandlers[msgType] = handler
}

// HandleCustomMessage handles a custom message type.
// This allows extending the WebSocket functionality with business-specific handlers.
func (h *MessageHandlers) HandleCustomMessage(ctx context.Context, client *Client, msgType string, data json.RawMessage) error {
	handler, exists := customHandlers[msgType]
	if !exists {
		return ErrUnknownMessageType
	}
	return handler(ctx, client, data)
}

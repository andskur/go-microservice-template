# WebSocket Module Guide

This guide explains how to use the WebSocket server module in the microservice template. The module provides real-time bidirectional communication with pub/sub support and room management.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [Architecture](#architecture)
5. [Message Protocol](#message-protocol)
6. [Client Integration](#client-integration)
7. [Service Layer Integration](#service-layer-integration)
8. [Custom Message Handlers](#custom-message-handlers)
9. [Testing](#testing)
10. [Production Considerations](#production-considerations)

## Overview

The WebSocket module provides:
- **Real-time communication**: Bidirectional WebSocket connections using `gorilla/websocket`
- **Pub/Sub with rooms**: Subscribe to channels/rooms for targeted message delivery
- **Broadcasting**: Send messages to all connected clients or specific rooms
- **Connection management**: Central hub manages all client connections
- **Connection limits**: Configurable global and per-room connection limits
- **Health checks**: Built-in health endpoint for monitoring
- **Service integration**: Access to the service layer for business logic

### Architecture

```
                                    ┌─────────────────┐
                                    │   HTTP Server   │
                                    │   (port 8081)   │
                                    └────────┬────────┘
                                             │
                              ┌──────────────┼──────────────┐
                              │              │              │
                         /ws endpoint   /health        (other)
                              │              │
                    ┌─────────▼─────────┐    │
                    │     Upgrader      │    │
                    │  HTTP → WebSocket │    │
                    └─────────┬─────────┘    │
                              │              │
                    ┌─────────▼─────────┐    │
                    │       Hub         │◄───┘
                    │  (Central Coord)  │
                    └─────────┬─────────┘
                              │
           ┌──────────────────┼──────────────────┐
           │                  │                  │
    ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐
    │   Client    │    │   Client    │    │   Client    │
    │  (conn 1)   │    │  (conn 2)   │    │  (conn N)   │
    └─────────────┘    └─────────────┘    └─────────────┘
           │                  │                  │
           │                  ▼                  │
           │         ┌───────────────┐          │
           └────────►│  Room Manager │◄─────────┘
                     │  (pub/sub)    │
                     └───────────────┘
```

## Quick Start

### 1. Enable WebSocket Module

```bash
export WEBSOCKET_ENABLED=true
export WEBSOCKET_PORT=8081
```

### 2. Run the Service

```bash
make run
```

### 3. Test Connection

```bash
# Install wscat
npm install -g wscat

# Connect
wscat -c ws://localhost:8081/ws
```

### 4. Test Pub/Sub

```bash
# Terminal 1: Connect client 1
wscat -c ws://localhost:8081/ws
> {"type":"subscribe","room":"chat"}

# Terminal 2: Connect client 2
wscat -c ws://localhost:8081/ws
> {"type":"subscribe","room":"chat"}
> {"type":"publish","room":"chat","data":{"message":"Hello!"}}

# Client 1 receives the message
```

## Configuration

### Full Configuration Example

```yaml
websocket:
  enabled: true
  host: "0.0.0.0"
  port: 8081
  timeout: "30s"
  read_buffer_size: 1024
  write_buffer_size: 1024
  max_message_size: 512000  # 500KB
  ping_interval: "54s"
  pong_wait: "60s"
  write_wait: "10s"
  
  limits:
    max_connections: 10000       # 0 = unlimited
    max_connections_per_room: 1000  # 0 = unlimited
```

### Environment Variables

```bash
WEBSOCKET_ENABLED=true
WEBSOCKET_HOST=0.0.0.0
WEBSOCKET_PORT=8081
WEBSOCKET_TIMEOUT=30s
WEBSOCKET_READ_BUFFER_SIZE=1024
WEBSOCKET_WRITE_BUFFER_SIZE=1024
WEBSOCKET_MAX_MESSAGE_SIZE=512000
WEBSOCKET_PING_INTERVAL=54s
WEBSOCKET_PONG_WAIT=60s
WEBSOCKET_WRITE_WAIT=10s
WEBSOCKET_LIMITS_MAX_CONNECTIONS=10000
WEBSOCKET_LIMITS_MAX_CONNECTIONS_PER_ROOM=1000
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | false | Enable WebSocket module |
| `host` | string | "0.0.0.0" | Server host interface |
| `port` | int | 8081 | Server port |
| `timeout` | string | "30s" | Connection timeout |
| `read_buffer_size` | int | 1024 | Read buffer size in bytes |
| `write_buffer_size` | int | 1024 | Write buffer size in bytes |
| `max_message_size` | int64 | 512000 | Maximum message size (500KB) |
| `ping_interval` | string | "54s" | Ping keepalive interval |
| `pong_wait` | string | "60s" | Pong response timeout |
| `write_wait` | string | "10s" | Write deadline |
| `limits.max_connections` | int | 0 | Global max connections (0 = unlimited) |
| `limits.max_connections_per_room` | int | 0 | Per-room limit (0 = unlimited) |

## Message Protocol

All messages are JSON-formatted.

### Client → Server Messages

#### Subscribe to Room
```json
{
  "type": "subscribe",
  "room": "room-name"
}
```

#### Unsubscribe from Room
```json
{
  "type": "unsubscribe",
  "room": "room-name"
}
```

#### Publish to Room
```json
{
  "type": "publish",
  "room": "room-name",
  "data": { "any": "payload" }
}
```
Note: Client must be subscribed to the room to publish.

#### Broadcast to All
```json
{
  "type": "broadcast",
  "data": { "any": "payload" }
}
```

#### Ping
```json
{
  "type": "ping"
}
```

### Server → Client Messages

#### Connected
```json
{
  "type": "connected",
  "data": {
    "client_id": "uuid-string",
    "message": "Connected to WebSocket server"
  },
  "timestamp": 1234567890123
}
```

#### Subscribed
```json
{
  "type": "subscribed",
  "room": "room-name",
  "data": {
    "room": "room-name",
    "client_count": 5
  },
  "timestamp": 1234567890123
}
```

#### Unsubscribed
```json
{
  "type": "unsubscribed",
  "room": "room-name",
  "data": {
    "room": "room-name"
  },
  "timestamp": 1234567890123
}
```

#### Message (from room or broadcast)
```json
{
  "type": "message",
  "room": "room-name",
  "data": { "any": "payload" },
  "client_id": "sender-uuid",
  "timestamp": 1234567890123
}
```

#### Error
```json
{
  "type": "error",
  "error": "error description",
  "timestamp": 1234567890123
}
```

#### Pong
```json
{
  "type": "pong",
  "timestamp": 1234567890123
}
```

### Error Messages

| Error | Description |
|-------|-------------|
| `invalid JSON format` | Message could not be parsed |
| `unknown message type` | Message type not recognized |
| `room name is required` | Room name was empty |
| `not subscribed to this room` | Tried to publish without subscription |
| `room is at maximum capacity` | Room connection limit reached |
| `server is at maximum capacity` | Global connection limit reached |

## Client Integration

### JavaScript/Browser

```javascript
class WebSocketClient {
  constructor(url) {
    this.url = url;
    this.ws = null;
    this.handlers = new Map();
  }

  connect() {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(this.url);
      
      this.ws.onopen = () => resolve(this);
      this.ws.onerror = (err) => reject(err);
      
      this.ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        const handler = this.handlers.get(msg.type);
        if (handler) handler(msg);
      };
    });
  }

  on(type, handler) {
    this.handlers.set(type, handler);
    return this;
  }

  subscribe(room) {
    this.ws.send(JSON.stringify({ type: 'subscribe', room }));
  }

  unsubscribe(room) {
    this.ws.send(JSON.stringify({ type: 'unsubscribe', room }));
  }

  publish(room, data) {
    this.ws.send(JSON.stringify({ type: 'publish', room, data }));
  }

  broadcast(data) {
    this.ws.send(JSON.stringify({ type: 'broadcast', data }));
  }

  close() {
    this.ws.close();
  }
}

// Usage
const client = new WebSocketClient('ws://localhost:8081/ws');

client
  .on('connected', (msg) => {
    console.log('Connected:', msg.data.client_id);
    client.subscribe('notifications');
  })
  .on('subscribed', (msg) => {
    console.log('Subscribed to:', msg.room);
  })
  .on('message', (msg) => {
    console.log('Message from', msg.room, ':', msg.data);
  })
  .on('error', (msg) => {
    console.error('Error:', msg.error);
  });

await client.connect();
```

### Go Client

```go
package main

import (
    "encoding/json"
    "log"
    "github.com/gorilla/websocket"
)

type Message struct {
    Type      string      `json:"type"`
    Room      string      `json:"room,omitempty"`
    Data      interface{} `json:"data,omitempty"`
    Error     string      `json:"error,omitempty"`
    Timestamp int64       `json:"timestamp,omitempty"`
}

func main() {
    conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8081/ws", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Read connected message
    _, data, _ := conn.ReadMessage()
    var connected Message
    json.Unmarshal(data, &connected)
    log.Printf("Connected: %v", connected.Data)

    // Subscribe to room
    sub := Message{Type: "subscribe", Room: "updates"}
    conn.WriteJSON(sub)

    // Read messages
    for {
        _, data, err := conn.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            break
        }
        var msg Message
        json.Unmarshal(data, &msg)
        log.Printf("Received: %+v", msg)
    }
}
```

## Service Layer Integration

The WebSocket module has access to the service layer, allowing handlers to perform business logic.

### Accessing Service in Custom Handlers

```go
// internal/websocket/handlers.go

func (h *MessageHandlers) HandleCustomMessage(ctx context.Context, client *Client, data json.RawMessage) error {
    // Access service layer
    user, err := h.service.GetUserByEmail(ctx, "test@example.com")
    if err != nil {
        return err
    }
    
    // Send response to client
    client.SendMessage(NewOutgoingMessage("user_data", user))
    return nil
}
```

### Broadcasting from Service Layer

You can broadcast messages from anywhere in your application by accessing the WebSocket module:

```go
// In your application code
func (app *App) NotifyAllClients(event string, data interface{}) {
    if app.wsModule != nil {
        msg := websocket.NewOutgoingMessage("notification", map[string]interface{}{
            "event": event,
            "data":  data,
        })
        app.wsModule.Broadcast(msg)
    }
}

// Broadcast to specific room
func (app *App) NotifyRoom(room string, data interface{}) {
    if app.wsModule != nil {
        msg := websocket.NewOutgoingMessage("room_update", data)
        app.wsModule.RoomBroadcast(room, msg)
    }
}
```

## Custom Message Handlers

### Registering Custom Handlers

```go
// internal/websocket/handlers.go

func init() {
    // Register custom handler for "get_user" message type
    RegisterCustomHandler("get_user", handleGetUser)
}

func handleGetUser(ctx context.Context, client *Client, data json.RawMessage) error {
    // Parse request
    var req struct {
        Email string `json:"email"`
    }
    if err := json.Unmarshal(data, &req); err != nil {
        client.SendMessage(NewErrorMessage("invalid request"))
        return nil
    }

    // Get service from handlers (you'll need to store reference)
    // user, err := service.GetUserByEmail(ctx, req.Email)
    
    // Send response
    client.SendMessage(NewOutgoingMessage("user_response", map[string]interface{}{
        "email": req.Email,
        // "user": user,
    }))
    
    return nil
}
```

### Extending HandleMessage

You can modify `HandleMessage` in `handlers.go` to support additional message types:

```go
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
    
    // Add custom handlers
    case "get_user":
        h.handleGetUser(client, msg)
    case "create_order":
        h.handleCreateOrder(client, msg)
    
    default:
        client.SendMessage(NewErrorMessage(ErrMsgUnknownType))
    }
}
```

## Testing

### Run WebSocket Tests

```bash
# Run all WebSocket tests
go test -v ./internal/websocket/...

# Run specific test
go test -v ./internal/websocket/... -run TestWebSocket_Integration
```

### Testing with wscat

```bash
# Basic connection test
wscat -c ws://localhost:8081/ws

# With headers
wscat -c ws://localhost:8081/ws -H "Authorization: Bearer token"
```

### Testing with curl (health endpoint)

```bash
curl http://localhost:8081/health
# {"status":"healthy","clients":0,"rooms":0}
```

### Unit Testing Handlers

```go
func TestCustomHandler(t *testing.T) {
    cfg := &config.WebSocketConfig{
        Host:            "127.0.0.1",
        Port:            0,  // Random port
        PongWait:        "60s",
        PingInterval:    "54s",
        WriteWait:       "10s",
        MaxMessageSize:  512000,
        Enabled:         true,
    }
    
    // Create module with mock service
    mockSvc := &mockService{}
    mod := NewModule(cfg, mockSvc)
    
    // Test your handlers...
}
```

## Production Considerations

### Connection Limits

Set appropriate limits based on your infrastructure:

```yaml
websocket:
  limits:
    max_connections: 50000        # Based on server memory
    max_connections_per_room: 5000  # Based on use case
```

### Memory Considerations

Each connection uses:
- ~20KB base memory
- + message buffer sizes (`read_buffer_size` + `write_buffer_size`)
- + send channel buffer (256 messages * avg message size)

For 10,000 connections with default settings: ~200-300MB

### Keepalive Settings

Tune keepalive for your network:

```yaml
websocket:
  ping_interval: "30s"   # Shorter for unstable networks
  pong_wait: "35s"       # Always > ping_interval
  write_wait: "10s"
```

### Load Balancing

When running multiple instances behind a load balancer:

1. **Sticky sessions**: Required for WebSocket connections
2. **Health checks**: Use `/health` endpoint
3. **Graceful shutdown**: Module handles connection draining

### Monitoring

Key metrics to track:
- `websocket_clients_total` - Total connected clients
- `websocket_rooms_total` - Number of active rooms
- `websocket_messages_per_second` - Message throughput
- `websocket_errors_total` - Error count

### Security

1. **Origin checking**: Implement in upgrader's `CheckOrigin`
2. **Rate limiting**: Add per-client message rate limits
3. **Message validation**: Validate all incoming message data
4. **Authentication**: Add token validation before upgrade (future enhancement)

### Example Production Config

```yaml
websocket:
  enabled: true
  host: "0.0.0.0"
  port: 8081
  timeout: "30s"
  read_buffer_size: 2048
  write_buffer_size: 2048
  max_message_size: 65536  # 64KB
  ping_interval: "30s"
  pong_wait: "35s"
  write_wait: "10s"
  
  limits:
    max_connections: 50000
    max_connections_per_room: 5000
```

## Troubleshooting

### Connection Drops

**Issue**: Clients disconnect unexpectedly

**Solutions**:
1. Check `ping_interval` and `pong_wait` settings
2. Verify network allows WebSocket upgrade
3. Check proxy timeout settings (nginx: `proxy_read_timeout`)

### High Memory Usage

**Issue**: Memory grows with connections

**Solutions**:
1. Reduce buffer sizes
2. Limit max connections
3. Monitor message sizes

### Messages Not Delivered

**Issue**: Published messages not reaching subscribers

**Solutions**:
1. Verify client is subscribed to room
2. Check for send buffer overflow (slow clients)
3. Review logs for errors

## Further Reading

- [gorilla/websocket Documentation](https://pkg.go.dev/github.com/gorilla/websocket)
- [RFC 6455 - WebSocket Protocol](https://datatracker.ietf.org/doc/html/rfc6455)
- [MODULE_DEVELOPMENT.md](./MODULE_DEVELOPMENT.md) - Module patterns

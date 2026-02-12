package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"microservice-template/config"
)

// testConfig creates a test WebSocket configuration.
func testConfig(port int) *config.WebSocketConfig {
	return &config.WebSocketConfig{
		Host:            "127.0.0.1",
		Port:            port,
		Timeout:         "5s",
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		MaxMessageSize:  512000,
		PingInterval:    "54s",
		PongWait:        "60s",
		WriteWait:       "10s",
		Enabled:         true,
		Limits: &config.WSLimitsConfig{
			MaxConnections:        10,
			MaxConnectionsPerRoom: 5,
		},
	}
}

func TestModule_Name(t *testing.T) {
	mod := NewModule(testConfig(0), nil)
	if mod.Name() != "websocket" {
		t.Errorf("expected name 'websocket', got '%s'", mod.Name())
	}
}

func TestModule_Lifecycle(t *testing.T) {
	cfg := testConfig(18081)
	mod := NewModule(cfg, nil)
	ctx := context.Background()

	// Test Init
	if err := mod.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Test Start
	if err := mod.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test HealthCheck
	if err := mod.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}

	// Test counters
	if mod.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", mod.ClientCount())
	}
	if mod.RoomCount() != 0 {
		t.Errorf("expected 0 rooms, got %d", mod.RoomCount())
	}

	// Test Stop
	if err := mod.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestModule_HealthCheckBeforeInit(t *testing.T) {
	mod := NewModule(testConfig(0), nil)
	ctx := context.Background()

	// HealthCheck should fail before Init
	if err := mod.HealthCheck(ctx); !errors.Is(err, ErrServerNotRunning) {
		t.Errorf("expected ErrServerNotRunning, got %v", err)
	}
}

func TestHub_Basic(t *testing.T) {
	limits := &config.WSLimitsConfig{
		MaxConnections:        10,
		MaxConnectionsPerRoom: 5,
	}
	hub := NewHub(limits)

	// Start hub in goroutine
	go hub.Run()
	defer hub.Stop()

	// Wait for hub to start
	time.Sleep(50 * time.Millisecond)

	if !hub.IsRunning() {
		t.Error("hub should be running")
	}

	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", hub.ClientCount())
	}

	if hub.RoomCount() != 0 {
		t.Errorf("expected 0 rooms, got %d", hub.RoomCount())
	}
}

func TestRoom_Basic(t *testing.T) {
	room := NewRoom("test-room")

	if room.Name() != "test-room" {
		t.Errorf("expected room name 'test-room', got '%s'", room.Name())
	}

	if !room.IsEmpty() {
		t.Error("new room should be empty")
	}

	if room.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", room.ClientCount())
	}
}

func TestRoomManager(t *testing.T) {
	rm := NewRoomManager()

	// Test GetOrCreate
	room1 := rm.GetOrCreate("room1")
	if room1 == nil {
		t.Fatal("GetOrCreate returned nil")
	}

	// Get same room again
	room1Again := rm.GetOrCreate("room1")
	if room1 != room1Again {
		t.Error("GetOrCreate should return same room for same name")
	}

	// Test Count
	if rm.Count() != 1 {
		t.Errorf("expected 1 room, got %d", rm.Count())
	}

	// Create another room
	rm.GetOrCreate("room2")
	if rm.Count() != 2 {
		t.Errorf("expected 2 rooms, got %d", rm.Count())
	}

	// Test Names
	names := rm.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}

	// Test Remove
	rm.Remove("room1")
	if rm.Count() != 1 {
		t.Errorf("expected 1 room after remove, got %d", rm.Count())
	}
}

func TestMessage_NewOutgoingMessage(t *testing.T) {
	msg := NewOutgoingMessage(MsgTypeMessage, map[string]string{"key": "value"})

	if msg.Type != MsgTypeMessage {
		t.Errorf("expected type '%s', got '%s'", MsgTypeMessage, msg.Type)
	}

	if msg.Timestamp == 0 {
		t.Error("timestamp should be set")
	}
}

func TestMessage_NewErrorMessage(t *testing.T) {
	msg := NewErrorMessage("test error")

	if msg.Type != MsgTypeError {
		t.Errorf("expected type '%s', got '%s'", MsgTypeError, msg.Type)
	}

	if msg.Error != "test error" {
		t.Errorf("expected error 'test error', got '%s'", msg.Error)
	}
}

func TestMessage_NewRoomMessage(t *testing.T) {
	msg := NewRoomMessage(MsgTypeSubscribed, "test-room", nil)

	if msg.Type != MsgTypeSubscribed {
		t.Errorf("expected type '%s', got '%s'", MsgTypeSubscribed, msg.Type)
	}

	if msg.Room != "test-room" {
		t.Errorf("expected room 'test-room', got '%s'", msg.Room)
	}
}

func TestMessage_ToJSON(t *testing.T) {
	msg := NewOutgoingMessage(MsgTypeMessage, "hello")

	data, err := msg.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var parsed OutgoingMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if parsed.Type != MsgTypeMessage {
		t.Errorf("expected type '%s', got '%s'", MsgTypeMessage, parsed.Type)
	}
}

func TestMessage_ParseIncomingMessage(t *testing.T) {
	data := []byte(`{"type":"subscribe","room":"test-room"}`)

	msg, err := ParseIncomingMessage(data)
	if err != nil {
		t.Fatalf("ParseIncomingMessage failed: %v", err)
	}

	if msg.Type != MsgTypeSubscribe {
		t.Errorf("expected type '%s', got '%s'", MsgTypeSubscribe, msg.Type)
	}

	if msg.Room != "test-room" {
		t.Errorf("expected room 'test-room', got '%s'", msg.Room)
	}
}

func TestMessage_ParseIncomingMessage_Invalid(t *testing.T) {
	data := []byte(`invalid json`)

	_, err := ParseIncomingMessage(data)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestWebSocket_Integration tests actual WebSocket connection.
func TestWebSocket_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := testConfig(18082)
	mod := NewModule(cfg, nil)
	ctx := context.Background()

	// Start module
	if err := mod.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := mod.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer mod.Stop(ctx)

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test health endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:18082/health", nil)
	if err != nil {
		t.Fatalf("create health check request failed: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("health check request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Connect WebSocket client
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:18082", Path: "/ws"}
	conn, resp2, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	if resp2 != nil && resp2.Body != nil {
		defer resp2.Body.Close()
	}
	defer conn.Close()

	// Wait for connection message
	time.Sleep(50 * time.Millisecond)

	// Should receive connected message
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read connected message failed: %v", err)
	}

	var connMsg OutgoingMessage
	if err := json.Unmarshal(message, &connMsg); err != nil {
		t.Fatalf("unmarshal connected message failed: %v", err)
	}

	if connMsg.Type != MsgTypeConnected {
		t.Errorf("expected type '%s', got '%s'", MsgTypeConnected, connMsg.Type)
	}

	// Check client count
	if mod.ClientCount() != 1 {
		t.Errorf("expected 1 client, got %d", mod.ClientCount())
	}

	// Subscribe to room
	subMsg := IncomingMessage{Type: MsgTypeSubscribe, Room: "test-room"}
	subData, _ := json.Marshal(subMsg)
	if err := conn.WriteMessage(websocket.TextMessage, subData); err != nil {
		t.Fatalf("write subscribe failed: %v", err)
	}

	// Wait for subscribe confirmation
	time.Sleep(50 * time.Millisecond)

	_, message, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read subscribed message failed: %v", err)
	}

	var subResp OutgoingMessage
	if err := json.Unmarshal(message, &subResp); err != nil {
		t.Fatalf("unmarshal subscribed message failed: %v", err)
	}

	if subResp.Type != MsgTypeSubscribed {
		t.Errorf("expected type '%s', got '%s'", MsgTypeSubscribed, subResp.Type)
	}

	if subResp.Room != "test-room" {
		t.Errorf("expected room 'test-room', got '%s'", subResp.Room)
	}

	// Check room count
	if mod.RoomCount() != 1 {
		t.Errorf("expected 1 room, got %d", mod.RoomCount())
	}

	// Unsubscribe
	unsubMsg := IncomingMessage{Type: MsgTypeUnsubscribe, Room: "test-room"}
	unsubData, _ := json.Marshal(unsubMsg)
	if err := conn.WriteMessage(websocket.TextMessage, unsubData); err != nil {
		t.Fatalf("write unsubscribe failed: %v", err)
	}

	// Wait for unsubscribe confirmation
	time.Sleep(50 * time.Millisecond)

	_, message, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read unsubscribed message failed: %v", err)
	}

	var unsubResp OutgoingMessage
	if err := json.Unmarshal(message, &unsubResp); err != nil {
		t.Fatalf("unmarshal unsubscribed message failed: %v", err)
	}

	if unsubResp.Type != MsgTypeUnsubscribed {
		t.Errorf("expected type '%s', got '%s'", MsgTypeUnsubscribed, unsubResp.Type)
	}

	// Room should be cleaned up
	time.Sleep(50 * time.Millisecond)
	if mod.RoomCount() != 0 {
		t.Errorf("expected 0 rooms after unsubscribe, got %d", mod.RoomCount())
	}
}

// TestWebSocket_ErrorHandling tests error handling.
func TestWebSocket_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := testConfig(18083)
	mod := NewModule(cfg, nil)
	ctx := context.Background()

	if err := mod.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := mod.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer mod.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	// Connect
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:18083", Path: "/ws"}
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	defer conn.Close()

	// Read connected message
	_, _, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read connected message failed: %v", err)
	}

	// Send invalid JSON
	if err := conn.WriteMessage(websocket.TextMessage, []byte("invalid json")); err != nil {
		t.Fatalf("write invalid json failed: %v", err)
	}

	// Should receive error
	time.Sleep(50 * time.Millisecond)
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error message failed: %v", err)
	}

	var errResp OutgoingMessage
	if err := json.Unmarshal(message, &errResp); err != nil {
		t.Fatalf("unmarshal error message failed: %v", err)
	}

	if errResp.Type != MsgTypeError {
		t.Errorf("expected type '%s', got '%s'", MsgTypeError, errResp.Type)
	}

	if !strings.Contains(errResp.Error, "invalid") {
		t.Errorf("expected error to contain 'invalid', got '%s'", errResp.Error)
	}

	// Send empty room subscribe
	emptyRoomMsg := IncomingMessage{Type: MsgTypeSubscribe, Room: ""}
	emptyRoomData, _ := json.Marshal(emptyRoomMsg)
	if err := conn.WriteMessage(websocket.TextMessage, emptyRoomData); err != nil {
		t.Fatalf("write empty room failed: %v", err)
	}

	// Should receive error
	time.Sleep(50 * time.Millisecond)
	_, message, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read empty room error failed: %v", err)
	}

	if err := json.Unmarshal(message, &errResp); err != nil {
		t.Fatalf("unmarshal empty room error failed: %v", err)
	}

	if errResp.Type != MsgTypeError {
		t.Errorf("expected type '%s', got '%s'", MsgTypeError, errResp.Type)
	}

	// Publish without subscription
	pubMsg := IncomingMessage{Type: MsgTypePublish, Room: "nonexistent", Data: json.RawMessage(`"test"`)}
	pubData, _ := json.Marshal(pubMsg)
	if err := conn.WriteMessage(websocket.TextMessage, pubData); err != nil {
		t.Fatalf("write publish failed: %v", err)
	}

	// Should receive not subscribed error
	time.Sleep(50 * time.Millisecond)
	_, message, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read not subscribed error failed: %v", err)
	}

	if err := json.Unmarshal(message, &errResp); err != nil {
		t.Fatalf("unmarshal not subscribed error failed: %v", err)
	}

	if errResp.Type != MsgTypeError {
		t.Errorf("expected type '%s', got '%s'", MsgTypeError, errResp.Type)
	}
}

// TestWebSocket_Ping tests ping/pong functionality.
func TestWebSocket_Ping(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := testConfig(18084)
	mod := NewModule(cfg, nil)
	ctx := context.Background()

	if err := mod.Init(ctx); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if err := mod.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer mod.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	// Connect
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:18084", Path: "/ws"}
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	defer conn.Close()

	// Read connected message
	_, _, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read connected message failed: %v", err)
	}

	// Send ping
	pingMsg := IncomingMessage{Type: MsgTypePing}
	pingData, _ := json.Marshal(pingMsg)
	if err := conn.WriteMessage(websocket.TextMessage, pingData); err != nil {
		t.Fatalf("write ping failed: %v", err)
	}

	// Should receive pong
	time.Sleep(50 * time.Millisecond)
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read pong message failed: %v", err)
	}

	var pongResp OutgoingMessage
	if err := json.Unmarshal(message, &pongResp); err != nil {
		t.Fatalf("unmarshal pong message failed: %v", err)
	}

	if pongResp.Type != MsgTypePong {
		t.Errorf("expected type '%s', got '%s'", MsgTypePong, pongResp.Type)
	}
}

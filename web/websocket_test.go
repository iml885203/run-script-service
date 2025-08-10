package web

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketHub_Creation(t *testing.T) {
	hub := NewWebSocketHub()

	// Hub should be initialized with proper channels
	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
}

func TestWebSocketHub_ClientManagement(t *testing.T) {
	hub := NewWebSocketHub()

	// Test that we can create clients and manage them
	client := &WebSocketClient{
		hub:  hub,
		send: make(chan []byte, 256),
	}

	assert.NotNil(t, client)
	assert.Equal(t, hub, client.hub)
	assert.NotNil(t, client.send)
}

func TestWebSocketMessage_Broadcasting(t *testing.T) {
	hub := NewWebSocketHub()

	// Test that we can send messages to broadcast channel
	testMessage := WebSocketMessage{
		Type:      "test",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"message": "hello world",
		},
	}

	messageBytes, err := json.Marshal(testMessage)
	require.NoError(t, err)
	assert.NotNil(t, messageBytes)

	// Test that hub broadcast channel can receive messages
	assert.NotNil(t, hub.broadcast)
}

func TestWebSocketMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name    string
		message WebSocketMessage
	}{
		{
			name: "script status message",
			message: WebSocketMessage{
				Type:      "script_status",
				Timestamp: time.Date(2025, 7, 31, 19, 0, 0, 0, time.UTC),
				Data: map[string]interface{}{
					"script_name": "test.sh",
					"status":      "running",
					"exit_code":   float64(0),
				},
			},
		},
		{
			name: "system metrics message",
			message: WebSocketMessage{
				Type:      "system_metrics",
				Timestamp: time.Date(2025, 7, 31, 19, 0, 0, 0, time.UTC),
				Data: map[string]interface{}{
					"cpu_percent":      45.2,
					"memory_percent":   67.8,
					"active_scripts":   float64(3),
					"total_executions": float64(156),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize to JSON
			jsonBytes, err := json.Marshal(tt.message)
			require.NoError(t, err)

			// Deserialize from JSON
			var deserialized WebSocketMessage
			err = json.Unmarshal(jsonBytes, &deserialized)
			require.NoError(t, err)

			// Verify fields
			assert.Equal(t, tt.message.Type, deserialized.Type)
			assert.Equal(t, tt.message.Timestamp.Unix(), deserialized.Timestamp.Unix())
			assert.Equal(t, tt.message.Data, deserialized.Data)
		})
	}
}

func TestWebSocketHub_BroadcastMessage(t *testing.T) {
	hub := NewWebSocketHub()

	// Test that BroadcastMessage creates proper JSON and sends to broadcast channel
	testData := map[string]interface{}{
		"script_name": "test.sh",
		"status":      "running",
	}

	err := hub.BroadcastMessage("script_status", testData)
	assert.NoError(t, err)

	// Verify message was sent to broadcast channel
	select {
	case message := <-hub.broadcast:
		var wsMessage WebSocketMessage
		err := json.Unmarshal(message, &wsMessage)
		require.NoError(t, err)

		assert.Equal(t, "script_status", wsMessage.Type)
		assert.Equal(t, testData, wsMessage.Data)
		assert.False(t, wsMessage.Timestamp.IsZero())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message to be sent to broadcast channel")
	}
}

func TestWebSocketHub_BroadcastScriptEvents(t *testing.T) {
	hub := NewWebSocketHub()

	// Test broadcasting different types of script events
	testCases := []struct {
		name     string
		msgType  string
		data     map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:    "script starting event",
			msgType: "script_status",
			data: map[string]interface{}{
				"script_name": "backup.sh",
				"status":      "starting",
				"exit_code":   0,
				"duration":    int64(0),
			},
			expected: map[string]interface{}{
				"script_name": "backup.sh",
				"status":      "starting",
				"exit_code":   float64(0), // JSON unmarshaling converts to float64
				"duration":    float64(0),
			},
		},
		{
			name:    "script completed event",
			msgType: "script_status",
			data: map[string]interface{}{
				"script_name": "backup.sh",
				"status":      "completed",
				"exit_code":   0,
				"duration":    int64(1234),
			},
			expected: map[string]interface{}{
				"script_name": "backup.sh",
				"status":      "completed",
				"exit_code":   float64(0),
				"duration":    float64(1234),
			},
		},
		{
			name:    "script failed event",
			msgType: "script_status",
			data: map[string]interface{}{
				"script_name": "deploy.sh",
				"status":      "failed",
				"exit_code":   1,
				"duration":    int64(500),
			},
			expected: map[string]interface{}{
				"script_name": "deploy.sh",
				"status":      "failed",
				"exit_code":   float64(1),
				"duration":    float64(500),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := hub.BroadcastMessage(tc.msgType, tc.data)
			assert.NoError(t, err)

			// Verify message was sent to broadcast channel
			select {
			case message := <-hub.broadcast:
				var wsMessage WebSocketMessage
				err := json.Unmarshal(message, &wsMessage)
				require.NoError(t, err)

				assert.Equal(t, tc.msgType, wsMessage.Type)
				assert.Equal(t, tc.expected, wsMessage.Data)
				assert.False(t, wsMessage.Timestamp.IsZero())
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Expected message to be sent to broadcast channel")
			}
		})
	}
}

func TestWebSocketHub_Run_ClientRegistration(t *testing.T) {
	hub := NewWebSocketHub()

	// Create a mock client
	mockClient := &WebSocketClient{
		hub:  hub,
		send: make(chan []byte, 256),
	}

	// Start the hub in a goroutine
	go hub.Run()

	// Register a client
	hub.register <- mockClient

	// Give the hub a moment to process
	time.Sleep(10 * time.Millisecond)

	// Verify the client was registered
	assert.Equal(t, 1, hub.GetConnectionCount())
}

func TestWebSocketHub_Run_ClientUnregistration(t *testing.T) {
	hub := NewWebSocketHub()

	// Create and pre-register a mock client
	mockClient := &WebSocketClient{
		hub:  hub,
		send: make(chan []byte, 256),
	}
	hub.clients[mockClient] = true

	// Start the hub in a goroutine
	go hub.Run()

	// Unregister the client
	hub.unregister <- mockClient

	// Give the hub a moment to process
	time.Sleep(10 * time.Millisecond)

	// Verify the client was unregistered
	assert.Equal(t, 0, hub.GetConnectionCount())
}

func TestWebSocketHub_Run_MessageBroadcast(t *testing.T) {
	hub := NewWebSocketHub()

	// Create mock clients
	client1 := &WebSocketClient{
		hub:  hub,
		send: make(chan []byte, 256),
	}
	client2 := &WebSocketClient{
		hub:  hub,
		send: make(chan []byte, 256),
	}

	// Pre-register clients
	hub.clients[client1] = true
	hub.clients[client2] = true

	// Start the hub in a goroutine
	go hub.Run()

	// Send a broadcast message
	testMessage := []byte(`{"type":"test","data":{"message":"hello"}}`)
	hub.broadcast <- testMessage

	// Give the hub a moment to process
	time.Sleep(10 * time.Millisecond)

	// Verify both clients received the message
	select {
	case msg := <-client1.send:
		assert.Equal(t, testMessage, msg)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Client1 should have received broadcast message")
	}

	select {
	case msg := <-client2.send:
		assert.Equal(t, testMessage, msg)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Client2 should have received broadcast message")
	}
}

func TestWebSocketHub_ConnectionLimit(t *testing.T) {
	hub := NewWebSocketHub()
	hub.maxConnections = 2

	// Start the hub in a goroutine
	go hub.Run()

	// Create clients with proper send channels but without actual WebSocket connections
	// We'll test the limit logic without involving actual WebSocket operations
	client1 := &WebSocketClient{hub: hub, send: make(chan []byte, 256)}
	client2 := &WebSocketClient{hub: hub, send: make(chan []byte, 256)}
	client3 := &WebSocketClient{hub: hub, send: make(chan []byte, 256)}

	// Register first two clients (should succeed)
	hub.register <- client1
	hub.register <- client2
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 2, hub.GetConnectionCount())

	// For the third client test, we need to account for the fact that
	// the hub will try to close the connection when limit is reached
	// Since we don't have real connections, we'll just verify the count doesn't increase
	originalCount := hub.GetConnectionCount()
	hub.register <- client3
	time.Sleep(10 * time.Millisecond)

	// Should still have only the original number of clients (limit enforced)
	assert.Equal(t, originalCount, hub.GetConnectionCount())
}

func TestWebSocketClient_Integration(t *testing.T) {
	// Set up a test WebSocket server
	hub := NewWebSocketHub()
	go hub.Run()

	// Set up Gin router with WebSocket handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		HandleWebSocket(hub, c)
	})

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Give the connection time to register
	time.Sleep(50 * time.Millisecond)

	// Verify client was registered
	assert.Equal(t, 1, hub.GetConnectionCount())

	// Test broadcasting a message
	testData := map[string]interface{}{
		"test": "message",
	}
	err = hub.BroadcastMessage("test_type", testData)
	require.NoError(t, err)

	// Read message from WebSocket
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, messageBytes, err := conn.ReadMessage()
	require.NoError(t, err)

	// Verify message content
	var receivedMessage WebSocketMessage
	err = json.Unmarshal(messageBytes, &receivedMessage)
	require.NoError(t, err)

	assert.Equal(t, "test_type", receivedMessage.Type)
	assert.Equal(t, testData, receivedMessage.Data)

	// Close connection and verify cleanup
	conn.Close()
	time.Sleep(50 * time.Millisecond)

	// Connection should be cleaned up
	assert.Equal(t, 0, hub.GetConnectionCount())
}

func TestWebSocketClient_MessageHandling(t *testing.T) {
	// Set up a test WebSocket server
	hub := NewWebSocketHub()
	go hub.Run()

	// Set up Gin router with WebSocket handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) {
		HandleWebSocket(hub, c)
	})

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Give the connection time to register
	time.Sleep(50 * time.Millisecond)

	// Send multiple messages rapidly to test message queuing
	messages := []map[string]interface{}{
		{"id": 1, "message": "first"},
		{"id": 2, "message": "second"},
		{"id": 3, "message": "third"},
	}

	for i, data := range messages {
		err = hub.BroadcastMessage("batch_test", data)
		require.NoError(t, err, "Failed to broadcast message %d", i+1)
	}

	// Read messages (they might be combined into fewer frames due to message queuing)
	receivedMessages := make([]WebSocketMessage, 0, len(messages))

	// Read up to the number of messages we sent, with timeout for each read
	for i := 0; i < len(messages); i++ {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			// If we can't read more messages, break (might be combined)
			break
		}

		// Messages might be combined with newlines, split them
		parts := strings.Split(string(messageBytes), "\n")
		for _, part := range parts {
			if strings.TrimSpace(part) == "" {
				continue
			}

			var msg WebSocketMessage
			err = json.Unmarshal([]byte(part), &msg)
			if err != nil {
				// Skip invalid JSON parts
				continue
			}
			receivedMessages = append(receivedMessages, msg)
		}
	}

	// Verify all messages were received
	assert.Equal(t, len(messages), len(receivedMessages))

	// Verify message content
	for i, expected := range messages {
		found := false
		for _, received := range receivedMessages {
			if received.Type == "batch_test" {
				if idVal, ok := received.Data["id"]; ok && idVal == float64(expected["id"].(int)) {
					assert.Equal(t, expected["message"], received.Data["message"])
					found = true
					break
				}
			}
		}
		assert.True(t, found, "Message %d not found in received messages", i+1)
	}
}

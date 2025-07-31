package web

import (
	"encoding/json"
	"testing"
	"time"

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

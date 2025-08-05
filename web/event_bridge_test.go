package web

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"run-script-service/service"
)

func TestEventBridge_Creation(t *testing.T) {
	wsHub := NewWebSocketHub()
	eventBroadcaster := service.NewEventBroadcaster()

	bridge := NewEventBridge(wsHub, eventBroadcaster)
	defer bridge.Close()

	assert.NotNil(t, bridge)
	assert.Equal(t, wsHub, bridge.wsHub)
	assert.Equal(t, eventBroadcaster, bridge.eventBroadcaster)
	assert.NotNil(t, bridge.events)
	assert.NotNil(t, bridge.unsubscribe)
}

func TestEventBridge_EventProcessing(t *testing.T) {
	wsHub := NewWebSocketHub()
	eventBroadcaster := service.NewEventBroadcaster()

	bridge := NewEventBridge(wsHub, eventBroadcaster)
	defer bridge.Close()

	// Create a script status event
	event := service.NewScriptStatusEvent("test.sh", "completed", 0, 1500)

	// Broadcast the event through the service broadcaster
	eventBroadcaster.Broadcast(event)

	// Give the bridge time to process the event
	time.Sleep(50 * time.Millisecond)

	// Check that the WebSocket hub received the message
	select {
	case message := <-wsHub.broadcast:
		var wsMessage WebSocketMessage
		err := json.Unmarshal(message, &wsMessage)
		require.NoError(t, err)

		assert.Equal(t, "script_status", wsMessage.Type)
		assert.Equal(t, "test.sh", wsMessage.Data["script_name"])
		assert.Equal(t, "completed", wsMessage.Data["status"])
		assert.Equal(t, float64(0), wsMessage.Data["exit_code"]) // JSON unmarshaling converts to float64
		assert.Equal(t, float64(1500), wsMessage.Data["duration"])
		assert.NotEmpty(t, wsMessage.Data["timestamp"])
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected WebSocket message to be received")
	}
}

func TestEventBridge_MultipleEvents(t *testing.T) {
	wsHub := NewWebSocketHub()
	eventBroadcaster := service.NewEventBroadcaster()

	bridge := NewEventBridge(wsHub, eventBroadcaster)
	defer bridge.Close()

	// Create multiple events
	events := []*service.ScriptStatusEvent{
		service.NewScriptStatusEvent("script1.sh", "starting", 0, 0),
		service.NewScriptStatusEvent("script1.sh", "completed", 0, 1000),
		service.NewScriptStatusEvent("script2.sh", "starting", 0, 0),
		service.NewScriptStatusEvent("script2.sh", "failed", 1, 500),
	}

	// Broadcast all events
	for _, event := range events {
		eventBroadcaster.Broadcast(event)
	}

	// Give the bridge time to process all events
	time.Sleep(100 * time.Millisecond)

	// Collect all WebSocket messages
	receivedMessages := make([]WebSocketMessage, 0, 4)
	for i := 0; i < 4; i++ {
		select {
		case message := <-wsHub.broadcast:
			var wsMessage WebSocketMessage
			err := json.Unmarshal(message, &wsMessage)
			require.NoError(t, err)
			receivedMessages = append(receivedMessages, wsMessage)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("Expected to receive message %d", i+1)
		}
	}

	assert.Len(t, receivedMessages, 4)

	// Verify each message
	expectedScripts := []string{"script1.sh", "script1.sh", "script2.sh", "script2.sh"}
	expectedStatuses := []string{"starting", "completed", "starting", "failed"}
	expectedExitCodes := []float64{0, 0, 0, 1}

	for i, msg := range receivedMessages {
		assert.Equal(t, "script_status", msg.Type)
		assert.Equal(t, expectedScripts[i], msg.Data["script_name"])
		assert.Equal(t, expectedStatuses[i], msg.Data["status"])
		assert.Equal(t, expectedExitCodes[i], msg.Data["exit_code"])
	}
}

func TestEventBridge_Close(t *testing.T) {
	wsHub := NewWebSocketHub()
	eventBroadcaster := service.NewEventBroadcaster()

	bridge := NewEventBridge(wsHub, eventBroadcaster)

	// Close the bridge
	bridge.Close()

	// After closing, events should not be processed
	event := service.NewScriptStatusEvent("test.sh", "completed", 0, 1000)
	eventBroadcaster.Broadcast(event)

	// Give time for processing (which shouldn't happen)
	time.Sleep(50 * time.Millisecond)

	// No message should be received
	select {
	case <-wsHub.broadcast:
		t.Fatal("No message should be received after bridge is closed")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message received
	}
}

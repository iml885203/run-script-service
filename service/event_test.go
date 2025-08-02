package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScriptStatusEvent tests the creation of script status events
func TestScriptStatusEvent_Creation(t *testing.T) {
	event := NewScriptStatusEvent("test.sh", "running", 0, 0)

	assert.NotNil(t, event)
	assert.Equal(t, "test.sh", event.ScriptName)
	assert.Equal(t, "running", event.Status)
	assert.Equal(t, 0, event.ExitCode)
	assert.Equal(t, int64(0), event.Duration)
	assert.False(t, event.Timestamp.IsZero())
}

// TestScriptStatusEvent_JSON tests JSON serialization of script status events
func TestScriptStatusEvent_JSON(t *testing.T) {
	timestamp := time.Date(2025, 7, 31, 19, 0, 0, 0, time.UTC)
	event := &ScriptStatusEvent{
		ScriptName: "backup.sh",
		Status:     "completed",
		ExitCode:   0,
		Duration:   1234,
		Timestamp:  timestamp,
	}

	jsonData := event.ToJSON()

	expectedData := map[string]interface{}{
		"script_name": "backup.sh",
		"status":      "completed",
		"exit_code":   0,
		"duration":    int64(1234),
		"timestamp":   timestamp.Format(time.RFC3339),
	}

	assert.Equal(t, expectedData, jsonData)
}

// TestEventBroadcaster tests the event broadcasting system
func TestEventBroadcaster_Creation(t *testing.T) {
	broadcaster := NewEventBroadcaster()

	assert.NotNil(t, broadcaster)
	assert.NotNil(t, broadcaster.listeners)
}

// TestEventBroadcaster_Subscribe tests subscribing to events
func TestEventBroadcaster_Subscribe(t *testing.T) {
	broadcaster := NewEventBroadcaster()
	events := make(chan *ScriptStatusEvent, 10)

	// Subscribe to events
	unsubscribe := broadcaster.Subscribe(events)

	assert.NotNil(t, unsubscribe)
	assert.Len(t, broadcaster.listeners, 1)

	// Unsubscribe
	unsubscribe()
	assert.Len(t, broadcaster.listeners, 0)
}

// TestEventBroadcaster_Broadcast tests broadcasting events to subscribers
func TestEventBroadcaster_Broadcast(t *testing.T) {
	broadcaster := NewEventBroadcaster()
	events1 := make(chan *ScriptStatusEvent, 10)
	events2 := make(chan *ScriptStatusEvent, 10)

	// Subscribe two listeners
	unsubscribe1 := broadcaster.Subscribe(events1)
	unsubscribe2 := broadcaster.Subscribe(events2)
	defer unsubscribe1()
	defer unsubscribe2()

	// Create and broadcast an event
	event := NewScriptStatusEvent("test.sh", "running", 0, 0)
	broadcaster.Broadcast(event)

	// Both listeners should receive the event
	select {
	case receivedEvent := <-events1:
		assert.Equal(t, event.ScriptName, receivedEvent.ScriptName)
		assert.Equal(t, event.Status, receivedEvent.Status)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event to be received by listener 1")
	}

	select {
	case receivedEvent := <-events2:
		assert.Equal(t, event.ScriptName, receivedEvent.ScriptName)
		assert.Equal(t, event.Status, receivedEvent.Status)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event to be received by listener 2")
	}
}

// TestEventBroadcaster_NonBlockingBroadcast tests that broadcasting doesn't block on full channels
func TestEventBroadcaster_NonBlockingBroadcast(t *testing.T) {
	broadcaster := NewEventBroadcaster()

	// Create a channel with no buffer (will block on send)
	events := make(chan *ScriptStatusEvent)
	unsubscribe := broadcaster.Subscribe(events)
	defer unsubscribe()

	// Broadcasting should not block even if the channel is not being read
	event := NewScriptStatusEvent("test.sh", "running", 0, 0)

	// This should complete quickly without blocking
	done := make(chan bool, 1)
	go func() {
		broadcaster.Broadcast(event)
		done <- true
	}()

	select {
	case <-done:
		// Good, broadcast completed without blocking
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Broadcast should not block on full channels")
	}
}

// TestEventBroadcaster_MultipleEvents tests broadcasting multiple events
func TestEventBroadcaster_MultipleEvents(t *testing.T) {
	broadcaster := NewEventBroadcaster()
	events := make(chan *ScriptStatusEvent, 10)

	unsubscribe := broadcaster.Subscribe(events)
	defer unsubscribe()

	// Broadcast multiple events
	event1 := NewScriptStatusEvent("test1.sh", "running", 0, 0)
	event2 := NewScriptStatusEvent("test2.sh", "completed", 0, 1000)
	event3 := NewScriptStatusEvent("test3.sh", "failed", 1, 500)

	broadcaster.Broadcast(event1)
	broadcaster.Broadcast(event2)
	broadcaster.Broadcast(event3)

	// Verify all events are received
	receivedEvents := make([]*ScriptStatusEvent, 0, 3)
	for i := 0; i < 3; i++ {
		select {
		case event := <-events:
			receivedEvents = append(receivedEvents, event)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Expected to receive event %d", i+1)
		}
	}

	require.Len(t, receivedEvents, 3)
	assert.Equal(t, "test1.sh", receivedEvents[0].ScriptName)
	assert.Equal(t, "test2.sh", receivedEvents[1].ScriptName)
	assert.Equal(t, "test3.sh", receivedEvents[2].ScriptName)
}

// TestScriptStatusEvent_StatusValidation tests different status values
func TestScriptStatusEvent_StatusValidation(t *testing.T) {
	validStatuses := []string{"running", "completed", "failed", "starting"}

	for _, status := range validStatuses {
		event := NewScriptStatusEvent("test.sh", status, 0, 0)
		assert.Equal(t, status, event.Status)
	}
}

package web

import (
	"run-script-service/service"
)

// EventBridge connects service events to WebSocket broadcasting
type EventBridge struct {
	wsHub            *WebSocketHub
	eventBroadcaster *service.EventBroadcaster
	events           chan *service.ScriptStatusEvent
	unsubscribe      func()
}

// NewEventBridge creates a bridge between service events and WebSocket hub
func NewEventBridge(wsHub *WebSocketHub, eventBroadcaster *service.EventBroadcaster) *EventBridge {
	events := make(chan *service.ScriptStatusEvent, 100)
	unsubscribe := eventBroadcaster.Subscribe(events)

	bridge := &EventBridge{
		wsHub:            wsHub,
		eventBroadcaster: eventBroadcaster,
		events:           events,
		unsubscribe:      unsubscribe,
	}

	// Start processing events
	go bridge.processEvents()

	return bridge
}

// processEvents processes script status events and broadcasts them via WebSocket
func (eb *EventBridge) processEvents() {
	for event := range eb.events {
		// Convert service event to WebSocket message format
		data := map[string]interface{}{
			"script_name": event.ScriptName,
			"status":      event.Status,
			"exit_code":   event.ExitCode,
			"duration":    event.Duration,
			"timestamp":   event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Broadcast via WebSocket
		if err := eb.wsHub.BroadcastMessage("script_status", data); err != nil {
			// Log error but continue processing
			// In a production system, you might want proper logging here
		}
	}
}

// Close stops the event bridge
func (eb *EventBridge) Close() {
	if eb.unsubscribe != nil {
		eb.unsubscribe()
	}
	close(eb.events)
}

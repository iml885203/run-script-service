package service

import (
	"sync"
	"time"
)

// ScriptStatusEvent represents a script status change event
type ScriptStatusEvent struct {
	ScriptName string    `json:"script_name"`
	Status     string    `json:"status"` // "starting", "running", "completed", "failed"
	ExitCode   int       `json:"exit_code"`
	Duration   int64     `json:"duration"` // Duration in milliseconds
	Timestamp  time.Time `json:"timestamp"`
}

// NewScriptStatusEvent creates a new script status event
func NewScriptStatusEvent(scriptName, status string, exitCode int, duration int64) *ScriptStatusEvent {
	return &ScriptStatusEvent{
		ScriptName: scriptName,
		Status:     status,
		ExitCode:   exitCode,
		Duration:   duration,
		Timestamp:  time.Now(),
	}
}

// ToJSON converts the event to a JSON-compatible map
func (e *ScriptStatusEvent) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"script_name": e.ScriptName,
		"status":      e.Status,
		"exit_code":   e.ExitCode,
		"duration":    e.Duration,
		"timestamp":   e.Timestamp.Format(time.RFC3339),
	}
}

// EventBroadcaster manages event broadcasting to multiple listeners
type EventBroadcaster struct {
	listeners []chan<- *ScriptStatusEvent
	mutex     sync.RWMutex
}

// NewEventBroadcaster creates a new event broadcaster
func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		listeners: make([]chan<- *ScriptStatusEvent, 0),
	}
}

// Subscribe adds a listener to receive events
// Returns an unsubscribe function
func (eb *EventBroadcaster) Subscribe(eventChan chan<- *ScriptStatusEvent) func() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	eb.listeners = append(eb.listeners, eventChan)

	// Return unsubscribe function
	return func() {
		eb.mutex.Lock()
		defer eb.mutex.Unlock()

		for i, listener := range eb.listeners {
			if listener == eventChan {
				// Remove this listener from the slice
				eb.listeners = append(eb.listeners[:i], eb.listeners[i+1:]...)
				break
			}
		}
	}
}

// Broadcast sends an event to all subscribers
// This is non-blocking - if a listener's channel is full, the event is dropped for that listener
func (eb *EventBroadcaster) Broadcast(event *ScriptStatusEvent) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	for _, listener := range eb.listeners {
		select {
		case listener <- event:
			// Event sent successfully
		default:
			// Channel is full, drop the event for this listener
		}
	}
}

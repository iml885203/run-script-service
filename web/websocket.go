package web

import (
	"encoding/json"
	"time"
)

// WebSocketMessage represents a message sent through WebSocket
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	hub  *WebSocketHub
	conn interface{} // Will be websocket.Conn when we add gorilla/websocket
	send chan []byte
}

// WebSocketHub manages WebSocket connections and message broadcasting
type WebSocketHub struct {
	// Registered clients
	clients map[*WebSocketClient]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *WebSocketClient

	// Unregister requests from clients
	unregister chan *WebSocketClient
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		broadcast:  make(chan []byte),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		clients:    make(map[*WebSocketClient]bool),
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastMessage sends a message to all connected clients
func (h *WebSocketHub) BroadcastMessage(msgType string, data map[string]interface{}) error {
	message := WebSocketMessage{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case h.broadcast <- messageBytes:
	default:
		// Channel is full, skip message
	}

	return nil
}

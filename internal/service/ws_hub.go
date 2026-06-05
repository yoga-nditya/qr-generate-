package service

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"qris-payment/internal/model"
)

// Hub manages active WebSocket connections grouped by order ID.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool
}

// NewHub creates and returns a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*websocket.Conn]bool),
	}
}

// Register adds a WebSocket connection for the given order ID.
func (h *Hub) Register(orderID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[orderID] == nil {
		h.clients[orderID] = make(map[*websocket.Conn]bool)
	}
	h.clients[orderID][conn] = true
	log.Printf("[WS] Registered connection: order=%s total=%d", orderID, len(h.clients[orderID]))
}

// Unregister removes a WebSocket connection for the given order ID.
func (h *Hub) Unregister(orderID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[orderID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, orderID)
		}
	}
	log.Printf("[WS] Unregistered connection: order=%s", orderID)
}

// Broadcast sends a message to all active connections for the given order ID.
func (h *Hub) Broadcast(orderID string, msg model.WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.clients[orderID]
	if !ok || len(conns) == 0 {
		log.Printf("[WS] No active clients for order=%s", orderID)
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WS] Failed to marshal message: %v", err)
		return
	}

	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[WS] Failed to write to client: %v", err)
		}
	}
	log.Printf("[WS] Broadcast sent to %d client(s): order=%s status=%s", len(conns), orderID, msg.Status)
}

package service

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"qris-payment/internal/model"
)

type Hub struct {
	clients map[string]map[*websocket.Conn]bool
	mu      sync.RWMutex
}

var GlobalHub = &Hub{
	clients: make(map[string]map[*websocket.Conn]bool),
}

func (h *Hub) Register(orderID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[orderID] == nil {
		h.clients[orderID] = make(map[*websocket.Conn]bool)
	}
	h.clients[orderID][conn] = true
	log.Printf("[WS] ✅ Register  order=%s  total=%d", orderID, len(h.clients[orderID]))
}

func (h *Hub) Unregister(orderID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.clients[orderID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, orderID)
		}
	}
	log.Printf("[WS] ❌ Unregister order=%s", orderID)
}

func (h *Hub) Broadcast(orderID string, msg model.WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.clients[orderID]
	if !ok || len(conns) == 0 {
		log.Printf("[WS] ⚠️  No clients for order=%s", orderID)
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WS] Marshal error: %v", err)
		return
	}

	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("[WS] Write error: %v", err)
		}
	}
	log.Printf("[WS] 📡 Broadcasted to %d client(s) order=%s status=%s",
		len(conns), orderID, msg.Status)
}

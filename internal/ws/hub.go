package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu    sync.RWMutex
	conns map[int]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{conns: make(map[int]*websocket.Conn)}
}

func (h *Hub) Add(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = conn
}

func (h *Hub) Remove(userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, userID)
}

func (h *Hub) Notify(userID int, msg any) {
	h.mu.RLock()
	conn := h.conns[userID]
	h.mu.RUnlock()

	if conn != nil {
		_ = conn.WriteJSON(msg)
	}
}

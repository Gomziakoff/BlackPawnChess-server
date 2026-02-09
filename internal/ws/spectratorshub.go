package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type SpectratorHub struct {
	mu    sync.RWMutex
	games map[int]map[int]*websocket.Conn
}

func NewSpectratorHub() *SpectratorHub {
	return &SpectratorHub{games: make(map[int]map[int]*websocket.Conn)}
}

func (g *SpectratorHub) Add(gameID, userID int, conn *websocket.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.games[gameID] == nil {
		g.games[gameID] = make(map[int]*websocket.Conn)
	}
	g.games[gameID][userID] = conn
}

func (g *SpectratorHub) Remove(gameID, userID int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.games[gameID] == nil {
		return
	}

	delete(g.games[gameID], userID)

	if len(g.games[gameID]) == 0 {
		delete(g.games, gameID)
	}
}

func (g *SpectratorHub) Broadcast(gameID int, msg any) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, conn := range g.games[gameID] {
		_ = conn.WriteJSON(msg)
	}
}

func (h *SpectratorHub) Count(gameID int) int {
	if h == nil {
		return 0
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.games[gameID] == nil {
		return 0
	}
	return len(h.games[gameID])
}

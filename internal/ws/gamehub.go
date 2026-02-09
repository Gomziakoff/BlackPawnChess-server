package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type GameHub struct {
	mu    sync.RWMutex
	games map[int]map[int]*websocket.Conn
}

func NewGameHub() *GameHub {
	return &GameHub{games: make(map[int]map[int]*websocket.Conn)}
}

func (g *GameHub) Add(gameID, userID int, conn *websocket.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.games[gameID] == nil {
		g.games[gameID] = make(map[int]*websocket.Conn)
	}
	g.games[gameID][userID] = conn
}

func (g *GameHub) Remove(gameID, userID int) {
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

func (g *GameHub) Broadcast(gameID int, msg any) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for _, conn := range g.games[gameID] {
		_ = conn.WriteJSON(msg)
	}
}

func (g *GameHub) Notify(gameID, userID int, msg any) {
	g.mu.RLock()
	conn := g.games[gameID][userID]
	g.mu.RUnlock()

	if conn != nil {
		_ = conn.WriteJSON(msg)
	}
}

func (h *GameHub) Crowd(gameID int, whiteID, blackID int) map[string]bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	crowd := map[string]bool{
		"white": false,
		"black": false,
	}

	if players, ok := h.games[gameID]; ok {
		if _, ok := players[whiteID]; ok {
			crowd["white"] = true
		}
		if _, ok := players[blackID]; ok {
			crowd["black"] = true
		}
	}

	return crowd
}

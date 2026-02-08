package ws

import (
	gameservice "BlackPawnChess-server/internal/gameService"
	"BlackPawnChess-server/internal/matchmaking"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub         *Hub
	gameHub     *GameHub
	matchmaker  matchmaking.Matchmaker
	gameService gameservice.GameService
}

func NewHandler(hub *Hub, gameHub *GameHub, matchmaker matchmaking.Matchmaker, gameService gameservice.GameService) *Handler {
	return &Handler{hub: hub, gameHub: gameHub, matchmaker: matchmaker, gameService: gameService}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) WS(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id type"})
		return
	}

	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	h.hub.Add(userID, conn)
	defer h.hub.Remove(userID)

	for {
		var msg IncomingMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("WS read error:", err)
			return
		}
		log.Printf("WS incoming from user %d: T=%s D=%s\n", userID, msg.T, string(msg.D)) //*
		switch msg.T {
		case "seek":
			h.hub.Notify(userID, OutgoingMessage{
				T: "seek:received",
			})

			ctx := context.Background()

			opponent, err := h.matchmaker.Seek(ctx, userID)
			if err != nil {
				h.hub.Notify(userID, OutgoingMessage{
					T: "error",
					D: err,
				})
				continue
			}

			if opponent == 0 {
				continue
			}

			gameID, err := h.gameService.StartGame(ctx, userID, opponent)
			if err != nil {
				h.hub.Notify(userID, OutgoingMessage{T: "error", D: "game creation failed"})
				h.hub.Notify(opponent, OutgoingMessage{T: "error", D: "game creation failed"})
				continue
			}

			redirectMsg := OutgoingMessage{
				T: "redirect",
				D: gin.H{
					"gameId": gameID,
					"url":    "/game/" + strconv.Itoa(gameID),
				},
			}

			h.hub.Notify(userID, redirectMsg)
			h.hub.Notify(opponent, redirectMsg)

		case "stop-seek":
			ctx := context.Background()
			err := h.matchmaker.Cancel(ctx, userID)
			if err != nil {
				h.hub.Notify(userID, OutgoingMessage{T: "error", D: "game creation failed"})
			}
		}
	}
}

func (h *Handler) GameWS(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id type"})
		return
	}

	gameIDStr := c.Param("id")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}

	_, err = h.gameService.GetGameState(gameID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	whiteID, blackID, err := h.gameService.GetPlayers(gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot get players"})
		return
	}

	if userID != whiteID && userID != blackID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a player"})
		return
	}

	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	h.gameHub.Add(gameID, userID, conn)
	defer func() {
		h.gameHub.Remove(gameID, userID)

		crowd := h.gameHub.Crowd(gameID, whiteID, blackID)
		h.gameHub.Broadcast(gameID, OutgoingMessage{
			T: "crowd",
			D: crowd,
		})
	}()

	crowd := h.gameHub.Crowd(gameID, whiteID, blackID)

	h.gameHub.Broadcast(gameID, OutgoingMessage{
		T: "crowd",
		D: crowd,
	})

	for {
		var msg IncomingMessage
		if err := conn.ReadJSON(&msg); err != nil {
			return
		}

		switch msg.T {
		case "move":
			var payload MovePayload
			if err := json.Unmarshal(msg.D, &payload); err != nil {
				h.gameHub.Notify(gameID, userID, OutgoingMessage{T: "error", D: "invalid move payload"})
				continue
			}

			msg, err := h.gameService.MakeMove(gameID, userID, payload.U)
			if err != nil {
				h.gameHub.Notify(gameID, userID, OutgoingMessage{T: "error", D: err.Error()})
				continue
			}
			h.gameHub.Notify(gameID, userID, OutgoingMessage{T: "ack", D: payload.A})

			h.gameHub.Broadcast(gameID, msg)

		case "resign":
			msg, err := h.gameService.Resign(gameID, userID)
			if err != nil {
				h.gameHub.Notify(gameID, userID, OutgoingMessage{T: "error", D: err.Error()})
				continue
			}

			h.gameHub.Broadcast(gameID, msg)
		}
	}
}

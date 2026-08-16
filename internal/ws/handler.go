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
	hub           *Hub
	gameHub       *GameHub
	spectratorHub *SpectratorHub
	matchmaker    matchmaking.Matchmaker
	gameService   gameservice.GameService
}

func NewHandler(hub *Hub, gameHub *GameHub, spectratorHub *SpectratorHub, matchmaker matchmaking.Matchmaker, gameService *gameservice.Service) *Handler {
	h := &Handler{hub: hub, gameHub: gameHub, spectratorHub: spectratorHub, matchmaker: matchmaker, gameService: gameService}

	gameService.SetGameEventHandler(func(gameID int, msg gameservice.OutgoingMessage) {
		h.gameHub.Broadcast(gameID, msg)
	})

	return h
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
		switch msg.T {
		case "seek":
			h.hub.Notify(userID, OutgoingMessage{
				T: "seek:received",
			})
			var payload SeekPayload
			if err := json.Unmarshal(msg.D, &payload); err != nil {
				h.hub.Notify(userID, OutgoingMessage{
					T: "error",
					D: "invalid seek payload",
				})
				continue
			}

			ctx := context.Background()

			opponent, err := h.matchmaker.Seek(ctx, userID, payload.InitialTime, payload.Increment)
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

			gameID, err := h.gameService.StartGame(ctx, userID, opponent, payload.InitialTime, payload.Increment)
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
		case "ping":
			h.hub.Notify(userID, OutgoingMessage{T: "pong"})
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

		h.broadcastCrowd(gameID)
	}()

	h.broadcastCrowd(gameID)

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

			msgs, err := h.gameService.MakeMove(gameID, userID, payload.U)
			if err != nil {
				h.gameHub.Notify(gameID, userID, OutgoingMessage{T: "error", D: err.Error()})
				continue
			}
			h.gameHub.Notify(gameID, userID, OutgoingMessage{T: "ack", D: payload.A})

			for _, m := range msgs {
				h.gameHub.Broadcast(gameID, m)
				h.spectratorHub.Broadcast(gameID, m)
			}

		case "resign":
			msg, err := h.gameService.Resign(gameID, userID)
			if err != nil {
				h.gameHub.Notify(gameID, userID, OutgoingMessage{T: "error", D: err.Error()})
				continue
			}

			h.gameHub.Broadcast(gameID, msg)
			h.spectratorHub.Broadcast(gameID, msg)

		case "flag":
			msg, err := h.gameService.Flag(gameID)
			if err != nil {
				h.gameHub.Notify(gameID, userID, OutgoingMessage{T: "error", D: err.Error()})
				continue
			}
			h.gameHub.Broadcast(gameID, msg)
			h.spectratorHub.Broadcast(gameID, msg)
		case "ping":
			h.gameHub.Notify(gameID, userID, OutgoingMessage{T: "pong"})
		}
	}
}

func (h *Handler) WatchWS(c *gin.Context) {
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

	conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
	h.spectratorHub.Add(gameID, userID, conn)
	defer func() {
		h.spectratorHub.Remove(gameID, userID)

		h.broadcastCrowd(gameID)
	}()

	h.broadcastCrowd(gameID)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *Handler) BuildCrowd(gameID int, whiteID, blackID int) map[string]interface{} {
	crowd := make(map[string]interface{})

	players := h.gameHub.Crowd(gameID, whiteID, blackID)
	crowd["white"] = players["white"]
	crowd["black"] = players["black"]
	crowd["spectators"] = h.spectratorHub.Count(gameID)

	return crowd
}

func (h *Handler) broadcastCrowd(gameID int) {
	whiteID, blackID, _ := h.gameService.GetPlayers(gameID)

	crowd := h.BuildCrowd(gameID, whiteID, blackID)

	h.gameHub.Broadcast(gameID, OutgoingMessage{
		T: "crowd",
		D: crowd,
	})

	h.spectratorHub.Broadcast(gameID, OutgoingMessage{
		T: "crowd",
		D: crowd,
	})
}

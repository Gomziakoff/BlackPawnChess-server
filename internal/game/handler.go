package game

import (
	gameservice "BlackPawnChess-server/internal/gameService"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	gameService gameservice.GameService
}

func NewHandler(gameService gameservice.GameService) *Handler {
	return &Handler{gameService: gameService}
}

func (h *Handler) GetGame(c *gin.Context) {
	gameIDStr := c.Param("id")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	snapshot, err := h.gameService.GetGame(gameID)
	c.JSON(200, snapshot)
}

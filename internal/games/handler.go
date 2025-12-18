package games

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	repo  *Repository
	redis *RedisRepo
}

func NewHandler(repo *Repository, redis *RedisRepo) *Handler {
	return &Handler{repo: repo, redis: redis}
}

func (h *Handler) SearchGame(c *gin.Context) {
	ctx := context.Background()
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDValue.(int)

	inQueue, _ := h.redis.IsInSearchQueue(ctx, userID)
	if inQueue {
		c.JSON(http.StatusOK, gin.H{"status": GameWaiting})
		return
	}

	opponentID, err := h.redis.PopFromSearchQueue(ctx)
	if err == redis.Nil {
		_ = h.redis.AddToSearchQueue(ctx, userID)
		c.JSON(http.StatusOK, gin.H{"status": GameWaiting})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var whiteID, blackID int
	if userID%2 == 0 {
		whiteID, blackID = userID, opponentID
	} else {
		whiteID, blackID = opponentID, userID
	}

	game, err := h.repo.CreateGame(ctx, whiteID, &blackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.repo.SetGameActive(ctx, game.ID, blackID)
	game.Status = GameActive

	c.JSON(http.StatusOK, gin.H{
		"status": GameActive,
		"game":   game,
	})
}

func (h *Handler) CancelSearch(c *gin.Context) {
	ctx := context.Background()
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDValue.(int)

	if err := h.redis.RemoveFromSearchQueue(ctx, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

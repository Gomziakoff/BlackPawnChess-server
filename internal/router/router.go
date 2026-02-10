package router

import (
	"BlackPawnChess-server/internal/auth"
	"BlackPawnChess-server/internal/game"
	"BlackPawnChess-server/internal/sessions"
	"BlackPawnChess-server/internal/ws"
	"time"

	"github.com/gin-gonic/gin"
)

func Register(
	r *gin.Engine,
	authHandler *auth.Handler,
	wsHandler *ws.Handler,
	gameHandler *game.Handler,
	sessionManager *sessions.Manager,
) {
	api := r.Group("/api/v1")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // фронт
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true, // важно для cookie сессии
		MaxAge:           12 * time.Hour,
	}))

	registerAuthRoutes(api, authHandler, sessionManager)
	registerWebSocketRoutes(api, wsHandler, sessionManager)
	registerGameRoutes(api, gameHandler, sessionManager)
}

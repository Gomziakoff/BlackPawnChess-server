package router

import (
	"BlackPawnChess-server/internal/auth"
	"BlackPawnChess-server/internal/game"
	"BlackPawnChess-server/internal/sessions"
	"BlackPawnChess-server/internal/ws"

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

	registerAuthRoutes(api, authHandler, sessionManager)
	registerWebSocketRoutes(api, wsHandler, sessionManager)
	registerGameRoutes(api, gameHandler)
}

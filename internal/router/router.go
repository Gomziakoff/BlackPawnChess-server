package router

import (
	"BlackPawnChess-server/internal/auth"
	"BlackPawnChess-server/internal/games"
	"BlackPawnChess-server/internal/sessions"
	"BlackPawnChess-server/internal/ws"

	"github.com/gin-gonic/gin"
)

func Register(
	r *gin.Engine,
	authHandler *auth.Handler,
	gameHandler *games.Handler,
	wsHandler *ws.Handler,
	sessionManager *sessions.Manager,
) {
	api := r.Group("/api/v1")

	registerAuthRoutes(api, authHandler, sessionManager)
	registerGameRoutes(api, gameHandler, sessionManager)
	registerWebSocketRoutes(api, wsHandler, sessionManager)
}

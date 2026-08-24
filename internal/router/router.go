package router

import (
	"BlackPawnChess-server/internal/auth"
	"BlackPawnChess-server/internal/game"
	"BlackPawnChess-server/internal/sessions"
	"BlackPawnChess-server/internal/sitemap"
	"BlackPawnChess-server/internal/ws"

	"github.com/gin-gonic/gin"
)

func Register(
	r *gin.Engine,
	authHandler *auth.Handler,
	wsHandler *ws.Handler,
	gameHandler *game.Handler,
	sessionManager *sessions.Manager,
	sitemapHandler *sitemap.Handler,
) {
	api := r.Group("/api/v1")

	registerAuthRoutes(api, authHandler, sessionManager)
	registerWebSocketRoutes(api, wsHandler, sessionManager)
	registerGameRoutes(api, gameHandler, sessionManager)
	registerSitemapRoutes(api, sitemapHandler)
}

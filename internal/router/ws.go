package router

import (
	"BlackPawnChess-server/internal/middleware"
	"BlackPawnChess-server/internal/sessions"
	"BlackPawnChess-server/internal/ws"

	"github.com/gin-gonic/gin"
)

func registerWebSocketRoutes(
	api *gin.RouterGroup,
	handler *ws.Handler,
	sessionManager *sessions.Manager,
) {
	wsGroup := api.Group("/ws")
	wsGroup.Use(middleware.Auth(sessionManager))
	{
		wsGroup.GET("", handler.WS)
		wsGroup.GET("/game/:id", handler.GameWS)
		wsGroup.GET("/watch/:id", handler.WatchWS)
	}
}

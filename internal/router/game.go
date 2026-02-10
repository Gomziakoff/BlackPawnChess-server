package router

import (
	"BlackPawnChess-server/internal/game"
	"BlackPawnChess-server/internal/middleware"
	"BlackPawnChess-server/internal/sessions"

	"github.com/gin-gonic/gin"
)

func registerGameRoutes(
	api *gin.RouterGroup,
	handler *game.Handler,
	sessionManager *sessions.Manager,
) {
	gameGroup := api.Group("/game")
	gameGroup.Use(middleware.Auth(sessionManager))
	{
		gameGroup.GET("/:id", handler.GetGame)
	}
}

package router

import (
	"BlackPawnChess-server/internal/games"
	"BlackPawnChess-server/internal/middleware"
	"BlackPawnChess-server/internal/sessions"

	"github.com/gin-gonic/gin"
)

func registerGameRoutes(
	api *gin.RouterGroup,
	handler *games.Handler,
	sessionManager *sessions.Manager,
) {
	gameGroup := api.Group("/games")
	gameGroup.Use(middleware.Auth(sessionManager))
	{
		gameGroup.POST("/search", handler.SearchGame)
		gameGroup.POST("/cancel-search", handler.CancelSearch)
	}
}

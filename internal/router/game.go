package router

import (
	"BlackPawnChess-server/internal/game"

	"github.com/gin-gonic/gin"
)

func registerGameRoutes(
	api *gin.RouterGroup,
	handler *game.Handler,
) {
	authGroup := api.Group("/game")
	{
		authGroup.GET("/:id", handler.GetGame)
	}
}

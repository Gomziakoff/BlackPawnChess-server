package router

import (
	"BlackPawnChess-server/internal/sitemap"

	"github.com/gin-gonic/gin"
)

func registerSitemapRoutes(
	api *gin.RouterGroup,
	handler *sitemap.Handler,
) {
	api.GET("/sitemap.xml", handler.GetSitemap)
}

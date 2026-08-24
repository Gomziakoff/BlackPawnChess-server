package router

import (
	"BlackPawnChess-server/internal/auth"
	"BlackPawnChess-server/internal/middleware"
	"BlackPawnChess-server/internal/sessions"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(
	api *gin.RouterGroup,
	handler *auth.Handler,
	sessionManager *sessions.Manager,
) {
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/guest", handler.Guest)
	}

	protected := authGroup.Group("/")
	protected.Use(middleware.Auth(sessionManager))
	{
		protected.POST("/logout", handler.Logout)
		protected.GET("/me", handler.Me)
	}
}

package middleware

import (
	"BlackPawnChess-server/internal/sessions"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth(sm *sessions.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, err := sm.GetUserID(c.Request.Context(), sessionID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}

		if err := sm.Refresh(c.Request.Context(), sessionID); err != nil {
			c.Error(err)
		}

		c.Set("user_id", userID)

		c.Next()
	}
}

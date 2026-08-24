package auth

import (
	"BlackPawnChess-server/internal/models"
	"BlackPawnChess-server/pkg/hashing"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	CreateGuest(ctx context.Context) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByUserID(ctx context.Context, userID string) (*models.User, error)
}

type SessionManager interface {
	CreateSession(ctx context.Context, UserID int) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type Handler struct {
	users    UserRepository
	sessions SessionManager
}

func NewHandler(users UserRepository, sessions SessionManager) *Handler {
	return &Handler{
		users:    users,
		sessions: sessions,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := hashing.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password hash failed"})
	}

	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hash,
	}

	if err := h.users.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user already exists"})
		return
	}

	sessionID, err := h.sessions.CreateSession(c.Request.Context(), user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session creation failed"})
		return
	}

	setSessionCookie(c, sessionID)

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.Id,
		"email":    user.Email,
		"username": user.Username,
	})
}

func (h *Handler) Guest(c *gin.Context) {
	user, err := h.users.CreateGuest(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create guest user"})
		return
	}

	sessionID, err := h.sessions.CreateSession(c.Request.Context(), user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session creation failed"})
		return
	}

	setSessionCookie(c, sessionID)
	c.JSON(http.StatusCreated, gin.H{
		"id":       user.Id,
		"username": user.Username,
		"IsGuest":  true,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.users.FindByEmail(c.Request.Context(), req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if err := hashing.CheckPasswordHash(user.PasswordHash, req.Password); err == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	print(err)

	sessionID, err := h.sessions.CreateSession(c.Request.Context(), user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session creation failed"})
		return
	}

	setSessionCookie(c, sessionID)

	c.JSON(http.StatusOK, gin.H{
		"id":       user.Id,
		"email":    user.Email,
		"username": user.Username,
	})
}

func (h *Handler) Me(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id type"})
		return
	}
	user, err := h.users.FindByUserID(c, strconv.Itoa(userID))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		_ = h.sessions.DeleteSession(c.Request.Context(), sessionID)
	}
	ClearSessionCookie(c)
	c.Status(http.StatusNoContent)
}

func setSessionCookie(c *gin.Context, sessionID string) {
	c.SetCookie(
		"session_id",
		sessionID,
		int((7 * 24 * time.Hour).Seconds()),
		"/",
		"",
		true,
		true,
	)
}

func ClearSessionCookie(c *gin.Context) {
	c.SetCookie(
		"session_id",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)
}

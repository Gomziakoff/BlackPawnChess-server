package sessions

import (
	"BlackPawnChess-server/internal/storage/rdb"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("SESSION NOT FOUND")
)

type Manager struct {
	rdb *rdb.Redis
	ttl time.Duration
}

func NewManager(redis *rdb.Redis, ttl time.Duration) *Manager {
	return &Manager{
		rdb: redis,
		ttl: ttl,
	}
}

func (m *Manager) CreateSession(
	ctx context.Context,
	userID int,
) (string, error) {
	sessionID := uuid.NewString()

	session := rdb.Session{
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := m.rdb.SaveSession(ctx, sessionID, session); err != nil {
		return "", err
	}

	return sessionID, nil
}

func (m *Manager) GetUserID(
	ctx context.Context,
	sessionID string,
) (int, error) {
	session, err := m.rdb.GetSession(ctx, sessionID)
	if err != nil {
		return 0, err
	}
	if session == nil {
		return 0, nil
	}

	return session.UserID, nil
}

func (m *Manager) DeleteSession(
	ctx context.Context,
	sessionID string,
) error {
	return m.rdb.DeleteSession(ctx, sessionID)
}

func (m *Manager) Refresh(
	ctx context.Context,
	sessionID string,
) error {
	return m.rdb.RefreshSession(ctx, sessionID, m.ttl)
}

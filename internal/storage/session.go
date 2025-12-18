package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Session struct {
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

const SessionTTL = 7 * 24 * time.Hour

func (r *Redis) SaveSession(
	ctx context.Context,
	sessionID string,
	session Session,
) error {
	key := fmt.Sprintf("session:%s", sessionID)

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return r.Client.Set(ctx, key, data, SessionTTL).Err()
}

func (r *Redis) GetSession(
	ctx context.Context,
	sessionID string,
) (*Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	data, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *Redis) DeleteSession(
	ctx context.Context,
	sessionID string,
) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Client.Del(ctx, key).Err()
}

func (r *Redis) RefreshSession(
	ctx context.Context,
	sessionID string,
	ttl time.Duration,
) error {
	key := fmt.Sprintf("session:%s", sessionID)

	ok, err := r.Client.Expire(ctx, key, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return redis.Nil
	}

	return nil
}

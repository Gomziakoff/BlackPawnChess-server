package matchmaking

import (
	"BlackPawnChess-server/internal/storage/rdb"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisMatchmaker struct {
	redis *rdb.Redis
}

func NewRedisMatchmaker(redis *rdb.Redis) *RedisMatchmaker {
	return &RedisMatchmaker{redis: redis}
}

func (m *RedisMatchmaker) Seek(ctx context.Context, userID int, initialTime int, increment int) (int, error) {

	userKey := fmt.Sprintf("seeking:%d", userID)
	queueKey := fmt.Sprintf("queue:%d:%d", initialTime, increment)

	exists, err := m.redis.Client.Exists(ctx, userKey).Result()
	if err != nil {
		return 0, err
	}
	if exists > 0 {
		return 0, nil
	}

	value := fmt.Sprintf("%d:%d", initialTime, increment)
	if err := m.redis.Client.Set(ctx, userKey, value, 5*time.Minute).Err(); err != nil {
		return 0, err
	}

	opponent, err := m.redis.Client.LPop(ctx, queueKey).Int()
	if err == redis.Nil {
		if err := m.redis.Client.RPush(ctx, queueKey, userID).Err(); err != nil {
			_ = m.redis.Client.Del(ctx, userKey)
			return 0, err
		}
		return 0, nil
	}
	if err != nil {
		_ = m.redis.Client.Del(ctx, userKey)
		return 0, err
	}
	_ = m.redis.Client.Del(ctx, userKey)
	_ = m.redis.Client.Del(ctx, fmt.Sprintf("seeking:%d", opponent))

	return opponent, nil
}

func (m *RedisMatchmaker) Cancel(ctx context.Context, userID int) error {
	userKey := fmt.Sprintf("seeking:%d", userID)
	value, err := m.redis.Client.Get(ctx, userKey).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}

	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		_ = m.redis.Client.Del(ctx, userKey)
		return nil
	}

	initialTime := parts[0]
	increment := parts[1]

	queueKey := fmt.Sprintf("queue:%s:%s", initialTime, increment)

	if _, err := m.redis.Client.LRem(ctx, queueKey, 0, userID).Result(); err != nil {
		return err
	}
	_ = m.redis.Client.Del(ctx, userKey)

	return nil
}

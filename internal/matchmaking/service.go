package matchmaking

import (
	"BlackPawnChess-server/internal/storage/rdb"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisMatchmaker struct {
	redis *rdb.Redis
}

func NewRedisMatchmaker(redis *rdb.Redis) *RedisMatchmaker {
	return &RedisMatchmaker{redis: redis}
}

func (m *RedisMatchmaker) Seek(ctx context.Context, userID int) (int, error) {
	userKey := fmt.Sprintf("seeking:%d", userID)

	exists, err := m.redis.Client.Exists(ctx, userKey).Result()
	if err != nil {
		return 0, err
	}
	if exists > 0 {
		return 0, nil
	}

	err = m.redis.Client.Set(ctx, userKey, 1, 5*time.Minute).Err()
	if err != nil {
		return 0, err
	}

	opponent, err := m.redis.Client.LPop(ctx, "queue").Int()
	if err == redis.Nil {
		if err := m.redis.Client.RPush(ctx, "queue", userID).Err(); err != nil {
			_ = m.redis.Client.Del(ctx, userKey)
			return 0, err
		}
		return 0, nil
	} else if err != nil {
		_ = m.redis.Client.Del(ctx, userKey)
		return 0, err
	}

	_ = m.redis.Client.Del(ctx, userKey)
	_ = m.redis.Client.Del(ctx, fmt.Sprintf("seeking:%d", opponent))

	return opponent, nil
}

func (m *RedisMatchmaker) Cancel(ctx context.Context, userID int) error {
	if _, err := m.redis.Client.LRem(ctx, "queue", 0, userID).Result(); err != nil {
		return err
	}

	userKey := fmt.Sprintf("seeking:%d", userID)
	_ = m.redis.Client.Del(ctx, userKey)

	return nil
}

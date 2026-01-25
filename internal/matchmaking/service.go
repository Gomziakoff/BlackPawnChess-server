package matchmaking

import (
	"BlackPawnChess-server/internal/storage/rdb"
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisMatchmaker struct {
	redis *rdb.Redis
}

func NewRedisMatchmaker(redis *rdb.Redis) *RedisMatchmaker {
	return &RedisMatchmaker{redis: redis}
}

func (m *RedisMatchmaker) Seek(ctx context.Context, userID int) (int, error) {
	opponent, err := m.redis.Client.LPop(ctx, "queue").Int()
	if err == redis.Nil {
		_ = m.redis.Client.RPush(ctx, "queue", userID)
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return opponent, nil
}

func (m *RedisMatchmaker) Cancel(ctx context.Context, userID int) error {
	_, err := m.redis.Client.LRem(ctx, "queue", 0, userID).Result()
	return err
}

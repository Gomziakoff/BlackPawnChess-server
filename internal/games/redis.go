package games

import (
	"BlackPawnChess-server/internal/storage/rdb"
	"context"
	"fmt"
	"strconv"
)

type RedisRepo struct {
	redis *rdb.Redis
}

const searchQueueKey = "search_queue:default"

func NewRedisRepo(redis *rdb.Redis) *RedisRepo {
	return &RedisRepo{redis: redis}
}

func (r *RedisRepo) AddMove(ctx context.Context, gameID int, move string) error {
	key := fmt.Sprintf("game:%d:moves", gameID)
	return r.redis.Client.RPush(ctx, key, move).Err()
}

func (r *RedisRepo) GetMoves(ctx context.Context, gameID int) ([]string, error) {
	key := fmt.Sprintf("game:%d:moves", gameID)
	return r.redis.Client.LRange(ctx, key, 0, -1).Result()
}

func (r *RedisRepo) ClearMoves(ctx context.Context, gameID int) error {
	key := fmt.Sprintf("game:%d:moves", gameID)
	return r.redis.Client.Del(ctx, key).Err()
}

func (r *RedisRepo) AddToSearchQueue(ctx context.Context, userID int) error {
	return r.redis.Client.RPush(ctx, searchQueueKey, strconv.Itoa(int(userID))).Err()
}

func (r *RedisRepo) PopFromSearchQueue(ctx context.Context) (int, error) {
	val, err := r.redis.Client.LPop(ctx, searchQueueKey).Result()
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (r *RedisRepo) RemoveFromSearchQueue(ctx context.Context, userID int) error {
	return r.redis.Client.LRem(ctx, searchQueueKey, 0, strconv.Itoa(int(userID))).Err()
}

func (r *RedisRepo) IsInSearchQueue(ctx context.Context, userID int) (bool, error) {
	list, err := r.redis.Client.LRange(ctx, searchQueueKey, 0, -1).Result()
	if err != nil {
		return false, err
	}
	uidStr := strconv.Itoa(int(userID))
	for _, v := range list {
		if v == uidStr {
			return true, nil
		}
	}
	return false, nil
}

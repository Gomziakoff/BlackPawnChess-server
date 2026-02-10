package gameservice

import (
	"BlackPawnChess-server/internal/storage/rdb"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db    *gorm.DB
	redis *rdb.Redis
}

func NewRepository(db *gorm.DB, redis *rdb.Redis) *Repository {
	return &Repository{db: db, redis: redis}
}

func (r *Repository) CreateGame(ctx context.Context, game *Game) error {
	return r.db.WithContext(ctx).Create(game).Error
}

func (r *Repository) GetGame(ctx context.Context, gameID int) (Game, error) {
	var g Game
	if err := r.db.WithContext(ctx).First(&g, gameID).Error; err != nil {
		return Game{}, err
	}
	return g, nil
}

func (r *Repository) GetPlayers(ctx context.Context, gameID int) (int, int, error) {
	var g Game
	if err := r.db.WithContext(ctx).First(&g, gameID).Error; err != nil {
		return 0, 0, err
	}
	if g.BlackUserID == nil {
		return g.WhiteUserID, 0, nil
	}
	return g.WhiteUserID, *g.BlackUserID, nil
}

func (r *Repository) FinishGame(ctx context.Context, gameID int, state *GameState, whiteDiff, blackDiff int, winner string, status GameStatus) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&Game{}).
		Where("id = ?", gameID).
		Updates(map[string]interface{}{
			"status":            status,
			"winner":            winner,
			"moves_uci":         state.MovesUCI,
			"fen":               state.FEN,
			"white_time_left":   state.WhiteTimeLeft,
			"black_time_left":   state.BlackTimeLeft,
			"white_rating_diff": whiteDiff,
			"black_rating_diff": blackDiff,
			"finished_at":       &now,
		}).Error
}

func (r *Repository) SaveState(ctx context.Context, state *GameState) error {
	data, _ := json.Marshal(state)
	return r.redis.Client.Set(ctx, r.stateKey(state.GameID), data, 0).Err()
}

func (r *Repository) GetState(ctx context.Context, gameID int) (*GameState, error) {
	data, err := r.redis.Client.Get(ctx, r.stateKey(gameID)).Result()
	if err != nil {
		return nil, err
	}

	var state GameState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *Repository) DeleteState(ctx context.Context, gameID int) error {
	return r.redis.Client.Del(ctx, r.stateKey(gameID)).Err()
}

func (r *Repository) stateKey(gameID int) string {
	return fmt.Sprintf("game:%d", gameID)
}

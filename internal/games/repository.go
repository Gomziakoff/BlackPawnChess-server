package games

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db    *gorm.DB
	redis *RedisRepo
}

func NewRepository(db *gorm.DB, redis *RedisRepo) *Repository {
	return &Repository{db: db, redis: redis}
}

func (r *Repository) CreateGame(ctx context.Context, whitePlayerID int, blackPlayerID *int) (*Game, error) {
	game := &Game{
		WhiteUserID: whitePlayerID,
		BlackUserID: blackPlayerID,
		Status:      GameWaiting,
		MovesUCI:    "",
		CreatedAt:   time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(game).Error; err != nil {
		return nil, err
	}

	return game, nil
}

func (r *Repository) SetGameActive(ctx context.Context, gameID int, blackUserID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var game Game
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&game, gameID).Error; err != nil {
			return err
		}

		if game.Status != GameWaiting {
			return errors.New("game is not in waiting status")
		}

		game.BlackUserID = &blackUserID
		game.Status = GameActive

		return tx.Save(&game).Error
	})
}

func (r *Repository) FinishGame(ctx context.Context, gameID int) error {
	moves, err := r.redis.GetMoves(ctx, gameID)
	if err != nil {
		return err
	}

	movesUCI := strings.Join(moves, " ")

	now := time.Now()
	err = r.db.WithContext(ctx).Model(&Game{}).Where("id = ?", gameID).Updates(map[string]interface{}{
		"moves_uci":   movesUCI,
		"status":      GameFinished,
		"finished_at": &now,
	}).Error
	if err != nil {
		return err
	}

	return r.redis.ClearMoves(ctx, gameID)
}

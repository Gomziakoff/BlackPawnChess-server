package pdb

import (
	"BlackPawnChess-server/internal/models"
	"context"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("user already exists")
		}
		return err
	}
	return nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindByUserID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateRatings(
	ctx context.Context,
	whiteID, blackID int,
	score float64,
	speed string,
) (int, int, error) {
	white, err := r.FindByUserID(ctx, strconv.Itoa(whiteID))
	if err != nil {
		return 0, 0, err
	}
	if white == nil {
		return 0, 0, errors.New("player not found")
	}

	black, err := r.FindByUserID(ctx, strconv.Itoa(blackID))
	if err != nil {
		return 0, 0, err
	}
	if black == nil {
		return 0, 0, errors.New("opponent not found")
	}

	whiteRating, _ := getRatingAndGames(white, speed)
	blackRating, _ := getRatingAndGames(black, speed)

	kWhite := kFactor(white, speed)
	kBlack := kFactor(black, speed)

	newWhiteRating, whiteDiff := calculateElo(whiteRating, blackRating, score, kWhite)
	newBlackRating, blackDiff := calculateElo(blackRating, whiteRating, 1-score, kBlack)

	setRatingAndIncrementGames(white, speed, newWhiteRating)
	setRatingAndIncrementGames(black, speed, newBlackRating)

	return whiteDiff, blackDiff, r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(white).Error; err != nil {
			return err
		}
		if err := tx.Save(black).Error; err != nil {
			return err
		}
		return nil
	})
}

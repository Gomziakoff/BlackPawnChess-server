package games

import "time"

type GameStatus string

const (
	GameWaiting  GameStatus = "waiting"
	GameActive   GameStatus = "active"
	GameFinished GameStatus = "finished"
)

type Game struct {
	ID          int        `gorm:"primaryKey"`
	WhiteUserID int        `gorm:"not null"`
	BlackUserID *int       `gorm:"null"`
	Status      GameStatus `gorm:"type:varchar(16);not null"`
	MovesUCI    string     `gorm:"type:text;not null;default:''"`
	CreatedAt   time.Time
	FinishedAt  *time.Time
}

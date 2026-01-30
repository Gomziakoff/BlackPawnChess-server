package gameservice

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

type GameState struct {
	GameID      int    `json:"game_id"`
	WhiteUserID int    `json:"white_id"`
	BlackUserID int    `json:"black_id"`
	FEN         string `json:"fen"`
	MovesUCI    string `json:"uci"`
	Turn        int    `json:"turn"`
}

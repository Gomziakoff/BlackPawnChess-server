package gameservice

import "time"

type GameStatus string

const (
	GameWaiting  GameStatus = "waiting"
	GameActive   GameStatus = "active"
	GameFinished GameStatus = "finished"
)

type Game struct {
	ID              int        `gorm:"primaryKey"`
	WhiteUserID     int        `gorm:"not null"`
	BlackUserID     *int       `gorm:"null"`
	Status          GameStatus `gorm:"type:varchar(16);not null"`
	MovesUCI        string     `gorm:"type:text;not null;default:''"`
	Speed           string     `gorm:"type:text;not null;default:'rapid'"`
	InitialTime     int        `gorm:"not null"`
	Increment       int        `gorm:"not null"`
	WhiteTimeLeft   int        `gorm:"not null"`
	BlackTimeLeft   int        `gorm:"not null"`
	Turn            int        `gorm:"not null"`
	LastMoveAt      int64      `gorm:"not null"`
	WhiteRatingDiff int        `gorm:"not null;default:0"`
	BlackRatingDiff int        `gorm:"not null;default:0"`
	CreatedAt       time.Time
	FinishedAt      *time.Time
}

type GameState struct {
	GameID        int    `json:"game_id"`
	WhiteUserID   int    `json:"white_id"`
	BlackUserID   int    `json:"black_id"`
	FEN           string `json:"fen"`
	MovesUCI      string `json:"uci"`
	Turn          int    `json:"turn"`
	Speed         string `json:"speed"`
	InitialTime   int    `json:"initial_time"`
	Increment     int    `json:"increment"`
	WhiteTimeLeft int    `json:"white_time"`
	BlackTimeLeft int    `json:"black_time"`
	LastMoveAt    int64  `json:"last_move_at"`
}

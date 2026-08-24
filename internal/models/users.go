package models

import (
	"time"
)

type User struct {
	Id           int    `gorm:"primaryKey"`
	Username     string `gorm:"size:32;not null"`
	Email        string `gorm:"type:varchar(100);unique;not null"`
	PasswordHash string `gorm:"not null"`
	IsGuest      bool   `gorm:"default:false"`

	EloClassical int `gorm:"default:1200"`
	EloRapid     int `gorm:"default:1200"`
	EloBlitz     int `gorm:"default:1200"`
	EloBullet    int `gorm:"default:1200"`

	GamesClassical int `gorm:"default:0"`
	GamesRapid     int `gorm:"default:0"`
	GamesBlitz     int `gorm:"default:0"`
	GamesBullet    int `gorm:"default:0"`

	AvatarURL string
	Title     string
	Country   string
	Role      string `gorm:"type:varchar(16);default:'user'"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

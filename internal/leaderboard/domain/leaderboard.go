package domain

import (
	"time"

	"github.com/google/uuid"
)

type LeaderboardRepository interface {
}
type Participant struct {
	LeaderboardID uuid.UUID `gorm:"type:char(36);not null;index"`
	ClassName     string    `gorm:"type:varchar(255);not null"`
	TotalKudos    int       `gorm:"default:0"`
}

type Leaderboard struct {
	ID           uuid.UUID      `gorm:"type:char(36);primaryKey"`
	Direction    string         `gorm:"type:varchar(255);not null"`
	StartDate    time.Time      `gorm:"not null"`
	EndDate      time.Time      `gorm:"not null"`
	Prize        Prize          `gorm:"embedded;embeddedPrefix:prize_"`
	Participants []*Participant `gorm:"foreignKey:LeaderboardID"`
}

type Prize struct {
	Name        string `gorm:"type:varchar(255)"`
	Description string `gorm:"type:text"`
	PhotoURL    string `gorm:"type:varchar(512)"`
}

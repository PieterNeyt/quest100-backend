package database

import (
	"Quest100Backend/internal/leaderboard/domain"
	"log"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(
		&domain.Leaderboard{},
		&domain.LeaderboardClass{},
	); err != nil {
		log.Printf("Failed to migrate leaderboard database: %v", err)
	}
}

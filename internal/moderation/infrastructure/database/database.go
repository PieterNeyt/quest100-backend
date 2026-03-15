package database

import (
	"Quest100Backend/internal/moderation/domain"
	"log"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(domain.Report{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
}

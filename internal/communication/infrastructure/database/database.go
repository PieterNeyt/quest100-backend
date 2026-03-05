package database

import (
	"Quest100Backend/internal/communication/domain"
	"log"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(&domain.Chat{}); err != nil {
		log.Printf("Failed to migrate Chat: %v", err)
	}

	if err := db.AutoMigrate(&domain.Message{}); err != nil {
		log.Printf("Failed to migrate Message: %v", err)
	}

	if err := db.AutoMigrate(&domain.ChatMember{}); err != nil {
		log.Printf("Failed to migrate ChatMember: %v", err)
	}
}

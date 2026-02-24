package database

import (
	"Quest100Backend/internal/event/domain"
	"log"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(&domain.Event{}); err != nil {
		log.Printf("Failed to migrate Event: %v", err)
	}

	if err := db.AutoMigrate(&domain.EventAttendee{}); err != nil {
		log.Printf("Failed to migrate EventAttendee: %v", err)
	}

	if err := db.AutoMigrate(&domain.ChatMessage{}); err != nil {
		log.Printf("Failed to migrate ChatMessage: %v", err)
	}
}

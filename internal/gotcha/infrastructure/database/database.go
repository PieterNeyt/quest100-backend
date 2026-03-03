package database

import (
	"Quest100Backend/internal/gotcha/domain"
	"log"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(&domain.Game{}); err != nil {
		log.Printf("Failed to migrate Game: %v", err)
	}
	if err := db.AutoMigrate(&domain.Participant{}); err != nil {
		log.Printf("Failed to migrate Participant: %v", err)
	}
	if err := db.AutoMigrate(&domain.Kill{}); err != nil {
		log.Printf("Failed to migrate Kill: %v", err)
	}
	if err := db.AutoMigrate(&domain.KillLike{}); err != nil {
		log.Printf("Failed to migrate KillLike: %v", err)
	}
	if err := db.AutoMigrate(&domain.Prop{}); err != nil {
		log.Printf("Failed to migrate Prop: %v", err)
	}
}

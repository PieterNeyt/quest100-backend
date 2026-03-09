package database

import (
	"Quest100Backend/internal/gotcha/domain"
	"log"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(&domain.GotchaGame{}); err != nil {
		log.Printf("Failed to migrate GotchaGame: %v", err)
	}
	if err := db.AutoMigrate(&domain.Participant{}); err != nil {
		log.Printf("Failed to migrate Participant: %v", err)
	}
	if err := db.AutoMigrate(&domain.GotchaKill{}); err != nil {
		log.Printf("Failed to migrate GotchaKill: %v", err)
	}
	if err := db.AutoMigrate(&domain.GotchaKillLike{}); err != nil {
		log.Printf("Failed to migrate GotchaKillLike: %v", err)
	}
	if err := db.AutoMigrate(&domain.GotchaProp{}); err != nil {
		log.Printf("Failed to migrate GotchaProp: %v", err)
	}

}

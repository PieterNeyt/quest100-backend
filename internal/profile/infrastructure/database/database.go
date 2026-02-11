package database

import (
	"Quest100Backend/internal/profile/domain"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	err := db.AutoMigrate(&domain.Profile{})
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}

	seedDatabase(db)
}

func seedDatabase(db *gorm.DB) {
	hardcodedID, _ := uuid.Parse("00000000-0000-0000-0000-000000000000")

	hugo := domain.Profile{
		ID:        hardcodedID,
		FirstName: "Hugo",
		LastName:  "Dor",
		Email:     "dorhugo@student.kdg.be",
		Kudos:     0,
	}

	err := db.Where(domain.Profile{ID: hardcodedID}).FirstOrCreate(&hugo).Error
	if err != nil {
		log.Printf("Could not seed database: %v", err)
	}
}

package database

import (
	"Quest100Backend/internal/profile/domain"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(&domain.Profile{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
	if err := db.AutoMigrate(&domain.KudosEntry{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
	if err := db.AutoMigrate(&domain.PlayerStats{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}
	if err := db.AutoMigrate(&domain.AwardHistoryEntry{}); err != nil {
		log.Printf("Failed to migrate database: %v", err)
	}

	seedDatabase(db)
}

func seedDatabase(db *gorm.DB) {
	hardcodedID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")

	hugo := domain.Profile{
		ID:                hardcodedID,
		FirstName:         "Hugo",
		LastName:          "Dor",
		Email:             "dorhugo@student.kdg.be",
		Kudos:             0,
		ArchetypeID:       1,
		PrefferedLanguage: domain.NL,
	}

	award := domain.AwardHistoryEntry{
		RecieverID: uuid.New(),
		ProfileID:  uuid.New(),
	}
	err := db.Where(domain.Profile{ID: hardcodedID}).FirstOrCreate(&hugo).Error
	errr := db.Where(domain.AwardHistoryEntry{RecieverID: award.RecieverID}).FirstOrCreate(&award).Error
	if err != nil {
		log.Printf("Could not seed database: %v", err)
	}
	if errr != nil {
		log.Printf("Could not seed database: %v", errr)
	}
}

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

	errAttendance := db.AutoMigrate(&domain.AttendanceRecord{})
	if errAttendance != nil {
		log.Printf("Failed to migrate AttendanceRecord: %v", errAttendance)
	}

	seedDatabase(db)
}

func seedDatabase(db *gorm.DB) {
	hardcodedID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")

	hugo := domain.Profile{
		ID:                hardcodedID,
		FirstName:         "Jon",
		LastName:          "Beton",
		Email:             "jon.beton@student.kdg.be",
		Kudos:             0,
		ArchetypeID:       1,
		Campus:            "Campus Stad",
		PreferredLanguage: domain.NL,
	}

	err := db.Where(domain.Profile{ID: hardcodedID}).FirstOrCreate(&hugo).Error
	if err != nil {
		log.Printf("Could not seed database: %v", err)
	}
}

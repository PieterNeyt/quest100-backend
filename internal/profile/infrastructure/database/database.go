package database

import (
	"Quest100Backend/internal/profile/domain"
	"log"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	errProfile := db.AutoMigrate(&domain.Profile{})
	if errProfile != nil {
		log.Printf("Failed to migrate database: %v", errProfile)
	}

	errKudosEntry := db.AutoMigrate(&domain.KudosEntry{})
	if errKudosEntry != nil {
		log.Printf("Failed to migrate database: %v", errKudosEntry)
	}

	errPlayerStats := db.AutoMigrate(&domain.PlayerStats{})
	if errPlayerStats != nil {
		log.Printf("Failed to migrate database: %v", errKudosEntry)
	}

	errAttendance := db.AutoMigrate(&domain.AttendanceRecord{})
	if errAttendance != nil {
		log.Printf("Failed to migrate AttendanceRecord: %v", errAttendance)
	}

	// seedDatabase(db)
}

/*func seedDatabase(db *gorm.DB) {
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

	err := db.Where(domain.Profile{ID: hardcodedID}).FirstOrCreate(&hugo).Error
	if err != nil {
		log.Printf("Could not seed database: %v", err)
	}
}
*/

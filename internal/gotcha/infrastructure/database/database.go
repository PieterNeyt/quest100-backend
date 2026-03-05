package database

import (
	"Quest100Backend/internal/gotcha/domain"
	"log"

	"github.com/google/uuid"
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

	seedProps(db)
}

func seedProps(db *gorm.DB) {
	props := []domain.GotchaProp{
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000001"), Name: "Waterpistooltje"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000002"), Name: "Schuimzwaard"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000003"), Name: "Rubberen kip"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000004"), Name: "Nerf-pistool"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000005"), Name: "Bananenschil"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000006"), Name: "Speelgoedmes"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000007"), Name: "Confettikanon"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000008"), Name: "Spiegeltje"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000009"), Name: "Luchtgitaar"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000010"), Name: "Plastic slang"},
	}

	for _, prop := range props {
		if err := db.Where(domain.GotchaProp{ID: prop.ID}).FirstOrCreate(&prop).Error; err != nil {
			log.Printf("Could not seed prop %s: %v", prop.Name, err)
		}
	}
}

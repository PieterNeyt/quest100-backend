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
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000001"), NameNL: "Waterpistooltje", NameEN: "Watergun"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000002"), NameNL: "Schuimzwaard", NameEN: "Foam swoard"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000003"), NameNL: "Rubberen kip", NameEN: "Rubber chicken"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000004"), NameNL: "Nerf-pistool", NameEN: "Nerf gun"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000005"), NameNL: "Bananenschil", NameEN: "Banana peel"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000006"), NameNL: "Speelgoedmes", NameEN: "Toy knife"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000007"), NameNL: "Confettikanon", NameEN: "confetti cannon"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000008"), NameNL: "Spiegeltje", NameEN: "Mirror"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000009"), NameNL: "Boek", NameEN: "Book"},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000010"), NameNL: "Brood", NameEN: "Bread"},
	}

	for _, prop := range props {
		if err := db.Where(domain.GotchaProp{ID: prop.ID}).FirstOrCreate(&prop).Error; err != nil {
			log.Printf("Could not seed prop %s: %v", prop.ID, err)
		}
	}
}

package database

import (
	"Quest100Backend/internal/minigame/nerdle/domain"
	"fmt"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(
		&domain.NerdleGame{},
		&domain.NerdleSession{},
		&domain.NerdleAttempt{},
	); err != nil {
		panic(fmt.Sprintf("nerdle: failed to migrate database: %v", err))
	}
}

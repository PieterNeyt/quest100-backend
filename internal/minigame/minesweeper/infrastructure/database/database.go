package database

import (
	"Quest100Backend/internal/minigame/minesweeper/domain"
	"fmt"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(&domain.MinesweeperSession{}); err != nil {
		panic(fmt.Sprintf("minesweeper: failed to migrate database: %v", err))
	}
}

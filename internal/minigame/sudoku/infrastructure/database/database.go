package database

import (
	"Quest100Backend/internal/minigame/sudoku/domain"
	"fmt"

	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	if err := db.AutoMigrate(&domain.SudokuGame{}, &domain.SudokuSession{}); err != nil {
		panic(fmt.Sprintf("sudoku: failed to migrate database: %v", err))
	}
}

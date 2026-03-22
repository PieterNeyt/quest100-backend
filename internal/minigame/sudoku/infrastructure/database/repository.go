package database

import (
	"Quest100Backend/internal/minigame/sudoku/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SudokuRepository struct {
	db *gorm.DB
}

func NewSudokuRepository(db *gorm.DB) *SudokuRepository {
	return &SudokuRepository{db: db}
}

func (r *SudokuRepository) GetGameByDate(date time.Time) (*domain.SudokuGame, error) {
	var game domain.SudokuGame
	result := r.db.Where("date = ?", date).First(&game)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &domain.GameNotFoundError{Date: date}
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return &game, nil
}

func (r *SudokuRepository) CreateGame(game *domain.SudokuGame) error {
	if err := r.db.Create(game).Error; err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}
	return nil
}

func (r *SudokuRepository) GetSessionByProfileAndGame(profileID uuid.UUID, gameID uuid.UUID) (*domain.SudokuSession, error) {
	var session domain.SudokuSession
	result := r.db.
		Where("profile_id = ? AND game_id = ?", profileID, gameID).
		First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &domain.SessionNotFoundError{ProfileID: profileID, GameID: gameID}
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return &session, nil
}

func (r *SudokuRepository) CreateSession(session *domain.SudokuSession) error {
	if err := r.db.Create(session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *SudokuRepository) SaveSession(session *domain.SudokuSession) error {
	if err := r.db.Save(session).Error; err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

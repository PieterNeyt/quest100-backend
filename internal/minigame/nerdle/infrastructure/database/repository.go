package database

import (
	"Quest100Backend/internal/minigame/nerdle/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NerdleRepository struct {
	db *gorm.DB
}

func NewNerdleRepository(db *gorm.DB) *NerdleRepository {
	return &NerdleRepository{db: db}
}

func (r *NerdleRepository) GetGameByDate(date time.Time) (*domain.NerdleGame, error) {
	var game domain.NerdleGame
	err := r.db.Where("date = ?", date).First(&game).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &domain.GameNotFoundError{Date: date}
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &game, nil
}

func (r *NerdleRepository) CreateGame(game *domain.NerdleGame) error {
	if err := r.db.Create(game).Error; err != nil {
		return fmt.Errorf("failed to create nerdle game: %w", err)
	}
	return nil
}

func (r *NerdleRepository) GetSessionByProfileAndGame(profileID uuid.UUID, gameID uuid.UUID) (*domain.NerdleSession, error) {
	var session domain.NerdleSession
	err := r.db.Where("profile_id = ? AND game_id = ?", profileID, gameID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found for profile %s on game %s", profileID, gameID)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &session, nil
}

func (r *NerdleRepository) CreateSession(session *domain.NerdleSession) error {
	if err := r.db.Create(session).Error; err != nil {
		return fmt.Errorf("failed to create nerdle session: %w", err)
	}
	return nil
}

func (r *NerdleRepository) SaveSession(session *domain.NerdleSession) error {
	if err := r.db.Save(session).Error; err != nil {
		return fmt.Errorf("failed to save nerdle session: %w", err)
	}
	return nil
}

func (r *NerdleRepository) AddAttempt(attempt *domain.NerdleAttempt) error {
	if err := r.db.Create(attempt).Error; err != nil {
		return fmt.Errorf("failed to save nerdle attempt: %w", err)
	}
	return nil
}

func (r *NerdleRepository) GetAttemptsBySession(sessionID uuid.UUID) ([]domain.NerdleAttempt, error) {
	var attempts []domain.NerdleAttempt
	err := r.db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&attempts).Error
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return attempts, nil
}

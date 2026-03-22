package database

import (
	"Quest100Backend/internal/minigame/minesweeper/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MinesweeperRepository struct {
	db *gorm.DB
}

func NewMinesweeperRepository(db *gorm.DB) *MinesweeperRepository {
	return &MinesweeperRepository{db: db}
}

func (r *MinesweeperRepository) GetSessionByProfileAndDate(profileID uuid.UUID, date time.Time) (*domain.MinesweeperSession, error) {
	var session domain.MinesweeperSession
	result := r.db.
		Where("profile_id = ? AND game_date = ?", profileID, date).
		First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &domain.SessionNotFoundError{ProfileID: profileID, Date: date}
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return &session, nil
}

func (r *MinesweeperRepository) CreateSession(session *domain.MinesweeperSession) error {
	if err := r.db.Create(session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *MinesweeperRepository) SaveSession(session *domain.MinesweeperSession) error {
	if err := r.db.Save(session).Error; err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

package database

import (
	"Quest100Backend/internal/minigame/nerdle/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) GetGameByDate(date time.Time) (*domain.NerdleGame, error) {
	var game domain.NerdleGame
	result := r.db.
		Where("date = ?", date).
		First(&game)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &domain.GameNotFoundError{Date: date}
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return &game, nil
}

func (r *GameRepository) CreateGame(game *domain.NerdleGame) error {
	result := r.db.Create(game)
	if result.Error != nil {
		return fmt.Errorf("failed to create nerdle game: %w", result.Error)
	}
	return nil
}

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) GetSessionByProfileAndGame(profileID uuid.UUID, gameID uuid.UUID) (*domain.NerdleSession, error) {
	var session domain.NerdleSession
	result := r.db.
		Where("profile_id = ? AND game_id = ?", profileID, gameID).
		First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found for profile %s on game %s", profileID, gameID)
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return &session, nil
}

func (r *SessionRepository) CreateSession(session *domain.NerdleSession) error {
	result := r.db.Create(session)
	if result.Error != nil {
		return fmt.Errorf("failed to create nerdle session: %w", result.Error)
	}
	return nil
}

func (r *SessionRepository) SaveSession(session *domain.NerdleSession) error {
	result := r.db.Save(session)
	if result.Error != nil {
		return fmt.Errorf("failed to save nerdle session: %w", result.Error)
	}
	return nil
}

func (r *SessionRepository) AddAttempt(attempt *domain.NerdleAttempt) error {
	result := r.db.Create(attempt)
	if result.Error != nil {
		return fmt.Errorf("failed to save nerdle attempt: %w", result.Error)
	}
	return nil
}

func (r *SessionRepository) GetAttemptsBySession(sessionID uuid.UUID) ([]domain.NerdleAttempt, error) {
	var attempts []domain.NerdleAttempt
	result := r.db.
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&attempts)
	if result.Error != nil {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return attempts, nil
}

type NerdleRepository struct {
	gameRepo    *GameRepository
	sessionRepo *SessionRepository
}

func NewNerdleRepository(db *gorm.DB) *NerdleRepository {
	return &NerdleRepository{
		gameRepo:    NewGameRepository(db),
		sessionRepo: NewSessionRepository(db),
	}
}

func (r *NerdleRepository) GetGameByDate(date time.Time) (*domain.NerdleGame, error) {
	return r.gameRepo.GetGameByDate(date)
}

func (r *NerdleRepository) CreateGame(game *domain.NerdleGame) error {
	return r.gameRepo.CreateGame(game)
}

func (r *NerdleRepository) GetSessionByProfileAndGame(profileID uuid.UUID, gameID uuid.UUID) (*domain.NerdleSession, error) {
	return r.sessionRepo.GetSessionByProfileAndGame(profileID, gameID)
}

func (r *NerdleRepository) CreateSession(session *domain.NerdleSession) error {
	return r.sessionRepo.CreateSession(session)
}

func (r *NerdleRepository) SaveSession(session *domain.NerdleSession) error {
	return r.sessionRepo.SaveSession(session)
}

func (r *NerdleRepository) AddAttempt(attempt *domain.NerdleAttempt) error {
	return r.sessionRepo.AddAttempt(attempt)
}

func (r *NerdleRepository) GetAttemptsBySession(sessionID uuid.UUID) ([]domain.NerdleAttempt, error) {
	return r.sessionRepo.GetAttemptsBySession(sessionID)
}

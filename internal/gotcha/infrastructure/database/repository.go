package database

import (
	"Quest100Backend/internal/gotcha/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── GotchaGame ──────────────────────────────────────────────────────────────

type gameRepository struct{ db *gorm.DB }

func NewGameRepository(db *gorm.DB) domain.GameRepository {
	return &gameRepository{db}
}

func (r *gameRepository) SaveGame(game *domain.GotchaGame) error {
	return r.db.Omit("Participants").Save(game).Error
}

func (r *gameRepository) GetGameByCampus(campus string) (*domain.GotchaGame, error) {
	var game domain.GotchaGame
	err := r.db.Where("campus = ?", campus).
		Order("created_at DESC").First(&game).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no active game for campus %s", campus)
	}
	return &game, err
}

func (r *gameRepository) GetGameByFinishedCampus(campus string) (*domain.GotchaGame, error) {
	var game domain.GotchaGame
	err := r.db.Where("campus = ? AND status = ?", campus, domain.StatusFinished).
		Order("updated_at DESC").First(&game).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no finished game for campus %s", campus)
	}
	return &game, err
}

func (r *gameRepository) GetGameByID(id uuid.UUID) (*domain.GotchaGame, error) {
	var game domain.GotchaGame
	err := r.db.First(&game, "id = ?", id).Error
	return &game, err
}

func (r *gameRepository) GetAllActiveGames() ([]*domain.GotchaGame, error) {
	var games []*domain.GotchaGame
	err := r.db.Where("status = ?", domain.StatusActive).Find(&games).Error
	return games, err
}

func (r *gameRepository) GetAllOptInGames() ([]*domain.GotchaGame, error) {
	var games []*domain.GotchaGame
	err := r.db.Where("status = ?", domain.StatusOptIn).Find(&games).Error
	return games, err
}

// ─── Participant ──────────────────────────────────────────────────────────────

type participantRepository struct{ db *gorm.DB }

func NewParticipantRepository(db *gorm.DB) domain.ParticipantRepository {
	return &participantRepository{db}
}

func (r *participantRepository) SaveParticipant(p *domain.Participant) error {
	return r.db.Save(p).Error
}

func (r *participantRepository) GetParticipantsByGame(gameID uuid.UUID) ([]*domain.Participant, error) {
	var list []*domain.Participant
	err := r.db.Where("game_id = ?", gameID).Find(&list).Error
	return list, err
}

func (r *participantRepository) GetParticipant(gameID, profileID uuid.UUID) (*domain.Participant, error) {
	var p domain.Participant
	err := r.db.Where("game_id = ? AND profile_id = ?", gameID, profileID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *participantRepository) GetExpiredParticipants(before interface{}) ([]*domain.Participant, error) {
	var list []*domain.Participant
	activeGameIDs := r.db.Model(&domain.GotchaGame{}).
		Select("id").
		Where("status = ?", domain.StatusActive)
	err := r.db.
		Where("is_alive = true").
		Where("target_id IS NOT NULL").
		Where("kill_deadline < ?", before).
		Where("pending_kill_at IS NULL").
		Where("game_id IN (?)", activeGameIDs).
		Find(&list).Error
	return list, err
}

func (r *participantRepository) DeleteParticipant(gameID, profileID uuid.UUID) error {
	return r.db.Where("game_id = ? AND profile_id = ?", gameID, profileID).
		Delete(&domain.Participant{}).Error
}

// ─── GotchaKill ──────────────────────────────────────────────────────────────

type killRepository struct{ db *gorm.DB }

func NewKillRepository(db *gorm.DB) domain.KillRepository {
	return &killRepository{db}
}

func (r *killRepository) SaveKill(kill *domain.GotchaKill) error {
	return r.db.Save(kill).Error
}

func (r *killRepository) GetKillByID(id uuid.UUID) (*domain.GotchaKill, error) {
	var kill domain.GotchaKill
	err := r.db.Preload("Likes").First(&kill, "id = ?", id).Error
	return &kill, err
}

func (r *killRepository) GetKillFeed(gameID uuid.UUID, limit, offset int) ([]*domain.GotchaKill, error) {
	var kills []*domain.GotchaKill
	err := r.db.Preload("Likes").Where("game_id = ?", gameID).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&kills).Error
	return kills, err
}

func (r *killRepository) GetApprovedKillFeed(gameID uuid.UUID, limit, offset int) ([]*domain.GotchaKill, error) {
	var kills []*domain.GotchaKill
	err := r.db.Preload("Likes").
		Where("game_id = ? AND status = ?", gameID, domain.KillApproved).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&kills).Error
	return kills, err
}

func (r *killRepository) GetApprovedKillsByGame(gameID uuid.UUID) ([]*domain.GotchaKill, error) {
	var kills []*domain.GotchaKill
	err := r.db.Where("game_id = ? AND status = ?", gameID, domain.KillApproved).
		Order("created_at ASC").Find(&kills).Error
	return kills, err
}

func (r *killRepository) GetPendingKills(gameID uuid.UUID) ([]*domain.GotchaKill, error) {
	var kills []*domain.GotchaKill
	err := r.db.Preload("Likes").
		Where("game_id = ? AND status = ?", gameID, domain.KillPending).
		Order("created_at ASC").Find(&kills).Error
	return kills, err
}

func (r *killRepository) GetOldestPendingKill(gameID uuid.UUID) (*domain.GotchaKill, error) {
	var kill domain.GotchaKill
	err := r.db.Preload("Likes").
		Where("game_id = ? AND status = ?", gameID, domain.KillPending).
		Order("created_at ASC").
		First(&kill).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no pending kills for game %s", gameID)
	}
	return &kill, err
}

func (r *killRepository) HasPendingKill(gameID, hunterID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.GotchaKill{}).
		Where("game_id = ? AND hunter_id = ? AND status = ?", gameID, hunterID, domain.KillPending).
		Count(&count).Error
	return count > 0, err
}

func (r *killRepository) CountPendingKills(gameID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&domain.GotchaKill{}).
		Where("game_id = ? AND status = ?", gameID, domain.KillPending).
		Count(&count).Error
	return int(count), err
}

func (r *killRepository) SaveKillLike(like *domain.GotchaKillLike) error {
	return r.db.Save(like).Error
}

func (r *killRepository) DeleteKillLike(killID, profileID uuid.UUID) error {
	return r.db.Where("kill_id = ? AND profile_id = ?", killID, profileID).
		Delete(&domain.GotchaKillLike{}).Error
}

func (r *killRepository) HasLiked(killID, profileID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.GotchaKillLike{}).
		Where("kill_id = ? AND profile_id = ?", killID, profileID).
		Count(&count).Error
	return count > 0, err
}

// ─── Prop ────────────────────────────────────────────────────────────────────
type propRepository struct{ db *gorm.DB }

func NewPropRepository(db *gorm.DB) domain.PropRepository {
	return &propRepository{db}
}

func (r *propRepository) GetRandomProp() (*domain.GotchaProp, error) {
	var prop domain.GotchaProp
	err := r.db.Order("RANDOM()").First(&prop).Error
	return &prop, err
}

func (r *propRepository) GetPropByID(id uuid.UUID) (*domain.GotchaProp, error) {
	var prop domain.GotchaProp
	err := r.db.First(&prop, "id = ?", id).Error
	return &prop, err
}

func (r *propRepository) SaveProp(prop *domain.GotchaProp) error {
	return r.db.Save(prop).Error
}

func (r *propRepository) DeleteProp(id uuid.UUID) error {
	return r.db.Delete(&domain.GotchaProp{}, "id = ?", id).Error
}

func (r *propRepository) GetAllProps() ([]*domain.GotchaProp, error) {
	var props []*domain.GotchaProp
	err := r.db.Order("name_en ASC").Find(&props).Error
	return props, err
}

var _ = time.Now

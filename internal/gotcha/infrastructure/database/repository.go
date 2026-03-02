package database

import (
	"Quest100Backend/internal/gotcha/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gameRepository struct{ db *gorm.DB }

func NewGameRepository(db *gorm.DB) domain.GameRepository {
	return &gameRepository{db}
}

func (r *gameRepository) SaveGame(game *domain.Game) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(game).Error
}

func (r *gameRepository) GetGameByCampus(campus string) (*domain.Game, error) {
	var game domain.Game
	err := r.db.Where("campus = ? AND status != ?", campus, domain.StatusFinished).
		Order("created_at DESC").First(&game).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no active game for campus %s", campus)
	}
	return &game, err
}

func (r *gameRepository) GetGameByID(id uuid.UUID) (*domain.Game, error) {
	var game domain.Game
	err := r.db.First(&game, "id = ?", id).Error
	return &game, err
}

func (r *gameRepository) GetAllActiveGames() ([]*domain.Game, error) {
	var games []*domain.Game
	err := r.db.Where("status = ?", domain.StatusActive).Find(&games).Error
	return games, err
}

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

func (r *participantRepository) GetExpiredParticipants(before time.Time) ([]*domain.Participant, error) {
	var list []*domain.Participant
	err := r.db.
		Joins("JOIN gotcha_games ON gotcha_games.id = gotcha_participants.game_id").
		Where("gotcha_participants.is_alive = true AND gotcha_participants.target_id IS NOT NULL AND gotcha_participants.kill_deadline < ? AND gotcha_games.status = ?", before, domain.StatusActive).
		Find(&list).Error
	return list, err
}

type killRepository struct{ db *gorm.DB }

func NewKillRepository(db *gorm.DB) domain.KillRepository {
	return &killRepository{db}
}

func (r *killRepository) SaveKill(kill *domain.Kill) error {
	return r.db.Save(kill).Error
}

func (r *killRepository) GetKillByID(id uuid.UUID) (*domain.Kill, error) {
	var kill domain.Kill
	err := r.db.First(&kill, "id = ?", id).Error
	return &kill, err
}

func (r *killRepository) GetKillFeed(gameID uuid.UUID, limit, offset int) ([]*domain.Kill, error) {
	var kills []*domain.Kill
	err := r.db.
		Preload("Likes").
		Where("game_id = ? AND status = ?", gameID, domain.KillApproved).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&kills).Error
	return kills, err
}

func (r *killRepository) GetPendingKills(gameID uuid.UUID) ([]*domain.Kill, error) {
	var kills []*domain.Kill
	err := r.db.Where("game_id = ? AND status = ?", gameID, domain.KillPending).Find(&kills).Error
	return kills, err
}

func (r *killRepository) SaveKillLike(like *domain.KillLike) error {
	return r.db.Save(like).Error
}

func (r *killRepository) DeleteKillLike(killID, profileID uuid.UUID) error {
	return r.db.Where("kill_id = ? AND profile_id = ?", killID, profileID).Delete(&domain.KillLike{}).Error
}

type propRepository struct{ db *gorm.DB }

func NewPropRepository(db *gorm.DB) domain.PropRepository {
	return &propRepository{db}
}

func (r *propRepository) GetRandomProp() (*domain.Prop, error) {
	var prop domain.Prop
	err := r.db.Order("RANDOM()").First(&prop).Error
	return &prop, err
}

func (r *propRepository) SaveProp(prop *domain.Prop) error {
	return r.db.Save(prop).Error
}

func (r *propRepository) GetAllProps() ([]*domain.Prop, error) {
	var props []*domain.Prop
	err := r.db.Find(&props).Error
	return props, err
}

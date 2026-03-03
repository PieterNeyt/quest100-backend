package database

import (
	"Quest100Backend/internal/profile/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *profileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) GetProfiles() (*[]domain.Profile, error) {
	var profiles []domain.Profile

	result := r.db.Find(&profiles)

	if result.Error != nil {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	return &profiles, nil
}

func (r *profileRepository) GetProfileById(profileId uuid.UUID) (*domain.Profile, error) {
	var profile domain.Profile

	result := r.db.Debug().
		Preload("KudosHistory").
		Preload("AttendanceRecords").
		First(&profile, "id = ?", profileId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("profile with id %s not found", profileId)
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	return &profile, nil
}

func (r *profileRepository) UpdateProfile(profile *domain.Profile) error {
	result := r.db.Where("id = ?", profile.ID).Save(profile)
	if result.Error != nil {
		return fmt.Errorf("failed to save profile: %w", result.Error)
	}
	return nil
}

func (r *profileRepository) SaveProfile(profile *domain.Profile) error {
	result := r.db.Save(profile)
	if result.Error != nil {
		return fmt.Errorf("failed to save profile: %w", result.Error)
	}
	return nil
}

func (r *profileRepository) AddAwardHistoryEntry(senderId uuid.UUID, receiverId uuid.UUID) error {
	entry := domain.AwardHistoryEntry{
		RecieverID: receiverId,
		SenderID:   senderId,
	}

	result := r.db.Create(&entry)
	if result.Error != nil {
		return fmt.Errorf("failed to save award history entry: %w", result.Error)
	}

	return nil
}

func (r *profileRepository) HasSentAward(senderId uuid.UUID, receiverId uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Table("award_history_entries").
		Where("sender_id = ? AND reciever_id = ?", senderId, receiverId).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *profileRepository) GetSentAwardReceivers(senderId uuid.UUID) ([]uuid.UUID, error) {
	var receiverIds []uuid.UUID

	err := r.db.
		Table("award_history_entries").
		Where("sender_id = ?", senderId).
		Pluck("reciever_id", &receiverIds).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch sent award receivers: %w", err)
	}

	return receiverIds, nil
}

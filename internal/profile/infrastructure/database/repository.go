package database

import (
	"Quest100Backend/internal/profile/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) GetProfiles() (*[]domain.Profile, error) {
	var profiles []domain.Profile

	result := r.db.Find(&profiles)

	if result.Error != nil {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	return &profiles, nil
}

func (r *ProfileRepository) GetProfileById(profileId uuid.UUID) (*domain.Profile, error) {
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

func (r *ProfileRepository) UpdateProfile(profile *domain.Profile) error {
	result := r.db.Where("id = ?", profile.ID).Save(profile)
	if result.Error != nil {
		return fmt.Errorf("failed to save profile: %w", result.Error)
	}
	return nil
}

func (r *ProfileRepository) SaveProfile(profile *domain.Profile) error {
	result := r.db.Omit("PlayerStats").Create(profile)
	if result.Error != nil {
		return result.Error
	}

	result = r.db.Create(&profile.PlayerStats)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *ProfileRepository) AddAwardHistoryEntry(senderId uuid.UUID, receiverId uuid.UUID) error {
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

func (r *ProfileRepository) HasSentAward(senderId uuid.UUID, receiverId uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Table("award_history_entries").
		Where("sender_id = ? AND reciever_id = ?", senderId, receiverId).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ProfileRepository) GetSentAwardReceivers(senderId uuid.UUID) ([]uuid.UUID, error) {
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

func (r *ProfileRepository) GetProfileStats(profileId uuid.UUID) (domain.ProfileStats, error) {
	var stats domain.ProfileStats

	result := r.db.
		Table("profile_stats").
		Where("profile_id = ?", profileId).
		First(&stats)

	if result.Error != nil {
		return domain.ProfileStats{}, fmt.Errorf("failed to fetch profile statistics: %w", result.Error)
	}

	return stats, nil
}
func (r *ProfileRepository) GetCampusByProfileID(profileId uuid.UUID) (string, error) {
	var campus string
	result := r.db.
		Model(&domain.Profile{}).
		Select("campus").
		Where("id = ?", profileId).
		Scan(&campus)
	if result.Error != nil {
		return "", fmt.Errorf("failed to fetch campus: %w", result.Error)
	}
	if campus == "" {
		return "", fmt.Errorf("campus not set on profile, sync first")
	}
	return campus, nil
}

func (r *ProfileRepository) GetLastKudosEntries(profileId uuid.UUID, limit int) ([]domain.KudosEntry, error) {
	var entries []domain.KudosEntry

	result := r.db.
		Where("profile_id = ?", profileId).
		Order("date DESC").
		Limit(limit).
		Find(&entries)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch kudos entries: %w", result.Error)
	}

	return entries, nil
}

func (r *ProfileRepository) GetKudoEntryById(id uuid.UUID) (*domain.KudosEntry, error) {
	var entry domain.KudosEntry
	if err := r.db.First(&entry, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

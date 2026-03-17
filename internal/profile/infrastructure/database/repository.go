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

func (r *ProfileRepository) GetDefaultBodyAsset() (*domain.Asset, error) {
	var asset domain.Asset
	result := r.db.First(&asset, "name = ? AND category = ?", "blue gopher", "Body")
	if result.Error != nil {
		return nil, result.Error
	}
	return &asset, nil
}

func (r *ProfileRepository) GetDefaultEyesAsset() (*domain.Asset, error) {
	var asset domain.Asset
	result := r.db.First(&asset, "name = ? AND category = ?", "crazy eyes", "Eyes")
	if result.Error != nil {
		return nil, result.Error
	}
	return &asset, nil
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
		Preload("Avatar").Preload("Avatar.Body").Preload("Avatar.Eyes").Preload("Avatar.Shirts").
		Preload("Avatar.Hair").Preload("Avatar.FacialHair").Preload("Avatar.Glasses").Preload("Avatar.Accessories").
		Preload("Avatar.Extras").
		Preload("Assets").
		First(&profile, "id = ?", profileId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("profile with id %s not found", profileId)
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	return &profile, nil
}

func (r *ProfileRepository) SaveProfile(profile *domain.Profile) error {

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("PlayerStats", "KudosHistory", "Avatar", "Assets").Save(profile).Error; err != nil {
			return err
		}

		if err := tx.Save(&profile.Avatar).Error; err != nil {
			return err
		}

		if err := tx.Save(profile).Error; err != nil {
			return err
		}
		return nil
	})
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

func (r *ProfileRepository) GetAllAssets() (*[]domain.Asset, error) {
	var assets []domain.Asset
	result := r.db.Find(&assets)

	if result.Error != nil {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return &assets, nil
}

func (r *ProfileRepository) GetProfileAssets(profileId uuid.UUID) (*[]domain.Asset, error) {
	var assets []domain.Asset
	if err := r.db.Model(&domain.Profile{ID: profileId}).Association("Assets").Find(&assets); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &assets, nil
}

func (r *ProfileRepository) GetProfileAvatar(profileId uuid.UUID) (*domain.Avatar, error) {
	var avatar domain.Avatar
	if err := r.db.Model(&domain.Profile{ID: profileId}).Preload("Body").Preload("Eyes").Preload("Shirts").
		Preload("Hair").Preload("FacialHair").Preload("Glasses").Preload("Accessories").
		Preload("Extras").Association("Avatar").Find(&avatar); err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &avatar, nil
}

func (r *ProfileRepository) GetAssetById(assetId string) (*domain.Asset, error) {
	var asset domain.Asset
	result := r.db.First(&asset, "id = ?", assetId)
	if result.Error != nil {
		return nil, fmt.Errorf("asset with id %s not found: %w", assetId, result.Error)
	}
	return &asset, nil
}

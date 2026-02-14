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

func NewProfileRepository(db *gorm.DB) domain.ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) GetProfileById(profileId uuid.UUID) (*domain.Profile, error) {
	var profile domain.Profile

	result := r.db.Debug().Preload("KudosHistory").First(&profile, "id = ?", profileId)

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

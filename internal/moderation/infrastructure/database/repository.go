package database

import (
	"Quest100Backend/internal/moderation/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModerationRepository struct {
	db *gorm.DB
}

func NewModerationRepository(db *gorm.DB) *ModerationRepository {
	return &ModerationRepository{db: db}
}

func (r *ModerationRepository) SaveReport(report *domain.Report) error {
	result := r.db.Create(report)
	if result.Error != nil {
		return fmt.Errorf("failed to save report: %w", result.Error)
	}
	return nil
}

func (r *ModerationRepository) GetReports() (*[]domain.Report, error) {
	var reports []domain.Report

	result := r.db.Find(&reports)
	if result.Error != nil {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	return &reports, nil
}

func (r *ModerationRepository) GetReportById(id uuid.UUID) (*domain.Report, error) {
	var report domain.Report

	result := r.db.First(&report, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("report with id %s not found", id)
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	return &report, nil
}

func (r *ModerationRepository) ResolveReport(id uuid.UUID) error {
	result := r.db.Model(&domain.Report{}).
		Where("id = ?", id).
		Update("resolved", true)

	if result.Error != nil {
		return fmt.Errorf("failed to resolve report: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("report with id %s not found", id)
	}

	return nil
}

func (r *ModerationRepository) HasOpenReport(targetID uuid.UUID, channelType domain.ChannelType) (bool, error) {
	var count int64
	result := r.db.Model(&domain.Report{}).
		Where("target_id = ? AND channel_type = ? AND resolved = false", targetID, channelType).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check open reports: %w", result.Error)
	}

	return count > 0, nil
}

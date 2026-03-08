package database

import (
	"Quest100Backend/internal/moderation/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) SaveReport(report *domain.Report) error {
	result := r.db.Create(report)
	if result.Error != nil {
		return fmt.Errorf("failed to save report: %w", result.Error)
	}
	return nil
}

func (r *ReportRepository) GetReports() (*[]domain.Report, error) {
	var reports []domain.Report

	result := r.db.Find(&reports)
	if result.Error != nil {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	return &reports, nil
}

func (r *ReportRepository) GetReportById(id uuid.UUID) (*domain.Report, error) {
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

func (r *ReportRepository) ResolveReport(id uuid.UUID) error {
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

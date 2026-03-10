package application

import (
	"Quest100Backend/internal/moderation/domain"
	"fmt"

	"github.com/google/uuid"
)

type ModerationService interface {
	CreateReport(userID, targetID uuid.UUID, channelType domain.ChannelType, reportType domain.ReportType, message string) (*domain.Report, error)
	GetReports() (*[]domain.Report, error)
	GetReportById(id uuid.UUID) (*domain.Report, error)
	ResolveReport(id uuid.UUID) error
}

type moderationService struct {
	moderationRepo domain.ModerationRepository
}

func NewModerationService(moderationRepo domain.ModerationRepository) ModerationService {
	return &moderationService{moderationRepo: moderationRepo}
}

func (s *moderationService) CreateReport(userID, targetID uuid.UUID, channelType domain.ChannelType, reportType domain.ReportType, message string) (*domain.Report, error) {
	report, err := domain.NewReport(userID, targetID, channelType, reportType, message)
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	if err := s.moderationRepo.SaveReport(report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	return report, nil
}

func (s *moderationService) GetReports() (*[]domain.Report, error) {
	return s.moderationRepo.GetReports()
}

func (s *moderationService) GetReportById(id uuid.UUID) (*domain.Report, error) {
	return s.moderationRepo.GetReportById(id)
}

func (s *moderationService) ResolveReport(id uuid.UUID) error {
	if err := s.moderationRepo.ResolveReport(id); err != nil {
		return fmt.Errorf("failed to resolve report: %w", err)
	}
	return nil
}

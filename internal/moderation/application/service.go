package application

import (
	"Quest100Backend/internal/moderation/domain"
	"fmt"

	"github.com/google/uuid"
)

type ReportService interface {
	CreateReport(userID, targetID uuid.UUID, channelType domain.ChannelType, reportType domain.ReportType, message string) (*domain.Report, error)
	GetReports() (*[]domain.Report, error)
	GetReportById(id uuid.UUID) (*domain.Report, error)
	ResolveReport(id uuid.UUID) error
}

type reportService struct {
	reportRepo domain.ReportRepository
}

func NewReportService(reportRepo domain.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

func (s *reportService) CreateReport(userID, targetID uuid.UUID, channelType domain.ChannelType, reportType domain.ReportType, message string) (*domain.Report, error) {
	report, err := domain.NewReport(userID, targetID, channelType, reportType, message)
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	if err := s.reportRepo.SaveReport(report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	return report, nil
}

func (s *reportService) GetReports() (*[]domain.Report, error) {
	return s.reportRepo.GetReports()
}

func (s *reportService) GetReportById(id uuid.UUID) (*domain.Report, error) {
	return s.reportRepo.GetReportById(id)
}

func (s *reportService) ResolveReport(id uuid.UUID) error {
	if err := s.reportRepo.ResolveReport(id); err != nil {
		return fmt.Errorf("failed to resolve report: %w", err)
	}
	return nil
}

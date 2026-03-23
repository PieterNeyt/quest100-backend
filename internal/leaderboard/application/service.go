package application

import (
	"Quest100Backend/internal/leaderboard/api/dto"
	"Quest100Backend/internal/leaderboard/domain"
	profileDomain "Quest100Backend/internal/profile/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LeaderboardService interface {
	GetAllCoursesWithClasses() ([]*profileDomain.Course, error)
	CreateLeaderboard(req dto.CreateLeaderboardRequest) (*domain.Leaderboard, error)
	UpdateLeaderboard(id uuid.UUID, req dto.UpdateLeaderboardRequest) (*domain.Leaderboard, error)
	GetAllLeaderboards() ([]*domain.Leaderboard, error)
	GetLeaderboardByID(id uuid.UUID) (*domain.Leaderboard, error)
	GetLeaderboardByCourseId(id uuid.UUID) ([]*domain.Leaderboard, error)
	UpdateLeaderboardStatuses() error
}

type leaderboardService struct {
	leaderboardRepo domain.LeaderboardRepository
}

func NewLeaderboardService(leaderboardRepo domain.LeaderboardRepository) LeaderboardService {
	return &leaderboardService{
		leaderboardRepo: leaderboardRepo,
	}
}

func (s *leaderboardService) UpdateLeaderboardStatuses() error {
	return s.leaderboardRepo.UpdateLeaderboardStatuses()
}

func (s *leaderboardService) GetLeaderboardByID(id uuid.UUID) (*domain.Leaderboard, error) {
	return s.leaderboardRepo.GetLeaderboardByID(id)
}

func (s *leaderboardService) GetLeaderboardByCourseId(id uuid.UUID) ([]*domain.Leaderboard, error) {
	return s.leaderboardRepo.GetLeaderboardByCourseId(id)
}

func (s *leaderboardService) GetAllLeaderboards() ([]*domain.Leaderboard, error) {
	return s.leaderboardRepo.GetAllLeaderboards()
}

func (s *leaderboardService) GetAllCoursesWithClasses() ([]*profileDomain.Course, error) {
	return s.leaderboardRepo.GetAllCoursesWithClasses()
}

func (s *leaderboardService) CreateLeaderboard(req dto.CreateLeaderboardRequest) (*domain.Leaderboard, error) {
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	overlaps, err := s.leaderboardRepo.HasOverlappingLeaderboard(req.CourseID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to check for overlapping leaderboards: %w", err)
	}
	if overlaps {
		return nil, fmt.Errorf("leaderboard overlaps with an existing leaderboard for this course")
	}

	classes, err := s.leaderboardRepo.GetClassesByCourseID(req.CourseID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch classes for course: %w", err)
	}

	now := time.Now()
	isActive := !now.Before(req.StartDate) && now.Before(req.EndDate)

	lb := &domain.Leaderboard{
		ID:        uuid.New(),
		CourseId:  req.CourseID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Active:    isActive,
		Prize: domain.Prize{
			Name:        req.Prize.Name,
			Description: req.Prize.Description,
			PhotoURL:    req.Prize.PhotoURL,
		},
	}

	for _, class := range classes {
		lb.Classes = append(lb.Classes, &domain.LeaderboardClass{
			LeaderboardID: lb.ID,
			ClassId:       class.Id,
			TotalKudos:    0,
		})
	}

	if err := s.leaderboardRepo.CreateLeaderboard(lb); err != nil {
		return nil, err
	}
	return lb, nil
}

func (s *leaderboardService) UpdateLeaderboard(id uuid.UUID, req dto.UpdateLeaderboardRequest) (*domain.Leaderboard, error) {
	lb, err := s.leaderboardRepo.GetLeaderboardByID(id)
	if err != nil {
		return nil, fmt.Errorf("leaderboard not found: %w", err)
	}

	if req.StartDate != nil {
		lb.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		lb.EndDate = *req.EndDate
	}
	if lb.EndDate.Before(lb.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	overlaps, err := s.leaderboardRepo.HasOverlappingLeaderboardExcludingId(lb.CourseId, lb.StartDate, lb.EndDate, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check for overlapping leaderboards: %w", err)
	}
	if overlaps {
		return nil, fmt.Errorf("leaderboard overlaps with an existing leaderboard for this course")
	}

	if req.Prize != nil {
		lb.Prize = domain.Prize{
			Name:        req.Prize.Name,
			Description: req.Prize.Description,
			PhotoURL:    req.Prize.PhotoURL,
		}
	}

	now := time.Now()
	lb.Active = !now.Before(lb.StartDate) && now.Before(lb.EndDate)

	if err := s.leaderboardRepo.UpdateLeaderboard(lb); err != nil {
		return nil, err
	}
	return lb, nil
}

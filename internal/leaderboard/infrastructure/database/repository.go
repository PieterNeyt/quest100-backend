package database

import (
	"Quest100Backend/internal/leaderboard/domain"
	profileDomain "Quest100Backend/internal/profile/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaderboardRepository struct {
	db *gorm.DB
}

func NewLeaderboardRepository(db *gorm.DB) *LeaderboardRepository {
	return &LeaderboardRepository{db: db}
}

func (r *LeaderboardRepository) GetAllCoursesWithClasses() ([]*profileDomain.Course, error) {
	var courses []*profileDomain.Course
	if err := r.db.Preload("Classes").Find(&courses).Error; err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *LeaderboardRepository) GetClassesByCourseID(courseID uuid.UUID) ([]*profileDomain.Class, error) {
	var classes []*profileDomain.Class
	if err := r.db.Where("course_id = ?", courseID).Find(&classes).Error; err != nil {
		return nil, err
	}
	return classes, nil
}

func (r *LeaderboardRepository) CreateLeaderboard(lb *domain.Leaderboard) error {
	return r.db.Create(lb).Error
}

func (r *LeaderboardRepository) GetLeaderboardByID(id uuid.UUID) (*domain.Leaderboard, error) {
	var lb domain.Leaderboard
	if err := r.db.Preload("Classes").First(&lb, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &lb, nil
}

func (r *LeaderboardRepository) GetLeaderboardByCourseId(id uuid.UUID) ([]*domain.Leaderboard, error) {
	var leaderboards []*domain.Leaderboard
	if err := r.db.
		Preload("Classes").
		Where("course_id = ?", id).
		Find(&leaderboards).Error; err != nil {
		return nil, err
	}
	return leaderboards, nil
}

func (r *LeaderboardRepository) HasOverlappingLeaderboard(courseID uuid.UUID, startDate, endDate time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Leaderboard{}).
		Where("course_id = ? AND start_date < ? AND end_date > ?", courseID, endDate, startDate).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}
	return count > 0, nil
}

func (r *LeaderboardRepository) HasOverlappingLeaderboardExcludingId(courseID uuid.UUID, startDate, endDate time.Time, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Leaderboard{}).
		Where("course_id = ? AND start_date < ? AND end_date > ? AND id != ?", courseID, endDate, startDate, excludeID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}
	return count > 0, nil
}

func (r *LeaderboardRepository) UpdateLeaderboardStatuses() error {
	now := time.Now()

	if err := r.db.Model(&domain.Leaderboard{}).
		Where("end_date < ? AND active = ?", now, true).
		Update("active", false).Error; err != nil {
		return fmt.Errorf("failed to close finished leaderboards: %w", err)
	}

	if err := r.db.Model(&domain.Leaderboard{}).
		Where("start_date <= ? AND end_date >= ? AND active = ?", now, now, false).
		Update("active", true).Error; err != nil {
		return fmt.Errorf("failed to activate leaderboards: %w", err)
	}

	return nil
}

func (r *LeaderboardRepository) GetActiveLeaderboardByCourseId(courseID uuid.UUID) (*domain.Leaderboard, error) {
	var lb domain.Leaderboard
	now := time.Now()
	err := r.db.
		Preload("Classes").
		Where("course_id = ? AND start_date <= ? AND end_date >= ?", courseID, now, now).
		First(&lb).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // geen actief leaderboard, geen fout
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &lb, nil
}

func (r *LeaderboardRepository) UpdateLeaderboard(lb *domain.Leaderboard) error {
	return r.db.Model(lb).Updates(map[string]interface{}{
		"start_date":        lb.StartDate,
		"end_date":          lb.EndDate,
		"prize_name":        lb.Prize.Name,
		"prize_description": lb.Prize.Description,
		"prize_photo_url":   lb.Prize.PhotoURL,
	}).Error
}

func (r *LeaderboardRepository) GetAllLeaderboards() ([]*domain.Leaderboard, error) {
	var leaderboards []*domain.Leaderboard
	if err := r.db.Preload("Classes").Find(&leaderboards).Error; err != nil {
		return nil, err
	}
	return leaderboards, nil
}

func (r *LeaderboardRepository) GetLeaderboardWithStandings(id uuid.UUID) (*domain.Leaderboard, error) {
	var lb domain.Leaderboard
	if err := r.db.Preload("Classes").First(&lb, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &lb, nil
}

func (r *LeaderboardRepository) AddKudosToLeaderboardClass(leaderboardId uuid.UUID, classId uuid.UUID, kudos int) error {
	result := r.db.
		Model(&domain.LeaderboardClass{}).
		Where("leaderboard_id = ? AND class_id = ?", leaderboardId, classId).
		UpdateColumn("total_kudos", gorm.Expr("total_kudos + ?", kudos))

	if result.Error != nil {
		return fmt.Errorf("failed to update leaderboard class kudos: %w", result.Error)
	}
	return nil
}

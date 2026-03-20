package database

import (
	"Quest100Backend/internal/leaderboard/domain"
	profileDomain "Quest100Backend/internal/profile/domain"
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

func (r *LeaderboardRepository) GetActiveLeaderboardByCourseId(courseID uuid.UUID) (*domain.Leaderboard, error) {
	var lb domain.Leaderboard
	now := time.Now()
	if err := r.db.
		Preload("Classes").
		Where("course_id = ? AND start_date <= ? AND end_date >= ?", courseID, now, now).
		First(&lb).Error; err != nil {
		return nil, err
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

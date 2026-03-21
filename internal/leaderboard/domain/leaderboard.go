package domain

import (
	"Quest100Backend/internal/profile/domain"
	"time"

	"github.com/google/uuid"
)

type LeaderboardRepository interface {
	GetAllCoursesWithClasses() ([]*domain.Course, error)
	GetClassesByCourseID(courseID uuid.UUID) ([]*domain.Class, error)
	GetAllLeaderboards() ([]*Leaderboard, error)
	CreateLeaderboard(lb *Leaderboard) error
	GetLeaderboardByID(id uuid.UUID) (*Leaderboard, error)
	UpdateLeaderboard(lb *Leaderboard) error
	GetLeaderboardWithStandings(id uuid.UUID) (*Leaderboard, error)
	GetLeaderboardByCourseId(id uuid.UUID) ([]*Leaderboard, error)
	GetActiveLeaderboardByCourseId(courseID uuid.UUID) (*Leaderboard, error)
	HasOverlappingLeaderboard(courseID uuid.UUID, startDate, endDate time.Time) (bool, error)
	HasOverlappingLeaderboardExcludingId(courseID uuid.UUID, startDate, endDate time.Time, excludeID uuid.UUID) (bool, error)
	AddKudosToLeaderboardClass(leaderboardId uuid.UUID, classId uuid.UUID, kudos int) error
}

type LeaderboardClass struct {
	LeaderboardID uuid.UUID `gorm:"type:char(36);primaryKey"`
	ClassId       uuid.UUID `gorm:"type:char(36);primaryKey"`
	ClassName     string    `gorm:"-"`
	TotalKudos    int       `gorm:"default:0"`
}

type Leaderboard struct {
	ID        uuid.UUID           `gorm:"type:char(36);primaryKey"`
	CourseId  uuid.UUID           `gorm:"type:char(36);not null"`
	StartDate time.Time           `gorm:"not null"`
	EndDate   time.Time           `gorm:"not null"`
	Prize     Prize               `gorm:"embedded;embeddedPrefix:prize_"`
	Classes   []*LeaderboardClass `gorm:"foreignKey:LeaderboardID"`
}

func (lb *Leaderboard) IsActive() bool {
	now := time.Now()
	return now.After(lb.StartDate) && now.Before(lb.EndDate)
}

func (lb *Leaderboard) IsFinished() bool {
	return time.Now().After(lb.EndDate)
}

type Prize struct {
	Name        string `gorm:"type:varchar(255)"`
	Description string `gorm:"type:text"`
	PhotoURL    string `gorm:"type:varchar(512)"`
}

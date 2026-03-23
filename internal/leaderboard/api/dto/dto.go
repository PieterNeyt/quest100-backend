package dto

import (
	"time"

	"github.com/google/uuid"
)

type PrizeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PhotoURL    string `json:"photoUrl"`
}

type CreateLeaderboardRequest struct {
	CourseID  uuid.UUID    `json:"courseId" binding:"required"`
	StartDate time.Time    `json:"startDate" binding:"required"`
	EndDate   time.Time    `json:"endDate" binding:"required"`
	Prize     PrizeRequest `json:"prize"`
}

type UpdateLeaderboardRequest struct {
	StartDate *time.Time    `json:"startDate"`
	EndDate   *time.Time    `json:"endDate"`
	Prize     *PrizeRequest `json:"prize"`
}

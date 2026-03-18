package dto

import (
	"time"

	"github.com/google/uuid"
)

type PrizeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PhotoURL    string `json:"photo_url"`
}

type CreateLeaderboardRequest struct {
	CourseID  uuid.UUID    `json:"course_id" binding:"required"`
	StartDate time.Time    `json:"start_date" binding:"required"`
	EndDate   time.Time    `json:"end_date" binding:"required"`
	Prize     PrizeRequest `json:"prize"`
}

type UpdateLeaderboardRequest struct {
	StartDate *time.Time    `json:"start_date"`
	EndDate   *time.Time    `json:"end_date"`
	Prize     *PrizeRequest `json:"prize"`
}

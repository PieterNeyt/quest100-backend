package api

import "time"

type CreateGameRequest struct {
	Campus            string    `json:"campus" binding:"required"`
	StartDate         time.Time `json:"startDate" binding:"required"`
	KillDeadlineHours int       `json:"killDeadlineHours" binding:"required,min=1"`
}

type ReviewKillRequest struct {
	Approve bool `json:"approve"`
}

type SubmitKillRequest struct {
	VictimID string `json:"victimId" binding:"required"`
	PhotoURL string `json:"photoUrl" binding:"required"`
}

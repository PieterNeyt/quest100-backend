package api

import (
	"time"

	"github.com/google/uuid"
)

type CreateGameRequest struct {
	Campus            string    `json:"campus" binding:"required"`
	StartDate         time.Time `json:"startDate"`
	KillDeadlineHours int       `json:"killDeadlineHours"`
}

type UpdateStartDateRequest struct {
	StartDate         time.Time `json:"startDate" binding:"required"`
	KillDeadlineHours int       `json:"killDeadlineHours"`
}

type SubmitKillRequest struct {
	PhotoURL string `json:"photoUrl" binding:"required"`
}

type ReviewKillRequest struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason,omitempty"`
}

type ProfileSummary struct {
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	ProfilePicture *string   `json:"profilePicture"`
}

type PropSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type KillFeedItem struct {
	ID         uuid.UUID      `json:"id"`
	GameID     uuid.UUID      `json:"gameId"`
	PhotoURL   string         `json:"photoUrl"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"createdAt"`
	ReviewedAt *time.Time     `json:"reviewedAt,omitempty"`
	Hunter     ProfileSummary `json:"hunter"`
	Victim     ProfileSummary `json:"victim"`
	Prop       *PropSummary   `json:"prop,omitempty"`
	LikeCount  int            `json:"likeCount"`
	LikedByMe  bool           `json:"likedByMe"`
}

// TargetInfoResponse holds the current player's target and assigned prop.
type TargetInfoResponse struct {
	Target       *ProfileSummary `json:"target"`
	AssignedProp *PropSummary    `json:"assignedProp"`
}

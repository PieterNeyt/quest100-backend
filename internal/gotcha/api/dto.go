package api

import (
	"time"

	"github.com/google/uuid"
)

type CreateGameRequest struct {
	Campus             string    `json:"campus" binding:"required"`
	StartDate          time.Time `json:"startDate"`
	KillDeadlineHours  int       `json:"killDeadlineHours"`
	PrizePhotoBase64   string    `json:"prizePhotoBase64"`
	PrizeDescriptionEN string    `json:"prizeDescriptionEN"`
	PrizeDescriptionNL string    `json:"prizeDescriptionNL"`
}

type UpdateStartDateRequest struct {
	StartDate          time.Time `json:"startDate" binding:"required"`
	KillDeadlineHours  int       `json:"killDeadlineHours"`
	PrizePhotoBase64   string    `json:"prizePhotoBase64"`
	PrizeDescriptionEN string    `json:"prizeDescriptionEN"`
	PrizeDescriptionNL string    `json:"prizeDescriptionNL"`
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

type TargetInfoResponse struct {
	Target       *ProfileSummary `json:"target"`
	AssignedProp *PropSummary    `json:"assignedProp"`
	KillDeadline *time.Time      `json:"killDeadline"`
}

type EndScreenKillNode struct {
	KillID    uuid.UUID      `json:"killId"`
	Hunter    ProfileSummary `json:"hunter"`
	Victim    ProfileSummary `json:"victim"`
	Prop      *PropSummary   `json:"prop,omitempty"`
	PhotoURL  string         `json:"photoUrl"`
	CreatedAt time.Time      `json:"createdAt"`
}

type EndScreenStats struct {
	TotalKills        int    `json:"totalKills"`
	TotalParticipants int    `json:"totalParticipants"`
	FastestKillSecs   int    `json:"fastestKillSecs"`
	MostKillsName     string `json:"mostKillsName"`
	MostKillsCount    int    `json:"mostKillsCount"`
}

type EndScreenResponse struct {
	Winner             *ProfileSummary     `json:"winner"`
	WinnerKillCount    int                 `json:"winnerKillCount"`
	PrizePhotoBase64   string              `json:"prizePhotoBase64,omitempty"`
	PrizeDescriptionEN string              `json:"prizeDescriptionEN,omitempty"`
	PrizeDescriptionNL string              `json:"prizeDescriptionNL,omitempty"`
	Stats              EndScreenStats      `json:"stats"`
	Kills              []EndScreenKillNode `json:"kills"`
}

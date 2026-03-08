package api

import (
	"time"

	"github.com/google/uuid"
)

// ─── Game requests ────────────────────────────────────────────────────────────

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

// ─── Kill requests ────────────────────────────────────────────────────────────

type SubmitKillRequest struct {
	PhotoBase64 string `json:"photoBase64" binding:"required"`
}

type ReviewKillRequest struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason,omitempty"`
}

// ─── Prop requests ────────────────────────────────────────────────────────────

type CreatePropRequest struct {
	NameEN string `json:"nameEN" binding:"required"`
	NameNL string `json:"nameNL" binding:"required"`
}

type UpdatePropRequest struct {
	NameEN string `json:"nameEN" binding:"required"`
	NameNL string `json:"nameNL" binding:"required"`
}

// ─── Shared response types ────────────────────────────────────────────────────

type ProfileSummary struct {
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	ProfilePicture *string   `json:"profilePicture"`
}

// PropSummary is returned inside kill feed / target info — contains both langs.
type PropSummary struct {
	ID     uuid.UUID `json:"id"`
	NameEN string    `json:"nameEN"`
	NameNL string    `json:"nameNL"`
}

type KillFeedItem struct {
	ID          uuid.UUID      `json:"id"`
	GameID      uuid.UUID      `json:"gameId"`
	PhotoBase64 string         `json:"photoBase64"`
	Status      string         `json:"status"`
	CreatedAt   time.Time      `json:"createdAt"`
	ReviewedAt  *time.Time     `json:"reviewedAt,omitempty"`
	Hunter      ProfileSummary `json:"hunter"`
	Victim      ProfileSummary `json:"victim"`
	Prop        *PropSummary   `json:"prop,omitempty"`
	LikeCount   int            `json:"likeCount"`
	LikedByMe   bool           `json:"likedByMe"`
}

type TargetInfoResponse struct {
	Target       *ProfileSummary `json:"target"`
	AssignedProp *PropSummary    `json:"assignedProp"`
	KillDeadline *time.Time      `json:"killDeadline"`
}

type EndScreenKillNode struct {
	KillID      uuid.UUID      `json:"killId"`
	Hunter      ProfileSummary `json:"hunter"`
	Victim      ProfileSummary `json:"victim"`
	Prop        *PropSummary   `json:"prop,omitempty"`
	PhotoBase64 string         `json:"photoBase64"`
	CreatedAt   time.Time      `json:"createdAt"`
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

// PropResponse is the full prop object returned by the props admin endpoints.
type PropResponse struct {
	ID     uuid.UUID `json:"id"`
	NameEN string    `json:"nameEN"`
	NameNL string    `json:"nameNL"`
}

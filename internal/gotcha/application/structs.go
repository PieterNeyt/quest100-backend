package application

import (
	"Quest100Backend/internal/gotcha/domain"
	"time"

	"github.com/google/uuid"
)

type ProfileSummary struct {
	ID             uuid.UUID
	FirstName      string
	LastName       string
	ProfilePicture *string
}

type PropSummary struct {
	ID     uuid.UUID
	NameEN string
	NameNL string
}

type KillFeedItem struct {
	ID          uuid.UUID
	GameID      uuid.UUID
	PhotoBase64 string
	Status      domain.KillStatus
	CreatedAt   time.Time
	ReviewedAt  *time.Time
	Hunter      ProfileSummary
	Victim      ProfileSummary
	Prop        *PropSummary
	LikeCount   int
	LikedByMe   bool
}

type TargetInfo struct {
	Target       *ProfileSummary
	AssignedProp *PropSummary
	KillDeadline *time.Time
}

type EndScreenKillNode struct {
	KillID           uuid.UUID
	Hunter           ProfileSummary
	Victim           ProfileSummary
	Prop             *PropSummary
	PhotoBase64      string
	CreatedAt        time.Time
	LikeCount        int
	TargetAssignedAt *time.Time
}

type EndScreenStats struct {
	TotalKills        int
	TotalParticipants int
	FastestKillSecs   int
	MostKillsName     string
	MostKillsCount    int
}

// AwardCategory represents the category of a game award.
type AwardCategory string

const (
	AwardCategoryCore   AwardCategory = "core"
	AwardCategorySkill  AwardCategory = "skill"
	AwardCategorySocial AwardCategory = "social"
	AwardCategoryProp   AwardCategory = "prop"
	AwardCategoryMeme   AwardCategory = "meme"
	AwardCategoryGame   AwardCategory = "game"
)

type GameAward struct {
	ID             string
	Category       AwardCategory
	TitleKey       string
	DescriptionKey string
	Profile        *ProfileSummary
	Profiles       []ProfileSummary
	Count          *int
	PropName       string
	Day            string
}

type EndScreen struct {
	GameID             uuid.UUID
	Winner             *ProfileSummary
	WinnerKillCount    int
	PrizePhotoBase64   string
	PrizeDescriptionEN string
	PrizeDescriptionNL string
	Stats              EndScreenStats
	Kills              []EndScreenKillNode
	Awards             []GameAward
}

type GameSummary struct {
	ID                 uuid.UUID
	Campus             string
	Status             domain.GameStatus
	StartDate          time.Time
	UpdatedAt          time.Time
	WinnerID           *uuid.UUID
	Winner             *ProfileSummary
	WinnerKillCount    int
	TotalParticipants  int
	TotalKills         int
	PrizeDescriptionEN string
	PrizeDescriptionNL string
}

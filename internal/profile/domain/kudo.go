package domain

import (
	"time"

	"github.com/google/uuid"
)

type KudoType int

const (
	KudoKnowledge KudoType = iota
	KudoAttendance
	KudoTeamwork
	KudoAtmosphere
	KudoEngagement
)

type KudosEntry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;"`
	ProfileID uuid.UUID `gorm:"type:uuid;index;"`
	Amount    int
	Reason    string
	Type      KudoType
	Date      time.Time `gorm:"autoCreateTime"`
}

type AwardHistoryEntry struct {
	RecieverID uuid.UUID `gorm:"type:uuid;primaryKey;"`
	ProfileID  uuid.UUID `gorm:"type:uuid;index;"`
}

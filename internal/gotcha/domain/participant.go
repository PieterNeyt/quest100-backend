package domain

import (
	"time"

	"github.com/google/uuid"
)

func (Participant) TableName() string { return "gotcha_participants" }

type Participant struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	GameID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"gameId"`
	ProfileID    uuid.UUID  `gorm:"type:uuid;not null" json:"profileId"`
	TargetID     *uuid.UUID `gorm:"type:uuid" json:"targetId,omitempty"`
	IsAlive      bool       `gorm:"default:true" json:"isAlive"`
	KillDeadline time.Time  `json:"killDeadline"`
	KilledAt     *time.Time `json:"killedAt,omitempty"`
	KilledBy     *uuid.UUID `gorm:"type:uuid" json:"killedBy,omitempty"`
	OptedInAt    time.Time  `json:"optedInAt"`

	AssignedPropID *uuid.UUID `gorm:"type:uuid" json:"assignedPropId,omitempty"`
	PendingKillAt  *time.Time `json:"pendingKillAt,omitempty"`
}

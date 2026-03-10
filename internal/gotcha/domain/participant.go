package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

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

func (p *Participant) CanSubmitKill() error {
	if !p.IsAlive {
		return fmt.Errorf("you are eliminated and cannot submit kills")
	}
	if p.TargetID == nil {
		return fmt.Errorf("you have no assigned target")
	}
	return nil
}

func (p *Participant) IsTimedOut() bool {
	return p.IsAlive && !p.KillDeadline.IsZero() && p.KillDeadline.Before(time.Now())
}

func (p *Participant) ClearPendingKill() {
	p.PendingKillAt = nil
}

func (p *Participant) MarkPendingKill() {
	now := time.Now()
	p.PendingKillAt = &now
}

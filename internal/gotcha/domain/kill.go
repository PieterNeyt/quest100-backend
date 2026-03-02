package domain

import (
	"time"

	"github.com/google/uuid"
)

type KillStatus string

const (
	KillPending  KillStatus = "PENDING"
	KillApproved KillStatus = "APPROVED"
	KillDenied   KillStatus = "DENIED"
)

type Kill struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	GameID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"gameId"`
	HunterID   uuid.UUID  `gorm:"type:uuid;not null" json:"hunterId"`
	VictimID   uuid.UUID  `gorm:"type:uuid;not null" json:"victimId"`
	PhotoURL   string     `gorm:"type:text;not null" json:"photoUrl"`
	PropID     *uuid.UUID `gorm:"type:uuid" json:"propId,omitempty"`
	Status     KillStatus `gorm:"type:varchar(20)" json:"status"`
	ReviewedBy *uuid.UUID `gorm:"type:uuid" json:"reviewedBy,omitempty"`
	ReviewedAt *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`

	Likes []KillLike `gorm:"foreignKey:KillID" json:"likes,omitempty"`
}

type KillLike struct {
	KillID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"killId"`
	ProfileID uuid.UUID `gorm:"type:uuid;primaryKey" json:"profileId"`
	LikedAt   time.Time `json:"likedAt"`
}

type Prop struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name string    `gorm:"type:varchar(100);not null" json:"name"`
}

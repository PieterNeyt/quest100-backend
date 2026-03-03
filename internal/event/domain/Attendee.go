package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventAttendee struct {
	EventID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"eventId"`
	ProfileID uuid.UUID `gorm:"type:uuid;primaryKey" json:"profileId"`
	JoinedAt  time.Time `json:"joinedAt"`
}

type AttendeeResponse struct {
	EventID   uuid.UUID `json:"eventId"`
	ProfileID uuid.UUID `json:"profileId"`
	JoinedAt  time.Time `json:"joinedAt"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Photo     *string   `json:"customProfilePicture"`
}

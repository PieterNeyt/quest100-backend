package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventAttendee struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventID   uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_event_profile_unique" json:"eventId"`
	ProfileID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_event_profile_unique" json:"profileId"`
	JoinedAt  time.Time `json:"joinedAt"`
}

type AttendeeResponse struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"eventId"`
	ProfileID uuid.UUID `json:"profileId"`
	JoinedAt  time.Time `json:"joinedAt"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Photo     *string   `json:"customProfilePicture"`
}

func (e *Event) IsAttendee(profileID uuid.UUID) bool {
	for _, a := range e.Attendees {
		if a.ProfileID == profileID {
			return true
		}
	}
	return false
}

func (e *Event) Join(profileID uuid.UUID) error {
	if e.IsAttendee(profileID) {
		return &AlreadyAttendingError{ProfileID: profileID, EventID: e.ID}
	}
	if e.IsFull() {
		return &EventFullError{EventID: e.ID, Message: "Event is full"}
	}
	e.Attendees = append(e.Attendees, EventAttendee{
		ID:        uuid.New(),
		EventID:   e.ID,
		ProfileID: profileID,
		JoinedAt:  time.Now(),
	})
	return nil
}

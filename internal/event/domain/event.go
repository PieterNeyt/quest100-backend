package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EventRepository interface {
	SaveEvent(event *Event) error
	GetEventByID(id uuid.UUID) (*Event, error)
	GetAllEvents() ([]*Event, error)
	DeleteEvent(id uuid.UUID) error
	RemoveAttendee(eventID uuid.UUID, profileID uuid.UUID) error
	GetEventByIDWithProfiles(id uuid.UUID) (*EventWithProfiles, error)
}

type EventCategory string

type EventVisibility string

const (
	EventVisibilityPublic EventVisibility = "public"
	EventVisibilityHidden EventVisibility = "hidden"
)

type EventWithProfiles struct {
	Event
	Attendees []AttendeeResponse `json:"attendees"`
}

type Event struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	Title        string          `gorm:"not null" json:"title"`
	Description  string          `gorm:"type:text" json:"description"`
	Photo        *string         `gorm:"type:text" json:"photo"`
	Category     EventCategory   `gorm:"type:varchar(20)" json:"category"`
	Visibility   EventVisibility `gorm:"type:varchar(10);not null;default:'public'" json:"visibility"`
	OrganizerID  uuid.UUID       `gorm:"type:uuid;not null" json:"organizerId"`
	EventDate    time.Time       `json:"eventDate"`
	MaxAttendees *int            `json:"maxAttendees"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`

	Attendees []EventAttendee `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE" json:"attendees"`
}

func CreateEvent(
	title, description string,
	photo *string,
	category EventCategory,
	organizerID uuid.UUID,
	eventDate time.Time,
	maxAttendees *int,
) *Event {
	return &Event{
		ID:           uuid.New(),
		Title:        title,
		Description:  description,
		Photo:        photo,
		Category:     category,
		Visibility:   EventVisibilityPublic,
		OrganizerID:  organizerID,
		EventDate:    eventDate,
		MaxAttendees: maxAttendees,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Attendees:    []EventAttendee{},
	}
}

func (e *Event) Update(
	title, description string,
	photo *string,
	category EventCategory,
	eventDate time.Time,
	maxAttendees *int,
	newOrganizerID *uuid.UUID,
) error {
	e.Title = title
	e.Description = description
	e.Photo = photo
	e.Category = category
	e.EventDate = eventDate
	e.UpdatedAt = time.Now()

	if maxAttendees != nil && len(e.Attendees) > *maxAttendees {
		return fmt.Errorf(
			"cannot reduce max attendees to %d because %d users are already attending",
			*maxAttendees,
			len(e.Attendees),
		)
	}
	e.MaxAttendees = maxAttendees

	if newOrganizerID != nil {
		if !e.IsAttendee(*newOrganizerID) {
			return fmt.Errorf("new organizer must be an existing attendee")
		}
		e.OrganizerID = *newOrganizerID
	}

	return nil
}

func (e *Event) Hide() {
	e.Visibility = EventVisibilityHidden
	e.UpdatedAt = time.Now()
}

func (e *Event) IsOrganizer(profileID uuid.UUID) bool {
	return e.OrganizerID == profileID
}

func (e *Event) IsFull() bool {
	if e.MaxAttendees == nil {
		return false
	}
	return len(e.Attendees) >= *e.MaxAttendees
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
		EventID:   e.ID,
		ProfileID: profileID,
		JoinedAt:  time.Now(),
	})
	return nil
}

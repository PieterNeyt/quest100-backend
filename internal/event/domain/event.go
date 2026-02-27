package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventRepository interface {
	SaveEvent(event *Event) error
	GetEventByID(id uuid.UUID) (*Event, error)
	GetAllEvents() ([]*Event, error)
	GetEventsByCategory(category EventCategory) ([]*Event, error)
	UpdateEvent(event *Event) error
	DeleteEvent(id uuid.UUID) error
	RemoveAttendee(eventID uuid.UUID, profileID uuid.UUID) error
}

type EventCategory string

const (
	CategorySports  EventCategory = "SPORTS"
	CategoryGaming  EventCategory = "GAMING"
	CategoryStudy   EventCategory = "STUDY"
	CategoryFood    EventCategory = "FOOD"
	CategoryMusic   EventCategory = "MUSIC"
	CategoryOutdoor EventCategory = "OUTDOOR"
	CategorySocial  EventCategory = "SOCIAL"
	CategoryOther   EventCategory = "OTHER"
)

type Event struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	Title        string        `gorm:"not null" json:"title"`
	Description  string        `gorm:"type:text" json:"description"`
	Photo        *string       `gorm:"type:text" json:"photo"`
	Category     EventCategory `gorm:"type:varchar(20)" json:"category"`
	OrganizerID  uuid.UUID     `gorm:"type:uuid;not null" json:"organizerId"`
	EventDate    time.Time     `json:"eventDate"`
	MaxAttendees *int          `json:"maxAttendees"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`

	Attendees []EventAttendee `gorm:"foreignKey:EventID;references:ID" json:"attendees"`
}

func CreateEvent(
	title, description string,
	photo *string,
	category EventCategory,
	organizerID uuid.UUID,
	eventDate time.Time,
	maxAttendees *int,
) *Event {
	eventID := uuid.New()
	return &Event{
		ID:           eventID,
		Title:        title,
		Description:  description,
		Photo:        photo,
		Category:     category,
		OrganizerID:  organizerID,
		EventDate:    eventDate,
		MaxAttendees: maxAttendees,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Attendees: []EventAttendee{
			{
				ID:        uuid.New(),
				EventID:   eventID,
				ProfileID: organizerID,
				JoinedAt:  time.Now(),
			},
		},
	}
}

func (e *Event) IsOrganizer(profileID uuid.UUID) bool {
	return e.OrganizerID == profileID
}

func (e *Event) IsFull() bool {
	if e.MaxAttendees == nil {
		return false
	}
	return e.GoingCount() >= *e.MaxAttendees
}

func (e *Event) GoingCount() int {
	return len(e.Attendees)
}

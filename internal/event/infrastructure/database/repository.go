package database

import (
	"Quest100Backend/internal/event/domain"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) domain.EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) SaveEvent(event *domain.Event) error {
	if err := r.db.Session(&gorm.Session{FullSaveAssociations: true}).
		Save(event).Error; err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}
	return nil
}

func (r *eventRepository) GetEventByID(id uuid.UUID) (*domain.Event, error) {
	var event domain.Event
	result := r.db.
		Preload("Attendees").
		First(&event, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("event with id %s not found", id)
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}
	return &event, nil
}

func (r *eventRepository) GetAllEvents() ([]*domain.Event, error) {
	var events []*domain.Event
	if err := r.db.Preload("Attendees").
		Where("visibility = ?", domain.EventVisibilityPublic).
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}
	return events, nil
}

func (r *eventRepository) DeleteEvent(id uuid.UUID) error {
	if err := r.db.Delete(&domain.Event{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

func (r *eventRepository) RemoveAttendee(eventID uuid.UUID, profileID uuid.UUID) error {
	result := r.db.
		Where("event_id = ? AND profile_id = ?", eventID, profileID).
		Delete(&domain.EventAttendee{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove attendee: %w", result.Error)
	}
	return nil
}

func (r *eventRepository) GetEventByIDWithProfiles(id uuid.UUID) (*domain.EventWithProfiles, error) {
	var event domain.Event
	result := r.db.Preload("Attendees").First(&event, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("event with id %s not found", id)
		}
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	enriched := &domain.EventWithProfiles{Event: event, Attendees: []domain.AttendeeResponse{}}

	for _, a := range event.Attendees {
		var p struct {
			FirstName            string  `gorm:"column:first_name"`
			LastName             string  `gorm:"column:last_name"`
			CustomProfilePicture *string `gorm:"column:custom_profile_picture"`
		}
		r.db.Table("profiles").
			Select("first_name, last_name, custom_profile_picture").
			Where("id = ?", a.ProfileID).
			Scan(&p)

		enriched.Attendees = append(enriched.Attendees, domain.AttendeeResponse{
			EventID:   a.EventID,
			ProfileID: a.ProfileID,
			JoinedAt:  a.JoinedAt,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Photo:     p.CustomProfilePicture,
		})
	}
	return enriched, nil
}

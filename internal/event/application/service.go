package application

import (
	"Quest100Backend/internal/event/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CreateEventInput struct {
	Title        string
	Description  string
	Photo        *string
	Category     domain.EventCategory
	OrganizerID  uuid.UUID
	EventDate    time.Time
	MaxAttendees *int
}

type UpdateEventInput struct {
	Title          string
	Description    string
	Photo          *string
	Category       domain.EventCategory
	EventDate      time.Time
	MaxAttendees   *int
	NewOrganizerID *uuid.UUID
}

type EventService interface {
	CreateEvent(input CreateEventInput) (*domain.Event, error)
	GetEventByID(id uuid.UUID) (*domain.Event, error)
	GetEventByIDWithProfiles(id uuid.UUID) (*domain.EventWithProfiles, error)
	GetAllEvents() ([]*domain.Event, error)
	UpdateEvent(eventID uuid.UUID, requestingProfileID uuid.UUID, input UpdateEventInput) (*domain.EventWithProfiles, error)
	DeleteEvent(eventID uuid.UUID, requestingProfileID uuid.UUID) error
	JoinEvent(eventID uuid.UUID, profileID uuid.UUID) (*domain.Event, error)
	LeaveEvent(eventID uuid.UUID, profileID uuid.UUID) error
}

type eventService struct {
	eventRepo domain.EventRepository
}

func NewEventService(eventRepo domain.EventRepository) EventService {
	return &eventService{eventRepo: eventRepo}
}
func (s *eventService) CreateEvent(input CreateEventInput) (*domain.Event, error) {
	event := domain.CreateEvent(
		input.Title,
		input.Description,
		input.Photo,
		input.Category,
		input.OrganizerID,
		input.EventDate,
		input.MaxAttendees,
	)

	if err := s.eventRepo.SaveEvent(event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	return s.JoinEvent(event.ID, input.OrganizerID)
}

func (s *eventService) GetEventByID(id uuid.UUID) (*domain.Event, error) {
	return s.eventRepo.GetEventByID(id)
}

func (s *eventService) GetEventByIDWithProfiles(id uuid.UUID) (*domain.EventWithProfiles, error) {
	return s.eventRepo.GetEventByIDWithProfiles(id)
}

func (s *eventService) GetAllEvents() ([]*domain.Event, error) {
	return s.eventRepo.GetAllEvents()
}

func (s *eventService) UpdateEvent(eventID uuid.UUID, requestingProfileID uuid.UUID, input UpdateEventInput) (*domain.EventWithProfiles, error) {
	event, err := s.eventRepo.GetEventByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}
	if !event.IsOrganizer(requestingProfileID) {
		return nil, &domain.UnauthorizedError{Message: "only the organizer can update this event"}
	}
	if err := event.Update(
		input.Title,
		input.Description,
		input.Photo,
		input.Category,
		input.EventDate,
		input.MaxAttendees,
		input.NewOrganizerID,
	); err != nil {
		return nil, err
	}
	if err := s.eventRepo.SaveEvent(event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}
	// Geef enriched versie terug
	return s.eventRepo.GetEventByIDWithProfiles(eventID)
}

func (s *eventService) DeleteEvent(eventID uuid.UUID, requestingProfileID uuid.UUID) error {
	event, err := s.eventRepo.GetEventByID(eventID)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}
	if !event.IsOrganizer(requestingProfileID) {
		return &domain.UnauthorizedError{Message: "only the organizer can delete this event"}
	}
	return s.eventRepo.DeleteEvent(eventID)
}

func (s *eventService) JoinEvent(eventID uuid.UUID, profileID uuid.UUID) (*domain.Event, error) {
	event, err := s.eventRepo.GetEventByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}
	if err := event.Join(profileID); err != nil {
		return nil, err
	}
	if err := s.eventRepo.SaveEvent(event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}
	return event, nil
}

func (s *eventService) LeaveEvent(eventID uuid.UUID, profileID uuid.UUID) error {
	event, err := s.eventRepo.GetEventByID(eventID)
	if err != nil {
		return fmt.Errorf("event not found: %w", err)
	}
	if event.IsOrganizer(profileID) {
		return &domain.UnauthorizedError{Message: "organizer cannot cancel attendance without transferring ownership or deleting the event"}
	}
	if !event.IsAttendee(profileID) {
		return &domain.NotAttendingError{ProfileID: profileID, EventID: eventID, Message: "not attending this event"}
	}
	if err := s.eventRepo.RemoveAttendee(eventID, profileID); err != nil {
		return fmt.Errorf("failed to cancel attendance: %w", err)
	}
	return nil
}

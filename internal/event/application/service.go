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
	GetAllEvents() ([]*domain.Event, error)
	GetEventsByCategory(category domain.EventCategory) ([]*domain.Event, error)
	UpdateEvent(eventID uuid.UUID, requestingProfileID uuid.UUID, input UpdateEventInput) (*domain.Event, error)
	DeleteEvent(eventID uuid.UUID, requestingProfileID uuid.UUID) error
	JoinEvent(eventID uuid.UUID, profileID uuid.UUID) (*domain.Event, error)
	CancelEvent(eventID uuid.UUID, profileID uuid.UUID) error
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
	return event, nil
}

func (s *eventService) GetEventByID(id uuid.UUID) (*domain.Event, error) {
	return s.eventRepo.GetEventByID(id)
}

func (s *eventService) GetAllEvents() ([]*domain.Event, error) {
	return s.eventRepo.GetAllEvents()
}

func (s *eventService) GetEventsByCategory(category domain.EventCategory) ([]*domain.Event, error) {
	return s.eventRepo.GetEventsByCategory(category)
}

func (s *eventService) UpdateEvent(eventID uuid.UUID, requestingProfileID uuid.UUID, input UpdateEventInput) (*domain.Event, error) {
	event, err := s.eventRepo.GetEventByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}
	if !event.IsOrganizer(requestingProfileID) {
		return nil, &domain.UnauthorizedError{Message: "only the organizer can update this event"}
	}

	event.Title = input.Title
	event.Description = input.Description
	event.Photo = input.Photo
	event.Category = input.Category
	event.EventDate = input.EventDate
	event.MaxAttendees = input.MaxAttendees

	if input.NewOrganizerID != nil {
		if !event.IsAttendee(*input.NewOrganizerID) {
			return nil, fmt.Errorf("new organizer must be an existing attendee")
		}
		event.OrganizerID = *input.NewOrganizerID
	}

	if err := s.eventRepo.UpdateEvent(event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}
	return event, nil
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
	if event.IsAttendee(profileID) {
		return nil, &domain.AlreadyAttendingError{ProfileID: profileID, EventID: eventID}
	}
	if err := event.Join(profileID); err != nil {
		return nil, err
	}
	if err := s.eventRepo.UpdateEvent(event); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}
	return event, nil
}

func (s *eventService) CancelEvent(eventID uuid.UUID, profileID uuid.UUID) error {
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

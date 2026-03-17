package application

import (
	"Quest100Backend/internal/communication/application"
	"Quest100Backend/internal/event/domain"
	moderationApp "Quest100Backend/internal/moderation/application"
	moderationDomain "Quest100Backend/internal/moderation/domain"
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
	GetReportedEvenByID(id uuid.UUID) (*domain.Event, error)
}

type eventService struct {
	eventRepo      domain.EventRepository
	chatServ       application.ChatService
	moderationServ moderationApp.ModerationService
}

func NewEventService(eventRepo domain.EventRepository, chatServ application.ChatService, moderationServ moderationApp.ModerationService) EventService {
	return &eventService{
		eventRepo:      eventRepo,
		chatServ:       chatServ,
		moderationServ: moderationServ,
	}
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
	if err := s.chatServ.CreateChatRoom(event.ID); err != nil {
		return nil, fmt.Errorf("failed to create chat room: %w", err)
	}
	return s.JoinEvent(event.ID, input.OrganizerID)
}

func (s *eventService) GetEventByID(id uuid.UUID) (*domain.Event, error) {
	return s.eventRepo.GetEventByID(id)
}

func (s *eventService) GetReportedEvenByID(id uuid.UUID) (*domain.Event, error) {
	return s.eventRepo.GetReportedEvenByID(id)
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

	HasOpenEventReport, err := s.moderationServ.HasOpenEventReport(eventID, moderationDomain.ChannelTypeEvent)
	if err != nil {
		return fmt.Errorf("failed to check open events reports: %w", err)
	}

	if HasOpenEventReport {
		event.Hide()
		if err := s.eventRepo.SaveEvent(event); err != nil {
			return fmt.Errorf("failed to hide event with open report: %w", err)
		}
		return nil
	}

	hasOpenMessageReport, err := s.moderationServ.HasOpenMessageReport(eventID, moderationDomain.ChannelTypeMessage)
	if err != nil {
		return fmt.Errorf("failed to check open message reports: %w", err)
	}
	if !hasOpenMessageReport {
		err := s.chatServ.DeleteChatRoom(requestingProfileID, eventID)
		if err != nil {
			return err
		}
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
	if err := s.chatServ.JoinChatRoom(profileID, event.ID); err != nil {
		return nil, fmt.Errorf("failed to join chat room: %w", err)
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
	if err := s.chatServ.LeaveChatRoom(profileID, event.ID); err != nil {
		return fmt.Errorf("failed to leave chatroom: %w", err)
	}
	return nil
}

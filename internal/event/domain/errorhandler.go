package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type EventFullError struct {
	EventID uuid.UUID
	Message string
}

func (e *EventFullError) Error() string {
	return fmt.Sprintf("%s (eventId: %s)", e.Message, e.EventID)
}

type NotAttendingError struct {
	ProfileID uuid.UUID
	EventID   uuid.UUID
	Message   string
}

func (e *NotAttendingError) Error() string {
	return fmt.Sprintf("%s (profileId: %s, eventId: %s)", e.Message, e.ProfileID, e.EventID)
}

type AlreadyAttendingError struct {
	ProfileID uuid.UUID
	EventID   uuid.UUID
}

func (e *AlreadyAttendingError) Error() string {
	return fmt.Sprintf("already attending this event (profileId: %s, eventId: %s)", e.ProfileID, e.EventID)
}

type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

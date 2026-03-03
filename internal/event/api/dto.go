package api

import (
	"Quest100Backend/internal/event/domain"
	"time"

	"github.com/google/uuid"
)

type CreateEventRequest struct {
	Title        string               `json:"title" binding:"required"`
	Description  string               `json:"description"`
	Photo        *string              `json:"photo"`
	Category     domain.EventCategory `json:"category" binding:"required"`
	EventDate    time.Time            `json:"eventDate" binding:"required"`
	MaxAttendees *int                 `json:"maxAttendees"`
}

type UpdateEventRequest struct {
	Title          string               `json:"title" binding:"required"`
	Description    string               `json:"description"`
	Photo          *string              `json:"photo"`
	Category       domain.EventCategory `json:"category" binding:"required"`
	EventDate      time.Time            `json:"eventDate" binding:"required"`
	MaxAttendees   *int                 `json:"maxAttendees"`
	NewOrganizerID *uuid.UUID           `json:"organizerID"`
}

package api

import (
	"Quest100Backend/internal/event/application"
	"Quest100Backend/internal/event/domain"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventHandler struct {
	eventService application.EventService
}

func NewEventHandler(eventService application.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

func getProfileID(c *gin.Context) (uuid.UUID, bool) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return uuid.Nil, false
	}
	return profileID.(uuid.UUID), true
}

func (h *EventHandler) GetAllEvents(c *gin.Context) {
	events, err := h.eventService.GetAllEvents()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) GetEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventService.GetEventByIDWithProfiles(eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	profileID, ok := getProfileID(c)
	if !ok {
		return
	}

	var body CreateEventRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.eventService.CreateEvent(application.CreateEventInput{
		Title:        body.Title,
		Description:  body.Description,
		Photo:        body.Photo,
		Category:     body.Category,
		OrganizerID:  profileID,
		EventDate:    body.EventDate,
		MaxAttendees: body.MaxAttendees,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	profileID, ok := getProfileID(c)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var body UpdateEventRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := application.UpdateEventInput{
		Title:          body.Title,
		Description:    body.Description,
		Photo:          body.Photo,
		Category:       body.Category,
		EventDate:      body.EventDate,
		MaxAttendees:   body.MaxAttendees,
		NewOrganizerID: body.NewOrganizerID,
	}

	event, err := h.eventService.UpdateEvent(eventID, profileID, input)
	if err != nil {
		var unauthorized *domain.UnauthorizedError
		if errors.As(err, &unauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) DeleteEvent(c *gin.Context) {
	profileID, ok := getProfileID(c)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := h.eventService.DeleteEvent(eventID, profileID); err != nil {
		var unauthorized *domain.UnauthorizedError
		if errors.As(err, &unauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EventHandler) JoinEvent(c *gin.Context) {
	profileID, ok := getProfileID(c)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventService.JoinEvent(eventID, profileID)
	if err != nil {
		var fullErr *domain.EventFullError
		if errors.As(err, &fullErr) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		var alreadyErr *domain.AlreadyAttendingError
		if errors.As(err, &alreadyErr) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) LeaveEvent(c *gin.Context) {
	profileID, ok := getProfileID(c)
	if !ok {
		return
	}

	eventID, err := uuid.Parse(c.Param("eventId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := h.eventService.LeaveEvent(eventID, profileID); err != nil {
		var unauthorized *domain.UnauthorizedError
		if errors.As(err, &unauthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		var notAttending *domain.NotAttendingError
		if errors.As(err, &notAttending) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

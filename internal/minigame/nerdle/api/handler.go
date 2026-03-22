package api

import (
	"Quest100Backend/internal/minigame/nerdle/application"
	"Quest100Backend/internal/minigame/nerdle/domain"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NerdleHandler struct {
	nerdleService application.NerdleService
}

func NewNerdleHandler(nerdleService application.NerdleService) *NerdleHandler {
	return &NerdleHandler{nerdleService: nerdleService}
}

func (h *NerdleHandler) GetSession(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	session, err := h.nerdleService.GetOrCreateSession(profileID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to load session"})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *NerdleHandler) SubmitGuess(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	var req SubmitGuessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body must contain 'guess'"})
		return
	}

	resp, err := h.nerdleService.SubmitGuess(profileID.(uuid.UUID), req.Guess)
	if err != nil {
		var alreadySolved *domain.AlreadySolvedError
		if errors.As(err, &alreadySolved) {
			c.JSON(http.StatusConflict, gin.H{"error": "You have already solved today's puzzle"})
			return
		}
		var maxAttempts *domain.MaxAttemptsReachedError
		if errors.As(err, &maxAttempts) {
			c.JSON(http.StatusConflict, gin.H{"error": "No attempts remaining for today's puzzle"})
			return
		}
		var invalidGuess *domain.InvalidGuessError
		if errors.As(err, &invalidGuess) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to process guess"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

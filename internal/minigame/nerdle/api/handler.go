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

func (h *NerdleHandler) GetTodayStatus(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}
	profileUUID := profileID.(uuid.UUID)

	game, err := h.nerdleService.GetTodayGame()
	if err != nil {
		var notFound *domain.GameNotFoundError
		if errors.As(err, &notFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No puzzle available for today"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to load today's game"})
		return
	}

	session, err := h.nerdleService.GetOrCreateSession(profileUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to load session"})
		return
	}

	attemptsUsed := len(session.Attempts)
	attemptsLeft := domain.MaxAttempts - attemptsUsed
	if attemptsLeft < 0 {
		attemptsLeft = 0
	}

	c.JSON(http.StatusOK, GameStatusResponse{
		GameID:       game.ID,
		Date:         game.Date,
		HasSession:   true,
		Solved:       session.Solved,
		AttemptsUsed: attemptsUsed,
		AttemptsLeft: attemptsLeft,
		MaxAttempts:  domain.MaxAttempts,
		CompletedAt:  session.CompletedAt,
	})
}

func (h *NerdleHandler) GetSession(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}
	profileUUID := profileID.(uuid.UUID)

	session, err := h.nerdleService.GetOrCreateSession(profileUUID)
	if err != nil {
		var notFound *domain.GameNotFoundError
		if errors.As(err, &notFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No puzzle available for today"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to load session"})
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
	profileUUID := profileID.(uuid.UUID)

	var req SubmitGuessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body must contain 'guess'"})
		return
	}

	resp, err := h.nerdleService.SubmitGuess(profileUUID, req.Guess)
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

		var notFound *domain.GameNotFoundError
		if errors.As(err, &notFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No puzzle available for today"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to process guess"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

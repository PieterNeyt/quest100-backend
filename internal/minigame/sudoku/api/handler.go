package api

import (
	"Quest100Backend/internal/minigame/sudoku/application"
	"Quest100Backend/internal/minigame/sudoku/domain"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SudokuHandler struct {
	sudokuService application.SudokuService
}

func NewSudokuHandler(sudokuService application.SudokuService) *SudokuHandler {
	return &SudokuHandler{sudokuService: sudokuService}
}

func (h *SudokuHandler) GetSession(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	resp, err := h.sudokuService.GetSessionResponse(profileID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to load session"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *SudokuHandler) SubmitMove(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	var req SubmitMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body must contain 'row', 'col', and 'value'"})
		return
	}

	resp, err := h.sudokuService.SubmitMove(profileID.(uuid.UUID), req.Row, req.Col, req.Value)
	if err != nil {
		var alreadyCompleted *domain.AlreadyCompletedError
		if errors.As(err, &alreadyCompleted) {
			c.JSON(http.StatusConflict, gin.H{"error": "Today's sudoku is already completed"})
			return
		}
		var invalidMove *domain.InvalidMoveError
		if errors.As(err, &invalidMove) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to process move"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

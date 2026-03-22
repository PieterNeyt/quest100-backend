package api

import (
	"Quest100Backend/internal/minigame/minesweeper/application"
	"Quest100Backend/internal/minigame/minesweeper/domain"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MinesweeperHandler struct {
	minesweeperService application.MinesweeperService
}

func NewMinesweeperHandler(minesweeperService application.MinesweeperService) *MinesweeperHandler {
	return &MinesweeperHandler{minesweeperService: minesweeperService}
}

func (h *MinesweeperHandler) GetSession(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	resp, err := h.minesweeperService.GetSessionResponse(profileID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to load session"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *MinesweeperHandler) SubmitMove(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	var req SubmitMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request body must contain 'action', 'row', and 'col'"})
		return
	}

	resp, err := h.minesweeperService.SubmitMove(profileID.(uuid.UUID), domain.MoveAction(req.Action), req.Row, req.Col)
	if err != nil {
		var alreadyCompleted *domain.AlreadyCompletedError
		if errors.As(err, &alreadyCompleted) {
			c.JSON(http.StatusConflict, gin.H{"error": "Today's game is already completed"})
			return
		}
		var invalidMove *domain.InvalidMoveError
		if errors.As(err, &invalidMove) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to process move"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

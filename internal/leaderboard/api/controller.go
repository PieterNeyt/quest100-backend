package api

import (
	"Quest100Backend/internal/leaderboard/api/dto"
	"Quest100Backend/internal/leaderboard/application"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LeaderboardHandler struct {
	leaderboardService application.LeaderboardService
}

func NewLeaderboardHandler(leaderboardService application.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
	}
}

func (h *LeaderboardHandler) GetAllCoursesWithClasses(c *gin.Context) {
	courses, err := h.leaderboardService.GetAllCoursesWithClasses()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, courses)
}

func (h *LeaderboardHandler) CreateLeaderboard(c *gin.Context) {
	var req dto.CreateLeaderboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	leaderboard, err := h.leaderboardService.CreateLeaderboard(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, leaderboard)
}

func (h *LeaderboardHandler) UpdateLeaderboard(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid leaderboard id"})
		return
	}

	var req dto.UpdateLeaderboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	leaderboard, err := h.leaderboardService.UpdateLeaderboard(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, leaderboard)
}

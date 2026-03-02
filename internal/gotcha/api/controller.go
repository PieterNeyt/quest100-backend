package api

import (
	"Quest100Backend/internal/gotcha/application"
	"Quest100Backend/internal/gotcha/domain"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GotchaHandler struct {
	service application.GotchaService
}

func NewGotchaHandler(service application.GotchaService) *GotchaHandler {
	return &GotchaHandler{service}
}

func profileID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("profileID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}
	return v.(uuid.UUID), true
}

func campus(c *gin.Context) string {
	v, _ := c.Get("campus")
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func (h *GotchaHandler) CreateGame(c *gin.Context) {
	var body CreateGameRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	game, err := h.service.CreateGame(body.Campus, body.StartDate, body.KillDeadlineHours)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, game)
}

func (h *GotchaHandler) OptIn(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam := campus(c)
	if cam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "campus not found in token"})
		return
	}
	if err := h.service.OptIn(cam, pid); err != nil {
		var alreadyIn *domain.AlreadyOptedInError
		if errors.As(err, &alreadyIn) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "opted in successfully"})
}

func (h *GotchaHandler) SubmitKill(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam := campus(c)

	var body SubmitKillRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	victimID, err := uuid.Parse(body.VictimID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid victimId"})
		return
	}

	status, err := h.service.GetMyStatus(cam, pid)
	if err != nil || status == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not a participant"})
		return
	}

	kill, err := h.service.SubmitKill(status.GameID, pid, victimID, body.PhotoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, kill)
}

func (h *GotchaHandler) ReviewKill(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	killID, err := uuid.Parse(c.Param("killId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid killId"})
		return
	}
	var body ReviewKillRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.ReviewKill(killID, pid, body.Approve); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *GotchaHandler) GetFeed(c *gin.Context) {
	cam := campus(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	kills, err := h.service.GetKillFeed(cam, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, kills)
}

func (h *GotchaHandler) LikeKill(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	killID, _ := uuid.Parse(c.Param("killId"))
	if err := h.service.LikeKill(killID, pid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *GotchaHandler) UnlikeKill(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	killID, _ := uuid.Parse(c.Param("killId"))
	if err := h.service.UnlikeKill(killID, pid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *GotchaHandler) GetMyStatus(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam := campus(c)
	status, err := h.service.GetMyStatus(cam, pid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not a participant"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *GotchaHandler) GetLeaderboard(c *gin.Context) {
	cam := campus(c)
	board, err := h.service.GetLeaderboard(cam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, board)
}

func (h *GotchaHandler) GetPendingKills(c *gin.Context) {
	cam := campus(c)

	status, err := h.service.GetMyStatus(cam, uuid.Nil)
	_ = status
	_ = err
	// TODO: implementeren functie
	c.JSON(http.StatusNotImplemented, gin.H{"message": "implement via service.GetPendingKills(campus)"})
}

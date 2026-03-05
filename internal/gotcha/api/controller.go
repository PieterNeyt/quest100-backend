package api

import (
	"Quest100Backend/internal/gotcha/application"
	"Quest100Backend/internal/gotcha/domain"
	profileApp "Quest100Backend/internal/profile/application"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GotchaHandler struct {
	service        application.GotchaService
	profileService profileApp.ProfileService
}

func NewGotchaHandler(service application.GotchaService, profileService profileApp.ProfileService) *GotchaHandler {
	return &GotchaHandler{service, profileService}
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

func (h *GotchaHandler) getCampus(c *gin.Context, pid uuid.UUID) (string, error) {
	if cam := campus(c); cam != "" {
		return cam, nil
	}
	profile, err := h.profileService.GetProfileById(pid)
	if err != nil {
		return "", err
	}
	if profile.Campus == "" {
		return "", fmt.Errorf("campus not set on profile, sync first")
	}
	return profile.Campus, nil
}

func (h *GotchaHandler) GetCurrentGame(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	game, err := h.service.GetCurrentGame(cam)
	if err != nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	c.JSON(http.StatusOK, game)
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

func (h *GotchaHandler) UpdateStartDate(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var body UpdateStartDateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	game, err := h.service.UpdateStartDate(cam, body.StartDate, body.KillDeadlineHours)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, game)
}

func (h *GotchaHandler) OptIn(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

func (h *GotchaHandler) OptOut(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.OptOut(cam, pid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *GotchaHandler) SubmitKill(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body SubmitKillRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	kill, err := h.service.SubmitKill(cam, pid, body.PhotoURL)
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
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, err := h.service.GetKillFeed(cam, pid, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := make([]KillFeedItem, 0, len(items))
	for _, item := range items {
		apiItem := KillFeedItem{
			ID:         item.ID,
			GameID:     item.GameID,
			PhotoURL:   item.PhotoURL,
			Status:     string(item.Status),
			CreatedAt:  item.CreatedAt,
			ReviewedAt: item.ReviewedAt,
			Hunter: ProfileSummary{
				ID:             item.Hunter.ID,
				FirstName:      item.Hunter.FirstName,
				LastName:       item.Hunter.LastName,
				ProfilePicture: item.Hunter.ProfilePicture,
			},
			Victim: ProfileSummary{
				ID:             item.Victim.ID,
				FirstName:      item.Victim.FirstName,
				LastName:       item.Victim.LastName,
				ProfilePicture: item.Victim.ProfilePicture,
			},
			LikeCount: item.LikeCount,
			LikedByMe: item.LikedByMe,
		}
		if item.Prop != nil {
			apiItem.Prop = &PropSummary{ID: item.Prop.ID, Name: item.Prop.Name}
		}
		response = append(response, apiItem)
	}

	c.JSON(http.StatusOK, response)
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
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	status, err := h.service.GetMyStatus(cam, pid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not a participant"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *GotchaHandler) GetLeaderboard(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	board, err := h.service.GetLeaderboard(cam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, board)
}

func (h *GotchaHandler) GetPendingKills(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = cam
	c.JSON(http.StatusNotImplemented, gin.H{"message": "implement via service.GetPendingKills(campus)"})
}

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

// ─── helpers ──────────────────────────────────────────────────────────────────

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

func propToResponse(p *domain.GotchaProp) PropResponse {
	return PropResponse{ID: p.ID, NameEN: p.NameEN, NameNL: p.NameNL}
}

func propSummaryFromDomain(p *domain.GotchaProp) *PropSummary {
	if p == nil {
		return nil
	}
	return &PropSummary{ID: p.ID, NameEN: p.NameEN, NameNL: p.NameNL}
}

// ─── Game ─────────────────────────────────────────────────────────────────────

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
	game, err := h.service.CreateGame(body.Campus, body.StartDate, body.KillDeadlineHours, body.PrizePhotoBase64, body.PrizeDescriptionEN, body.PrizeDescriptionNL)
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
	game, err := h.service.UpdateStartDate(cam, body.StartDate, body.KillDeadlineHours, body.PrizePhotoBase64, body.PrizeDescriptionEN, body.PrizeDescriptionNL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, game)
}

// ─── Opt-in / out ─────────────────────────────────────────────────────────────

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

// ─── Kills ────────────────────────────────────────────────────────────────────

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
	kill, err := h.service.SubmitKill(cam, pid, body.PhotoBase64)
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

// ─── Feed ─────────────────────────────────────────────────────────────────────

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
			ID:          item.ID,
			GameID:      item.GameID,
			PhotoBase64: item.PhotoBase64,
			Status:      string(item.Status),
			CreatedAt:   item.CreatedAt,
			ReviewedAt:  item.ReviewedAt,
			Hunter:      ProfileSummary{ID: item.Hunter.ID, FirstName: item.Hunter.FirstName, LastName: item.Hunter.LastName, ProfilePicture: item.Hunter.ProfilePicture},
			Victim:      ProfileSummary{ID: item.Victim.ID, FirstName: item.Victim.FirstName, LastName: item.Victim.LastName, ProfilePicture: item.Victim.ProfilePicture},
			LikeCount:   item.LikeCount,
			LikedByMe:   item.LikedByMe,
		}
		if item.Prop != nil {
			apiItem.Prop = &PropSummary{ID: item.Prop.ID, NameEN: item.Prop.NameEN, NameNL: item.Prop.NameNL}
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

// ─── Me ───────────────────────────────────────────────────────────────────────

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

func (h *GotchaHandler) GetTargetInfo(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	info, err := h.service.GetTargetInfo(cam, pid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	resp := TargetInfoResponse{KillDeadline: info.KillDeadline}
	if info.Target != nil {
		resp.Target = &ProfileSummary{ID: info.Target.ID, FirstName: info.Target.FirstName, LastName: info.Target.LastName, ProfilePicture: info.Target.ProfilePicture}
	}
	if info.AssignedProp != nil {
		resp.AssignedProp = &PropSummary{ID: info.AssignedProp.ID, NameEN: info.AssignedProp.NameEN, NameNL: info.AssignedProp.NameNL}
	}
	c.JSON(http.StatusOK, resp)
}

// ─── Leaderboard / end screen ─────────────────────────────────────────────────

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

func (h *GotchaHandler) GetEndScreen(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data, err := h.service.GetEndScreen(cam)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	resp := EndScreenResponse{
		WinnerKillCount:    data.WinnerKillCount,
		PrizePhotoBase64:   data.PrizePhotoBase64,
		PrizeDescriptionEN: data.PrizeDescriptionEN,
		PrizeDescriptionNL: data.PrizeDescriptionNL,
		Stats: EndScreenStats{
			TotalKills:        data.Stats.TotalKills,
			TotalParticipants: data.Stats.TotalParticipants,
			FastestKillSecs:   data.Stats.FastestKillSecs,
			MostKillsName:     data.Stats.MostKillsName,
			MostKillsCount:    data.Stats.MostKillsCount,
		},
	}
	if data.Winner != nil {
		resp.Winner = &ProfileSummary{ID: data.Winner.ID, FirstName: data.Winner.FirstName, LastName: data.Winner.LastName, ProfilePicture: data.Winner.ProfilePicture}
	}
	kills := make([]EndScreenKillNode, 0, len(data.Kills))
	for _, k := range data.Kills {
		node := EndScreenKillNode{
			KillID:           k.KillID,
			Hunter:           ProfileSummary{ID: k.Hunter.ID, FirstName: k.Hunter.FirstName, LastName: k.Hunter.LastName, ProfilePicture: k.Hunter.ProfilePicture},
			Victim:           ProfileSummary{ID: k.Victim.ID, FirstName: k.Victim.FirstName, LastName: k.Victim.LastName, ProfilePicture: k.Victim.ProfilePicture},
			PhotoBase64:      k.PhotoBase64,
			CreatedAt:        k.CreatedAt,
			LikeCount:        k.LikeCount,
			TargetAssignedAt: k.TargetAssignedAt,
		}
		if k.Prop != nil {
			node.Prop = &PropSummary{ID: k.Prop.ID, NameEN: k.Prop.NameEN, NameNL: k.Prop.NameNL}
		}
		kills = append(kills, node)
	}
	resp.Kills = kills
	c.JSON(http.StatusOK, resp)
}

// ─── Pending kill review ──────────────────────────────────────────────────────

func (h *GotchaHandler) GetNextPendingKill(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.GetNextPendingKill(cam, pid)
	if err != nil {
		c.Status(http.StatusNoContent)
		return
	}
	apiItem := KillFeedItem{
		ID:          item.ID,
		GameID:      item.GameID,
		PhotoBase64: item.PhotoBase64,
		Status:      string(item.Status),
		CreatedAt:   item.CreatedAt,
		ReviewedAt:  item.ReviewedAt,
		Hunter:      ProfileSummary{ID: item.Hunter.ID, FirstName: item.Hunter.FirstName, LastName: item.Hunter.LastName, ProfilePicture: item.Hunter.ProfilePicture},
		Victim:      ProfileSummary{ID: item.Victim.ID, FirstName: item.Victim.FirstName, LastName: item.Victim.LastName, ProfilePicture: item.Victim.ProfilePicture},
		LikeCount:   item.LikeCount,
		LikedByMe:   item.LikedByMe,
	}
	if item.Prop != nil {
		apiItem.Prop = &PropSummary{ID: item.Prop.ID, NameEN: item.Prop.NameEN, NameNL: item.Prop.NameNL}
	}
	c.JSON(http.StatusOK, apiItem)
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
	items, err := h.service.GetPendingKills(cam, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response := make([]KillFeedItem, 0, len(items))
	for _, item := range items {
		apiItem := KillFeedItem{
			ID:          item.ID,
			GameID:      item.GameID,
			PhotoBase64: item.PhotoBase64,
			Status:      string(item.Status),
			CreatedAt:   item.CreatedAt,
			Hunter:      ProfileSummary{ID: item.Hunter.ID, FirstName: item.Hunter.FirstName, LastName: item.Hunter.LastName, ProfilePicture: item.Hunter.ProfilePicture},
			Victim:      ProfileSummary{ID: item.Victim.ID, FirstName: item.Victim.FirstName, LastName: item.Victim.LastName, ProfilePicture: item.Victim.ProfilePicture},
		}
		if item.Prop != nil {
			apiItem.Prop = &PropSummary{ID: item.Prop.ID, NameEN: item.Prop.NameEN, NameNL: item.Prop.NameNL}
		}
		response = append(response, apiItem)
	}
	c.JSON(http.StatusOK, response)
}

func (h *GotchaHandler) GetPendingKillCount(c *gin.Context) {
	pid, ok := profileID(c)
	if !ok {
		return
	}
	cam, err := h.getCampus(c, pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	count, err := h.service.GetPendingKillCount(cam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// ─── Props CRUD ───────────────────────────────────────────────────────────────

func (h *GotchaHandler) GetAllProps(c *gin.Context) {
	props, err := h.service.GetAllProps()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]PropResponse, 0, len(props))
	for _, p := range props {
		response = append(response, propToResponse(p))
	}
	c.JSON(http.StatusOK, response)
}

func (h *GotchaHandler) CreateProp(c *gin.Context) {
	var body CreatePropRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prop, err := h.service.CreateProp(body.NameEN, body.NameNL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, propToResponse(prop))
}

func (h *GotchaHandler) UpdateProp(c *gin.Context) {
	propID, err := uuid.Parse(c.Param("propId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid propId"})
		return
	}
	var body UpdatePropRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prop, err := h.service.UpdateProp(propID, body.NameEN, body.NameNL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, propToResponse(prop))
}

func (h *GotchaHandler) DeleteProp(c *gin.Context) {
	propID, err := uuid.Parse(c.Param("propId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid propId"})
		return
	}
	if err := h.service.DeleteProp(propID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

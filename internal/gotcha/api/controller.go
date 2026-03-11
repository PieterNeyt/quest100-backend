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

// Game

func (h *GotchaHandler) GetCurrentGame(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}
	game, err := h.service.GetCurrentGame(pid)
	if err != nil {
		c.Status(http.StatusNoContent)
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

// History

func (h *GotchaHandler) GetGameHistory(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}

	summaries, err := h.service.GetGameHistory(pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response := make([]GameSummaryResponse, 0, len(summaries))
	for _, s := range summaries {
		item := GameSummaryResponse{
			ID:                 s.ID,
			Campus:             s.Campus,
			Status:             string(s.Status),
			StartDate:          s.StartDate,
			FinishedAt:         s.UpdatedAt,
			WinnerKillCount:    s.WinnerKillCount,
			TotalParticipants:  s.TotalParticipants,
			TotalKills:         s.TotalKills,
			PrizeDescriptionEN: s.PrizeDescriptionEN,
			PrizeDescriptionNL: s.PrizeDescriptionNL,
		}
		if s.Winner != nil {
			item.Winner = &ProfileSummary{ID: s.Winner.ID, FirstName: s.Winner.FirstName, LastName: s.Winner.LastName, ProfilePicture: s.Winner.ProfilePicture}
		}
		response = append(response, item)
	}
	c.JSON(http.StatusOK, response)
}

// Opt-in / out

func (h *GotchaHandler) OptIn(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}

	if err := h.service.OptIn(pid); err != nil {
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
	pid, ok := h.profile(c)
	if !ok {
		return
	}

	if err := h.service.OptOut(pid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Kills

func (h *GotchaHandler) SubmitKill(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}

	var body SubmitKillRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	kill, err := h.service.SubmitKill(pid, body.PhotoBase64)
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
	killID, ok := parseUUIDParam(c, "killId")
	if !ok {
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

// Feed

func (h *GotchaHandler) GetFeed(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, err := h.service.GetKillFeed(pid, limit, offset)
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

func (h *GotchaHandler) GetMyStatus(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}
	status, err := h.service.GetMyStatus(pid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not a participant"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *GotchaHandler) GetTargetInfo(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}

	info, err := h.service.GetTargetInfo(pid)
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

// Leaderboard / end screen

func (h *GotchaHandler) GetLeaderboard(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}
	board, err := h.service.GetLeaderboard(pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, board)
}

func (h *GotchaHandler) GetEndScreen(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}
	data, err := h.service.GetEndScreen(pid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, buildEndScreenResponse(data))
}

func (h *GotchaHandler) GetEndScreenByID(c *gin.Context) {
	gameID, ok := parseUUIDParam(c, "gameId")
	if !ok {
		return
	}
	data, err := h.service.GetEndScreenByID(gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, buildEndScreenResponse(data))
}

func buildEndScreenResponse(data *application.EndScreen) EndScreenResponse {
	resp := EndScreenResponse{
		GameID:             data.GameID,
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
	return resp
}

// Pending kill review

func (h *GotchaHandler) GetNextPendingKill(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}
	item, err := h.service.GetNextPendingKill(pid)
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
	pid, ok := h.profile(c)
	if !ok {
		return
	}
	items, err := h.service.GetPendingKills(pid)
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
	pid, ok := h.profile(c)
	if !ok {
		return
	}
	count, err := h.service.GetPendingKillCount(pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

// Props CRUD

func (h *GotchaHandler) GetAllProps(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}
	props, err := h.service.GetAllPropsByGame(pid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response := make([]PropResponse, 0, len(props))
	for _, p := range props {
		response = append(response, propToResponse(p))
	}
	c.JSON(http.StatusOK, response)
}

func (h *GotchaHandler) CreateProp(c *gin.Context) {
	pid, ok := h.profile(c)
	if !ok {
		return
	}
	var body CreatePropRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prop, err := h.service.CreateProp(pid, body.NameEN, body.NameNL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, propToResponse(prop))
}

func (h *GotchaHandler) UpdateProp(c *gin.Context) {
	propID, ok := parseUUIDParam(c, "propId")
	if !ok {
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
	propID, ok := parseUUIDParam(c, "propId")
	if !ok {
		return
	}
	if err := h.service.DeleteProp(propID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

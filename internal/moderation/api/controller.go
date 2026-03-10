package api

import (
	"Quest100Backend/internal/moderation/application"
	"Quest100Backend/internal/moderation/domain"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ModerationHandler struct {
	moderationService application.ModerationService
}

func NewModerationHandler(moderationService application.ModerationService) *ModerationHandler {
	return &ModerationHandler{moderationService: moderationService}
}

func (h *ModerationHandler) CreateReport(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}
	profileUUID := profileID.(uuid.UUID)

	var body CreateReportRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.moderationService.CreateReport(
		profileUUID,
		body.TargetID,
		domain.ChannelType(body.ChannelType),
		domain.ReportType(body.ReportType),
		body.Message,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

func (h *ModerationHandler) GetReports(c *gin.Context) {
	reports, err := h.moderationService.GetReports()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

func (h *ModerationHandler) ResolveReport(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("reportId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID format"})
		return
	}

	if err := h.moderationService.ResolveReport(reportID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "report resolved"})
}

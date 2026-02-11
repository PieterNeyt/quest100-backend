package api

import (
	"Quest100Backend/internal/profile/application"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	profileService application.ProfileService
}

func NewProfileHandler(profileService application.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

func (h *ProfileHandler) HandleAttendance(c *gin.Context) {

	classId, err := uuid.Parse(c.Param("classId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID format"})
		return
	}

	//TIJDELIJKE OPLOSSING: We gebruiken een hardcoded profile ID omdat we nog geen authenticatie hebben
	profileId, err := uuid.Parse(os.Getenv("HARDCODED_PROFILE_ID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID configuration"})
		return
	}

	profile, err := h.profileService.HandleAttendance(classId, profileId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Attendance recorded successfully",
		"profile": profile,
	})
}

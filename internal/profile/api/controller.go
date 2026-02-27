package api

import (
	"Quest100Backend/internal/profile/application"
	"Quest100Backend/internal/profile/domain"
	"net/http"

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
func (h *ProfileHandler) UpdateLanguage(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}
	profileId := profileID.(uuid.UUID)

	var input struct {
		Language string `json:"language" binding:"required,oneof=NL EN"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid language. Use NL or EN"})
		return
	}

	profile, err := h.profileService.GetProfileById(profileId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	profile.PreferredLanguage = domain.Language(input.Language)

	if err := h.profileService.UpdateProfile(profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update language"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "language updated",
		"language": input.Language,
	})
}

func (h *ProfileHandler) HandleAttendance(c *gin.Context) {

	classId, err := uuid.Parse(c.Param("classId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid class ID format"})
		return
	}

	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}
	profileId := profileID.(uuid.UUID)

	profile, kudosEarned, alreadyRegistered, err := h.profileService.HandleAttendance(classId, profileId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if alreadyRegistered {
		c.JSON(http.StatusOK, gin.H{
			"message":           "Attendance already registered",
			"alreadyRegistered": true,
			"kudosEarned":       kudosEarned,
			"profile":           profile,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":           "Attendance recorded successfully",
		"alreadyRegistered": false,
		"kudosEarned":       kudosEarned,
		"profile":           profile,
	})
}

func (h *ProfileHandler) Sync(c *gin.Context) {
	token := c.GetHeader("X-Graph-Token")

	graphProfile, err := h.profileService.GetGraphProfile(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.profileService.Sync(graphProfile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	microsoftPicture := ""

	base64Img, err := h.profileService.GetGraphProfilePicture(token)
	if err == nil {
		microsoftPicture = base64Img
	}

	response := SyncProfileResponse{
		Profile:                 profile,
		MicrosoftProfilePicture: microsoftPicture,
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProfileHandler) UpdateProfilePicture(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}
	profileId := profileID.(uuid.UUID)

	var body struct {
		ProfilePicture string `json:"profilePicture" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	profile, err := h.profileService.UpdateProfilePicture(profileId, body.ProfilePicture)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *ProfileHandler) DeleteProfilePicture(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}
	profileId := profileID.(uuid.UUID)

	profile, err := h.profileService.DeleteProfilePicture(profileId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *ProfileHandler) GiveAwardTo(c *gin.Context) {
	var transaction AwardTransaction

	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	senderID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	senderUUID, ok := senderID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID format"})
		return
	}

	if err := h.profileService.GiveAwardTo(senderUUID, transaction.Receiver, transaction.Type, transaction.Message); err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kudos succesfully given to"})
}
func (h *ProfileHandler) GetProfiles(c *gin.Context) {
	profiles, err := h.profileService.GetProfiles()
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, profiles)
}

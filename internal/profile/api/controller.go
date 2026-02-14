package api

import (
	"Quest100Backend/internal/profile/application"
	"Quest100Backend/internal/profile/domain"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

func (h *ProfileHandler) extractProfileIDFromToken(c *gin.Context) (uuid.UUID, error) {

	authHeader := c.GetHeader("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, _, _ := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	claims := token.Claims.(jwt.MapClaims)

	oidClaim, ok := claims["oid"].(string)
	if !ok {
		return uuid.Nil, &InvalidTokenError{Message: "Missing oid claim"}
	}

	return uuid.Parse(oidClaim)
}

func (h *ProfileHandler) UpdateLanguage(c *gin.Context) {
	profileId, err := h.extractProfileIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

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

	profile.PrefferedLanguage = domain.Language(input.Language)

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

	profileId, err := h.extractProfileIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	profile, kudosEarned, alreadyRegistered, err := h.profileService.HandleAttendance(classId, profileId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := "Attendance recorded successfully"
	if alreadyRegistered {
		message = "Attendance already registered"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           message,
		"alreadyRegistered": alreadyRegistered,
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

	c.JSON(http.StatusOK, profile)
}

type InvalidTokenError struct {
	Message string
}

func (e *InvalidTokenError) Error() string {
	return e.Message
}

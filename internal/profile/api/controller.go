package api

import (
	"Quest100Backend/internal/profile/application"
	"Quest100Backend/internal/profile/domain"
	"net/http"
	"slices"
	"strconv"

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
	classIdStr := c.Param("classId")
	classId, err := strconv.Atoi(classIdStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Class ID not found"})
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

		if profile.CustomProfilePicture == nil {
			updatedProfile, updateErr := h.profileService.UpdateProfilePicture(profile.ID, base64Img)
			if updateErr == nil {
				profile = updatedProfile
			}
		}
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
func (h *ProfileHandler) GetProfileById(c *gin.Context) {
	idStr := c.Param("id")
	profileId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}

	profile, err := h.profileService.GetProfileById(profileId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	profile, err := h.profileService.GiveAwardTo(
		senderUUID,
		transaction.Receiver,
		transaction.Type,
		transaction.Message,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, profile)
}
func (h *ProfileHandler) GetProfiles(c *gin.Context) {
	profiles, err := h.profileService.GetProfiles()
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, profiles)
}

func (h *ProfileHandler) GetProfilesForAwards(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	profileUUID, ok := profileID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID format"})
		return
	}

	var ProfileAwards, err = h.profileService.GetProfilesWithAward(profileUUID)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, ProfileAwards)
}

func (h *ProfileHandler) GetProfilesStatistics(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	profileUUID, ok := profileID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID format"})
		return
	}

	ProfileStats, err := h.profileService.GetProfilesStatistics(profileUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, ProfileStats)
}

func (h *ProfileHandler) GetLastKudosEntries(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	profileUUID, ok := profileID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID format"})
		return
	}

	entries, err := h.profileService.GetLastKudosEntries(profileUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entries)
}

func (h *ProfileHandler) GetAvatarItems(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	profileUUID, ok := profileID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID format"})
		return
	}

	assets, err := h.profileService.GetAllAssets()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profileAssets, err := h.profileService.GetProfileAssets(profileUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	equippedAssets, err := h.profileService.GetEquippedAssets(profileUUID)

	categoryMap := make(map[string]*CategoryDTO)
	var categoryOrder []string

	for _, asset := range assets {
		if _, exists := categoryMap[asset.Category]; !exists {
			categoryMap[asset.Category] = &CategoryDTO{Name: asset.Category, Items: []AssetDTO{}}
			categoryOrder = append(categoryOrder, asset.Category)
		}

		categoryMap[asset.Category].Items = append(categoryMap[asset.Category].Items, AssetDTO{
			ID:       asset.ID,
			Name:     asset.Name,
			Category: asset.Category,
			Price:    asset.Price,
			IsOwned: slices.ContainsFunc(profileAssets, func(item domain.Asset) bool {
				return item.ID == asset.ID
			}),
			Link: asset.Link,
			Equipped: slices.ContainsFunc(equippedAssets, func(item domain.Asset) bool {
				return item.ID == asset.ID
			}),
			Thumbnail: asset.Thumbnail,
		})
	}

	var responseArray []CategoryDTO

	for _, catName := range categoryOrder {
		responseArray = append(responseArray, *categoryMap[catName])
	}

	c.JSON(http.StatusOK, responseArray)
}

func (h *ProfileHandler) BuyAvatarItem(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
	}

	profileUUID, ok := profileID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID format"})
		return
	}

	avatarId := c.Param("id")
	if avatarId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id format"})
		return
	}

	if err := h.profileService.BuyAsset(profileUUID, avatarId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *ProfileHandler) ToggleAvatarItem(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	profileUUID, ok := profileID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID format"})
		return
	}

	avatarId := c.Param("id")
	if avatarId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id format"})
		return
	}

	assets, err := h.profileService.ToggleAsset(profileUUID, avatarId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var assetArray []AssetDTO
	for _, asset := range assets {
		assetArray = append(assetArray, AssetDTO{
			ID:        asset.ID,
			Name:      asset.Name,
			Category:  asset.Category,
			Price:     asset.Price,
			IsOwned:   true,
			Link:      asset.Link,
			Equipped:  true,
			Thumbnail: asset.Thumbnail,
		})
	}

	c.JSON(http.StatusOK, assetArray)
}

func (h *ProfileHandler) ProxyAsset(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url required"})
		return
	}

	contentType, body, err := h.profileService.FetchExternalAsset(url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer body.Close()

	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Cache-Control", "public, max-age=86400")
	c.DataFromReader(http.StatusOK, -1, contentType, body, nil)
}

func (h *ProfileHandler) GetKudoEntrieById(c *gin.Context) {
	idStr := c.Param("id")
	kudoId, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kudo entry id"})
		return
	}

	entry, err := h.profileService.GetKudoEntryById(kudoId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "kudo entry not found"})
		return
	}

	c.JSON(http.StatusOK, entry)
}
func (h *ProfileHandler) GetTodayAgenda(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Profile ID not found"})
		return
	}

	items, err := h.profileService.GetTodayAgendaWithAttendance(profileID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

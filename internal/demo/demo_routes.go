package demo

import (
	"Quest100Backend/internal/profile/application"
	"Quest100Backend/internal/profile/domain"
	"Quest100Backend/internal/profile/infrastructure/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SetupDemoRoutes registers the demo login endpoint.
// In production: delete this file and remove the one call in RegisterRoutes.
func SetupDemoRoutes(r *gin.Engine, db *gorm.DB) {
	profileRepo := database.NewProfileRepository(db)
	profileService := application.NewProfileService(profileRepo)

	r.POST("/api/demo/login", func(c *gin.Context) {
		var body struct {
			FirstName string `json:"firstName" binding:"required"`
			LastName  string `json:"lastName" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "firstName and lastName required"})
			return
		}

		// Create a deterministic UUID based on name so same name = same account
		// Using uuid v5 with a fixed namespace
		namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
		profileID := uuid.NewSHA1(namespace, []byte("demo:"+body.FirstName+":"+body.LastName))

		// Get or create the demo profile
		profile, err := profileService.GetProfileById(profileID)
		if err != nil {
			// Profile doesn't exist yet, create it
			graphProfile := &domain.GraphProfile{
				Id:                profileID,
				Name:              body.FirstName,
				Surname:           body.LastName,
				Mail:              "demo_" + body.FirstName + "_" + body.LastName + "@demo.local",
				PreferredLanguage: domain.ENG,
			}
			profile, err = profileService.Sync(graphProfile)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create demo profile"})
				return
			}
		}

		token, err := GenerateDemoToken(profile.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":   token,
			"profile": profile,
		})
	})
}

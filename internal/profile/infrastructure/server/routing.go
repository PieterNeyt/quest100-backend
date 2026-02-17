package server

import (
	"Quest100Backend/internal/profile/api"
	"Quest100Backend/internal/profile/application"
	"Quest100Backend/internal/profile/infrastructure/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupProfileRoutes(r *gin.RouterGroup, db *gorm.DB) {
	profileRepo := database.NewProfileRepository(db)
	profileService := application.NewProfileService(profileRepo)
	profileHandler := api.NewProfileHandler(profileService)

	profileGroup := r.Group("/profiles")
	{
		profileGroup.POST("/attendance/:classId", profileHandler.HandleAttendance)
		profileGroup.GET("/sync", profileHandler.Sync)
		profileGroup.PUT("/language", profileHandler.UpdateLanguage)
	}
}

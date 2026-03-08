package server

import (
	"Quest100Backend/internal/profile/api"
	"Quest100Backend/internal/profile/application"

	"github.com/gin-gonic/gin"
)

func SetupProfileRoutes(r *gin.RouterGroup, profServ application.ProfileService) {
	profileHandler := api.NewProfileHandler(profServ)

	profileGroup := r.Group("/profiles")
	{
		profileGroup.POST("/attendance/:classId", profileHandler.HandleAttendance)
		profileGroup.POST("/award", profileHandler.GiveAwardTo)
		profileGroup.GET("/award", profileHandler.GetProfilesForAwards)
		profileGroup.GET("/stats", profileHandler.GetProfilesStatistics)
		profileGroup.GET("/sync", profileHandler.Sync)
		profileGroup.GET("", profileHandler.GetProfiles)
		profileGroup.PUT("/language", profileHandler.UpdateLanguage)
		profileGroup.PUT("/picture", profileHandler.UpdateProfilePicture)
		profileGroup.DELETE("/picture", profileHandler.DeleteProfilePicture)
	}
}

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
		profileGroup.GET("/kudos/recent", profileHandler.GetLastKudosEntries)
		profileGroup.GET("/kudos/:id", profileHandler.GetKudoEntrieById)
		profileGroup.PUT("/language", profileHandler.UpdateLanguage)
		profileGroup.PUT("/class", profileHandler.UpdateClass)
		profileGroup.PUT("/picture", profileHandler.UpdateProfilePicture)
		profileGroup.DELETE("/picture", profileHandler.DeleteProfilePicture)
		profileGroup.GET("/:id", profileHandler.GetProfileById)
		profileGroup.GET("/assets", profileHandler.GetAvatarItems)
		profileGroup.PUT("/assets/:id", profileHandler.BuyAvatarItem)
		profileGroup.PUT("/avatar/:id", profileHandler.ToggleAvatarItem)
		profileGroup.GET("/proxy/asset", profileHandler.ProxyAsset)
	}
}

package server

import (
	"Quest100Backend/internal/moderation/api"
	"Quest100Backend/internal/moderation/application"

	"github.com/gin-gonic/gin"
)

func SetupModerationRoutes(r *gin.RouterGroup, modSer application.ModerationService) {
	modHandler := api.NewModerationHandler(modSer)

	moderationGroup := r.Group("/moderation")
	{
		moderationGroup.POST("/report", modHandler.CreateReport)
	}
}

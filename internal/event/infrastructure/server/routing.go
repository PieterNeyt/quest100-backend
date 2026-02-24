package server

import (
	"Quest100Backend/internal/event/api"
	"Quest100Backend/internal/event/application"
	"Quest100Backend/internal/event/infrastructure/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupEventRoutes(r *gin.RouterGroup, db *gorm.DB) {
	eventRepo := database.NewEventRepository(db)
	eventService := application.NewEventService(eventRepo)
	eventHandler := api.NewEventHandler(eventService)

	eventGroup := r.Group("/events")
	{
		eventGroup.GET("", eventHandler.GetAllEvents)
		eventGroup.POST("", eventHandler.CreateEvent)
		eventGroup.GET("/:eventId", eventHandler.GetEvent)
		eventGroup.PUT("/:eventId", eventHandler.UpdateEvent)
		eventGroup.DELETE("/:eventId", eventHandler.DeleteEvent)
		eventGroup.POST("/:eventId/attendance", eventHandler.JoinEvent)
		eventGroup.DELETE("/:eventId/attendance", eventHandler.CancelEvent)

	}
}

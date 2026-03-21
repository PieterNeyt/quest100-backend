package server

import (
	"Quest100Backend/internal/event/api"
	"Quest100Backend/internal/event/application"
	"Quest100Backend/internal/infrastructure/auth"

	"github.com/gin-gonic/gin"
)

func SetupEventRoutes(r *gin.RouterGroup, eventServ application.EventService) {
	eventHandler := api.NewEventHandler(eventServ)

	eventGroup := r.Group("/events")
	{
		eventGroup.GET("", eventHandler.GetAllEvents)
		eventGroup.POST("", eventHandler.CreateEvent)
		eventGroup.GET("/:eventId", eventHandler.GetEvent)
		eventGroup.GET("/:eventId/reported", auth.RequireRole(auth.Admin), eventHandler.GetReportedEvenByID)
		eventGroup.PUT("/:eventId", eventHandler.UpdateEvent)
		eventGroup.DELETE("/:eventId", eventHandler.DeleteEvent)
		eventGroup.POST("/:eventId/attendance", eventHandler.JoinEvent)
		eventGroup.DELETE("/:eventId/attendance", eventHandler.LeaveEvent)

	}
}

package server

import (
	"Quest100Backend/internal/communication/api"
	"Quest100Backend/internal/communication/application"

	"github.com/gin-gonic/gin"
)

func SetupCommunicationsRoutes(r *gin.RouterGroup, chatServ application.ChatService) {
	commHandler := api.NewCommunicationHandler(chatServ)

	chatGroup := r.Group("/chat")
	{
		chatGroup.GET("/:id", commHandler.GetAllMessagesOfChatroom)
	}
}

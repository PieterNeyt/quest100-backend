package api

import (
	"Quest100Backend/internal/communication/application"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CommunicationHandler struct {
	chatService application.ChatService
}

func NewCommunicationHandler(chatService application.ChatService) *CommunicationHandler {
	return &CommunicationHandler{chatService: chatService}
}

func (h *CommunicationHandler) GetAllMessagesOfChatroom(c *gin.Context) {
	profileID, exists := c.Get("profileID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no profile id found"})
		return
	}
	profileId := profileID.(uuid.UUID)
	chatId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messages, err := h.chatService.GetAllMessagesOfChatRoom(chatId, profileId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, messages)
}

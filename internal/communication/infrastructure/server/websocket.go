package server

import (
	"Quest100Backend/internal/communication/application"
	"Quest100Backend/internal/communication/domain"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:4200"
	},
}

func handleWebSocket(c *gin.Context, hub *Hub) {
	token := c.Query("token")
	if token == "" {
		log.Printf("Profile ID not found")
		return
	}

	id, err := uuid.Parse(token)
	if err != nil {
		log.Printf("Profile ID is not a UUID")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade: %v", err)
		return
	}

	client := &Client{
		hub:         hub,
		conn:        conn,
		send:        make(chan *domain.Message, 256),
		userID:      id,
		activeRooms: make(map[uuid.UUID]bool),
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func SetupWebSocketRoutes(r *gin.Engine, chatServ application.ChatService, hub *Hub) {
	hub.chatService = chatServ

	r.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c, hub)
	})
}

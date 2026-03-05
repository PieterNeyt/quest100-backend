package server

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type WsMessage struct {
	Type        string    `json:"type"`
	SenderID    uuid.UUID `json:"senderId"`
	RecipientID uuid.UUID `json:"recipientId,omitempty"`
	RoomID      uuid.UUID `json:"roomId,omitempty"`
	Content     string    `json:"content,omitempty"`
}

func (h *Hub) handleMessage(rawMessage []byte) {
	var msg WsMessage
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		log.Printf("Error decoding message: %v", err)
		return
	}

	log.Printf("Received message: %s", msg.Type)
	switch msg.Type {
	case "private":
		//h.handlePrivateMessage(msg, rawMessage)
		fmt.Println("private message")
	case "group":
		h.handleGroupMessage(msg, rawMessage)
	case "join":
		h.handleJoinRoom(msg)
		//case "leave":
		//	h.handleLeaveRoom(msg)
	}
}

func (h *Hub) handleJoinRoom(msg WsMessage) {
	err := h.chatService.IsUserInChat(msg.SenderID, msg.RoomID)
	if err != nil {
		log.Printf("Access denied: User %s -> Room %s: %s", msg.SenderID, msg.RoomID, err)
		return
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	client := h.users[msg.SenderID]
	if h.rooms[msg.RoomID] == nil {
		h.rooms[msg.RoomID] = make(map[*Client]bool)
	}
	h.rooms[msg.RoomID][client] = true
	client.activeRooms[msg.RoomID] = true
	log.Printf("Joined room: User %s -> Room %s", msg.SenderID, msg.RoomID)
}

func (h *Hub) handleGroupMessage(msg WsMessage, message []byte) {
	if err := h.chatService.CreateMessage(msg.RoomID, msg.SenderID, msg.Content); err != nil {
		return
	}
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if clients, ok := h.rooms[msg.RoomID]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				h.unregister <- client
			}
		}
	}
}

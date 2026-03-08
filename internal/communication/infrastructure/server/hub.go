package server

import (
	"Quest100Backend/internal/communication/application"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	users        map[uuid.UUID]*Client
	rooms        map[uuid.UUID]map[*Client]bool
	register     chan *Client
	unregister   chan *Client
	routeMessage chan []byte
	mutex        sync.RWMutex
	chatService  application.ChatService
}

func NewHub() *Hub {
	return &Hub{
		users:        make(map[uuid.UUID]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		rooms:        make(map[uuid.UUID]map[*Client]bool),
		routeMessage: make(chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case message := <-h.routeMessage:
			h.handleMessage(message)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.users[client.userID] = client
	log.Printf("User %s connected", client.userID)
}

func (h *Hub) handleUnregister(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.users[client.userID]; ok {
		delete(h.users, client.userID)
		for roomID := range client.activeRooms {
			if clients, exists := h.rooms[roomID]; exists {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.rooms, roomID)
				}
			}
		}
		close(client.send)
		log.Printf("User %s disconnected", client.userID)
	}
}

type HubSnapshot struct {
	TotalConnected int                 `json:"total_connected"`
	Rooms          map[string][]string `json:"rooms"`
}

func (h *Hub) GetSnapshot() HubSnapshot {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	snapshot := HubSnapshot{
		TotalConnected: len(h.users),
		Rooms:          make(map[string][]string),
	}

	for roomID, clients := range h.rooms {
		var clientList []string
		for client := range clients {
			clientList = append(clientList, client.userID.String())
		}
		snapshot.Rooms[roomID.String()] = clientList
	}

	return snapshot
}

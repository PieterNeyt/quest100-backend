package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "http://localhost:4200"
	},
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

type Hub struct {
	users        map[string]*Client
	rooms        map[string]map[*Client]bool
	register     chan *Client
	unregister   chan *Client
	routeMessage chan []byte
	mutex        sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		users:        make(map[string]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		rooms:        make(map[string]map[*Client]bool),
		routeMessage: make(chan []byte),
	}
}

type ChatMessage struct {
	Type        string `json:"type"`
	SenderID    string `json:"senderId"`
	RecipientID string `json:"recipientId,omitempty"`
	RoomID      string `json:"roomId,omitempty"`
	Text        string `json:"text"`
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.users[client.userID] = client
			h.mutex.Unlock()
			log.Printf("User %s connected", client.userID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.users[client.userID]; ok {
				delete(h.users, client.userID)
				for roomID, clients := range h.rooms {
					if _, ok := clients[client]; ok {
						delete(h.rooms[roomID], client)
					}
				}
				close(client.send)
				log.Printf("User %s disconnected", client.userID)
			}
			h.mutex.Unlock()

		case message := <-h.routeMessage:
			var msg ChatMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error decoding message: %v", err)
				continue
			}

			h.mutex.RLock()

			if msg.Type == "private" {
				if targetClient, ok := h.users[msg.RecipientID]; ok {
					select {
					case targetClient.send <- message:
					default:
						h.mutex.RUnlock()
						h.unregister <- targetClient
						h.mutex.RLock()
					}
				}
				if senderClient, ok := h.users[msg.SenderID]; ok {
					senderClient.send <- message
				}

			} else if msg.Type == "group" {
				if clients, ok := h.rooms[msg.RoomID]; ok {
					for client := range clients {
						select {
						case client.send <- message:
						default:
							h.mutex.RUnlock()
							h.unregister <- client
							h.mutex.RLock()
						}
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (s *Server) handleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		log.Printf("Profile ID not found")
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade: %v", err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), userID: token}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		log.Printf("recv: %s", message)
		c.hub.routeMessage <- message
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			log.Printf("wrote: %s", message)
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

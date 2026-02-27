package server

import (
	profileRouting "Quest100Backend/internal/profile/infrastructure/server"
	qrcodeRouting "Quest100Backend/internal/util/qrcode/infrastructure/server"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Graph-Token"},
		AllowCredentials: true,
	}))

	api := r.Group("/api", AuthMiddleware())
	profileRouting.SetupProfileRoutes(api, s.db.GetDB())
	qrcodeRouting.SetupQRCodeRoutes(api)

	r.GET("/ws", s.handleWebSocket)

	return r
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// In production, implement a proper origin check
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler integrated into Gin
//func (s *Server) handleWebSocket(c *gin.Context) {
//	// Upgrade HTTP connection to WebSocket
//	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		log.Printf("Failed to upgrade connection: %v", err)
//		return
//	}
//
//	// Handle the connection in a new goroutine
//	go handleConnection(conn)
//}

func handleConnection(conn *websocket.Conn) {
	defer conn.Close()
	log.Println("Client connected via Gorilla/Gin")

	for {
		// Read message from client
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		log.Printf("Received: %s\n", string(p))

		// Echo message back to client
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

// Client represents a single chat user
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte // Buffered channel of outbound messages
	userID string      // Essential for 1-on-1 routing
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Map of userId to Client for private messages
	users map[string]*Client
	// Map of roomId to a set of Clients
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	// Channel for messages intended for the hub to route
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
	Type        string `json:"type"` // "private" or "group"
	SenderID    string `json:"senderId"`
	RecipientID string `json:"recipientId,omitempty"` // For private
	RoomID      string `json:"roomId,omitempty"`      // For group
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
				// Remove client from all rooms they were in
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
				// Send to specific user
				if targetClient, ok := h.users[msg.RecipientID]; ok {
					select {
					case targetClient.send <- message:
					default:
						// If channel is full, consider client disconnected
						h.mutex.RUnlock()
						h.unregister <- targetClient
						h.mutex.RLock()
					}
				}
				// Also send back to sender so they see their own message
				if senderClient, ok := h.users[msg.SenderID]; ok {
					senderClient.send <- message
				}

			} else if msg.Type == "group" {
				// Send to all clients in the room
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
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade: %v", err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Start read/write routines for this specific client
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// Configure connection limits if necessary
	c.conn.SetReadLimit(512)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		// Send raw message to hub for parsing and routing
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
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Simple check to see if there are more messages to send
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

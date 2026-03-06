package server

import (
	"Quest100Backend/internal/communication/domain"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan *domain.Message
	userID      uuid.UUID
	activeRooms map[uuid.UUID]bool
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
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(message); err != nil {
				return
			}

			log.Printf("wrote: %s", message)
			n := len(c.send)
			for i := 0; i < n; i++ {
				extraMsg := <-c.send
				if err := encoder.Encode(extraMsg); err != nil {
					break
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

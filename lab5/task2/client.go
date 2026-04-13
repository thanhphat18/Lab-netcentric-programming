package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	username string
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, rawMessage, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				log.Printf("unexpected close for %s: %v", c.username, err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			sendErrorToClient(c, "invalid JSON format")
			continue
		}

		if msg.Type == "" || msg.Content == "" {
			sendErrorToClient(c, "missing required fields: type and content")
			continue
		}

		switch msg.Type {
		case "chat":
			msg.Sender = c.username
			msg.Timestamp = time.Now().Format(time.RFC3339)

			data, err := json.Marshal(msg)
			if err != nil {
				sendErrorToClient(c, "internal server error")
				continue
			}

			c.hub.broadcast <- data
			log.Printf("[%s] %s: %s", msg.Type, c.username, msg.Content)

		default:
			sendErrorToClient(c, "unknown message type: "+msg.Type)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func sendErrorToClient(c *Client, errText string) {
	msg := Message{
		Type:    "error",
		Content: errText,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("marshal error failed:", err)
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		delete(c.hub.clients, c)
	}
}

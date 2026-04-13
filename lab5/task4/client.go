package main

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	username string
	room     string
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

		switch msg.Type {
		case "chat":
			if strings.TrimSpace(msg.Content) == "" {
				sendErrorToClient(c, "content is required")
				continue
			}

			out := Message{
				Type:      "chat",
				Sender:    c.username,
				Content:   msg.Content,
				Timestamp: nowRFC3339(),
				Room:      c.room,
			}

			data, err := json.Marshal(out)
			if err != nil {
				sendErrorToClient(c, "internal server error")
				continue
			}

			c.hub.broadcast <- &RoomMessage{
				room:    c.room,
				message: data,
			}

		case "join_room":
			newRoom := strings.TrimSpace(msg.Content)
			if newRoom == "" {
				sendErrorToClient(c, "room name is required")
				continue
			}
			if newRoom == c.room {
				sendInfoToClient(c, "already in room "+newRoom, c.room)
				continue
			}

			oldRoom := c.room
			c.hub.moveClientToRoom(c, newRoom)
			sendInfoToClient(c, "moved from "+oldRoom+" to "+newRoom, newRoom)

		case "list_rooms":
			roomList := c.hub.listRooms()
			out := Message{
				Type:  "room_list",
				Room:  c.room,
				Rooms: roomList,
			}
			data, err := json.Marshal(out)
			if err != nil {
				sendErrorToClient(c, "internal server error")
				continue
			}
			select {
			case c.send <- data:
			default:
				c.hub.unregister <- c
			}

		case "dm":
			if strings.TrimSpace(msg.Content) == "" {
				sendErrorToClient(c, "dm content is required")
				continue
			}
			if strings.TrimSpace(msg.Recipient) == "" {
				sendErrorToClient(c, "recipient is required")
				continue
			}
			if msg.Recipient == c.username {
				sendErrorToClient(c, "cannot send DM to yourself")
				continue
			}

			c.hub.sendDirect(c, msg.Recipient, msg.Content)

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
		Room:    c.room,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("marshal error failed:", err)
		return
	}

	select {
	case c.send <- data:
	default:
		c.hub.unregister <- c
	}
}

func sendInfoToClient(c *Client, text string, room string) {
	msg := Message{
		Type:    "info",
		Content: text,
		Room:    room,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("marshal info failed:", err)
		return
	}

	select {
	case c.send <- data:
	default:
		c.hub.unregister <- c
	}
}

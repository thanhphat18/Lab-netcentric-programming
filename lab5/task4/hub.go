package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
)

type RoomMessage struct {
	room    string
	message []byte
}

type Hub struct {
	rooms      map[string]map[*Client]bool
	usernames  map[string]*Client
	broadcast  chan *RoomMessage
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		usernames:  make(map[string]*Client),
		broadcast:  make(chan *RoomMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			// Reject duplicate username
			if _, exists := h.usernames[client.username]; exists {
				sendErrorDirect(client, "username already in use: "+client.username)
				close(client.send)
				client.conn.Close()
				continue
			}

			h.usernames[client.username] = client

			if _, ok := h.rooms[client.room]; !ok {
				h.rooms[client.room] = make(map[*Client]bool)
			}
			h.rooms[client.room][client] = true

			log.Printf("client registered: %s joined room %s", client.username, client.room)

			joinMsg := Message{
				Type:    "join",
				Sender:  client.username,
				Content: fmt.Sprintf("joined room %s", client.room),
				Room:    client.room,
			}
			h.broadcastToRoom(client.room, joinMsg)

		case client := <-h.unregister:
			if existing, ok := h.usernames[client.username]; ok && existing == client {
				delete(h.usernames, client.username)
			}

			if roomClients, ok := h.rooms[client.room]; ok {
				if _, exists := roomClients[client]; exists {
					delete(roomClients, client)

					leaveMsg := Message{
						Type:    "leave",
						Sender:  client.username,
						Content: fmt.Sprintf("left room %s", client.room),
						Room:    client.room,
					}
					h.broadcastToRoom(client.room, leaveMsg)

					if len(roomClients) == 0 {
						delete(h.rooms, client.room)
					}
				}
			}

			select {
			case <-client.send:
			default:
			}
			closeIfOpen(client.send)

			log.Printf("client unregistered: %s", client.username)

		case roomMsg := <-h.broadcast:
			roomClients, ok := h.rooms[roomMsg.room]
			if !ok {
				continue
			}

			for client := range roomClients {
				select {
				case client.send <- roomMsg.message:
				default:
					if existing, ok := h.usernames[client.username]; ok && existing == client {
						delete(h.usernames, client.username)
					}
					close(client.send)
					delete(roomClients, client)
				}
			}
		}
	}
}

func (h *Hub) moveClientToRoom(client *Client, newRoom string) {
	if newRoom == "" || newRoom == client.room {
		return
	}

	oldRoom := client.room

	if roomClients, ok := h.rooms[oldRoom]; ok {
		delete(roomClients, client)

		leaveMsg := Message{
			Type:    "leave",
			Sender:  client.username,
			Content: fmt.Sprintf("left room %s", oldRoom),
			Room:    oldRoom,
		}
		h.broadcastToRoom(oldRoom, leaveMsg)

		if len(roomClients) == 0 {
			delete(h.rooms, oldRoom)
		}
	}

	client.room = newRoom

	if _, ok := h.rooms[newRoom]; !ok {
		h.rooms[newRoom] = make(map[*Client]bool)
	}
	h.rooms[newRoom][client] = true

	joinMsg := Message{
		Type:    "join",
		Sender:  client.username,
		Content: fmt.Sprintf("joined room %s", newRoom),
		Room:    newRoom,
	}
	h.broadcastToRoom(newRoom, joinMsg)
}

func (h *Hub) listRooms() []RoomInfo {
	rooms := make([]RoomInfo, 0, len(h.rooms))
	for roomName, clients := range h.rooms {
		rooms = append(rooms, RoomInfo{
			Name:  roomName,
			Count: len(clients),
		})
	}
	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].Name < rooms[j].Name
	})
	return rooms
}

func (h *Hub) sendDirect(sender *Client, recipientUsername string, content string) {
	recipient, ok := h.usernames[recipientUsername]
	if !ok {
		sendErrorToClient(sender, "recipient not online: "+recipientUsername)
		return
	}

	// Send message to recipient only
	dmToRecipient := Message{
		Type:      "dm",
		Sender:    sender.username,
		Recipient: recipient.username,
		Content:   content,
		Timestamp: nowRFC3339(),
		Room:      sender.room,
	}

	dataRecipient, err := json.Marshal(dmToRecipient)
	if err != nil {
		sendErrorToClient(sender, "internal server error")
		return
	}

	select {
	case recipient.send <- dataRecipient:
	default:
		sendErrorToClient(sender, "recipient is unavailable: "+recipientUsername)
		return
	}

	// Confirmation to sender
	dmConfirm := Message{
		Type:      "dm_sent",
		Sender:    sender.username,
		Recipient: recipient.username,
		Content:   content,
		Timestamp: nowRFC3339(),
		Room:      sender.room,
	}

	dataSender, err := json.Marshal(dmConfirm)
	if err != nil {
		sendErrorToClient(sender, "internal server error")
		return
	}

	select {
	case sender.send <- dataSender:
	default:
	}
}

func (h *Hub) broadcastToRoom(room string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("marshal room message failed:", err)
		return
	}

	roomClients, ok := h.rooms[room]
	if !ok {
		return
	}

	for client := range roomClients {
		select {
		case client.send <- data:
		default:
			if existing, ok := h.usernames[client.username]; ok && existing == client {
				delete(h.usernames, client.username)
			}
			close(client.send)
			delete(roomClients, client)
		}
	}
}

func sendErrorDirect(client *Client, text string) {
	msg := Message{
		Type:    "error",
		Content: text,
		Room:    client.room,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case client.send <- data:
	default:
	}
}

func closeIfOpen(ch chan []byte) {
	defer func() {
		recover()
	}()
	close(ch)
}

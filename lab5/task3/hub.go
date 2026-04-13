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
	broadcast  chan *RoomMessage
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan *RoomMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			if _, ok := h.rooms[client.room]; !ok {
				h.rooms[client.room] = make(map[*Client]bool)
			}
			h.rooms[client.room][client] = true

			log.Printf("client registered: %s joined room %s (%d users)",
				client.username, client.room, len(h.rooms[client.room]))

			joinMsg := Message{
				Type:    "join",
				Sender:  client.username,
				Content: fmt.Sprintf("joined room %s", client.room),
				Room:    client.room,
			}
			h.broadcastToRoom(client.room, joinMsg)

		case client := <-h.unregister:
			if roomClients, ok := h.rooms[client.room]; ok {
				if _, exists := roomClients[client]; exists {
					delete(roomClients, client)
					close(client.send)

					leaveMsg := Message{
						Type:    "leave",
						Sender:  client.username,
						Content: fmt.Sprintf("left room %s", client.room),
						Room:    client.room,
					}
					h.broadcastToRoom(client.room, leaveMsg)

					if len(roomClients) == 0 {
						delete(h.rooms, client.room)
						log.Printf("deleted empty room: %s", client.room)
					} else {
						log.Printf("client unregistered: %s left room %s (%d users remain)",
							client.username, client.room, len(roomClients))
					}
				}
			}

		case roomMsg := <-h.broadcast:
			roomClients, ok := h.rooms[roomMsg.room]
			if !ok {
				continue
			}

			for client := range roomClients {
				select {
				case client.send <- roomMsg.message:
				default:
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

	// Remove from old room
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
			log.Printf("deleted empty room: %s", oldRoom)
		}
	}

	// Add to new room
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

	log.Printf("%s moved from %s to %s", client.username, oldRoom, newRoom)
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
			close(client.send)
			delete(roomClients, client)
		}
	}
}

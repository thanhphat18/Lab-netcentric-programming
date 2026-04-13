package main

import (
	"encoding/json"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("client registered: %s (%d total)", client.username, len(h.clients))

			joinMsg := Message{
				Type:    "join",
				Sender:  client.username,
				Content: "joined the chat",
			}
			if data, err := json.Marshal(joinMsg); err == nil {
				for c := range h.clients {
					select {
					case c.send <- data:
					default:
						close(c.send)
						delete(h.clients, c)
					}
				}
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("client unregistered: %s (%d remaining)", client.username, len(h.clients))

				leaveMsg := Message{
					Type:    "leave",
					Sender:  client.username,
					Content: "left the chat",
				}
				if data, err := json.Marshal(leaveMsg); err == nil {
					for c := range h.clients {
						select {
						case c.send <- data:
						default:
							close(c.send)
							delete(h.clients, c)
						}
					}
				}
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

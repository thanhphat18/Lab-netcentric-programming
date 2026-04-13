package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type RoomInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Message struct {
	Type      string     `json:"type"`
	Sender    string     `json:"sender,omitempty"`
	Content   string     `json:"content,omitempty"`
	Room      string     `json:"room,omitempty"`
	Rooms     []RoomInfo `json:"rooms,omitempty"`
	Timestamp string     `json:"timestamp,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}

	room := r.URL.Query().Get("room")
	if room == "" {
		room = "general"
	}

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		username: username,
		room:     room,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func main() {
	hub := newHub()
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	http.Handle("/", http.FileServer(http.Dir("static")))

	log.Println("Room chat server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

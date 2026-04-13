---
title: Web Socket
tags: [template]
---

<!-- Change the following info. Leave group empty if you dont have any group -->
# Net-Centric Programming Lab 05 Report
Full Name: Chau Thanh Phat
Student ID: ITITIU21135
___
**Content**
Learn to build real-time WebSocket services in Go, covering connection management, JSON message protocols, broadcasting with the Hub pattern, rooms, private messaging, and production-quality connection lifecycle handling.
___

## Task 1:
In Task 1, I implemented a basic WebSocket JSON echo server in Go using the gorilla/websocket package. The server upgrades HTTP connections to WebSocket, continuously reads incoming messages in a read loop, parses them from JSON into a Message struct, validates the required fields, and routes processing based on the type field. Valid chat messages are echoed back with a server-generated RFC3339 timestamp, while invalid JSON, missing fields, and unknown message types return structured JSON error responses. A browser client was also created to connect to ws://localhost:8080/ws, send chat messages, and display normal responses and errors.

'''
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Upgrader converts HTTP connection to WebSocket connection.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development only.
		return true
	},
}

// Message defines the JSON structure for WebSocket communication.
type Message struct {
	Type      string `json:"type"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
}

// sendError sends a JSON error response to the client.
func sendError(conn *websocket.Conn, errText string) {
	errMsg := Message{
		Type:    "error",
		Content: errText,
	}

	data, err := json.Marshal(errMsg)
	if err != nil {
		log.Println("marshal error response failed:", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("write error response failed:", err)
	}
}

// handleWebSocket upgrades the connection and processes messages.
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Step 1: Upgrade HTTP -> WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Printf("client connected: %s", conn.RemoteAddr())

	// Step 2: Read loop
	for {
		_, rawMessage, err := conn.ReadMessage()
		if err != nil {
			log.Printf("client disconnected: %s (%v)", conn.RemoteAddr(), err)
			break
		}

		// Step 3: Parse JSON
		var msg Message
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Printf("invalid JSON from %s: %v", conn.RemoteAddr(), err)
			sendError(conn, "invalid JSON format")
			continue
		}

		// Step 4: Validate required fields
		if msg.Type == "" || msg.Content == "" {
			sendError(conn, "missing required fields: type and content")
			continue
		}

		// Step 5: Log received message
		log.Printf("[%s] %s: %s", msg.Type, msg.Sender, msg.Content)

		// Step 6: Route by message type
		switch msg.Type {
		case "chat":
			msg.Timestamp = time.Now().Format(time.RFC3339)

			response, err := json.Marshal(msg)
			if err != nil {
				log.Println("marshal chat response failed:", err)
				sendError(conn, "internal server error")
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
				log.Println("write chat response failed:", err)
				break
			}

		default:
			sendError(conn, "unknown message type: "+msg.Type)
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir("static")))

	log.Println("JSON Echo server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

'''
### HTML file
'''
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>WebSocket JSON Echo</title>
  <style>
    body {
      font-family: monospace;
      max-width: 700px;
      margin: 40px auto;
    }

    #log {
      border: 1px solid #ccc;
      padding: 10px;
      height: 320px;
      overflow-y: auto;
      background: #f9f9f9;
      margin-bottom: 10px;
    }

    .sent { color: blue; }
    .received { color: green; }
    .error { color: red; }
    .system { color: gray; font-style: italic; }

    input {
      padding: 6px;
      margin-right: 6px;
    }

    button {
      padding: 6px 14px;
    }
  </style>
</head>
<body>
  <h2>WebSocket JSON Echo</h2>

  <div id="log"></div>

  <input type="text" id="name" placeholder="Your name" value="Alice" style="width:120px;" />
  <input type="text" id="msg" placeholder="Type a message..." style="width:320px;" />
  <button onclick="send()">Send</button>

  <script>
    const logDiv = document.getElementById("log");
    const nameInput = document.getElementById("name");
    const msgInput = document.getElementById("msg");

    function appendLog(text, cls) {
      const div = document.createElement("div");
      div.textContent = text;
      div.className = cls;
      logDiv.appendChild(div);
      logDiv.scrollTop = logDiv.scrollHeight;
    }

    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = () => {
      appendLog("Connected", "system");
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);

      if (msg.type === "error") {
        appendLog("Error: " + msg.content, "error");
      } else {
        appendLog(
          `← [${msg.type}] ${msg.sender}: ${msg.content} (${msg.timestamp})`,
          "received"
        );
      }
    };

    ws.onclose = () => {
      appendLog("Disconnected", "system");
    };

    ws.onerror = () => {
      appendLog("WebSocket error", "error");
    };

    function send() {
      const text = msgInput.value.trim();
      if (!text || ws.readyState !== WebSocket.OPEN) return;

      const msg = {
        type: "chat",
        sender: nameInput.value.trim() || "Anonymous",
        content: text
      };

      ws.send(JSON.stringify(msg));
      appendLog(`→ [${msg.type}] ${msg.sender}: ${msg.content}`, "sent");
      msgInput.value = "";
    }

    msgInput.addEventListener("keypress", (e) => {
      if (e.key === "Enter") send();
    });
  </script>
</body>
</html>
'''

## Task 2: 
In Task 2, I implemented the Hub pattern to support multi-client broadcasting over WebSocket. A central Hub goroutine manages connected clients through register, unregister, and broadcast channels. Each WebSocket connection is represented by a Client with two goroutines: readPump() reads and validates JSON chat messages, then forwards them to the Hub, while writePump() sends outbound messages and periodic ping frames to keep the connection alive. Join and leave notifications are broadcast to all connected users, and a Go terminal client was added to test communication alongside the browser client.

main.go
'''
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type      string `json:"type"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
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

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		username: username,
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

	log.Println("Chat server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

'''

hub.go
'''
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
'''

client.go
'''
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

'''

static > index.html
'''
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Broadcast Chat</title>
  <style>
    body { font-family: monospace; max-width: 700px; margin: 40px auto; }
    #log {
      border: 1px solid #ccc;
      padding: 10px;
      height: 320px;
      overflow-y: auto;
      background: #f9f9f9;
      margin-bottom: 10px;
    }
    .sent { color: blue; }
    .chat { color: green; }
    .join { color: purple; font-style: italic; }
    .leave { color: brown; font-style: italic; }
    .error { color: red; }
    .system { color: gray; font-style: italic; }
    input { padding: 6px; margin-right: 6px; }
    button { padding: 6px 14px; }
  </style>
</head>
<body>
  <h2>Broadcast Chat</h2>

  <div id="log"></div>

  <input type="text" id="name" placeholder="Your name" value="Alice" style="width:120px;" />
  <button onclick="connect()">Connect</button>
  <br><br>

  <input type="text" id="msg" placeholder="Type a message..." style="width:320px;" />
  <button onclick="send()">Send</button>

  <script>
    const logDiv = document.getElementById("log");
    const nameInput = document.getElementById("name");
    const msgInput = document.getElementById("msg");
    let ws = null;

    function appendLog(text, cls) {
      const div = document.createElement("div");
      div.textContent = text;
      div.className = cls;
      logDiv.appendChild(div);
      logDiv.scrollTop = logDiv.scrollHeight;
    }

    function connect() {
      if (ws && ws.readyState === WebSocket.OPEN) {
        appendLog("Already connected", "system");
        return;
      }

      const username = encodeURIComponent(nameInput.value.trim() || "Anonymous");
      ws = new WebSocket(`ws://localhost:8080/ws?username=${username}`);

      ws.onopen = () => appendLog("Connected", "system");

      ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);

        if (msg.type === "error") {
          appendLog("Error: " + msg.content, "error");
        } else if (msg.type === "join") {
          appendLog(`+ ${msg.sender} ${msg.content}`, "join");
        } else if (msg.type === "leave") {
          appendLog(`- ${msg.sender} ${msg.content}`, "leave");
        } else if (msg.type === "chat") {
          appendLog(`← ${msg.sender}: ${msg.content} (${msg.timestamp})`, "chat");
        } else {
          appendLog(JSON.stringify(msg), "system");
        }
      };

      ws.onclose = () => appendLog("Disconnected", "system");
      ws.onerror = () => appendLog("WebSocket error", "error");
    }

    function send() {
      const text = msgInput.value.trim();
      if (!ws || ws.readyState !== WebSocket.OPEN || !text) return;

      const msg = {
        type: "chat",
        sender: nameInput.value.trim() || "Anonymous",
        content: text
      };

      ws.send(JSON.stringify(msg));
      appendLog(`→ ${msg.sender}: ${msg.content}`, "sent");
      msgInput.value = "";
    }

    msgInput.addEventListener("keypress", (e) => {
      if (e.key === "Enter") send();
    });
  </script>
</body>
</html>
'''

go client > main.go
'''
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type      string `json:"type"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run . <username>")
	}

	username := os.Args[1]
	url := fmt.Sprintf("ws://localhost:8080/ws?username=%s", username)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	log.Println("connected to", url)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("disconnected:", err)
				return
			}

			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				fmt.Printf("← raw: %s\n", message)
				continue
			}

			switch msg.Type {
			case "join":
				fmt.Printf("[JOIN] %s %s\n", msg.Sender, msg.Content)
			case "leave":
				fmt.Printf("[LEAVE] %s %s\n", msg.Sender, msg.Content)
			case "chat":
				fmt.Printf("[CHAT] %s: %s (%s)\n", msg.Sender, msg.Content, msg.Timestamp)
			case "error":
				fmt.Printf("[ERROR] %s\n", msg.Content)
			default:
				fmt.Printf("[RAW] %s\n", message)
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type messages and press Enter. Ctrl+C to quit.")

	for {
		select {
		case <-done:
			return

		case <-interrupt:
			log.Println("shutting down...")
			err := conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"),
			)
			if err != nil {
				log.Println("close error:", err)
			}
			return

		default:
			if scanner.Scan() {
				text := scanner.Text()
				if text == "" {
					continue
				}

				msg := Message{
					Type:    "chat",
					Sender:  username,
					Content: text,
				}

				data, err := json.Marshal(msg)
				if err != nil {
					log.Println("marshal error:", err)
					continue
				}

				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Println("write error:", err)
					return
				}
			}
		}
	}
}
'''

### Task 3:
In Task 3, I extended the Hub pattern to support room-based targeted broadcasting. The Hub now stores clients using rooms map[string]map[*Client]bool, where each room name maps to a set of connected clients. Each client has a room field, defaults to general, and may also join a room from the WebSocket query parameter. Chat messages are wrapped in a RoomMessage so the Hub can broadcast only to the intended room. I also implemented join_room to move a client between rooms with correct leave and join notifications, and list_rooms to return active room names and user counts. Empty rooms are removed from the Hub automatically when the last user leaves. This ensures messages stay isolated to the correct room while preserving the same readPump/writePump and ping/pong architecture from Task 2.

main.go
'''
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

'''

hub.go
'''
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

'''

client.go
'''
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
	maxMessageSize = 512
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
				Timestamp: time.Now().Format(time.RFC3339),
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

			log.Printf("[chat][room=%s] %s: %s", c.room, c.username, msg.Content)

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
'''

static > index.html
'''
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Room Chat</title>
  <style>
    body { font-family: monospace; max-width: 800px; margin: 40px auto; }
    #log {
      border: 1px solid #ccc;
      padding: 10px;
      height: 360px;
      overflow-y: auto;
      background: #f9f9f9;
      margin-bottom: 10px;
    }
    .sent { color: blue; }
    .chat { color: green; }
    .join { color: purple; font-style: italic; }
    .leave { color: brown; font-style: italic; }
    .info { color: darkcyan; }
    .error { color: red; }
    .system { color: gray; font-style: italic; }
    input { padding: 6px; margin-right: 6px; }
    button { padding: 6px 12px; margin-right: 4px; }
    #currentRoom { font-weight: bold; }
  </style>
</head>
<body>
  <h2>Room Chat</h2>

  <div>
    <input id="name" value="Alice" placeholder="Username" style="width:120px;" />
    <input id="room" value="general" placeholder="Room" style="width:120px;" />
    <button onclick="connect()">Connect</button>
    <button onclick="listRooms()">List Rooms</button>
  </div>

  <p>Current room: <span id="currentRoom">-</span></p>

  <div id="log"></div>

  <div>
    <input id="msg" placeholder="Type a message..." style="width:320px;" />
    <button onclick="sendChat()">Send</button>
  </div>

  <div style="margin-top:10px;">
    <input id="newRoom" placeholder="New room name" style="width:180px;" />
    <button onclick="joinRoom()">Join Room</button>
  </div>

  <script>
    let ws = null;
    const logDiv = document.getElementById("log");
    const currentRoomSpan = document.getElementById("currentRoom");

    function appendLog(text, cls) {
      const div = document.createElement("div");
      div.textContent = text;
      div.className = cls;
      logDiv.appendChild(div);
      logDiv.scrollTop = logDiv.scrollHeight;
    }

    function connect() {
      if (ws && ws.readyState === WebSocket.OPEN) {
        appendLog("Already connected", "system");
        return;
      }

      const username = encodeURIComponent(document.getElementById("name").value.trim() || "Anonymous");
      const room = encodeURIComponent(document.getElementById("room").value.trim() || "general");

      ws = new WebSocket(`ws://localhost:8080/ws?username=${username}&room=${room}`);

      ws.onopen = () => {
        currentRoomSpan.textContent = decodeURIComponent(room);
        appendLog("Connected", "system");
      };

      ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);

        if (msg.room) {
          currentRoomSpan.textContent = msg.room;
        }

        switch (msg.type) {
          case "chat":
            appendLog(`[${msg.room}] ${msg.sender}: ${msg.content} (${msg.timestamp})`, "chat");
            break;
          case "join":
            appendLog(`[${msg.room}] + ${msg.sender} ${msg.content}`, "join");
            break;
          case "leave":
            appendLog(`[${msg.room}] - ${msg.sender} ${msg.content}`, "leave");
            break;
          case "room_list":
            appendLog("Active rooms: " + msg.rooms.map(r => `${r.name}(${r.count})`).join(", "), "info");
            break;
          case "info":
            appendLog(msg.content, "info");
            break;
          case "error":
            appendLog("Error: " + msg.content, "error");
            break;
          default:
            appendLog(JSON.stringify(msg), "system");
        }
      };

      ws.onclose = () => appendLog("Disconnected", "system");
      ws.onerror = () => appendLog("WebSocket error", "error");
    }

    function sendChat() {
      const input = document.getElementById("msg");
      const text = input.value.trim();
      if (!ws || ws.readyState !== WebSocket.OPEN || !text) return;

      const msg = {
        type: "chat",
        content: text
      };

      ws.send(JSON.stringify(msg));
      appendLog(`→ ${text}`, "sent");
      input.value = "";
    }

    function joinRoom() {
      const room = document.getElementById("newRoom").value.trim();
      if (!ws || ws.readyState !== WebSocket.OPEN || !room) return;

      ws.send(JSON.stringify({
        type: "join_room",
        content: room
      }));
    }

    function listRooms() {
      if (!ws || ws.readyState !== WebSocket.OPEN) return;

      ws.send(JSON.stringify({
        type: "list_rooms"
      }));
    }

    document.getElementById("msg").addEventListener("keypress", (e) => {
      if (e.key === "Enter") sendChat();
    });
  </script>
</body>
</html>
'''

go-client > main.go
'''
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

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

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run . <username> [room]")
	}

	username := os.Args[1]
	room := "general"
	if len(os.Args) >= 3 {
		room = os.Args[2]
	}

	url := fmt.Sprintf("ws://localhost:8080/ws?username=%s&room=%s", username, room)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	log.Println("connected to", url)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("disconnected:", err)
				return
			}

			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				fmt.Printf("← raw: %s\n", message)
				continue
			}

			switch msg.Type {
			case "chat":
				fmt.Printf("[CHAT][%s] %s: %s (%s)\n", msg.Room, msg.Sender, msg.Content, msg.Timestamp)
			case "join":
				fmt.Printf("[JOIN][%s] %s %s\n", msg.Room, msg.Sender, msg.Content)
			case "leave":
				fmt.Printf("[LEAVE][%s] %s %s\n", msg.Room, msg.Sender, msg.Content)
			case "room_list":
				fmt.Printf("[ROOMS] ")
				for i, r := range msg.Rooms {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Printf("%s(%d)", r.Name, r.Count)
				}
				fmt.Println()
			case "info":
				fmt.Printf("[INFO] %s\n", msg.Content)
			case "error":
				fmt.Printf("[ERROR] %s\n", msg.Content)
			default:
				fmt.Printf("[RAW] %s\n", string(message))
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Commands:")
	fmt.Println("  normal text         -> send chat")
	fmt.Println("  /join roomname      -> switch room")
	fmt.Println("  /rooms              -> list rooms")
	fmt.Println("  Ctrl+C              -> quit")

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("shutting down...")
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"),
			)
			return
		default:
			if scanner.Scan() {
				text := scanner.Text()
				if text == "" {
					continue
				}

				var msg Message

				if len(text) > 6 && text[:6] == "/join " {
					msg = Message{
						Type:    "join_room",
						Content: text[6:],
					}
				} else if text == "/rooms" {
					msg = Message{
						Type: "list_rooms",
					}
				} else {
					msg = Message{
						Type:    "chat",
						Content: text,
					}
				}

				data, err := json.Marshal(msg)
				if err != nil {
					log.Println("marshal error:", err)
					continue
				}

				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Println("write error:", err)
					return
				}
			}
		}
	}
}

'''

### Task 4
In Task 4, I extended the room-based WebSocket chat service to support direct private messaging between users. The Hub now maintains a usernames map[string]*Client, allowing the server to route messages to a specific connected user by username. During registration, usernames are checked for duplicates and rejected if already in use. I added a new dm message type with a recipient field. When a client sends a DM, the server looks up the target client in the username map and sends the message only to that user, while also returning a confirmation to the sender. If the recipient is not online, the sender receives a structured JSON error response. This preserves the existing room broadcast system while adding one-to-one messaging functionality.

main.go
'''
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type RoomInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Message struct {
	Type      string     `json:"type"`
	Sender    string     `json:"sender,omitempty"`
	Recipient string     `json:"recipient,omitempty"`
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

func nowRFC3339() string {
	return time.Now().Format(time.RFC3339)
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

	log.Println("DM chat server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

'''
hub.go
'''
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

'''
client.go
'''
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

'''
static > index.html
'''
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Room Chat + DM</title>
  <style>
    body { font-family: monospace; max-width: 850px; margin: 40px auto; }
    #log {
      border: 1px solid #ccc;
      padding: 10px;
      height: 380px;
      overflow-y: auto;
      background: #f9f9f9;
      margin-bottom: 10px;
    }
    .sent { color: blue; }
    .chat { color: green; }
    .join { color: purple; font-style: italic; }
    .leave { color: brown; font-style: italic; }
    .dm { color: darkorange; font-weight: bold; }
    .dm-sent { color: teal; }
    .info { color: darkcyan; }
    .error { color: red; }
    .system { color: gray; font-style: italic; }
    input { padding: 6px; margin-right: 6px; margin-bottom: 6px; }
    button { padding: 6px 12px; margin-right: 4px; }
    #currentRoom { font-weight: bold; }
  </style>
</head>
<body>
  <h2>Room Chat + Private Messaging</h2>

  <div>
    <input id="name" value="Alice" placeholder="Username" style="width:120px;" />
    <input id="room" value="general" placeholder="Room" style="width:120px;" />
    <button onclick="connect()">Connect</button>
    <button onclick="listRooms()">List Rooms</button>
  </div>

  <p>Current room: <span id="currentRoom">-</span></p>

  <div id="log"></div>

  <div>
    <input id="msg" placeholder="Room message..." style="width:320px;" />
    <button onclick="sendChat()">Send Room Chat</button>
  </div>

  <div style="margin-top:10px;">
    <input id="newRoom" placeholder="New room name" style="width:180px;" />
    <button onclick="joinRoom()">Join Room</button>
  </div>

  <div style="margin-top:10px;">
    <input id="recipient" placeholder="DM recipient" style="width:140px;" />
    <input id="dmMsg" placeholder="Private message..." style="width:320px;" />
    <button onclick="sendDM()">Send DM</button>
  </div>

  <script>
    let ws = null;
    const logDiv = document.getElementById("log");
    const currentRoomSpan = document.getElementById("currentRoom");

    function appendLog(text, cls) {
      const div = document.createElement("div");
      div.textContent = text;
      div.className = cls;
      logDiv.appendChild(div);
      logDiv.scrollTop = logDiv.scrollHeight;
    }

    function connect() {
      if (ws && ws.readyState === WebSocket.OPEN) {
        appendLog("Already connected", "system");
        return;
      }

      const username = encodeURIComponent(document.getElementById("name").value.trim() || "Anonymous");
      const room = encodeURIComponent(document.getElementById("room").value.trim() || "general");

      ws = new WebSocket(`ws://localhost:8080/ws?username=${username}&room=${room}`);

      ws.onopen = () => {
        currentRoomSpan.textContent = decodeURIComponent(room);
        appendLog("Connected", "system");
      };

      ws.onmessage = (e) => {
        const msg = JSON.parse(e.data);

        if (msg.room) {
          currentRoomSpan.textContent = msg.room;
        }

        switch (msg.type) {
          case "chat":
            appendLog(`[${msg.room}] ${msg.sender}: ${msg.content} (${msg.timestamp})`, "chat");
            break;
          case "join":
            appendLog(`[${msg.room}] + ${msg.sender} ${msg.content}`, "join");
            break;
          case "leave":
            appendLog(`[${msg.room}] - ${msg.sender} ${msg.content}`, "leave");
            break;
          case "room_list":
            appendLog("Active rooms: " + msg.rooms.map(r => `${r.name}(${r.count})`).join(", "), "info");
            break;
          case "dm":
            appendLog(`[DM from ${msg.sender}] ${msg.content} (${msg.timestamp})`, "dm");
            break;
          case "dm_sent":
            appendLog(`[DM to ${msg.recipient}] ${msg.content} (${msg.timestamp})`, "dm-sent");
            break;
          case "info":
            appendLog(msg.content, "info");
            break;
          case "error":
            appendLog("Error: " + msg.content, "error");
            break;
          default:
            appendLog(JSON.stringify(msg), "system");
        }
      };

      ws.onclose = () => appendLog("Disconnected", "system");
      ws.onerror = () => appendLog("WebSocket error", "error");
    }

    function sendChat() {
      const input = document.getElementById("msg");
      const text = input.value.trim();
      if (!ws || ws.readyState !== WebSocket.OPEN || !text) return;

      ws.send(JSON.stringify({
        type: "chat",
        content: text
      }));
      appendLog(`→ room: ${text}`, "sent");
      input.value = "";
    }

    function sendDM() {
      const recipient = document.getElementById("recipient").value.trim();
      const text = document.getElementById("dmMsg").value.trim();
      if (!ws || ws.readyState !== WebSocket.OPEN || !recipient || !text) return;

      ws.send(JSON.stringify({
        type: "dm",
        recipient: recipient,
        content: text
      }));

      document.getElementById("dmMsg").value = "";
    }

    function joinRoom() {
      const room = document.getElementById("newRoom").value.trim();
      if (!ws || ws.readyState !== WebSocket.OPEN || !room) return;

      ws.send(JSON.stringify({
        type: "join_room",
        content: room
      }));
    }

    function listRooms() {
      if (!ws || ws.readyState !== WebSocket.OPEN) return;

      ws.send(JSON.stringify({
        type: "list_rooms"
      }));
    }

    document.getElementById("msg").addEventListener("keypress", (e) => {
      if (e.key === "Enter") sendChat();
    });

    document.getElementById("dmMsg").addEventListener("keypress", (e) => {
      if (e.key === "Enter") sendDM();
    });
  </script>
</body>
</html>
'''
go-client > main.go
'''
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
)

type RoomInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Message struct {
	Type      string     `json:"type"`
	Sender    string     `json:"sender,omitempty"`
	Recipient string     `json:"recipient,omitempty"`
	Content   string     `json:"content,omitempty"`
	Room      string     `json:"room,omitempty"`
	Rooms     []RoomInfo `json:"rooms,omitempty"`
	Timestamp string     `json:"timestamp,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run . <username> [room]")
	}

	username := os.Args[1]
	room := "general"
	if len(os.Args) >= 3 {
		room = os.Args[2]
	}

	url := fmt.Sprintf("ws://localhost:8080/ws?username=%s&room=%s", username, room)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	log.Println("connected to", url)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("disconnected:", err)
				return
			}

			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				fmt.Printf("← raw: %s\n", string(message))
				continue
			}

			switch msg.Type {
			case "chat":
				fmt.Printf("[CHAT][%s] %s: %s (%s)\n", msg.Room, msg.Sender, msg.Content, msg.Timestamp)
			case "join":
				fmt.Printf("[JOIN][%s] %s %s\n", msg.Room, msg.Sender, msg.Content)
			case "leave":
				fmt.Printf("[LEAVE][%s] %s %s\n", msg.Room, msg.Sender, msg.Content)
			case "room_list":
				fmt.Printf("[ROOMS] ")
				for i, r := range msg.Rooms {
					if i > 0 {
						fmt.Print(", ")
					}
					fmt.Printf("%s(%d)", r.Name, r.Count)
				}
				fmt.Println()
			case "dm":
				fmt.Printf("[DM FROM %s] %s (%s)\n", msg.Sender, msg.Content, msg.Timestamp)
			case "dm_sent":
				fmt.Printf("[DM TO %s] %s (%s)\n", msg.Recipient, msg.Content, msg.Timestamp)
			case "info":
				fmt.Printf("[INFO] %s\n", msg.Content)
			case "error":
				fmt.Printf("[ERROR] %s\n", msg.Content)
			default:
				fmt.Printf("[RAW] %s\n", string(message))
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Commands:")
	fmt.Println("  normal text            -> send room chat")
	fmt.Println("  /join roomname         -> switch room")
	fmt.Println("  /rooms                 -> list rooms")
	fmt.Println("  /dm username message   -> send direct message")
	fmt.Println("  Ctrl+C                 -> quit")

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("shutting down...")
			conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"),
			)
			return
		default:
			if scanner.Scan() {
				text := strings.TrimSpace(scanner.Text())
				if text == "" {
					continue
				}

				var msg Message

				if strings.HasPrefix(text, "/join ") {
					msg = Message{
						Type:    "join_room",
						Content: strings.TrimSpace(text[6:]),
					}
				} else if text == "/rooms" {
					msg = Message{
						Type: "list_rooms",
					}
				} else if strings.HasPrefix(text, "/dm ") {
					parts := strings.SplitN(text[4:], " ", 2)
					if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
						fmt.Println("usage: /dm username message")
						continue
					}
					msg = Message{
						Type:      "dm",
						Recipient: strings.TrimSpace(parts[0]),
						Content:   strings.TrimSpace(parts[1]),
					}
				} else {
					msg = Message{
						Type:    "chat",
						Content: text,
					}
				}

				data, err := json.Marshal(msg)
				if err != nil {
					log.Println("marshal error:", err)
					continue
				}

				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Println("write error:", err)
					return
				}
			}
		}
	}
}

'''
### Task 5:

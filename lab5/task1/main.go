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

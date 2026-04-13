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

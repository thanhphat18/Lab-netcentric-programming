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

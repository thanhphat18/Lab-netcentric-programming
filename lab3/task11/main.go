package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type ChatRoom struct {
	clients map[net.Conn]string
	mu      sync.Mutex
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		clients: make(map[net.Conn]string),
	}
}

// Send a message to all clients except sender.
// If sender == nil, send to everyone.
func (cr *ChatRoom) broadcast(message string, sender net.Conn) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	for conn := range cr.clients {
		if sender != nil && conn == sender {
			continue
		}
		_, err := fmt.Fprint(conn, message)
		if err != nil {
			fmt.Println("Error broadcasting to client:", err)
		}
	}
}

func (cr *ChatRoom) addClient(conn net.Conn, username string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.clients[conn] = username
}

func (cr *ChatRoom) removeClient(conn net.Conn) string {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	username := cr.clients[conn]
	delete(cr.clients, conn)
	return username
}

func handleChatClient(conn net.Conn, cr *ChatRoom) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	// First line from client = username
	if !scanner.Scan() {
		return
	}
	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		username = "Anonymous"
	}

	cr.addClient(conn, username)
	fmt.Printf("%s joined the chat\n", username)

	joinMsg := fmt.Sprintf("*** %s joined the chat ***\n", username)
	cr.broadcast(joinMsg, nil)

	// Read chat messages line by line
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		fmt.Printf("Broadcasting from %s: %s\n", username, message)
		formatted := fmt.Sprintf("[%s]: %s\n", username, message)

		// Send only to other clients
		cr.broadcast(formatted, conn)
	}

	// Client disconnected
	leftUser := cr.removeClient(conn)
	if leftUser != "" {
		fmt.Printf("%s left the chat\n", leftUser)
		leaveMsg := fmt.Sprintf("*** %s left the chat ***\n", leftUser)
		cr.broadcast(leaveMsg, nil)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Read error:", err)
	}
}

func chatRoomServer() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("Error starting chat server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Chat server listening on :9000")

	chatRoom := NewChatRoom()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleChatClient(conn, chatRoom)
	}
}

func chatRoomClient(username string) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Send username first
	_, err = fmt.Fprintln(conn, username)
	if err != nil {
		fmt.Println("Error sending username:", err)
		return
	}

	// Read messages from server in a goroutine
	go func() {
		serverScanner := bufio.NewScanner(conn)
		for serverScanner.Scan() {
			fmt.Println(serverScanner.Text())
		}
		if err := serverScanner.Err(); err != nil {
			fmt.Println("Disconnected from server:", err)
		} else {
			fmt.Println("Server closed the connection.")
		}
		os.Exit(0)
	}()

	// Read user input and send to server
	inputScanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Enter username: %s\n", username)
	fmt.Println("Type messages (type 'exit' to quit):")

	for inputScanner.Scan() {
		text := strings.TrimSpace(inputScanner.Text())

		if text == "exit" {
			break
		}
		if text == "" {
			continue
		}

		_, err := fmt.Fprintln(conn, text)
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}
	}

	fmt.Println("Leaving chat...")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go server")
		fmt.Println("  go run main.go client <username>")
		return
	}

	mode := os.Args[1]

	switch mode {
	case "server":
		chatRoomServer()
	case "client":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a username.")
			fmt.Println("Example: go run main.go client Alice")
			return
		}
		chatRoomClient(os.Args[2])
	default:
		fmt.Println("Unknown mode. Use 'server' or 'client'.")
	}
}

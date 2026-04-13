package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Message struct {
	Username  string
	Content   string
	Timestamp time.Time
}

type UserStatus struct {
	Username string
	Status   string // "online", "typing", "away"
	LastSeen time.Time
}

type HybridChatServer struct {
	tcpClients     map[net.Conn]string
	udpClients     map[string]*net.UDPAddr
	messageHistory []Message
	userStatuses   map[string]*UserStatus
	mu             sync.RWMutex
}

func NewHybridChatServer() *HybridChatServer {
	return &HybridChatServer{
		tcpClients:     make(map[net.Conn]string),
		udpClients:     make(map[string]*net.UDPAddr),
		messageHistory: make([]Message, 0, 100),
		userStatuses:   make(map[string]*UserStatus),
	}
}

func (s *HybridChatServer) addMessage(username, content string) Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := Message{
		Username:  username,
		Content:   content,
		Timestamp: time.Now(),
	}

	if len(s.messageHistory) >= 100 {
		s.messageHistory = s.messageHistory[1:]
	}
	s.messageHistory = append(s.messageHistory, msg)
	return msg
}

func (s *HybridChatServer) updateStatus(username, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userStatuses[username] = &UserStatus{
		Username: username,
		Status:   status,
		LastSeen: time.Now(),
	}
}

func (s *HybridChatServer) setUDPClient(username string, addr *net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.udpClients[username] = addr
}

func (s *HybridChatServer) registerTCPClient(conn net.Conn, username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tcpClients[conn] = username
}

func (s *HybridChatServer) removeTCPClient(conn net.Conn) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	username := s.tcpClients[conn]
	delete(s.tcpClients, conn)
	return username
}

func (s *HybridChatServer) sendHistory(conn net.Conn, count int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if count <= 0 {
		count = 10
	}
	if count > len(s.messageHistory) {
		count = len(s.messageHistory)
	}

	start := len(s.messageHistory) - count
	for _, msg := range s.messageHistory[start:] {
		line := fmt.Sprintf("[%s] %s: %s\n",
			msg.Timestamp.Format("15:04:05"),
			msg.Username,
			msg.Content,
		)
		_, _ = conn.Write([]byte(line))
	}
}

func (s *HybridChatServer) broadcastMessage(msg Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	line := fmt.Sprintf("[%s] %s: %s\n",
		msg.Timestamp.Format("15:04:05"),
		msg.Username,
		msg.Content,
	)

	for conn := range s.tcpClients {
		_, err := conn.Write([]byte(line))
		if err != nil {
			fmt.Println("Error broadcasting TCP message:", err)
		}
	}
}

func (s *HybridChatServer) broadcastStatus(conn *net.UDPConn, username, status string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	message := fmt.Sprintf("STATUS_UPDATE:%s:%s", username, status)
	for _, addr := range s.udpClients {
		_, err := conn.WriteToUDP([]byte(message), addr)
		if err != nil {
			fmt.Println("Error broadcasting UDP status:", err)
		}
	}
}

func validStatus(status string) bool {
	return status == "online" || status == "typing" || status == "away"
}

func (s *HybridChatServer) handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	var username string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "USER:"):
			u := strings.TrimSpace(strings.TrimPrefix(line, "USER:"))
			if u != "" {
				username = u
				s.registerTCPClient(conn, username)
				fmt.Printf("%s connected via TCP\n", username)
				_, _ = conn.Write([]byte("*** Joined chat room ***\n"))
			}

		case strings.HasPrefix(line, "HISTORY:"):
			countStr := strings.TrimSpace(strings.TrimPrefix(line, "HISTORY:"))
			count, err := strconv.Atoi(countStr)
			if err != nil {
				count = 10
			}
			s.sendHistory(conn, count)

		case strings.HasPrefix(line, "MSG:"):
			parts := strings.SplitN(line, ":", 3)
			if len(parts) != 3 {
				continue
			}

			msgUser := strings.TrimSpace(parts[1])
			content := strings.TrimSpace(parts[2])
			if msgUser == "" || content == "" {
				continue
			}

			if username == "" {
				username = msgUser
				s.registerTCPClient(conn, username)
				fmt.Printf("%s connected via TCP\n", username)
			}

			msg := s.addMessage(msgUser, content)
			fmt.Printf("[TCP] %s: %s\n", msgUser, content)
			s.broadcastMessage(msg)
		}
	}

	leftUser := s.removeTCPClient(conn)
	if leftUser != "" {
		s.updateStatus(leftUser, "away")
		fmt.Printf("%s disconnected\n", leftUser)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("TCP read error:", err)
	}
}

func (s *HybridChatServer) startTCPServer() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("TCP Server listening on :9000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			continue
		}
		go s.handleTCPConnection(conn)
	}
}

func (s *HybridChatServer) startUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", ":9001")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("UDP Server listening on :9001")

	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading UDP packet:", err)
			continue
		}

		line := strings.TrimSpace(string(buffer[:n]))
		if !strings.HasPrefix(line, "STATUS:") {
			continue
		}

		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}

		username := strings.TrimSpace(parts[1])
		status := strings.TrimSpace(parts[2])

		if username == "" || !validStatus(status) {
			continue
		}

		s.setUDPClient(username, clientAddr)
		s.updateStatus(username, status)

		fmt.Printf("[UDP] %s: %s\n", username, status)
		s.broadcastStatus(conn, username, status)
	}
}

func startHybridServer() {
	server := NewHybridChatServer()

	fmt.Println("=== Hybrid Chat Server ===")
	go server.startTCPServer()
	go server.startUDPServer()

	select {}
}

// ---------------- CLIENT ----------------

func sendUDPStatus(conn *net.UDPConn, username, status string) {
	message := fmt.Sprintf("STATUS:%s:%s", username, status)
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending status:", err)
	}
}

func startHybridClient(username string) {
	// TCP connection for chat/history
	tcpConn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Error connecting TCP:", err)
		return
	}
	defer tcpConn.Close()

	// UDP socket for status updates
	udpServerAddr, err := net.ResolveUDPAddr("udp", "localhost:9001")
	if err != nil {
		fmt.Println("Error resolving UDP server:", err)
		return
	}

	udpConn, err := net.DialUDP("udp", nil, udpServerAddr)
	if err != nil {
		fmt.Println("Error connecting UDP:", err)
		return
	}
	defer udpConn.Close()

	fmt.Println("Connected to chat server (TCP: 9000, UDP: 9001)")
	fmt.Printf("Enter username: %s\n", username)

	// Register username on TCP
	_, _ = fmt.Fprintf(tcpConn, "USER:%s\n", username)

	// Send initial status
	sendUDPStatus(udpConn, username, "online")

	// Read TCP chat/history
	go func() {
		scanner := bufio.NewScanner(tcpConn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// Read UDP status updates
	go func() {
		buffer := make([]byte, 1024)
		for {
			_ = udpConn.SetReadDeadline(time.Now().Add(5 * time.Second))
			n, err := udpConn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return
			}

			line := strings.TrimSpace(string(buffer[:n]))
			if strings.HasPrefix(line, "STATUS_UPDATE:") {
				parts := strings.SplitN(line, ":", 3)
				if len(parts) == 3 {
					statusUser := parts[1]
					status := parts[2]

					switch status {
					case "typing":
						fmt.Printf("Status Update: %s is typing...\n", statusUser)
					case "online":
						fmt.Printf("Status Update: %s is online\n", statusUser)
					case "away":
						fmt.Printf("Status Update: %s is away\n", statusUser)
					}
				}
			}
		}
	}()

	// Heartbeat every 5 seconds
	stopHeartbeat := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sendUDPStatus(udpConn, username, "online")
			case <-stopHeartbeat:
				return
			}
		}
	}()

	// User input
	inputScanner := bufio.NewScanner(os.Stdin)
	for inputScanner.Scan() {
		text := strings.TrimSpace(inputScanner.Text())

		if text == "" {
			continue
		}

		if text == "exit" {
			sendUDPStatus(udpConn, username, "away")
			close(stopHeartbeat)
			fmt.Println("Leaving chat...")
			return
		}

		if strings.HasPrefix(text, "/history ") {
			countStr := strings.TrimSpace(strings.TrimPrefix(text, "/history "))
			_, _ = fmt.Fprintf(tcpConn, "HISTORY:%s\n", countStr)
			continue
		}

		if text == "/away" {
			sendUDPStatus(udpConn, username, "away")
			continue
		}

		if text == "/online" {
			sendUDPStatus(udpConn, username, "online")
			continue
		}

		sendUDPStatus(udpConn, username, "typing")
		_, err := fmt.Fprintf(tcpConn, "MSG:%s:%s\n", username, text)
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}
		sendUDPStatus(udpConn, username, "online")
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go server")
		fmt.Println("  go run main.go client <username>")
		return
	}

	switch os.Args[1] {
	case "server":
		startHybridServer()
	case "client":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a username.")
			fmt.Println("Example: go run main.go client Alice")
			return
		}
		startHybridClient(os.Args[2])
	default:
		fmt.Println("Unknown mode. Use 'server' or 'client'.")
	}
}

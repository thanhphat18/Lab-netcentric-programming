package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type PingResult struct {
	ServerAddr string
	RTT        time.Duration
	Success    bool
	ServerID   string
}

// UDP server that responds to PING with PONG <server_id> <timestamp>
func pingServer(port int, serverID string) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Error resolving UDP address for %s: %v\n", serverID, err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("Error starting %s on port %d: %v\n", serverID, port, err)
		return
	}
	defer conn.Close()

	fmt.Printf("%s listening on :%d\n", serverID, port)

	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Read error on %s: %v\n", serverID, err)
			continue
		}

		message := strings.TrimSpace(string(buffer[:n]))
		if message != "PING" {
			continue
		}

		response := fmt.Sprintf("PONG %s %d", serverID, time.Now().Unix())
		_, err = conn.WriteToUDP([]byte(response), clientAddr)
		if err != nil {
			fmt.Printf("Write error on %s: %v\n", serverID, err)
		}
	}
}

// Send one ping and measure RTT
func pingOnce(serverAddr string, timeout time.Duration) PingResult {
	result := PingResult{
		ServerAddr: serverAddr,
		Success:    false,
	}

	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return result
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return result
	}
	defer conn.Close()

	start := time.Now()

	_, err = conn.Write([]byte("PING"))
	if err != nil {
		return result
	}

	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return result
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return result
	}

	response := strings.TrimSpace(string(buffer[:n]))
	parts := strings.Split(response, " ")

	// Expected format: PONG <server_id> <timestamp>
	if len(parts) != 3 || parts[0] != "PONG" {
		return result
	}

	result.ServerID = parts[1]
	result.RTT = time.Since(start)
	result.Success = true
	return result
}

// Ping all servers concurrently every 2 seconds
func pingMonitor(servers []string) {
	timeout := 1 * time.Second

	for {
		fmt.Println("Pinging servers...")

		resultsChan := make(chan PingResult, len(servers))
		var wg sync.WaitGroup

		for _, server := range servers {
			wg.Add(1)
			go func(addr string) {
				defer wg.Done()
				resultsChan <- pingOnce(addr, timeout)
			}(server)
		}

		wg.Wait()
		close(resultsChan)

		// Store results by address so printing order matches input order
		resultMap := make(map[string]PingResult)
		for result := range resultsChan {
			resultMap[result.ServerAddr] = result
		}

		fmt.Printf("%-18s %-10s %s\n", "Server", "Status", "RTT")
		fmt.Println("------------------------------------------")

		var totalRTT time.Duration
		successCount := 0

		for _, server := range servers {
			result := resultMap[server]

			if result.Success {
				fmt.Printf("%-18s %-10s %v\n", server, "✓ Online", result.RTT)
				totalRTT += result.RTT
				successCount++
			} else {
				fmt.Printf("%-18s %-10s -\n", server, "✗ Timeout")
			}
		}

		if successCount > 0 {
			avgRTT := totalRTT / time.Duration(successCount)
			fmt.Printf("Average RTT: %v\n", avgRTT)
		} else {
			fmt.Println("Average RTT: -")
		}

		successRate := float64(successCount) / float64(len(servers)) * 100
		fmt.Printf("Success Rate: %.2f%% (%d/%d)\n", successRate, successCount, len(servers))
		fmt.Println()

		time.Sleep(2 * time.Second)
	}
}

func main() {
	fmt.Println("=== UDP Ping Monitor ===")
	fmt.Println("Starting ping servers on ports: 9001, 9002, 9003")

	// Start only two servers so 9003 times out like the lab sample
	go pingServer(9001, "server-1")
	go pingServer(9002, "server-2")
	// go pingServer(9003, "server-3") // Uncomment if you want all three online

	time.Sleep(500 * time.Millisecond)

	servers := []string{
		"localhost:9001",
		"localhost:9002",
		"localhost:9003",
	}

	pingMonitor(servers)
}

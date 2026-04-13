package main

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ServiceInfo struct {
	Name    string
	Address string
	Port    int
}

var (
	services      []ServiceInfo
	servicesMu    sync.RWMutex
	discoveryOnce sync.Once
)

// Returns a usable local IP for display.
// Falls back to 127.0.0.1 if it cannot determine one.
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || localAddr.IP == nil {
		return "127.0.0.1"
	}
	return localAddr.IP.String()
}

// Register one service and ensure the shared discovery listener is running.
func discoveryServer(serviceName string, servicePort int) {
	servicesMu.Lock()
	services = append(services, ServiceInfo{
		Name:    serviceName,
		Address: getLocalIP(),
		Port:    servicePort,
	})
	servicesMu.Unlock()

	discoveryOnce.Do(func() {
		go startDiscoveryListener()
	})
}

func startDiscoveryListener() {
	addr, err := net.ResolveUDPAddr("udp4", ":8083")
	if err != nil {
		fmt.Println("Error resolving discovery address:", err)
		return
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Println("Error starting discovery listener:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Discovery listener running on :8083")

	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading discovery request:", err)
			continue
		}

		message := strings.TrimSpace(string(buffer[:n]))
		if message != "DISCOVER" {
			continue
		}

		servicesMu.RLock()
		currentServices := make([]ServiceInfo, len(services))
		copy(currentServices, services)
		servicesMu.RUnlock()

		for _, svc := range currentServices {
			response := fmt.Sprintf("SERVICE:%s:%s:%d", svc.Name, svc.Address, svc.Port)
			_, err := conn.WriteToUDP([]byte(response), clientAddr)
			if err != nil {
				fmt.Println("Error sending service response:", err)
			}
		}
	}
}

func discoverServices() []ServiceInfo {
	// Use an ephemeral local port to send and receive responses.
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		fmt.Println("Error creating UDP socket:", err)
		return nil
	}
	defer conn.Close()

	// Lab hint mentions SetWriteBuffer(1024).
	_ = conn.SetWriteBuffer(1024)

	discoveryMessage := []byte("DISCOVER")

	// Send to broadcast address required by the lab.
	broadcastAddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:8083")
	if err == nil {
		_, _ = conn.WriteToUDP(discoveryMessage, broadcastAddr)
	}

	// Also send to localhost so local single-machine testing works reliably.
	localAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:8083")
	if err == nil {
		_, _ = conn.WriteToUDP(discoveryMessage, localAddr)
	}

	found := make(map[string]ServiceInfo)
	buffer := make([]byte, 1024)
	deadline := time.Now().Add(3 * time.Second)

	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(250 * time.Millisecond))

		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			fmt.Println("Error reading discovery response:", err)
			break
		}

		response := strings.TrimSpace(string(buffer[:n]))
		parts := strings.SplitN(response, ":", 4)
		if len(parts) != 4 || parts[0] != "SERVICE" {
			continue
		}

		port, err := strconv.Atoi(parts[3])
		if err != nil {
			continue
		}

		service := ServiceInfo{
			Name:    parts[1],
			Address: parts[2],
			Port:    port,
		}

		key := fmt.Sprintf("%s|%s|%d", service.Name, service.Address, service.Port)
		found[key] = service
	}

	result := make([]ServiceInfo, 0, len(found))
	for _, svc := range found {
		result = append(result, svc)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Port < result[j].Port
	})

	return result
}

func main() {
	fmt.Println("=== Service Discovery ===")
	fmt.Println("Starting services:")

	discoveryServer("Database Service", 5432)
	fmt.Println("- Database Service on port 5432")

	discoveryServer("Web Service", 8080)
	fmt.Println("- Web Service on port 8080")

	discoveryServer("API Service", 3000)
	fmt.Println("- API Service on port 3000")

	time.Sleep(500 * time.Millisecond)

	fmt.Println("Discovering services...")
	foundServices := discoverServices()

	fmt.Printf("Found %d services:\n", len(foundServices))
	fmt.Printf("%-20s %-15s %s\n", "Service Name", "Address", "Port")
	fmt.Println("------------------------------------------------")

	for _, svc := range foundServices {
		fmt.Printf("%-20s %-15s %d\n", svc.Name, svc.Address, svc.Port)
	}

	fmt.Println("Discovery complete!")
}

package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	tcpAddr      = "127.0.0.1:9100"
	udpAddr      = "127.0.0.1:9101"
	udpChunkSize = 508 // keep UDP packets small
	iterations   = 5
)

// -------------------- TCP SERVER --------------------

func tcpServer(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("TCP server error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("TCP test server listening on :" + port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("TCP accept error:", err)
			continue
		}
		go handleTCPConn(conn)
	}
}

func handleTCPConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Read 4-byte size header
	header := make([]byte, 4)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return
	}
	dataSize := binary.BigEndian.Uint32(header)

	// Read exactly dataSize bytes
	_, err = io.CopyN(io.Discard, conn, int64(dataSize))
	if err != nil {
		return
	}

	_, _ = conn.Write([]byte("OK"))
}

// -------------------- TCP CLIENT TEST --------------------

func tcpTransferTest(dataSize int) time.Duration {
	payload := make([]byte, dataSize)

	start := time.Now()

	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		fmt.Println("TCP dial error:", err)
		return 0
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Send size header
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, uint32(dataSize))

	if err := writeAll(conn, header); err != nil {
		fmt.Println("TCP header write error:", err)
		return 0
	}

	// Send payload
	if err := writeAll(conn, payload); err != nil {
		fmt.Println("TCP payload write error:", err)
		return 0
	}

	// Read ACK
	ack := make([]byte, 2)
	_, err = io.ReadFull(conn, ack)
	if err != nil {
		fmt.Println("TCP ack read error:", err)
		return 0
	}

	return time.Since(start)
}

// -------------------- UDP SERVER --------------------

type udpTransferState struct {
	expected int
	received int
}

func udpServer(port string) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		fmt.Println("UDP resolve error:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("UDP server error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("UDP test server listening on :" + port)

	states := make(map[string]*udpTransferState)
	buffer := make([]byte, 2048)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("UDP read error:", err)
			continue
		}

		clientKey := clientAddr.String()
		packet := buffer[:n]
		text := string(packet)

		// First packet should be metadata: SIZE:<n>
		if strings.HasPrefix(text, "SIZE:") {
			sizeStr := strings.TrimPrefix(text, "SIZE:")
			size, err := strconv.Atoi(sizeStr)
			if err != nil || size < 0 {
				continue
			}
			states[clientKey] = &udpTransferState{
				expected: size,
				received: 0,
			}
			continue
		}

		state, ok := states[clientKey]
		if !ok {
			continue
		}

		state.received += n

		if state.received >= state.expected {
			_, _ = conn.WriteToUDP([]byte("OK"), clientAddr)
			delete(states, clientKey)
		}
	}
}

// -------------------- UDP CLIENT TEST --------------------

func udpTransferTest(dataSize int) time.Duration {
	payload := make([]byte, dataSize)

	serverAddr, err := net.ResolveUDPAddr("udp", udpAddr)
	if err != nil {
		fmt.Println("UDP resolve error:", err)
		return 0
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("UDP dial error:", err)
		return 0
	}
	defer conn.Close()

	start := time.Now()

	// Send metadata first
	meta := fmt.Sprintf("SIZE:%d", dataSize)
	_, err = conn.Write([]byte(meta))
	if err != nil {
		fmt.Println("UDP metadata write error:", err)
		return 0
	}

	// Send payload in small chunks
	for i := 0; i < dataSize; i += udpChunkSize {
		end := i + udpChunkSize
		if end > dataSize {
			end = dataSize
		}

		_, err = conn.Write(payload[i:end])
		if err != nil {
			fmt.Println("UDP payload write error:", err)
			return 0
		}
	}

	// Wait for ACK with timeout
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	ack := make([]byte, 16)
	_, err = conn.Read(ack)
	if err != nil {
		fmt.Println("UDP ack read error:", err)
		return 0
	}

	return time.Since(start)
}

// -------------------- PERFORMANCE RUNNER --------------------

func runPerformanceTest() {
	sizes := []int{
		1 * 1024,
		10 * 1024,
		100 * 1024,
		1 * 1024 * 1024,
	}

	fmt.Println("=== TCP vs UDP Performance Comparison ===")
	fmt.Println("Testing data transfer speeds...")
	fmt.Printf("%-10s | %-12s | %-12s | %s\n", "Data Size", "TCP Avg Time", "UDP Avg Time", "UDP Faster By")
	fmt.Println("-----------|--------------|--------------|---------------")

	var totalImprovement float64
	validRows := 0

	for _, size := range sizes {
		var tcpTotal time.Duration
		var udpTotal time.Duration

		for i := 0; i < iterations; i++ {
			tcpTotal += tcpTransferTest(size)
			udpTotal += udpTransferTest(size)
		}

		tcpAvg := tcpTotal / iterations
		udpAvg := udpTotal / iterations

		improvement := 0.0
		if tcpAvg > 0 {
			improvement = (float64(tcpAvg-udpAvg) / float64(tcpAvg)) * 100
		}

		totalImprovement += improvement
		validRows++

		fmt.Printf(
			"%-10s | %-12s | %-12s | %.0f%%\n",
			formatSize(size),
			formatDuration(tcpAvg),
			formatDuration(udpAvg),
			math.Round(improvement),
		)
	}

	avgImprovement := 0.0
	if validRows > 0 {
		avgImprovement = totalImprovement / float64(validRows)
	}

	fmt.Println("Summary:")
	fmt.Println("- UDP is usually faster than TCP in this local benchmark")
	fmt.Printf("- Average speed improvement: %.1f%%\n", avgImprovement)
	fmt.Println("- TCP provides reliability, UDP provides speed")
}

// -------------------- HELPERS --------------------

func writeAll(conn net.Conn, data []byte) error {
	totalWritten := 0
	for totalWritten < len(data) {
		n, err := conn.Write(data[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += n
	}
	return nil
}

func formatSize(size int) string {
	switch size {
	case 1024:
		return "1 KB"
	case 10 * 1024:
		return "10 KB"
	case 100 * 1024:
		return "100 KB"
	case 1024 * 1024:
		return "1 MB"
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2f µs", float64(d.Microseconds()))
	}
	return fmt.Sprintf("%.2f ms", float64(d.Microseconds())/1000.0)
}

func main() {
	go tcpServer("9100")
	go udpServer("9101")

	time.Sleep(500 * time.Millisecond)
	runPerformanceTest()
}

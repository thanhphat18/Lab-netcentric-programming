package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func fileTransferServer() {
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Println("Error starting file transfer server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("File Transfer Server listening on :8081")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleFileTransfer(conn)
	}
}

func handleFileTransfer(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Client connected")

	reader := bufio.NewReader(conn)

	// Step 1: Read metadata line: "filename:size\n"
	metadata, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading metadata:", err)
		return
	}

	metadata = strings.TrimSpace(metadata)
	parts := strings.SplitN(metadata, ":", 2)
	if len(parts) != 2 {
		fmt.Println("Invalid metadata format")
		fmt.Fprintln(conn, "ERROR: invalid metadata format")
		return
	}

	// Use filepath.Base to avoid directory traversal / weird paths
	filename := filepath.Base(parts[0])

	fileSize, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || fileSize < 0 {
		fmt.Println("Invalid file size")
		fmt.Fprintln(conn, "ERROR: invalid file size")
		return
	}

	fmt.Printf("Receiving file: %s (%d bytes)\n", filename, fileSize)

	// Step 2: Ensure destination directory exists
	err = os.MkdirAll("./received", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		fmt.Fprintln(conn, "ERROR: cannot create directory")
		return
	}

	// Step 3: Create output file
	savePath := filepath.Join("./received", filename)
	outFile, err := os.Create(savePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		fmt.Fprintln(conn, "ERROR: cannot create file")
		return
	}
	defer outFile.Close()

	// Step 4: Read exactly fileSize bytes from the connection
	received, err := io.CopyN(outFile, reader, fileSize)
	if err != nil {
		fmt.Println("Error receiving file data:", err)
		fmt.Fprintln(conn, "ERROR: file transfer failed")
		return
	}

	if received != fileSize {
		fmt.Printf("Incomplete file received: got %d, expected %d\n", received, fileSize)
		fmt.Fprintln(conn, "ERROR: incomplete file")
		return
	}

	fmt.Printf("File saved successfully to %s\n", savePath)

	// Step 5: Send confirmation
	fmt.Fprintln(conn, "OK")
}

func sendFile(filename string, serverAddr string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	fileSize := info.Size()
	baseName := filepath.Base(filename)

	fmt.Printf("Sending file: %s\n", baseName)

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server")

	// Step 1: Send metadata line
	_, err = fmt.Fprintf(conn, "%s:%d\n", baseName, fileSize)
	if err != nil {
		fmt.Println("Error sending metadata:", err)
		return
	}

	// Step 2: Send file in 1KB chunks
	buffer := make([]byte, 1024)
	var sent int64 = 0

	for {
		n, readErr := file.Read(buffer)
		if n > 0 {
			totalWritten := 0

			// Handle partial writes safely
			for totalWritten < n {
				written, writeErr := conn.Write(buffer[totalWritten:n])
				if writeErr != nil {
					fmt.Println("Error sending file data:", writeErr)
					return
				}
				totalWritten += written
			}

			sent += int64(n)
			percentage := float64(sent) / float64(fileSize) * 100
			fmt.Printf("\rSent: %d/%d bytes (%.2f%%)", sent, fileSize, percentage)
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			fmt.Println("\nError reading file:", readErr)
			return
		}
	}

	fmt.Println()

	// Step 3: Wait for confirmation
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading confirmation:", err)
		return
	}

	response = strings.TrimSpace(response)
	if response == "OK" {
		fmt.Println("✓ File transferred successfully!")
	} else {
		fmt.Println("Server response:", response)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go server")
		fmt.Println("  go run main.go client <filename> [serverAddr]")
		return
	}

	mode := os.Args[1]

	switch mode {
	case "server":
		fileTransferServer()

	case "client":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a filename.")
			fmt.Println("Example: go run main.go client document.txt")
			return
		}

		serverAddr := "localhost:8081"
		if len(os.Args) >= 4 {
			serverAddr = os.Args[3]
		}

		sendFile(os.Args[2], serverAddr)

	default:
		fmt.Println("Unknown mode. Use 'server' or 'client'.")
	}
}

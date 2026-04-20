package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "task2/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCalculatorClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test 1: Addition
	fmt.Println("=== Test 1: Addition ===")
	resp, err := client.Calculate(ctx, &pb.CalculateRequest{
		A:         10,
		B:         5,
		Operation: "add",
	})
	if err != nil {
		log.Fatalf("Addition failed: %v", err)
	}
	fmt.Printf("Result: %.2f + %.2f = %.2f\n", 10.0, 5.0, resp.Result)

	// Test 2: Division
	fmt.Println("\n=== Test 2: Division ===")
	resp, err = client.Calculate(ctx, &pb.CalculateRequest{
		A:         20,
		B:         4,
		Operation: "divide",
	})
	if err != nil {
		log.Fatalf("Division failed: %v", err)
	}
	fmt.Printf("Result: %.2f / %.2f = %.2f\n", 20.0, 4.0, resp.Result)

	// Test 3: Division by zero
	fmt.Println("\n=== Test 3: Division by Zero ===")
	_, err = client.Calculate(ctx, &pb.CalculateRequest{
		A:         10,
		B:         0,
		Operation: "divide",
	})
	if err != nil {
		st, _ := status.FromError(err)
		fmt.Printf("Expected error: %s\n", st.Message())
	}

	// Test 4: Square root
	fmt.Println("\n=== Test 4: Square Root ===")
	sqrtResp, err := client.SquareRoot(ctx, &pb.SquareRootRequest{
		Number: 16,
	})
	if err != nil {
		log.Fatalf("SquareRoot failed: %v", err)
	}
	fmt.Printf("Result: sqrt(%.2f) = %.2f\n", 16.0, sqrtResp.Result)

	// Test 5: Negative square root
	fmt.Println("\n=== Test 5: Negative Square Root ===")
	_, err = client.SquareRoot(ctx, &pb.SquareRootRequest{
		Number: -4,
	})
	if err != nil {
		st, _ := status.FromError(err)
		fmt.Printf("Expected error: %s\n", st.Message())
	}

	// Test 6: History
	fmt.Println("\n=== Test 6: History ===")
	historyResp, err := client.GetHistory(ctx, &pb.HistoryRequest{})
	if err != nil {
		log.Fatalf("GetHistory failed: %v", err)
	}

	fmt.Printf("Calculations: %d\n", historyResp.Count)
	for i, calc := range historyResp.Calculations {
		fmt.Printf("%d. %s\n", i+1, calc)
	}
}
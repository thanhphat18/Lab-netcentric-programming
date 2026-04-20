package main

import (
	"context"
	"fmt"
	"log"
	"time"
	
	pb "simplegRPC/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:50051", 
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	
	// Create client
	client := pb.NewGreeterClient(conn)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	
	// Call SayHello
	fmt.Println("Calling SayHello...")
	helloResp, err := client.SayHello(ctx, &pb.HelloRequest{
		Name: "John",
	})
	if err != nil {
		log.Fatalf("SayHello failed: %v", err)
	}
	
	fmt.Printf("Response: %s\n", helloResp.Message)
	fmt.Printf("Call count: %d\n\n", helloResp.Count)
	
	// Call SayGoodbye
	fmt.Println("Calling SayGoodbye...")
	byeResp, err := client.SayGoodbye(ctx, &pb.HelloRequest{
		Name: "John",
	})
	if err != nil {
		log.Fatalf("SayGoodbye failed: %v", err)
	}
	
	fmt.Printf("Response: %s\n", byeResp.Message)
	fmt.Printf("Call count: %d\n", byeResp.Count)
}
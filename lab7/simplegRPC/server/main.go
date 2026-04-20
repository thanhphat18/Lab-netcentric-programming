package main

import (
	"context"
	"fmt"
	"log"
	"net"
	
	pb "simplegRPC/proto"
	"google.golang.org/grpc"
)

// Server implements the Greeter service
type server struct {
	pb.UnimplementedGreeterServer  // Embed for forward compatibility
	callCount int                   // Track number of calls
}

// SayHello implements the SayHello RPC method
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	s.callCount++
	
	log.Printf("Received SayHello request: name=%s", req.Name)
	
	// Create response
	message := fmt.Sprintf("Hello, %s! Welcome to gRPC.", req.Name)
	
	return &pb.HelloResponse{
		Message: message,
		Count:   int32(s.callCount),
	}, nil
}

// SayGoodbye implements the SayGoodbye RPC method
func (s *server) SayGoodbye(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	s.callCount++
	
	log.Printf("Received SayGoodbye request: name=%s", req.Name)
	
	message := fmt.Sprintf("Goodbye, %s! Thanks for using gRPC.", req.Name)
	
	return &pb.HelloResponse{
		Message: message,
		Count:   int32(s.callCount),
	}, nil
}

func main() {
	// Create TCP listener
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	// Create gRPC server
	grpcServer := grpc.NewServer()
	
	// Register our service
	pb.RegisterGreeterServer(grpcServer, &server{})
	
	log.Println("🚀 gRPC server listening on :50051")
	
	// Start serving
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "task4/proto"

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

	client := pb.NewBookCatalogClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test 1: search books by title
	fmt.Println("=== Test 1: Search Books by Title ===")
	searchResp, err := client.SearchBooks(ctx, &pb.SearchBooksRequest{
		Query: "go",
		Field: "title",
	})
	if err != nil {
		log.Fatalf("SearchBooks failed: %v", err)
	} else {
		fmt.Printf("Found %d books: ", searchResp.Count)
		for _, b := range searchResp.Books {
			fmt.Printf("- %s by %s\n", b.Title, b.Author)
		}
	}
	
	// Test 2: search books by author
	fmt.Println("\n=== Test 2: Search Books by Author ===")
	searchResp, err = client.SearchBooks(ctx, &pb.SearchBooksRequest{
		Query: "martin",
		Field: "author",
	})
	if err != nil {
		log.Printf("SearchBooks failed: %v", err)
	} else {
		fmt.Printf("Found %d books: ", searchResp.Count)
		for _, b := range searchResp.Books {
			fmt.Printf("- %s by %s\n", b.Title, b.Author)
		}
	}

	//Test 3: Filter by price
	fmt.Println("\n=== Test 3: Filter Books by Price ===")
	filterResp, err := client.FilterBooks(ctx, &pb.FilterBooksRequest{
		MinPrice: 20.0,
		MaxPrice: 40.0,
	})
	if err != nil {
		log.Printf("FilterBooks failed: %v", err)
	} else {
		fmt.Printf("Found %d books in price range $20-$40:\n", filterResp.Count)
		for _, b := range filterResp.Books {
			fmt.Printf("- %s by %s - $%.2f\n", b.Title, b.Author, b.Price)
		}
	}

	//Test 4: filter by year
	fmt.Println("\n=== Test 4: Filter Books by Year ===")
	filterResp, err = client.FilterBooks(ctx, &pb.FilterBooksRequest{
		MinYear: 2010,
	})
	if err != nil {
		log.Printf("FilterBooks failed: %v", err)
	} else {
		fmt.Printf("Found %d books published after 2010:\n", filterResp.Count)
		for _, b := range filterResp.Books {
			fmt.Printf("- %s by %s - Published in %d\n", b.Title, b.Author, b.PublishedYear)
		}
	}

	//Test 5: Get statistics
	fmt.Println("\n=== Test 5: Get Book Statistics ===")
	statsResp, err := client.GetStats(ctx, &pb.GetStatsRequest{})
	if err != nil {
		log.Printf("GetStatistics failed: %v", err)
	} else {
		fmt.Printf("Total books: %d\n", statsResp.TotalBooks)
		fmt.Printf("Average price: $%.2f\n", statsResp.AveragePrice)
		fmt.Printf("Books in stock: %d\n", statsResp.TotalStock)
	}

	//Test 6: error cases
	_, err = client.SearchBooks(ctx, &pb.SearchBooksRequest{
		Query: "",
		Field: "title",
	})
	if err != nil{
		st, _ := status.FromError(err)
		fmt.Printf("\nExpected error for empty query: %s\n", st.Message())
	}
	_, err = client.FilterBooks(ctx, &pb.FilterBooksRequest{
		MinPrice: 50.0,
		MaxPrice: 20.0,
	})
	if err != nil{
		st, _ := status.FromError(err)
		fmt.Printf("\nExpected error for invalid price range: %s\n", st.Message())
	}
}
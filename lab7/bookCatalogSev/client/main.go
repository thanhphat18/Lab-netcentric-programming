package main

import (
	"context"
	"fmt"
	"log"
	"time"
	
	pb "bookCatalogSev/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to server
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	
	client := pb.NewBookCatalogClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Test 1: List all books
	fmt.Println("=== Test 1: List Books ===")
	listResp, err := client.ListBooks(ctx, &pb.ListBooksRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Fatalf("ListBooks failed: %v", err)
	}
	
	fmt.Printf("Total books: %d\n", listResp.Total)
	for i, book := range listResp.Books {
		fmt.Printf("%d. %s by %s - $%.2f\n", i+1, book.Title, book.Author, book.Price)
	}
	
	// Test 2: Get specific book
	fmt.Println("\n=== Test 2: Get Book ===")
	getResp, err := client.GetBook(ctx, &pb.GetBookRequest{Id: 1})
	if err != nil {
		log.Printf("GetBook failed: %v", err)
	} else {
		book := getResp.Book
		fmt.Printf("Book: %s\n", book.Title)
		fmt.Printf("Author: %s\n", book.Author)
		fmt.Printf("Price: $%.2f\n", book.Price)
		fmt.Printf("Stock: %d\n", book.Stock)
	}
	
	// Test 3: Create new book
	fmt.Println("\n=== Test 3: Create Book ===")
	createResp, err := client.CreateBook(ctx, &pb.CreateBookRequest{
		Title:         "Learning Go",
		Author:        "Jon Bodner",
		Isbn:          "978-1492077213",
		Price:         44.99,
		Stock:         10,
		PublishedYear: 2021,
	})
	if err != nil {
		log.Printf("CreateBook failed: %v", err)
	} else {
		fmt.Printf("Created book ID: %d\n", createResp.Book.Id)
		fmt.Printf("Title: %s\n", createResp.Book.Title)
	}
	
	// Test 4: Search books
	fmt.Println("\n=== Test 4: Search Books ===")
	searchResp, err := client.SearchBooks(ctx, &pb.SearchBooksRequest{
		Query: "go",
	})
	if err != nil {
		log.Printf("SearchBooks failed: %v", err)
	} else {
		fmt.Printf("Found %d books:\n", searchResp.Count)
		for _, book := range searchResp.Books {
			fmt.Printf("- %s by %s\n", book.Title, book.Author)
		}
	}
	
	// Test 5: Update book
	fmt.Println("\n=== Test 5: Update Book ===")
	updateResp, err := client.UpdateBook(ctx, &pb.UpdateBookRequest{
		Id:            1,
		Title:         "The Go Programming Language (Updated)",
		Author:        "Alan Donovan",
		Isbn:          "978-0134190440",
		Price:         35.99,
		Stock:         25,
		PublishedYear: 2015,
	})
	if err != nil {
		log.Printf("UpdateBook failed: %v", err)
	} else {
		fmt.Printf("Updated book: %s\n", updateResp.Book.Title)
		fmt.Printf("New price: $%.2f\n", updateResp.Book.Price)
	}
	
	// Test 6: Delete book (commented out to preserve data)
	// fmt.Println("\n=== Test 6: Delete Book ===")
	// deleteResp, err := client.DeleteBook(ctx, &pb.DeleteBookRequest{Id: 4})
	// if err != nil {
	// 	log.Printf("DeleteBook failed: %v", err)
	// } else {
	// 	fmt.Printf("%s\n", deleteResp.Message)
	// }
	
	fmt.Println("\n✓ All tests completed!")
}
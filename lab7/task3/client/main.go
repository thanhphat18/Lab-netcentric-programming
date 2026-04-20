package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "bookcatalog-grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// Test 1: List All Books
	fmt.Println("=== Test 1: List All Books ===")
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

	// Test 2: Get Book
	fmt.Println("\n=== Test 2: Get Book ===")
	getResp, err := client.GetBook(ctx, &pb.GetBookRequest{Id: 1})
	if err != nil {
		log.Printf("GetBook failed: %v", err)
	} else {
		fmt.Printf("Book ID: %d\n", getResp.Book.Id)
		fmt.Printf("Title: %s\n", getResp.Book.Title)
		fmt.Printf("Author: %s\n", getResp.Book.Author)
		fmt.Printf("Price: $%.2f\n", getResp.Book.Price)
	}

	// Test 3: Create Book
	fmt.Println("\n=== Test 3: Create Book ===")
	createResp, err := client.CreateBook(ctx, &pb.CreateBookRequest{
		Title:         "Go Web Services",
		Author:        "John Smith",
		Isbn:          "978-1234567890",
		Price:         36.99,
		Stock:         9,
		PublishedYear: 2022,
	})
	if err != nil {
		log.Printf("CreateBook failed: %v", err)
	} else {
		fmt.Printf("Created book ID: %d\n", createResp.Book.Id)
		fmt.Printf("Title: %s\n", createResp.Book.Title)
	}

	// Test 4: Update Book
	fmt.Println("\n=== Test 4: Update Book ===")
	updateResp, err := client.UpdateBook(ctx, &pb.UpdateBookRequest{
		Id:            1,
		Title:         "The Go Programming Language (2nd Edition)",
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

	// Test 5: Delete Book
	fmt.Println("\n=== Test 5: Delete Book ===")
	deleteResp, err := client.DeleteBook(ctx, &pb.DeleteBookRequest{Id: createResp.Book.Id})
	if err != nil {
		log.Printf("DeleteBook failed: %v", err)
	} else {
		fmt.Printf("%s\n", deleteResp.Message)
	}

	// Test 6: Pagination
	fmt.Println("\n=== Test 6: Pagination ===")
	page1, err := client.ListBooks(ctx, &pb.ListBooksRequest{
		Page:     1,
		PageSize: 3,
	})
	if err != nil {
		log.Printf("Pagination page 1 failed: %v", err)
	} else {
		fmt.Printf("Page 1: %d books\n", len(page1.Books))
		for _, b := range page1.Books {
			fmt.Printf("- %s\n", b.Title)
		}
	}

	page2, err := client.ListBooks(ctx, &pb.ListBooksRequest{
		Page:     2,
		PageSize: 3,
	})
	if err != nil {
		log.Printf("Pagination page 2 failed: %v", err)
	} else {
		fmt.Printf("Page 2: %d books\n", len(page2.Books))
		for _, b := range page2.Books {
			fmt.Printf("- %s\n", b.Title)
		}
	}

	fmt.Println("\nTask 3 client tests completed")
}
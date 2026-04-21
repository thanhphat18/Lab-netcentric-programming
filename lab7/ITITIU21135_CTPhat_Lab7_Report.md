---
title: gRCP
tags: [template]
---

# Net-Centric Programming Lab 07 Report
Full Name: Chau Thanh Phat
Student ID: ITITIU21135
___
**Content**
Learn to build high-performance gRPC services using Protocol Buffers,
implementing microservice communication patterns with type safety and automatic code
generation.
___
## Task 1: Protocol Buffer Basic
- create book.proto, define Book, BookCategory, DetailedBook, and Author
- compile .proto file
- write main.go to create objects, serialize with protobuf, deserialize, and print values
- grading to check serialization/deserializtion works correctly
### main.go
'''
package main

import (
	"fmt"
	"log"

	pb "task1/proto"  // Import generated code
	"google.golang.org/protobuf/proto"

)

func main() {
	//Create a book
	book := &pb.Book {
		Id: 1,
		Title: "The Go Programming Language",
		Author: "Alan A. A. Donovan",
		Isbn: "978-0134190440",
		Price: 35.99,
		Stock: 15,
		PublishedYear: 2015,
	}
	fmt.Printf("Book: %v\n", book)

	//Create detailed book
	detailedBook := &pb.DetailedBook{
		Book: book,
		Category: pb.BookCategory_NONFICTION,
		Description: "A comprehensive introduction to Go programming",
		Tags: []string{"programming", "go","technical"},
		Rating: 4.5,
	}
	fmt.Printf("Detailed Book: %v\n", detailedBook)
	fmt.Printf("Category: %s\n", detailedBook.Category)
	fmt.Printf("Tags: %v\n", detailedBook.Tags)
	fmt.Printf("Rating: %.1f\n", detailedBook.Rating)

	//Serialize Book to bytes
	data, err := proto.Marshal(book) //convert Book object into compact protobuf binary bytes
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Serialized Book size: %d bytes\n", len(data))

	//Deserialize bytes back to Book
	newBook := &pb.Book{}
	err = proto.Unmarshal(data, newBook) //reconstruct Book from those type
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deserialized Book: %v\n", newBook)

	//Create an author with multiple books
	author := &pb.Author{
		Id:        1,
		Name:      "Robert C. Martin",
		Bio:       "Software engineer and author of clean code principles.",
		BirthYear: 1952,
		Books: []*pb.Book{
			{
				Id:            2,
				Title:         "Clean Code",
				Author:        "Robert C. Martin",
				Isbn:          "978-0132350884",
				Price:         42.50,
				Stock:         20,
				PublishedYear: 2008,
			},
			{
				Id:            3,
				Title:         "Clean Architecture",
				Author:        "Robert C. Martin",
				Isbn:          "978-0134494166",
				Price:         45.99,
				Stock:         12,
				PublishedYear: 2017,
			},
		},
	}

	fmt.Printf("\nAuthor: %s\n", author.Name)
	fmt.Printf("Books written: %d\n", len(author.Books))
	for i,b := range author.Books{
		fmt.Printf("%d. %s\n", i+1, b.Title)
	}

}

'''
### book.proto
'''
syntax = "proto3"; 

package bookstore;
option go_package = "task1/proto";

message Book {
    int32 id = 1;
    string title = 2;
    string author = 3;
    string isbn = 4;
    float price = 5;
    int32 stock = 6;
    int32 published_year = 7;
}

enum BookCategory {
  UNKNOWN = 0;
  FICTION = 1;
  NONFICTION = 2;
  SCIFI = 3;
  FANTASY = 4;
  MYSTERY = 5;
  BIOGRAPHY = 6;
}

message DetailedBook {
  Book book = 1;
  BookCategory category = 2;
  string description = 3;
  repeated string tags = 4;
  float rating = 5;
}

message Author {
  int32 id = 1;
  string name = 2;
  string bio = 3;
  int32 birth_year = 4;
  repeated Book books = 5;
}
'''
___
## Task 2: Remote function call system
- build a Calculator gRPC service with three RPC methods: Calculate, SquareRoot, GetHistory
- server must support add/substract/multiply/divide
- server must store calculation history in memory
- server must return proper errors for division by zero and negative squareroot
- client must test normal cases, error cases, and print the history
### proto/calculator.proto
'''
syntax = "proto3";

package calculator;
option go_package = "task2/proto";

message CalculateRequest {
  float a = 1;
  float b = 2;
  string operation = 3; // "add", "subtract", "multiply", "divide"
}

message CalculateResponse {
  float result = 1;
  string operation = 2;
}

message SquareRootRequest {
  float number = 1;
}

message SquareRootResponse {
  float result = 1;
}

message HistoryRequest {
}

message HistoryResponse {
  repeated string calculations = 1;
  int32 count = 2;
}

service Calculator {
  rpc Calculate(CalculateRequest) returns (CalculateResponse);
  rpc SquareRoot(SquareRootRequest) returns (SquareRootResponse);
  rpc GetHistory(HistoryRequest) returns (HistoryResponse);
}
'''
### client/main.go
'''
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
'''
### server/main.go
'''
package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"

	pb "task2/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type calculatorServer struct {
	pb.UnimplementedCalculatorServer
	history []string
}

func (s *calculatorServer) Calculate(ctx context.Context, req *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	log.Printf("Calculate: %.2f %s %.2f", req.A, req.Operation, req.B)

	var result float32

	switch req.Operation {
	case "add":
		result = req.A + req.B
	case "subtract":
		result = req.A - req.B
	case "multiply":
		result = req.A * req.B
	case "divide":
		if req.B == 0 {
			return nil, status.Error(codes.InvalidArgument, "cannot divide by zero")
		}
		result = req.A / req.B
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown operation: %s", req.Operation)
	}

	historyEntry := fmt.Sprintf("%.2f %s %.2f = %.2f", req.A, req.Operation, req.B, result)
	s.history = append(s.history, historyEntry)

	return &pb.CalculateResponse{
		Result:    result,
		Operation: req.Operation,
	}, nil
}

func (s *calculatorServer) SquareRoot(ctx context.Context, req *pb.SquareRootRequest) (*pb.SquareRootResponse, error) {
	log.Printf("SquareRoot: %.2f", req.Number)

	if req.Number < 0 {
		return nil, status.Errorf(codes.InvalidArgument,
			"cannot calculate square root of negative number: %.2f", req.Number)
	}

	result := float32(math.Sqrt(float64(req.Number)))

	historyEntry := fmt.Sprintf("sqrt(%.2f) = %.2f", req.Number, result)
	s.history = append(s.history, historyEntry)

	return &pb.SquareRootResponse{
		Result: result,
	}, nil
}

func (s *calculatorServer) GetHistory(ctx context.Context, req *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	log.Println("GetHistory called")

	return &pb.HistoryResponse{
		Calculations: s.history,
		Count:        int32(len(s.history)),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterCalculatorServer(grpcServer, &calculatorServer{
		history: []string{},
	})

	log.Println("Calculator gRPC server listening on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
'''
___
## Task 3: Real Microservice with persistence
- book_service.proto
- server with SQLite initialization
- seed data with at least 5 books
- RPC method: GetBook, CreateBook, UpdateBook, DeleteBook, ListBook
- proper error handling
- use of context for DB operations
- client tests for CRUD and pagination
### proto/book_service.proto
'''
syntax = "proto3";

package bookservice;
option go_package = "bookcatalog-grpc/proto";

message Book {
  int32 id = 1;
  string title = 2;
  string author = 3;
  string isbn = 4;
  float price = 5;
  int32 stock = 6;
  int32 published_year = 7;
}

message GetBookRequest {
  int32 id = 1;
}

message GetBookResponse {
  Book book = 1;
}

message CreateBookRequest {
  string title = 1;
  string author = 2;
  string isbn = 3;
  float price = 4;
  int32 stock = 5;
  int32 published_year = 6;
}

message CreateBookResponse {
  Book book = 1;
}

message UpdateBookRequest {
  int32 id = 1;
  string title = 2;
  string author = 3;
  string isbn = 4;
  float price = 5;
  int32 stock = 6;
  int32 published_year = 7;
}

message UpdateBookResponse {
  Book book = 1;
}

message DeleteBookRequest {
  int32 id = 1;
}

message DeleteBookResponse {
  bool success = 1;
  string message = 2;
}

message ListBooksRequest {
  int32 page = 1;
  int32 page_size = 2;
}

message ListBooksResponse {
  repeated Book books = 1;
  int32 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}

service BookCatalog {
  rpc GetBook(GetBookRequest) returns (GetBookResponse);
  rpc CreateBook(CreateBookRequest) returns (CreateBookResponse);
  rpc UpdateBook(UpdateBookRequest) returns (UpdateBookResponse);
  rpc DeleteBook(DeleteBookRequest) returns (DeleteBookResponse);
  rpc ListBooks(ListBooksRequest) returns (ListBooksResponse);
}
'''
### client/main.go
'''
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
'''
### server/main.go
'''
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	pb "bookcatalog-grpc/proto"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type bookCatalogServer struct {
	pb.UnimplementedBookCatalogServer
	db *sql.DB
}

func newServer(db *sql.DB) *bookCatalogServer {
	return &bookCatalogServer{db: db}
}

func (s *bookCatalogServer) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	log.Printf("GetBook called: id=%d", req.Id)

	var book pb.Book
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, title, author, isbn, price, stock, published_year FROM books WHERE id = ?",
		req.Id,
	).Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, &book.Price, &book.Stock, &book.PublishedYear)

	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "book not found: id=%d", req.Id)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	return &pb.GetBookResponse{Book: &book}, nil
}

func (s *bookCatalogServer) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.CreateBookResponse, error) {
	log.Printf("CreateBook called: title=%s", req.Title)

	if req.Title == "" || req.Author == "" {
		return nil, status.Error(codes.InvalidArgument, "title and author are required")
	}
	if req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be positive")
	}
	if req.Stock < 0 {
		return nil, status.Error(codes.InvalidArgument, "stock cannot be negative")
	}

	result, err := s.db.ExecContext(
		ctx,
		"INSERT INTO books (title, author, isbn, price, stock, published_year) VALUES (?, ?, ?, ?, ?, ?)",
		req.Title, req.Author, req.Isbn, req.Price, req.Stock, req.PublishedYear,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create book: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get inserted id: %v", err)
	}

	book := &pb.Book{
		Id:            int32(id),
		Title:         req.Title,
		Author:        req.Author,
		Isbn:          req.Isbn,
		Price:         req.Price,
		Stock:         req.Stock,
		PublishedYear: req.PublishedYear,
	}

	return &pb.CreateBookResponse{Book: book}, nil
}

func (s *bookCatalogServer) UpdateBook(ctx context.Context, req *pb.UpdateBookRequest) (*pb.UpdateBookResponse, error) {
	log.Printf("UpdateBook called: id=%d", req.Id)

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid book id is required")
	}
	if req.Title == "" || req.Author == "" {
		return nil, status.Error(codes.InvalidArgument, "title and author are required")
	}
	if req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be positive")
	}
	if req.Stock < 0 {
		return nil, status.Error(codes.InvalidArgument, "stock cannot be negative")
	}

	result, err := s.db.ExecContext(
		ctx,
		"UPDATE books SET title=?, author=?, isbn=?, price=?, stock=?, published_year=? WHERE id=?",
		req.Title, req.Author, req.Isbn, req.Price, req.Stock, req.PublishedYear, req.Id,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update book: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check update result: %v", err)
	}
	if rowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "book not found: id=%d", req.Id)
	}

	book := &pb.Book{
		Id:            req.Id,
		Title:         req.Title,
		Author:        req.Author,
		Isbn:          req.Isbn,
		Price:         req.Price,
		Stock:         req.Stock,
		PublishedYear: req.PublishedYear,
	}

	return &pb.UpdateBookResponse{Book: book}, nil
}

func (s *bookCatalogServer) DeleteBook(ctx context.Context, req *pb.DeleteBookRequest) (*pb.DeleteBookResponse, error) {
	log.Printf("DeleteBook called: id=%d", req.Id)

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid book id is required")
	}

	result, err := s.db.ExecContext(ctx, "DELETE FROM books WHERE id = ?", req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete book: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check delete result: %v", err)
	}
	if rowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "book not found: id=%d", req.Id)
	}

	return &pb.DeleteBookResponse{
		Success: true,
		Message: fmt.Sprintf("Book %d deleted successfully", req.Id),
	}, nil
}

func (s *bookCatalogServer) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	log.Printf("ListBooks called: page=%d, pageSize=%d", req.Page, req.PageSize)

	page := req.Page
	if page < 1 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 3
	}

	offset := (page - 1) * pageSize

	var total int32
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books").Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count books: %v", err)
	}

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT id, title, author, isbn, price, stock, published_year FROM books ORDER BY id LIMIT ? OFFSET ?",
		pageSize, offset,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list books: %v", err)
	}
	defer rows.Close()

	var books []*pb.Book
	for rows.Next() {
		var book pb.Book
		err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, &book.Price, &book.Stock, &book.PublishedYear)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan row: %v", err)
		}
		books = append(books, &book)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "row iteration error: %v", err)
	}

	return &pb.ListBooksResponse{
		Books:    books,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./books.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		isbn TEXT,
		price REAL NOT NULL,
		stock INTEGER DEFAULT 0,
		published_year INTEGER
	)
	`)
	if err != nil {
		return nil, err
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		sampleBooks := []struct {
			title  string
			author string
			isbn   string
			price  float32
			stock  int
			year   int
		}{
			{"The Go Programming Language", "Alan Donovan", "978-0134190440", 39.99, 15, 2015},
			{"Clean Code", "Robert Martin", "978-0132350884", 42.50, 20, 2008},
			{"Design Patterns", "Gang of Four", "978-0201633610", 54.99, 8, 1994},
			{"Learning Go", "Jon Bodner", "978-1492077213", 44.99, 10, 2021},
			{"Clean Architecture", "Robert Martin", "978-0134494166", 45.99, 12, 2017},
		}

		for _, b := range sampleBooks {
			_, err := db.Exec(
				"INSERT INTO books (title, author, isbn, price, stock, published_year) VALUES (?, ?, ?, ?, ?, ?)",
				b.title, b.author, b.isbn, b.price, b.stock, b.year,
			)
			if err != nil {
				return nil, err
			}
		}
		log.Println("Database seeded with sample books")
	}

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterBookCatalogServer(grpcServer, newServer(db))

	log.Println("BookCatalog gRPC server listening on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
'''
___
## Task 4: Extended version of Book service
- add three new RPCs - SearchBooks, FilterBooks, and GetStats
### proto/book_service.proto
'''
syntax = "proto3";

package bookservice;
option go_package = "task4/proto";

message Book {
  int32 id = 1;
  string title = 2;
  string author = 3;
  string isbn = 4;
  float price = 5;
  int32 stock = 6;
  int32 published_year = 7;
}

message GetBookRequest {
    int32 id = 1;
}

message GetBookResponse {
    Book book = 1;
}

message CreateBookRequest {
    string title = 1;
    string author = 2;
    string isbn = 3;
    float price = 4;
    int32 stock = 5;
    int32 published_year = 6;
}

message CreateBookResponse {
    Book book = 1;
}

message UpdateBookRequest {
    int32 id = 1;
    string title = 2;
    string author = 3;
    string isbn = 4;
    float price = 5;
    int32 stock = 6;
    int32 published_year = 7;
}

message UpdateBookResponse {
    Book book = 1;
}

message DeleteBookRequest {
    int32 id = 1;
}

message DeleteBookResponse {
    bool success = 1;
    string message = 2;
}

message ListBooksRequest {
    int32 page = 1;
    int32 page_size = 2;
}

message ListBooksResponse {
    repeated Book books = 1;
    int32 total = 2;
    int32 page = 3;
    int32 page_size = 4;
}

message SearchBooksRequest{
    string query = 1;
    string field = 2; //"title", "isbn", or "all"
}

message SearchBooksResponse {
    repeated Book books = 1;
    int32 count = 2;
    string query = 3;
}

message FilterBooksRequest {
    float min_price = 1;
    float max_price = 2;
    int32 min_year = 3;
    int32 max_year = 4;
}

message FilterBooksResponse {
    repeated Book books = 1;
    int32 count = 2;
}

message GetStatsRequest {}

message GetStatsResponse{
    int32 total_books = 1;
    float average_price = 2;
    int32 total_stock = 3;
    int32 earliest_year = 4;
    int32 latest_year = 5;
}



service BookCatalog {
    rpc GetBook(GetBookRequest) returns (GetBookResponse);
    rpc CreateBook(CreateBookRequest) returns (CreateBookResponse);
    rpc UpdateBook(UpdateBookRequest) returns (UpdateBookResponse);
    rpc DeleteBook(DeleteBookRequest) returns (DeleteBookResponse);
    rpc ListBooks(ListBooksRequest) returns (ListBooksResponse);

    rpc SearchBooks(SearchBooksRequest) returns (SearchBooksResponse);
    rpc FilterBooks(FilterBooksRequest) returns (FilterBooksResponse);
    rpc GetStats(GetStatsRequest) returns (GetStatsResponse);
}
'''
### client/main.go
'''
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
'''
### server/main.go
'''
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	pb "task4/proto"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type bookCatalogServer struct {
	pb.UnimplementedBookCatalogServer
	db *sql.DB
}

func newServer(db *sql.DB) *bookCatalogServer {
	return &bookCatalogServer{db: db}
}

func (s *bookCatalogServer) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	log.Printf("GetBook called: id=%d", req.Id)

	var book pb.Book
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, title, author, isbn, price, stock, published_year FROM books WHERE id = ?",
		req.Id,
	).Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, &book.Price, &book.Stock, &book.PublishedYear)

	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "book not found: id=%d", req.Id)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	return &pb.GetBookResponse{Book: &book}, nil
}

func (s *bookCatalogServer) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.CreateBookResponse, error) {
	log.Printf("CreateBook called: title=%s", req.Title)

	if req.Title == "" || req.Author == "" {
		return nil, status.Error(codes.InvalidArgument, "title and author are required")
	}
	if req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be positive")
	}
	if req.Stock < 0 {
		return nil, status.Error(codes.InvalidArgument, "stock cannot be negative")
	}

	result, err := s.db.ExecContext(
		ctx,
		"INSERT INTO books (title, author, isbn, price, stock, published_year) VALUES (?, ?, ?, ?, ?, ?)",
		req.Title, req.Author, req.Isbn, req.Price, req.Stock, req.PublishedYear,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create book: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get inserted id: %v", err)
	}

	book := &pb.Book{
		Id:            int32(id),
		Title:         req.Title,
		Author:        req.Author,
		Isbn:          req.Isbn,
		Price:         req.Price,
		Stock:         req.Stock,
		PublishedYear: req.PublishedYear,
	}

	return &pb.CreateBookResponse{Book: book}, nil
}

func (s *bookCatalogServer) UpdateBook(ctx context.Context, req *pb.UpdateBookRequest) (*pb.UpdateBookResponse, error) {
	log.Printf("UpdateBook called: id=%d", req.Id)

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid book id is required")
	}
	if req.Title == "" || req.Author == "" {
		return nil, status.Error(codes.InvalidArgument, "title and author are required")
	}
	if req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be positive")
	}
	if req.Stock < 0 {
		return nil, status.Error(codes.InvalidArgument, "stock cannot be negative")
	}

	result, err := s.db.ExecContext(
		ctx,
		"UPDATE books SET title=?, author=?, isbn=?, price=?, stock=?, published_year=? WHERE id=?",
		req.Title, req.Author, req.Isbn, req.Price, req.Stock, req.PublishedYear, req.Id,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update book: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check update result: %v", err)
	}
	if rowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "book not found: id=%d", req.Id)
	}

	book := &pb.Book{
		Id:            req.Id,
		Title:         req.Title,
		Author:        req.Author,
		Isbn:          req.Isbn,
		Price:         req.Price,
		Stock:         req.Stock,
		PublishedYear: req.PublishedYear,
	}

	return &pb.UpdateBookResponse{Book: book}, nil
}

func (s *bookCatalogServer) DeleteBook(ctx context.Context, req *pb.DeleteBookRequest) (*pb.DeleteBookResponse, error) {
	log.Printf("DeleteBook called: id=%d", req.Id)

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid book id is required")
	}

	result, err := s.db.ExecContext(ctx, "DELETE FROM books WHERE id = ?", req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete book: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check delete result: %v", err)
	}
	if rowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "book not found: id=%d", req.Id)
	}

	return &pb.DeleteBookResponse{
		Success: true,
		Message: fmt.Sprintf("Book %d deleted successfully", req.Id),
	}, nil
}

func (s *bookCatalogServer) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	log.Printf("ListBooks called: page=%d, pageSize=%d", req.Page, req.PageSize)

	page := req.Page
	if page < 1 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 3
	}

	offset := (page - 1) * pageSize

	var total int32
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books").Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count books: %v", err)
	}

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT id, title, author, isbn, price, stock, published_year FROM books ORDER BY id LIMIT ? OFFSET ?",
		pageSize, offset,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list books: %v", err)
	}
	defer rows.Close()

	var books []*pb.Book
	for rows.Next() {
		var book pb.Book
		err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, &book.Price, &book.Stock, &book.PublishedYear)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan row: %v", err)
		}
		books = append(books, &book)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "row iteration error: %v", err)
	}

	return &pb.ListBooksResponse{
		Books:    books,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

//-----------------------------------------

func (s *bookCatalogServer) SearchBooks(ctx context.Context, req *pb.SearchBooksRequest) (*pb.SearchBooksResponse, error) {
	log.Printf("SearchBooks: query=%s, field=%s", req.Query, req.Field)

	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "search query required")
	}

	var sqlQuery string
	var args []interface{}
	searchPattern := "%" + req.Query + "%"

	switch req.Field {
	case "title":
		sqlQuery = "SELECT id, title, author, isbn, price, stock, published_year FROM books WHERE title LIKE ?"
		args = []interface{}{searchPattern}
	case "author":
		sqlQuery = "SELECT id, title, author, isbn, price, stock, published_year FROM books WHERE author LIKE ?"
		args = []interface{}{searchPattern}
	case "isbn":
		sqlQuery = "SELECT id, title, author, isbn, price, stock, published_year FROM books WHERE isbn = ?"
		args = []interface{}{req.Query}
	case "all", "":
		sqlQuery = `SELECT id, title, author, isbn, price, stock, published_year
		            FROM books
		            WHERE title LIKE ? OR author LIKE ? OR isbn LIKE ?`
		args = []interface{}{searchPattern, searchPattern, searchPattern}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid field: %s", req.Field)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search failed: %v", err)
	}
	defer rows.Close()

	var books []*pb.Book
	for rows.Next() {
		var book pb.Book
		if err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, &book.Price, &book.Stock, &book.PublishedYear); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan search result: %v", err)
		}
		books = append(books, &book)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "search rows error: %v", err)
	}

	return &pb.SearchBooksResponse{
		Books: books,
		Count: int32(len(books)),
		Query: req.Query,
	}, nil
}

func (s *bookCatalogServer) FilterBooks(ctx context.Context, req *pb.FilterBooksRequest) (*pb.FilterBooksResponse, error) {
	log.Printf("FilterBooks: price[%.2f-%.2f], year[%d-%d]",
		req.MinPrice, req.MaxPrice, req.MinYear, req.MaxYear)

	if req.MinPrice < 0 || req.MaxPrice < 0 {
		return nil, status.Error(codes.InvalidArgument, "price cannot be negative")
	}
	if req.MaxPrice > 0 && req.MinPrice > req.MaxPrice {
		return nil, status.Error(codes.InvalidArgument, "min_price cannot be greater than max_price")
	}
	if req.MinYear < 0 || req.MaxYear < 0 {
		return nil, status.Error(codes.InvalidArgument, "year cannot be negative")
	}
	if req.MaxYear > 0 && req.MinYear > req.MaxYear {
		return nil, status.Error(codes.InvalidArgument, "min_year cannot be greater than max_year")
	}

	query := "SELECT id, title, author, isbn, price, stock, published_year FROM books WHERE 1=1"
	var args []interface{}

	if req.MinPrice > 0 {
		query += " AND price >= ?"
		args = append(args, req.MinPrice)
	}
	if req.MaxPrice > 0 {
		query += " AND price <= ?"
		args = append(args, req.MaxPrice)
	}
	if req.MinYear > 0 {
		query += " AND published_year >= ?"
		args = append(args, req.MinYear)
	}
	if req.MaxYear > 0 {
		query += " AND published_year <= ?"
		args = append(args, req.MaxYear)
	}

	query += " ORDER BY id"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "filter failed: %v", err)
	}
	defer rows.Close()

	var books []*pb.Book
	for rows.Next() {
		var book pb.Book
		if err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, &book.Price, &book.Stock, &book.PublishedYear); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan filtered result: %v", err)
		}
		books = append(books, &book)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "filter rows error: %v", err)
	}

	return &pb.FilterBooksResponse{
		Books: books,
		Count: int32(len(books)),
	}, nil
}

func (s *bookCatalogServer) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	log.Println("GetStats called")

	var stats pb.GetStatsResponse

	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books").Scan(&stats.TotalBooks)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count books: %v", err)
	}

	err = s.db.QueryRowContext(ctx, "SELECT COALESCE(AVG(price), 0) FROM books").Scan(&stats.AveragePrice)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate average price: %v", err)
	}

	err = s.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(stock), 0) FROM books").Scan(&stats.TotalStock)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate total stock: %v", err)
	}

	err = s.db.QueryRowContext(ctx, "SELECT COALESCE(MIN(published_year), 0), COALESCE(MAX(published_year), 0) FROM books").
		Scan(&stats.EarliestYear, &stats.LatestYear)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate year range: %v", err)
	}

	return &stats, nil
}

//-----------------------------------------

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./books.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		isbn TEXT,
		price REAL NOT NULL,
		stock INTEGER DEFAULT 0,
		published_year INTEGER
	)
	`)
	if err != nil {
		return nil, err
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		sampleBooks := []struct {
			title  string
			author string
			isbn   string
			price  float32
			stock  int
			year   int
		}{
			{"The Go Programming Language", "Alan Donovan", "978-0134190440", 39.99, 15, 2015},
			{"Clean Code", "Robert Martin", "978-0132350884", 42.50, 20, 2008},
			{"Design Patterns", "Gang of Four", "978-0201633610", 54.99, 8, 1994},
			{"Learning Go", "Jon Bodner", "978-1492077213", 44.99, 10, 2021},
			{"Clean Architecture", "Robert Martin", "978-0134494166", 45.99, 12, 2017},
		}

		for _, b := range sampleBooks {
			_, err := db.Exec(
				"INSERT INTO books (title, author, isbn, price, stock, published_year) VALUES (?, ?, ?, ?, ?, ?)",
				b.title, b.author, b.isbn, b.price, b.stock, b.year,
			)
			if err != nil {
				return nil, err
			}
		}
		log.Println("Database seeded with sample books")
	}

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterBookCatalogServer(grpcServer, newServer(db))

	log.Println("BookCatalog gRPC server listening on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
'''
___
## Task 5: 
- build multi service gRPC system
- define author_service.proto
### author-service/main.go
'''
package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	authorpb "task5/proto/author"
	bookpb "task5/proto/book"
)

type authorCatalogServer struct {
	authorpb.UnimplementedAuthorCatalogServer
	db         *sql.DB
	bookClient bookpb.BookCatalogClient
}

func newServer(db *sql.DB, bookClient bookpb.BookCatalogClient) *authorCatalogServer {
	return &authorCatalogServer{
		db:         db,
		bookClient: bookClient,
	}
}

func (s *authorCatalogServer) GetAuthor(ctx context.Context, req *authorpb.GetAuthorRequest) (*authorpb.GetAuthorResponse, error) {
	var author authorpb.Author

	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, bio, birth_year, country FROM authors WHERE id = ?`, req.Id).
		Scan(&author.Id, &author.Name, &author.Bio, &author.BirthYear, &author.Country)

	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "author not found: id=%d", req.Id)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	return &authorpb.GetAuthorResponse{Author: &author}, nil
}

func (s *authorCatalogServer) CreateAuthor(ctx context.Context, req *authorpb.CreateAuthorRequest) (*authorpb.CreateAuthorResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO authors (name, bio, birth_year, country)
		 VALUES (?, ?, ?, ?)`,
		req.Name, req.Bio, req.BirthYear, req.Country,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create author: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get inserted id: %v", err)
	}

	author := &authorpb.Author{
		Id:        int32(id),
		Name:      req.Name,
		Bio:       req.Bio,
		BirthYear: req.BirthYear,
		Country:   req.Country,
	}

	return &authorpb.CreateAuthorResponse{Author: author}, nil
}

func (s *authorCatalogServer) ListAuthors(ctx context.Context, req *authorpb.ListAuthorsRequest) (*authorpb.ListAuthorsResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var total int32
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM authors`).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count authors: %v", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, bio, birth_year, country
		 FROM authors
		 LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list authors: %v", err)
	}
	defer rows.Close()

	var authors []*authorpb.Author
	for rows.Next() {
		var author authorpb.Author
		if err := rows.Scan(&author.Id, &author.Name, &author.Bio, &author.BirthYear, &author.Country); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan row: %v", err)
		}
		authors = append(authors, &author)
	}

	return &authorpb.ListAuthorsResponse{
		Authors: authors,
		Total:   total,
	}, nil
}

func (s *authorCatalogServer) GetAuthorBooks(ctx context.Context, req *authorpb.GetAuthorBooksRequest) (*authorpb.GetAuthorBooksResponse, error) {
	var author authorpb.Author

	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, bio, birth_year, country FROM authors WHERE id = ?`, req.AuthorId).
		Scan(&author.Id, &author.Name, &author.Bio, &author.BirthYear, &author.Country)

	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "author not found: id=%d", req.AuthorId)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	bookResp, err := s.bookClient.GetBooksByAuthor(ctx, &bookpb.GetBooksByAuthorRequest{
		AuthorId: req.AuthorId,
	})
	if err != nil {
		log.Printf("failed to get books from book service: %v", err)
		return &authorpb.GetAuthorBooksResponse{
			Author:    &author,
			Books:     nil,
			BookCount: 0,
		}, nil
	}

	var bookSummaries []*authorpb.BookSummary
	for _, book := range bookResp.Books {
		bookSummaries = append(bookSummaries, &authorpb.BookSummary{
			Id:            book.Id,
			Title:         book.Title,
			Price:         book.Price,
			PublishedYear: book.PublishedYear,
		})
	}

	return &authorpb.GetAuthorBooksResponse{
		Author:    &author,
		Books:     bookSummaries,
		BookCount: int32(len(bookSummaries)),
	}, nil
}

func connectToBookService() (bookpb.BookCatalogClient, error) {
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return bookpb.NewBookCatalogClient(conn), nil
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./authors.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS authors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		bio TEXT,
		birth_year INTEGER,
		country TEXT
	)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	bookClient, err := connectToBookService()
	if err != nil {
		log.Fatalf("failed to connect to book service: %v", err)
	}

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	authorpb.RegisterAuthorCatalogServer(grpcServer, newServer(db, bookClient))

	log.Println("Author Catalog gRPC server listening on :50052")
	log.Println("Connected to Book Catalog service on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
'''
### book-service/main.go
'''
package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "task5/proto/book"
)

type bookCatalogServer struct {
	pb.UnimplementedBookCatalogServer
	db *sql.DB
}

func newServer(db *sql.DB) *bookCatalogServer {
	return &bookCatalogServer{db: db}
}

func (s *bookCatalogServer) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	var book pb.Book

	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, author_id, isbn, price, stock, published_year
		 FROM books WHERE id = ?`, req.Id).
		Scan(&book.Id, &book.Title, &book.AuthorId, &book.Isbn, &book.Price, &book.Stock, &book.PublishedYear)

	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "book not found: id=%d", req.Id)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "database error: %v", err)
	}

	return &pb.GetBookResponse{Book: &book}, nil
}

func (s *bookCatalogServer) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.CreateBookResponse, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.AuthorId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "author_id must be positive")
	}
	if req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be positive")
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO books (title, author_id, isbn, price, stock, published_year)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		req.Title, req.AuthorId, req.Isbn, req.Price, req.Stock, req.PublishedYear,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create book: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get inserted id: %v", err)
	}

	book := &pb.Book{
		Id:            int32(id),
		Title:         req.Title,
		AuthorId:      req.AuthorId,
		Isbn:          req.Isbn,
		Price:         req.Price,
		Stock:         req.Stock,
		PublishedYear: req.PublishedYear,
	}

	return &pb.CreateBookResponse{Book: book}, nil
}

func (s *bookCatalogServer) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	var total int32
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM books`).Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count books: %v", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, author_id, isbn, price, stock, published_year
		 FROM books
		 LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list books: %v", err)
	}
	defer rows.Close()

	var books []*pb.Book
	for rows.Next() {
		var book pb.Book
		if err := rows.Scan(&book.Id, &book.Title, &book.AuthorId, &book.Isbn, &book.Price, &book.Stock, &book.PublishedYear); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan row: %v", err)
		}
		books = append(books, &book)
	}

	return &pb.ListBooksResponse{
		Books:    books,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *bookCatalogServer) GetBooksByAuthor(ctx context.Context, req *pb.GetBooksByAuthorRequest) (*pb.GetBooksByAuthorResponse, error) {
	if req.AuthorId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "author_id must be positive")
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, author_id, isbn, price, stock, published_year
		 FROM books WHERE author_id = ?`, req.AuthorId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query books by author: %v", err)
	}
	defer rows.Close()

	var books []*pb.Book
	for rows.Next() {
		var book pb.Book
		if err := rows.Scan(&book.Id, &book.Title, &book.AuthorId, &book.Isbn, &book.Price, &book.Stock, &book.PublishedYear); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan row: %v", err)
		}
		books = append(books, &book)
	}

	return &pb.GetBooksByAuthorResponse{
		Books: books,
		Count: int32(len(books)),
	}, nil
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./books.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author_id INTEGER NOT NULL,
		isbn TEXT,
		price REAL NOT NULL,
		stock INTEGER DEFAULT 0,
		published_year INTEGER
	)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("failed to initialize db: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterBookCatalogServer(grpcServer, newServer(db))

	log.Println("Book Catalog gRPC server listening on :50051")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
'''
### proto/author
'''
syntax = "proto3";

package authorservice;
option go_package = "task5-microservices/proto/author";

message Author {
  int32 id = 1;
  string name = 2;
  string bio = 3;
  int32 birth_year = 4;
  string country = 5;
}

message GetAuthorRequest {
  int32 id = 1;
}

message GetAuthorResponse {
  Author author = 1;
}

message CreateAuthorRequest {
  string name = 1;
  string bio = 2;
  int32 birth_year = 3;
  string country = 4;
}

message CreateAuthorResponse {
  Author author = 1;
}

message ListAuthorsRequest {
  int32 page = 1;
  int32 page_size = 2;
}

message ListAuthorsResponse {
  repeated Author authors = 1;
  int32 total = 2;
}

message GetAuthorBooksRequest {
  int32 author_id = 1;
}

message BookSummary {
  int32 id = 1;
  string title = 2;
  float price = 3;
  int32 published_year = 4;
}

message GetAuthorBooksResponse {
  Author author = 1;
  repeated BookSummary books = 2;
  int32 book_count = 3;
}

service AuthorCatalog {
  rpc GetAuthor(GetAuthorRequest) returns (GetAuthorResponse);
  rpc CreateAuthor(CreateAuthorRequest) returns (CreateAuthorResponse);
  rpc ListAuthors(ListAuthorsRequest) returns (ListAuthorsResponse);
  rpc GetAuthorBooks(GetAuthorBooksRequest) returns (GetAuthorBooksResponse);
}
'''
### proto/book
'''
syntax = "proto3";

package bookservice;
option go_package = "task5/proto/book";

message Book {
  int32 id = 1;
  string title = 2;
  int32 author_id = 3;
  string isbn = 4;
  float price = 5;
  int32 stock = 6;
  int32 published_year = 7;
}

message GetBookRequest {
  int32 id = 1;
}

message GetBookResponse {
  Book book = 1;
}

message CreateBookRequest {
  string title = 1;
  int32 author_id = 2;
  string isbn = 3;
  float price = 4;
  int32 stock = 5;
  int32 published_year = 6;
}

message CreateBookResponse {
  Book book = 1;
}

message ListBooksRequest {
  int32 page = 1;
  int32 page_size = 2;
}

message ListBooksResponse {
  repeated Book books = 1;
  int32 total = 2;
  int32 page = 3;
  int32 page_size = 4;
}

message GetBooksByAuthorRequest {
  int32 author_id = 1;
}

message GetBooksByAuthorResponse {
  repeated Book books = 1;
  int32 count = 2;
}

service BookCatalog {
  rpc GetBook(GetBookRequest) returns (GetBookResponse);
  rpc CreateBook(CreateBookRequest) returns (CreateBookResponse);
  rpc ListBooks(ListBooksRequest) returns (ListBooksResponse);
  rpc GetBooksByAuthor(GetBooksByAuthorRequest) returns (GetBooksByAuthorResponse);
}
'''
### client/main.go
'''
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	authorpb "task5/proto/author"
	bookpb "task5/proto/book"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	bookConn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer bookConn.Close()

	authorConn, err := grpc.Dial("localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer authorConn.Close()

	bookClient := bookpb.NewBookCatalogClient(bookConn)
	authorClient := authorpb.NewAuthorCatalogClient(authorConn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("=== Microservice Demo ===\n")

	fmt.Println("1. Creating author...")
	authorResp, err := authorClient.CreateAuthor(ctx, &authorpb.CreateAuthorRequest{
		Name:      "Martin Fowler",
		Bio:       "Software development expert",
		BirthYear: 1963,
		Country:   "UK",
	})
	if err != nil {
		log.Fatalf("failed to create author: %v", err)
	}

	fmt.Printf("Created author: %s (ID: %d)\n\n", authorResp.Author.Name, authorResp.Author.Id)

	fmt.Println("2. Creating books for author...")
	book1, err := bookClient.CreateBook(ctx, &bookpb.CreateBookRequest{
		Title:         "Refactoring",
		AuthorId:      authorResp.Author.Id,
		Isbn:          "978-0134757599",
		Price:         49.99,
		Stock:         15,
		PublishedYear: 2018,
	})
	if err != nil {
		log.Fatalf("failed to create first book: %v", err)
	}
	fmt.Printf("Created book: %s\n", book1.Book.Title)

	book2, err := bookClient.CreateBook(ctx, &bookpb.CreateBookRequest{
		Title:         "Patterns of Enterprise Application Architecture",
		AuthorId:      authorResp.Author.Id,
		Isbn:          "978-0321127426",
		Price:         54.99,
		Stock:         8,
		PublishedYear: 2002,
	})
	if err != nil {
		log.Fatalf("failed to create second book: %v", err)
	}
	fmt.Printf("Created book: %s\n\n", book2.Book.Title)

	fmt.Println("3. Fetching author's books (cross-service call)...")
	booksResp, err := authorClient.GetAuthorBooks(ctx, &authorpb.GetAuthorBooksRequest{
		AuthorId: authorResp.Author.Id,
	})
	if err != nil {
		log.Fatalf("failed to get author books: %v", err)
	}

	fmt.Printf("Author: %s\n", booksResp.Author.Name)
	fmt.Printf("Books written: %d\n", booksResp.BookCount)
	for i, book := range booksResp.Books {
		fmt.Printf("%d. %s (%d)\n", i+1, book.Title, book.PublishedYear)
	}

	fmt.Println("\nMicroservice demo completed!")
}
'''

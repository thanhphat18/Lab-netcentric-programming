package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"strings"
	
	pb "bookCatalogSev/proto"
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

// GetBook retrieves a book by ID
func (s *bookCatalogServer) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	log.Printf("GetBook called: id=%d", req.Id)
	
	var book pb.Book
	err := s.db.QueryRowContext(ctx,
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

// CreateBook adds a new book
func (s *bookCatalogServer) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.CreateBookResponse, error) {
	log.Printf("CreateBook called: title=%s", req.Title)
	
	// Validate input
	if req.Title == "" || req.Author == "" {
		return nil, status.Error(codes.InvalidArgument, "title and author are required")
	}
	if req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "price must be positive")
	}
	
	// Insert into database
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO books (title, author, isbn, price, stock, published_year) VALUES (?, ?, ?, ?, ?, ?)",
		req.Title, req.Author, req.Isbn, req.Price, req.Stock, req.PublishedYear,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create book: %v", err)
	}
	
	// Get generated ID
	id, _ := result.LastInsertId()
	
	// Return created book
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

// UpdateBook modifies an existing book
func (s *bookCatalogServer) UpdateBook(ctx context.Context, req *pb.UpdateBookRequest) (*pb.UpdateBookResponse, error) {
	log.Printf("UpdateBook called: id=%d", req.Id)
	
	// Validate input
	if req.Title == "" || req.Author == "" {
		return nil, status.Error(codes.InvalidArgument, "title and author are required")
	}
	
	// Update in database
	result, err := s.db.ExecContext(ctx,
		"UPDATE books SET title=?, author=?, isbn=?, price=?, stock=?, published_year=? WHERE id=?",
		req.Title, req.Author, req.Isbn, req.Price, req.Stock, req.PublishedYear, req.Id,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update book: %v", err)
	}
	
	// Check if book exists
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "book not found: id=%d", req.Id)
	}
	
	// Return updated book
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

// DeleteBook removes a book
func (s *bookCatalogServer) DeleteBook(ctx context.Context, req *pb.DeleteBookRequest) (*pb.DeleteBookResponse, error) {
	log.Printf("DeleteBook called: id=%d", req.Id)
	
	result, err := s.db.ExecContext(ctx, "DELETE FROM books WHERE id = ?", req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete book: %v", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "book not found: id=%d", req.Id)
	}
	
	return &pb.DeleteBookResponse{
		Success: true,
		Message: fmt.Sprintf("Book %d deleted successfully", req.Id),
	}, nil
}

// ListBooks returns paginated books
func (s *bookCatalogServer) ListBooks(ctx context.Context, req *pb.ListBooksRequest) (*pb.ListBooksResponse, error) {
	log.Printf("ListBooks called: page=%d, pageSize=%d", req.Page, req.PageSize)
	
	// Default pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	offset := (page - 1) * pageSize
	
	// Get total count
	var total int32
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM books").Scan(&total)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count books: %v", err)
	}
	
	// Get books
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, title, author, isbn, price, stock, published_year FROM books LIMIT ? OFFSET ?",
		pageSize, offset,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list books: %v", err)
	}
	defer rows.Close()
	
	var books []*pb.Book
	for rows.Next() {
		var book pb.Book
		err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, 
			&book.Price, &book.Stock, &book.PublishedYear)
		if err != nil {
			continue
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

// SearchBooks finds books by title or author
func (s *bookCatalogServer) SearchBooks(ctx context.Context, req *pb.SearchBooksRequest) (*pb.SearchBooksResponse, error) {
	log.Printf("SearchBooks called: query=%s", req.Query)
	
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "search query required")
	}
	
	searchPattern := "%" + strings.ToLower(req.Query) + "%"
	
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, title, author, isbn, price, stock, published_year FROM books WHERE LOWER(title) LIKE ? OR LOWER(author) LIKE ?",
		searchPattern, searchPattern,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search failed: %v", err)
	}
	defer rows.Close()
	
	var books []*pb.Book
	for rows.Next() {
		var book pb.Book
		err := rows.Scan(&book.Id, &book.Title, &book.Author, &book.Isbn, 
			&book.Price, &book.Stock, &book.PublishedYear)
		if err != nil {
			continue
		}
		books = append(books, &book)
	}
	
	return &pb.SearchBooksResponse{
		Books: books,
		Count: int32(len(books)),
	}, nil
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./books.db")
	if err != nil {
		return nil, err
	}
	
	// Create table
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
	
	// Seed data
	var count int
	db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
	if count == 0 {
		books := []struct {
			title, author, isbn string
			price               float32
			stock, year         int
		}{
			{"The Go Programming Language", "Alan Donovan", "978-0134190440", 39.99, 15, 2015},
			{"Clean Code", "Robert Martin", "978-0132350884", 42.50, 20, 2008},
			{"Design Patterns", "Gang of Four", "978-0201633610", 54.99, 8, 1994},
		}
		
		for _, b := range books {
			db.Exec("INSERT INTO books (title, author, isbn, price, stock, published_year) VALUES (?, ?, ?, ?, ?, ?)",
				b.title, b.author, b.isbn, b.price, b.stock, b.year)
		}
		log.Println("✓ Database seeded with sample books")
	}
	
	return db, nil
}

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	
	// Create listener
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	// Create gRPC server
	grpcServer := grpc.NewServer()
	
	// Register service
	pb.RegisterBookCatalogServer(grpcServer, newServer(db))
	
	log.Println("🚀 BookCatalog gRPC server listening on :50051")
	
	// Start serving
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
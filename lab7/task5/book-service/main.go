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
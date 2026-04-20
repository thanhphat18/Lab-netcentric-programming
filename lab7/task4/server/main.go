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
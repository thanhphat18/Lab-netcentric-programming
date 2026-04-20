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
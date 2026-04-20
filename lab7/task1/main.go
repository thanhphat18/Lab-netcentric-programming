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

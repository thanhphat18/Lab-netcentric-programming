package main

import (
	"fmt"
	"log"
	
	"google.golang.org/protobuf/proto"
	pb "protocolBuffers/proto"  // Import generated code
)

func main() {
	// Create a Person message
	person := &pb.Person{
		Name:    "John Doe",
		Age:     30,
		Email:   "john@example.com",
		Hobbies: []string{"reading", "gaming", "coding"},
	}
	
	// Print the message
	fmt.Printf("Person: %v\n", person)
	fmt.Printf("Name: %s\n", person.Name)
	fmt.Printf("Age: %d\n", person.Age)
	fmt.Printf("Hobbies: %v\n", person.Hobbies)
	
	// Serialize to bytes (for network transmission)
	data, err := proto.Marshal(person)
	if err != nil {
		log.Fatal("Marshaling error:", err)
	}
	fmt.Printf("\nSerialized size: %d bytes\n", len(data))
	
	// Deserialize from bytes
	newPerson := &pb.Person{}
	err = proto.Unmarshal(data, newPerson)
	if err != nil {
		log.Fatal("Unmarshaling error:", err)
	}
	
	fmt.Printf("\nDeserialized person: %v\n", newPerson)
	
	// Create nested message
	personWithAddr := &pb.PersonWithAddress{
		Person: person,
		Address: &pb.Address{
			Street:  "123 Main St",
			City:    "New York",
			Country: "USA",
			ZipCode: 10001,
		},
	}
	
	fmt.Printf("\nPerson with address: %v\n", personWithAddr)
	
	// Using enum
	book := &pb.Book{
		Id:        1,
		Title:     "1984",
		Author:    "George Orwell",
		Genre:     pb.BookGenre_FICTION,
		Price:     19.99,
		Available: true,
	}
	
	fmt.Printf("\nBook: %v\n", book)
	fmt.Printf("Genre: %s\n", book.Genre)
}
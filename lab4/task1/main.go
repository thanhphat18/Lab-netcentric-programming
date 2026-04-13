package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)
// Website page → download HTML → parse HTML into a tree → find each book block → extract data from each block → store in []Book → print summary → save as JSON.

type Book struct {
	Title        string `json:"title"`
	Price        string `json:"price"`
	Rating       string `json:"rating"`
	Availability string `json:"availability"`
	ImageURL     string `json:"image_url"`
}

func main() {
	url := "http://books.toscrape.com/catalogue/page-1.html"

	fmt.Println("Scraping books from:", url)

	books, err := scrapeBooks(url)
	if err != nil {
		fmt.Println("Error scraping books:", err)
		return
	}

	fmt.Printf("Found %d books\n\n", len(books))

	// Print all books
	for i, book := range books {
		fmt.Printf("Book %d:\n", i+1)
		fmt.Println(" Title:", book.Title)
		fmt.Println(" Price:", book.Price)
		fmt.Println(" Rating:", book.Rating)
		fmt.Println(" Availability:", book.Availability)
		fmt.Println(" Image URL:", book.ImageURL)
		fmt.Println()
	}

	minPrice, maxPrice, avgPrice := calculatePriceStats(books)

	fmt.Println("Summary:")
	fmt.Println(" Total books:", len(books))
	fmt.Printf(" Price range: £%.2f - £%.2f\n", minPrice, maxPrice)
	fmt.Printf(" Average price: £%.2f\n", avgPrice)

	err = saveBooksToJSON(books, "books.json")
	if err != nil {
		fmt.Println("Error saving JSON:", err)
		return
	}

	fmt.Printf("Saved %d books to books.json\n", len(books))
}

func scrapeBooks(url string) ([]Book, error) {
	//send http request to the page and get the response
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// convert web into tree structure
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var books []Book

	// find and extract data from each book block
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "article" && hasClass(n, "product_pod") {
			book := extractBookData(n)
			books = append(books, book)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	return books, nil
}

func extractBookData(n *html.Node) Book {
	book := Book{}

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode {
			// Title: <h3><a title="...">
			if node.Data == "a" {
				if title, ok := getAttr(node, "title"); ok && book.Title == "" {
					book.Title = strings.TrimSpace(title)
				}
			}

			// Image: <img src="...">
			if node.Data == "img" {
				if src, ok := getAttr(node, "src"); ok && book.ImageURL == "" {
					book.ImageURL = "http://books.toscrape.com/" + strings.TrimPrefix(src, "../")
				}
			}

			// Price: <p class="price_color">£51.77</p>
			if node.Data == "p" && hasClass(node, "price_color") {
				book.Price = strings.TrimSpace(getTextContent(node))
			}

			// Availability: <p class="availability">In stock</p>
			if node.Data == "p" && hasClass(node, "availability") {
				book.Availability = cleanWhitespace(getTextContent(node))
			}

			// Rating: <p class="star-rating Three">
			if node.Data == "p" && hasClass(node, "star-rating") {
				book.Rating = extractRating(node)
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)

	return book
}

func saveBooksToJSON(books []Book, filename string) error {
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func hasClass(n *html.Node, className string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, className) {
			return true
		}
	}
	return false
}

func getAttr(n *html.Node, key string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}
	return "", false
}

func getTextContent(n *html.Node) string {
	var builder strings.Builder

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			builder.WriteString(node.Data)
			builder.WriteString(" ")
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return builder.String()
}

func cleanWhitespace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func extractRating(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			parts := strings.Fields(attr.Val)
			for _, part := range parts {
				if part != "star-rating" {
					return part
				}
			}
		}
	}
	return ""
}

func parsePrice(price string) float64 {
	cleaned := strings.ReplaceAll(price, "£", "")
	value, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0
	}
	return value
}

func calculatePriceStats(books []Book) (float64, float64, float64) {
	if len(books) == 0 {
		return 0, 0, 0
	}

	minPrice := parsePrice(books[0].Price)
	maxPrice := minPrice
	total := 0.0

	for _, book := range books {
		price := parsePrice(book.Price)
		if price < minPrice {
			minPrice = price
		}
		if price > maxPrice {
			maxPrice = price
		}
		total += price
	}

	avgPrice := total / float64(len(books))
	return minPrice, maxPrice, avgPrice
}

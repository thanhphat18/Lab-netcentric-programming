package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Book struct {
	Title        string `json:"title"`
	Price        string `json:"price"`
	Rating       string `json:"rating"`
	Availability string `json:"availability"`
	ImageURL     string `json:"image_url"`
}

type ScraperStats struct {
	PagesScraped int
	BooksFound   int
	Errors       int
	StartTime    time.Time
	EndTime      time.Time
}

func main() {
	baseURL := "http://books.toscrape.com/catalogue/page-1.html"
	maxPages := 5

	fmt.Printf("Starting paginated scraper...\n")
	fmt.Printf("Max pages: %d\n\n", maxPages)

	books, stats, err := scrapePaginatedBooks(baseURL, maxPages)
	if err != nil {
		fmt.Printf("Fatal error: %v\n", err)
		return
	}

	printStats(stats)

	err = saveBooksToJSON(books, "paginated_books.json")
	if err != nil {
		fmt.Printf("Error saving JSON: %v\n", err)
		return
	}

	fmt.Printf("Saved %d books to paginated_books.json\n", len(books))
}

func scrapePaginatedBooks(baseURL string, maxPages int) ([]Book, *ScraperStats, error) {
	stats := &ScraperStats{
		StartTime: time.Now(),
	}

	// keep track of all books across pages
	var allBooks []Book
	currentURL := baseURL

	//loop through pages 
	for page := 1; page <= maxPages; page++ {
		books, doc, err := scrapeBooksWithDoc(currentURL)
		if err != nil {
			stats.Errors++
			fmt.Printf("Error scraping page %d: %v\n", page, err)

			// simple retry once
			time.Sleep(1 * time.Second)
			books, doc, err = scrapeBooksWithDoc(currentURL)
			if err != nil {
				stats.Errors++
				fmt.Printf("Retry failed for page %d: %v\n", page, err)
				break
			}

			_ = doc
		}

		fmt.Printf("Scraping page %d/%d... Found %d books\n", page, maxPages, len(books))

		allBooks = append(allBooks, books...)
		stats.PagesScraped++
		stats.BooksFound += len(books)

		// find next page
		nextURL, hasNext := getNextPageURL(doc, currentURL)
		if !hasNext {
			break
		}

		currentURL = nextURL

		// rate limit: 1 request per second
		time.Sleep(1 * time.Second)
	}

	stats.EndTime = time.Now()
	return allBooks, stats, nil
}

func scrapeBooksWithDoc(pageURL string) ([]Book, *html.Node, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var books []Book

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "article" && hasClass(n, "product_pod") {
			book := extractBookData(n)
			books = append(books, book)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return books, doc, nil
}

func extractBookData(n *html.Node) Book {
	book := Book{}

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "a" {
				if title, ok := getAttr(node, "title"); ok && book.Title == "" {
					book.Title = strings.TrimSpace(title)
				}
			}

			if node.Data == "img" {
				if src, ok := getAttr(node, "src"); ok && book.ImageURL == "" {
					book.ImageURL = "http://books.toscrape.com/" + strings.TrimPrefix(src, "../")
				}
			}

			if node.Data == "p" && hasClass(node, "price_color") {
				book.Price = strings.TrimSpace(getTextContent(node))
			}

			if node.Data == "p" && hasClass(node, "availability") {
				book.Availability = cleanWhitespace(getTextContent(node))
			}

			if node.Data == "p" && hasClass(node, "star-rating") {
				book.Rating = extractRating(node)
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)
	return book
}

func getNextPageURL(doc *html.Node, currentURL string) (string, bool) {
	var nextHref string

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if nextHref != "" {
			return
		}

		// Look for: <li class="next"><a href="page-2.html">
		if n.Type == html.ElementNode && n.Data == "li" && hasClass(n, "next") {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "a" {
					if href, ok := getAttr(c, "href"); ok {
						nextHref = href
						return
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)

	if nextHref == "" {
		return "", false
	}

	base, err := url.Parse(currentURL)
	if err != nil {
		return "", false
	}

	rel, err := url.Parse(nextHref)
	if err != nil {
		return "", false
	}

	return base.ResolveReference(rel).String(), true
}

func printStats(stats *ScraperStats) {
	duration := stats.EndTime.Sub(stats.StartTime).Seconds()

	fmt.Println("\n=== Scraping Statistics ===")
	fmt.Printf("Pages scraped: %d\n", stats.PagesScraped)
	fmt.Printf("Total books found: %d\n", stats.BooksFound)
	fmt.Printf("Errors: %d\n", stats.Errors)
	fmt.Printf("Duration: %.1f seconds\n", duration)

	if stats.PagesScraped > 0 {
		avg := float64(stats.BooksFound) / float64(stats.PagesScraped)
		fmt.Printf("Average books per page: %.1f\n", avg)
	}
}

func saveBooksToJSON(books []Book, filename string) error {
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
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

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			builder.WriteString(node.Data)
			builder.WriteString(" ")
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)
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

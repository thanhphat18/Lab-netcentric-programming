---
title: Lab report template
tags: [template]
---

# Net-Centric Programming Lab 01 Report
-- *Student 1: Chau Thanh Phat. ID: ITITIU21135*
___
**Content**
Learn web scraping, HTTP APIs, and data collection techniques in Go to build robust data collection systems.
___

## Task 1.1:
- If you want to mention anything
- Include a link if you want: [Google](https://www.google.com)

'''package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

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
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var books []Book

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

```


## Task 2.1: 

<!-- Copy your source code here for Task 2 -->
<!-- Remember to keep the ```  -->

```package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const TMDBBaseURL = "https://api.themoviedb.org/3"
const TMDBImageBaseURL = "https://image.tmdb.org/t/p/w500"

// Raw search response from TMDB
type TMDBSearchResponse struct {
	Page         int               `json:"page"`
	Results      []TMDBMovieResult `json:"results"`
	TotalResults int               `json:"total_results"`
}

// One movie item inside search results
type TMDBMovieResult struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	GenreIDs    []int   `json:"genre_ids"`
	PosterPath  string  `json:"poster_path"`
}

// Genre list response
type TMDBGenreResponse struct {
	Genres []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
}

// Movie details response
type TMDBMovieDetailsResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	PosterPath  string  `json:"poster_path"`
	Genres      []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
}

// Simplified movie struct required by lab
type Movie struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Overview    string   `json:"overview"`
	ReleaseDate string   `json:"release_date"`
	Rating      float64  `json:"rating"`
	Genres      []string `json:"genres"`
	PosterURL   string   `json:"poster_url"`
}

type TMDBClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	GenreMap   map[int]string
}

func NewTMDBClient(apiKey string) *TMDBClient {
	return &TMDBClient{
		APIKey:  apiKey,
		BaseURL: TMDBBaseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		GenreMap: make(map[int]string),
	}
}

func (c *TMDBClient) loadGenres() error {
	endpoint := fmt.Sprintf("%s/genre/movie/list?api_key=%s", c.BaseURL, url.QueryEscape(c.APIKey))

	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch genres: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("genre request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var genreResp TMDBGenreResponse
	err = json.NewDecoder(resp.Body).Decode(&genreResp)
	if err != nil {
		return fmt.Errorf("failed to decode genres JSON: %w", err)
	}

	for _, genre := range genreResp.Genres {
		c.GenreMap[genre.ID] = genre.Name
	}

	return nil
}

func (c *TMDBClient) searchMovies(query string) ([]Movie, error) {
	escapedQuery := url.QueryEscape(query)
	endpoint := fmt.Sprintf("%s/search/movie?api_key=%s&query=%s", c.BaseURL, url.QueryEscape(c.APIKey), escapedQuery)

	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var searchResp TMDBSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode search JSON: %w", err)
	}

	var movies []Movie
	for _, result := range searchResp.Results {
		movie := Movie{
			ID:          result.ID,
			Title:       result.Title,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			Rating:      result.VoteAverage,
			Genres:      c.convertGenreIDs(result.GenreIDs),
			PosterURL:   buildPosterURL(result.PosterPath),
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (c *TMDBClient) getMovieDetails(id int) (*Movie, error) {
	endpoint := fmt.Sprintf("%s/movie/%d?api_key=%s", c.BaseURL, id, url.QueryEscape(c.APIKey))

	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("details request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var details TMDBMovieDetailsResponse
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		return nil, fmt.Errorf("failed to decode details JSON: %w", err)
	}

	var genres []string
	for _, g := range details.Genres {
		genres = append(genres, g.Name)
	}

	movie := &Movie{
		ID:          details.ID,
		Title:       details.Title,
		Overview:    details.Overview,
		ReleaseDate: details.ReleaseDate,
		Rating:      details.VoteAverage,
		Genres:      genres,
		PosterURL:   buildPosterURL(details.PosterPath),
	}

	return movie, nil
}

func (c *TMDBClient) convertGenreIDs(ids []int) []string {
	var genres []string
	for _, id := range ids {
		if name, ok := c.GenreMap[id]; ok {
			genres = append(genres, name)
		}
	}
	return genres
}

func buildPosterURL(posterPath string) string {
	if posterPath == "" {
		return ""
	}
	return TMDBImageBaseURL + posterPath
}

func saveMoviesToJSON(movies []Movie, filename string) error {
	data, err := json.MarshalIndent(movies, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal movies JSON: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

func main() {
	apiKey := "515e8dee0de90b76191f7b2023a8f75d" // Replace with your TMDB API key

	client := NewTMDBClient(apiKey)

	fmt.Println("Loading movie genres...")
	err := client.loadGenres()
	if err != nil {
		fmt.Printf("Error loading genres: %v\n", err)
		return
	}
	fmt.Printf("Loaded %d genres\n", len(client.GenreMap))

	query := "inception"
	fmt.Printf("Searching for: %s\n", query)

	movies, err := client.searchMovies(query)
	if err != nil {
		fmt.Printf("Error searching movies: %v\n", err)
		return
	}

	fmt.Printf("Found %d movies\n", len(movies))

	for i, movie := range movies {
		fmt.Printf("\nMovie %d:\n", i+1)
		fmt.Printf(" ID: %d\n", movie.ID)
		fmt.Printf(" Title: %s\n", movie.Title)
		fmt.Printf(" Release Date: %s\n", movie.ReleaseDate)
		fmt.Printf(" Rating: %.1f/10\n", movie.Rating)
		fmt.Printf(" Genres: %v\n", movie.Genres)

		overview := movie.Overview
		if len(overview) > 60 {
			overview = overview[:60] + "..."
		}
		fmt.Printf(" Overview: %s\n", overview)
	}

	err = saveMoviesToJSON(movies, "tmdb_results.json")
	if err != nil {
		fmt.Printf("Error saving JSON: %v\n", err)
		return
	}

	fmt.Printf("\nSaved %d movies to tmdb_results.json\n", len(movies))
}

```
## Task 2.2:
'''
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

const TMDBBaseURL = "https://api.themoviedb.org/3"
const TMDBImageBaseURL = "https://image.tmdb.org/t/p/w500"

// =========================
// Part A: TMDB client
// =========================

type TMDBSearchResponse struct {
	Page         int               `json:"page"`
	Results      []TMDBMovieResult `json:"results"`
	TotalResults int               `json:"total_results"`
}

type TMDBMovieResult struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	GenreIDs    []int   `json:"genre_ids"`
	PosterPath  string  `json:"poster_path"`
}

type TMDBGenreResponse struct {
	Genres []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
}

type TMDBMovieDetailsResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	PosterPath  string  `json:"poster_path"`
	Genres      []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
}

type Movie struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Overview    string   `json:"overview"`
	ReleaseDate string   `json:"release_date"`
	Rating      float64  `json:"rating"`
	Genres      []string `json:"genres"`
	PosterURL   string   `json:"poster_url"`
}

type TMDBClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	GenreMap   map[int]string
}

func NewTMDBClient(apiKey string) *TMDBClient {
	return &TMDBClient{
		APIKey:  apiKey,
		BaseURL: TMDBBaseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		GenreMap: make(map[int]string),
	}
}

func (c *TMDBClient) loadGenres() error {
	endpoint := fmt.Sprintf("%s/genre/movie/list?api_key=%s", c.BaseURL, url.QueryEscape(c.APIKey))

	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch genres: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("genre request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var genreResp TMDBGenreResponse
	if err := json.NewDecoder(resp.Body).Decode(&genreResp); err != nil {
		return fmt.Errorf("failed to decode genres JSON: %w", err)
	}

	for _, genre := range genreResp.Genres {
		c.GenreMap[genre.ID] = genre.Name
	}

	return nil
}

func (c *TMDBClient) searchMovies(query string) ([]Movie, error) {
	escapedQuery := url.QueryEscape(query)
	endpoint := fmt.Sprintf("%s/search/movie?api_key=%s&query=%s", c.BaseURL, url.QueryEscape(c.APIKey), escapedQuery)

	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var searchResp TMDBSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search JSON: %w", err)
	}

	var movies []Movie
	for _, result := range searchResp.Results {
		movie := Movie{
			ID:          result.ID,
			Title:       result.Title,
			Overview:    result.Overview,
			ReleaseDate: result.ReleaseDate,
			Rating:      result.VoteAverage,
			Genres:      c.convertGenreIDs(result.GenreIDs),
			PosterURL:   buildPosterURL(result.PosterPath),
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (c *TMDBClient) getMovieDetails(id int) (*Movie, error) {
	endpoint := fmt.Sprintf("%s/movie/%d?api_key=%s", c.BaseURL, id, url.QueryEscape(c.APIKey))

	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("details request failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var details TMDBMovieDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode details JSON: %w", err)
	}

	var genres []string
	for _, g := range details.Genres {
		genres = append(genres, g.Name)
	}

	movie := &Movie{
		ID:          details.ID,
		Title:       details.Title,
		Overview:    details.Overview,
		ReleaseDate: details.ReleaseDate,
		Rating:      details.VoteAverage,
		Genres:      genres,
		PosterURL:   buildPosterURL(details.PosterPath),
	}

	return movie, nil
}

func (c *TMDBClient) convertGenreIDs(ids []int) []string {
	var genres []string
	for _, id := range ids {
		if name, ok := c.GenreMap[id]; ok {
			genres = append(genres, name)
		}
	}
	return genres
}

func buildPosterURL(posterPath string) string {
	if posterPath == "" {
		return ""
	}
	return TMDBImageBaseURL + posterPath
}

// =========================
// Part B: Task 2.2 structs
// =========================

type MovieInfo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Director    string   `json:"director,omitempty"`
	Year        int      `json:"year"`
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
	Rating      float64  `json:"rating"`
	Duration    int      `json:"duration_minutes,omitempty"`
	Source      string   `json:"source"`
	LastUpdated string   `json:"last_updated"`
}

type MovieSource interface {
	GetMovies(query string, limit int) ([]MovieInfo, error)
	GetName() string
}

// =========================
// Part C: Sources
// =========================

type TMDBSource struct {
	client *TMDBClient
}

func NewTMDBSource(apiKey string) *TMDBSource {
	return &TMDBSource{
		client: NewTMDBClient(apiKey),
	}
}

func (t *TMDBSource) GetName() string {
	return "TMDB"
}

func (t *TMDBSource) GetMovies(query string, limit int) ([]MovieInfo, error) {
	if len(t.client.GenreMap) == 0 {
		if err := t.client.loadGenres(); err != nil {
			return nil, fmt.Errorf("failed to load TMDB genres: %w", err)
		}
	}

	movies, err := t.client.searchMovies(query)
	if err != nil {
		return nil, err
	}

	var results []MovieInfo
	for i, movie := range movies {
		if i >= limit {
			break
		}

		results = append(results, MovieInfo{
			ID:          fmt.Sprintf("%d", movie.ID),
			Title:       movie.Title,
			Year:        extractYear(movie.ReleaseDate),
			Description: movie.Overview,
			Genres:      movie.Genres,
			Rating:      movie.Rating,
			Source:      "TMDB",
			LastUpdated: time.Now().Format(time.RFC3339),
		})
	}

	return results, nil
}

type MockScraperSource struct {
	name string
}

func NewMockScraperSource(name string) *MockScraperSource {
	return &MockScraperSource{name: name}
}

func (m *MockScraperSource) GetName() string {
	return m.name
}

func (m *MockScraperSource) GetMovies(query string, limit int) ([]MovieInfo, error) {
	allMovies := []MovieInfo{
		{
			ID:          "mock-1",
			Title:       "Spider-Man",
			Director:    "Sam Raimi",
			Year:        2002,
			Description: "Peter Parker becomes Spider-Man after being bitten by a genetically altered spider.",
			Genres:      []string{"Action", "Fantasy"},
			Rating:      7.3,
			Duration:    121,
			Source:      m.name,
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		{
			ID:          "mock-2",
			Title:       "Spider-Man: Homecoming",
			Director:    "Jon Watts",
			Year:        2017,
			Description: "Peter Parker balances high school life with being Spider-Man.",
			Genres:      []string{"Action", "Adventure", "Science Fiction"},
			Rating:      7.4,
			Duration:    133,
			Source:      m.name,
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		{
			ID:          "mock-3",
			Title:       "The Amazing Spider-Man",
			Director:    "Marc Webb",
			Year:        2012,
			Description: "A new take on Peter Parker's origin story.",
			Genres:      []string{"Action", "Adventure", "Fantasy"},
			Rating:      6.9,
			Duration:    136,
			Source:      m.name,
			LastUpdated: time.Now().Format(time.RFC3339),
		},
		{
			ID:          "mock-4",
			Title:       "Spider-Man 2",
			Director:    "Sam Raimi",
			Year:        2004,
			Description: "Peter Parker struggles to balance life and hero duties.",
			Genres:      []string{"Action", "Adventure"},
			Rating:      7.8,
			Duration:    127,
			Source:      m.name,
			LastUpdated: time.Now().Format(time.RFC3339),
		},
	}

	query = strings.ToLower(strings.TrimSpace(query))
	var results []MovieInfo

	for _, movie := range allMovies {
		if strings.Contains(strings.ToLower(movie.Title), query) {
			results = append(results, movie)
		}
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// =========================
// Part D: Aggregator
// =========================

type MovieAggregator struct {
	sources []MovieSource
}

func NewMovieAggregator(sources ...MovieSource) *MovieAggregator {
	return &MovieAggregator{sources: sources}
}

func (a *MovieAggregator) Search(query string, limit int) ([]MovieInfo, error) {
	return aggregateMovies(a.sources, query, limit)
}

func aggregateMovies(sources []MovieSource, query string, limit int) ([]MovieInfo, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	var allMovies []MovieInfo
	var errors []string

	for _, source := range sources {
		wg.Add(1)

		go func(src MovieSource) {
			defer wg.Done()

			movies, err := src.GetMovies(query, limit)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Sprintf("%s: %v", src.GetName(), err))
				mu.Unlock()
				return
			}

			fmt.Printf("Querying %s... Found %d results\n", src.GetName(), len(movies))

			mu.Lock()
			allMovies = append(allMovies, movies...)
			mu.Unlock()
		}(source)
	}

	wg.Wait()

	if len(allMovies) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("all sources failed: %s", strings.Join(errors, "; "))
	}

	uniqueMovies := deduplicateMovies(allMovies)

	sort.Slice(uniqueMovies, func(i, j int) bool {
		if uniqueMovies[i].Rating == uniqueMovies[j].Rating {
			return uniqueMovies[i].Title < uniqueMovies[j].Title
		}
		return uniqueMovies[i].Rating > uniqueMovies[j].Rating
	})

	return uniqueMovies, nil
}

func deduplicateMovies(movies []MovieInfo) []MovieInfo {
	var unique []MovieInfo

	for _, movie := range movies {
		duplicateIndex := -1

		for i, existing := range unique {
			similarity := calculateSimilarity(movie.Title, existing.Title)
			if similarity >= 0.8 && (movie.Year == 0 || existing.Year == 0 || movie.Year == existing.Year) {
				duplicateIndex = i
				break
			}
		}

		if duplicateIndex == -1 {
			unique = append(unique, movie)
		} else {
			unique[duplicateIndex] = mergeMovies(unique[duplicateIndex], movie)
		}
	}

	return unique
}

func calculateSimilarity(title1, title2 string) float64 {
	t1 := strings.ToLower(strings.TrimSpace(title1))
	t2 := strings.ToLower(strings.TrimSpace(title2))

	if t1 == t2 {
		return 1.0
	}

	if strings.Contains(t1, t2) || strings.Contains(t2, t1) {
		return 0.8
	}

	return 0.0
}

func mergeMovies(a, b MovieInfo) MovieInfo {
	merged := a

	if b.Rating > merged.Rating {
		merged.Rating = b.Rating
	}
	if merged.Description == "" && b.Description != "" {
		merged.Description = b.Description
	}
	if merged.Director == "" && b.Director != "" {
		merged.Director = b.Director
	}
	if merged.Year == 0 && b.Year != 0 {
		merged.Year = b.Year
	}
	if merged.Duration == 0 && b.Duration != 0 {
		merged.Duration = b.Duration
	}

	merged.Genres = mergeGenres(merged.Genres, b.Genres)

	if !strings.Contains(merged.Source, b.Source) {
		merged.Source = merged.Source + ", " + b.Source
	}

	merged.LastUpdated = time.Now().Format(time.RFC3339)
	return merged
}

func mergeGenres(a, b []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, genre := range a {
		g := strings.TrimSpace(genre)
		if g != "" && !seen[g] {
			seen[g] = true
			result = append(result, g)
		}
	}

	for _, genre := range b {
		g := strings.TrimSpace(genre)
		if g != "" && !seen[g] {
			seen[g] = true
			result = append(result, g)
		}
	}

	sort.Strings(result)
	return result
}

// =========================
// Part E: Reporting + JSON
// =========================

func generateReport(movies []MovieInfo) {
	fmt.Println("\n=== Movie Aggregation Report ===")
	fmt.Printf("Total movies found: %d\n\n", len(movies))

	sourceCount := make(map[string]int)
	for _, movie := range movies {
		parts := strings.Split(movie.Source, ",")
		for _, part := range parts {
			source := strings.TrimSpace(part)
			if source != "" {
				sourceCount[source]++
			}
		}
	}

	fmt.Println("Movies by Source:")
	var sourceNames []string
	for source := range sourceCount {
		sourceNames = append(sourceNames, source)
	}
	sort.Strings(sourceNames)
	for _, source := range sourceNames {
		fmt.Printf(" - %s: %d movies\n", source, sourceCount[source])
	}

	genreCount := make(map[string]int)
	for _, movie := range movies {
		for _, genre := range movie.Genres {
			genreCount[genre]++
		}
	}

	type GenreStat struct {
		Name  string
		Count int
	}

	var genreStats []GenreStat
	for genre, count := range genreCount {
		genreStats = append(genreStats, GenreStat{Name: genre, Count: count})
	}

	sort.Slice(genreStats, func(i, j int) bool {
		if genreStats[i].Count == genreStats[j].Count {
			return genreStats[i].Name < genreStats[j].Name
		}
		return genreStats[i].Count > genreStats[j].Count
	})

	fmt.Println("\nTop Genres:")
	for i, stat := range genreStats {
		if i >= 5 {
			break
		}
		fmt.Printf(" - %s: %d\n", stat.Name, stat.Count)
	}

	var totalRating float64
	ratedCount := 0
	for _, movie := range movies {
		if movie.Rating > 0 {
			totalRating += movie.Rating
			ratedCount++
		}
	}

	if ratedCount > 0 {
		avgRating := totalRating / float64(ratedCount)
		fmt.Printf("\nAverage Rating: %.2f/10\n", avgRating)
	}
}

func saveToJSON(movies []MovieInfo, filename string) error {
	data, err := json.MarshalIndent(movies, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	fmt.Printf("\nSaved %d movies to %s\n", len(movies), filename)
	return nil
}

// =========================
// Part F: Helpers
// =========================

func extractYear(dateStr string) int {
	if len(dateStr) >= 4 {
		var year int
		fmt.Sscanf(dateStr[:4], "%d", &year)
		return year
	}
	return 0
}

// =========================
// Part G: main
// =========================

func main() {
	apiKey := "515e8dee0de90b76191f7b2023a8f75d" // Replace with your real TMDB key

	aggregator := NewMovieAggregator(
		NewTMDBSource(apiKey),
		NewMockScraperSource("MovieScraper"),
	)

	query := "spider-man"

	fmt.Println("=== Multi-Source Movie Aggregator ===")
	fmt.Printf("Searching for: %s\n\n", query)

	movies, err := aggregator.Search(query, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nFound %d movies after deduplication\n", len(movies))

	for i, movie := range movies {
		if i < 3 {
			fmt.Printf("\n%d. %s (%d)\n", i+1, movie.Title, movie.Year)
			fmt.Printf(" Source: %s\n", movie.Source)
			fmt.Printf(" Rating: %.1f/10\n", movie.Rating)
			fmt.Printf(" Genres: %v\n", movie.Genres)
		}
	}

	generateReport(movies)

	if err := saveToJSON(movies, "aggregated_movies.json"); err != nil {
		fmt.Printf("Error saving JSON: %v\n", err)
	}
}

'''

## Task 3.1:
'''
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

	var allBooks []Book
	currentURL := baseURL

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

'''
## Task 4.1:
'''
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const TMDBBaseURL = "https://api.themoviedb.org/3"

// =========================
// Models
// =========================

type TMDBSearchResponse struct {
	Page    int `json:"page"`
	Results []struct {
		ID          int     `json:"id"`
		Title       string  `json:"title"`
		Overview    string  `json:"overview"`
		ReleaseDate string  `json:"release_date"`
		VoteAverage float64 `json:"vote_average"`
		GenreIDs    []int   `json:"genre_ids"`
		PosterPath  string  `json:"poster_path"`
	} `json:"results"`
	TotalResults int `json:"total_results"`
}

type TMDBGenreResponse struct {
	Genres []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
}

type Movie struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Overview    string   `json:"overview"`
	ReleaseDate string   `json:"release_date"`
	Rating      float64  `json:"rating"`
	Genres      []string `json:"genres"`
	PosterURL   string   `json:"poster_url"`
}

type MovieInfo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Director    string   `json:"director,omitempty"`
	Year        int      `json:"year"`
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
	Rating      float64  `json:"rating"`
	Duration    int      `json:"duration_minutes,omitempty"`
	Source      string   `json:"source"`
	LastUpdated string   `json:"last_updated"`
}

type MovieDatabase struct {
	Movies      map[string]MovieInfo `json:"movies"`
	Genres      map[string][]string  `json:"genres"`
	Directors   map[string][]string  `json:"directors"`
	Years       map[int][]string     `json:"years"`
	LastUpdated time.Time            `json:"last_updated"`
	TotalCount  int                  `json:"total_count"`
}

// =========================
// TMDB Client
// =========================

type TMDBClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	GenreMap   map[int]string
}

func NewTMDBClient(apiKey string) *TMDBClient {
	return &TMDBClient{
		APIKey:  apiKey,
		BaseURL: TMDBBaseURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		GenreMap: make(map[int]string),
	}
}

func (c *TMDBClient) loadGenres() error {
	endpoint := fmt.Sprintf("%s/genre/movie/list?api_key=%s", c.BaseURL, url.QueryEscape(c.APIKey))

	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to fetch genres: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("genre API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var genreResp TMDBGenreResponse
	if err := json.NewDecoder(resp.Body).Decode(&genreResp); err != nil {
		return fmt.Errorf("failed to parse genre response: %w", err)
	}

	for _, g := range genreResp.Genres {
		c.GenreMap[g.ID] = g.Name
	}

	return nil
}

func (c *TMDBClient) searchMovies(query string) ([]Movie, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query cannot be empty")
	}

	endpoint := fmt.Sprintf(
		"%s/search/movie?api_key=%s&query=%s",
		c.BaseURL,
		url.QueryEscape(c.APIKey),
		url.QueryEscape(query),
	)

	resp, err := c.HTTPClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to search movies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var searchResp TMDBSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	var movies []Movie
	for _, item := range searchResp.Results {
		var genreNames []string
		for _, id := range item.GenreIDs {
			if name, ok := c.GenreMap[id]; ok {
				genreNames = append(genreNames, name)
			}
		}

		posterURL := ""
		if item.PosterPath != "" {
			posterURL = "https://image.tmdb.org/t/p/w500" + item.PosterPath
		}

		movies = append(movies, Movie{
			ID:          item.ID,
			Title:       item.Title,
			Overview:    item.Overview,
			ReleaseDate: item.ReleaseDate,
			Rating:      item.VoteAverage,
			Genres:      genreNames,
			PosterURL:   posterURL,
		})
	}

	return movies, nil
}

// =========================
// Movie Database
// =========================

func NewMovieDatabase() *MovieDatabase {
	return &MovieDatabase{
		Movies:    make(map[string]MovieInfo),
		Genres:    make(map[string][]string),
		Directors: make(map[string][]string),
		Years:     make(map[int][]string),
	}
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

func (db *MovieDatabase) Add(movie MovieInfo) error {
	if movie.ID == "" {
		return errors.New("movie ID cannot be empty")
	}

	if _, exists := db.Movies[movie.ID]; exists {
		return fmt.Errorf("movie already exists: %s", movie.ID)
	}

	db.Movies[movie.ID] = movie

	for _, genre := range movie.Genres {
		if !containsString(db.Genres[genre], movie.ID) {
			db.Genres[genre] = append(db.Genres[genre], movie.ID)
		}
	}

	if movie.Director != "" {
		if !containsString(db.Directors[movie.Director], movie.ID) {
			db.Directors[movie.Director] = append(db.Directors[movie.Director], movie.ID)
		}
	}

	if movie.Year > 0 {
		if !containsString(db.Years[movie.Year], movie.ID) {
			db.Years[movie.Year] = append(db.Years[movie.Year], movie.ID)
		}
	}

	db.TotalCount++
	return nil
}

func (db *MovieDatabase) Get(id string) (*MovieInfo, error) {
	movie, exists := db.Movies[id]
	if !exists {
		return nil, fmt.Errorf("movie not found: %s", id)
	}
	return &movie, nil
}

func (db *MovieDatabase) Search(query string) ([]MovieInfo, error) {
	var results []MovieInfo
	query = strings.ToLower(strings.TrimSpace(query))

	for _, movie := range db.Movies {
		if strings.Contains(strings.ToLower(movie.Title), query) {
			results = append(results, movie)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Rating > results[j].Rating
	})

	return results, nil
}

func (db *MovieDatabase) GetByGenre(genre string) ([]MovieInfo, error) {
	var results []MovieInfo
	movieIDs, exists := db.Genres[genre]
	if !exists {
		return results, nil
	}

	for _, id := range movieIDs {
		if movie, err := db.Get(id); err == nil {
			results = append(results, *movie)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Rating > results[j].Rating
	})

	return results, nil
}

func (db *MovieDatabase) GetByYear(year int) ([]MovieInfo, error) {
	var results []MovieInfo
	movieIDs, exists := db.Years[year]
	if !exists {
		return results, nil
	}

	for _, id := range movieIDs {
		if movie, err := db.Get(id); err == nil {
			results = append(results, *movie)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Rating > results[j].Rating
	})

	return results, nil
}

func (db *MovieDatabase) GetByDirector(director string) ([]MovieInfo, error) {
	var results []MovieInfo
	movieIDs, exists := db.Directors[director]
	if !exists {
		return results, nil
	}

	for _, id := range movieIDs {
		if movie, err := db.Get(id); err == nil {
			results = append(results, *movie)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Rating > results[j].Rating
	})

	return results, nil
}

func (db *MovieDatabase) removeIDFromIndex(index map[string][]string, key, id string) {
	ids := index[key]
	var updated []string
	for _, existingID := range ids {
		if existingID != id {
			updated = append(updated, existingID)
		}
	}
	if len(updated) == 0 {
		delete(index, key)
	} else {
		index[key] = updated
	}
}

func (db *MovieDatabase) removeIDFromYearIndex(year int, id string) {
	ids := db.Years[year]
	var updated []string
	for _, existingID := range ids {
		if existingID != id {
			updated = append(updated, existingID)
		}
	}
	if len(updated) == 0 {
		delete(db.Years, year)
	} else {
		db.Years[year] = updated
	}
}

func (db *MovieDatabase) Delete(id string) error {
	movie, exists := db.Movies[id]
	if !exists {
		return fmt.Errorf("movie not found: %s", id)
	}

	for _, genre := range movie.Genres {
		db.removeIDFromIndex(db.Genres, genre, id)
	}

	if movie.Director != "" {
		db.removeIDFromIndex(db.Directors, movie.Director, id)
	}

	if movie.Year > 0 {
		db.removeIDFromYearIndex(movie.Year, id)
	}

	delete(db.Movies, id)
	db.TotalCount--
	return nil
}

func (db *MovieDatabase) Update(movie MovieInfo) error {
	if movie.ID == "" {
		return errors.New("movie ID cannot be empty")
	}

	if _, exists := db.Movies[movie.ID]; !exists {
		return fmt.Errorf("movie not found: %s", movie.ID)
	}

	if err := db.Delete(movie.ID); err != nil {
		return err
	}

	return db.Add(movie)
}

func (db *MovieDatabase) Save(filename string) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write database file: %w", err)
	}

	return nil
}

func (db *MovieDatabase) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read database file: %w", err)
	}

	if err := json.Unmarshal(data, db); err != nil {
		return fmt.Errorf("failed to unmarshal database: %w", err)
	}

	return nil
}

func (db *MovieDatabase) PrintStatistics() {
	fmt.Println("\n=== Movie Database Statistics ===")
	fmt.Printf("Total Movies: %d\n", db.TotalCount)
	fmt.Printf("Total Genres: %d\n", len(db.Genres))
	fmt.Printf("Total Directors: %d\n", len(db.Directors))
	fmt.Printf("Year Range: %d entries\n\n", len(db.Years))

	fmt.Println("Movies by Genre:")
	genreKeys := make([]string, 0, len(db.Genres))
	for genre := range db.Genres {
		genreKeys = append(genreKeys, genre)
	}
	sort.Strings(genreKeys)

	for _, genre := range genreKeys {
		fmt.Printf(" - %s: %d movies\n", genre, len(db.Genres[genre]))
	}

	// Rating distribution
	var low, medium, high int
	for _, movie := range db.Movies {
		switch {
		case movie.Rating < 5:
			low++
		case movie.Rating < 7:
			medium++
		default:
			high++
		}
	}

	fmt.Println("\nRating Distribution:")
	fmt.Printf(" - Low (<5): %d\n", low)
	fmt.Printf(" - Medium (5-6.9): %d\n", medium)
	fmt.Printf(" - High (>=7): %d\n", high)

	// Decade distribution
	decades := make(map[int]int)
	for _, movie := range db.Movies {
		if movie.Year > 0 {
			decade := (movie.Year / 10) * 10
			decades[decade]++
		}
	}

	var decadeKeys []int
	for d := range decades {
		decadeKeys = append(decadeKeys, d)
	}
	sort.Ints(decadeKeys)

	fmt.Println("\nMovies by Decade:")
	for _, d := range decadeKeys {
		fmt.Printf(" - %ds: %d\n", d, decades[d])
	}
}

// =========================
// Collection Pipeline
// =========================

func extractYear(dateStr string) int {
	if len(dateStr) >= 4 {
		var year int
		fmt.Sscanf(dateStr[:4], "%d", &year)
		return year
	}
	return 0
}

func buildMovieDatabase(apiKey string) (*MovieDatabase, error) {
	db := NewMovieDatabase()
	client := NewTMDBClient(apiKey)

	fmt.Println("Loading movie genres from TMDB...")
	if err := client.loadGenres(); err != nil {
		return nil, fmt.Errorf("failed to load genres: %w", err)
	}
	fmt.Printf("Loaded %d genres\n", len(client.GenreMap))

	searchQueries := []string{
		"action", "comedy", "drama", "horror",
		"sci-fi", "romance", "animation", "thriller",
		"marvel", "star wars", "batman", "classic",
	}

	fmt.Println("\nBuilding movie database...")
	for i, query := range searchQueries {
		fmt.Printf("[%d/%d] Searching for: %s\n", i+1, len(searchQueries), query)

		movies, err := client.searchMovies(query)
		if err != nil {
			fmt.Printf(" Error: %v\n", err)
			continue
		}

		added := 0
		for _, movie := range movies {
			movieInfo := MovieInfo{
				ID:          fmt.Sprintf("%d", movie.ID),
				Title:       movie.Title,
				Year:        extractYear(movie.ReleaseDate),
				Description: movie.Overview,
				Genres:      movie.Genres,
				Rating:      movie.Rating,
				Source:      "TMDB",
				LastUpdated: time.Now().Format(time.RFC3339),
			}

			if _, exists := db.Movies[movieInfo.ID]; !exists {
				if err := db.Add(movieInfo); err == nil {
					added++
				}
			}
		}

		fmt.Printf(" Added %d new movies (found %d total)\n", added, len(movies))

		// Rate limiting
		time.Sleep(1 * time.Second)
	}

	db.LastUpdated = time.Now()
	fmt.Println("\nDatabase building complete!")
	return db, nil
}

// =========================
// Main
// =========================

func main() {
	fmt.Println("=== Movie Database Builder ===\n")

	apiKey := "515e8dee0de90b76191f7b2023a8f75d"
	if apiKey == "" || apiKey == "YOUR_TMDB_API_KEY" {
		fmt.Println("Please replace YOUR_TMDB_API_KEY with your real TMDB API key.")
		return
	}

	fmt.Println("Starting database collection...")
	db, err := buildMovieDatabase(apiKey)
	if err != nil {
		fmt.Printf("Error building database: %v\n", err)
		return
	}

	db.PrintStatistics()

	fmt.Println("\n=== Testing Search Functions ===")

	fmt.Println("\nSearching for 'spider':")
	results, _ := db.Search("spider")
	for i, movie := range results {
		if i < 3 {
			fmt.Printf(" %d. %s (%d) - Rating: %.1f\n",
				i+1, movie.Title, movie.Year, movie.Rating)
		}
	}

	if len(db.Genres) > 0 {
		var firstGenre string
		for genre := range db.Genres {
			firstGenre = genre
			break
		}

		fmt.Printf("\nMovies in genre '%s':\n", firstGenre)
		genreMovies, _ := db.GetByGenre(firstGenre)
		for i, movie := range genreMovies {
			if i < 3 {
				fmt.Printf(" %d. %s (%d)\n", i+1, movie.Title, movie.Year)
			}
		}
	}

	filename := "movie_database.json"
	if err := db.Save(filename); err != nil {
		fmt.Printf("Error saving database: %v\n", err)
		return
	}

	fileInfo, err := os.Stat(filename)
	if err != nil {
		fmt.Printf("Database saved, but failed to read file info: %v\n", err)
		return
	}

	sizeKB := fileInfo.Size() / 1024
	fmt.Printf("\n✓ Database saved successfully!\n")
	fmt.Printf(" File: %s\n", filename)
	fmt.Printf(" Size: %d KB\n", sizeKB)
	fmt.Printf(" Movies: %d\n", db.TotalCount)
}

'''
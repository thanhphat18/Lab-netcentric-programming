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

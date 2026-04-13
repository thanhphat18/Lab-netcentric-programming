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

	fmt.Println("\nSearching for 'batman':")
	results, _ := db.Search("batman")
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

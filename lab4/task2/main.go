package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

//TMDB website/API → Go sends request → TMDB returns JSON → Go converts JSON into structs → Go prints and saves data

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

	// converts JSON from response to Go struct
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

	query := "Harry Potter"
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

package models

// Game represents a board game from the user's wishlist
type Game struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

// StoreResult represents the availability result from a single store
type StoreResult struct {
	Store    string  `json:"store"`
	Found    bool    `json:"found"`
	InStock  bool    `json:"inStock"`
	Price    string  `json:"price"`
	PriceNum float64 `json:"priceNum"`
	URL      string  `json:"url"`
	Title    string  `json:"title"`
	Error    string  `json:"error,omitempty"`
}

// GameResult represents all store results for a single game
type GameResult struct {
	Name    string        `json:"name"`
	Results []StoreResult `json:"results"`
}

// CheckRequest is the API request format
type CheckRequest struct {
	Games []Game `json:"games"`
}

// CheckResponse is the API response format
type CheckResponse struct {
	Results []GameResult   `json:"results"`
	Summary map[string]int `json:"summary"`
}

package stores

import (
	"net/http"
	"time"

	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/shopify"
)

// Store represents a board game store with checking capabilities
type Store interface {
	Name() string
	Check(gameName string) models.StoreResult
}

// HTTPClient is the shared HTTP client for all stores
var HTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

// ShopifyClient is the shared Shopify API client
var ShopifyClient = &shopify.Client{
	HTTPClient: HTTPClient,
}

// GetAllStores returns all available store implementations
func GetAllStores() []Store {
	return []Store{
		NewBoardGameBliss(),
		NewGames401(),
		NewGreatBoardGames(),
	}
}

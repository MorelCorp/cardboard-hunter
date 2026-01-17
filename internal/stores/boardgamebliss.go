package stores

import (
	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/shopify"
)

// BoardGameBliss represents the Board Game Bliss store
type BoardGameBliss struct {
	name    string
	baseURL string
}

// NewBoardGameBliss creates a new Board Game Bliss store checker
func NewBoardGameBliss() *BoardGameBliss {
	return &BoardGameBliss{
		name:    "Board Game Bliss",
		baseURL: "https://www.boardgamebliss.com",
	}
}

// Name returns the store name
func (s *BoardGameBliss) Name() string {
	return s.name
}

// Check searches for a game at Board Game Bliss
func (s *BoardGameBliss) Check(gameName string) models.StoreResult {
	products, err := ShopifyClient.Search(s.baseURL, gameName)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}

	matches := shopify.FindMatches(gameName, products, s.baseURL, 5)
	return shopify.BuildStoreResult(s.name, matches, gameName)
}

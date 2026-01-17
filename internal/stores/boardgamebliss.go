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
	result := models.StoreResult{
		Store: s.name,
	}

	products, err := ShopifyClient.Search(s.baseURL, gameName)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	matchedResult, found := shopify.FindBestMatch(gameName, products, s.baseURL)
	if found {
		result = matchedResult
		result.Store = s.name
	}

	return result
}

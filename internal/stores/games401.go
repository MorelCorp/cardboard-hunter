package stores

import (
	"strings"

	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/utils"
)

// Games401 represents the 401 Games store
type Games401 struct {
	name    string
	baseURL string
}

// NewGames401 creates a new 401 Games store checker
func NewGames401() *Games401 {
	return &Games401{
		name:    "401 Games",
		baseURL: "https://store.401games.ca",
	}
}

// Name returns the store name
func (s *Games401) Name() string {
	return s.name
}

// Check searches for a game at 401 Games
func (s *Games401) Check(gameName string) models.StoreResult {
	result := models.StoreResult{
		Store: s.name,
	}

	products, err := ShopifyClient.Search(s.baseURL, gameName)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if len(products) == 0 {
		return result
	}

	// Find best match (filtering out non-board-game products like TCG singles)
	for _, product := range products {
		if isTCGProduct(product.Title) {
			continue
		}

		if utils.FuzzyMatch(gameName, product.Title) {
			result.Found = true
			result.Title = product.Title
			result.URL = s.baseURL + product.URL
			result.InStock = product.Available
			result.Price = product.Price
			result.PriceNum = utils.ParsePrice(product.Price)
			return result
		}
	}

	// Fallback to first non-TCG result if nothing matched
	for _, product := range products {
		if isTCGProduct(product.Title) {
			continue
		}
		result.Found = true
		result.Title = product.Title
		result.URL = s.baseURL + product.URL
		result.InStock = product.Available
		result.Price = product.Price
		result.PriceNum = utils.ParsePrice(product.Price)
		return result
	}

	return result
}

// isTCGProduct checks if a product is a TCG single, sleeve, or booster
func isTCGProduct(title string) bool {
	titleLower := strings.ToLower(title)
	return strings.Contains(titleLower, "sleeve") ||
		strings.Contains(titleLower, "single") ||
		strings.Contains(titleLower, "booster")
}

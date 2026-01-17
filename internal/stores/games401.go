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
	products, err := ShopifyClient.Search(s.baseURL, gameName)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}

	var matches []models.ProductMatch
	for _, product := range products {
		if isTCGProduct(product.Title) || utils.ShouldExclude(product.Title) {
			continue
		}
		if utils.FuzzyMatch(gameName, product.Title) {
			matches = append(matches, models.ProductMatch{
				Title:    product.Title,
				URL:      s.baseURL + product.URL,
				Price:    product.Price,
				PriceNum: utils.ParsePrice(product.Price),
				InStock:  product.Available,
			})
			if len(matches) >= 5 {
				break
			}
		}
	}

	if len(matches) == 0 {
		return models.StoreResult{Store: s.name}
	}

	// If exact title match exists, use only that
	for _, m := range matches {
		if utils.ExactTitleMatch(gameName, m.Title) {
			matches = []models.ProductMatch{m}
			break
		}
	}

	first := matches[0]
	return models.StoreResult{
		Store:    s.name,
		Found:    true,
		Title:    first.Title,
		URL:      first.URL,
		Price:    first.Price,
		PriceNum: first.PriceNum,
		InStock:  first.InStock,
		Matches:  matches,
	}
}

// isTCGProduct checks if a product is a TCG single, sleeve, or booster
func isTCGProduct(title string) bool {
	titleLower := strings.ToLower(title)
	return strings.Contains(titleLower, "sleeve") ||
		strings.Contains(titleLower, "single") ||
		strings.Contains(titleLower, "booster")
}

package stores

import (
	"strings"

	"cardboard-hunter/internal/config"
	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/shopify"
	"cardboard-hunter/internal/utils"
)

// ShopifyChecker implements checking for Shopify-based stores
type ShopifyChecker struct {
	cfg *config.StoreConfig
}

// NewShopifyChecker creates a new Shopify checker from config
func NewShopifyChecker(cfg *config.StoreConfig) *ShopifyChecker {
	return &ShopifyChecker{cfg: cfg}
}

func (c *ShopifyChecker) Check(gameName string) models.StoreResult {
	products, err := ShopifyClient.Search(c.cfg.BaseURL, gameName)
	if err != nil {
		return models.StoreResult{Store: c.cfg.Name, Error: err.Error()}
	}

	var excludes []string
	if c.cfg.Shopify != nil {
		excludes = c.cfg.Shopify.ExcludePatterns
	}

	var matches []models.ProductMatch
	for _, p := range products {
		if shouldExcludeByPatterns(p.Title, excludes) || utils.ShouldExclude(p.Title) {
			continue
		}
		if utils.FuzzyMatch(gameName, p.Title) {
			matches = append(matches, models.ProductMatch{
				Title:    p.Title,
				URL:      c.cfg.BaseURL + p.URL,
				Price:    p.Price,
				PriceNum: utils.ParsePrice(p.Price),
				InStock:  p.Available,
			})
			if len(matches) >= 5 {
				break
			}
		}
	}

	return shopify.BuildStoreResult(c.cfg.Name, matches, gameName)
}

func shouldExcludeByPatterns(title string, patterns []string) bool {
	t := strings.ToLower(title)
	for _, p := range patterns {
		if strings.Contains(t, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

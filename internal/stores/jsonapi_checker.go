package stores

import (
	"encoding/json"
	"io"
	"net/url"
	"strings"

	"cardboard-hunter/internal/config"
	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/utils"
)

// JSONAPIChecker implements checking for JSON API stores
type JSONAPIChecker struct {
	cfg *config.StoreConfig
}

// NewJSONAPIChecker creates a new JSON API checker from config
func NewJSONAPIChecker(cfg *config.StoreConfig) *JSONAPIChecker {
	return &JSONAPIChecker{cfg: cfg}
}

func (c *JSONAPIChecker) Check(gameName string) models.StoreResult {
	if c.cfg.JSONAPI == nil {
		return models.StoreResult{Store: c.cfg.Name, Error: "no jsonApi config"}
	}

	searchURL := c.cfg.BaseURL + strings.Replace(
		c.cfg.JSONAPI.SearchPath, "{query}", url.QueryEscape(gameName), 1)

	resp, err := HTTPClient.Get(searchURL)
	if err != nil {
		return models.StoreResult{Store: c.cfg.Name, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.StoreResult{Store: c.cfg.Name, Error: err.Error()}
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return models.StoreResult{Store: c.cfg.Name, Error: err.Error()}
	}

	products := c.extractProducts(data)
	fields := c.cfg.JSONAPI.Fields

	var matches []models.ProductMatch
	for _, p := range products {
		title := getString(p, fields.Title)
		if utils.ShouldExclude(title) || !utils.FuzzyMatch(gameName, title) {
			continue
		}

		price := strings.TrimSpace(getString(p, fields.Price))
		matches = append(matches, models.ProductMatch{
			Title:    title,
			URL:      getString(p, fields.URL),
			Price:    price,
			PriceNum: utils.ParsePrice(price),
			InStock:  c.determineStock(p),
		})

		if len(matches) >= 5 {
			break
		}
	}

	return buildResult(c.cfg.Name, matches, gameName)
}

func (c *JSONAPIChecker) extractProducts(data map[string]any) []map[string]any {
	path := c.cfg.JSONAPI.ProductsPath
	val, ok := data[path]
	if !ok {
		return nil
	}

	arr, ok := val.([]any)
	if !ok {
		return nil
	}

	var products []map[string]any
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			products = append(products, m)
		}
	}
	return products
}

func (c *JSONAPIChecker) determineStock(p map[string]any) bool {
	fields := c.cfg.JSONAPI.Fields

	if fields.Quantity != "" {
		if qty := getNumber(p, fields.Quantity); qty > 0 {
			return true
		}
	}

	if fields.StockStatus != "" && c.cfg.JSONAPI.InStockValue != "" {
		status := getString(p, fields.StockStatus)
		if status == c.cfg.JSONAPI.InStockValue {
			return true
		}
	}

	return fields.Quantity == "" && fields.StockStatus == ""
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getNumber(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return n
		case int:
			return float64(n)
		}
	}
	return 0
}

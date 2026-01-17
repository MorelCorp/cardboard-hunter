package shopify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/utils"
)

// SearchResponse represents the Shopify search API response
type SearchResponse struct {
	Resources struct {
		Results struct {
			Products []Product `json:"products"`
		} `json:"results"`
	} `json:"resources"`
}

// Product represents a Shopify product in search results
type Product struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Price     string `json:"price"`
	Available bool   `json:"available"`
}

// Client handles Shopify API requests
type Client struct {
	HTTPClient *http.Client
}

// Search performs a product search on a Shopify store
func (c *Client) Search(baseURL, gameName string) ([]Product, error) {
	searchURL := fmt.Sprintf(
		"%s/search/suggest.json?q=%s&resources[type]=product&resources[limit]=10",
		baseURL,
		url.QueryEscape(gameName),
	)

	resp, err := c.HTTPClient.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data SearchResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data.Resources.Results.Products, nil
}

// FindMatches finds all matching products from search results (up to limit)
func FindMatches(gameName string, products []Product, baseURL string, limit int) []models.ProductMatch {
	var matches []models.ProductMatch
	for _, product := range products {
		if utils.ShouldExclude(product.Title) {
			continue
		}
		if utils.FuzzyMatch(gameName, product.Title) {
			matches = append(matches, models.ProductMatch{
				Title:    product.Title,
				URL:      baseURL + product.URL,
				Price:    product.Price,
				PriceNum: utils.ParsePrice(product.Price),
				InStock:  product.Available,
			})
			if len(matches) >= limit {
				break
			}
		}
	}
	return matches
}

// BuildStoreResult creates a StoreResult from matches
// If an exact title match exists, returns only that match
func BuildStoreResult(storeName string, matches []models.ProductMatch, gameName string) models.StoreResult {
	if len(matches) == 0 {
		return models.StoreResult{Store: storeName}
	}

	// Check for exact title match - if found, use only that
	for _, m := range matches {
		if utils.ExactTitleMatch(gameName, m.Title) {
			matches = []models.ProductMatch{m}
			break
		}
	}

	first := matches[0]
	return models.StoreResult{
		Store:    storeName,
		Found:    true,
		Title:    first.Title,
		URL:      first.URL,
		Price:    first.Price,
		PriceNum: first.PriceNum,
		InStock:  first.InStock,
		Matches:  matches,
	}
}

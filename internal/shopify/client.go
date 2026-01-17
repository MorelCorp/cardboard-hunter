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

// FindBestMatch finds the best matching product from search results
func FindBestMatch(gameName string, products []Product, baseURL string) (models.StoreResult, bool) {
	var result models.StoreResult

	if len(products) == 0 {
		return result, false
	}

	// Find best match using fuzzy matching
	for _, product := range products {
		if utils.FuzzyMatch(gameName, product.Title) {
			result.Found = true
			result.Title = product.Title
			result.URL = baseURL + product.URL
			result.InStock = product.Available
			result.Price = product.Price
			result.PriceNum = utils.ParsePrice(product.Price)
			return result, true
		}
	}

	// Fallback to first result if nothing matched but we have results
	if len(products) > 0 {
		product := products[0]
		result.Found = true
		result.Title = product.Title
		result.URL = baseURL + product.URL
		result.InStock = product.Available
		result.Price = product.Price
		result.PriceNum = utils.ParsePrice(product.Price)
		return result, true
	}

	return result, false
}

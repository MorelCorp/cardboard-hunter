package stores

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/utils"
)

type BoardGamesNMore struct {
	name    string
	baseURL string
}

type journal3Response struct {
	Products []struct {
		Name        string `json:"name"`
		Price       string `json:"price"`
		Href        string `json:"href"`
		Quantity    int    `json:"quantity"`
		StockStatus string `json:"stock_status"`
	} `json:"products"`
}

func NewBoardGamesNMore() *BoardGamesNMore {
	return &BoardGamesNMore{
		name:    "Board Games N More",
		baseURL: "https://www.boardgamesnmore.com",
	}
}

func (s *BoardGamesNMore) Name() string { return s.name }

func (s *BoardGamesNMore) Check(gameName string) models.StoreResult {
	searchURL := fmt.Sprintf("%s/index.php?route=journal3/search&search=%s",
		s.baseURL, url.QueryEscape(gameName))

	resp, err := HTTPClient.Get(searchURL)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}

	var data journal3Response
	if err := json.Unmarshal(body, &data); err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}

	var matches []models.ProductMatch
	for _, p := range data.Products {
		if utils.ShouldExclude(p.Name) || !utils.FuzzyMatch(gameName, p.Name) {
			continue
		}

		// Clean price string (remove HTML if any)
		price := strings.TrimSpace(p.Price)

		matches = append(matches, models.ProductMatch{
			Title:    p.Name,
			URL:      p.Href,
			Price:    price,
			PriceNum: utils.ParsePrice(price),
			InStock:  p.Quantity > 0 || p.StockStatus == "In Stock",
		})
		if len(matches) >= 5 {
			break
		}
	}

	if len(matches) == 0 {
		return models.StoreResult{Store: s.name}
	}

	// Prioritize exact match
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

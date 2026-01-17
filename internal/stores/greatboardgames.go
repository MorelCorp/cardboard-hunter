package stores

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/utils"
)

type GreatBoardGames struct {
	name    string
	baseURL string
}

func NewGreatBoardGames() *GreatBoardGames {
	return &GreatBoardGames{
		name:    "Great Board Games",
		baseURL: "https://www.greatboardgames.ca",
	}
}

func (s *GreatBoardGames) Name() string {
	return s.name
}

func (s *GreatBoardGames) Check(gameName string) models.StoreResult {
	result := models.StoreResult{Store: s.name}

	searchURL := fmt.Sprintf("%s/search?q=%s", s.baseURL, url.QueryEscape(gameName))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	html := string(body)

	// Split HTML by product-card divs
	cardSplitter := regexp.MustCompile(`<div class="product-card`)
	titlePattern := regexp.MustCompile(`<a href="(https://www\.greatboardgames\.ca/games/[^"]+)"[^>]*class="text-dark"[^>]*>([^<]+)</a>`)
	pricePattern := regexp.MustCompile(`<span[^>]*>\$([0-9]+\.[0-9]{2})</span>`)

	parts := cardSplitter.Split(html, -1)
	for _, cardHTML := range parts[1:] { // Skip first part (before any card)
		titleMatch := titlePattern.FindStringSubmatch(cardHTML)
		if titleMatch == nil {
			continue
		}

		title := strings.TrimSpace(titleMatch[2])
		if !utils.FuzzyMatch(gameName, title) {
			continue
		}

		result.Found = true
		result.Title = title
		result.URL = titleMatch[1]
		result.InStock = !strings.Contains(cardHTML, "Out of Stock")

		if priceMatch := pricePattern.FindStringSubmatch(cardHTML); priceMatch != nil {
			result.Price = "$" + priceMatch[1]
			result.PriceNum = utils.ParsePrice(priceMatch[1])
		}
		return result
	}

	return result
}

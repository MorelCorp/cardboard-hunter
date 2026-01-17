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
	searchURL := fmt.Sprintf("%s/search?q=%s", s.baseURL, url.QueryEscape(gameName))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}

	html := string(body)

	cardSplitter := regexp.MustCompile(`<div class="product-card`)
	titlePattern := regexp.MustCompile(`<a href="(https://www\.greatboardgames\.ca/games/[^"]+)"[^>]*class="text-dark"[^>]*>([^<]+)</a>`)
	pricePattern := regexp.MustCompile(`<span[^>]*>\$([0-9]+\.[0-9]{2})</span>`)

	var matches []models.ProductMatch
	parts := cardSplitter.Split(html, -1)
	for _, cardHTML := range parts[1:] {
		titleMatch := titlePattern.FindStringSubmatch(cardHTML)
		if titleMatch == nil {
			continue
		}

		title := strings.TrimSpace(titleMatch[2])
		if utils.ShouldExclude(title) || !utils.FuzzyMatch(gameName, title) {
			continue
		}

		match := models.ProductMatch{
			Title:   title,
			URL:     titleMatch[1],
			InStock: !strings.Contains(cardHTML, "Out of Stock"),
		}
		if priceMatch := pricePattern.FindStringSubmatch(cardHTML); priceMatch != nil {
			match.Price = "$" + priceMatch[1]
			match.PriceNum = utils.ParsePrice(priceMatch[1])
		}
		matches = append(matches, match)
		if len(matches) >= 5 {
			break
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

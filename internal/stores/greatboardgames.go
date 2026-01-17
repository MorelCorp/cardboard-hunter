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

// GreatBoardGames represents the Great Board Games store
type GreatBoardGames struct {
	name    string
	baseURL string
}

// NewGreatBoardGames creates a new Great Board Games store checker
func NewGreatBoardGames() *GreatBoardGames {
	return &GreatBoardGames{
		name:    "Great Board Games",
		baseURL: "https://www.greatboardgames.ca",
	}
}

// Name returns the store name
func (s *GreatBoardGames) Name() string {
	return s.name
}

// Check searches for a game at Great Board Games
func (s *GreatBoardGames) Check(gameName string) models.StoreResult {
	result := models.StoreResult{
		Store: s.name,
	}

	searchURL := fmt.Sprintf(
		"%s/search?q=%s",
		s.baseURL,
		url.QueryEscape(gameName),
	)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; WishlistChecker/1.0)")

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

	// Parse search results using regex
	// Looking for product cards with title, price, and URL
	productPattern := regexp.MustCompile(`<a[^>]*href="(/games/[^"]+)"[^>]*>.*?<h3[^>]*>([^<]+)</h3>.*?\$([0-9.]+)`)
	matches := productPattern.FindAllStringSubmatch(html, -1)

	if len(matches) == 0 {
		// Try alternate pattern
		s.parseAlternateFormat(html, gameName, &result)
	} else {
		s.parseStandardFormat(matches, gameName, &result)
	}

	return result
}

// parseStandardFormat parses the standard product card format
func (s *GreatBoardGames) parseStandardFormat(matches [][]string, gameName string, result *models.StoreResult) {
	for _, match := range matches {
		title := strings.TrimSpace(match[2])
		if utils.FuzzyMatch(gameName, title) {
			result.Found = true
			result.Title = title
			result.URL = s.baseURL + match[1]
			result.Price = "$" + match[3]
			result.PriceNum = utils.ParsePrice(match[3])
			result.InStock = true // If it shows up in search, likely in stock
			break
		}
	}
}

// parseAlternateFormat parses an alternate HTML format
func (s *GreatBoardGames) parseAlternateFormat(html, gameName string, result *models.StoreResult) {
	titlePattern := regexp.MustCompile(`href="(/games/[^"]+)"[^>]*>\s*<[^>]*>\s*([^<]+)`)
	pricePattern := regexp.MustCompile(`\$([0-9]+\.[0-9]{2})`)

	titleMatches := titlePattern.FindAllStringSubmatch(html, 10)
	priceMatches := pricePattern.FindAllStringSubmatch(html, 10)

	for i, tm := range titleMatches {
		title := strings.TrimSpace(tm[2])
		if utils.FuzzyMatch(gameName, title) {
			result.Found = true
			result.Title = title
			result.URL = s.baseURL + tm[1]
			result.InStock = !strings.Contains(html, "Out of Stock") && !strings.Contains(html, "Sold Out")
			if i < len(priceMatches) {
				result.Price = "$" + priceMatches[i][1]
				result.PriceNum = utils.ParsePrice(priceMatches[i][1])
			}
			break
		}
	}
}

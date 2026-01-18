package stores

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/utils"
)

type LaRevanche struct {
	name    string
	baseURL string
}

func NewLaRevanche() *LaRevanche {
	return &LaRevanche{
		name:    "La Revanche",
		baseURL: "https://boutique.larevanche.ca",
	}
}

func (s *LaRevanche) Name() string { return s.name }

func (s *LaRevanche) Check(gameName string) models.StoreResult {
	searchURL := fmt.Sprintf("%s/search?q=%s", s.baseURL, url.QueryEscape(gameName))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept-Language", "fr-CA,fr;q=0.9,en;q=0.8")

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

	// Parse products from gtag view_item_list event
	// Format: gtag('event', 'view_item_list', {"items":[{...}]})
	gtagPattern := regexp.MustCompile(`gtag\('event',\s*'view_item_list',\s*(\{[^;]+\})`)
	gtagMatch := gtagPattern.FindStringSubmatch(html)

	if gtagMatch == nil {
		return models.StoreResult{Store: s.name}
	}

	// Extract individual products from items array
	itemPattern := regexp.MustCompile(`"item_id"\s*:\s*"([^"]+)"[^}]*"item_name"\s*:\s*"([^"]+)"[^}]*"price"\s*:\s*(\d+(?:\.\d+)?)`)
	itemMatches := itemPattern.FindAllStringSubmatch(gtagMatch[1], -1)

	// Also extract product URLs from the page
	urlPattern := regexp.MustCompile(`href="(https://boutique\.larevanche\.ca/fc/[^"]+\.html)"`)
	urlMatches := urlPattern.FindAllStringSubmatch(html, -1)
	urls := make(map[string]string) // slug -> full URL
	for _, m := range urlMatches {
		// Extract slug from URL for matching
		slug := strings.TrimSuffix(strings.TrimPrefix(m[1], "https://boutique.larevanche.ca/fc/"), ".html")
		urls[slug] = m[1]
	}

	var matches []models.ProductMatch
	for _, item := range itemMatches {
		itemID := item[1]
		title := item[2]
		priceStr := item[3]

		if utils.ShouldExclude(title) || !utils.FuzzyMatch(gameName, title) {
			continue
		}

		priceNum, _ := strconv.ParseFloat(priceStr, 64)

		// Try to find URL by matching product slug
		productURL := ""
		titleSlug := slugify(title)
		for slug, u := range urls {
			if strings.Contains(slug, titleSlug) || strings.Contains(titleSlug, slug) {
				productURL = u
				break
			}
		}
		if productURL == "" {
			// Fallback: construct URL from item ID or skip
			productURL = fmt.Sprintf("%s/search?q=%s", s.baseURL, url.QueryEscape(title))
		}

		// Check stock status by looking for "Hors stock" near product references
		inStock := !strings.Contains(html, itemID) || !containsNear(html, itemID, "Hors stock", 500)

		matches = append(matches, models.ProductMatch{
			Title:    title,
			URL:      productURL,
			Price:    fmt.Sprintf("$%.2f", priceNum),
			PriceNum: priceNum,
			InStock:  inStock,
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

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "[français]", "francais")
	s = strings.ReplaceAll(s, "[anglais]", "anglais")
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func containsNear(html, anchor, search string, distance int) bool {
	idx := strings.Index(html, anchor)
	if idx == -1 {
		return false
	}
	start := idx - distance
	if start < 0 {
		start = 0
	}
	end := idx + len(anchor) + distance
	if end > len(html) {
		end = len(html)
	}
	return strings.Contains(html[start:end], search)
}

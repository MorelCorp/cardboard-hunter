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

type LeValet struct {
	name    string
	baseURL string
}

func NewLeValet() *LeValet {
	return &LeValet{
		name:    "Le Valet d'Coeur",
		baseURL: "https://levalet.com",
	}
}

func (s *LeValet) Name() string { return s.name }

func (s *LeValet) Check(gameName string) models.StoreResult {
	searchURL := fmt.Sprintf("%s/fr/catalogsearch/result/?q=%s", s.baseURL, url.QueryEscape(gameName))

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept-Language", "fr-CA,fr;q=0.9")

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

	// Split by product items
	cardSplitter := regexp.MustCompile(`<li[^>]*class="[^"]*product-item[^"]*"`)
	// Product link pattern: <a> with class="product-item-link" and href (in any order)
	titlePattern := regexp.MustCompile(`<a[^>]*href="([^"]+)"[^>]*class="product-item-link"[^>]*>([^<]+)</a>`)
	titlePatternAlt := regexp.MustCompile(`<a[^>]*class="product-item-link"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`)
	// French price: 44,99 $CA or 44,99$CA or data-price-amount="44.99"
	pricePattern := regexp.MustCompile(`data-price-amount="([^"]+)"`)
	pricePatternAlt := regexp.MustCompile(`(\d+)[,.](\d{2})\s*\$`)

	var matches []models.ProductMatch
	parts := cardSplitter.Split(html, -1)

	for i, cardHTML := range parts {
		if i == 0 {
			continue // Skip content before first product
		}

		titleMatch := titlePattern.FindStringSubmatch(cardHTML)
		if titleMatch == nil {
			titleMatch = titlePatternAlt.FindStringSubmatch(cardHTML)
		}
		if titleMatch == nil {
			continue
		}

		title := strings.TrimSpace(titleMatch[2])
		productURL := titleMatch[1]

		if utils.ShouldExclude(title) || !utils.FuzzyMatch(gameName, title) {
			continue
		}

		// Determine stock status
		// In stock: has "Ajouter au panier" and no "Rupture" or "Hors d'impression"
		inStock := strings.Contains(cardHTML, "Ajouter au panier") &&
			!strings.Contains(cardHTML, "Rupture") &&
			!strings.Contains(cardHTML, "Hors d'impression")

		match := models.ProductMatch{
			Title:   title,
			URL:     productURL,
			InStock: inStock,
		}

		// Parse price - try data-price-amount first, then French format
		if priceMatch := pricePattern.FindStringSubmatch(cardHTML); priceMatch != nil {
			match.PriceNum = utils.ParsePrice(priceMatch[1])
			match.Price = fmt.Sprintf("$%.2f", match.PriceNum)
		} else if priceMatch := pricePatternAlt.FindStringSubmatch(cardHTML); priceMatch != nil {
			match.Price = "$" + priceMatch[1] + "." + priceMatch[2]
			match.PriceNum = utils.ParsePrice(priceMatch[1] + "." + priceMatch[2])
		}

		matches = append(matches, match)
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

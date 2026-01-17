package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

type Game struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

type StoreResult struct {
	Store     string  `json:"store"`
	Found     bool    `json:"found"`
	InStock   bool    `json:"inStock"`
	Price     string  `json:"price"`
	PriceNum  float64 `json:"priceNum"`
	URL       string  `json:"url"`
	Title     string  `json:"title"`
	Error     string  `json:"error,omitempty"`
}

type GameResult struct {
	Name    string        `json:"name"`
	Results []StoreResult `json:"results"`
}

type CheckRequest struct {
	Games []Game `json:"games"`
}

type CheckResponse struct {
	Results  []GameResult     `json:"results"`
	Summary  map[string]int   `json:"summary"`
}

// Store configurations
type Store struct {
	Name    string
	BaseURL string
	Checker func(gameName string) StoreResult
}

var stores = []Store{
	{
		Name:    "Board Game Bliss",
		BaseURL: "https://www.boardgamebliss.com",
		Checker: checkBoardGameBliss,
	},
	{
		Name:    "401 Games",
		BaseURL: "https://store.401games.ca",
		Checker: check401Games,
	},
	{
		Name:    "Great Board Games",
		BaseURL: "https://www.greatboardgames.ca",
		Checker: checkGreatBoardGames,
	},
}

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

func main() {
	// Serve static files (need to strip "static/" prefix from embedded FS)
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	// API endpoint
	http.HandleFunc("/api/check", handleCheck)

	port := "8080"
	fmt.Printf("🎲 Wishlist Checker running at http://localhost:%s\n", port)
	fmt.Println("Open your browser to start checking game availability!")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	results := make([]GameResult, len(req.Games))
	summary := make(map[string]int)

	// Initialize summary
	for _, store := range stores {
		summary[store.Name] = 0
	}

	// Process games with limited concurrency
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 3) // Limit concurrent game checks

	for i, game := range req.Games {
		wg.Add(1)
		go func(idx int, g Game) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			gameResult := checkGame(g.Name)
			results[idx] = gameResult
		}(i, game)
	}

	wg.Wait()

	// Calculate summary
	for _, gr := range results {
		for _, sr := range gr.Results {
			if sr.Found && sr.InStock {
				summary[sr.Store]++
			}
		}
	}

	response := CheckResponse{
		Results: results,
		Summary: summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func checkGame(gameName string) GameResult {
	result := GameResult{
		Name:    gameName,
		Results: make([]StoreResult, len(stores)),
	}

	var wg sync.WaitGroup
	for i, store := range stores {
		wg.Add(1)
		go func(idx int, s Store) {
			defer wg.Done()
			result.Results[idx] = s.Checker(gameName)
		}(i, store)
	}
	wg.Wait()

	return result
}

// Board Game Bliss - Shopify store
func checkBoardGameBliss(gameName string) StoreResult {
	result := StoreResult{
		Store: "Board Game Bliss",
	}

	searchURL := fmt.Sprintf(
		"https://www.boardgamebliss.com/search/suggest.json?q=%s&resources[type]=product&resources[limit]=10",
		url.QueryEscape(gameName),
	)

	resp, err := httpClient.Get(searchURL)
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

	var data struct {
		Resources struct {
			Results struct {
				Products []struct {
					Title     string `json:"title"`
					URL       string `json:"url"`
					Price     string `json:"price"`
					Available bool   `json:"available"`
				} `json:"products"`
			} `json:"results"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		result.Error = err.Error()
		return result
	}

	products := data.Resources.Results.Products
	if len(products) == 0 {
		return result
	}

	// Find best match
	for _, product := range products {
		if fuzzyMatch(gameName, product.Title) {
			result.Found = true
			result.Title = product.Title
			result.URL = "https://www.boardgamebliss.com" + product.URL
			result.InStock = product.Available
			result.Price = product.Price
			result.PriceNum = parsePrice(product.Price)
			break
		}
	}

	// Fallback to first result if nothing matched but we have results
	if !result.Found && len(products) > 0 {
		product := products[0]
		result.Found = true
		result.Title = product.Title
		result.URL = "https://www.boardgamebliss.com" + product.URL
		result.InStock = product.Available
		result.Price = product.Price
		result.PriceNum = parsePrice(product.Price)
	}

	return result
}

// 401 Games - Shopify store
func check401Games(gameName string) StoreResult {
	result := StoreResult{
		Store: "401 Games",
	}

	searchURL := fmt.Sprintf(
		"https://store.401games.ca/search/suggest.json?q=%s&resources[type]=product&resources[limit]=10&resources[options][fields]=title,product_type,variants.title",
		url.QueryEscape(gameName),
	)

	resp, err := httpClient.Get(searchURL)
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

	var data struct {
		Resources struct {
			Results struct {
				Products []struct {
					Title     string `json:"title"`
					URL       string `json:"url"`
					Price     string `json:"price"`
					Available bool   `json:"available"`
				} `json:"products"`
			} `json:"results"`
		} `json:"resources"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		result.Error = err.Error()
		return result
	}

	products := data.Resources.Results.Products
	if len(products) == 0 {
		return result
	}

	// Find best match (filtering out non-board-game products)
	for _, product := range products {
		titleLower := strings.ToLower(product.Title)
		// Skip TCG singles, sleeves, etc.
		if strings.Contains(titleLower, "sleeve") ||
			strings.Contains(titleLower, "single") ||
			strings.Contains(titleLower, "booster") {
			continue
		}

		if fuzzyMatch(gameName, product.Title) {
			result.Found = true
			result.Title = product.Title
			result.URL = "https://store.401games.ca" + product.URL
			result.InStock = product.Available
			result.Price = product.Price
			result.PriceNum = parsePrice(product.Price)
			break
		}
	}

	// Fallback to first non-TCG result if nothing matched
	if !result.Found {
		for _, product := range products {
			titleLower := strings.ToLower(product.Title)
			if strings.Contains(titleLower, "sleeve") ||
				strings.Contains(titleLower, "single") ||
				strings.Contains(titleLower, "booster") {
				continue
			}
			result.Found = true
			result.Title = product.Title
			result.URL = "https://store.401games.ca" + product.URL
			result.InStock = product.Available
			result.Price = product.Price
			result.PriceNum = parsePrice(product.Price)
			break
		}
	}

	return result
}

// Great Board Games - Custom site (HTML scraping)
func checkGreatBoardGames(gameName string) StoreResult {
	result := StoreResult{
		Store: "Great Board Games",
	}

	// GBG uses a simple search
	searchURL := fmt.Sprintf(
		"https://www.greatboardgames.ca/search?q=%s",
		url.QueryEscape(gameName),
	)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; WishlistChecker/1.0)")

	resp, err := httpClient.Do(req)
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

	// Parse search results using regex (fragile but works for now)
	// Looking for product cards with title, price, and URL
	productPattern := regexp.MustCompile(`<a[^>]*href="(/games/[^"]+)"[^>]*>.*?<h3[^>]*>([^<]+)</h3>.*?\$([0-9.]+)`)
	matches := productPattern.FindAllStringSubmatch(html, -1)

	if len(matches) == 0 {
		// Try alternate pattern
		titlePattern := regexp.MustCompile(`href="(/games/[^"]+)"[^>]*>\s*<[^>]*>\s*([^<]+)`)
		pricePattern := regexp.MustCompile(`\$([0-9]+\.[0-9]{2})`)
		
		titleMatches := titlePattern.FindAllStringSubmatch(html, 10)
		priceMatches := pricePattern.FindAllStringSubmatch(html, 10)

		for i, tm := range titleMatches {
			title := strings.TrimSpace(tm[2])
			if fuzzyMatch(gameName, title) {
				result.Found = true
				result.Title = title
				result.URL = "https://www.greatboardgames.ca" + tm[1]
				result.InStock = !strings.Contains(html, "Out of Stock") && !strings.Contains(html, "Sold Out")
				if i < len(priceMatches) {
					result.Price = "$" + priceMatches[i][1]
					result.PriceNum = parsePrice(priceMatches[i][1])
				}
				break
			}
		}
	} else {
		for _, match := range matches {
			title := strings.TrimSpace(match[2])
			if fuzzyMatch(gameName, title) {
				result.Found = true
				result.Title = title
				result.URL = "https://www.greatboardgames.ca" + match[1]
				result.Price = "$" + match[3]
				result.PriceNum = parsePrice(match[3])
				result.InStock = true // If it shows up in search, likely in stock
				break
			}
		}
	}

	return result
}

// Helper functions

func fuzzyMatch(search, title string) bool {
	searchLower := strings.ToLower(search)
	titleLower := strings.ToLower(title)

	// Exact substring match
	if strings.Contains(titleLower, searchLower) {
		return true
	}

	// Check if all words from search appear in title
	searchWords := strings.Fields(searchLower)
	allFound := true
	for _, word := range searchWords {
		if len(word) < 3 {
			continue
		}
		if !strings.Contains(titleLower, word) {
			allFound = false
			break
		}
	}

	return allFound
}

func parsePrice(priceStr string) float64 {
	// Remove currency symbols and parse
	cleaned := strings.ReplaceAll(priceStr, "$", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "CAD", "")
	cleaned = strings.TrimSpace(cleaned)

	var price float64
	fmt.Sscanf(cleaned, "%f", &price)
	return price
}

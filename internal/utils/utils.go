package utils

import (
	"fmt"
	"strings"
)

// FuzzyMatch checks if a search term matches a title using fuzzy matching logic
func FuzzyMatch(search, title string) bool {
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

// ParsePrice extracts a numeric price from a price string
func ParsePrice(priceStr string) float64 {
	// Remove currency symbols and parse
	cleaned := strings.ReplaceAll(priceStr, "$", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, "CAD", "")
	cleaned = strings.TrimSpace(cleaned)

	var price float64
	fmt.Sscanf(cleaned, "%f", &price)
	return price
}

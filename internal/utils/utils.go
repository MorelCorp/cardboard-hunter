package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var wordSplitter = regexp.MustCompile(`[\s\-:,()\[\]]+`)

func splitIntoWords(s string) []string {
	parts := wordSplitter.Split(s, -1)
	words := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			words = append(words, p)
		}
	}
	return words
}

// ExactTitleMatch checks if title matches search exactly (ignoring case)
func ExactTitleMatch(search, title string) bool {
	return strings.EqualFold(strings.TrimSpace(search), strings.TrimSpace(title))
}

// ShouldExclude checks if a product should be filtered out
func ShouldExclude(title string) bool {
	t := strings.ToLower(title)
	return strings.Contains(t, "pre-order") ||
		strings.Contains(t, "preorder") ||
		strings.Contains(t, "extension") ||
		strings.Contains(t, "expansion")
}

// FuzzyMatch checks if search term matches title using word boundary matching
func FuzzyMatch(search, title string) bool {
	searchLower := strings.ToLower(search)
	titleLower := strings.ToLower(title)
	titleWords := splitIntoWords(titleLower)

	searchWords := strings.Fields(searchLower)
	if len(searchWords) == 1 {
		// Single word: must match as complete word
		for _, w := range titleWords {
			if w == searchLower {
				return true
			}
		}
		return false
	}

	// Multi-word: all search words must be present as complete words
	for _, sw := range searchWords {
		found := false
		for _, tw := range titleWords {
			if tw == sw {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
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

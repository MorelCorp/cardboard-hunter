package checker

import (
	"sync"

	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/stores"
)

// Checker handles game availability checking across multiple stores
type Checker struct {
	stores []stores.Store
}

// New creates a new Checker with all available stores
func New() *Checker {
	return &Checker{
		stores: stores.GetAllStores(),
	}
}

// CheckGame checks a single game across all stores concurrently
func (c *Checker) CheckGame(gameName string) models.GameResult {
	result := models.GameResult{
		Name:    gameName,
		Results: make([]models.StoreResult, len(c.stores)),
	}

	var wg sync.WaitGroup
	for i, store := range c.stores {
		wg.Add(1)
		go func(idx int, s stores.Store) {
			defer wg.Done()
			result.Results[idx] = s.Check(gameName)
		}(i, store)
	}
	wg.Wait()

	return result
}

// CheckGames checks multiple games with limited concurrency
func (c *Checker) CheckGames(games []models.Game) []models.GameResult {
	results := make([]models.GameResult, len(games))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 3) // Limit concurrent game checks

	for i, game := range games {
		wg.Add(1)
		go func(idx int, g models.Game) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results[idx] = c.CheckGame(g.Name)
		}(i, game)
	}

	wg.Wait()
	return results
}

// CalculateSummary calculates how many in-stock games each store has
func (c *Checker) CalculateSummary(results []models.GameResult) map[string]int {
	summary := make(map[string]int)

	// Initialize summary for all stores
	for _, store := range c.stores {
		summary[store.Name()] = 0
	}

	// Count in-stock games per store
	for _, gr := range results {
		for _, sr := range gr.Results {
			if sr.Found && sr.InStock {
				summary[sr.Store]++
			}
		}
	}

	return summary
}

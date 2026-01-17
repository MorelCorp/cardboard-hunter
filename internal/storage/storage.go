package storage

import (
	"encoding/json"
	"os"
	"sync"

	"cardboard-hunter/internal/models"
)

const defaultStorageFile = "games.json"

// Storage handles persisting game lists to disk
type Storage struct {
	filepath string
	mu       sync.RWMutex
}

// New creates a new Storage instance
func New(filepath string) *Storage {
	if filepath == "" {
		filepath = defaultStorageFile
	}
	return &Storage{
		filepath: filepath,
	}
}

// LoadGames loads the game list from disk
func (s *Storage) LoadGames() ([]models.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(s.filepath); os.IsNotExist(err) {
		return []models.Game{}, nil
	}

	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return nil, err
	}

	var games []models.Game
	if err := json.Unmarshal(data, &games); err != nil {
		return nil, err
	}

	return games, nil
}

// SaveGames saves the game list to disk
func (s *Storage) SaveGames(games []models.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(games, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filepath, data, 0644)
}

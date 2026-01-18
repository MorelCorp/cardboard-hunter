package config

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
)

//go:embed defaults/*.json defaults/stores/*.json
var embeddedConfigs embed.FS

// Loader handles configuration loading with external override support
type Loader struct {
	configDir string
}

// NewLoader creates a config loader. If configDir is empty, uses embedded configs only.
func NewLoader(configDir string) *Loader {
	return &Loader{configDir: configDir}
}

// LoadStoresConfig loads the main stores configuration
func (l *Loader) LoadStoresConfig() (*StoresConfig, error) {
	data, err := l.readFile("stores.json")
	if err != nil {
		return nil, err
	}
	var cfg StoresConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadStoreConfig loads an individual store's configuration
func (l *Loader) LoadStoreConfig(ref StoreRef) (*StoreConfig, error) {
	if ref.Builtin {
		return nil, nil
	}
	data, err := l.readFile(ref.File)
	if err != nil {
		return nil, err
	}
	var cfg StoreConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// readFile attempts to read from external config dir first, then embedded
func (l *Loader) readFile(relPath string) ([]byte, error) {
	if l.configDir != "" {
		extPath := filepath.Join(l.configDir, relPath)
		if data, err := os.ReadFile(extPath); err == nil {
			return data, nil
		}
	}
	return embeddedConfigs.ReadFile("defaults/" + relPath)
}

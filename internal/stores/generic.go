package stores

import (
	"cardboard-hunter/internal/config"
	"cardboard-hunter/internal/models"
)

// GenericStore wraps config-driven store implementations
type GenericStore struct {
	cfg     *config.StoreConfig
	checker checker
}

type checker interface {
	Check(gameName string) models.StoreResult
}

// NewGenericStore creates a store from configuration
func NewGenericStore(cfg *config.StoreConfig) *GenericStore {
	var c checker
	switch cfg.Type {
	case config.StoreTypeShopify:
		c = NewShopifyChecker(cfg)
	case config.StoreTypeHTMLScraper:
		c = NewScraperChecker(cfg)
	case config.StoreTypeJSONAPI:
		c = NewJSONAPIChecker(cfg)
	}
	return &GenericStore{cfg: cfg, checker: c}
}

func (s *GenericStore) Name() string {
	return s.cfg.Name
}

func (s *GenericStore) Check(gameName string) models.StoreResult {
	if s.checker == nil {
		return models.StoreResult{Store: s.cfg.Name, Error: "unknown store type"}
	}
	return s.checker.Check(gameName)
}

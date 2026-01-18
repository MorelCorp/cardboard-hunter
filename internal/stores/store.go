package stores

import (
	"net/http"
	"os"
	"time"

	"cardboard-hunter/internal/config"
	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/shopify"
)

// Store represents a board game store with checking capabilities
type Store interface {
	Name() string
	Check(gameName string) models.StoreResult
}

// HTTPClient is the shared HTTP client for all stores
var HTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

// ShopifyClient is the shared Shopify API client
var ShopifyClient = &shopify.Client{
	HTTPClient: HTTPClient,
}

// GetAllStores returns all available store implementations
func GetAllStores() []Store {
	configDir := os.Getenv("CARDBOARD_CONFIG_DIR")
	loader := config.NewLoader(configDir)

	mainCfg, err := loader.LoadStoresConfig()
	if err != nil {
		return builtinStores()
	}

	var stores []Store
	for _, ref := range mainCfg.Stores {
		if ref.Builtin {
			if s := getBuiltinStore(ref.ID); s != nil {
				stores = append(stores, s)
			}
			continue
		}

		storeCfg, err := loader.LoadStoreConfig(ref)
		if err != nil || storeCfg == nil || !storeCfg.Enabled {
			continue
		}
		stores = append(stores, NewGenericStore(storeCfg))
	}

	if len(stores) == 0 {
		return builtinStores()
	}
	return stores
}

func getBuiltinStore(id string) Store {
	switch id {
	case "larevanche":
		return NewLaRevanche()
	default:
		return nil
	}
}

func builtinStores() []Store {
	return []Store{
		NewLaRevanche(),
	}
}

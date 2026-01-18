package config

// StoreType represents the type of store checker
type StoreType string

const (
	StoreTypeShopify     StoreType = "shopify"
	StoreTypeHTMLScraper StoreType = "html_scraper"
	StoreTypeJSONAPI     StoreType = "json_api"
)

// StoresConfig is the main configuration file
type StoresConfig struct {
	Version  int           `json:"version"`
	Stores   []StoreRef    `json:"stores"`
	Defaults DefaultConfig `json:"defaults"`
}

// StoreRef references a store in the main config
type StoreRef struct {
	ID      string `json:"id"`
	File    string `json:"file,omitempty"`
	Builtin bool   `json:"builtin,omitempty"`
}

// DefaultConfig holds default settings
type DefaultConfig struct {
	MaxMatches int    `json:"maxMatches"`
	Timeout    string `json:"timeout"`
}

// StoreConfig represents a single store's configuration
type StoreConfig struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Enabled bool              `json:"enabled"`
	Type    StoreType         `json:"type"`
	BaseURL string            `json:"baseURL"`
	Headers map[string]string `json:"headers,omitempty"`
	Shopify *ShopifyConfig    `json:"shopify,omitempty"`
	Scraper *ScraperConfig    `json:"scraper,omitempty"`
	JSONAPI *JSONAPIConfig    `json:"jsonApi,omitempty"`
}

// ShopifyConfig for Shopify-based stores
type ShopifyConfig struct {
	ExcludePatterns []string `json:"excludePatterns,omitempty"`
}

// ScraperConfig for HTML scraping stores
type ScraperConfig struct {
	SearchPath           string         `json:"searchPath"`
	CardSplitter         string         `json:"cardSplitter"`
	TitlePatterns        []string       `json:"titlePatterns,omitempty"`
	TitleGroups          CaptureGroups  `json:"titleGroups"`
	PricePatterns        []PricePattern `json:"pricePatterns,omitempty"`
	PricePrefix          string         `json:"pricePrefix"`
	OutOfStockIndicators []string       `json:"outOfStockIndicators,omitempty"`
	InStockIndicators    []string       `json:"inStockIndicators,omitempty"`
	StockLogic           string         `json:"stockLogic,omitempty"` // "out_of_stock" (default) or "in_stock_required"
}

// CaptureGroups maps named captures to group indices
type CaptureGroups struct {
	URL   int `json:"url"`
	Title int `json:"title"`
}

// PricePattern defines a price extraction pattern
type PricePattern struct {
	Pattern string           `json:"pattern"`
	Groups  PriceCaptureMode `json:"groups"`
}

// PriceCaptureMode defines how to extract price from regex groups
type PriceCaptureMode struct {
	Amount  int `json:"amount,omitempty"`  // single group with full price
	Dollars int `json:"dollars,omitempty"` // group with dollar part
	Cents   int `json:"cents,omitempty"`   // group with cents part
}

// JSONAPIConfig for JSON API stores
type JSONAPIConfig struct {
	SearchPath   string       `json:"searchPath"`
	ProductsPath string       `json:"productsPath"`
	Fields       JSONFieldMap `json:"fields"`
	InStockValue string       `json:"inStockValue,omitempty"`
}

// JSONFieldMap maps product fields to JSON keys
type JSONFieldMap struct {
	Title       string `json:"title"`
	Price       string `json:"price"`
	URL         string `json:"url"`
	Quantity    string `json:"quantity,omitempty"`
	StockStatus string `json:"stockStatus,omitempty"`
}

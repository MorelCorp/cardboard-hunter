package stores

import (
	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/shopify"
)

type LaPioche struct {
	name    string
	baseURL string
}

func NewLaPioche() *LaPioche {
	return &LaPioche{
		name:    "La Pioche",
		baseURL: "https://boutiquelapioche.com",
	}
}

func (s *LaPioche) Name() string { return s.name }

func (s *LaPioche) Check(gameName string) models.StoreResult {
	products, err := ShopifyClient.Search(s.baseURL, gameName)
	if err != nil {
		return models.StoreResult{Store: s.name, Error: err.Error()}
	}

	matches := shopify.FindMatches(gameName, products, s.baseURL, 5)
	return shopify.BuildStoreResult(s.name, matches, gameName)
}

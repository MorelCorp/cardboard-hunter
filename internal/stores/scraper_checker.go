package stores

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"cardboard-hunter/internal/config"
	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/utils"
)

// ScraperChecker implements checking for HTML scraping stores
type ScraperChecker struct {
	cfg          *config.StoreConfig
	cardSplitter *regexp.Regexp
	titleRegexps []*regexp.Regexp
	priceRegexps []priceRegexp
}

type priceRegexp struct {
	re     *regexp.Regexp
	groups config.PriceCaptureMode
}

// NewScraperChecker creates a new HTML scraper checker from config
func NewScraperChecker(cfg *config.StoreConfig) *ScraperChecker {
	sc := &ScraperChecker{cfg: cfg}

	if cfg.Scraper == nil {
		return sc
	}

	sc.cardSplitter = regexp.MustCompile(cfg.Scraper.CardSplitter)

	for _, p := range cfg.Scraper.TitlePatterns {
		sc.titleRegexps = append(sc.titleRegexps, regexp.MustCompile(p))
	}

	for _, pp := range cfg.Scraper.PricePatterns {
		sc.priceRegexps = append(sc.priceRegexps, priceRegexp{
			re:     regexp.MustCompile(pp.Pattern),
			groups: pp.Groups,
		})
	}

	return sc
}

func (c *ScraperChecker) Check(gameName string) models.StoreResult {
	if c.cfg.Scraper == nil {
		return models.StoreResult{Store: c.cfg.Name, Error: "no scraper config"}
	}

	searchURL := c.cfg.BaseURL + strings.Replace(
		c.cfg.Scraper.SearchPath, "{query}", url.QueryEscape(gameName), 1)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return models.StoreResult{Store: c.cfg.Name, Error: err.Error()}
	}

	for k, v := range c.cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return models.StoreResult{Store: c.cfg.Name, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.StoreResult{Store: c.cfg.Name, Error: err.Error()}
	}

	html := string(body)
	parts := c.cardSplitter.Split(html, -1)

	var matches []models.ProductMatch
	for i, cardHTML := range parts {
		if i == 0 {
			continue
		}

		titleMatch := c.findTitleMatch(cardHTML)
		if titleMatch == nil {
			continue
		}

		title := strings.TrimSpace(titleMatch[c.cfg.Scraper.TitleGroups.Title])
		productURL := titleMatch[c.cfg.Scraper.TitleGroups.URL]

		if utils.ShouldExclude(title) || !utils.FuzzyMatch(gameName, title) {
			continue
		}

		price, priceNum := c.extractPrice(cardHTML)
		matches = append(matches, models.ProductMatch{
			Title:    title,
			URL:      productURL,
			Price:    price,
			PriceNum: priceNum,
			InStock:  c.determineStock(cardHTML),
		})

		if len(matches) >= 5 {
			break
		}
	}

	return buildResult(c.cfg.Name, matches, gameName)
}

func (c *ScraperChecker) findTitleMatch(cardHTML string) []string {
	for _, re := range c.titleRegexps {
		if m := re.FindStringSubmatch(cardHTML); m != nil {
			return m
		}
	}
	return nil
}

func (c *ScraperChecker) determineStock(cardHTML string) bool {
	scfg := c.cfg.Scraper

	for _, indicator := range scfg.OutOfStockIndicators {
		if strings.Contains(cardHTML, indicator) {
			return false
		}
	}

	if scfg.StockLogic == "in_stock_required" {
		for _, indicator := range scfg.InStockIndicators {
			if strings.Contains(cardHTML, indicator) {
				return true
			}
		}
		return false
	}

	return true
}

func (c *ScraperChecker) extractPrice(cardHTML string) (string, float64) {
	for _, pp := range c.priceRegexps {
		m := pp.re.FindStringSubmatch(cardHTML)
		if m == nil {
			continue
		}

		if pp.groups.Amount > 0 && pp.groups.Amount < len(m) {
			price := utils.ParsePrice(m[pp.groups.Amount])
			return fmt.Sprintf("%s%.2f", c.cfg.Scraper.PricePrefix, price), price
		}

		if pp.groups.Dollars > 0 && pp.groups.Cents > 0 &&
			pp.groups.Dollars < len(m) && pp.groups.Cents < len(m) {
			priceStr := m[pp.groups.Dollars] + "." + m[pp.groups.Cents]
			price := utils.ParsePrice(priceStr)
			return c.cfg.Scraper.PricePrefix + priceStr, price
		}
	}
	return "", 0
}

func buildResult(storeName string, matches []models.ProductMatch, gameName string) models.StoreResult {
	if len(matches) == 0 {
		return models.StoreResult{Store: storeName}
	}

	for _, m := range matches {
		if utils.ExactTitleMatch(gameName, m.Title) {
			matches = []models.ProductMatch{m}
			break
		}
	}

	first := matches[0]
	return models.StoreResult{
		Store:    storeName,
		Found:    true,
		Title:    first.Title,
		URL:      first.URL,
		Price:    first.Price,
		PriceNum: first.PriceNum,
		InStock:  first.InStock,
		Matches:  matches,
	}
}

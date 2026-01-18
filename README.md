# Cardboard Hunter

Check board game availability across multiple stores. Single Go binary with embedded web UI.

## Supported Stores

- **Board Game Bliss** (boardgamebliss.com) — Shopify
- **401 Games** (store.401games.ca) — Shopify, filters TCG products
- **Great Board Games** (greatboardgames.ca) — HTML scraper
- **La Pioche** (lapioche.ca) — Shopify
- **Board Games N More** (boardgamesnmore.com) — Shopify
- **Le Valet** (levalet.com) — HTML scraper
- **La Revanche** (larevanche.ca) — JSON API (builtin)

## Building & Running

Requires Go 1.21+

```bash
# Windows
build.bat

# Or directly
go build -o cardboard-hunter.exe .
./cardboard-hunter
```

The browser opens automatically on launch. Use the "Exit App" button to close.

## Features

### Wishlist Management
- Add/remove games, drag-and-drop reordering
- Priority-based ordering (top = most wanted)
- Star items as "must-have" — stores with starred items always rank first
- Import/export as text file
- Persisted server-side in `games.json`

### Two Result Views

**Table View** — Traditional grid showing each game × store with availability and price

**Cart View** — Ranks stores by how well they can fulfill your wishlist:
- Each store shown as a "cart" with top X available items
- Stores ranked by score (higher priority items = more points)
- Starred items give +1000 bonus to stores that have them
- Price comparison: "BEST" badge or "+$X.XX" difference vs cheapest
- Configurable limit (show top 5, 10, etc. items per store)

### Scoring Algorithm

```
Points per item = (TotalItems - Priority + 1)
```

Example with 10 items:
- Priority 1 = 10 pts, Priority 2 = 9 pts, ..., Priority 10 = 1 pt
- Store with items #2,3,4,5 → 9+8+7+6 = 30 pts
- Store with only #1 → 10 pts
- More coverage beats single high-priority item

Starred items: Any store with a starred item gets +1000 bonus.

### Price Comparison

- Compares prices across ALL stores (including out-of-stock)
- Filters out prices under $5 to avoid TCG/accessory false matches
- Shows "BEST" or price difference in cart view

## Project Structure

```
cardboard-hunter/
├── main.go                     # HTTP server, handlers
├── build.bat                   # Windows build script
├── internal/
│   ├── models/models.go        # Data structures (Game, StoreResult, etc.)
│   ├── checker/checker.go      # Concurrent game checking
│   ├── config/
│   │   ├── types.go            # Config structs
│   │   ├── loader.go           # Config loading (embedded + external)
│   │   └── defaults/
│   │       ├── stores.json     # Main store list
│   │       └── stores/*.json   # Individual store configs
│   ├── stores/
│   │   ├── store.go            # Store interface + registry
│   │   ├── shopify.go          # Shopify checker (config-driven)
│   │   ├── scraper.go          # HTML scraper (config-driven)
│   │   └── larevanche.go       # Builtin: La Revanche (custom JSON API)
│   ├── storage/storage.go      # games.json persistence
│   └── utils/utils.go          # FuzzyMatch, ParsePrice helpers
├── static/index.html           # Embedded web UI (all HTML/CSS/JS)
└── games.json                  # User's saved wishlist
```

## Store Configuration

Stores are defined in JSON config files embedded in the binary. Three store types are supported:

### Shopify Stores

```json
{
  "id": "boardgamebliss",
  "name": "Board Game Bliss",
  "enabled": true,
  "type": "shopify",
  "baseURL": "https://www.boardgamebliss.com",
  "shopify": {
    "excludePatterns": ["TCG", "Sleeve"]
  }
}
```

### HTML Scraper Stores

```json
{
  "id": "greatboardgames",
  "name": "Great Board Games",
  "enabled": true,
  "type": "html_scraper",
  "baseURL": "https://www.greatboardgames.ca",
  "headers": {"User-Agent": "..."},
  "scraper": {
    "searchPath": "/search?q={query}",
    "cardSplitter": "<div class=\"product-card",
    "titlePatterns": ["<a href=\"([^\"]+)\"[^>]*>([^<]+)</a>"],
    "titleGroups": {"url": 1, "title": 2},
    "pricePatterns": [{"pattern": "\\$([0-9.]+)", "groups": {"amount": 1}}],
    "pricePrefix": "$",
    "outOfStockIndicators": ["Out of Stock"],
    "stockLogic": "out_of_stock"
  }
}
```

### Adding a New Store

1. Create `internal/config/defaults/stores/newstore.json`
2. Add reference to `internal/config/defaults/stores.json`:
   ```json
   {"id": "newstore", "file": "stores/newstore.json"}
   ```

For stores that don't fit the Shopify or scraper patterns, create a builtin implementation in `internal/stores/` and mark it with `"builtin": true` in stores.json.

## API Endpoints

- `GET /` — Serves web UI
- `GET /api/games` — Load saved wishlist
- `POST /api/games` — Save wishlist
- `POST /api/check` — Check availability (returns results + summary)
- `POST /api/shutdown` — Exit the application

## Data Models

```go
type Game struct {
    Name     string `json:"name"`
    Priority int    `json:"priority"`
    Starred  bool   `json:"starred"`
}

type StoreResult struct {
    Store    string  `json:"store"`
    Found    bool    `json:"found"`
    InStock  bool    `json:"inStock"`
    Price    string  `json:"price"`
    PriceNum float64 `json:"priceNum"`
    URL      string  `json:"url"`
    Title    string  `json:"title"`
    Error    string  `json:"error,omitempty"`
}
```

## Notes

- Fuzzy matching: searches where title contains search terms
- 401 Games filters out TCG sleeves/singles/boosters
- Prices in CAD
- HTML scraping (Great Board Games) may break if site changes

## License

MIT

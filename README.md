# Board Game Wishlist Checker

Check board game availability across Canadian retailers. Single Go binary with embedded web UI.

## Supported Stores

- **Board Game Bliss** (boardgamebliss.com) — Shopify API
- **401 Games** (store.401games.ca) — Shopify API, filters TCG products
- **Great Board Games** (greatboardgames.ca) — HTML scraping

## Building & Running

Requires Go 1.21+

```bash
go build -o cardboard-hunter .
./cardboard-hunter
```

Open http://localhost:8080

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
├── internal/
│   ├── models/models.go        # Data structures (Game, StoreResult, etc.)
│   ├── checker/checker.go      # Concurrent game checking
│   ├── stores/
│   │   ├── store.go            # Store interface
│   │   ├── shopify.go          # Shared Shopify client
│   │   ├── boardgamebliss.go   # Board Game Bliss implementation
│   │   ├── games401.go         # 401 Games (with TCG filtering)
│   │   └── greatboardgames.go  # Great Board Games (HTML scraping)
│   ├── storage/storage.go      # games.json persistence
│   └── utils/utils.go          # FuzzyMatch, ParsePrice helpers
├── static/index.html           # Embedded web UI (all HTML/CSS/JS)
└── games.json                  # User's saved wishlist
```

## Adding Stores

1. Create `internal/stores/newstore.go` implementing the `Store` interface
2. Register in `internal/stores/store.go`

Shopify stores are straightforward (see `boardgamebliss.go`). Others need HTML scraping.

## API Endpoints

- `GET /` — Serves web UI
- `GET /api/games` — Load saved wishlist
- `POST /api/games` — Save wishlist
- `POST /api/check` — Check availability (returns results + summary)

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

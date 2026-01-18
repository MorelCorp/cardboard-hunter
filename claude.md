# cardboard-hunter

Board game availability checker for Canadian retailers. Go binary with embedded web UI.

## Code Style

- **Concise over verbose.** No explanatory comments for obvious code. No wrappers that add nothing.
- **Flat over nested.** Early returns. Avoid deep indentation.
- **Standard library first.** Only add dependencies when they genuinely earn their keep.
- **Error handling:** Return errors up, handle them once at the appropriate level.

## Interaction Rules

- **Be direct.** If there's a better approach, say so. If my idea is flawed, tell me why.
- **No fluff.** Skip the preamble, skip the "great question!", skip restating what I said.
- **Peer mode.** Treat this like a code review between equals. Push back when warranted.
- **Show, don't lecture.** Code examples over explanations when possible.

## Architecture

```
main.go                         # HTTP server, routes, browser launch
internal/
  config/
    types.go                    # StoreConfig, ShopifyConfig, ScraperConfig structs
    loader.go                   # Loads embedded + external configs
    defaults/
      stores.json               # Store registry (id -> file mapping)
      stores/*.json             # Individual store configs
  checker/checker.go            # Concurrent game checking across stores
  stores/
    store.go                    # Store interface, registry, factory
    shopify.go                  # Config-driven Shopify checker
    scraper.go                  # Config-driven HTML scraper
    larevanche.go               # Builtin store (custom JSON API)
  models/models.go              # Game, StoreResult, CheckRequest/Response
  storage/storage.go            # games.json persistence
  utils/utils.go                # FuzzyMatch, ParsePrice
static/index.html               # Embedded SPA (vanilla JS)
```

## Store Types

**Shopify** (`type: "shopify"`): Uses `/products.json?title=query` API. Config specifies `excludePatterns` for filtering unwanted products.

**HTML Scraper** (`type: "html_scraper"`): Fetches search page, splits by `cardSplitter`, extracts title/price/stock via regex patterns defined in config.

**Builtin**: For stores needing custom logic (e.g., La Revanche's JSON API). Marked `"builtin": true` in stores.json, implemented in `internal/stores/`.

## Adding a Store

**Config-driven (preferred):**
1. Create `internal/config/defaults/stores/storename.json`
2. Add to `internal/config/defaults/stores.json`
3. Rebuild

**Builtin (custom logic needed):**
1. Implement `Store` interface in `internal/stores/storename.go`
2. Register in `internal/stores/store.go` registry
3. Add to stores.json with `"builtin": true`

## Key Patterns

- **Fuzzy matching**: `utils.FuzzyMatch(title, query)` - all query words must appear in title (case-insensitive)
- **Price parsing**: `utils.ParsePrice(str)` - extracts float from "$XX.XX" format
- **Concurrency**: `checker.CheckGames()` runs all stores in parallel per game
- **Config loading**: Embedded defaults, can be overridden by external `config/` dir

## Frontend

Single HTML file with vanilla JS. Key state:
- `wishlist[]` - games with name, priority, starred
- `lastResults` - cached API response for re-rendering
- `selectedMatches{}` - user's match selections per game/store
- `viewMode` - "table" or "cart"

## API

- `GET /api/games` - load wishlist
- `POST /api/games` - save wishlist
- `POST /api/check` - check availability, returns `{results, summary}`
- `POST /api/shutdown` - exit app

## Build

```bash
build.bat           # Windows
go build .          # Any platform
```

Browser auto-opens on launch. "Exit App" button calls `/api/shutdown`.

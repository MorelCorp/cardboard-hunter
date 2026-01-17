# 🎲 Board Game Wishlist Checker

A local tool to check board game availability across Canadian retailers.

## Supported Stores

- **Board Game Bliss** (boardgamebliss.com)
- **401 Games** (store.401games.ca)
- **Great Board Games** (greatboardgames.ca)

## Building

Requires Go 1.21+

```bash
# Build for your current platform
go build -o wishlist-checker .

# Or build for specific platforms:
# Windows
GOOS=windows GOARCH=amd64 go build -o wishlist-checker.exe .

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o wishlist-checker-mac .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o wishlist-checker-mac-arm .

# Linux
GOOS=linux GOARCH=amd64 go build -o wishlist-checker-linux .
```

## Running

```bash
./wishlist-checker
```

Then open http://localhost:8080 in your browser.

## Usage

1. **Add games** — Type game names and click "Add" (or press Enter)
2. **Prioritize** — Use ↑/↓ buttons to reorder your wishlist
3. **Check** — Click "Check Availability" to search all stores
4. **Review** — See which store has the most games in stock, compare prices

## Features

- Wishlist saved locally in browser (persists between sessions)
- Import/export wishlist as text file (one game per line)
- Parallel store checking for speed
- Best price highlighting (⭐)
- Direct links to product pages

## How It Works

The app runs a local Go server that:
1. Serves the web UI (embedded in the binary)
2. Provides an API endpoint to check stores
3. Queries each store in parallel:
   - Board Game Bliss & 401 Games: Uses Shopify's JSON search API
   - Great Board Games: Scrapes HTML search results

## Notes

- Search matching is fuzzy — it looks for games where the title contains your search terms
- 401 Games sells TCG singles which can pollute results; the app filters out sleeves/singles/boosters
- Great Board Games parsing is HTML-based and may break if they change their site structure
- Prices are in CAD

## Adding More Stores

To add a new store, add a `Store` entry in `main.go` with a checker function:

```go
{
    Name:    "New Store",
    BaseURL: "https://newstore.ca",
    Checker: checkNewStore,
},
```

Then implement `checkNewStore(gameName string) StoreResult`.

## License

MIT — do whatever you want with it.

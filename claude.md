# cardboard-hunter

Board game availability checker for Canadian retailers. Go binary with embedded web UI.

## Code Style

- **Concise over verbose.** No explanatory comments for obvious code. No wrappers that add nothing.
- **Flat over nested.** Early returns. Avoid deep indentation.
- **Standard library first.** Only add dependencies when they genuinely earn their keep.
- **Error handling:** Return errors up, handle them once at the appropriate level. No `if err != nil { return err }` chains when a helper would be cleaner.

## Interaction Rules

- **Be direct.** If there's a better approach, say so. If my idea is flawed, tell me why.
- **No fluff.** Skip the preamble, skip the "great question!", skip restating what I said.
- **Peer mode.** Treat this like a code review between equals. Push back when warranted.
- **Show, don't lecture.** Code examples over explanations when possible.

## Project Context

- Stores use either Shopify JSON API or HTML scraping
- Search matching is fuzzy/contains-based
- Results need filtering (401 Games TCG pollution)
- Single binary deployment - UI is embedded

## When Adding Stores

1. Add `Store` entry in main.go
2. Implement checker function: `func checkStoreName(gameName string) StoreResult`
3. Shopify stores are straightforward; others likely need goquery or similar
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"cardboard-hunter/internal/checker"
	"cardboard-hunter/internal/models"
	"cardboard-hunter/internal/storage"
)

//go:embed static/*
var staticFiles embed.FS

var store *storage.Storage

func main() {
	// Initialize storage
	store = storage.New("games.json")
	// Serve static files (need to strip "static/" prefix from embedded FS)
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	// API endpoints
	http.HandleFunc("/api/check", handleCheck)
	http.HandleFunc("/api/games", handleGames)
	http.HandleFunc("/api/shutdown", handleShutdown)

	port := "8080"
	url := fmt.Sprintf("http://localhost:%s", port)
	fmt.Printf("🎲 Cardboard Hunter running at %s\n", url)
	go openBrowser(url)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Run()
}

func handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "shutting down"})
	go func() {
		os.Exit(0)
	}()
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create checker and process games
	c := checker.New()
	results := c.CheckGames(req.Games)
	summary := c.CalculateSummary(results)

	response := models.CheckResponse{
		Results: results,
		Summary: summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Load saved games
		games, err := store.LoadGames()
		if err != nil {
			http.Error(w, "Failed to load games", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(games)

	case http.MethodPost:
		// Save games
		var games []models.Game
		if err := json.NewDecoder(r.Body).Decode(&games); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := store.SaveGames(games); err != nil {
			http.Error(w, "Failed to save games", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

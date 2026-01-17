package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"cardboard-hunter/internal/checker"
	"cardboard-hunter/internal/models"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// Serve static files (need to strip "static/" prefix from embedded FS)
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	// API endpoint
	http.HandleFunc("/api/check", handleCheck)

	port := "8080"
	fmt.Printf("🎲 Wishlist Checker running at http://localhost:%s\n", port)
	fmt.Println("Open your browser to start checking game availability!")
	log.Fatal(http.ListenAndServe(":"+port, nil))
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

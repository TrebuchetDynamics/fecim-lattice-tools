package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// DocsHistory manages recent documents and favorites with thread-safe persistence
type DocsHistory struct {
	Recent     []string `json:"recent"`    // Last 10 viewed (LRU order)
	Favorites  []string `json:"favorites"` // Starred docs
	mu         sync.RWMutex
	configPath string
}

// NewDocsHistory loads or creates history from .omc/docs-history.json
func NewDocsHistory() *DocsHistory {
	h := &DocsHistory{
		Recent:     make([]string, 0),
		Favorites:  make([]string, 0),
		configPath: getHistoryPath(),
	}
	h.Load()
	return h
}

// AddRecent adds a document to recent list (front), removes duplicates, caps at 10
func (h *DocsHistory) AddRecent(path string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove path if it already exists
	filtered := make([]string, 0, len(h.Recent))
	for _, p := range h.Recent {
		if p != path {
			filtered = append(filtered, p)
		}
	}

	// Add to front
	h.Recent = append([]string{path}, filtered...)

	// Cap at 10
	if len(h.Recent) > 10 {
		h.Recent = h.Recent[:10]
	}

	// Save asynchronously
	go h.Save()
}

// GetRecent returns recent docs list
func (h *DocsHistory) GetRecent() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]string, len(h.Recent))
	copy(result, h.Recent)
	return result
}

// ToggleFavorite adds or removes from favorites
func (h *DocsHistory) ToggleFavorite(path string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if already favorited
	for i, p := range h.Favorites {
		if p == path {
			// Remove from favorites
			h.Favorites = append(h.Favorites[:i], h.Favorites[i+1:]...)
			go h.Save()
			return
		}
	}

	// Add to favorites
	h.Favorites = append(h.Favorites, path)
	go h.Save()
}

// IsFavorite checks if document is favorited
func (h *DocsHistory) IsFavorite(path string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, p := range h.Favorites {
		if p == path {
			return true
		}
	}
	return false
}

// GetFavorites returns favorites list
func (h *DocsHistory) GetFavorites() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]string, len(h.Favorites))
	copy(result, h.Favorites)
	return result
}

// Save persists to disk
func (h *DocsHistory) Save() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(h.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(h.configPath, data, 0644)
}

// Load reads from disk
func (h *DocsHistory) Load() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	data, err := os.ReadFile(h.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, start fresh
			return nil
		}
		return err
	}

	return json.Unmarshal(data, h)
}

// getHistoryPath returns the path to docs-history.json
func getHistoryPath() string {
	// Try .omc in current working directory
	if _, err := os.Stat(".omc"); err == nil {
		return filepath.Join(".omc", "docs-history.json")
	}
	// Create .omc if needed
	os.MkdirAll(".omc", 0755)
	return filepath.Join(".omc", "docs-history.json")
}

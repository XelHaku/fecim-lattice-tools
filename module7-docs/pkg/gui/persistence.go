//go:build legacy_fyne

package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	sharedio "fecim-lattice-tools/shared/io"
)

// DocsHistory manages recent documents and favorites with thread-safe persistence
type DocsHistory struct {
	Recent       []string `json:"recent"`    // Last 10 viewed (LRU order)
	Favorites    []string `json:"favorites"` // Starred docs
	favoritesMap map[string]bool
	mu           sync.RWMutex
	configPath   string
}

// NewDocsHistory loads or creates history from .omc/docs-history.json
func NewDocsHistory() *DocsHistory {
	h := &DocsHistory{
		Recent:       make([]string, 0),
		Favorites:    make([]string, 0),
		favoritesMap: make(map[string]bool),
		configPath:   getHistoryPath(),
	}
	h.Load()
	return h
}

func (h *DocsHistory) rebuildFavoritesMapLocked() {
	h.favoritesMap = make(map[string]bool, len(h.Favorites))
	unique := make([]string, 0, len(h.Favorites))
	for _, p := range h.Favorites {
		if p == "" {
			continue
		}
		if h.favoritesMap[p] {
			continue
		}
		h.favoritesMap[p] = true
		unique = append(unique, p)
	}
	h.Favorites = unique
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
	if h.favoritesMap == nil {
		h.rebuildFavoritesMapLocked()
	}

	if h.favoritesMap[path] {
		delete(h.favoritesMap, path)
		filtered := make([]string, 0, len(h.Favorites))
		for _, p := range h.Favorites {
			if p != path {
				filtered = append(filtered, p)
			}
		}
		h.Favorites = filtered
		h.mu.Unlock()
		go h.Save()
		return
	}

	h.favoritesMap[path] = true
	h.Favorites = append(h.Favorites, path)
	h.mu.Unlock()
	go h.Save()
}

// IsFavorite checks if document is favorited
func (h *DocsHistory) IsFavorite(path string) bool {
	h.mu.RLock()
	if h.favoritesMap == nil {
		h.mu.RUnlock()
		h.mu.Lock()
		h.rebuildFavoritesMapLocked()
		h.mu.Unlock()
		h.mu.RLock()
	}
	_, ok := h.favoritesMap[path]
	h.mu.RUnlock()
	return ok
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

	return sharedio.SaveJSON(h.configPath, h)
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

	type historyList struct {
		Recent    []string `json:"recent"`
		Favorites []string `json:"favorites"`
	}
	var list historyList
	listErr := json.Unmarshal(data, &list)
	if listErr == nil {
		h.Recent = list.Recent
		h.Favorites = list.Favorites
		h.rebuildFavoritesMapLocked()
		return nil
	}

	type historyMap struct {
		Recent    []string        `json:"recent"`
		Favorites map[string]bool `json:"favorites"`
	}
	var mapped historyMap
	mapErr := json.Unmarshal(data, &mapped)
	if mapErr == nil {
		h.Recent = mapped.Recent
		h.favoritesMap = make(map[string]bool, len(mapped.Favorites))
		h.Favorites = h.Favorites[:0]
		for path, fav := range mapped.Favorites {
			if fav {
				h.favoritesMap[path] = true
				h.Favorites = append(h.Favorites, path)
			}
		}
		return nil
	}

	return mapErr
}

// getHistoryPath returns the path to docs-history.json
func getHistoryPath() string {
	// Try .omc in current working directory
	if _, err := os.Stat(".omc"); err == nil {
		return filepath.Join(".omc", "docs-history.json")
	}
	// Create .omc if needed (error is non-fatal; caller handles write failure)
	_ = os.MkdirAll(".omc", 0755)
	return filepath.Join(".omc", "docs-history.json")
}

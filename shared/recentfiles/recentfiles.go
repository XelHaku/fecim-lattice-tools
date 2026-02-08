// Package recentfiles provides tracking of recently opened/saved files
// across FeCIM Lattice Tools sessions. Supports configs, exports, and projects.
package recentfiles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

// FileType categorizes recent files
type FileType string

const (
	FileTypeConfig  FileType = "config"  // Configuration/preset files (.json configs)
	FileTypeExport  FileType = "export"  // Exported data (.csv, .json, .png)
	FileTypeProject FileType = "project" // Project files or directories
	FileTypePreset  FileType = "preset"  // Preset files
	FileTypeAny     FileType = "any"     // Used for filtering - matches all types
)

// Preference keys for persistence
const (
	prefKeyRecentFiles = "recent_files_data"
	maxRecentFiles     = 20 // Maximum files to track per type
	maxTotalFiles      = 50 // Maximum total files across all types
)

// RecentFile represents a single recently accessed file
type RecentFile struct {
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Type       FileType  `json:"type"`
	Module     string    `json:"module,omitempty"` // Which module opened it (hysteresis, crossbar, etc.)
	AccessedAt time.Time `json:"accessed_at"`
	Size       int64     `json:"size,omitempty"` // File size in bytes (if known)
	Exists     bool      `json:"-"`              // Runtime check - not persisted
}

// Manager handles tracking and persistence of recent files
type Manager struct {
	mu       sync.RWMutex
	files    []*RecentFile
	prefs    fyne.Preferences
	onChange []func([]*RecentFile)
}

// persistedData is the format stored in preferences
type persistedData struct {
	Version int           `json:"version"`
	Files   []*RecentFile `json:"files"`
}

// NewManager creates a new recent files manager
// Pass nil for prefs to use in-memory only (for testing)
func NewManager(prefs fyne.Preferences) *Manager {
	m := &Manager{
		files: make([]*RecentFile, 0),
		prefs: prefs,
	}
	m.load()
	return m
}

// Add records a file access. If the file already exists, it updates the access time.
func (m *Manager) Add(path string, fileType FileType, module string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clean up path
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Get file info
	var size int64
	if info, err := os.Stat(absPath); err == nil {
		size = info.Size()
	}

	// Check if file already exists in list
	now := time.Now()
	found := false
	for _, f := range m.files {
		if f.Path == absPath {
			f.AccessedAt = now
			f.Module = module
			f.Size = size
			f.Type = fileType // Update type in case it changed
			found = true
			break
		}
	}

	if !found {
		// Add new entry
		m.files = append(m.files, &RecentFile{
			Path:       absPath,
			Name:       filepath.Base(absPath),
			Type:       fileType,
			Module:     module,
			AccessedAt: now,
			Size:       size,
		})
	}

	// Sort by access time (most recent first)
	sort.Slice(m.files, func(i, j int) bool {
		return m.files[i].AccessedAt.After(m.files[j].AccessedAt)
	})

	// Enforce limits
	m.enforceLimits()

	// Save to preferences
	m.saveNoLock()

	// Notify listeners
	m.notifyChange()
}

// AddConfig is a convenience method for adding config files
func (m *Manager) AddConfig(path, module string) {
	m.Add(path, FileTypeConfig, module)
}

// AddExport is a convenience method for adding export files
func (m *Manager) AddExport(path, module string) {
	m.Add(path, FileTypeExport, module)
}

// AddProject is a convenience method for adding project files
func (m *Manager) AddProject(path, module string) {
	m.Add(path, FileTypeProject, module)
}

// AddPreset is a convenience method for adding preset files
func (m *Manager) AddPreset(path, module string) {
	m.Add(path, FileTypePreset, module)
}

// Remove removes a file from the recent list
func (m *Manager) Remove(path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, _ := filepath.Abs(path)
	for i, f := range m.files {
		if f.Path == absPath {
			m.files = append(m.files[:i], m.files[i+1:]...)
			m.saveNoLock()
			m.notifyChange()
			return true
		}
	}
	return false
}

// Clear removes all recent files, optionally filtered by type
func (m *Manager) Clear(fileType FileType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if fileType == FileTypeAny {
		m.files = make([]*RecentFile, 0)
	} else {
		newFiles := make([]*RecentFile, 0)
		for _, f := range m.files {
			if f.Type != fileType {
				newFiles = append(newFiles, f)
			}
		}
		m.files = newFiles
	}

	m.saveNoLock()
	m.notifyChange()
}

// ClearAll removes all recent files
func (m *Manager) ClearAll() {
	m.Clear(FileTypeAny)
}

// List returns all recent files, optionally filtered by type
func (m *Manager) List(fileType FileType) []*RecentFile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*RecentFile, 0)
	for _, f := range m.files {
		if fileType == FileTypeAny || f.Type == fileType {
			// Check if file still exists
			exists := false
			if _, err := os.Stat(f.Path); err == nil {
				exists = true
			}
			entry := &RecentFile{
				Path:       f.Path,
				Name:       f.Name,
				Type:       f.Type,
				Module:     f.Module,
				AccessedAt: f.AccessedAt,
				Size:       f.Size,
				Exists:     exists,
			}
			result = append(result, entry)
		}
	}
	return result
}

// ListAll returns all recent files
func (m *Manager) ListAll() []*RecentFile {
	return m.List(FileTypeAny)
}

// ListConfigs returns recent config files
func (m *Manager) ListConfigs() []*RecentFile {
	return m.List(FileTypeConfig)
}

// ListExports returns recent export files
func (m *Manager) ListExports() []*RecentFile {
	return m.List(FileTypeExport)
}

// ListProjects returns recent project files
func (m *Manager) ListProjects() []*RecentFile {
	return m.List(FileTypeProject)
}

// ListPresets returns recent preset files
func (m *Manager) ListPresets() []*RecentFile {
	return m.List(FileTypePreset)
}

// ListByModule returns recent files for a specific module
func (m *Manager) ListByModule(module string) []*RecentFile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*RecentFile, 0)
	for _, f := range m.files {
		if f.Module == module {
			exists := false
			if _, err := os.Stat(f.Path); err == nil {
				exists = true
			}
			entry := &RecentFile{
				Path:       f.Path,
				Name:       f.Name,
				Type:       f.Type,
				Module:     f.Module,
				AccessedAt: f.AccessedAt,
				Size:       f.Size,
				Exists:     exists,
			}
			result = append(result, entry)
		}
	}
	return result
}

// GetMostRecent returns the most recently accessed file of a given type
func (m *Manager) GetMostRecent(fileType FileType) *RecentFile {
	files := m.List(fileType)
	if len(files) > 0 {
		return files[0]
	}
	return nil
}

// Count returns the number of recent files, optionally filtered by type
func (m *Manager) Count(fileType FileType) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if fileType == FileTypeAny {
		return len(m.files)
	}

	count := 0
	for _, f := range m.files {
		if f.Type == fileType {
			count++
		}
	}
	return count
}

// OnChange registers a callback for when the recent files list changes
func (m *Manager) OnChange(callback func([]*RecentFile)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onChange = append(m.onChange, callback)
}

// CleanupMissing removes files that no longer exist on disk
func (m *Manager) CleanupMissing() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	newFiles := make([]*RecentFile, 0)
	for _, f := range m.files {
		if _, err := os.Stat(f.Path); err == nil {
			newFiles = append(newFiles, f)
		} else {
			removed++
		}
	}
	m.files = newFiles

	if removed > 0 {
		m.saveNoLock()
		m.notifyChange()
	}

	return removed
}

// enforceLimits ensures we don't exceed maximum file counts
func (m *Manager) enforceLimits() {
	// Already sorted by access time (most recent first)

	// First, enforce per-type limits
	typeCounts := make(map[FileType]int)
	newFiles := make([]*RecentFile, 0)
	for _, f := range m.files {
		typeCounts[f.Type]++
		if typeCounts[f.Type] <= maxRecentFiles {
			newFiles = append(newFiles, f)
		}
	}
	m.files = newFiles

	// Then enforce total limit
	if len(m.files) > maxTotalFiles {
		m.files = m.files[:maxTotalFiles]
	}
}

// load reads recent files from preferences
func (m *Manager) load() {
	if m.prefs == nil {
		return
	}

	data := m.prefs.String(prefKeyRecentFiles)
	if data == "" {
		return
	}

	var persisted persistedData
	if err := json.Unmarshal([]byte(data), &persisted); err != nil {
		// Silently ignore corrupt data
		return
	}

	m.files = persisted.Files
	if m.files == nil {
		m.files = make([]*RecentFile, 0)
	}
}

// saveNoLock saves recent files to preferences (caller must hold lock)
func (m *Manager) saveNoLock() {
	if m.prefs == nil {
		return
	}

	persisted := persistedData{
		Version: 1,
		Files:   m.files,
	}

	data, err := json.Marshal(persisted)
	if err != nil {
		return
	}

	m.prefs.SetString(prefKeyRecentFiles, string(data))
}

// Save forces a save to preferences
func (m *Manager) Save() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveNoLock()
}

// notifyChange calls all registered change callbacks
func (m *Manager) notifyChange() {
	// Make a copy of files for callbacks
	files := make([]*RecentFile, len(m.files))
	copy(files, m.files)

	// Make a copy of callbacks (we're already holding the lock)
	callbacks := make([]func([]*RecentFile), len(m.onChange))
	copy(callbacks, m.onChange)

	// Release lock before calling callbacks
	go func() {
		for _, cb := range callbacks {
			cb(files)
		}
	}()
}

// FormatAccessTime returns a human-readable relative time string
func FormatAccessTime(t time.Time) string {
	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "Just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return formatDuration(mins, "minute")
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return formatDuration(hours, "hour")
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "Yesterday"
		}
		return formatDuration(days, "day")
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return formatDuration(weeks, "week")
	default:
		return t.Format("Jan 2, 2006")
	}
}

func formatDuration(n int, unit string) string {
	if n == 1 {
		return "1 " + unit + " ago"
	}
	return intToStrSmall(n) + " " + unit + "s ago"
}

func intToStrSmall(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		n = -n
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// FormatSize returns a human-readable file size
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes < KB:
		return fmtBytes(bytes, 1, "B")
	case bytes < MB:
		return fmtBytes(bytes, KB, "KB")
	case bytes < GB:
		return fmtBytes(bytes, MB, "MB")
	default:
		return fmtBytes(bytes, GB, "GB")
	}
}

func fmtBytes(bytes int64, divisor int64, unit string) string {
	if divisor == 1 {
		return intToStr(bytes) + " " + unit
	}
	whole := bytes / divisor
	frac := (bytes % divisor) * 10 / divisor
	if frac == 0 {
		return intToStr(whole) + " " + unit
	}
	return intToStr(whole) + "." + intToStr(frac) + " " + unit
}

func intToStr(n int64) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

package recentfiles

import (
	"sync"

	"fyne.io/fyne/v2"
)

var (
	globalManager *Manager
	globalMu      sync.Mutex
)

// SetGlobalManager sets the global recent files manager instance
func SetGlobalManager(m *Manager) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalManager = m
}

// GetGlobalManager returns the global recent files manager
// Returns nil if not initialized
func GetGlobalManager() *Manager {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalManager
}

// InitGlobal initializes the global manager with the given preferences
// This is a convenience function for app initialization
func InitGlobal(prefs fyne.Preferences) *Manager {
	m := NewManager(prefs)
	SetGlobalManager(m)
	return m
}

// GlobalAdd adds a file to the global recent files manager
// No-op if manager not initialized
func GlobalAdd(path string, fileType FileType, module string) {
	m := GetGlobalManager()
	if m != nil {
		m.Add(path, fileType, module)
	}
}

// GlobalAddConfig adds a config file to recent files
func GlobalAddConfig(path, module string) {
	GlobalAdd(path, FileTypeConfig, module)
}

// GlobalAddExport adds an export file to recent files
func GlobalAddExport(path, module string) {
	GlobalAdd(path, FileTypeExport, module)
}

// GlobalAddProject adds a project file to recent files
func GlobalAddProject(path, module string) {
	GlobalAdd(path, FileTypeProject, module)
}

// GlobalAddPreset adds a preset file to recent files
func GlobalAddPreset(path, module string) {
	GlobalAdd(path, FileTypePreset, module)
}

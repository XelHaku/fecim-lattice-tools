package presets

import (
	"sync"
)

var (
	globalManager     *Manager
	globalManagerOnce sync.Once
	globalManagerMu   sync.Mutex
)

// Global returns the global preset manager instance
func Global() *Manager {
	globalManagerOnce.Do(func() {
		globalManager = NewManager(DefaultPresetsDir())
	})
	return globalManager
}

// SetGlobal sets the global preset manager (for testing or custom initialization)
func SetGlobal(m *Manager) {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()
	globalManager = m
}

// InitGlobal initializes the global preset manager with a custom directory
func InitGlobal(presetsDir string) *Manager {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()

	// If already initialized with the same dir, return existing
	if globalManager != nil {
		return globalManager
	}

	globalManager = NewManager(presetsDir)
	return globalManager
}

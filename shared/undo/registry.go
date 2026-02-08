package undo

import (
	"sync"
)

// registry holds the global undo manager for the application.
var (
	globalManager *Manager
	registryMu    sync.RWMutex
)

// SetGlobalManager sets the global undo manager.
// This should be called once during application initialization.
func SetGlobalManager(m *Manager) {
	registryMu.Lock()
	defer registryMu.Unlock()
	globalManager = m
}

// GetGlobalManager returns the global undo manager.
// Returns nil if no manager has been set.
func GetGlobalManager() *Manager {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return globalManager
}

// Execute executes a command using the global manager.
// If no global manager is set, the command is executed without undo support.
func Execute(cmd Command) {
	m := GetGlobalManager()
	if m != nil {
		m.Execute(cmd)
	} else {
		// No undo manager, just execute the command directly
		cmd.Execute()
	}
}

// Undo undoes the last command using the global manager.
// Returns false if no manager is set or undo stack is empty.
func Undo() bool {
	m := GetGlobalManager()
	if m != nil {
		return m.Undo()
	}
	return false
}

// Redo redoes the last undone command using the global manager.
// Returns false if no manager is set or redo stack is empty.
func Redo() bool {
	m := GetGlobalManager()
	if m != nil {
		return m.Redo()
	}
	return false
}

// CanUndo returns true if there are commands to undo.
func CanUndo() bool {
	m := GetGlobalManager()
	if m != nil {
		return m.CanUndo()
	}
	return false
}

// CanRedo returns true if there are commands to redo.
func CanRedo() bool {
	m := GetGlobalManager()
	if m != nil {
		return m.CanRedo()
	}
	return false
}

// BeginGroup starts a command group using the global manager.
func BeginGroup(description string) {
	m := GetGlobalManager()
	if m != nil {
		m.BeginGroup(description)
	}
}

// EndGroup ends a command group using the global manager.
func EndGroup() {
	m := GetGlobalManager()
	if m != nil {
		m.EndGroup()
	}
}

// Clear clears all undo/redo history using the global manager.
func Clear() {
	m := GetGlobalManager()
	if m != nil {
		m.Clear()
	}
}

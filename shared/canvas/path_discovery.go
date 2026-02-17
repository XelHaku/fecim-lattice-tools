// Package utils provides common utility functions for FeCIM tools
package utils

import (
	"os"
	"path/filepath"
)

// FindDirectory searches for a directory by name in common locations.
// It checks paths relative to the working directory and the executable.
// Returns the absolute path if found, empty string otherwise.
func FindDirectory(dirName string) string {
	candidates := []string{
		dirName,
		filepath.Join("..", dirName),
		filepath.Join("..", "..", dirName),
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, dirName),
			filepath.Join(exeDir, "..", dirName),
			filepath.Join(exeDir, "..", "..", dirName),
		)
	}

	for _, candidate := range candidates {
		if abs, err := filepath.Abs(candidate); err == nil {
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				return abs
			}
		}
	}

	return ""
}

// FindDirectoryWithMarker searches for a directory containing a specific marker file.
// This is useful when you need to verify the directory contains expected content.
// For example: FindDirectoryWithMarker("data", "pretrained_weights.json")
// Returns the directory path (not absolute) if found, empty string otherwise.
func FindDirectoryWithMarker(dirName, markerFile string) string {
	candidates := []string{
		dirName,
		filepath.Join("..", dirName),
		filepath.Join("..", "..", dirName),
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, dirName),
			filepath.Join(exeDir, "..", dirName),
			filepath.Join(exeDir, "..", "..", dirName),
		)
	}

	// First pass: look for directory with marker file
	for _, candidate := range candidates {
		markerPath := filepath.Join(candidate, markerFile)
		if _, err := os.Stat(markerPath); err == nil {
			return candidate
		}
	}

	return ""
}

// FindModuleDataDir locates a module's data directory.
// It searches for module-specific paths and verifies with an optional marker file.
// Example: FindModuleDataDir("module3-mnist", "pretrained_weights.json")
func FindModuleDataDir(moduleName, markerFile string) string {
	candidates := []string{
		filepath.Join(moduleName, "data"),
		"data",
		filepath.Join("..", "data"),
		filepath.Join("..", "..", moduleName, "data"),
		filepath.Join("..", moduleName, "data"),
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, moduleName, "data"),
			filepath.Join(exeDir, "data"),
			filepath.Join(exeDir, "..", moduleName, "data"),
		)
	}

	// If marker file specified, look for it
	if markerFile != "" {
		for _, candidate := range candidates {
			markerPath := filepath.Join(candidate, markerFile)
			if _, err := os.Stat(markerPath); err == nil {
				return candidate
			}
		}
	}

	// Fallback: return first existing directory
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return ""
}

// FindFile searches for a file by name in common locations.
// Returns the absolute path if found, empty string otherwise.
func FindFile(fileName string) string {
	candidates := []string{
		fileName,
		filepath.Join("..", fileName),
		filepath.Join("..", "..", fileName),
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, fileName),
			filepath.Join(exeDir, "..", fileName),
			filepath.Join(exeDir, "..", "..", fileName),
		)
	}

	for _, candidate := range candidates {
		if abs, err := filepath.Abs(candidate); err == nil {
			if info, err := os.Stat(abs); err == nil && !info.IsDir() {
				return abs
			}
		}
	}

	return ""
}

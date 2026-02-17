package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDirectory(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Save current working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to temp directory for test
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer os.Chdir(origWd)

	// Test finding the directory
	result := FindDirectory("testdir")
	if result == "" {
		t.Error("FindDirectory should find testdir")
	}

	// Test non-existent directory
	result = FindDirectory("nonexistent")
	if result != "" {
		t.Errorf("FindDirectory should return empty for nonexistent, got: %s", result)
	}
}

func TestFindDirectoryWithMarker(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create data directory: %v", err)
	}

	// Create a marker file
	markerPath := filepath.Join(dataDir, "marker.json")
	if err := os.WriteFile(markerPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create marker file: %v", err)
	}

	// Save and change working directory
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Test finding directory with marker
	result := FindDirectoryWithMarker("data", "marker.json")
	if result == "" {
		t.Error("FindDirectoryWithMarker should find data with marker.json")
	}

	// Test with wrong marker
	result = FindDirectoryWithMarker("data", "wrong.json")
	if result != "" {
		t.Errorf("FindDirectoryWithMarker should return empty for wrong marker, got: %s", result)
	}
}

func TestFindModuleDataDir(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	moduleDataDir := filepath.Join(tmpDir, "module-test", "data")
	if err := os.MkdirAll(moduleDataDir, 0755); err != nil {
		t.Fatalf("failed to create module data directory: %v", err)
	}

	// Create a marker file
	markerPath := filepath.Join(moduleDataDir, "weights.json")
	if err := os.WriteFile(markerPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create marker file: %v", err)
	}

	// Save and change working directory
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Test finding module data directory with marker
	result := FindModuleDataDir("module-test", "weights.json")
	if result == "" {
		t.Error("FindModuleDataDir should find module-test/data with weights.json")
	}

	// Test without marker (should still find directory)
	result = FindModuleDataDir("module-test", "")
	if result == "" {
		t.Error("FindModuleDataDir should find module-test/data without marker")
	}

	// Test non-existent module
	result = FindModuleDataDir("nonexistent", "weights.json")
	if result != "" {
		t.Errorf("FindModuleDataDir should return empty for nonexistent module, got: %s", result)
	}
}

func TestFindFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Save and change working directory
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Test finding the file
	result := FindFile("testfile.txt")
	if result == "" {
		t.Error("FindFile should find testfile.txt")
	}

	// Test non-existent file
	result = FindFile("nonexistent.txt")
	if result != "" {
		t.Errorf("FindFile should return empty for nonexistent file, got: %s", result)
	}
}

func TestFindDirectoryReturnsAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "abstest")
	os.MkdirAll(testDir, 0755)

	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	result := FindDirectory("abstest")
	if result == "" {
		t.Fatal("FindDirectory should find abstest")
	}

	if !filepath.IsAbs(result) {
		t.Errorf("FindDirectory should return absolute path, got: %s", result)
	}
}

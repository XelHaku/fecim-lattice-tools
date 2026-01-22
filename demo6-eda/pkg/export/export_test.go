// pkg/export/export_test.go
package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"demo6-eda/pkg/compiler"
)

func getTestMapping() *compiler.CrossbarMapping {
	weights := [][]float64{
		{0.1, -0.2, 0.3},
		{-0.4, 0.5, -0.6},
	}
	config := compiler.DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8
	mapping, _ := compiler.Compile(weights, config)
	return mapping
}

func TestExportJSON(t *testing.T) {
	mapping := getTestMapping()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.json")

	err := ExportJSON(mapping, path)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Exported JSON file is empty")
	}

	// Check it contains expected fields
	content := string(data)
	if !strings.Contains(content, "cells") {
		t.Error("JSON should contain 'cells' field")
	}
	if !strings.Contains(content, "stats") {
		t.Error("JSON should contain 'stats' field")
	}

	t.Logf("JSON export: %d bytes", len(data))
}

func TestExportCSV(t *testing.T) {
	mapping := getTestMapping()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.csv")

	err := ExportCSV(mapping, path)
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	content := string(data)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	// Should have header + 6 data rows
	expectedRows := 7 // 1 header + 6 cells
	if len(lines) != expectedRows {
		t.Errorf("Expected %d rows, got %d", expectedRows, len(lines))
	}

	// Check header
	if !strings.Contains(lines[0], "row,col,weight") {
		t.Error("CSV header missing expected columns")
	}

	t.Logf("CSV export: %d rows", len(lines))
}

func TestExportSPICE(t *testing.T) {
	mapping := getTestMapping()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.sp")

	err := ExportSPICE(mapping, path, 1.8)
	if err != nil {
		t.Fatalf("ExportSPICE failed: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	content := string(data)

	// Check SPICE structure
	if !strings.Contains(content, ".param VDD") {
		t.Error("SPICE should contain VDD parameter")
	}
	if !strings.Contains(content, "Word Line Drivers") {
		t.Error("SPICE should contain word line drivers")
	}
	if !strings.Contains(content, "FeFET Cells") {
		t.Error("SPICE should contain FeFET cells")
	}
	if !strings.Contains(content, ".end") {
		t.Error("SPICE should end with .end")
	}

	t.Logf("SPICE export: %d bytes", len(data))
}

func TestGenerateSPICE(t *testing.T) {
	mapping := getTestMapping()
	spice := GenerateSPICE(mapping, 1.8)

	// Should generate valid SPICE
	if !strings.HasPrefix(spice, "*") {
		t.Error("SPICE should start with comment")
	}
	if !strings.HasSuffix(strings.TrimSpace(spice), ".end") {
		t.Error("SPICE should end with .end")
	}

	// Count resistor elements (should be 6 for 2x3 matrix)
	resistorCount := strings.Count(spice, "R_")
	if resistorCount != 6 {
		t.Errorf("Expected 6 resistors, got %d", resistorCount)
	}
}

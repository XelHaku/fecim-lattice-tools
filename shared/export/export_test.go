package export

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExporter_ExportCSV(t *testing.T) {
	tmpDir := t.TempDir()
	exporter := NewExporter(tmpDir, "test")

	headers := []string{"x", "y", "value"}
	rows := [][]string{
		{"1", "2", "3.14"},
		{"4", "5", "6.28"},
	}

	result := exporter.ExportCSV(headers, rows)
	if result.Error != nil {
		t.Fatalf("ExportCSV failed: %v", result.Error)
	}

	if result.Format != FormatCSV {
		t.Errorf("Expected format CSV, got %s", result.Format)
	}

	if !strings.HasSuffix(result.FilePath, ".csv") {
		t.Errorf("Expected .csv extension, got %s", result.FilePath)
	}

	// Verify file exists and has content
	content, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	if !strings.Contains(string(content), "x,y,value") {
		t.Error("CSV header not found in output")
	}

	if !strings.Contains(string(content), "1,2,3.14") {
		t.Error("CSV row data not found in output")
	}
}

func TestExporter_ExportCSVFromFloats(t *testing.T) {
	tmpDir := t.TempDir()
	exporter := NewExporter(tmpDir, "floats")

	headers := []string{"x", "y"}
	col1 := []float64{1.5, 2.5, 3.5}
	col2 := []float64{10.0, 20.0, 30.0}

	result := exporter.ExportCSVFromFloats(headers, col1, col2)
	if result.Error != nil {
		t.Fatalf("ExportCSVFromFloats failed: %v", result.Error)
	}

	content, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	if !strings.Contains(string(content), "1.5") {
		t.Error("Float value not found in output")
	}
}

func TestExporter_ExportJSON(t *testing.T) {
	tmpDir := t.TempDir()
	exporter := NewExporter(tmpDir, "config")

	data := map[string]interface{}{
		"name":    "test",
		"version": 1.0,
		"params": map[string]int{
			"rows": 64,
			"cols": 64,
		},
	}

	result := exporter.ExportJSON(data)
	if result.Error != nil {
		t.Fatalf("ExportJSON failed: %v", result.Error)
	}

	if result.Format != FormatJSON {
		t.Errorf("Expected format JSON, got %s", result.Format)
	}

	if !strings.HasSuffix(result.FilePath, ".json") {
		t.Errorf("Expected .json extension, got %s", result.FilePath)
	}

	content, err := os.ReadFile(result.FilePath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	if !strings.Contains(string(content), `"name": "test"`) {
		t.Error("JSON data not found in output")
	}
}

func TestExporter_ExportPNG(t *testing.T) {
	tmpDir := t.TempDir()
	exporter := NewExporter(tmpDir, "image")

	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}

	result := exporter.ExportPNG(img)
	if result.Error != nil {
		t.Fatalf("ExportPNG failed: %v", result.Error)
	}

	if result.Format != FormatPNG {
		t.Errorf("Expected format PNG, got %s", result.Format)
	}

	if !strings.HasSuffix(result.FilePath, ".png") {
		t.Errorf("Expected .png extension, got %s", result.FilePath)
	}

	// Verify file exists
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		t.Error("PNG file was not created")
	}
}

func TestCSVData(t *testing.T) {
	data := NewCSVData("A", "B", "C")
	data.AddRow("1", "2", "3")
	data.AddRowFromFloats(1.5, 2.5, 3.5)
	data.AddRowFromInts(10, 20, 30)

	if len(data.Headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(data.Headers))
	}

	if len(data.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(data.Rows))
	}

	if data.Rows[0][0] != "1" {
		t.Errorf("Expected '1', got '%s'", data.Rows[0][0])
	}

	if data.Rows[1][0] != "1.5" {
		t.Errorf("Expected '1.5', got '%s'", data.Rows[1][0])
	}

	if data.Rows[2][0] != "10" {
		t.Errorf("Expected '10', got '%s'", data.Rows[2][0])
	}
}

func TestQuickExport_JSON(t *testing.T) {
	tmpDir := t.TempDir()

	data := map[string]string{"test": "value"}
	result := QuickExport(tmpDir, "quick", FormatJSON, data)

	if result.Error != nil {
		t.Fatalf("QuickExport failed: %v", result.Error)
	}

	if !strings.HasSuffix(result.FilePath, ".json") {
		t.Errorf("Expected .json extension, got %s", result.FilePath)
	}
}

func TestQuickExport_CSV(t *testing.T) {
	tmpDir := t.TempDir()

	data := NewCSVData("col1", "col2")
	data.AddRow("a", "b")

	result := QuickExport(tmpDir, "quick", FormatCSV, data)

	if result.Error != nil {
		t.Fatalf("QuickExport failed: %v", result.Error)
	}

	if !strings.HasSuffix(result.FilePath, ".csv") {
		t.Errorf("Expected .csv extension, got %s", result.FilePath)
	}
}

func TestExportMetadata(t *testing.T) {
	meta := NewExportMetadata("test-module")

	if meta.ModuleName != "test-module" {
		t.Errorf("Expected module name 'test-module', got '%s'", meta.ModuleName)
	}

	if meta.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", meta.Version)
	}

	if meta.CustomFields == nil {
		t.Error("CustomFields should not be nil")
	}
}

func TestExporter_EnsuresOutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "deep", "dir")
	exporter := NewExporter(nestedDir, "test")

	data := map[string]string{"test": "value"}
	result := exporter.ExportJSON(data)

	if result.Error != nil {
		t.Fatalf("Export failed: %v", result.Error)
	}

	// Verify nested directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("Nested directory was not created")
	}
}

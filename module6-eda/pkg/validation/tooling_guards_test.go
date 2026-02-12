package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/openlane"
)

func TestValidateVerilog_MissingFile(t *testing.T) {
	err := ValidateVerilog("/does/not/exist.v")
	if err == nil {
		t.Fatal("expected missing file error")
	}
	if !strings.Contains(err.Error(), "verilog file not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateVerilogWithCell_MissingArrayFile(t *testing.T) {
	err := ValidateVerilogWithCell("/does/not/exist_array.v", "/does/not/exist_cell.v")
	if err == nil {
		t.Fatal("expected missing array file error")
	}
	if !strings.Contains(err.Error(), "array verilog not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateVerilogWithCell_MissingCellFile(t *testing.T) {
	tmpDir := t.TempDir()
	arrayPath := filepath.Join(tmpDir, "array.v")
	if err := os.WriteFile(arrayPath, []byte("module arr; endmodule\n"), 0644); err != nil {
		t.Fatalf("write array file: %v", err)
	}

	err := ValidateVerilogWithCell(arrayPath, "/does/not/exist_cell.v")
	if err == nil {
		t.Fatal("expected missing cell file error")
	}
	if !strings.Contains(err.Error(), "cell verilog not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateLayoutImage_MissingDEF(t *testing.T) {
	tmpDir := t.TempDir()
	lefPath := filepath.Join(tmpDir, "cell.lef")
	outputPath := filepath.Join(tmpDir, "layout.png")
	if err := os.WriteFile(lefPath, []byte("MACRO dummy\nEND dummy\n"), 0644); err != nil {
		t.Fatalf("write lef file: %v", err)
	}

	result, err := GenerateLayoutImage("/does/not/exist.def", lefPath, outputPath, nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Success {
		t.Fatal("expected failure result")
	}
	if !strings.Contains(result.Error, "DEF file not found") {
		t.Fatalf("unexpected result error: %q", result.Error)
	}
}

func TestGenerateLayoutImage_MissingLEF(t *testing.T) {
	tmpDir := t.TempDir()
	defPath := filepath.Join(tmpDir, "design.def")
	outputPath := filepath.Join(tmpDir, "layout.png")
	if err := os.WriteFile(defPath, []byte("VERSION 5.8 ;\nEND DESIGN\n"), 0644); err != nil {
		t.Fatalf("write def file: %v", err)
	}

	result, err := GenerateLayoutImage(defPath, "/does/not/exist.lef", outputPath, nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Success {
		t.Fatal("expected failure result")
	}
	if !strings.Contains(result.Error, "LEF file not found") {
		t.Fatalf("unexpected result error: %q", result.Error)
	}
}

func TestRunCellUsageReport_MissingDEF(t *testing.T) {
	result, err := RunCellUsageReport("/does/not/exist.def", nil, nil)
	if err == nil {
		t.Fatal("expected error for missing DEF")
	}
	if result != nil {
		t.Fatal("expected nil result")
	}
	if !strings.Contains(err.Error(), "DEF file not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsToolAvailabilityHelpers_AgreeWithDetectMode(t *testing.T) {
	manager := openlane.NewManager()
	expected := manager.DetectMode() != openlane.ModeNone

	if got := IsOpenROADAvailable(manager); got != expected {
		t.Fatalf("IsOpenROADAvailable()=%v, expected %v", got, expected)
	}
	if got := IsKLayoutAvailable(manager); got != expected {
		t.Fatalf("IsKLayoutAvailable()=%v, expected %v", got, expected)
	}
}

package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateYosysSchematic_MissingVerilog(t *testing.T) {
	result, err := GenerateYosysSchematic("/does/not/exist.v", "/tmp/out", "top", "passive", nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Success {
		t.Fatal("expected unsuccessful result")
	}
	if !strings.Contains(result.Error, "Verilog file not found") {
		t.Fatalf("unexpected error text: %q", result.Error)
	}
}

func TestGenerateOpenROADImage_MissingDEF(t *testing.T) {
	tmpDir := t.TempDir()
	lefPath := filepath.Join(tmpDir, "cell.lef")
	if err := os.WriteFile(lefPath, []byte("MACRO dummy\nEND dummy\n"), 0644); err != nil {
		t.Fatalf("write LEF: %v", err)
	}

	result, err := GenerateOpenROADImage("/does/not/exist.def", lefPath, filepath.Join(tmpDir, "out.png"), nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Success {
		t.Fatal("expected unsuccessful result")
	}
	if !strings.Contains(result.Error, "DEF file not found") {
		t.Fatalf("unexpected error text: %q", result.Error)
	}
}

func TestGenerateOpenROADImage_MissingLEF(t *testing.T) {
	tmpDir := t.TempDir()
	defPath := filepath.Join(tmpDir, "design.def")
	if err := os.WriteFile(defPath, []byte("VERSION 5.8 ;\nEND DESIGN\n"), 0644); err != nil {
		t.Fatalf("write DEF: %v", err)
	}

	result, err := GenerateOpenROADImage(defPath, "/does/not/exist.lef", filepath.Join(tmpDir, "out.png"), nil, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Success {
		t.Fatal("expected unsuccessful result")
	}
	if !strings.Contains(result.Error, "LEF file not found") {
		t.Fatalf("unexpected error text: %q", result.Error)
	}
}

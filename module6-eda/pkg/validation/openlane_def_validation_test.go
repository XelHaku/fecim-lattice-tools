package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateDEF_InvalidComponentsDeclarationFormat(t *testing.T) {
	tmpDir := t.TempDir()
	defPath := filepath.Join(tmpDir, "bad_components.def")

	content := `VERSION 5.8 ;
DESIGN test ;
UNITS DISTANCE MICRONS 1000 ;
DIEAREA ( 0 0 ) ( 1000 1000 ) ;

COMPONENTS two ;
    - cell_0 fecim_bitcell + FIXED ( 0 0 ) N ;
END COMPONENTS

END DESIGN
`
	if err := os.WriteFile(defPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test DEF: %v", err)
	}

	err := ValidateDEF(defPath)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "invalid COMPONENTS declaration format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetDEFStats_MissingFile(t *testing.T) {
	_, err := GetDEFStats("/does/not/exist.def")
	if err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestDetectArchitectureFromDEF_FromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{name: "passive fallback", content: "MACRO fecim_bitcell", expected: "passive"},
		{name: "1t1r", content: "- u0 fecim_1t1r + FIXED ( 0 0 ) N ;", expected: "1t1r"},
		{name: "2t1r", content: "- u0 fecim_2t1r + FIXED ( 0 0 ) N ;", expected: "2t1r"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			defPath := filepath.Join(tmpDir, "arch.def")
			if err := os.WriteFile(defPath, []byte(tc.content), 0644); err != nil {
				t.Fatalf("write failed: %v", err)
			}

			got := detectArchitectureFromDEF(defPath)
			if got != tc.expected {
				t.Fatalf("detectArchitectureFromDEF() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestDetectArchitectureFromDEF_MissingFileDefaultsToPassive(t *testing.T) {
	if got := detectArchitectureFromDEF("/does/not/exist.def"); got != "passive" {
		t.Fatalf("got %q, want passive", got)
	}
}

func TestParsePlacementOutput_DetectsViolations(t *testing.T) {
	output := `
Info: running check_placement
WARNING overlap at cell u1
ERROR placement failed for region core
Found 2 unplaced instances
`
	violations, count := parsePlacementOutput(output)
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}
	if len(violations) != 3 {
		t.Fatalf("len(violations) = %d, want 3", len(violations))
	}

	issues := map[string]bool{}
	for _, v := range violations {
		issues[v.Issue] = true
	}

	for _, issue := range []string{"overlap", "placement_error", "unplaced"} {
		if !issues[issue] {
			t.Fatalf("missing issue type %q in %+v", issue, violations)
		}
	}
}

func TestParseCellUsageOutput_ParsesSectionAndMetrics(t *testing.T) {
	output := `
=== CELL USAGE ===
fecim_1t1r 10
filler 2
=== END ===
Total 12
Utilization: 76.5%
`

	result := parseCellUsageOutput(output)
	if result.TotalCells != 12 {
		t.Fatalf("TotalCells = %d, want 12", result.TotalCells)
	}
	if result.CellTypes["fecim_1t1r"] != 10 {
		t.Fatalf("fecim_1t1r count = %d, want 10", result.CellTypes["fecim_1t1r"])
	}
	if result.CellTypes["filler"] != 2 {
		t.Fatalf("filler count = %d, want 2", result.CellTypes["filler"])
	}
	if result.UtilizationPct != 76.5 {
		t.Fatalf("UtilizationPct = %v, want 76.5", result.UtilizationPct)
	}
}

func TestRunPlacementCheckWithCell_DEFNotFound(t *testing.T) {
	result, err := RunPlacementCheckWithCell("/does/not/exist.def", "cells/fecim_bitcell/fecim_bitcell.lef", nil, nil)
	if err == nil {
		t.Fatal("expected error for missing DEF")
	}
	if result != nil {
		t.Fatal("expected nil result on DEF stat failure")
	}
	if !strings.Contains(err.Error(), "DEF file not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPlacementCheck_DEFNotFound(t *testing.T) {
	result, err := RunPlacementCheck("/does/not/exist.def", nil, nil)
	if err == nil {
		t.Fatal("expected error for missing DEF")
	}
	if result != nil {
		t.Fatal("expected nil result on missing DEF")
	}
	if !strings.Contains(err.Error(), "DEF file not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

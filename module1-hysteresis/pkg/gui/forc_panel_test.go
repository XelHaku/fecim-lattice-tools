package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFORCGridDimensionsMatchExpected(t *testing.T) {
	out, err := runFORCWorkflow(0.25, 1.0, 17, 23)
	if err != nil {
		t.Fatalf("runFORCWorkflow error: %v", err)
	}
	if out.Matrix.Rows != 17 {
		t.Fatalf("rows=%d want 17", out.Matrix.Rows)
	}
	if out.Matrix.Cols != 23 {
		t.Fatalf("cols=%d want 23", out.Matrix.Cols)
	}
}

func TestFORCDerivativePipelineNonZeroNearEc(t *testing.T) {
	out, err := runFORCWorkflow(0.25, 1.0, 31, 31)
	if err != nil {
		t.Fatalf("runFORCWorkflow error: %v", err)
	}
	if !strings.Contains(out.Stats, "peak_density=") {
		t.Fatalf("missing peak density stats: %s", out.Stats)
	}
	if strings.Contains(out.Stats, "peak_density=0.000000e+00") {
		t.Fatalf("expected non-zero density near Ec, got %s", out.Stats)
	}
}

func TestFORCExportCSVHeadersAndRowCount(t *testing.T) {
	out, err := runFORCWorkflow(0.25, 1.0, 11, 13)
	if err != nil {
		t.Fatalf("runFORCWorkflow error: %v", err)
	}
	dir := t.TempDir()
	sweepPath := filepath.Join(dir, "sweep.csv")
	if err := ExportFORCSweepCSV(out.Sweep, sweepPath); err != nil {
		t.Fatalf("ExportFORCSweepCSV error: %v", err)
	}
	b, err := os.ReadFile(sweepPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if lines[0] != "reversal_field,E,P" {
		t.Fatalf("bad header: %q", lines[0])
	}
	wantRows := 1 + (11 * 12 / 2)
	if len(lines) != wantRows {
		t.Fatalf("row count=%d want %d", len(lines), wantRows)
	}

	matrixPath := filepath.Join(dir, "matrix.csv")
	if err := ExportFORCMatrixCSV(out.Matrix, matrixPath); err != nil {
		t.Fatalf("ExportFORCMatrixCSV error: %v", err)
	}
	mb, _ := os.ReadFile(matrixPath)
	if !strings.HasPrefix(string(mb), "Ea,Eb,density") {
		t.Fatalf("matrix header missing")
	}
}

func TestFORCMetadataJSONSchemaValid(t *testing.T) {
	dir := t.TempDir()
	metaPath := filepath.Join(dir, "meta.json")
	params := map[string]any{"Emax": 1.0, "numReversals": 21, "resolution": 21}
	if err := ExportFORCMetadata("HZO", "Preisach", params, metaPath); err != nil {
		t.Fatalf("ExportFORCMetadata error: %v", err)
	}
	b, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	for _, k := range []string{"timestamp", "git_commit", "material", "engine", "params"} {
		if _, ok := got[k]; !ok {
			t.Fatalf("missing key %q", k)
		}
	}
}

func TestFORCPanelCreatesWithoutPanic(t *testing.T) {
	a := NewApp()
	if a == nil {
		t.Fatal("NewApp returned nil")
	}
	p := a.createFORCPanel()
	if p == nil {
		t.Fatal("FORC panel is nil")
	}
}

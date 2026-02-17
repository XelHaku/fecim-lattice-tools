package gui_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

// M2-GUI-03: If export exists, verify CSV output contains correct dimensions.
// ExportWeightsCSV is a long-format export: one record per cell plus header.
func TestM2GUI03_ExportWeightsCSV_HasCorrectDimensions(t *testing.T) {
	t.Parallel()

	const (
		rows = 12
		cols = 7
	)

	arr, err := crossbar.NewArray(&crossbar.Config{
		Rows:             rows,
		Cols:             cols,
		NoiseLevel:       0,
		ADCBits:          8,
		DACBits:          8,
		UseGPU:           false,
		ConductanceModel: crossbar.ConductanceLinear,
		Endurance:        crossbar.DefaultEnduranceConfig(),
		HalfSelect:       crossbar.DefaultHalfSelectConfig(),
	})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			arr.ProgramWeight(i, j, float64(i*cols+j)/float64(rows*cols))
		}
	}

	dir := t.TempDir()
	out := filepath.Join(dir, "weights.csv")
	if err := arr.ExportWeightsCSV(out); err != nil {
		t.Fatalf("ExportWeightsCSV: %v", err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("open exported CSV: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	recs, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read exported CSV: %v", err)
	}

	want := rows*cols + 1
	if len(recs) != want {
		t.Fatalf("CSV record count mismatch: got=%d want=%d", len(recs), want)
	}

	header := recs[0]
	if len(header) < 2 || header[0] != "row" || header[1] != "col" {
		t.Fatalf("unexpected CSV header: %v", header)
	}

	seen := make(map[[2]int]bool, rows*cols)
	for idx, rec := range recs[1:] {
		if len(rec) < 2 {
			t.Fatalf("record %d too short: %v", idx+1, rec)
		}
		ri, err := strconv.Atoi(rec[0])
		if err != nil {
			t.Fatalf("record %d invalid row: %q", idx+1, rec[0])
		}
		cj, err := strconv.Atoi(rec[1])
		if err != nil {
			t.Fatalf("record %d invalid col: %q", idx+1, rec[1])
		}
		if ri < 0 || ri >= rows || cj < 0 || cj >= cols {
			t.Fatalf("record %d out of bounds: row=%d col=%d (rows=%d cols=%d)", idx+1, ri, cj, rows, cols)
		}
		key := [2]int{ri, cj}
		if seen[key] {
			t.Fatalf("duplicate cell in export: row=%d col=%d", ri, cj)
		}
		seen[key] = true
	}
	if len(seen) != rows*cols {
		t.Fatalf("expected %d unique cells, got %d", rows*cols, len(seen))
	}
}

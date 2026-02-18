package export

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratePublicationsCSVHeaders(t *testing.T) {
	dir := t.TempDir()
	_, err := GeneratePublicationsCSV(dir, PublicationsData{
		PELoop:          []XYPoint{{X: -1, Y: -20}, {X: 1, Y: 20}},
		ISPPConvergence: []XYPoint{{X: 1, Y: 0.45}},
		ReadMarginSweep: []XYPoint{{X: 0.1, Y: 45}},
	})
	if err != nil {
		t.Fatalf("GeneratePublicationsCSV failed: %v", err)
	}

	expect := map[string][]string{
		"pe_loop.csv":           {"electric_field_mv_cm", "polarization_uc_cm2"},
		"ispp_convergence.csv":  {"pulse_index", "threshold_voltage_v"},
		"read_margin_sweep.csv": {"delta_vread_v", "read_margin_mv"},
	}

	for name, headers := range expect {
		file, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("open %s: %v", name, err)
		}
		r := csv.NewReader(file)
		row, err := r.Read()
		_ = file.Close()
		if err != nil {
			t.Fatalf("read header %s: %v", name, err)
		}
		if len(row) != len(headers) || row[0] != headers[0] || row[1] != headers[1] {
			t.Fatalf("header mismatch for %s: got=%v want=%v", name, row, headers)
		}
	}
}

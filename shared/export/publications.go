package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// XYPoint is figure-ready tabular data for publications exports.
type XYPoint struct {
	X float64
	Y float64
}

// PublicationsData bundles common physics result curves.
type PublicationsData struct {
	PELoop          []XYPoint
	ISPPConvergence []XYPoint
	ReadMarginSweep []XYPoint
}

// GeneratePublicationsCSV writes one CSV per physics result for manuscript-ready workflows.
// Files generated:
//   - pe_loop.csv
//   - ispp_convergence.csv
//   - read_margin_sweep.csv
func GeneratePublicationsCSV(outputDir string, data PublicationsData) ([]string, error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create publications output dir: %w", err)
	}

	files := []struct {
		name    string
		headers []string
		rows    []XYPoint
	}{
		{name: "pe_loop.csv", headers: []string{"electric_field_mv_cm", "polarization_uc_cm2"}, rows: data.PELoop},
		{name: "ispp_convergence.csv", headers: []string{"pulse_index", "threshold_voltage_v"}, rows: data.ISPPConvergence},
		{name: "read_margin_sweep.csv", headers: []string{"delta_vread_v", "read_margin_mv"}, rows: data.ReadMarginSweep},
	}

	written := make([]string, 0, len(files))
	for _, f := range files {
		path := filepath.Join(outputDir, f.name)
		if err := writeXYCSV(path, f.headers, f.rows); err != nil {
			return written, err
		}
		written = append(written, path)
	}
	return written, nil
}

func writeXYCSV(path string, headers []string, rows []XYPoint) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create csv %s: %w", path, err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write(headers); err != nil {
		return fmt.Errorf("write header %s: %w", path, err)
	}
	for _, p := range rows {
		record := []string{
			strconv.FormatFloat(p.X, 'f', -1, 64),
			strconv.FormatFloat(p.Y, 'f', -1, 64),
		}
		if err := w.Write(record); err != nil {
			return fmt.Errorf("write row %s: %w", path, err)
		}
	}
	return nil
}

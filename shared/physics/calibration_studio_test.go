package physics

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestCalibrationStudioFitKnownDataset_RMSEUnder10Percent(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "pe.csv")

	// Synthetic Preisach-like dataset.
	data := "E,P\n"
	for e := -3.0e8; e <= 3.0e8; e += 3.0e7 {
		p := 0.24 * math.Tanh((e-1.0e7)/7.5e7)
		data += fmt.Sprintf("%.6e,%.6e\n", e, p)
	}
	if err := os.WriteFile(csvPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	points, err := ImportCalibrationCSV(csvPath)
	if err != nil {
		t.Fatalf("ImportCalibrationCSV failed: %v", err)
	}
	bundle := FitCalibration(points, CalibrationModelPreisach, 3000, 42)
	if bundle.RelativeRMSE >= 0.10 {
		t.Fatalf("Relative RMSE=%.4f, want <0.10", bundle.RelativeRMSE)
	}
}

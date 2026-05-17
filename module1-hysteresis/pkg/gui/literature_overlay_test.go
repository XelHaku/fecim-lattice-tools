//go:build legacy_fyne

package gui

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestLoadLiteratureCSV_Valid(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "lit.csv")
	if err := os.WriteFile(p, []byte("E_MV_cm,P_uC_cm2\n-1.0,-10\n0,0\n1.0,10\n"), 0644); err != nil {
		t.Fatal(err)
	}
	ds, err := LoadLiteratureCSV(p)
	if err != nil {
		t.Fatalf("expected valid CSV, got error: %v", err)
	}
	if len(ds.EField) != 3 || len(ds.Polarization) != 3 {
		t.Fatalf("unexpected lengths E=%d P=%d", len(ds.EField), len(ds.Polarization))
	}
}

func TestLoadLiteratureCSV_Invalid(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bad.csv")
	if err := os.WriteFile(p, []byte("E_MV_cm,P_uC_cm2\nabc,1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadLiteratureCSV(p)
	if err == nil {
		t.Fatal("expected error for invalid numeric value")
	}
	if !strings.Contains(err.Error(), "malformed CSV row") {
		t.Fatalf("expected clear malformed CSV error, got: %v", err)
	}
}

func TestLoadLiteratureJSON_Valid(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "lit.json")
	body := `{"source":"Park et al. 2015","material":"HZO","E_V_m":[-1e8,0,1e8],"P_C_m2":[-0.2,0,0.2]}`
	if err := os.WriteFile(p, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	ds, err := LoadLiteratureJSON(p)
	if err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if ds.Source != "Park et al. 2015" || ds.Material != "HZO" {
		t.Fatalf("metadata parse failed: %#v", ds)
	}
}

func TestUnitNormalization(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "lit.csv")
	if err := os.WriteFile(p, []byte("E_MV_cm,P_uC_cm2\n1.25,20\n1.50,25\n"), 0644); err != nil {
		t.Fatal(err)
	}
	ds, err := LoadLiteratureCSV(p)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ds.EField[0], 1.25e8; math.Abs(got-want) > 1e-9 {
		t.Fatalf("E normalization mismatch: got=%g want=%g", got, want)
	}
	if got, want := ds.Polarization[0], 0.20; math.Abs(got-want) > 1e-12 {
		t.Fatalf("P normalization mismatch: got=%g want=%g", got, want)
	}
}

func TestFitMetrics_KnownData(t *testing.T) {
	sim := &LiteratureDataset{EField: []float64{0, 10}, Polarization: []float64{0, 10}}
	lit := &LiteratureDataset{EField: []float64{0, 5, 10}, Polarization: []float64{0, 6, 10}}
	m := ComputeFitMetrics(sim, lit)
	if m.NSamples != 3 {
		t.Fatalf("NSamples mismatch: got=%d want=3", m.NSamples)
	}
	if math.Abs(m.RMSE-math.Sqrt(1.0/3.0)) > 1e-12 {
		t.Fatalf("RMSE mismatch: got=%.12f", m.RMSE)
	}
	if math.Abs(m.MAE-(1.0/3.0)) > 1e-12 {
		t.Fatalf("MAE mismatch: got=%.12f", m.MAE)
	}
	if math.Abs(m.MaxErr-1.0) > 1e-12 {
		t.Fatalf("MaxErr mismatch: got=%.12f", m.MaxErr)
	}
}

func TestOverlayPanel_NoPanic(t *testing.T) {
	test.NewApp()
	a := NewApp()
	_ = a.createLiteratureOverlayPanel()
	a.SetLiteratureDataset(&LiteratureDataset{
		Source:       "Park et al. 2015",
		Material:     "HZO",
		EField:       []float64{-1e8, 0, 1e8},
		Polarization: []float64{-0.2, 0, 0.2},
		Units:        DataUnits{EFieldUnit: "V/m", PolarizationUnit: "C/m2"},
	})
}

//go:build legacy_fyne

package gui

import (
	"strings"
	"testing"
)

func TestComputeSneakPathMetrics(t *testing.T) {
	currents := [][]float64{
		{1e-6, 2e-6, 0},
		{3e-6, 9e-6, 4e-6},
		{0, 5e-6, 6e-6},
	}
	// target at (1,1) is excluded.
	metrics := computeSneakPathMetrics(currents, 1, 1)

	if metrics.AffectedCells != 6 {
		t.Fatalf("affected cells: got %d want 6", metrics.AffectedCells)
	}
	if metrics.HalfSelectCells != 4 {
		t.Fatalf("half-select cells: got %d want 4", metrics.HalfSelectCells)
	}
	if metrics.SneakOnlyCells != 2 {
		t.Fatalf("sneak-only cells: got %d want 2", metrics.SneakOnlyCells)
	}
	if got := metrics.TotalSneakCurrentA; got < 20.9e-6 || got > 21.1e-6 {
		t.Fatalf("total sneak current: got %.9g A want 21e-6 A", got)
	}
	if len(metrics.TopAffectedCells) != 3 {
		t.Fatalf("top cells len: got %d want 3", len(metrics.TopAffectedCells))
	}
	if metrics.TopAffectedCells[0].Row != 2 || metrics.TopAffectedCells[0].Col != 2 {
		t.Fatalf("top cell: got (%d,%d) want (2,2)", metrics.TopAffectedCells[0].Row, metrics.TopAffectedCells[0].Col)
	}
}

func TestFormatSneakPathSummary(t *testing.T) {
	metrics := SneakPathMetrics{
		TotalSneakCurrentA: 2.3e-6,
		AffectedCells:      5,
		HalfSelectCells:    3,
		SneakOnlyCells:     2,
		TopAffectedCells:   []SneakCellImpact{{Row: 4, Col: 7, CurrentA: 9e-7}},
	}
	got := formatSneakPathSummary(metrics)
	if got == "" {
		t.Fatal("empty summary")
	}
	if want := "affected=5"; !strings.Contains(got, want) {
		t.Fatalf("summary %q missing %q", got, want)
	}
	if want := "[4,7]"; !strings.Contains(got, want) {
		t.Fatalf("summary %q missing %q", got, want)
	}
}

func TestComputeSneakPathMetricsEmpty(t *testing.T) {
	metrics := computeSneakPathMetrics(nil, 0, 0)
	if metrics.AffectedCells != 0 {
		t.Fatalf("expected 0 affected cells for nil input, got %d", metrics.AffectedCells)
	}
	summary := formatSneakPathSummary(metrics)
	if !strings.Contains(summary, "0 affected") {
		t.Fatalf("empty summary should mention 0 affected: %q", summary)
	}
}

func TestPeripheralSnapshotCSV(t *testing.T) {
	weights := [][]int{{5, 10}, {15, 20}}
	voltages := [][]float64{{0.3, 0.0}, {0.3, 0.15}}
	currents := [][]float64{{1e-6, 0}, {3e-6, 2e-6}}
	rows := buildPeripheralSnapshotRows(weights, voltages, currents, 0, 0)
	if len(rows) != 4 {
		t.Fatalf("expected 4 snapshot rows, got %d", len(rows))
	}
	if !rows[0].IsTargetCell {
		t.Fatal("row 0 should be target cell")
	}
	if !rows[1].IsHalfSelected {
		t.Fatal("(0,1) should be half-selected")
	}
	csv := peripheralSnapshotCSV(rows)
	if len(csv.Headers) != 7 {
		t.Fatalf("expected 7 headers, got %d", len(csv.Headers))
	}
	if len(csv.Rows) != 4 {
		t.Fatalf("expected 4 CSV rows, got %d", len(csv.Rows))
	}
}

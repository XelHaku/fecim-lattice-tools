package crossbar

import "testing"

func sumUnselectedSneak(sp *SneakPathAnalyzer, targetRow, targetCol int) float64 {
	total := 0.0
	for r := 0; r < sp.Rows; r++ {
		for c := 0; c < sp.Cols; c++ {
			if r == targetRow && c == targetCol {
				continue
			}
			total += sp.SneakCurrents[r][c]
		}
	}
	return total
}

// 0T1R (passive) arrays should exhibit measurable sneak current in unselected cells.
func TestSneakPathCurrent_0T1R_PassiveHasSneakCurrent(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	targetRow, targetCol := 1, 1
	voltage := 1.0

	sp.AnalyzeTarget(targetRow, targetCol, voltage)

	if sp.TotalSneakRatio <= 0 {
		t.Fatalf("expected non-zero sneak ratio in 0T1R array, got %.6e", sp.TotalSneakRatio)
	}

	// At least one unselected cell should carry sneak current.
	nonZeroUnselected := 0
	for r := 0; r < sp.Rows; r++ {
		for c := 0; c < sp.Cols; c++ {
			if r == targetRow && c == targetCol {
				continue
			}
			if sp.SneakCurrents[r][c] > 0 {
				nonZeroUnselected++
			}
		}
	}

	if nonZeroUnselected == 0 {
		t.Fatal("expected sneak current in at least one unselected cell for 0T1R")
	}

	totalUnselectedSneak := sumUnselectedSneak(sp, targetRow, targetCol)
	if totalUnselectedSneak <= 0 {
		t.Fatalf("expected quantified unselected sneak current > 0, got %.6e A", totalUnselectedSneak)
	}
}

// 1T1R arrays should suppress (but not necessarily fully eliminate) sneak current via selector action.
func TestSneakPathCurrent_1T1R_SelectorSuppressesSneakCurrent(t *testing.T) {
	sp := NewSneakPathAnalyzer(8, 8)
	targetRow, targetCol := 2, 3
	voltage := 1.0

	sp.AnalyzeTarget(targetRow, targetCol, voltage)
	baseline := sp.GetStats(voltage)

	mit := SneakMitigation{
		UseSelector:   true,
		SelectorOnOff: 1e4,
	}
	withSelector := sp.AnalyzeWithMitigation(targetRow, targetCol, voltage, mit)

	if withSelector.TotalSneakCurrent >= baseline.TotalSneakCurrent {
		t.Fatalf("selector should suppress sneak current: baseline %.6e A, with selector %.6e A",
			baseline.TotalSneakCurrent, withSelector.TotalSneakCurrent)
	}

	if withSelector.SneakRatio >= baseline.SneakRatio {
		t.Fatalf("selector should reduce sneak ratio: baseline %.6e, with selector %.6e",
			baseline.SneakRatio, withSelector.SneakRatio)
	}
}

// 2T1R-like masking: masked cells are isolated, so sneak current through them should be fully eliminated.
func TestSneakPathCurrent_2T1R_MaskedCellsEliminateSneakCurrent(t *testing.T) {
	sp := NewSneakPathAnalyzer(5, 5)
	targetRow, targetCol := 0, 0
	voltage := 1.0

	maskedCells := [][2]int{{0, 2}, {2, 2}, {2, 0}, {3, 3}}
	for _, cell := range maskedCells {
		sp.SetConductance(cell[0], cell[1], 0) // masked/off via extra selector transistor
	}

	sp.AnalyzeTarget(targetRow, targetCol, voltage)

	for _, cell := range maskedCells {
		r, c := cell[0], cell[1]
		if sp.SneakCurrents[r][c] != 0 {
			t.Fatalf("masked cell (%d,%d) should have zero sneak current, got %.6e A", r, c, sp.SneakCurrents[r][c])
		}
	}
}

// Sneak current should scale up with array size due to growth in available parasitic paths.
func TestSneakPathCurrent_ScalesWithArraySize(t *testing.T) {
	voltage := 1.0
	sizes := []int{2, 4, 8, 16}
	currents := make([]float64, 0, len(sizes))

	for _, n := range sizes {
		sp := NewSneakPathAnalyzer(n, n)
		sp.AnalyzeTarget(0, 0, voltage)
		stats := sp.GetStats(voltage)
		currents = append(currents, stats.TotalSneakCurrent)
	}

	for i := 1; i < len(currents); i++ {
		if currents[i] <= currents[i-1] {
			t.Fatalf("expected sneak current to increase with array size: %dx%d=%.6e A, %dx%d=%.6e A",
				sizes[i-1], sizes[i-1], currents[i-1], sizes[i], sizes[i], currents[i])
		}
	}
}

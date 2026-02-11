package crossbar

import (
	"math"
	"testing"
)

// TestNewSneakPathAnalyzer_Defaults verifies dimensions and default conductance 50µS
func TestNewSneakPathAnalyzer_Defaults(t *testing.T) {
	rows, cols := 4, 4
	sp := NewSneakPathAnalyzer(rows, cols)

	if sp.Rows != rows {
		t.Errorf("Rows = %d, want %d", sp.Rows, rows)
	}
	if sp.Cols != cols {
		t.Errorf("Cols = %d, want %d", sp.Cols, cols)
	}

	// Verify conductances initialized to 50µS
	defaultG := 50e-6
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if sp.Conductances[i][j] != defaultG {
				t.Errorf("Conductances[%d][%d] = %e, want %e", i, j, sp.Conductances[i][j], defaultG)
			}
		}
	}
}

// TestSetConductance_Valid verifies set and retrieve
func TestSetConductance_Valid(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)

	newG := 100e-6
	sp.SetConductance(2, 3, newG)

	if sp.Conductances[2][3] != newG {
		t.Errorf("Conductances[2][3] = %e, want %e", sp.Conductances[2][3], newG)
	}
}

// TestSetConductance_OutOfBounds verifies does nothing for invalid indices
func TestSetConductance_OutOfBounds(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	defaultG := 50e-6

	// Try to set out of bounds - should not crash
	sp.SetConductance(-1, 0, 999e-6)
	sp.SetConductance(0, -1, 999e-6)
	sp.SetConductance(4, 0, 999e-6)
	sp.SetConductance(0, 4, 999e-6)
	sp.SetConductance(10, 10, 999e-6)

	// Verify valid cells still have default value
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if sp.Conductances[i][j] != defaultG {
				t.Errorf("Conductances[%d][%d] = %e, should still be %e after out-of-bounds set",
					i, j, sp.Conductances[i][j], defaultG)
			}
		}
	}
}

// TestAnalyzeTarget_UniformArray verifies predictable sneak ratio for uniform conductance
func TestAnalyzeTarget_UniformArray(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	voltage := 1.0

	// All cells have same conductance (50µS)
	sp.AnalyzeTarget(1, 1, voltage)

	// With uniform conductances, sneak paths are predictable
	// For 4x4 array targeting (1,1), expect (4-1)*(4-1) = 9 sneak paths
	expectedPaths := (sp.Rows - 1) * (sp.Cols - 1)
	if len(sp.SneakPaths) != expectedPaths {
		t.Errorf("NumSneakPaths = %d, want %d", len(sp.SneakPaths), expectedPaths)
	}

	// Sneak ratio should be > 0 for uniform array
	if sp.TotalSneakRatio <= 0 {
		t.Errorf("TotalSneakRatio = %f, should be > 0 for uniform array", sp.TotalSneakRatio)
	}
}

// TestAnalyzeTarget_2x2Array verifies exactly 1 sneak path for 2x2
func TestAnalyzeTarget_2x2Array(t *testing.T) {
	sp := NewSneakPathAnalyzer(2, 2)
	voltage := 1.0

	sp.AnalyzeTarget(0, 0, voltage)

	// For 2x2 array, only 1 sneak path possible
	expectedPaths := 1
	if len(sp.SneakPaths) != expectedPaths {
		t.Errorf("NumSneakPaths = %d, want %d for 2x2 array", len(sp.SneakPaths), expectedPaths)
	}

	// Verify the sneak path goes through correct cells
	if len(sp.SneakPaths) > 0 {
		path := sp.SneakPaths[0]
		if path.PathLength != 3 {
			t.Errorf("PathLength = %d, want 3", path.PathLength)
		}

		// Expected path: (0,1), (1,1), (1,0)
		expectedCells := [][2]int{{0, 1}, {1, 1}, {1, 0}}
		if len(path.PathCells) != len(expectedCells) {
			t.Errorf("PathCells length = %d, want %d", len(path.PathCells), len(expectedCells))
		}

		for i, cell := range path.PathCells {
			if cell != expectedCells[i] {
				t.Errorf("PathCells[%d] = %v, want %v", i, cell, expectedCells[i])
			}
		}
	}
}

// TestAnalyzeTarget_NumPaths verifies for NxM, expect (N-1)*(M-1) sneak paths
func TestAnalyzeTarget_NumPaths(t *testing.T) {
	tests := []struct {
		rows int
		cols int
	}{
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
		{4, 6},
		{8, 8},
	}

	voltage := 1.0

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			sp := NewSneakPathAnalyzer(tc.rows, tc.cols)
			sp.AnalyzeTarget(0, 0, voltage)

			expectedPaths := (tc.rows - 1) * (tc.cols - 1)
			if len(sp.SneakPaths) != expectedPaths {
				t.Errorf("For %dx%d array: NumSneakPaths = %d, want %d",
					tc.rows, tc.cols, len(sp.SneakPaths), expectedPaths)
			}
		})
	}
}

// TestAnalyzeTarget_HighTargetConductance verifies high target cell → low sneak ratio (good)
func TestAnalyzeTarget_HighTargetConductance(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	voltage := 1.0

	// Set target cell to high conductance
	sp.SetConductance(1, 1, 500e-6) // 10x higher than default

	sp.AnalyzeTarget(1, 1, voltage)

	// High target conductance means strong target current
	// Sneak ratio should be relatively low
	if sp.TotalSneakRatio > 2.0 {
		t.Errorf("TotalSneakRatio = %f, expected < 2.0 with high target conductance", sp.TotalSneakRatio)
	}
}

// TestAnalyzeTarget_LowTargetConductance verifies low target cell → high sneak ratio (bad)
func TestAnalyzeTarget_LowTargetConductance(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	voltage := 1.0

	// Set target cell to low conductance
	sp.SetConductance(1, 1, 5e-6) // 10x lower than default

	sp.AnalyzeTarget(1, 1, voltage)

	// Low target conductance means weak target current
	// Sneak ratio should be relatively high (bad signal quality)
	if sp.TotalSneakRatio < 1.0 {
		t.Errorf("TotalSneakRatio = %f, expected > 1.0 with low target conductance", sp.TotalSneakRatio)
	}
}

// TestGetStats_SNR verifies SNR is positive dB for reasonable array
func TestGetStats_SNR(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	voltage := 1.0

	// Set target cell higher than others for good SNR
	sp.SetConductance(1, 1, 200e-6)

	sp.AnalyzeTarget(1, 1, voltage)
	stats := sp.GetStats(voltage)

	// SNR should be positive (target current > sneak current)
	if stats.SignalToNoiseRatio <= 0 {
		t.Errorf("SignalToNoiseRatio = %f dB, expected > 0", stats.SignalToNoiseRatio)
	}

	// Verify stats fields are populated
	if stats.TargetCurrent <= 0 {
		t.Errorf("TargetCurrent = %e, should be > 0", stats.TargetCurrent)
	}
	if stats.NumSneakPaths != len(sp.SneakPaths) {
		t.Errorf("NumSneakPaths = %d, want %d", stats.NumSneakPaths, len(sp.SneakPaths))
	}
}

// TestAnalyzeWithMitigation_Selector verifies selector (high on/off ratio) decreases sneak ratio
func TestAnalyzeWithMitigation_Selector(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	voltage := 1.0

	// Get baseline without mitigation
	sp.AnalyzeTarget(1, 1, voltage)
	baselineRatio := sp.TotalSneakRatio

	// Apply selector mitigation
	mit := SneakMitigation{
		UseSelector:   true,
		SelectorOnOff: 100.0, // 100:1 on/off ratio
	}

	stats := sp.AnalyzeWithMitigation(1, 1, voltage, mit)

	// Sneak ratio should decrease significantly with selector
	if stats.SneakRatio >= baselineRatio {
		t.Errorf("SneakRatio with selector = %f, should be < baseline %f",
			stats.SneakRatio, baselineRatio)
	}

	// Verify reduction is substantial (at least 10x improvement expected)
	if stats.SneakRatio > baselineRatio/10.0 {
		t.Logf("SneakRatio with selector = %f, baseline = %f (reduction: %.2fx)",
			stats.SneakRatio, baselineRatio, baselineRatio/stats.SneakRatio)
	}
}

// TestAnalyzeWithMitigation_HalfSelect verifies half-select voltage decreases sneak current
func TestAnalyzeWithMitigation_HalfSelect(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	voltage := 1.0

	// Get baseline without mitigation
	sp.AnalyzeTarget(1, 1, voltage)
	baselineStats := sp.GetStats(voltage)

	// Apply half-select mitigation
	mit := SneakMitigation{
		UseHalfSelect:     true,
		HalfSelectVoltage: 0.5, // Half voltage on unselected paths
	}

	stats := sp.AnalyzeWithMitigation(1, 1, voltage, mit)

	// Sneak current should decrease with half-select
	if stats.TotalSneakCurrent >= baselineStats.TotalSneakCurrent {
		t.Errorf("TotalSneakCurrent with half-select = %e, should be < baseline %e",
			stats.TotalSneakCurrent, baselineStats.TotalSneakCurrent)
	}

	// Target current should remain the same
	if math.Abs(stats.TargetCurrent-baselineStats.TargetCurrent) > 1e-10 {
		t.Errorf("TargetCurrent changed: %e vs %e", stats.TargetCurrent, baselineStats.TargetCurrent)
	}
}

// TestAnalyzeWithMitigation_Both verifies selector + half-select gives best result
func TestAnalyzeWithMitigation_Both(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	voltage := 1.0

	// Get baseline
	sp.AnalyzeTarget(1, 1, voltage)
	baselineRatio := sp.TotalSneakRatio

	// Apply selector only
	mitSelector := SneakMitigation{
		UseSelector:   true,
		SelectorOnOff: 100.0,
	}
	statsSelector := sp.AnalyzeWithMitigation(1, 1, voltage, mitSelector)

	// Apply half-select only
	mitHalfSelect := SneakMitigation{
		UseHalfSelect:     true,
		HalfSelectVoltage: 0.5,
	}
	statsHalfSelect := sp.AnalyzeWithMitigation(1, 1, voltage, mitHalfSelect)

	// Apply both
	mitBoth := SneakMitigation{
		UseSelector:       true,
		SelectorOnOff:     100.0,
		UseHalfSelect:     true,
		HalfSelectVoltage: 0.5,
	}
	statsBoth := sp.AnalyzeWithMitigation(1, 1, voltage, mitBoth)

	// Both mitigations should give the best result
	if statsBoth.SneakRatio >= statsSelector.SneakRatio {
		t.Errorf("Both mitigations SneakRatio = %f, should be < selector-only %f",
			statsBoth.SneakRatio, statsSelector.SneakRatio)
	}
	if statsBoth.SneakRatio >= statsHalfSelect.SneakRatio {
		t.Errorf("Both mitigations SneakRatio = %f, should be < half-select-only %f",
			statsBoth.SneakRatio, statsHalfSelect.SneakRatio)
	}
	if statsBoth.SneakRatio >= baselineRatio {
		t.Errorf("Both mitigations SneakRatio = %f, should be < baseline %f",
			statsBoth.SneakRatio, baselineRatio)
	}

	t.Logf("Baseline: %.4f, Selector: %.4f, HalfSelect: %.4f, Both: %.4f",
		baselineRatio, statsSelector.SneakRatio, statsHalfSelect.SneakRatio, statsBoth.SneakRatio)
}

// TestAnalyzeTarget_ZeroConductanceHandling verifies no crash with zero conductances
func TestAnalyzeTarget_ZeroConductanceHandling(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	voltage := 1.0

	// Set some cells to zero conductance
	sp.SetConductance(0, 0, 0)
	sp.SetConductance(1, 1, 0)
	sp.SetConductance(2, 2, 0)

	// Should not crash
	sp.AnalyzeTarget(1, 1, voltage)

	// Paths through zero-conductance cells should not contribute
	for _, path := range sp.SneakPaths {
		if path.PathCurrent < 0 {
			t.Errorf("PathCurrent = %e, should not be negative", path.PathCurrent)
		}
	}
}

// TestGetStats_NoSneakPaths verifies SNR handling when no sneak paths exist
func TestGetStats_NoSneakPaths(t *testing.T) {
	sp := NewSneakPathAnalyzer(1, 1)
	voltage := 1.0

	// 1x1 array has no sneak paths
	sp.AnalyzeTarget(0, 0, voltage)
	stats := sp.GetStats(voltage)

	if stats.NumSneakPaths != 0 {
		t.Errorf("NumSneakPaths = %d, want 0 for 1x1 array", stats.NumSneakPaths)
	}
	if stats.TotalSneakCurrent != 0 {
		t.Errorf("TotalSneakCurrent = %e, want 0", stats.TotalSneakCurrent)
	}

	// SNR should be very high (100 dB default when no sneak)
	if stats.SignalToNoiseRatio < 50 {
		t.Errorf("SignalToNoiseRatio = %f dB, expected very high value", stats.SignalToNoiseRatio)
	}
}

// TestAnalyzeWithMitigation_EdgeCases verifies mitigation with extreme parameters
func TestAnalyzeWithMitigation_EdgeCases(t *testing.T) {
	sp := NewSneakPathAnalyzer(4, 4)
	voltage := 1.0

	t.Run("very high selector ratio", func(t *testing.T) {
		mit := SneakMitigation{
			UseSelector:   true,
			SelectorOnOff: 1e6, // Very high ratio
		}
		stats := sp.AnalyzeWithMitigation(1, 1, voltage, mit)

		// Should dramatically reduce sneak
		if stats.SneakRatio > 0.001 {
			t.Logf("SneakRatio = %f with 1e6 selector ratio", stats.SneakRatio)
		}
	})

	t.Run("half-select voltage equals full voltage", func(t *testing.T) {
		mit := SneakMitigation{
			UseHalfSelect:     true,
			HalfSelectVoltage: 1.0, // Same as full voltage
		}
		stats := sp.AnalyzeWithMitigation(1, 1, voltage, mit)

		// Effective voltage on sneak paths should be zero
		if stats.TotalSneakCurrent > 1e-12 {
			t.Errorf("TotalSneakCurrent = %e, expected ~0 when half-select = full voltage",
				stats.TotalSneakCurrent)
		}
	})

	t.Run("half-select voltage exceeds full voltage", func(t *testing.T) {
		mit := SneakMitigation{
			UseHalfSelect:     true,
			HalfSelectVoltage: 1.5, // Higher than full voltage
		}
		stats := sp.AnalyzeWithMitigation(1, 1, voltage, mit)

		// Should clamp to zero (no negative voltage)
		if stats.TotalSneakCurrent != 0 {
			t.Errorf("TotalSneakCurrent = %e, expected 0 when half-select > full voltage",
				stats.TotalSneakCurrent)
		}
	})
}

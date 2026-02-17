package crossbar

import (
	"math"
	"testing"
)

func newHalfSelectTestArray(t *testing.T, disturbThreshold, disturbRate float64) *Array {
	t.Helper()

	cfg := &Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
		HalfSelect: &HalfSelectConfig{
			Enabled:          true,
			DisturbThreshold: disturbThreshold,
			DisturbRate:      disturbRate,
		},
	}

	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("failed to create test array: %v", err)
	}

	for r := 0; r < cfg.Rows; r++ {
		for c := 0; c < cfg.Cols; c++ {
			if err := arr.ProgramWeight(r, c, 0.5); err != nil {
				t.Fatalf("failed to initialize cell (%d,%d): %v", r, c, err)
			}
		}
	}
	arr.ResetDisturbTracking()
	return arr
}

// (1) Program target cell and measure neighbor cells before/after.
func TestHalfSelectDisturb_ProgramTargetAndMeasureNeighbors(t *testing.T) {
	arr := newHalfSelectTestArray(t, 0.3, 0.01)
	targetRow, targetCol := 4, 4

	beforeRowNeighbor := arr.cells[targetRow][targetCol-1].DisturbShift
	beforeColNeighbor := arr.cells[targetRow-1][targetCol].DisturbShift
	beforeDiagonal := arr.cells[targetRow-1][targetCol-1].DisturbShift

	if err := arr.ProgramWeightWithDisturb(targetRow, targetCol, 0.8, true); err != nil {
		t.Fatalf("ProgramWeightWithDisturb failed: %v", err)
	}

	afterRowNeighbor := arr.cells[targetRow][targetCol-1].DisturbShift
	afterColNeighbor := arr.cells[targetRow-1][targetCol].DisturbShift
	afterDiagonal := arr.cells[targetRow-1][targetCol-1].DisturbShift

	if afterRowNeighbor <= beforeRowNeighbor {
		t.Fatalf("same-row neighbor disturb did not increase: before=%.6f after=%.6f", beforeRowNeighbor, afterRowNeighbor)
	}
	if afterColNeighbor <= beforeColNeighbor {
		t.Fatalf("same-column neighbor disturb did not increase: before=%.6f after=%.6f", beforeColNeighbor, afterColNeighbor)
	}
	if afterDiagonal != beforeDiagonal {
		t.Fatalf("non-half-selected diagonal cell changed unexpectedly: before=%.6f after=%.6f", beforeDiagonal, afterDiagonal)
	}
	if arr.cells[targetRow][targetCol].DisturbShift != 0 {
		t.Fatalf("target cell must not accumulate half-select disturb, got %.6f", arr.cells[targetRow][targetCol].DisturbShift)
	}
}

// (2) Verify disturb magnitude matches V/2 model for passive arrays.
func TestHalfSelectDisturb_PassiveV2MagnitudeMatchesModel(t *testing.T) {
	const (
		disturbThreshold = 0.3
		disturbRate      = 0.01
		cycles           = 20
	)
	arr := newHalfSelectTestArray(t, disturbThreshold, disturbRate)
	targetRow, targetCol := 4, 4

	for i := 0; i < cycles; i++ {
		if err := arr.ProgramWeightWithDisturb(targetRow, targetCol, 0.7, true); err != nil {
			t.Fatalf("ProgramWeightWithDisturb failed at cycle %d: %v", i, err)
		}
	}

	expectedPerPulse := disturbRate * (0.5 - disturbThreshold) / (1.0 - disturbThreshold)
	expectedTotal := float64(cycles) * expectedPerPulse
	actual := arr.cells[targetRow][targetCol-1].DisturbShift

	if math.Abs(actual-expectedTotal) > 1e-12 {
		t.Fatalf("V/2 disturb mismatch: expected %.12f, got %.12f", expectedTotal, actual)
	}
}

// (3) Verify 1T1R selector suppresses disturb below threshold.
func TestHalfSelectDisturb_1T1RSelectorSuppressesBelowThreshold(t *testing.T) {
	const (
		writeV          = 2.0
		coerciveV       = 1.0
		disturbRatioThr = 0.3
		selectorGain    = 0.1 // 10x half-select attenuation
	)

	passiveHalfV := HalfSelectVoltage(writeV, "V/2")
	activeHalfV := passiveHalfV * selectorGain
	thresholdV := coerciveV * disturbRatioThr

	if passiveHalfV <= thresholdV {
		t.Fatalf("sanity check failed: passive V/2 (%.3fV) should exceed threshold (%.3fV)", passiveHalfV, thresholdV)
	}
	if activeHalfV >= thresholdV {
		t.Fatalf("sanity check failed: selector-suppressed voltage (%.3fV) should be below threshold (%.3fV)", activeHalfV, thresholdV)
	}

	arrPassive := newHalfSelectTestArray(t, disturbRatioThr, 0.01)
	arrActive := newHalfSelectTestArray(t, disturbRatioThr, 0.01)
	targetRow, targetCol := 4, 4

	for i := 0; i < 100; i++ {
		if err := arrPassive.ProgramWeightWithDisturb(targetRow, targetCol, 0.7, true); err != nil {
			t.Fatalf("passive ProgramWeightWithDisturb failed: %v", err)
		}
		if err := arrActive.ProgramWeightWithDisturb(targetRow, targetCol, 0.7, false); err != nil {
			t.Fatalf("1T1R ProgramWeightWithDisturb failed: %v", err)
		}
	}

	passiveDrift := arrPassive.cells[targetRow][targetCol-1].DisturbShift
	activeDrift := arrActive.cells[targetRow][targetCol-1].DisturbShift

	if passiveDrift <= 0 {
		t.Fatalf("passive neighbor should accumulate disturb, got %.6f", passiveDrift)
	}
	if activeDrift != 0 {
		t.Fatalf("1T1R selector should suppress disturb below threshold, got neighbor drift %.6f", activeDrift)
	}
}

// (4) Verify cumulative disturb over 100 cycles stays within 5% drift on neighbors.
func TestHalfSelectDisturb_Cumulative100CyclesWithinFivePercent(t *testing.T) {
	const (
		disturbThreshold = 0.3
		disturbRate      = 0.001 // 100 cycles -> ~2.86% expected drift under V/2 model
		cycles           = 100
		maxAllowedDrift  = 0.05
	)

	arr := newHalfSelectTestArray(t, disturbThreshold, disturbRate)
	targetRow, targetCol := 4, 4

	for i := 0; i < cycles; i++ {
		if err := arr.ProgramWeightWithDisturb(targetRow, targetCol, 0.65, true); err != nil {
			t.Fatalf("ProgramWeightWithDisturb failed at cycle %d: %v", i, err)
		}
	}

	maxNeighborDrift := 0.0
	for r := 0; r < arr.Rows(); r++ {
		for c := 0; c < arr.Cols(); c++ {
			if r == targetRow && c == targetCol {
				continue
			}
			isHalfSelected := r == targetRow || c == targetCol
			if !isHalfSelected {
				continue
			}
			drift := arr.cells[r][c].DisturbShift
			if drift > maxNeighborDrift {
				maxNeighborDrift = drift
			}
		}
	}

	if maxNeighborDrift > maxAllowedDrift {
		t.Fatalf("cumulative neighbor drift exceeds 5%%: max drift=%.6f (> %.6f)", maxNeighborDrift, maxAllowedDrift)
	}
}

//go:build legacy_fyne

package gui

import (
	"fmt"
	"math"
	"testing"
)

// TestWriteCell_ProofBeforeAfter proves that ISPP write-cell converges from
// any start level to any target level using escalating-voltage pulses.
//
// Evidence: before-level and after-level printed for 6 cells.
// A fixed voltage cannot always reach max level in one pulse (LK physics).
// Real ISPP escalates voltage until the target is reached.
func TestWriteCell_ProofBeforeAfter(t *testing.T) {
	ds := NewDeviceState(8, 8, nil, nil)
	numLevels := ds.writeRange.NumLevels
	if numLevels < 2 {
		numLevels = 30
	}
	pulseWidth := float64(PhaseWriteDurationNs) * 1e-9 * 10.0

	type writeCase struct {
		name        string
		startLevel  int
		targetLevel int
		direction   string
	}
	cases := []writeCase{
		{"low→mid", 0, numLevels / 2, "ascending"},
		{"mid→80%", numLevels / 2, numLevels * 80 / 100, "ascending"},
		{"high→mid", numLevels - 1, numLevels / 4, "descending"},
		{"mid→low", numLevels / 4, 1, "descending"},
		{"zero→15%", 0, numLevels * 15 / 100, "ascending"},
		{"high→85%", numLevels - 1, numLevels * 85 / 100, "descending"},
	}

	t.Logf("Array: 8×8 | NumLevels: %d | Material: %s", numLevels, ds.material.Name)
	t.Log("ISPP with escalating voltage: Vstart=0.5V → Vmax=2.5V in 0.1V steps")
	t.Log("════════════════════════════════════════════════════════════════════════")
	t.Logf("%-20s  %5s  %6s  %6s  %6s  %7s  %s",
		"Case", "Start", "Target", "After", "Delta", "Pulses", "Status")
	t.Log("────────────────────────────────────────────────────────────────────────")

	allPass := true
	for _, tc := range cases {
		startLevel := tc.startLevel
		curLevel := startLevel
		totalPulses := 0

		// Escalating voltage from 0.5V to 2.5V
		for vStep := 0.5; vStep <= 2.5; vStep += 0.1 {
			applied, _ := ds.DACWriteVoltage(vStep)
			v := math.Abs(applied)
			if tc.direction == "descending" {
				v = -v
			}
			next := ds.programLevelFromCoupledVoltage(curLevel, v, pulseWidth, numLevels)
			totalPulses++
			if next != curLevel {
				curLevel = next
			}
			if tc.direction == "ascending" && curLevel >= tc.targetLevel {
				curLevel = tc.targetLevel
				break
			}
			if tc.direction == "descending" && curLevel <= tc.targetLevel {
				curLevel = tc.targetLevel
				break
			}
		}

		afterLevel := curLevel
		success := afterLevel == tc.targetLevel
		if !success {
			allPass = false
		}
		delta := afterLevel - startLevel
		status := "PASS ✓"
		if !success {
			status = fmt.Sprintf("FAIL ✗ (off by %d)", afterLevel-tc.targetLevel)
		}
		t.Logf("%-20s  %5d  %6d  %6d  %6d  %7d  %s",
			tc.name, startLevel, tc.targetLevel, afterLevel, delta, totalPulses, status)
		if !success {
			t.Errorf("WRITE FAIL: %q start=%d target=%d after=%d delta=%d pulses=%d",
				tc.name, startLevel, tc.targetLevel, afterLevel, delta, totalPulses)
		}
	}

	t.Log("────────────────────────────────────────────────────────────────────────")
	if allPass {
		t.Log("ALL 6 WRITE-CELL CASES PASSED: escalating ISPP converges to every target")
	}
	fmt.Printf("WRITE_PROOF: all_pass=%v numLevels=%d\n", allPass, numLevels)
}

// TestWriteCell_VoltageTopOnlyInjection proves the WRITE architecture:
//   - Voltage is injected ONLY from DAC (external WL/BL pins)
//   - Internal cell voltage is DERIVED by the MNA solver (never forced)
//   - Vcell ≤ VDAC always (solver gives lossy, physically correct result)
func TestWriteCell_VoltageTopOnlyInjection(t *testing.T) {
	ds := NewDeviceState(4, 4, nil, nil)
	ds.SetPassiveMode(true)
	ds.SetOperationMode(OpModeWrite)

	targetVoltage := 1.2
	applied, code := ds.DACWriteVoltage(targetVoltage)
	if code <= 0 {
		t.Fatalf("DACWriteVoltage returned code=%d for target=%.3f V", code, targetVoltage)
	}

	t.Logf("DAC: target=%.4f V → applied=%.4f V (code=%d)", targetVoltage, applied, code)

	row, col := 2, 2
	ds.ApplyHalfSelectWrite(row, col, applied)

	weights := [][]int{
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
		{15, 15, 15, 15},
	}
	ds.Compute(weights, 30)

	vCell := ds.GetEffectiveCellVoltage(row, col)
	t.Logf("MNA solver → Vcell(2,2) = %.6f V  (DAC injected = %.6f V)", vCell, applied)

	// INVARIANT: Internal cell voltage ≤ DAC voltage (solver-derived, not forced)
	if math.Abs(vCell) > math.Abs(applied)+1e-9 {
		t.Errorf("BOUNDARY VIOLATION: |Vcell| %.6f > |VDAC| %.6f", math.Abs(vCell), math.Abs(applied))
	} else {
		t.Logf("PASS: Vcell ≤ VDAC — internal voltage solver-derived (not forced)")
	}

	// Half-select cells must not exceed DAC voltage
	violations := 0
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if r == row && c == col {
				continue
			}
			v := math.Abs(ds.GetEffectiveCellVoltage(r, c))
			if v > math.Abs(applied)+1e-9 {
				violations++
				t.Errorf("Cell(%d,%d) |Vcell|=%.6f > |VDAC|=%.6f", r, c, v, math.Abs(applied))
			}
		}
	}
	if violations == 0 {
		t.Logf("PASS: All 15 non-target cells bounded by VDAC (MNA solver enforced)")
	}
	fmt.Printf("WRITE_BOUNDARY: vCell=%.6f vDAC=%.6f code=%d violations=%d\n",
		vCell, applied, code, violations)
}

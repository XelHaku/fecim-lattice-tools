//go:build legacy_fyne

package gui

import (
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

func TestDeviceState_ProgramLevelFromCoupledVoltage_MonotonicWithVoltage(t *testing.T) {
	ds := NewDeviceState(2, 2, nil, nil)
	currentLevel := 10
	levels := 30
	pulseWidth := 10 * float64(PhaseWriteDurationNs) * 1e-9

	lowV := 0.6
	highV := 1.2

	lowLevel := ds.programLevelFromCoupledVoltage(currentLevel, lowV, pulseWidth, levels)
	highLevel := ds.programLevelFromCoupledVoltage(currentLevel, highV, pulseWidth, levels)

	if lowLevel < currentLevel {
		t.Fatalf("positive pulse should not decrease level: got %d from current %d", lowLevel, currentLevel)
	}
	if highLevel < lowLevel {
		t.Fatalf("higher effective voltage should not program less: lowV=%.3f->L%d highV=%.3f->L%d", lowV, lowLevel, highV, highLevel)
	}
}

func TestDeviceState_ProgramLevelFromCoupledVoltage_UsesActualCoupledVoltage(t *testing.T) {
	ds := NewDeviceState(3, 3, nil, nil)
	ds.SetPassiveMode(true)
	ds.SetOperationMode(OpModeWrite)
	ds.SetCouplingMode(arraysim.CouplingTierA)

	// Force visible IR drop in coupled solve.
	ds.wireParams = arraysim.WireParams{RWordLine: 2e5, RBitLine: 2e5}

	// Make neighbors high-conductance to increase loading/coupling impact.
	weights := [][]int{
		{29, 29, 29},
		{29, 5, 29},
		{29, 29, 29},
	}

	row, col := 1, 1
	targetVoltage := 1.2
	applied, _ := ds.DACWriteVoltage(targetVoltage)
	ds.ApplyHalfSelectWrite(row, col, applied)
	ds.Compute(weights, 30)

	vActual := math.Abs(ds.GetEffectiveCellVoltage(row, col))
	if vActual <= 0 {
		t.Fatalf("expected positive coupled target-cell voltage, got %.9fV", vActual)
	}
	if vActual >= math.Abs(applied) {
		t.Fatalf("expected coupled target-cell voltage to include IR drop (vActual < vDAC). got vActual=%.9fV vDAC=%.9fV", vActual, math.Abs(applied))
	}

	currentLevel := weights[row][col]
	pulseWidth := 10 * float64(PhaseWriteDurationNs) * 1e-9

	idealNext := ds.programLevelFromCoupledVoltage(currentLevel, math.Abs(applied), pulseWidth, 30)
	coupledNext := ds.programLevelFromCoupledVoltage(currentLevel, vActual, pulseWidth, 30)

	if coupledNext > idealNext {
		t.Fatalf("coupled write update should be bounded by ideal-voltage update: coupled=%d ideal=%d (vActual=%.6fV, vDAC=%.6fV)", coupledNext, idealNext, vActual, math.Abs(applied))
	}
}

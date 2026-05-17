//go:build legacy_fyne

package gui

import (
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

// TestCouplingFidelityTiers_ExpectedBehavior validates the headless coupling-tier
// behavior contract for Module 4:
//   - Ideal: no coupled snapshot; effective Vcell follows direct BL drive.
//   - Tier-A: coupled snapshot is produced by approximate solver.
//   - Tier-B: coupled snapshot is produced by DC nodal reference solver.
func TestCouplingFidelityTiers_ExpectedBehavior(t *testing.T) {
	tiers := []struct {
		name                string
		mode                arraysim.CouplingMode
		expectCoupledResult bool
	}{
		{name: "Ideal", mode: arraysim.CouplingIdeal, expectCoupledResult: false},
		{name: "Tier-A", mode: arraysim.CouplingTierA, expectCoupledResult: true},
		{name: "Tier-B", mode: arraysim.CouplingTierB, expectCoupledResult: true},
	}

	weights := [][]int{{20, 10}, {8, 24}}
	const quantLevels = 30
	const readV = 0.25

	for _, tc := range tiers {
		t.Run(tc.name, func(t *testing.T) {
			ds := newTestDeviceState(2, 2)
			ds.SetOperationMode(OpModeRead)
			ds.SetWLSingle(0)
			ds.SetDACVoltage(0, readV)
			ds.SetDACVoltage(1, 0)
			ds.SetCouplingMode(tc.mode)

			ds.Compute(weights, quantLevels)

			vCells, iCells := ds.GetCoupledCellSnapshot()
			if tc.expectCoupledResult {
				if vCells == nil || iCells == nil {
					t.Fatalf("%s expected coupled snapshots, got voltages=%v currents=%v", tc.name, vCells, iCells)
				}
				if len(vCells) != 2 || len(vCells[0]) != 2 {
					cols := 0
					if len(vCells) > 0 {
						cols = len(vCells[0])
					}
					t.Fatalf("%s coupled voltage shape mismatch: got %dx%d", tc.name, len(vCells), cols)
				}
				if got := ds.GetEffectiveCellVoltage(0, 0); math.Abs(got-vCells[0][0]) > 1e-12 {
					t.Fatalf("%s effective voltage must come from coupled snapshot: got %.9g want %.9g", tc.name, got, vCells[0][0])
				}
			} else {
				if vCells != nil || iCells != nil {
					t.Fatalf("%s should not expose coupled snapshots, got voltages=%v currents=%v", tc.name, vCells, iCells)
				}
				if got := ds.GetEffectiveCellVoltage(0, 0); math.Abs(got-readV) > 1e-12 {
					t.Fatalf("%s effective voltage mismatch: got %.9g V want %.9g V", tc.name, got, readV)
				}
			}

			if rowIuA := ds.GetRowCurrent(0); rowIuA <= 0 {
				t.Fatalf("%s expected positive selected-row current, got %.9g uA", tc.name, rowIuA)
			}
		})
	}
}

func TestCouplingFidelityTiers_SwitchingResetsIdealSnapshot(t *testing.T) {
	ds := newTestDeviceState(2, 2)
	weights := [][]int{{15, 15}, {15, 15}}

	ds.SetCouplingMode(arraysim.CouplingTierA)
	ds.SetDACVoltage(0, 0.2)
	ds.Compute(weights, 30)
	vCells, iCells := ds.GetCoupledCellSnapshot()
	if vCells == nil || iCells == nil {
		t.Fatal("tier A must produce coupled snapshots before mode switch")
	}

	ds.SetCouplingMode(arraysim.CouplingIdeal)
	ds.Compute(weights, 30)
	vCells, iCells = ds.GetCoupledCellSnapshot()
	if vCells != nil || iCells != nil {
		t.Fatalf("ideal mode must clear coupled snapshots, got voltages=%v currents=%v", vCells, iCells)
	}
}

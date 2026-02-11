package gui

import (
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// TestReadCoupling_SignedPerCellVI validates signed per-cell READ voltages/currents
// in passive biasing for selected + half-selected cells, including polarity reversal.
func TestReadCoupling_SignedPerCellVI(t *testing.T) {
	ds := newTestDeviceState(2, 2)
	ds.SetPassiveMode(true)
	ds.SetOperationMode(OpModeRead)
	ds.SetDACRangeMode(DACRangeRead)
	ds.SetCouplingMode(arraysim.CouplingTierA)

	// Use identical conductance in all cells to isolate V-sign and geometry effects.
	weights := [][]int{{15, 15}, {15, 15}}

	tests := []struct {
		name             string
		halfBiasV        float64
		wantSelectedSign float64
	}{
		{
			name:             "positive selected + half-selected read",
			halfBiasV:        +0.2,
			wantSelectedSign: +1,
		},
		{
			name:             "negative selected + half-selected read",
			halfBiasV:        -0.2,
			wantSelectedSign: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 2x2 passive read bias pattern (row0/col0 selected):
			// WL0=+/-V/2, BL0=-/+V/2, others=0.
			// Selected (0,0): +/-V; half-selected (0,1) and (1,0): +/-V/2.
			ds.wlVoltages[0] = tc.halfBiasV
			ds.wlVoltages[1] = 0
			ds.dacVoltages[0] = -tc.halfBiasV
			ds.dacVoltages[1] = 0

			ds.Compute(weights, 30)

			vSel := ds.GetEffectiveCellVoltage(0, 0)
			vHalfRow := ds.GetEffectiveCellVoltage(0, 1)
			vHalfCol := ds.GetEffectiveCellVoltage(1, 0)
			iSel := ds.GetCoupledCellCurrent(0, 0)
			iHalfRow := ds.GetCoupledCellCurrent(0, 1)
			iHalfCol := ds.GetCoupledCellCurrent(1, 0)

			if tc.wantSelectedSign*vSel <= 0 {
				t.Fatalf("selected V sign mismatch: got %.6g V", vSel)
			}
			if tc.wantSelectedSign*iSel <= 0 {
				t.Fatalf("selected I sign mismatch: got %.6g A", iSel)
			}
			if tc.wantSelectedSign*vHalfRow <= 0 {
				t.Fatalf("half-selected row V sign mismatch: got %.6g V", vHalfRow)
			}
			if tc.wantSelectedSign*vHalfCol <= 0 {
				t.Fatalf("half-selected col V sign mismatch: got %.6g V", vHalfCol)
			}
			if tc.wantSelectedSign*iHalfRow <= 0 {
				t.Fatalf("half-selected row I sign mismatch: got %.6g A", iHalfRow)
			}
			if tc.wantSelectedSign*iHalfCol <= 0 {
				t.Fatalf("half-selected col I sign mismatch: got %.6g A", iHalfCol)
			}

			// Magnitude checks with tolerance:
			// selected V is nominally 2x half-selected V in this bias scheme;
			// current ratio follows V ratio because G is equal across cells.
			const ratioTol = 0.08 // 8% to absorb Tier-A fixed-point + finite wire drop.
			absRatio := func(a, b float64) float64 {
				if b == 0 {
					return math.Inf(1)
				}
				return math.Abs(a) / math.Abs(b)
			}
			assertNear := func(label string, got, want, tol float64) {
				t.Helper()
				if math.Abs(got-want) > tol {
					t.Fatalf("%s: got %.6f, want %.6f ± %.6f", label, got, want, tol)
				}
			}

			assertNear("|Vsel|/|Vhalf-row|", absRatio(vSel, vHalfRow), 2.0, ratioTol)
			assertNear("|Vsel|/|Vhalf-col|", absRatio(vSel, vHalfCol), 2.0, ratioTol)
			assertNear("|Isel|/|Ihalf-row|", absRatio(iSel, iHalfRow), 2.0, ratioTol)
			assertNear("|Isel|/|Ihalf-col|", absRatio(iSel, iHalfCol), 2.0, ratioTol)
		})
	}
}

func TestReadCoupling_DefaultsToTierA(t *testing.T) {
	ds := newTestDeviceState(2, 2)
	ds.SetOperationMode(OpModeRead)
	ds.SetWLSingle(0)
	ds.SetDACPreset(DACReadPreset, 0.2)

	weights := [][]int{{15, 15}, {15, 15}}
	ds.Compute(weights, 30)

	vCells, iCells := ds.GetCoupledCellSnapshot()
	if vCells == nil || iCells == nil {
		t.Fatalf("expected coupled Tier-A snapshots in READ default path, got voltages=%v currents=%v", vCells, iCells)
	}
	if got := ds.GetCouplingMode(); got != arraysim.CouplingTierA {
		t.Fatalf("default coupling mode mismatch: got=%v want=%v", got, arraysim.CouplingTierA)
	}
}

func TestReadCoupling_MaterialSelectionChangesReadCurrent(t *testing.T) {
	ds := newTestDeviceState(1, 1)
	ds.SetOperationMode(OpModeRead)
	ds.SetWLSingle(0)
	ds.SetDACPreset(DACReadPreset, 0.25)

	weights := [][]int{{15}}

	ds.SetMaterial(sharedphysics.FeCIMMaterial())
	ds.Compute(weights, 30)
	feCIMCurrentUA := ds.GetRowCurrent(0)

	ds.SetMaterial(sharedphysics.LiteratureSuperlattice())
	ds.Compute(weights, 30)
	superCurrentUA := ds.GetRowCurrent(0)

	if feCIMCurrentUA <= 0 || superCurrentUA <= 0 {
		t.Fatalf("expected positive READ currents, got FeCIM=%.6g uA superlattice=%.6g uA", feCIMCurrentUA, superCurrentUA)
	}

	// Materials have different Gmin/Gmax; READ current should reflect that.
	if math.Abs(superCurrentUA-feCIMCurrentUA) < 1e-6 {
		t.Fatalf("material-dependent READ current not observed: FeCIM=%.9f uA superlattice=%.9f uA", feCIMCurrentUA, superCurrentUA)
	}
}

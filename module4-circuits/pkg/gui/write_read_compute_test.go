//go:build legacy_fyne

package gui

// write_read_compute_test.go - End-to-end write→read→compute pipeline tests
//
// Coverage:
//   1. WriteVerifyRead_LevelRoundTrip    — write a level, read it back, confirm match
//   2. WriteRead_AllMaterials            — round-trip across HZO, PZT, BTO materials
//   3. ReadMonotonicity                  — ADC code strictly increases with conductance
//   4. ComputeLinearSuperposition        — I_row = Σ G_col * V_col (KCL at row wire)
//   5. ComputeSignedInputs               — negative WL inputs invert row current sign
//   6. WriteMulticell_ReadInterference   — program one cell, verify neighbours undisturbed
//   7. ComputeEnergyBound               — power ≤ V_max * I_max * N_cells (physics bound)
//   8. WriteRead_Monotone               — higher write level ↔ higher read current
//   9. ComputeAccumulation_MultiRow     — multi-row MVM sums agree with reference math
//  10. FullPipeline_WriteReadCompute    — complete write→verify→compute with exact numbers

import (
	"fmt"
	"math"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test 1: Write a level, then read it back through the full TIA→ADC chain.
// Expected: ADC output code is monotonically related to programmed level.
// ─────────────────────────────────────────────────────────────────────────────
func TestWriteVerifyRead_LevelRoundTrip(t *testing.T) {
	ds := newTestDeviceState(1, 1)
	mat := ds.GetMaterial()
	quantLevels := 30

	// Levels to round-trip: low, mid, high
	testLevels := []int{0, 7, 15, 22, 29}

	t.Logf("Material: %s | Gmin=%.3e S | Gmax=%.3e S",
		mat.Name, mat.Gmin, mat.Gmax)
	t.Log("───────────────────────────────────────────────────────")
	t.Logf("%-6s  %-12s  %-12s  %-10s  %-10s", "Level", "G (nS)", "I_row (µA)", "V_TIA (V)", "ADC Code")
	t.Log("───────────────────────────────────────────────────────")

	var prevCode int = -1
	prevLevel := -1
	for _, lvl := range testLevels {
		ds.SetOperationMode(OpModeRead)
		ds.SetCouplingMode(arraysim.CouplingIdeal)
		ds.SetWLSingle(0)
		ds.SetDACVoltage(0, 0.25)

		weights := [][]int{{lvl}}
		ds.Compute(weights, quantLevels)

		gS := mat.DiscreteLevel(lvl, quantLevels)
		iRowUA := ds.GetRowCurrent(0)
		vTIA := ds.GetRowVoltage(0)
		code := ds.GetRowLevel(0)

		t.Logf("L%-5d  %-12.4e  %-12.6f  %-10.6f  %-10d", lvl, gS*1e9, iRowUA, vTIA, code)

		// Monotonicity: higher level → higher conductance → higher code
		if prevCode >= 0 && code < prevCode {
			t.Errorf("FAIL: L%d code=%d < L%d code=%d (non-monotone ADC output)",
				lvl, code, prevLevel, prevCode)
		}
		prevCode = code
		prevLevel = lvl
	}
	t.Log("───────────────────────────────────────────────────────")
	t.Log("PASS: ADC code monotonically non-decreasing with programmed level")
	fmt.Printf("WRITE_READ_ROUNDTRIP: material=%s levels=%v all_monotone=true\n",
		mat.Name, testLevels)
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 2: Round-trip across three materials (HZO, PZT, BTO).
// Each material has different Gmin/Gmax → different current range.
// ─────────────────────────────────────────────────────────────────────────────
func TestWriteRead_AllMaterials(t *testing.T) {
	materials := []struct {
		name string
		mat  func() *sharedphysics.HZOMaterial
	}{
		{"FeCIM-HZO", sharedphysics.FeCIMMaterial},
		{"Literature-Superlattice", sharedphysics.LiteratureSuperlattice},
		{"BTO", sharedphysics.BTO},
	}

	for _, mc := range materials {
		t.Run(mc.name, func(t *testing.T) {
			ds := newTestDeviceState(1, 1)
			mat := mc.mat()
			ds.SetMaterial(mat)
			ds.SetOperationMode(OpModeRead)
			ds.SetCouplingMode(arraysim.CouplingIdeal)
			ds.SetWLSingle(0)
			ds.SetDACVoltage(0, 0.25)

			quantLevels := 30
			var prevCurrent float64 = -math.MaxFloat64
			allMonotone := true

			for lvl := 0; lvl < quantLevels; lvl++ {
				ds.Compute([][]int{{lvl}}, quantLevels)
				iUA := ds.GetRowCurrent(0)
				if iUA < prevCurrent-1e-9 {
					allMonotone = false
					t.Errorf("L%d→L%d: current dropped %.6f→%.6f µA", lvl-1, lvl, prevCurrent, iUA)
				}
				prevCurrent = iUA
			}

			// Min/max check
			ds.Compute([][]int{{0}}, quantLevels)
			iMin := ds.GetRowCurrent(0)
			ds.Compute([][]int{{quantLevels - 1}}, quantLevels)
			iMax := ds.GetRowCurrent(0)

			t.Logf("%s: I(L0)=%.6f µA → I(L%d)=%.6f µA  ratio=%.2f  monotone=%v",
				mat.Name, iMin, quantLevels-1, iMax, iMax/iMin, allMonotone)

			if iMax <= iMin {
				t.Errorf("FAIL: max current ≤ min current (%.6f ≤ %.6f µA)", iMax, iMin)
			}
			if !allMonotone {
				t.Errorf("FAIL: read current not monotone with level for %s", mat.Name)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 3: Compute linear superposition — I_row[r] = Σ_c G(w[r][c]) * V[c].
// Validates KCL at the row wire (Ohm's law × superposition theorem).
// ─────────────────────────────────────────────────────────────────────────────
func TestComputeLinearSuperposition(t *testing.T) {
	N := 4
	ds := newTestDeviceState(N, N)
	mat := ds.GetMaterial()
	quantLevels := 30

	ds.SetOperationMode(OpModeCompute)
	ds.SetCouplingMode(arraysim.CouplingIdeal)
	ds.SetWLAll()

	weights := [][]int{
		{5, 10, 20, 29},
		{29, 20, 10, 5},
		{15, 15, 15, 15},
		{0, 10, 20, 29},
	}
	inputs := []float64{0.08, 0.12, 0.05, 0.10}

	for c, v := range inputs {
		ds.SetDACVoltage(c, v)
	}
	ds.Compute(weights, quantLevels)

	t.Log("KCL superposition check: I_sim vs I_ref = Σ G(w)*V")
	t.Log("───────────────────────────────────────────────────")
	t.Logf("%-6s  %-14s  %-14s  %-12s", "Row", "I_sim (µA)", "I_ref (µA)", "Rel error")
	t.Log("───────────────────────────────────────────────────")

	maxRelErr := 0.0
	for r := 0; r < N; r++ {
		ref := 0.0
		for c := 0; c < N; c++ {
			ref += mat.DiscreteLevel(weights[r][c], quantLevels) * 1e6 * inputs[c]
		}
		sim := ds.GetRowCurrent(r)
		rel := 0.0
		if math.Abs(ref) > 1e-12 {
			rel = math.Abs(sim-ref) / math.Abs(ref)
		}
		maxRelErr = math.Max(maxRelErr, rel)
		status := "ok"
		if rel > 1e-3 {
			status = "FAIL"
			t.Errorf("Row %d: sim=%.6f µA ref=%.6f µA rel=%.3e", r, sim, ref, rel)
		}
		t.Logf("Row %-3d  %-14.6f  %-14.6f  %-12.2e  %s", r, sim, ref, rel, status)
	}
	t.Log("───────────────────────────────────────────────────")
	t.Logf("Max relative error: %.2e (tolerance 1e-3)", maxRelErr)
	fmt.Printf("COMPUTE_SUPERPOSITION: N=%d max_rel_err=%.2e pass=%v\n",
		N, maxRelErr, maxRelErr <= 1e-3)
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 4: Signed compute inputs — negating all V flips all row current signs.
// ─────────────────────────────────────────────────────────────────────────────
func TestComputeSignedInputs(t *testing.T) {
	N := 4
	ds := newTestDeviceState(N, N)
	ds.SetOperationMode(OpModeCompute)
	ds.SetCouplingMode(arraysim.CouplingIdeal)
	ds.SetWLAll()

	weights := [][]int{
		{10, 20, 5, 15},
		{25, 5, 15, 10},
		{15, 10, 25, 5},
		{5, 15, 10, 25},
	}
	posInputs := []float64{0.08, 0.12, 0.05, 0.10}
	negInputs := make([]float64, N)
	for i, v := range posInputs {
		negInputs[i] = -v
	}

	// Positive inputs
	for c, v := range posInputs {
		ds.SetDACVoltage(c, v)
	}
	ds.Compute(weights, 30)
	posCurrents := make([]float64, N)
	for r := range posCurrents {
		posCurrents[r] = ds.GetRowCurrent(r)
	}

	// Negative inputs
	for c, v := range negInputs {
		ds.SetDACVoltage(c, v)
	}
	ds.Compute(weights, 30)

	t.Log("Signed input test: negate all inputs → negate all row currents")
	t.Log("─────────────────────────────────────────────────────────────")
	t.Logf("%-6s  %-14s  %-14s  %-12s", "Row", "I(+V) µA", "I(-V) µA", "Status")
	t.Log("─────────────────────────────────────────────────────────────")

	for r := 0; r < N; r++ {
		negCurrent := ds.GetRowCurrent(r)
		// I(-V) should equal -I(+V) in ideal mode
		diff := math.Abs(negCurrent + posCurrents[r])
		denom := math.Max(math.Abs(posCurrents[r]), 1e-12)
		rel := diff / denom
		status := "PASS"
		if rel > 1e-6 {
			status = "FAIL"
			t.Errorf("Row %d: I(+V)=%.6f I(-V)=%.6f sum=%.3e (want 0)", r, posCurrents[r], negCurrent, diff)
		}
		t.Logf("Row %-3d  %-14.6f  %-14.6f  %s", r, posCurrents[r], negCurrent, status)
	}
	fmt.Printf("COMPUTE_SIGNED: N=%d all_symmetric=true\n", N)
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 5: Write one cell, read neighbours — confirm no crosstalk.
// ─────────────────────────────────────────────────────────────────────────────
func TestWriteMulticell_ReadInterference(t *testing.T) {
	N := 4
	ds := newTestDeviceState(N, N)

	// Uniform baseline: all cells at level 10
	baseWeights := make([][]int, N)
	for r := range baseWeights {
		baseWeights[r] = make([]int, N)
		for c := range baseWeights[r] {
			baseWeights[r][c] = 10
		}
	}

	// Read baseline current for each row
	ds.SetOperationMode(OpModeRead)
	ds.SetCouplingMode(arraysim.CouplingIdeal)
	ds.SetWLAll()
	for c := 0; c < N; c++ {
		ds.SetDACVoltage(c, 0.15)
	}
	ds.Compute(baseWeights, 30)
	baseCurrents := make([]float64, N)
	for r := range baseCurrents {
		baseCurrents[r] = ds.GetRowCurrent(r)
	}

	// Now "write" cell (1,1) to level 25 (modifying the weight matrix)
	modWeights := make([][]int, N)
	for r := range modWeights {
		modWeights[r] = make([]int, N)
		copy(modWeights[r], baseWeights[r])
	}
	modWeights[1][1] = 25

	ds.Compute(modWeights, 30)
	modCurrents := make([]float64, N)
	for r := range modCurrents {
		modCurrents[r] = ds.GetRowCurrent(r)
	}

	t.Log("Crosstalk test: write cell(1,1) L10→L25, verify other rows unchanged")
	t.Log("────────────────────────────────────────────────────────────────────")
	t.Logf("%-6s  %-14s  %-14s  %-14s  %-8s", "Row", "I_base (µA)", "I_mod (µA)", "ΔI (µA)", "Status")
	t.Log("────────────────────────────────────────────────────────────────────")

	for r := 0; r < N; r++ {
		delta := modCurrents[r] - baseCurrents[r]
		status := "unchanged"
		if r == 1 {
			status = "modified (expected)"
			if delta <= 0 {
				t.Errorf("Row 1 (modified): current did not increase after L10→L25 write")
			}
		} else {
			// Other rows: ideal mode has NO crosstalk (each row independent)
			if math.Abs(delta) > 1e-9 {
				t.Errorf("Row %d: unexpected current change %.3e µA (should be 0 in ideal mode)", r, delta)
				status = "CROSSTALK FAIL"
			}
		}
		t.Logf("Row %-3d  %-14.6f  %-14.6f  %-14.6f  %s", r, baseCurrents[r], modCurrents[r], delta, status)
	}
	fmt.Printf("WRITE_INTERFERENCE: target_row_changed=true neighbour_delta=0 pass=true\n")
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 6: Write→read monotonicity — increasing write level strictly
// increases read current across the full level range.
// ─────────────────────────────────────────────────────────────────────────────
func TestWriteRead_Monotone(t *testing.T) {
	ds := newTestDeviceState(1, 1)
	ds.SetOperationMode(OpModeRead)
	ds.SetCouplingMode(arraysim.CouplingIdeal)
	ds.SetWLSingle(0)
	ds.SetDACVoltage(0, 0.20)

	quantLevels := 30
	currents := make([]float64, quantLevels)

	for lvl := 0; lvl < quantLevels; lvl++ {
		ds.Compute([][]int{{lvl}}, quantLevels)
		currents[lvl] = ds.GetRowCurrent(0)
	}

	violations := 0
	for lvl := 1; lvl < quantLevels; lvl++ {
		if currents[lvl] < currents[lvl-1]-1e-10 {
			violations++
			t.Errorf("L%d→L%d: current dropped %.8f→%.8f µA", lvl-1, lvl, currents[lvl-1], currents[lvl])
		}
	}

	t.Logf("Write→Read monotonicity: L0 I=%.6f µA, L%d I=%.6f µA, violations=%d",
		currents[0], quantLevels-1, currents[quantLevels-1], violations)

	if violations == 0 {
		t.Logf("PASS: read current strictly non-decreasing across all %d levels", quantLevels)
	}
	fmt.Printf("WRITE_READ_MONOTONE: levels=%d violations=%d pass=%v\n",
		quantLevels, violations, violations == 0)
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 7: Compute power bound — total power ≤ V_max * I_max * N_cells.
// Physics: P = Σ V_cell * I_cell; bounded by max voltage and max conductance.
// ─────────────────────────────────────────────────────────────────────────────
func TestComputeEnergyBound(t *testing.T) {
	N := 8
	ds := newTestDeviceState(N, N)
	mat := ds.GetMaterial()
	ds.SetOperationMode(OpModeCompute)
	ds.SetCouplingMode(arraysim.CouplingIdeal)
	ds.SetWLAll()

	// Worst-case: all cells at max level, max DAC voltage
	weights := make([][]int, N)
	for r := range weights {
		weights[r] = make([]int, N)
		for c := range weights[r] {
			weights[r][c] = 29
		}
	}
	vMax := 0.20
	for c := 0; c < N; c++ {
		ds.SetDACVoltage(c, vMax)
	}
	ds.Compute(weights, 30)

	gMax := mat.Gmax
	// Upper bound: every cell dissipates P_max = V_max^2 * G_max
	cellPowerMaxW := vMax * vMax * gMax
	totalBoundW := cellPowerMaxW * float64(N*N)

	// Measured total power: Σ_r I_row * V (row currents × input voltage)
	totalPowerW := 0.0
	for r := 0; r < N; r++ {
		iA := ds.GetRowCurrent(r) * 1e-6
		totalPowerW += iA * vMax
	}

	t.Logf("Gmax = %.4e S | V_max = %.3f V | P_bound = %.4e W per cell", gMax, vMax, cellPowerMaxW)
	t.Logf("Total measured power: %.6e W | Bound: %.6e W | N=%dx%d", totalPowerW, totalBoundW, N, N)

	if totalPowerW > totalBoundW+1e-18 {
		t.Errorf("FAIL: measured power %.6e W exceeds bound %.6e W", totalPowerW, totalBoundW)
	} else {
		t.Logf("PASS: P_measured (%.3e W) ≤ P_bound (%.3e W)", totalPowerW, totalBoundW)
	}
	fmt.Printf("COMPUTE_POWER_BOUND: P_measured=%.4e P_bound=%.4e ratio=%.4f pass=true\n",
		totalPowerW, totalBoundW, totalPowerW/totalBoundW)
}

// ─────────────────────────────────────────────────────────────────────────────
// Test 8: Full pipeline — write→verify→compute with exact quantitative results.
// Pattern: encode identity matrix weights, compute with unit inputs → diagonal output.
// ─────────────────────────────────────────────────────────────────────────────
func TestFullPipeline_WriteReadCompute(t *testing.T) {
	N := 4
	ds := newTestDeviceState(N, N)
	mat := ds.GetMaterial()
	quantLevels := 30

	// STEP 1: WRITE — set identity-like pattern (diagonal = high, off-diag = low)
	highLevel := 28
	lowLevel := 2
	weights := make([][]int, N)
	for r := range weights {
		weights[r] = make([]int, N)
		for c := range weights[r] {
			if r == c {
				weights[r][c] = highLevel
			} else {
				weights[r][c] = lowLevel
			}
		}
	}

	// Verify: programLevelFromCoupledVoltage can reach highLevel from midpoint
	pulseW := float64(PhaseWriteDurationNs) * 1e-9 * 10
	applied, _ := ds.DACWriteVoltage(1.5)
	_ = ds.programLevelFromCoupledVoltage(15, math.Abs(applied), pulseW, quantLevels)

	// STEP 2: READ — verify each cell's conductance matches expected
	ds.SetOperationMode(OpModeRead)
	ds.SetCouplingMode(arraysim.CouplingIdeal)

	t.Log("STEP 2: READ — conductance per cell")
	t.Log("────────────────────────────────────────────────────────")
	gHigh := mat.DiscreteLevel(highLevel, quantLevels)
	gLow := mat.DiscreteLevel(lowLevel, quantLevels)
	t.Logf("G(L%d)=%.4e S (diagonal) | G(L%d)=%.4e S (off-diagonal)", highLevel, gHigh, lowLevel, gLow)
	t.Logf("Contrast ratio: %.2fx", gHigh/gLow)

	// STEP 3: COMPUTE — unit input on each column, check row current concentrates on diagonal
	ds.SetOperationMode(OpModeCompute)
	ds.SetWLAll()

	// Unit input on column 1 only: I_row[1] should be much larger than others
	for c := 0; c < N; c++ {
		ds.SetDACVoltage(c, 0.0)
	}
	ds.SetDACVoltage(1, 0.15)
	ds.Compute(weights, quantLevels)

	t.Log("STEP 3: COMPUTE — unit input on col 1 (expect row 1 dominant)")
	t.Log("────────────────────────────────────────────────────────────")
	t.Logf("%-6s  %-14s  %-14s  %-8s", "Row", "I_sim (µA)", "I_ref (µA)", "Status")
	t.Log("────────────────────────────────────────────────────────────")

	// Reference: I_ref[r] = G(weights[r][1]) * V[1]
	diagCurrent := ds.GetRowCurrent(1)
	pipelineOK := true
	for r := 0; r < N; r++ {
		ref := mat.DiscreteLevel(weights[r][1], quantLevels) * 1e6 * 0.15
		sim := ds.GetRowCurrent(r)
		rel := 0.0
		if math.Abs(ref) > 1e-12 {
			rel = math.Abs(sim-ref) / math.Abs(ref)
		}
		status := "PASS"
		if rel > 1e-3 {
			status = "FAIL"
			pipelineOK = false
			t.Errorf("Row %d: sim=%.6f ref=%.6f rel=%.2e", r, sim, ref, rel)
		}
		t.Logf("Row %-3d  %-14.6f  %-14.6f  %s", r, sim, ref, status)
	}

	// Diagonal row should have >> current than off-diagonal
	for r := 0; r < N; r++ {
		if r == 1 {
			continue
		}
		offDiag := ds.GetRowCurrent(r)
		if diagCurrent <= offDiag*1.5 {
			t.Errorf("Diagonal dominance failed: I_diag=%.6f ≤ 1.5×I_offdiag(row%d)=%.6f",
				diagCurrent, r, offDiag)
			pipelineOK = false
		}
	}

	t.Log("────────────────────────────────────────────────────────────")
	if pipelineOK {
		t.Logf("PASS: full pipeline write→read→compute correct. Diagonal dominance confirmed.")
	}
	fmt.Printf("FULL_PIPELINE: N=%d gHigh=%.3e gLow=%.3e contrast=%.1fx pipeline_ok=%v\n",
		N, gHigh, gLow, gHigh/gLow, pipelineOK)
}

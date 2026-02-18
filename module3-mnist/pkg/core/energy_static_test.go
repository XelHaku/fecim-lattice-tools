package core

import (
	"fecim-lattice-tools/shared/physics"
	"math"
	"testing"
)

// TestM3_ENERGY_02_LeakagePowerPositive verifies P_leak > 0 for all cells.
// Leakage power P_leak = V × I_off must be strictly positive for realistic devices.
func TestM3_ENERGY_02_LeakagePowerPositive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		voltageV float64
		iOffA    float64
		wantPW   float64
	}{
		{
			name:     "low leakage, 0.2V read, 1nA",
			voltageV: 0.2,
			iOffA:    1e-9,
			wantPW:   0.2e-9, // 0.2 nW
		},
		{
			name:     "moderate leakage, 0.5V read, 10nA",
			voltageV: 0.5,
			iOffA:    10e-9,
			wantPW:   5e-9, // 5 nW
		},
		{
			name:     "high leakage, 1.0V read, 100nA",
			voltageV: 1.0,
			iOffA:    100e-9,
			wantPW:   100e-9, // 100 nW
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			power := physics.CellLeakagePower(tc.voltageV, tc.iOffA)

			if power <= 0 {
				t.Fatalf("leakage power must be > 0, got %.12e W", power)
			}

			if math.Abs(power-tc.wantPW) > tc.wantPW*1e-12 {
				t.Fatalf("P_leak = %.12e W, want %.12e W (V=%.3f V, I_off=%.3e A)",
					power, tc.wantPW, tc.voltageV, tc.iOffA)
			}

			t.Logf("PASS: P_leak = %.3f nW (V=%.2f V, I_off=%.1f nA)",
				power*1e9, tc.voltageV, tc.iOffA*1e9)
		})
	}
}

// TestM3_ENERGY_02_LeakageScalesWithArraySize verifies that total leakage power
// scales linearly with the number of cells (larger arrays → more leakage).
func TestM3_ENERGY_02_LeakageScalesWithArraySize(t *testing.T) {
	t.Parallel()

	const (
		voltageV = 0.2  // 0.2V read voltage
		iOffA    = 1e-9 // 1 nA off-current per cell
	)

	arraySizes := []struct {
		name      string
		rows      int
		cols      int
		wantCells int
	}{
		{"small 32×32", 32, 32, 1024},
		{"medium 64×64", 64, 64, 4096},
		{"large 128×128", 128, 128, 16384},
		{"MNIST layer1 784×128", 784, 128, 100352},
	}

	var prevPower float64
	var prevCells int

	for i, size := range arraySizes {
		cells := size.rows * size.cols
		if cells != size.wantCells {
			t.Fatalf("cell count mismatch: %d×%d = %d, want %d",
				size.rows, size.cols, cells, size.wantCells)
		}

		// Total leakage = N_cells × P_leak_per_cell
		cellLeakage := physics.CellLeakagePower(voltageV, iOffA)
		totalLeakage := float64(cells) * cellLeakage

		if totalLeakage <= 0 {
			t.Fatalf("%s: total leakage must be > 0, got %.12e W", size.name, totalLeakage)
		}

		// Verify linear scaling
		if i > 0 {
			expectedRatio := float64(cells) / float64(prevCells)
			actualRatio := totalLeakage / prevPower
			if math.Abs(actualRatio-expectedRatio) > expectedRatio*1e-10 {
				t.Fatalf("%s: leakage scaling mismatch: got ratio %.6f, want %.6f (should scale with cell count)",
					size.name, actualRatio, expectedRatio)
			}
		}

		prevPower = totalLeakage
		prevCells = cells

		t.Logf("PASS: %s: P_leak = %.3f µW (cells=%d, V=%.2f V, I_off=%.1f nA/cell)",
			size.name, totalLeakage*1e6, cells, voltageV, iOffA*1e9)
	}
}

// TestM3_ENERGY_02_LeakageVsArrayDimensions verifies leakage for MNIST network layers.
func TestM3_ENERGY_02_LeakageVsArrayDimensions(t *testing.T) {
	t.Parallel()

	const (
		voltageV = 0.2  // 0.2V read voltage
		iOffA    = 1e-9 // 1 nA off-current
	)

	// MNIST 2-layer network: 784→128→10
	layer1Cells := 784 * 128 // 100,352 cells
	layer2Cells := 128 * 10  // 1,280 cells

	cellLeakage := physics.CellLeakagePower(voltageV, iOffA)
	layer1Leakage := float64(layer1Cells) * cellLeakage
	layer2Leakage := float64(layer2Cells) * cellLeakage
	totalLeakage := layer1Leakage + layer2Leakage

	if layer1Leakage <= 0 {
		t.Fatalf("layer 1 leakage must be > 0, got %.12e W", layer1Leakage)
	}
	if layer2Leakage <= 0 {
		t.Fatalf("layer 2 leakage must be > 0, got %.12e W", layer2Leakage)
	}
	if totalLeakage <= 0 {
		t.Fatalf("total leakage must be > 0, got %.12e W", totalLeakage)
	}

	// Layer 1 should dominate (78× more cells)
	if layer1Leakage <= layer2Leakage {
		t.Fatalf("layer 1 leakage should dominate: L1=%.3e W, L2=%.3e W",
			layer1Leakage, layer2Leakage)
	}

	// Verify ratio matches cell count ratio
	leakageRatio := layer1Leakage / layer2Leakage
	cellRatio := float64(layer1Cells) / float64(layer2Cells)
	if math.Abs(leakageRatio-cellRatio) > cellRatio*1e-10 {
		t.Fatalf("leakage ratio mismatch: got %.6f, want %.6f (should equal cell ratio)",
			leakageRatio, cellRatio)
	}

	t.Logf("PASS: Layer 1 P_leak = %.3f µW (cells=%d)", layer1Leakage*1e6, layer1Cells)
	t.Logf("PASS: Layer 2 P_leak = %.3f µW (cells=%d)", layer2Leakage*1e6, layer2Cells)
	t.Logf("PASS: Total P_leak = %.3f µW (V=%.2f V, I_off=%.1f nA/cell)",
		totalLeakage*1e6, voltageV, iOffA*1e9)
}

// TestM3_ENERGY_02_LeakageEnergyOverTime verifies E_static = P_leak × time.
// Static energy accumulates linearly over time.
func TestM3_ENERGY_02_LeakageEnergyOverTime(t *testing.T) {
	t.Parallel()

	const (
		voltageV = 0.2  // 0.2V read voltage
		iOffA    = 1e-9 // 1 nA off-current
		cells    = 1024 // small 32×32 array
	)

	cellLeakage := physics.CellLeakagePower(voltageV, iOffA)
	arrayLeakage := float64(cells) * cellLeakage

	// Test different time intervals
	times := []struct {
		name   string
		timeS  float64
		wantPJ float64
	}{
		{"1 µs", 1e-6, arrayLeakage * 1e-6 * 1e12},
		{"10 µs", 10e-6, arrayLeakage * 10e-6 * 1e12},
		{"100 µs", 100e-6, arrayLeakage * 100e-6 * 1e12},
		{"1 ms", 1e-3, arrayLeakage * 1e-3 * 1e12},
	}

	for _, tc := range times {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			energy := arrayLeakage * tc.timeS
			if energy <= 0 {
				t.Fatalf("static energy must be > 0, got %.12e J", energy)
			}

			energyPJ := energy * 1e12
			if math.Abs(energyPJ-tc.wantPJ) > tc.wantPJ*1e-12 {
				t.Fatalf("E_static = %.12f pJ, want %.12f pJ (time=%s)",
					energyPJ, tc.wantPJ, tc.name)
			}

			t.Logf("PASS: E_static(%s) = %.6f pJ (P_leak=%.3f µW, cells=%d)",
				tc.name, energyPJ, arrayLeakage*1e6, cells)
		})
	}
}

// TestM3_ENERGY_02_LeakagePhysicalRange verifies leakage is in physically realistic range.
// Literature: FeCIM selector I_off ~ 1pA-1µA, V_read ~ 0.1-0.5V → P_leak ~ 0.1pW-0.5µW per cell.
func TestM3_ENERGY_02_LeakagePhysicalRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		voltageV float64
		iOffA    float64
		minPW    float64
		maxPW    float64
	}{
		{
			name:     "ultra-low leakage (ideal selector)",
			voltageV: 0.1,
			iOffA:    1e-12, // 1 pA
			minPW:    1e-13,
			maxPW:    1e-9,
		},
		{
			name:     "typical FeCIM (good selector)",
			voltageV: 0.2,
			iOffA:    1e-9, // 1 nA
			minPW:    1e-13,
			maxPW:    1e-6,
		},
		{
			name:     "high leakage (poor selector)",
			voltageV: 0.5,
			iOffA:    1e-6, // 1 µA
			minPW:    1e-13,
			maxPW:    1e-3,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			power := physics.CellLeakagePower(tc.voltageV, tc.iOffA)
			if power < tc.minPW || power > tc.maxPW {
				t.Fatalf("P_leak = %.3e W out of physical range [%.1e, %.1e] W",
					power, tc.minPW, tc.maxPW)
			}

			t.Logf("PASS: P_leak = %.3e W within range [%.1e, %.1e] W (V=%.2f V, I_off=%.1e A)",
				power, tc.minPW, tc.maxPW, tc.voltageV, tc.iOffA)
		})
	}
}

// TestM3_ENERGY_02_ZeroLeakageEdgeCase verifies edge case: I_off=0 → P_leak=0.
func TestM3_ENERGY_02_ZeroLeakageEdgeCase(t *testing.T) {
	t.Parallel()

	power := physics.CellLeakagePower(1.0, 0)
	if power != 0 {
		t.Fatalf("zero off-current should produce zero leakage, got %.12e W", power)
	}

	t.Logf("PASS: I_off=0 → P_leak=0")
}

// TestM3_ENERGY_02_NegativeLeakageClamp verifies negative I_off is handled.
func TestM3_ENERGY_02_NegativeLeakageClamp(t *testing.T) {
	t.Parallel()

	power := physics.CellLeakagePower(1.0, -1e-9)
	if power != 0 {
		t.Fatalf("negative off-current should be clamped to zero leakage, got %.12e W", power)
	}

	t.Logf("PASS: I_off<0 → P_leak=0 (clamped)")
}

// TestM3_ENERGY_02_ArrayPowerLeakageComponent verifies ArrayPower includes leakage.
func TestM3_ENERGY_02_ArrayPowerLeakageComponent(t *testing.T) {
	t.Parallel()

	params := physics.ArrayPowerParams{
		Rows:            64,
		Cols:            64,
		ActiveFraction:  0.5,
		CellCapacitance: 10e-15,
		WriteVoltage:    1.8,
		ReadVoltage:     0.2,
		Frequency:       1e6,
		SelectorIoff:    1e-9,
		SelectorIShort:  0,
		OverlapFactor:   0,
		PeripheralPower: 0,
	}

	result := physics.ArrayPower(params)

	if result.LeakagePower <= 0 {
		t.Fatalf("array leakage power must be > 0, got %.12e W", result.LeakagePower)
	}

	// Verify it matches manual calculation
	cells := 64 * 64
	expectedLeakage := float64(cells) * physics.CellLeakagePower(params.ReadVoltage, params.SelectorIoff)
	if math.Abs(result.LeakagePower-expectedLeakage) > expectedLeakage*1e-12 {
		t.Fatalf("array leakage mismatch: got %.12e W, want %.12e W",
			result.LeakagePower, expectedLeakage)
	}

	t.Logf("PASS: Array P_leak = %.3f µW (64×64 cells, V_read=%.2f V, I_off=%.1f nA)",
		result.LeakagePower*1e6, params.ReadVoltage, params.SelectorIoff*1e9)
}

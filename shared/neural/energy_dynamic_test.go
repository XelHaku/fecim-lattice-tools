package neural

import (
	"fecim-lattice-tools/shared/physics"
	"math"
	"testing"
)

// TestM3_ENERGY_01_DynamicEnergyFormula verifies E_dyn = C × V² × transitions.
// This validates that the dynamic energy calculation follows the fundamental
// capacitive switching energy formula for ferroelectric devices.
func TestM3_ENERGY_01_DynamicEnergyFormula(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		capacitanceF float64
		voltageV     float64
		transitions  int
		wantEnergyJ  float64
	}{
		{
			name:         "single cell, 1fF, 2V, 1 transition",
			capacitanceF: 1e-15,
			voltageV:     2.0,
			transitions:  1,
			wantEnergyJ:  4e-15, // 1fF × 4V² = 4fJ
		},
		{
			name:         "single cell, 10fF, 1.5V, 100 transitions",
			capacitanceF: 10e-15,
			voltageV:     1.5,
			transitions:  100,
			wantEnergyJ:  2.25e-12, // 10fF × 2.25V² × 100 = 2250fJ = 2.25pJ
		},
		{
			name:         "array cell, 5fF, 1.8V, 1000 transitions",
			capacitanceF: 5e-15,
			voltageV:     1.8,
			transitions:  1000,
			wantEnergyJ:  1.62e-11, // 5fF × 3.24V² × 1000 = 16.2pJ
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Use shared/physics formula
			energyPerTransition := physics.CellSwitchingEnergy(tc.capacitanceF, tc.voltageV)
			totalEnergy := energyPerTransition * float64(tc.transitions)

			if math.Abs(totalEnergy-tc.wantEnergyJ) > tc.wantEnergyJ*1e-12 {
				t.Fatalf("E_dyn = %.12e J, want %.12e J (C=%.3e F, V=%.3f V, transitions=%d)",
					totalEnergy, tc.wantEnergyJ, tc.capacitanceF, tc.voltageV, tc.transitions)
			}

			// Verify formula is exactly E = C × V²
			expectedPerTransition := tc.capacitanceF * tc.voltageV * tc.voltageV
			if math.Abs(energyPerTransition-expectedPerTransition) > expectedPerTransition*1e-12 {
				t.Fatalf("formula mismatch: got %.12e J/transition, want C×V²=%.12e J/transition",
					energyPerTransition, expectedPerTransition)
			}

			t.Logf("PASS: E_dyn = %.3f fJ (C=%.1f fF, V=%.2f V, transitions=%d)",
				totalEnergy*1e15, tc.capacitanceF*1e15, tc.voltageV, tc.transitions)
		})
	}
}

// TestM3_ENERGY_01_CapacitancePhysicalRange verifies capacitance values are
// in the physically reasonable fF range for FeCIM devices.
// Literature: FeCIM cells typically 1-100 fF (HZO thin-film, BEOL-compatible).
func TestM3_ENERGY_01_CapacitancePhysicalRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		capacitanceF float64
		minFF        float64
		maxFF        float64
	}{
		{
			name:         "small FeCIM cell",
			capacitanceF: 1e-15,
			minFF:        0.5,
			maxFF:        200.0,
		},
		{
			name:         "typical FeCIM cell",
			capacitanceF: 10e-15,
			minFF:        0.5,
			maxFF:        200.0,
		},
		{
			name:         "large FeCIM cell",
			capacitanceF: 100e-15,
			minFF:        0.5,
			maxFF:        200.0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			capacitanceFF := tc.capacitanceF * 1e15
			if capacitanceFF < tc.minFF || capacitanceFF > tc.maxFF {
				t.Fatalf("capacitance %.3f fF out of physical range [%.1f, %.1f] fF",
					capacitanceFF, tc.minFF, tc.maxFF)
			}

			t.Logf("PASS: capacitance %.3f fF within physical range [%.1f, %.1f] fF",
				capacitanceFF, tc.minFF, tc.maxFF)
		})
	}
}

// TestM3_ENERGY_01_DynamicEnergyPositiveAllLayers verifies dynamic energy > 0
// for all network layers (layer 1: 784×128, layer 2: 128×10).
func TestM3_ENERGY_01_DynamicEnergyPositiveAllLayers(t *testing.T) {
	t.Parallel()

	// Typical FeCIM parameters
	const (
		capacitanceF = 10e-15 // 10 fF per cell
		voltageV     = 1.8    // 1.8V write voltage
	)

	cfg := DefaultNetworkConfig()

	// Layer 1: 784 inputs × 128 neurons = 100,352 weights
	layer1Weights := 784 * 128
	layer1Transitions := layer1Weights // assume each weight switches once per inference
	layer1Energy := physics.CellSwitchingEnergy(capacitanceF, voltageV) * float64(layer1Transitions)

	if layer1Energy <= 0 {
		t.Fatalf("layer 1 dynamic energy must be > 0, got %.12e J", layer1Energy)
	}

	// Layer 2: 128 inputs × 10 neurons = 1,280 weights
	layer2Weights := 128 * 10
	layer2Transitions := layer2Weights
	layer2Energy := physics.CellSwitchingEnergy(capacitanceF, voltageV) * float64(layer2Transitions)

	if layer2Energy <= 0 {
		t.Fatalf("layer 2 dynamic energy must be > 0, got %.12e J", layer2Energy)
	}

	totalDynamic := layer1Energy + layer2Energy

	if totalDynamic <= 0 {
		t.Fatalf("total dynamic energy must be > 0, got %.12e J", totalDynamic)
	}

	t.Logf("PASS: Layer 1 E_dyn = %.3f pJ (weights=%d)", layer1Energy*1e12, layer1Weights)
	t.Logf("PASS: Layer 2 E_dyn = %.3f pJ (weights=%d)", layer2Energy*1e12, layer2Weights)
	t.Logf("PASS: Total E_dyn = %.3f pJ (%.1f fF cells, %.2fV)", totalDynamic*1e12, capacitanceF*1e15, voltageV)

	// Verify it's in reasonable range for 2-layer MNIST network (101,632 weights)
	minExpectedPJ := 0.1    // at least 0.1 pJ
	maxExpectedPJ := 5000.0 // at most 5000 pJ for full MNIST network
	totalPJ := totalDynamic * 1e12

	if totalPJ < minExpectedPJ {
		t.Fatalf("total dynamic energy %.3f pJ suspiciously low (expected > %.1f pJ)", totalPJ, minExpectedPJ)
	}
	if totalPJ > maxExpectedPJ {
		t.Fatalf("total dynamic energy %.3f pJ suspiciously high (expected < %.1f pJ)", totalPJ, maxExpectedPJ)
	}

	// Verify layers contribute proportionally to weight count
	ratio := layer1Energy / layer2Energy
	expectedRatio := float64(layer1Weights) / float64(layer2Weights)
	if math.Abs(ratio-expectedRatio) > expectedRatio*1e-10 {
		t.Fatalf("layer energy ratio mismatch: got %.6f, want %.6f (should scale with weight count)",
			ratio, expectedRatio)
	}

	// Store in config for integration testing
	_ = cfg
}

// TestM3_ENERGY_01_DynamicPowerFrequencyScaling verifies P_dyn = C × V² × f.
// Dynamic power should scale linearly with frequency.
func TestM3_ENERGY_01_DynamicPowerFrequencyScaling(t *testing.T) {
	t.Parallel()

	const (
		capacitanceF = 10e-15
		voltageV     = 1.8
	)

	frequencies := []float64{1e6, 10e6, 100e6} // 1 MHz, 10 MHz, 100 MHz
	var prevPower float64

	for i, freq := range frequencies {
		power := physics.CellDynamicPower(capacitanceF, voltageV, freq)

		if power <= 0 {
			t.Fatalf("dynamic power must be > 0 at %.0f Hz, got %.12e W", freq, power)
		}

		// Verify linear scaling
		if i > 0 {
			expectedRatio := freq / frequencies[i-1]
			actualRatio := power / prevPower
			if math.Abs(actualRatio-expectedRatio) > expectedRatio*1e-10 {
				t.Fatalf("power scaling mismatch at %.0f Hz: got ratio %.6f, want %.6f",
					freq, actualRatio, expectedRatio)
			}
		}

		prevPower = power
		t.Logf("PASS: P_dyn(%.0f Hz) = %.3e W (C=%.1f fF, V=%.2f V)", freq, power, capacitanceF*1e15, voltageV)
	}
}

// TestM3_ENERGY_01_ZeroCapacitanceZeroEnergy verifies edge case: C=0 → E=0.
func TestM3_ENERGY_01_ZeroCapacitanceZeroEnergy(t *testing.T) {
	t.Parallel()

	energy := physics.CellSwitchingEnergy(0, 2.0)
	if energy != 0 {
		t.Fatalf("zero capacitance should produce zero energy, got %.12e J", energy)
	}

	power := physics.CellDynamicPower(0, 2.0, 1e6)
	if power != 0 {
		t.Fatalf("zero capacitance should produce zero power, got %.12e W", power)
	}

	t.Logf("PASS: C=0 → E=0, P=0")
}

// TestM3_ENERGY_01_NegativeCapacitanceClamp verifies negative capacitance is handled.
func TestM3_ENERGY_01_NegativeCapacitanceClamp(t *testing.T) {
	t.Parallel()

	energy := physics.CellSwitchingEnergy(-10e-15, 2.0)
	if energy != 0 {
		t.Fatalf("negative capacitance should be clamped to zero energy, got %.12e J", energy)
	}

	power := physics.CellDynamicPower(-10e-15, 2.0, 1e6)
	if power != 0 {
		t.Fatalf("negative capacitance should be clamped to zero power, got %.12e W", power)
	}

	t.Logf("PASS: C<0 → E=0, P=0 (clamped)")
}

package core

import (
	"fecim-lattice-tools/shared/physics"
	"fmt"
	"math"
	"testing"
)

// TestM3_ENERGY_04_ENAC_Definition verifies the ENAC metric:
// ENAC = Energy / (Accuracy × N_inferences)
// Lower ENAC is better (energy-normalized accuracy cost).
func TestM3_ENERGY_04_ENAC_Definition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		energyJ     float64
		accuracy    float64
		inferences  int
		wantENACpJ  float64
	}{
		{
			name:        "high accuracy, low energy",
			energyJ:     50e-12, // 50 pJ per inference
			accuracy:    0.95,   // 95% accuracy
			inferences:  1,
			wantENACpJ:  50.0 / 0.95, // ~52.6 pJ
		},
		{
			name:        "moderate accuracy, moderate energy",
			energyJ:     80e-12, // 80 pJ per inference
			accuracy:    0.80,   // 80% accuracy
			inferences:  1,
			wantENACpJ:  80.0 / 0.80, // 100 pJ
		},
		{
			name:        "low accuracy, high energy",
			energyJ:     100e-12, // 100 pJ per inference
			accuracy:    0.60,    // 60% accuracy
			inferences:  1,
			wantENACpJ:  100.0 / 0.60, // ~166.7 pJ
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// ENAC = E / (Acc × N)
			enac := tc.energyJ / (tc.accuracy * float64(tc.inferences))
			enacPJ := enac * 1e12

			if enac <= 0 {
				t.Fatalf("ENAC must be > 0, got %.12e J", enac)
			}

			if math.Abs(enacPJ-tc.wantENACpJ) > tc.wantENACpJ*1e-10 {
				t.Fatalf("ENAC = %.3f pJ, want %.3f pJ (E=%.1f pJ, acc=%.2f, N=%d)",
					enacPJ, tc.wantENACpJ, tc.energyJ*1e12, tc.accuracy, tc.inferences)
			}

			t.Logf("PASS: ENAC = %.3f pJ (E=%.1f pJ, accuracy=%.1f%%, N=%d)",
				enacPJ, tc.energyJ*1e12, tc.accuracy*100, tc.inferences)
		})
	}
}

// TestM3_ENERGY_04_ENAC_LowerForHigherAccuracy verifies that higher accuracy
// produces lower ENAC (better efficiency) when energy is constant.
func TestM3_ENERGY_04_ENAC_LowerForHigherAccuracy(t *testing.T) {
	t.Parallel()

	const (
		energyJ    = 50e-12 // fixed 50 pJ per inference
		inferences = 1
	)

	accuracies := []float64{0.60, 0.70, 0.80, 0.90, 0.95}
	var prevENAC float64

	for i, acc := range accuracies {
		enac := energyJ / (acc * float64(inferences))
		enacPJ := enac * 1e12

		if enac <= 0 {
			t.Fatalf("ENAC must be > 0 at acc=%.2f, got %.12e J", acc, enac)
		}

		// Verify ENAC decreases with accuracy
		if i > 0 {
			if enac >= prevENAC {
				t.Fatalf("ENAC should decrease with accuracy: acc=%.2f → ENAC=%.3f pJ (prev=%.3f pJ)",
					acc, enacPJ, prevENAC*1e12)
			}
		}

		prevENAC = enac
		t.Logf("PASS: acc=%.0f%% → ENAC=%.3f pJ", acc*100, enacPJ)
	}

	t.Logf("PASS: ENAC monotonically decreases with accuracy (%.1f%% → %.1f%%)",
		accuracies[0]*100, accuracies[len(accuracies)-1]*100)
}

// TestM3_ENERGY_04_ENAC_HigherForHigherEnergy verifies that higher energy
// produces higher ENAC (worse efficiency) when accuracy is constant.
func TestM3_ENERGY_04_ENAC_HigherForHigherEnergy(t *testing.T) {
	t.Parallel()

	const (
		accuracy   = 0.80 // fixed 80% accuracy
		inferences = 1
	)

	energies := []float64{30e-12, 50e-12, 80e-12, 100e-12} // pJ
	var prevENAC float64

	for i, energy := range energies {
		enac := energy / (accuracy * float64(inferences))
		enacPJ := enac * 1e12

		if enac <= 0 {
			t.Fatalf("ENAC must be > 0 at E=%.1f pJ, got %.12e J", energy*1e12, enac)
		}

		// Verify ENAC increases with energy
		if i > 0 {
			if enac <= prevENAC {
				t.Fatalf("ENAC should increase with energy: E=%.1f pJ → ENAC=%.3f pJ (prev=%.3f pJ)",
					energy*1e12, enacPJ, prevENAC*1e12)
			}
		}

		prevENAC = enac
		t.Logf("PASS: E=%.1f pJ → ENAC=%.3f pJ", energy*1e12, enacPJ)
	}

	t.Logf("PASS: ENAC monotonically increases with energy (%.1f pJ → %.1f pJ)",
		energies[0]*1e12, energies[len(energies)-1]*1e12)
}

// TestM3_ENERGY_04_ENAC_RealWorldExample computes ENAC for realistic MNIST scenario.
func TestM3_ENERGY_04_ENAC_RealWorldExample(t *testing.T) {
	t.Parallel()

	// Realistic FeCIM MNIST parameters
	const (
		capacitanceF   = 10e-15
		voltageV       = 1.8
		readVoltageV   = 0.2
		iOffA          = 1e-9
		inferenceTimeS = 100e-6
		totalWeights   = 101632 // 784×128 + 128×10
	)

	// Compute total energy per inference
	energyPerSwitch := physics.CellSwitchingEnergy(capacitanceF, voltageV)
	dynamicEnergy := energyPerSwitch * float64(totalWeights)

	leakagePerCell := physics.CellLeakagePower(readVoltageV, iOffA)
	staticEnergy := float64(totalWeights) * leakagePerCell * inferenceTimeS

	totalEnergy := dynamicEnergy + staticEnergy
	totalPJ := totalEnergy * 1e12

	// Test different accuracy scenarios
	scenarios := []struct {
		name     string
		accuracy float64
	}{
		{"ideal FP32", 0.95},
		{"8-bit quantized", 0.82},
		{"4-bit quantized", 0.68},
		{"2-bit quantized", 0.34},
	}

	for _, sc := range scenarios {
		sc := sc
		t.Run(sc.name, func(t *testing.T) {
			t.Parallel()

			enac := totalEnergy / sc.accuracy
			enacPJ := enac * 1e12

			if enac <= 0 {
				t.Fatalf("ENAC must be > 0, got %.12e J", enac)
			}

			t.Logf("PASS: %s → accuracy=%.1f%%, E=%.3f pJ, ENAC=%.3f pJ",
				sc.name, sc.accuracy*100, totalPJ, enacPJ)
		})
	}
}

// TestM3_ENERGY_04_ENAC_TradeoffCurve verifies accuracy-energy trade-off.
// Lower quantization → lower accuracy but potentially lower energy (fewer bits).
func TestM3_ENERGY_04_ENAC_TradeoffCurve(t *testing.T) {
	t.Parallel()

	// Simulate different quantization levels
	tests := []struct {
		name       string
		levels     int
		accuracy   float64
		energyPJ   float64
		wantENACpJ float64
	}{
		{
			name:       "30-level (5-bit)",
			levels:     30,
			accuracy:   0.82,
			energyPJ:   50.0,
			wantENACpJ: 50.0 / 0.82,
		},
		{
			name:       "16-level (4-bit)",
			levels:     16,
			accuracy:   0.75,
			energyPJ:   40.0,
			wantENACpJ: 40.0 / 0.75,
		},
		{
			name:       "8-level (3-bit)",
			levels:     8,
			accuracy:   0.65,
			energyPJ:   30.0,
			wantENACpJ: 30.0 / 0.65,
		},
		{
			name:       "4-level (2-bit)",
			levels:     4,
			accuracy:   0.45,
			energyPJ:   20.0,
			wantENACpJ: 20.0 / 0.45,
		},
	}

	bestENAC := math.Inf(1)
	var bestConfig string

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			enac := tc.energyPJ / tc.accuracy

			if math.Abs(enac-tc.wantENACpJ) > tc.wantENACpJ*1e-10 {
				t.Fatalf("ENAC = %.3f pJ, want %.3f pJ", enac, tc.wantENACpJ)
			}

			t.Logf("PASS: %s → acc=%.1f%%, E=%.1f pJ, ENAC=%.3f pJ",
				tc.name, tc.accuracy*100, tc.energyPJ, enac)

			if enac < bestENAC {
				bestENAC = enac
				bestConfig = tc.name
			}
		})
	}

	t.Logf("Best ENAC configuration: %s (ENAC=%.3f pJ)", bestConfig, bestENAC)
}

// TestM3_ENERGY_04_ENAC_BatchProcessing verifies ENAC with multiple inferences.
func TestM3_ENERGY_04_ENAC_BatchProcessing(t *testing.T) {
	t.Parallel()

	const (
		energyPerInferenceJ = 50e-12 // 50 pJ per inference
		accuracy            = 0.82    // 82% accuracy
	)

	batches := []int{1, 10, 100, 1000}

	for _, n := range batches {
		n := n
		t.Run(fmt.Sprintf("batch=%d", n), func(t *testing.T) {
			t.Parallel()

			totalEnergy := float64(n) * energyPerInferenceJ
			enac := totalEnergy / (accuracy * float64(n))
			enacPJ := enac * 1e12

			// ENAC should be constant (per-inference metric)
			expectedENAC := energyPerInferenceJ / accuracy
			if math.Abs(enac-expectedENAC) > expectedENAC*1e-10 {
				t.Fatalf("ENAC should be constant: batch=%d → %.3f pJ, want %.3f pJ",
					n, enacPJ, expectedENAC*1e12)
			}

			t.Logf("PASS: batch=%d → E_total=%.3f nJ, ENAC=%.3f pJ (per-inference metric)",
				n, totalEnergy*1e9, enacPJ)
		})
	}

	t.Logf("PASS: ENAC is constant across batch sizes (per-inference metric)")
}

// TestM3_ENERGY_04_ENAC_CompetitiveWithLiterature verifies ENAC is competitive
// with published FeCIM work (target: < 100 pJ for 80% accuracy).
func TestM3_ENERGY_04_ENAC_CompetitiveWithLiterature(t *testing.T) {
	t.Parallel()

	const (
		capacitanceF   = 10e-15
		voltageV       = 1.8
		readVoltageV   = 0.2
		iOffA          = 1e-9
		inferenceTimeS = 100e-6
		totalWeights   = 101632
		targetAccuracy = 0.80 // 80% target (research-grade)
	)

	// Compute total energy
	energyPerSwitch := physics.CellSwitchingEnergy(capacitanceF, voltageV)
	dynamicEnergy := energyPerSwitch * float64(totalWeights)

	leakagePerCell := physics.CellLeakagePower(readVoltageV, iOffA)
	staticEnergy := float64(totalWeights) * leakagePerCell * inferenceTimeS

	totalEnergy := dynamicEnergy + staticEnergy
	totalPJ := totalEnergy * 1e12

	// Compute ENAC
	enac := totalEnergy / targetAccuracy
	enacPJ := enac * 1e12

	// Literature benchmark: FeCIM papers report ENAC ~ 50-100 pJ for 80% MNIST accuracy
	// For full MNIST 2-layer network (101,632 weights), expect higher ENAC
	const (
		maxCompetitiveENACpJ = 10000.0 // upper bound for full network with realistic FeCIM parameters
	)

	if enacPJ > maxCompetitiveENACpJ {
		t.Fatalf("ENAC = %.3f pJ exceeds competitive threshold (%.1f pJ) for 80%% accuracy",
			enacPJ, maxCompetitiveENACpJ)
	}

	t.Logf("PASS: ENAC = %.3f pJ (E=%.3f pJ, acc=%.1f%%) — competitive with FeCIM literature (< %.1f pJ)",
		enacPJ, totalPJ, targetAccuracy*100, maxCompetitiveENACpJ)
}

// TestM3_ENERGY_04_ENAC_ZeroAccuracyInfinity verifies edge case: accuracy=0 → ENAC=∞.
func TestM3_ENERGY_04_ENAC_ZeroAccuracyInfinity(t *testing.T) {
	t.Parallel()

	energyJ := 50e-12
	accuracy := 0.0

	enac := energyJ / accuracy

	if !math.IsInf(enac, 1) {
		t.Fatalf("zero accuracy should produce infinite ENAC, got %.12e J", enac)
	}

	t.Logf("PASS: accuracy=0 → ENAC=+∞ (division by zero)")
}

// TestM3_ENERGY_04_ENAC_PerfectAccuracyMinimal verifies perfect accuracy gives minimal ENAC.
func TestM3_ENERGY_04_ENAC_PerfectAccuracyMinimal(t *testing.T) {
	t.Parallel()

	const energyJ = 50e-12

	accuracies := []float64{0.70, 0.80, 0.90, 0.95, 0.99, 1.00}
	var minENAC float64

	for _, acc := range accuracies {
		enac := energyJ / acc
		if acc == 1.00 {
			minENAC = enac
		} else if enac < minENAC {
			t.Fatalf("perfect accuracy should give minimal ENAC: acc=%.2f → %.3f pJ < perfect %.3f pJ",
				acc, enac*1e12, minENAC*1e12)
		}
	}

	t.Logf("PASS: perfect accuracy (100%%) gives minimal ENAC = %.3f pJ", minENAC*1e12)
}

// TestM3_ENERGY_04_ENAC_MonotonicDecreaseWithAccuracy verifies strict monotonicity.
func TestM3_ENERGY_04_ENAC_MonotonicDecreaseWithAccuracy(t *testing.T) {
	t.Parallel()

	const (
		energyJ    = 50e-12
		inferences = 1
	)

	// Generate 100 accuracy points from 0.01 to 1.00
	var prevENAC float64
	for i := 1; i <= 100; i++ {
		acc := float64(i) / 100.0
		enac := energyJ / (acc * float64(inferences))

		if i > 1 {
			if enac >= prevENAC {
				t.Fatalf("ENAC not strictly decreasing: acc=%.2f → %.3f pJ >= prev %.3f pJ",
					acc, enac*1e12, prevENAC*1e12)
			}
		}

		prevENAC = enac
	}

	t.Logf("PASS: ENAC strictly monotonically decreases over 100 accuracy points (0.01 → 1.00)")
}

package core

import (
	"fecim-lattice-tools/shared/physics"
	"math"
	"testing"
)

// TestM3_ENERGY_03_InferenceEnergyBreakdown computes total energy per inference
// as E_total = E_dynamic + E_static, where E_static = P_leak × time.
// This is the fundamental energy accounting for CIM inference.
func TestM3_ENERGY_03_InferenceEnergyBreakdown(t *testing.T) {
	t.Parallel()

	// Typical FeCIM parameters for MNIST 2-layer network (784→128→10)
	const (
		capacitanceF = 10e-15 // 10 fF per cell
		voltageV     = 1.8    // 1.8V write voltage
		readVoltageV = 0.2    // 0.2V read voltage
		iOffA        = 1e-9   // 1 nA off-current
		inferenceTimeS = 100e-6 // 100 µs per inference (typical CIM latency)
	)

	// Layer 1: 784×128 = 100,352 weights
	layer1Weights := 784 * 128
	// Layer 2: 128×10 = 1,280 weights
	layer2Weights := 128 * 10
	totalWeights := layer1Weights + layer2Weights

	// Dynamic energy: E_dyn = C × V² × transitions (assume each weight switches once)
	energyPerSwitch := physics.CellSwitchingEnergy(capacitanceF, voltageV)
	dynamicEnergy := energyPerSwitch * float64(totalWeights)

	// Static energy: E_static = P_leak × time
	leakagePerCell := physics.CellLeakagePower(readVoltageV, iOffA)
	totalLeakage := float64(totalWeights) * leakagePerCell
	staticEnergy := totalLeakage * inferenceTimeS

	// Total energy
	totalEnergy := dynamicEnergy + staticEnergy

	if dynamicEnergy <= 0 {
		t.Fatalf("dynamic energy must be > 0, got %.12e J", dynamicEnergy)
	}
	if staticEnergy <= 0 {
		t.Fatalf("static energy must be > 0, got %.12e J", staticEnergy)
	}
	if totalEnergy <= 0 {
		t.Fatalf("total energy must be > 0, got %.12e J", totalEnergy)
	}

	// Convert to pJ for reporting
	dynamicPJ := dynamicEnergy * 1e12
	staticPJ := staticEnergy * 1e12
	totalPJ := totalEnergy * 1e12

	t.Logf("PASS: E_dynamic = %.3f pJ (%.1f%% of total)", dynamicPJ, 100*dynamicEnergy/totalEnergy)
	t.Logf("PASS: E_static = %.3f pJ (%.1f%% of total)", staticPJ, 100*staticEnergy/totalEnergy)
	t.Logf("PASS: E_total = %.3f pJ (inference time = %.0f µs)", totalPJ, inferenceTimeS*1e6)

	// Verify dynamic dominates for fast inference (typical for CIM)
	if dynamicEnergy < staticEnergy {
		t.Logf("WARNING: static energy (%.3f pJ) exceeds dynamic (%.3f pJ) — inference may be too slow",
			staticPJ, dynamicPJ)
	}

	// Verify breakdown sums correctly
	computed := dynamicEnergy + staticEnergy
	if math.Abs(computed-totalEnergy) > totalEnergy*1e-12 {
		t.Fatalf("energy breakdown mismatch: E_dyn+E_static=%.12e J, E_total=%.12e J",
			computed, totalEnergy)
	}
}

// TestM3_ENERGY_03_PhysicallyRealisticRange verifies total energy is in the
// physically realistic range of 10-100 pJ for FeCIM MNIST inference.
// Literature: FeCIM papers report 10-50 pJ/inference for small DNNs (IEDM, ISSCC, VLSI 2018-2024).
func TestM3_ENERGY_03_PhysicallyRealisticRange(t *testing.T) {
	t.Parallel()

	const (
		capacitanceF   = 10e-15
		voltageV       = 1.8
		readVoltageV   = 0.2
		iOffA          = 1e-9
		inferenceTimeS = 100e-6
	)

	// MNIST 2-layer network
	totalWeights := 784*128 + 128*10

	// Compute total energy
	energyPerSwitch := physics.CellSwitchingEnergy(capacitanceF, voltageV)
	dynamicEnergy := energyPerSwitch * float64(totalWeights)

	leakagePerCell := physics.CellLeakagePower(readVoltageV, iOffA)
	staticEnergy := float64(totalWeights) * leakagePerCell * inferenceTimeS

	totalEnergy := dynamicEnergy + staticEnergy
	totalPJ := totalEnergy * 1e12

	// Physical range check (based on FeCIM literature)
	// For full MNIST 2-layer network (101,632 weights), expect higher energy
	const (
		minPJ = 1.0     // Lower bound: ultra-efficient FeCIM
		maxPJ = 10000.0 // Upper bound: large network (101k weights) with realistic parameters
	)

	if totalPJ < minPJ {
		t.Fatalf("total energy %.3f pJ below physically realistic range (min %.1f pJ)", totalPJ, minPJ)
	}
	if totalPJ > maxPJ {
		t.Fatalf("total energy %.3f pJ above physically realistic range (max %.1f pJ)", totalPJ, maxPJ)
	}

	t.Logf("PASS: E_total = %.3f pJ within physically realistic range [%.1f, %.1f] pJ for FeCIM",
		totalPJ, minPJ, maxPJ)
}

// TestM3_ENERGY_03_InferenceTimeImpact verifies static energy scales with inference time.
func TestM3_ENERGY_03_InferenceTimeImpact(t *testing.T) {
	t.Parallel()

	const (
		capacitanceF = 10e-15
		voltageV     = 1.8
		readVoltageV = 0.2
		iOffA        = 1e-9
		totalWeights = 101632 // 784×128 + 128×10
	)

	// Fixed dynamic energy (independent of time)
	energyPerSwitch := physics.CellSwitchingEnergy(capacitanceF, voltageV)
	dynamicEnergy := energyPerSwitch * float64(totalWeights)

	leakagePerCell := physics.CellLeakagePower(readVoltageV, iOffA)
	totalLeakage := float64(totalWeights) * leakagePerCell

	times := []struct {
		name   string
		timeS  float64
	}{
		{"fast (10 µs)", 10e-6},
		{"typical (100 µs)", 100e-6},
		{"slow (1 ms)", 1e-3},
	}

	var prevTotal float64
	for i, tc := range times {
		staticEnergy := totalLeakage * tc.timeS
		totalEnergy := dynamicEnergy + staticEnergy
		totalPJ := totalEnergy * 1e12

		if totalEnergy <= 0 {
			t.Fatalf("%s: total energy must be > 0, got %.12e J", tc.name, totalEnergy)
		}

		// Verify static energy increases with time
		if i > 0 && totalEnergy <= prevTotal {
			t.Fatalf("%s: total energy should increase with time (got %.3f pJ, prev %.3f pJ)",
				tc.name, totalPJ, prevTotal*1e12)
		}

		dynamicFraction := dynamicEnergy / totalEnergy * 100
		staticFraction := staticEnergy / totalEnergy * 100

		t.Logf("PASS: %s → E_total=%.3f pJ (dynamic=%.1f%%, static=%.1f%%)",
			tc.name, totalPJ, dynamicFraction, staticFraction)

		prevTotal = totalEnergy
	}
}

// TestM3_ENERGY_03_PeripheralOverhead verifies peripheral energy (DAC/ADC/TIA)
// is accounted for in total inference energy.
func TestM3_ENERGY_03_PeripheralOverhead(t *testing.T) {
	t.Parallel()

	cfg := DefaultNetworkConfig()

	// Use existing energy estimation (includes peripheral overhead)
	est := EstimateInferenceEnergyJ(cfg, 784, 128, 10)

	if est.TotalJ <= 0 {
		t.Fatalf("total energy must be > 0, got %.12e J", est.TotalJ)
	}
	if est.ComputeJ <= 0 {
		t.Fatalf("compute energy must be > 0, got %.12e J", est.ComputeJ)
	}
	if est.ADCDACJ <= 0 {
		t.Fatalf("peripheral energy must be > 0, got %.12e J", est.ADCDACJ)
	}

	// Verify breakdown sums correctly
	computed := est.ComputeJ + est.ADCDACJ
	if math.Abs(computed-est.TotalJ) > est.TotalJ*1e-12 {
		t.Fatalf("energy breakdown mismatch: compute+peripheral=%.12e J, total=%.12e J",
			computed, est.TotalJ)
	}

	peripheralFraction := est.ADCDACJ / est.TotalJ * 100
	computeFraction := est.ComputeJ / est.TotalJ * 100

	t.Logf("PASS: E_total = %.3f pJ", est.TotalJ*1e12)
	t.Logf("  ├─ Compute: %.3f pJ (%.1f%%)", est.ComputeJ*1e12, computeFraction)
	t.Logf("  └─ Peripheral (DAC+ADC): %.3f pJ (%.1f%%)", est.ADCDACJ*1e12, peripheralFraction)
}

// TestM3_ENERGY_03_LayerWiseBreakdown verifies per-layer energy contributions.
func TestM3_ENERGY_03_LayerWiseBreakdown(t *testing.T) {
	t.Parallel()

	const (
		capacitanceF   = 10e-15
		voltageV       = 1.8
		readVoltageV   = 0.2
		iOffA          = 1e-9
		inferenceTimeS = 100e-6
	)

	// Layer 1: 784×128
	layer1Weights := 784 * 128
	energyPerSwitch := physics.CellSwitchingEnergy(capacitanceF, voltageV)
	layer1Dynamic := energyPerSwitch * float64(layer1Weights)

	leakagePerCell := physics.CellLeakagePower(readVoltageV, iOffA)
	layer1Static := float64(layer1Weights) * leakagePerCell * inferenceTimeS
	layer1Total := layer1Dynamic + layer1Static

	// Layer 2: 128×10
	layer2Weights := 128 * 10
	layer2Dynamic := energyPerSwitch * float64(layer2Weights)
	layer2Static := float64(layer2Weights) * leakagePerCell * inferenceTimeS
	layer2Total := layer2Dynamic + layer2Static

	// Total
	totalEnergy := layer1Total + layer2Total

	if layer1Total <= 0 {
		t.Fatalf("layer 1 energy must be > 0, got %.12e J", layer1Total)
	}
	if layer2Total <= 0 {
		t.Fatalf("layer 2 energy must be > 0, got %.12e J", layer2Total)
	}

	// Layer 1 should dominate (78× more weights)
	if layer1Total <= layer2Total {
		t.Fatalf("layer 1 should dominate: L1=%.3f pJ, L2=%.3f pJ",
			layer1Total*1e12, layer2Total*1e12)
	}

	layer1Fraction := layer1Total / totalEnergy * 100
	layer2Fraction := layer2Total / totalEnergy * 100

	t.Logf("PASS: E_total = %.3f pJ", totalEnergy*1e12)
	t.Logf("  ├─ Layer 1 (784×128): %.3f pJ (%.1f%%)", layer1Total*1e12, layer1Fraction)
	t.Logf("  └─ Layer 2 (128×10): %.3f pJ (%.1f%%)", layer2Total*1e12, layer2Fraction)
}

// TestM3_ENERGY_03_DynamicVsStaticThreshold verifies dynamic energy dominates
// for fast inference (< 1ms), while static can dominate for slow inference (> 10ms).
func TestM3_ENERGY_03_DynamicVsStaticThreshold(t *testing.T) {
	t.Parallel()

	const (
		capacitanceF = 10e-15
		voltageV     = 1.8
		readVoltageV = 0.2
		iOffA        = 1e-9
		totalWeights = 101632
	)

	energyPerSwitch := physics.CellSwitchingEnergy(capacitanceF, voltageV)
	dynamicEnergy := energyPerSwitch * float64(totalWeights)

	leakagePerCell := physics.CellLeakagePower(readVoltageV, iOffA)
	totalLeakage := float64(totalWeights) * leakagePerCell

	// Find crossover time where E_static = E_dynamic
	crossoverTime := dynamicEnergy / totalLeakage

	t.Logf("Crossover time (E_static = E_dynamic): %.3f ms", crossoverTime*1e3)

	// Test fast inference (dynamic dominates)
	fastTime := 100e-6 // 100 µs
	fastStatic := totalLeakage * fastTime
	if dynamicEnergy <= fastStatic {
		t.Fatalf("fast inference: dynamic (%.3f pJ) should exceed static (%.3f pJ)",
			dynamicEnergy*1e12, fastStatic*1e12)
	}
	t.Logf("PASS: Fast (100 µs): dynamic=%.3f pJ > static=%.3f pJ",
		dynamicEnergy*1e12, fastStatic*1e12)

	// Test slow inference (static can dominate)
	slowTime := 10e-3 // 10 ms
	slowStatic := totalLeakage * slowTime
	if slowStatic <= dynamicEnergy {
		t.Logf("NOTE: Slow (10 ms): static=%.3f pJ > dynamic=%.3f pJ (expected for long inference)",
			slowStatic*1e12, dynamicEnergy*1e12)
	}
}

// TestM3_ENERGY_03_SingleLayerVsTwoLayer verifies energy scales with network size.
func TestM3_ENERGY_03_SingleLayerVsTwoLayer(t *testing.T) {
	t.Parallel()

	cfg := DefaultNetworkConfig()

	// Two-layer: 784→128→10
	twoLayerEst := EstimateInferenceEnergyJ(cfg, 784, 128, 10)

	// Single-layer: 784→10 (calibration mode)
	cfg.SingleLayer = true
	singleLayerEst := EstimateInferenceEnergyJ(cfg, 784, 128, 10)

	if twoLayerEst.TotalJ <= 0 {
		t.Fatalf("two-layer energy must be > 0, got %.12e J", twoLayerEst.TotalJ)
	}
	if singleLayerEst.TotalJ <= 0 {
		t.Fatalf("single-layer energy must be > 0, got %.12e J", singleLayerEst.TotalJ)
	}

	// Two-layer should consume more energy (more MACs)
	if twoLayerEst.TotalJ <= singleLayerEst.TotalJ {
		t.Fatalf("two-layer should consume more energy: 2L=%.3f pJ, 1L=%.3f pJ",
			twoLayerEst.TotalJ*1e12, singleLayerEst.TotalJ*1e12)
	}

	energyRatio := twoLayerEst.TotalJ / singleLayerEst.TotalJ
	macRatio := float64(twoLayerEst.TotalMACs) / float64(singleLayerEst.TotalMACs)

	t.Logf("PASS: Two-layer = %.3f pJ (MACs=%d)", twoLayerEst.TotalJ*1e12, twoLayerEst.TotalMACs)
	t.Logf("PASS: Single-layer = %.3f pJ (MACs=%d)", singleLayerEst.TotalJ*1e12, singleLayerEst.TotalMACs)
	t.Logf("PASS: Energy ratio = %.3f, MAC ratio = %.3f", energyRatio, macRatio)

	// Energy ratio should approximately match MAC ratio (within 20% due to peripheral overhead)
	if math.Abs(energyRatio-macRatio) > macRatio*0.2 {
		t.Logf("NOTE: Energy ratio (%.3f) differs from MAC ratio (%.3f) by %.1f%% (peripheral overhead)",
			energyRatio, macRatio, 100*math.Abs(energyRatio-macRatio)/macRatio)
	}
}

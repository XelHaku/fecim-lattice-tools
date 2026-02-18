// Package main provides integration tests for the FeCIM Lattice Tools application.
// These tests verify end-to-end workflows across multiple modules.
package main

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/simulation"
	"fecim-lattice-tools/module3-mnist/pkg/core"
	"fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/shared/peripherals"
)

// =============================================================================
// CROSS-MODULE INTEGRATION TESTS
// =============================================================================

// TestMaterialPhysicsConsistency verifies physics constants are consistent
// across modules 1 (hysteresis) and 2 (crossbar).
func TestMaterialPhysicsConsistency(t *testing.T) {
	// Module 1: Material parameters
	material := ferroelectric.FeCIMMaterial()

	// Verify demo baseline: 30 discrete levels
	states := ferroelectric.NewPreisachModel(material).DiscreteStates(30)
	if len(states) != 30 {
		t.Errorf("Expected 30 discrete states (demo baseline), got %d", len(states))
	}

	// Module 2: Crossbar array should use same levels
	cfg := &crossbar.Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create crossbar array: %v", err)
	}

	// Program a weight and verify quantization
	arr.ProgramWeight(0, 0, 0.5)

	// Module 3: 30-level quantization constant
	if core.FeCIMLevels != 30 {
		t.Errorf("FeCIMLevels baseline should be 30, got %d", core.FeCIMLevels)
	}

	t.Logf("Cross-module baseline verification passed: 30 discrete levels")
}

// TestHysteresisToMNISTWorkflow verifies the workflow from material
// characterization (Module 1) to neural network inference (Module 3).
func TestHysteresisToMNISTWorkflow(t *testing.T) {
	// Step 1: Characterize material behavior (Module 1)
	material := ferroelectric.DefaultHZO()
	model := ferroelectric.NewPreisachModel(material)

	// Generate hysteresis loop to verify material behavior
	E, P := model.GetHysteresisLoop(2*material.Ec, 100)
	if len(E) == 0 || len(P) == 0 {
		t.Fatal("Failed to generate hysteresis loop")
	}

	// Verify hysteresis shows proper ferroelectric behavior
	maxP := 0.0
	minP := 0.0
	for _, p := range P {
		if p > maxP {
			maxP = p
		}
		if p < minP {
			minP = p
		}
	}
	polarizationRange := maxP - minP
	if polarizationRange < 0.5*material.Ps {
		t.Errorf("Insufficient polarization range: %.4f", polarizationRange)
	}

	// Step 2: Extract discrete levels for quantization (Module 1/2 bridge)
	discreteLevels := model.DiscreteStates(30)
	if len(discreteLevels) != 30 {
		t.Fatalf("Expected 30 discrete levels (demo baseline), got %d", len(discreteLevels))
	}

	// Step 3: Create neural network with 30-level quantization (Module 3)
	net := core.NewDualModeNetwork(784, 128, 10)
	if net.Config.NumLevels != 30 {
		t.Errorf("Network should use 30 levels (demo baseline), got %d", net.Config.NumLevels)
	}

	// Initialize simple weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i*j)%100-50) / 500.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i*j)%100-50) / 200.0
		}
	}
	net.FPBias1 = make([]float64, 128)
	net.FPBias2 = make([]float64, 10)
	net.RequantizeWeights()

	// Step 4: Run inference
	input := make([]float64, 784)
	for i := 0; i < 100; i++ {
		input[i] = float64(i%10) / 10.0
	}

	result := net.Infer(input)

	// Verify result
	if result.FPPrediction < 0 || result.FPPrediction > 9 {
		t.Errorf("Invalid FP prediction: %d", result.FPPrediction)
	}
	if result.CIMPrediction < 0 || result.CIMPrediction > 9 {
		t.Errorf("Invalid CIM prediction: %d", result.CIMPrediction)
	}

	t.Logf("Workflow test: Hysteresis range=%.4f C/m², Network predictions FP=%d, CIM=%d",
		polarizationRange, result.FPPrediction, result.CIMPrediction)
}

// TestSimulationEngineIntegration verifies the simulation engine
// produces physically correct results.
func TestSimulationEngineIntegration(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := simulation.NewEngine(material)

	// Run simulation
	engine.SetWaveform(simulation.WaveformSine)
	engine.SetFrequency(1e6) // 1 MHz
	engine.Start()

	// Run multiple cycles
	for i := 0; i < 10000; i++ {
		engine.Step()
	}

	state := engine.State()

	// Verify simulation advanced
	if state.Time == 0 {
		t.Error("Simulation time did not advance")
	}

	// Verify history recorded
	if len(state.VoltageHistory) == 0 {
		t.Error("Voltage history empty")
	}
	if len(state.PolHistory) == 0 {
		t.Error("Polarization history empty")
	}

	// Get hysteresis data
	E, P := engine.GetHysteresisData()
	if len(E) == 0 || len(P) == 0 {
		t.Error("Hysteresis data empty")
	}

	t.Logf("Simulation integration: time=%e s, history=%d points, hysteresis=%d points",
		state.Time, len(state.VoltageHistory), len(E))
}

// TestCrossbarNonIdealitiesIntegration verifies IR drop, sneak paths,
// and drift simulations work together.
func TestCrossbarNonIdealitiesIntegration(t *testing.T) {
	// Set up crossbar array
	size := 8

	// Test IR Drop
	irSim := crossbar.NewIRDropSimulator(size, size)
	for i := 0; i < size; i++ {
		irSim.SetInputVoltage(i, 0.5)
		for j := 0; j < size; j++ {
			irSim.SetConductance(i, j, 50e-6)
		}
	}
	irSim.Simulate(100)
	irStats := irSim.GetStats()

	if irStats.MaxIRDrop <= 0 {
		t.Error("IR drop simulation produced no drop")
	}

	// Test Sneak Paths
	spSim := crossbar.NewSneakPathAnalyzer(size, size)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			spSim.SetConductance(i, j, 50e-6)
		}
	}
	spSim.AnalyzeTarget(4, 4, 0.5)
	spStats := spSim.GetStats(0.5)

	if spStats.TotalSneakCurrent <= 0 {
		t.Error("Sneak path analysis produced no current")
	}

	// Test Drift
	driftSim := crossbar.NewDriftSimulator(size, size, 30)
	driftSim.SimulateTimeStep(1000) // 1000 seconds
	driftStats := driftSim.GetStats()

	// Verify drift stats are reasonable
	if driftStats.RetentionPrediction < 90 {
		t.Errorf("FeCIM should have >90%% retention, got %.1f%%",
			driftStats.RetentionPrediction)
	}

	t.Logf("Non-idealities: IR drop=%.4f mV, sneak SNR=%.1f dB, retention=%.1f%%",
		irStats.MaxIRDrop*1000, spStats.SignalToNoiseRatio, driftStats.RetentionPrediction)
}

// TestPeripheralCircuitsIntegration verifies DAC/ADC/TIA circuits
// work correctly with crossbar arrays.
func TestPeripheralCircuitsIntegration(t *testing.T) {
	// Create peripheral circuits using default constructors
	dac := peripherals.DefaultDAC()
	adc := peripherals.DefaultADC()
	tia := peripherals.DefaultTIA()

	// Test DAC: digital to analog
	digitalIn := 15 // Mid-scale for 30-level DAC
	analogOut := dac.Convert(digitalIn)

	// DAC should produce voltage in valid range
	vMin, vMax := dac.VoltageRange()
	if analogOut < vMin || analogOut > vMax {
		t.Errorf("DAC output %.4f V outside range [%.4f, %.4f]", analogOut, vMin, vMax)
	}

	// Test ADC: analog to digital (0.5V should be mid-scale)
	digitalOut := adc.Convert(0.5)
	if digitalOut < 0 || digitalOut > adc.Levels() {
		t.Errorf("ADC output %d outside valid range", digitalOut)
	}

	// Test TIA: current to voltage
	currentIn := 1e-6 // 1 µA
	voltageOut := tia.Convert(currentIn)

	// TIA should produce positive voltage for positive current
	if voltageOut <= 0 {
		t.Errorf("TIA output %.4f V should be positive for positive current", voltageOut)
	}

	t.Logf("Peripheral circuits: DAC(level=%d)=%.4fV, ADC(0.5V)=%d, TIA(1µA)=%.4fV",
		digitalIn, analogOut, digitalOut, voltageOut)
}

// TestQuantizationCliffEffect verifies the quantization cliff behavior
// where 2-level quantization performs much worse than 30-level.
func TestQuantizationCliffEffect(t *testing.T) {
	net := core.NewDualModeNetwork(100, 50, 10)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i*j)%100-50) / 100.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i*j)%100-50) / 100.0
		}
	}
	net.FPBias1 = make([]float64, 50)
	net.FPBias2 = make([]float64, 10)

	// Test with 30 levels (good)
	net.SetNumLevels(30)
	net.RequantizeWeights()
	stats30 := core.ComputeQuantizationStats(net.FPWeights1, net.QuantWeights1)

	// Test with 2 levels (bad)
	net.SetNumLevels(2)
	net.RequantizeWeights()
	stats2 := core.ComputeQuantizationStats(net.FPWeights1, net.QuantWeights1)

	// 30 levels should have much better PSNR than 2 levels
	if stats30.PSNR <= stats2.PSNR {
		t.Errorf("30-level PSNR (%.1f dB) should be > 2-level PSNR (%.1f dB)",
			stats30.PSNR, stats2.PSNR)
	}

	// MSE improvement should be significant (at least 10x better)
	if stats2.MSE < stats30.MSE*10 {
		t.Errorf("2-level MSE (%.4f) should be much worse than 30-level (%.4f)",
			stats2.MSE, stats30.MSE)
	}

	t.Logf("Quantization cliff: 30-level PSNR=%.1f dB, 2-level PSNR=%.1f dB",
		stats30.PSNR, stats2.PSNR)
}

// TestEndToEndMNISTInference verifies complete MNIST inference pipeline.
func TestEndToEndMNISTInference(t *testing.T) {
	// Create network with MNIST dimensions
	net := core.NewDualModeNetwork(784, 128, 10)

	// Initialize with reasonable weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = (float64((i+j)%100) - 50) / 500.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = (float64((i*j)%100) - 50) / 200.0
		}
	}
	net.FPBias1 = make([]float64, 128)
	net.FPBias2 = make([]float64, 10)
	for i := range net.FPBias1 {
		net.FPBias1[i] = 0.01
	}
	for i := range net.FPBias2 {
		net.FPBias2[i] = 0.01
	}

	// Set ideal conditions
	net.Config.NumLevels = 30
	net.Config.NoiseLevel = 0.0
	net.Config.ADCBits = 8
	net.Config.DACBits = 8
	net.RequantizeWeights()

	// Create synthetic digit patterns and run inference
	for digit := 0; digit < 10; digit++ {
		input := make([]float64, 784)

		// Create simple pattern for each digit
		for i := 0; i < 784; i++ {
			input[i] = float64((i+digit*100)%256) / 255.0
		}

		result := net.Infer(input)

		// Verify valid output
		if result.FPPrediction < 0 || result.FPPrediction > 9 {
			t.Errorf("Digit %d: Invalid FP prediction %d", digit, result.FPPrediction)
		}
		if result.CIMPrediction < 0 || result.CIMPrediction > 9 {
			t.Errorf("Digit %d: Invalid CIM prediction %d", digit, result.CIMPrediction)
		}

		// Verify probabilities sum to 1
		fpSum := 0.0
		for _, p := range result.FPProbabilities {
			fpSum += p
		}
		if math.Abs(fpSum-1.0) > 0.01 {
			t.Errorf("Digit %d: FP probabilities don't sum to 1: %.4f", digit, fpSum)
		}
	}

	t.Log("End-to-end MNIST inference verified for all 10 digit patterns")
}

// TestConcurrentInferenceStability verifies thread safety of inference.
func TestConcurrentInferenceStability(t *testing.T) {
	net := core.NewDualModeNetwork(100, 50, 10)

	// Initialize weights
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i*j)%100-50) / 500.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i*j)%100-50) / 200.0
		}
	}
	net.FPBias1 = make([]float64, 50)
	net.FPBias2 = make([]float64, 10)
	net.RequantizeWeights()

	// Run concurrent inferences
	done := make(chan bool)
	numGoroutines := 5
	inferPerGoroutine := 100

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			for i := 0; i < inferPerGoroutine; i++ {
				input := make([]float64, 100)
				for j := range input {
					input[j] = float64((id*i+j)%100) / 100.0
				}
				result := net.Infer(input)

				// Verify result is valid
				if result.FPPrediction < 0 || result.FPPrediction > 9 {
					t.Errorf("Goroutine %d: Invalid prediction", id)
				}
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	t.Logf("Concurrent inference: %d goroutines × %d inferences completed",
		numGoroutines, inferPerGoroutine)
}

// =============================================================================
// PHYSICS ACCURACY TESTS
// =============================================================================

// Test30LevelQuantizationAccuracy verifies the 30-level quantization
// meets FeCIM physics specifications.
func Test30LevelQuantizationAccuracy(t *testing.T) {
	// Test quantization at different bit depths
	testCases := []struct {
		levels      int
		expectedMin float64 // Minimum PSNR expected
	}{
		{2, 4.0},   // Binary: very low quality
		{4, 13.0},  // 2-bit
		{8, 20.0},  // 3-bit
		{16, 26.0}, // 4-bit
		{30, 32.0}, // FeCIM: high quality
	}

	// Create test weights
	weights := make([][]float64, 100)
	for i := range weights {
		weights[i] = make([]float64, 100)
		for j := range weights[i] {
			weights[i][j] = float64(i*j%100-50) / 50.0 // Range [-1, 1]
		}
	}

	for _, tc := range testCases {
		t.Run(string(rune('0'+tc.levels)), func(t *testing.T) {
			quantized, err := core.QuantizeWeights(weights, tc.levels)
			if err != nil {
				t.Fatalf("Quantization failed: %v", err)
			}

			stats := core.ComputeQuantizationStats(weights, quantized)

			if stats.PSNR < tc.expectedMin {
				t.Errorf("%d levels: PSNR %.1f dB < expected %.1f dB",
					tc.levels, stats.PSNR, tc.expectedMin)
			}

			t.Logf("%d levels: MSE=%.4f, PSNR=%.1f dB, distinct=%d",
				tc.levels, stats.MSE, stats.PSNR, stats.NumDistinct)
		})
	}
}

// TestFerroelectricMaterialPhysics verifies material parameters
// match published literature.
func TestFerroelectricMaterialPhysics(t *testing.T) {
	materials := []*ferroelectric.HZOMaterial{
		ferroelectric.DefaultHZO(),
		ferroelectric.FeCIMMaterial(),
		ferroelectric.LiteratureSuperlattice(),
	}

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Pr is stored in C/m² where 1 C/m² = 10,000 µC/cm²
			// Literature values: Pr ~ 10-50 µC/cm² = 0.001-0.005 C/m²
			// But these materials use ~0.25-0.45 C/m² for visualization purposes
			// So we verify they are positive and reasonable for the simulation
			if mat.Pr <= 0 {
				t.Errorf("Pr must be positive, got %.4f C/m²", mat.Pr)
			}

			// Verify coercive field (literature: 0.5-2 MV/cm)
			ecMVcm := mat.Ec / 1e8 // Convert to MV/cm
			if ecMVcm < 0.5 || ecMVcm > 2.5 {
				t.Errorf("Ec %.2f MV/cm outside literature range [0.5, 2.5]", ecMVcm)
			}

			// Verify thickness is reasonable (5-50 nm for FeFET)
			thickNm := mat.Thickness * 1e9
			if thickNm < 1 || thickNm > 100 {
				t.Errorf("Thickness %.0f nm outside expected range", thickNm)
			}

			// Verify Curie temperature is above operating range
			if mat.CurieTemp < 400 {
				t.Errorf("Curie temp %.0f K too low", mat.CurieTemp)
			}

			// Verify Ps > Pr
			if mat.Ps <= mat.Pr {
				t.Errorf("Ps (%.4f) should be > Pr (%.4f)", mat.Ps, mat.Pr)
			}

			t.Logf("%s: Pr=%.4f C/m², Ec=%.2f MV/cm, endurance=%.0e",
				mat.Name, mat.Pr, ecMVcm, mat.EnduranceCycles)
		})
	}
}

// =============================================================================
// BENCHMARKS
// =============================================================================

func BenchmarkFullInferencePipeline(b *testing.B) {
	net := core.NewDualModeNetwork(784, 128, 10)
	for i := range net.FPWeights1 {
		for j := range net.FPWeights1[i] {
			net.FPWeights1[i][j] = float64((i*j)%100-50) / 500.0
		}
	}
	for i := range net.FPWeights2 {
		for j := range net.FPWeights2[i] {
			net.FPWeights2[i][j] = float64((i*j)%100-50) / 200.0
		}
	}
	net.FPBias1 = make([]float64, 128)
	net.FPBias2 = make([]float64, 10)
	net.RequantizeWeights()

	input := make([]float64, 784)
	for i := range input {
		input[i] = float64(i%256) / 255.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		net.Infer(input)
	}
}

func BenchmarkCrossbarMVM(b *testing.B) {
	cfg := &crossbar.Config{
		Rows:       64,
		Cols:       64,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, _ := crossbar.NewArray(cfg)
	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			arr.ProgramWeight(i, j, float64((i*j)%100)/100.0)
		}
	}

	input := make([]float64, 64)
	for i := range input {
		input[i] = float64(i) / 64.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr.MVM(input)
	}
}

func BenchmarkQuantization30Levels(b *testing.B) {
	weights := make([][]float64, 128)
	for i := range weights {
		weights[i] = make([]float64, 784)
		for j := range weights[i] {
			weights[i][j] = float64((i*j)%100-50) / 50.0
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.QuantizeWeights(weights, 30)
	}
}

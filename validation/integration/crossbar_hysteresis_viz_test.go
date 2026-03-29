// Package integration provides comprehensive integration tests that exercise
// multiple FeCIM modules together. These tests verify crossbar + hysteresis +
// visualization workflows and end-to-end functionality.
package integration

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module1-hysteresis/pkg/simulation"
	"fecim-lattice-tools/module2-crossbar/pkg/visualization"
	"fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/shared/physics"
)

// =============================================================================
// HYSTERESIS → CROSSBAR INTEGRATION TESTS
// =============================================================================

// TestHysteresisModelToCrossbarProgramming verifies that conductance values
// derived from hysteresis simulation can be correctly programmed into crossbar arrays.
func TestHysteresisModelToCrossbarProgramming(t *testing.T) {
	// Step 1: Create material and Preisach model
	material := ferroelectric.DefaultHZO()
	model := ferroelectric.NewPreisachModel(material)

	// Step 2: Generate discrete states from hysteresis model (DefaultLevels baseline)
	discreteStates := model.DiscreteStates(physics.DefaultLevels)
	if len(discreteStates) != physics.DefaultLevels {
		t.Fatalf("Expected %d discrete states (demo baseline), got %d", physics.DefaultLevels, len(discreteStates))
	}

	// Step 3: Create crossbar array
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
	defer arr.Destroy()

	// Step 4: Program weights derived from polarization states
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			// Map cell (i,j) to a discrete state level
			level := (i*8 + j) % physics.DefaultLevels
			state := discreteStates[level]

			// Convert normalized polarization to conductance [0,1]
			// NormalizedP is in [-1, 1], map to [0, 1]
			conductance := (state.NormalizedP + 1.0) / 2.0

			err := arr.ProgramWeight(i, j, conductance)
			if err != nil {
				t.Errorf("Failed to program weight at (%d,%d): %v", i, j, err)
			}
		}
	}

	// Step 5: Verify conductances are correctly quantized
	matrix := arr.GetConductanceMatrix()
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			g := matrix[i][j]
			// Verify quantization to DefaultLevels (step = 1/(DefaultLevels-1))
			level := crossbar.GetLevel(g)
			if level < 0 || level >= physics.DefaultLevels {
				t.Errorf("Invalid level %d at (%d,%d)", level, i, j)
			}
		}
	}

	t.Log("Hysteresis → Crossbar programming integration verified")
}

// TestHysteresisDrivenMVM verifies that MVM operations work correctly
// with hysteresis-derived conductance values.
func TestHysteresisDrivenMVM(t *testing.T) {
	// Create material and extract conductance mapping
	material := ferroelectric.FeCIMMaterial()
	model := ferroelectric.NewPreisachModel(material)

	// Create crossbar with FeCIM-calibrated conductances
	cfg := &crossbar.Config{
		Rows:             16,
		Cols:             16,
		NoiseLevel:       0.0,
		ADCBits:          8,
		DACBits:          8,
		ConductanceModel: crossbar.ConductanceLinear,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	defer arr.Destroy()

	// Generate hysteresis loop and extract representative polarizations
	E, P := model.GetHysteresisLoop(2*material.Ec, 100)
	if len(E) == 0 || len(P) == 0 {
		t.Fatal("Failed to generate hysteresis loop")
	}

	// Program weights from hysteresis loop ascending branch
	ascendingStart := 0
	ascendingEnd := len(P) / 2
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			// Sample from ascending branch
			idx := ascendingStart + (ascendingEnd-ascendingStart)*(i*16+j)/(16*16)
			pol := P[idx]
			// Normalize and convert to conductance
			normPol := pol / material.Ps
			conductance := (normPol + 1.0) / 2.0
			arr.ProgramWeight(i, j, conductance)
		}
	}

	// Perform MVM with unit input
	input := make([]float64, 16)
	for i := range input {
		input[i] = 1.0
	}

	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	// Verify output dimensions and valid values
	if len(output) != 16 {
		t.Errorf("Expected 16 outputs, got %d", len(output))
	}

	for i, val := range output {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			t.Errorf("Invalid output[%d] = %v", i, val)
		}
		if val < 0 || val > 1 {
			t.Errorf("Output[%d] = %f out of normalized range", i, val)
		}
	}

	t.Logf("Hysteresis-driven MVM: input sum=%.2f, output range=[%.4f, %.4f]",
		16.0, output[0], output[len(output)-1])
}

// TestSimulationEngineToCrossbar verifies the workflow from
// time-domain simulation to crossbar programming.
func TestSimulationEngineToCrossbar(t *testing.T) {
	// Step 1: Run hysteresis simulation
	material := ferroelectric.DefaultHZO()
	engine := simulation.NewEngine(material)
	engine.SetWaveform(simulation.WaveformSine)
	engine.SetFrequency(1e6) // 1 MHz
	engine.Start()

	// Run simulation to collect polarization data
	numSteps := 5000
	polarizations := make([]float64, 0, numSteps)

	for i := 0; i < numSteps; i++ {
		engine.Step()
		state := engine.State()
		polarizations = append(polarizations, state.NormPol)
	}

	engine.Stop()

	// Verify simulation produced data
	if len(polarizations) == 0 {
		t.Fatal("Simulation produced no polarization data")
	}

	// Find min/max polarization to verify hysteresis behavior
	minP, maxP := polarizations[0], polarizations[0]
	for _, p := range polarizations {
		if p < minP {
			minP = p
		}
		if p > maxP {
			maxP = p
		}
	}

	polRange := maxP - minP
	if polRange < 0.5 {
		t.Errorf("Insufficient polarization range: %.4f (expected > 0.5)", polRange)
	}

	// Step 2: Create crossbar and program with simulation-derived conductances
	cfg := &crossbar.Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.01,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	defer arr.Destroy()

	// Sample polarizations for weight programming
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			idx := (i*8 + j) * numSteps / 64
			normPol := polarizations[idx]
			conductance := (normPol + 1.0) / 2.0
			arr.ProgramWeight(i, j, conductance)
		}
	}

	// Verify MVM still works
	input := make([]float64, 8)
	for i := range input {
		input[i] = 0.5
	}

	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM after simulation programming failed: %v", err)
	}

	if len(output) != 8 {
		t.Errorf("Expected 8 outputs, got %d", len(output))
	}

	t.Logf("Simulation→Crossbar: %d steps, pol range=[%.2f, %.2f], MVM output[0]=%.4f",
		numSteps, minP, maxP, output[0])
}

// =============================================================================
// CROSSBAR → VISUALIZATION INTEGRATION TESTS
// =============================================================================

// TestCrossbarStateVisualization verifies the terminal visualizer
// correctly displays crossbar array state.
func TestCrossbarStateVisualization(t *testing.T) {
	// Create and program crossbar
	cfg := &crossbar.Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	defer arr.Destroy()

	// Program with gradient pattern for easy visual verification
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			// Diagonal gradient
			conductance := float64(i+j) / 14.0
			arr.ProgramWeight(i, j, conductance)
		}
	}

	// Create visualizer
	viz := visualization.NewTerminalVisualizer(arr, false)

	// Capture output (we can't easily capture stdout, but we verify no panic)
	viz.ShowCrossbarState()

	// Verify the visualizer can be called with color mode too
	vizColor := visualization.NewTerminalVisualizer(arr, true)
	vizColor.ShowCrossbarState()

	t.Log("Crossbar visualization integration verified (no panic, output generated)")
}

// TestMVMOperationVisualization verifies MVM operation display.
func TestMVMOperationVisualization(t *testing.T) {
	cfg := &crossbar.Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	defer arr.Destroy()

	// Program identity-ish matrix
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			if i == j {
				arr.ProgramWeight(i, j, 0.8)
			} else {
				arr.ProgramWeight(i, j, 0.1)
			}
		}
	}

	// Run MVM
	input := make([]float64, 8)
	for i := range input {
		input[i] = float64(i+1) / 8.0
	}

	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	// Visualize
	viz := visualization.NewTerminalVisualizer(arr, false)
	viz.ShowMVMOperation(input, output)

	t.Log("MVM visualization integration verified")
}

// TestIRDropVisualization verifies IR drop analysis visualization.
func TestIRDropVisualization(t *testing.T) {
	size := 8

	// Create crossbar array
	cfg := &crossbar.Config{
		Rows:       size,
		Cols:       size,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	defer arr.Destroy()

	// Program with uniform conductances
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	// Create input voltages
	input := make([]float64, size)
	for i := range input {
		input[i] = 0.5
	}

	// Get IR drop analysis from array
	wireParams := crossbar.DefaultWireParams()
	analysis := arr.AnalyzeIRDrop(input, wireParams)
	if analysis == nil {
		t.Fatal("IR drop analysis returned nil")
	}

	// Visualize IR drop
	viz := visualization.NewTerminalVisualizer(arr, true)
	viz.ShowIRDropAnalysis(analysis)

	// Verify analysis contains valid data
	if analysis.MaxIRDrop < 0 {
		t.Error("IR drop should be non-negative")
	}
	if analysis.AvgIRDrop < 0 {
		t.Error("Average IR drop should be non-negative")
	}

	t.Logf("IR drop visualization: max=%.4f, avg=%.4f, variance=%.6f",
		analysis.MaxIRDrop, analysis.AvgIRDrop, analysis.IRDropVariance)
}

// TestSneakPathVisualization verifies sneak path analysis visualization.
func TestSneakPathVisualization(t *testing.T) {
	size := 8

	// Create crossbar array
	cfg := &crossbar.Config{
		Rows:       size,
		Cols:       size,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	defer arr.Destroy()

	// Program with uniform conductances
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			arr.ProgramWeight(i, j, 0.5)
		}
	}

	// Analyze for a specific target cell
	targetRow, targetCol := 4, 4
	analysis := arr.AnalyzeSneakPaths(targetRow, targetCol)
	if analysis == nil {
		t.Fatal("Sneak path analysis returned nil")
	}

	// Visualize sneak paths
	viz := visualization.NewTerminalVisualizer(arr, true)
	viz.ShowSneakPathAnalysis(analysis, targetRow, targetCol)

	// Verify analysis validity
	if analysis.TotalSignal < 0 {
		t.Error("Signal current should be non-negative")
	}
	if analysis.TotalSneak < 0 {
		t.Error("Sneak current should be non-negative")
	}

	t.Logf("Sneak path visualization: signal=%.6f, sneak=%.6f, max ratio=%.4f",
		analysis.TotalSignal, analysis.TotalSneak, analysis.MaxSneakRatio)
}

// =============================================================================
// END-TO-END WORKFLOW TESTS
// =============================================================================

// TestFullHysteresisCrossbarVisualizationPipeline tests the complete workflow
// from hysteresis characterization through crossbar operation to visualization.
func TestFullHysteresisCrossbarVisualizationPipeline(t *testing.T) {
	// ===== STAGE 1: Material Characterization =====
	material := ferroelectric.FeCIMMaterial()
	model := ferroelectric.NewPreisachModel(material)

	// Generate and verify hysteresis loop
	Emax := 2 * material.Ec
	_, P := model.GetHysteresisLoop(Emax, 100)

	// Verify hysteresis characteristics
	var maxP, minP float64 = P[0], P[0]
	for _, p := range P {
		if p > maxP {
			maxP = p
		}
		if p < minP {
			minP = p
		}
	}

	polarizationWindow := maxP - minP
	expectedWindow := 2 * material.Ps * 0.8 // Allow 20% variation
	if polarizationWindow < expectedWindow {
		t.Errorf("Polarization window %.4f below expected %.4f", polarizationWindow, expectedWindow)
	}

	t.Logf("Stage 1 - Material: Ps=%.4f, Ec=%.2e, P_window=%.4f",
		material.Ps, material.Ec, polarizationWindow)

	// ===== STAGE 2: Weight Quantization =====
	levels := physics.DefaultLevels // FeCIM baseline

	// Create weight matrix from hysteresis-derived values
	weights := make([][]float64, 16)
	for i := range weights {
		weights[i] = make([]float64, 16)
		for j := range weights[i] {
			// Sample from hysteresis loop
			idx := (i*16 + j) % len(P)
			normPol := P[idx] / material.Ps
			weights[i][j] = normPol
		}
	}

	// Quantize weights
	quantized := make([][]float64, 16)
	for i := range quantized {
		quantized[i] = make([]float64, 16)
		for j := range quantized[i] {
			// Map [-1,1] to [0,1] for conductance
			normalized := (weights[i][j] + 1.0) / 2.0
			// Quantize to levels
			quantized[i][j] = physics.QuantizeTo30Levels(normalized)
		}
	}

	// Calculate quantization error
	var mse float64
	for i := range weights {
		for j := range weights[i] {
			originalNorm := (weights[i][j] + 1.0) / 2.0
			diff := originalNorm - quantized[i][j]
			mse += diff * diff
		}
	}
	mse /= float64(16 * 16)
	psnr := -10 * math.Log10(mse+1e-10)

	const minPSNRdB = 30.0 // minimum acceptable PSNR for FeCIM quantization
	if psnr < minPSNRdB {
		t.Errorf("Quantization PSNR %.1f dB below expected (%.0f dB)", psnr, minPSNRdB)
	}

	t.Logf("Stage 2 - Quantization: %d levels, MSE=%.6f, PSNR=%.1f dB", levels, mse, psnr)

	// ===== STAGE 3: Crossbar Programming =====
	cfg := &crossbar.Config{
		Rows:             16,
		Cols:             16,
		NoiseLevel:       0.02,
		ADCBits:          8,
		DACBits:          8,
		ConductanceModel: crossbar.ConductanceLinear,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create crossbar: %v", err)
	}
	defer arr.Destroy()

	// Program quantized weights
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			arr.ProgramWeight(i, j, quantized[i][j])
		}
	}

	// Verify programming
	programmed := arr.GetConductanceMatrix()
	var programmingError float64
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			diff := programmed[i][j] - quantized[i][j]
			programmingError += diff * diff
		}
	}
	programmingError /= float64(16 * 16)

	if programmingError > 1e-6 {
		t.Errorf("Programming error %.6f higher than expected", programmingError)
	}

	t.Logf("Stage 3 - Programming: %dx%d array, error=%.2e", 16, 16, programmingError)

	// ===== STAGE 4: MVM Operation =====
	input := make([]float64, 16)
	for i := range input {
		input[i] = float64(i+1) / 16.0
	}

	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}

	// Verify output characteristics
	var outSum float64
	for _, v := range output {
		outSum += v
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Error("MVM produced invalid output")
		}
	}

	t.Logf("Stage 4 - MVM: input_sum=%.2f, output_sum=%.4f", 8.5, outSum)

	// ===== STAGE 5: Visualization =====
	viz := visualization.NewTerminalVisualizer(arr, false)

	// These should not panic
	viz.ShowCrossbarState()
	viz.ShowMVMOperation(input, output)

	t.Log("Stage 5 - Visualization: All displays rendered successfully")
	t.Log("Full pipeline test PASSED")
}

// TestCrossbarNonIdealitiesFullWorkflow tests the complete non-idealities
// workflow with visualization.
func TestCrossbarNonIdealitiesFullWorkflow(t *testing.T) {
	size := 16

	// ===== Create visualization array =====
	cfg := &crossbar.Config{Rows: size, Cols: size, ADCBits: 8, DACBits: 8, NoiseLevel: 0.0}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	defer arr.Destroy()

	// Program with gradient
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			arr.ProgramWeight(i, j, float64(i+j)/float64(2*size))
		}
	}

	// ===== MVM with non-idealities =====
	input := make([]float64, size)
	for i := range input {
		input[i] = 0.5
	}

	// ===== IR Drop Analysis =====
	wireParams := crossbar.DefaultWireParams()
	irAnalysis := arr.AnalyzeIRDrop(input, wireParams)

	// ===== Sneak Path Analysis =====
	spAnalysis := arr.AnalyzeSneakPaths(8, 8)

	// ===== Drift Simulation =====
	driftSim := crossbar.NewDriftSimulator(size, size, physics.DefaultLevels)
	driftSim.SimulateTimeStep(3600) // 1 hour
	driftStats := driftSim.GetStats()

	// Get ideal output
	idealOutput, _ := arr.MVM(input)

	// Simulate actual output with IR drop effect (simplified)
	actualOutput := make([]float64, size)
	for i := range idealOutput {
		// Apply IR drop degradation
		degradation := 1.0 - irAnalysis.MaxIRDrop*float64(i)/float64(size)
		actualOutput[i] = idealOutput[i] * degradation
	}

	// ===== Visualize =====
	viz := visualization.NewTerminalVisualizer(arr, true)
	viz.ShowCrossbarState()
	viz.ShowIRDropAnalysis(irAnalysis)
	viz.ShowSneakPathAnalysis(spAnalysis, 8, 8)
	viz.ShowMVMWithNonidealities(input, idealOutput, actualOutput, irAnalysis)

	// ===== Verify results =====
	// Note: IR drop might be very small or zero depending on implementation
	if irAnalysis.MaxIRDrop < 0 {
		t.Error("IR drop should be non-negative")
	}
	if spAnalysis.TotalSneak < 0 {
		t.Error("Sneak current should be non-negative")
	}
	if driftStats.RetentionPrediction < 90 {
		t.Errorf("FeCIM retention should be >90%%, got %.1f%%", driftStats.RetentionPrediction)
	}

	t.Logf("Non-idealities workflow: IR_max=%.4f, sneak_ratio=%.4f, retention=%.1f%%",
		irAnalysis.MaxIRDrop, spAnalysis.MaxSneakRatio, driftStats.RetentionPrediction)
}

// =============================================================================
// PHYSICS CONSISTENCY TESTS
// =============================================================================

// TestQuantizationLevelConsistencyAcrossModules verifies that the 30-level
// quantization is consistent between hysteresis, crossbar, and physics modules.
func TestQuantizationLevelConsistencyAcrossModules(t *testing.T) {
	// Module 1: Hysteresis discrete states
	material := ferroelectric.FeCIMMaterial()
	model := ferroelectric.NewPreisachModel(material)
	hysteresisStates := model.DiscreteStates(physics.DefaultLevels)

	// Module 2: Crossbar quantization constant
	crossbarLevels := crossbar.DefaultQuantizationLevels

	// Shared physics: Quantization function
	physicsLevels := physics.DefaultLevels

	// Verify consistency
	if len(hysteresisStates) != physics.DefaultLevels {
		t.Errorf("Hysteresis states: expected %d, got %d", physics.DefaultLevels, len(hysteresisStates))
	}
	if crossbarLevels != physics.DefaultLevels {
		t.Errorf("Crossbar levels: expected %d, got %d", physics.DefaultLevels, crossbarLevels)
	}
	if physicsLevels != physics.DefaultLevels {
		t.Errorf("Physics levels: expected %d, got %d", physics.DefaultLevels, physicsLevels)
	}

	// Verify quantization function produces exactly 30 distinct values
	distinctValues := make(map[float64]bool)
	for i := 0; i < 1000; i++ {
		input := float64(i) / 999.0
		quantized := physics.QuantizeTo30Levels(input)
		distinctValues[quantized] = true
	}

	if len(distinctValues) != physics.DefaultLevels {
		t.Errorf("Quantization produced %d distinct values, expected %d", len(distinctValues), physics.DefaultLevels)
	}

	// Verify crossbar.QuantizeToLevels matches physics.QuantizeTo30Levels
	for i := 0; i < 100; i++ {
		input := float64(i) / 99.0
		crossbarQ := crossbar.QuantizeToLevels(input)
		physicsQ := physics.QuantizeTo30Levels(input)

		if crossbarQ != physicsQ {
			t.Errorf("Quantization mismatch at %.4f: crossbar=%.4f, physics=%.4f",
				input, crossbarQ, physicsQ)
		}
	}

	t.Log("Quantization level consistency verified across all modules")
}

// TestConductanceRangeConsistency verifies physical conductance ranges
// are consistent between modules.
func TestConductanceRangeConsistency(t *testing.T) {
	// Crossbar module constants
	crossbarGMin := crossbar.GMin
	crossbarGMax := crossbar.GMax

	// Physics module constants
	physicsGMin := physics.GMin
	physicsGMax := physics.GMax

	// Verify consistency
	if crossbarGMin != physicsGMin {
		t.Errorf("GMin mismatch: crossbar=%e, physics=%e", crossbarGMin, physicsGMin)
	}
	if crossbarGMax != physicsGMax {
		t.Errorf("GMax mismatch: crossbar=%e, physics=%e", crossbarGMax, physicsGMax)
	}

	// Verify range is physically reasonable (matches shared physics constants)
	expectedGMin := physics.GMin  // 1 µS
	expectedGMax := physics.GMax  // 100 µS

	if math.Abs(crossbarGMin-expectedGMin) > 1e-9 {
		t.Errorf("GMin %.0f µS outside expected 1 µS", crossbarGMin*1e6)
	}
	if math.Abs(crossbarGMax-expectedGMax) > 1e-9 {
		t.Errorf("GMax %.0f µS outside expected 100 µS", crossbarGMax*1e6)
	}

	t.Logf("Conductance range verified: [%.0f, %.0f] µS", crossbarGMin*1e6, crossbarGMax*1e6)
}

// TestMaterialParameterConsistency verifies material parameters are
// consistent across the hysteresis and crossbar modules.
func TestMaterialParameterConsistency(t *testing.T) {
	materials := []*ferroelectric.HZOMaterial{
		ferroelectric.DefaultHZO(),
		ferroelectric.FeCIMMaterial(),
		ferroelectric.LiteratureSuperlattice(),
	}

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			// Create Preisach model
			model := ferroelectric.NewPreisachModel(mat)

			// Verify Ps > Pr (saturation > remanent)
			if mat.Ps <= mat.Pr {
				t.Errorf("Invalid: Ps (%.4f) should be > Pr (%.4f)", mat.Ps, mat.Pr)
			}

			// Verify Ec is positive
			if mat.Ec <= 0 {
				t.Errorf("Coercive field should be positive: %.2e", mat.Ec)
			}

			// Generate hysteresis loop and verify shape
			_, P := model.GetHysteresisLoop(2*mat.Ec, 100)

			// Find coercive field points (where P crosses zero)
			coerciveCount := 0
			for i := 1; i < len(P); i++ {
				if (P[i-1] <= 0 && P[i] > 0) || (P[i-1] >= 0 && P[i] < 0) {
					coerciveCount++
				}
			}

			// Should have at least 2 zero-crossings (up and down)
			if coerciveCount < 2 {
				t.Errorf("Expected at least 2 coercive crossings, got %d", coerciveCount)
			}

			t.Logf("%s: Ps=%.4f, Pr=%.4f, Ec=%.2e, crossings=%d",
				mat.Name, mat.Ps, mat.Pr, mat.Ec, coerciveCount)
		})
	}
}

// =============================================================================
// STRESS AND EDGE CASE TESTS
// =============================================================================

// TestLargeCrossbarWithHysteresisData tests a larger crossbar array
// with hysteresis-derived data.
func TestLargeCrossbarWithHysteresisData(t *testing.T) {
	// Create larger array
	size := 64
	cfg := &crossbar.Config{
		Rows:       size,
		Cols:       size,
		NoiseLevel: 0.02,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create %dx%d array: %v", size, size, err)
	}
	defer arr.Destroy()

	// Generate hysteresis data
	material := ferroelectric.DefaultHZO()
	model := ferroelectric.NewPreisachModel(material)
	_, P := model.GetHysteresisLoop(2*material.Ec, size*size)

	// Program all cells
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			idx := (i*size + j) % len(P)
			normPol := P[idx] / material.Ps
			conductance := (normPol + 1.0) / 2.0
			arr.ProgramWeight(i, j, conductance)
		}
	}

	// Run MVM
	input := make([]float64, size)
	for i := range input {
		input[i] = float64(i) / float64(size)
	}

	output, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM failed on %dx%d array: %v", size, size, err)
	}

	if len(output) != size {
		t.Errorf("Expected %d outputs, got %d", size, len(output))
	}

	// Verify no NaN/Inf in output
	for i, v := range output {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("Invalid output[%d] = %v", i, v)
		}
	}

	t.Logf("Large array test: %dx%d, MVM completed successfully", size, size)
}

// TestExtremeTemperatureIntegration tests the workflow at extreme temperatures.
func TestExtremeTemperatureIntegration(t *testing.T) {
	temperatures := []float64{200, 300, 400, 500} // Kelvin

	material := ferroelectric.DefaultHZO()

	for _, temp := range temperatures {
		t.Run(string(rune('0'+int(temp)))+"K", func(t *testing.T) {
			model := ferroelectric.NewPreisachModel(material)
			model.SetTemperature(temp)

			// Verify model still works
			_, P := model.GetHysteresisLoop(2*material.Ec, 50)

			if len(P) == 0 {
				t.Error("Failed to generate hysteresis at temperature")
			}

			// At high temps approaching Curie, polarization should decrease
			var maxP float64
			for _, p := range P {
				if math.Abs(p) > maxP {
					maxP = math.Abs(p)
				}
			}

			if temp < material.CurieTemp {
				if maxP < 0.01 {
					t.Errorf("Polarization too low at T=%v K", temp)
				}
			}

			t.Logf("T=%vK: max|P|=%.4f C/m²", temp, maxP)
		})
	}
}

// TestVisualizationWithEmptyArray tests visualization edge cases.
func TestVisualizationWithEmptyArray(t *testing.T) {
	cfg := &crossbar.Config{
		Rows:    4,
		Cols:    4,
		ADCBits: 8,
		DACBits: 8,
	}
	arr, err := crossbar.NewArray(cfg)
	if err != nil {
		t.Fatalf("Failed to create array: %v", err)
	}
	defer arr.Destroy()

	// Don't program anything - all zeros
	viz := visualization.NewTerminalVisualizer(arr, false)

	// Should not panic with all-zero array
	viz.ShowCrossbarState()

	// MVM with zeros
	input := make([]float64, 4)
	output, _ := arr.MVM(input)
	viz.ShowMVMOperation(input, output)

	t.Log("Visualization with empty/zero array verified")
}

// =============================================================================
// BENCHMARKS
// =============================================================================

func BenchmarkHysteresisToCrossbarWorkflow(b *testing.B) {
	material := ferroelectric.DefaultHZO()
	model := ferroelectric.NewPreisachModel(material)

	cfg := &crossbar.Config{
		Rows:       16,
		Cols:       16,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, _ := crossbar.NewArray(cfg)
	defer arr.Destroy()

	_, P := model.GetHysteresisLoop(2*material.Ec, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Program weights
		for row := 0; row < 16; row++ {
			for col := 0; col < 16; col++ {
				idx := (row*16 + col) % len(P)
				conductance := (P[idx]/material.Ps + 1.0) / 2.0
				arr.ProgramWeight(row, col, conductance)
			}
		}

		// Run MVM
		input := make([]float64, 16)
		for j := range input {
			input[j] = 0.5
		}
		arr.MVM(input)
	}
}

func BenchmarkFullPipelineVisualization(b *testing.B) {
	material := ferroelectric.DefaultHZO()
	model := ferroelectric.NewPreisachModel(material)

	cfg := &crossbar.Config{
		Rows:       8,
		Cols:       8,
		NoiseLevel: 0.0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, _ := crossbar.NewArray(cfg)
	defer arr.Destroy()

	_, P := model.GetHysteresisLoop(2*material.Ec, 64)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			idx := (i*8 + j) % len(P)
			conductance := (P[idx]/material.Ps + 1.0) / 2.0
			arr.ProgramWeight(i, j, conductance)
		}
	}

	input := make([]float64, 8)
	for i := range input {
		input[i] = 0.5
	}

	// Pre-compute MVM
	output, _ := arr.MVM(input)

	// Redirect stdout to suppress output during benchmark
	var buf bytes.Buffer
	_ = buf // Visualizer writes to stdout directly

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		viz := visualization.NewTerminalVisualizer(arr, false)
		viz.ShowCrossbarState()
		viz.ShowMVMOperation(input, output)
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// captureOutput is a helper to capture stdout output (simplified version).
// In production, this would redirect os.Stdout.
func captureOutput(f func()) string {
	var buf strings.Builder
	f()
	return buf.String()
}

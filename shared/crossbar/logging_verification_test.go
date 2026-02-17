package crossbar

import (
	"testing"
)

// TestLoggingIntegration verifies that logging infrastructure is properly integrated.
// This test doesn't check log output (which would be implementation-specific),
// but ensures logging calls don't break functionality.
func TestLoggingIntegration(t *testing.T) {
	t.Run("Array creation logs without error", func(t *testing.T) {
		cfg := &Config{
			Rows:       4,
			Cols:       4,
			NoiseLevel: 0.05,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, err := NewArray(cfg)
		if err != nil {
			t.Fatalf("NewArray failed: %v", err)
		}
		if arr == nil {
			t.Fatal("NewArray returned nil")
		}
	})

	t.Run("QuantizeToLevels logs without error", func(t *testing.T) {
		result := QuantizeToLevels(0.5)
		if result < 0 || result > 1 {
			t.Errorf("QuantizeToLevels returned out-of-range value: %f", result)
		}
	})

	t.Run("ProgramWeight logs without error", func(t *testing.T) {
		cfg := &Config{
			Rows:       4,
			Cols:       4,
			NoiseLevel: 0,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, _ := NewArray(cfg)
		err := arr.ProgramWeight(0, 0, 0.5)
		if err != nil {
			t.Fatalf("ProgramWeight failed: %v", err)
		}
	})

	t.Run("MVM logs without error", func(t *testing.T) {
		cfg := &Config{
			Rows:       4,
			Cols:       4,
			NoiseLevel: 0,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, _ := NewArray(cfg)
		input := []float64{1, 0.5, 0.25, 0}
		output, err := arr.MVM(input)
		if err != nil {
			t.Fatalf("MVM failed: %v", err)
		}
		if len(output) != 4 {
			t.Errorf("Expected 4 outputs, got %d", len(output))
		}
	})

	t.Run("VMM logs without error", func(t *testing.T) {
		cfg := &Config{
			Rows:       4,
			Cols:       4,
			NoiseLevel: 0,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, _ := NewArray(cfg)
		input := []float64{1, 0.5, 0.25, 0}
		output, err := arr.VMM(input)
		if err != nil {
			t.Fatalf("VMM failed: %v", err)
		}
		if len(output) != 4 {
			t.Errorf("Expected 4 outputs, got %d", len(output))
		}
	})

	t.Run("AnalyzeIRDrop logs without error", func(t *testing.T) {
		cfg := &Config{
			Rows:       4,
			Cols:       4,
			NoiseLevel: 0,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, _ := NewArray(cfg)
		arr.ProgramWeight(0, 0, 0.5)
		input := []float64{1, 1, 1, 1}
		result := arr.AnalyzeIRDrop(input, nil)
		if result == nil {
			t.Fatal("AnalyzeIRDrop returned nil")
		}
	})

	t.Run("AnalyzeSneakPaths logs without error", func(t *testing.T) {
		cfg := &Config{
			Rows:       4,
			Cols:       4,
			NoiseLevel: 0,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, _ := NewArray(cfg)
		arr.ProgramWeight(0, 0, 0.5)
		result := arr.AnalyzeSneakPaths(0, 0)
		if result == nil {
			t.Fatal("AnalyzeSneakPaths returned nil")
		}
	})

	t.Run("MVMWithIRDrop logs without error", func(t *testing.T) {
		cfg := &Config{
			Rows:       4,
			Cols:       4,
			NoiseLevel: 0,
			ADCBits:    8,
			DACBits:    8,
		}
		arr, _ := NewArray(cfg)
		input := []float64{1, 1, 1, 1}
		output, analysis, err := arr.MVMWithIRDrop(input, nil)
		if err != nil {
			t.Fatalf("MVMWithIRDrop failed: %v", err)
		}
		if len(output) != 4 {
			t.Errorf("Expected 4 outputs, got %d", len(output))
		}
		if analysis == nil {
			t.Fatal("MVMWithIRDrop returned nil analysis")
		}
	})

	t.Run("ComputeError logs without error", func(t *testing.T) {
		ideal := []float64{1, 0.8, 0.6, 0.4}
		actual := []float64{0.9, 0.75, 0.55, 0.38}
		err := ComputeError(ideal, actual)
		if err < 0 || err > 1 {
			t.Errorf("ComputeError returned out-of-range value: %f", err)
		}
	})

	t.Run("NewDriftSimulator logs without error", func(t *testing.T) {
		sim := NewDriftSimulator(4, 4, 30)
		if sim == nil {
			t.Fatal("NewDriftSimulator returned nil")
		}
	})

	t.Run("DriftSimulator.SimulateTimeStep logs without error", func(t *testing.T) {
		sim := NewDriftSimulator(4, 4, 30)
		sim.SimulateTimeStep(1000) // 1000 seconds
		if sim.Time != 1000 {
			t.Errorf("Expected time=1000, got %f", sim.Time)
		}
	})

	t.Run("NewIRDropSimulator logs without error", func(t *testing.T) {
		sim := NewIRDropSimulator(4, 4)
		if sim == nil {
			t.Fatal("NewIRDropSimulator returned nil")
		}
	})

	t.Run("IRDropSimulator.Simulate logs without error", func(t *testing.T) {
		sim := NewIRDropSimulator(4, 4)
		sim.SetAllInputs([]float64{1, 1, 1, 1})
		sim.Simulate(10)
		// Just verify it doesn't crash
	})

	t.Run("NewSneakPathAnalyzer logs without error", func(t *testing.T) {
		analyzer := NewSneakPathAnalyzer(4, 4)
		if analyzer == nil {
			t.Fatal("NewSneakPathAnalyzer returned nil")
		}
	})

	t.Run("SneakPathAnalyzer.AnalyzeTarget logs without error", func(t *testing.T) {
		analyzer := NewSneakPathAnalyzer(4, 4)
		analyzer.AnalyzeTarget(1, 1, 1.0)
		// Just verify it doesn't crash
	})
}

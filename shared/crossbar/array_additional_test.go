package crossbar

import (
	"math"
	"testing"
)

func TestArrayAdditionalOperationsCoverage(t *testing.T) {
	arr, err := NewArray(&Config{Rows: 3, Cols: 3, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatal(err)
	}

	// Program matrix success path
	weights := [][]float64{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}}
	if err := arr.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("ProgramWeightMatrix success path failed: %v", err)
	}

	// MVM CPU path and stats
	out, err := arr.MVM([]float64{0.2, 0.5, 0.8})
	if err != nil {
		t.Fatalf("MVM failed: %v", err)
	}
	if len(out) != arr.Rows() {
		t.Fatalf("unexpected output length: got %d, want %d", len(out), arr.Rows())
	}
	reads, writes := arr.GetStats()
	if reads == 0 || writes == 0 {
		t.Fatalf("expected non-zero stats, got reads=%d writes=%d", reads, writes)
	}

	// Conductance matrix helpers
	eff := arr.GetEffectiveConductanceMatrix()
	if len(eff) != 3 || len(eff[0]) != 3 {
		t.Fatalf("unexpected effective matrix shape: %dx%d", len(eff), len(eff[0]))
	}
	cfgCopy := arr.GetConfig()
	if cfgCopy.Rows != 3 || cfgCopy.Cols != 3 {
		t.Fatalf("GetConfig returned wrong dimensions: %+v", cfgCopy)
	}

	// Cell stats path
	cellStats, err := arr.GetCellStats(1, 1)
	if err != nil {
		t.Fatalf("GetCellStats failed: %v", err)
	}
	if cellStats.Row != 1 || cellStats.Col != 1 {
		t.Fatalf("unexpected cell stats coordinates: %+v", cellStats)
	}

	// Cycle aging/reset paths
	arr.AgeCycles(10)
	aged, err := arr.GetCellStats(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if aged.SwitchingCount < 10 {
		t.Fatalf("expected switching count to age, got %d", aged.SwitchingCount)
	}
	arr.ResetCycleCounts()
	reset, err := arr.GetCellStats(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if reset.SwitchingCount != 0 {
		t.Fatalf("expected switching count reset, got %d", reset.SwitchingCount)
	}

	// Cleanup path with nil GPU accelerator
	arr.Destroy()
}

func TestArrayAdditionalErrorAndBranchCoverage(t *testing.T) {
	arr, err := NewArray(&Config{Rows: 2, Cols: 2, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatal(err)
	}

	// ProgramWeightMatrix row overflow
	if err := arr.ProgramWeightMatrix([][]float64{{0, 0}, {0, 0}, {0, 0}}); err == nil {
		t.Fatal("expected row overflow error")
	}
	// ProgramWeightMatrix col overflow
	if err := arr.ProgramWeightMatrix([][]float64{{0, 0, 0}}); err == nil {
		t.Fatal("expected col overflow error")
	}

	// MVM input overflow
	if _, err := arr.MVM([]float64{0, 0, 0}); err == nil {
		t.Fatal("expected MVM input overflow error")
	}

	// GetCellStats out of range
	if _, err := arr.GetCellStats(-1, 0); err == nil {
		t.Fatal("expected GetCellStats bounds error")
	}

	// initialNoiseFactor with process variation override and sigma<=0
	fNoNoise := initialNoiseFactor(&Config{NoiseLevel: -1})
	if fNoNoise != 1.0 {
		t.Fatalf("expected 1.0 for non-positive sigma, got %f", fNoNoise)
	}

	fPVNoNoise := initialNoiseFactor(&Config{NoiseLevel: 0.5, ProcessVariation: &ProcessVariationConfig{DeviceSigma: 0}})
	if fPVNoNoise != 1.0 {
		t.Fatalf("expected 1.0 for zero process sigma, got %f", fPVNoNoise)
	}

	// Trigger clamping branch statistically with huge sigma
	clampedSeen := false
	for i := 0; i < 2000; i++ {
		if initialNoiseFactor(&Config{NoiseLevel: 10}) == 0 {
			clampedSeen = true
			break
		}
	}
	if !clampedSeen {
		t.Fatal("expected to observe clamped noise factor == 0 with large sigma")
	}

	// initGPU branches (no GPU required): false branch, then true-attempt branch
	arr.initGPU() // UseGPU false in config -> early return

	gpuArr, err := NewArray(&Config{Rows: 2, Cols: 2, NoiseLevel: 0, ADCBits: 8, DACBits: 8, UseGPU: true})
	if err != nil {
		t.Fatal(err)
	}
	gpuArr.initGPU()
	if !gpuArr.gpuInitialized {
		t.Fatal("expected gpuInitialized after init attempt")
	}
	gpuArr.uploadVariationToGPU() // nil-safe when accelerator unavailable
	gpuArr.Destroy()

	// SetConductanceTable size validation
	if err := arr.SetConductanceTable([]float64{1, 2}); err == nil {
		t.Fatal("expected conductance table size error")
	}

	// Disturb shift clamping in GetPhysicalConductanceForCell
	arr.config.HalfSelect = &HalfSelectConfig{Enabled: true}
	arr.cells[0][0].Conductance = 0.9
	arr.cells[0][0].DisturbShift = 0.5
	gHigh := arr.GetPhysicalConductanceForCell(0, 0)
	if math.IsNaN(gHigh) || gHigh <= 0 {
		t.Fatalf("unexpected conductance with high disturb shift: %f", gHigh)
	}
	arr.cells[0][0].DisturbShift = -2
	gLow := arr.GetPhysicalConductanceForCell(0, 0)
	if math.IsNaN(gLow) || gLow <= 0 {
		t.Fatalf("unexpected conductance with low disturb shift: %f", gLow)
	}
}

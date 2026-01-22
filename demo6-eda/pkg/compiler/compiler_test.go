// pkg/compiler/compiler_test.go
package compiler

import (
	"testing"
)

func TestCompile_Basic(t *testing.T) {
	weights := [][]float64{
		{0.1, -0.2, 0.3},
		{-0.4, 0.5, -0.6},
	}

	config := DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8

	mapping, err := Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Check basic stats
	if mapping.Stats.UsedCells != 6 {
		t.Errorf("Expected 6 used cells, got %d", mapping.Stats.UsedCells)
	}

	if mapping.Stats.Utilization < 0 || mapping.Stats.Utilization > 1 {
		t.Errorf("Invalid utilization: %f", mapping.Stats.Utilization)
	}

	// Check all cells have valid conductance
	for _, cell := range mapping.Cells {
		if cell.Conductance < config.GMin || cell.Conductance > config.GMax {
			t.Errorf("Cell conductance %f outside range [%f, %f]",
				cell.Conductance, config.GMin, config.GMax)
		}
	}

	t.Logf("Stats: %+v", mapping.Stats)
}

func TestCompile_Quantization(t *testing.T) {
	// Test that quantization preserves information
	weights := [][]float64{{0.0, 0.5, 1.0, -0.5, -1.0}}

	config := DefaultConfig()
	config.Levels = 30
	config.ArrayRows = 8
	config.ArrayCols = 8

	mapping, err := Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// PSNR should be reasonable for 30 levels
	if mapping.Stats.QuantPSNR < 20 {
		t.Errorf("PSNR too low: %f dB", mapping.Stats.QuantPSNR)
	}

	t.Logf("PSNR: %.2f dB, MSE: %.6f", mapping.Stats.QuantPSNR, mapping.Stats.QuantMSE)
}

func TestCompile_SizeValidation(t *testing.T) {
	weights := [][]float64{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, // 10 cols
	}

	config := DefaultConfig()
	config.ArrayCols = 5 // Too small

	_, err := Compile(weights, config)
	if err == nil {
		t.Error("Expected error for oversized weights")
	}
}

func TestCompile_EmptyMatrix(t *testing.T) {
	weights := [][]float64{}

	config := DefaultConfig()
	_, err := Compile(weights, config)
	if err == nil {
		t.Error("Expected error for empty matrix")
	}
}

func TestCompile_8x8Sample(t *testing.T) {
	// Sample from plan file
	weights := [][]float64{
		{0.1, -0.2, 0.3, -0.4, 0.5, -0.6, 0.7, -0.8},
		{-0.1, 0.2, -0.3, 0.4, -0.5, 0.6, -0.7, 0.8},
		{0.15, -0.25, 0.35, -0.45, 0.55, -0.65, 0.75, -0.85},
		{-0.15, 0.25, -0.35, 0.45, -0.55, 0.65, -0.75, 0.85},
		{0.05, -0.15, 0.25, -0.35, 0.45, -0.55, 0.65, -0.75},
		{-0.05, 0.15, -0.25, 0.35, -0.45, 0.55, -0.65, 0.75},
		{0.2, -0.3, 0.4, -0.5, 0.6, -0.7, 0.8, -0.9},
		{-0.2, 0.3, -0.4, 0.5, -0.6, 0.7, -0.8, 0.9},
	}

	config := DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8

	mapping, err := Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Check expected stats
	if mapping.Stats.UsedCells != 64 {
		t.Errorf("Expected 64 used cells, got %d", mapping.Stats.UsedCells)
	}

	if mapping.Stats.Utilization != 1.0 {
		t.Errorf("Expected 100%% utilization, got %f", mapping.Stats.Utilization)
	}

	t.Logf("8x8 sample compiled successfully:")
	t.Logf("  UsedCells: %d", mapping.Stats.UsedCells)
	t.Logf("  UniqueLevels: %d", mapping.Stats.UniqueLevels)
	t.Logf("  WeightRange: [%.2f, %.2f]", mapping.Stats.WeightMin, mapping.Stats.WeightMax)
	t.Logf("  QuantPSNR: %.2f dB", mapping.Stats.QuantPSNR)
}

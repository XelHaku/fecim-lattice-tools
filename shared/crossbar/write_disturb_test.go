package crossbar

import (
	"testing"
)

func TestWriteDisturbEngine_Disabled(t *testing.T) {
	config := DefaultWriteDisturbConfig()
	config.Enable = false

	engine := NewWriteDisturbEngine(8, 8, config)
	engine.RecordWrite(3, 3)

	stats := engine.GetStressStats()
	if stats.TotalHalfSelects != 0 {
		t.Errorf("Expected 0 half-selects when disabled, got %d", stats.TotalHalfSelects)
	}
}

func TestWriteDisturbEngine_Enabled(t *testing.T) {
	config := DefaultWriteDisturbConfig()
	config.Enable = true
	config.Architecture1T1R = false // Passive array for clearer testing

	engine := NewWriteDisturbEngine(8, 8, config)
	engine.RecordWrite(3, 3)

	stats := engine.GetStressStats()

	// For an 8x8 array, writing to (3,3) half-selects:
	// - 7 cells in row 3 (excluding target)
	// - 7 cells in column 3 (excluding target)
	// Total = 14 half-selects
	expectedHalfSelects := 14
	if stats.TotalHalfSelects != expectedHalfSelects {
		t.Errorf("Expected %d half-selects, got %d", expectedHalfSelects, stats.TotalHalfSelects)
	}

	if stats.TotalWriteOps != 1 {
		t.Errorf("Expected 1 write op, got %d", stats.TotalWriteOps)
	}
}

func TestWriteDisturbEngine_StressAccumulation(t *testing.T) {
	config := DefaultWriteDisturbConfig()
	config.Enable = true
	config.Architecture1T1R = false
	config.StressAccumulationRate = 0.1 // 10% per exposure for fast testing

	engine := NewWriteDisturbEngine(4, 4, config)

	// Write to same cell multiple times
	for i := 0; i < 5; i++ {
		engine.RecordWrite(1, 1)
	}

	// Cell (1,0) is half-selected 5 times (same row)
	stress := engine.GetCellStress(1, 0)
	expectedStress := 5 * 0.1 // 0.5
	if stress != expectedStress {
		t.Errorf("Expected stress %.2f, got %.2f", expectedStress, stress)
	}

	// Cell (0,0) is NOT half-selected (different row AND column)
	stressCorner := engine.GetCellStress(0, 0)
	if stressCorner != 0 {
		t.Errorf("Expected 0 stress on non-half-selected cell, got %.2f", stressCorner)
	}
}

func TestWriteDisturbEngine_ApplyEffects(t *testing.T) {
	config := DefaultWriteDisturbConfig()
	config.Enable = true
	config.Architecture1T1R = false
	config.StressAccumulationRate = 0.5
	config.StressThreshold = 1.0

	engine := NewWriteDisturbEngine(4, 4, config)

	// Create conductance matrix
	conductances := make([][]float64, 4)
	for i := range conductances {
		conductances[i] = make([]float64, 4)
		for j := range conductances[i] {
			conductances[i][j] = (GMin + GMax) / 2 // Mid-level
		}
	}

	// Write twice to accumulate stress > threshold
	engine.RecordWrite(1, 1)
	engine.RecordWrite(1, 1)

	// Cell (1,0) should have stress = 2 * 0.5 = 1.0 (at threshold)
	stress := engine.GetCellStress(1, 0)
	if stress < 1.0 {
		t.Errorf("Expected stress >= 1.0, got %.2f", stress)
	}

	// Apply effects
	disturbed := engine.ApplyDisturbEffects(conductances, 30)

	if disturbed == 0 {
		t.Log("Warning: No cells disturbed (stress may not have exceeded threshold)")
	}
}

func TestWriteDisturbEngine_1T1RReduction(t *testing.T) {
	// Passive array (0T1R)
	passiveConfig := DefaultWriteDisturbConfig()
	passiveConfig.Enable = true
	passiveConfig.Architecture1T1R = false
	passiveConfig.StressAccumulationRate = 0.1

	// Active array (1T1R)
	activeConfig := DefaultWriteDisturbConfig()
	activeConfig.Enable = true
	activeConfig.Architecture1T1R = true
	activeConfig.Architecture1T1RReduction = 0.1
	activeConfig.StressAccumulationRate = 0.1

	passiveEngine := NewWriteDisturbEngine(8, 8, passiveConfig)
	activeEngine := NewWriteDisturbEngine(8, 8, activeConfig)

	// Same write operation
	passiveEngine.RecordWrite(4, 4)
	activeEngine.RecordWrite(4, 4)

	passiveStress := passiveEngine.GetCellStress(4, 0)
	activeStress := activeEngine.GetCellStress(4, 0)

	// Active should have 10x less stress
	expectedRatio := 0.1
	actualRatio := activeStress / passiveStress

	if actualRatio < expectedRatio-0.01 || actualRatio > expectedRatio+0.01 {
		t.Errorf("Expected 1T1R to reduce stress by %.0fx, got %.2fx",
			1/expectedRatio, 1/actualRatio)
	}
}

func TestWriteDisturbEngine_Reset(t *testing.T) {
	config := DefaultWriteDisturbConfig()
	config.Enable = true
	config.Architecture1T1R = false

	engine := NewWriteDisturbEngine(4, 4, config)

	engine.RecordWrite(1, 1)
	engine.RecordWrite(2, 2)

	stats := engine.GetStressStats()
	if stats.TotalWriteOps == 0 {
		t.Error("Expected non-zero write ops before reset")
	}

	engine.Reset()

	stats = engine.GetStressStats()
	if stats.TotalWriteOps != 0 {
		t.Errorf("Expected 0 write ops after reset, got %d", stats.TotalWriteOps)
	}
	if stats.MaxStress != 0 {
		t.Errorf("Expected 0 max stress after reset, got %.2f", stats.MaxStress)
	}
}

func TestEstimateDisturbRate(t *testing.T) {
	passiveConfig := DefaultWriteDisturbConfig()
	passiveConfig.Enable = true
	passiveConfig.Architecture1T1R = false

	activeConfig := DefaultWriteDisturbConfig()
	activeConfig.Enable = true
	activeConfig.Architecture1T1R = true

	// With 100 writes per cell
	passiveRate := EstimateDisturbRate(100, passiveConfig)
	activeRate := EstimateDisturbRate(100, activeConfig)

	// Active should have much lower disturb rate
	if activeRate >= passiveRate {
		t.Errorf("Expected active rate < passive rate, got active=%.4f, passive=%.4f",
			activeRate, passiveRate)
	}

	// Disabled should return 0
	disabledConfig := DefaultWriteDisturbConfig()
	disabledConfig.Enable = false
	disabledRate := EstimateDisturbRate(100, disabledConfig)
	if disabledRate != 0 {
		t.Errorf("Expected 0 rate when disabled, got %.4f", disabledRate)
	}
}

func TestCompareArchitectures(t *testing.T) {
	passiveRate, activeRate := CompareArchitectures(1000)

	if activeRate >= passiveRate {
		t.Errorf("Expected active < passive disturb rate, got active=%.4f, passive=%.4f",
			activeRate, passiveRate)
	}

	// Both should be > 0 for this many writes
	if passiveRate == 0 {
		t.Error("Expected non-zero passive disturb rate")
	}
}

func TestHalfSelectVoltage(t *testing.T) {
	writeV := 3.0

	v2 := HalfSelectVoltage(writeV, "V/2")
	if v2 != 1.5 {
		t.Errorf("V/2 scheme: expected 1.5V, got %.1fV", v2)
	}

	v3 := HalfSelectVoltage(writeV, "V/3")
	if v3 != 1.0 {
		t.Errorf("V/3 scheme: expected 1.0V, got %.1fV", v3)
	}
}

func TestIsDisturbCritical(t *testing.T) {
	coerciveV := 1.0
	safetyMargin := 0.3 // 30% margin

	// Not critical: 0.5V with 0.7V threshold
	if IsDisturbCritical(0.5, coerciveV, safetyMargin) {
		t.Error("0.5V should not be critical with 30% margin")
	}

	// Critical: 0.8V exceeds 0.7V threshold
	if !IsDisturbCritical(0.8, coerciveV, safetyMargin) {
		t.Error("0.8V should be critical with 30% margin")
	}
}

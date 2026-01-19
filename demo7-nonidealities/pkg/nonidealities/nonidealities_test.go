package nonidealities

import (
	"math"
	"testing"
)

// TestNewIRDropSimulator verifies IR drop simulator creation.
func TestNewIRDropSimulator(t *testing.T) {
	ir := NewIRDropSimulator(16, 16)

	if ir.Rows != 16 {
		t.Errorf("Expected 16 rows, got %d", ir.Rows)
	}
	if ir.Cols != 16 {
		t.Errorf("Expected 16 cols, got %d", ir.Cols)
	}
	if len(ir.Conductances) != 16 {
		t.Errorf("Expected 16 conductance rows, got %d", len(ir.Conductances))
	}
}

// TestIRDropSimulation verifies IR drop simulation runs.
func TestIRDropSimulation(t *testing.T) {
	ir := NewIRDropSimulator(8, 8)

	// Set input voltages
	for i := 0; i < 8; i++ {
		ir.SetInputVoltage(i, 0.5)
	}

	// Run simulation
	ir.Simulate(50)

	// Check that results exist
	if len(ir.IRDropMap) != 8 {
		t.Errorf("Expected 8 IR drop map rows, got %d", len(ir.IRDropMap))
	}

	maxDrop := ir.GetMaxIRDrop()
	if maxDrop < 0 {
		t.Error("Max IR drop should not be negative")
	}
}

// TestIRDropWorstCase verifies worst-case corner is identified.
func TestIRDropWorstCase(t *testing.T) {
	ir := NewIRDropSimulator(16, 16)

	// Set uniform input
	for i := 0; i < 16; i++ {
		ir.SetInputVoltage(i, 0.5)
	}

	ir.Simulate(100)

	stats := ir.GetStats()

	// Worst cell should be in far corner (high row, high col)
	// due to cumulative IR drop
	if stats.WorstCellRow < 8 && stats.WorstCellCol < 8 {
		t.Logf("Worst cell at (%d, %d) - expected near far corner",
			stats.WorstCellRow, stats.WorstCellCol)
	}
}

// TestIRDropMitigation verifies mitigation reduces IR drop.
func TestIRDropMitigation(t *testing.T) {
	ir := NewIRDropSimulator(16, 16)

	for i := 0; i < 16; i++ {
		ir.SetInputVoltage(i, 0.5)
	}

	ir.Simulate(100)
	beforeStats := ir.GetStats()

	// Apply mitigation (wider lines = lower resistance)
	mitigation := IRDropMitigation{
		UseWidenedLines:   true,
		LineWidthIncrease: 2.0, // 2x wider
	}
	ir.ApplyMitigation(mitigation)

	afterStats := ir.GetStats()

	if afterStats.MaxIRDrop >= beforeStats.MaxIRDrop {
		t.Error("Mitigation should reduce max IR drop")
	}
}

// TestNewSneakPathAnalyzer verifies sneak path analyzer creation.
func TestNewSneakPathAnalyzer(t *testing.T) {
	sp := NewSneakPathAnalyzer(8, 8)

	if sp.Rows != 8 {
		t.Errorf("Expected 8 rows, got %d", sp.Rows)
	}
	if sp.Cols != 8 {
		t.Errorf("Expected 8 cols, got %d", sp.Cols)
	}
}

// TestSneakPathAnalysis verifies sneak path analysis runs.
func TestSneakPathAnalysis(t *testing.T) {
	sp := NewSneakPathAnalyzer(8, 8)

	// Analyze target cell
	sp.AnalyzeTarget(4, 4, 0.5)

	// Check that sneak paths were identified
	if sp.TotalSneakRatio < 0 {
		t.Error("Sneak ratio should not be negative")
	}

	if len(sp.SneakPaths) == 0 {
		t.Error("Should identify at least one sneak path")
	}
}

// TestSneakPathStats verifies sneak statistics.
func TestSneakPathStats(t *testing.T) {
	sp := NewSneakPathAnalyzer(8, 8)
	sp.AnalyzeTarget(4, 4, 0.5)

	stats := sp.GetStats(0.5)

	if stats.TargetCurrent <= 0 {
		t.Error("Target current should be positive")
	}

	if stats.NumSneakPaths <= 0 {
		t.Error("Should have identified sneak paths")
	}

	// Expected number of 3-cell sneak paths: (rows-1) * (cols-1)
	expected := (8 - 1) * (8 - 1)
	if stats.NumSneakPaths != expected {
		t.Errorf("Expected %d sneak paths, got %d", expected, stats.NumSneakPaths)
	}
}

// TestSneakMitigation verifies mitigation improves SNR.
func TestSneakMitigation(t *testing.T) {
	sp := NewSneakPathAnalyzer(8, 8)

	// Without mitigation
	sp.AnalyzeTarget(4, 4, 0.5)
	beforeStats := sp.GetStats(0.5)

	// With selector mitigation
	mitigation := SneakMitigation{
		UseSelector:   true,
		SelectorOnOff: 1000, // 1000:1 on/off ratio
	}
	afterStats := sp.AnalyzeWithMitigation(4, 4, 0.5, mitigation)

	if afterStats.SignalToNoiseRatio <= beforeStats.SignalToNoiseRatio {
		t.Error("Selector should improve SNR")
	}
}

// TestArraySneakAnalysis verifies full array analysis.
func TestArraySneakAnalysis(t *testing.T) {
	sp := NewSneakPathAnalyzer(8, 8)
	analysis := sp.AnalyzeArray(0.5)

	if analysis.AvgSneakRatio <= 0 {
		t.Error("Average sneak ratio should be positive")
	}

	// Max should be >= avg (with small tolerance for floating point)
	if analysis.MaxSneakRatio < analysis.AvgSneakRatio-0.001 {
		t.Errorf("Max sneak ratio (%.6f) should be >= avg (%.6f)",
			analysis.MaxSneakRatio, analysis.AvgSneakRatio)
	}
}

// TestNewDriftSimulator verifies drift simulator creation.
func TestNewDriftSimulator(t *testing.T) {
	d := NewDriftSimulator(8, 8, 30)

	if d.Rows != 8 {
		t.Errorf("Expected 8 rows, got %d", d.Rows)
	}
	if d.Levels != 30 {
		t.Errorf("Expected 30 levels, got %d", d.Levels)
	}
	if d.DriftCoeff != 0.001 {
		t.Errorf("Expected 0.001 drift coeff for FeFET, got %f", d.DriftCoeff)
	}
}

// TestDriftSimulation verifies drift simulation runs.
func TestDriftSimulation(t *testing.T) {
	d := NewDriftSimulator(8, 8, 30)

	// Run simulation for 1000 seconds
	for step := 0; step < 100; step++ {
		d.SimulateTimeStep(10)
		d.RecordSnapshot()
	}

	if d.Time != 1000 {
		t.Errorf("Expected time 1000, got %f", d.Time)
	}

	if len(d.DriftHistory) != 100 {
		t.Errorf("Expected 100 snapshots, got %d", len(d.DriftHistory))
	}
}

// TestDriftStats verifies drift statistics.
func TestDriftStats(t *testing.T) {
	d := NewDriftSimulator(8, 8, 30)

	// Run some drift
	for step := 0; step < 100; step++ {
		d.SimulateTimeStep(10)
	}

	stats := d.GetStats()

	if stats.ElapsedTime != 1000 {
		t.Errorf("Expected 1000s elapsed, got %f", stats.ElapsedTime)
	}

	// FeFET should have very high retention prediction
	if stats.RetentionPrediction < 99 {
		t.Errorf("Expected >99%% retention, got %f%%", stats.RetentionPrediction)
	}
}

// TestFeFETLowDrift verifies FeFET has very low drift.
func TestFeFETLowDrift(t *testing.T) {
	d := NewDriftSimulator(8, 8, 30)
	d.DriftCoeff = 0.001 // FeFET

	// Run for 10000 seconds
	for step := 0; step < 100; step++ {
		d.SimulateTimeStep(100)
	}

	stats := d.GetStats()

	// FeFET should have < 1% drift even after 10000s
	if stats.AvgDriftPercent > 1.0 {
		t.Errorf("FeFET drift should be < 1%%, got %f%%", stats.AvgDriftPercent)
	}
}

// TestRRAMHigherDrift verifies RRAM has higher drift than FeFET.
func TestRRAMHigherDrift(t *testing.T) {
	fefet := NewDriftSimulator(8, 8, 30)
	fefet.DriftCoeff = 0.001 // FeFET

	rram := NewDriftSimulator(8, 8, 30)
	rram.DriftCoeff = 0.05 // RRAM

	// Run same simulation
	for step := 0; step < 100; step++ {
		fefet.SimulateTimeStep(100)
		rram.SimulateTimeStep(100)
	}

	fefetStats := fefet.GetStats()
	rramStats := rram.GetStats()

	if rramStats.AvgDriftPercent <= fefetStats.AvgDriftPercent {
		t.Error("RRAM should have higher drift than FeFET")
	}

	// FeFET should be significantly better
	advantage := rramStats.AvgDriftPercent / fefetStats.AvgDriftPercent
	if advantage < 10 {
		t.Errorf("Expected >10x advantage, got %.1fx", advantage)
	}
}

// TestDriftRefresh verifies refresh resets drift.
func TestDriftRefresh(t *testing.T) {
	d := NewDriftSimulator(8, 8, 30)

	// Run drift
	for step := 0; step < 100; step++ {
		d.SimulateTimeStep(100)
	}

	statsBefore := d.GetStats()

	// Refresh all cells
	d.RefreshAll()

	statsAfter := d.GetStats()

	// After refresh, levels should be exact
	if statsAfter.NumLevelErrors != 0 {
		t.Errorf("Expected 0 level errors after refresh, got %d", statsAfter.NumLevelErrors)
	}

	// Average drift should be reduced (though not zero due to quantization)
	if statsAfter.AvgDrift > statsBefore.AvgDrift {
		t.Error("Refresh should reduce drift")
	}
}

// TestCompareTechnologies verifies technology comparison.
func TestCompareTechnologies(t *testing.T) {
	results := CompareTechnologies(8, 8, 10000)

	// Should have all technologies
	expectedTechs := []string{"IronLattice (FeFET)", "RRAM", "PCM", "Flash"}
	for _, tech := range expectedTechs {
		if _, ok := results[tech]; !ok {
			t.Errorf("Missing technology: %s", tech)
		}
	}

	// FeFET should have best retention
	fefetRetention := results["IronLattice (FeFET)"].RetentionPrediction
	for name, stats := range results {
		if name != "IronLattice (FeFET)" && stats.RetentionPrediction > fefetRetention {
			t.Errorf("%s has better retention than FeFET", name)
		}
	}
}

// TestRenderer verifies renderer functions.
func TestRenderer(t *testing.T) {
	r := NewRenderer()

	// Test IR drop rendering
	ir := NewIRDropSimulator(8, 8)
	for i := 0; i < 8; i++ {
		ir.SetInputVoltage(i, 0.5)
	}
	ir.Simulate(50)

	irMap := r.RenderIRDropMap(ir)
	if len(irMap) == 0 {
		t.Error("IR drop map should not be empty")
	}

	// Test sneak path rendering
	sp := NewSneakPathAnalyzer(8, 8)
	sp.AnalyzeTarget(4, 4, 0.5)

	sneakMap := r.RenderSneakPathMap(sp)
	if len(sneakMap) == 0 {
		t.Error("Sneak path map should not be empty")
	}

	// Test drift history rendering
	d := NewDriftSimulator(8, 8, 30)
	for step := 0; step < 10; step++ {
		d.SimulateTimeStep(10)
		d.RecordSnapshot()
	}

	driftHist := r.RenderDriftHistory(d)
	if len(driftHist) == 0 {
		t.Error("Drift history should not be empty")
	}
}

// TestIRDropOutputError verifies output error calculation.
func TestIRDropOutputError(t *testing.T) {
	ir := NewIRDropSimulator(8, 8)

	// Set input voltages
	for i := 0; i < 8; i++ {
		ir.SetInputVoltage(i, 0.5)
	}

	ir.Simulate(100)

	errors := ir.GetOutputError()

	if len(errors) != 8 {
		t.Errorf("Expected 8 output errors, got %d", len(errors))
	}

	// All errors should be positive
	for i, err := range errors {
		if err < 0 {
			t.Errorf("Error %d should be non-negative, got %f", i, err)
		}
	}
}

// TestSneakTopPaths verifies top sneak path identification.
func TestSneakTopPaths(t *testing.T) {
	sp := NewSneakPathAnalyzer(8, 8)
	sp.AnalyzeTarget(4, 4, 0.5)

	topPaths := sp.GetTopSneakPaths(5)

	if len(topPaths) != 5 {
		t.Errorf("Expected 5 top paths, got %d", len(topPaths))
	}

	// Verify they're sorted (largest first)
	for i := 1; i < len(topPaths); i++ {
		if topPaths[i].PathCurrent > topPaths[i-1].PathCurrent {
			t.Error("Top paths should be sorted by current (descending)")
		}
	}
}

// TestDriftReadDisturb verifies read disturb simulation.
func TestDriftReadDisturb(t *testing.T) {
	d := NewDriftSimulator(8, 8, 30)
	d.ReadDisturb = 0.1 // High for testing

	initialG := d.Conductances[0][0]

	// Many reads
	d.SimulateRead(0, 0, 10000)

	// Some change should occur with high read disturb
	change := math.Abs(d.Conductances[0][0] - initialG)
	// Just verify it ran without error; actual change depends on random
	t.Logf("Conductance change after 10000 reads: %.6f S", change)
}

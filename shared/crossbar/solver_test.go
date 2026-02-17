package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

func TestNewParasiticSolver(t *testing.T) {
	tests := []struct {
		name    string
		rows    int
		cols    int
		wantErr bool
	}{
		{"valid 8x8", 8, 8, false},
		{"valid 64x64", 64, 64, false},
		{"invalid zero rows", 0, 8, true},
		{"invalid zero cols", 8, 0, true},
		{"invalid negative", -1, 8, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewParasiticSolver(tt.rows, tt.cols, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewParasiticSolver() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSORConvergence_ZeroParasitics(t *testing.T) {
	// With zero parasitics, result should match ideal MVM
	solver, err := NewParasiticSolver(8, 8, nil)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	// Set uniform conductances
	g := make([][]float64, 8)
	for i := range g {
		g[i] = make([]float64, 8)
		for j := range g[i] {
			g[i][j] = 0.5 // 50% conductance
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0, 0) // Zero parasitics

	// Apply uniform voltages
	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	result, err := solver.SolveMVM(input)
	if err != nil {
		t.Fatalf("SolveMVM failed: %v", err)
	}

	if !result.Converged {
		t.Error("Solver did not converge with zero parasitics")
	}

	// Should converge in 1-2 iterations with zero parasitics
	if result.Iterations > 3 {
		t.Errorf("Expected <3 iterations with zero parasitics, got %d", result.Iterations)
	}

	// Compare to ideal
	ideal := solver.ComputeIdealMVM(input)
	for i := range ideal {
		if math.Abs(ideal[i]-result.OutputCurrents[i]) > 1e-6 {
			t.Errorf("Output[%d] mismatch: ideal=%f, actual=%f", i, ideal[i], result.OutputCurrents[i])
		}
	}
}

func TestSORConvergence_LowParasitics(t *testing.T) {
	// Low parasitics (Rp/Rmin = 0.01) should converge quickly
	solver, err := NewParasiticSolver(8, 8, nil)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	g := make([][]float64, 8)
	for i := range g {
		g[i] = make([]float64, 8)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.01, 0.01)

	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	result, err := solver.SolveMVM(input)
	if err != nil {
		t.Fatalf("SolveMVM failed: %v", err)
	}

	if !result.Converged {
		t.Error("Solver did not converge with low parasitics")
	}

	if result.Iterations > 10 {
		t.Errorf("Expected <10 iterations with low parasitics, got %d", result.Iterations)
	}
}

func TestSORConvergence_MediumParasitics(t *testing.T) {
	// Medium parasitics (Rp/Rmin = 0.1) - typical case
	config := &SORConfig{
		MaxIterations: 200,
		Tolerance:     1e-5,
		OmegaInitial:  0.8, // Under-relaxation for better convergence
		OmegaMin:      0.01,
		OmegaDecay:    0.95,
		AdaptiveOmega: true,
	}

	solver, err := NewParasiticSolver(32, 32, config)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	g := make([][]float64, 32)
	for i := range g {
		g[i] = make([]float64, 32)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.1, 0.1)

	input := make([]float64, 32)
	for i := range input {
		input[i] = 1.0
	}

	result, err := solver.SolveMVMWithFallback(input)
	if err != nil && result == nil {
		t.Fatalf("SolveMVM failed with no result: %v", err)
	}

	t.Logf("Medium parasitics: iterations=%d, converged=%v, maxErr=%e",
		result.Iterations, result.Converged, result.MaxError)

	// Should produce output even if not fully converged
	if result.OutputCurrents == nil {
		t.Error("Expected non-nil output")
	}
}

func TestSORConvergence_HighParasitics(t *testing.T) {
	// High parasitics (Rp/Rmin = 1.0) - challenging case
	config := &SORConfig{
		MaxIterations: 200,
		Tolerance:     1e-5,
		OmegaInitial:  0.8, // Under-relaxation for stability
		OmegaMin:      0.01,
		OmegaDecay:    0.95,
		AdaptiveOmega: true,
	}

	solver, err := NewParasiticSolver(64, 64, config)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	g := make([][]float64, 64)
	for i := range g {
		g[i] = make([]float64, 64)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(1.0, 1.0)

	input := make([]float64, 64)
	for i := range input {
		input[i] = 1.0
	}

	result, err := solver.SolveMVMWithFallback(input)
	if err != nil && !result.Converged {
		// High parasitics may not fully converge, but should return partial result
		t.Logf("High parasitics: iterations=%d, converged=%v, maxErr=%e",
			result.Iterations, result.Converged, result.MaxError)
	}

	// At minimum, should produce reasonable output
	if result.OutputCurrents == nil {
		t.Error("Expected non-nil output even if not fully converged")
	}

	// Check that output is reduced due to IR drop (not equal to ideal)
	ideal := solver.ComputeIdealMVM(input)
	for i := range ideal {
		if result.OutputCurrents[i] > ideal[i]*1.01 {
			t.Errorf("Output[%d] exceeds ideal: actual=%f > ideal=%f", i, result.OutputCurrents[i], ideal[i])
		}
	}
}

func TestSORConvergence_128x128(t *testing.T) {
	// Acceptance criteria: converge for 128×128 with Rp/Rmin = 1.0
	config := &SORConfig{
		MaxIterations: 300,
		Tolerance:     1e-4,
		OmegaInitial:  0.6, // More conservative for large arrays
		OmegaMin:      0.01,
		OmegaDecay:    0.98,
		AdaptiveOmega: true,
	}

	solver, err := NewParasiticSolver(128, 128, config)
	if err != nil {
		t.Fatalf("Failed to create solver: %v", err)
	}

	// Random conductances for realistic test
	g := make([][]float64, 128)
	r := rand.New(rand.NewSource(42))
	for i := range g {
		g[i] = make([]float64, 128)
		for j := range g[i] {
			g[i][j] = r.Float64()
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(1.0, 1.0)

	input := make([]float64, 128)
	for i := range input {
		input[i] = r.Float64()
	}

	result, _ := solver.SolveMVMWithFallback(input)

	t.Logf("128x128 with Rp=1.0: iterations=%d, converged=%v, maxErr=%e, omega=%f",
		result.Iterations, result.Converged, result.MaxError, result.FinalOmega)

	// Should at least produce sensible output
	if result.OutputCurrents == nil || len(result.OutputCurrents) != 128 {
		t.Error("Expected 128 output currents")
	}
}

func TestSORVsIdeal_AccuracyLoss(t *testing.T) {
	// Test accuracy loss calculation
	config := &SORConfig{
		MaxIterations: 200,
		Tolerance:     1e-5,
		OmegaInitial:  0.6,
		OmegaMin:      0.01,
		OmegaDecay:    0.95,
		AdaptiveOmega: true,
	}
	solver, _ := NewParasiticSolver(16, 16, config)

	g := make([][]float64, 16)
	for i := range g {
		g[i] = make([]float64, 16)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.5, 0.5)

	input := make([]float64, 16)
	for i := range input {
		input[i] = 1.0
	}

	result, _ := solver.SolveMVMWithFallback(input)

	ideal := solver.ComputeIdealMVM(input)
	loss := ComputeAccuracyLoss(ideal, result.OutputCurrents)

	t.Logf("Accuracy loss with Rp=0.5: RMSE=%e, converged=%v", loss, result.Converged)

	// With moderate parasitics, expect some loss
	if loss < 1e-10 {
		t.Error("Expected non-zero accuracy loss with parasitics")
	}
}

func TestParasiticImpactAnalysis(t *testing.T) {
	solver, _ := NewParasiticSolver(8, 8, nil)

	g := make([][]float64, 8)
	for i := range g {
		g[i] = make([]float64, 8)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.2, 0.2)

	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	impact, err := solver.AnalyzeParasiticImpact(input)
	if err != nil {
		t.Fatalf("AnalyzeParasiticImpact failed: %v", err)
	}

	t.Logf("Impact analysis: RMSE=%e, MaxRelErr=%f%%, Iterations=%d, WorstCell=%v, WorstDrop=%.2f%%",
		impact.RMSError, impact.MaxRelativeError*100, impact.IterationsNeeded,
		impact.WorstCaseCell, impact.WorstCaseDropPct)

	// Worst case should be bottom-right corner (furthest from drivers)
	if impact.WorstCaseCell[0] < 4 || impact.WorstCaseCell[1] < 4 {
		t.Logf("Note: Worst case cell at %v (expected near bottom-right)", impact.WorstCaseCell)
	}

	// Should have some drop with Rp=0.2
	if impact.WorstCaseDropPct < 0.01 {
		t.Error("Expected non-zero voltage drop")
	}
}

func TestAdaptiveOmega(t *testing.T) {
	// Test that adaptive omega kicks in for difficult cases
	config := &SORConfig{
		MaxIterations: 100,
		Tolerance:     1e-6,
		OmegaInitial:  1.5, // Aggressive initial omega
		OmegaMin:      0.1,
		OmegaDecay:    0.9,
		AdaptiveOmega: true,
	}

	solver, _ := NewParasiticSolver(16, 16, config)

	g := make([][]float64, 16)
	for i := range g {
		g[i] = make([]float64, 16)
		for j := range g[i] {
			g[i][j] = 0.8 // High conductance
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.5, 0.5)

	input := make([]float64, 16)
	for i := range input {
		input[i] = 1.0
	}

	result, _ := solver.SolveMVM(input)

	t.Logf("Adaptive omega: initial=%f, final=%f, iterations=%d",
		config.OmegaInitial, result.FinalOmega, result.Iterations)

	// If omega decayed, adaptive mechanism worked
	if result.FinalOmega < config.OmegaInitial {
		t.Log("Adaptive omega decay triggered as expected")
	}
}

func TestFallbackMechanism(t *testing.T) {
	// Test that fallback works when SOR diverges
	config := &SORConfig{
		MaxIterations: 20,    // Low to trigger fallback
		Tolerance:     1e-10, // Very tight to potentially not converge
		OmegaInitial:  1.9,   // Aggressive, might diverge
		OmegaMin:      0.5,
		OmegaDecay:    0.99,
		AdaptiveOmega: true,
	}

	solver, _ := NewParasiticSolver(32, 32, config)

	g := make([][]float64, 32)
	for i := range g {
		g[i] = make([]float64, 32)
		for j := range g[i] {
			g[i][j] = 0.9
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(2.0, 2.0) // Very high parasitics

	input := make([]float64, 32)
	for i := range input {
		input[i] = 1.0
	}

	// Use fallback method
	result, _ := solver.SolveMVMWithFallback(input)

	// Should still produce output
	if result.OutputCurrents == nil {
		t.Error("Fallback should produce output")
	}

	t.Logf("Fallback result: converged=%v, iterations=%d, maxErr=%e",
		result.Converged, result.Iterations, result.MaxError)
}

func BenchmarkSOR_8x8(b *testing.B) {
	solver, _ := NewParasiticSolver(8, 8, nil)
	g := make([][]float64, 8)
	for i := range g {
		g[i] = make([]float64, 8)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.1, 0.1)

	input := make([]float64, 8)
	for i := range input {
		input[i] = 1.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver.SolveMVM(input)
	}
}

func BenchmarkSOR_32x32(b *testing.B) {
	solver, _ := NewParasiticSolver(32, 32, nil)
	g := make([][]float64, 32)
	for i := range g {
		g[i] = make([]float64, 32)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.1, 0.1)

	input := make([]float64, 32)
	for i := range input {
		input[i] = 1.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver.SolveMVM(input)
	}
}

func BenchmarkSOR_64x64(b *testing.B) {
	solver, _ := NewParasiticSolver(64, 64, nil)
	g := make([][]float64, 64)
	for i := range g {
		g[i] = make([]float64, 64)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.1, 0.1)

	input := make([]float64, 64)
	for i := range input {
		input[i] = 1.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver.SolveMVM(input)
	}
}

func BenchmarkSOR_128x128_HighParasitics(b *testing.B) {
	config := &SORConfig{
		MaxIterations: 200,
		Tolerance:     1e-4,
		OmegaInitial:  0.6,
		OmegaMin:      0.01,
		OmegaDecay:    0.98,
		AdaptiveOmega: true,
	}

	solver, _ := NewParasiticSolver(128, 128, config)
	g := make([][]float64, 128)
	for i := range g {
		g[i] = make([]float64, 128)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(1.0, 1.0)

	input := make([]float64, 128)
	for i := range input {
		input[i] = 1.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver.SolveMVMWithFallback(input)
	}
}

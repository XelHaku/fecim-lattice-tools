package crossbar

import (
	"math"
	"math/rand"
	"testing"
)

// TestOptimizedSolverCorrectness verifies the optimized solver produces
// identical results to the original solver.
func TestOptimizedSolverCorrectness(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	rows, cols := 32, 32
	cfg := DefaultSORConfig()
	cfg.MaxIterations = 500
	cfg.Tolerance = 1e-6
	cfg.OmegaInitial = 1.0
	cfg.AdaptiveOmega = false

	origSolver, err := NewParasiticSolver(rows, cols, cfg)
	if err != nil {
		t.Fatalf("NewParasiticSolver: %v", err)
	}

	optSolver, err := NewOptimizedParasiticSolver(rows, cols, cfg)
	if err != nil {
		t.Fatalf("NewOptimizedParasiticSolver: %v", err)
	}

	g := make([][]float64, rows)
	for i := range g {
		g[i] = make([]float64, cols)
		for j := range g[i] {
			g[i][j] = 0.3 + 0.4*rng.Float64()
		}
	}

	origSolver.SetConductances(g)
	optSolver.SetConductances(g)
	origSolver.SetParasitics(0.0005, 0.0005)
	optSolver.SetParasitics(0.0005, 0.0005)

	applied := make([]float64, cols)
	for j := range applied {
		applied[j] = 0.5 + 0.5*rng.Float64()
	}

	origResult, err := origSolver.SolveMVM(applied)
	if err != nil {
		t.Fatalf("Original SolveMVM failed: %v", err)
	}

	optResult, err := optSolver.SolveMVM(applied)
	if err != nil {
		t.Fatalf("Optimized SolveMVM failed: %v", err)
	}

	// Verify same number of iterations
	if origResult.Iterations != optResult.Iterations {
		t.Errorf("Iteration count mismatch: original=%d, optimized=%d",
			origResult.Iterations, optResult.Iterations)
	}

	// Verify output currents match
	tolerance := 1e-10
	for j := 0; j < cols; j++ {
		diff := math.Abs(origResult.OutputCurrents[j] - optResult.OutputCurrents[j])
		if diff > tolerance {
			t.Errorf("Output current mismatch at col %d: original=%v, optimized=%v, diff=%v",
				j, origResult.OutputCurrents[j], optResult.OutputCurrents[j], diff)
		}
	}

	// Verify device voltages match
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			diff := math.Abs(origResult.DeviceVoltages[i][j] - optResult.DeviceVoltages[i][j])
			if diff > tolerance {
				t.Errorf("Device voltage mismatch at (%d,%d): original=%v, optimized=%v",
					i, j, origResult.DeviceVoltages[i][j], optResult.DeviceVoltages[i][j])
			}
		}
	}

	// Test fast path produces same output currents
	fastResult, iters, err := optSolver.SolveMVMFast(applied)
	if err != nil {
		t.Fatalf("SolveMVMFast failed: %v", err)
	}

	if iters != origResult.Iterations {
		t.Errorf("Fast path iteration count mismatch: original=%d, fast=%d",
			origResult.Iterations, iters)
	}

	for j := 0; j < cols; j++ {
		diff := math.Abs(origResult.OutputCurrents[j] - fastResult[j])
		if diff > tolerance {
			t.Errorf("Fast path output mismatch at col %d: original=%v, fast=%v",
				j, origResult.OutputCurrents[j], fastResult[j])
		}
	}
}

// TestOptimizedSolverReusability verifies the solver can be reused multiple times
// with different inputs without accumulating errors.
func TestOptimizedSolverReusability(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	rows, cols := 16, 16
	cfg := DefaultSORConfig()
	cfg.MaxIterations = 200
	cfg.Tolerance = 1e-5
	cfg.AdaptiveOmega = false

	solver, err := NewOptimizedParasiticSolver(rows, cols, cfg)
	if err != nil {
		t.Fatalf("NewOptimizedParasiticSolver: %v", err)
	}

	g := make([][]float64, rows)
	for i := range g {
		g[i] = make([]float64, cols)
		for j := range g[i] {
			g[i][j] = 0.5
		}
	}
	solver.SetConductances(g)
	solver.SetParasitics(0.0001, 0.0001)

	// Run multiple times with same input
	applied := make([]float64, cols)
	for j := range applied {
		applied[j] = 1.0
	}

	var firstResult []float64
	for run := 0; run < 10; run++ {
		result, _, err := solver.SolveMVMFast(applied)
		if err != nil {
			t.Fatalf("Run %d failed: %v", run, err)
		}

		if run == 0 {
			firstResult = make([]float64, len(result))
			copy(firstResult, result)
		} else {
			for j := 0; j < cols; j++ {
				if math.Abs(firstResult[j]-result[j]) > 1e-12 {
					t.Errorf("Run %d: output differs at col %d: first=%v, current=%v",
						run, j, firstResult[j], result[j])
				}
			}
		}
	}

	// Run with different random inputs
	for run := 0; run < 5; run++ {
		for j := range applied {
			applied[j] = 0.2 + 0.8*rng.Float64()
		}
		_, _, err := solver.SolveMVMFast(applied)
		if err != nil {
			t.Fatalf("Random run %d failed: %v", run, err)
		}
	}
}

// TestOptimizedSolverIdealMVM verifies the ideal MVM computation matches original.
func TestOptimizedSolverIdealMVM(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	rows, cols := 8, 8
	cfg := DefaultSORConfig()

	origSolver, _ := NewParasiticSolver(rows, cols, cfg)
	optSolver, _ := NewOptimizedParasiticSolver(rows, cols, cfg)

	g := make([][]float64, rows)
	for i := range g {
		g[i] = make([]float64, cols)
		for j := range g[i] {
			g[i][j] = rng.Float64()
		}
	}

	origSolver.SetConductances(g)
	optSolver.SetConductances(g)

	applied := make([]float64, cols)
	for j := range applied {
		applied[j] = rng.Float64()
	}

	origIdeal := origSolver.ComputeIdealMVM(applied)
	optIdeal := optSolver.ComputeIdealMVM(applied)

	for j := 0; j < cols; j++ {
		if math.Abs(origIdeal[j]-optIdeal[j]) > 1e-15 {
			t.Errorf("Ideal MVM mismatch at col %d: original=%v, optimized=%v",
				j, origIdeal[j], optIdeal[j])
		}
	}
}

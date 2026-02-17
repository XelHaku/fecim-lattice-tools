package crossbar

import (
	"math/rand"
	"testing"
)

// makeTestConductances creates a conductance matrix for benchmarks.
func makeTestConductances(rng *rand.Rand, rows, cols int) [][]float64 {
	g := make([][]float64, rows)
	for i := range g {
		g[i] = make([]float64, cols)
		for j := range g[i] {
			g[i][j] = 0.45 + 0.10*rng.Float64()
		}
	}
	return g
}

// makeAppliedVoltages creates applied voltages for benchmarks.
func makeAppliedVoltages(rng *rand.Rand, cols int) []float64 {
	applied := make([]float64, cols)
	for j := range applied {
		applied[j] = 0.2 + 0.8*rng.Float64()
	}
	return applied
}

func benchmarkArrayMVM(b *testing.B, rows, cols int) {
	rng := rand.New(rand.NewSource(7))
	cfg := &Config{
		Rows:       rows,
		Cols:       cols,
		NoiseLevel: 0,
		ADCBits:    8,
		DACBits:    8,
	}
	arr, err := NewArray(cfg)
	if err != nil {
		b.Fatalf("NewArray: %v", err)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			arr.cells[i][j].Conductance = 0.2 + 0.6*rng.Float64()
		}
	}

	input := make([]float64, cols)
	for j := range input {
		input[j] = 0.1 + 0.8*rng.Float64()
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out, err := arr.MVM(input)
		if err != nil {
			b.Fatalf("MVM: %v", err)
		}
		if len(out) != rows {
			b.Fatalf("unexpected output length: got %d want %d", len(out), rows)
		}
	}
}

func BenchmarkArrayMVM_8x8(b *testing.B) {
	benchmarkArrayMVM(b, 8, 8)
}

func BenchmarkArrayMVM_32x32(b *testing.B) {
	benchmarkArrayMVM(b, 32, 32)
}

func BenchmarkParasiticSolverSolveMVM_64x64(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	rows, cols := 64, 64
	cfg := DefaultSORConfig()
	cfg.MaxIterations = 500
	cfg.Tolerance = 1e-5
	cfg.OmegaInitial = 1.2
	solver, err := NewParasiticSolver(rows, cols, cfg)
	if err != nil {
		b.Fatalf("NewParasiticSolver: %v", err)
	}

	g := makeTestConductances(rng, rows, cols)
	solver.SetConductances(g)
	solver.SetParasitics(0.001, 0.001)
	applied := makeAppliedVoltages(rng, cols)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := solver.SolveMVM(applied)
		if err != nil {
			b.Fatalf("SolveMVM: %v", err)
		}
		if res == nil || len(res.OutputCurrents) == 0 {
			b.Fatalf("unexpected empty result")
		}
	}
}

// BenchmarkOptimizedParasiticSolverSolveMVM_64x64 benchmarks the optimized solver.
func BenchmarkOptimizedParasiticSolverSolveMVM_64x64(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	rows, cols := 64, 64
	cfg := DefaultSORConfig()
	cfg.MaxIterations = 500
	cfg.Tolerance = 1e-5
	cfg.OmegaInitial = 1.2
	solver, err := NewOptimizedParasiticSolver(rows, cols, cfg)
	if err != nil {
		b.Fatalf("NewOptimizedParasiticSolver: %v", err)
	}

	g := makeTestConductances(rng, rows, cols)
	solver.SetConductances(g)
	solver.SetParasitics(0.001, 0.001)
	applied := makeAppliedVoltages(rng, cols)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := solver.SolveMVM(applied)
		if err != nil {
			b.Fatalf("SolveMVM: %v", err)
		}
		if res == nil || len(res.OutputCurrents) == 0 {
			b.Fatalf("unexpected empty result")
		}
	}
}

// BenchmarkOptimizedParasiticSolverSolveMVMFast_64x64 benchmarks the fast path.
func BenchmarkOptimizedParasiticSolverSolveMVMFast_64x64(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	rows, cols := 64, 64
	cfg := DefaultSORConfig()
	cfg.MaxIterations = 500
	cfg.Tolerance = 1e-5
	cfg.OmegaInitial = 1.2
	solver, err := NewOptimizedParasiticSolver(rows, cols, cfg)
	if err != nil {
		b.Fatalf("NewOptimizedParasiticSolver: %v", err)
	}

	g := makeTestConductances(rng, rows, cols)
	solver.SetConductances(g)
	solver.SetParasitics(0.001, 0.001)
	applied := makeAppliedVoltages(rng, cols)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, _, err := solver.SolveMVMFast(applied)
		if err != nil {
			b.Fatalf("SolveMVMFast: %v", err)
		}
		if len(res) == 0 {
			b.Fatalf("unexpected empty result")
		}
	}
}

// BenchmarkParasiticSolverSolveMVM_128x128 benchmarks larger array sizes.
// Uses very low parasitic resistance to ensure convergence.
func BenchmarkParasiticSolverSolveMVM_128x128(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	rows, cols := 128, 128
	cfg := DefaultSORConfig()
	cfg.MaxIterations = 200
	cfg.Tolerance = 1e-4
	cfg.OmegaInitial = 1.0
	cfg.AdaptiveOmega = false // Disable for stability in benchmarks
	solver, err := NewParasiticSolver(rows, cols, cfg)
	if err != nil {
		b.Fatalf("NewParasiticSolver: %v", err)
	}

	g := makeTestConductances(rng, rows, cols)
	solver.SetConductances(g)
	solver.SetParasitics(0.0001, 0.0001) // Very low parasitic for convergence
	applied := makeAppliedVoltages(rng, cols)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := solver.SolveMVM(applied)
		if err != nil {
			b.Fatalf("SolveMVM: %v", err)
		}
		if res == nil || len(res.OutputCurrents) == 0 {
			b.Fatalf("unexpected empty result")
		}
	}
}

// BenchmarkOptimizedParasiticSolverSolveMVM_128x128 benchmarks optimized larger arrays.
func BenchmarkOptimizedParasiticSolverSolveMVM_128x128(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	rows, cols := 128, 128
	cfg := DefaultSORConfig()
	cfg.MaxIterations = 200
	cfg.Tolerance = 1e-4
	cfg.OmegaInitial = 1.0
	cfg.AdaptiveOmega = false
	solver, err := NewOptimizedParasiticSolver(rows, cols, cfg)
	if err != nil {
		b.Fatalf("NewOptimizedParasiticSolver: %v", err)
	}

	g := makeTestConductances(rng, rows, cols)
	solver.SetConductances(g)
	solver.SetParasitics(0.0001, 0.0001)
	applied := makeAppliedVoltages(rng, cols)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := solver.SolveMVM(applied)
		if err != nil {
			b.Fatalf("SolveMVM: %v", err)
		}
		if res == nil || len(res.OutputCurrents) == 0 {
			b.Fatalf("unexpected empty result")
		}
	}
}

// BenchmarkOptimizedParasiticSolverSolveMVMFast_128x128 benchmarks the fast path for larger arrays.
func BenchmarkOptimizedParasiticSolverSolveMVMFast_128x128(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	rows, cols := 128, 128
	cfg := DefaultSORConfig()
	cfg.MaxIterations = 200
	cfg.Tolerance = 1e-4
	cfg.OmegaInitial = 1.0
	cfg.AdaptiveOmega = false
	solver, err := NewOptimizedParasiticSolver(rows, cols, cfg)
	if err != nil {
		b.Fatalf("NewOptimizedParasiticSolver: %v", err)
	}

	g := makeTestConductances(rng, rows, cols)
	solver.SetConductances(g)
	solver.SetParasitics(0.0001, 0.0001)
	applied := makeAppliedVoltages(rng, cols)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, _, err := solver.SolveMVMFast(applied)
		if err != nil {
			b.Fatalf("SolveMVMFast: %v", err)
		}
		if len(res) == 0 {
			b.Fatalf("unexpected empty result")
		}
	}
}

func BenchmarkDriftSimulatorSimulateTimeStep_128x128(b *testing.B) {
	rows, cols := 128, 128
	sim := NewDriftSimulator(rows, cols, DefaultQuantizationLevels)
	sim.Temperature = 358 // ~85C
	sim.DriftCoeff = 1e-3

	// Make RNG deterministic: the simulator uses math/rand global.
	rand.Seed(1)

	dt := 1000.0 // seconds
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sim.SimulateTimeStep(dt)
	}
}

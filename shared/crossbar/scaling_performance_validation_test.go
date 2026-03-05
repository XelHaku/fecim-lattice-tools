package crossbar

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"testing"
)

func newRandomOptimizedSolverForBench(rng *rand.Rand, n int) (*OptimizedParasiticSolver, []float64) {
	cfg := &SORConfig{
		MaxIterations: 15000,
		Tolerance:     1e-4,
		OmegaInitial:  1.0,
		OmegaMin:      0.02,
		OmegaDecay:    0.95,
		AdaptiveOmega: true,
	}
	s, err := NewOptimizedParasiticSolver(n, n, cfg)
	if err != nil {
		panic(err)
	}
	s.SetParasitics(0.002, 0.002)
	s.SetConductances(randomConductanceMatrix(rng, n, n))
	x := randomVector(rng, n)
	return s, x
}

// M2-SCL-02: Benchmark MVM throughput.
func BenchmarkM2SCL02_MVMThroughput_SolveMVMFast(b *testing.B) {
	sizes := []int{8, 16, 32}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			rng := rand.New(rand.NewSource(int64(1000 + n)))
			s, x := newRandomOptimizedSolverForBench(rng, n)

			for i := 0; i < 5; i++ {
				if _, _, err := s.SolveMVMFast(x); err != nil {
					b.Fatalf("warmup SolveMVMFast: %v", err)
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, _, err := s.SolveMVMFast(x); err != nil {
					b.Fatalf("SolveMVMFast: %v", err)
				}
			}
		})
	}
}

// M2-SCL-03: Verify memory footprint scaling is ~O(N^2).
func TestM2SCL03_MemoryFootprint_ScalesLikeN2(t *testing.T) {
	if testing.Short() {
		t.Skip("scaling memory/solver stress is slow and may exceed iteration budgets; skip in -short")
	}
	// Do not run in parallel with other package tests: this case samples
	// runtime MemStats TotalAlloc and is sensitive to unrelated allocator noise.

	sizes := []int{8, 16, 32}
	const runs = 40

	type sample struct {
		n        int
		bytesRun float64
		allocs   float64
	}

	samples := make([]sample, 0, len(sizes))

	for _, n := range sizes {
		rng := rand.New(rand.NewSource(int64(4242 + n)))
		s, x := newRandomOptimizedSolverForBench(rng, n)

		runtime.GC()
		var ms0 runtime.MemStats
		runtime.ReadMemStats(&ms0)

		allocs := testing.AllocsPerRun(runs, func() {
			if _, _, err := s.SolveMVMFast(x); err != nil {
				t.Fatalf("SolveMVMFast(n=%d): %v", n, err)
			}
		})

		runtime.GC()
		var ms1 runtime.MemStats
		runtime.ReadMemStats(&ms1)

		bytesPerRun := float64(ms1.TotalAlloc-ms0.TotalAlloc) / float64(runs)
		samples = append(samples, sample{n: n, bytesRun: bytesPerRun, allocs: allocs})
	}

	const slack = 6.0
	for i := 1; i < len(samples); i++ {
		prev := samples[i-1]
		curr := samples[i]

		k := float64(curr.n) / float64(prev.n)
		expected := k * k
		measured := curr.bytesRun / math.Max(prev.bytesRun, 1)
		if measured > expected*slack {
			t.Fatalf("allocation bytes scaled worse than O(N^2): n=%d bytes/run=%.1f -> n=%d bytes/run=%.1f, ratio=%.2f, bound=%.2f (expected=%.2f slack=%.1f)",
				prev.n, prev.bytesRun, curr.n, curr.bytesRun, measured, expected*slack, expected, slack)
		}
	}

	for _, s := range samples {
		t.Logf("n=%-3d bytes/run=%.1f allocs/run=%.2f", s.n, s.bytesRun, s.allocs)
	}

	// The race detector instruments memory operations, inflating AllocsPerRun
	// counts (e.g., 1 → 355). Only assert allocation counts without -race.
	if !raceDetectorEnabled && samples[len(samples)-1].allocs > 5 {
		t.Fatalf("expected low allocations on SolveMVMFast, got allocs/run=%.2f at n=%d", samples[len(samples)-1].allocs, samples[len(samples)-1].n)
	}
}

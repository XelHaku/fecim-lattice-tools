# Benchmarks (hot paths)

This repo includes microbenchmarks for the hottest inner loops:

- **Solver step:** `shared/physics.LKSolver.Step`
- **Hysteresis evaluation:** `module1-hysteresis/pkg/ferroelectric.PreisachModel.Update` and `.Polarization`
- **Crossbar update / iterative solve:** `module2-crossbar/pkg/crossbar.ParasiticSolver.SolveMVM`
- **Crossbar update (optimized):** `module2-crossbar/pkg/crossbar.OptimizedParasiticSolver.SolveMVM`
- **Crossbar drift update:** `module2-crossbar/pkg/crossbar.DriftSimulator.SimulateTimeStep`
- **GUI rendering:** `shared/utils.DrawRectFast`, `FillImageFast`, `DrawGridFast`

## Run

```bash
# Convenience wrapper
make bench

# Run all benchmarks, skip unit tests
go test ./... -run ^$ -bench . -benchmem

# Run only the hot-path set
go test ./... -run ^$ -bench "BenchmarkLKSolverStep|BenchmarkPreisachModel|BenchmarkParasiticSolver|BenchmarkOptimizedParasiticSolver|BenchmarkDriftSimulator" -benchmem

# Compare original vs optimized crossbar solver
go test ./module2-crossbar/pkg/crossbar -run ^$ -bench "BenchmarkParasiticSolver|BenchmarkOptimizedParasiticSolver" -benchmem

# GUI rendering benchmarks
go test ./shared/utils -run ^$ -bench "BenchmarkDraw|BenchmarkFill|BenchmarkDigitCanvas" -benchmem
```

## Profiling (pprof)

### CPU profile

```bash
go test ./module2-crossbar/pkg/crossbar \
  -run ^$ -bench BenchmarkOptimizedParasiticSolverSolveMVM_64x64 \
  -benchmem -cpuprofile cpu.out

go tool pprof -http=:0 cpu.out
```

### Memory profile

```bash
go test ./module2-crossbar/pkg/crossbar \
  -run ^$ -bench BenchmarkParasiticSolverSolveMVM_64x64 \
  -benchmem -memprofile mem.out

go tool pprof -http=:0 mem.out
```

### Tips

- Prefer `-run ^$` to avoid unit tests changing cache/branch state.
- Use `-count 5` to see variance.
- For allocation hot spots, `go test -benchmem` + `mem.out` is usually faster to act on than CPU.
- Use `OptimizedParasiticSolver` for production workloads; use `SolveMVMFast()` when only outputs are needed.

---

## Baseline numbers (AMD Ryzen 9 5900X, linux/amd64)

### Physics Solvers

```
BenchmarkLKSolverStep-24                               65 ns/op      0 B/op      0 allocs/op
BenchmarkLKSolverStep_StiffImplicitPath-24             70 ns/op      0 B/op      0 allocs/op
```

### Hysteresis (Preisach Model)

```
BenchmarkPreisachModelUpdate_SineExcitation-24        186 ns/op     27 B/op      1 allocs/op
BenchmarkPreisachModelPolarization_NoHistoryMutation-24 140 ns/op   24 B/op      1 allocs/op
```

### Crossbar Solver: Original vs Optimized

| Benchmark | Time | Memory | Allocs |
|-----------|------|--------|--------|
| **64x64 Array** ||||
| ParasiticSolver.SolveMVM | 11.3 ms | 20.8 MB | 39,262 |
| OptimizedParasiticSolver.SolveMVM | 8.5 ms | 70 KB | 132 |
| OptimizedParasiticSolver.SolveMVMFast | 8.4 ms | 512 B | 1 |
| **128x128 Array** ||||
| ParasiticSolver.SolveMVM | 4.8 ms | 8.5 MB | 8,258 |
| OptimizedParasiticSolver.SolveMVM | 3.6 ms | 270 KB | 260 |
| OptimizedParasiticSolver.SolveMVMFast | 3.5 ms | 1 KB | 1 |

**Improvement Summary (64x64):**
- **25% faster** execution time
- **99.7% less memory** (20.8 MB → 70 KB)
- **99.7% fewer allocations** (39,262 → 132)
- **Fast path:** 99.998% less memory (20.8 MB → 512 B)

### Drift Simulator

```
BenchmarkDriftSimulatorSimulateTimeStep_128x128-24    462 µs/op     48 B/op      5 allocs/op
```

### GUI Rendering: Original vs Fast

| Benchmark | Original | Fast | Speedup |
|-----------|----------|------|---------|
| DrawRect (300×300) | 550 µs | 95 µs | **5.8×** |
| FillImage (500×500) | 3,678 µs | 15 µs | **240×** |
| DrawGradientRect (300×300) | 1,434 µs | 86 µs | **17×** |
| DrawGrid (28×28 cells) | N/A | 22 µs | - |
| DigitCanvas full render | N/A | 45 µs | - |

**Key insight:** The fast drawing functions use direct `img.Pix` slice manipulation instead of `img.Set()`, eliminating interface dispatch and color conversion overhead.

---

## Optimization Notes

### OptimizedParasiticSolver

The original `ParasiticSolver.SolveMVM` allocated 5 work matrices **per iteration** inside the convergence loop:
- `IsumCol`, `IsumRow` (cumulative currents)
- `VdropsCol`, `VdropsRow` (voltage drops)
- `VerrMat` (voltage error)

For a 64×64 array with ~40 iterations, this meant ~5 × 64 × 64 × 8 × 40 ≈ 6.5 MB of allocations per call.

The `OptimizedParasiticSolver` pre-allocates all work buffers in `NewOptimizedParasiticSolver()` and reuses them across calls. The API is identical:

```go
// Drop-in replacement
solver, _ := NewOptimizedParasiticSolver(rows, cols, config)
solver.SetConductances(g)
solver.SetParasitics(rpRow, rpCol)
result, err := solver.SolveMVM(appliedVoltages)

// Or for maximum speed when only outputs are needed:
outputs, iterations, err := solver.SolveMVMFast(appliedVoltages)
```

### Fast Drawing Functions

Located in `shared/utils/drawing_fast.go`:

- `DrawRectFast()` - Direct pixel fill with boundary clipping
- `FillImageFast()` - Full image fill using row-copy optimization
- `DrawHLineFast()` / `DrawVLineFast()` - Fast line drawing
- `DrawGridFast()` - Efficient grid rendering for crossbar displays
- `DrawGradientRectFast()` - Vertical gradient with per-row color calculation
- `BlendPixelFast()` - Alpha blending with early-out for opaque/transparent

All functions take `color.RGBA` directly to avoid color model conversion overhead.

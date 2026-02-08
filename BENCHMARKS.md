# Benchmarks (hot paths)

This repo includes microbenchmarks for the hottest inner loops:

- **Solver step:** `shared/physics.LKSolver.Step`
- **Hysteresis evaluation:** `module1-hysteresis/pkg/ferroelectric.PreisachModel.Update` and `.Polarization`
- **Crossbar update / iterative solve:** `module2-crossbar/pkg/crossbar.ParasiticSolver.SolveMVM`
- **Crossbar drift update:** `module2-crossbar/pkg/crossbar.DriftSimulator.SimulateTimeStep`

## Run

```bash
# Convenience wrapper
make bench

# Run all benchmarks, skip unit tests
go test ./... -run ^$ -bench . -benchmem

# Run only the hot-path set
go test ./... -run ^$ -bench "BenchmarkLKSolverStep|BenchmarkEngineStep|BenchmarkPreisachModel|BenchmarkParasiticSolverSolveMVM|BenchmarkDriftSimulatorSimulateTimeStep" -benchmem

# Use a regex + multiple samples (variance)
BENCH=BenchmarkLKSolverStep BENCH_COUNT=5 make bench
```

## Profiling (pprof)

### CPU profile

```bash
go test ./module2-crossbar/pkg/crossbar \
  -run ^$ -bench BenchmarkParasiticSolverSolveMVM_64x64 \
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
- Use `-count 5` to see variance:
  ```bash
  go test ./... -run ^$ -bench BenchmarkLKSolverStep -benchmem -count 5
  ```
- If you suspect inlining or dead-code elimination, assign results to a package-level sink.
- For allocation hot spots, `go test -benchmem` + `mem.out` is usually faster to act on than CPU.

## Baseline numbers (AMD Ryzen 9 5900X, linux/amd64, Go toolchain as installed)

Captured via:

```bash
go test ./... -run ^$ -bench "BenchmarkLKSolverStep|BenchmarkEngineStep|BenchmarkPreisachModel|BenchmarkParasiticSolverSolveMVM_64x64|BenchmarkDriftSimulatorSimulateTimeStep_128x128" -benchmem
```

```
BenchmarkLKSolverStep-24                              84.90 ns/op     0 B/op      0 allocs/op
BenchmarkLKSolverStep_StiffImplicitPath-24            86.30 ns/op     0 B/op      0 allocs/op

BenchmarkEngineStep-24                              1369 ns/op     529 B/op      9 allocs/op

BenchmarkPreisachModelUpdate_SineExcitation-24        277.4 ns/op     27 B/op      1 allocs/op
BenchmarkPreisachModelPolarization_NoHistoryMutation-24 205.6 ns/op   24 B/op      1 allocs/op

BenchmarkParasiticSolverSolveMVM_64x64-24          19300 µs/op   20875091 B/op  39262 allocs/op
BenchmarkDriftSimulatorSimulateTimeStep_128x128-24    694.9 µs/op     48 B/op      5 allocs/op
```

Notes:

- `ParasiticSolver.SolveMVM` currently allocates several large matrices per call (and per-iteration), so `B/op` and `allocs/op` are expected to be high.
- The drift simulator uses `math/rand`'s global RNG; benchmarks seed it for determinism.

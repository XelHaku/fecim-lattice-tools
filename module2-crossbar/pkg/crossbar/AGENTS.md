<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module2-crossbar/pkg/crossbar

## Purpose

Implements ferroelectric crossbar array simulation with matrix-vector multiplication (MVM), IR drop modeling, sneak-path effects, thermal effects, write disturbance, and device drift. Supports quantization to 30 discrete conductance levels, temperature-dependent behavior, and comprehensive non-ideality modeling. Core simulation engine for analog in-memory computing demonstrations.

## Key Files

| File | Description |
|------|-------------|
| `array.go` | Main array type and MVM implementation (26KB). Manages device conductance, quantization, concurrent access. |
| `nonidealities.go` | Non-ideality models: IR drop, sneak paths, write disturb, access time (27KB). |
| `enhanced.go` | Enhanced MVM with coupled non-ideality (33KB). IR drop adjusts device current; sneak paths couple WL/BL currents. |
| `irdrop.go` | IR drop calculation: WL/BL resistance, parasitic current (9.6KB). |
| `sneakpath.go` | Sneak-path model: unselected cells leak current into bit lines (8.6KB). |
| `drift.go` | Conductance drift over time: stress-dependent and temperature-dependent (15KB). |
| `temperature.go` | Thermal modeling: local heating, temperature-dependent conductance (14KB). |
| `gpu_mvm.go` | GPU acceleration for large arrays (12KB, optional CUDA backend). |

## For AI Agents

### Working In This Directory

**Default Quantization:**

- 30 discrete conductance levels (conference claim; educational simulation)
- Alias to `shared/physics.DefaultLevels`
- Range: GMin to GMax (shared/physics constants)
- Conductance model: Linear/Exponential/Lookup tables

**MVM Core Algorithm:**

1. Apply selected WL voltage to selected rows
2. For each bit line: sum selected cell currents + sneak paths + IR drop
3. Quantize sensed voltage to output conductance range
4. Return conductance matrix-vector result

**Non-idealities Integration:**

- IR drop: Reduces effective voltage on each cell, scales current
- Sneak paths: Unselected cells leak current; scales with unselected cell count and voltage
- Write disturb: Partial writes to unselected cells during row programming
- Drift: Conductance changes over time; stress-accelerated and temperature-dependent
- Thermal: Self-heating updates temperature; affects mobility and conductance

**Stress Testing:**

- 50+ test files covering edge cases: boundary conditions, scaling, concurrent access
- Coupled non-ideality tests: IR drop + sneak paths together
- Statistical validation: Monte Carlo device variation
- GPU stress tests (if CUDA available)

### Testing Requirements

```bash
# Run all crossbar tests
go test ./module2-crossbar/pkg/crossbar -v

# Run MVM core tests
go test ./module2-crossbar/pkg/crossbar -run TestMVM -v

# Run non-ideality coupling tests
go test ./module2-crossbar/pkg/crossbar -run TestCoupledNonidealities -v

# Run IR drop + sneak path interaction
go test ./module2-crossbar/pkg/crossbar -run TestDriftIRDrop -v

# Run boundary and edge case tests
go test ./module2-crossbar/pkg/crossbar -run TestBoundary -v

# Run statistical validation
go test ./module2-crossbar/pkg/crossbar -run TestStatistical -v

# Run concurrent stress test (race detection)
go test ./module2-crossbar/pkg/crossbar -run TestConcurrentStress -race -v

# Run GPU tests (if CUDA available)
go test ./module2-crossbar/pkg/crossbar -run TestGPU -v

# Benchmark MVM
go test ./module2-crossbar/pkg/crossbar -bench BenchmarkMVM -benchmem
```

### Common Patterns

- **Array creation**: `NewArray(rows, cols, noiseLevel, conductanceModel)` with config
- **Quantization**: `QuantizeTo30Levels(float64) int` for 30-state simulation
- **MVM**: `ComputeMVM(inputVector) ([]float64, error)` applies selected WL and reads all BLs
- **Non-ideality control**: Enable/disable via array config or non-ideality struct fields
- **Temperature update**: Computed from power dissipation; affects all subsequent MVMs
- **Drift**: Applied at end of each operation; cumulative over time

## Dependencies

### Internal

- `shared/physics` - Conductance models, quantization, units
- `shared/logging` - Array operation logging
- GPU integration (optional): `shared/gpu` (CUDA bindings, not always available)

### External

- `math` (Go stdlib) - Exp, sqrt, etc.
- `math/rand` (Go stdlib) - Device variation
- `sync` (Go stdlib) - Concurrent array access
- `fmt`, `os` (Go stdlib) - Logging, error handling

<!-- MANUAL: Last edited 2026-02-13. MVM + coupled non-idealities stable. 50+ edge case tests ensure robustness. -->

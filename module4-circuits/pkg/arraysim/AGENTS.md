<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module4-circuits/pkg/arraysim

## Purpose

Implements array simulation specialized for peripheral circuits module. Extends core crossbar simulation with architecture-specific features: tiered device models (Tier A, Tier B), program scheduling, read-margin analysis, transient behavior, SPICE export, and design-space exploration. Provides higher-level abstractions over raw crossbar MVM for circuit applications.

## Key Files

| File | Description |
|------|-------------|
| `types.go` | Array configuration and operation types (6.4KB). Tier selection, device model enums. |
| `tier_a.go` | Simplified single-device-per-cell model (0.9KB). Fast for prototyping. |
| `tier_b.go` | Full multi-transistor circuit model (7.6KB). Realistic for tape-out. |
| `array_ispp.go` | ISPP integration for write scheduling (6.8KB). |
| `program_scheduler.go` | Program order optimization for convergence (4.7KB). |
| `read_margin_analysis.go` | Read-margin characterization and statistics (6KB). |
| `transient.go` | Transient response analysis (3.5KB). |
| `sensechain.go` | Sense-amplifier chain simulation (4.9KB). |
| `spice_export.go` | NetList generation for external simulation (3.4KB). |
| Design space exploration utilities | `design_space_exploration.go`, `mixed_precision_planner.go`, `process_variation_mc.go` |

## For AI Agents

### Working In This Directory

**Tier Selection:**

- **Tier A** (simple): Single device per cell, fast MVM, good for system-level exploration
- **Tier B** (realistic): Full multi-transistor circuit (TFT access, parallel cap, etc.), slower but accurate

**Program Scheduler:**

- Determines order of write operations for optimal convergence
- Can use level-order (ascending/descending), random, or heuristic scheduling
- Impacts program time and number of pulses

**Read-Margin Analysis:**

- Characterizes margin between 0-state and 1-state conductance
- Statistical analysis of worst-case margins given device variation
- Critical for reliability assessment

**SPICE Export:**

- Generates `.sp` netlist for external validation in Cadence, NGSPICE, etc.
- Tier B models map directly to SPICE subcircuits

### Testing Requirements

```bash
# Run all arraysim tests
go test ./module4-circuits/pkg/arraysim -v

# Run Tier A tests
go test ./module4-circuits/pkg/arraysim -run TestTierA -v

# Run Tier B tests
go test ./module4-circuits/pkg/arraysim -run TestTierB -v

# Run ISPP integration tests
go test ./module4-circuits/pkg/arraysim -run TestArrayISPP -v

# Run program scheduler tests
go test ./module4-circuits/pkg/arraysim -run TestProgramScheduler -v

# Run read-margin analysis tests
go test ./module4-circuits/pkg/arraysim -run TestReadMarginAnalysis -v

# Run SPICE export tests
go test ./module4-circuits/pkg/arraysim -run TestSPICEExport -v

# Run design-space exploration tests
go test ./module4-circuits/pkg/arraysim -run TestDesignSpaceExploration -v

# Run Monte Carlo process variation
go test ./module4-circuits/pkg/arraysim -run TestProcessVariationMC -v

# Run mixed-precision planner
go test ./module4-circuits/pkg/arraysim -run TestMixedPrecisionPlanner -v
```

### Common Patterns

- **Array creation**: `NewArray(rows, cols, tierType, config)` with `tierType` enum
- **MVM with non-idealities**: Call underlying crossbar; Tier B applies post-processing
- **Read margin**: `ComputeReadMargin()` returns worst-case and statistical margins
- **ISPP schedule**: `ScheduleWrites(targets)` returns ordered list of (row, col, level) tuples
- **Transient analysis**: `ComputeTransient(inputVector, timePoints)` returns V(t)
- **SPICE export**: `ExportSPICE(filename, tierModel)` writes netlist

## Dependencies

### Internal

- `module2-crossbar/pkg/crossbar` - Core MVM, IR drop, sneak paths
- `shared/physics` - Material models, conductance, units
- `shared/peripherals` - DAC, ADC, TIA, sense-amplifier models
- `shared/logging` - Operation logging

### External

- `math` (Go stdlib) - Numerical operations
- `fmt`, `os`, `io` (Go stdlib) - File I/O, logging

<!-- MANUAL: Last edited 2026-02-13. Tier A/B abstraction allows fast vs. realistic tradeoff. -->

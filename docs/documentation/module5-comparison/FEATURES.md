<!-- Category: Features | Module: module5-comparison | Reading time: ~3 min -->
# Module 5 Features: Architecture Comparison

> Feature inventory for the technology comparison module.

---

## Core Features

| Feature | Status |
|---------|--------|
| Architecture profiles (CPU, GPU, FeCIM) | IMPLEMENTED |
| Workload-based comparison engine | IMPLEMENTED |
| Energy per inference calculation | IMPLEMENTED |
| Throughput comparison (TOPS) | IMPLEMENTED |
| Efficiency comparison (TOPS/W) | IMPLEMENTED |
| Latency comparison | IMPLEMENTED |
| Chart rendering (bar charts, tables) | IMPLEMENTED |
| Summary advantage display | IMPLEMENTED |

---

## ROI Calculator

| Feature | Status |
|---------|--------|
| Configurable MACs per inference | IMPLEMENTED |
| Configurable energy per MAC (fJ) | IMPLEMENTED |
| Configurable inference rate | IMPLEMENTED |
| Monthly cost projection | IMPLEMENTED |
| Annual savings projection | IMPLEMENTED |
| Server scale multiplier | IMPLEMENTED |
| Electricity price input | IMPLEMENTED |

---

## GUI

| Feature | Status |
|---------|--------|
| Workload selection | IMPLEMENTED |
| Architecture toggle | IMPLEMENTED |
| Side-by-side metric display | IMPLEMENTED |
| Chart visualization | IMPLEMENTED |

---

## What This Module Does NOT Do

- Does not include real hardware measurements.
- Does not model SRAM, ReRAM, or MRAM (those comparisons would
  belong in a future module or extension).
- Does not include system-level overhead (memory controllers, host
  CPU, data preprocessing).
- Does not provide financial advice -- ROI numbers are model outputs.

---

## Extension Points

- Add new workload definitions or benchmark profiles.
- Extend architecture parameters with additional fields (e.g., area,
  process node).
- Add new chart types or visual summaries.

---

## Where It Lives in Code

| Path | Purpose |
|------|---------|
| `module5-comparison/pkg/comparison/architecture.go` | Architecture profiles |
| `module5-comparison/pkg/comparison/render.go` | Rendering |
| `module5-comparison/pkg/gui/app.go` | GUI entry point |
| `module5-comparison/pkg/gui/widgets.go` | ROI calculator widgets |

---

## Status Legend

- **IMPLEMENTED**: Code exists, tests pass, feature is functional

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*

# 1T1R Architecture Review - Dr. Shin Checkpoint

## Overview

This document addresses the Phase 2 requirement for Dr. Shin's review of 1T1R sneak path considerations in the FeCIM Lattice Generator.

## Architecture Comparison

### Passive Crossbar
```
     BL[0]    BL[1]
       │        │
WL[0]──■────────■──
       │        │
WL[1]──■────────■──
       │        │
```
- **Ports**: WL[], BL[], VDD, VSS
- **Pros**: Simple, dense (0.46μm cell pitch)
- **Cons**: Sneak path currents through unselected cells

### 1T1R Crossbar
```
     BL[0]    BL[1]    SL[0]    SL[1]
       │        │        │        │
WL[0]──┼────────┼────────┼────────┼──
       │        │        │        │
WL[1]──┼────────┼────────┼────────┼──
       │        │        │        │
```
- **Ports**: WL[], BL[], SL[], VDD, VSS
- **Pros**: Sneak path mitigation via select transistor
- **Cons**: Larger cell (0.92μm pitch), more routing

## Sneak Path Analysis

### Problem in Passive Arrays

In a passive crossbar, when reading cell (0,0):
1. Apply voltage to WL[0], ground BL[0]
2. Sneak current flows: WL[0] → Cell(0,1) → BL[1] → Cell(1,1) → WL[1] → ...
3. This parasitic current corrupts the read signal

**Impact**: Read accuracy degrades as array size increases. For N×N arrays, sneak path error ∝ N².

### 1T1R Solution

In 1T1R architecture:
1. Each cell has a select transistor controlled by WL
2. When WL is low, transistor is OFF → cell isolated
3. SL (Source Line) provides controlled current path only when transistor is ON

**Key Design Decision**: SL is per-column (shared by cells in same column)
- Rationale: During read, only one row is active (WL high)
- All other rows have transistors OFF, blocking sneak paths
- Per-column SL simplifies routing while maintaining isolation

## Implementation Details

### Verilog Output

**Passive**:
```verilog
module fecim_crossbar (
    input  wire [N-1:0] WL,
    inout  wire [M-1:0] BL,
    input  wire VDD,
    input  wire VSS
);
```

**1T1R**:
```verilog
module fecim_crossbar (
    input  wire [N-1:0] WL,  // Gate of select transistor
    inout  wire [M-1:0] BL,  // Drain of select transistor
    input  wire [M-1:0] SL,  // Source of FeFET (through select)
    input  wire VDD,
    input  wire VSS
);
```

### DEF Output

1T1R DEF includes:
- SL[M-1:0] pins at bottom edge (opposite from BL at top)
- SL nets connecting all cells in each column
- Larger cell placement (0.92μm vs 0.46μm pitch)

## Recommendations for Large Arrays

| Array Size | Recommended Architecture |
|------------|-------------------------|
| ≤ 32×32    | Passive (simpler, denser) |
| 64×64      | Either (application dependent) |
| ≥ 128×128  | 1T1R (sneak path critical) |

## Validation

Both architectures validated with Yosys:
- ✓ Passive: 16 cells (4×4), 10 wire bits
- ✓ 1T1R: 4 cells (2×2), 8 wire bits including SL

## Files Generated

| Architecture | Verilog | DEF | LEF Cell |
|-------------|---------|-----|----------|
| Passive | lattice.v | placement.def | fecim_bit.stub.lef |
| 1T1R | lattice_1t1r.v | placement_1t1r.def | fecim_1t1r.stub.lef |

## Sign-off

This implementation correctly handles:
- [x] Conditional SL[] port generation for 1T1R
- [x] Cell-type selection based on architecture
- [x] Per-column SL routing in DEF
- [x] Larger cell pitch for 1T1R (transistor overhead)

---
*Dr. Shin Review Checkpoint - Phase 2 Architecture Configuration*
*Date: 2026-01-23*

# Autopilot Specification: CIM Deep Research Documentation

## Executive Summary

Comprehensive research on Compute-in-Memory (CIM) technology fundamentals, creating documentation that enables building accurate CIM simulators. Focus on physics, mathematics, and practical implementation.

## Status: ✅ COMPLETE

**Completed:** 2026-01-25

## Deliverables

### Documentation Created

| File | Purpose | Status |
|------|---------|--------|
| `docs/cim-physics.md` | CIM fundamentals - von Neumann bottleneck, Ohm's law, Kirchhoff's law, energy analysis | ✅ Complete |
| `docs/cim-mathematics.md` | Full equation derivations - MVM, IR drop, sneak paths, quantization, noise models | ✅ Complete |
| `docs/cim-devices.md` | Device technologies - FeFET vs ReRAM vs PCM vs MRAM, 30-level mechanism, Preisach model | ✅ Complete |
| `docs/cim-simulation.md` | Simulation tools - CrossSim, NeuroSim, our implementation comparison, validation methodology | ✅ Complete |
| `docs/cim-equations.md` | Quick reference - all key equations in one place with variable definitions and typical values | ✅ Complete |
| `papers/cim/README.md` | Paper summary - 300+ research sources organized by category | ✅ Complete |

### Research Questions Answered

#### 1. Fundamentals ✅
- [x] What exactly happens physically during in-memory computation?
- [x] How does Ohm's Law enable multiplication? (V × G = I)
- [x] How does Kirchhoff's Current Law enable addition? (ΣI)
- [x] Why is this faster/more efficient than von Neumann?

#### 2. Crossbar Physics ✅
- [x] How does current flow through a crossbar array?
- [x] What determines the voltage at each node?
- [x] How do parasitic resistances affect computation?
- [x] What is the complete circuit model for an N×M array?

#### 3. Mathematics ✅
- [x] Derive the matrix-vector multiplication equation for crossbar
- [x] What is the accuracy loss formula from quantization?
- [x] How to calculate IR drop mathematically?
- [x] How to model sneak path currents?
- [x] What is the noise model for analog computation?

#### 4. Device Physics (FeFET/ReRAM/PCM) ✅
- [x] How does conductance map to stored weight?
- [x] What causes device-to-device variation?
- [x] What causes read noise?
- [x] What causes drift over time?
- [x] How many bits per cell is achievable? Why?

#### 5. System Architecture ✅
- [x] How are large matrices tiled across multiple arrays?
- [x] How do DAC/ADC affect precision?
- [x] What is the energy breakdown (array vs peripherals)?
- [x] How does batch size affect efficiency?

#### 6. Accuracy Analysis ✅
- [x] How does quantization affect neural network accuracy?
- [x] How do non-idealities combine to affect accuracy?
- [x] What is the theoretical maximum accuracy for N-bit weights?
- [x] How does CIM accuracy compare to digital at same bit-width?

## Key Findings

### CIM Physics
- **Multiplication**: Ohm's Law (I = G × V) at each crosspoint
- **Addition**: Kirchhoff's Current Law (automatic current summation)
- **Parallelism**: N×M MACs in single timestep
- **Energy**: 220-1000× advantage over von Neumann (DRAM 200pJ vs MAC 0.25pJ)

### 30-Level Quantization (Dr. Tour)
- 30 states = 4.9 bits/cell
- Intermediate polarization states in HZO superlattice
- Each domain switches independently (stochastic)
- 87% MNIST accuracy achieved (vs 88% theoretical max)

### Non-Ideality Impact
| Array Size | IR Drop | Accuracy Loss |
|------------|---------|---------------|
| 64×64 | 5% | 0.3% |
| 128×128 | 12% | 1.2% |
| 256×256 | 22% | 3.5% |

### Key Parameters (from CLAUDE.md)
- Pr: 15-34 µC/cm² (Nature Commun. 2025)
- Ec: 1.0-1.5 MV/cm (Nature Commun. 2025)
- Endurance: 10⁹ demonstrated, 10¹² target
- FeFET drift: ν = 0.001 (50× better than RRAM)

## Codebase Integration

All documentation references verified code locations:
- `module2-crossbar/pkg/crossbar/array.go` - MVM implementation
- `module2-crossbar/pkg/crossbar/nonidealities.go` - IR drop, sneak paths
- `module2-crossbar/pkg/crossbar/irdrop.go` - Iterative IR drop solver
- `module2-crossbar/pkg/crossbar/sneakpath.go` - Sneak path analyzer
- `module2-crossbar/pkg/crossbar/drift.go` - Drift simulation
- `module1-hysteresis/pkg/ferroelectric/preisach.go` - Preisach model

## Identified Gaps for Future Work

1. **PyTorch integration** - CrossSim has it, we don't
2. **GPU acceleration** - Would speed up large array simulations
3. **Hardware-aware training** - Need noise injection during training
4. **TensorRT integration** - For production deployment

## Sources

300+ research papers cataloged in `papers/cim/README.md` covering:
- Foundational physics (PMC, IEEE, Nature)
- Mathematics & modeling (arXiv, IEEE TCAD)
- Device-specific (Nature Commun, J. Appl. Phys.)
- Simulation tools (CrossSim, NeuroSim)

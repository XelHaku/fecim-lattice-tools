<!-- Category: Features | Module: module2-crossbar | Reading time: ~6 min -->
# Module 2 Features: Crossbar Array Simulator

> Feature reference for the crossbar module -- what the GUI shows,
> what non-idealities are modeled, and how to configure the simulation.

---

## Feature Reference Table

| Feature | GUI Location | What It Shows | Physics Model |
|---------|-------------|---------------|---------------|
| Conductance Heatmap | Main tab | Cell-by-cell conductance values (color-coded) | 30-level quantization |
| IR Drop Map | IR Drop tab | Voltage distribution across array | Cumulative R*I model |
| Sneak Path Visualization | Sneak tab | Parasitic current flow through unselected cells | 3-cell path enumeration |
| MVM Computation | Action panel | Matrix-vector multiply result with error metrics | Ideal + non-idealities |
| Error Dashboard | Results panel | RMSE, max error, accuracy loss estimates | Comparison: ideal vs actual |
| Energy Report | Results panel | Per-operation energy breakdown (array, ADC, DAC) | Literature-based model |
| Architecture Selector | Controls panel | Switch between 0T1R and 1T1R | Sneak isolation factor |
| Temperature Control | Controls panel | Set operating temperature (77K to 400K) | TCR + window scaling |

---

## Non-Ideality Controls

| Effect | Toggle | Key Parameter | Default |
|--------|--------|--------------|---------|
| IR Drop | EnableIRDrop | R_wordline, R_bitline | 2.5 ohm/pitch |
| Sneak Paths | EnableSneakPaths | Architecture (0T1R/1T1R) | 0T1R |
| Device Variation | EnableVariation | sigma | 5% |
| Drift | EnableDrift | nu, time | 0.001, 1 year |
| Temperature | Temperature slider | T (K) | 300 K |

Each effect can be toggled independently. When all are enabled, the combined
impact chain is:

```
Ideal output
  + ADC/DAC quantization   (~1-2% RMSE)
  + IR drop                 (~2-3% RMSE)
  + Device variation        (~1-2% RMSE)
  + Sneak paths (0T1R)      (~5-20% RMSE)
  = Actual output
```

Switching to 1T1R reduces sneak path impact to <0.1%, bringing total RMSE
to 1-2%.

---

## Conductance Models

| Model | When to Use | How to Select |
|-------|-------------|---------------|
| Linear | Quick prototyping, education | Default |
| Exponential | Realistic FeFET simulation | Config.ConductanceModel |
| Lookup Table | Calibrated from measurements | Load via ConductanceTable |

---

## Architectures

| Architecture | Cell Structure | Sneak Isolation | Density | Use Case |
|--------------|---------------|----------------|---------|----------|
| 0T1R (Passive) | Resistor only | None (full sneak) | Very high | Research, small arrays |
| 1T1R | 1 transistor + resistor | 1000x | Moderate | Production systems |

---

## Array Operations

| Operation | What It Does | Access |
|-----------|-------------|--------|
| MVM | Matrix-vector multiply (inference) | Run button / API |
| VMM | Vector-matrix multiply (transpose) | API |
| Write-Verify | Program cell to target level iteratively | API |
| Differential MVM | Signed weights via two arrays (I+ - I-) | API |
| Accuracy Degradation | Stepwise breakdown of error sources | Analysis button |

---

## Export Formats

| Format | Contents | Use Case |
|--------|----------|----------|
| JSON | Array stats, MVM results, energy metrics | Reporting, archiving |
| CSV | Weight matrix with conductances | Spreadsheet analysis |
| NumPy (.npy) | Multi-dimensional weight arrays | Python ML integration |
| SPICE | Netlist with resistance values | Circuit simulation |

---

## Running Module 2

```bash
# GUI (default)
./fecim-lattice-tools crossbar

# With specific array size
./fecim-lattice-tools crossbar --rows 64 --cols 64

# 1T1R architecture
./fecim-lattice-tools crossbar --architecture 1T1R
```

---

## Integration with Other Modules

| Source Module | What It Provides |
|--------------|-----------------|
| Module 1 (Hysteresis) | P-E curves drive conductance values |
| Module 3 (MNIST) | Crossbar executes MVM in network layers |
| Module 4 (Circuits) | ADC/DAC bit-widths, TIA configuration |

---

## Known Limitations

1. No transistor-level simulation (behavioral model only)
2. Simplified wire models (no parasitic capacitance)
3. Default parameters are educational baselines, not production-calibrated
4. Drift coefficients estimated from retention data, not directly measured
5. Full sneak path calculation becomes slow above 32x32 (simplified model used)
6. Not thread-safe for concurrent operations on the same array

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*

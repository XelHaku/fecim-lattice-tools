# Module 2: Crossbar - Features

Matrix-Vector Multiply (MVM) simulator with non-idealities and visualization.

---

## Features

- **Analog MVM/VMM** - Kirchhoff-law current summation (I = G x V)
- **30-Level Quantization** - Demo baseline (configurable)
- **Conductance Models** - Linear, exponential, and lookup-table mapping
- **Non-Idealities** - IR drop, sneak paths, drift, RC delay, process variation, endurance, half-select disturb
- **GPU Acceleration (Optional)** - Compute-shader MVM with CPU fallback
- **Heatmaps & Tabs** - Conductance, IR drop, sneak paths, drift
- **External Tool Checks** - CrossSim and BadCrossbar install/status validation
- **Weight I/O** - Save/load weight matrices and stats

---

## GUI Tabs

| Tab | Focus |
|---|---|
| **Conductance** | Ideal MVM and heatmap visualization |
| **IR Drop** | Line resistance + voltage drop analysis |
| **Sneak Paths** | Parasitic current analysis |
| **Drift** | Conductance drift over time |

---

## Default Parameters (From Code)

| Parameter | Default | Notes |
|---|---:|---|
| Levels | 30 | Demo baseline (configurable) |
| Gmin / Gmax | 10 uS / 100 uS | Conductance range |
| Wire R | 2.5 Ohm/cell | `DefaultWireParams()` |
| Wire C | 0.2 fF/cell | `DefaultWireParams()` |
| ADC / DAC | 6 bits / 8 bits | GUI defaults |
| Noise | 1% | GUI default |

---

## Notes

- Non-idealities are configurable; several are disabled by default for performance.
- GPU acceleration is optional and auto-falls back to CPU when unavailable.
- Defaults are simulation assumptions; cite before external use (DOI: (add)).
- External tool checks detect installs only; they do not run external simulators.

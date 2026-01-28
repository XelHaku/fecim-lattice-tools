# Module 2: Crossbar - Features

## Features

- **Analog Matrix-Vector Multiply (MVM)** — Physics-based I = G × V computation
- **30-Level Quantization** — Discrete conductance states [0/29 ... 29/29]
- **Non-Ideality Simulation** — IR drop, sneak paths, drift, temperature, device variation
- **Architecture Comparison** — 0T1R (passive) vs 1T1R (gated)
- **Real-Time Heatmaps** — Conductance, IR drop, sneak paths, drift visualization
- **GPU Acceleration** — Optional Vulkan compute shaders
- **Neural Network Integration** — Multi-layer networks with hardware-aware training

## Physics Models

| Non-Ideality | Model | Impact |
|--------------|-------|--------|
| **IR Drop** | WL/BL resistance (2.5Ω/cell @ 45nm) | 10-20% voltage loss in large arrays |
| **Sneak Paths** | 3-cell parasitic loops | 5-20% error (0T1R), ~0.001% (1T1R) |
| **Drift** | Power-law + logarithmic + Arrhenius | <0.5 level over 10 years |
| **Temperature** | Arrhenius activation (4K-500K) | Cryogenic: 1.5× window |
| **Variation** | Gaussian + spatial gradients | 2% device-to-device |

## Key Parameters

| Parameter | Value | Notes |
|-----------|-------|-------|
| Levels | 30 | Quantized analog states |
| Gmin | 10 µS | OFF-state conductance |
| Gmax | 100 µS | ON-state conductance |
| R_wire | 2.5 Ω/cell | Word/bit line resistance (45nm) |
| Drift coeff | 0.0005-0.001 | Literature vs assumed |

## Conductance Models

1. **Linear** — G = Gmin + norm × (Gmax - Gmin)
2. **Exponential** — G = Gmin × exp(ln(Gmax/Gmin) × norm)
3. **Lookup** — Calibration table from measurements

## Architecture Comparison

| Metric | 0T1R (Passive) | 1T1R (Gated) |
|--------|----------------|--------------|
| Density | 4F² | 8-12F² |
| Sneak Error | 5-20% | ~0.001% |
| Complexity | Lowest | Higher |

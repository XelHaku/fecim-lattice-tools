<!-- Category: Features | Module: module1-hysteresis | Reading time: ~6 min -->
# Module 1 Features: Hysteresis Simulator

> Feature reference for the ferroelectric hysteresis module --
> what each GUI element shows and which physics model drives it.

---

## Feature Reference Table

| Feature | GUI Location | What It Shows | Physics Model |
|---------|-------------|---------------|---------------|
| P-E Loop | Hysteresis tab (main plot) | Real-time polarization vs electric field curve | Preisach (default) or L-K |
| Discrete Levels | Level indicator bar | Current polarization quantized to 0-29 | Linear discretization of P/Ps |
| ISPP Write | Write mode (automatic) | Pulse-by-pulse convergence to target level | WriteController (binary search) |
| Minor Loops | Enabled when reversing field mid-sweep | Inner loops from partial traversal | Preisach turning-point stack |
| Material Selector | Dropdown in controls panel | Switch between HZO, Superlattice, AlScN, etc. | Material parameter set swap |
| Temperature Slider | Controls panel | Adjust Ec(T), Ps(T) in real time | Linear scaling model |
| Waveform Modes | Mode selector (Sine/Triangle/Square/Manual/Random/Write) | Different E-field drive patterns | Each mode drives physics engine differently |
| WRITE/READ Indicator | Status bar | Shows whether current |E| exceeds Ec | Threshold comparison |
| PUND Panel | PUND tab | Positive-Up Negative-Down pulse sequence results | RunPUNDSimulation() |
| FORC Density | FORC tab | First-Order Reversal Curve density raster | NewFORCDensityRaster() |
| Wake-Up | Integrated into write cycling | Gradual Pr increase over initial cycles | WakeUpPolarization() |

---

## Waveform Modes

| Mode | What It Does | Physics Demonstrated |
|------|-------------|---------------------|
| Manual | Direct E-field control via slider | Explore hysteresis freely |
| Sine | Smooth oscillation, full loop traversal | Complete major loop cycling |
| Triangle | Linear ramps up and down | See switching thresholds (Ec) clearly |
| Square | Snap between +/- extremes | Fast switching between saturated states |
| Random Walk | Random steps across field range | Multi-level storage demonstration |
| Write/Read | ISPP controller sequences through targets | Full memory programming cycle |

---

## Physics Engine Options

| Engine | Type | Best For |
|--------|------|----------|
| Preisach (default) | Phenomenological, hysteron-based | General hysteresis, analog states, minor loops |
| Landau-Khalatnikov | Thermodynamic, mean-field | Phase transitions, temperature effects |

Switch engines in the GUI or via CLI flag `--engine lk`.

---

## Programming Features (ISPP)

| Feature | Status | Notes |
|---------|--------|-------|
| Binary search convergence | Implemented | Adaptive [V_min, V_max] bounds |
| Guard-band correction | Implemented | Max 2 fine-tuning pulses for boundary states |
| Overshoot detection and recovery | Implemented | Minimal bounds widening, not full reset |
| ACCEPT +/-1 after 8+ overshoots | Implemented | Handles physics-limited convergence |
| Overshoot limit (30) triggers success | Implemented | Material cannot hold target; not a failure |
| Per-target verification logging | Implemented | Metrics surfaced in UI log |

---

## Material Library

| Material | Pr (uC/cm^2) | Ec (MV/cm) | Tau (ns) | Tc (C) | Analog States |
|----------|-------------|-----------|---------|--------|---------------|
| Standard HZO | 25 | 1.2 | 1 | 450 | ~30 |
| FeCIM HZO | 30 | 1.0 | 10 | 450 | 30 |
| Superlattice | 50 | 0.85 | 0.36 | 500 | 30+ |
| Cryogenic HZO | 75 | 1.5 | 1 | 450 | 30 |
| AlScN | 120 | 5.0 | 10 | 1000 | 8-16 |
| HZO-32 | 20 | 1.0 | 10 | 450 | 32 |
| FTJ-140 | 25 | 1.2 | 20 | 450 | 140 |

See PHYSICS.md for full parameter tables.

---

## Visualization Modes

| Mode | GUI | TUI | Headless |
|------|-----|-----|----------|
| Real-time P-E plot | Yes | Yes (ASCII) | No |
| Interactive controls | Yes | Yes | No |
| Data export (JSON/CSV) | Yes | Yes | Yes |
| Level bar indicator | Yes | No | No |

---

## Running Module 1

```bash
# GUI (default)
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools hysteresis

# TUI (terminal)
./fecim-lattice-tools hysteresis --tui

# Headless (batch/test)
./fecim-lattice-tools hysteresis --headless

# Specific engine
./fecim-lattice-tools --mode hysteresis --engine lk
```

---

## Integration with Other Modules

| Target Module | What Module 1 Provides |
|---------------|----------------------|
| Module 2 (Crossbar) | Conductance-polarization mapping: level --> G |
| Module 3 (MNIST) | Discrete levels for weight quantization |
| Module 4 (Circuits) | Ec, Pr for peripheral voltage range design |

---

## Known Limitations

1. L-K engine is educational, not device-calibrated
2. Wake-up and fatigue are UI-level models, not physics-accurate
3. Switching time (tau) is defined but not used in real-time visualization
4. FORC calibration not used; tanh Everett approximation instead
5. Material presets include illustrative values; cite before external use

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*

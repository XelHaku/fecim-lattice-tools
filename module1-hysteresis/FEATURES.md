# Module 1: Hysteresis - Features

P-E Curve Simulator for Ferroelectric Memory

---

## Simulation Method

**Core Model: Mayergoyz Classical Preisach**

The simulator uses a unified Preisach model based on Mayergoyz's mathematical framework (1991):

- **Hysteron Grid** - 50×50 to 100×100 bistable switching elements
- **Distribution Function** - Bivariate Gaussian or Lorentzian weighting
- **Hysteron Integration** - Each hysteron switches UP at field α, DOWN at field β; polarization is the weighted sum of all hysteron states

**Physical Effects (Always Active)**

All effects are combined simultaneously in every calculation:

| Effect | Implementation |
|--------|----------------|
| **Temperature** | Modifies Ec via Curie-Weiss: `Ec(T) = Ec0 × (1 - T/Tc)^β` |
| **Fatigue** | Degradation factor: `P × (1 - fatigueRate × cycles)` |
| **Wake-up** | Modifies distribution weights during initialization |
| **Substrate Strain** | Shifts Ec via electrostrictive coupling |
| **KAI Dynamics** | Time-resolved domain growth: `P(t) = Ps × (1 - exp(-(t/τ)^n))` |

---

## Features

- **Interactive P-E Loop** - Real-time hysteresis curves with animated polarization
- **Multi-Level Memory Demo** - Program/read discrete states with ISPP verification
- **8+ Material Library** - Each material defines Ps, Ec, Pr, τ, and level count
- **Temperature Control** - 4K to 723K range, reinitializes hysteron grid
- **Waveform Modes** - Manual, sine, triangle, write/read demo, time-resolved
- **Calibration System** - Temperature-aware multi-level calibration
- **Multiple Run Modes** - Fyne GUI, TUI, headless ASCII, Vulkan graphics

---

## Materials

| Material | Levels | Notes |
|----------|--------|-------|
| HZO (Si-doped) | 30 | Baseline |
| FeCIM HZO | 30 | Dr. Tour specs |
| Literature Superlattice | 64 | Cheema 2020 |
| Cryogenic HZO | 30 | 75 µC/cm² at 4K |
| HZO Standard 32 | 32 | Oh 2017 |
| HZO FTJ 140 | 140 | Song 2024 |
| AlScN | 8-16 | High Pr (120 µC/cm²) |

---

## GUI Components

- **P-E Hysteresis Plot** - Real-time polarization vs field
- **Level Indicator** - Visual gauge for discrete levels
- **Phase Indicator** - State machine: RESET → SETTLE → WRITE → READ → VERIFY
- **Material Picker** - Searchable list with property tables
- **Stability Indicator** - Color-coded level stability warnings

---

## Export

- JSON export with metadata (material, temperature, parameters)
- CSV export for data analysis
- Debug log export (cycle-by-cycle data, energy tracking)

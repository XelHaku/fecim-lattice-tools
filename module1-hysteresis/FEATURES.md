# Module 1: Hysteresis - Features

P-E curve simulator for ferroelectric memory physics.

---

## Physics Engines

- **Preisach (quasi-static)** - Mayergoyz-based hysteresis with memory and discrete states.
- **Landau-Khalatnikov (dynamic)** - Time-resolved switching engine for educational visualization.

> Note: The Landau engine is intended for interactive learning, not calibrated device modeling.

---

## Features

- **Interactive P-E Loop** - Real-time hysteresis curve with polarization animation
- **Discrete Level Programming** - Write/Read/Verify state machine with ISPP-style calibration
- **Material Library** - HZO baseline, FeCIM baseline, literature superlattice, cryogenic HZO, 32-level HZO, 140-level FTJ, AlScN
- **Temperature Control** - 200-700 K slider with temperature-aware calibration cache
- **Waveform Modes** - Manual, sine, triangle, write/read demo, time-resolved switching
- **Multi-Mode UI** - Fyne GUI, TUI, headless ASCII, Vulkan renderer

---

## Materials (From `shared/physics`)

| Material | Levels | Notes |
|---|---:|---|
| HZO (Si-doped) | 30 | Baseline demo material |
| FeCIM HZO | 30 | Conference-claim baseline (pending peer review) |
| Literature Superlattice | 64 | Academic best-case preset |
| Cryogenic HZO | 30 | Cryo preset (see material config) |
| HZO Standard 32 | 32 | Multi-level demonstration preset |
| HZO FTJ 140 | 140 | High-level FTJ preset |
| AlScN | 8-16 | High-Pr material preset |

---

## GUI Components

- **P-E Hysteresis Plot** - Polarization vs field
- **Level Indicator** - Discrete level gauge
- **Phase Indicator** - RESET -> SETTLE -> WRITE -> READ -> VERIFY
- **Material Picker** - Searchable list with property tables
- **Calibration Status** - Per-temperature calibration state and interpolation

---

## Export

- JSON export with metadata (material, temperature, parameters)
- CSV export for data analysis
- Debug logs for calibration and write/verify steps

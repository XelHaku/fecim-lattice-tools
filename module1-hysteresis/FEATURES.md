# Module 1: Hysteresis - Features

## Features

- **Interactive P-E Loop Visualization** — Real-time hysteresis curves with animated polarization switching
- **30-Level Analog Memory Demo** — Program/read discrete FeCIM states with write verification
- **Multiple Run Modes** — Fyne GUI, TUI, headless ASCII, Vulkan graphics
- **8 Material Library** — HZO variants, AlScN, cryogenic, FTJ (140 states)
- **Waveform Control** — Sine, triangle, square, manual slider
- **Educational Slides** — Ferroelectric physics explanations built-in

## Physics Models

| Model | Description |
|-------|-------------|
| **Preisach (Basic)** | Hyperbolic tangent switching with history-dependent minor loops |
| **Mayergoyz Preisach** | Full 40×40 hysteron grid, bivariate Gaussian distribution |
| **KAI Switching** | Kolmogorov-Avrami-Ishibashi time-resolved domain switching |
| **Temperature Effects** | Curie-Weiss law for Ec(T), Pr(T), Arrhenius for τ(T) |
| **Fatigue/Wake-up** | Stretched exponential endurance degradation |

## Key Parameters

| Parameter | Value | Notes |
|-----------|-------|-------|
| FeCIM Levels | 30 | 4.91 bits/cell |
| Pr (RT) | 15-34 µC/cm² | Material-dependent |
| Pr (4K) | 75 µC/cm² | Cryogenic enhanced |
| Ec | 0.6-5.0 MV/cm | Material-dependent |
| Switching τ | 1-20 ns | Temperature-dependent |
| Endurance | 10⁸-10¹² cycles | Material-dependent |
| Retention | 10-100 years @ 85°C | Arrhenius model |

## Materials Available

1. HZO (Si-doped) — Baseline 30 levels
2. FeCIM HZO — Dr. Tour's specs
3. Literature Superlattice — 64 levels (Cheema 2020)
4. Cryogenic HZO — 75 µC/cm² at 4K
5. HZO Standard 32 — Peer-reviewed (Oh 2017)
6. HZO FTJ 140 — 140 states (Song 2024)
7. AlScN — High Pr (120 µC/cm²), 8-16 levels

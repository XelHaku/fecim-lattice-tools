# Hysteresis Demo Documentation

**FeCIM Visualizer - Ferroelectric P-E Curve Demo**

> *"It's got 30 discrete states. So it's not 0-1-0-1."* — Dr. external research group  
> *Simulation baseline (unverified).*

**Complexity:** Beginner (Graphics only)

---

## Overview

The Hysteresis demo provides an interactive visualization of ferroelectric hysteresis in HfO2-ZrO2 (HZO) superlattice materials. This demo illustrates the fundamental physics of ferroelectric memory cells that enable FeCIM's compute-in-memory technology.

### What This Demo Shows

1. **P-E Hysteresis Loop** — The characteristic polarization-electric field curve of ferroelectric materials
2. **30 Discrete States (Baseline)** — Demo baseline for multi-level cell (MLC) storage (~4.9 bits/cell; simulation baseline)
3. **Preisach Hysteresis Model** — Physics-accurate simulation of domain switching
4. **Real-time Simulation** — Interactive control of electric field and waveforms
5. **Write/Read Operations** — Demonstrates non-volatile memory behavior

---

## Quick Start

```bash
# From project root
./launch.sh

# Or build and run directly
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools

# Then select the "Hysteresis" tab
```

---

## UI Layout

```
┌───────────────────────────────────────────────────────────────────────────────────────────┐
│  FeCIM Ferroelectric Hysteresis Visualization                                             │
│  "It's got 30 discrete states. So it's not 0-1-0-1." — Dr. external research group                     │
├───────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                           │
│ ┌────────┐ ┌──────────────────────────┐ ┌───┐ ┌───────────────────┬───────────────────┐  │
│ │ Memory │ │   P-E Hysteresis Loop    │ │30 │ │ Controls          │ What You're       │  │
│ │  Cell  │ │                          │ │ L │ │                   │ Seeing            │  │
│ │ ┌────┐ │ │  P (µC/cm²)              │ │ E │ │ Material: [HZO v] │                   │  │
│ │ │ 24 │ │ │   40 ┼    ╭──────╮       │ │ V │ │ Waveform: [Demo v]│ WRITE/READ DEMO   │  │
│ │ └────┘ │ │  +Pr ┼────╯      │       │ │ E │ │ E-field: ███░░░░  │                   │  │
│ │        │ │   20 ┼           │       │ │ L │ │ Frequency: 0.5 Hz │ 1. WRITE: E>Ec    │  │
│ │ Level  │ │    0 ┼───────────┼─→ E   │ │ S │ │ Trail: 500 pts    │    sets state     │  │
│ │ 24/30  │ │  -20 ┼           │       │ │   │ │ [Pause] [Reset]   │ 2. HOLD: E=0      │  │
│ │        │ │  -Pr ┼────╮      │       │ │ ▓ │ ├───────────────────┤    P persists!    │  │
│ │Positive│ │  -40 ┼    ╰──────╯       │ │ ▓ │ │ Current State     │ 3. READ: E<Ec     │  │
│ │   P    │ │      -1  -Ec 0 +Ec  1    │ │ ▓ │ │ E: 0.85 MV/cm     │    no change      │  │
│ └────────┘ └──────────────────────────┘ │ ░ │ │ P: 25.3 µC/cm²    ├───────────────────┤  │
│                                         │ ░ │ │ Level: 24/30      │ Memory Log        │  │
│ This is the cell                        └───┘ │ Mode: [WRITE]     │                   │  │
│                                               │                   │ >> WRITE(28)      │  │
│                                               │                   │    HOLD @ 27      │  │
│                                               │                   │ << READ...        │  │
│                                               │                   │    Got: 27 [OK]   │  │
│                                               │                   │ >> WRITE(5)       │  │
│                                               └───────────────────┴───────────────────┘  │
│  ● Write/Read Demo | WRITING 5...                                                        │
└───────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Waveform Modes

| Mode | Description |
|------|-------------|
| **Manual** | Drag slider to control E-field directly |
| **Sine Wave** | Continuous sweep traces full hysteresis loop |
| **Triangle Wave** | Linear ramps show Ec switching thresholds |
| **Square Wave** | Instant jumps show rapid state flipping |
| **Random Walk** | Picks random target levels, demonstrates multi-level storage |
| **Write/Read Demo** | Full memory operation cycle: WRITE → HOLD → READ |

---

## GUI Controls

- **E-field Slider**: Drag to control electric field (Manual mode)
- **Waveform Dropdown**: Select input waveform type
- **Material Dropdown**: Switch between HZO variants
- **Physics Engine Dropdown**: Toggle between L‑K (dynamic) and Preisach (quasi‑static)
- **Frequency Slider**: Adjust speed (affects all auto modes)
- **Trail Slider**: Adjust plot history length
- **Pause/Resume Button**: Control simulation
- **Reset Button**: Clear history and restart

---

## Visual Indicators

- **Memory Cell**: Color-coded square showing current level (1-30)
- **P-E Plot**: Real-time hysteresis curve with Ec/Pr markers
- **Level Bar**: 30-segment vertical indicator
- **Mode Indicator**: Shows WRITE (|E|>Ec) or READ (|E|<Ec)
- **Educational Slide**: Context-sensitive explanations
- **Memory Log**: Real-time read/write operation log

---

## Physics Model

For detailed physics, see [../hysteresis/hysteresis.physics.md](../hysteresis/hysteresis.physics.md).

### Quick Summary

The demo supports two physics engines:

- **L‑K (dynamic)**: Landau‑Khalatnikov time‑domain solver (same equation as headless).
- **Preisach (quasi‑static)**: Mayergoyz hysteron stack for fast loop visualization.

Preisach form (when selected):

```
P(E) = ∫∫ μ(α, β) γ_αβ dα dβ  →  Discretized: P = Σ μᵢ × γᵢ
```

**Key principle:** The hysteresis loop is EMERGENT from the sum of microscopic hysterons, not drawn.

### Write vs Read Operations

```
WRITE: |E| > Ec  → Polarization changes (crosses coercive field)
READ:  |E| < Ec  → Polarization unchanged, state sensed non-destructively
```

### ISPP Write/Read Demo (Implementation Details)

The Write/Read demo runs a multi‑phase **ISPP (Incremental Step Pulse Programming)**
controller to reach a target discrete level. The implementation is split across:

- `module1-hysteresis/pkg/gui/simulation.go` (phase sequencing)
- `module1-hysteresis/pkg/controller/writer.go` (ISPP pulse/verify loop)

#### Step Sequencing

**Outer demo phases (simulation loop):**

1. **PREP (Phase 0)**  
   **Saturate to the opposite polarity** of the target using `±2 × Ec` until
   `|P| ≥ 0.75 × Ps` (upper targets → negative saturation, lower targets → positive).
2. **WRITE (Phase 2)**  
   Delegates to `WriteController` for the ISPP pulse loop (apply/wait/verify).
3. **DISPLAY (Phase 5)**  
   Report success/failure, update stats, and select the next target level.

**ISPP pulse loop (inside `WriteController`):**

- **Apply**: ramp to the next pulse field (`CurrentField`).
- **Wait**: hold briefly so the field reaches the target.
- **Verify**: return to 0 V/m and read the new level.
- **Adjust**: incremental step sizing based on level error; use calibration on the first pulse.
- **Resetting**: overshoot recovery uses **reverse‑direction correction pulses** (no full saturation).

#### Termination Criteria

- **Success**: `currentLevel == targetLevel` (strict equality).
- **Failure**: `PulseCount >= MaxRetries` (default 50 pulses).
- **Overshoot**: crossing the target on the *wrong hysteresis branch* → enter `RESETTING`
  and apply reverse‑direction correction pulses.

#### Parameter Choices (Physical Meaning)

| Parameter | Location | Meaning |
|-----------|----------|---------|
| `EcField` | `writer.go` | Coercive field baseline (V/m). |
| `MaxField` | `writer.go` | Maximum programming field; default `~2.5 × Ec`. |
| `PulseDuration` | `simulation.go` | Pulse width per ISPP step; set to ~40% of the phase duration so the ramp can settle. |
| `VMin`, `VMax` | `writer.go` | Tracking bounds for the **absolute** field magnitude (diagnostics + step sizing). |
| `FromSaturation` | `writer.go` | `true` after PREP saturation; enables calibration hints for the first pulse. |
| `CalibManager` | `algo/calibration.go` | Stores per‑level calibrated fields; used only for the **first** ISPP pulse. |

#### Autonomous Runtime Recalibration

The WRD controller now recalibrates **during runtime** when convergence is poor:

- **Trigger conditions** (defaults):
  - `overshoots ≥ 2` in a single target cycle, or
  - `pulses ≥ 12` without hitting the target.
- **Execution**: recalibration runs **between targets** (DISPLAY phase) to avoid
  corrupting the active hysteresis state.
- **Persistence**: the new calibration is saved to `data/calibrations/*.json`.

#### Constraints / Limits

- **Field bounds**: `VMin ≥ 0`, `VMax ≤ MaxField`.
- **Overshoot reset**: uses a **deep reset** of `±1.5 × MaxField` with sign based on direction.
- **Reset direction**: locked at overshoot detection to prevent sign flips during reset; next pulse sign is re-derived from target vs. current.
- **Retry limit**: `MaxRetries = 50` (configurable).
- **Directionality**: pulse sign derives from target vs. current level (and target branch when reset).
- **Overshoot detection**: compares the last verified level to the post‑pulse level using the pulse direction to detect true crossings.
- **Quantization**: level readout uses `normalizedP` → discrete level mapping (0–N‑1).

#### Headless ISPP (`--mode hysteresis --engine preisach|lk`)

The headless diagnostics path uses the **same WRD phase machine + WriteController**
as the GUI (`module1-hysteresis/pkg/controller`). By default it runs the **Preisach
quasi‑static engine** (`module1-hysteresis/pkg/ferroelectric/preisach.go`) for WRD/ISPP
parity with the GUI. Use `--engine lk` to switch to the dynamic Landau‑Khalatnikov
solver (`shared/physics/landau.go`) for performance profiling and equation checks.
Targets are resolved to **discrete levels** from the conductance mapping (Gmin/Gmax),
then the controller steers field pulses to hit those levels exactly.

**Sequence:**
1. **PREP**: saturate to the **opposite polarity** of the target (`±2 × Ec`, threshold `0.75 × Ps`).
2. **WRITE (controller)**: Apply → Wait → Verify
   - **Apply**: pulse field from calibration (if available) or `~1 × Ec` toward target.
   - **Wait**: hold field near target for the pulse window.
   - **Verify**: settle to 0, read level via active engine (`P → G → level`).
3. **Resetting**: if the level crosses the target in the wrong direction, apply reverse‑direction correction pulses.
4. **DISPLAY**: ramp to 0 and advance to the next target.

**Termination:**
- **Success**: exact level match.
- **Failure**: `MaxRetries` exceeded (controller returns `FAILED`).

**Headless defaults (Feb 2026):**
| Parameter | Value | Meaning |
|-----------|-------|---------|
| `MaxField` | `2.5 × Ec` | Safe upper bound in E‑field |
| `PulseDuration` | `τ` | Characteristic switching time (material) |
| `MaxRetries` | `50` | Max program‑verify pulses before `FAILED` |
| `dtNominal` | `min(1e‑4, τ / 10,000)` | Nominal L‑K step (GUI‑aligned, stability‑clamped) |
| `dtMin` | `min(1e‑6, dtNominal)` | Reduced step near ±Ec |
| `dtMax` | `min(0.025, τ)` | Cap for stability |

**Headless multi‑step validation:** `cmd/fecim-lattice-tools/mode.go` runs a 3‑step
sequence (`pos-1`, `pos-2`, `neg-1`) to confirm end‑to‑end ISPP convergence across
positive and negative branches without forcing a full reset between each step.

**Authoritative validation (headless‑first):** Headless mode is the acceptance gate
for physics + ISPP correctness. Run `./launch.sh --logger --verbosity debug --mode hysteresis --engine preisach`
(default engine) and confirm:
- `ISPP` logs show `APPLY → WAIT → VERIFY → (RESETTING)` sequences per step.
- CSV rows show `wrd_target_level` matching `controller_target_level` during WRITE/VERIFY.
- Headless runs emit a full‑resolution CSV at `logs/hysteresis-<material>-<timestamp>.csv` (same schema as GUI)
  with `controller_*` fields for ISPP state transitions.

For L‑K equation verification and perf counters, run
`./launch.sh --logger --verbosity debug --mode hysteresis --engine lk` and confirm:
- `lk-solver` logs include `E_applied`, `E_dep`, `E_eff`, `dG_dP`, `rho_eff`, `Alpha`, `Beta`, `Gamma`, `K_dep`.
- `ISPP` logs show `APPLY → WAIT → VERIFY → (RESETTING)` sequences per step.
GUI runs are **illustrative only**; physics verification is done headlessly.

### Key Parameters (HZO Materials)

| Parameter | Default HZO | Optimized | FeCIM |
|-----------|-------------|-----------|-------|
| Pr (µC/cm²) | 25 | 45 | 30 |
| Ps (µC/cm²) | 30 | 50 | 35 |
| Ec (MV/cm) | 1.2 | 0.8 | 1.0 |
| τ (ns) | 1.0 | 0.5 | 10* |
| Endurance | 10¹⁰ | 10¹² | 10¹¹ |

*τ is defined but NOT used in real-time visualization (quasistatic approximation).

---

## Architecture

```
module1-hysteresis/
├── cmd/demo/
│   └── main.go              # Entry point (standalone mode)
├── pkg/
│   ├── ferroelectric/       # Physics engine
│   │   ├── preisach.go      # Basic Preisach model
│   │   ├── preisach_advanced.go  # Full Mayergoyz model
│   │   ├── material.go      # HZO material parameters
│   │   └── render.go        # ASCII rendering utilities
│   └── gui/
│       ├── gui.go           # Standalone GUI application
│       └── embedded.go      # Embeddable app for unified visualizer
└── shaders/                 # (Reserved for future Vulkan mode)
```

---

## Testing

```bash
# Run module tests
cd module1-hysteresis
go test ./...

# Run ferroelectric package tests with verbose output
go test ./pkg/ferroelectric -v
```

---

## Troubleshooting

### GUI (Fyne) fails to start

**Linux:** Install required dependencies:
```bash
# Debian/Ubuntu
sudo apt-get install libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel

# Arch
sudo pacman -S mesa libxcursor libxrandr libxinerama libxi
```

---

## References

1. Mayergoyz, I.D. "Mathematical Models of Hysteresis" IEEE Trans. Mag. (1986)
2. Böscke et al. "Ferroelectricity in HfO₂ Thin Films" APL (2011)
3. Park et al. "Ferroelectricity in Doped Hafnium Oxide" Adv. Mater. (2015)
4. Dr. external research group, "FeCIM Presentation" (Nov 2024)
5. Bartic et al. "Preisach Model for Ferroelectric Capacitors" J. Appl. Phys. (2001)

---

*This document is part of the FeCIM Visualizer project. For beginner explanations, see [../hysteresis/hysteresis.ELI5.md](../hysteresis/hysteresis.ELI5.md). For deep physics, see [../hysteresis/hysteresis.physics.md](../hysteresis/hysteresis.physics.md). For research references, see [../hysteresis/hysteresis.research.md](../hysteresis/hysteresis.research.md).*

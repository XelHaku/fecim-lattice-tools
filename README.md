# Ferroelectric CIM Lattice Tools

**Educational visualization suite for Ferroelectric Compute-in-Memory (FeCIM) concepts (simulation-only)**

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![Fyne](https://img.shields.io/badge/Fyne-2.7.2-blue?logo=go)](https://fyne.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)]()
[![Modules](https://img.shields.io/badge/Modules-7-brightgreen.svg)]()
[![CI](https://github.com/your-org/fecim-lattice-tools/actions/workflows/ci.yml/badge.svg)](https://github.com/your-org/fecim-lattice-tools/actions/workflows/ci.yml)

> **Status**: Education phase (simulation-only). See `docs/project/STATUS.md`.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Command Line Options](#command-line-options)
- [CLI Subcommands](#cli-subcommands)
- [Modules](#modules)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [Claims Policy](#claims-policy)
- [Technical Stack](#technical-stack)
- [Repository Structure](#repository-structure)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

FeCIM Lattice Tools is a Go/Fyne application that visualizes ferroelectric hysteresis, crossbar MVM behavior, quantization/noise effects, and a small MNIST demo. It is a **simulator** meant for learning and exploration. Values shown in the UI are **model parameters**, not device measurements.

**Simulation defaults:**
- Default conductance quantization: **30 levels** (configurable).
- Material presets and temperature controls are provided for exploration. See `module1-hysteresis/pkg/ferroelectric/material.go` for defaults.

---

## Quick Start

```bash
# Clone and run
git clone https://github.com/your-org/fecim-lattice-tools.git
cd fecim-lattice-tools
./launch.sh
```

Or build manually:

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools
```

---

## Installation

See `INSTALLATION.md` for prerequisites, optional dependencies, and platform-specific setup.

### Running Tests

```bash
go test ./...                              # See CI for latest status
go test -v ./module2-crossbar/pkg/crossbar # Crossbar tests only
go test -race ./...                        # Race detector (optional)

# CI-like settings (recommended when reproducing CI failures)
make test-ci
make test-race-ci
```

See: `docs/development/TESTING.md` and `docs/development/CI.md`.

### Headless / Non-GUI Usage

The GUI requires a display server, but several workflows are supported without one:

```bash
# Headless hysteresis diagnostics (prints + optional CSV log)
./fecim-lattice-tools --mode hysteresis

# Hysteresis subcommand headless ASCII mode
./fecim-lattice-tools hysteresis --headless --material superlattice
```

See: `docs/development/HEADLESS.md`.

### Benchmarks

Microbenchmarks for hot loops live alongside unit tests.

```bash
make bench
# Or targeted:
BENCH=BenchmarkEngineStep BENCH_COUNT=5 make bench
```

See: `BENCHMARKS.md`.

### Command Line Options

```bash
./launch.sh [options]
# Or: ./fecim-lattice-tools [options]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--logger` | off | Enable file logging to `logs/<timestamp>-fecim.log` |
| `--verbosity` | info | Log level: `off`, `info`, `debug`, `trace` |

**Examples:**
```bash
./launch.sh --logger --verbosity debug  # Enable logging with debug output
./launch.sh --verbosity trace           # Console-only trace output (no file)
```

### CLI Subcommands

All modules support standalone CLI operation with common flags:

```bash
# JSON output for scripting
fecim-lattice-tools hysteresis --headless --json --material superlattice

# Quiet mode (results only)
fecim-lattice-tools circuits cli --all --json --quiet

# Config file support
fecim-lattice-tools mnist cli --config config.yaml --evaluate

# Batch processing
fecim-lattice-tools hysteresis --headless --batch materials.txt --json
```

**Common CLI Flags (all subcommands):**
| Flag | Description |
|------|-------------|
| `--json` | Output results as JSON |
| `-q, --quiet` | Suppress informational output |
| `-c, --config FILE` | Load from YAML/JSON config |
| `--batch FILE` | Process multiple items from file |
| `-o, --output FILE` | Write to file (default: stdout) |

See `docs/CLI.md` for complete CLI reference.

---

## Modules

The visualizer includes 7 interconnected modules:

```
PHYSICS → COMPUTE → APPLICATION → SYSTEM → BUSINESS → TOOLING → REFERENCE

┌────────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐
│  Module 1  │───▶│  Module 2  │───▶│  Module 3  │    │  Module 7  │
│ Hysteresis │    │  Crossbar  │    │   MNIST    │    │    Docs    │
└────────────┘    └────────────┘    └────────────┘    └────────────┘
      │                                    │                 ▲
      ▼                                    ▼                 │
┌────────────┐    ┌────────────┐    ┌────────────┐          │
│  Module 4  │◀───│  Module 5  │◀───│  Module 6  │──────────┘
│  Circuits  │    │ Comparison │    │    EDA     │
└────────────┘    └────────────┘    └────────────┘
```

| Module | Focus | Description |
|--------|-------|-------------|
| **1. Hysteresis** | Physics | Preisach + Landau-Khalatnikov engines, multi-level state visualization |
| **2. Crossbar** | Compute | MVM operations with IR drop, sneak paths, drift, and noise |
| **3. MNIST** | Application | Dual-mode inference (FP32 vs CIM) with adjustable quantization/noise |
| **4. Circuits** | System | DAC/ADC/TIA behavior and peripheral timing/precision tradeoffs |
| **5. Comparison** | Business | Model-based, clearly labeled comparisons and projections |
| **6. EDA Suite** | Tooling | Generate illustrative JSON/CSV/SPICE/Verilog/DEF outputs (not signoff-ready) |
| **7. Docs** | Reference | In-app documentation browser and glossary |

---

## Keyboard Shortcuts

All modules support comprehensive keyboard shortcuts for efficient navigation and control. Press `?` or `/` in any module to see the full list.

### Common Shortcuts (All Modules)

| Shortcut | Action |
|----------|--------|
| `Ctrl+S` | Save/Export data |
| `Ctrl+E` | Export data |
| `Ctrl+R` | Reset simulation/view |
| `Space` | Pause/Resume animation |
| `↑/↓/←/→` | Navigate (context-dependent) |
| `Tab` | Next tab/view |
| `Shift+Tab` | Previous tab/view |
| `+/=` | Increase value |
| `-` | Decrease value |
| `?` or `/` | Show help dialog |

### Module-Specific Shortcuts

**Module 1 (Hysteresis):**
- `E/D` - Adjust E-field (Manual mode)
- `T/G` - Adjust temperature (±25K)
- `F/V` - Adjust frequency (×2 or ÷2)
- `W` - Cycle waveform mode
- `R` - Reset simulation

**Module 2 (Crossbar):**
- `M` - Run MVM operation
- `N` - Randomize weights
- `A` - Toggle architecture (0T1R/1T1R/2T1R)
- `T/G` - Adjust temperature

**Module 3 (MNIST):**
- `C` - Clear canvas
- `R` - Load random test digit
- `E` - Evaluate network
- `L` - Load test data

**Module 4 (Circuits):**
- `P` - Program selected cell
- `R` - Read selected cell
- `C` - Run compute operation
- `W/D/M` - Switch Write/Read/Compute mode
- `A` - Toggle animation

**Module 5 (Comparison):**
- `N/P` - Next/Previous demo phase
- `←/→` - Cycle workloads

**Module 6 (EDA):**
- `1/2` - Jump to view by number
- `Space` - Cycle views

The keyboard system is implemented in `shared/keyboard/` and can be extended for new modules.

---

## Claims Policy

This repository does **not** present hardware performance claims. External scientific claims (if any) are tracked in `docs/comparison/HONESTY_AUDIT.md`. If a claim is not listed there, treat it as **unverified**.

Physics acceptance criteria are defined in `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md`.
GUI minimum supported size is documented in `docs/development/GUI_MINIMUMS.md` (1024×768).

---

## Technical Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.24+ |
| GUI Framework | Fyne 2.7.2 |
| Physics Model | Preisach/Mayergoyz + Landau-Khalatnikov (educational) |
| Compute | Crossbar MVM simulation |
| Non-Idealities | IR drop, sneak paths, drift |
| Tests | See CI (`go test ./...`) |
| Documentation | Markdown + in-app viewer |

---

## Repository Structure

```
fecim-lattice-tools/
├── cmd/fecim-lattice-tools/    # Unified GUI entry point
├── module1-hysteresis/         # P-E curve physics
├── module2-crossbar/           # MVM + non-idealities
├── module3-mnist/              # Neural network demo
├── module4-circuits/           # Peripheral circuits
├── module5-comparison/         # Technology comparisons (model-based)
├── module6-eda/                # Design suite (educational)
├── module7-docs/               # Documentation browser
├── shared/                     # Common theme, logging, widgets
├── docs/                       # Markdown documentation files
├── data/                       # Calibration data
└── go.mod
```

---

## Contributing

Contributions are welcome. Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Follow existing code patterns (see `CLAUDE.md` for conventions)
4. Run tests (`go test ./...`)
5. Submit a pull request

---

## License

MIT License

This is an independent educational visualization project. Ferroelectric CIM research originates from the broader research community. No official affiliation.

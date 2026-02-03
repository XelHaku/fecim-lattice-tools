# CLAUDE.md - FeCIM Lattice Tools

## For AI Agents

**Full reference:** See `docs/development/scriptReference.md` for detailed lookups.

| I need to... | Look in |
|--------------|---------|
| Find a function | `docs/development/scriptReference.md#quick-function-lookups` |
| Fix an error | `docs/development/scriptReference.md#error-resolution-guide` |
| Add a feature | `docs/development/scriptReference.md#decision-trees` |
| Check thread safety | `docs/development/scriptReference.md#thread-safety-guide` |
| Fix Fyne GUI issues | `docs/development/GUI/FYNE_NOTES.md` |
| Run/understand tests | `docs/development/TESTING.md` |
| Review UI analysis | `docs/development/HYPER_ANALYSIS_REPORT.md` |
| EDA documentation | `docs/eda/README.md` |
| OpenLane integration | `docs/eda/guides/integration.md` |
| EDA CLI reference | `docs/eda/references/cli-reference.md` |

## Overview

Go-based lattice tool suite for Ferroelectric Compute-in-Memory (FeCIM) visualization and simulation. It includes configurable material presets, crossbar models, and an educational EDA pipeline.

**Status**: Education phase (simulation-only). See `docs/project/STATUS.md`.

**Core concept**: The simulator quantizes conductance to a default of 30 discrete levels (configurable). This is a **simulation baseline**, not a validated hardware claim.

> **Historical reference**: Dr. external research group, COSM 2025 - [Transcript](docs/video-transcripts/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md) (conference material; not peer-reviewed).

## Build & Run

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools
# Or: ./launch.sh
```

## Key Rules

### Do
- Use `fyne.Do(func() { ... })` for all UI updates from goroutines
- Quantize to 30 levels: `crossbar.QuantizeTo30Levels(value)`
- Follow embedded app interface: `BuildContent()`, `Start()`, `Stop()`
- Run `go test ./...` before committing

### Don't
- Add demos without implementing embedded interface
- Use blocking operations on main UI thread
- Commit binaries

## Project Structure

```
cmd/fecim-lattice-tools/     # Main unified app entry point
module1-hysteresis/       # P-E curve, Preisach model
module2-crossbar/         # MVM, non-idealities (IR drop, sneak paths, drift)
module3-mnist/            # Neural network digit recognition
module4-circuits/         # DAC/ADC/TIA peripherals
module5-comparison/       # Technology comparison
module6-eda/              # EDA tools
shared/                   # Theme, widgets, logging
```

## Model Defaults (Simulation Parameters)

The project includes **preset parameters** for education and visualization. Treat these as **simulation defaults**, not validated device measurements.

- Material presets: `module1-hysteresis/pkg/ferroelectric/material.go`
- Crossbar defaults: `module2-crossbar/pkg/crossbar/array.go`
- EDA defaults: `module6-eda/pkg/core/config.go`

## Accuracy & Honesty Policy

Scientific accuracy over marketing claims. Full audit: `docs/comparison/HONESTY_AUDIT.md`.

### Verified External Claim (Current Audit)

- **98.24% MNIST accuracy** in HZO FTJ reservoir computing (Journal of Alloys and Compounds 2025). This is **not** a FeCIM device claim and should not be attributed to this simulator.

### Unverified/Removed Claims (Do Not Present as Facts)

- 30 analog states for Tour device (conference-only reference)
- 87% MNIST accuracy (conference-only reference)
- Energy multipliers vs NAND or GPUs without peer-reviewed measurement evidence

## Testing

```bash
go test ./...                            # See CI for latest status
go test ./module2-crossbar/pkg/crossbar  # Crossbar only
```

Full test documentation: `docs/development/TESTING.md`

## Git Conventions

- Commit: `type: description` (feat, fix, docs, refactor, test, chore)
- Run tests before pushing

## Ignore

- `logs/`, `output/`, `docs/archive/`

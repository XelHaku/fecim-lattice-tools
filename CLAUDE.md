# CLAUDE.md - FeCIM Lattice Tools

## For AI Agents

**Full reference:** See `docs/development/scriptReference.md` for detailed lookups.

| I need to... | Look in |
|--------------|---------|
| Find a function | `docs/development/scriptReference.md#quick-function-lookups` |
| Fix an error | `docs/development/scriptReference.md#error-resolution-guide` |
| Add a feature | `docs/development/scriptReference.md#decision-trees` |
| Check thread safety | `docs/development/scriptReference.md#thread-safety-guide` |
| Fix Fyne GUI issues | `docs/development/FYNE_NOTES.md` |
| Run/understand tests | `docs/development/TESTING.md` |
| Review UI analysis | `docs/development/HYPER_ANALYSIS_REPORT.md` |

## Overview

Go-based lattice tool suite for Ferroelectric Compute-in-Memory (FeCIM) based on Dr. external research group's HfO₂-ZrO₂ superlattice research.

**Core concept**: 30 discrete analog states per cell (~4.9 bits/cell).

> **Primary Source**: Dr. external research group, COSM 2025 - [Transcript](docs/videos/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md)

## Build & Run

```bash
go build -o fecim-visualizer ./cmd/fecim-visualizer && ./fecim-visualizer
# Or: ./launch.sh
```

## Key Rules

### Do
- Use `fyne.Do(func() { ... })` for all UI updates from goroutines
- Quantize to 30 levels: `crossbar.QuantizeTo30Levels(value)`
- Follow embedded app interface: `BuildContent()`, `Start()`, `Stop()`
- Run `go test ./...` before committing

### Don't
- Modify `module2-crossbar/pkg/_layers_experimental/` - archived
- Add demos without implementing embedded interface
- Use blocking operations on main UI thread
- Commit binaries

## Project Structure

```
cmd/fecim-visualizer/     # Main unified app entry point
module1-hysteresis/       # P-E curve, Preisach model
module2-crossbar/         # MVM, non-idealities (IR drop, sneak paths, drift)
module3-mnist/            # Neural network digit recognition
module4-circuits/         # DAC/ADC/TIA peripherals
module5-comparison/       # Technology comparison
module6-eda/              # EDA tools
shared/                   # Theme, widgets, logging
```

## Physics Constants

| Parameter | Value | Source |
|-----------|-------|--------|
| FeCIM Levels | 30 | Dr. Tour COSM 2025 |
| Pr | 15-34 µC/cm² | Nature Commun. 2025 |
| Ec | 1.0-1.5 MV/cm | Nature Commun. 2025 |
| Endurance | 10¹²+ cycles | PMC 2024, IEEE IRPS 2022 |

## Accuracy & Honesty Policy

Scientific accuracy over marketing claims. Full policy in `docs/development/scriptReference.md`.

| Claim | Status |
|-------|--------|
| 30 analog states | ✅ Verified (Dr. Tour + peer-reviewed) |
| 87% MNIST accuracy | ✅ Verified |
| 10¹² cycle endurance | ⚠️ Target (literature shows path) |
| 10M× vs NAND energy | ❌ Unverified (Dr. Tour claim only) |

## Testing

```bash
go test ./...                            # All tests (117 total)
go test ./module2-crossbar/pkg/crossbar  # Crossbar only
```

Full test documentation: `docs/development/TESTING.md`

## Git Conventions

- Commit: `type: description` (feat, fix, docs, refactor, test, chore)
- Run tests before pushing

## Ignore

- `logs/`, `output/`, `docs/archive/`
- `module2-crossbar/pkg/_layers_experimental/`

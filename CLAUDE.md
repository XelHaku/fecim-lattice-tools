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

## Overview

Go-based lattice tool suite for Ferroelectric Compute-in-Memory (FeCIM) based on Dr. external research group's HfO₂-ZrO₂ superlattice research.

**Core concept**: 30 discrete analog states per cell (~4.9 bits/cell).

> **Primary Source**: Dr. external research group, COSM 2025 - [Transcript](docs/video-transcripts/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md)

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
- Modify `module2-crossbar/pkg/_layers_experimental/` - archived
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

## Physics Constants

| Parameter | Value | Source |
|-----------|-------|--------|
| FeCIM Levels | 30 | Dr. Tour COSM 2025 (32-140 demonstrated by others) |
| Pr | 15-34 µC/cm² (RT), 75 µC/cm² (4K) | Nature Commun. 2025, Adv. Elec. Mat. 2024 |
| Ec | 0.6-1.5 MV/cm | Nature Commun. 2025, Nano Letters 2024 |
| Endurance (demonstrated) | 10⁹-10¹² cycles | IEEE IRPS 2022, Nano Letters 2024 (V:HfO₂) |
| 3D Integration | 22nm BEOL demonstrated | CEA-Leti December 2024 |

## Accuracy & Honesty Policy

Scientific accuracy over marketing claims. Full audit: `docs/comparison/HONESTY_AUDIT.md`.

### Verified Claims (Peer-Reviewed)

| Claim | Status | Evidence |
|-------|--------|----------|
| Pr: 15-34 µC/cm² | ✅ Verified | Nature Commun. 2025 (HZO measurements) |
| Pr: 75 µC/cm² @ 4K | ✅ Verified | Adv. Elec. Mat. 2024 (cryogenic) |
| Ec: 0.6-1.5 MV/cm | ✅ Verified | Nature Commun. 2025, Nano Letters 2024 |
| 32-140 analog states | ✅ Verified | Oh 2017 (32), Song 2024 (140) |
| 25-100× vs NAND | ✅ Verified | Samsung Nature 2025 |
| 10⁹ cycle endurance | ✅ Verified | IEEE IRPS 2022 |
| 10¹² cycle endurance | ✅ Verified | Nano Letters 2024 (V:HfO₂), Science 2024 |
| 96.6% MNIST accuracy | ✅ Verified | Nature Communications 2023 |
| 98.24% MNIST accuracy | ✅ Verified | ScienceDirect 2025 (FTJ reservoir) |
| 3D BEOL @ 22nm | ✅ Verified | CEA-Leti December 2024 |
| Grade 0 automotive | ✅ Verified | Fraunhofer IPMS 2024 (AEC-Q100) |
| Cryogenic 5K operation | ✅ Verified | IEEE 2024, Frontiers 2024 |

### Unverified/Removed Claims (Conference Only)

| Claim | Status | Source |
|-------|--------|--------|
| 30 analog states (Tour device) | ⚠️ Unverified | COSM 2025 (not peer-reviewed) |
| 87% MNIST accuracy (Tour) | ❌ REMOVED | Removed from tool - below peer-reviewed 96.6-98.24% |
| 10M× vs NAND energy | ❌ REMOVED | No measurement data exists (verified: 25-100×) |

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

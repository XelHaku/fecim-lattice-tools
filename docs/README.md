# FeCIM Lattice Tools Documentation

> Ferroelectric Compute-in-Memory simulation suite with a configurable 30-level default baseline.

---

## Quick Start

| Doc | Purpose |
|-----|---------|
| [ELI5.md](ELI5.md) | Concepts explained simply |
| [RUNBOOK.md](RUNBOOK.md) | Build, run, and deploy |
| [FEATURES.md](FEATURES.md) | Complete feature reference |
| [../CONTRIBUTING.md](../CONTRIBUTING.md) | How to contribute |
| [project/STATUS.md](project/STATUS.md) | Project phase, validation, and CI status |

## Build & Run

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools
# Or: ./launch.sh
```

### Command-Line Options

| Flag | Description |
|------|-------------|
| `--logger` | Enable file logging to `logs/` |
| `--verbosity <level>` | Logging level: `off`, `info`, `debug`, `trace` |
| `--calibrate` | Run hysteresis calibration (no GUI) |
| `--material <name>` | Material to calibrate (default: all) |
| `--force` | Force recalibration even if valid |
| `--verify` | Verify calibration accuracy |
| `--list-materials` | List available materials and exit |

---

## By Module

| Module | Folder | Features | Start With |
|--------|--------|----------|------------|
| 1. Hysteresis | [hysteresis/](hysteresis/) | P-E loops, Preisach model, 8 materials, temperature effects | [ELI5](hysteresis/../hysteresis/hysteresis.ELI5.md), [Physics](hysteresis/../hysteresis/hysteresis.physics.md) |
| 2. Crossbar | [crossbar/](crossbar/) | MVM, IR drop, sneak paths, 0T1R/1T1R/2T1R | [ELI5](crossbar/educational/crossbar.ELI5.md), [Physics](crossbar/educational/../educational/crossbar.physics.md), [Voltage Rules](crossbar/reference/VOLTAGE_RULES.md) |
| 3. Neural Network | [neural-network/](neural-network/) | MNIST inference, FP32 vs CIM comparison, quantization | [ELI5](neural-network/mnist.ELI5.md), [Demo](neural-network/mnist.demo.md) |
| 4. Peripheral Circuits | [peripheral-circuits/](peripheral-circuits/) | DAC/ADC/TIA, 4-phase write, ISPP | [ELI5](peripheral-circuits/circuits.ELI5.md) |
| 5. Comparison | [comparison/](comparison/) | CPU vs GPU vs FeCIM, data center projections | [ELI5](comparison/cim.ELI5.md), [Honesty Audit](comparison/HONESTY_AUDIT.md) |
| 6. EDA Design Suite | [eda/](eda/) | RTL-to-GDSII, 8 export formats, SKY130/GF180 PDKs | [ELI5](eda/guides/eli5.md), [Workflow](eda/WORKFLOW.md) |
| 7. Docs Viewer | - | Glossary (100+ terms), full-text search, markdown rendering | Embedded in app |

---

## Shared Components

| Component | Location | Purpose |
|-----------|----------|---------|
| Material Picker | `shared/widgets/` | Searchable material dialog with property tables |
| Theme System | `shared/theme/` | Consistent FeCIM color palette |
| Physics Config | `config/physics/` | Physics engine and material loading |
| Material Library | `config/materials.yaml` | 8 ferroelectric materials with 25+ properties |
| Logging | `shared/logging/` | Structured logging with verbosity levels |
| Recording | `shared/recording/` | FFmpeg screen capture with audio |

### Key Widgets

- **MaterialPicker** - Dialog with searchable list, 8 property categories, scientific unit formatting
- **ColorLegend** - Scalable color bar with dual-scale support
- **AdaptiveLayout** - Responsive design for desktop/tablet/mobile
- **ArchitectureSelector** - 0T1R/1T1R/2T1R toggle

---

## Calibration System

Temperature-aware multi-level calibration for discrete state mapping. Calibration settings live in code and configuration files (see `config/` and module documentation).

Run calibration:
```bash
./fecim-lattice-tools --calibrate --material fecim_hzo --verify
```

---

## Research & References

| Topic | Location |
|-------|----------|
| Research papers | [research-papers/](research-papers/) |
| Dr. Tour's research | [tour-group-ironlattice-research.md](tour-group-ironlattice-research.md) |
| Material analysis | [superlattice-material-analysis.md](superlattice-material-analysis.md) |
| Open source tools | [opensource-tools/](opensource-tools/) |
| PDK reference | [pdk-reference/](pdk-reference/) |
| Video transcripts | [video-transcripts/](video-transcripts/) |

---

## For Developers

| Doc | Purpose |
|-----|---------|
| [ARCHITECTURE.md](development/ARCHITECTURE.md) | System design |
| [TESTING.md](development/TESTING.md) | Test guide (see CI for current status) |
| [scriptReference.md](development/scriptReference.md) | Function lookups |
| [FYNE_NOTES.md](development/GUI/FYNE_NOTES.md) | Fyne GUI development |
| [HYPER_ANALYSIS_REPORT.md](development/HYPER_ANALYSIS_REPORT.md) | UI analysis |

---

## Accuracy & Honesty

This repository prioritizes simulation accuracy and clear labeling of assumptions. External scientific claims (if any) are tracked in [comparison/HONESTY_AUDIT.md](comparison/HONESTY_AUDIT.md). If a claim is not listed there, treat it as **unverified**.

---

## See Also

- **Project root:** [../CLAUDE.md](../CLAUDE.md) - AI agent instructions and quick reference
- **Root README:** [../README.md](../README.md) - Full project overview
- **About:** [about/](about/) - App info, contributors, thanks
- **Glossary:** [GLOSSARY.md](GLOSSARY.md) - 100+ ferroelectric terms

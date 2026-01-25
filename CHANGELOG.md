# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Changed
- **Module 6: Array Builder** - FeCIM crossbar array builder for OpenLane integration
  - Generates LEF, Liberty, Verilog, and DEF files for OpenLane flow
  - Added three operation modes: Storage (NAND), Memory (DRAM), Compute (AI)
  - New `GenerateDesign()` API supporting all three modes
  - CLI tool updated with `-mode` flag (storage, memory, compute)
  - All export formats (JSON, CSV, SPICE, Verilog, DEF) work with all modes

## [0.9.0] - 2026-01-23

### Added
- **Module 6: FeCIM Design Suite** with 7-tab interface
  - Tab 1: Cell Builder (LEF/Liberty/Verilog generation)
  - Tab 2: Layout (visual crossbar grid)
  - Tab 3: HDL (Verilog + DEF generation)
  - Tab 6: Export (JSON, CSV, SPICE, Verilog, DEF)
  - Tab 7: Learn (OpenLane documentation)
- **OpenLane integration documentation**
  - `docs/eda/eda.integration.md` - Complete integration guide
  - `docs/eda/eda.demo.md` - Demo interface guide
  - `docs/eda/OPENLANE_STUDY.md` - Validated findings from OpenLane source
  - `docs/eda/module6-openlane-integration.md` - Quick-start guide
  - `docs/eda/module6-technical-plan.md` - Implementation roadmap
- **Working examples**
  - `examples/01-basic-8x8/` - Basic crossbar array generation
  - `examples/02-mnist-layer/` - Neural network layer mapping
  - `examples/03-openlane-integration/` - Full OpenLane workflow
- **CLI tool** for automated/headless EDA file generation
- **Verilog netlist generation** with passive/1T1R support
- **DEF placement file generation** with FIXED keyword

### Improved
- Root README with Quick Start, team info, key quotes
- Module 6 README with detailed OpenLane integration
- Package documentation (godoc comments)

## [0.8.0] - 2026-01-22

### Added
- Unified tabbed GUI application (`cmd/fecim-visualizer/`)
- Modules 1-5 fully implemented and tested
- 117 passing tests across all modules
- Lattice generator with fractal placement algorithm

### Changed
- Renamed `demo*` directories to `module*` for clarity
- Consolidated non-idealities into Module 2 tabs

## [0.7.0] - 2026-01-20

### Added
- **Module 3: MNIST Classifier**
  - Dual-mode inference (floating-point vs CIM)
  - 87% accuracy demonstration
  - Non-ideality impact visualization
- **Module 4: Peripheral Circuits**
  - DAC/ADC visualization
  - TIA design
  - INL/DNL analysis
- **Module 5: Technology Comparison**
  - Energy comparison bar chart
  - Competitive technology matrix
  - Data center savings calculator

## [0.6.0] - 2026-01-18

### Added
- **Module 2: Crossbar Array**
  - Matrix-vector multiply (MVM) visualization
  - IR drop analysis with heatmap
  - Sneak path detection
  - Drift simulation over time
- 30-level quantization throughout

## [0.5.0] - 2026-01-15

### Added
- **Module 1: Hysteresis Visualization**
  - P-E curve animation
  - Preisach model implementation
  - Multi-cell array visualization

### Changed
- Migrated to Fyne v2.7.2 GUI framework

## [0.1.0] - 2026-01-10

### Added
- Initial project structure
- Basic ferroelectric physics simulation
- Proof of concept visualization

---

## Roadmap

### v1.0.0 (Target: Q1 2026)
- [ ] OpenLane flow integration testing
- [ ] Custom FeCIM cell LEF/GDS in Magic
- [ ] Liberty timing model generation
- [ ] ngspice simulation bridge

### v1.1.0 (Target: Q2 2026)
- [ ] Design space explorer (Tab 4)
- [ ] Multi-layer stacked crossbar support
- [ ] Automated DRC/LVS validation

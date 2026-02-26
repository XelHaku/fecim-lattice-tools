# FeCIM Lattice Tools

**A scientific desktop application for ferroelectric compute-in-memory (FeCIM) research and education.** Simulates HZO/PZT/BTO ferroelectric devices across 7 integrated modules: hysteresis (Preisach + Landau-Khalatnikov), crossbar arrays with IR drop and sneak paths, MNIST inference at 80% accuracy through the full CIM pipeline, peripheral circuits (DAC/ADC/TIA), technology comparison, EDA export (SPICE/Verilog/Liberty/DEF/LEF), and interactive documentation.

Built on **published physics** — Materlik 2015, Park 2015, Alessandri 2018, Guo 2018 — with core parameters cited or explicitly marked educational. Verified by 3717+ automated tests across 70 packages, Kirchhoff-law current verification, and cross-tool comparison harnesses. Reproducible: clone, run one script, verify internal model claims.

**For:** Physics/EE researchers, graduate students, device engineers working on ferroelectric memory, neuromorphic computing, or compute-in-memory architectures.

> This repository is a simulation and educational toolkit (not a silicon measurement report).

---

## Badges

<!-- Replace with repo-specific badges when publishing/releasing -->
[![Build](https://img.shields.io/badge/build-passing-brightgreen)](#)
[![Coverage](https://img.shields.io/badge/coverage-placeholder-lightgrey)](#)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](https://go.dev)
[![Fyne](https://img.shields.io/badge/Fyne-2.5.3%2B-5f5fff)](https://fyne.io)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

---

## Table of Contents

- [Features (7 Modules)](#features-7-modules)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Screenshot Placeholders](#screenshot-placeholders)
- [Physics Models Overview](#physics-models-overview)
- [Dependencies](#dependencies)
- [Repository Layout](#repository-layout)
- [Module READMEs](#module-readmes)
- [License](#license)

---

## Features (7 Modules)

1. **Module 1 — Hysteresis** (`module1-hysteresis`)  
   Preisach/LK ferroelectric switching, P-E loop analysis, and material presets.

2. **Module 2 — Crossbar** (`module2-crossbar`)  
   Crossbar MVM simulation with non-idealities (including IR drop/sneak effects).

3. **Module 3 — MNIST** (`module3-mnist`)  
   End-to-end inference flow to study algorithm/hardware interaction under CIM constraints.

4. **Module 4 — Circuits** (`module4-circuits`)  
   Peripheral-circuit abstractions (read/program paths, front-end behavior).

5. **Module 5 — Comparison** (`module5-comparison`)  
   Comparative analysis views across operating conditions and design assumptions.

6. **Module 6 — EDA** (`module6-eda`)  
   EDA-oriented utilities, exports, and integration hooks.

7. **Module 7 — Docs** (`module7-docs`)  
   Integrated documentation and references for rapid onboarding.

---

## Architecture

```text
                         ┌─────────────────────────────┐
                         │     cmd/fecim-lattice-tools │
                         │   (GUI/CLI entrypoint)      │
                         └──────────────┬──────────────┘
                                        │
                        ┌───────────────┴────────────────┐
                        │            shared/              │
                        │ common UI, utilities, logging   │
                        └───────┬───────────────┬─────────┘
                                │               │
      ┌─────────────────────────┘               └──────────────────────────┐
      │                                                                    │
┌─────▼──────┐   ┌─────────────┐   ┌────────────┐   ┌──────────────┐   ┌──▼──────────┐
│ module1    │──▶│ module2     │──▶│ module3    │   │ module4      │   │ module5     │
│ hysteresis │   │ crossbar    │   │ mnist      │   │ circuits     │   │ comparison  │
└────────────┘   └─────────────┘   └────────────┘   └──────────────┘   └────┬────────┘
       │                 │                  │                    │             │
       └─────────────────┴──────────────────┴────────────────────┴─────────────┘
                                        │
                                 ┌──────▼──────┐
                                 │ module6-eda │
                                 └──────┬──────┘
                                        │
                                 ┌──────▼──────┐
                                 │ module7-docs│
                                 └─────────────┘
```

---

## Quick Start

```bash
git clone https://github.com/your-org/fecim-lattice-tools.git
cd fecim-lattice-tools

go run ./cmd/fecim-lattice-tools
```

Or build a binary:

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
```

Run tests:

```bash
go test -race ./...
```

Run the one-command reproducibility validation pack:

```bash
bash scripts/reproduce_validation.sh
# optional report capture:
# bash scripts/reproduce_validation.sh > artifacts/repro-report.txt 2>&1
```

---

## Screenshot Placeholders

> Replace these with actual captures from each module.

- `docs/assets/screenshots/module1-hysteresis.png`
- `docs/assets/screenshots/module2-crossbar.png`
- `docs/assets/screenshots/module3-mnist.png`
- `docs/assets/screenshots/module4-circuits.png`
- `docs/assets/screenshots/module5-comparison.png`
- `docs/assets/screenshots/module6-eda.png`
- `docs/assets/screenshots/module7-docs.png`

Markdown template:

```md
![Module 1 Hysteresis Screenshot](docs/assets/screenshots/module1-hysteresis.png)
```

---

## Physics Models Overview

### 1) Preisach Ferroelectric Model
- Represents ferroelectric polarization as an ensemble of switching units (hysterons).
- Captures hysteresis memory and minor-loop behavior.
- Used for P-E loop dynamics and state trajectory exploration.

### 2) Landau–Khalatnikov (LK) Dynamics
- Time-domain ferroelectric polarization evolution derived from free-energy minimization.
- Useful for switching transients, field-dependent behavior, and dynamic effects.
- Complements Preisach-style static/phenomenological modeling.

### 3) Crossbar IR Drop Model
- Simulates finite wire resistance and voltage drops across rows/columns.
- Quantifies effective cell bias distortion and MVM accuracy degradation.
- Supports study of array scaling limits and compensation strategies.

---

## Dependencies

- **Go**: `1.24+`
- **Fyne**: `2.5.3+`

See also:
- [`go.mod`](./go.mod)
- [`INSTALLATION.md`](./docs/1-getting-started/installation.md)

---

## Repository Layout

```text
fecim-lattice-tools/
├── cmd/
├── module1-hysteresis/
├── module2-crossbar/
├── module3-mnist/
├── module4-circuits/
├── module5-comparison/
├── module6-eda/
├── module7-docs/
├── shared/
├── docs/
├── data/
└── validation/
```

---

## Module READMEs

- [Module 1 — Hysteresis](./module1-hysteresis/README.md)
- [Module 2 — Crossbar](./module2-crossbar/README.md)
- [Module 3 — MNIST](./module3-mnist/README.md)
- [Module 4 — Circuits](./module4-circuits/README.md)
- [Module 5 — Comparison](./module5-comparison/README.md)
- [Module 6 — EDA](./module6-eda/README.md)
- [Module 7 — Docs](./module7-docs/README.md)

Additional docs:
- [Contributing](./CONTRIBUTING.md)
- [Changelog](./CHANGELOG.md)

---

## License

This project is licensed under the **MIT License**. See [LICENSE](./LICENSE).

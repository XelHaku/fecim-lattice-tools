# Module 1: Hysteresis Loop Visualization

**Navigation:** [← Back to Learn](../README.md) | [ELI5](./eli5.md) | [Physics](./physics.md) | [Features](./features.md) | [Tools](./tools.md)

---

## Overview

Module 1 simulates ferroelectric hysteresis loops using Preisach and Landau-Khalatnikov models. It provides real-time visualization of polarization vs electric field (P-E curves) with support for multiple materials, temperatures, and programming modes.

**Key Concept:** Ferroelectric materials exhibit hysteresis—the polarization (P) depends not just on the current electric field (E), but on the history of applied fields. This "memory" enables non-volatile storage and analog compute-in-memory.

---

## Quick Links

### For Beginners
- **[ELI5 Explanation](./eli5.md)** - Start here if you're new to ferroelectrics
- **[Run Modes Guide](./run-modes.md)** - GUI, TUI, headless, and Vulkan modes

### For Developers
- **[Physics Reference](./physics.md)** - Equations, models, and implementation details
- **[Features](./features.md)** - What the module does and how to extend it
- **[Materials Guide](./materials.md)** - Material parameter sets and calibration

### For Researchers
- **[Open-Source Tools](./tools.md)** - Integration with external simulation tools

---

## Module Contents

```
module1-hysteresis/
├── cmd/hysteresis/          # CLI entry point
├── pkg/
│   ├── ferroelectric/       # Physics models (Preisach, L-K)
│   ├── controller/          # ISPP write controller
│   ├── gui/                 # Fyne GUI implementation
│   ├── tui/                 # Bubble Tea TUI
│   └── render/              # Vulkan renderer (optional)
└── docs/2-learn/module1-hysteresis/  # This documentation
```

---

## Quick Start

### GUI Mode (Default)
```bash
fecim-lattice-tools hysteresis
```

### Terminal UI
```bash
fecim-lattice-tools hysteresis --tui
```

### Headless Mode
```bash
fecim-lattice-tools hysteresis --headless --material superlattice
```

### With Landau-Khalatnikov Engine
```bash
fecim-lattice-tools --mode hysteresis --engine lk
```

See [Run Modes](./run-modes.md) for complete CLI reference.

---

## What You'll Learn

1. **Hysteresis Fundamentals**
   - What is polarization and electric field?
   - Why ferroelectrics have memory
   - Reading P-E loop diagrams

2. **Physics Models**
   - Preisach model (hysteron-based)
   - Landau-Khalatnikov (thermodynamic)
   - When to use each model

3. **Memory Operations**
   - WRITE phase (programming states)
   - HOLD phase (retention)
   - READ phase (sensing)
   - VERIFY phase (checking correctness)

4. **Material Properties**
   - Coercive field (Ec)
   - Remanent polarization (Pr)
   - Saturation (Ps)
   - Temperature dependence

---

## Key Features

- **Real-time visualization** of P-E hysteresis loops
- **Multiple waveform modes**: manual, sine, triangle, square, random walk
- **Write/Read/Verify demo** with ISPP-style programming
- **Material library**: HZO, superlattice, AlScN, cryogenic variants
- **Temperature simulation** with calibration caching
- **Physics engine selector**: Preisach (default) or Landau-Khalatnikov
- **30-level quantization** (configurable baseline)
- **Multi-platform**: GUI (Fyne), TUI (Bubble Tea), headless, Vulkan

---

## Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [eli5.md](./eli5.md) | Plain-language introduction | Beginners |
| [physics.md](./physics.md) | Physics equations and models | Developers, researchers |
| [features.md](./features.md) | Feature list and architecture | Developers |
| [tools.md](./tools.md) | Open-source tool ecosystem | Researchers |
| [materials.md](./materials.md) | Material parameters | All |
| [run-modes.md](./run-modes.md) | CLI usage guide | Developers |

---

## Evidence Status

- **Demonstrated:** Repository structure, navigation behavior, and code paths are implemented and verifiable from source/tests
- **Modeled:** Equations, defaults, and performance estimates are simulator models unless explicitly tied to cited measured data
- **Aspirational:** Production-scale or silicon-parity claims are roadmap intent and must not be reported as demonstrated results

See [honesty-audit.md](../../4-research/honesty-audit.md) for complete accuracy policy.

---

## Related Modules

- **[Module 2: Crossbar](../module2-crossbar/README.md)** - Uses conductance from hysteresis model
- **[Module 3: MNIST](../module3-mnist/README.md)** - Uses quantized states from hysteresis
- **[Module 4: Circuits](../module4-circuits/README.md)** - DAC/ADC peripherals for programming

---

## Source Code

- **GitHub:** [module1-hysteresis/](../../../module1-hysteresis/)
- **Physics models:** `pkg/ferroelectric/preisach.go`, `preisach_advanced.go`
- **GUI:** `pkg/gui/gui.go`, `pkg/gui/simulation.go`
- **Controller:** `pkg/controller/writer.go`

---

**Last Updated:** 2026-02-16
**Maintainer:** FeCIM Lattice Tools Project

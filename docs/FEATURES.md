# FeCIM Lattice Tools - Features

End-to-end educational toolchain for Ferroelectric Compute-in-Memory (FeCIM).

**Core Concept:** Discrete analog states per cell (demo baseline: 30; higher-state presets are illustrative and unverified). See `docs/4-research/honesty-audit.md` (DOI: (add)).

---

## Module Summary

| Module | Purpose |
|---|---|
| 1. Hysteresis | P-E curve simulator + material presets |
| 2. Crossbar | MVM simulator with non-idealities |
| 3. MNIST | FP vs CIM inference demo |
| 4. Circuits | DAC/ADC/TIA signal-chain demo |
| 5. Comparison | Business-case visualization (model-based) |
| 6. EDA | Array compiler + export pipeline |
| 7. Docs | In-app documentation browser |

---

## Module Highlights

### Module 1: Hysteresis

- Preisach and Landau-Khalatnikov engines
- Multi-level write/read/verify with simulated calibration
- Temperature-aware calibration cache

### Module 2: Crossbar

- Analog MVM/VMM with 30-level quantization
- IR drop, sneak paths, drift, RC delay, variation, endurance
- Optional GPU MVM acceleration + tool validation widgets

### Module 3: MNIST

- Dual-mode FP vs CIM inference
- Quantization and noise controls
- Confusion matrix + activation views + energy widget

### Module 4: Circuits

- DAC -> Pump -> ferroelectric cell model -> TIA -> ADC pipeline
- INL/DNL analysis, voltage-zone visualization
- Architecture voltage-rule visualization (0T1R/1T1R/2T1R)

### Module 5: Comparison

- CPU/GPU/FeCIM modeled comparisons
- Workload library (MNIST -> LLM-70B)
- Data-center scaling and market visualization (scenario inputs)

### Module 6: EDA

- Storage/Memory/Compute modes
- JSON/CSV/SPICE/Verilog/DEF exports
- OpenLane integration helpers and validation tools

### Module 7: Docs

- Curriculum-first documentation viewer with search, ToC, glossary

---

## Shared Infrastructure

- Unified FeCIM theme
- Centralized logging
- Shared physics and peripherals libraries

---

## Cross-Module Workflow

```
Module 1 (Calibrate) -> Module 2 (Simulate) -> Module 3 (Infer)
                                v
                        Module 4 (Circuits)
                                v
                        Module 5 (Comparison)
                                v
                        Module 6 (Export)
```

Workflow is conceptual; modules can be used independently.

---

## See Also

- [Module 1 FEATURES.md](../module1-hysteresis/FEATURES.md)
- [Module 2 FEATURES.md](../module2-crossbar/FEATURES.md)
- [Module 3 FEATURES.md](../module3-mnist/FEATURES.md)
- [Module 4 FEATURES.md](../module4-circuits/FEATURES.md)
- [Module 5 FEATURES.md](../module5-comparison/FEATURES.md)
- [Module 6 FEATURES.md](../module6-eda/FEATURES.md)

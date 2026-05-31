# Verification vs Validation in FeCIM Lattice Tools

## Core distinction

- **Verification** = _Are we solving the equations/code correctly?_  
  Evidence: unit tests, regression tests, golden files, deterministic replay checks.
- **Validation** = _Do the equations/models match physical reality?_  
  Evidence: calibration to measured data, literature comparisons, device-to-device agreement.

In short:
- Verification checks software correctness.
- Validation checks physical truthfulness.

## How this maps to this repo

### Module 1 (Hysteresis / Ferroelectric Physics)
- **Verification status:** Strong. Extensive tests for Preisach/LK behavior, ISPP sequencing, convergence, and regressions.
- **Validation status:** Partial-to-moderate. Material presets and calibration workflows are present; some presets are literature-informed, but not all paths are tied to direct silicon measurement in this repo.

### Module 2 (Crossbar)
- **Verification status:** Moderate. Algorithmic flows and data handling are testable and covered.
- **Validation status:** Limited. Primarily architectural/simulation behavior; depends on Module 1+4 model realism.

### Module 3 (MNIST / Inference)
- **Verification status:** Moderate-to-strong. Inference and pipeline logic are test-driven.
- **Validation status:** Limited. Accuracy trends are simulation-based and depend on upstream model fidelity.

### Module 4 (Circuits / Array + Peripheral)
- **Verification status:** Strong. Sense-chain, coupling tiers, ISPP controller behavior, and solver consistency are tested.
- **Validation status:** Partial. Contains physically motivated constraints and literature-aware defaults, but many operating bounds remain estimated unless calibrated to measured silicon.

### Module 5 (Comparison)
- **Verification status:** Moderate. Comparative calculations/aggregation are testable.
- **Validation status:** Derived. Inherits validity from upstream modules and datasets.

### Module 6 (EDA Export)
- **Verification status:** Strong for format/tool flow checks (generation, syntax, cross-checks).
- **Validation status:** Low for signoff metrics. Exported timing/power are model-derived estimates, not foundry signoff characterization.

### Shared / Validation packages
- **Verification status:** Strong intent and growing coverage via shared tests and golden references.
- **Validation status:** Depends on which source data each workflow uses (measured vs estimated).

## Practical policy

1. Treat passing tests as **verification evidence**, not physical proof.
2. Treat literature/measurement agreement as **validation evidence**.
3. Any estimated characterization (especially Module 6 exports) must be labeled clearly for non-signoff use.
4. For production-grade signoff, use measured-device characterization flows (e.g., Liberate/SiliconSmart with silicon/SPICE post-layout data).

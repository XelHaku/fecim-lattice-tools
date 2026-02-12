# Module 1: Hysteresis - Physics

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Prerequisites

- Electric field and polarization basics
- Units and dimensional analysis
- Simple integrals and summations

## Core Model

- Hysteresis is modeled as history-dependent polarization versus electric field.
- The Preisach model represents the material as a weighted sum of bistable elements.
- Each element switches at thresholds and contributes to total polarization.

## Key Equations (Simplified)

```
P(E) = ∬ μ(α,β) γ_{α,β}(E) dα dβ
γ_{α,β}(E) ∈ { -1, +1 }
Pr = P(E = 0) after saturation
Ec ≈ field where P crosses 0 on the major loop
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| E | Electric field | V/m (or MV/cm) |
| P | Polarization | C/m^2 (or uC/cm^2) |
| Ec | Coercive field | V/m |
| Pr | Remanent polarization | C/m^2 |
| α | Hysteron up-switch threshold | V/m |
| β | Hysteron down-switch threshold | V/m |
| μ(α,β) | Preisach density/weighting (normalized in code) | scaled to yield P |
| γ_{α,β} | Hysteron state (+1 or -1) | unitless |

## Assumptions And Limits

- Quasi-static loops, no high-frequency domain dynamics.
- Uniform material properties, no spatial gradients.
- Preisach elements are idealized, not microstructural.
- Parameters are simulation defaults unless cited; do not treat numeric values as measured device data.

## Where It Lives In Code

- `module1-hysteresis/pkg/ferroelectric/preisach.go`
- `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go`
- `module1-hysteresis/pkg/ferroelectric/material.go`
- `module1-hysteresis/pkg/gui/embedded.go`

## Sources

- `docs/video-transcripts/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md`
- `docs/research-papers/by-topic/01-ferroelectric-materials/`
- `docs/development/scriptReference.md#demo-1-hysteresis-module1-hysteresis`
- HZO P-E loop examples (DOI: (add))

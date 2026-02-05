# Preisach Hysteresis Model (Module 1)

**Scope**
This document summarizes the Preisach hysteresis model, key literature, and how it is implemented in this repository. It is scoped to the **scalar, rate‑independent** Preisach model used in the module 1 hysteresis simulation.

**What Preisach Is (Short Definition)**
The Preisach model represents a hysteretic material as a superposition of many elementary relay operators (hysterons) with different switching thresholds. A distribution over these thresholds encodes the material’s memory.

**Why It Matters for Ferroelectrics**
Preisach‑type models are widely used in ferroelectrics and other ferroic systems because they capture history‑dependent polarization and minor loops with a compact, physically interpretable structure.

**Key Literature (Foundational + Ferroelectric‑Relevant)**
- F. Preisach (1935): Original hysteresis formulation in magnetism. Often cited as the origin of the Preisach model.
- I. D. Mayergoyz (1986, 1991): Formal theory and identification of Preisach models, including conditions for applicability.
- C. R. Pike, A. P. Roberts, K. L. Verosub (1999): First‑Order Reversal Curve (FORC) method for extracting Preisach distributions.
- R. D. Groot et al. (2018, Nature Communications): Direct experimental evidence of Preisach distributions in ferroelectrics.
- A. Sutor, S. Rupitsch, R. Lerch (2010): Preisach‑based model validated on ferroelectric hysteresis.
- F. Wolf et al. (2011): Dynamic and rate‑dependent extensions of Preisach for ferroelectric actuators.

**Core Concepts (Theory Summary)**
- **Hysterons**: Ideal relays with upper/lower thresholds (α, β). Each hysteron outputs ±1 depending on the past input path.
- **Preisach Plane**: The (α, β) half‑plane (α ≥ β). A density μ(α,β) weights the contribution of each hysteron.
- **Output**: The model output is the weighted sum/integral of hysterons. In continuous form, the output is an integral over the Preisach plane.
- **Wipe‑Out Property**: New extreme inputs erase prior smaller loops, leaving only a reduced set of turning points.
- **Congruency Property**: Minor loops of the same amplitude are geometrically congruent.
- **Everett Function**: A two‑argument function derived from μ(α,β) used to compute polarization on minor loops efficiently. See general Preisach reviews for definitions and usage.

**Validity Conditions**
Preisach‑type models are appropriate for systems exhibiting **wipe‑out** and **congruency**, which are necessary/sufficient conditions for scalar Preisach behavior in the classical theory (see Mayergoyz 1986).

**Identification & Measurement**
- **FORC** is a practical method to infer the Preisach distribution from measured hysteresis data.
- Everett functions or Preisach densities can be fit to major loops and a family of minor loops.

**Ferroelectric‑Specific Notes**
- Preisach distributions have been experimentally extracted in ferroelectrics, linking macroscopic hysteresis to microscopic switching kinetics.
- Classical Preisach is **rate‑independent**; dynamic ferroelectric behavior often requires rate‑dependent or “viscous” extensions.

**Our Implementation (Code‑Level Behavior)**

**Key Files**
- `shared/physics/preisach.go`
- `module1-hysteresis/pkg/ferroelectric/preisach.go`
- `shared/physics/material.go`
- `module1-hysteresis/pkg/ferroelectric/level_bins.go` (MLC guard‑band binning)

**Algorithmic Structure**
- A **turning‑point stack** stores input extrema; direction changes push turning points.
- The **wipe‑out** rule removes nested loops when a new extreme is reached.
- Polarization is computed by summing Everett contributions over the stack geometry.

**Everett Function Used**
- The Everett function is implemented with a **tanh‑based surrogate** to reproduce a smooth S‑shaped major loop.
- The distribution width `Delta` is **tuned per material** so the remanent polarization matches `Pr` (fallback: `0.25 * Ec`), and the saturation field is `E_sat = 5 * Ec`.

**Reversible (Nonlinear) Relaxation**
- We add a **nonlinear reversible component** derived from permittivity (no tuning knobs):
  - Linear susceptibility: `χ = ε0(εr − 1)`
  - Saturating reversible polarization: `P_rev_sat = χ * Ec`
  - Reversible curve: `P_rev(E) = P_rev_sat * tanh(E / Ec)`
- To preserve total saturation **Ps**, the irreversible saturation is reduced to:
  - `Ps_irrev = Ps − P_rev_sat`
- This yields curved relaxation when E decreases, without manual parameters.

**MLC Level Binning + Guard Band (Read Margin)**
- We model discrete MLC levels as **bins with guard bands** rather than a hard round-to-nearest only.
- Levels are **evenly spaced in polarization**, not in electric field, and use an **effective** saturation range.
- `effectivePs = Ps * rangeFrac`.
- `rangeFrac` comes from `TargetRangeFrac` in the material (defaults to `0.98` via `module1-hysteresis/pkg/gui/gui.go`).
- Spacing: `step = 2 * effectivePs / (NumLevels - 1)` (`module1-hysteresis/pkg/ferroelectric/level_bins.go`).
- Level 1 center: `-effectivePs`.
- Level N center: `+effectivePs`.
- GUI level mapping (used for `a.discreteLevel`) normalizes by the effective range.
- `levelNorm = clamp(P / effectivePs, -1, 1)`.
- `level_index = round((levelNorm + 1) / 2 * (NumLevels - 1))`.
- Stored as 0-based `level_index`; logged as both `level_index` and `level = level_index + 1` (`module1-hysteresis/pkg/gui/physics_engine.go`, `module1-hysteresis/pkg/gui/data_logger.go`).
- Each bin has a **guard fraction** (default `0.15`) that marks edge regions as unsafe for reliable read.
- The guard band **does not change bin width**; it only shrinks the "safe" region used during verify (`module1-hysteresis/pkg/ferroelectric/level_bins.go`).
- Implementation: `module1-hysteresis/pkg/ferroelectric/level_bins.go` (`LevelBins.LevelForP`).

**Write‑Verify Behavior with Guard Bands**
- During the ISPP write‑verify loop, the simulation treats "in-guard" reads as **not yet valid** even if the level index matches the target.
- The controller receives a guard‑direction hint and applies a small correction pulse toward the bin center.
- This models real MLC behavior where the sense amplifier rejects edge states and forces extra program/verify iterations.

**Short Example (30 Levels)**
Assume `Ps = 0.55 C/m²`, `N = 30`, `GuardFrac = 0.15`, `rangeFrac = 0.98`.

*Example values are illustrative; cite before external use (DOI: (add)).*

```text
effectivePs = Ps * rangeFrac = 0.55 * 0.98 = 0.539 C/m²
Step = 2*effectivePs/(N-1) = 2*0.539/29 ≈ 0.03717 C/m²
Guard width = 0.15*Step ≈ 0.00558 C/m²

Level k center: Pk = -effectivePs + (k-1)*Step
Safe zone: [Pk - 0.5*Step + Guard, Pk + 0.5*Step - Guard]
```

If a read lands inside the guard zone, the write-verify loop treats it as "not yet valid"
and nudges the state toward the bin center.

**Guard Band Diagram (One Bin)**

```text
|---edge guard---|======= safe =======|---edge guard---|
Pk - 0.5*Step    Pk                   Pk + 0.5*Step
```

**Tuning Note**
If you want the outer levels closer to full saturation, increase `target_range_frac` in `config/materials.yaml`
(e.g., `literature_superlattice.target_range_frac`).

**Temperature & Stress Effects**
- Effective Ec and Ps are updated via linear scaling in the material model.
- After temperature/stress updates, reversible parameters are recomputed and the Everett saturation is adjusted accordingly.

**What This Implementation Does Not Model**
- Rate dependence and switching kinetics (quasi‑static only).
- Spatial domain structure or domain wall dynamics.
- Multi‑axis coupling (stress‑electric interactions beyond heuristic scaling).
- Stochastic switching distributions from device‑level physics.

**Numerical Validation in Repo**
- Golden hysteresis loop regression test files validate stability of outputs.
- Tests check loop symmetry, closure, and basic remanence behavior.

**Where to Extend**
- Replace tanh Everett with an identified μ(α,β) from measured loops.
- Add rate‑dependent Preisach or viscous/dynamic operators for frequency effects.
- Couple Preisach with a Landau equilibrium solver for the anhysteretic curve.

## References (External)

1. Preisach, F. “Über die magnetische Nachwirkung.” Zeitschrift für Physik 94, 277–302 (1935).
2. Mayergoyz, I. D. “Mathematical models of hysteresis.” Phys. Rev. Lett. 56, 1518 (1986). DOI: 10.1103/PhysRevLett.56.1518
3. Mayergoyz, I. D. Mathematical Models of Hysteresis. Springer (1991). DOI: 10.1007/978-1-4612-3028-1
4. Pike, C. R., Roberts, A. P., Verosub, K. L. “Characterizing interactions in fine magnetic particle systems using first order reversal curves.” J. Appl. Phys. 85, 6660–6667 (1999). DOI: 10.1063/1.370176
5. Groot, R. D. et al. “Physical reality of the Preisach model for organic ferroelectrics.” Nat. Commun. 9, 3805 (2018). DOI: 10.1038/s41467-018-06717-w
6. Sutor, A., Rupitsch, S., Lerch, R. “A Preisach‑based hysteresis model for magnetic and ferroelectric hysteresis.” Appl. Phys. A 100, 425–430 (2010). DOI: 10.1007/s00339-010-5884-9
7. Wolf, F. et al. “Modeling and measurement of creep‑ and rate‑dependent hysteresis in ferroelectric actuators.” Sensors and Actuators A: Physical 172(1), 245–252 (2011). DOI: 10.1016/j.sna.2011.02.026
8. Besa, V., et al. “Review of the Preisach Model and Its Applications.” Micromachines 14(1), 177 (2023). DOI: 10.3390/mi14010177

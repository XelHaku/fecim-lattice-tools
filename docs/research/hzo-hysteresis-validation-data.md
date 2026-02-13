# HZO Hysteresis Loop Validation — Experimental Data Sources & Calibration Plan

**Purpose:** Anchor the fecim-lattice-tools Preisach + Landau-Khalatnikov hysteresis simulator to real experimental data.
**Date:** 2026-02-13
**Priority:** This is the #1 validation gap. Education layer depends on correct physics.

---

## 1. Published Experimental Parameter Values for HZO (Hf0.5Zr0.5O2)

### Compiled from peer-reviewed literature:

| Parameter | Symbol | Range | Typical (10nm TiN/HZO/TiN) | Unit | Key Sources |
|-----------|--------|-------|-----------------------------|------|-------------|
| Remnant Polarization | Pr | 5–30 | 14–24 | µC/cm² | [1][2][3][5] |
| Coercive Field (+) | Ec+ | 0.8–2.0 | 1.0–1.9 | MV/cm | [1][2][3][5] |
| Coercive Field (-) | Ec- | -0.8 to -2.0 | -0.85 to -1.4 | MV/cm | [1][2][5] |
| Saturation Polarization | Ps | 10–35 | 14–25 | µC/cm² | [1][2] |
| Film Thickness | t | 5–20 | 10 | nm | [1][2][3] |
| Relative Permittivity | εr | 25–40 | 33 | — | [1] |
| Endurance | cycles | 10⁹–10¹¹ | 10¹⁰ | cycles | [3][4] |
| Retention | time | >10 years (extrapolated) | 10⁵ s measured | s | [3] |
| Polarization Offset | Poffset | 0–2 | ~0.5 | µC/cm² | [1] |

### Specific Measured Values (HIGH CONFIDENCE — directly from papers):

**From [1] (Nature Scientific Reports 2021, 10nm HZO, 5 minor loops fitted):**
- Ps = 14.66 µC/cm²
- Pr = 14.55 µC/cm²
- Ec+ = 1.96 MV/cm
- Ec- = -0.85 MV/cm
- Poffset = -0.59 µC/cm²
- εFE = 33.79
- **This paper provides the exact Preisach model fitting methodology we should replicate.**

**From [5] (Nature Communications 2020, TiN/HZO/TiN ferroelectric diode):**
- Pr ≈ 17 µC/cm²
- Ec ≈ -1.4 / +1.9 MV/cm (asymmetric)
- After 400°C rapid thermal annealing

**From [6] (unlab.tech 2025, 20nm HZO programmable antenna):**
- Pr = 24 µC/cm²
- Consistent with published values for 20nm-thick HZO

---

## 2. Key Papers (Priority-Ranked for Calibration)

### PRIORITY 1 — Use this paper's data and method:

**[1] "Extraction of Preisach model parameters for fluorite-structure ferroelectrics and antiferroelectrics"**
- Authors: (UC Berkeley group)
- Journal: Nature Scientific Reports, 2021
- DOI: 10.1038/s41598-021-91492-w
- URL: https://www.nature.com/articles/s41598-021-91492-w
- **WHY:** This paper does EXACTLY what we need:
  - Uses the same tanh-based Preisach model: P_sat(E) = Ps·tanh(s·(E - Ec)) + Poffset
  - Provides the fitting algorithm (interior-point minimization of squared error)
  - Extracts 6 parameters: Ps, Pr, Ec+, Ec-, Poffset, εFE
  - Tests on 10nm HZO with real experimental data
  - Publishes exact fitted values (see table above)
  - Also handles antiferroelectric double-hysteresis loops
  - **The Preisach model equations (Eqs 1-5) map directly to our TanhEverett implementation**

### PRIORITY 2 — Foundational HZO papers:

**[2] Böscke, T.S., Müller, J., Bräuhaus, D., Schröder, U. & Böttger, U.**
- "Ferroelectricity in Hafnium Oxide Thin Films"
- Applied Physics Letters 99, 102903 (2011)
- DOI: 10.1063/1.3634052
- **WHY:** Original discovery paper. Baseline Pr/Ec values for Si-doped HfO2.

**[3] Trentzsch, M. et al.**
- "A 28nm HKMG super low power embedded NVM technology based on ferroelectric FETs"
- IEDM 2016
- **WHY:** Production-relevant 28nm FeFET. Published timing (50ns write, 5ns read). Already referenced in our Module 6 liberty export.

**[4] Park, M.H. et al.**
- Various papers on HZO characterization (thickness series, wake-up, fatigue)
- Multiple journals, 2015-2020
- **WHY:** Comprehensive thickness-dependent data. Shows how Pr/Ec change with film thickness (5.5nm to 20nm). Critical for validating thickness scaling in our model.

### PRIORITY 3 — Additional validation:

**[5] "A highly CMOS compatible hafnia-based ferroelectric diode"**
- Nature Communications, 2020
- DOI: 10.1038/s41467-020-15159-2
- URL: https://www.nature.com/articles/s41467-020-15159-2
- **WHY:** Publishes Pr ≈ 17 µC/cm², Ec ≈ -1.4/+1.9 MV/cm for TiN/HZO/TiN.

**[6] "Metastable ferroelectricity driven by depolarization fields in ultrathin Hf0.5Zr0.5O2"**
- Communications Physics, 2022
- DOI: 10.1038/s42005-022-00951-x
- URL: https://www.nature.com/articles/s42005-022-00951-x
- **WHY:** Uses Landau-Devonshire theory to explain pinched loops in ultrathin HZO. Relevant for our L-K model validation.

---

## 3. Open Datasets

### Status: No raw P-E loop datasets found in public repositories.

Searched: GitHub, Zenodo, Materials Project, NIST Materials Data Repository.

**Reality:** Ferroelectric P-E loop data is almost never published as raw CSV/JSON. It lives in paper figures. This means we MUST digitize from published figures.

### Digitization Tools:

| Tool | URL | Cost | Notes |
|------|-----|------|-------|
| WebPlotDigitizer | https://automeris.io | Free (web) | Best for P-E loops. Handles XY scatter well. |
| Engauge Digitizer | https://github.com/markummitchell/engauge-digitizer | Free (desktop) | Open source, good for batch work. |
| Energent.ai | https://energent.ai | Paid | AI-assisted, fastest but costs money. |

**Recommended:** WebPlotDigitizer (free, web-based, well-documented).

### Best Figures to Digitize:

1. **[1] Figure 4a** — 10nm HZO, 5 minor loops + reconstructed saturated loop. This is the PRIMARY calibration target because the paper also publishes the exact fitted parameters, so we can verify our fitting matches theirs.

2. **[5] Figure 1a** — TiN/HZO/TiN P-V loop after RTA. Clean single saturated loop.

3. **[4] Park et al. thickness series** — Multiple loops at different thicknesses. Critical for validating thickness scaling.

**Expected digitization accuracy:** ±1-2% on well-printed figures. P-E loops have smooth curves, which digitize well.

---

## 4. Calibration Strategy

### Step-by-Step Plan:

#### Phase 1: Reproduce Paper [1]'s Results (1-2 days)

1. **Digitize Figure 4a from [1]** using WebPlotDigitizer
   - Extract 5 minor loop traces as CSV (E in MV/cm, Q in µC/cm²)
   - Save to `validation/data/hzo_10nm_berkeley_loops.csv`

2. **Implement the fitting algorithm from [1]:**
   - The paper uses: P_sat(E) = Ps·tanh(s·(E - Ec)) + Poffset
   - This is essentially our `TanhEverett` in `module1-hysteresis/pkg/ferroelectric/preisach.go`
   - Add Q_FE(E) = P_FE(E) + ε₀·εFE·E (include linear dielectric term)
   - Minimize squared error using Levenberg-Marquardt or interior-point

3. **Verify our fitted parameters match theirs:**
   - Target: Ps = 14.66, Pr = 14.55, Ec+ = 1.96, Ec- = -0.85, εFE = 33.79
   - Acceptable: within 5% relative error on each parameter

4. **Produce "Simulation vs Experiment" overlay plot data:**
   - Our simulated P-E loop overlaid on digitized experimental data
   - Report RMSE in µC/cm² and relative error on Pr, Ec

#### Phase 2: Validate Against Additional Data (2-3 days)

5. **Digitize [5] Figure 1a** (TiN/HZO/TiN single loop)
   - Fit our model to this independent dataset
   - Compare extracted Pr, Ec to paper's reported values

6. **Thickness scaling validation:**
   - Digitize Park et al. loops at 5.5nm, 9.2nm, 14nm, 20nm
   - Fit each, plot Pr vs thickness and Ec vs thickness
   - Compare trend to published thickness-dependence studies

#### Phase 3: Integration into Tool (1-2 days)

7. **Add `validation/calibration/` package:**
   - `hzo_calibration.go` — loads experimental CSV, runs fitting, reports error metrics
   - `hzo_calibration_test.go` — regression test that verifies fit quality stays within tolerance
   - Golden files for each calibrated dataset

8. **Add to Module 1 GUI:**
   - "Load Experimental Data" button (CSV import)
   - Overlay experimental points on simulated P-E curve
   - Display fit quality metrics (RMSE, R², relative error on Pr/Ec)

### Fitting Parameters:

| Parameter | What to Fit | Initial Guess | Bounds |
|-----------|-------------|---------------|--------|
| Ps | Saturation polarization | 0.9 × max(Q) | [1, 50] µC/cm² |
| Pr | Remnant polarization | Ps - 1 | [0.5, Ps] µC/cm² |
| Ec+ | Forward coercive field | 0.5 × max(E) | [0.1, 5] MV/cm |
| Ec- | Backward coercive field | 0.5 × min(E) | [-5, -0.1] MV/cm |
| Poffset | Polarization offset | (max(Q)+min(Q))/2 | [-5, 5] µC/cm² |
| εFE | Linear permittivity | dQ/dE at max(E) / ε₀ | [10, 100] |

### Validation Metrics:

| Metric | Target | How to Compute |
|--------|--------|----------------|
| RMSE on P-E loop | < 1.0 µC/cm² | sqrt(mean((P_sim - P_exp)²)) |
| Relative error on Pr | < 5% | abs(Pr_sim - Pr_exp) / Pr_exp |
| Relative error on Ec | < 10% | abs(Ec_sim - Ec_exp) / Ec_exp |
| R² | > 0.99 | 1 - SS_res/SS_tot |
| Minor loop prediction | Qualitative match | Visual overlay + RMSE per loop |

---

## 5. What This Unlocks

Once Phase 1 is complete:
- **Education:** Students can see REAL data overlaid on simulation. "This is what a real HZO film does. This is what our model predicts. Here's the error."
- **Research:** The simulator becomes anchored to reality. Claims about multi-level storage, endurance, etc. can be referenced back to calibrated parameters.
- **Credibility:** Any reviewer or PI who asks "where's your validation?" gets a clear answer with published data, fit quality, and error metrics.

This is the single highest-value task for fecim-lattice-tools right now.

---

## References

[1] "Extraction of Preisach model parameters for fluorite-structure ferroelectrics and antiferroelectrics", Scientific Reports 11, 12460 (2021). DOI: 10.1038/s41598-021-91492-w

[2] Böscke, T.S. et al., "Ferroelectricity in Hafnium Oxide Thin Films", Appl. Phys. Lett. 99, 102903 (2011). DOI: 10.1063/1.3634052

[3] Trentzsch, M. et al., "A 28nm HKMG super low power embedded NVM technology based on ferroelectric FETs", IEDM 2016.

[4] Park, M.H. et al., various HZO characterization papers (2015-2020).

[5] "A highly CMOS compatible hafnia-based ferroelectric diode", Nature Communications 11, 1391 (2020). DOI: 10.1038/s41467-020-15159-2

[6] "Metastable ferroelectricity driven by depolarization fields in ultrathin Hf0.5Zr0.5O2", Communications Physics 5, 178 (2022). DOI: 10.1038/s42005-022-00951-x

[7] "Identification of Preisach hysteresis model parameters using genetic algorithms", Journal of King Saud University - Science (2019).

[8] "A Preisach-based hysteresis model for magnetic and ferroelectric hysteresis", Applied Physics A 100, 375 (2010). DOI: 10.1007/s00339-010-5884-9

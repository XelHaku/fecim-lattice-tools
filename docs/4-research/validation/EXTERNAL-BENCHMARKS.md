# External Benchmarks: Literature Validation Reference

**Version:** 1.0 | **Date:** 2026-03-29 | **Status:** Active

This document records every published numeric benchmark that simulator defaults are compared against, the exact source reference, and the comparison outcome.

---

## Methodology

Each parameter is validated using **range-check**: if the simulator default falls within the published measured spread, the check passes. This confirms the simulator is not divergent from reality; it does not claim exact calibration.

Test suite: `validation/literature/external_benchmarks_test.go`

---

## Parameter Comparisons

### 1. Remnant Polarization (Pr) -- DefaultHZO

| Field | Value |
|-------|-------|
| **Simulator default** | 24.5 uC/cm^2 |
| **Literature range** | 15--34 uC/cm^2 |
| **Source** | Park et al., Adv. Mater. 27, 1811 (2015) |
| **DOI** | [10.1002/adma.201404531](https://doi.org/10.1002/adma.201404531) |
| **Reference location** | Table 1, Fig. 3 (10nm ALD Hf0.5Zr0.5O2, various doping/anneal) |
| **Outcome** | **PASS** -- 24.5 is within [15, 34] |
| **Action** | None required |

### 2. Coercive Field (Ec) -- DefaultHZO

| Field | Value |
|-------|-------|
| **Simulator default** | 1.2 MV/cm |
| **Literature range** | 0.8--1.5 MV/cm |
| **Source** | Park et al., Adv. Mater. 27, 1811 (2015) |
| **DOI** | [10.1002/adma.201404531](https://doi.org/10.1002/adma.201404531) |
| **Reference location** | Table 1, Fig. 3 |
| **Outcome** | **PASS** -- 1.2 is within [0.8, 1.5] |
| **Action** | None required |

### 3. Remnant Polarization (Pr) -- FeCIMMaterial

| Field | Value |
|-------|-------|
| **Simulator default** | 30 uC/cm^2 |
| **Literature range** | 15--40 uC/cm^2 |
| **Source (lower)** | Park et al., Adv. Mater. 27, 1811 (2015), DOI: [10.1002/adma.201404531](https://doi.org/10.1002/adma.201404531) |
| **Source (upper)** | Cheema et al., Nature 580, 478 (2020), DOI: [10.1038/s41586-020-2208-x](https://doi.org/10.1038/s41586-020-2208-x) |
| **Reference location** | Park Table 1; Cheema Fig. 2 (ultrathin stacks >40 uC/cm^2) |
| **Outcome** | **PASS** -- 30 is within [15, 40] |
| **Note** | FeCIM Pr is labeled ESTIMATED in source code. No FeCIM-specific Pr measurement is published. |
| **Action** | Update when FeCIM device Pr is published in peer-reviewed literature |

### 4. Endurance -- DefaultHZO

| Field | Value |
|-------|-------|
| **Simulator default** | 10^10 cycles |
| **Literature range** | 10^9--10^12 cycles |
| **Source (lower)** | Mueller et al., IEEE IEDM 2013, DOI: [10.1109/IEDM.2013.6724605](https://doi.org/10.1109/IEDM.2013.6724605) |
| **Source (upper)** | Science Advances 2024, DOI: [10.1126/sciadv.adp0174](https://doi.org/10.1126/sciadv.adp0174) |
| **Reference location** | Mueller: Fig. 4 (10nm HZO endurance); SciAdv: Fig. 3 (HZO+MoS2) |
| **Outcome** | **PASS** -- 10^10 is within [10^9, 10^12] |
| **Action** | None required |

### 5. Cycle-to-Cycle Variation (C2C) -- DefaultHZO

| Field | Value |
|-------|-------|
| **Simulator default** | 3% (C2CSigmaBase = 0.03) |
| **Literature range** | ~3--5% for 4-state FeFET |
| **Source** | Soliman et al., Nature Communications 14, 6348 (2023) |
| **DOI** | [10.1038/s41467-023-42110-y](https://doi.org/10.1038/s41467-023-42110-y) |
| **Reference location** | Fig. 2, Supplementary Fig. 5 (Vth sigma = 38--40 mV, 4 states) |
| **Outcome** | **PASS** -- 3% is within [1%, 8%] (broadened for multi-level arrays) |
| **Note** | Sigma is reported as Vth variation; conversion to conductance C2C is approximate |
| **Action** | None required; consider narrowing range when more multi-level C2C data is available |

### 6. LGD Coefficients (beta, gamma) -- DefaultHZO

| Field | Value |
|-------|-------|
| **Simulator default** | beta = -6.72e8 J*m^5/C^4, gamma = 1.95e10 J*m^9/C^6 |
| **Literature value** | beta = -6.72e8, gamma = 1.95e10 (exact) |
| **Source** | Materlik et al., J. Appl. Phys. 117, 134109 (2015) |
| **DOI** | [10.1063/1.4916229](https://doi.org/10.1063/1.4916229) |
| **Reference location** | Table I |
| **Outcome** | **PASS** -- exact match |
| **Action** | None required |

### 7. Remnant Polarization (Pr) -- MaterlikHfO2

| Field | Value |
|-------|-------|
| **Simulator default** | 20 uC/cm^2 |
| **Literature range** | 10--30 uC/cm^2 |
| **Source** | Materlik et al., J. Appl. Phys. 117, 134109 (2015) |
| **DOI** | [10.1063/1.4916229](https://doi.org/10.1063/1.4916229) |
| **Reference location** | Fig. 5, Table II (LGD-derived Pr varies with coefficients and thickness) |
| **Outcome** | **PASS** -- 20 is within [10, 30] |
| **Action** | None required |

---

## Parameters Without External Validation

The following simulator defaults **lack** direct external measurement comparisons:

| Parameter | Preset | Current Value | Reason |
|-----------|--------|---------------|--------|
| Retention time | DefaultHZO | 3.15e9 s (100 yr @ 85C) | Extrapolated, no single paper covers 100-year retention |
| Imprint field | DefaultHZO | 1 MV/m | Educational default, no specific reference |
| NLS parameters (Tau0NLS, EaNLS, NLSSigma) | DefaultHZO | Various | Based on Guo et al., APL 2018 (NLSSigma); others are typical values |
| Series resistance | All presets | 50 Ohm | Simulation default, device-specific in practice |
| Viscosity (rho) | DefaultHZO | 0.05 | Range 0.005--0.05 per Alessandri et al., IEEE EDL 2018; not validated against specific device |
| Electrostriction (Q11, Q12) | DefaultHZO | 0.089, -0.026 | Q12 has known DFT discrepancy (see struct comment) |
| NumLevels | FeCIMMaterial | 30 | Conference claim (COSM 2025); pending peer review |

---

## Update Procedure

1. When a new paper is published with measured data for a parameter we use, add a row to this document.
2. Add a corresponding test case in `external_benchmarks_test.go`.
3. If the simulator default falls outside the new measured range, open an issue and decide whether to update the default or document the divergence.
4. Update the honesty audit (`docs/4-research/honesty-audit.md`) accordingly.

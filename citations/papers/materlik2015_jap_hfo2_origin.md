# The Origin of Ferroelectricity in Hf1-xZrxO2: A Computational Investigation and a Surface Energy Model

**Key:** `materlik2015_jap_hfo2_origin`
**DOI:** `10.1063/1.4916229`
**Year:** 2015
**Venue:** Journal of Applied Physics, 117, 134109
**Authors:** Materlik, R. and Kuenneth, C. and Kersch, A.
**Tags:** `#hfo2` `#dft` `#lgd` `#landau` `#ferroelectric` `#core`
**Status:** `deep-read`
**PDF:** `not stored`
**Added:** 2026-05-04

---

## TL;DR

DFT calculations explaining why Hf1-xZrxO2 is ferroelectric. Provided Landau-Ginzburg-Devonshire (LGD) coefficients: β = -6.72e8 J·m⁵/C⁴, γ = 1.95e10 J·m⁹/C⁶. These coefficients drive the Landau-Khalatnikov solver in FeCIM.

## Why It Matters For FeCIM Lattice Tools

Provides the LGD coefficients (β, γ) used in `shared/physics/landau.go` LKSolver. The temperature-dependent reference points (200-500K) calibrate the Curie-Weiss model. Without this paper, the LK solver has no physics basis for HZO switching dynamics.

## Key Facts

### Landau Coefficients (Orthorhombic Pca21 HfO2)

- **β (LGD first-order nonlinearity)** = `-6.720e8` `J·m⁵/C⁴`
  - **Source location:** Table I, computational results
  - **Conditions:** 0 K DFT ground state, orthorhombic Pca21 phase
  - **Evidence level:** peer-reviewed
  - **Notes:** Negative β indicates first-order phase transition. Used directly in DefaultHZO, FeCIMMaterial, and MaterlikHfO2 presets.

- **γ (LGD stability coefficient)** = `1.950e10` `J·m⁹/C⁶`
  - **Source location:** Table I
  - **Conditions:** Same as β
  - **Evidence level:** peer-reviewed
  - **Notes:** Positive γ ensures thermodynamic stability at large P. Used in all HZO presets.

- **Curie Temperature** = `598` `K`
  - **Source location:** Computational estimate
  - **Conditions:** Pca21 orthorhombic phase
  - **Evidence level:** peer-reviewed
  - **Notes:** FeCIM uses Tc=723K for DefaultHZO (Park 2015 experimental midpoint). The discrepancy reflects computational vs experimental Tc.

- **Curie Constant** = `5.3e5` `K`
  - **Source location:** Derived from computational LGD
  - **Conditions:** Same as above
  - **Evidence level:** peer-reviewed

### Temperature-Dependent LGD

- **α(T) reference points at 200K, 300K, 400K, 500K**
  - **Source location:** `shared/physics/temperature_calibration.go`
  - **Evidence level:** peer-reviewed (via Materlik data)
  - **Notes:** Used in `HZO10nmTemperatureCalibration()` to validate Curie-Weiss model.

## Methodology

Density functional theory (DFT) calculations using VASP. Pca21 orthorhombic phase identified as the ferroelectric phase. Surface energy model explains why the orthorhombic phase is stabilized in thin films despite being metastable in bulk.

## Limitations

- 0 K ground state; finite-temperature effects estimated but not directly computed
- Single composition (Hf0.5Zr0.5O2); doping effects not modeled
- No explicit modeling of grain boundaries, oxygen vacancies, or electrode interfaces

## Cited In

- [ ] `shared/physics/landau.go:139-141` - LKSolver defaults (β, γ)
- [ ] `shared/physics/material.go` - DefaultHZO preset
- [ ] `shared/physics/temperature_calibration.go` - Temperature validation points
- [ ] `shared/viewmodel/hysteresis/viewmodel.go` - LK solver integration

## Related Sources

- `park2015_advmat_hzo` - Experimental confirmation of ferroelectricity in HZO
- `guo2018_apl_nls` - NLS switching using Materlik LGD parameters

## Verification

FeCIM LK solver uses β=-6.72e8 and γ=1.95e10 directly from this paper. The LK04 mitigation (ConfigureFromMaterial) adjusts α to match experimental Pr at the cost of internal Ec consistency. See `PHYSICS_REALISM_AUDIT.md` for known solver limitations.

## Notes

There is a DOI discrepancy in the codebase: `10.1063/1.4916229` (paper/refs.bib, landau.go) vs `10.1063/1.4916707` (calibration_targets.go). Both resolve to the same J. Appl. Phys. article. The primary DOI (1.4916229) is used here.

---

## BibTeX

```bibtex
@article{materlik2015,
  author = {Materlik, R. and Kuenneth, C. and Kersch, A.},
  title = {The origin of ferroelectricity in Hf1-xZrxO2: A computational investigation and a surface energy model},
  journal = {Journal of Applied Physics},
  volume = {117},
  pages = {134109},
  year = {2015},
  doi = {10.1063/1.4916229}
}
```

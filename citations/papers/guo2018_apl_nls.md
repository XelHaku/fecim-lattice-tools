# Nucleation-Limited Polarization Switching in Ferroelectric HfO2 Capacitors

**Key:** `guo2018_apl_nls`
**DOI:** `10.1063/1.5038038`
**Year:** 2018
**Venue:** Applied Physics Letters, 112, 262903
**Authors:** Guo, R. and others
**Tags:** `#hfo2` `#nls` `#switching` `#merz` `#core`
**Status:** `read`
**PDF:** `not stored`
**Added:** 2026-05-04

---

## TL;DR

Extended nucleation-limited switching (NLS) model for ferroelectric HfO2 capacitors. Quantified switching-time distribution (log-normal, σ=1.5) and provided the τ_inf = 100 ps intrinsic switching time used in FeCIM.

## Why It Matters For FeCIM Lattice Tools

Provides the NLS statistical parameters (NLSSigma=1.5, TauInf=1e-10s) used as defaults in the LK solver. These are cited in `shared/physics/landau.go` as "HfO2 default (Guo et al., APL 112, 262903, 2018)". Critical for accurate stochastic switching simulation in ISPP write-verify and polydomain ensemble mode.

## Key Facts

### NLS Distribution

- **NLSSigma** = `1.5` (log-normal, dimensionless)
  - **Source location:** Switching time distribution fit
  - **Conditions:** HfO2 capacitors, room temperature
  - **Evidence level:** peer-reviewed
  - **Notes:** Determines the width of the switching-time distribution. Larger σ means wider spread in nucleation times. Used as LKSolver.NLSSigma default.

- **τ_inf** = `1.0e-10` `s` (100 ps)
  - **Source location:** Merz law extrapolation
  - **Conditions:** Infinite applied field limit
  - **Evidence level:** peer-reviewed
  - **Notes:** Intrinsic attempt time for polarization reversal. Used as LKSolver.TauInf default.

## Methodology

Pulse-switching measurements on HfO2-based ferroelectric capacitors. Switching time extracted from transient current. NLS model (Merz law: τ(E) = τ_inf · exp(Ea/E)) fit to field-dependent switching data. Log-normal distribution model for nucleation site statistics.

## Limitations

- Single device type (planar HfO2 capacitor)
- Room temperature only
- Switching time statistics assume independent nucleation sites (no grain-grain interaction)

## Cited In

- [ ] `shared/physics/landau.go:156-157` - NLSSigma and TauInf defaults
- [ ] `shared/physics/nls.go` - NLS model parameters
- [ ] `AGENTS.md` - Three-tier system reference

## Related Sources

- `alessandri2018_ieee_edl_switching` - Earlier NLS characterization of HZO
- `materlik2015_jap_hfo2_origin` - LGD coefficients for switching energetics

## Verification

FeCIM uses NLSSigma=1.5 and TauInf=1.0e-10s as LK solver defaults. These are directly cited from Guo 2018. The NLS model is enabled by default (LKSolver.UseNLS = true) and can be toggled per-material.

---

## BibTeX

```bibtex
@article{guo2018,
  author = {Guo, R. and others},
  title = {Nucleation-limited polarization switching in ferroelectric HfO2 capacitors},
  journal = {Applied Physics Letters},
  volume = {112},
  pages = {262903},
  year = {2018},
  doi = {10.1063/1.5038038}
}
```

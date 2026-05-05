# Switching Dynamics of Ferroelectric HZO

**Key:** `alessandri2018_ieee_edl_switching`
**DOI:** `10.1109/LED.2018.2872124`
**Year:** 2018
**Venue:** IEEE Electron Device Letters, 39(11), 1780-1783
**Authors:** Alessandri, A. and others
**Tags:** `#hzo` `#switching` `#nls` `#dynamics` `#core`
**Status:** `read`
**PDF:** `not stored`
**Added:** 2026-05-04

---

## TL;DR

Characterized polarization switching dynamics in HZO ferroelectric capacitors using nucleation-limited switching (NLS) model. Provided the foundational switching time parameters used in FeCIM's LK solver NLS mode.

## Why It Matters For FeCIM Lattice Tools

Provides the NLS (Nucleation-Limited Switching) parameters for the Landau-Khalatnikov solver: activation field, intrinsic switching time, and switching time statistics. Used in `shared/physics/landau.go` and `shared/physics/nls.go`. Critical for accurate ISPP write timing in Module 1 and Module 4.

## Key Facts

### NLS Parameters

- **Activation Field (Ea)** = `1.9` `MV/cm` (Merz law)
  - **Source location:** NLS model fit
  - **Conditions:** HZO capacitor, room temperature
  - **Evidence level:** peer-reviewed
  - **Notes:** FeCIM default is 19 MV/cm in SI (LKSolver.ActivationField = 1.9e9 V/m). This is the Merz activation field for nucleation-limited switching.

- **Intrinsic Switching Time (τ_inf)** = `~100` `ps`
  - **Source location:** NLS model fit
  - **Conditions:** Infinite field extrapolation
  - **Evidence level:** peer-reviewed
  - **Notes:** FeCIM uses TauInf = 1.0e-10 s (100 ps). Cited with Guo et al. APL 2018.

- **Switching-time dispersion (σ)** = `1.5` (log-normal, dimensionless)
  - **Source location:** Measured switching time statistics
  - **Conditions:** HZO at room temperature
  - **Evidence level:** peer-reviewed
  - **Notes:** Used in LKSolver.NLSSigma. From Guo et al. APL 112, 262903 (2018), who built on Alessandri's framework.

## Methodology

Pulse-switching measurements on TiN/HZO/TiN capacitors. NLS model (Merz law) fit to measured switching times vs applied field. Log-normal distribution of switching times attributed to nucleation site statistics.

## Limitations

- Single device geometry (planar capacitor)
- Room temperature only (temperature-dependent NLS not characterized here)
- No endurance-dependent switching dynamics

## Cited In

- [ ] `shared/physics/landau.go:154-158` - NLS defaults
- [ ] `shared/physics/nls.go` - NLS model implementation
- [ ] `docs/3-develop/known-limitations.md` - Cited as NLS source

## Related Sources

- `guo2018_apl_nls` - Extended NLS characterization
- `materlik2015_jap_hfo2_origin` - LGD coefficients for the switching model

## Verification

FeCIM uses ActivationField=1.9e9 V/m, TauInf=1.0e-10s, NLSSigma=1.5. These are directly from the Alessandri/Guo NLS framework. The values are used as solver defaults and can be overridden per-material.

## Notes

The three-tier AGENTS.md system references Alessandri 2018 along with Materlik 2015, Park 2015, and Guo 2018 as the four foundational papers for HZO physics. The ρ (viscosity) parameter range reported in this paper is used for the DefaultHZO RhoViscosity but the specific value (0.05 Ohm·m) is the FeCIM default, not directly from this paper.

---

## BibTeX

```bibtex
@article{alessandri2018,
  author = {Alessandri, A. and others},
  title = {Switching Dynamics of Ferroelectric HZO},
  journal = {IEEE Electron Device Letters},
  volume = {39},
  number = {11},
  pages = {1780--1783},
  year = {2018},
  doi = {10.1109/LED.2018.2872124}
}
```

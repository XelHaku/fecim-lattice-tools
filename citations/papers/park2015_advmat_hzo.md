# Ferroelectricity and Antiferroelectricity of Doped Thin HfO2-Based Films

**Key:** `park2015_advmat_hzo`
**DOI:** `10.1002/adma.201404531`
**Year:** 2015
**Venue:** Advanced Materials, 27(11), 1811-1831
**Authors:** Park, M. H. and others
**Tags:** `#hzo` `#ferroelectric` `#pe-loop` `#pr` `#ec` `#core`
**Status:** `deep-read`
**PDF:** `not stored`
**Added:** 2026-05-04

---

## TL;DR

Confirmed ferroelectricity in Si-doped Hf0.5Zr0.5O2 thin films (10 nm). Reported Pr = 24 µC/cm² and Ec = 1.0 MV/cm at room temperature. This is the foundational paper for FeCIM's DefaultHZO material preset.

## Why It Matters For FeCIM Lattice Tools

Provides the core material parameters for the DefaultHZO preset used in Module 1 (Hysteresis) and Module 2 (Crossbar) simulations. Pr and Ec values anchor the Preisach and Landau-Khalatnikov solvers. The Fig 2a P-E loop is digitized as `experimental-data/hzo/pe-loops/park2015_advmat_hzo_10nm_fig2a.json`.

## Key Facts

### Ferroelectric Properties (HZO 10nm, Si-doped)

- **Pr (Remanent Polarization)** = `24` `µC/cm²`
  - **Source location:** Fig. 2a
  - **Conditions:** 10 nm Hf0.5Zr0.5O2, TiN electrodes, Si-doped, room temperature, 1 kHz triangular wave
  - **Evidence level:** peer-reviewed
  - **Notes:** Baseline for FeCIM DefaultHZO preset. Literature range: 15-34 µC/cm² across doping/temperature.

- **Ec (Coercive Field)** = `1.0` `MV/cm`
  - **Source location:** Fig. 2a
  - **Conditions:** Same as Pr
  - **Evidence level:** peer-reviewed
  - **Notes:** Literature range: 0.8-1.5 MV/cm.

## Methodology

P-E hysteresis measured via Sawyer-Tower circuit on TiN/HZO/TiN capacitors. Temperature-dependent measurements from 80K to 400K. XRD confirms orthorhombic Pca21 phase. Antiferroelectric behavior observed at lower doping concentrations.

## Limitations

- Single doping concentration (Si); other dopants (Al, Gd, Y, La) produce different Pr/Ec ranges
- Thin film only (10 nm); thickness scaling not fully characterized here
- No endurance data beyond initial cycles

## Cited In

- [ ] `shared/physics/material.go` - DefaultHZO() preset
- [ ] `shared/viewmodel/hysteresis/snapshot.go` - material summary
- [ ] `validation/literature/external_benchmarks_test.go` - Pr/Ec range check

## Related Sources

- `materlik2015_jap_hfo2_origin` - Computational origin of HfO2 ferroelectricity, provides LGD coefficients
- `cheema2020_nature_hzo_superlattice` - HZO/ZrO2 superlattice achieving higher Pr
- `kim2020_materials_tin_hzo` - TiN/HZO wake-up behavior

## Verification

FeCIM DefaultHZO uses Pr=24.5 µC/cm² and Ec=1.2 MV/cm — within reported literature range. Digitized Fig 2a loop used for calibration validation in `validation/testdata/literature/park2015_fig2a_hzo_10nm.csv`.

## Notes

Park 2015 is the single most-cited paper in FeCIM Lattice Tools and the anchor for all HZO-based material presets. The reported Pr=24 µC/cm² and Ec=1.0 MV/cm should be used as the citation reference; FeCIM DefaultHZO values (24.5, 1.2) are conservative midpoints.

---

## BibTeX

```bibtex
@article{park2015,
  author = {Park, M. H. and others},
  title = {Ferroelectricity and Antiferroelectricity of Doped Thin HfO2-Based Films},
  journal = {Advanced Materials},
  volume = {27},
  number = {11},
  pages = {1811--1831},
  year = {2015},
  doi = {10.1002/adma.201404531}
}
```

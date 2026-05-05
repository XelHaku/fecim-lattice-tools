# Ueber die magnetische Nachwirkung

**Key:** `preisach1935_zphys`
**DOI:** `10.1007/BF01349418`
**Year:** 1935
**Venue:** Zeitschrift fuer Physik, 94, 277-302
**Authors:** Preisach, F.
**Tags:** `#preisach` `#hysteresis` `#foundational` `#core`
**Status:** `read`
**PDF:** `not stored`
**Added:** 2026-05-04

---

## TL;DR

The foundational paper that introduced the Preisach model of hysteresis. Describes a mathematical framework where hysteresis is decomposed into a distribution of elementary bistable units (hysterons) on the (α,β) half-plane.

## Why It Matters For FeCIM Lattice Tools

The Preisach model implemented in `shared/physics/preisach.go` and `module1-hysteresis/pkg/ferroelectric/preisach.go` is based directly on this work. The product-form Everett function, wipe-out property, and Preisach plane are all from Preisach 1935.

## Key Facts

### Preisach Model Definition

- **Preisach density** = `μ(α, β)` defined on half-plane `α > β`
  - **Source location:** Original text
  - **Evidence level:** peer-reviewed
  - **Notes:** FeCIM uses TanhEverett (sech² distribution) as the default density. The product-form Everett function ensures non-negative output.

- **Wipe-out property** = `deletion of minor loops upon crossing previous turning points`
  - **Source location:** Original description of hysteretic memory
  - **Evidence level:** peer-reviewed
  - **Notes:** Implemented in PreisachStack turning-point tracking logic.

## Methodology

Purely theoretical/mathematical. Defines the Preisach operator as an integral over elementary rectangular hysteresis loops (hysterons) with thresholds α (up-switching) and β (down-switching).

## Limitations

- Original paper is for magnetic hysteresis; application to ferroelectrics is an analogy
- No rate-dependence (purely quasi-static)
- No temperature dependence
- No physical mechanism (purely phenomenological)

## Cited In

- [ ] `shared/physics/preisach.go`
- [ ] `module1-hysteresis/pkg/ferroelectric/preisach.go`
- [ ] `shared/physics/tanh_everett.go`
- [ ] `docs/3-develop/known-limitations.md`

## Related Sources

- `mayergoyz2003_hysteresis_models` - Modern mathematical treatment of Preisach models

## Notes

The Preisach model was originally developed for magnetic materials but has been successfully applied to ferroelectrics. The key adaptation for FeCIM is using the Everett function (product-form) to compute polarization, calibrated to HZO material parameters.

---

## BibTeX

```bibtex
@article{preisach1935,
  author = {Preisach, F.},
  title = {Ueber die magnetische Nachwirkung},
  journal = {Zeitschrift fuer Physik},
  volume = {94},
  pages = {277--302},
  year = {1935},
  doi = {10.1007/BF01349418}
}
```

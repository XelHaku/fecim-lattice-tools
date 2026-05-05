# Mathematical Models of Hysteresis and Their Applications

**Key:** `mayergoyz2003_hysteresis_models`
**Year:** 2003
**Venue:** Academic Press (book)
**Authors:** Mayergoyz, I. D.
**Tags:** `#preisach` `#hysteresis` `#textbook` `#core`
**Status:** `read`
**PDF:** `not stored`
**Added:** 2026-05-04

---

## TL;DR

The definitive modern mathematical treatment of Preisach-type hysteresis models. Provides rigorous mathematical framework for the Preisach operator, including identification theorems and numerical implementation guidance.

## Why It Matters For FeCIM Lattice Tools

Provides the mathematical foundation for the Preisach model implementation, particularly the Everett function formulation. The product-form Everett function used in FeCIM (`Everett = (1+tanh((α-Ec)/Δ)) * (1-tanh((β+Ec)/Δ)) * Ps/4`) follows Mayergoyz's mathematical framework.

## Key Facts

### Everett Function

- **Product-form Everett** = analytically derived from Preisach density integral
  - **Source location:** Chapters on Preisach model identification
  - **Evidence level:** textbook (derived result)
  - **Notes:** The product-form ensures non-negative output (important fix applied Feb 2026).

## Methodology

Mathematical. Covers Preisach model definition, identification from experimental data, numerical implementation, and vector extensions.

## Limitations

- Textbook, not primary research
- Does not cover ferroelectric-specific adaptations
- No temperature or rate-dependent extensions

## Cited In

- [ ] `shared/physics/preisach.go` - Preisach model implementation
- [ ] `shared/physics/tanh_everett.go` - Everett function adapter
- [ ] `docs/3-develop/known-limitations.md`

## Related Sources

- `preisach1935_zphys` - Original Preisach model
- `khalatnikov1954_zheft_lk` - Landau-Khalatnikov model (complementary dynamic model)

---

## BibTeX

```bibtex
@book{mayergoyz2003,
  author = {Mayergoyz, I. D.},
  title = {Mathematical Models of Hysteresis and Their Applications},
  publisher = {Academic Press},
  year = {2003}
}
```

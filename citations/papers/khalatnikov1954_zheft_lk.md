# On the Anomalous Absorption of Sound Near a Second Order Phase Transition Point

**Key:** `khalatnikov1954_zheft_lk`
**Year:** 1954
**Venue:** Zhurnal Eksperimentalnoi i Teoreticheskoi Fiziki, 26, 677
**Authors:** Khalatnikov, I. M.
**Tags:** `#landau` `#khaltanikov` `#dynamics` `#foundational` `#core`
**Status:** `read`
**PDF:** `not stored`
**Added:** 2026-05-04

---

## TL;DR

Introduced the time-dependent extension of Landau's phase transition theory, now known as the Landau-Khalatnikov (LK) equation. The master equation: ρ·dP/dt = -∂G/∂P + E(t) describes the relaxation dynamics of an order parameter near a phase transition.

## Why It Matters For FeCIM Lattice Tools

The Landau-Khalatnikov equation is the foundation of the dynamic hysteresis solver in `shared/physics/landau.go`. The 551-line LK solver with RK4 integration, implicit Newton fallback, NLS coupling, and depolarization field all trace back to this equation.

## Key Facts

### LK Equation

- **Master equation** = `ρ · dP/dt = E_applied - dG/dP`
  - **Source location:** Original text
  - **Evidence level:** peer-reviewed
  - **Notes:** This is the base form. FeCIM extends it with depolarization field (K_dep·P), Langevin noise, series resistance coupling, and NLS modulation. In SI units: ρ in Ohm·m, P in C/m², E in V/m, G in J/m³.

## Methodology

Theoretical physics. Derives the time-dependent Landau equation by adding a dissipative term (viscosity ρ) to the Landau free energy functional. Originally applied to sound absorption near the λ-point of liquid helium.

## Limitations

- Original paper is for liquid helium λ-transition, not ferroelectrics
- The adaptation to ferroelectrics uses the LGD free energy (αP²/2 + βP⁴/4 + γP⁶/6) rather than the original order-parameter expansion
- No treatment of domain walls or nucleation

## Cited In

- [ ] `shared/physics/landau.go` - LK solver implementation
- [ ] `shared/viewmodel/hysteresis/viewmodel.go` - LK solver integration
- [ ] `docs/3-develop/known-limitations.md`

## Related Sources

- `materlik2015_jap_hfo2_origin` - LGD coefficients for HZO (β, γ)
- `alessandri2018_ieee_edl_switching` - NLS switching using LK dynamics

---

## BibTeX

```bibtex
@article{khalatnikov1954,
  author = {Khalatnikov, I. M.},
  title = {On the anomalous absorption of sound near a second order phase transition point},
  journal = {Zhurnal Eksperimentalnoi i Teoreticheskoi Fiziki},
  volume = {26},
  pages = {677},
  year = {1954}
}
```

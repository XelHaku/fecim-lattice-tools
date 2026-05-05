# CrossSim: A Hardware/Software Co-Design Tool for Analog In-Memory Computing

**Key:** `crosssim2024_sandia`
**DOI:** `10.2172/2563881`
**Year:** 2024
**Venue:** Sandia National Laboratories (SAND2024-05171C)
**Authors:** Foulk, James W. and Wahby, William and Xiao, Tianyao P. and Feinberg, Benjamin and Bennett, Christopher and Musuvathy, Srideep S.
**Tags:** `#crosssim` `#crossbar` `#validation` `#mvm` `#core`
**Status:** `skimmed`
**PDF:** `not stored`
**Added:** 2026-05-04

---

## TL;DR

CrossSim is a hardware/software co-design tool for analog in-memory computing (analog IMC) developed at Sandia National Laboratories. It models crossbar arrays with non-idealities and is the primary external validation target for FeCIM's crossbar module.

## Why It Matters For FeCIM Lattice Tools

FeCIM has ported CrossSim's SOR (Successive Over-Relaxation) solver to `shared/crossbar/solver.go` as the reference MVM implementation. CrossSim is the primary external validation target for Module 2 (Crossbar). The `crosssim_reference_8x8.json` test vectors were generated from CrossSim.

## Key Facts

### Architecture

- **Solver type** = `SOR (Successive Over-Relaxation)` for IR drop
- **Programming error models** = `5 models (ideal, uniform noise, level-dependent, write-verify, measured)`
  - **Source location:** CrossSim documentation
  - **Evidence level:** technical report
  - **Notes:** FeCIM has ported all 5 error models to `shared/crossbar/solver.go`.

## Methodology

Iterative SOR solver for crossbar MVM with IR drop. Nonlinear device models (I-V characteristics). Support for digital and analog peripheral models. Python-based with GPU acceleration.

## Limitations

- Python only (FeCIM port is Go)
- Assumes continuous conductance (not quantized by default)
- Not specifically optimized for FeFET devices (RRAM-focused)

## Cited In

- [ ] `shared/crossbar/solver.go` - SOR solver port
- [ ] `validation/external/crosssim_interop_test.go` - Interop test harness
- [ ] `validation/testdata/literature/crosssim_reference_8x8.json` - Reference vectors
- [ ] `docs/4-research/opensource-tools/opensource-crossbar.md` - Analysis doc

## Related Sources

- `badcrossbar` (Joksas et al., SoftwareX 2020) - Alternative nodal analysis validation

---

## BibTeX

```bibtex
@techreport{crosssim2024,
  author = {Foulk, James W. and Wahby, William and Xiao, Tianyao P. and
            Feinberg, Benjamin and Bennett, Christopher and Musuvathy, Srideep S.},
  title = {CrossSim: a hardware/software co-design tool for analog in-memory computing},
  institution = {Sandia National Laboratories},
  year = {2024},
  number = {SAND2024-05171C},
  doi = {10.2172/2563881}
}
```

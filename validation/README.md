# Validation

## Audience

This folder is for researchers, graduate students, reviewers, and contributors who need to see how FeCIM Lattice Tools checks its simulation claims. It is not a replacement for silicon measurements. It is the evidence layer for a public educational and research simulator.

## What Validation Means Here

Unit tests show that code paths run. Validation shows whether the modeled behavior respects physics, numerical conservation laws, external tools, or published reference data.

Every strong project claim should map to:

- a command someone can run
- a pass/fail threshold
- a generated artifact or fixture
- a limitation statement when the evidence is partial

Generated validation artifacts are written under `output/validation/` and are intentionally not tracked in git unless they are promoted to a curated fixture under `validation/testdata/`.

## One-Command Reproduction

```bash
bash scripts/reproduce_validation.sh
```

For the public Module 2 conservation gate only:

```bash
go test -v ./validation/module2/... -run TestModule2KCLConservation_PublicValidation
```

That test writes:

```text
output/validation/module2/kcl_conservation.json
```

## Current Executable Validation

| Area | Evidence | Command |
|---|---|---|
| Module 1 hysteresis | Literature-backed P-E loop checks against digitized datasets, including Park 2015 HZO | `go test -v ./validation/literature/...` |
| Module 2 crossbar | Kirchhoff current conservation over 100 deterministic random parasitic arrays | `go test -v ./validation/module2/...` |
| Module 2 external comparison | NumPy/SciPy and ngspice comparison harnesses where external tools are installed | `go test -v ./validation/external/...` |
| Module 4 circuits | KCL/KVL and sense-chain regression checks | `go test -v ./validation/... -run 'Module4|SenseChain'` |
| Module 6 EDA | Verilog sanity/lint and OpenLane smoke tests where tools are installed | `go test -v ./validation/external/... -run 'Verilog|OpenLane'` |
| Configuration | YAML/JSON validation for array, calibration, Preisach, weight, and OpenLane configs | `go test -v ./validation/configvalidator/...` |

External tool checks are optional by design. If `ngspice`, Yosys, OpenLane, or Python scientific packages are not installed, those tests skip the external execution while keeping structural checks active.

## Module 2 KCL Gate

`validation/module2/kcl_conservation_test.go` checks the Module 2 parasitic crossbar solver against Kirchhoff's Current Law.

The test:

- builds 100 deterministic random arrays
- varies array shape, conductance, wire parasitics, and applied voltages
- solves each parasitic matrix-vector multiply
- reconstructs cumulative row and column currents
- checks that every node conserves current
- requires maximum KCL residual below `1e-9 A`
- emits a JSON report with seed, threshold, maximum residual, and worst case

This proves current conservation inside the solver. It does not prove agreement with fabricated devices or SPICE by itself; those are separate validation layers.

## Package Structure

| Path | Purpose |
|---|---|
| `literature/` | Digitized reference data and literature-backed physics validation |
| `module2/` | Public Module 2 conservation and crossbar validation gates |
| `external/` | Optional external-tool comparisons, including ngspice and OpenLane checks |
| `integration/` | Cross-module validation between physics, crossbar, inference, and EDA paths |
| `configvalidator/` | Rule-based config validation engine and CLI |
| `testdata/` | Curated fixtures and golden references |
| `output/` | Generated local artifacts, ignored by git |

Core statistical helpers live in the root validation package:

- `literature.go` loads published experimental datasets
- `statistics.go` provides KS, chi-square, RMSE, MAE, and correlation metrics
- `interfaces.go` defines shared validation interfaces
- `readiness_report.go` generates release-readiness summaries

## Trust Boundaries

The repository is currently an educational simulation toolkit. A passing validation suite means the simulator is internally consistent and matches selected external references within declared tolerances. It does not mean the repository reports new silicon measurements.

Known public-facing limits:

- Some literature datasets are digitized from figures and carry digitization uncertainty.
- External comparisons depend on installed local tools and versions.
- MNIST results are pipeline demonstrations unless accompanied by a full training/inference artifact and confusion matrix.
- EDA export validity is staged: syntax and smoke tests are not the same as a clean full physical implementation.

See `validation/PLANNED.md` for the remaining validation roadmap.

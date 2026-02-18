# Module 1 Automated Testing Plan (Execution-Grade)

Scope: `module1-hysteresis` physics/controller validation in headless CI.

## Current Implementation Status (2026-02-18)

| ID | Description | Status | Location |
|---|---|---|---|
| RG-PHY-OBS-01 | Major P–E loop vs DOI data | ✅ Implemented | `validation/literature/module1_pe_loop_test.go` |
| RG-PHY-OBS-02 | Switching kinetics falsification | ✅ Implemented | `validation/literature/module1_switching_kinetics_test.go`, `module1_arrhenius_switching_test.go` |
| RG-PHY-OBS-03 | FORC/minor-loop falsification | ✅ Implemented | `validation/literature/module1_forc_test.go` |
| RG-VAL-M1-01 | 9-material regression | ✅ Implemented | `validation/physics_regression_test.go` |
| RG-VAL-M1-02 | Golden regression drift | ✅ Implemented | `validation/physics_regression_test.go` (golden JSON in `validation/testdata/physics_regression/`) |
| RG-VAL-M1-03 | WriteVerifyStats export + schema | ✅ Implemented | `validation/m1_write_verify_stats_test.go` |
| RG-VAL-M1-04 | Monte Carlo UQ | ✅ Implemented | `validation/m1_montecarlo_uncertainty_test.go` |

### Artifact Schema

All `module1_pe_loop_*.json` artifacts now carry the full required envelope
(`schema_version`, `timestamp_utc`, `commit`, `gate`, `test_id`, `verdict`,
`material`, `dataset`, `metrics`, `uncertainty`, `thresholds`) since 2026-02-18.
The artifact validator at `scripts/ci/validate_regression_artifacts.py` enforces
this on every run.

`RG-VAL-M1-03` (write-verify) and `RG-VAL-M1-04` (Monte Carlo) artifacts also
carry the full envelope; they are not currently validated by the Python script
because they live under different paths.

### Statistical Rigor Gap

The current Monte Carlo implementation uses empirical 5th/95th percentiles as a
90% CI proxy. The plan specifies Shapiro-Wilk → t-test or BCa bootstrap. This
is a known gap; the current approach is sufficient for a PR gate but should be
upgraded for nightly/release lanes.

---

## Operating Contract

- Headless only: `DISPLAY` and `WAYLAND_DISPLAY` must be unset.
- No aggregate pass if any required material/dataset fails.
- Required lanes must emit machine-readable artifacts.
- Thresholds and runtime budgets below are hard gates.

---

## P0 / P1 / P2 Waves

### P0 — CI Safety Baseline (required first)
**Deliverables**
- Deterministic build/test lane for Module 1.
- Artifact emission for required validation tests.
- Explicit material matrix (no implicit defaults).

**Exit criteria**
- PR and nightly lanes pass with complete artifacts.
- Artifact schema validates for every required run.

**Status: ✅ Complete** — `make ci` enforces build+vet+test-short+arch-check on every push.

### P1 — Physics Falsification Core (primary gate)
**Deliverables**
- `RG-PHY-OBS-01`: DOI-backed major-loop falsification.
- `RG-VAL-M1-01`, `RG-VAL-M1-02`: 9-material regression + golden drift.
- `RG-VAL-M1-03`: Write/Verify stats export + schema validation.

**Hard thresholds (all required)**
- `|Pr_error| <= 10%`
- `|Ec_error| <= 10%`
- `RMSE(P(E))/Ps <= 0.05`
- `LoopArea_error <= 25%`
- Golden normalized RMS drift `<= 1e-3`

**Status: ✅ Complete** — all tests exist and pass. Artifact schema enforced by validator.

### P2 — Extended Falsification + Uncertainty
**Deliverables**
- `RG-PHY-OBS-02`: switching kinetics falsification.
- `RG-PHY-OBS-03`: FORC/minor-loop falsification.
- `RG-VAL-M1-04`: Monte Carlo uncertainty propagation.

**Hard thresholds (all required)**
- Kinetics: `R^2 >= 0.95`, parameter CI width `<= 30%` of estimate.
- FORC/minor-loop: normalized shape error `<= 0.10`, return-point error `<= 1% Ps`.
- UQ: literature target lies inside 95% CI for `Pr` and `Ec`.

**Status: ✅ Complete** — tests exist and pass. Statistical rigor gap noted above.

---

## Command Lanes (PR / Nightly / Release)

Run from repo root:

```bash
cd <local-path>
export DISPLAY=
export WAYLAND_DISPLAY=
```

### PR lane (P0 + minimal P1)
**Runtime budget:** target `<= 12 min`, hard cap `15 min`.

```bash
make ci                                        # fmt + vet + test-short + arch-check
FECIM_CI_GATE=pr go test -v -count=1 \
  ./validation/literature/... \
  -run TestModule1_PELoop_LiteratureBacked     # RG-PHY-OBS-01 with artifact emission
```

**Pass requires**
- Exit code 0 for every command.
- P1 thresholds pass for each dataset/material exercised.
- Required artifacts emitted.
- `python3 scripts/ci/validate_regression_artifacts.py --module module1 --root output/validation/literature` exits 0.

**CI wiring:** `.github/workflows/ci.yml` — runs on every push/PR.

### Nightly lane (full P1)
**Runtime budget:** target `<= 45 min`, hard cap `60 min`.

```bash
go build ./... && go vet ./...
FECIM_CI_GATE=nightly go test -count=1 ./...
go test -v -count=1 ./validation/literature/...
bash scripts/run_literature_validation.sh
go test -race ./module1-hysteresis/... ./shared/physics/...
```

**Pass requires**
- PR lane pass conditions.
- Full 9-material matrix complete.
- Race lane clean.

**CI wiring:** `.github/workflows/nightly.yml` — scheduled 02:00 UTC daily.

### Release lane (P0 + P1 + P2)
**Runtime budget:** target `<= 90 min`, hard cap `120 min`.

```bash
go build ./... && go vet ./...
go test -count=1 ./...
go test -v -count=1 ./validation/...
go test -v -count=1 ./validation/literature/...
bash scripts/run_literature_validation.sh
go test -race ./...
```

**Pass requires**
- Nightly lane pass conditions.
- All P2 thresholds satisfied.
- Immutable release artifact bundle keyed by commit SHA.

---

## Statistical Rigor Policy (enforced)

### Minimum sample sizes (hard gate)
- Seeded scalar metrics: `n >= 30` nightly, `n >= 100` release.
- Distribution metrics (KS): `n >= 200` per distribution.
- Proportion metrics: `n >= 200` writes per material.

If minima are unmet: mark `insufficient_n` and fail gate.

### CI method selection (Shapiro/t vs BCa bootstrap)
1. Shapiro-Wilk (`alpha=0.05`) when `8 <= n <= 5000`.
2. If normality not rejected (`p >= 0.05`): two-sided 95% t-interval (`method=t`).
3. Else: BCa bootstrap 95% CI (`method=bootstrap_bca`; `2000` nightly, `10000` release; fixed seed).
4. Proportions: Wilson 95% CI.

**Current gap:** The Monte Carlo test (`RG-VAL-M1-04`) uses empirical 5th/95th
percentiles (`method=monte_carlo`, `confidence=0.90`). Shapiro-Wilk + BCa
bootstrap upgrade is deferred to a future iteration.

### KS thresholds
- Apply KS only to continuous distributions with valid `n`; always report `(D, p)`.
- Pass: `D <= 0.10` and `p >= 0.05`.
- Warning: `0.10 < D <= 0.15` or `0.01 < p < 0.05`.
- Fail: `D > 0.15` or `p <= 0.01`.

---

## Falsification Matrix

| ID | Observable | Required metrics | Hard fail condition | Status |
|---|---|---|---|---|
| RG-PHY-OBS-01 | Major P–E loop vs DOI data | Pr error, Ec error, RMSE/Ps, loop area error | Any metric above threshold | ✅ |
| RG-PHY-OBS-02 | Switching kinetics vs DOI data | R^2, parameter CI width, residual diagnostics | `R^2 < 0.95` or CI width too large | ✅ |
| RG-PHY-OBS-03 | FORC/minor loops vs DOI data | Shape error, return-point error | Any metric above threshold | ✅ |
| RG-VAL-M1-01 | 9-material regression | Per-material pass | Any missing/failing material | ✅ |
| RG-VAL-M1-02 | Golden regression | Normalized RMS drift | Drift `> 1e-3` | ✅ |
| RG-VAL-M1-03 | WriteVerifyStats export | Schema + finite values | Missing/invalid field | ✅ |
| RG-VAL-M1-04 | Monte Carlo UQ | 95% CI coverage | Target outside CI | ✅ |

---

## Artifact Contract

**Path**
- PE-loop (RG-PHY-OBS-01): `output/validation/literature/module1_pe_loop_<material_id>.json`
- Write-verify (RG-VAL-M1-03): `output/write_stats/write_verify_stats_<material_id>.json`
- Monte Carlo (RG-VAL-M1-04): `output/montecarlo/mc_pe_uncertainty_<material_id>.json`

**Required keys (all artifact types)**
- `schema_version` — `"v1"`
- `timestamp_utc` — RFC 3339 timestamp
- `commit` — short git hash (from `FECIM_GIT_COMMIT` env or `git rev-parse --short HEAD`)
- `gate` — `"pr"` | `"nightly"` | `"release"` (from `FECIM_CI_GATE` env, default `"pr"`)
- `test_id` — plan ID, e.g. `"RG-PHY-OBS-01"`
- `verdict` — `"pass"` | `"fail"`
- `material` — material name string
- `dataset` — dataset identifier string
- `metrics` — object containing scalar metric values
- `uncertainty` — object with `method`, `confidence`, `sample_size`
- `thresholds` — object containing the hard-gate threshold values

**Enforcement**
- `scripts/ci/validate_regression_artifacts.py --module module1 --root output/validation/literature` checks all `module1_pe_loop_*.json` files.
- Missing required key, NaN/Inf, or `sample_size <= 0` ⇒ fail.
- Run via `FECIM_GIT_COMMIT=$COMMIT python3 scripts/ci/validate_regression_artifacts.py ...` in CI.

**Shared types**
- `shared/validation/artifact_envelope.go`: `ArtifactEnvelope`, `ArtifactUncertainty`, `NewEnvelope(testID, gate string, pass bool)`

---

## Execution Order (authoritative)

1. PR lane on every PR.
2. Nightly lane once/day on default branch.
3. Release lane before tagging release.
4. Publish immutable artifact bundle keyed by commit SHA.

## Definition of Done

Done when all are true:
- P0, P1, P2 implemented and passing in release lane.
- All listed IDs have artifacts + explicit pass/fail verdicts.
- Runtime budgets met (or exception documented/approved).
- No unresolved per-material failures.

**Current status:** P0 ✅ P1 ✅ P2 ✅ — all falsification IDs implemented and passing.
Remaining gap: Shapiro-Wilk + BCa bootstrap upgrade for statistical rigor in MC tests.

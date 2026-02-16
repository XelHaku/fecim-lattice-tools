# Module 1 Automated Testing Plan (Execution-Ready, Enforceable)

Scope: `module1-hysteresis` physics + controller validation in headless CI.

## 1) Objective and Operating Rules

**Objective:** falsify (not just regress) Module 1 behavior against DOI-backed observables, with deterministic artifacts and explicit pass/fail thresholds.

**Hard rules**
- Required lanes are headless only (`DISPLAY` and `WAYLAND_DISPLAY` unset).
- Every required test emits machine-readable artifacts.
- No aggregate pass if any material/dataset fails.
- Commands and runtime budgets in this plan are binding for CI gates.

## 2) Phased Delivery (P0/P1/P2)

## P0 — CI Safety Baseline (must exist before any claim)

Deliverables
- Deterministic build/test baseline for Module 1 lanes.
- Artifact emission wired into required tests.
- Material-explicit execution matrix (no implicit defaults).

Acceptance
- Gate commands (Section 3) run green in PR and nightly.
- Artifact schema (Section 5) validates for every required run.

## P1 — Physics Falsification Core (primary scientific gate)

Deliverables
- DOI-backed major-loop falsification (`RG-PHY-OBS-01`).
- 9-material deep regression + golden loop drift checks (`RG-VAL-M1-01`, `RG-VAL-M1-02`).
- Write/Verify stats exported and validated (`RG-VAL-M1-03`).

Acceptance thresholds (must all pass)
- `|Pr_error| <= 10%`
- `|Ec_error| <= 10%`
- `RMSE(P(E))/Ps <= 0.05`
- `LoopArea_error <= 25%`
- Golden drift: normalized RMS drift `<= 1e-3` vs approved baseline artifact.

## P2 — Extended Falsification + Uncertainty

Deliverables
- Switching kinetics falsification (`RG-PHY-OBS-02`) with explicit fit quality.
- FORC/minor-loop falsification (`RG-PHY-OBS-03`).
- Monte Carlo uncertainty propagation (`RG-VAL-M1-04`).

Acceptance thresholds (must all pass)
- Kinetics fit: `R^2 >= 0.95` and parameter CI width <= 30% of estimate.
- FORC/minor-loop: normalized shape error `<= 0.10`, return-point error `<= 1% Ps`.
- Uncertainty: literature target metric lies inside 95% CI for `Pr` and `Ec`.

## 3) CI Gates with Exact Commands + Runtime Budgets

All commands executed from repo root:

```bash
cd <local-path>
export DISPLAY=
export WAYLAND_DISPLAY=
```

## PR Gate (target <= 12 min, hard cap 15 min)

Purpose: fail fast on regressions; run P0 + minimal P1.

```bash
go build ./... && go vet ./...
go test -short -count=1 ./...
go test -v -count=1 ./validation/literature/... -run TestModule1_PELoop_LiteratureBacked
```

Pass criteria
- Exit code 0 for every command.
- Required falsification thresholds pass for each dataset/material in run.
- Artifacts generated for each falsification test invocation.

## Nightly Gate (target <= 45 min, hard cap 60 min)

Purpose: full P1 + broad stability checks.

```bash
go build ./... && go vet ./...
go test -count=1 ./...
go test -v -count=1 ./validation/literature/...
bash scripts/run_literature_validation.sh
go test -race ./module1-hysteresis/... ./shared/physics/...
```

Pass criteria
- All PR criteria, plus race-free execution.
- 9-material matrix complete (no missing material verdicts).

## Release Gate (target <= 90 min, hard cap 120 min)

Purpose: P0 + P1 + P2 publication-grade evidence.

```bash
go build ./... && go vet ./...
go test -count=1 ./...
go test -v -count=1 ./validation/...
go test -v -count=1 ./validation/literature/...
bash scripts/run_literature_validation.sh
go test -race ./...
```

Pass criteria
- All nightly criteria, plus P2 thresholds met.
- Release artifact bundle produced (Section 5) with immutable commit hash.

## 4) Falsification Matrix (enforceable)

| ID | Observable | Required metric(s) | Fail condition |
|---|---|---|---|
| RG-PHY-OBS-01 | Major P–E loop vs DOI data | Pr error, Ec error, RMSE/Ps, loop area error | Any metric over threshold |
| RG-PHY-OBS-02 | Switching kinetics vs DOI data | R^2, parameter CI width, residual diagnostics | R^2 < 0.95 or CI too wide |
| RG-PHY-OBS-03 | FORC/minor loops vs DOI data | Shape error, return-point error | Any metric over threshold |
| RG-VAL-M1-01 | 9-material regression | per-material pass count | Any material missing/failing |
| RG-VAL-M1-02 | Golden regression | normalized RMS drift | Drift > 1e-3 |
| RG-VAL-M1-03 | WriteVerifyStats export | schema compliance + finite values | Missing/invalid field |
| RG-VAL-M1-04 | Monte Carlo UQ | 95% CI coverage | Target outside CI |

## 5) Artifact Contract (required JSON schema)

Each required test writes one JSON artifact under:
- `output/validation/module1/<gate>/<test_id>/<material>/<dataset>.json`

Minimal schema (required keys)

```json
{
  "schema_version": "m1.validation.v1",
  "timestamp_utc": "RFC3339",
  "commit": "<git sha>",
  "gate": "pr|nightly|release",
  "test_id": "RG-PHY-OBS-01",
  "material": {
    "name": "string",
    "Ec_Vm": 0,
    "Ps_Cm2": 0,
    "Pr_Cm2": 0,
    "thickness_m": 0,
    "Gmin_S": 0,
    "Gmax_S": 0
  },
  "dataset": {
    "doi": "string",
    "source_ref": "figure/table identifier",
    "units": {"E": "MV/cm", "P": "uC/cm2"}
  },
  "metrics": {
    "pr_error_pct": 0,
    "ec_error_pct": 0,
    "rmse_over_ps": 0,
    "loop_area_error_pct": 0,
    "r2": 0,
    "return_point_error_over_ps": 0
  },
  "thresholds": {
    "pr_error_pct_max": 10,
    "ec_error_pct_max": 10,
    "rmse_over_ps_max": 0.05,
    "loop_area_error_pct_max": 25
  },
  "verdict": "pass|fail",
  "notes": "optional"
}
```

Schema enforcement
- Missing required key, NaN/Inf metric, or unit mismatch => automatic fail.
- `verdict` must be derivable from metrics + thresholds (no manual override).

## 6) Execution Order (single source of truth)

1. Run PR gate on every PR.
2. Run nightly gate once per day on default branch.
3. Run release gate before tagging release.
4. Publish artifact bundle and keep immutable by commit hash.

## 7) Definition of Done

Plan execution is complete when:
- P0, P1, P2 all implemented and passing at release gate.
- All falsification IDs above have artifacts + explicit pass/fail verdicts.
- Runtime budgets are met (or documented exception approved with reason).
- No unresolved per-material failures.

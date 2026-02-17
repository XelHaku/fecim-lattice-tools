# Module 4 Automated Testing Plan (Execution-Ready)

**Applies to:** `module4-circuits`  
**Goal:** enforce research-grade, headless, reproducible validation for READ / WRITE / COMPUTE behavior with falsifiable thresholds and auditable artifacts.

---

## 1) Non-Negotiable Rules

1. **Headless-only for required gates**
   - Required gates must run with `DISPLAY` and `WAYLAND_DISPLAY` unset.
   - No `xvfb-run` in required lanes.
2. **Material-explicit testing**
   - Every case binds explicit material; no implicit defaults.
   - Verdicts are per-material; aggregate cannot mask a single material failure.
3. **Single source of physics truth**
   - Headless and GUI must share physics code paths (no duplicated equations).
4. **Deterministic evidence**
   - Fixed seed required in gated suites.
   - Every gate emits machine-readable artifacts.

---

## 2) Phase Plan (P0 / P1 / P2)

## P0 — Mandatory PR Safety Net (Blocker Phase)

**Objective:** prevent silent physics regressions in daily development.

**Exit criteria (all required):**
- READ and COMPUTE immutability pass.
- WRITE locality + broadcast guard pass.
- KCL/KVL + basic power conservation pass.
- No NaN/Inf.
- Deterministic run reproducibility (same seed -> bitwise-equal key outputs).
- Artifact schema v2 emitted for every gated run.

**Priority tests (must gate PR):**
- T1 invariants (immutability, locality, determinism, finite outputs, conductance bounds).
- Broadcast guard thresholding.
- Core solver correctness (Tier-A/Tier-B residual checks).

---

## P1 — Quantitative Fidelity (PR + Nightly)

**Objective:** move from invariant checks to measurable circuit/peripheral fidelity.

**Exit criteria (all required):**
- DAC/ADC/TIA characterization automated and gated.
- IR-drop scaling and solver diagnostics gated.
- Signal-chain trace + uncertainty fields emitted.
- Statistical confidence checks (bootstrap CI + distribution checks) implemented.

**Priority tests (gated by lane):**
- INL/DNL/ENOB/SNR/settling checks.
- MVM PSNR + BER checks.
- Tier-B condition-number and convergence diagnostics.

---

## P2 — Release-Grade Robustness (Nightly + Release)

**Objective:** robustness across process/temperature/size/architecture/material space.

**Exit criteria (all required for release):**
- Extended matrix coverage completed (or approved DOE sampling plan with justification).
- Retention/endurance/PVT suites pass thresholds.
- Cross-model/SPICE agreement within limits.
- Trend-regression checks integrated (no silent drift vs previous release baseline).

---

## 3) Enforced Gates by Pipeline Stage

| Stage | Scope | Must Pass | Runtime Budget |
|---|---|---|---|
| **PR gate** | Fast deterministic subset | P0 full + P1 fast subset | **<= 8 min** |
| **Nightly gate** | Broad matrix + diagnostics | P0 + full P1 + selected P2 | **<= 60 min** |
| **Release gate** | Full validation and evidence freeze | All P0/P1/P2 required suites | **<= 180 min** |

If runtime exceeds budget, gate **fails** unless an explicit waiver is merged before run.

### Runtime impact of statistical policy (enforced cap)

| Stage | Added stats workload | Max added wall-time vs non-stat run |
|---|---|---|
| **PR** | `n` minima + 1000-resample bootstrap where required | **<= +90s** |
| **Nightly** | full `n` minima + 2000-resample bootstrap + KS drift checks | **<= +12 min** |
| **Release** | release `n` minima + 10000-resample bootstrap + full KS drift matrix | **<= +35 min** |

If added wall-time exceeds the cap, pipeline must reduce matrix breadth (not sample-size minima) or parallelize execution; otherwise gate fails.

---

## 4) Exact Commands (Canonical)

Run from repo root: `<local-path>`

## 4.1 PR Gate Commands (required)

```bash
cd <local-path>
unset DISPLAY WAYLAND_DISPLAY

go build ./...
go vet ./...

go test -short -count=1 ./...

# Module 4 fast headless regression
bash scripts/run_module4_fast_gate.sh
bash scripts/run_headless_module4_regressions.sh
bash scripts/run_headless_ispp_regressions.sh
```

## 4.2 Nightly Gate Commands (required)

```bash
cd <local-path>
unset DISPLAY WAYLAND_DISPLAY

bash scripts/ci/go-test-all.sh
bash scripts/run_headless_module4_regressions.sh
bash scripts/run_headless_ispp_regressions.sh

go test -count=1 -v ./module4-circuits/pkg/arraysim/...
```

## 4.3 Release Gate Commands (required)

```bash
cd <local-path>
unset DISPLAY WAYLAND_DISPLAY

bash scripts/ci/go-test-all.sh
bash scripts/ci/go-test-race.sh
bash scripts/run_headless_module4_regressions.sh
bash scripts/run_headless_ispp_regressions.sh

go test -count=1 -v ./validation/...
go test -count=1 -v ./module4-circuits/pkg/arraysim/...
```

---

## 5) Falsification Thresholds (Hard Fail)

A gate fails when **any** threshold below is violated.

### 5.0 Statistical and uncertainty policy (mandatory)

**Sample-size minima (per metric, per material, per operating corner):**
- **PR gate (fast subset):** scalar `n >= 10`; proportions `n >= 60`; distribution tests `n >= 80`.
- **Nightly gate:** scalar `n >= 30`; proportions `n >= 200`; distribution tests `n >= 200`.
- **Release gate:** scalar `n >= 100`; proportions `n >= 500`; distribution tests `n >= 500`.
- If a minimum is not met, verdict must be `FAIL_INSUFFICIENT_N` (no pass-by-omission).

**CI method policy (95% two-sided, deterministic):**
- `n < 8`: CI is invalid for gated metrics -> gate fail.
- Continuous metrics, `8 <= n <= 5000`: run Shapiro-Wilk (`alpha=0.05`).
  - `p >= 0.05` -> Student-t CI.
  - `p < 0.05` -> BCa bootstrap CI.
- Continuous metrics, `n > 5000`: skip Shapiro-Wilk; use BCa bootstrap CI.
- Bootstrap resamples: PR `1000`, nightly `2000`, release `10000`; fixed seed required.
- Proportion metrics: Wilson CI only (never normal approximation).

**KS drift policy (two-sample, against pinned baseline artifact):**
- Baseline and candidate must both meet distribution-test minimum `n`.
- `p < 0.01` **and** `D >= 0.15` -> fail.
- `0.01 <= p < 0.05` **or** `0.10 <= D < 0.15` -> warn.
- Otherwise pass.
- Missing baseline reference or KS metadata on required drift metrics -> gate fail.

- Missing CI metadata, method/seed/resample count, or sample-size fields is a gate failure.

## 5.1 Physics/solver integrity
- KCL residual: `max_residual <= 1e-6`
- Tier-A dense residual: `||Ax-b||/||b|| < 1e-12`
- Tier-B PCG residual: `||r||/||b|| < 1e-8`
- Tier-B condition number: warn `>1e8`, fail `>1e12`
- No NaN/Inf in any reported metric.

## 5.2 Operation invariants
- READ immutability: `0` changed cells.
- COMPUTE immutability: `0` changed cells.
- WRITE locality:
  - 0T1R: changes only on selected row/column set.
  - 1T1R: non-target change `<=1` level.
  - 2T1R: non-target change `==0` levels.
- Broadcast guard default: fail when `count(non-target Δlevel > 3) > 3`.

## 5.3 Accuracy and peripheral metrics
- MVM PSNR: `>= 30 dB`
- MVM BER (task patterns): `< 5%`
- TIA SNR: `> 40 dB`
- TIA dynamic range: `> 60 dB`
- TIA settling: `< 80 ns`
- DAC/ADC monotonicity: no missing codes; DNL `>-1 LSB` everywhere.
- DAC/ADC INL: `< 1 LSB` (nominal corner).
- ADC ENOB: `>= nominal_bits - 1`

## 5.4 Thermodynamic consistency
- Non-negative dissipation for all resistive elements.
- Power closure: `|Pin - Pdiss| / Pin < 1%`
- Cumulative energy monotonicity during transient/write sequences.

---

## 6) Required Artifact Set (per gated run)

Output root:

```text
output/regression/module4/<timestamp>/
```

Minimum required files:

1. `summary.json` (run metadata, git hash, seed, pass/fail)
2. `matrix.json` (architectures/tier/size/materials covered)
3. `thresholds.json` (effective thresholds used)
4. `results.json` (all measured metrics + verdicts)
5. `material_snapshot.json` (Ec, Ps, Pr, thickness, Gmin/Gmax, quant_levels)
6. `signal_chain_trace.json` (DAC->Array->TIA->ADC stage values)
7. `solver_diagnostics.json` (residuals, iterations, condition number)
8. `thermodynamics.json` (power/energy closure checks)
9. `confidence_ledger.json` (`measured|estimated|placeholder` parameter tags)
10. `provenance.json` with:
   - `run_id`, `timestamp_utc`, `git_hash`, `git_dirty`
   - `pipeline_stage` (`pr|nightly|release`), `command_manifest` (exact commands)
   - `baseline_artifact_ref` (path + commit hash) for drift comparisons
   - `data_sources[]` (`id`, `type`, `path_or_uri`, `sha256`, `doi?`, `license?`)
   - `toolchain` (`go_version`, `os`, `arch`, script versions)
11. `uncertainty.json` with required per-metric fields:
   - `metric_id`, `units`, `material`, `corner`, `n`
   - `ci_level` (=0.95), `ci_method` (`t|bootstrap_bca|wilson`)
   - `estimate`, `ci_low`, `ci_high`
   - `normality_test` (`name`, `p_value`, `alpha`, `result`) when applicable
   - `bootstrap` (`resamples`, `seed`) when applicable
   - `ks_drift` (`baseline_ref`, `n_candidate`, `n_baseline`, `D`, `p_value`, `verdict`) when applicable

**Artifact validity rules:**
- Missing required artifact = gate fail.
- Invalid JSON or schema mismatch = gate fail.
- Artifact `git_hash` must match tested commit.

---

## 7) Coverage Matrix Policy

## PR matrix (required fast subset)
- Architectures: `0T1R, 1T1R`
- Coupling tiers: `Ideal, TierA`
- Sizes: `2x2, 8x8`
- Materials: `fecim_hzo, literature_superlattice, default_hzo`
- Operations: `READ, WRITE, COMPUTE, WRC`

## Nightly matrix (required expanded subset)
- Architectures: `0T1R, 1T1R, 2T1R`
- Coupling tiers: `Ideal, TierA, TierB`
- Sizes: `2x2, 8x8, 16x16, 32x32`
- Materials: 9-material set
- Includes PVT/process and statistical diagnostics

## Release matrix
- Full nightly requirements plus race check and validation suites.
- If DOE sampling is used instead of full combinatorial coverage, sampling method and confidence rationale must be archived in `matrix.json`.

---

## 8) Enforceability Rules

1. **No “informational only” for P0 checks** — P0 is always blocking.
2. **Threshold changes require PR diff** in this file and artifacted before/after comparison.
3. **New materials cannot bypass gates**; they must inherit full threshold set.
4. **Flaky-test policy**: if nondeterministic behavior is detected, gate fails until root cause is fixed or thresholded with signed rationale.
5. **Waivers are time-bounded** and must include issue ID, owner, and expiry date.

---

## 9) Minimal Ownership and Review Contract

- **Module owner:** maintains thresholds and matrix definitions.
- **CI owner:** maintains command execution + artifact publishing.
- **Reviewer requirement:** no merge when PR gate missing artifacts or threshold failures.

---

## 10) Definition of Done

This plan is considered implemented when:
- PR, nightly, and release pipelines execute the exact commands above.
- Required artifacts are emitted and schema-validated.
- P0/P1/P2 thresholds are machine-checked and blocking.
- Runtime budgets are met (or waived with tracked expiry).

---

## 11) Change Log

- **2026-02-16:** Refactored to concise execution-ready format; removed duplicate catalog tables; converted broad requirements into enforceable phase gates, exact commands, runtime budgets, hard falsification thresholds, and mandatory artifacts.
- **2026-02-16 (stats tighten):** Added stage-specific sample-size minima, deterministic CI method policy, KS decision rules with effect-size guard, required `provenance.json`, strict `uncertainty.json` schema fields, and explicit PR/nightly/release runtime-impact caps.

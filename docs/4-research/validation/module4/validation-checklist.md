# Module 4 (Circuits) Validation Checklist

**Purpose:** end-to-end validation of Module 4 circuits, data fidelity, and GUI correctness against literature and external reference tools.
**Scope:** module4-circuits (READ/WRITE/COMPUTE paths, DAC/ADC/TIA, array simulator, GUI controls).
**Status:** living checklist; every item must link to evidence (artifact path + DOI or reference tool).

---

## 0) Pre-flight (required for every validation run)
- [ ] Toolchain recorded in `provenance.json` (go version, OS, arch, git hash, git dirty).
- [ ] `DISPLAY` and `WAYLAND_DISPLAY` **unset** for headless gates.
- [ ] Material explicitly bound (no defaults).
- [ ] All artifacts emitted under `output/regression/module4/<timestamp>/`.

---

## 1) Data Fidelity vs Literature (physics + circuits)

### 1.1 Device/array physics consistency
- [ ] **M1–M4 physics contract** verified (mapping P↔G matches `shared/physics/transfer.go`).
- [ ] Polarization units: C/m² internally; report µC/cm² to user.
- [ ] Conductance levels: monotonic, within [Gmin, Gmax], 30 levels by default.

### 1.2 Peripheral circuit metrics (must map to literature bounds)
- [ ] DAC monotonicity (no missing codes), DNL > -1 LSB, INL < 1 LSB (nominal).
- [ ] ADC ENOB ≥ (nominal_bits - 1) with noise model documented.
- [ ] TIA SNR > 40 dB, dynamic range > 60 dB, settling < 80 ns.
- [ ] Signal-chain trace (DAC → Array → TIA → ADC) logged and unit-checked.

### 1.3 Thermodynamics / power integrity
- [ ] Non-negative dissipation for resistive elements.
- [ ] Power closure |Pin - Pdiss| / Pin < 1%.
- [ ] Energy monotonicity during writes; no negative cumulative energy.

### 1.4 Literature-linked validation
- [ ] Each metric above includes DOI-backed reference range (stored in `provenance.json`).
- [ ] Error metrics: RMSE/Ps <= 0.05 where applicable; Pr/Ec error <= 10% for calibrated datasets.

---

## 2) Array & Operation Invariants (behavioral correctness)
- [ ] READ immutability: 0 changed cells.
- [ ] COMPUTE immutability: 0 changed cells.
- [ ] WRITE locality:
  - [ ] 0T1R: only selected row/column changes.
  - [ ] 1T1R: non-target change ≤ 1 level.
  - [ ] 2T1R: non-target change = 0 levels.
- [ ] Broadcast guard: count(non-target Δlevel > 3) ≤ 3.
- [ ] Solver residuals: Tier-A < 1e-12, Tier-B < 1e-8.
- [ ] No NaN/Inf anywhere in arrays or metrics.

---

## 3) Coverage Matrix (combinatorial validation)

### 3.1 Architecture coverage
- [ ] 0T1R
- [ ] 1T1R
- [ ] 2T1R

### 3.2 Coupling tiers
- [ ] Ideal
- [ ] TierA
- [ ] TierB

### 3.3 Array sizes
- [ ] 2x2
- [ ] 8x8
- [ ] 16x16
- [ ] 32x32

### 3.4 Materials (explicit)
- [ ] fecim_hzo
- [ ] literature_superlattice
- [ ] default_hzo
- [ ] additional literature materials (list in matrix.json)

---

## 4) GUI Validation (controls, visuals, and correctness)

### 4.1 Control-to-model mapping (no phantom controls)
- [ ] **Preisach view**: no temperature slider if temperature not modeled.
- [ ] **Landau-Khalatnikov view**: temperature slider only if LK coefficients are T-dependent in code.
- [ ] Sliders map to actual parameters; ranges match model constraints.
- [ ] Buttons (run/stop/reset) wired to real actions; no dead controls.

### 4.2 GUI regression tests (headless)
- [ ] `go test -count=1 ./module4-circuits/pkg/gui -run 'Parity|ReadChain|Layout'`
- [ ] `go test -count=1 ./module4-circuits/pkg/gui/unified/...`

### 4.3 Manual visual verification (screenshot-based evidence)
- [ ] Layout aligns with GUI spec in `docs/3-develop/gui/GUI.module4*.md`.
- [ ] Legends/units shown (µS, V, ns, dB) and numerically consistent.
- [ ] All tabs render, no blank panels.
- [ ] Accessibility: text alternatives exist for plots; keyboard navigation works.

**Evidence required:** attach screenshot or text dump of widget tree + parameter snapshot.

---

## 5) Cross-tool Validation (external references)

### 5.1 Circuit validation vs SPICE
- [ ] ngspice or PySpice sweep matches DAC/ADC curves within tolerance.
- [ ] Document netlist parameters and run scripts under `validation/external/`.

### 5.2 Array validation vs badcrossbar / CrossSim
- [ ] IR drop vs array size compared to `badcrossbar` (exact nodal analysis).
- [ ] MVM error metrics compared to CrossSim (GPU reference) when available.

---

## 6) Required Artifacts per Run
- [ ] `summary.json`
- [ ] `matrix.json`
- [ ] `thresholds.json`
- [ ] `results.json`
- [ ] `material_snapshot.json`
- [ ] `signal_chain_trace.json`
- [ ] `solver_diagnostics.json`
- [ ] `thermodynamics.json`
- [ ] `confidence_ledger.json`
- [ ] `provenance.json`
- [ ] `uncertainty.json`

---

## 7) Pass/Fail Criteria
- [ ] All thresholds from `docs/3-develop/automation/module4-automated-testing-plan.md` satisfied.
- [ ] Any missing artifact or insufficient N → FAIL.
- [ ] Any GUI control without model mapping → FAIL.

---

## 8) Next Actions (fill as work progresses)
- [ ] Populate DOI references for each circuit metric.
- [ ] Implement external-tool comparison scripts in `validation/external/`.
- [ ] Add GUI widget tree dump to CI artifacts for visual diff.
- [ ] Add dedicated Module 4 validation lane to `scripts/module4_automation.sh` output.

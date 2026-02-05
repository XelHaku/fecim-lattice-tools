---
Module: module4-circuits
Name: Unified GUI Controls — Semantics, Units, and Physics Meaning
Package: fecim-lattice-tools/module4-circuits/pkg/gui
Primary UI File: module4-circuits/pkg/gui/tab_unified.go
State Layer: module4-circuits/pkg/gui/device_state.go
Last Updated: 2026-02-05
---

# Module 4 Unified GUI — Control Semantics & Units (READ/WRITE/COMPUTE)

This document is a **control-by-control contract** for the unified Module 4 GUI (OPERATIONS view). It answers:

- What each control *means physically* (biasing, gating, or algorithmic intent)
- What units the user is entering/seeing (V, code, state index)
- What the bounds are and where clamping occurs

> Design intent: **Mode-first UX**. The active `OpMode` determines the *allowed voltage window* and *word-line activation pattern*.

## Shared definitions (used by multiple controls)

- **State / level index**: An integer `L` in `[0, quantLevels-1]` that represents a *discrete FeFET conductance / polarization state*. It is **not** a voltage.
- **Vc (coercive voltage)**: Derived from material hysteresis (`material.CoerciveVoltage()`), conceptually `Vc ≈ Ec × thickness`.
- **Read range**: `Vread ∈ [0, readRange.Max]` where `readRange.Max ≈ FieldMinRatio × Vc` (from `physics.yaml`), then bounded for practicality.
- **Write range**: `Vwrite ∈ [writeRange.Min, writeRange.Max]` where `writeRange.Min ≈ Vc` and `writeRange.Max ≈ FieldMaxRatio × Vc`.
- **MVM physics**: For compute, columns apply voltages `Vj` and each active row produces current `Ii = Σj(Gij × Vj)`.

File references:
- Mode + ranges: `pkg/gui/device_state.go` (`OpMode`, `VoltageRange`, `SetOperationMode`, `updateVoltageRanges`)
- UI wiring: `pkg/gui/tab_unified.go` and `pkg/gui/tab_unified_actions.go`

---

## Control: Mode buttons — READ / WRITE / COMPUTE

**UI:** `READ`, `WRITE`, `COMPUTE` buttons (`createUnifiedConfigModeRow`, `setOperationMode`).

**Physics meaning:** Selects the **biasing regime** (safe read, programming, or analog compute) and the **word-line activation pattern** (single-row vs all-rows).

### READ mode
- **Word lines (1T1R / 2T1R):** `WLSingle(selectedRow)` (one row enabled).
- **Word lines (0T1R passive):** all WLs effectively always on (no gating).
- **Bit lines / DAC:** all columns grounded except the selected column, which is driven to a small `Vread`.

**Units:** volts (V) on BL/DAC.

**Bounds / clamping:**
- `Vread` is derived from `readRange.Max` and should remain **well below** `Vc` to avoid disturb.
- In the UI, a typical default is ~`0.4 × readRange.Max`.

### WRITE mode
- **Word lines (1T1R / 2T1R):** `WLSingle(selectedRow)`.
- **Word lines (0T1R passive):** all WLs always active; **half-select mitigation required**.
- **Voltages:** the GUI stays at **0V idle** until the user presses **Program Cell**.

**Units:**
- Write target control is a **state index** (see next section).
- Applied pulses are **volts** (V), within the write range.

**Bounds / clamping:**
- Target level is clamped to `[0, quantLevels-1]`.
- Pulse voltage is clamped to the material-derived write range, and additionally to practical DAC limits.

### COMPUTE mode
- **Word lines (1T1R / 2T1R):** `WLAll()` (all rows enabled) to perform a dense MVM.
- **Word lines (0T1R passive):** naturally behaves like all-rows-active (no gating).
- **Bit lines / DAC:** per-column voltages come from the **input vector** (digital codes mapped to volts).

**Units:** input is a **byte-like code** (0–255), mapped to volts.

**Bounds / clamping:**
- Input codes are clamped to `[0, 255]` before conversion.
- Voltage mapping is clamped implicitly by using `readRange.Max` as the full-scale.

---

## Control: Architecture toggle — PASSIVE (0T1R) / 1T1R / 2T1R

**UI:** Architecture toggle buttons (`createArchitectureToggle`).

**Physics meaning:** Chooses the **array access structure**:

- **0T1R (PASSIVE):** no access transistor. WL gating is not available. Sneak paths exist.
- **1T1R:** per-row access transistor. WL gating isolates unselected rows.
- **2T1R:** planned / partial support; conceptually provides further isolation/decoupling.

**Bounds / enforcement:**
- In **0T1R**, WL selection controls are ignored at the state layer (`DeviceState.isPassive`) and the UI forces all WLs effectively active.
- In **1T1R/2T1R**, WL activation follows the current mode (`WLSingle` in READ/WRITE, `WLAll` in COMPUTE).

---

## Control: Program Cell (WRITE action)

**UI:** `Program Cell` button (`onUnifiedProgram`, `runISPPWithAnimation`, `applyWriteVoltages`).

**Physics meaning:** Executes a simplified **ISPP-style write/verify** loop:

1. Apply a programming pulse (`Vprog`)
2. Read/verify using a safe read bias
3. Adjust pulse amplitude and iterate until the target state is reached (or max iterations)

**Units:**
- Target is a **state index**.
- Pulses are **volts**.

**Bounds / clamping:**
- Target level clamped to `[0, quantLevels-1]`.
- Pulse voltages limited to the material-derived write range.

---

## Control: Target Write Level slider (WRITE panel)

**UI:** Slider in `createWriteModePanel` with label `L:n` and voltage preview.

**Physics meaning:** Sets the intended **final polarization/conductance state** of the selected cell.

- The slider value `L` is a **discrete state index**, not a voltage.
- The GUI computes a **preview voltage** using a linear map across the write range:

`Vpreview(L) = writeMin + (L/(N-1)) × (writeMax-writeMin)`

This preview represents the *nominal pulse amplitude* associated with that state; the actual ISPP loop may apply a sequence around this level.

**Units:**
- Slider: integer state index `L`.
- Preview label: volts (V).

**Bounds / clamping:**
- Slider clamps to integer steps.
- Computed preview voltage clamps to `[writeMin, writeMax]`.

---

## Control: Compute input vector entries (COMPUTE panel)

**UI:** `x0 … xN-1` entries, each expecting `0–255` (`createComputeModePanel`, `onInputVectorEntryChanged`).

**Physics meaning:** Defines per-column **applied input voltage** for MVM compute.

**Mapping (digital → analog):**
- Each entry is a code `d ∈ [0,255]` mapped linearly to the **read/compute-safe** voltage range:

`Vj = (d/255) × readRange.Max`

So:
- `0 → 0V`
- `255 → readRange.Max`

**Units:** input = unitless code; output = volts (V).

**Bounds / clamping:**
- UI clamps the code to `[0,255]`.
- The resulting voltage is bounded by construction to `[0, readRange.Max]`.

---

## Control: “Set All (V)” DAC entry (manual override)

**UI:** `Set All (V)` entry in the toolbar (`createDACInputSection`, `setAllUnifiedDACVoltages`).

**Physics meaning:** Applies a manual **bit-line bias** to *all columns*.

- In READ, this should typically be avoided (READ wants only one selected column biased; others grounded).
- In COMPUTE, this can be used as a quick way to set a uniform input vector.
- In WRITE, the safe default is 0V until **Program Cell** applies the write pulse.

**Units:** volts (V).

**Bounds / clamping notes:**
- For physical realism, manual voltages should respect the current mode’s range:
  - READ/COMPUTE: `0 … readRange.Max`
  - WRITE: `writeRange.Min … writeRange.Max` (when intentionally applying write biases)
- Additional hard caps may be applied (e.g., `MaxPracticalVoltage`) to avoid unrealistic voltages.

---

## Passive (0T1R) V/2 half-select pattern (WRITE)

When programming a passive array, unselected cells cannot be isolated. The GUI implements a **symmetric ±V/2 half-select scheme**:

Let the requested program amplitude be `Vwrite`.

- **Selected WL (target row):** `+Vwrite/2`
- **Selected BL (target col):** `−Vwrite/2`
- **All other WL/BL:** `0V`

Effective cell voltage is `Vcell = VWL − VBL`:

- **Target cell:** `(+V/2) − (−V/2) = V` (full switching)
- **Half-selected (same row):** `(+V/2) − 0 = V/2` (should be below switching threshold)
- **Half-selected (same col):** `0 − (−V/2) = V/2` (should be below switching threshold)
- **Unselected:** `0 − 0 = 0`

**Physics meaning:** Reduces write disturb by ensuring any non-target cell sees at most ~`V/2`.

**Bounds / clamping:** ensure `V/2 < Vc` for minimal disturb; this is one reason `writeRange.Min` should be chosen with margin.

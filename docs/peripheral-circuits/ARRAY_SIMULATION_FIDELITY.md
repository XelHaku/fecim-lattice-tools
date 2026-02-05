# Module 4 Array Coupling + Simulation Fidelity (Planned)

> **HONESTY_AUDIT:** This document describes **current behavior** where known and **planned improvements** where not yet implemented. Where implementation details are uncertain, text is marked **(planned / unvalidated)**.

Module 4 (peripheral circuits) sits at the boundary between:

- **digital controls** (DAC codes, row/column selection),
- a **crossbar array** (conductances + interconnect), and
- a **sense chain** (TIA → ADC).

This page defines **what “array coupling” means**, provides **architecture-aware rules** (0T1R / 1T1R / 2T1R), and introduces **fidelity tiers** for array simulation.

---

## 1) Definitions (Terms we use consistently)

### 1.1 Vcell (cell voltage)

**Vcell** is the voltage actually applied *across the storage element* of a single cell.

A useful abstraction is:

- Let the local electrical potentials at cell *(r,c)* be:
  - \(V_{WL}(r,c)\): local potential on the word-line node near that cell
  - \(V_{BL}(r,c)\): local potential on the bit-line node near that cell

Then the (signed) cell voltage is:

\[
V_{cell}(r,c) = V_{WL}(r,c) - V_{BL}(r,c)
\]

**Why it matters:** in non-ideal arrays, \(V_{WL}(r,c)\) and \(V_{BL}(r,c)\) can differ from the *driver* voltages because of:

- **IR drop** along line resistances,
- loading from other cells, and
- (in passive arrays) unintended conduction paths.

### 1.2 Sneak paths

A **sneak path** is any current path that contributes to the measured column/row current but **does not flow through the intended selected cell(s)**.

- Most relevant in **0T1R (passive)** arrays.
- Greatly reduced in **1T1R** because unselected cells are isolated by access transistors (ideally).

**Operational definition for simulation:**

- If a current flows through a cell that is not logically selected by the operation (READ/WRITE target), that current is treated as sneak contribution.

### 1.3 Half-select disturb

**Half-select disturb** occurs when *unselected* cells see a partial \(|V_{cell}|\) during a READ or WRITE.

- In passive arrays, write biasing (e.g., V/2 scheme) intentionally applies partial voltages to reduce disturb.
- In transistor-gated arrays, half-select disturb is mostly an issue of **leakage** and **gate/drain coupling** (not modeled in the current educational implementation).

**Rule of thumb (unvalidated, literature-style):** disturb risk increases as \(|V_{cell}|\) approaches the coercive switching region.

> DOI:(add) for half-select disturb analysis in FeFET/FeRAM-style crossbars.

### 1.4 IR drop (interconnect voltage drop)

**IR drop** is the voltage loss along the WL/BL metal lines due to finite line resistance.

For a 1D segment of resistance \(R_{seg}\) carrying current \(I\):

\[
\Delta V = I \cdot R_{seg}
\]

In arrays, this becomes a distributed network; the resulting \(V_{cell}(r,c)\) varies across the grid.

> **HONESTY_AUDIT:** The repo contains conceptual IR-drop discussion; a full DC nodal solver integration for Module 4 is **planned** (see fidelity tiers).

---

## 2) Architecture-aware array coupling (0T1R vs 1T1R vs 2T1R)

### 2.1 What “coupling” means here

For Module 4, “array coupling” is the mapping:

**(DAC voltages + word-line enables + stored conductances) → (measurable currents)**

That mapping depends on the cell access architecture.

### 2.2 0T1R (passive) coupling

In a passive array, there is no per-cell access device. In the simplest educational approximation, all cells on active lines contribute to the sensed current.

- **Pros:** minimal parameters; easy to teach \(I=GV\) and KCL summation.
- **Cons:** ignores or compresses sneak paths / IR drop into a single “error term” unless a nodal solver is used.

**Planned coupling improvement:**
- Add explicit bias schemes for READ and WRITE (e.g., V/2, V/3) and a DC nodal solver option to compute sneak contributions (planned / unvalidated).

### 2.3 1T1R coupling

In 1T1R, an access transistor gates each cell.

**Coupling consequence:**
- READ/WRITE: only the selected row (or selected rows) conducts if the access device is OFF elsewhere.
- COMPUTE: typically all access devices are ON (or a defined mask is ON), enabling full MVM.

**Modeling note:**
- The current simulator’s “1T1R” mode is primarily **behavioral** (row masking). It does **not** yet model transistor \(R_{on}\), \(R_{off}\), or body effect (planned).

### 2.4 2T1R coupling (planned)

2T1R is commonly used to decouple read and write paths:

- One transistor gates the **read** path, optimizing for low disturb and good sensing.
- A second transistor gates the **write** path, allowing higher voltages/currents without corrupting the sense circuitry.

**Planned simulator implication:**
- Separate enable masks for READ and WRITE.
- Potentially different effective series resistances in read vs write.

> DOI:(add) for 2T1R FeFET/RCIM array architecture references.

---

## 3) Fidelity tiers (how accurate is the array model?)

This project benefits from making “fidelity” explicit. Different tasks (GUI education vs research sweeps vs verification) need different cost/accuracy trade-offs.

### Tier 0 — Ideal (current baseline for many paths)

**Intent:** fastest, clearest teaching model.

Assumptions:
- No sneak paths
- No IR drop
- No line capacitance

Typical equation:

\[
I_{row}(r) = \sum_{c} G(r,c) \cdot V_{in}(c)
\]

### Tier 1 — Approx (lumped non-idealities)

**Intent:** introduce realism without a full circuit solve.

Examples of approximations (pick one; (planned)):
- Add a global sneak fraction: \(I_{meas} = I_{ideal} (1 + \epsilon_{sneak})\)
- Add a per-row effective voltage scaling for IR drop: \(V_{eff}(r) = \alpha(r) V_{in}\)

> **HONESTY_AUDIT:** Approx models can mislead if not calibrated; treat them as “didactic noise knobs” unless fit to measured data.

### Tier 2 — DC Nodal / MNA (Modified Nodal Analysis)

**Intent:** compute a physically consistent DC operating point of a resistive crossbar network.

- Nodes: WL/BL intersections (and potentially intermediate segments)
- Elements: cell conductances + line resistances
- Solve: \(A x = b\) for node voltages

Outputs:
- \(V_{cell}(r,c)\) everywhere
- exact DC currents including sneak paths and IR drop

> DOI:(add) for passive crossbar nodal methods (e.g., SoftwareX/badcossbar-style) and MNA references.

### Tier 3 — Transient (time domain)

**Intent:** capture settling, bandwidth, and dynamic effects.

Adds:
- line capacitances
- TIA finite bandwidth / settling
- ADC sample/hold timing

This tier is computationally heavier and is **planned** for focused verification, not default GUI operation.

---

## 4) Sense chain model: TIA + ADC and measurable current range

### 4.1 TIA model (transimpedance)

In Module 4, the TIA converts an input current to a voltage.

Idealized formula:

\[
V_{out} = I_{in} \cdot R_{TIA} + V_{offset}
\]

Units:
- \(I_{in}\): amperes (A)
- \(R_{TIA}\): ohms (Ω)
- \(V_{out}\): volts (V)

Practical constraints (typical in code models):

- **output clamp:** \(V_{out}\) is limited to \([V_{out,min}, V_{out,max}]\)
- **input clamp:** \(I_{in}\) may be limited to \([ -I_{max}, +I_{max}]\)

> **HONESTY_AUDIT:** The simulator’s TIA is a simplified behavioral model. Noise and bandwidth are modeled in aggregate; stability, op-amp internal poles, and layout parasitics are not modeled.

### 4.2 ADC quantization model

For an N-bit ADC with reference range \([V_{ref,low}, V_{ref,high}]\):

- LSB size:
\[
V_{LSB} = \frac{V_{ref,high} - V_{ref,low}}{2^N - 1}
\]

- Code mapping (one common convention):
\[
code = \mathrm{round}\left( \frac{V_{in}-V_{ref,low}}{V_{LSB}} \right)
\]

with clamp to \([0, 2^N-1]\).

### 4.3 Measurable current range (TIA + ADC together)

The sense chain imposes a **measurable input current range**.

Ignoring input clamp for a moment, the ADC input range corresponds to a current range:

\[
I_{min,ADC} = \frac{V_{ref,low} - V_{offset}}{R_{TIA}}
\]
\[
I_{max,ADC} = \frac{V_{ref,high} - V_{offset}}{R_{TIA}}
\]

Then include TIA input limit (if modeled):

\[
I_{max,meas} = \min(I_{max,ADC}, I_{max,TIA})
\]

**Example (illustrative):**
- \(R_{TIA}=10\,k\Omega\)
- \(V_{ref,low}=0\,V\), \(V_{ref,high}=1\,V\)
- \(V_{offset}=5\,mV\)

\[
I_{max,ADC} \approx \frac{1.0-0.005}{10\,000} = 9.95\times 10^{-5} \; A = 99.5\,\mu A
\]

This aligns with the typical “~100 µA full scale” behavior.

> **HONESTY_AUDIT:** Whether these numbers match the current defaults depends on the exact values in `shared/peripherals`. Treat the example as a unit-checked derivation, not a measured spec.

### 4.4 Implication for array sizing

As array dimensions grow, worst-case row/column currents can exceed \(I_{max,meas}\), causing:

- TIA saturation → ADC rail codes
- reduced effective resolution

Planned improvement:
- Provide an analysis utility: given \(G_{max}\), \(V_{in,max}\), and active columns, estimate worst-case \(I\) and warn when the sense chain saturates.

---

## 5) Testing strategy (headless, table-driven)

Module 4 already has good unit tests for individual peripherals. Planned additions focus on **architecture-aware coupling** and **fidelity tiers**.

### 5.1 Goals

- Tests run headless: `go test ./...` (no GUI)
- Table-driven cases with explicit units
- Deterministic: no randomness unless seeded

### 5.2 Suggested test categories (planned)

1. **Sense-chain range tests**
   - Given TIA+ADC defaults, compute the theoretical \(I_{max,meas}\) and verify clamping behavior.

2. **Architecture coupling tests**
   - Same weight matrix + same DAC voltages:
     - 0T1R (all-active) vs 1T1R (masked rows) should produce different currents.

3. **Fidelity tier golden tests**
   - Small arrays (e.g., 2×2, 3×3) with known resistances.
   - Tier 2 (DC nodal) results compared against precomputed values.

### 5.3 Table-driven pattern (Go)

```go
func TestSenseChain_CurrentToCode(t *testing.T) {
    type tc struct {
        name string
        iInA float64
        wantCode int
    }

    tests := []tc{
        {name: "zero", iInA: 0, wantCode: 0},
        {name: "fullscale", iInA: 100e-6, wantCode: 31},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // compute: V = TIA(i), code = ADC(V)
        })
    }
}
```

> **HONESTY_AUDIT:** The exact expected codes depend on rounding convention and clamp rules. If conventions change, update tests to match the documented convention.

---

## 6) Where this plugs into the repo

- Peripheral models: `shared/peripherals/*`
- Module 4 GUI behavioral state: `module4-circuits/pkg/gui/device_state.go`
- Crossbar / array research notes: `docs/internal-analysis/crossbar-arrays.md`
- Operations doc (0T1R vs 1T1R): `docs/peripheral-circuits/circuits.operations.md`

---

## 7) TODO checklist (documentation-driven)

- [ ] Add explicit “fidelity tier” selector to GUI (planned)
- [ ] Add DC nodal solver for passive sneak paths (planned)
- [ ] Implement 2T1R masks (planned)
- [ ] Add headless test suite for coupling + tiers (planned)


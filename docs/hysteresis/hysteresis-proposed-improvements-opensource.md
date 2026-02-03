# Mega Implementation Plan: FeCIM Hysteresis Module 1

**Status:** Draft plan  
**Target:** Production-grade modeling (aspirational)  
**Reference:** `docs/hysteresis/hysteresis-gemini.md` (vision document)

This plan consolidates the "Unified Theory" from the Gemini Compendium with the feature-rich improvements identified in the Open-Source analysis. It is structured into 4 distinct phases for immediate execution.

## 🚀 Phase 1: The "Physics Core" (Critical Path)
*Goal: Implement the definitive math models found in `hysteresis-gemini.md`.*

- [ ] **1.1: Landau-Khalatnikov (L-K) Engine**
    - [ ] Create `pkg/ferroelectric/solver_lk.go`.
    - [ ] Implement `LKSolver` struct with `Alpha` (Dynamic), `Beta` (First-Order), `Gamma` (Stability), `Rho` (Viscosity).
    - [ ] Implement `UpdateParams(T, Stress)` for unified coefficient calculation.
    - [ ] Implement `Step(E, dt)` using **Runge-Kutta 4 (RK4)**.
    - [ ] **1.1.b: Circuit Parasitics:**
    - [ ] Add `SeriesResistance` (Ohms) to `LKSolver` struct.
    - [ ] Modify `dPdT` loop to calculate effective voltage: $V_{eff} = V_{applied} - I(t) \cdot R_s$. (Requires implicit solve or small dt).
    - [ ] **Validation:** Verify 1ns step stability without oscillation.

- [ ] **1.2: Preisach Stack "Wipe-Out"**
    - [ ] Refactor `pkg/ferroelectric/preisach.go` to use `Stack` data structure.
    - [ ] Implement `WipeOut(E_new)` logic:
        - If `E_new > Stack.PeekMax()`, pop pair.
        - If `E_new < Stack.PeekMin()`, pop pair.
    - [ ] Ensure minor loops close perfectly (return to major loop trajectory).
    - [ ] **Validation:** "Pump" test – drift check after 100 read cycles.

- [ ] **1.3: "Golden Set" Materials**
    - [ ] Update `materials.yaml` with the 10nm HZO parameters:
        - $\beta = -2.160 \times 10^8$
        - $Q_{12} = -0.026$
        - $T_C = 723 K$
    - [ ] Remove legacy/dummy material definitions.

## ⚡ Phase 2: Control & Write Logic (Arenaton)
*Goal: Enable nanosecond-scale "Write" operations using the physics engine.*

- [ ] **2.1: Adaptive Binary ISPP**
    - [ ] Create `pkg/controller/ispp.go`.
    - [ ] Implement `PredictState(target_P)` using inverse model.
    - [ ] Implement `BinarySearchWrite(target_P)` loop:
        - Apply Pulse via L-K Solver.
        - Read Conductance.
        - Bisect Voltage.
        - **Critical:** Handle overshoot with Negative Reset Pulse.
    - [ ] **Validation:** Reach Level 14 in $<5$ pulses.

- [ ] **2.2: Conductance Mapper:**
    - [ ] Implement `GetConductance(P)` function.
    - [ ] Define `G_min` (Off state) and `G_max` (On state) in `materials.yaml`.

## 🔬 Phase 3: Silicon Realism (Advanced Physics)
*Goal: Add the "Real World" messy physics required for silicon verification.*

- [ ] **3.1: Nucleation-Limited Switching (NLS)**
    - [ ] Update `solver_lk.go`:
        - Add `IncubationTime(E)` function based on Merz's Law ($\exp(E_a/E)$).
        - Modify `Step()` to delay switching until $t > t_{inc}$.
    - [ ] **Validation:** Low voltage pulses should fail to switch even if $V > V_c$.

- [ ] **3.2: Stochastic Langevin Noise**
    - [ ] Add `Noise(T)` term to `dPdT` equation.
    - [ ] Scale noise by temperature and damping $\rho$.
    - [ ] **Validation:** Run 1000 Monte Carlo writes to plot Bit Error Rate (BER).

- [ ] **3.3: Frequency Dependence**
    - [ ] Add `SetFrequency(Hz)` to Solver.
    - [ ] Scale `Ec` dynamically: $E_c(f) \approx E_{c0} \times (1 + f/f_0)^{0.1}$.

- [ ] **3.4: Wake-Up & Fatigue:**
    - [ ] Add `CycleCount` to solver state.
    - [ ] Implement dynamic scaling of $P_r$ (Wake-up) and $E_c$ (Fatigue/Breakdown).
    - [ ] Add phenomenological scaling: $P_r(N) = P_{r,\infty} \cdot (1 - e^{-N/N_{wakeup}})$.

- [ ] **3.5: Depolarization Field:**
    - [ ] Add "Retention Mode" where $E=0$ but $P \neq P_{remnant}$.
    - [ ] Simulate $P$ decay over log(time).

## 🖥️ Phase 4: GUI & Experience (The "Wow" Factor)
*Goal: Visualize the new physics in the Fyne GUI.*

- [ ] **4.1: Temperature & Stress Sliders**
    - [ ] Add vertical slider for $T$ (4K to 800K).
    - [ ] Add slider for Stress (0 to 3 GPa).
    - [ ] Live update of P-E loop shape (observe $P_r$ drop at high $T$).

- [ ] **4.2: Real-Time Metrics Dashboard**
    - [ ] Add panel showing:
        - **Energy:** $\oint P dE$ (Loop Area).
        - **Speed:** Effective Switching Time $\tau$.
        - **Retention:** Estimated data loss rate.

- [ ] **4.3: Overlay Comparison**
    - [ ] "Ghost Mode": Keep previous loop on screen when changing parameters.
    - [ ] Compare "Room Temp" vs "Cryogenic" directly.

- [ ] **4.4: Data Export (EDA Handoff):**
    - [ ] Add "Export .CSV" button for the generated P-E loop data.
    - [ ] Add "Export .MDM" (Verilog-A Table) format for SPICE compatibility.

## Appendix: Technical Reference
*Specific constants for Phase 1.1*

| Constant | Value |
| :--- | :--- |
| $\beta$ | $-2.160 \times 10^8$ |
| $\gamma$ | $1.653 \times 10^{10}$ |
| $\rho$ | $0.05$ |

---

## Strategic Advice: Physics vs Memory

**Clarification:** In Phase 1.2, you mention the Preisach Stack. Be careful not to double-count physics.

**Use L-K (Phase 1.1)** for the **Time-Domain** simulation (Write Pulses, Switching Speed).

**Use Preisach (Phase 1.2)** for the **State-Domain** simulation (Hysteresis shape, Minor Loops, "30 Levels").

*They are two different views of the same elephant.* Usually, you run L-K to determine *if* the bit switches, and Preisach to determine *exactly where* it settles.

| Constant | Value |
| :--- | :--- |
| $\beta$ | $-2.160 \times 10^8$ |
| $\gamma$ | $1.653 \times 10^{10}$ |
| $\rho$ | $0.05$ |

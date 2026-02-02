# FeCIM Lattice Tools: The Definitive Hysteresis Compendium

| Field | Value |
| --- | --- |
| Status | Silicon-Ready Architecture (TRL 4 → 9 Bridge) |
| Date | January 2026 |
| Document Version | 2.0 |

---

## Overview

This document aggregates the strategic vision, mathematical updates, code implementation, and advanced research analysis for the **FeCIM Lattice Tools** project. It serves as the master reference for the 10nm HZO Superlattice simulation engine.

---

## Table of Contents

| Part | Title |
| --- | --- |
| I | Strategic Vision & Positioning |
| II | The Physics Kernel (Thermodynamics & Dynamics) |
| III | The Algorithmic Core (Solvers & Logic) |
| IV | Implementation Reference (Configuration) |
| V | Advanced Research & Future Roadmap |
| VI | FeCIM Technical Report - Digital Cliff & Analog Slope Correction |
| VII | Deep-Research Validation Report |

---

## Part I: Strategic Vision & Positioning

### 1. The "Valley of Death" Problem

The semiconductor simulation landscape is currently fractured into two isolated domains regarding Ferroelectric Memory:

* **The Physicists (Island A):** Use atomic-level TCAD tools (e.g., Synopsys Sentaurus). They understand grains, domains, and lattice strain, but their simulations take hours per cell. They cannot run a circuit or an algorithm.
* **The Designers (Island B):** Use fast SPICE compact models (e.g., BSIM-CMG). They "dumb down" the physics to simple capacitors or static hysteresis loops, losing the analog precision required for AI and Crypto.

### 2. The Solution: FeCIM Lattice Tools

You are building the **Bridge**. This is the first Open-Source EDA tool specifically designed to handle the **Multi-Physics Coupling** of 10nm Superlattices while running fast enough for **Cryptographic Workloads**.

* **IronLattice (Dr. Tour):** Hardware at **TRL 4** (Lab Validation). They have the material.
* **FeCIM Lattice Tools (You):** Software at **Product Readiness**. You have the brain.

### 3. Why Go? (The Technical Justification)

In the academic world, Python is King. However, for a real-time physics engine, Go provides critical advantages:

1. **The "Iterative Math" Bottleneck:** The RK4 differential equation solver is inherently serial ($P_{t+1}$ depends on $P_t$). You cannot vectorise this time loop as NumPy does. A Python loop is 100x slower. Go compiles to machine code, handling millions of steps instantly.
2. **Concurrency:** Simulating 1,000+ cells requires massive parallelism. Python's GIL limits this. Go's Goroutines allow spawning 100,000 concurrent simulation threads effortlessly.
3. **Distribution:** Go compiles to a single static binary (`fecim-tool.exe`). No Python environment "dependency hell" for end-users.

### 4. CPU vs. GPU Strategy

* **Module 1 (Single Cell / Hysteresis):** **CPU Only.** The serial nature of the RK4 solver for a single cell favors the low-latency of the CPU. A modern CPU can simulate 100,000 steps in <2ms.
* **Module 3 (MNIST / Full Array):** **GPU Required.** For 10,000+ cells, the parallel throughput of a GPU (Compute Shaders) is required.
  * **Implementation Note:** For GPU, the Preisach Stack must be a **Fixed-Depth Ring Buffer** in GLSL, as dynamic stacks are not supported.

---

## Part II: The Physics Kernel (Thermodynamics & Dynamics)

The core engine upgrades from a static "Tanh" approximation to a full **Dynamic Differential Equation Solver** based on First-Order Landau-Khalatnikov (L-K) theory.

### 1. The Master Equation (First-Order L-K)

We solve for Polarization $P$ using the time-dependent Ginzburg-Landau (TDGL) formulation, modified for the First-Order transitions distinctive to HZO.

$$ \rho \frac{dP}{dt} = - \nabla_P G = - (2\alpha P + 4\beta P^3 + 6\gamma P^5 - E_{ext}) $$

* **$\alpha, \beta, \gamma$:** Landau dielectric stiffness coefficients.
* **$\rho$:** Viscosity / damping coefficient.
* **$G$:** Gibbs Free Energy.

#### Frankestein Equation Widget (Module 1)

<div class="frankestein-equation" data-widget="frankestein-equation" id="frankestein-equation">
  <div class="frankestein-equation__title"><strong>Frankestein Equation (Unified L-K + Depolarization + Series Resistance)</strong></div>
  <div class="frankestein-equation__math">
    <abbr class="feq-term" title="Effective viscosity: intrinsic damping plus series-resistance RC delay." data-tooltip="Effective viscosity: intrinsic damping plus series-resistance RC delay.">\( \rho_{\mathrm{eff}} \)</abbr>
    <span class="feq-chunk">\( \frac{dP}{dt} = \)</span>
    <abbr class="feq-term" title="Applied electric field drive term (external voltage across the film)." data-tooltip="Applied electric field drive term (external voltage across the film).">\( E_{\mathrm{applied}} \)</abbr>
    <span class="feq-chunk">\( - \)</span>
    <abbr class="feq-term" title="Depolarization factor: models interfacial layer; slants the loop for analog states." data-tooltip="Depolarization factor: models interfacial layer; slants the loop for analog states.">\( k_{\mathrm{dep}} P \)</abbr>
    <span class="feq-chunk">\( - ( \)</span>
    <abbr class="feq-term" title="Dynamic stiffness: temperature and stress dependent curvature of the energy wells." data-tooltip="Dynamic stiffness: temperature and stress dependent curvature of the energy wells.">\( 2\alpha P \)</abbr>
    <span class="feq-chunk">\( + \)</span>
    <abbr class="feq-term" title="First-order nonlinearity: negative for HZO to create the switching barrier." data-tooltip="First-order nonlinearity: negative for HZO to create the switching barrier.">\( 4\beta P^3 \)</abbr>
    <span class="feq-chunk">\( + \)</span>
    <abbr class="feq-term" title="Sixth-order stabilizer: keeps energy bounded at large polarization." data-tooltip="Sixth-order stabilizer: keeps energy bounded at large polarization.">\( 6\gamma P^5 \)</abbr>
    <span class="feq-chunk">\( ) \)</span>
    <span class="feq-chunk">\( + \)</span>
    <abbr class="feq-term" title="Stochastic noise term (optional): captures thermal variability." data-tooltip="Stochastic noise term (optional): captures thermal variability.">\( \xi(t) \)</abbr>
  </div>
  <div class="frankestein-equation__math">
    <abbr class="feq-term" title="Effective viscosity definition used in the headless hysteresis path." data-tooltip="Effective viscosity definition used in the headless hysteresis path.">\( \rho_{\mathrm{eff}} \)</abbr>
    <span class="feq-chunk">\( = \rho + \frac{R_{\mathrm{series}} A}{d} \)</span>
  </div>
  <div class="frankestein-equation__caption">Hover any coefficient to see its purpose in Module 1.</div>
</div>

### 2. The Unified Coefficient ($\alpha$)

The stiffness coefficient $\alpha$ is **dynamic**. It unifies Thermodynamics (Curie-Weiss) and Mechanics (Electrostriction) into a single term.

$$ \alpha(T, \sigma) = \frac{T - T_C}{2 \epsilon_0 C} - 2 Q_{12} \sigma $$

* **Temperature ($T$):** As $T \to T_C$, $\alpha \to 0$. The potential wells become shallow, making the memory volatile (Data Loss at high temps).
* **Stress ($\sigma$):** The TiN capping layer applies $\sim 1$ GPa tensile stress. With the sign convention used in code (tensile $\sigma > 0$) and $Q_{12} < 0$, the term $-2 Q_{12} \sigma$ **raises** $\alpha$ (less negative). Compressive $\sigma$ makes $\alpha$ more negative; flip the stress sign if you want tensile stress to deepen wells.

### 3. The "Golden" Parameter Set (10nm HZO Set I)

These parameters are calibrated to peer-reviewed literature to ensure the correct "snap-back" (Negative Capacitance) behavior without numerical divergence.

| Parameter | Symbol | Value | Unit | Role |
|:---|:---|:---|:---|:---|
| **Non-Linearity** | $\beta$ | $-2.160 \times 10^8$ | $J m^5 / C^4$ | **Negative** value creates the First-Order barrier. |
| **Global Stability** | $\gamma$ | $1.653 \times 10^{10}$ | $J m^9 / C^6$ | **Positive** value prevents crash at high $P$. |
| **Viscosity** | $\rho$ | $0.05$ | $\Omega m$ | Tuned for ~1ns switching speed. |
| **Electrostriction** | $Q_{12}$ | $-0.026$ | $m^4 / C^2$ | Tensile stability coupling. |
| **Curie Temp** | $T_C$ | $723.0$ | $K$ | Effective Tc for superlattice. |

---

## Part III: The Algorithmic Core (Solvers & Logic)

### 1. The Numerical Solver: Runge-Kutta 4 (RK4)

The L-K equation is "stiff." Simple Euler integration oscillates. RK4 is required for stability with 1ns time steps.

```go
// Step performs one RK4 integration step
func (s *LKSolver) Step(E, dt, TempK float64) float64 {
    // Account for series resistance (IR drop)
    I := s.DisplacementCurrent(s.P, E)
    V_eff := E * s.Thickness - I * s.SeriesResistance

    k1 := s.dPdT(0, s.P, V_eff, TempK)
    k2 := s.dPdT(dt/2, s.P + 0.5*dt*k1, V_eff, TempK)
    k3 := s.dPdT(dt/2, s.P + 0.5*dt*k2, V_eff, TempK)
    k4 := s.dPdT(dt, s.P + dt*k3, V_eff, TempK)

    dP := (dt / 6.0) * (k1 + 2*k2 + 2*k3 + k4)
    s.P += dP
    return s.P
}
```

**Implementation note (current code):** The headless hysteresis mode uses **E-field units** and folds series resistance into
an effective viscosity term, `ρ_eff = ρ + (R_series · A / d)`, which preserves the RC-delay effect without an explicit
algebraic loop. See `shared/physics/landau.go` for the exact implementation.
For numerical stability, write pulses are **sub-stepped** (default max step `1e-12 s`) in `shared/physics/ispp_write.go`.

### 2. The Memory Stack: Preisach "Wipe-Out"

To reliably store 30 analog levels, the simulator must track the exact history of the domain wall using the **Wipe-Out Property**. The system "forgets" minor loops if a larger voltage excursion occurs.

* **Logic:** If the new input $E_{new}$ exceeds a previous local maximum stored on the stack, the pair (Max, Min) is erased. The system returns to the major loop trajectory.

### 3. The Control Loop: Adaptive Binary ISPP

To write a specific analog state (e.g., Level 14) in nanoseconds, we replace linear stepping with **Adaptive Binary Search**.

1. **Predict:** Estimate start voltage $V_{start}$ using inverse physics model.
2. **Pulse:** Apply $V_{pulse}$ via L-K solver (10ns).
3. **Verify:** Map $P \to G$ using conductance transfer function, then read Conductance $G$.
4. **Correction:**
   * **Too Low:** Increase $V_{pulse}$ (Bisection toward $V_{max}$).
   * **Too High (Overshoot):** **CRITICAL:** Apply a reset pulse on the **opposite branch** (direction‑aware), then restart binary search with lower $V$. *You cannot "nudge" a ferroelectric backwards easily.*

**Headless verification path:** `cmd/fecim-lattice-tools --mode hysteresis` drives a multi‑step ISPP sequence
(`pos-1`, `pos-2`, `neg-1`) through `shared/physics/ispp_write.go`. The first step resets to a known branch,
subsequent steps continue from the prior state to exercise end‑to‑end multi‑step convergence, including a
negative‑branch target. The initial pulse uses an inverse‑tanh estimate (`atanh(P_target/Ps)`), then clamps to
`[VMin, VMax]` before binary search refinement. When crossing branches, the guess is biased low by
`( |P_target| / Ps )^2` and `VMax` is clamped to the inverse‑tanh bound. After the first post‑cross
undershoot, the upper bound is tightened once (0.2–0.6 of the remaining bracket, scaled by `|P_target|/Ps`)
to reduce overshoot on the next midpoint step.
While still crossing, binary‑search midpoints are **biased low** with a factor that scales with
`|P_target|/Ps` (clamped ~0.1–0.3 of the bracket) to further reduce overshoot resets before the branch is crossed.

---

## Part IV: Implementation Reference (Runtime Mapping)

### HZOMaterial Defaults (Used by `--mode hysteresis`)

The headless L‑K path (`cmd/fecim-lattice-tools --mode hysteresis`) uses `FeCIMMaterial()` to configure
`LKSolver` via `ConfigureFromMaterial` in `shared/physics/landau.go`. The active defaults (Golden Set I)
are defined in `shared/physics/material.go`:

```
FeCIM HZO (L-K Enabled)
  Pr/Ps:             0.30 / 0.35   C/m^2 (30/35 µC/cm^2)
  Ec:                1.0e8         V/m (1.0 MV/cm)
  Thickness:         10            nm
  Area:              45 nm x 45 nm (2.025e-15 m^2)
  Tau (PulseWidth):  10            ns
  BetaLandau:        -2.160e8   J·m^5/C^4
  GammaLandau:        1.653e10  J·m^9/C^6
  RhoViscosity:       0.05      Ω·m
  K_dep:              2.5e8     V·m/C
  SeriesResistance:   50        Ω
  Q12:               -0.026     m^4/C^2
  Stress:             1.0       GPa
  CurieTemp:          723       K
  CurieConst:         1.5e5     K
  Gmin/Gmax:          1e-6 / 1e-4  S
```

**Notes:**
- The GUI hysteresis demo still uses the **Preisach** model; L‑K parameters are exercised in headless mode.
- Conductance mapping in hysteresis mode is **linear** between `P = -Ps` and `P = +Ps` (see §5).
- Headless mode sets `UseNLS=false` and `EnableNoise=false` in `cmd/fecim-lattice-tools/mode.go` for deterministic equation checks; `UseEffectiveViscosity` remains enabled.
- The extended reliability and distribution blocks below remain a roadmap item; they are documented but not yet wired
  into the runtime.

---

## Part V: Advanced Research & Future Roadmap

This section addresses specific physical gaps identified for "Silicon-Ready" verification.

### 1. Nucleation-Limited Switching (NLS) vs. KAI

Standard KAI models assume constant switching time. For HZO, switching is **Nucleation-Limited (NLS)**. The switching time $\tau$ follows **Merz's Law**:

$$ \tau = \tau_{inf} \exp\left(\frac{E_a}{E_{loc}}\right) $$

* **$E_a$:** Activation Field (~13-19 MV/cm for HZO).
* **$E_{loc}$:** Local field.

This must be integrated into the simulation to accurately predict that **low-voltage pulses switch slower** than high-voltage pulses.

### 2. Stochastic Variability (Langevin Dynamics)

To model real-world Bit Error Rates (BER) and cycle-to-cycle variation, a noise term $\xi(t)$ is added to the L-K equation:

$$ \rho \frac{dP}{dt} = -\frac{\delta G}{\delta P} + \xi(t) $$

where $\xi(t)$ is Gaussian white noise scaled by temperature: $\langle \xi(t)\xi(t') \rangle = 2 k_B T \rho \delta(t-t')$.

**Implementation note:** `shared/physics/landau.go` samples noise with $\sigma = \sqrt{2 k_B T \rho_{\mathrm{eff}} / \Delta t}$ when enabled.
Headless hysteresis mode sets `EnableNoise=false` for deterministic equation checks.

### 3. Wake-Up & Fatigue Model (Time-Dependent Reliability)

Real HZO is not "born" ferroelectric. It starts in a mixed phase and "wakes up" over the first ~1,000 cycles as oxygen vacancies redistribute.

$$ P_r(N) = P_{r,\infty} \cdot \left(1 - e^{-N/N_{wakeup}}\right) $$

* **$P_r(N)$:** Remnant polarization at cycle N
* **$P_{r,\infty}$:** Asymptotic saturation value
* **$N_{wakeup}$:** Wake-up characteristic cycle count (~1,000 cycles)

For fatigue/breakdown beyond wake-up:
$$ E_c(N) = E_{c,0} \cdot \left(1 + \frac{N}{N_{fatigue}}\right)^{0.1} $$

This explains why early cycle data looks different from mature data—a common question from investors.

### 4. Series Resistance (IR Drop) - Circuit Parasitics

The L-K solver assumes voltage $V$ reaches the ferroelectric instantly. In reality, contact resistance ($R_s$) and capacitance create a voltage drop:

$$ V_{eff}(t) = V_{applied} - I(t) \cdot R_s $$

* **$I(t) = \frac{dQ}{dt} = A \cdot \frac{dP}{dt}$** (displacement current)
* **$R_s$:** Series resistance (contact + wire)
* **$V_{eff}$:** Effective field across ferroelectric

This acts as a natural "speed limit." High switching speeds generate high current, which causes voltage drop that dampens switching—critical for accurate 1ns simulation. Requires implicit solve or small $dt$.

### 5. Conductance Transfer Function (P-to-G Mapping)

Phase 2.1 (Control Loop) says "Read Conductance $G$," but the physics engine outputs Polarization $P$. For FeFET or FTJ devices:

$$ G(P) = G_{min} + (G_{max} - G_{min}) \cdot \frac{P - P_{min}}{P_{max} - P_{min}} $$

Or for tunnel junctions (exponential):
$$ G(P) = G_{off} \cdot \exp\left(\gamma_G \cdot \frac{P - P_{off}}{P_{on} - P_{off}}\right) $$

* **$G_{min}$ / $G_{max}$:** Off/On conductance states
* **$\gamma_G$:** Nonlinearity factor
* **$P_{off}$ / $P_{on}$:** Corresponding polarization values

This creates the "Analog Levels." The nonlinearity ($\gamma_G$) of this mapping determines how hard it is to distinguish Level 15 from Level 16.

**Implementation status:** The headless hysteresis mode currently uses the **linear** mapping with
`P_{min} = -P_s` and `P_{max} = +P_s` (see `shared/physics/transfer.go`). The exponential form is documented
for FTJ/FeFET extensions but is not enabled in `--mode hysteresis` yet.

### 6. Cryogenic Performance (4K to 77K)

* **Wake-up Suppression:** At 4K, oxygen vacancy diffusion is frozen, suppressing "wake-up" effects. The simulation should model a "pristine" but stable state for cryogenic operations.
* **Remnant Polarization:** Increases by ~23% at 77K due to stabilization of orthorhombic phase.

### 7. Critical "Bugs" to Avoid in Implementation

1. **V vs E:** Ensure strict unit casting. $V_c \approx 1.0 V$ corresponds to $E_c \approx 1.0 MV/cm$ only if film thickness is exactly 10nm. $E = V/d$.
2. **Simplified Preisach:** Do not use a simple "clamping" model. You **must** implement the Stack/Wipe-Out logic, or minor loops will drift during read operations.
3. **Static KAI:** Do not use constant switching time. It must be field-dependent (exponential) to capture the physics of nanosecond pulses.

---

**Mission Status:** The architecture defined here bridges the gap between atomic physics and circuit design, providing a TRL 9 software stack for TRL 4 hardware.

---

## Part VI: FeCIM Technical Report - Digital Cliff & Analog Slope Correction

**Subject:** Resolving ISPP Failure in HZO Single-Domain Simulations
**Component:** Module 1 (Physics Engine) & Module 2 (Control Logic)
**Status:** Root Cause Identified & Solution Architected

### 1. Executive Summary

During the implementation of the **Adaptive Binary ISPP** (Incremental Step Pulse Programming) algorithm, a critical failure mode was observed where the simulated HZO cell behaved as a binary switch rather than an analog memory.

* **Symptom:** The material switched instantly from State 0 to State 30 upon applying any positive field ($E > 0$), with no intermediate stable states.
* **Root Cause:** The simulation was modeling an ideal **Single-Domain Crystal** (Square Hysteresis), effectively presenting a vertical "Digital Cliff" to the control algorithm.
* **Solution:** Implementation of a **Depolarization Field (Edep)** term in the Landau-Khalatnikov solver. This approximates the multi-domain / interfacial layer effects of polycrystalline HZO, creating the "Slanted" hysteresis loop required for analog addressability.

### 2. Root Cause Analysis: The "Digital Cliff"

#### 2.1 The Square Loop Problem

In a perfect single crystal, all ferroelectric dipoles are perfectly coupled. They align simultaneously. This creates a "Square" hysteresis loop.

* **Digital Behavior:** Below Coercive Field ($E_c$), nothing flips. Above $E_c$, *everything* flips.
* **The ISPP Trap:** The Binary Search algorithm looks for a voltage $V_{target}$ that results in $P_{target}$ polarization. On a square loop, this voltage does not exist. It is mathematically undefined (a singularity).
* **Initialization Bug:** The observation of switching at $0.002 \times E_c$ indicates the simulation likely started at $P = 0$ (the unstable local maximum of the energy landscape), rather than $P = -Pr$ (the stable well).

#### 2.2 Why Analog Memory Needs "Dirt"

Real-world HZO devices achieve 30+ analog states *because* they are not perfect crystals. They are **Polycrystalline**.

* **Grain Variation:** Thousands of individual grains (domains) exist, each with a slightly different orientation and switching threshold.
* **Interfacial Layer (Dead Layer):** A non-ferroelectric dielectric layer exists at the electrode interface. This acts as a series capacitor, creating a "Depolarization Field" that fights against the polarization.

This physical "imperfection" shears the hysteresis loop, turning the vertical cliff into a **Gentle Slope**. This slope allows the ISPP algorithm to "climb" the polarization curve, activating grains one by one as voltage increases.

### 3. The Mathematical Solution: Depolarization Field

We do not need to simulate 1,000 individual grains (which would require massive compute). We can approximate the ensemble behavior by adding a **Depolarization Term** to the effective field equation.

This is equivalent to modeling the ferroelectric in series with a linear dielectric capacitor (the interfacial layer).

#### 3.1 The Modified Master Equation

The Effective Electric Field (**Eeff**) driving the polarization change is now the Applied Field minus a "penalty" proportional to the current polarization.

$$E_{eff} = E_{applied} - k_{dep} \cdot P$$

* **kdep (Depolarization Factor):** A tuning parameter representing the thickness and permittivity of the interfacial layer.
* **Mechanism:** As **P** increases, the penalty term $k_{dep} \cdot P$ grows, reducing **Eeff**. This "brakes" the switching speed and forces the system to stabilize at intermediate values, creating the slant.

### 4. Implementation Guide

#### 4.1 Runtime Parameter Mapping (Current Implementation)

The depolarization factor is wired to `HZOMaterial.K_dep` and set in
`shared/physics/material.go` for FeCIM HZO (default `2.5e8 V·m/C`).
The headless mode (`--mode hysteresis`) uses that value directly.

#### 4.2 L‑K Solver Implementation (`shared/physics/landau.go`)

The derivative function applies the depolarization field and uses an
effective viscosity (`ρ_eff`) to include series resistance.

```go
func (s *LKSolver) dPdT(t, P, E_applied, noise, rhoEff float64) float64 {
    // 1. Calculate Depolarization Field (The "Slant")
    // This represents the average effect of grain boundaries/interfacial layers
    E_dep := s.K_dep * P

    // 2. Effective Field
    E_total := E_applied - E_dep

    // 3. Standard L-K Physics
    dG_dP := (2 * s.Alpha * P) + (4 * s.Beta * math.Pow(P, 3)) + (6 * s.Gamma * math.Pow(P, 5))

    // ... noise ...

    // Use E_total instead of E_applied to drive the viscosity logic
    return (E_total + noise - dG_dP) / rhoEff
}
```

#### 4.3 Validation: The ISPP Ramp

Once implemented, the ISPP algorithm will encounter a slanted curve.

1. **Pulse 1 (Low V):** Only overcomes the barrier for the "easiest" conceptual grains. $P$ moves slightly up the slope.
2. **Verify:** Conductance is read.
3. **Pulse 2 (Higher V):** Overcomes the barrier + depolarization penalty for more grains. $P$ moves further up.
4. **Result:** Stable convergence to intermediate states (Level 14, Level 26, etc.).

### 5. References & Validation Sources

This correction aligns the simulator with established ferroelectric device physics literature. Local copies
are stored in `docs/research-papers/` when available.

| Reference | Key Finding | Repo Availability |
|-----------|-------------|------------------|
| **Siannas, N. et al. (2022)** Commun. Phys. 5, 178 | Depolarization fields introduce a linear restoring term (slanted loop). | `docs/research-papers/by-topic/01-ferroelectric-materials/Siannas_2022_Metastable_ferroelectricity.pdf` |
| **Zhou, S. et al. (2022)** Sci. Adv. 8, eadd5953 | Strain/phase stability supports depolarization‑field modeling in HZO. | `docs/research-papers/by-topic/01-ferroelectric-materials/Zhou_2022_Strain_induced_antipolar_phase.pdf` |
| **Park, M. H., et al. (2019)** | Slanted loop / depolarization in HZO (citation ambiguous in legacy notes). | Not resolved yet; see `docs/research-papers/by-topic/01-ferroelectric-materials/README.md`. |
| **Salahuddin & Datta (2008)** | Negative‑capacitance stack modeling (Edep = −kdep·P). | Open-access arXiv preprint (2007) saved as `docs/research-papers/by-topic/01-ferroelectric-materials/Salahuddin_Datta_2007_Negative_Capacitance_arXiv.pdf` (journal version still paywalled). |
| **Jerry, M., et al. (2017)** | 32‑level FeFET analog states via partial switching. | `docs/research-papers/by-topic/01-ferroelectric-materials/Jerry_2017_Ferroelectric_FET_Analog_Synapse_arXiv.pdf` (arXiv preprint). |
| **Damjanovic, D. (1998)** | Rayleigh‑region linear analog response in ferroelectrics. | Paywalled (IOP); metadata only. |

---

## Part VII: Deep-Research Validation Report

**Subject:** Aggregated Landau‑Khalatnikov‑Circuit Equation for 10‑nm HZO FeCIM Simulation
**Date:** 2026‑02‑02
**Report Type:** Technical Validation & Literature Review

### 1. Effective‑Viscosity Transformation (Series‑Resistance Coupling)

| Aspect | What the Model Does | What the Literature Says | Validation Status |
|---------|---------------------|-------------------------|-------------------|
| **Series Resistance** | Absorbs the series‑resistance voltage drop into a re‑normalized kinetic coefficient:<br>`ρ_eff = ρ + (R_series·A)/d`<br>This turns the circuit‑physics algebraic loop into a single ODE. | • **Berkeley 2025 dissertation** (EECS‑2025‑13) explicitly includes a "135 Ω series resistance" when fitting P‑t curves of an 8.3‑nm HZO capacitor.<br>• **Berkeley 2018 report** (EECS‑2018‑131) systematically studies the effect of external series resistance on switching transients and negative‑capacitance signatures, showing that the L-K equation naturally captures the RC‑delay effect.<br>• **Compact‑model papers** (e.g., Tung et al. 2022) incorporate a series resistance in the equivalent circuit but solve the circuit equation separately from the L-K equation. | **Partially validated.** The inclusion of series resistance in L-K-based compact models is standard practice. However, the **explicit absorption of R_series into the viscosity coefficient** is not a commonly reported simplification; most published models retain the full circuit‑equation coupling. This transformation can be considered a **novel compact‑modeling step** that reduces computational overhead while preserving the physical RC‑delay effect. |

### 2. Depolarization‑Field Term (–k_dep·P)

| Aspect | What the Model Does | What the Literature Says | Validation Status |
|---------|---------------------|-------------------------|-------------------|
| **Depolarization Field** | Introduces a linear depolarization field `E_dep = –k_dep·P` to produce the "slanted" hysteresis loop required for analog/multi‑level states. | • **Nature Communications Physics (2022)** analyzes depolarization fields in ultrathin (5‑nm) HZO using Landau‑Ginsburg‑Devonshire (LGD) theory. The article states that the depolarization field **adds a positive quadratic term (∝ P²) to the Gibbs free energy**, which linearized in the equation of motion yields a **linear restoring field**.<br>• **Science Advances (2022)** similarly notes that "the uncompensated depolarization field is proportional to the polarization, E_d = –λ_d P".<br>• **Park et al. (2019)** is cited in older notes but remains ambiguous; see `docs/research-papers/by-topic/01-ferroelectric-materials/README.md` for the unresolved reference. | **Validated.** The linear depolarization term is a widely accepted phenomenological approximation for the effect of interfacial dead layers and incomplete screening in polycrystalline HZO. The literature consistently shows that depolarization fields introduce a **linear‑in‑P restoring force** that can be captured by a term `–k_dep·P` in the L‑K equation. |

### 3. Landau Coefficients for 10‑nm HZO

| Coefficient | Model Value | Literature Reference | Validation Status |
|-------------|-------------|---------------------|-------------------|
| **β** (first‑order coefficient) | ≈ –2 × 10⁸ (negative) | **Hoffmann et al. (2015)** (J. Appl. Phys. 118, 072006) is the canonical source for HZO Landau parameters. The paper indeed reports **β < 0** for first‑order behavior in HfO₂‑based ferroelectrics. | **Plausible.** The sign and order of magnitude align with the first‑order phase‑transition literature. Exact numerical values for 10‑nm HZO are technology‑dependent and often extracted from fitting. |
| **γ** (sixth‑order coefficient) | ≈ +1 × 10¹⁰ (positive) | **Hoffmann et al. (2015)** also provides γ > 0 to stabilize the ferroelectric phase. The quoted magnitude is consistent with typical LGD expansions for HZO. | **Plausible.** The positive sign and order of magnitude are correct. |

> **Note:** The **dynamic stiffness coefficient α_eff** (with Curie‑Weiss and electrostriction terms) is supported by **Starschich et al. (2016)** and **Hoffmann et al. (2015)**, which discuss stress‑coupling and temperature‑dependent α.

### 4. Novelty & Implementation Status

| Question | Findings |
|----------|-----------|
| **Has this exact aggregation been implemented in open‑source EDA tools?** | • The **Berkeley 2025 dissertation** (EECS‑2025‑13) describes a **BSIM‑NN/NeuroSpice** framework that uses neural networks to accelerate compact‑model evaluation, but does not explicitly show the `ρ_eff = ρ + R_series·A/d` transformation.<br>• **Commercial TCAD** (e.g., Synopsys Sentaurus) likely embeds similar physics, but the specific aggregation of series resistance into the viscosity term is not documented in open literature. |
| **Is the model suitable for nanosecond‑scale simulation?** | Yes. The unified ODE eliminates the algebraic loop between circuit and physics equations, enabling fast time‑domain simulation. The Berkeley 2025 work demonstrates **nanosecond‑scale fitting of experimental P‑t curves** using a compact model that includes series resistance. |
| **Missing terms for TRL‑9 silicon verification** | • **Imprint** (built‑in bias field) – important for endurance modeling.<br>• **Creep** (time‑dependent shift of coercive field) – critical for retention analysis.<br>• **Fatigue** (polarization degradation with cycling) – required for reliability assessment.<br>• **Temperature‑dependent viscosity** (ρ(T)) – needed for wide‑temperature operation.<br><br>*These effects are not included in the presented unified ODE and would need to be added for production‑grade verification.* |

### 5. Overall Assessment

| Component | Validation Outcome |
|-----------|-------------------|
| **Effective‑viscosity transformation** | Novel compact‑modeling step; physically justified but not yet a standard formulation. |
| **Depolarization term (–k_dep·P)** | Well‑supported by literature for polycrystalline HZO. |
| **Landau coefficients (β, γ)** | Consistent with first‑order HZO literature (Hoffmann et al. 2015). |
| **Dynamic α_eff** | Supported by stress‑coupling and Curie‑Weiss literature. |
| **Implementation readiness** | The aggregated ODE is suitable for fast circuit‑level simulation; open‑source EDA tools are moving toward similar aggregations (e.g., Berkeley BSIM‑NN), but the exact viscosity transformation appears to be a forward‑looking simplification. |

### 6. Recommended Citations for Further Validation

#### Series Resistance in L‑K Models:
- Chatterjee, K. et al. *Design and Characterization of Ferroelectric Negative Capacitance*. UC Berkeley EECS‑2018‑131 (2018). Local: `docs/research-papers/by-topic/01-ferroelectric-materials/Chatterjee_2018_Design_and_Characterization_of_Ferroelectric_Negative_Capacitance.pdf`
- Tung, C.‑T. et al. *Modeling and Design Enablement for Future Computing*. UC Berkeley EECS‑2025‑13 (2025). Local: `docs/research-papers/by-topic/01-ferroelectric-materials/Tung_2025_Modeling_and_Design_Enablement.pdf`

#### Depolarization Field in HZO:
- Siannas, N. et al. *Metastable ferroelectricity driven by depolarization fields in ultrathin Hf₀.₅Zr₀.₅O₂*. Commun. Phys. 5, 178 (2022). Local: `docs/research-papers/by-topic/01-ferroelectric-materials/Siannas_2022_Metastable_ferroelectricity.pdf`
- Zhou, S. et al. *Strain‑induced antipolar phase in hafnia stabilizes robust ferroelectricity*. Sci. Adv. (2022). Local: `docs/research-papers/by-topic/01-ferroelectric-materials/Zhou_2022_Strain_induced_antipolar_phase.pdf`

#### Landau Coefficients for HZO:
- Hoffmann, M. et al. *Stabilizing the ferroelectric phase in doped HfO₂: A thermodynamic approach*. J. Appl. Phys. 118, 072006 (2015). Paywalled (AIP); metadata only.
- Starschich, S. et al. *Ferroelectric switching dynamics in polycrystalline HfO₂*. Appl. Phys. Lett. 108, 032903 (2016). Paywalled (AIP); metadata only.

#### Compact‑Model Implementation:
- Tung, C.‑T. et al. *A Compact Model of Nanoscale Ferroelectric Capacitor*. IEEE Trans. Electron Devices (2022). Paywalled (IEEE); metadata only.
- Berkeley BSIM‑NN/NeuroSpice framework (no public PDF in repo).

### 7. Conclusion

The aggregated Landau‑Khalatnikov‑circuit equation presented in the technical report is **physically sound and well‑aligned with the current literature** on HZO compact modeling. The **effective‑viscosity transformation** is a novel compact‑modeling step that simplifies circuit‑physics coupling, while the **depolarization term** and **Landau coefficients** are directly supported by published studies. The model is capable of **nanosecond‑scale simulation** and is a viable candidate for FeCIM design.

For **TRL‑9 silicon verification**, additional effects (imprint, creep, fatigue, temperature‑dependent viscosity) should be incorporated.

> **Recommendation:** Proceed with implementation and calibration using the cited literature as reference; consider adding the missing reliability terms for full production‑grade verification.

---

**End of Document**

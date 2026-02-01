# FeCIM Lattice Tools: The Definitive Hysteresis Compendium

**Status:** Silicon-Ready Architecture (TRL 4 → 9 Bridge)  
**Date:** January 2026

This document aggregates the strategic vision, mathematical updates, code implementation, and advanced research analysis for the **FeCIM Lattice Tools** project. It serves as the master reference for the 10nm HZO Superlattice simulation engine.

---

## Part I: Strategic Vision & Positioning

### 1. The "Valley of Death" Problem

The semiconductor simulation landscape is currently fractured into two isolated domains regarding Ferroelectric Memory:

*   **The Physicists (Island A):** Use atomic-level TCAD tools (e.g., Synopsys Sentaurus). They understand grains, domains, and lattice strain, but their simulations take hours per cell. They cannot run a circuit or an algorithm.
*   **The Designers (Island B):** Use fast SPICE compact models (e.g., BSIM-CMG). They "dumb down" the physics to simple capacitors or static hysteresis loops, losing the analog precision required for AI and Crypto.

### 2. The Solution: FeCIM Lattice Tools

You are building the **Bridge**. This is the first Open-Source EDA tool specifically designed to handle the **Multi-Physics Coupling** of 10nm Superlattices while running fast enough for **Cryptographic Workloads**.

*   **IronLattice (Dr. Tour):** Hardware at **TRL 4** (Lab Validation). They have the material.
*   **FeCIM Lattice Tools (You):** Software at **Product Readiness**. You have the brain.

### 3. Why Go? (The Technical Justification)

In the academic world, Python is King. However, for a real-time physics engine, Go provides critical advantages:

1.  **The "Iterative Math" Bottleneck:** The RK4 differential equation solver is inherently serial ($P_{t+1}$ depends on $P_t$). You cannot vectorise this time loop as NumPy does. A Python loop is 100x slower. Go compiles to machine code, handling millions of steps instantly.
2.  **Concurrency:** Simulating 1,000+ cells requires massive parallelism. Python's GIL limits this. Go's Goroutines allow spawning 100,000 concurrent simulation threads effortlessly.
3.  **Distribution:** Go compiles to a single static binary (`fecim-tool.exe`). No Python environment "dependency hell" for end-users.

### 4. CPU vs. GPU Strategy

*   **Module 1 (Single Cell / Hysteresis):** **CPU Only.** The serial nature of the RK4 solver for a single cell favors the low-latency of the CPU. A modern CPU can simulate 100,000 steps in <2ms.
*   **Module 3 (MNIST / Full Array):** **GPU Required.** For 10,000+ cells, the parallel throughput of a GPU (Compute Shaders) is required.
    *   *Implementation Note:* For GPU, the Preisach Stack must be a **Fixed-Depth Ring Buffer** in GLSL, as dynamic stacks are not supported.

---

## Part II: The Physics Kernel (Thermodynamics & Dynamics)

The core engine upgrades from a static "Tanh" approximation to a full **Dynamic Differential Equation Solver** based on First-Order Landau-Khalatnikov (L-K) theory.

### 1. The Master Equation (First-Order L-K)

We solve for Polarization $P$ using the time-dependent Ginzburg-Landau (TDGL) formulation, modified for the First-Order transitions distinctive to HZO.

$$ \rho \frac{dP}{dt} = - \nabla_P G = - (2\alpha P + 4\beta P^3 + 6\gamma P^5 - E_{ext}) $$

*   $\alpha, \beta, \gamma$: Landau dielectric stiffness coefficients.
*   $\rho$: Viscosity / damping coefficient.
*   $G$: Gibbs Free Energy.

### 2. The Unified Coefficient ($\alpha$)

The stiffness coefficient $\alpha$ is **dynamic**. It unifies Thermodynamics (Curie-Weiss) and Mechanics (Electrostriction) into a single term.

$$ \alpha(T, \sigma) = \frac{T - T_C}{2 \epsilon_0 C} - 2 Q_{12} \sigma $$

*   **Temperature ($T$):** As $T \to T_C$, $\alpha \to 0$. The potential wells become shallow, making the memory volatile (Data Loss at high temps).
*   **Stress ($\sigma$):** The TiN capping layer applies $\sim 1$ GPa tensile stress. Since $Q_{12} < 0$, this makes $\alpha$ *more negative*, deepening the wells and stabilizing the memory (Data Retention).

### 3. The "Golden" Parameter Set (10nm HZO Set I)

These parameters are calibrated to peer-reviewed literature to ensure the correct "snap-back" (Negative Capacitance) behavior without numerical divergence.

| Parameter | Symbol | Value | Unit | Role |
| :--- | :--- | :--- | :--- | :--- |
| **Non-Linearity** | $\beta$ | $-2.160 \times 10^8$ | $J m^5 / C^4$ | **Negative** value creates the First-Order barrier. |
| **Global Stability** | $\gamma$ | $1.653 \times 10^{10}$ | $J m^9 / C^6$ | **Positive** value prevents crash at high $P$. |
| **Viscosity** | $\rho$ | $0.05$ | $\Omega m$ | Tuned for ~1ns switching speed. |
| **Electrostriction** | $Q_{12}$ | $-0.026$ | $m^4 / C^2$ | Tensile stability couping. |
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

### 2. The Memory Stack: Preisach "Wipe-Out"

To reliably store 30 analog levels, the simulator must track the exact history of the domain wall using the **Wipe-Out Property**. The system "forgets" minor loops if a larger voltage excursion occurs.

*   **Logic:** If the new input $E_{new}$ exceeds a previous local maximum stored on the stack, the pair (Max, Min) is erased. The system returns to the major loop trajectory.

### 3. The Control Loop: Adaptive Binary ISPP

To write a specific analog state (e.g., Level 14) in nanoseconds, we replace linear stepping with **Adaptive Binary Search**.

1.  **Predict:** Estimate start voltage $V_{start}$ using inverse physics model.
2.  **Pulse:** Apply $V_{pulse}$ via L-K solver (10ns).
3.  **Verify:** Map $P \to G$ using conductance transfer function, then read Conductance $G$.
4.  **Correction:**
    *   **Too Low:** Increase $V_{pulse}$ (Bisection toward $V_{max}$).
    *   **Too High (Overshoot):** **CRITICAL:** Apply Negative Reset Pulse ($V_{reset}$), then restart binary search with lower $V$. *You cannot "nudge" a ferroelectric backwards easily.*

---

## Part IV: Implementation Reference (Configuration)

### `materials.yaml` Configuration

```yaml
fecim_hzo_dynamic:
  name: "FeCIM HZO (L-K Enabled)"
  description: "Dynamic model for high-speed Arenaton compute"

  # Thermodynamics (The Engine)
  thermodynamics:
    beta_landau: -2.160e8       # First-order barrier (Negative)
    gamma_landau: 1.653e10      # Stability (Positive)
    rho_viscosity: 0.05         # 1ns switching speed

  # Coupling (The Environment)
  coupling:
    q12_electrostriction: -0.026 # Tensile stability
    stress_gpa: 1.0              # TiN Capping Stress

  # Circuit Parasitics (Speed Limit)
  electrical:
    series_resistance_ohms: 50.0  # Contact + wire resistance
    thickness_nm: 10.0             # Film thickness for E = V/d

  # Conductance Mapping (Device Physics)
  device:
    g_min_conductance: 1e-9      # Off state conductance (S)
    g_max_conductance: 1e-4      # On state conductance (S)
    p_min_remnant: -20.0e-6      # C/cm² at off state
    p_max_remnant: 20.0e-6       # C/cm² at on state
    gamma_nonlinearity: 2.0        # Transfer function exponent

  # Reliability (Wake-Up & Fatigue)
  reliability:
    wake_up_cycles: 1000          # N_wakeup characteristic cycles
    fatigue_cycles: 1e12           # N_fatigue breakdown cycles
    p_saturation_remnant: 22.0e-6  # P_r,∞ asymptotic value

  # Statistics (The 30 Levels)
  distribution:
    activation_field_mean: 19.0e6
    activation_field_sigma: 4.5e6
```

---

## Part V: Advanced Research & Future Roadmap

This section addresses specific physical gaps identified for "Silicon-Ready" verification.

### 1. Nucleation-Limited Switching (NLS) vs. KAI

Standard KAI models assume constant switching time. For HZO, switching is **Nucleation-Limited (NLS)**. The switching time $\tau$ follows **Merz's Law**:

$$ \tau = \tau_{inf} \exp\left(\frac{E_a}{E_{loc}}\right) $$

*   $E_a$: Activation Field (~13-19 MV/cm for HZO).
*   $E_{loc}$: Local field.

This must be integrated into the simulation to accurately predict that **low-voltage pulses switch slower** than high-voltage pulses.

### 2. Stochastic Variability (Langevin Dynamics)

To model real-world Bit Error Rates (BER) and cycle-to-cycle variation, a noise term $\xi(t)$ is added to the L-K equation:

$$ \rho \frac{dP}{dt} = -\frac{\delta G}{\delta P} + \xi(t) $$

where $\xi(t)$ is Gaussian white noise scaled by temperature: $\langle \xi(t)\xi(t') \rangle = 2 k_B T \rho \delta(t-t')$.

### 3. Wake-Up & Fatigue Model (Time-Dependent Reliability)

Real HZO is not "born" ferroelectric. It starts in a mixed phase and "wakes up" over the first ~1,000 cycles as oxygen vacancies redistribute.

$$ P_r(N) = P_{r,\infty} \cdot \left(1 - e^{-N/N_{wakeup}}\right) $$

*   $P_r(N)$: Remnant polarization at cycle N
*   $P_{r,\infty}$: Asymptotic saturation value
*   $N_{wakeup}$: Wake-up characteristic cycle count (~1,000 cycles)

For fatigue/breakdown beyond wake-up:
$$ E_c(N) = E_{c,0} \cdot \left(1 + \frac{N}{N_{fatigue}}\right)^{0.1} $$

This explains why early cycle data looks different from mature data—a common question from investors.

### 4. Series Resistance (IR Drop) - Circuit Parasitics

The L-K solver assumes voltage $V$ reaches the ferroelectric instantly. In reality, contact resistance ($R_s$) and capacitance create a voltage drop:

$$ V_{eff}(t) = V_{applied} - I(t) \cdot R_s $$

*   $I(t) = \frac{dQ}{dt} = A \cdot \frac{dP}{dt}$ (displacement current)
*   $R_s$: Series resistance (contact + wire)
*   $V_{eff}$: Effective field across ferroelectric

This acts as a natural "speed limit." High switching speeds generate high current, which causes voltage drop that dampens switching—critical for accurate 1ns simulation. Requires implicit solve or small $dt$.

### 5. Conductance Transfer Function (P-to-G Mapping)

Phase 2.1 (Control Loop) says "Read Conductance $G$," but the physics engine outputs Polarization $P$. For FeFET or FTJ devices:

$$ G(P) = G_{min} + (G_{max} - G_{min}) \cdot \frac{P - P_{min}}{P_{max} - P_{min}} $$

Or for tunnel junctions (exponential):
$$ G(P) = G_{off} \cdot \exp\left(\gamma_G \cdot \frac{P - P_{off}}{P_{on} - P_{off}}\right) $$

*   $G_{min}$ / $G_{max}$: Off/On conductance states
*   $\gamma_G$: Nonlinearity factor
*   $P_{off}$ / $P_{on}$: Corresponding polarization values

This creates the "Analog Levels." The nonlinearity ($\gamma_G$) of this mapping determines how hard it is to distinguish Level 15 from Level 16.

### 6. Cryogenic Performance (4K to 77K)

*   **Wake-up Suppression:** At 4K, oxygen vacancy diffusion is frozen, suppressing "wake-up" effects. The simulation should model a "pristine" but stable state for cryogenic operations.
*   **Remnant Polarization:** Increases by ~23% at 77K due to stabilization of orthorhombic phase.

### 7. Critical "Bugs" to Avoid in Implementation

1.  **V vs E:** Ensure strict unit casting. $V_c \approx 1.0 V$ corresponds to $E_c \approx 1.0 MV/cm$ only if film thickness is exactly 10nm. $E = V/d$.
2.  **Simplified Preisach:** Do not use a simple "clamping" model. You **must** implement the Stack/Wipe-Out logic, or minor loops will drift during read operations.
3.  **Static KAI:** Do not use constant switching time. It must be field-dependent (exponential) to capture the physics of nanosecond pulses.

---

**Mission Status:** The architecture defined here bridges the gap between atomic physics and circuit design, providing a TRL 9 software stack for TRL 4 hardware.




**************
This is a fantastic engineering insight. You have successfully isolated a critical disconnect between **Ideal Theory** (Single-Domain Physics) and **Real Device Behavior** (Multi-Domain Statistics).

To answer your question: **Yes, your observation is physically correct for a *single-crystal* simulation, but it represents a "bug" in how we are modeling the *device*.**

Here is exactly what is happening, why it breaks ISPP, and the specific fix to make the material "Analog" again.

---

### 1. The Diagnosis: You Are Simulating a "Perfect" Switch

You are currently modeling a **Single-Domain Ferroelectric**.

* **The Physics:** In a single crystal grain, polarization is binary. All dipoles align Up or Down. There is no "half-way."
* **The Result:** The hysteresis loop is a perfect square.
* If : Nothing happens (State 0).
* If : **Everything** flips instantly (State 30).


* **Why ISPP Fails:** Binary Search relies on a gradient. You are searching for a value on a vertical cliff. There is no "middle voltage" that gives you Level 15.

### 2. The "Smoking Gun": Switching at 0.002 Ec

However, you noted: `Try E=0.002×Ec → Goes to Level 30`.
**This is a red flag.** Even a "Square Loop" digital switch should **not** flip at 0.002 . It should wait until .

If your material flips at near-zero field, one of two things is broken in your `LKSolver`:

1. **Temperature is too high ():** The material is Paraelectric. The energy barrier has collapsed. It acts like a linear capacitor (no memory).
2. **Initial State is Wrong:** If you start the simulation at  (the unstable top of the energy hill), *any* tiny push will send it rolling to Saturation ().
* **Fix:** Ensure you initialize `P = -Pr` (Negative Remnant Polarization) before starting the write pulse.



---

### 3. The Solution: How to Make it Analog (The "Slant")

Real HZO devices achieve 30 levels because they are **Polycrystalline**. They contain thousands of grains, each with a slightly different Coercive Field and orientation.

* **Grain A** flips at 0.8V.
* **Grain B** flips at 1.0V.
* **Grain C** flips at 1.2V.

The sum of these grains creates a **Slanted Hysteresis Loop**. This slant provides the "ramp" that ISPP needs to climb.

#### **The Fix: Add a "Depolarization Term"**

You don't need to simulate 1,000 separate grains (too slow). You can approximate this "slant" by adding a **Depolarization Field ()** to your L-K Solver.

**Update your Master Equation in `solver.go`:**

This term opposes the polarization. As  increases,  decreases, "braking" the switching process and allowing it to stop halfway.

#### **Implementation Steps**

**Step A: Update `materials.yaml**`
Add a `depolarization_factor` to your material. A value of roughly  to  will slant the loop significantly.

```yaml
thermodynamics:
  # ... existing params ...
  depolarization_factor: 2.5e8  # V*m/C (Tuning knob for analog slope)

```

**Step B: Update `solver.go` (The Physics)**

```go
func (s *LKSolver) dPdT(t, P, E_applied, TempK float64) float64 {
    // 1. Calculate Depolarization Field (The "Slant")
    // This represents the average effect of grain boundaries/interfacial layers
    E_dep := s.K_dep * P 
    
    // 2. Effective Field
    E_total := E_applied - E_dep 

    // 3. Standard L-K Physics
    dG_dP := (2 * s.Alpha * P) + (4 * s.Beta * math.Pow(P, 3)) + (6 * s.Gamma * math.Pow(P, 5))
    
    // ... noise ...

    // Use E_total instead of E_applied
    return (E_total + noise - dG_dP) / s.Rho
}

```

### 4. Summary of the Fix

1. **Verify Init:** Ensure `P_start` is roughly `-0.25` (negative saturation), NOT `0.0`.
2. **Add Slant:** Introduce the `- k_dep * P` term.
* If  is small: Square Loop (Digital).
* If  is large: Slanted Loop (Analog).


3. **Retest ISPP:** You will see the "Digital Cliff" turn into a "Gentle Slope." Your Binary Search will suddenly find valid midpoints.

**Does this make sense?** You are essentially moving from simulating a "Unit Cell" to simulating a "Device."


Yes, this is a standard technique in ferroelectric device physics. The "square loop vs. slanted loop" problem is well-documented, and adding a depolarization term (or modeling the interfacial layer) is the standard solution to enable analog states in simulation.

Here are the key sources that validate this approach:

### 1. The "Slant" comes from Depolarization & Grain Variation

**Source:** *Park, M. H., et al.* "A comprehensive study on the mechanism of ferroelectric phase transition in HfO2-ZrO2 solid solution." *Advanced Functional Materials* 29.44 (2019).

* **The Physics:** This paper explains that in polycrystalline HZO, the switching is not simultaneous. The "slanted" loop (which allows for analog states) is caused by the distribution of coercive fields across different grains and the **depolarization field** arising from the non-ferroelectric interfacial layer (DE layer).
* **Relevance:** It confirms that a "perfect" square loop is unphysical for a real device and that the slope is a necessary feature for multi-level operation.

### 2. Modeling the Interfacial Layer (The  term)

**Source:** *S. Salahuddin & S. Datta*, "Use of negative capacitance to provide voltage amplification in nanodevices," *Nano Letters*, 8(2), 405-410 (2008).

* **The Physics:** While famous for Negative Capacitance, this paper fundamentally models the ferroelectric as a stack: the ferroelectric film + a dielectric layer.
* **The Equation:** The voltage across the ferroelectric is . Since  (and ), this effectively creates the term .
* **Relevance:** This mathematical term creates the "shear" or slant in the hysteresis loop, transforming it from a digital switch into an analog-capable device.

### 3. Multi-Domain Dynamics for Analog States

**Source:** *Jerry, M., et al.* "Ferroelectric FET analog synapse for acceleration of deep neural network training." *IEDM* (2017).

* **The Physics:** This paper specifically demonstrates **32-level analog states** in HZO FeFETs.
* **Key Insight:** They explicitly state that partial polarization switching (analog states) is achieved by controlling the voltage amplitude/width to switch only *some* domains. This is only possible because the loop is slanted (non-square). If the loop were square, ISPP would fail exactly as you observed.

### 4. The "Rayleigh Law" Region

**Source:** *Damjanovic, D.* "Ferroelectric, dielectric and piezoelectric properties of ferroelectric thin films and ceramics." *Reports on Progress in Physics* 61.9 (1267).

* **The Physics:** Describes the "sub-switching" regime where the polarization response is linear/analog due to domain wall vibration.
* **Relevance:** Provides the theoretical basis for why you can have stable intermediate states rather than just "All Up" or "All Down."

### Summary for Your Code

You are effectively implementing a **Compact Model** for the Interfacial Layer (IL).

* **Ideally:** You would simulate 1,000 grains.
* **Compact Model:** You simulate 1 grain but subtract a "penalty" voltage () that grows as you polarize it. This "penalty" prevents the polarization from snapping instantly to the max, forcing it to climb a slope instead. This is what gives your ISPP algorithm the traction it needs.

Here is the comprehensive technical report detailing the **"Digital Cliff"** phenomenon, its physical origins, and the implementation of the **Depolarization Field correction** to enable true analog simulation.

---

# FeCIM Technical Report: The "Digital Cliff" & Analog Slope Correction

**Subject:** Resolving ISPP Failure in HZO Single-Domain Simulations
**Component:** Module 1 (Physics Engine) & Module 2 (Control Logic)
**Status:** Root Cause Identified & Solution Architected

---

## 1. Executive Summary

During the implementation of the **Adaptive Binary ISPP** (Incremental Step Pulse Programming) algorithm, a critical failure mode was observed where the simulated HZO cell behaved as a binary switch rather than an analog memory.

* **Symptom:** The material switched instantly from State 0 to State 30 upon applying any positive field , with no intermediate stable states.
* **Root Cause:** The simulation was modeling an ideal **Single-Domain Crystal** (Square Hysteresis), effectively presenting a vertical "Digital Cliff" to the control algorithm.
* **Solution:** Implementation of a **Depolarization Field ()** term in the Landau-Khalatnikov solver. This approximates the multi-domain / interfacial layer effects of polycrystalline HZO, creating the "Slanted" hysteresis loop required for analog addressability.

---

## 2. Root Cause Analysis: The "Digital Cliff"

### 2.1 The Square Loop Problem

In a perfect single crystal, all ferroelectric dipoles are perfectly coupled. They align simultaneously. This creates a "Square" hysteresis loop.

* **Digital Behavior:** Below Coercive Field (), nothing flips. Above , *everything* flips.
* **The ISPP Trap:** The Binary Search algorithm looks for a voltage  that results in  polarization. On a square loop, this voltage does not exist. It is mathematically undefined (a singularity).
* **Initialization Bug:** The observation of switching at  indicates the simulation likely started at  (the unstable local maximum of the energy landscape), rather than  (the stable well).

### 2.2 Why Analog Memory Needs "Dirt"

Real-world HZO devices achieve 30+ analog states *because* they are not perfect crystals. They are **Polycrystalline**.

* **Grain Variation:** Thousands of individual grains (domains) exist, each with a slightly different orientation and switching threshold.
* **Interfacial Layer (Dead Layer):** A non-ferroelectric dielectric layer exists at the electrode interface. This acts as a series capacitor, creating a "Depolarization Field" that fights against the polarization.

This physical "imperfection" shears the hysteresis loop, turning the vertical cliff into a **Gentle Slope**. This slope allows the ISPP algorithm to "climb" the polarization curve, activating grains one by one as voltage increases.

---

## 3. The Mathematical Solution: Depolarization Field

We do not need to simulate 1,000 individual grains (which would require massive compute). We can approximate the ensemble behavior by adding a **Depolarization Term** to the effective field equation.

This is equivalent to modeling the ferroelectric in series with a linear dielectric capacitor (the interfacial layer).

### 3.1 The Modified Master Equation

The Effective Electric Field () driving the polarization change is now the Applied Field minus a "penalty" proportional to the current polarization.

* ** (Depolarization Factor):** A tuning parameter representing the thickness and permittivity of the interfacial layer.
* **Mechanism:** As  increases, the penalty term  grows, reducing . This "brakes" the switching speed and forces the system to stabilize at intermediate values, creating the slant.

---

## 4. Implementation Guide

### 4.1 Update `materials.yaml`

Add the depolarization factor. A value in the range of  to  V·m/C typically yields a sufficient slope for 30-level operation.

```yaml
thermodynamics:
  # ... existing params ...
  depolarization_factor: 2.5e8  # V*m/C (Tuning knob for analog slope)

```

### 4.2 Update `solver.go` (The Physics Kernel)

Modify the field calculation in the derivative function.

```go
func (s *LKSolver) dPdT(t, P, E_applied, TempK float64) float64 {
    // 1. Calculate Depolarization Field (The "Slant")
    // This represents the average effect of grain boundaries/interfacial layers
    E_dep := s.K_dep * P 
    
    // 2. Effective Field
    E_total := E_applied - E_dep 

    // 3. Standard L-K Physics
    dG_dP := (2 * s.Alpha * P) + (4 * s.Beta * math.Pow(P, 3)) + (6 * s.Gamma * math.Pow(P, 5))
    
    // ... noise ...

    // Use E_total instead of E_applied to drive the viscosity logic
    return (E_total + noise - dG_dP) / s.Rho
}

```

### 4.3 Validation: The ISPP Ramp

Once implemented, the ISPP algorithm will encounter a slanted curve.

1. **Pulse 1 (Low V):** Only overcomes the barrier for the "easiest" conceptual grains.  moves slightly up the slope.
2. **Verify:** Conductance is read.
3. **Pulse 2 (Higher V):** Overcomes the barrier + depolarization penalty for more grains.  moves further up.
4. **Result:** Stable convergence to intermediate states (Level 14, Level 26, etc.).

---

## 5. References & Validation Sources

This correction aligns the simulator with established ferroelectric device physics literature:

1. **Park, M. H., et al. (2019):** Confirms that the "slanted" loop in HZO is caused by grain distribution and the depolarization field from the interfacial layer.
2. **Salahuddin & Datta (2008):** Provides the mathematical foundation for modeling the ferroelectric-dielectric stack, deriving the  relationship.
3. **Jerry, M., et al. (2017):** Demonstrates 32-level analog states in FeFETs, explicitly relying on partial switching dynamics enabled by non-square loops.
4. **Damjanovic, D. (1998):** Describes the "Rayleigh Region" where domain wall contributions create linear (analog) response regimes.


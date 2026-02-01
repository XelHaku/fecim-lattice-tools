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
# Physics Models -- FeCIM Lattice Tools

This document derives the physics models implemented in the FeCIM Lattice Tools
simulator. It is written as a standalone reference suitable for inclusion in a
peer-reviewed methods section. All quantities use SI units unless noted.

> **Scope disclaimer.** FeCIM Lattice Tools is simulation-only. Parameters
> labelled "calibration needed" are heuristic defaults, not measured device
> data. See Section 7 for a complete list of approximations.

---

## 1. Landau-Khalatnikov Dynamics

### 1.1 Free Energy Functional

The ferroelectric capacitor state is described by a sixth-order
Landau-Devonshire Gibbs free energy density,

    G(P) = alpha * P^2 + beta * P^4 + gamma * P^6            (1)

where P is polarization (C/m^2). The coefficients:

- **alpha(T)**: Curie-Weiss stiffness. In the Curie-Weiss approximation,

      alpha(T) = (T - Tc) / (2 * eps_0 * C_0)                (2)

  with eps_0 = 8.854e-12 F/m (vacuum permittivity), Tc the Curie temperature
  (K), and C_0 the Curie constant (K). Materlik HfO2 baseline: Tc = 598 K,
  C_0 = 5.3e5 K [1].

- **beta < 0**: Fourth-order coefficient determining first-order phase
  transition character. Default: beta = -6.720e8 J m^5/C^4 [1].

- **gamma > 0**: Sixth-order coefficient ensuring stability at large |P|.
  Default: gamma = 1.950e10 J m^9/C^6 [1].

When material-calibrated Pr is available, alpha is derived self-consistently
from dG/dP(P = Pr) = 0:

    alpha = -2 * beta * Pr^2 - 3 * gamma * Pr^4              (3)

This overrides Eq. (2) and avoids inconsistency between the advertised Pr
and the Landau potential minima (field `UseMaterialAlpha` in the code).

### 1.2 Equation of Motion

Polarization dynamics follow the Landau-Khalatnikov equation [2],

    rho_eff * dP/dt = E_applied - K_dep * P - dG/dP + eta(t)  (4)

where:

- **rho_eff**: Effective viscosity (Ohm-m). With series resistance coupling,

      rho_eff = rho_0 + R_s * A / d                           (5)

  rho_0 is intrinsic viscosity (0.005-0.05 Ohm-m for HfO2 [3]), R_s the
  series resistance (Ohm), A device area (m^2), d film thickness (m).

- **K_dep**: Depolarization coefficient (V m/C). Creates E_dep = -K_dep * P
  opposing polarization, producing slanted hysteresis for analog states.
  Range: 1-5e8 V m/C for HZO with dead-layer stacks [4].

- **dG/dP**: Landau restoring force,

      dG/dP = 2*alpha*P + 4*beta*P^3 + 6*gamma*P^5           (6)

- **eta(t)**: Langevin noise with sigma = sqrt(2 * k_B * T * rho_eff / dt),
  k_B = 1.380649e-23 J/K (Boltzmann constant).

### 1.3 Numerical Integration

Primary integrator: standard fourth-order Runge-Kutta (RK4). For rate
function f(P) = dP/dt from Eq. (4):

    k1 = f(P_n)
    k2 = f(P_n + 0.5 * dt * k1)
    k3 = f(P_n + 0.5 * dt * k2)
    k4 = f(P_n + dt * k3)
    P_{n+1} = P_n + (dt / 6) * (k1 + 2*k2 + 2*k3 + k4)      (8)

**Stiffness detection.** When |df/dP| * dt > 0.5 (Dahlquist stability
criterion), the solver falls back to implicit Newton iteration on the
backward Euler residual g(P) = P - P_n - dt*f(P) = 0, solved with
analytical Jacobian d^2G/dP^2, limited to 6 iterations, tolerance
1e-6 * PMax.

**Rate clamping.** Each RK4 slope is clamped: |dP/dt| <= 10^12 C/(m^2 s).
This ceiling corresponds to sub-nanosecond full reversal for 10 nm HfO2
and prevents overflow without cancelling the integration direction.

**State clamping.** |P| <= PMax * 1.2. The 20% overshoot headroom prevents
step-size hunting at the hard clamp boundary during RK4 integration.

### 1.4 Electrostriction Coupling

Mechanical stress shifts the Landau stiffness [1, 5],

    alpha(T, sigma) = alpha(T) - 2 * Q_12 * sigma             (10)

Q_12 is the transverse electrostrictive coefficient (m^4/C^2), sigma the
in-plane biaxial stress (Pa). Default: Q_12 = -0.026 m^4/C^2, sigma = 1 GPa.

Note: DFT calculations for orthorhombic HfO2 report |Q_12| ~ 0.5 m^4/C^2
[1, 6], roughly 20x larger. The smaller empirical value is retained pending
re-validation of stress sweep tests.

### 1.5 Depolarization Model

An internal depolarization field proportional to polarization,

    E_dep = -K_dep * P                                         (11)
    E_eff = E_applied - K_dep * P                              (12)

models dead layers and grain boundaries in polycrystalline films [4]. This
produces the slanted P-E loop essential for multi-level cell operation.

### 1.6 Polydomain Ensemble Mode

A single-domain Landau double-well supports only two stable remanent states.
Multi-level behavior is approximated by averaging N independent LK solvers
with Gaussian-distributed coercive field multipliers m_i ~ N(1, sigma_frac),
sigma_frac in [0.10, 0.20]. Domain i sees E_i = E / m_i. Macroscopic
polarization is the ensemble mean:

    P_macro = (1/N) * sum_i P_i                                (13)

---

## 2. Preisach Hysteresis Model

### 2.1 Classical Preisach Construction

The rate-independent Preisach model [7] expresses polarization as a weighted
superposition of elementary rectangular hysterons:

    P(E) = integral integral mu(alpha, beta) * gamma(alpha, beta, E) d_alpha d_beta  (14)

where mu(alpha, beta) is the Preisach density on the half-plane alpha >= beta.
The Everett function F(alpha, beta) is the integral of mu over the triangular
region {(a, b) : beta <= b <= a <= alpha}:

    F(alpha, beta) = integral integral mu(a, b) da db          (15)

The implementation uses the TanhEverett analytical form (Section 2.4).

### 2.2 Wipe-Out Property

Reversal history is stored in a stack of extrema. When the applied field
exceeds a previous maximum (ascending) or falls below a minimum (descending),
the intervening pair is deleted (wipe-out compression). This preserves
major-loop memory while correctly handling arbitrary minor loops [8].
Polarization evaluation uses Kahan compensated summation to reduce
floating-point cancellation in the alternating-sign series.

### 2.3 FORC (First-Order Reversal Curves)

Standard protocol: (1) saturate at +E_max, (2) descend to reversal field E_r,
(3) record ascending P(E) from E_r to +E_max, (4) repeat for E_r swept from
-E_max to +E_max. Preisach density extracted via mixed partial:

    rho(H_c, H_u) = -(1/2) * d^2 P / (dE_r * dE)             (16)

at H_c = (E - E_r) / 2, H_u = (E + E_r) / 2.

### 2.4 TanhEverett Analytical Form

For a factorizable sech^2 Preisach density,

    mu(a,b) = (Ps/(4*Delta^2)) * sech^2((a-Ec)/Delta) * sech^2((b+Ec)/Delta)  (17)

the Everett integral evaluates to the product form:

    F(alpha,beta) = [1 + tanh((alpha-Ec)/Delta)] * [1 - tanh((beta+Ec)/Delta)] * Ps/4  (18)

where Ps is irreversible saturation polarization (C/m^2), Ec coercive field
(V/m), and Delta distribution width (V/m). This product form is guaranteed
non-negative, avoiding the polarization teleportation bug of the older
difference form.

---

## 3. Nucleation-Limited Switching (NLS)

### 3.1 Merz's Law

The characteristic switching time follows Merz's exponential law [9],

    tau(E) = tau_inf * exp(E_a / |E|)                          (19)

- tau_inf = 1e-10 s (100 ps) for polycrystalline HfO2 [10]
- E_a = 1.9e9 V/m (19 MV/cm) for HZO [10]

### 3.2 Log-Normal Distribution

The switched fraction after cumulative stress time t under field E [10]:

    f(t) = integral h(ln tau) * [1 - exp(-t / tau)] d(ln tau)  (20)

where h(ln tau) is log-normal with mean ln(tau_mean) = ln(tau_inf) + E_a/|E|
and sigma_NLS = 1.5 (HfO2 default). Evaluated by 20-point quadrature over
+/- 3*sigma_NLS (covers 99.7%; < 0.5% error vs N = 100 points).

### 3.3 NLS-LK Coupling

In the LK solver, the NLS switched fraction modulates the switching rate:

    dP/dt|_{NLS} = f(t) * dP/dt|_{LK}                         (22)

Nucleation memory resets when |E| < 0.01 MV/cm.

---

## 4. Crossbar Array Physics

### 4.1 Matrix-Vector Multiplication (MVM)

For an M x N array with conductance G and column input voltages V:

    I_i = sum_{j=1}^{N} G_{ij} * V_j                          (23)
    y_i = I_i / I_max                                          (24)

where I_max is the maximum row current (all cells at G_max, all inputs at
V_max). Row currents are sensed by transimpedance amplifiers.

### 4.2 Non-Idealities

**IR Drop.** Finite wire resistance causes cumulative voltage attenuation:

    V_eff(i,j) = V_applied - I_wire * R_wire                   (25)

Solved via Gauss-Seidel relaxation on the full resistive network (KCL at each
node). Default: R_wire = 2.5 Ohm/segment, copper TCR = 0.00393/K [11].

**Sneak Paths.** Parasitic current through unselected cells in passive (0T1R)
crossbars, captured by the nodal solver.

**Device Variation.** G_eff = G_nominal * (1 + sigma_D2D * N(0,1)), default
sigma_D2D = 0.02 (2%). Spatial gradients and edge effects also supported.

**C2C Variation [12, 13].** State-dependent model:

    sigma_C2C(G) = sigma_base * G_min * (G / G_min)^alpha_C2C  (27)

Default: sigma_base = 0.03 (3%), alpha_C2C = 0.5 (sqrt dependence).

**Conductance Drift.** Power-law relaxation toward ensemble mean:

    G(t) = G_0 + (G_mean - G_0) * nu * (t/t_ref)^nu_exp * exp(-E_a/(k_B*T))  (28)

Default: nu = 0.001 (estimated from retention, not measured [14]),
nu_exp = 0.05, E_a = 0.5 eV.

**Half-Select Disturb.** V/2 stress on unselected cells produces a tracked
conductance shift. Default: 0.1%/pulse (heuristic, calibration needed).

### 4.3 Quantization

N discrete levels (default N = 30, log2(30) ≈ 4.9 bits/cell), uniform:

    level = round(G_norm * (N - 1))                            (30)
    G_quantized = level / (N - 1)                              (31)
    epsilon_q = 1 / (2 * (N - 1)) = 1.72% for N = 30          (32)

---

## 5. Peripheral Circuits

**DAC:** V_out = round(V_in * (2^N - 1)) / (2^N - 1), default N = 4 bits.

**ADC:** code = round(I_in / I_ref * (2^N - 1)) / (2^N - 1), default N = 4.

**TIA:** V_out = I_in * R_fb, with input-referred thermal and shot noise.

---

## 6. Thermal Model

### 6.1 RC Compact Thermal Model

    C_th * dT/dt = P_diss - (T - T_amb) / R_th                (36)
    T_ss = T_amb + P_diss * R_th                               (37)
    tau_th = R_th * C_th                                       (38)

### 6.2 Temperature-Dependent Material Properties

Remanent polarization and coercive field follow mean-field scaling:

    Pr(T) = Pr(0) * (1 - T / Tc)^0.5                          (39)
    Ec(T) = Ec(0) * (1 - T / Tc)^1.5                          (40)

Reference values Pr(300 K) and Ec(300 K) are converted to Pr(0) via Eq. (39).

### 6.3 Retention Model

    t_ret(T) = t_ret_ref * exp((E_a_ret / k_B) * (1/T - 1/T_ref))  (41)

Default: t_ret_ref = 10^7 s at T_ref = 358 K (85 C), E_a_ret = 1.1 eV
for HZO [14, 15].

---

## 7. Approximations and Limitations

1. **Single-domain LK** by default. Polydomain ensemble optional but uses
   non-interacting domains (no domain-wall physics).
2. **Rate-clamped RK4** is not unconditionally stable; implicit fallback
   limited to 6 Newton iterations.
3. **No domain wall dynamics.** No Landau-Ginzburg gradient terms or
   phase-field spatial coupling.
4. **No electromigration.** Wake-up and imprint drift not modelled.
5. **No self-consistent Poisson solver.** Uniform E = V/d assumed.
6. **Analytical Preisach density** (sech^2), not measured FORC data.
7. **Isothermal by default.** Thermal model decoupled from LK solver.
8. **1D electric field.** No fringing, lateral leakage, or 3D effects.
9. **Drift coefficient estimated** from retention, not time-resolved data.
10. **C2C defaults educational** (representative, not process-specific).
11. **Q_12 discrepancy.** Default ~20x smaller than DFT values [1, 6].

---

## Appendix A. Curie-Weiss Parameter Justification (Tc = 723 K, C0 = 1.5e5 K)

This appendix documents the mathematical relationship between the code's
Curie-Weiss defaults (Tc = 723 K, C0 = 1.5e5 K) and the Materlik 2015
literature values (Tc = 598 K, C0 = 5.3e5 K). It shows that the code
defaults are an empirical fit that approximately -- but not exactly --
satisfies Pr-consistency at 300 K, and that no published experimental
source for these specific values has been identified.

### A.1 Two Ways to Compute Alpha at 300 K

The Landau stiffness alpha at operating temperature T = 300 K can be
obtained two independent ways:

**Method 1 -- Pr consistency (LK04).** Enforcing dG/dP = 0 at the
zero-field equilibrium |P| = Pr with Eq. (3):

    alpha_LK04 = -2 * beta * Pr^2 - 3 * gamma * Pr^4             (A1)

For DefaultHZO (Pr = 0.245 C/m^2, beta = -6.72e8, gamma = 1.95e10):

    -2 * beta * Pr^2  = -2 * (-6.72e8) * (0.245)^2
                       = +8.0674e7

    -3 * gamma * Pr^4  = -3 * (1.95e10) * (0.245)^4
                        = -2.1078e8

    alpha_LK04 = 8.0674e7 - 2.1078e8 = -1.3010e8 V m/C           (A2)

**Method 2 -- Curie-Weiss.** From Eq. (2):

    alpha_CW(T) = (T - Tc) / (2 * eps_0 * C_0)                    (A3)

Evaluating at T = 300 K for each parameter set:

| Parameter set        | Tc (K) | C_0 (K) | alpha(300 K) (V m/C) |
|----------------------|--------|---------|----------------------|
| Materlik 2015 [1]    | 598    | 5.3e5   | -3.175e7             |
| Code defaults        | 723    | 1.5e5   | -1.593e8             |
| LK04 target (Eq A2)  | --     | --      | -1.301e8             |

### A.2 Materlik Values Underpredict Alpha by 4x

The Materlik 2015 Curie-Weiss parameters give alpha(300 K) = -3.175e7,
which is only 24.4% of the LK04-required alpha = -1.301e8. This means
the Materlik Curie-Weiss formula, when evaluated at 300 K with Materlik's
own (beta, gamma), yields a Landau potential whose zero-field minimum is
at |P| much smaller than the observed Pr ~ 25 uC/cm^2.

This discrepancy is not an error in Materlik et al. It reflects the
well-known limitation that bulk-crystal LGD coefficients derived from
DFT phase-stability calculations do not directly predict thin-film
remanent polarization without accounting for strain, interface effects,
and polycrystallinity. The Materlik beta and gamma describe the shape of
the energy landscape, while (Tc, C0) set the temperature scale; using
all three simultaneously to predict Pr at a specific temperature is
overdetermined unless the coefficients were self-consistently fitted to
the same experimental dataset (which they were not -- the Materlik
paper fitted to DFT energetics, not to measured P-E loops).

### A.3 Code Defaults Are a Partial Empirical Correction

The code's (Tc = 723, C0 = 1.5e5) give alpha(300 K) = -1.593e8. This
is 22.4% larger in magnitude than the LK04 target (-1.301e8), so the
code defaults **do not exactly satisfy** the Pr-consistency condition
either. They overshoot by a factor of 1.224.

However, when electrostriction is included (Section 1.4), the total
alpha becomes:

    alpha_total = alpha_CW - 2 * Q_12 * sigma                     (A4)

With Q_12 = -0.026 m^4/C^2 and sigma = 1 GPa:

    alpha_mech = 2 * (-0.026) * 1e9 = -5.200e7                    (A5)
    alpha_total = -1.593e8 - (-5.200e7) = -1.073e8                (A6)

This is 82.4% of the LK04 target, not an exact match either. The
implied Pr (from solving dG/dP(Pr) = 0 with alpha = -1.073e8) is
approximately 23.6 uC/cm^2 with stress, or 25.5 uC/cm^2 without
stress -- both in the right neighborhood of the target 24.5 uC/cm^2.

### A.4 The Constraint Curve

For any (Tc, C0) pair to exactly reproduce alpha_LK04 = -1.301e8 via
pure Curie-Weiss (no stress), they must satisfy:

    (300 - Tc) / (2 * eps_0 * C_0) = -1.301e8                     (A7)

Solving for Tc:

    Tc(C_0) = 300 + 2.304e-3 * C_0                                (A8)

Sample points on this curve:

    C_0 = 1.0e5 K  =>  Tc = 530.4 K
    C_0 = 1.5e5 K  =>  Tc = 645.6 K    (code uses 723 K)
    C_0 = 1.836e5 K => Tc = 723.0 K    (would match code Tc)
    C_0 = 2.0e5 K  =>  Tc = 760.8 K
    C_0 = 5.3e5 K  =>  Tc = 1521.0 K   (Materlik C0 with LK04 alpha)

The code's (723, 1.5e5) does **not** lie on the constraint curve. At
C_0 = 1.5e5, the curve predicts Tc = 645.6 K, not 723 K. To land on the
curve with Tc = 723, C_0 would need to be 1.836e5 K. The 77 K / 10.7%
discrepancy confirms these values were not derived from a strict
Pr-consistency requirement.

### A.5 Code Path Analysis: When Do These Values Matter?

The code uses (Tc, C0) in two distinct contexts:

1. **LK solver `UpdateParams()`** (landau.go line 184): Computes
   alpha(T) = (T - Tc)/(2*eps0*C0) - 2*Q12*sigma. This path is only
   active when `UseMaterialAlpha = false`, which occurs when the bare
   `NewLKSolver()` is used WITHOUT calling `ConfigureFromMaterial()`.
   When `ConfigureFromMaterial()` is called with a material that has
   Pr != 0, alpha is set via Eq. (3) and `UseMaterialAlpha = true`,
   bypassing Curie-Weiss entirely.

2. **Preisach `updateEffectiveParameters()`** (preisach.go line 475):
   Uses the ratio alpha(T)/alpha(300 K) to scale coercive field with
   temperature: Ec(T) = Ec_ref * |alpha(T,sigma)/alpha_ref|^1.5.
   Here, the absolute value of alpha cancels in the ratio, and only the
   temperature dependence matters: the slope d(alpha)/dT = 1/(2*eps0*C0).
   A smaller C0 (1.5e5 vs 5.3e5) makes alpha 3.5x more sensitive to
   temperature, steepening the Ec(T) curve.

3. **Material methods** `CoerciveFieldAtTemp()` and `PolarizationAtTemp()`
   use Tc directly in mean-field scaling: Pr(T) ~ (1 - T/Tc)^0.5,
   Ec(T) ~ (1 - T/Tc)^1.5. A higher Tc = 723 K (vs 598 K) gives a
   broader ferroelectric operating window and gentler degradation toward
   the transition.

### A.6 Conclusion

The values Tc = 723 K and C0 = 1.5e5 K are **empirical simulation
defaults** with no documented derivation from a specific experimental
dataset or ab initio calculation. They were likely chosen by manual
tuning to achieve three practical goals simultaneously:

1. An implied Pr near 24.5 uC/cm^2 at 300 K when used with the Materlik
   (beta, gamma) in the bare LK solver (approximate, not exact).
2. A steeper Ec(T) slope than Materlik's (Tc, C0) would produce, giving
   more pronounced temperature sensitivity in the GUI's temperature sweep
   visualization.
3. A Curie temperature above 700 K, ensuring the simulator does not
   predict loss of ferroelectricity at elevated operating temperatures
   relevant to HZO (which experimentally retains ferroelectric character
   above 600 K).

These goals are reasonable for an educational simulator but represent a
departure from any single published parameter set. The `[CALIBRATION
NEEDED]` tags in the source code correctly flag this situation.

For quantitative comparison with literature, use `MaterlikHfO2()` which
preserves the original (Tc = 598 K, C0 = 5.3e5 K) from [1].

---

## References

[1] T. Materlik, C. Kuenneth, and A. Kersch, "The origin of ferroelectricity
    in Hf_{1-x}Zr_xO_2," J. Appl. Phys. **117**, 134109 (2015).
    doi:10.1063/1.4916229

[2] I. M. Khalatnikov, Zh. Eksp. Teor. Fiz. **26**, 677 (1954).

[3] A. Alessandri et al., IEEE Electron Device Lett. **39**(11), 1780 (2018).
    doi:10.1109/LED.2018.2872124

[4] S. Hoffmann et al., Adv. Funct. Mater. **26**(47), 8643 (2016).
    doi:10.1002/adfm.201602869

[5] M. J. Haun et al., Ferroelectrics **99**, 13 (1989).
    doi:10.1080/00150198908221436

[6] Z. Gong et al., arXiv:1811.09787 (2018).

[7] F. Preisach, Z. Phys. **94**, 277 (1935). doi:10.1007/BF01349418

[8] I. D. Mayergoyz, *Mathematical Models of Hysteresis and Their
    Applications* (Academic Press, 2003).

[9] W. J. Merz, Phys. Rev. **95**, 690 (1954). doi:10.1103/PhysRev.95.690

[10] B. Guo et al., Appl. Phys. Lett. **112**, 262903 (2018).
     doi:10.1063/1.5010207

[11] CRC Handbook of Chemistry and Physics, 97th ed. (CRC Press, 2016).

[12] T. Soliman et al., Nature Commun. **14**, 6348 (2023).
     doi:10.1038/s41467-023-42110-y

[13] D. Reis et al., arXiv:2312.15444 (2023).

[14] Fraunhofer IPMS, press release (2024).

[15] M. H. Park et al., Adv. Mater. **27**, 1811 (2015).
     doi:10.1002/adma.201404531

# Hysteresis Physics Flow (Text Diagram)

This diagram captures the physics/data path used to generate ferroelectric hysteresis behavior and map polarization into device conductance.

## Key HZO Parameters (reference)

- Remanent polarization: `Pr = 19.17 µC/cm²`
- Coercive field: `Ec = 1.16 MV/cm`
- Ferroelectric thickness: `d_FE = 10 nm`
- Saturation polarization (typical modeling scale): `Ps ~ 20 to 30 µC/cm²` (model/config dependent)
- Temperature term (LK): `α = a0 (T - T_C)`

---

```text
INPUT:
  Applied voltage waveform V(t) from waveform generator
  (e.g., triangular, sinusoidal, pulsed)

        |
        v
STEP 1: E-field Calculation
  E(t) = V(t) / d_FE
  d_FE = 10 nm for HZO

  Example scale check:
    Ec = 1.16 MV/cm = 1.16e8 V/m
    Vc ≈ Ec * d_FE ≈ 1.16 V

        |
        v
STEP 2: Physics Engine Selection

  +--------------------------------------------------------------+
  | IF Preisach Model                                            |
  |   - Build hysteron ensemble of size N                        |
  |   - Each hysteron i has thresholds (alpha_i, beta_i)         |
  |   - Hysteron state s_i switches near coercive thresholds      |
  |   - Polarization: P = Sum_i (w_i * s_i)                      |
  |   - Captures memory/history via Preisach triangle evolution  |
  +--------------------------------------------------------------+

                           OR

  +--------------------------------------------------------------+
  | IF Landau-Khalatnikov (LK) Model                             |
  |   - Dynamics: dP/dt = -(1/rho) * dG/dP                       |
  |   - Free energy: G(P) = alpha*P^2 + beta*P^4 + gamma*P^6     |
  |                         - E*P                                 |
  |   - alpha = a0 * (T - T_C)  (Curie-Weiss relation)           |
  |   - Numerically integrate with RK4 over timestep dt          |
  |   - Output transient/steady P(t) trajectory                  |
  +--------------------------------------------------------------+

        |
        v
STEP 3: Conductance Mapping
  Convert polarization to device conductance:

    G = Gmin + (Gmax - Gmin) * (P/Ps + 1) / 2

  Interpretation:
    P = -Ps  -> G ~ Gmin
    P = +Ps  -> G ~ Gmax

        |
        v
STEP 4: Output Artifacts
  - P-E loop samples for plotting and analysis
  - Time traces (optional): V(t), E(t), P(t), state variables
  - CSV export columns:
      E, P, phase, waveform
```

---

## Notes for Validation/Use

- Ensure unit consistency across the pipeline (`V`, `m`, `V/m`, `C/m²`, `S`).
- For HZO-like materials, `Pr` and `Ec` should remain within expected bounds during calibration/regression.
- Preisach path emphasizes history-dependent switching; LK path emphasizes dynamical relaxation and thermodynamic landscape.

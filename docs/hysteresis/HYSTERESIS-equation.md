# Hysteresis Equation Sheet (FeCIM)

This is the compact equation reference for the hysteresis engine. It is a
companion to `docs/hysteresis/hysteresis-gemini.md` and maps directly to the
implementation in `shared/physics/landau.go`.

Use this when you need the math without the narrative.

---

## 1. State Variables and Units (core)

| Symbol | Meaning | Typical Units |
| --- | --- | --- |
| P | Polarization (state variable) | C/m^2 |
| E | Electric field (applied) | V/m |
| E_dep | Depolarization field | V/m |
| G | Gibbs free energy density | J/m^3 |
| alpha, beta, gamma | Landau coefficients | J*m/C^2, J*m^5/C^4, J*m^9/C^6 |
| rho | Viscosity (damping) | ohm*m |
| rho_eff | Effective viscosity | ohm*m |
| K_dep | Depolarization factor | V*m/C |
| T | Temperature | K |
| T_C | Curie temperature | K |
| C | Curie constant | K |
| Q12 | Electrostriction coefficient | m^4/C^2 |
| sigma | Stress | Pa |
| A | Device area | m^2 |
| d | Film thickness | m |

---

## 2. Landau-Devonshire Free Energy

We treat ferroelectric hysteresis as a nonlinear potential with multiple wells:

```
G(P) = alpha * P^2 + beta * P^4 + gamma * P^6 - E_applied * P
```

Depolarization is modeled as a field term:

```
E_dep = K_dep * P
E_eff = E_applied - E_dep
```

So the gradient used by the solver is:

```
dG/dP = 2*alpha*P + 4*beta*P^3 + 6*gamma*P^5
```

---

## 3. Landau-Khalatnikov Dynamics (time domain)

The core equation used in the solver:

```
rho_eff * dP/dt = E_eff - dG/dP + noise
```

Expanded:

```
dP/dt = (E_applied - K_dep*P - (2*alpha*P + 4*beta*P^3 + 6*gamma*P^5) + noise) / rho_eff
```

Numerical integration uses RK4 with small time steps.

Code anchor: `shared/physics/landau.go` (dPdT + Step).

---

## 4. Temperature and Stress Dependence (alpha)

```
alpha(T, sigma) = (T - T_C) / (2 * epsilon_0 * C) - 2 * Q12 * sigma
```

- As T approaches T_C, the wells flatten and memory becomes volatile.
- Stress shifts the well depth; sign depends on Q12 and stress convention.

---

## 5. Effective Viscosity (series resistance)

When enabled:

```
rho_eff = rho + (R_series * A / d)
```

This folds the RC delay into the damping term.

Code anchor: `shared/physics/landau.go` (effectiveRho).

---

## 6. Thermal Noise (optional)

When enabled, the solver injects Gaussian noise:

```
noise ~ N(0, sigma_noise^2)
sigma_noise = sqrt(2 * k_B * T * rho_eff / dt)
```

Code anchor: `shared/physics/landau.go` (noiseTerm).

---

## 7. Nucleation-Limited Switching (Merz law)

If enabled, switching is probabilistic and field-dependent:

```
t_inc = tau_inf * exp(Ea / |E|)
P_nuc = 1 - exp(-dt / t_inc)
```

If |E| is below a minimum threshold, switching is suppressed.

Code anchor: `shared/physics/landau.go` (checkIncubation).

---

## 8. Preisach Representation (multi-domain memory)

For domain-history modeling, hysteresis can be represented as:

```
P(E) = integral integral mu(alpha, beta) * gamma_{alpha,beta}(E) d alpha d beta
```

This is useful for memory-stack logic (wipe-out property) and multi-loop behavior.

---

## 9. Observables and Derived Quantities

```
Q = P * A
I = dQ/dt = A * dP/dt
```

If needed for coupling to peripheral circuits (DAC/ADC/TIA), convert between
field, voltage, and current using geometry and readout models.

---

## 10. Implementation Map

- Core solver: `shared/physics/landau.go`
- Material parameters: `shared/physics/material.go`
- ISPP write strategy: `shared/physics/ispp_write.go`
- Hysteresis module UI: `module1-hysteresis/pkg/gui/`


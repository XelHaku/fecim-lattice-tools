# Module 2: Crossbar - Physics

## Prerequisites

- Ohm's law
- Matrix-vector multiplication
- Basic circuit networks

## Core Model

- A crossbar computes y = G * v, where G is conductance.
- Non-idealities such as IR drop and sneak paths perturb results.
- Quantization and noise approximate device limits.

## Key Equations (Simplified)

```
I = G * V
I_i = sum_j G_ij * V_j
V_drop ≈ I * R_wire
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| G | Conductance | Siemens |
| V | Input voltage | Volts |
| I | Output current | Amps |
| R_wire | Wire resistance | Ohms |
| G_ij | Cell conductance at row i, col j (normalized in baseline MVM) | unitless (0–1) or Siemens |
| V_j | Input voltage at column j | Volts |
| I_i | Row-summed output current | Amps |
| V_drop | Voltage drop along wire | Volts |
| N_cols | Number of columns in array | count |

## Assumptions And Limits

- Linear conductance model for ideal MVM.
- Non-idealities are simplified and not device-calibrated.
- Noise is modeled as additive perturbations.
- Baseline MVM uses normalized conductance; physical units are applied in analysis paths.

## Where It Lives In Code

- `module2-crossbar/pkg/crossbar/array.go`
- `module2-crossbar/pkg/crossbar/nonidealities.go`
- `module2-crossbar/pkg/crossbar/irdrop.go`
- `module2-crossbar/pkg/crossbar/sneakpath.go`

## Sources

- `docs/development/scriptReference.md#demo-2-crossbar-module2-crossbar`
- `docs/ELI5.md`

<!-- Category: Physics | Module: module4-circuits | Reading time: ~8 min -->
# Module 4 Physics: Peripheral Circuit Models

> Equations and models for the DAC, TIA, ADC, charge pump, and ISPP
> write controller used in the FeCIM circuit interface.

All models here are **simulation-only** -- idealized transfer functions
with optional noise paths. They are not SPICE-calibrated and no
transistor-level modeling is performed.

---

## DAC: Digital-to-Analog Converter

**Code location:** `shared/peripherals/dac.go`

The DAC maps a digital code to an analog voltage using endpoint-inclusive
mapping.

```
Levels:    L = 2^N
LSB:       LSB = (V_refHigh - V_refLow) / (L - 1)
Output:    V_dac(code) = V_refLow + clamp(code, 0, L-1) * LSB
```

| Parameter | Meaning | Units |
|-----------|---------|-------|
| N | Resolution | bits |
| L | Number of levels | count |
| V_refLow, V_refHigh | Reference voltage endpoints | V |
| LSB | Least significant bit step | V |
| code | Digital input | unitless integer |

**Nonlinearity path:** Optional INL/DNL perturbations are added before
output. DNL is tested within +/-1 LSB.

**Why 4-bit is the literature sweet spot:** At 4 bits (16 levels), the
DAC adds manageable area and power while providing sufficient input
resolution for most neural network inference workloads. Going to 5+ bits
doubles the circuit complexity for diminishing accuracy returns.

---

## TIA: Transimpedance Amplifier

**Code location:** `shared/peripherals/tia.go`

The TIA converts the crossbar output current to a voltage.

```
Ideal:     V_out,raw = I_in * R_tia + V_offset
Clamped:   V_out = clamp(V_out,raw, 0, V_max)
```

**Noise path** (ConvertWithNoise):

```
V_noise,rms = I_noise,rms * R_tia * sqrt(BW)
```

The demo path adds a deterministic +RMS offset and re-clamps.

**Virtual ground condition:** The TIA holds its input node at
approximately 0 V. In passive (0T1R) arrays, this means all wordlines
sit at 0 V -- the TIA enforces the ground reference for the entire
row side of the crossbar.

| Parameter | Meaning | Units |
|-----------|---------|-------|
| R_tia | Feedback resistance | Ohms |
| I_in | Input current from crossbar | A |
| V_offset | Input offset voltage | V |
| V_max | Output clamp ceiling | V |
| BW | Noise bandwidth | Hz |
| I_noise,rms | Input-referred noise density | A/sqrt(Hz) |

---

## ADC: Analog-to-Digital Converter

**Code location:** `shared/peripherals/adc.go`

The ADC quantizes TIA output voltage back to a digital code.

```
Levels:    L = 2^N
LSB:       LSB = (V_refHigh - V_refLow) / (L - 1)
Clamp:     V = clamp(V_in, V_refLow, V_refHigh)
Quantize:  code = round((V - V_refLow) / (V_refHigh - V_refLow) * (L - 1))
Final:     code = clamp(code, 0, L-1)
```

Quantization uses round-to-nearest with ties half-up. Ideal
quantization error bound: approximately +/-0.5 LSB.

**Energy dominance:** ADC typically consumes 40-60% of total system
energy in CIM architectures. This makes ADC resolution and architecture
(SAR, flash, sigma-delta) a critical design choice.

| Parameter | Meaning | Units |
|-----------|---------|-------|
| N | Resolution | bits |
| V_refLow, V_refHigh | Reference voltage endpoints | V |
| code | Digital output | unitless integer |

**SAR noise model:** The ADC includes optional thermal noise,
metastability noise, and reference drift models. ENOB and SNR are
reported via analysis methods.

---

## Charge Pump

**Code location:** `shared/peripherals/chargepump.go`

The charge pump boosts supply voltage for programming pulses using a
Dickson-style topology.

```
Ideal:       V_ideal = sign(V_target) * (Stages + 1) * V_in
Losses:      V_th,drop = Stages * V_diode
             R_out = 1 / (C_fly * f_clk)
             V_ir,drop = |I_load| * R_out
Actual:      |V_actual| = max(|V_ideal| - V_th,drop - V_ir,drop, 0)
             V_actual = sign(V_ideal) * |V_actual|
Regulation:  if |V_actual| > |V_target|: clamp to +/-|V_target|
```

| Parameter | Meaning | Units |
|-----------|---------|-------|
| Stages | Number of pump stages | count |
| V_in | Input supply voltage | V |
| V_diode | Per-stage diode drop | V |
| C_fly | Flying capacitance | F |
| f_clk | Clock frequency | Hz |
| I_load | Output load current | A |

---

## Noise Models

**Code location:** `shared/peripherals/noise.go`

The simulator composes multiple noise sources by variance summation:

```
Thermal:       V^2 = 4 * k * T * R * BW
Shot:          I^2 = 2 * q * I_dc * BW
Flicker (1/f): V^2 = K_f / (C_ox * A * f)
Quantization:  V^2 = LSB^2 / 12
Total:         V^2_total = sum of all variance terms
SNR(dB):       20 * log10(V_signal / sqrt(V^2_total))
```

---

## 0T1R Passive Mode Constraint

In a 0T1R passive crossbar, the DAC drives the bitline (column) and
TIA holds all wordlines at 0 V (virtual ground). Wordline voltages
cannot be independently controlled.

**Write behavior:**

```
  All WLs = 0 V (TIA virtual ground)
  Selected BL = -V_write (DAC output)
  --> All cells in selected column see full V_write
  --> Same-row cells in other columns see 0 V (safe)
```

This means writing to one cell in a column exposes the entire column
to the write voltage. The V/2 half-select scheme used in 1T1R/2T1R
architectures is physically impossible in 0T1R.

---

## ISPP: Incremental Step Pulse Programming

The ISPP controller performs a binary-search-style convergence to
program a cell to a target conductance level.

```
  Algorithm:
  1. Set search bracket [V_min, V_max]
  2. Apply voltage pulse at V_mid = (V_min + V_max) / 2
  3. Read back current level
  4. If level == target: SUCCESS
  5. If level < target:  V_min = V_mid  (need more voltage)
  6. If level > target:  V_max = V_mid  (need less voltage)
  7. Repeat from step 2

  Termination conditions:
  - Target level reached             --> SUCCESS
  - Overshoot limit reached (30)     --> SUCCESS (physics-limited)
  - Maximum iterations reached       --> FAILED
  - Bracket collapsed (V_min >= V_max) --> widen minimally and retry
```

**Guard-band logic:** After reaching the target, the controller
applies 1-2 additional pulses slightly above/below to verify stability.
Guard pulses are limited to prevent direction-flip overshoots.

**Physics limitation:** Materials with sharp switching thresholds
(like HZO) may not be able to maintain mid-range levels at zero
applied field. Repeated overshoots in this case indicate the
controller has bracketed the target voltage -- reaching the overshoot
limit triggers success, not failure.

---

## Cell Physics Integration

**Code location:** `shared/physics/cell_geometry.go`

The circuit interface connects to ferroelectric physics through these
conversions:

```
Electric field:    E = V / t_film        (t_film = 10 nm default)
Charge:            Q = P * A             (A = 100 nm^2 default)
Conductance:       G = sigma * A / t_film
```

The polarization-to-conductance transfer function maps the ferroelectric
state to a measurable conductance level, quantized to 30 discrete
levels by default.

---

## Parameters and Units Summary

| Symbol | Meaning | Units |
|--------|---------|-------|
| V_ref | Reference voltage span | V |
| N | Converter resolution | bits |
| code | Digital code | unitless integer |
| I_in | Input current | A |
| R_tia | TIA feedback resistance | Ohms |
| LSB | Least significant bit size | V |
| V_out | Output voltage | V |
| BW | Noise bandwidth | Hz |
| k | Boltzmann constant | J/K |
| T | Temperature | K |
| q | Electron charge | C |
| t_film | Ferroelectric film thickness | m |
| A | Cell area | m^2 |

---

## Assumptions and Limits

- All transfer functions are idealized and static (no full transient).
- No transistor-level modeling; not SPICE-accurate.
- Noise/nonlinearity are simplified when enabled.
- Timing and energy estimates are analytic placeholders.
- The charge pump assumes ideal flying capacitors.

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*

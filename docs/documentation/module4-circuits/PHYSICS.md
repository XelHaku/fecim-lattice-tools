# Module 4: Circuits - Physics

## Prerequisites

- Ohm's law
- Basic sampling and quantization
- Simple RC concepts

## Model Honesty Labels (Reported vs Validated)

Per `docs/comparison/HONESTY_AUDIT.md`, this module is **simulation-only**:

- **Validated (code-level):** unit-consistent analytic models with regression tests (no SPICE deck equivalence is claimed).
- **Assumed / illustrative (not silicon-validated):** default parameters and any "typical" ADC/DAC energy/timing numbers are teaching placeholders unless explicitly cited and verified.

## Core Model

- DAC maps digital codes to analog voltages.
- TIA maps analog currents to voltages.
- ADC maps analog voltages back to digital codes.

## Key Equations (Implemented in `shared/peripherals`)

### DAC (`shared/peripherals/dac.go`)

- Levels: `L = 2^N`
- LSB (code step):
  - `LSB = (VrefHigh - VrefLow) / (L - 1)`
- Ideal transfer (with code clamp to `[0, L-1]`):
  - `V_dac(code) = VrefLow + code * LSB`

Notes:
- The model uses endpoint-inclusive mapping (`L-1` in denominator), matching the code.
- Optional nonlinearity path adds deterministic INL/DNL perturbations before output.

### TIA (`shared/peripherals/tia.go`)

- Ideal transimpedance + offset:
  - `V_out,raw = I_in * R_tia + V_offset`
- Output clamp:
  - `V_out = clamp(V_out,raw, 0, V_max)`

Noise path (`ConvertWithNoise`):
- Input-referred RMS noise density to output RMS voltage:
  - `V_noise,rms = I_noise,rms * R_tia * sqrt(BW)`
- Demo path adds a deterministic +RMS offset and re-clamps.

### ADC (`shared/peripherals/adc.go`)

- Levels: `L = 2^N`
- LSB:
  - `LSB = (VrefHigh - VrefLow) / (L - 1)`
- Input clamp first: `V = clamp(V_in, VrefLow, VrefHigh)`
- Quantization (round-to-nearest, ties half-up):
  - `code = round((V - VrefLow)/(VrefHigh - VrefLow) * (L - 1))`
  - then clamp `code` to `[0, L-1]`
- Quantization error bound (ideal mid-tread assumption): `≈ ±0.5 LSB`

### Charge Pump (`shared/peripherals/chargepump.go`)

- Ideal Dickson magnitude (sign follows target polarity):
  - `V_ideal = sign(V_target) * (Stages + 1) * V_in`
- Non-ideal losses:
  - `V_th,drop = Stages * V_diode`
  - `R_out ≈ 1 / (C_fly * f_clk)`
  - `V_ir,drop = |I_load| * R_out`
- Actual output magnitude and sign:
  - `|V_actual| = max(|V_ideal| - V_th,drop - V_ir,drop, 0)`
  - `V_actual = sign(V_ideal) * |V_actual|`
- Regulation clamp to configured target magnitude:
  - if `|V_actual| > |V_target|`, output is clamped to `±|V_target|` with target sign.

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| Vref | Reference voltage span (often `Vmax - Vmin`) | Volts (V) |
| Vmin, Vmax | ADC/DAC input/output range endpoints | Volts (V) |
| N | Converter resolution | bits |
| code | Digital code | unitless integer |
| I_in | Input current | Amps (A) |
| R_tia | TIA resistance | Ohms (Ω) |
| LSB | Least significant bit size | Volts (V) |
| V_out | Output voltage | Volts (V) |

## Assumptions And Limits

- Idealized, static transfer functions by default (no full transient behavior).
- No transistor-level modeling; **not** SPICE-accurate.
- Noise/nonlinearity are simplified when enabled.
- Timing and energy estimates (when shown) are analytic placeholders unless specifically verified.

## Where It Lives In Code

- `shared/peripherals/dac.go`
- `shared/peripherals/adc.go`
- `shared/peripherals/tia.go`
- `shared/peripherals/chargepump.go`
- `module4-circuits/pkg/gui/app.go`

## Sources

- `docs/development/scriptReference.md#demo-4-circuits-module4-circuits`
- `shared/peripherals/analysis.go`

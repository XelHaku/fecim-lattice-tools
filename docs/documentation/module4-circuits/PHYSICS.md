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

## Key Equations (Simplified)

### DAC (Ideal linear code-to-voltage)

- `LSB = Vref / (2^N)`
- `V_dac(code) = Vmin + code * LSB` (clamped to `[Vmin, Vmax]`)

### TIA (Ohmic transimpedance)

- `V_out = I_in * R_tia`

### ADC (Uniform quantization)

- `code = round((V_in - Vmin) / LSB)` (clamped to valid code range)
- `QuantizationError ≈ ±0.5 LSB`

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

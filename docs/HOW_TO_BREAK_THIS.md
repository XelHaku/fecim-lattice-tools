# How To Break This Tool

This document records stress cases and known trust boundaries. Some entries are confirmed limitations; others are adversarial tests that should be run through The Crucible.

## Module 1: Hysteresis

Stress cases:

- electric fields far beyond the material preset range
- temperatures outside characterized range
- negative remnant polarization
- zero coercive field
- tiny Preisach grids versus high-resolution grids
- minor loops that should remain bounded by major loops
- unit conversions between `uC/cm^2` and `C/m^2`

Expected response:

- fail clearly for invalid inputs
- label extrapolation
- avoid presenting unsupported results as physical predictions

## Module 2: Crossbar

Stress cases:

- `1x1` arrays
- `64x64`, `128x128`, and larger arrays
- all-zero conductance matrices
- all-maximum conductance matrices
- negative voltages
- zero wire resistance
- extremely high wire resistance
- passive arrays where sneak paths should be visible

Expected response:

- conserve current where the model applies
- avoid divide-by-zero metrics
- show IR drop in the physically expected direction
- label unsupported regimes

## Module 3: MNIST

Stress cases:

- zero training epochs
- one-level quantization
- high-level quantization near FP32 behavior
- all-zero weights
- corrupted input images
- multiple seeds

Expected response:

- report deterministic metrics when seeded
- fail gracefully for invalid data
- avoid stale accuracy claims

## Module 4: Peripherals

Stress cases:

- DAC code above valid range
- `V_ref = 0`
- ADC input beyond range
- TIA current beyond compliance
- high-frequency assumptions
- stochastic noise with fixed seeds

Expected response:

- clamp or reject invalid values explicitly
- report saturation
- document idealized formulas and missing non-idealities

## Module 6: EDA

Stress cases:

- degenerate generated arrays
- large generated arrays
- invalid dimensions
- overlapping physical placements
- placeholder timing data
- full OpenLane flow

Expected response:

- generated files parse when claimed
- placeholder data is labeled
- no claim of tape-out readiness without signoff artifacts


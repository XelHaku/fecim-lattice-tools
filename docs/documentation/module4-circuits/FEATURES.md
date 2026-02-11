# Module 4: Circuits - Features

## What This Module Does

- Models DAC, ADC, TIA, and charge-pump behavior for FeCIM arrays.
- Estimates timing and power for peripheral operations.
- Visualizes signal flow, voltage zones, and architecture-specific rules.

## Primary Components

All peripheral circuit physics for Module 4 is sourced from `shared/peripherals`:

- `shared/peripherals/dac.go`
- `shared/peripherals/adc.go`
- `shared/peripherals/tia.go`
- `shared/peripherals/chargepump.go`
- `shared/peripherals/analysis.go`
- `module4-circuits/pkg/gui/app.go`

## Key Workflows

- Convert digital inputs to analog voltages for array drive.
- Convert array currents into voltages and digital codes.
- Estimate timing and power breakdown for conversions.

## Extension Points

- Add new ADC/DAC architectures or nonlinearity models.
- Extend power analysis with additional blocks.
- Connect to exported SPICE netlists from Module 6.

## Known Limitations

- Behavior is analytic, not SPICE-accurate.
- Parameter defaults are for teaching, not silicon tuning.
- Timing is approximate and does not include full routing effects.

// Package peripherals provides shared peripheral circuit configurations for FeCIM systems.
package peripherals

import "fecim-lattice-tools/shared/physics"

// Standard peripheral configuration constants.
// Updated 2026-03-18: Changed ADC default from 4-bit to 6-bit. 4-bit produced only
// 5 distinct codes for 30 conductance levels (effective ~2.3 bits). 6-bit (64 codes)
// resolves all 30 levels with margin.
const (
	// DefaultBits is the standard resolution for ADC (6 bits = 64 levels).
	// DAC remains 4-bit; this constant now reflects the ADC default.
	DefaultBits = 6

	// DefaultLevels is the number of discrete ADC levels (2^6 = 64).
	DefaultLevels = 64

	// FeCIMLevels is the baseline number of analog states used in this demo.
	// Re-exported from shared/physics for backward compatibility.
	FeCIMLevels = physics.DefaultLevels
)

// DAC reference voltage constants
const (
	// DACVrefHigh is the high reference voltage for write operations (+1.5V).
	// Heuristic simulation baseline for bipolar write sweeps.
	DACVrefHigh = 1.5

	// DACVrefLow is the low reference voltage for write operations (-1.5V).
	// Heuristic simulation baseline for bipolar write sweeps.
	DACVrefLow = -1.5

	// DACSettleTime is the typical DAC settling time in nanoseconds.
	DACSettleTime = 10.0
)

// ADC reference voltage constants
const (
	// ADCVrefHigh is the high reference voltage for read operations (1.0V).
	// Set to align with default TIA full-scale output.
	ADCVrefHigh = 1.0

	// ADCVrefLow is the low reference voltage for read operations (0.0V).
	ADCVrefLow = 0.0

	// ADCConversionTime is the typical SAR ADC conversion time in nanoseconds.
	ADCConversionTime = 50.0
)

// Nonlinearity specifications (in LSB)
const (
	// DefaultINL is a placeholder integral nonlinearity for SAR-style studies.
	DefaultINL = 0.5

	// DefaultDNL is a placeholder differential nonlinearity for SAR-style studies.
	DefaultDNL = 0.25
)

// NOTE: Full DAC, ADC, TIA, ChargePump structs and their methods are now in:
//   - dac.go (DAC struct with DefaultDAC())
//   - adc.go (ADC struct with DefaultADC(), ADCType enum)
//   - tia.go (TIA struct with DefaultTIA())
//   - chargepump.go (ChargePump struct with DefaultChargePump())
//   - analysis.go (INL/DNL analysis, timing, power breakdown)
//
// The simplified Config structs that were here have been removed to avoid
// duplication. Use the full structs from the individual files instead.

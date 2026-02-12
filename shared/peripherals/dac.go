// Package peripherals provides peripheral circuit models for ferroelectric CIM systems.
package peripherals

import (
	"fecim-lattice-tools/shared/logging"
)

var log = logging.NewLogger("peripherals")

// DAC represents a Digital-to-Analog Converter for crossbar write operations.
// 5-bit DAC maps 30 discrete levels to write voltages.
type DAC struct {
	Bits       int     // Resolution in bits (5 for 30 levels)
	VrefHigh   float64 // High reference voltage (+1.5V)
	VrefLow    float64 // Low reference voltage (-1.5V)
	INL        float64 // Integral nonlinearity (LSB)
	DNL        float64 // Differential nonlinearity (LSB)
	SettleTime float64 // Settling time (ns)
}

// DefaultDAC returns a DAC configured for FeCIM 30-level operation.
func DefaultDAC() *DAC {
	dac := &DAC{
		Bits:       5,    // 32 levels, we use 30
		VrefHigh:   1.5,  // +1.5V for positive write
		VrefLow:    -1.5, // -1.5V for negative write
		INL:        0.5,  // 0.5 LSB INL
		DNL:        0.25, // 0.25 LSB DNL
		SettleTime: 10,   // 10 ns settling time
	}
	log.Calculation("DefaultDAC", map[string]interface{}{
		"bits":        dac.Bits,
		"vref_high":   dac.VrefHigh,
		"vref_low":    dac.VrefLow,
		"inl":         dac.INL,
		"dnl":         dac.DNL,
		"settle_time": dac.SettleTime,
	}, dac)
	return dac
}

// EnableLogging re-initializes the peripherals logger after file logging is enabled.
// Call logging.EnableFileLogging() first, then this function.
func EnableLogging() {
	log = logging.NewLogger("peripherals")
}

// Levels returns the number of discrete output levels.
func (d *DAC) Levels() int {
	return 1 << d.Bits
}

// Convert maps a digital level (0-29) to an analog voltage.
func (d *DAC) Convert(level int) float64 {
	log.Input("DAC.Convert", map[string]interface{}{
		"level":     level,
		"bits":      d.Bits,
		"vref_high": d.VrefHigh,
		"vref_low":  d.VrefLow,
	})

	if level < 0 {
		level = 0
	}
	maxLevel := d.Levels() - 1
	if level > maxLevel {
		level = maxLevel
	}

	// Linear interpolation between Vref voltages
	fraction := float64(level) / float64(maxLevel)
	voltage := d.VrefLow + fraction*(d.VrefHigh-d.VrefLow)

	log.Calculation("DAC.Convert", map[string]interface{}{
		"level": level,
	}, voltage)

	return voltage
}

// ConvertWithNonlinearity adds INL/DNL errors to conversion at nominal PVT.
func (d *DAC) ConvertWithNonlinearity(level int) float64 {
	log.Input("DAC.ConvertWithNonlinearity", map[string]interface{}{
		"level": level,
		"inl":   d.INL,
		"dnl":   d.DNL,
	})

	result := d.ConvertWithCondition(level, referenceTemperatureK, CornerTypical)

	log.Calculation("DAC.ConvertWithNonlinearity", map[string]interface{}{
		"level": level,
	}, result)

	return result
}

// VoltageRange returns the full output voltage range.
func (d *DAC) VoltageRange() (min, max float64) {
	return d.VrefLow, d.VrefHigh
}

// Resolution returns the voltage per LSB.
func (d *DAC) Resolution() float64 {
	return (d.VrefHigh - d.VrefLow) / float64(d.Levels()-1)
}

// EnergyPerConversion estimates energy consumption per DAC conversion.
// Based on typical switched-capacitor DAC.
func (d *DAC) EnergyPerConversion() float64 {
	log.Input("DAC.EnergyPerConversion", map[string]interface{}{
		"bits":      d.Bits,
		"vref_high": d.VrefHigh,
		"vref_low":  d.VrefLow,
	})

	// Energy ~ C_eff * (Vspan/2)^2 * 2^N (differential, switched-cap DAC)
	// C_eff is an effective unit capacitor calibrated to typical 5-bit DAC energy.
	capacitance := 0.2e-15 // 0.2 fF effective unit capacitor
	levels := float64(d.Levels())
	vrefSpan := d.VrefHigh - d.VrefLow
	vref := vrefSpan / 2.0

	energy := capacitance * vref * vref * levels // ~15 fJ typical

	log.Calculation("DAC.EnergyPerConversion", map[string]interface{}{
		"capacitance": capacitance,
		"levels":      levels,
		"vref_span":   vrefSpan,
		"vref":        vref,
	}, energy)

	return energy
}

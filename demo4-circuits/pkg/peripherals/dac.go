// Package peripherals provides peripheral circuit models for ferroelectric CIM systems.
package peripherals

import (
	"math"
)

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
	return &DAC{
		Bits:       5,    // 32 levels, we use 30
		VrefHigh:   1.5,  // +1.5V for positive write
		VrefLow:    -1.5, // -1.5V for negative write
		INL:        0.5,  // 0.5 LSB INL
		DNL:        0.25, // 0.25 LSB DNL
		SettleTime: 10,   // 10 ns settling time
	}
}

// Levels returns the number of discrete output levels.
func (d *DAC) Levels() int {
	return 1 << d.Bits
}

// Convert maps a digital level (0-29) to an analog voltage.
func (d *DAC) Convert(level int) float64 {
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

	return voltage
}

// ConvertWithNonlinearity adds INL/DNL errors to conversion.
func (d *DAC) ConvertWithNonlinearity(level int) float64 {
	idealVoltage := d.Convert(level)

	// LSB size
	lsb := (d.VrefHigh - d.VrefLow) / float64(d.Levels()-1)

	// Add INL error (varies with code)
	inlError := d.INL * lsb * math.Sin(math.Pi*float64(level)/float64(d.Levels()-1))

	// Add DNL error (random per level)
	dnlError := d.DNL * lsb * (0.5 - float64(level%3)/2.0)

	return idealVoltage + inlError + dnlError
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
	// Energy ~ C * Vref^2 * 2^N
	// Typical: ~1 fJ/conversion-step for 65nm CMOS
	capacitance := 1e-15 // 1 fF unit capacitor
	levels := float64(d.Levels())
	vref := (d.VrefHigh - d.VrefLow) / 2

	return capacitance * vref * vref * levels // ~15 fJ typical
}

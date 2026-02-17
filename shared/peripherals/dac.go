// Package peripherals provides peripheral circuit models for ferroelectric CIM systems.
package peripherals

import (
	"fecim-lattice-tools/shared/logging"
)

var log = logging.NewLogger("peripherals")

// DACEncoding describes the internal code representation of a DAC.
// The output voltage is the same regardless of encoding; the difference is in
// switching energy and glitch behavior during code transitions.
type DACEncoding int

const (
	// DACEncodingBinary is a standard binary-weighted DAC.
	// Pro: fewest unit cells (2^N); Con: code-dependent glitch (e.g. 7→8 flips all bits).
	DACEncodingBinary DACEncoding = iota
	// DACEncodingThermometer uses 2^N-1 unary unit cells, one per output level.
	// Pro: monotone transitions, zero code-dependent glitch; Con: more cells/area.
	DACEncodingThermometer
)

// DAC models the write-path digital-to-analog converter that generates program
// voltages (V) from discrete control codes for ferroelectric state updates.
//
// Default 4-bit configuration (literature optimal) provides 16 codes.
// Can be increased to 5-bit (32 codes) or higher for high-precision applications.
type DAC struct {
	Bits       int         // Resolution in bits (4-bit default per 2024-2025 literature)
	VrefHigh   float64     // High reference voltage (+1.5V)
	VrefLow    float64     // Low reference voltage (-1.5V)
	INL        float64     // Integral nonlinearity (LSB)
	DNL        float64     // Differential nonlinearity (LSB)
	SettleTime float64     // Settling time (ns)
	Encoding   DACEncoding // Internal code representation (binary default)
}

// DefaultDAC returns a DAC configured for FeCIM 30-level operation.
func DefaultDAC() *DAC {
	dac := &DAC{
		Bits:       4,    // 16 levels (4-bit optimal per 2024-2025 literature consensus)
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

// Convert maps a digital code to output voltage (V) across the configured
// reference span, with input clamped to valid code range.
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

// Resolution returns DAC output granularity in V/LSB.
func (d *DAC) Resolution() float64 {
	return (d.VrefHigh - d.VrefLow) / float64(d.Levels()-1)
}

// EnergyPerConversion estimates per-update DAC energy (J) using a simplified
// switched-capacitor scaling model.
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

// LIT-P3-04: Thermometer/segmented DAC encoding.
// LIT-P3-05: DAC glitch energy model.

// ThermometerDAC returns a DAC pre-configured for thermometer encoding.
//
// A thermometer DAC uses 2^N-1 unary unit cells. At level k, exactly k cells
// are "on". Only one cell toggles per single-step transition — eliminating the
// code-dependent glitch that occurs in binary DACs at major carry transitions
// (e.g. 0111→1000 flips all 4 bits simultaneously).
//
// Reference: IEEE JSSC 2018 — "Glitch-free DAC for FeRAM write paths".
func ThermometerDAC(bits int) *DAC {
	if bits < 1 {
		bits = 1
	}
	dac := DefaultDAC()
	dac.Bits = bits
	dac.Encoding = DACEncodingThermometer
	return dac
}

// GlitchTransitions returns the number of unit-cell transitions when the DAC
// switches from code `from` to code `to`.
//
// Binary DAC: transitions = popcount(from XOR to) — bits that must switch.
// Thermometer DAC: transitions = |to - from| — only one cell per step, no
// simultaneous multi-bit toggle regardless of code distance.
func (d *DAC) GlitchTransitions(from, to int) int {
	maxCode := d.Levels() - 1
	if from < 0 {
		from = 0
	}
	if from > maxCode {
		from = maxCode
	}
	if to < 0 {
		to = 0
	}
	if to > maxCode {
		to = maxCode
	}
	switch d.Encoding {
	case DACEncodingThermometer:
		// Thermometer: each step changes exactly one cell — no simultaneous glitch.
		diff := to - from
		if diff < 0 {
			diff = -diff
		}
		return diff
	default: // DACEncodingBinary
		// Binary: count differing bits (popcount of XOR).
		xorVal := from ^ to
		count := 0
		for xorVal != 0 {
			count += xorVal & 1
			xorVal >>= 1
		}
		return count
	}
}

// GlitchEnergy returns the energy (Joules) dissipated by code-dependent
// glitching when the DAC transitions from code `from` to code `to`.
//
// Model: each switching unit cell contributes E_glitch = C_unit × V_lsb²
// where C_unit = 0.2 fF (same as EnergyPerConversion unit cap) and
// V_lsb = (VrefHigh-VrefLow)/(2^N-1).
//
// Binary DAC: glitch = popcount(from XOR to) × C_unit × V_lsb²
// Thermometer DAC: glitch ≈ 0 (single-step transitions, no multi-bit glitch)
func (d *DAC) GlitchEnergy(from, to int) float64 {
	const cUnit = 0.2e-15 // 0.2 fF unit capacitor
	vLSB := (d.VrefHigh - d.VrefLow) / float64(d.Levels()-1)
	transitions := d.GlitchTransitions(from, to)
	if d.Encoding == DACEncodingThermometer {
		// Thermometer: glitch only when multiple cells switch simultaneously.
		// Single-step transitions have zero glitch; larger jumps are sequential.
		return 0
	}
	return float64(transitions) * cUnit * vLSB * vLSB
}

// Package arraysim provides approximate array coupling solvers for module4-circuits.
package arraysim

import (
	"math"
	"math/rand"
)

// TIAConfig captures the minimal transimpedance amplifier model.
type TIAConfig struct {
	Rf   float64 // Feedback resistance (Ohm)
	Vref float64 // Output reference (V)
	Vmin float64 // Minimum output rail (V)
	Vmax float64 // Maximum output rail (V)
}

// ADCConfig captures the minimal ADC model for sense conversion.
//
// Quantization policy:
//   - The ADC code is computed by rounding to the nearest integer code (ties half-up).
//   - Inputs are clamped to [Vmin, Vmax] before quantization.
//   - The output code is clamped to [0, 2^Bits-1].
//
// This matches the GUI/CPU peripheral model in shared/peripherals/adc.go and the GPU shader
// in module4-circuits/shaders/adc.comp.
type ADCConfig struct {
	Bits int     // Resolution in bits
	Vmin float64 // Minimum input voltage (V)
	Vmax float64 // Maximum input voltage (V)
}

// SenseChain chains a TIA and ADC model for current sensing.
type SenseChain struct {
	TIA TIAConfig
	ADC ADCConfig
}

// SenseResult captures the sense conversion outputs with saturation flags.
type SenseResult struct {
	Vout         float64
	Code         int
	TIASaturated bool
	ADCSaturated bool
}

// ConvertCurrent converts a single row current to TIA output and ADC code.
func (s SenseChain) ConvertCurrent(currentA float64) SenseResult {
	vout := s.TIA.Vref + currentA*s.TIA.Rf
	tiaSat := false
	if vout < s.TIA.Vmin {
		vout = s.TIA.Vmin
		tiaSat = true
	}
	if vout > s.TIA.Vmax {
		vout = s.TIA.Vmax
		tiaSat = true
	}

	adcBits := s.ADC.Bits
	if adcBits < 1 {
		adcBits = 1
	}
	levels := 1 << adcBits
	adcSat := false
	adcV := vout
	if adcV < s.ADC.Vmin {
		adcV = s.ADC.Vmin
		adcSat = true
	}
	if adcV > s.ADC.Vmax {
		adcV = s.ADC.Vmax
		adcSat = true
	}
	if tiaSat {
		adcSat = true
	}

	fraction := 0.0
	span := s.ADC.Vmax - s.ADC.Vmin
	if span > 0 {
		fraction = (adcV - s.ADC.Vmin) / span
	}
	// Round-to-nearest, ties half-up (same convention as shared/peripherals/adc.go).
	code := int(math.Floor(fraction*float64(levels-1) + 0.5))
	if code < 0 {
		code = 0
	}
	if code >= levels {
		code = levels - 1
	}

	return SenseResult{
		Vout:         vout,
		Code:         code,
		TIASaturated: tiaSat,
		ADCSaturated: adcSat,
	}
}

// ConvertCurrents converts a slice of row currents.
func (s SenseChain) ConvertCurrents(currents []float64) []SenseResult {
	results := make([]SenseResult, len(currents))
	for i, current := range currents {
		results[i] = s.ConvertCurrent(current)
	}
	return results
}

func (s SenseChain) ConvertCurrentWithNoise(currentA float64, rng *rand.Rand) SenseResult {
	if rng == nil {
		return s.ConvertCurrent(currentA)
	}
	if s.TIA.Rf <= 0 {
		return s.ConvertCurrent(currentA)
	}
	kT := 1.38e-23 * 300.0
	bw := 100e6
	noiseSigma := math.Sqrt(4 * kT * bw / s.TIA.Rf)
	noiseI := rng.NormFloat64() * noiseSigma
	return s.ConvertCurrent(currentA + noiseI)
}

// EffectiveVoltageRange returns the measurable voltage window after TIA rails and ADC references are applied.
func (s SenseChain) EffectiveVoltageRange() (float64, float64) {
	vmin := s.TIA.Vmin
	vmax := s.TIA.Vmax
	if s.ADC.Vmin > vmin {
		vmin = s.ADC.Vmin
	}
	if s.ADC.Vmax < vmax {
		vmax = s.ADC.Vmax
	}
	return vmin, vmax
}

// CurrentRange returns the measurable input current window derived from the effective voltage range.
func (s SenseChain) CurrentRange() (float64, float64) {
	if s.TIA.Rf <= 0 {
		return 0, 0
	}
	vmin, vmax := s.EffectiveVoltageRange()
	if vmax < vmin {
		return 0, 0
	}
	imin := (vmin - s.TIA.Vref) / s.TIA.Rf
	imax := (vmax - s.TIA.Vref) / s.TIA.Rf
	return imin, imax
}

// CurrentLSB returns the current represented by one ADC LSB within the effective voltage range.
func (s SenseChain) CurrentLSB() float64 {
	if s.TIA.Rf <= 0 {
		return 0
	}
	adcBits := s.ADC.Bits
	if adcBits < 1 {
		adcBits = 1
	}
	levels := 1 << adcBits
	if levels < 2 {
		return 0
	}
	vmin, vmax := s.EffectiveVoltageRange()
	span := vmax - vmin
	if span <= 0 {
		return 0
	}
	lsbV := span / float64(levels-1)
	return lsbV / s.TIA.Rf
}

package peripherals

import (
	"math"
)

// ADC represents an Analog-to-Digital Converter for crossbar read operations.
// 5-bit ADC quantizes current/voltage to 30 discrete levels.
type ADC struct {
	Bits           int     // Resolution in bits
	VrefHigh       float64 // High reference voltage
	VrefLow        float64 // Low reference voltage
	INL            float64 // Integral nonlinearity (LSB)
	DNL            float64 // Differential nonlinearity (LSB)
	ConversionTime float64 // Conversion time (ns)
	Type           ADCType // Architecture type
}

// ADCType specifies the ADC architecture.
type ADCType int

const (
	ADCTypeSAR        ADCType = iota // Successive Approximation Register
	ADCTypeFlash                     // Flash (parallel)
	ADCTypeSigmaDelta                // Sigma-Delta (oversampling)
)

// DefaultADC returns an ADC configured for FeCIM read operations.
func DefaultADC() *ADC {
	return &ADC{
		Bits:           5,    // 32 levels, we use 30
		VrefHigh:       1.0,  // Max sense voltage
		VrefLow:        0.0,  // Min sense voltage
		INL:            0.5,  // 0.5 LSB INL
		DNL:            0.25, // 0.25 LSB DNL
		ConversionTime: 50,   // 50 ns for SAR
		Type:           ADCTypeSAR,
	}
}

// Levels returns the number of discrete output levels.
func (a *ADC) Levels() int {
	return 1 << a.Bits
}

// Convert performs ADC conversion from voltage to digital level.
func (a *ADC) Convert(voltage float64) int {
	// Clamp input
	if voltage < a.VrefLow {
		voltage = a.VrefLow
	}
	if voltage > a.VrefHigh {
		voltage = a.VrefHigh
	}

	// Linear quantization
	fraction := (voltage - a.VrefLow) / (a.VrefHigh - a.VrefLow)
	level := int(fraction*float64(a.Levels()-1) + 0.5)

	if level >= a.Levels() {
		level = a.Levels() - 1
	}

	return level
}

// ConvertWithNonlinearity adds INL/DNL errors to conversion.
func (a *ADC) ConvertWithNonlinearity(voltage float64) int {
	// Add noise before quantization
	lsb := (a.VrefHigh - a.VrefLow) / float64(a.Levels()-1)

	// INL error (code-dependent)
	idealLevel := a.Convert(voltage)
	inlOffset := a.INL * lsb * math.Sin(math.Pi*float64(idealLevel)/float64(a.Levels()-1))

	// DNL error (random per level)
	dnlOffset := a.DNL * lsb * (0.5 - float64(idealLevel%5)/4.0)

	// Re-convert with errors
	return a.Convert(voltage + inlOffset + dnlOffset)
}

// Resolution returns the voltage per LSB.
func (a *ADC) Resolution() float64 {
	return (a.VrefHigh - a.VrefLow) / float64(a.Levels()-1)
}

// ENOB returns the Effective Number of Bits considering nonlinearity.
func (a *ADC) ENOB() float64 {
	// ENOB = Bits - log2(1 + INL^2 + DNL^2)^0.5
	noiseFactorSq := 1.0 + a.INL*a.INL + a.DNL*a.DNL
	enob := float64(a.Bits) - math.Log2(math.Sqrt(noiseFactorSq))
	return enob
}

// EnergyPerConversion estimates energy per ADC conversion.
func (a *ADC) EnergyPerConversion() float64 {
	switch a.Type {
	case ADCTypeSAR:
		// SAR: ~1-10 fJ/conversion-step
		return 5e-15 * float64(a.Bits) // ~25 fJ for 5-bit
	case ADCTypeFlash:
		// Flash: ~50-100 fJ/level (2^N comparators)
		return 50e-15 * float64(a.Levels())
	case ADCTypeSigmaDelta:
		// Sigma-Delta: higher due to oversampling
		return 100e-15 * float64(a.Bits)
	default:
		return 25e-15
	}
}

// TheoreticalSNR returns theoretical signal-to-noise ratio in dB.
func (a *ADC) TheoreticalSNR() float64 {
	// SNR = 6.02 * N + 1.76 dB
	return 6.02*float64(a.Bits) + 1.76
}

// EffectiveSNR returns SNR accounting for nonlinearity.
func (a *ADC) EffectiveSNR() float64 {
	return 6.02*a.ENOB() + 1.76
}

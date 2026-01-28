// Package peripherals provides shared peripheral circuit configurations for FeCIM systems.
package peripherals

// Standard peripheral configuration constants for FeCIM 30-level operation.
const (
	// DefaultBits is the standard resolution for ADC/DAC (5 bits = 32 levels, we use 30).
	DefaultBits = 5

	// DefaultLevels is the number of discrete levels (2^5 = 32, but FeCIM uses 30).
	DefaultLevels = 32

	// FeCIMLevels is the actual number of analog states used in FeCIM.
	FeCIMLevels = 30
)

// DAC reference voltage constants
const (
	// DACVrefHigh is the high reference voltage for write operations (+1.5V).
	DACVrefHigh = 1.5

	// DACVrefLow is the low reference voltage for write operations (-1.5V).
	DACVrefLow = -1.5

	// DACSettleTime is the typical DAC settling time in nanoseconds.
	DACSettleTime = 10.0
)

// ADC reference voltage constants
const (
	// ADCVrefHigh is the high reference voltage for read operations (1.0V).
	ADCVrefHigh = 1.0

	// ADCVrefLow is the low reference voltage for read operations (0.0V).
	ADCVrefLow = 0.0

	// ADCConversionTime is the typical SAR ADC conversion time in nanoseconds.
	ADCConversionTime = 50.0
)

// Nonlinearity specifications (in LSB)
const (
	// DefaultINL is the typical integral nonlinearity.
	DefaultINL = 0.5

	// DefaultDNL is the typical differential nonlinearity.
	DefaultDNL = 0.25
)

// DACConfig holds DAC configuration parameters.
type DACConfig struct {
	Bits       int     // Resolution in bits
	VrefHigh   float64 // High reference voltage (V)
	VrefLow    float64 // Low reference voltage (V)
	INL        float64 // Integral nonlinearity (LSB)
	DNL        float64 // Differential nonlinearity (LSB)
	SettleTime float64 // Settling time (ns)
}

// DefaultDACConfig returns the standard DAC configuration for FeCIM.
func DefaultDACConfig() DACConfig {
	return DACConfig{
		Bits:       DefaultBits,
		VrefHigh:   DACVrefHigh,
		VrefLow:    DACVrefLow,
		INL:        DefaultINL,
		DNL:        DefaultDNL,
		SettleTime: DACSettleTime,
	}
}

// ADCType specifies the ADC architecture.
type ADCType int

const (
	// ADCTypeSAR is a Successive Approximation Register ADC (low power, moderate speed).
	ADCTypeSAR ADCType = iota

	// ADCTypeFlash is a Flash/Parallel ADC (high speed, high power).
	ADCTypeFlash

	// ADCTypeSigmaDelta is a Sigma-Delta ADC (high resolution, low speed).
	ADCTypeSigmaDelta
)

// String returns the string representation of the ADC type.
func (t ADCType) String() string {
	switch t {
	case ADCTypeSAR:
		return "SAR"
	case ADCTypeFlash:
		return "Flash"
	case ADCTypeSigmaDelta:
		return "Sigma-Delta"
	default:
		return "Unknown"
	}
}

// ADCConfig holds ADC configuration parameters.
type ADCConfig struct {
	Bits           int     // Resolution in bits
	VrefHigh       float64 // High reference voltage (V)
	VrefLow        float64 // Low reference voltage (V)
	INL            float64 // Integral nonlinearity (LSB)
	DNL            float64 // Differential nonlinearity (LSB)
	ConversionTime float64 // Conversion time (ns)
	Type           ADCType // Architecture type
}

// DefaultADCConfig returns the standard ADC configuration for FeCIM.
func DefaultADCConfig() ADCConfig {
	return ADCConfig{
		Bits:           DefaultBits,
		VrefHigh:       ADCVrefHigh,
		VrefLow:        ADCVrefLow,
		INL:            DefaultINL,
		DNL:            DefaultDNL,
		ConversionTime: ADCConversionTime,
		Type:           ADCTypeSAR,
	}
}

// Resolution calculates the voltage per LSB for a given configuration.
func (c DACConfig) Resolution() float64 {
	levels := 1 << c.Bits
	return (c.VrefHigh - c.VrefLow) / float64(levels-1)
}

// Resolution calculates the voltage per LSB for a given configuration.
func (c ADCConfig) Resolution() float64 {
	levels := 1 << c.Bits
	return (c.VrefHigh - c.VrefLow) / float64(levels-1)
}

// Levels returns the number of discrete levels for the DAC.
func (c DACConfig) Levels() int {
	return 1 << c.Bits
}

// Levels returns the number of discrete levels for the ADC.
func (c ADCConfig) Levels() int {
	return 1 << c.Bits
}

// TIAConfig holds Transimpedance Amplifier configuration.
type TIAConfig struct {
	Gain      float64 // Transimpedance gain (Ohms)
	Bandwidth float64 // Bandwidth (Hz)
	Noise     float64 // Input-referred noise (V RMS)
	Offset    float64 // Input offset voltage (V)
}

// DefaultTIAConfig returns the standard TIA configuration for FeCIM.
func DefaultTIAConfig() TIAConfig {
	return TIAConfig{
		Gain:      10e3,  // 10 kOhm
		Bandwidth: 100e6, // 100 MHz
		Noise:     100e-6, // 100 µV RMS
		Offset:    5e-3,  // 5 mV
	}
}

// SettlingTime estimates TIA settling time from bandwidth (10-90% rise time).
func (c TIAConfig) SettlingTime() float64 {
	// Rise time ≈ 0.35 / BW
	return 0.35 / c.Bandwidth
}

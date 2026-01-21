package peripherals

import (
	"math"
)

// TIA represents a Transimpedance Amplifier for current-to-voltage conversion.
// Used in crossbar read path to sense column currents.
type TIA struct {
	Gain             float64 // Transimpedance gain (Ohms)
	Bandwidth        float64 // -3dB bandwidth (Hz)
	InputNoiseRMS    float64 // Input-referred noise (A/sqrt(Hz))
	OutputOffset     float64 // Output offset voltage (V)
	MaxInputCurrent  float64 // Maximum input current (A)
	MaxOutputVoltage float64 // Maximum output voltage (V)
}

// DefaultTIA returns a TIA configured for crossbar sense operations.
func DefaultTIA() *TIA {
	return &TIA{
		Gain:             10e3,   // 10 kΩ transimpedance
		Bandwidth:        100e6,  // 100 MHz bandwidth
		InputNoiseRMS:    1e-12,  // 1 pA/sqrt(Hz) input noise
		OutputOffset:     5e-3,   // 5 mV output offset
		MaxInputCurrent:  100e-6, // 100 µA max input
		MaxOutputVoltage: 1.0,    // 1V max output
	}
}

// Convert performs current-to-voltage conversion.
func (t *TIA) Convert(current float64) float64 {
	// Vout = Iin * Gain + Offset
	output := current*t.Gain + t.OutputOffset

	// Clamp to output range
	if output < 0 {
		output = 0
	}
	if output > t.MaxOutputVoltage {
		output = t.MaxOutputVoltage
	}

	return output
}

// ConvertWithNoise adds thermal noise to conversion.
func (t *TIA) ConvertWithNoise(current float64) float64 {
	idealOutput := t.Convert(current)

	// Calculate RMS noise voltage
	// Vnoise = Inoise * Gain * sqrt(BW)
	noiseVoltage := t.InputNoiseRMS * t.Gain * math.Sqrt(t.Bandwidth)

	// For simulation, use gaussian approximation
	// In real implementation, would use random noise
	noiseContribution := noiseVoltage * 0.1 // ~10% of RMS for demo

	return idealOutput + noiseContribution
}

// SNR returns the signal-to-noise ratio for a given input current.
func (t *TIA) SNR(current float64) float64 {
	signal := current * t.Gain
	noise := t.InputNoiseRMS * t.Gain * math.Sqrt(t.Bandwidth)

	if noise == 0 {
		return math.Inf(1)
	}

	return 20 * math.Log10(signal/noise) // dB
}

// MinDetectableCurrent returns minimum detectable current (SNR=1).
func (t *TIA) MinDetectableCurrent() float64 {
	// I_min = Inoise * sqrt(BW)
	return t.InputNoiseRMS * math.Sqrt(t.Bandwidth)
}

// DynamicRange returns the dynamic range in dB.
func (t *TIA) DynamicRange() float64 {
	minCurrent := t.MinDetectableCurrent()
	return 20 * math.Log10(t.MaxInputCurrent/minCurrent)
}

// SettlingTime estimates the step response settling time.
func (t *TIA) SettlingTime() float64 {
	// Single-pole settling: t = ln(1/accuracy) / (2*pi*BW)
	// For 0.1% accuracy: t ≈ 7 / (2*pi*BW)
	accuracy := 0.001 // 0.1%
	return math.Log(1/accuracy) / (2 * math.Pi * t.Bandwidth)
}

// PowerConsumption estimates TIA power based on bandwidth and gain.
func (t *TIA) PowerConsumption() float64 {
	// Typical: P ≈ 2 * kT * BW * Gain / η
	// Simplified estimate
	kT := 4.14e-21 // kT at 300K
	efficiency := 0.1
	return 2 * kT * t.Bandwidth * t.Gain / efficiency
}

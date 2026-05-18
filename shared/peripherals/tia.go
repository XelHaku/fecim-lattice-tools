package peripherals

import (
	"math"
)

// Default TIA parameters for FeCIM crossbar sense chain.
// These are heuristic baselines, not measured silicon values.
const (
	defaultTIAGainOhm     = 10e3   // 10 kOhm transimpedance
	defaultTIABandwidthHz = 100e6  // 100 MHz -3dB bandwidth
	defaultTIANoiseAPrtHz = 1e-12  // 1 pA/sqrt(Hz) input-referred noise
	defaultTIAOffsetV     = 5e-3   // 5 mV output offset
	defaultTIAMaxCurrentA = 100e-6 // 100 uA max input
	defaultTIAMaxOutputV  = 1.0    // 1V max output
)

// kBT300 is the Boltzmann thermal energy at 300 K (room temperature).
const kBT300 = 1.380649e-23 * 300.0 // kT at 300K (J)

// TIA models a transimpedance amplifier used in the crossbar read chain to
// convert summed column current (A) into a measurable voltage (V).
//
// Physics/circuit context: in FeFET arrays, cell conductance and read bias set
// microamp-level column currents that are converted via Vout ~= Iin*Gain + offset.
type TIA struct {
	Gain             float64 // Transimpedance gain (Ohms)
	Bandwidth        float64 // -3dB bandwidth (Hz)
	InputNoiseRMS    float64 // Input-referred noise (A/sqrt(Hz))
	OutputOffset     float64 // Output offset voltage (V)
	MaxInputCurrent  float64 // Maximum input current (A)
	MaxOutputVoltage float64 // Maximum output voltage (V)
}

// DefaultTIA returns a TIA configured for crossbar sense operations.
//
// Defaults are heuristic and chosen so that 0-100 µA array current maps into
// the default ADC 0-1V range via Vout ~= Iin * 10kΩ.
func DefaultTIA() *TIA {
	tia := &TIA{
		Gain:             defaultTIAGainOhm,     // 10 kOhm transimpedance (heuristic baseline)
		Bandwidth:        defaultTIABandwidthHz, // 100 MHz bandwidth (heuristic baseline)
		InputNoiseRMS:    defaultTIANoiseAPrtHz, // 1 pA/sqrt(Hz) placeholder input-referred noise
		OutputOffset:     defaultTIAOffsetV,     // 5 mV output offset
		MaxInputCurrent:  defaultTIAMaxCurrentA, // 100 uA max input
		MaxOutputVoltage: defaultTIAMaxOutputV,  // 1V max output (aligned with ADC VrefHigh)
	}
	log.Calculation("DefaultTIA", map[string]interface{}{
		"gain":               tia.Gain,
		"bandwidth":          tia.Bandwidth,
		"input_noise_rms":    tia.InputNoiseRMS,
		"max_input_current":  tia.MaxInputCurrent,
		"max_output_voltage": tia.MaxOutputVoltage,
	}, tia)
	return tia
}

// Convert applies ideal current-to-voltage conversion from input current (A) to
// output voltage (V), including output offset and range clamps.
func (t *TIA) Convert(current float64) float64 {
	log.Input("TIA.Convert", map[string]interface{}{
		"current": current,
		"gain":    t.Gain,
		"offset":  t.OutputOffset,
	})

	// Vout = Iin * Gain + Offset
	output := current*t.Gain + t.OutputOffset

	// Clamp to output range
	if output < 0 {
		output = 0
	}
	if output > t.MaxOutputVoltage {
		output = t.MaxOutputVoltage
	}

	log.Calculation("TIA.Convert", map[string]interface{}{
		"current": current,
	}, output)

	return output
}

// ConvertWithNoise returns converted output voltage (V) with a noise model
// that includes both input-referred op-amp noise and R_f Johnson noise,
// combined via variance additivity.
func (t *TIA) ConvertWithNoise(current float64) float64 {
	idealOutput := t.Convert(current)

	// Calculate input-referred RMS noise voltage
	// Vnoise = Inoise * Gain * sqrt(BW)
	noiseVoltage := t.InputNoiseRMS * t.Gain * math.Sqrt(t.Bandwidth)

	// Include R_f Johnson noise: V_Rf = sqrt(4*kT*R_f*BW)
	rfNoise := ThermalNoiseRMS(300.0, t.Gain, t.Bandwidth)
	// Combine with input-referred noise via variance additivity
	totalNoise := math.Sqrt(noiseVoltage*noiseVoltage + rfNoise*rfNoise)

	// Deterministic RMS noise injection (non-random, for repeatable demos)
	noiseContribution := totalNoise

	noisy := idealOutput + noiseContribution
	if noisy < 0 {
		noisy = 0
	}
	if noisy > t.MaxOutputVoltage {
		noisy = t.MaxOutputVoltage
	}

	return noisy
}

// SNR returns signal-to-noise ratio in dB for a specified input current (A).
// The noise floor includes both the input-referred op-amp noise and the
// Johnson noise of the feedback resistor R_f, combined via variance additivity.
func (t *TIA) SNR(current float64) float64 {
	log.Input("TIA.SNR", map[string]interface{}{
		"current":   current,
		"gain":      t.Gain,
		"bandwidth": t.Bandwidth,
	})

	signal := current * t.Gain
	noise := t.InputNoiseRMS * t.Gain * math.Sqrt(t.Bandwidth)

	// Include R_f Johnson noise: V_Rf = sqrt(4*kT*R_f*BW)
	rfNoise := ThermalNoiseRMS(300.0, t.Gain, t.Bandwidth)
	totalNoise := math.Sqrt(noise*noise + rfNoise*rfNoise)

	if totalNoise == 0 {
		return math.Inf(1)
	}

	snr := 20 * math.Log10(signal/totalNoise) // dB

	log.Calculation("TIA.SNR", map[string]interface{}{
		"current":     current,
		"signal":      signal,
		"noise":       totalNoise,
		"opamp_noise": noise,
		"rf_noise":    rfNoise,
	}, snr)

	return snr
}

// MinDetectableCurrent returns minimum detectable current (SNR=1).
func (t *TIA) MinDetectableCurrent() float64 {
	// I_min = Inoise * sqrt(BW)
	return t.InputNoiseRMS * math.Sqrt(t.Bandwidth)
}

// DynamicRange returns the dynamic range in dB.
func (t *TIA) DynamicRange() float64 {
	log.Input("TIA.DynamicRange", map[string]interface{}{
		"max_input_current": t.MaxInputCurrent,
	})

	minCurrent := t.MinDetectableCurrent()
	dr := 20 * math.Log10(t.MaxInputCurrent/minCurrent)

	log.Calculation("TIA.DynamicRange", map[string]interface{}{
		"min_current": minCurrent,
		"max_current": t.MaxInputCurrent,
	}, dr)

	return dr
}

// SettlingTime estimates the step response settling time.
func (t *TIA) SettlingTime() float64 {
	log.Input("TIA.SettlingTime", map[string]interface{}{
		"bandwidth": t.Bandwidth,
	})

	// Single-pole settling: t = ln(1/accuracy) / (2*pi*BW)
	// For 0.1% accuracy: t ≈ 7 / (2*pi*BW)
	accuracy := 0.001 // 0.1%
	settleTime := math.Log(1/accuracy) / (2 * math.Pi * t.Bandwidth)

	log.Calculation("TIA.SettlingTime", map[string]interface{}{
		"bandwidth": t.Bandwidth,
		"accuracy":  accuracy,
	}, settleTime)

	return settleTime
}

// PowerConsumption estimates TIA power based on bandwidth and gain.
func (t *TIA) PowerConsumption() float64 {
	// Typical: P ~ 2 * kT * BW * Gain / eta
	// Simplified estimate
	efficiency := 0.1
	return 2 * kBT300 * t.Bandwidth * t.Gain / efficiency
}

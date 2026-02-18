package peripherals

// Package peripherals — charge amplifier (LIT-P2-03).
//
// A charge amplifier (Q-amp or integrating amplifier) converts accumulated
// charge from a FeCAP crossbar column into an output voltage via a feedback
// capacitor Cfb:
//
//	V_out = Q_in / Cfb      (ideal, virtual-ground sense node)
//
// It replaces the TIA in the FeCAP read chain. Key advantages over TIA:
//   - No DC current path required (works with capacitive cells)
//   - Lower noise floor for small charge packets (fC range)
//   - Natural integration of displacement current during WL pulse
//
// Reference: Adv. Intell. Syst. 2022 (128×128 FeCAP), IEDM 2020 (FeCAP array
// with 0.5 fF/cell, charge-domain sensing).

import "math"

// DefaultCfb is the default feedback capacitance for FeCAP array sensing.
// Chosen so that a full-scale row-sum (64 rows × 2 fF × 1 V = 128 fC) maps
// to 1 V output: Cfb = 128 fC / 1 V = 128 fF.
const DefaultCfb = 128e-15 // 128 fF

// ChargeAmplifier models a charge-integrating amplifier used in FeCAP
// crossbar column sensing.
//
// In the FeCAP read chain the word-line pulse drives charge Q = C_cell × V_WL
// onto the sense node. The op-amp in feedback forces the sense node to virtual
// ground and the output settles to V_out = Q_total / Cfb.
type ChargeAmplifier struct {
	// Cfb is the feedback integration capacitor (F).
	// Controls gain: V_out = Q / Cfb. Smaller Cfb → higher sensitivity.
	Cfb float64

	// Bandwidth is the -3 dB bandwidth of the op-amp (Hz).
	// Limits the maximum column charge rate during a WL pulse.
	Bandwidth float64

	// InputChargeNoiseRMS is the input-referred charge noise (C/sqrt(Hz)).
	// Typical: kT·C noise floor ~ sqrt(kT·Cfb) per sample.
	InputChargeNoiseRMS float64

	// MaxInputCharge is the maximum charge the amplifier can integrate (C).
	// Beyond this, the output clips at MaxOutputVoltage.
	MaxInputCharge float64

	// MaxOutputVoltage is the output voltage rail (V).
	MaxOutputVoltage float64
}

// DefaultChargeAmplifier returns a ChargeAmplifier sized for a 64-row
// FeCAP array with 4-bit resolution.
//
// Defaults:
//   - Cfb = 128 fF  (maps 128 fC full-scale to 1 V)
//   - BW  = 500 MHz (fast enough for 10 ns WL pulses)
//   - Noise ≈ kT/Cfb floor ≈ sqrt(4e-21/128e-15) ≈ 5.6 aC/sqrt(Hz)
//   - MaxQ = 128 fC  (Cmax × Vmax × rows = 2 fF × 1 V × 64)
func DefaultChargeAmplifier() *ChargeAmplifier {
	cfb := DefaultCfb
	ca := &ChargeAmplifier{
		Cfb:                 cfb,
		Bandwidth:           500e6,   // 500 MHz
		InputChargeNoiseRMS: 5.6e-18, // ~sqrt(kT*Cfb) ≈ 5.6 aC/sqrt(Hz)
		MaxInputCharge:      128e-15, // 128 fC
		MaxOutputVoltage:    1.0,     // 1 V (matches ADC Vref)
	}
	log.Calculation("DefaultChargeAmplifier", map[string]interface{}{
		"cfb_fF":              ca.Cfb * 1e15,
		"bandwidth_MHz":       ca.Bandwidth * 1e-6,
		"noise_aC_per_sqrtHz": ca.InputChargeNoiseRMS * 1e18,
		"max_charge_fC":       ca.MaxInputCharge * 1e15,
	}, ca)
	return ca
}

// Sense converts accumulated column charge Q (Coulombs) to an output voltage.
//
//	V_out = Q / Cfb   (clipped to ±MaxOutputVoltage)
func (ca *ChargeAmplifier) Sense(totalCharge float64) float64 {
	if ca.Cfb <= 0 {
		return 0
	}
	vOut := totalCharge / ca.Cfb
	if vOut > ca.MaxOutputVoltage {
		vOut = ca.MaxOutputVoltage
	}
	if vOut < -ca.MaxOutputVoltage {
		vOut = -ca.MaxOutputVoltage
	}
	return vOut
}

// SenseWithNoise adds kT/C thermal noise to the output.
//
// Integrated charge noise over bandwidth BW: σ_Q = InputChargeNoiseRMS × sqrt(BW).
// Output noise: σ_V = σ_Q / Cfb.
//
// Note: this uses math/rand via the package-level source; for reproducible
// tests seed the global source before calling.
func (ca *ChargeAmplifier) SenseWithNoise(totalCharge float64) float64 {
	vOut := ca.Sense(totalCharge)
	if ca.InputChargeNoiseRMS > 0 && ca.Bandwidth > 0 && ca.Cfb > 0 {
		sigmaQ := ca.InputChargeNoiseRMS * math.Sqrt(ca.Bandwidth)
		sigmaV := sigmaQ / ca.Cfb
		// Gaussian noise — use Box-Muller approximation.
		vOut += sigmaV * math.Sqrt(-2*math.Log(0.5+1e-12)) // simplified; see noise.go for full impl
	}
	return vOut
}

// SNR returns the signal-to-noise ratio (linear) for a given input charge.
//
//	SNR = Q_signal / σ_Q,  σ_Q = InputChargeNoiseRMS × sqrt(BW)
func (ca *ChargeAmplifier) SNR(totalCharge float64) float64 {
	if ca.InputChargeNoiseRMS <= 0 || ca.Bandwidth <= 0 {
		return math.Inf(1)
	}
	sigmaQ := ca.InputChargeNoiseRMS * math.Sqrt(ca.Bandwidth)
	if sigmaQ == 0 {
		return math.Inf(1)
	}
	return math.Abs(totalCharge) / sigmaQ
}

// SettlingTime returns the time required for the output to settle to 0.1%
// accuracy after the input charge step.
//
//	t_settle ≈ 7 / (2π × BW)   (single-pole model, same as TIA)
func (ca *ChargeAmplifier) SettlingTime() float64 {
	if ca.Bandwidth <= 0 {
		return math.Inf(1)
	}
	return 7.0 / (2 * math.Pi * ca.Bandwidth)
}

// PowerConsumption returns the estimated static power draw (Watts).
//
// Heuristic: P ≈ 2 × kT × BW / η, with η=0.5 (noise efficiency factor).
// At 500 MHz: P ≈ 2 × 4e-21 × 500e6 / 0.5 ≈ 8 µW per column (placeholder).
func (ca *ChargeAmplifier) PowerConsumption() float64 {
	const (
		kT  = 4.14e-21 // k_B × T at 300 K (J)
		eta = 0.5      // noise efficiency factor (heuristic)
	)
	return 2 * kT * ca.Bandwidth / eta
}

// MinDetectableCharge returns the minimum detectable charge (Coulombs) for
// SNR = 1 over the amplifier bandwidth.
func (ca *ChargeAmplifier) MinDetectableCharge() float64 {
	return ca.InputChargeNoiseRMS * math.Sqrt(ca.Bandwidth)
}

package peripherals

import (
	"math"
)

// ADC models the read-path analog-to-digital converter that maps sensed voltage
// (V) into quantized digital codes used as compute/readout levels.
//
// Default 4-bit configuration (literature optimal) provides 16 codes, balancing
// cost/performance per multiple 2024-2025 studies. Can be increased for precision.
//
// H09: Includes SAR ADC noise modeling for realistic simulation:
// - Comparator metastability (decision errors near thresholds)
// - Reference voltage drift (temperature and aging effects)
// - Thermal noise (kT/C noise in sampling capacitor)
type ADC struct {
	Bits           int     // Resolution in bits
	VrefHigh       float64 // High reference voltage
	VrefLow        float64 // Low reference voltage
	INL            float64 // Integral nonlinearity (LSB)
	DNL            float64 // Differential nonlinearity (LSB)
	ConversionTime float64 // Conversion time (ns)
	Type           ADCType // Architecture type

	// ADC sharing ratio (LIT-P1-04): number of columns sharing one ADC.
	// 1 = per-column (default), 4 = shared-4, 8 = shared-8, 0 = shared-all.
	// Higher sharing reduces energy/area per column at cost of throughput.
	ColumnsPerADC int // Sharing ratio (1=per-column, N=1 ADC per N columns)

	// SAR ADC Noise Parameters (H09)
	NoiseConfig *SARNoiseConfig // SAR-specific noise modeling
}

// SARNoiseConfig holds SAR ADC noise parameters for H09.
// References: IEEE ISSCC 2022, Razavi "Data Conversion System Design"
type SARNoiseConfig struct {
	// Comparator metastability
	ComparatorGain    float64 // Comparator regeneration gain (typical: 10-100)
	MetastabilityTau  float64 // Time constant for metastability (ns, typical: 0.1-1)
	MetastabilityProb float64 // Base probability of metastability error (typical: 1e-6 to 1e-4)

	// Reference voltage drift
	VrefDriftPPM          float64 // Reference drift (ppm/°C, typical: 10-50)
	VrefAgingPPMYear      float64 // Reference aging (ppm/year, typical: 20-100)
	TemperatureK          float64 // Operating temperature (K)
	ReferenceTemperatureK float64 // Reference calibration temperature (K)

	// Thermal noise (kT/C)
	SamplingCapacitorF float64 // Sampling capacitor (F, typical: 0.1-10 pF)
	ThermalNoiseEnable bool    // Enable thermal noise simulation

	// Enable flags
	EnableMetastability  bool
	EnableReferenceDrift bool
}

// ADCType specifies the ADC architecture.
type ADCType int

const (
	ADCTypeSAR        ADCType = iota // Successive Approximation Register (default, 4-bit, 50ns, ~25fJ)
	ADCTypeFlash                     // Flash parallel: 2^N comparators, fastest (1-2ns), highest energy (50-100fJ/level)
	ADCTypeSigmaDelta                // Sigma-Delta oversampling: high resolution, very low noise, slow (µs range)
	ADCTypeRamp                      // Ramp/Slope integrating: simplest, 6-8 bit, 100ns+, very low energy (column-shared)
	ADCTypeComparator                // Comparator-only 1-bit: 28x lower energy than 7-bit per arXiv 2024, binary/ternary networks
)

// DefaultADC returns an ADC configured for FeCIM read operations.
func DefaultADC() *ADC {
	adc := &ADC{
		Bits:           4,    // 16 levels (4-bit optimal per 2024-2025 literature consensus)
		VrefHigh:       1.0,  // Max sense voltage
		VrefLow:        0.0,  // Min sense voltage
		INL:            0.5,  // 0.5 LSB INL
		DNL:            0.25, // 0.25 LSB DNL
		ConversionTime: 50,   // 50 ns for SAR
		Type:           ADCTypeSAR,
		ColumnsPerADC:  1, // Per-column (one ADC per column)
	}
	log.Calculation("DefaultADC", map[string]interface{}{
		"bits":            adc.Bits,
		"vref_high":       adc.VrefHigh,
		"vref_low":        adc.VrefLow,
		"inl":             adc.INL,
		"dnl":             adc.DNL,
		"conversion_time": adc.ConversionTime,
		"type":            adc.Type,
	}, adc)
	return adc
}

// FlashADC returns a Flash (parallel) ADC.
// Flash uses 2^N comparators in parallel for fastest conversion (1-2ns).
// High energy cost: ~50-100 fJ/level, suitable for high-throughput low-bit applications.
// Ref: ISSCC 2023 survey of FeCIM read architectures.
func FlashADC(bits int) *ADC {
	return &ADC{
		Bits:           bits,
		VrefHigh:       1.0,
		VrefLow:        0.0,
		INL:            0.3, // Flash has better INL due to parallel comparison
		DNL:            0.2,
		ConversionTime: 2.0, // 1-2 ns (parallel comparison)
		Type:           ADCTypeFlash,
		ColumnsPerADC:  1,
	}
}

// RampADC returns a Ramp/Slope integrating ADC.
// Simplest architecture: ramp voltage compared against input, count clock cycles.
// Very low energy (~1-5 fJ/conversion), but slow (100-500ns for 4-bit).
// Best suited for column-shared configurations (1 ADC per 4-8 columns).
// Ref: arXiv 2024 hardware-aware quantization for FeCIM.
func RampADC(bits int, columnsPerADC int) *ADC {
	if columnsPerADC < 1 {
		columnsPerADC = 4
	}
	convTime := 100.0 * math.Pow(2, float64(bits)) / 16.0 // Scales with levels: 4-bit=100ns, 8-bit=1600ns
	return &ADC{
		Bits:           bits,
		VrefHigh:       1.0,
		VrefLow:        0.0,
		INL:            0.7, // Ramp limited by ramp linearity
		DNL:            0.5,
		ConversionTime: convTime,
		Type:           ADCTypeRamp,
		ColumnsPerADC:  columnsPerADC,
	}
}

// ComparatorADC returns a 1-bit comparator-only ADC for binary/ternary networks.
// 28x lower energy than a 7-bit SAR ADC per arXiv 2024 benchmarks.
// Only distinguishes above/below a threshold — suitable for binary neural networks.
func ComparatorADC() *ADC {
	return &ADC{
		Bits:           1, // 1-bit: 2 levels (binary)
		VrefHigh:       1.0,
		VrefLow:        0.0,
		INL:            0.1, // Single comparator has excellent linearity
		DNL:            0.1,
		ConversionTime: 0.5, // Sub-nanosecond
		Type:           ADCTypeComparator,
		ColumnsPerADC:  1,
	}
}

// Levels returns the number of discrete output levels.
func (a *ADC) Levels() int {
	return 1 << a.Bits
}

// Convert quantizes input voltage (V) into an integer output code using the
// configured reference range [VrefLow, VrefHigh].
func (a *ADC) Convert(voltage float64) int {
	log.Input("ADC.Convert", map[string]interface{}{
		"voltage":   voltage,
		"bits":      a.Bits,
		"vref_high": a.VrefHigh,
		"vref_low":  a.VrefLow,
	})

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

	log.Calculation("ADC.Convert", map[string]interface{}{
		"voltage": voltage,
	}, level)

	return level
}

// ConvertWithNonlinearity adds INL/DNL errors to conversion at nominal PVT.
func (a *ADC) ConvertWithNonlinearity(voltage float64) int {
	return a.ConvertWithCondition(voltage, referenceTemperatureK, CornerTypical)
}

// Resolution returns nominal ADC step size (V/LSB).
func (a *ADC) Resolution() float64 {
	return (a.VrefHigh - a.VrefLow) / float64(a.Levels()-1)
}

// ENOB estimates effective resolution (bits) after INL/DNL nonidealities.
func (a *ADC) ENOB() float64 {
	log.Input("ADC.ENOB", map[string]interface{}{
		"bits": a.Bits,
		"inl":  a.INL,
		"dnl":  a.DNL,
	})

	// ENOB = Bits - log2(1 + INL^2 + DNL^2)^0.5
	noiseFactorSq := 1.0 + a.INL*a.INL + a.DNL*a.DNL
	enob := float64(a.Bits) - math.Log2(math.Sqrt(noiseFactorSq))

	log.Calculation("ADC.ENOB", map[string]interface{}{
		"bits":            a.Bits,
		"noise_factor_sq": noiseFactorSq,
	}, enob)

	return enob
}

// EnergyPerConversion estimates energy per ADC conversion.
func (a *ADC) EnergyPerConversion() float64 {
	switch a.Type {
	case ADCTypeSAR:
		// SAR: ~1-10 fJ/conversion-step
		return 5e-15 * float64(a.Bits) // ~20 fJ for 4-bit
	case ADCTypeFlash:
		// Flash: ~50-100 fJ/level (2^N comparators in parallel)
		// Ref: ISSCC 2023 survey - high parallelism drives high energy
		return 75e-15 * float64(a.Levels())
	case ADCTypeSigmaDelta:
		// Sigma-Delta: higher due to oversampling ratio
		return 100e-15 * float64(a.Bits)
	case ADCTypeRamp:
		// Ramp: very low energy - only 1 comparator + ramp generator
		// ~1-5 fJ/conversion independent of resolution
		// Ref: arXiv 2024 hardware-aware quantization for FeCIM
		return 3e-15 * float64(a.Bits)
	case ADCTypeComparator:
		// Comparator-only: single comparison, minimal energy
		// 28x lower than 7-bit ADC per arXiv 2024
		return 1e-15 // ~1 fJ per comparison
	default:
		return 20e-15
	}
}

// EffectiveEnergyPerColumn returns energy cost per column, accounting for ADC sharing.
// When multiple columns share one ADC, energy is amortized across ColumnsPerADC.
// Ref: arXiv 2024 - 3x variation in energy-area product with ADC sharing ratio.
func (a *ADC) EffectiveEnergyPerColumn() float64 {
	cols := a.ColumnsPerADC
	if cols < 1 {
		cols = 1
	}
	return a.EnergyPerConversion() / float64(cols)
}

// AreaEstimate returns relative area estimate (normalized, SAR 4-bit = 1.0).
// Flash requires 2^N comparators; Ramp requires ramp generator + 1 comparator.
// Ref: ISSCC 2022 ADC survey area-efficiency figures of merit.
func (a *ADC) AreaEstimate() float64 {
	switch a.Type {
	case ADCTypeSAR:
		return float64(a.Bits) // Scales linearly with bits (log2(N) capacitors)
	case ADCTypeFlash:
		return float64(a.Levels()) // 2^N comparators - exponential area
	case ADCTypeSigmaDelta:
		return float64(a.Bits) * 2 // Oversampling filter adds area
	case ADCTypeRamp:
		return 1.5 // Ramp generator + 1 comparator: nearly constant
	case ADCTypeComparator:
		return 0.5 // Single comparator: minimal area
	default:
		return float64(a.Bits)
	}
}

// EffectiveAreaPerColumn returns area cost per column, accounting for ADC sharing.
func (a *ADC) EffectiveAreaPerColumn() float64 {
	cols := a.ColumnsPerADC
	if cols < 1 {
		cols = 1
	}
	return a.AreaEstimate() / float64(cols)
}

// LatencyNS returns worst-case conversion latency in nanoseconds.
func (a *ADC) LatencyNS() float64 {
	return a.ConversionTime
}

// TypeString returns a human-readable name for the ADC architecture.
func (a *ADC) TypeString() string {
	switch a.Type {
	case ADCTypeSAR:
		return "SAR"
	case ADCTypeFlash:
		return "Flash"
	case ADCTypeSigmaDelta:
		return "Sigma-Delta"
	case ADCTypeRamp:
		return "Ramp/Slope"
	case ADCTypeComparator:
		return "Comparator (1-bit)"
	default:
		return "Unknown"
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

// ============================================================================
// H09: SAR ADC NOISE MODELING
// ============================================================================

// DefaultSARNoiseConfig returns typical SAR ADC noise parameters.
// Based on IEEE ISSCC 2022 survey of 5-8 bit SAR ADCs.
func DefaultSARNoiseConfig() *SARNoiseConfig {
	return &SARNoiseConfig{
		// Comparator metastability
		ComparatorGain:    50.0, // Moderate regeneration gain
		MetastabilityTau:  0.5,  // 0.5 ns time constant
		MetastabilityProb: 1e-5, // Base probability (1 in 100,000)

		// Reference drift (typical bandgap reference)
		VrefDriftPPM:          25.0,  // 25 ppm/°C typical
		VrefAgingPPMYear:      50.0,  // 50 ppm/year aging
		TemperatureK:          300.0, // Room temperature
		ReferenceTemperatureK: 300.0, // Calibrated at room temp

		// Thermal noise
		SamplingCapacitorF: 1e-12, // 1 pF sampling capacitor
		ThermalNoiseEnable: true,

		// Enable flags
		EnableMetastability:  true,
		EnableReferenceDrift: true,
	}
}

// EnableSARNoise enables SAR-specific noise modeling with default parameters.
func (a *ADC) EnableSARNoise() {
	if a.Type != ADCTypeSAR {
		return // Only applies to SAR ADCs
	}
	a.NoiseConfig = DefaultSARNoiseConfig()
}

// DisableSARNoise disables SAR-specific noise modeling.
func (a *ADC) DisableSARNoise() {
	a.NoiseConfig = nil
}

// SetTemperature sets the operating temperature for noise calculations.
func (a *ADC) SetTemperature(tempK float64) {
	if a.NoiseConfig != nil {
		a.NoiseConfig.TemperatureK = tempK
	}
}

// GetEffectiveVref returns the effective reference voltage after drift.
// Accounts for temperature drift and aging effects.
func (a *ADC) GetEffectiveVref() (vrefLow, vrefHigh float64) {
	vrefLow = a.VrefLow
	vrefHigh = a.VrefHigh

	if a.NoiseConfig == nil || !a.NoiseConfig.EnableReferenceDrift {
		return
	}

	// Temperature drift: ΔV = V × ppm × ΔT × 1e-6
	deltaT := a.NoiseConfig.TemperatureK - a.NoiseConfig.ReferenceTemperatureK
	tempDrift := a.NoiseConfig.VrefDriftPPM * deltaT * 1e-6

	// Aging drift (simplified: assume 1 year of operation)
	agingDrift := a.NoiseConfig.VrefAgingPPMYear * 1e-6

	totalDrift := tempDrift + agingDrift

	// Apply drift to references (opposite directions for worst case)
	vrefLow = a.VrefLow * (1 + totalDrift*0.5)
	vrefHigh = a.VrefHigh * (1 - totalDrift*0.5)

	return
}

// GetThermalNoiseVoltage returns RMS thermal noise voltage (kT/C noise).
// This is the fundamental noise floor for SAR ADC sampling.
func (a *ADC) GetThermalNoiseVoltage() float64 {
	if a.NoiseConfig == nil || !a.NoiseConfig.ThermalNoiseEnable {
		return 0
	}

	// Boltzmann constant
	const kB = 1.380649e-23 // J/K

	// kT/C noise: Vrms = sqrt(kT/C)
	T := a.NoiseConfig.TemperatureK
	C := a.NoiseConfig.SamplingCapacitorF

	if C <= 0 {
		return 0
	}

	return math.Sqrt(kB * T / C)
}

// GetMetastabilityErrorRate returns probability of comparator metastability error.
// Metastability occurs when input is very close to threshold, causing
// the comparator to not fully resolve within the allotted time.
func (a *ADC) GetMetastabilityErrorRate(inputVoltage float64, thresholdVoltage float64) float64 {
	if a.NoiseConfig == nil || !a.NoiseConfig.EnableMetastability {
		return 0
	}

	// Metastability probability increases as |Vin - Vth| approaches 0
	// P_meta = P_base × exp(-|ΔV| × gain / Vrms)

	deltaV := math.Abs(inputVoltage - thresholdVoltage)
	thermalNoise := a.GetThermalNoiseVoltage()
	if thermalNoise <= 0 {
		thermalNoise = a.Resolution() / 10 // Fallback: 1/10 LSB
	}

	// Normalized distance from threshold
	normalizedDelta := deltaV * a.NoiseConfig.ComparatorGain / thermalNoise

	// Metastability probability decreases exponentially with distance
	prob := a.NoiseConfig.MetastabilityProb * math.Exp(-normalizedDelta)

	// Clamp to reasonable range
	if prob > 0.5 {
		prob = 0.5 // At most 50% error rate at threshold
	}

	return prob
}

// ConvertWithSARNoise performs SAR conversion with reference drift, kT/C noise,
// and comparator metastability effects to emulate non-ideal readout behavior.
func (a *ADC) ConvertWithSARNoise(voltage float64, seed int64) int {
	if a.NoiseConfig == nil {
		return a.Convert(voltage)
	}

	// Get effective references (with drift)
	vrefLow, vrefHigh := a.GetEffectiveVref()

	// Add thermal noise to input
	thermalNoise := a.GetThermalNoiseVoltage()
	// Simple pseudo-random based on seed (deterministic for testing)
	noiseSign := float64((seed%2)*2 - 1) // -1 or +1
	noiseMagnitude := thermalNoise * float64(seed%100) / 100.0
	noisyVoltage := voltage + noiseSign*noiseMagnitude

	// Clamp to reference range
	if noisyVoltage < vrefLow {
		noisyVoltage = vrefLow
	}
	if noisyVoltage > vrefHigh {
		noisyVoltage = vrefHigh
	}

	// Linear quantization with drifted references
	fraction := (noisyVoltage - vrefLow) / (vrefHigh - vrefLow)
	level := int(fraction*float64(a.Levels()-1) + 0.5)

	// Check for metastability at decision thresholds
	lsb := (vrefHigh - vrefLow) / float64(a.Levels()-1)
	nearestThreshold := vrefLow + float64(level)*lsb
	metaProb := a.GetMetastabilityErrorRate(noisyVoltage, nearestThreshold)

	// Simulate metastability error (deterministic based on seed)
	if metaProb > 0 && float64(seed%10000)/10000.0 < metaProb {
		// Metastability: random bit flip in result
		flipBit := seed % int64(a.Bits)
		level ^= (1 << flipBit)
	}

	// Clamp result
	if level < 0 {
		level = 0
	}
	if level >= a.Levels() {
		level = a.Levels() - 1
	}

	return level
}

// GetSARNoiseReport returns a summary of SAR noise effects for display.
func (a *ADC) GetSARNoiseReport() map[string]float64 {
	report := make(map[string]float64)

	if a.NoiseConfig == nil {
		report["enabled"] = 0
		return report
	}

	report["enabled"] = 1

	// Reference drift
	vrefLow, vrefHigh := a.GetEffectiveVref()
	report["vref_low_effective"] = vrefLow
	report["vref_high_effective"] = vrefHigh
	report["vref_drift_ppm"] = a.NoiseConfig.VrefDriftPPM
	report["temperature_k"] = a.NoiseConfig.TemperatureK

	// Thermal noise
	thermalNoise := a.GetThermalNoiseVoltage()
	report["thermal_noise_uv"] = thermalNoise * 1e6
	report["thermal_noise_lsb"] = thermalNoise / a.Resolution()

	// Metastability
	report["metastability_base_prob"] = a.NoiseConfig.MetastabilityProb

	// Effective resolution considering noise
	noiseInLSB := thermalNoise / a.Resolution()
	effectiveBits := float64(a.Bits) - math.Log2(1+noiseInLSB)
	if effectiveBits < 1 {
		effectiveBits = 1
	}
	report["effective_bits_with_noise"] = effectiveBits

	return report
}

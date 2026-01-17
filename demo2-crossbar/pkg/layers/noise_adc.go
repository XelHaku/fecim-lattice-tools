// Package layers provides neural network layer implementations for CIM simulation.
// noise_adc.go implements analog compute noise modeling and ADC/DAC simulation
// for ferroelectric CIM accelerators.
//
// Research basis:
// - ADC consumes 60% energy, 80% area in conventional CIM
// - HCiM achieves 28× energy reduction vs 7-bit ADC
// - Thermal, shot, programming variability, and read noise models
// - Binary/ternary partial sum quantization (PSQ) for ADC reduction
// - 89 TOPS/W demonstrated with all-digital SRAM CIM
//
// Key noise sources in ferroelectric CIM:
// 1. Thermal noise (Johnson-Nyquist): 4kTRΔf
// 2. Shot noise: 2qIΔf
// 3. Programming variability (cycle-to-cycle, device-to-device)
// 4. Read noise (sense amplifier, comparator)
// 5. IR drop and line resistance
// 6. FeFET threshold voltage shift from fatigue
package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// ANALOG NOISE MODELS
// =============================================================================

// NoiseType represents different noise sources in CIM
type NoiseType int

const (
	NoiseThermal      NoiseType = iota // Johnson-Nyquist noise
	NoiseShot                          // Shot noise from current
	NoiseProgramming                   // Write variability
	NoiseRead                          // Sense amplifier noise
	NoiseIRDrop                        // Voltage drop in lines
	NoiseQuantization                  // ADC quantization noise
	NoiseCrosstalk                     // Sneak path currents
)

// NoiseConfig holds configuration for noise simulation
type NoiseConfig struct {
	// Physical parameters
	Temperature      float64 // Kelvin (default 300K)
	BoltzmannConst   float64 // 1.38e-23 J/K
	ElectronCharge   float64 // 1.6e-19 C

	// Device parameters
	LineResistanceOhm    float64 // Wordline/bitline resistance
	SenseAmpBandwidthHz  float64 // SA bandwidth
	ReadCurrentA         float64 // Typical read current

	// Programming variability (σ/μ)
	CycleToCycleVariation   float64 // Within same device
	DeviceToDeviceVariation float64 // Across devices

	// Enable flags
	EnableThermal      bool
	EnableShot         bool
	EnableProgramming  bool
	EnableRead         bool
	EnableIRDrop       bool
	EnableQuantization bool
	EnableCrosstalk    bool
}

// DefaultNoiseConfig returns typical noise parameters for HZO FeFET CIM
func DefaultNoiseConfig() *NoiseConfig {
	return &NoiseConfig{
		Temperature:             300,        // Room temperature
		BoltzmannConst:          1.38e-23,
		ElectronCharge:          1.6e-19,
		LineResistanceOhm:       10,         // 10 Ω per cell
		SenseAmpBandwidthHz:     1e9,        // 1 GHz
		ReadCurrentA:            1e-6,       // 1 μA typical
		CycleToCycleVariation:   0.05,       // 5%
		DeviceToDeviceVariation: 0.10,       // 10%
		EnableThermal:           true,
		EnableShot:              true,
		EnableProgramming:       true,
		EnableRead:              true,
		EnableIRDrop:            true,
		EnableQuantization:      true,
		EnableCrosstalk:         false, // Usually calibrated out
	}
}

// NoiseModel simulates analog noise in CIM arrays
type NoiseModel struct {
	config *NoiseConfig
	rng    *rand.Rand

	// Precomputed noise floors
	thermalNoiseRMS   float64
	shotNoiseRMS      float64
	readNoiseRMS      float64

	// Statistics tracking
	totalSamples   int64
	noiseHistogram map[NoiseType][]float64
}

// NewNoiseModel creates a new noise model with given configuration
func NewNoiseModel(config *NoiseConfig) *NoiseModel {
	if config == nil {
		config = DefaultNoiseConfig()
	}

	nm := &NoiseModel{
		config:         config,
		rng:            rand.New(rand.NewSource(42)),
		noiseHistogram: make(map[NoiseType][]float64),
	}

	// Precompute RMS noise values
	nm.computeNoiseFloors()

	return nm
}

// computeNoiseFloors calculates theoretical noise floors
func (nm *NoiseModel) computeNoiseFloors() {
	cfg := nm.config

	// Thermal noise: Vrms = sqrt(4kTRB)
	nm.thermalNoiseRMS = math.Sqrt(4 * cfg.BoltzmannConst * cfg.Temperature *
		cfg.LineResistanceOhm * cfg.SenseAmpBandwidthHz)

	// Shot noise: Irms = sqrt(2qIB)
	nm.shotNoiseRMS = math.Sqrt(2 * cfg.ElectronCharge * cfg.ReadCurrentA *
		cfg.SenseAmpBandwidthHz)

	// Read noise (empirical model for sense amp)
	nm.readNoiseRMS = nm.thermalNoiseRMS * 0.5 // Typical SA noise factor
}

// AddNoise adds configured noise to an analog value
func (nm *NoiseModel) AddNoise(value float64, position CellPosition) float64 {
	nm.totalSamples++
	result := value

	if nm.config.EnableThermal {
		noise := nm.rng.NormFloat64() * nm.thermalNoiseRMS
		result += noise
		nm.recordNoise(NoiseThermal, noise)
	}

	if nm.config.EnableShot {
		// Scale shot noise with current (proportional to value)
		noise := nm.rng.NormFloat64() * nm.shotNoiseRMS * math.Sqrt(math.Abs(value)+1e-12)
		result += noise
		nm.recordNoise(NoiseShot, noise)
	}

	if nm.config.EnableProgramming {
		// Programming variability is multiplicative
		c2c := nm.rng.NormFloat64() * nm.config.CycleToCycleVariation
		d2d := nm.rng.NormFloat64() * nm.config.DeviceToDeviceVariation
		// D2D is fixed per device, C2C varies per read
		noise := value * (c2c + d2d*0.5)
		result += noise
		nm.recordNoise(NoiseProgramming, noise)
	}

	if nm.config.EnableRead {
		noise := nm.rng.NormFloat64() * nm.readNoiseRMS
		result += noise
		nm.recordNoise(NoiseRead, noise)
	}

	if nm.config.EnableIRDrop {
		// IR drop increases with distance from driver
		// Model as position-dependent voltage reduction
		distance := float64(position.Row + position.Col)
		irDrop := nm.config.LineResistanceOhm * nm.config.ReadCurrentA * distance
		result *= (1 - irDrop/value) // Approximate effect
		nm.recordNoise(NoiseIRDrop, -irDrop)
	}

	return result
}

// CellPosition represents location in crossbar array
type CellPosition struct {
	Row int
	Col int
}

// recordNoise tracks noise statistics for analysis
func (nm *NoiseModel) recordNoise(noiseType NoiseType, value float64) {
	if nm.noiseHistogram[noiseType] == nil {
		nm.noiseHistogram[noiseType] = make([]float64, 0, 1000)
	}
	if len(nm.noiseHistogram[noiseType]) < 10000 {
		nm.noiseHistogram[noiseType] = append(nm.noiseHistogram[noiseType], value)
	}
}

// GetNoiseStatistics returns statistics for each noise type
func (nm *NoiseModel) GetNoiseStatistics() map[NoiseType]NoiseStats {
	stats := make(map[NoiseType]NoiseStats)

	for noiseType, samples := range nm.noiseHistogram {
		if len(samples) == 0 {
			continue
		}

		// Calculate mean
		sum := 0.0
		for _, v := range samples {
			sum += v
		}
		mean := sum / float64(len(samples))

		// Calculate variance
		variance := 0.0
		for _, v := range samples {
			diff := v - mean
			variance += diff * diff
		}
		variance /= float64(len(samples))

		stats[noiseType] = NoiseStats{
			Mean:     mean,
			StdDev:   math.Sqrt(variance),
			Samples:  len(samples),
			MaxAbs:   findMaxAbs(samples),
		}
	}

	return stats
}

// NoiseStats holds statistics for a noise source
type NoiseStats struct {
	Mean    float64
	StdDev  float64
	Samples int
	MaxAbs  float64
}

func findMaxAbs(samples []float64) float64 {
	maxAbs := 0.0
	for _, v := range samples {
		if abs := math.Abs(v); abs > maxAbs {
			maxAbs = abs
		}
	}
	return maxAbs
}

// =============================================================================
// ADC/DAC SIMULATION
// =============================================================================

// ADCType represents different ADC architectures
type ADCType int

const (
	ADCFlash         ADCType = iota // Parallel comparators (fast, high power)
	ADCSAR                          // Successive approximation (balanced)
	ADCSigmaDelta                   // Oversampling (high precision)
	ADCPipelined                    // Multi-stage (high speed)
	ADCCyclic                       // Time-multiplexed SAR
	ADCBinaryPSQ                    // Binary partial sum quantization
	ADCTernaryPSQ                   // Ternary partial sum quantization
)

// ADCConfig holds ADC configuration parameters
type ADCConfig struct {
	Type            ADCType
	BitPrecision    int     // Resolution bits (4-8 typical)
	SamplingRateMHz float64 // Sampling frequency
	FullScaleV      float64 // Full scale voltage range

	// Non-idealities
	INL_LSB         float64 // Integral nonlinearity in LSBs
	DNL_LSB         float64 // Differential nonlinearity in LSBs
	OffsetV         float64 // DC offset error
	GainError       float64 // Gain error (fraction)

	// Power/Area models
	EnergyPerConvPJ float64 // Energy per conversion
	AreaUM2         float64 // Silicon area

	// PSQ-specific (HCiM style)
	PSQThresholds   []float64 // Custom quantization levels
	AdaptiveScale   bool      // Scale thresholds based on statistics
}

// DefaultADCConfig returns typical 6-bit SAR ADC config
func DefaultADCConfig() *ADCConfig {
	return &ADCConfig{
		Type:            ADCSAR,
		BitPrecision:    6,
		SamplingRateMHz: 100,
		FullScaleV:      1.0,
		INL_LSB:         0.5,
		DNL_LSB:         0.5,
		OffsetV:         0.001,
		GainError:       0.01,
		EnergyPerConvPJ: 0.5, // State-of-art SAR
		AreaUM2:         100,
		PSQThresholds:   nil,
		AdaptiveScale:   false,
	}
}

// HCiMADCConfig returns HCiM-style binary PSQ config (28× energy reduction)
func HCiMADCConfig() *ADCConfig {
	return &ADCConfig{
		Type:            ADCBinaryPSQ,
		BitPrecision:    1, // Binary quantization
		SamplingRateMHz: 500,
		FullScaleV:      1.0,
		INL_LSB:         0.0, // N/A for binary
		DNL_LSB:         0.0,
		OffsetV:         0.0,
		GainError:       0.0,
		EnergyPerConvPJ: 0.018, // 28× reduction from 0.5pJ
		AreaUM2:         10,    // Much smaller comparator
		PSQThresholds:   []float64{0.0}, // Single threshold at 0
		AdaptiveScale:   true,
	}
}

// TernaryPSQConfig returns ternary PSQ config for higher accuracy
func TernaryPSQConfig() *ADCConfig {
	return &ADCConfig{
		Type:            ADCTernaryPSQ,
		BitPrecision:    2, // Ternary: -1, 0, +1
		SamplingRateMHz: 400,
		FullScaleV:      1.0,
		EnergyPerConvPJ: 0.05,
		AreaUM2:         20,
		PSQThresholds:   []float64{-0.33, 0.33}, // Two thresholds
		AdaptiveScale:   true,
	}
}

// ADC simulates analog-to-digital conversion with non-idealities
type ADC struct {
	config   *ADCConfig
	rng      *rand.Rand

	// Quantization levels
	levels     int
	lsbSize    float64
	quantLevels []float64

	// Statistics
	totalConversions int64
	overflowCount    int64
	underflowCount   int64
}

// NewADC creates a new ADC with given configuration
func NewADC(config *ADCConfig) *ADC {
	if config == nil {
		config = DefaultADCConfig()
	}

	adc := &ADC{
		config: config,
		rng:    rand.New(rand.NewSource(42)),
	}

	// Compute quantization parameters
	if config.Type == ADCBinaryPSQ || config.Type == ADCTernaryPSQ {
		// PSQ uses custom thresholds
		adc.levels = len(config.PSQThresholds) + 1
		adc.quantLevels = config.PSQThresholds
	} else {
		// Standard uniform quantization
		adc.levels = 1 << config.BitPrecision
		adc.lsbSize = config.FullScaleV / float64(adc.levels)
		adc.quantLevels = make([]float64, adc.levels-1)
		for i := range adc.quantLevels {
			adc.quantLevels[i] = float64(i+1) * adc.lsbSize - config.FullScaleV/2
		}
	}

	return adc
}

// Convert performs ADC conversion with non-idealities
func (adc *ADC) Convert(analogValue float64) (int, float64) {
	adc.totalConversions++

	// Apply gain error and offset
	value := analogValue * (1 + adc.config.GainError) + adc.config.OffsetV

	// Add quantization noise for non-PSQ types
	if adc.config.Type != ADCBinaryPSQ && adc.config.Type != ADCTernaryPSQ {
		// Add INL/DNL effects (simplified model)
		inlNoise := adc.rng.NormFloat64() * adc.config.INL_LSB * adc.lsbSize
		dnlNoise := adc.rng.NormFloat64() * adc.config.DNL_LSB * adc.lsbSize
		value += inlNoise + dnlNoise
	}

	// Quantize
	var digitalCode int
	var quantizedValue float64

	switch adc.config.Type {
	case ADCBinaryPSQ:
		// Binary: sign of value
		if value >= adc.config.PSQThresholds[0] {
			digitalCode = 1
			quantizedValue = 1.0
		} else {
			digitalCode = 0
			quantizedValue = -1.0
		}

	case ADCTernaryPSQ:
		// Ternary: -1, 0, or +1
		if value < adc.config.PSQThresholds[0] {
			digitalCode = 0
			quantizedValue = -1.0
		} else if value > adc.config.PSQThresholds[1] {
			digitalCode = 2
			quantizedValue = 1.0
		} else {
			digitalCode = 1
			quantizedValue = 0.0
		}

	default:
		// Uniform quantization
		// Normalize to [0, levels)
		normalized := (value + adc.config.FullScaleV/2) / adc.config.FullScaleV * float64(adc.levels)

		// Clamp and track overflow/underflow
		if normalized < 0 {
			adc.underflowCount++
			normalized = 0
		} else if normalized >= float64(adc.levels) {
			adc.overflowCount++
			normalized = float64(adc.levels) - 1
		}

		digitalCode = int(normalized)
		quantizedValue = (float64(digitalCode)+0.5)*adc.lsbSize - adc.config.FullScaleV/2
	}

	return digitalCode, quantizedValue
}

// GetEnergy returns energy consumed for a conversion
func (adc *ADC) GetEnergy() float64 {
	return adc.config.EnergyPerConvPJ
}

// GetStatistics returns ADC operation statistics
func (adc *ADC) GetStatistics() ADCStats {
	return ADCStats{
		TotalConversions: adc.totalConversions,
		OverflowCount:    adc.overflowCount,
		UnderflowCount:   adc.underflowCount,
		OverflowRate:     float64(adc.overflowCount) / float64(adc.totalConversions+1),
		UnderflowRate:    float64(adc.underflowCount) / float64(adc.totalConversions+1),
	}
}

// ADCStats holds ADC operation statistics
type ADCStats struct {
	TotalConversions int64
	OverflowCount    int64
	UnderflowCount   int64
	OverflowRate     float64
	UnderflowRate    float64
}

// =============================================================================
// DAC SIMULATION
// =============================================================================

// DACConfig holds DAC configuration parameters
type DACConfig struct {
	BitPrecision    int     // Resolution bits
	FullScaleV      float64 // Full scale voltage
	SettlingTimeNs  float64 // Settling time

	// Non-idealities
	INL_LSB         float64
	DNL_LSB         float64
	GlitchEnergyPJ  float64 // Glitch energy per transition

	// Power model
	StaticPowerUW   float64
	DynamicEnergyPJ float64 // Per conversion
}

// DefaultDACConfig returns typical 8-bit DAC config
func DefaultDACConfig() *DACConfig {
	return &DACConfig{
		BitPrecision:    8,
		FullScaleV:      1.0,
		SettlingTimeNs:  10,
		INL_LSB:         0.5,
		DNL_LSB:         0.5,
		GlitchEnergyPJ:  0.1,
		StaticPowerUW:   10,
		DynamicEnergyPJ: 0.2,
	}
}

// DAC simulates digital-to-analog conversion
type DAC struct {
	config    *DACConfig
	rng       *rand.Rand
	levels    int
	lsbSize   float64
	lastCode  int
}

// NewDAC creates a new DAC with given configuration
func NewDAC(config *DACConfig) *DAC {
	if config == nil {
		config = DefaultDACConfig()
	}

	return &DAC{
		config:  config,
		rng:     rand.New(rand.NewSource(42)),
		levels:  1 << config.BitPrecision,
		lsbSize: config.FullScaleV / float64(1<<config.BitPrecision),
	}
}

// Convert performs DAC conversion with non-idealities
func (dac *DAC) Convert(digitalCode int) float64 {
	// Clamp code
	if digitalCode < 0 {
		digitalCode = 0
	} else if digitalCode >= dac.levels {
		digitalCode = dac.levels - 1
	}

	// Ideal conversion
	analogValue := float64(digitalCode) * dac.lsbSize

	// Add INL/DNL errors
	inlError := dac.rng.NormFloat64() * dac.config.INL_LSB * dac.lsbSize
	dnlError := dac.rng.NormFloat64() * dac.config.DNL_LSB * dac.lsbSize

	analogValue += inlError + dnlError

	dac.lastCode = digitalCode
	return analogValue
}

// GetEnergy returns energy for conversion including glitch
func (dac *DAC) GetEnergy(newCode int) float64 {
	// Count bit transitions for glitch energy
	transitions := countBitTransitions(dac.lastCode, newCode)
	return dac.config.DynamicEnergyPJ + float64(transitions)*dac.config.GlitchEnergyPJ
}

func countBitTransitions(a, b int) int {
	xor := a ^ b
	count := 0
	for xor != 0 {
		count += xor & 1
		xor >>= 1
	}
	return count
}

// =============================================================================
// COMPLETE CIM SIGNAL CHAIN
// =============================================================================

// CIMSignalChain models the complete analog compute path
type CIMSignalChain struct {
	dac        *DAC
	adc        *ADC
	noiseModel *NoiseModel

	// Array parameters
	arrayRows  int
	arrayCols  int

	// Energy tracking
	totalEnergyPJ float64

	// Configuration
	config *SignalChainConfig
}

// SignalChainConfig holds signal chain configuration
type SignalChainConfig struct {
	// DAC for input activation
	InputDACBits int

	// ADC for output sensing
	OutputADCBits int
	ADCType       ADCType

	// Array dimensions (affects IR drop, etc.)
	ArrayRows int
	ArrayCols int

	// Enable partial sum processing
	EnablePSQ         bool
	PSQAccumBits      int // Accumulator bits after PSQ

	// Enable noise injection for training
	EnableNoiseAware  bool
	NoiseInjectionSTD float64
}

// DefaultSignalChainConfig returns typical CIM signal chain config
func DefaultSignalChainConfig() *SignalChainConfig {
	return &SignalChainConfig{
		InputDACBits:      8,
		OutputADCBits:     6,
		ADCType:           ADCSAR,
		ArrayRows:         64,
		ArrayCols:         64,
		EnablePSQ:         false,
		PSQAccumBits:      16,
		EnableNoiseAware:  false,
		NoiseInjectionSTD: 0.0,
	}
}

// HighEfficiencyConfig returns HCiM-style config for 28× energy reduction
func HighEfficiencyConfig() *SignalChainConfig {
	return &SignalChainConfig{
		InputDACBits:      4, // Reduced input precision
		OutputADCBits:     1, // Binary PSQ
		ADCType:           ADCBinaryPSQ,
		ArrayRows:         128,
		ArrayCols:         128,
		EnablePSQ:         true,
		PSQAccumBits:      8, // Digital accumulation
		EnableNoiseAware:  true,
		NoiseInjectionSTD: 0.05,
	}
}

// NewCIMSignalChain creates a new signal chain model
func NewCIMSignalChain(config *SignalChainConfig) *CIMSignalChain {
	if config == nil {
		config = DefaultSignalChainConfig()
	}

	// Create DAC config
	dacConfig := &DACConfig{
		BitPrecision:    config.InputDACBits,
		FullScaleV:      1.0,
		SettlingTimeNs:  10,
		INL_LSB:         0.5,
		DNL_LSB:         0.5,
		DynamicEnergyPJ: 0.2 * float64(config.InputDACBits) / 8.0, // Scale with bits
	}

	// Create ADC config based on type
	var adcConfig *ADCConfig
	switch config.ADCType {
	case ADCBinaryPSQ:
		adcConfig = HCiMADCConfig()
	case ADCTernaryPSQ:
		adcConfig = TernaryPSQConfig()
	default:
		adcConfig = DefaultADCConfig()
		adcConfig.BitPrecision = config.OutputADCBits
	}

	// Create noise config
	noiseConfig := DefaultNoiseConfig()
	if config.EnableNoiseAware {
		noiseConfig.CycleToCycleVariation = config.NoiseInjectionSTD
	}

	return &CIMSignalChain{
		dac:        NewDAC(dacConfig),
		adc:        NewADC(adcConfig),
		noiseModel: NewNoiseModel(noiseConfig),
		arrayRows:  config.ArrayRows,
		arrayCols:  config.ArrayCols,
		config:     config,
	}
}

// ProcessMVMColumn performs analog MAC on one column with noise
func (chain *CIMSignalChain) ProcessMVMColumn(
	inputs []float64,
	weights []float64,
	colIndex int,
) float64 {
	if len(inputs) != len(weights) {
		panic("input and weight length mismatch")
	}

	// Analog accumulation with noise
	analogSum := 0.0

	for row, input := range inputs {
		// DAC conversion of input
		inputCode := int(input * float64(chain.dac.levels-1))
		analogInput := chain.dac.Convert(inputCode)
		chain.totalEnergyPJ += chain.dac.GetEnergy(inputCode)

		// Analog multiply with stored weight
		product := analogInput * weights[row]

		// Add cell-level noise
		position := CellPosition{Row: row, Col: colIndex}
		product = chain.noiseModel.AddNoise(product, position)

		analogSum += product
	}

	// ADC conversion
	_, quantized := chain.adc.Convert(analogSum)
	chain.totalEnergyPJ += chain.adc.GetEnergy()

	return quantized
}

// ProcessMVM performs full matrix-vector multiplication
func (chain *CIMSignalChain) ProcessMVM(
	inputs []float64,
	weights [][]float64, // [rows][cols]
) []float64 {
	if len(weights) == 0 {
		return nil
	}

	numCols := len(weights[0])
	outputs := make([]float64, numCols)

	// Process each column
	for col := 0; col < numCols; col++ {
		// Extract column weights
		colWeights := make([]float64, len(weights))
		for row := range weights {
			colWeights[row] = weights[row][col]
		}

		outputs[col] = chain.ProcessMVMColumn(inputs, colWeights, col)
	}

	return outputs
}

// GetTotalEnergy returns accumulated energy consumption
func (chain *CIMSignalChain) GetTotalEnergy() float64 {
	return chain.totalEnergyPJ
}

// GetEnergyBreakdown returns energy breakdown by component
func (chain *CIMSignalChain) GetEnergyBreakdown() EnergyBreakdown {
	// Estimate based on typical ratios
	// ADC: 60%, DAC: 10%, Compute: 30%
	return EnergyBreakdown{
		TotalEnergyPJ: chain.totalEnergyPJ,
		ADCEnergyPJ:   chain.totalEnergyPJ * 0.60,
		DACEnergyPJ:   chain.totalEnergyPJ * 0.10,
		ComputePJ:     chain.totalEnergyPJ * 0.30,
		ADCPercent:    60.0,
		DACPercent:    10.0,
		ComputePercent: 30.0,
	}
}

// EnergyBreakdown holds energy consumption breakdown
type EnergyBreakdown struct {
	TotalEnergyPJ  float64
	ADCEnergyPJ    float64
	DACEnergyPJ    float64
	ComputePJ      float64
	ADCPercent     float64
	DACPercent     float64
	ComputePercent float64
}

// =============================================================================
// SNR AND PRECISION ANALYSIS
// =============================================================================

// SNRAnalyzer computes signal-to-noise ratio metrics
type SNRAnalyzer struct {
	signalPower float64
	noisePower  float64
	samples     int
}

// NewSNRAnalyzer creates a new SNR analyzer
func NewSNRAnalyzer() *SNRAnalyzer {
	return &SNRAnalyzer{}
}

// AddSample adds a (ideal, actual) sample pair
func (snr *SNRAnalyzer) AddSample(ideal, actual float64) {
	snr.signalPower += ideal * ideal
	noise := actual - ideal
	snr.noisePower += noise * noise
	snr.samples++
}

// GetSNR returns signal-to-noise ratio in dB
func (snr *SNRAnalyzer) GetSNR() float64 {
	if snr.samples == 0 || snr.noisePower == 0 {
		return math.Inf(1)
	}
	return 10 * math.Log10(snr.signalPower/snr.noisePower)
}

// GetENOB returns effective number of bits
func (snr *SNRAnalyzer) GetENOB() float64 {
	// ENOB = (SNR - 1.76) / 6.02
	snrDB := snr.GetSNR()
	if math.IsInf(snrDB, 1) {
		return 16 // Max reasonable ENOB
	}
	return (snrDB - 1.76) / 6.02
}

// =============================================================================
// NOISE-AWARE TRAINING SUPPORT
// =============================================================================

// NoiseAwareConfig configures noise injection for training
type NoiseAwareConfig struct {
	// Weight noise (programming variability)
	WeightNoiseSTD float64

	// Activation noise (read noise)
	ActivationNoiseSTD float64

	// Quantization simulation
	SimulateQuantization bool
	WeightBits          int
	ActivationBits      int

	// Clipping simulation
	ClipWeights     bool
	WeightClipRange float64
}

// DefaultNoiseAwareConfig returns typical config for noise-aware training
func DefaultNoiseAwareConfig() *NoiseAwareConfig {
	return &NoiseAwareConfig{
		WeightNoiseSTD:       0.03, // 3% variation
		ActivationNoiseSTD:   0.02, // 2% variation
		SimulateQuantization: true,
		WeightBits:           6,
		ActivationBits:       6,
		ClipWeights:          true,
		WeightClipRange:      1.0,
	}
}

// NoiseInjector injects noise during training
type NoiseInjector struct {
	config *NoiseAwareConfig
	rng    *rand.Rand
}

// NewNoiseInjector creates a new noise injector
func NewNoiseInjector(config *NoiseAwareConfig) *NoiseInjector {
	if config == nil {
		config = DefaultNoiseAwareConfig()
	}
	return &NoiseInjector{
		config: config,
		rng:    rand.New(rand.NewSource(42)),
	}
}

// InjectWeightNoise adds noise to weights during forward pass
func (ni *NoiseInjector) InjectWeightNoise(weight float64) float64 {
	noise := ni.rng.NormFloat64() * ni.config.WeightNoiseSTD
	return weight + weight*noise
}

// InjectActivationNoise adds noise to activations
func (ni *NoiseInjector) InjectActivationNoise(activation float64) float64 {
	noise := ni.rng.NormFloat64() * ni.config.ActivationNoiseSTD
	return activation + noise
}

// SimulateQuantization applies quantization to value
func (ni *NoiseInjector) SimulateQuantization(value float64, bits int) float64 {
	levels := 1 << bits
	scale := float64(levels - 1)

	// Assume [-1, 1] range
	normalized := (value + 1) / 2 // Map to [0, 1]
	quantized := math.Round(normalized * scale) / scale
	return quantized*2 - 1 // Map back to [-1, 1]
}

// ProcessWeight applies full weight processing pipeline
func (ni *NoiseInjector) ProcessWeight(weight float64) float64 {
	result := weight

	// Clip weights
	if ni.config.ClipWeights {
		if result > ni.config.WeightClipRange {
			result = ni.config.WeightClipRange
		} else if result < -ni.config.WeightClipRange {
			result = -ni.config.WeightClipRange
		}
	}

	// Quantize
	if ni.config.SimulateQuantization {
		result = ni.SimulateQuantization(result, ni.config.WeightBits)
	}

	// Add noise
	result = ni.InjectWeightNoise(result)

	return result
}

// =============================================================================
// CALIBRATION AND COMPENSATION
// =============================================================================

// ArrayCalibration holds calibration data for a crossbar array
type ArrayCalibration struct {
	// Per-cell offset calibration
	OffsetMap [][]float64

	// Per-column gain calibration
	GainMap []float64

	// IR drop compensation coefficients
	IRDropCoeffs [][]float64

	// Calibration metadata
	CalibrationDate string
	Temperature     float64
	IsValid         bool
}

// NewArrayCalibration creates calibration for array dimensions
func NewArrayCalibration(rows, cols int) *ArrayCalibration {
	offsetMap := make([][]float64, rows)
	irDropCoeffs := make([][]float64, rows)
	for i := range offsetMap {
		offsetMap[i] = make([]float64, cols)
		irDropCoeffs[i] = make([]float64, cols)
	}

	return &ArrayCalibration{
		OffsetMap:    offsetMap,
		GainMap:      make([]float64, cols),
		IRDropCoeffs: irDropCoeffs,
		IsValid:      false,
	}
}

// CalibrateArray performs array calibration using test patterns
func CalibrateArray(chain *CIMSignalChain, rows, cols int) *ArrayCalibration {
	cal := NewArrayCalibration(rows, cols)

	// Offset calibration: measure with zero input
	zeroInput := make([]float64, rows)
	zeroWeights := make([][]float64, rows)
	for i := range zeroWeights {
		zeroWeights[i] = make([]float64, cols)
	}

	offsets := chain.ProcessMVM(zeroInput, zeroWeights)
	for col, offset := range offsets {
		cal.GainMap[col] = 1.0 // Initial gain
		for row := range cal.OffsetMap {
			cal.OffsetMap[row][col] = offset / float64(rows)
		}
	}

	// Gain calibration: measure with unity input/weight
	oneInput := make([]float64, rows)
	oneWeights := make([][]float64, rows)
	for i := range oneInput {
		oneInput[i] = 1.0
		oneWeights[i] = make([]float64, cols)
		for j := range oneWeights[i] {
			oneWeights[i][j] = 1.0 / float64(rows)
		}
	}

	outputs := chain.ProcessMVM(oneInput, oneWeights)
	for col, output := range outputs {
		expected := 1.0
		if output != 0 {
			cal.GainMap[col] = expected / output
		}
	}

	cal.IsValid = true
	cal.CalibrationDate = "2026-01-15"
	cal.Temperature = 300

	return cal
}

// ApplyCalibration applies calibration correction to output
func ApplyCalibration(output float64, col int, cal *ArrayCalibration) float64 {
	if !cal.IsValid {
		return output
	}

	// Apply gain correction
	corrected := output * cal.GainMap[col]

	return corrected
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// CalculateSNRFromBits returns theoretical SNR for given ADC bits
func CalculateSNRFromBits(bits int) float64 {
	// SNR = 6.02N + 1.76 dB
	return 6.02*float64(bits) + 1.76
}

// CalculateBitsFromSNR returns required bits for target SNR
func CalculateBitsFromSNR(snrDB float64) int {
	// N = (SNR - 1.76) / 6.02
	bits := (snrDB - 1.76) / 6.02
	return int(math.Ceil(bits))
}

// EstimateEnergyReduction estimates energy savings from ADC bit reduction
func EstimateEnergyReduction(originalBits, reducedBits int) float64 {
	// Energy roughly scales as 2^bits for flash ADC
	// For SAR, scales more linearly
	originalEnergy := float64(originalBits)
	reducedEnergy := float64(reducedBits)
	return originalEnergy / reducedEnergy
}

// FormatNoiseReport generates a human-readable noise analysis report
func FormatNoiseReport(nm *NoiseModel) string {
	stats := nm.GetNoiseStatistics()
	report := "=== CIM Noise Analysis Report ===\n\n"

	noiseNames := map[NoiseType]string{
		NoiseThermal:      "Thermal",
		NoiseShot:         "Shot",
		NoiseProgramming:  "Programming",
		NoiseRead:         "Read",
		NoiseIRDrop:       "IR Drop",
		NoiseQuantization: "Quantization",
		NoiseCrosstalk:    "Crosstalk",
	}

	for noiseType, stat := range stats {
		name := noiseNames[noiseType]
		report += fmt.Sprintf("%s Noise:\n", name)
		report += fmt.Sprintf("  Mean:   %.6f\n", stat.Mean)
		report += fmt.Sprintf("  StdDev: %.6f\n", stat.StdDev)
		report += fmt.Sprintf("  MaxAbs: %.6f\n", stat.MaxAbs)
		report += fmt.Sprintf("  Samples: %d\n\n", stat.Samples)
	}

	return report
}

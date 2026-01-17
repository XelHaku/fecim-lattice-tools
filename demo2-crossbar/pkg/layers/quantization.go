// Package layers provides neural network layer implementations for crossbar-based CIM.
// quantization.go implements comprehensive quantization utilities for model compression
// and hardware deployment on crossbar arrays.
//
// Quantization strategies:
// - Post-training quantization (PTQ)
// - Quantization-aware training (QAT) simulation
// - Mixed-precision quantization
// - Per-channel vs per-tensor quantization
// - Symmetric vs asymmetric quantization
//
// CIM-specific considerations:
// - Weight quantization for conductance mapping
// - Activation quantization for DAC input
// - ADC output quantization simulation

package layers

import (
	"math"
	"sort"
)

// QuantizationMode defines the quantization approach
type QuantizationMode string

const (
	QuantModeSymmetric   QuantizationMode = "symmetric"
	QuantModeAsymmetric  QuantizationMode = "asymmetric"
	QuantModePerTensor   QuantizationMode = "per_tensor"
	QuantModePerChannel  QuantizationMode = "per_channel"
)

// QuantizationConfig configures quantization behavior
type QuantizationConfig struct {
	WeightBits      int              // Bits for weights (2-16)
	ActivationBits  int              // Bits for activations
	BiasParBits     int              // Bits for bias (usually higher)
	Mode            QuantizationMode // Symmetric or asymmetric
	PerChannel      bool             // Per-channel quantization
	CalibrationSize int              // Number of samples for calibration
	ClipPercentile  float64          // Percentile for outlier clipping
}

// DefaultQuantizationConfig returns default config
func DefaultQuantizationConfig() *QuantizationConfig {
	return &QuantizationConfig{
		WeightBits:      8,
		ActivationBits:  8,
		BiasParBits:     16,
		Mode:            QuantModeSymmetric,
		PerChannel:      false,
		CalibrationSize: 1000,
		ClipPercentile:  99.9,
	}
}

// CIMQuantizationConfig returns CIM-optimized config
func CIMQuantizationConfig(adcBits, dacBits, weightBits int) *QuantizationConfig {
	return &QuantizationConfig{
		WeightBits:      weightBits,
		ActivationBits:  dacBits,
		BiasParBits:     adcBits + 4, // Higher precision for accumulation
		Mode:            QuantModeSymmetric,
		PerChannel:      true,
		CalibrationSize: 1000,
		ClipPercentile:  99.5, // More aggressive clipping for CIM
	}
}

// QuantizationParams stores computed quantization parameters
type QuantizationParams struct {
	Scale      float64 // Scale factor
	ZeroPoint  int     // Zero point (asymmetric only)
	Min        float64 // Min representable value
	Max        float64 // Max representable value
	Bits       int
	Symmetric  bool
}

// Quantizer handles quantization operations
type Quantizer struct {
	Config *QuantizationConfig
}

// NewQuantizer creates a new quantizer
func NewQuantizer(config *QuantizationConfig) *Quantizer {
	if config == nil {
		config = DefaultQuantizationConfig()
	}
	return &Quantizer{Config: config}
}

// ============================================================================
// Quantization Parameter Computation
// ============================================================================

// ComputeWeightParams computes quantization parameters for weights
func (q *Quantizer) ComputeWeightParams(weights [][]float64) *QuantizationParams {
	if len(weights) == 0 {
		return nil
	}

	// Find min/max
	minVal, maxVal := weights[0][0], weights[0][0]
	for _, row := range weights {
		for _, val := range row {
			if val < minVal {
				minVal = val
			}
			if val > maxVal {
				maxVal = val
			}
		}
	}

	return q.computeParams(minVal, maxVal, q.Config.WeightBits, q.Config.Mode == QuantModeSymmetric)
}

// ComputeActivationParams computes params from calibration data
func (q *Quantizer) ComputeActivationParams(activations [][]float64) *QuantizationParams {
	if len(activations) == 0 {
		return nil
	}

	// Collect all values
	var allVals []float64
	for _, act := range activations {
		allVals = append(allVals, act...)
	}

	// Apply percentile clipping
	minVal, maxVal := q.computeClippedRange(allVals)

	return q.computeParams(minVal, maxVal, q.Config.ActivationBits, q.Config.Mode == QuantModeSymmetric)
}

// ComputePerChannelParams computes per-output-channel parameters
func (q *Quantizer) ComputePerChannelParams(weights [][]float64) []*QuantizationParams {
	if len(weights) == 0 {
		return nil
	}

	params := make([]*QuantizationParams, len(weights))
	for i, row := range weights {
		minVal, maxVal := row[0], row[0]
		for _, val := range row {
			if val < minVal {
				minVal = val
			}
			if val > maxVal {
				maxVal = val
			}
		}
		params[i] = q.computeParams(minVal, maxVal, q.Config.WeightBits, q.Config.Mode == QuantModeSymmetric)
	}
	return params
}

func (q *Quantizer) computeParams(minVal, maxVal float64, bits int, symmetric bool) *QuantizationParams {
	levels := float64(int(1) << bits)

	if symmetric {
		// Symmetric quantization: zero point = 0
		absMax := math.Max(math.Abs(minVal), math.Abs(maxVal))
		if absMax < 1e-10 {
			absMax = 1e-10
		}
		scale := absMax / ((levels - 1) / 2)
		return &QuantizationParams{
			Scale:     scale,
			ZeroPoint: 0,
			Min:       -absMax,
			Max:       absMax,
			Bits:      bits,
			Symmetric: true,
		}
	}

	// Asymmetric quantization
	if maxVal-minVal < 1e-10 {
		maxVal = minVal + 1e-10
	}
	scale := (maxVal - minVal) / (levels - 1)
	zeroPoint := int(math.Round(-minVal / scale))
	if zeroPoint < 0 {
		zeroPoint = 0
	}
	if zeroPoint >= int(levels) {
		zeroPoint = int(levels) - 1
	}

	return &QuantizationParams{
		Scale:     scale,
		ZeroPoint: zeroPoint,
		Min:       minVal,
		Max:       maxVal,
		Bits:      bits,
		Symmetric: false,
	}
}

func (q *Quantizer) computeClippedRange(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 1
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	lowIdx := int(float64(len(sorted)) * (100 - q.Config.ClipPercentile) / 100)
	highIdx := int(float64(len(sorted)) * q.Config.ClipPercentile / 100)

	if lowIdx >= len(sorted) {
		lowIdx = len(sorted) - 1
	}
	if highIdx >= len(sorted) {
		highIdx = len(sorted) - 1
	}

	return sorted[lowIdx], sorted[highIdx]
}

// ============================================================================
// Quantization Operations
// ============================================================================

// QuantizeWeights quantizes weight matrix
func (q *Quantizer) QuantizeWeights(weights [][]float64) ([][]int, *QuantizationParams) {
	params := q.ComputeWeightParams(weights)
	if params == nil {
		return nil, nil
	}

	quantized := make([][]int, len(weights))
	for i, row := range weights {
		quantized[i] = make([]int, len(row))
		for j, val := range row {
			quantized[i][j] = quantizeValue(val, params)
		}
	}
	return quantized, params
}

// QuantizeWeightsPerChannel quantizes with per-channel parameters
func (q *Quantizer) QuantizeWeightsPerChannel(weights [][]float64) ([][]int, []*QuantizationParams) {
	paramsPerChannel := q.ComputePerChannelParams(weights)
	if paramsPerChannel == nil {
		return nil, nil
	}

	quantized := make([][]int, len(weights))
	for i, row := range weights {
		quantized[i] = make([]int, len(row))
		for j, val := range row {
			quantized[i][j] = quantizeValue(val, paramsPerChannel[i])
		}
	}
	return quantized, paramsPerChannel
}

// QuantizeActivations quantizes activation vector
func (q *Quantizer) QuantizeActivations(activations []float64, params *QuantizationParams) []int {
	quantized := make([]int, len(activations))
	for i, val := range activations {
		quantized[i] = quantizeValue(val, params)
	}
	return quantized
}

// DequantizeWeights converts quantized weights back to float
func (q *Quantizer) DequantizeWeights(quantized [][]int, params *QuantizationParams) [][]float64 {
	weights := make([][]float64, len(quantized))
	for i, row := range quantized {
		weights[i] = make([]float64, len(row))
		for j, qval := range row {
			weights[i][j] = dequantizeValue(qval, params)
		}
	}
	return weights
}

// DequantizeActivations converts quantized activations back to float
func (q *Quantizer) DequantizeActivations(quantized []int, params *QuantizationParams) []float64 {
	activations := make([]float64, len(quantized))
	for i, qval := range quantized {
		activations[i] = dequantizeValue(qval, params)
	}
	return activations
}

func quantizeValue(val float64, params *QuantizationParams) int {
	levels := int(1) << params.Bits
	minQ := 0
	maxQ := levels - 1

	if params.Symmetric {
		minQ = -(levels / 2)
		maxQ = levels/2 - 1
	}

	q := int(math.Round(val/params.Scale)) + params.ZeroPoint
	if q < minQ {
		q = minQ
	}
	if q > maxQ {
		q = maxQ
	}
	return q
}

func dequantizeValue(qval int, params *QuantizationParams) float64 {
	return float64(qval-params.ZeroPoint) * params.Scale
}

// ============================================================================
// Fake Quantization (for QAT simulation)
// ============================================================================

// FakeQuantize applies quantization and immediately dequantizes (simulates quantization error)
func (q *Quantizer) FakeQuantize(weights [][]float64) [][]float64 {
	params := q.ComputeWeightParams(weights)
	if params == nil {
		return weights
	}

	fakeQuant := make([][]float64, len(weights))
	for i, row := range weights {
		fakeQuant[i] = make([]float64, len(row))
		for j, val := range row {
			qval := quantizeValue(val, params)
			fakeQuant[i][j] = dequantizeValue(qval, params)
		}
	}
	return fakeQuant
}

// FakeQuantizeWithParams uses provided params
func (q *Quantizer) FakeQuantizeWithParams(weights [][]float64, params *QuantizationParams) [][]float64 {
	fakeQuant := make([][]float64, len(weights))
	for i, row := range weights {
		fakeQuant[i] = make([]float64, len(row))
		for j, val := range row {
			qval := quantizeValue(val, params)
			fakeQuant[i][j] = dequantizeValue(qval, params)
		}
	}
	return fakeQuant
}

// StraightThroughEstimator computes gradient through fake quantization
func (q *Quantizer) StraightThroughEstimator(grad, weights [][]float64, params *QuantizationParams) [][]float64 {
	// STE: pass gradient through unchanged within quantization range
	// clip gradient outside range
	steGrad := make([][]float64, len(grad))
	for i, row := range grad {
		steGrad[i] = make([]float64, len(row))
		for j, g := range row {
			w := weights[i][j]
			if w < params.Min || w > params.Max {
				steGrad[i][j] = 0 // Clip gradient outside range
			} else {
				steGrad[i][j] = g
			}
		}
	}
	return steGrad
}

// ============================================================================
// Mixed Precision Quantization
// ============================================================================

// MixedPrecisionConfig defines per-layer precision
type MixedPrecisionConfig struct {
	LayerBits map[string]int // Layer name -> bit width
	Default   int            // Default bit width
}

// NewMixedPrecisionConfig creates a mixed precision config
func NewMixedPrecisionConfig(defaultBits int) *MixedPrecisionConfig {
	return &MixedPrecisionConfig{
		LayerBits: make(map[string]int),
		Default:   defaultBits,
	}
}

// SetLayerBits sets bit width for a specific layer
func (mpc *MixedPrecisionConfig) SetLayerBits(layerName string, bits int) {
	mpc.LayerBits[layerName] = bits
}

// GetLayerBits returns bit width for a layer
func (mpc *MixedPrecisionConfig) GetLayerBits(layerName string) int {
	if bits, ok := mpc.LayerBits[layerName]; ok {
		return bits
	}
	return mpc.Default
}

// MixedPrecisionQuantizer handles mixed precision quantization
type MixedPrecisionQuantizer struct {
	Config *MixedPrecisionConfig
	Base   *QuantizationConfig
}

// NewMixedPrecisionQuantizer creates mixed precision quantizer
func NewMixedPrecisionQuantizer(baseConfig *QuantizationConfig, mpConfig *MixedPrecisionConfig) *MixedPrecisionQuantizer {
	return &MixedPrecisionQuantizer{
		Config: mpConfig,
		Base:   baseConfig,
	}
}

// QuantizeLayer quantizes a specific layer with its assigned precision
func (mpq *MixedPrecisionQuantizer) QuantizeLayer(layerName string, weights [][]float64) ([][]int, *QuantizationParams) {
	bits := mpq.Config.GetLayerBits(layerName)

	config := *mpq.Base
	config.WeightBits = bits

	q := NewQuantizer(&config)
	return q.QuantizeWeights(weights)
}

// ============================================================================
// CIM-Specific Quantization
// ============================================================================

// ConductanceQuantizer maps weights to crossbar conductance values
type ConductanceQuantizer struct {
	MinCond    float64 // Minimum conductance
	MaxCond    float64 // Maximum conductance
	NumLevels  int     // Number of programmable levels
	Symmetric  bool    // Symmetric around mid-point
}

// NewConductanceQuantizer creates a conductance quantizer
func NewConductanceQuantizer(minCond, maxCond float64, bits int) *ConductanceQuantizer {
	return &ConductanceQuantizer{
		MinCond:   minCond,
		MaxCond:   maxCond,
		NumLevels: 1 << bits,
		Symmetric: true,
	}
}

// MapToConductance maps weights to conductance values
func (cq *ConductanceQuantizer) MapToConductance(weights [][]float64) ([][]float64, float64) {
	// Find weight range
	minW, maxW := weights[0][0], weights[0][0]
	for _, row := range weights {
		for _, val := range row {
			if val < minW {
				minW = val
			}
			if val > maxW {
				maxW = val
			}
		}
	}

	absMax := math.Max(math.Abs(minW), math.Abs(maxW))
	if absMax < 1e-10 {
		absMax = 1e-10
	}

	// Scale factor for later reconstruction
	scale := absMax

	// Map to conductances
	conductances := make([][]float64, len(weights))
	condRange := cq.MaxCond - cq.MinCond
	midCond := (cq.MaxCond + cq.MinCond) / 2

	for i, row := range weights {
		conductances[i] = make([]float64, len(row))
		for j, val := range row {
			// Normalize to [-1, 1]
			normalized := val / absMax

			// Map to conductance
			if cq.Symmetric {
				// Symmetric: -1 -> MinCond, 0 -> MidCond, 1 -> MaxCond
				cond := midCond + normalized*(condRange/2)
				conductances[i][j] = cq.quantizeConductance(cond)
			} else {
				// Asymmetric: 0 -> MinCond, 1 -> MaxCond
				normalized01 := (normalized + 1) / 2
				cond := cq.MinCond + normalized01*condRange
				conductances[i][j] = cq.quantizeConductance(cond)
			}
		}
	}

	return conductances, scale
}

// quantizeConductance snaps conductance to nearest programmable level
func (cq *ConductanceQuantizer) quantizeConductance(cond float64) float64 {
	// Clamp to range
	if cond < cq.MinCond {
		cond = cq.MinCond
	}
	if cond > cq.MaxCond {
		cond = cq.MaxCond
	}

	// Quantize to levels
	step := (cq.MaxCond - cq.MinCond) / float64(cq.NumLevels-1)
	level := math.Round((cond - cq.MinCond) / step)
	return cq.MinCond + level*step
}

// DifferentialPair maps weight to differential conductance pair
func (cq *ConductanceQuantizer) DifferentialPair(weight, absMax float64) (float64, float64) {
	normalized := weight / absMax
	midCond := (cq.MaxCond + cq.MinCond) / 2
	condRange := cq.MaxCond - cq.MinCond

	// G+ = mid + normalized * range/2
	// G- = mid - normalized * range/2
	gPlus := midCond + normalized*(condRange/2)
	gMinus := midCond - normalized*(condRange/2)

	return cq.quantizeConductance(gPlus), cq.quantizeConductance(gMinus)
}

// MapToDifferentialConductance creates differential pair arrays
func (cq *ConductanceQuantizer) MapToDifferentialConductance(weights [][]float64) (gPlus, gMinus [][]float64, scale float64) {
	// Find weight range
	absMax := 0.0
	for _, row := range weights {
		for _, val := range row {
			if math.Abs(val) > absMax {
				absMax = math.Abs(val)
			}
		}
	}
	if absMax < 1e-10 {
		absMax = 1e-10
	}

	gPlus = make([][]float64, len(weights))
	gMinus = make([][]float64, len(weights))

	for i, row := range weights {
		gPlus[i] = make([]float64, len(row))
		gMinus[i] = make([]float64, len(row))
		for j, val := range row {
			gPlus[i][j], gMinus[i][j] = cq.DifferentialPair(val, absMax)
		}
	}

	return gPlus, gMinus, absMax
}

// ============================================================================
// ADC/DAC Quantization Simulation
// ============================================================================

// ADCSimulator simulates ADC conversion
type ADCSimulator struct {
	Bits      int
	VrefMin   float64
	VrefMax   float64
	NoiseStd  float64 // ADC noise
}

// NewADCSimulator creates an ADC simulator
func NewADCSimulator(bits int, vrefMin, vrefMax, noiseStd float64) *ADCSimulator {
	return &ADCSimulator{
		Bits:     bits,
		VrefMin:  vrefMin,
		VrefMax:  vrefMax,
		NoiseStd: noiseStd,
	}
}

// Convert simulates ADC conversion
func (adc *ADCSimulator) Convert(voltage float64) int {
	// Add noise
	voltage += adc.NoiseStd * randNorm()

	// Clamp to reference range
	if voltage < adc.VrefMin {
		voltage = adc.VrefMin
	}
	if voltage > adc.VrefMax {
		voltage = adc.VrefMax
	}

	// Quantize
	levels := 1 << adc.Bits
	normalized := (voltage - adc.VrefMin) / (adc.VrefMax - adc.VrefMin)
	code := int(math.Round(normalized * float64(levels-1)))

	if code < 0 {
		code = 0
	}
	if code >= levels {
		code = levels - 1
	}

	return code
}

// ConvertBatch converts batch of voltages
func (adc *ADCSimulator) ConvertBatch(voltages []float64) []int {
	codes := make([]int, len(voltages))
	for i, v := range voltages {
		codes[i] = adc.Convert(v)
	}
	return codes
}

// DACSimulator simulates DAC conversion
type DACSimulator struct {
	Bits      int
	VoutMin   float64
	VoutMax   float64
	INL       float64 // Integral non-linearity (LSB)
	DNL       float64 // Differential non-linearity (LSB)
}

// NewDACSimulator creates a DAC simulator
func NewDACSimulator(bits int, voutMin, voutMax float64) *DACSimulator {
	return &DACSimulator{
		Bits:    bits,
		VoutMin: voutMin,
		VoutMax: voutMax,
		INL:     0.5, // Default 0.5 LSB
		DNL:     0.5,
	}
}

// Convert simulates DAC conversion
func (dac *DACSimulator) Convert(code int) float64 {
	levels := 1 << dac.Bits

	if code < 0 {
		code = 0
	}
	if code >= levels {
		code = levels - 1
	}

	// Ideal voltage
	normalized := float64(code) / float64(levels-1)
	voltage := dac.VoutMin + normalized*(dac.VoutMax-dac.VoutMin)

	// Add INL error (simplified)
	lsb := (dac.VoutMax - dac.VoutMin) / float64(levels-1)
	inlError := dac.INL * lsb * math.Sin(math.Pi*normalized)
	voltage += inlError

	return voltage
}

// ConvertBatch converts batch of codes
func (dac *DACSimulator) ConvertBatch(codes []int) []float64 {
	voltages := make([]float64, len(codes))
	for i, c := range codes {
		voltages[i] = dac.Convert(c)
	}
	return voltages
}

// ============================================================================
// Quantization Error Analysis
// ============================================================================

// QuantizationError computes quantization error metrics
type QuantizationError struct {
	MSE          float64 // Mean squared error
	RMSE         float64 // Root mean squared error
	MAE          float64 // Mean absolute error
	MaxError     float64 // Maximum absolute error
	SNR          float64 // Signal-to-noise ratio (dB)
	SQNR         float64 // Signal-to-quantization-noise ratio
	NumClipped   int     // Number of clipped values
}

// ComputeQuantizationError computes error between original and quantized
func ComputeQuantizationError(original, quantized [][]float64) *QuantizationError {
	if len(original) == 0 || len(original) != len(quantized) {
		return nil
	}

	var sumSqErr, sumAbsErr, maxErr, sumSigSq float64
	numClipped := 0
	count := 0

	for i, row := range original {
		for j, val := range row {
			qval := quantized[i][j]
			err := val - qval

			sumSqErr += err * err
			sumAbsErr += math.Abs(err)
			sumSigSq += val * val

			if math.Abs(err) > maxErr {
				maxErr = math.Abs(err)
			}

			if qval != val && (qval == quantized[i][j] || qval == quantized[i][j]) {
				// Check if clipped (simplified)
				numClipped++
			}

			count++
		}
	}

	n := float64(count)
	mse := sumSqErr / n
	rmse := math.Sqrt(mse)
	mae := sumAbsErr / n

	snr := 0.0
	sqnr := 0.0
	if mse > 0 {
		snr = 10 * math.Log10(sumSigSq/sumSqErr)
		sqnr = snr
	}

	return &QuantizationError{
		MSE:        mse,
		RMSE:       rmse,
		MAE:        mae,
		MaxError:   maxErr,
		SNR:        snr,
		SQNR:       sqnr,
		NumClipped: numClipped,
	}
}

// ============================================================================
// Calibration
// ============================================================================

// CalibrationData stores calibration samples
type CalibrationData struct {
	Activations [][]float64
	LayerName   string
}

// Calibrator calibrates quantization parameters using data
type Calibrator struct {
	Config      *QuantizationConfig
	LayerParams map[string]*QuantizationParams
}

// NewCalibrator creates a calibrator
func NewCalibrator(config *QuantizationConfig) *Calibrator {
	return &Calibrator{
		Config:      config,
		LayerParams: make(map[string]*QuantizationParams),
	}
}

// Calibrate computes quantization parameters from calibration data
func (cal *Calibrator) Calibrate(data []CalibrationData) {
	q := NewQuantizer(cal.Config)

	for _, d := range data {
		params := q.ComputeActivationParams(d.Activations)
		cal.LayerParams[d.LayerName] = params
	}
}

// GetParams returns calibrated parameters for a layer
func (cal *Calibrator) GetParams(layerName string) *QuantizationParams {
	return cal.LayerParams[layerName]
}

// ============================================================================
// Utility Functions
// ============================================================================

// Helper for normal random (Box-Muller)
var hasSpare bool
var spare float64

func randNorm() float64 {
	if hasSpare {
		hasSpare = false
		return spare
	}

	var u, v, s float64
	for {
		u = randFloat()*2 - 1
		v = randFloat()*2 - 1
		s = u*u + v*v
		if s != 0 && s < 1 {
			break
		}
	}

	s = math.Sqrt(-2 * math.Log(s) / s)
	spare = v * s
	hasSpare = true
	return u * s
}

// Simple random float (for simulation - in production use math/rand)
var randSeed uint64 = 12345

func randFloat() float64 {
	randSeed = randSeed*6364136223846793005 + 1442695040888963407
	return float64(randSeed>>33) / float64(1<<31)
}

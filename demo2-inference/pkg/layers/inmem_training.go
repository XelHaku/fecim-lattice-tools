// Package layers provides neural network layer implementations for CIM simulation.
// This file implements in-memory training and hardware-aware noise injection.
//
// In-Memory Training:
// - Outer product weight updates for memristor crossbars
// - Three-phase training: forward, backprop, weight update
// - Analog gradient computation
// - On-device incremental learning
//
// Hardware-Aware Noise Injection:
// - Device variation (D2D, C2C)
// - ADC/DAC quantization noise
// - IR drop modeling
// - Thermal and retention drift
// - STE-based robust training
//
// References:
// - Full-stack CIM System (Nature Communications 2025)
// - Memristor Crossbar Training (ScienceDirect 2025)
// - NeuroSim V1.5 (arXiv 2505.02314)
// - ASiM Framework (arXiv 2411.11022)
// - Hardware-Aware Training (Nature Communications 2023)

package layers

import (
	"math"
	"math/rand"
)

// ============================================================================
// Noise Model Types
// ============================================================================

// NoiseType enumerates CIM noise sources.
type NoiseType int

const (
	NOISE_D2D_VARIATION  NoiseType = iota // Device-to-device variation
	NOISE_C2C_VARIATION                   // Cycle-to-cycle variation
	NOISE_ADC_QUANT                       // ADC quantization
	NOISE_DAC_QUANT                       // DAC quantization
	NOISE_IR_DROP                         // Wire resistance voltage drop
	NOISE_THERMAL                         // Thermal noise
	NOISE_RETENTION                       // Retention drift
	NOISE_PROGRAMMING                     // Write variation
	NOISE_READ                            // Read noise
)

// NoiseConfig configures hardware noise model.
type NoiseConfig struct {
	// Device variation
	D2DVariation   float64 // Device-to-device σ (0.02-0.1)
	C2CVariation   float64 // Cycle-to-cycle σ (0.01-0.05)

	// ADC/DAC
	ADCBits        int     // ADC resolution (4-8)
	DACBits        int     // DAC resolution (4-8)
	ADCNonlinearity float64 // ADC INL/DNL (LSB)

	// Circuit effects
	IRDropFactor   float64 // IR drop severity (0-1)
	WireResistance float64 // Wire resistance (Ω)

	// Temporal effects
	ThermalSigma   float64 // Thermal noise σ
	RetentionRate  float64 // Retention drift rate per second
	Temperature    float64 // Operating temperature (K)

	// Programming
	ProgramNoise   float64 // Programming variation σ
	ReadNoise      float64 // Read noise σ

	// Control
	EnabledNoises  map[NoiseType]bool
}

// DefaultNoiseConfig returns realistic noise configuration.
func DefaultNoiseConfig() *NoiseConfig {
	return &NoiseConfig{
		D2DVariation:    0.05,   // 5% device variation
		C2CVariation:    0.02,   // 2% cycle variation
		ADCBits:         6,
		DACBits:         8,
		ADCNonlinearity: 0.5,    // 0.5 LSB
		IRDropFactor:    0.1,    // 10% IR drop
		WireResistance:  10.0,   // 10 Ω
		ThermalSigma:    0.01,   // 1% thermal
		RetentionRate:   0.001,  // 0.1%/s drift
		Temperature:     300.0,  // Room temperature
		ProgramNoise:    0.03,   // 3% programming
		ReadNoise:       0.01,   // 1% read noise
		EnabledNoises: map[NoiseType]bool{
			NOISE_D2D_VARIATION: true,
			NOISE_C2C_VARIATION: true,
			NOISE_ADC_QUANT:     true,
			NOISE_DAC_QUANT:     true,
			NOISE_IR_DROP:       true,
			NOISE_THERMAL:       true,
			NOISE_READ:          true,
		},
	}
}

// ============================================================================
// Hardware Noise Simulator
// ============================================================================

// NoiseSimulator simulates CIM hardware non-idealities.
type NoiseSimulator struct {
	config *NoiseConfig
	rng    *rand.Rand

	// Statistics
	TotalNoiseInjected int64
	SNREstimate        float64
}

// NewNoiseSimulator creates a new noise simulator.
func NewNoiseSimulator(config *NoiseConfig) *NoiseSimulator {
	return &NoiseSimulator{
		config: config,
		rng:    rand.New(rand.NewSource(42)),
	}
}

// InjectNoise applies all enabled noises to a value.
func (ns *NoiseSimulator) InjectNoise(value float64, noiseType NoiseType) float64 {
	if !ns.config.EnabledNoises[noiseType] {
		return value
	}

	ns.TotalNoiseInjected++

	switch noiseType {
	case NOISE_D2D_VARIATION:
		return ns.applyD2DVariation(value)
	case NOISE_C2C_VARIATION:
		return ns.applyC2CVariation(value)
	case NOISE_ADC_QUANT:
		return ns.applyADCQuantization(value)
	case NOISE_DAC_QUANT:
		return ns.applyDACQuantization(value)
	case NOISE_THERMAL:
		return ns.applyThermalNoise(value)
	case NOISE_READ:
		return ns.applyReadNoise(value)
	default:
		return value
	}
}

// applyD2DVariation applies device-to-device variation.
// Modeled as multiplicative Gaussian noise.
func (ns *NoiseSimulator) applyD2DVariation(value float64) float64 {
	noise := 1.0 + ns.rng.NormFloat64()*ns.config.D2DVariation
	return value * noise
}

// applyC2CVariation applies cycle-to-cycle variation.
func (ns *NoiseSimulator) applyC2CVariation(value float64) float64 {
	noise := ns.rng.NormFloat64() * ns.config.C2CVariation * math.Abs(value)
	return value + noise
}

// applyADCQuantization applies ADC quantization noise.
func (ns *NoiseSimulator) applyADCQuantization(value float64) float64 {
	levels := float64(int(1) << ns.config.ADCBits)

	// Assume value is normalized to [-1, 1]
	quantized := math.Round(value * levels / 2)
	quantized = math.Max(-levels/2, math.Min(levels/2-1, quantized))

	// Add INL/DNL error
	inlError := ns.rng.NormFloat64() * ns.config.ADCNonlinearity / levels
	return (quantized / (levels / 2)) + inlError
}

// applyDACQuantization applies DAC quantization noise.
func (ns *NoiseSimulator) applyDACQuantization(value float64) float64 {
	levels := float64(int(1) << ns.config.DACBits)

	quantized := math.Round(value * levels / 2)
	quantized = math.Max(-levels/2, math.Min(levels/2-1, quantized))

	return quantized / (levels / 2)
}

// applyThermalNoise applies thermal noise.
func (ns *NoiseSimulator) applyThermalNoise(value float64) float64 {
	// kT noise scaled by temperature
	tempFactor := ns.config.Temperature / 300.0
	noise := ns.rng.NormFloat64() * ns.config.ThermalSigma * tempFactor
	return value + noise
}

// applyReadNoise applies read noise.
func (ns *NoiseSimulator) applyReadNoise(value float64) float64 {
	noise := ns.rng.NormFloat64() * ns.config.ReadNoise * math.Abs(value)
	return value + noise
}

// ApplyIRDrop applies IR drop to a crossbar column.
func (ns *NoiseSimulator) ApplyIRDrop(columnCurrent []float64, wireR float64) []float64 {
	if !ns.config.EnabledNoises[NOISE_IR_DROP] {
		return columnCurrent
	}

	result := make([]float64, len(columnCurrent))
	cumulativeDrop := 0.0

	for i, current := range columnCurrent {
		// IR drop accumulates down the column
		drop := current * wireR * ns.config.IRDropFactor
		cumulativeDrop += drop
		result[i] = current * (1.0 - cumulativeDrop)
	}

	return result
}

// ApplyRetentionDrift applies time-dependent retention drift.
func (ns *NoiseSimulator) ApplyRetentionDrift(value float64, timeSeconds float64) float64 {
	if !ns.config.EnabledNoises[NOISE_RETENTION] {
		return value
	}

	// Logarithmic drift model
	drift := ns.config.RetentionRate * math.Log(1+timeSeconds)
	return value * (1.0 - drift)
}

// ============================================================================
// Noisy Crossbar Array
// ============================================================================

// NoisyCrossbar represents a crossbar with realistic noise.
type NoisyCrossbar struct {
	Rows, Cols    int
	Conductances  [][]float64 // Programmed conductances
	D2DVariation  [][]float64 // Per-device D2D variation factor
	noiseSimulator *NoiseSimulator

	// Statistics
	TotalMACs     int64
	EffectiveSNR  float64
}

// NewNoisyCrossbar creates a new noisy crossbar array.
func NewNoisyCrossbar(rows, cols int, config *NoiseConfig) *NoisyCrossbar {
	cb := &NoisyCrossbar{
		Rows:         rows,
		Cols:         cols,
		Conductances: make([][]float64, rows),
		D2DVariation: make([][]float64, rows),
		noiseSimulator: NewNoiseSimulator(config),
	}

	// Initialize D2D variation (fixed per device)
	for i := range cb.Conductances {
		cb.Conductances[i] = make([]float64, cols)
		cb.D2DVariation[i] = make([]float64, cols)
		for j := range cb.D2DVariation[i] {
			cb.D2DVariation[i][j] = 1.0 + rand.NormFloat64()*config.D2DVariation
		}
	}

	return cb
}

// Program programs weights with programming noise.
func (cb *NoisyCrossbar) Program(weights [][]float64) {
	for i := 0; i < cb.Rows && i < len(weights); i++ {
		for j := 0; j < cb.Cols && j < len(weights[i]); j++ {
			// Apply programming noise
			programmed := cb.noiseSimulator.InjectNoise(weights[i][j], NOISE_PROGRAMMING)
			// Apply D2D variation
			cb.Conductances[i][j] = programmed * cb.D2DVariation[i][j]
		}
	}
}

// MVM performs matrix-vector multiplication with noise.
func (cb *NoisyCrossbar) MVM(input []float64) []float64 {
	output := make([]float64, cb.Rows)

	// Apply DAC quantization to input
	quantizedInput := make([]float64, len(input))
	for j, val := range input {
		quantizedInput[j] = cb.noiseSimulator.InjectNoise(val, NOISE_DAC_QUANT)
	}

	// Column currents for IR drop calculation
	columnCurrents := make([]float64, cb.Rows)

	for i := 0; i < cb.Rows; i++ {
		var sum float64
		for j := 0; j < cb.Cols && j < len(quantizedInput); j++ {
			// Apply C2C variation and read noise
			conductance := cb.noiseSimulator.InjectNoise(cb.Conductances[i][j], NOISE_C2C_VARIATION)
			conductance = cb.noiseSimulator.InjectNoise(conductance, NOISE_READ)

			sum += conductance * quantizedInput[j]
		}

		// Apply thermal noise
		sum = cb.noiseSimulator.InjectNoise(sum, NOISE_THERMAL)
		columnCurrents[i] = sum
	}

	// Apply IR drop
	columnCurrents = cb.noiseSimulator.ApplyIRDrop(columnCurrents, cb.noiseSimulator.config.WireResistance)

	// Apply ADC quantization
	for i, current := range columnCurrents {
		output[i] = cb.noiseSimulator.InjectNoise(current, NOISE_ADC_QUANT)
	}

	cb.TotalMACs += int64(cb.Rows * cb.Cols)
	return output
}

// ============================================================================
// In-Memory Training
// ============================================================================

// TrainingPhase enumerates training phases.
type TrainingPhase int

const (
	PHASE_FORWARD TrainingPhase = iota
	PHASE_BACKWARD
	PHASE_UPDATE
)

// InMemTrainingConfig configures in-memory training.
type InMemTrainingConfig struct {
	LearningRate    float64
	Momentum        float64
	WeightDecay     float64

	// Hardware constraints
	MinConductance  float64 // Minimum programmable conductance
	MaxConductance  float64 // Maximum programmable conductance
	UpdateGranularity float64 // Minimum weight update step

	// Training mode
	UseOuterProduct bool    // Use outer product updates
	UseAnalogBP     bool    // Use analog backpropagation
	BatchSize       int

	// Noise-aware training
	InjectTrainingNoise bool
	NoiseConfig         *NoiseConfig
}

// DefaultInMemTrainingConfig returns default training configuration.
func DefaultInMemTrainingConfig() *InMemTrainingConfig {
	return &InMemTrainingConfig{
		LearningRate:      0.01,
		Momentum:          0.9,
		WeightDecay:       1e-4,
		MinConductance:    0.1,
		MaxConductance:    1.0,
		UpdateGranularity: 0.01,
		UseOuterProduct:   true,
		UseAnalogBP:       true,
		BatchSize:         32,
		InjectTrainingNoise: true,
		NoiseConfig:       DefaultNoiseConfig(),
	}
}

// InMemTrainer implements in-memory training for CIM.
type InMemTrainer struct {
	config         *InMemTrainingConfig
	crossbar       *NoisyCrossbar

	// Weight state
	Weights        [][]float64
	Velocity       [][]float64 // Momentum
	Gradients      [][]float64

	// Activations for backprop
	ForwardActs    [][]float64
	BackwardDeltas [][]float64

	// Statistics
	Epoch          int
	Loss           float64
	Iterations     int64
}

// NewInMemTrainer creates a new in-memory trainer.
func NewInMemTrainer(rows, cols int, config *InMemTrainingConfig) *InMemTrainer {
	trainer := &InMemTrainer{
		config:    config,
		crossbar:  NewNoisyCrossbar(rows, cols, config.NoiseConfig),
		Weights:   make([][]float64, rows),
		Velocity:  make([][]float64, rows),
		Gradients: make([][]float64, rows),
	}

	// Initialize weights
	scale := math.Sqrt(2.0 / float64(rows+cols))
	for i := range trainer.Weights {
		trainer.Weights[i] = make([]float64, cols)
		trainer.Velocity[i] = make([]float64, cols)
		trainer.Gradients[i] = make([]float64, cols)
		for j := range trainer.Weights[i] {
			trainer.Weights[i][j] = (rand.Float64()*2 - 1) * scale
		}
	}

	// Program initial weights
	trainer.crossbar.Program(trainer.Weights)

	return trainer
}

// Forward performs forward pass with optional noise.
func (t *InMemTrainer) Forward(input []float64) []float64 {
	// Store input for backprop
	t.ForwardActs = append(t.ForwardActs, input)

	if t.config.InjectTrainingNoise {
		// Use noisy crossbar
		return t.crossbar.MVM(input)
	}

	// Clean forward pass
	output := make([]float64, len(t.Weights))
	for i, row := range t.Weights {
		var sum float64
		for j, w := range row {
			if j < len(input) {
				sum += w * input[j]
			}
		}
		output[i] = sum
	}

	return output
}

// Backward computes gradients using backpropagation.
func (t *InMemTrainer) Backward(outputGrad []float64) []float64 {
	rows := len(t.Weights)
	cols := len(t.Weights[0])

	// Get last forward activation
	var input []float64
	if len(t.ForwardActs) > 0 {
		input = t.ForwardActs[len(t.ForwardActs)-1]
		t.ForwardActs = t.ForwardActs[:len(t.ForwardActs)-1]
	} else {
		input = make([]float64, cols)
	}

	// Compute weight gradients: dW = outer(delta, input)
	if t.config.UseOuterProduct {
		t.computeOuterProductGradients(outputGrad, input)
	} else {
		t.computeStandardGradients(outputGrad, input)
	}

	// Compute input gradients: dX = W^T @ delta
	inputGrad := make([]float64, cols)
	for j := 0; j < cols; j++ {
		var sum float64
		for i := 0; i < rows && i < len(outputGrad); i++ {
			sum += t.Weights[i][j] * outputGrad[i]
		}
		inputGrad[j] = sum
	}

	t.BackwardDeltas = append(t.BackwardDeltas, outputGrad)

	return inputGrad
}

// computeOuterProductGradients computes gradients using outer product.
// dW[i][j] = delta[i] × input[j]
func (t *InMemTrainer) computeOuterProductGradients(delta, input []float64) {
	for i := range t.Gradients {
		for j := range t.Gradients[i] {
			grad := 0.0
			if i < len(delta) && j < len(input) {
				grad = delta[i] * input[j]
			}
			t.Gradients[i][j] = grad
		}
	}
}

// computeStandardGradients computes gradients in standard way.
func (t *InMemTrainer) computeStandardGradients(delta, input []float64) {
	t.computeOuterProductGradients(delta, input)
}

// Update applies weight updates.
func (t *InMemTrainer) Update() {
	for i := range t.Weights {
		for j := range t.Weights[i] {
			// Momentum update
			t.Velocity[i][j] = t.config.Momentum*t.Velocity[i][j] -
				t.config.LearningRate*t.Gradients[i][j]

			// Weight update with decay
			update := t.Velocity[i][j] - t.config.WeightDecay*t.Weights[i][j]

			// Quantize update to hardware granularity
			if t.config.UpdateGranularity > 0 {
				update = math.Round(update/t.config.UpdateGranularity) *
					t.config.UpdateGranularity
			}

			// Apply update
			t.Weights[i][j] += update

			// Clamp to conductance range
			t.Weights[i][j] = math.Max(t.config.MinConductance,
				math.Min(t.config.MaxConductance, t.Weights[i][j]))
		}
	}

	// Reprogram crossbar with updated weights
	t.crossbar.Program(t.Weights)
	t.Iterations++
}

// TrainStep performs one training step (forward + backward + update).
func (t *InMemTrainer) TrainStep(input, target []float64) float64 {
	// Forward pass
	output := t.Forward(input)

	// Compute loss and output gradient (MSE)
	loss := 0.0
	outputGrad := make([]float64, len(output))
	for i, out := range output {
		if i < len(target) {
			diff := out - target[i]
			loss += diff * diff
			outputGrad[i] = 2.0 * diff / float64(len(target))
		}
	}
	loss /= float64(len(target))

	// Backward pass
	t.Backward(outputGrad)

	// Update weights
	t.Update()

	t.Loss = loss
	return loss
}

// ============================================================================
// Hardware-Aware Training (HAT)
// ============================================================================

// HATConfig configures hardware-aware training.
type HATConfig struct {
	// Noise injection during training
	TrainingNoiseScale float64 // Scale factor for training noise
	UseSTEGradients    bool    // Use STE for non-differentiable ops

	// Quantization-aware training
	QuantizationBits   int
	LearnQuantParams   bool // Learn scale/zero-point

	// Non-ideality models
	SimulateD2D        bool
	SimulateIRDrop     bool
	SimulateRetention  bool

	// Robustness
	AdversarialNoise   float64 // Adversarial noise for robustness
	DropoutRate        float64
}

// DefaultHATConfig returns default HAT configuration.
func DefaultHATConfig() *HATConfig {
	return &HATConfig{
		TrainingNoiseScale: 1.0,
		UseSTEGradients:    true,
		QuantizationBits:   8,
		LearnQuantParams:   false,
		SimulateD2D:        true,
		SimulateIRDrop:     true,
		SimulateRetention:  false,
		AdversarialNoise:   0.0,
		DropoutRate:        0.0,
	}
}

// HATTrainer implements hardware-aware training.
type HATTrainer struct {
	config         *HATConfig
	noiseSimulator *NoiseSimulator
	baseTrainer    *InMemTrainer

	// Learned quantization parameters
	WeightScale    float64
	WeightZero     float64
	ActScale       float64
	ActZero        float64

	// Statistics
	NoiseAcc       float64 // Noisy accuracy
	CleanAcc       float64 // Clean accuracy
	RobustnessGap  float64
}

// NewHATTrainer creates a new HAT trainer.
func NewHATTrainer(rows, cols int, config *HATConfig, noiseConfig *NoiseConfig) *HATTrainer {
	// Configure base trainer for noise-aware training
	trainConfig := DefaultInMemTrainingConfig()
	trainConfig.InjectTrainingNoise = true
	trainConfig.NoiseConfig = noiseConfig

	return &HATTrainer{
		config:         config,
		noiseSimulator: NewNoiseSimulator(noiseConfig),
		baseTrainer:    NewInMemTrainer(rows, cols, trainConfig),
		WeightScale:    1.0,
		ActScale:       1.0,
	}
}

// ForwardWithNoise performs forward pass with scaled noise.
func (t *HATTrainer) ForwardWithNoise(input []float64, training bool) []float64 {
	// Apply activation quantization
	quantInput := t.quantizeActivations(input)

	// Apply dropout during training
	if training && t.config.DropoutRate > 0 {
		quantInput = t.applyDropout(quantInput)
	}

	// Forward through base trainer (already has noise)
	output := t.baseTrainer.Forward(quantInput)

	// Scale noise during training
	if training {
		for i := range output {
			noise := rand.NormFloat64() * t.config.TrainingNoiseScale * 0.01
			output[i] += noise * math.Abs(output[i])
		}
	}

	return output
}

// quantizeActivations applies quantization-aware training.
func (t *HATTrainer) quantizeActivations(x []float64) []float64 {
	if t.config.QuantizationBits == 0 {
		return x
	}

	levels := float64(int(1) << t.config.QuantizationBits)
	result := make([]float64, len(x))

	for i, val := range x {
		// Quantize
		q := math.Round((val-t.ActZero)/t.ActScale*levels) / levels * t.ActScale + t.ActZero

		if t.config.UseSTEGradients {
			// STE: forward uses quantized, backward uses original
			result[i] = q
		} else {
			result[i] = q
		}
	}

	return result
}

// applyDropout applies dropout regularization.
func (t *HATTrainer) applyDropout(x []float64) []float64 {
	result := make([]float64, len(x))
	scale := 1.0 / (1.0 - t.config.DropoutRate)

	for i, val := range x {
		if rand.Float64() > t.config.DropoutRate {
			result[i] = val * scale
		}
	}

	return result
}

// ComputeSTEGradient computes gradient using Straight-Through Estimator.
// Forward: apply non-differentiable operation
// Backward: pass gradient through as identity
func (t *HATTrainer) ComputeSTEGradient(forwardOut, gradOutput []float64) []float64 {
	// STE passes gradient through unchanged for quantization
	return gradOutput
}

// CalibrateQuantization calibrates quantization parameters.
func (t *HATTrainer) CalibrateQuantization(calibrationData [][]float64) {
	// Find activation range
	minAct, maxAct := math.MaxFloat64, -math.MaxFloat64
	for _, data := range calibrationData {
		for _, val := range data {
			if val < minAct {
				minAct = val
			}
			if val > maxAct {
				maxAct = val
			}
		}
	}

	levels := float64(int(1) << t.config.QuantizationBits)
	t.ActScale = (maxAct - minAct) / levels
	if t.ActScale == 0 {
		t.ActScale = 1.0
	}
	t.ActZero = minAct

	// Calibrate weight quantization
	minW, maxW := math.MaxFloat64, -math.MaxFloat64
	for _, row := range t.baseTrainer.Weights {
		for _, w := range row {
			if w < minW {
				minW = w
			}
			if w > maxW {
				maxW = w
			}
		}
	}

	t.WeightScale = (maxW - minW) / levels
	if t.WeightScale == 0 {
		t.WeightScale = 1.0
	}
	t.WeightZero = minW
}

// ============================================================================
// Analog Gradient Computation
// ============================================================================

// AnalogGradientComputer computes gradients using analog circuits.
type AnalogGradientComputer struct {
	crossbar *NoisyCrossbar

	// Pulse parameters for weight update
	PulseDuration float64 // Pulse duration (ns)
	PulseAmplitude float64 // Pulse voltage (V)
	UpdateThreshold float64 // Minimum update threshold
}

// NewAnalogGradientComputer creates a new analog gradient computer.
func NewAnalogGradientComputer(crossbar *NoisyCrossbar) *AnalogGradientComputer {
	return &AnalogGradientComputer{
		crossbar:       crossbar,
		PulseDuration:  10.0,  // 10 ns
		PulseAmplitude: 1.0,   // 1 V
		UpdateThreshold: 0.001,
	}
}

// ComputeOuterProduct computes outer product ΔW = η × δ ⊗ x.
// This maps to simultaneous row/column pulse application.
func (agc *AnalogGradientComputer) ComputeOuterProduct(delta, input []float64, lr float64) [][]float64 {
	rows := len(delta)
	cols := len(input)
	gradients := make([][]float64, rows)

	for i := range gradients {
		gradients[i] = make([]float64, cols)
		for j := range gradients[i] {
			// Outer product: each element is product of row and column signals
			grad := lr * delta[i] * input[j]

			// Apply threshold (minimum detectable update)
			if math.Abs(grad) < agc.UpdateThreshold {
				grad = 0
			}

			gradients[i][j] = grad
		}
	}

	return gradients
}

// ApplyPulseUpdate applies pulse-based weight update.
func (agc *AnalogGradientComputer) ApplyPulseUpdate(weights [][]float64, gradients [][]float64) [][]float64 {
	rows := len(weights)
	cols := len(weights[0])
	updated := make([][]float64, rows)

	for i := range updated {
		updated[i] = make([]float64, cols)
		for j := range updated[i] {
			grad := 0.0
			if i < len(gradients) && j < len(gradients[i]) {
				grad = gradients[i][j]
			}

			// Pulse-based update: ΔG ∝ V × t
			// conductance increment proportional to pulse duration × amplitude
			pulseEffect := grad * agc.PulseDuration * agc.PulseAmplitude * 1e-9

			updated[i][j] = weights[i][j] + pulseEffect
		}
	}

	return updated
}

// ============================================================================
// On-Device Learning Support
// ============================================================================

// OnDeviceLearner supports incremental on-device learning.
type OnDeviceLearner struct {
	trainer        *InMemTrainer
	gradComputer   *AnalogGradientComputer

	// Streaming data buffer
	Buffer         [][]float64
	BufferLabels   [][]float64
	MaxBufferSize  int

	// Learning statistics
	SamplesLearned int64
	OnlineAccuracy float64
}

// NewOnDeviceLearner creates a new on-device learner.
func NewOnDeviceLearner(rows, cols int, config *InMemTrainingConfig) *OnDeviceLearner {
	trainer := NewInMemTrainer(rows, cols, config)

	return &OnDeviceLearner{
		trainer:       trainer,
		gradComputer:  NewAnalogGradientComputer(trainer.crossbar),
		Buffer:        make([][]float64, 0),
		BufferLabels:  make([][]float64, 0),
		MaxBufferSize: 100,
	}
}

// LearnSample performs one-shot learning on a single sample.
func (odl *OnDeviceLearner) LearnSample(input, target []float64) float64 {
	loss := odl.trainer.TrainStep(input, target)
	odl.SamplesLearned++
	return loss
}

// LearnBatch performs mini-batch learning.
func (odl *OnDeviceLearner) LearnBatch(inputs, targets [][]float64) float64 {
	var totalLoss float64

	for i := range inputs {
		if i < len(targets) {
			loss := odl.trainer.TrainStep(inputs[i], targets[i])
			totalLoss += loss
		}
	}

	return totalLoss / float64(len(inputs))
}

// BufferSample adds a sample to the buffer for later learning.
func (odl *OnDeviceLearner) BufferSample(input, target []float64) {
	if len(odl.Buffer) >= odl.MaxBufferSize {
		// Remove oldest
		odl.Buffer = odl.Buffer[1:]
		odl.BufferLabels = odl.BufferLabels[1:]
	}

	odl.Buffer = append(odl.Buffer, input)
	odl.BufferLabels = append(odl.BufferLabels, target)
}

// LearnFromBuffer trains on buffered samples.
func (odl *OnDeviceLearner) LearnFromBuffer() float64 {
	if len(odl.Buffer) == 0 {
		return 0
	}

	return odl.LearnBatch(odl.Buffer, odl.BufferLabels)
}

// Predict performs inference.
func (odl *OnDeviceLearner) Predict(input []float64) []float64 {
	return odl.trainer.Forward(input)
}

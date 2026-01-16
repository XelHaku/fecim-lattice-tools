// Package layers provides training loop utilities for CIM-based neural networks.
// Implements ANN-to-SNN conversion, hardware-aware training, and hybrid training strategies.
package layers

import (
	"math"
	"math/rand"
	"sort"
)

// ============================================================================
// ANN-TO-SNN CONVERSION
// ============================================================================

// ConversionConfig configures ANN-to-SNN conversion parameters.
type ConversionConfig struct {
	TimeSteps        int     // Number of simulation time steps
	ThresholdBalance bool    // Enable threshold balancing
	PercentileScale  float64 // Percentile for weight scaling (e.g., 99.9)
	RateCoding       bool    // Use rate coding for input encoding
	TemporalCoding   bool    // Use temporal coding (latency-based)
	BurstCoding      bool    // Use burst coding for higher precision
	AdaptiveThresh   bool    // Enable adaptive threshold neurons
	ChannelWise      bool    // Channel-wise threshold (for conv layers)
}

// DefaultConversionConfig returns standard conversion settings.
func DefaultConversionConfig() *ConversionConfig {
	return &ConversionConfig{
		TimeSteps:        100,
		ThresholdBalance: true,
		PercentileScale:  99.9,
		RateCoding:       true,
		TemporalCoding:   false,
		BurstCoding:      false,
		AdaptiveThresh:   false,
		ChannelWise:      true,
	}
}

// ANNtoSNNConverter converts trained ANN weights to SNN format.
type ANNtoSNNConverter struct {
	Config     *ConversionConfig
	Thresholds [][]float64 // Per-layer thresholds
	ScaleFactors []float64 // Per-layer scale factors
}

// NewANNtoSNNConverter creates a new converter.
func NewANNtoSNNConverter(config *ConversionConfig) *ANNtoSNNConverter {
	return &ANNtoSNNConverter{
		Config: config,
	}
}

// Convert transforms ANN layer weights to SNN-compatible format.
func (c *ANNtoSNNConverter) Convert(layers [][]float64) (*SNNModel, error) {
	model := &SNNModel{
		Weights:    make([][][]float64, len(layers)),
		Thresholds: make([][]float64, len(layers)),
		TimeSteps:  c.Config.TimeSteps,
	}

	// Step 1: Compute activation statistics for each layer
	c.Thresholds = make([][]float64, len(layers))
	c.ScaleFactors = make([]float64, len(layers))

	for i, layer := range layers {
		// Find maximum activation percentile
		maxAct := c.computeMaxActivation(layer)
		c.ScaleFactors[i] = maxAct

		// Threshold balancing: set threshold to max activation
		if c.Config.ThresholdBalance {
			numNeurons := len(layer)
			c.Thresholds[i] = make([]float64, numNeurons)
			for j := range c.Thresholds[i] {
				c.Thresholds[i][j] = maxAct
			}
		}

		model.Thresholds[i] = c.Thresholds[i]
	}

	// Step 2: Scale weights layer by layer
	for i, layer := range layers {
		scaledLayer := make([][]float64, len(layer))
		for j := 0; j < len(layer); j++ {
			scaledLayer[j] = make([]float64, 1)
			scale := 1.0
			if i > 0 {
				scale = c.ScaleFactors[i-1] / c.ScaleFactors[i]
			}
			scaledLayer[j][0] = layer[j] * scale
		}
		model.Weights[i] = scaledLayer
	}

	return model, nil
}

// computeMaxActivation finds the percentile-scaled maximum activation.
func (c *ANNtoSNNConverter) computeMaxActivation(weights []float64) float64 {
	absWeights := make([]float64, len(weights))
	for i, w := range weights {
		absWeights[i] = math.Abs(w)
	}
	sort.Float64s(absWeights)

	idx := int(float64(len(absWeights)) * c.Config.PercentileScale / 100.0)
	if idx >= len(absWeights) {
		idx = len(absWeights) - 1
	}
	return absWeights[idx]
}

// ThresholdBalance performs post-conversion threshold balancing.
func (c *ANNtoSNNConverter) ThresholdBalance(model *SNNModel, calibrationData [][]float64) {
	// Simulate forward pass to collect activation statistics
	for layerIdx := range model.Weights {
		maxActs := make([]float64, len(model.Thresholds[layerIdx]))

		for _, input := range calibrationData {
			// Forward through layer
			acts := c.simulateLayer(model.Weights[layerIdx], input)
			for j, act := range acts {
				if act > maxActs[j] {
					maxActs[j] = act
				}
			}
		}

		// Set threshold to max activation
		for j := range model.Thresholds[layerIdx] {
			if maxActs[j] > 0 {
				model.Thresholds[layerIdx][j] = maxActs[j]
			}
		}
	}
}

// simulateLayer performs a simple forward pass.
func (c *ANNtoSNNConverter) simulateLayer(weights [][]float64, input []float64) []float64 {
	output := make([]float64, len(weights))
	for i, w := range weights {
		sum := 0.0
		for j := range input {
			if j < len(w) {
				sum += w[j] * input[j]
			}
		}
		output[i] = math.Max(0, sum) // ReLU activation
	}
	return output
}

// SNNModel represents a converted SNN ready for spike-based simulation.
type SNNModel struct {
	Weights    [][][]float64
	Thresholds [][]float64
	TimeSteps  int
}

// ============================================================================
// SPIKE ENCODING
// ============================================================================

// SpikeEncoder converts continuous values to spike trains.
type SpikeEncoder struct {
	Config *ConversionConfig
}

// NewSpikeEncoder creates a new spike encoder.
func NewSpikeEncoder(config *ConversionConfig) *SpikeEncoder {
	return &SpikeEncoder{Config: config}
}

// RateEncode converts values to rate-coded spike trains.
func (e *SpikeEncoder) RateEncode(values []float64) [][]bool {
	spikes := make([][]bool, len(values))
	for i, v := range values {
		spikes[i] = make([]bool, e.Config.TimeSteps)
		prob := math.Max(0, math.Min(1, v))
		for t := 0; t < e.Config.TimeSteps; t++ {
			spikes[i][t] = rand.Float64() < prob
		}
	}
	return spikes
}

// TemporalEncode converts values to latency-coded spikes.
func (e *SpikeEncoder) TemporalEncode(values []float64) []int {
	latencies := make([]int, len(values))
	for i, v := range values {
		// Higher value = earlier spike (lower latency)
		if v > 0 {
			latency := int(float64(e.Config.TimeSteps) * (1.0 - math.Min(1, v)))
			latencies[i] = latency
		} else {
			latencies[i] = e.Config.TimeSteps // No spike
		}
	}
	return latencies
}

// BurstEncode converts values to burst-coded spikes (multiple spikes for higher precision).
func (e *SpikeEncoder) BurstEncode(values []float64, maxBurst int) []int {
	counts := make([]int, len(values))
	for i, v := range values {
		counts[i] = int(math.Max(0, math.Min(float64(maxBurst), v*float64(maxBurst))))
	}
	return counts
}

// ============================================================================
// CIM-AWARE TRAINING
// ============================================================================

// CIMTrainingConfig configures CIM-aware training parameters.
type CIMTrainingConfig struct {
	NoiseLevel       float64 // Conductance variation (0-0.1)
	QuantizationBits int     // Weight quantization bits
	ADCBits          int     // ADC precision bits
	DACBits          int     // DAC precision bits
	ClipMin          float64 // Weight clipping minimum
	ClipMax          float64 // Weight clipping maximum
	NonlineartiyFactor float64 // Programming nonlinearity
}

// DefaultCIMTrainingConfig returns standard CIM training settings.
func DefaultCIMTrainingConfig() *CIMTrainingConfig {
	return &CIMTrainingConfig{
		NoiseLevel:        0.05,
		QuantizationBits:  6,
		ADCBits:           6,
		DACBits:           8,
		ClipMin:           -1.0,
		ClipMax:           1.0,
		NonlineartiyFactor: 0.1,
	}
}

// CIMTrainer implements hardware-aware training.
type CIMTrainer struct {
	Config       *CIMTrainingConfig
	NoiseInjector *NoiseInjector
}

// NewCIMTrainer creates a new CIM-aware trainer.
func NewCIMTrainer(config *CIMTrainingConfig) *CIMTrainer {
	return &CIMTrainer{
		Config:       config,
		NoiseInjector: NewNoiseInjector(config),
	}
}

// NoiseInjector adds CIM-realistic noise during training.
type NoiseInjector struct {
	Config *CIMTrainingConfig
}

// NewNoiseInjector creates a new noise injector.
func NewNoiseInjector(config *CIMTrainingConfig) *NoiseInjector {
	return &NoiseInjector{Config: config}
}

// InjectNoise adds conductance variation noise to weights.
func (n *NoiseInjector) InjectNoise(weights [][]float64) [][]float64 {
	noisy := make([][]float64, len(weights))
	for i := range weights {
		noisy[i] = make([]float64, len(weights[i]))
		for j := range weights[i] {
			noise := rand.NormFloat64() * n.Config.NoiseLevel
			noisy[i][j] = weights[i][j] * (1 + noise)
		}
	}
	return noisy
}

// InjectADCNoise simulates ADC quantization noise.
func (n *NoiseInjector) InjectADCNoise(values []float64) []float64 {
	levels := math.Pow(2, float64(n.Config.ADCBits))
	quantized := make([]float64, len(values))
	for i, v := range values {
		// Quantize to ADC levels
		q := math.Round(v * levels) / levels
		// Add small quantization noise
		noise := (rand.Float64() - 0.5) / levels
		quantized[i] = q + noise
	}
	return quantized
}

// InjectDACNoise simulates DAC quantization on inputs.
func (n *NoiseInjector) InjectDACNoise(inputs []float64) []float64 {
	levels := math.Pow(2, float64(n.Config.DACBits))
	quantized := make([]float64, len(inputs))
	for i, v := range inputs {
		quantized[i] = math.Round(v*levels) / levels
	}
	return quantized
}

// ============================================================================
// DIRECT FEEDBACK ALIGNMENT (DFA)
// ============================================================================

// DFAConfig configures Direct Feedback Alignment training.
type DFAConfig struct {
	LearningRate   float64
	FeedbackScale  float64 // Random feedback matrix scale
	NoiseAwareness float64 // How much noise to inject during training
}

// DefaultDFAConfig returns standard DFA settings.
func DefaultDFAConfig() *DFAConfig {
	return &DFAConfig{
		LearningRate:   0.01,
		FeedbackScale:  0.1,
		NoiseAwareness: 0.05,
	}
}

// DFATrainer implements Direct Feedback Alignment for CIM.
type DFATrainer struct {
	Config          *DFAConfig
	FeedbackMatrices [][][]float64 // Random fixed feedback per layer
	NumLayers       int
	OutputSize      int
}

// NewDFATrainer creates a new DFA trainer.
func NewDFATrainer(config *DFAConfig, layerSizes []int, outputSize int) *DFATrainer {
	trainer := &DFATrainer{
		Config:    config,
		NumLayers: len(layerSizes),
		OutputSize: outputSize,
	}

	// Initialize random feedback matrices
	trainer.FeedbackMatrices = make([][][]float64, len(layerSizes))
	for l, size := range layerSizes {
		trainer.FeedbackMatrices[l] = make([][]float64, outputSize)
		for i := 0; i < outputSize; i++ {
			trainer.FeedbackMatrices[l][i] = make([]float64, size)
			for j := 0; j < size; j++ {
				trainer.FeedbackMatrices[l][i][j] = (rand.Float64() - 0.5) * 2 * config.FeedbackScale
			}
		}
	}

	return trainer
}

// ComputeGradient computes weight gradient using DFA.
func (t *DFATrainer) ComputeGradient(layerIdx int, error []float64, activations []float64) [][]float64 {
	B := t.FeedbackMatrices[layerIdx]

	// Compute feedback signal: B^T * error
	feedback := make([]float64, len(activations))
	for j := range activations {
		sum := 0.0
		for i := range error {
			if i < len(B) && j < len(B[i]) {
				sum += B[i][j] * error[i]
			}
		}
		feedback[j] = sum
	}

	// Gradient: feedback * activation^T
	grad := make([][]float64, len(feedback))
	for i := range feedback {
		grad[i] = make([]float64, len(activations))
		for j := range activations {
			grad[i][j] = t.Config.LearningRate * feedback[i] * activations[j]
		}
	}

	return grad
}

// ============================================================================
// HYBRID TRAINING
// ============================================================================

// HybridTrainingConfig configures two-phase training.
type HybridTrainingConfig struct {
	PretrainEpochs    int
	FineTuneEpochs    int
	PretrainLR        float64
	FineTuneLR        float64
	NoiseSchedule     []float64 // Noise level per fine-tune epoch
	TransferQuantize  bool      // Quantize when transferring to on-chip
}

// DefaultHybridConfig returns standard hybrid training settings.
func DefaultHybridConfig() *HybridTrainingConfig {
	return &HybridTrainingConfig{
		PretrainEpochs:   100,
		FineTuneEpochs:   20,
		PretrainLR:       0.001,
		FineTuneLR:       0.0001,
		NoiseSchedule:    []float64{0.01, 0.02, 0.03, 0.04, 0.05},
		TransferQuantize: true,
	}
}

// HybridTrainer implements off-chip pretrain + on-chip fine-tune.
type HybridTrainer struct {
	Config      *HybridTrainingConfig
	CIMConfig   *CIMTrainingConfig
	DFATrainer  *DFATrainer
}

// NewHybridTrainer creates a new hybrid trainer.
func NewHybridTrainer(config *HybridTrainingConfig, cimConfig *CIMTrainingConfig,
	layerSizes []int, outputSize int) *HybridTrainer {
	dfaConfig := &DFAConfig{
		LearningRate:   config.FineTuneLR,
		FeedbackScale:  0.1,
		NoiseAwareness: 0.05,
	}
	return &HybridTrainer{
		Config:     config,
		CIMConfig:  cimConfig,
		DFATrainer: NewDFATrainer(dfaConfig, layerSizes, outputSize),
	}
}

// TransferToCIM prepares weights for on-chip fine-tuning.
func (t *HybridTrainer) TransferToCIM(weights [][][]float64) [][][]float64 {
	transferred := make([][][]float64, len(weights))
	for l := range weights {
		transferred[l] = make([][]float64, len(weights[l]))
		for i := range weights[l] {
			transferred[l][i] = make([]float64, len(weights[l][i]))
			for j := range weights[l][i] {
				w := weights[l][i][j]
				// Clip to valid range
				w = math.Max(t.CIMConfig.ClipMin, math.Min(t.CIMConfig.ClipMax, w))
				// Quantize if enabled
				if t.Config.TransferQuantize {
					levels := math.Pow(2, float64(t.CIMConfig.QuantizationBits))
					w = math.Round(w*levels) / levels
				}
				transferred[l][i][j] = w
			}
		}
	}
	return transferred
}

// ============================================================================
// CONTINUAL LEARNING
// ============================================================================

// EWCConfig configures Elastic Weight Consolidation.
type EWCConfig struct {
	Lambda       float64 // Regularization strength
	NumSamples   int     // Samples for Fisher computation
}

// DefaultEWCConfig returns standard EWC settings.
func DefaultEWCConfig() *EWCConfig {
	return &EWCConfig{
		Lambda:     1000.0,
		NumSamples: 200,
	}
}

// EWCRegularizer implements Elastic Weight Consolidation for continual learning.
type EWCRegularizer struct {
	Config      *EWCConfig
	OptimalWeights [][][]float64 // Weights after task completion
	FisherInfo     [][][]float64 // Fisher information per weight
}

// NewEWCRegularizer creates a new EWC regularizer.
func NewEWCRegularizer(config *EWCConfig) *EWCRegularizer {
	return &EWCRegularizer{
		Config: config,
	}
}

// SaveTaskWeights stores optimal weights for a completed task.
func (e *EWCRegularizer) SaveTaskWeights(weights [][][]float64) {
	e.OptimalWeights = make([][][]float64, len(weights))
	for l := range weights {
		e.OptimalWeights[l] = make([][]float64, len(weights[l]))
		for i := range weights[l] {
			e.OptimalWeights[l][i] = make([]float64, len(weights[l][i]))
			copy(e.OptimalWeights[l][i], weights[l][i])
		}
	}
}

// ComputeFisher estimates Fisher information from gradients.
func (e *EWCRegularizer) ComputeFisher(gradients [][][][]float64) {
	if len(gradients) == 0 {
		return
	}

	numLayers := len(gradients[0])
	e.FisherInfo = make([][][]float64, numLayers)

	for l := 0; l < numLayers; l++ {
		rows := len(gradients[0][l])
		cols := len(gradients[0][l][0])
		e.FisherInfo[l] = make([][]float64, rows)
		for i := 0; i < rows; i++ {
			e.FisherInfo[l][i] = make([]float64, cols)
		}

		// Fisher ≈ E[grad²]
		for _, grad := range gradients {
			for i := range grad[l] {
				for j := range grad[l][i] {
					e.FisherInfo[l][i][j] += grad[l][i][j] * grad[l][i][j]
				}
			}
		}

		// Normalize
		n := float64(len(gradients))
		for i := range e.FisherInfo[l] {
			for j := range e.FisherInfo[l][i] {
				e.FisherInfo[l][i][j] /= n
			}
		}
	}
}

// ComputePenalty calculates EWC regularization penalty.
func (e *EWCRegularizer) ComputePenalty(currentWeights [][][]float64) float64 {
	if e.OptimalWeights == nil || e.FisherInfo == nil {
		return 0.0
	}

	penalty := 0.0
	for l := range currentWeights {
		for i := range currentWeights[l] {
			for j := range currentWeights[l][i] {
				diff := currentWeights[l][i][j] - e.OptimalWeights[l][i][j]
				penalty += e.FisherInfo[l][i][j] * diff * diff
			}
		}
	}

	return 0.5 * e.Config.Lambda * penalty
}

// ============================================================================
// TRAINING LOOP
// ============================================================================

// TrainingLoopConfig configures the main training loop.
type TrainingLoopConfig struct {
	Epochs          int
	BatchSize       int
	LearningRate    float64
	Momentum        float64
	WeightDecay     float64
	CIMNoise        bool
	CIMConfig       *CIMTrainingConfig
	UseDFA          bool
	DFAConfig       *DFAConfig
	UseEWC          bool
	EWCConfig       *EWCConfig
	GradientClip    float64
	EarlyStopPatience int
	ValidationSplit float64
}

// DefaultTrainingLoopConfig returns standard training settings.
func DefaultTrainingLoopConfig() *TrainingLoopConfig {
	return &TrainingLoopConfig{
		Epochs:            100,
		BatchSize:         32,
		LearningRate:      0.001,
		Momentum:          0.9,
		WeightDecay:       0.0001,
		CIMNoise:          true,
		CIMConfig:         DefaultCIMTrainingConfig(),
		UseDFA:            false,
		DFAConfig:         DefaultDFAConfig(),
		UseEWC:            false,
		EWCConfig:         DefaultEWCConfig(),
		GradientClip:      1.0,
		EarlyStopPatience: 10,
		ValidationSplit:   0.1,
	}
}

// TrainingLoop manages the complete training process.
type TrainingLoop struct {
	Config        *TrainingLoopConfig
	Weights       [][][]float64
	Velocities    [][][]float64 // For momentum
	BestWeights   [][][]float64
	BestLoss      float64
	NoiseInjector *NoiseInjector
	DFATrainer    *DFATrainer
	EWCReg        *EWCRegularizer
	History       *TrainingHistory
}

// TrainingHistory records training metrics.
type TrainingHistory struct {
	TrainLoss   []float64
	ValLoss     []float64
	TrainAcc    []float64
	ValAcc      []float64
	LearningRates []float64
}

// NewTrainingLoop creates a new training loop.
func NewTrainingLoop(config *TrainingLoopConfig, layerSizes []int, outputSize int) *TrainingLoop {
	loop := &TrainingLoop{
		Config:   config,
		BestLoss: math.MaxFloat64,
		History:  &TrainingHistory{},
	}

	if config.CIMNoise {
		loop.NoiseInjector = NewNoiseInjector(config.CIMConfig)
	}

	if config.UseDFA {
		loop.DFATrainer = NewDFATrainer(config.DFAConfig, layerSizes, outputSize)
	}

	if config.UseEWC {
		loop.EWCReg = NewEWCRegularizer(config.EWCConfig)
	}

	return loop
}

// InitializeWeights sets up initial weight matrices.
func (t *TrainingLoop) InitializeWeights(layerSizes []int) {
	numLayers := len(layerSizes) - 1
	t.Weights = make([][][]float64, numLayers)
	t.Velocities = make([][][]float64, numLayers)

	for l := 0; l < numLayers; l++ {
		inSize := layerSizes[l]
		outSize := layerSizes[l+1]

		t.Weights[l] = make([][]float64, outSize)
		t.Velocities[l] = make([][]float64, outSize)

		// Xavier initialization
		scale := math.Sqrt(2.0 / float64(inSize+outSize))
		for i := 0; i < outSize; i++ {
			t.Weights[l][i] = make([]float64, inSize)
			t.Velocities[l][i] = make([]float64, inSize)
			for j := 0; j < inSize; j++ {
				t.Weights[l][i][j] = rand.NormFloat64() * scale
			}
		}
	}
}

// Forward performs forward pass with optional noise injection.
func (t *TrainingLoop) Forward(input []float64) ([][]float64, []float64) {
	activations := make([][]float64, len(t.Weights)+1)
	activations[0] = input

	current := input
	for l, weights := range t.Weights {
		// Optionally inject noise for CIM-aware training
		w := weights
		if t.Config.CIMNoise && t.NoiseInjector != nil {
			w = t.NoiseInjector.InjectNoise(weights)
		}

		// Matrix-vector multiplication
		output := make([]float64, len(w))
		for i, row := range w {
			sum := 0.0
			for j, wj := range row {
				if j < len(current) {
					sum += wj * current[j]
				}
			}
			output[i] = math.Max(0, sum) // ReLU
		}

		activations[l+1] = output
		current = output
	}

	return activations, current
}

// ComputeLoss calculates cross-entropy loss.
func (t *TrainingLoop) ComputeLoss(output []float64, target int) float64 {
	// Softmax
	maxOut := output[0]
	for _, v := range output {
		if v > maxOut {
			maxOut = v
		}
	}

	expSum := 0.0
	for _, v := range output {
		expSum += math.Exp(v - maxOut)
	}

	logProb := output[target] - maxOut - math.Log(expSum)
	loss := -logProb

	// Add EWC penalty if enabled
	if t.Config.UseEWC && t.EWCReg != nil {
		loss += t.EWCReg.ComputePenalty(t.Weights)
	}

	return loss
}

// Backward computes gradients (simplified backprop or DFA).
func (t *TrainingLoop) Backward(activations [][]float64, target int) [][][]float64 {
	output := activations[len(activations)-1]

	// Softmax gradients
	maxOut := output[0]
	for _, v := range output {
		if v > maxOut {
			maxOut = v
		}
	}

	probs := make([]float64, len(output))
	expSum := 0.0
	for i, v := range output {
		probs[i] = math.Exp(v - maxOut)
		expSum += probs[i]
	}
	for i := range probs {
		probs[i] /= expSum
	}

	// Output error
	error := make([]float64, len(output))
	for i := range error {
		error[i] = probs[i]
		if i == target {
			error[i] -= 1.0
		}
	}

	gradients := make([][][]float64, len(t.Weights))

	if t.Config.UseDFA && t.DFATrainer != nil {
		// Direct Feedback Alignment
		for l := range t.Weights {
			gradients[l] = t.DFATrainer.ComputeGradient(l, error, activations[l])
		}
	} else {
		// Standard backpropagation
		delta := error
		for l := len(t.Weights) - 1; l >= 0; l-- {
			act := activations[l]
			gradients[l] = make([][]float64, len(t.Weights[l]))

			for i := range t.Weights[l] {
				gradients[l][i] = make([]float64, len(t.Weights[l][i]))
				for j := range t.Weights[l][i] {
					if j < len(act) {
						gradients[l][i][j] = delta[i] * act[j]
					}
				}
			}

			// Propagate delta (if not first layer)
			if l > 0 {
				newDelta := make([]float64, len(t.Weights[l][0]))
				for j := range newDelta {
					sum := 0.0
					for i := range t.Weights[l] {
						if j < len(t.Weights[l][i]) {
							sum += t.Weights[l][i][j] * delta[i]
						}
					}
					// ReLU derivative
					if activations[l][j] > 0 {
						newDelta[j] = sum
					}
				}
				delta = newDelta
			}
		}
	}

	return gradients
}

// UpdateWeights applies gradients with momentum and weight decay.
func (t *TrainingLoop) UpdateWeights(gradients [][][]float64) {
	for l := range t.Weights {
		for i := range t.Weights[l] {
			for j := range t.Weights[l][i] {
				if i < len(gradients[l]) && j < len(gradients[l][i]) {
					grad := gradients[l][i][j]

					// Gradient clipping
					if t.Config.GradientClip > 0 {
						grad = math.Max(-t.Config.GradientClip,
							math.Min(t.Config.GradientClip, grad))
					}

					// Weight decay
					grad += t.Config.WeightDecay * t.Weights[l][i][j]

					// Momentum
					t.Velocities[l][i][j] = t.Config.Momentum*t.Velocities[l][i][j] +
						t.Config.LearningRate*grad

					// Update
					t.Weights[l][i][j] -= t.Velocities[l][i][j]
				}
			}
		}
	}
}

// EvaluateAccuracy computes classification accuracy.
func (t *TrainingLoop) EvaluateAccuracy(inputs [][]float64, labels []int) float64 {
	correct := 0
	for i, input := range inputs {
		_, output := t.Forward(input)

		// Find argmax
		maxIdx := 0
		maxVal := output[0]
		for j, v := range output {
			if v > maxVal {
				maxVal = v
				maxIdx = j
			}
		}

		if maxIdx == labels[i] {
			correct++
		}
	}

	return float64(correct) / float64(len(inputs))
}

// ============================================================================
// MEMORY TECHNOLOGY COMPARISON
// ============================================================================

// MemoryTechnology represents different NVM technologies.
type MemoryTechnology int

const (
	MemoryFeFET MemoryTechnology = iota
	MemoryRRAM
	MemoryPCM
	MemoryMRAM
	MemorySRAM
)

// MemoryCharacteristics defines technology-specific parameters.
type MemoryCharacteristics struct {
	Technology     MemoryTechnology
	Name           string
	Endurance      float64 // Cycles
	WriteEnergy    float64 // Joules
	ReadEnergy     float64 // Joules
	WriteTime      float64 // Seconds
	ReadTime       float64 // Seconds
	OnOffRatio     float64 // Conductance ratio
	AnalogLevels   int     // Number of distinct states
	RetentionYears float64 // Data retention
	AreaFactor     float64 // Relative to SRAM (1.0)
	CIMSuitability float64 // Score 0-1
}

// GetMemoryCharacteristics returns typical values for each technology.
func GetMemoryCharacteristics(tech MemoryTechnology) *MemoryCharacteristics {
	switch tech {
	case MemoryFeFET:
		return &MemoryCharacteristics{
			Technology:     MemoryFeFET,
			Name:           "FeFET (HfO₂/ZrO₂)",
			Endurance:      1e12,
			WriteEnergy:    10e-15,  // 10 fJ
			ReadEnergy:     1e-15,   // 1 fJ
			WriteTime:      10e-9,   // 10 ns
			ReadTime:       1e-9,    // 1 ns
			OnOffRatio:     1e4,
			AnalogLevels:   64,      // 6-bit
			RetentionYears: 10,
			AreaFactor:     0.2,
			CIMSuitability: 0.95,
		}
	case MemoryRRAM:
		return &MemoryCharacteristics{
			Technology:     MemoryRRAM,
			Name:           "RRAM (HfOx)",
			Endurance:      1e6,
			WriteEnergy:    100e-15, // 100 fJ
			ReadEnergy:     10e-15,  // 10 fJ
			WriteTime:      10e-9,   // 10 ns
			ReadTime:       10e-9,   // 10 ns
			OnOffRatio:     100,
			AnalogLevels:   16,      // 4-bit
			RetentionYears: 10,
			AreaFactor:     0.1,
			CIMSuitability: 0.75,
		}
	case MemoryPCM:
		return &MemoryCharacteristics{
			Technology:     MemoryPCM,
			Name:           "PCM (GST)",
			Endurance:      1e8,
			WriteEnergy:    10e-12,  // 10 pJ
			ReadEnergy:     100e-15, // 100 fJ
			WriteTime:      100e-9,  // 100 ns
			ReadTime:       10e-9,   // 10 ns
			OnOffRatio:     1e3,
			AnalogLevels:   32,      // 5-bit
			RetentionYears: 10,
			AreaFactor:     0.15,
			CIMSuitability: 0.70,
		}
	case MemoryMRAM:
		return &MemoryCharacteristics{
			Technology:     MemoryMRAM,
			Name:           "STT-MRAM",
			Endurance:      1e15,
			WriteEnergy:    100e-15, // 100 fJ
			ReadEnergy:     10e-15,  // 10 fJ
			WriteTime:      10e-9,   // 10 ns
			ReadTime:       1e-9,    // 1 ns
			OnOffRatio:     3,       // TMR ratio
			AnalogLevels:   2,       // Binary only
			RetentionYears: 10,
			AreaFactor:     0.3,
			CIMSuitability: 0.50,
		}
	case MemorySRAM:
		return &MemoryCharacteristics{
			Technology:     MemorySRAM,
			Name:           "SRAM (Baseline)",
			Endurance:      1e18,    // Effectively unlimited
			WriteEnergy:    1e-15,   // 1 fJ
			ReadEnergy:     1e-15,   // 1 fJ
			WriteTime:      1e-9,    // 1 ns
			ReadTime:       1e-9,    // 1 ns
			OnOffRatio:     1e6,
			AnalogLevels:   2,       // Binary
			RetentionYears: 0,       // Volatile
			AreaFactor:     1.0,
			CIMSuitability: 0.60,
		}
	default:
		return GetMemoryCharacteristics(MemoryFeFET)
	}
}

// CompareMemoryTechnologies returns comparison of all technologies.
func CompareMemoryTechnologies() []*MemoryCharacteristics {
	return []*MemoryCharacteristics{
		GetMemoryCharacteristics(MemoryFeFET),
		GetMemoryCharacteristics(MemoryRRAM),
		GetMemoryCharacteristics(MemoryPCM),
		GetMemoryCharacteristics(MemoryMRAM),
		GetMemoryCharacteristics(MemorySRAM),
	}
}

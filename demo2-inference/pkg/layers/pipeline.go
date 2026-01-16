// Package layers provides neural network layer implementations for crossbar-based CIM.
// pipeline.go implements end-to-end inference pipeline for crossbar-based neural networks.
//
// The pipeline handles:
// - Model loading from checkpoints
// - Input preprocessing
// - Layer-by-layer forward pass with CIM simulation
// - Output post-processing
// - Accuracy evaluation
//
// CIM-specific features:
// - Crossbar array simulation with noise
// - Quantization at each layer
// - ADC/DAC conversion simulation
// - Multi-tile matrix operations

package layers

import (
	"fmt"
	"math"
	"time"
)

// InferencePipeline orchestrates neural network inference
type InferencePipeline struct {
	Model        *ModelCheckpoint
	Layers       []*PipelineLayer
	Config       *PipelineConfig
	Quantizer    *Quantizer
	CrossbarSim  *CrossbarSimulator
	Stats        *InferenceStats
}

// PipelineLayer represents a layer in the inference pipeline
type PipelineLayer struct {
	Name         string
	Type         string
	Weights      [][]float64
	Biases       []float64
	Activation   string
	InputShape   []int
	OutputShape  []int
	QuantParams  *QuantizationParams
	Conductances [][]float64 // For CIM simulation
}

// PipelineConfig configures the inference pipeline
type PipelineConfig struct {
	UseCIMSimulation  bool
	QuantizeInputs    bool
	QuantizeWeights   bool
	InputBits         int
	WeightBits        int
	ADCBits           int
	NoiseLevel        float64
	CrossbarSize      int
	BatchSize         int
	Verbose           bool
}

// DefaultPipelineConfig returns default pipeline configuration
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		UseCIMSimulation: true,
		QuantizeInputs:   true,
		QuantizeWeights:  true,
		InputBits:        8,
		WeightBits:       6,
		ADCBits:          8,
		NoiseLevel:       0.02,
		CrossbarSize:     64,
		BatchSize:        32,
		Verbose:          false,
	}
}

// IdealPipelineConfig returns config without CIM effects
func IdealPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		UseCIMSimulation: false,
		QuantizeInputs:   false,
		QuantizeWeights:  false,
		InputBits:        32,
		WeightBits:       32,
		ADCBits:          32,
		NoiseLevel:       0.0,
		CrossbarSize:     64,
		BatchSize:        32,
		Verbose:          false,
	}
}

// InferenceStats tracks inference statistics
type InferenceStats struct {
	TotalInferences  int
	TotalLatencyMs   float64
	AvgLatencyMs     float64
	LayerLatencies   map[string]float64
	MACOperations    int64
	MemoryAccessesKB float64
}

// NewInferencePipeline creates a new inference pipeline
func NewInferencePipeline(model *ModelCheckpoint, config *PipelineConfig) *InferencePipeline {
	if config == nil {
		config = DefaultPipelineConfig()
	}

	pipe := &InferencePipeline{
		Model:  model,
		Config: config,
		Stats: &InferenceStats{
			LayerLatencies: make(map[string]float64),
		},
	}

	// Initialize quantizer
	qconfig := &QuantizationConfig{
		WeightBits:     config.WeightBits,
		ActivationBits: config.InputBits,
		Mode:           QuantModeSymmetric,
	}
	pipe.Quantizer = NewQuantizer(qconfig)

	// Initialize crossbar simulator
	if config.UseCIMSimulation {
		pipe.CrossbarSim = NewCrossbarSimulator(config.CrossbarSize, config.NoiseLevel)
	}

	// Build pipeline layers from model
	pipe.buildLayers()

	return pipe
}

// buildLayers converts model checkpoints to pipeline layers
func (p *InferencePipeline) buildLayers() {
	p.Layers = make([]*PipelineLayer, len(p.Model.Layers))

	for i, layerCkpt := range p.Model.Layers {
		layer := &PipelineLayer{
			Name:       layerCkpt.Name,
			Type:       layerCkpt.Type,
			Weights:    layerCkpt.Weights,
			Biases:     layerCkpt.Biases,
			Activation: "relu", // Default, could be extracted from type
		}

		if len(layerCkpt.Shape) >= 2 {
			layer.InputShape = []int{layerCkpt.Shape[1]}
			layer.OutputShape = []int{layerCkpt.Shape[0]}
		}

		// Compute quantization parameters
		if p.Config.QuantizeWeights && len(layer.Weights) > 0 {
			layer.QuantParams = p.Quantizer.ComputeWeightParams(layer.Weights)
		}

		// Map to conductances for CIM
		if p.Config.UseCIMSimulation && len(layer.Weights) > 0 {
			cq := NewConductanceQuantizer(0.1, 1.0, p.Config.WeightBits)
			layer.Conductances, _ = cq.MapToConductance(layer.Weights)
		}

		p.Layers[i] = layer
	}
}

// Forward performs forward pass on input
func (p *InferencePipeline) Forward(input []float64) []float64 {
	startTime := time.Now()

	x := input

	// Quantize input if configured
	if p.Config.QuantizeInputs {
		x = p.quantizeActivations(x)
	}

	// Process each layer
	for _, layer := range p.Layers {
		layerStart := time.Now()

		if len(layer.Weights) > 0 {
			if p.Config.UseCIMSimulation {
				x = p.crossbarMVM(x, layer)
			} else {
				x = p.idealMVM(x, layer)
			}
		}

		// Add bias
		if len(layer.Biases) > 0 {
			for i := range x {
				if i < len(layer.Biases) {
					x[i] += layer.Biases[i]
				}
			}
		}

		// Apply activation
		x = p.applyActivation(x, layer.Activation)

		layerLatency := time.Since(layerStart).Seconds() * 1000
		p.Stats.LayerLatencies[layer.Name] = layerLatency
	}

	totalLatency := time.Since(startTime).Seconds() * 1000
	p.Stats.TotalInferences++
	p.Stats.TotalLatencyMs += totalLatency
	p.Stats.AvgLatencyMs = p.Stats.TotalLatencyMs / float64(p.Stats.TotalInferences)

	return x
}

// ForwardBatch performs forward pass on batch
func (p *InferencePipeline) ForwardBatch(inputs [][]float64) [][]float64 {
	outputs := make([][]float64, len(inputs))
	for i, input := range inputs {
		outputs[i] = p.Forward(input)
	}
	return outputs
}

// quantizeActivations quantizes input activations
func (p *InferencePipeline) quantizeActivations(x []float64) []float64 {
	levels := float64(int(1) << p.Config.InputBits)
	quantized := make([]float64, len(x))

	for i, val := range x {
		// Assume input in [0, 1]
		level := math.Round(val * (levels - 1))
		if level < 0 {
			level = 0
		}
		if level >= levels {
			level = levels - 1
		}
		quantized[i] = level / (levels - 1)
	}
	return quantized
}

// idealMVM performs ideal matrix-vector multiply
func (p *InferencePipeline) idealMVM(x []float64, layer *PipelineLayer) []float64 {
	rows := len(layer.Weights)
	cols := len(layer.Weights[0])

	output := make([]float64, rows)
	for i := 0; i < rows; i++ {
		sum := 0.0
		for j := 0; j < cols && j < len(x); j++ {
			sum += layer.Weights[i][j] * x[j]
		}
		output[i] = sum
	}

	// Count MACs
	p.Stats.MACOperations += int64(rows * cols)

	return output
}

// crossbarMVM performs CIM-simulated matrix-vector multiply
func (p *InferencePipeline) crossbarMVM(x []float64, layer *PipelineLayer) []float64 {
	if p.CrossbarSim == nil {
		return p.idealMVM(x, layer)
	}

	return p.CrossbarSim.MVM(x, layer.Conductances, p.Config.ADCBits)
}

// applyActivation applies activation function
func (p *InferencePipeline) applyActivation(x []float64, activation string) []float64 {
	result := make([]float64, len(x))

	switch activation {
	case "relu":
		for i, val := range x {
			if val > 0 {
				result[i] = val
			}
		}
	case "sigmoid":
		for i, val := range x {
			result[i] = 1.0 / (1.0 + math.Exp(-val))
		}
	case "tanh":
		for i, val := range x {
			result[i] = math.Tanh(val)
		}
	case "softmax":
		maxVal := x[0]
		for _, val := range x {
			if val > maxVal {
				maxVal = val
			}
		}
		sum := 0.0
		for i, val := range x {
			result[i] = math.Exp(val - maxVal)
			sum += result[i]
		}
		for i := range result {
			result[i] /= sum
		}
	default:
		copy(result, x)
	}

	return result
}

// Predict returns the predicted class
func (p *InferencePipeline) Predict(input []float64) int {
	output := p.Forward(input)

	// Find argmax
	maxIdx := 0
	maxVal := output[0]
	for i, val := range output {
		if val > maxVal {
			maxVal = val
			maxIdx = i
		}
	}
	return maxIdx
}

// PredictBatch returns predicted classes for batch
func (p *InferencePipeline) PredictBatch(inputs [][]float64) []int {
	predictions := make([]int, len(inputs))
	for i, input := range inputs {
		predictions[i] = p.Predict(input)
	}
	return predictions
}

// Evaluate evaluates accuracy on dataset
func (p *InferencePipeline) Evaluate(inputs [][]float64, labels []int) *EvaluationResult {
	startTime := time.Now()

	predictions := p.PredictBatch(inputs)

	correct := 0
	for i, pred := range predictions {
		if pred == labels[i] {
			correct++
		}
	}

	accuracy := float64(correct) / float64(len(labels))
	totalTime := time.Since(startTime).Seconds() * 1000

	// Compute confusion matrix
	numClasses := 0
	for _, l := range labels {
		if l >= numClasses {
			numClasses = l + 1
		}
	}

	confusion := make([][]int, numClasses)
	for i := range confusion {
		confusion[i] = make([]int, numClasses)
	}
	for i, pred := range predictions {
		if labels[i] < numClasses && pred < numClasses {
			confusion[labels[i]][pred]++
		}
	}

	return &EvaluationResult{
		Accuracy:        accuracy,
		NumSamples:      len(labels),
		NumCorrect:      correct,
		TotalTimeMs:     totalTime,
		AvgTimePerSample: totalTime / float64(len(labels)),
		ConfusionMatrix: confusion,
		Predictions:     predictions,
	}
}

// EvaluationResult contains evaluation metrics
type EvaluationResult struct {
	Accuracy         float64
	NumSamples       int
	NumCorrect       int
	TotalTimeMs      float64
	AvgTimePerSample float64
	ConfusionMatrix  [][]int
	Predictions      []int
}

// GetStats returns inference statistics
func (p *InferencePipeline) GetStats() *InferenceStats {
	return p.Stats
}

// ResetStats resets inference statistics
func (p *InferencePipeline) ResetStats() {
	p.Stats = &InferenceStats{
		LayerLatencies: make(map[string]float64),
	}
}

// ============================================================================
// Crossbar Simulator
// ============================================================================

// CrossbarSimulator simulates crossbar array operations
type CrossbarSimulator struct {
	ArraySize   int
	NoiseLevel  float64
	IRDropCoeff float64
}

// NewCrossbarSimulator creates a new crossbar simulator
func NewCrossbarSimulator(arraySize int, noiseLevel float64) *CrossbarSimulator {
	return &CrossbarSimulator{
		ArraySize:   arraySize,
		NoiseLevel:  noiseLevel,
		IRDropCoeff: 0.01, // 1% IR drop per row
	}
}

// MVM performs matrix-vector multiplication with noise simulation
func (cs *CrossbarSimulator) MVM(input []float64, conductances [][]float64, adcBits int) []float64 {
	if len(conductances) == 0 {
		return nil
	}

	rows := len(conductances)
	cols := len(conductances[0])

	output := make([]float64, rows)

	for i := 0; i < rows; i++ {
		sum := 0.0
		// IR drop increases with row position
		irDropFactor := 1.0 - cs.IRDropCoeff*float64(i)/float64(rows)

		for j := 0; j < cols && j < len(input); j++ {
			// Apply conductance with noise
			g := conductances[i][j]
			gNoise := g + cs.NoiseLevel*g*randNorm()

			// Multiply input voltage by conductance
			v := input[j]
			current := v * gNoise * irDropFactor

			sum += current
		}

		// Add read noise
		sum += cs.NoiseLevel * randNorm()

		// Simulate ADC quantization
		output[i] = cs.adcQuantize(sum, adcBits)
	}

	return output
}

// adcQuantize simulates ADC quantization
func (cs *CrossbarSimulator) adcQuantize(value float64, bits int) float64 {
	levels := float64(int(1) << bits)

	// Assume output range normalized to [-1, 1]
	normalized := value / 10.0 // Scale factor
	if normalized > 1 {
		normalized = 1
	}
	if normalized < -1 {
		normalized = -1
	}

	// Quantize
	level := math.Round((normalized + 1) / 2 * (levels - 1))
	if level < 0 {
		level = 0
	}
	if level >= levels {
		level = levels - 1
	}

	// Dequantize
	dequant := (level/(levels-1))*2 - 1
	return dequant * 10.0
}

// ============================================================================
// Pipeline Builder (Fluent API)
// ============================================================================

// PipelineBuilder provides fluent API for pipeline construction
type PipelineBuilder struct {
	config *PipelineConfig
	model  *ModelCheckpoint
}

// NewPipelineBuilder creates a new pipeline builder
func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		config: DefaultPipelineConfig(),
	}
}

// WithModel sets the model
func (pb *PipelineBuilder) WithModel(model *ModelCheckpoint) *PipelineBuilder {
	pb.model = model
	return pb
}

// WithCIMSimulation enables/disables CIM simulation
func (pb *PipelineBuilder) WithCIMSimulation(enabled bool) *PipelineBuilder {
	pb.config.UseCIMSimulation = enabled
	return pb
}

// WithQuantization sets quantization parameters
func (pb *PipelineBuilder) WithQuantization(inputBits, weightBits, adcBits int) *PipelineBuilder {
	pb.config.QuantizeInputs = true
	pb.config.QuantizeWeights = true
	pb.config.InputBits = inputBits
	pb.config.WeightBits = weightBits
	pb.config.ADCBits = adcBits
	return pb
}

// WithNoiseLevel sets noise level
func (pb *PipelineBuilder) WithNoiseLevel(noise float64) *PipelineBuilder {
	pb.config.NoiseLevel = noise
	return pb
}

// WithCrossbarSize sets crossbar array size
func (pb *PipelineBuilder) WithCrossbarSize(size int) *PipelineBuilder {
	pb.config.CrossbarSize = size
	return pb
}

// WithBatchSize sets batch size
func (pb *PipelineBuilder) WithBatchSize(size int) *PipelineBuilder {
	pb.config.BatchSize = size
	return pb
}

// Verbose enables verbose output
func (pb *PipelineBuilder) Verbose() *PipelineBuilder {
	pb.config.Verbose = true
	return pb
}

// Build constructs the pipeline
func (pb *PipelineBuilder) Build() (*InferencePipeline, error) {
	if pb.model == nil {
		return nil, fmt.Errorf("model required")
	}
	return NewInferencePipeline(pb.model, pb.config), nil
}

// ============================================================================
// Comparison Utilities
// ============================================================================

// CompareIdealVsCIM compares ideal vs CIM inference
func CompareIdealVsCIM(model *ModelCheckpoint, inputs [][]float64, labels []int) *ComparisonResult {
	// Ideal pipeline
	idealPipe := NewInferencePipeline(model, IdealPipelineConfig())
	idealResult := idealPipe.Evaluate(inputs, labels)

	// CIM pipeline
	cimPipe := NewInferencePipeline(model, DefaultPipelineConfig())
	cimResult := cimPipe.Evaluate(inputs, labels)

	// Compute differences
	return &ComparisonResult{
		IdealAccuracy:   idealResult.Accuracy,
		CIMAccuracy:     cimResult.Accuracy,
		AccuracyDrop:    idealResult.Accuracy - cimResult.Accuracy,
		IdealLatencyMs:  idealResult.AvgTimePerSample,
		CIMLatencyMs:    cimResult.AvgTimePerSample,
		Agreement:       computeAgreement(idealResult.Predictions, cimResult.Predictions),
	}
}

// ComparisonResult contains comparison metrics
type ComparisonResult struct {
	IdealAccuracy  float64
	CIMAccuracy    float64
	AccuracyDrop   float64
	IdealLatencyMs float64
	CIMLatencyMs   float64
	Agreement      float64 // Fraction of matching predictions
}

func computeAgreement(pred1, pred2 []int) float64 {
	if len(pred1) != len(pred2) {
		return 0
	}
	matches := 0
	for i := range pred1 {
		if pred1[i] == pred2[i] {
			matches++
		}
	}
	return float64(matches) / float64(len(pred1))
}

// ============================================================================
// Noise Sweep Analysis
// ============================================================================

// NoiseSweepResult contains results from noise sweep
type NoiseSweepResult struct {
	NoiseLevels []float64
	Accuracies  []float64
	Agreements  []float64
}

// SweepNoiseLevel evaluates accuracy across noise levels
func SweepNoiseLevel(model *ModelCheckpoint, inputs [][]float64, labels []int, noiseLevels []float64) *NoiseSweepResult {
	result := &NoiseSweepResult{
		NoiseLevels: noiseLevels,
		Accuracies:  make([]float64, len(noiseLevels)),
		Agreements:  make([]float64, len(noiseLevels)),
	}

	// Get ideal predictions for agreement comparison
	idealPipe := NewInferencePipeline(model, IdealPipelineConfig())
	idealPreds := idealPipe.PredictBatch(inputs)

	for i, noise := range noiseLevels {
		config := DefaultPipelineConfig()
		config.NoiseLevel = noise

		pipe := NewInferencePipeline(model, config)
		evalResult := pipe.Evaluate(inputs, labels)

		result.Accuracies[i] = evalResult.Accuracy
		result.Agreements[i] = computeAgreement(idealPreds, evalResult.Predictions)
	}

	return result
}

// ============================================================================
// Bit-Width Sweep Analysis
// ============================================================================

// BitWidthSweepResult contains results from bit-width sweep
type BitWidthSweepResult struct {
	BitWidths  []int
	Accuracies []float64
}

// SweepWeightBitWidth evaluates accuracy across weight bit widths
func SweepWeightBitWidth(model *ModelCheckpoint, inputs [][]float64, labels []int, bitWidths []int) *BitWidthSweepResult {
	result := &BitWidthSweepResult{
		BitWidths:  bitWidths,
		Accuracies: make([]float64, len(bitWidths)),
	}

	for i, bits := range bitWidths {
		config := DefaultPipelineConfig()
		config.WeightBits = bits
		config.NoiseLevel = 0.0 // Isolate quantization effect

		pipe := NewInferencePipeline(model, config)
		evalResult := pipe.Evaluate(inputs, labels)

		result.Accuracies[i] = evalResult.Accuracy
	}

	return result
}

// ============================================================================
// Model Profiling
// ============================================================================

// ModelProfile contains model profiling information
type ModelProfile struct {
	TotalParameters    int64
	TotalMACs         int64
	MemoryBytes       int64
	LayerProfiles     []LayerProfile
	CrossbarArrays    int
	EstimatedTOPS     float64
	EstimatedEnergy   float64 // picojoules per inference
}

// LayerProfile contains layer-level profiling
type LayerProfile struct {
	Name       string
	Type       string
	Parameters int64
	MACs       int64
	InputSize  int
	OutputSize int
	NumTiles   int
}

// ProfileModel computes model profile
func ProfileModel(model *ModelCheckpoint, crossbarSize int) *ModelProfile {
	profile := &ModelProfile{
		LayerProfiles: make([]LayerProfile, len(model.Layers)),
	}

	for i, layer := range model.Layers {
		lp := LayerProfile{
			Name: layer.Name,
			Type: layer.Type,
		}

		rows, cols := 0, 0
		if len(layer.Shape) >= 2 {
			rows = layer.Shape[0]
			cols = layer.Shape[1]
			lp.OutputSize = rows
			lp.InputSize = cols
		}

		// Parameters
		lp.Parameters = int64(rows * cols)
		if len(layer.Biases) > 0 {
			lp.Parameters += int64(len(layer.Biases))
		}

		// MACs
		lp.MACs = int64(rows * cols)

		// Tiles
		rowTiles := (rows + crossbarSize - 1) / crossbarSize
		colTiles := (cols + crossbarSize - 1) / crossbarSize
		lp.NumTiles = rowTiles * colTiles

		profile.LayerProfiles[i] = lp
		profile.TotalParameters += lp.Parameters
		profile.TotalMACs += lp.MACs
		profile.CrossbarArrays += lp.NumTiles
	}

	profile.MemoryBytes = profile.TotalParameters * 4 // float32

	// Estimate performance (theoretical)
	// Assuming 1 GHz clock, 1 MAC per cycle per crossbar
	profile.EstimatedTOPS = float64(profile.TotalMACs) / 1e12

	// Estimate energy (100 fJ per MAC for CIM)
	profile.EstimatedEnergy = float64(profile.TotalMACs) * 100 // fJ

	return profile
}

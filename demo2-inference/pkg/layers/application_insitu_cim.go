// application_insitu_cim.go - Application-Specific Optimization and In-Situ Training for CIM
// Iteration 145: Domain-specific accelerators and on-chip learning capabilities
//
// Application-Specific CIM Optimization:
// - Domain-specific accelerator configurations (NLP, vision, audio, medical)
// - Workload-aware optimization strategies
// - EdgeBERT-style transformer optimization
// - Latency-aware energy optimization
//
// In-Situ Training and Adaptation:
// - Error-aware probabilistic update (EaPU) for memristor training
// - Multi-tile residual learning for limited conductance states
// - Open-loop weight update scheme
// - Forward-Forward algorithm (BP-free training)
// - Gradient computation directly on crossbar arrays
//
// References:
// - All-in-One Analog AI Accelerator (Adv. Functional Materials 2025)
// - Error-Aware Probabilistic Training (Nature Communications 2025)
// - EdgeBERT (MICRO 2021)
// - Multi-tile Residual Learning (arXiv 2025)

package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// ============================================================================
// Application-Specific CIM Configuration
// ============================================================================

// ApplicationDomain defines supported application domains
type ApplicationDomain string

const (
	DomainNLP           ApplicationDomain = "nlp"
	DomainVision        ApplicationDomain = "vision"
	DomainAudio         ApplicationDomain = "audio"
	DomainMedical       ApplicationDomain = "medical"
	DomainAutonomous    ApplicationDomain = "autonomous"
	DomainWearable      ApplicationDomain = "wearable"
	DomainKeywordSpot   ApplicationDomain = "keyword_spotting"
	DomainRecommendation ApplicationDomain = "recommendation"
)

// DomainSpecificConfig holds optimized parameters for each application domain
type DomainSpecificConfig struct {
	Domain           ApplicationDomain
	ArraySize        int     // Optimal crossbar size
	WeightBits       int     // Weight precision
	ActivationBits   int     // Activation precision
	ADCBits          int     // ADC resolution
	BatchSize        int     // Optimal batch size
	LatencyBudgetUs  float64 // Latency constraint in microseconds
	PowerBudgetMw    float64 // Power budget in milliwatts
	AccuracyTarget   float64 // Minimum acceptable accuracy
	SparsityTarget   float64 // Target weight sparsity

	// Domain-specific optimizations
	EnablePruning      bool
	EnableQuantization bool
	EnableDistillation bool
	EnableEarlyExit    bool
}

// DomainConfigLibrary stores optimized configurations for each domain
var DomainConfigLibrary = map[ApplicationDomain]*DomainSpecificConfig{
	DomainNLP: {
		Domain:           DomainNLP,
		ArraySize:        128,
		WeightBits:       8,
		ActivationBits:   8,
		ADCBits:          8,
		BatchSize:        1,
		LatencyBudgetUs:  10000, // 10ms for real-time NLP
		PowerBudgetMw:    500,
		AccuracyTarget:   0.85,
		SparsityTarget:   0.5,
		EnablePruning:    true,
		EnableQuantization: true,
		EnableDistillation: true,
		EnableEarlyExit:  true,
	},
	DomainVision: {
		Domain:           DomainVision,
		ArraySize:        256,
		WeightBits:       6,
		ActivationBits:   8,
		ADCBits:          6,
		BatchSize:        4,
		LatencyBudgetUs:  33000, // 30 FPS
		PowerBudgetMw:    1000,
		AccuracyTarget:   0.90,
		SparsityTarget:   0.4,
		EnablePruning:    true,
		EnableQuantization: true,
		EnableDistillation: false,
		EnableEarlyExit:  false,
	},
	DomainKeywordSpot: {
		Domain:           DomainKeywordSpot,
		ArraySize:        64,
		WeightBits:       4,
		ActivationBits:   4,
		ADCBits:          4,
		BatchSize:        1,
		LatencyBudgetUs:  1000, // 1ms for real-time KWS
		PowerBudgetMw:    10,   // Ultra-low power
		AccuracyTarget:   0.95,
		SparsityTarget:   0.6,
		EnablePruning:    true,
		EnableQuantization: true,
		EnableDistillation: true,
		EnableEarlyExit:  false,
	},
	DomainMedical: {
		Domain:           DomainMedical,
		ArraySize:        128,
		WeightBits:       8,
		ActivationBits:   8,
		ADCBits:          8,
		BatchSize:        1,
		LatencyBudgetUs:  100000, // 100ms acceptable
		PowerBudgetMw:    200,
		AccuracyTarget:   0.98,   // High accuracy critical
		SparsityTarget:   0.3,
		EnablePruning:    false,  // Preserve accuracy
		EnableQuantization: false,
		EnableDistillation: false,
		EnableEarlyExit:  false,
	},
	DomainAutonomous: {
		Domain:           DomainAutonomous,
		ArraySize:        256,
		WeightBits:       8,
		ActivationBits:   8,
		ADCBits:          8,
		BatchSize:        1,
		LatencyBudgetUs:  5000, // 5ms for safety
		PowerBudgetMw:    5000,
		AccuracyTarget:   0.99, // Safety-critical
		SparsityTarget:   0.2,
		EnablePruning:    true,
		EnableQuantization: true,
		EnableDistillation: false,
		EnableEarlyExit:  true,
	},
	DomainWearable: {
		Domain:           DomainWearable,
		ArraySize:        32,
		WeightBits:       4,
		ActivationBits:   4,
		ADCBits:          4,
		BatchSize:        1,
		LatencyBudgetUs:  50000, // 50ms acceptable
		PowerBudgetMw:    5,     // Battery constrained
		AccuracyTarget:   0.90,
		SparsityTarget:   0.7,
		EnablePruning:    true,
		EnableQuantization: true,
		EnableDistillation: true,
		EnableEarlyExit:  true,
	},
}

// ApplicationOptimizer optimizes CIM configuration for specific domains
type ApplicationOptimizer struct {
	Config           *DomainSpecificConfig
	WorkloadProfile  *WorkloadProfile
	OptimizedMapping *LayerMapping
	Statistics       *OptimizationStats
}

// WorkloadProfile captures workload characteristics
type WorkloadProfile struct {
	LayerTypes       []string           // Conv, FC, Attention, etc.
	LayerSizes       [][]int            // Dimensions for each layer
	ComputeIntensity []float64          // MACs per layer
	MemoryAccess     []float64          // Bytes accessed per layer
	SparsityPattern  []float64          // Sparsity per layer
	ActivationRange  [][]float64        // Min/max activations
	WeightDistribution [][]float64      // Weight histograms
}

// LayerMapping maps layers to CIM resources
type LayerMapping struct {
	LayerToArray    map[int][]int      // Layer -> crossbar array IDs
	ArrayConfig     map[int]*ArrayOptConfig
	PipelineStages  [][]int            // Layers grouped by pipeline stage
	MemoryAlloc     map[int]int        // Layer -> memory bank
}

// ArrayOptConfig holds per-array optimization config
type ArrayOptConfig struct {
	ArrayID        int
	Size           int
	WeightBits     int
	ActivationBits int
	SparsityMask   [][]bool
	PruningRatio   float64
}

// OptimizationStats tracks optimization results
type OptimizationStats struct {
	OriginalLatencyUs    float64
	OptimizedLatencyUs   float64
	OriginalPowerMw      float64
	OptimizedPowerMw     float64
	OriginalAccuracy     float64
	OptimizedAccuracy    float64
	CompressionRatio     float64
	SparsityAchieved     float64
}

// NewApplicationOptimizer creates optimizer for specific domain
func NewApplicationOptimizer(domain ApplicationDomain) *ApplicationOptimizer {
	config, exists := DomainConfigLibrary[domain]
	if !exists {
		config = DomainConfigLibrary[DomainVision] // Default
	}

	return &ApplicationOptimizer{
		Config:     config,
		Statistics: &OptimizationStats{},
	}
}

// ProfileWorkload analyzes workload characteristics
func (ao *ApplicationOptimizer) ProfileWorkload(weights [][][]float64, activations [][]float64) *WorkloadProfile {
	profile := &WorkloadProfile{
		LayerTypes:         make([]string, len(weights)),
		LayerSizes:         make([][]int, len(weights)),
		ComputeIntensity:   make([]float64, len(weights)),
		MemoryAccess:       make([]float64, len(weights)),
		SparsityPattern:    make([]float64, len(weights)),
		ActivationRange:    make([][]float64, len(weights)),
		WeightDistribution: make([][]float64, len(weights)),
	}

	for i, layerWeights := range weights {
		rows := len(layerWeights)
		cols := 0
		if rows > 0 {
			cols = len(layerWeights[0])
		}

		profile.LayerSizes[i] = []int{rows, cols}
		profile.ComputeIntensity[i] = float64(rows * cols * 2) // MACs
		profile.MemoryAccess[i] = float64(rows*cols*4 + rows*4) // Bytes

		// Calculate sparsity
		totalWeights := 0
		zeroWeights := 0
		minW, maxW := math.MaxFloat64, -math.MaxFloat64

		for _, row := range layerWeights {
			for _, w := range row {
				totalWeights++
				if math.Abs(w) < 1e-6 {
					zeroWeights++
				}
				minW = math.Min(minW, w)
				maxW = math.Max(maxW, w)
			}
		}

		if totalWeights > 0 {
			profile.SparsityPattern[i] = float64(zeroWeights) / float64(totalWeights)
		}

		// Weight distribution histogram (10 bins)
		profile.WeightDistribution[i] = ao.computeHistogram(layerWeights, 10)

		// Classify layer type based on dimensions
		if rows == cols {
			profile.LayerTypes[i] = "FC"
		} else if rows > cols*4 {
			profile.LayerTypes[i] = "Conv"
		} else {
			profile.LayerTypes[i] = "Attention"
		}
	}

	ao.WorkloadProfile = profile
	return profile
}

// computeHistogram creates weight distribution histogram
func (ao *ApplicationOptimizer) computeHistogram(weights [][]float64, bins int) []float64 {
	// Find range
	minW, maxW := math.MaxFloat64, -math.MaxFloat64
	for _, row := range weights {
		for _, w := range row {
			minW = math.Min(minW, w)
			maxW = math.Max(maxW, w)
		}
	}

	if maxW <= minW {
		return make([]float64, bins)
	}

	histogram := make([]float64, bins)
	binWidth := (maxW - minW) / float64(bins)
	total := 0.0

	for _, row := range weights {
		for _, w := range row {
			binIdx := int((w - minW) / binWidth)
			if binIdx >= bins {
				binIdx = bins - 1
			}
			histogram[binIdx]++
			total++
		}
	}

	// Normalize
	if total > 0 {
		for i := range histogram {
			histogram[i] /= total
		}
	}

	return histogram
}

// OptimizeForDomain applies domain-specific optimizations
func (ao *ApplicationOptimizer) OptimizeForDomain(weights [][][]float64) ([][][]float64, *LayerMapping) {
	optimized := make([][][]float64, len(weights))
	mapping := &LayerMapping{
		LayerToArray: make(map[int][]int),
		ArrayConfig:  make(map[int]*ArrayOptConfig),
		MemoryAlloc:  make(map[int]int),
	}

	arrayID := 0

	for i, layerWeights := range weights {
		// Apply pruning if enabled
		if ao.Config.EnablePruning {
			layerWeights = ao.applyStructuredPruning(layerWeights, ao.Config.SparsityTarget)
		}

		// Apply quantization if enabled
		if ao.Config.EnableQuantization {
			layerWeights = ao.quantizeWeights(layerWeights, ao.Config.WeightBits)
		}

		// Map to crossbar arrays
		rows := len(layerWeights)
		cols := 0
		if rows > 0 {
			cols = len(layerWeights[0])
		}

		// Calculate number of arrays needed
		arraysNeeded := ((rows + ao.Config.ArraySize - 1) / ao.Config.ArraySize) *
			((cols + ao.Config.ArraySize - 1) / ao.Config.ArraySize)

		arrayIDs := make([]int, arraysNeeded)
		for j := 0; j < arraysNeeded; j++ {
			arrayIDs[j] = arrayID
			mapping.ArrayConfig[arrayID] = &ArrayOptConfig{
				ArrayID:        arrayID,
				Size:           ao.Config.ArraySize,
				WeightBits:     ao.Config.WeightBits,
				ActivationBits: ao.Config.ActivationBits,
				PruningRatio:   ao.Config.SparsityTarget,
			}
			arrayID++
		}
		mapping.LayerToArray[i] = arrayIDs
		mapping.MemoryAlloc[i] = i % 4 // Round-robin memory banks

		optimized[i] = layerWeights
	}

	// Create pipeline stages
	mapping.PipelineStages = ao.createPipelineStages(len(weights))

	ao.OptimizedMapping = mapping
	return optimized, mapping
}

// applyStructuredPruning applies block-structured pruning
func (ao *ApplicationOptimizer) applyStructuredPruning(weights [][]float64, targetSparsity float64) [][]float64 {
	pruned := make([][]float64, len(weights))
	blockSize := 4 // 4x4 block pruning

	for i, row := range weights {
		pruned[i] = make([]float64, len(row))
		copy(pruned[i], row)
	}

	// Calculate block importance scores
	numBlockRows := (len(weights) + blockSize - 1) / blockSize
	numBlockCols := 0
	if len(weights) > 0 {
		numBlockCols = (len(weights[0]) + blockSize - 1) / blockSize
	}

	type blockScore struct {
		row, col int
		score    float64
	}

	scores := make([]blockScore, 0, numBlockRows*numBlockCols)

	for br := 0; br < numBlockRows; br++ {
		for bc := 0; bc < numBlockCols; bc++ {
			score := 0.0
			count := 0

			for i := br * blockSize; i < min((br+1)*blockSize, len(weights)); i++ {
				for j := bc * blockSize; j < min((bc+1)*blockSize, len(weights[0])); j++ {
					score += weights[i][j] * weights[i][j]
					count++
				}
			}

			if count > 0 {
				scores = append(scores, blockScore{br, bc, score / float64(count)})
			}
		}
	}

	// Sort by importance (ascending)
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].score > scores[j].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Prune lowest importance blocks
	numToPrune := int(float64(len(scores)) * targetSparsity)
	for k := 0; k < numToPrune; k++ {
		br, bc := scores[k].row, scores[k].col
		for i := br * blockSize; i < min((br+1)*blockSize, len(pruned)); i++ {
			for j := bc * blockSize; j < min((bc+1)*blockSize, len(pruned[0])); j++ {
				pruned[i][j] = 0
			}
		}
	}

	return pruned
}

// quantizeWeights quantizes to specified bit width
func (ao *ApplicationOptimizer) quantizeWeights(weights [][]float64, bits int) [][]float64 {
	quantized := make([][]float64, len(weights))
	levels := float64(int(1) << bits)

	// Find range
	minW, maxW := math.MaxFloat64, -math.MaxFloat64
	for _, row := range weights {
		for _, w := range row {
			minW = math.Min(minW, w)
			maxW = math.Max(maxW, w)
		}
	}

	scale := (maxW - minW) / (levels - 1)
	if scale < 1e-10 {
		scale = 1e-10
	}

	for i, row := range weights {
		quantized[i] = make([]float64, len(row))
		for j, w := range row {
			// Quantize
			level := math.Round((w - minW) / scale)
			level = math.Max(0, math.Min(levels-1, level))
			// Dequantize
			quantized[i][j] = level*scale + minW
		}
	}

	return quantized
}

// createPipelineStages groups layers into pipeline stages
func (ao *ApplicationOptimizer) createPipelineStages(numLayers int) [][]int {
	// Simple grouping: 4 layers per stage
	layersPerStage := 4
	stages := make([][]int, 0)

	for i := 0; i < numLayers; i += layersPerStage {
		stage := make([]int, 0)
		for j := i; j < min(i+layersPerStage, numLayers); j++ {
			stage = append(stage, j)
		}
		stages = append(stages, stage)
	}

	return stages
}

// ============================================================================
// EdgeBERT-Style Transformer Optimization
// ============================================================================

// EdgeBERTConfig holds EdgeBERT optimization parameters
type EdgeBERTConfig struct {
	NumLayers        int
	HiddenSize       int
	NumHeads         int
	MaxSeqLength     int

	// Optimization parameters
	EarlyExitThreshold float64
	EntropyThreshold   float64
	DynamicWidth       bool
	AdaptiveDepth      bool

	// Hardware constraints
	LatencyBudgetMs float64
	PowerBudgetMw   float64
}

// EdgeBERTOptimizer optimizes BERT-like models for edge CIM
type EdgeBERTOptimizer struct {
	Config        *EdgeBERTConfig
	LayerLatency  []float64           // Estimated latency per layer
	LayerPower    []float64           // Estimated power per layer
	ExitScores    []float64           // Confidence scores for early exit
	Statistics    *EdgeBERTStats
}

// EdgeBERTStats tracks optimization statistics
type EdgeBERTStats struct {
	AverageExitLayer   float64
	LatencySavings     float64
	EnergySavings      float64
	AccuracyDrop       float64
	ThroughputImprove  float64
}

// NewEdgeBERTOptimizer creates EdgeBERT optimizer
func NewEdgeBERTOptimizer(config *EdgeBERTConfig) *EdgeBERTOptimizer {
	return &EdgeBERTOptimizer{
		Config:       config,
		LayerLatency: make([]float64, config.NumLayers),
		LayerPower:   make([]float64, config.NumLayers),
		ExitScores:   make([]float64, config.NumLayers),
		Statistics:   &EdgeBERTStats{},
	}
}

// OptimizeLayer applies layer-specific optimizations
func (ebo *EdgeBERTOptimizer) OptimizeLayer(layerIdx int, weights [][]float64,
	activations []float64) ([][]float64, []float64, bool) {

	// Estimate layer metrics
	ebo.estimateLayerMetrics(layerIdx, weights)

	// Check early exit condition
	exitScore := ebo.computeExitScore(activations)
	ebo.ExitScores[layerIdx] = exitScore

	shouldExit := ebo.Config.AdaptiveDepth &&
		exitScore > ebo.Config.EarlyExitThreshold &&
		layerIdx >= ebo.Config.NumLayers/3 // At least 1/3 layers

	// Apply width reduction if enabled
	optimizedWeights := weights
	if ebo.Config.DynamicWidth {
		optimizedWeights = ebo.applyDynamicWidth(weights, activations)
	}

	// Compute output
	output := ebo.computeLayerOutput(optimizedWeights, activations)

	return optimizedWeights, output, shouldExit
}

// estimateLayerMetrics estimates latency and power for layer
func (ebo *EdgeBERTOptimizer) estimateLayerMetrics(layerIdx int, weights [][]float64) {
	rows := len(weights)
	cols := 0
	if rows > 0 {
		cols = len(weights[0])
	}

	// Latency model: ~0.1us per 1000 MACs on CIM
	macs := float64(rows * cols)
	ebo.LayerLatency[layerIdx] = macs / 10000.0 // us

	// Power model: ~0.1mW per 1000 MACs
	ebo.LayerPower[layerIdx] = macs / 10000.0 // mW
}

// computeExitScore computes confidence for early exit
func (ebo *EdgeBERTOptimizer) computeExitScore(activations []float64) float64 {
	if len(activations) == 0 {
		return 0
	}

	// Compute entropy of softmax
	maxVal := activations[0]
	for _, v := range activations {
		if v > maxVal {
			maxVal = v
		}
	}

	expSum := 0.0
	for _, v := range activations {
		expSum += math.Exp(v - maxVal)
	}

	entropy := 0.0
	for _, v := range activations {
		p := math.Exp(v-maxVal) / expSum
		if p > 1e-10 {
			entropy -= p * math.Log(p)
		}
	}

	// Convert to confidence (low entropy = high confidence)
	maxEntropy := math.Log(float64(len(activations)))
	if maxEntropy > 0 {
		return 1.0 - entropy/maxEntropy
	}
	return 1.0
}

// applyDynamicWidth reduces width based on importance
func (ebo *EdgeBERTOptimizer) applyDynamicWidth(weights [][]float64,
	activations []float64) [][]float64 {

	if len(weights) == 0 || len(weights[0]) == 0 {
		return weights
	}

	// Keep top 75% of neurons by importance
	keepRatio := 0.75
	rows := len(weights)
	cols := len(weights[0])
	newCols := int(float64(cols) * keepRatio)

	// Compute column importance
	colImportance := make([]float64, cols)
	for j := 0; j < cols; j++ {
		for i := 0; i < rows; i++ {
			colImportance[j] += weights[i][j] * weights[i][j]
		}
	}

	// Get indices of top columns
	topCols := make([]int, newCols)
	used := make([]bool, cols)
	for k := 0; k < newCols; k++ {
		maxIdx := -1
		maxVal := -1.0
		for j := 0; j < cols; j++ {
			if !used[j] && colImportance[j] > maxVal {
				maxVal = colImportance[j]
				maxIdx = j
			}
		}
		if maxIdx >= 0 {
			topCols[k] = maxIdx
			used[maxIdx] = true
		}
	}

	// Create reduced weight matrix
	reduced := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		reduced[i] = make([]float64, newCols)
		for k, j := range topCols {
			reduced[i][k] = weights[i][j]
		}
	}

	return reduced
}

// computeLayerOutput computes MVM output
func (ebo *EdgeBERTOptimizer) computeLayerOutput(weights [][]float64,
	input []float64) []float64 {

	if len(weights) == 0 {
		return nil
	}

	output := make([]float64, len(weights))
	for i, row := range weights {
		sum := 0.0
		for j, w := range row {
			if j < len(input) {
				sum += w * input[j]
			}
		}
		output[i] = sum
	}

	return output
}

// ComputeStatistics calculates overall optimization statistics
func (ebo *EdgeBERTOptimizer) ComputeStatistics(actualExitLayers []int) *EdgeBERTStats {
	// Average exit layer
	sum := 0.0
	for _, exit := range actualExitLayers {
		sum += float64(exit)
	}
	if len(actualExitLayers) > 0 {
		ebo.Statistics.AverageExitLayer = sum / float64(len(actualExitLayers))
	}

	// Latency savings
	fullLatency := 0.0
	for _, lat := range ebo.LayerLatency {
		fullLatency += lat
	}

	avgLatency := 0.0
	for i := 0; i < int(ebo.Statistics.AverageExitLayer); i++ {
		if i < len(ebo.LayerLatency) {
			avgLatency += ebo.LayerLatency[i]
		}
	}

	if fullLatency > 0 {
		ebo.Statistics.LatencySavings = 1.0 - avgLatency/fullLatency
	}

	// Similar calculation for energy
	fullEnergy := 0.0
	for _, pow := range ebo.LayerPower {
		fullEnergy += pow
	}

	avgEnergy := 0.0
	for i := 0; i < int(ebo.Statistics.AverageExitLayer); i++ {
		if i < len(ebo.LayerPower) {
			avgEnergy += ebo.LayerPower[i]
		}
	}

	if fullEnergy > 0 {
		ebo.Statistics.EnergySavings = 1.0 - avgEnergy/fullEnergy
	}

	return ebo.Statistics
}

// ============================================================================
// In-Situ Training: Error-Aware Probabilistic Update (EaPU)
// ============================================================================

// EaPUConfig holds error-aware probabilistic update parameters
type EaPUConfig struct {
	WriteNoiseSigma   float64 // Device write noise standard deviation
	UpdateThreshold   float64 // Minimum gradient magnitude to trigger update
	ProbabilityDecay  float64 // Decay factor for update probability
	MaxUpdateAttempts int     // Maximum write attempts per weight

	// Learning parameters
	LearningRate     float64
	MomentumFactor   float64
	WeightDecay      float64
}

// EaPUTrainer implements error-aware probabilistic training
type EaPUTrainer struct {
	Config           *EaPUConfig
	Weights          [][]float64
	Gradients        [][]float64
	Momentum         [][]float64
	UpdateMask       [][]bool
	Statistics       *EaPUStats
}

// EaPUStats tracks training statistics
type EaPUStats struct {
	TotalUpdates       int
	SkippedUpdates     int
	UpdateRatio        float64
	AverageNoise       float64
	EnergyConsumption  float64
	AccuracyImprovement float64
}

// NewEaPUTrainer creates error-aware probabilistic trainer
func NewEaPUTrainer(config *EaPUConfig, weightRows, weightCols int) *EaPUTrainer {
	trainer := &EaPUTrainer{
		Config:     config,
		Weights:    make([][]float64, weightRows),
		Gradients:  make([][]float64, weightRows),
		Momentum:   make([][]float64, weightRows),
		UpdateMask: make([][]bool, weightRows),
		Statistics: &EaPUStats{},
	}

	for i := 0; i < weightRows; i++ {
		trainer.Weights[i] = make([]float64, weightCols)
		trainer.Gradients[i] = make([]float64, weightCols)
		trainer.Momentum[i] = make([]float64, weightCols)
		trainer.UpdateMask[i] = make([]bool, weightCols)
	}

	return trainer
}

// InitializeWeights sets initial weights
func (et *EaPUTrainer) InitializeWeights(weights [][]float64) {
	for i := range weights {
		copy(et.Weights[i], weights[i])
	}
}

// ComputeUpdateProbability computes probability of updating each weight
func (et *EaPUTrainer) ComputeUpdateProbability(gradient float64) float64 {
	// Probability based on gradient magnitude relative to noise
	gradMag := math.Abs(gradient)
	snr := gradMag / et.Config.WriteNoiseSigma

	// Sigmoid-based probability
	prob := 1.0 / (1.0 + math.Exp(-snr+2.0))

	// Apply threshold
	if gradMag < et.Config.UpdateThreshold {
		prob *= et.Config.ProbabilityDecay
	}

	return prob
}

// UpdateWeights performs probabilistic weight update
func (et *EaPUTrainer) UpdateWeights(gradients [][]float64) {
	for i := range gradients {
		for j := range gradients[i] {
			et.Gradients[i][j] = gradients[i][j]

			// Compute update probability
			prob := et.ComputeUpdateProbability(gradients[i][j])

			// Probabilistic update decision
			if rand.Float64() < prob {
				// Update momentum
				et.Momentum[i][j] = et.Config.MomentumFactor*et.Momentum[i][j] +
					et.Config.LearningRate*gradients[i][j]

				// Apply weight decay
				et.Weights[i][j] -= et.Weights[i][j] * et.Config.WeightDecay

				// Apply update with write noise
				noise := rand.NormFloat64() * et.Config.WriteNoiseSigma
				et.Weights[i][j] -= et.Momentum[i][j] + noise

				et.UpdateMask[i][j] = true
				et.Statistics.TotalUpdates++
				et.Statistics.AverageNoise += math.Abs(noise)
			} else {
				et.UpdateMask[i][j] = false
				et.Statistics.SkippedUpdates++
			}
		}
	}

	// Update statistics
	total := et.Statistics.TotalUpdates + et.Statistics.SkippedUpdates
	if total > 0 {
		et.Statistics.UpdateRatio = float64(et.Statistics.TotalUpdates) / float64(total)
		et.Statistics.AverageNoise /= float64(max(1, et.Statistics.TotalUpdates))
	}

	// Energy model: ~1pJ per weight update
	et.Statistics.EnergyConsumption += float64(et.Statistics.TotalUpdates) * 1e-12
}

// GetWeights returns current weights
func (et *EaPUTrainer) GetWeights() [][]float64 {
	return et.Weights
}

// ============================================================================
// Multi-Tile Residual Learning
// ============================================================================

// MultiTileConfig holds multi-tile residual learning parameters
type MultiTileConfig struct {
	NumTiles         int     // Number of crossbar tiles
	BitsPerTile      int     // Conductance bits per tile
	TotalBits        int     // Target total precision
	AsymmetryFactor  float64 // Update asymmetry (set vs reset)

	// Learning parameters
	LearningRate     float64
	ResidualDecay    float64
}

// MultiTileTrainer implements multi-tile residual learning
type MultiTileTrainer struct {
	Config         *MultiTileConfig
	TileWeights    [][][]float64     // Weights per tile
	ResidualErrors [][]float64       // Accumulated residual errors
	TileScales     []float64         // Scale factors per tile
	Statistics     *MultiTileStats
}

// MultiTileStats tracks multi-tile training statistics
type MultiTileStats struct {
	TileUtilization    []float64
	ResidualMagnitude  float64
	EffectivePrecision float64
	ConvergenceRate    float64
}

// NewMultiTileTrainer creates multi-tile residual learning trainer
func NewMultiTileTrainer(config *MultiTileConfig, rows, cols int) *MultiTileTrainer {
	trainer := &MultiTileTrainer{
		Config:         config,
		TileWeights:    make([][][]float64, config.NumTiles),
		ResidualErrors: make([][]float64, rows),
		TileScales:     make([]float64, config.NumTiles),
		Statistics:     &MultiTileStats{TileUtilization: make([]float64, config.NumTiles)},
	}

	// Initialize tiles with decreasing scale factors
	for t := 0; t < config.NumTiles; t++ {
		trainer.TileWeights[t] = make([][]float64, rows)
		for i := 0; i < rows; i++ {
			trainer.TileWeights[t][i] = make([]float64, cols)
		}
		// Exponentially decreasing scales
		trainer.TileScales[t] = math.Pow(0.5, float64(t))
	}

	for i := 0; i < rows; i++ {
		trainer.ResidualErrors[i] = make([]float64, cols)
	}

	return trainer
}

// DecomposeWeights decomposes high-precision weights to multi-tile representation
func (mtt *MultiTileTrainer) DecomposeWeights(targetWeights [][]float64) {
	rows := len(targetWeights)
	if rows == 0 {
		return
	}
	cols := len(targetWeights[0])

	// Initialize residual with target
	for i := 0; i < rows; i++ {
		copy(mtt.ResidualErrors[i], targetWeights[i])
	}

	// Sequentially decompose to each tile
	for t := 0; t < mtt.Config.NumTiles; t++ {
		scale := mtt.TileScales[t]
		levels := float64(int(1) << mtt.Config.BitsPerTile)

		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				// Quantize residual to tile precision
				normalized := mtt.ResidualErrors[i][j] / scale
				quantized := math.Round(normalized * (levels - 1) / 2)
				quantized = math.Max(-levels/2, math.Min(levels/2-1, quantized))

				// Store in tile
				mtt.TileWeights[t][i][j] = quantized * scale * 2 / (levels - 1)

				// Update residual
				mtt.ResidualErrors[i][j] -= mtt.TileWeights[t][i][j]
			}
		}
	}

	// Compute statistics
	mtt.computeDecompositionStats(targetWeights)
}

// computeDecompositionStats computes decomposition quality metrics
func (mtt *MultiTileTrainer) computeDecompositionStats(targetWeights [][]float64) {
	if len(targetWeights) == 0 {
		return
	}

	// Compute residual magnitude
	residualSum := 0.0
	targetSum := 0.0
	count := 0

	for i := range mtt.ResidualErrors {
		for j := range mtt.ResidualErrors[i] {
			residualSum += mtt.ResidualErrors[i][j] * mtt.ResidualErrors[i][j]
			if i < len(targetWeights) && j < len(targetWeights[i]) {
				targetSum += targetWeights[i][j] * targetWeights[i][j]
			}
			count++
		}
	}

	if count > 0 {
		mtt.Statistics.ResidualMagnitude = math.Sqrt(residualSum / float64(count))
	}

	// Compute effective precision
	if targetSum > 0 {
		snr := targetSum / residualSum
		mtt.Statistics.EffectivePrecision = math.Log2(snr) / 2
	}

	// Compute tile utilization
	for t := 0; t < mtt.Config.NumTiles; t++ {
		utilization := 0.0
		total := 0

		for i := range mtt.TileWeights[t] {
			for j := range mtt.TileWeights[t][i] {
				if math.Abs(mtt.TileWeights[t][i][j]) > 1e-10 {
					utilization++
				}
				total++
			}
		}

		if total > 0 {
			mtt.Statistics.TileUtilization[t] = utilization / float64(total)
		}
	}
}

// UpdateWithResidualLearning performs one training step with residual learning
func (mtt *MultiTileTrainer) UpdateWithResidualLearning(gradients [][]float64) {
	// Update tile 0 first (coarse weights)
	for t := 0; t < mtt.Config.NumTiles; t++ {
		scale := mtt.TileScales[t]
		levels := float64(int(1) << mtt.Config.BitsPerTile)

		for i := range gradients {
			for j := range gradients[i] {
				// Scaled gradient for this tile
				scaledGrad := gradients[i][j] * mtt.Config.LearningRate / scale

				// Asymmetric update
				if scaledGrad > 0 {
					scaledGrad *= mtt.Config.AsymmetryFactor
				}

				// Update with quantization
				oldVal := mtt.TileWeights[t][i][j] / scale
				newVal := oldVal - scaledGrad
				newVal = math.Round(newVal * (levels - 1) / 2)
				newVal = math.Max(-levels/2, math.Min(levels/2-1, newVal))
				mtt.TileWeights[t][i][j] = newVal * scale * 2 / (levels - 1)
			}
		}

		// Propagate residual to next tile
		if t < mtt.Config.NumTiles-1 {
			for i := range gradients {
				for j := range gradients[i] {
					oldVal := mtt.TileWeights[t][i][j]
					// Residual gradient for next tile
					mtt.ResidualErrors[i][j] = gradients[i][j] - oldVal
				}
			}
		}
	}
}

// ReconstructWeights reconstructs full-precision weights from tiles
func (mtt *MultiTileTrainer) ReconstructWeights() [][]float64 {
	if len(mtt.TileWeights) == 0 || len(mtt.TileWeights[0]) == 0 {
		return nil
	}

	rows := len(mtt.TileWeights[0])
	cols := len(mtt.TileWeights[0][0])

	reconstructed := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		reconstructed[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			for t := 0; t < mtt.Config.NumTiles; t++ {
				reconstructed[i][j] += mtt.TileWeights[t][i][j]
			}
		}
	}

	return reconstructed
}

// ============================================================================
// Forward-Forward Algorithm (BP-Free Training)
// ============================================================================

// ForwardForwardConfig holds FF algorithm parameters
type ForwardForwardConfig struct {
	NumLayers       int
	HiddenSizes     []int
	Threshold       float64 // Goodness threshold
	LearningRate    float64
	NumNegSamples   int     // Negative samples per positive
	LayerNormalize  bool
}

// ForwardForwardTrainer implements BP-free forward-forward training
type ForwardForwardTrainer struct {
	Config       *ForwardForwardConfig
	Weights      [][][]float64
	Biases       [][]float64
	Statistics   *FFStats
}

// FFStats tracks forward-forward training statistics
type FFStats struct {
	PositiveGoodness []float64
	NegativeGoodness []float64
	LayerAccuracy    []float64
	TotalEpochs      int
}

// NewForwardForwardTrainer creates FF trainer
func NewForwardForwardTrainer(config *ForwardForwardConfig, inputSize int) *ForwardForwardTrainer {
	trainer := &ForwardForwardTrainer{
		Config:     config,
		Weights:    make([][][]float64, config.NumLayers),
		Biases:     make([][]float64, config.NumLayers),
		Statistics: &FFStats{
			PositiveGoodness: make([]float64, config.NumLayers),
			NegativeGoodness: make([]float64, config.NumLayers),
			LayerAccuracy:    make([]float64, config.NumLayers),
		},
	}

	prevSize := inputSize
	for l := 0; l < config.NumLayers; l++ {
		hiddenSize := config.HiddenSizes[l]
		trainer.Weights[l] = make([][]float64, hiddenSize)
		trainer.Biases[l] = make([]float64, hiddenSize)

		// Xavier initialization
		scale := math.Sqrt(2.0 / float64(prevSize+hiddenSize))
		for i := 0; i < hiddenSize; i++ {
			trainer.Weights[l][i] = make([]float64, prevSize)
			for j := 0; j < prevSize; j++ {
				trainer.Weights[l][i][j] = rand.NormFloat64() * scale
			}
		}

		prevSize = hiddenSize
	}

	return trainer
}

// ComputeGoodness computes "goodness" (sum of squared activations)
func (fft *ForwardForwardTrainer) ComputeGoodness(activations []float64) float64 {
	goodness := 0.0
	for _, a := range activations {
		goodness += a * a
	}
	return goodness
}

// ForwardPass performs forward pass for one layer
func (fft *ForwardForwardTrainer) ForwardPass(layerIdx int, input []float64) []float64 {
	weights := fft.Weights[layerIdx]
	biases := fft.Biases[layerIdx]

	output := make([]float64, len(weights))

	for i := range weights {
		sum := biases[i]
		for j, w := range weights[i] {
			if j < len(input) {
				sum += w * input[j]
			}
		}
		// ReLU activation
		output[i] = math.Max(0, sum)
	}

	// Layer normalization if enabled
	if fft.Config.LayerNormalize {
		output = fft.layerNormalize(output)
	}

	return output
}

// layerNormalize applies layer normalization
func (fft *ForwardForwardTrainer) layerNormalize(x []float64) []float64 {
	if len(x) == 0 {
		return x
	}

	// Compute mean
	mean := 0.0
	for _, v := range x {
		mean += v
	}
	mean /= float64(len(x))

	// Compute variance
	variance := 0.0
	for _, v := range x {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(x))

	// Normalize
	std := math.Sqrt(variance + 1e-8)
	normalized := make([]float64, len(x))
	for i, v := range x {
		normalized[i] = (v - mean) / std
	}

	return normalized
}

// TrainLayer trains a single layer using forward-forward
func (fft *ForwardForwardTrainer) TrainLayer(layerIdx int, posInput, negInput []float64) {
	// Forward pass for positive sample
	posOutput := fft.ForwardPass(layerIdx, posInput)
	posGoodness := fft.ComputeGoodness(posOutput)

	// Forward pass for negative sample
	negOutput := fft.ForwardPass(layerIdx, negInput)
	negGoodness := fft.ComputeGoodness(negOutput)

	// Update statistics
	fft.Statistics.PositiveGoodness[layerIdx] = posGoodness
	fft.Statistics.NegativeGoodness[layerIdx] = negGoodness

	// Compute local loss gradients
	// Positive: want goodness > threshold
	// Negative: want goodness < threshold

	posLoss := math.Max(0, fft.Config.Threshold-posGoodness)
	negLoss := math.Max(0, negGoodness-fft.Config.Threshold)

	// Update weights based on local learning rule
	for i := range fft.Weights[layerIdx] {
		for j := range fft.Weights[layerIdx][i] {
			// Gradient for positive sample (increase goodness)
			if posLoss > 0 && i < len(posOutput) && j < len(posInput) {
				grad := 2 * posOutput[i] * posInput[j]
				fft.Weights[layerIdx][i][j] += fft.Config.LearningRate * grad
			}

			// Gradient for negative sample (decrease goodness)
			if negLoss > 0 && i < len(negOutput) && j < len(negInput) {
				grad := 2 * negOutput[i] * negInput[j]
				fft.Weights[layerIdx][i][j] -= fft.Config.LearningRate * grad
			}
		}
	}

	// Compute layer accuracy
	correct := 0
	if posGoodness > fft.Config.Threshold {
		correct++
	}
	if negGoodness < fft.Config.Threshold {
		correct++
	}
	fft.Statistics.LayerAccuracy[layerIdx] = float64(correct) / 2.0
}

// CreateNegativeSample creates negative sample by corrupting input
func (fft *ForwardForwardTrainer) CreateNegativeSample(input []float64, label int, numClasses int) []float64 {
	negative := make([]float64, len(input))
	copy(negative, input)

	// Embed wrong label into first numClasses dimensions
	for i := 0; i < numClasses && i < len(negative); i++ {
		if i == label {
			negative[i] = 0 // Remove correct label
		} else if rand.Float64() < 1.0/float64(numClasses-1) {
			negative[i] = 1 // Add random wrong label
		}
	}

	return negative
}

// ============================================================================
// Open-Loop Weight Update Scheme
// ============================================================================

// OpenLoopConfig holds open-loop update parameters
type OpenLoopConfig struct {
	PulseAmplitude   float64 // Voltage amplitude
	PulseWidth       float64 // Pulse width in ns
	NumPulses        int     // Number of pulses per update
	ConductanceRange []float64 // [min, max] conductance

	// Device characteristics
	SetSlope         float64 // Set (potentiation) slope
	ResetSlope       float64 // Reset (depression) slope
	NonlinearityExp  float64 // Nonlinearity exponent
}

// OpenLoopUpdater implements open-loop weight updates
type OpenLoopUpdater struct {
	Config           *OpenLoopConfig
	Conductances     [][]float64
	Statistics       *OpenLoopStats
}

// OpenLoopStats tracks update statistics
type OpenLoopStats struct {
	TotalPulses      int
	SuccessfulSets   int
	SuccessfulResets int
	AverageError     float64
	EnergyPerUpdate  float64
}

// NewOpenLoopUpdater creates open-loop updater
func NewOpenLoopUpdater(config *OpenLoopConfig, rows, cols int) *OpenLoopUpdater {
	updater := &OpenLoopUpdater{
		Config:       config,
		Conductances: make([][]float64, rows),
		Statistics:   &OpenLoopStats{},
	}

	// Initialize conductances
	for i := 0; i < rows; i++ {
		updater.Conductances[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			// Random initial conductance
			gMin := config.ConductanceRange[0]
			gMax := config.ConductanceRange[1]
			updater.Conductances[i][j] = gMin + rand.Float64()*(gMax-gMin)
		}
	}

	return updater
}

// ApplyPulses applies programming pulses to update conductance
func (olu *OpenLoopUpdater) ApplyPulses(row, col int, targetConductance float64) {
	current := olu.Conductances[row][col]
	gMin := olu.Config.ConductanceRange[0]
	gMax := olu.Config.ConductanceRange[1]

	// Determine if set or reset
	isSet := targetConductance > current

	// Apply pulses
	for p := 0; p < olu.Config.NumPulses; p++ {
		if math.Abs(current-targetConductance) < 0.01*(gMax-gMin) {
			break // Close enough
		}

		olu.Statistics.TotalPulses++

		if isSet {
			// Potentiation (set)
			delta := olu.Config.SetSlope * olu.Config.PulseAmplitude
			// Nonlinear update
			delta *= math.Pow(1-(current-gMin)/(gMax-gMin), olu.Config.NonlinearityExp)
			current = math.Min(gMax, current+delta)
			olu.Statistics.SuccessfulSets++
		} else {
			// Depression (reset)
			delta := olu.Config.ResetSlope * olu.Config.PulseAmplitude
			// Nonlinear update
			delta *= math.Pow((current-gMin)/(gMax-gMin), olu.Config.NonlinearityExp)
			current = math.Max(gMin, current-delta)
			olu.Statistics.SuccessfulResets++
		}
	}

	olu.Conductances[row][col] = current

	// Update error statistics
	error := math.Abs(current - targetConductance)
	olu.Statistics.AverageError = (olu.Statistics.AverageError*float64(olu.Statistics.TotalPulses-1) +
		error) / float64(olu.Statistics.TotalPulses)

	// Energy model: ~1pJ per pulse
	olu.Statistics.EnergyPerUpdate = float64(olu.Config.NumPulses) * 1e-12
}

// UpdateWeightMatrix updates entire weight matrix
func (olu *OpenLoopUpdater) UpdateWeightMatrix(targetWeights [][]float64) {
	gMin := olu.Config.ConductanceRange[0]
	gMax := olu.Config.ConductanceRange[1]

	// Find weight range
	wMin, wMax := math.MaxFloat64, -math.MaxFloat64
	for _, row := range targetWeights {
		for _, w := range row {
			wMin = math.Min(wMin, w)
			wMax = math.Max(wMax, w)
		}
	}

	if wMax <= wMin {
		wMax = wMin + 1
	}

	// Map weights to conductances and update
	for i := range targetWeights {
		for j, w := range targetWeights[i] {
			// Linear mapping from weight to conductance
			normalized := (w - wMin) / (wMax - wMin)
			targetG := gMin + normalized*(gMax-gMin)
			olu.ApplyPulses(i, j, targetG)
		}
	}
}

// GetWeights converts conductances back to weights
func (olu *OpenLoopUpdater) GetWeights() [][]float64 {
	gMin := olu.Config.ConductanceRange[0]
	gMax := olu.Config.ConductanceRange[1]

	weights := make([][]float64, len(olu.Conductances))
	for i := range olu.Conductances {
		weights[i] = make([]float64, len(olu.Conductances[i]))
		for j, g := range olu.Conductances[i] {
			// Map conductance to normalized weight [-1, 1]
			weights[i][j] = 2*(g-gMin)/(gMax-gMin) - 1
		}
	}

	return weights
}

// ============================================================================
// Integrated In-Situ Training System
// ============================================================================

// InSituTrainingSystem combines multiple training approaches
type InSituTrainingSystem struct {
	Domain           ApplicationDomain
	AppOptimizer     *ApplicationOptimizer
	EaPUTrainer      *EaPUTrainer
	MultiTileTrainer *MultiTileTrainer
	FFTrainer        *ForwardForwardTrainer
	OpenLoopUpdater  *OpenLoopUpdater

	// Training state
	CurrentEpoch     int
	BestAccuracy     float64
	TrainingHistory  []float64
}

// InSituConfig holds integrated system configuration
type InSituConfig struct {
	Domain            ApplicationDomain
	TrainingMethod    string // "eapu", "multitile", "forward_forward", "openloop"
	WeightRows        int
	WeightCols        int
	NumTiles          int
	BitsPerTile       int
	FFLayers          []int
	LearningRate      float64
}

// NewInSituTrainingSystem creates integrated training system
func NewInSituTrainingSystem(config *InSituConfig) *InSituTrainingSystem {
	system := &InSituTrainingSystem{
		Domain:          config.Domain,
		AppOptimizer:    NewApplicationOptimizer(config.Domain),
		TrainingHistory: make([]float64, 0),
	}

	// Initialize appropriate trainer
	switch config.TrainingMethod {
	case "eapu":
		eapuConfig := &EaPUConfig{
			WriteNoiseSigma:   0.1,
			UpdateThreshold:   0.01,
			ProbabilityDecay:  0.5,
			MaxUpdateAttempts: 10,
			LearningRate:      config.LearningRate,
			MomentumFactor:    0.9,
			WeightDecay:       0.0001,
		}
		system.EaPUTrainer = NewEaPUTrainer(eapuConfig, config.WeightRows, config.WeightCols)

	case "multitile":
		mtConfig := &MultiTileConfig{
			NumTiles:        config.NumTiles,
			BitsPerTile:     config.BitsPerTile,
			TotalBits:       config.NumTiles * config.BitsPerTile,
			AsymmetryFactor: 0.8,
			LearningRate:    config.LearningRate,
			ResidualDecay:   0.1,
		}
		system.MultiTileTrainer = NewMultiTileTrainer(mtConfig, config.WeightRows, config.WeightCols)

	case "forward_forward":
		ffConfig := &ForwardForwardConfig{
			NumLayers:      len(config.FFLayers),
			HiddenSizes:    config.FFLayers,
			Threshold:      2.0,
			LearningRate:   config.LearningRate,
			NumNegSamples:  1,
			LayerNormalize: true,
		}
		inputSize := config.WeightCols
		if inputSize == 0 {
			inputSize = 784 // Default for MNIST
		}
		system.FFTrainer = NewForwardForwardTrainer(ffConfig, inputSize)

	case "openloop":
		olConfig := &OpenLoopConfig{
			PulseAmplitude:   1.0,
			PulseWidth:       100,
			NumPulses:        10,
			ConductanceRange: []float64{1e-6, 100e-6},
			SetSlope:         0.1,
			ResetSlope:       0.08,
			NonlinearityExp:  2.0,
		}
		system.OpenLoopUpdater = NewOpenLoopUpdater(olConfig, config.WeightRows, config.WeightCols)
	}

	return system
}

// TrainStep performs one training step
func (ists *InSituTrainingSystem) TrainStep(input, target []float64, gradients [][]float64) float64 {
	var loss float64

	if ists.EaPUTrainer != nil {
		ists.EaPUTrainer.UpdateWeights(gradients)
		// Compute loss
		weights := ists.EaPUTrainer.GetWeights()
		loss = ists.computeLoss(weights, input, target)
	}

	if ists.MultiTileTrainer != nil {
		ists.MultiTileTrainer.UpdateWithResidualLearning(gradients)
		weights := ists.MultiTileTrainer.ReconstructWeights()
		loss = ists.computeLoss(weights, input, target)
	}

	if ists.OpenLoopUpdater != nil {
		// Convert gradients to target weights
		current := ists.OpenLoopUpdater.GetWeights()
		for i := range current {
			for j := range current[i] {
				if i < len(gradients) && j < len(gradients[i]) {
					current[i][j] -= 0.01 * gradients[i][j]
				}
			}
		}
		ists.OpenLoopUpdater.UpdateWeightMatrix(current)
		weights := ists.OpenLoopUpdater.GetWeights()
		loss = ists.computeLoss(weights, input, target)
	}

	ists.CurrentEpoch++
	ists.TrainingHistory = append(ists.TrainingHistory, loss)

	return loss
}

// computeLoss computes MSE loss
func (ists *InSituTrainingSystem) computeLoss(weights [][]float64, input, target []float64) float64 {
	if len(weights) == 0 || len(input) == 0 {
		return 0
	}

	// Forward pass
	output := make([]float64, len(weights))
	for i, row := range weights {
		sum := 0.0
		for j, w := range row {
			if j < len(input) {
				sum += w * input[j]
			}
		}
		output[i] = sum
	}

	// MSE loss
	loss := 0.0
	for i := range output {
		if i < len(target) {
			diff := output[i] - target[i]
			loss += diff * diff
		}
	}

	return loss / float64(len(output))
}

// GetTrainingSummary returns training summary
func (ists *InSituTrainingSystem) GetTrainingSummary() string {
	summary := fmt.Sprintf("In-Situ Training Summary for %s domain:\n", ists.Domain)
	summary += fmt.Sprintf("  Epochs completed: %d\n", ists.CurrentEpoch)

	if len(ists.TrainingHistory) > 0 {
		lastLoss := ists.TrainingHistory[len(ists.TrainingHistory)-1]
		summary += fmt.Sprintf("  Final loss: %.6f\n", lastLoss)
	}

	if ists.EaPUTrainer != nil {
		stats := ists.EaPUTrainer.Statistics
		summary += fmt.Sprintf("  EaPU update ratio: %.2f%%\n", stats.UpdateRatio*100)
		summary += fmt.Sprintf("  Average noise: %.6f\n", stats.AverageNoise)
	}

	if ists.MultiTileTrainer != nil {
		stats := ists.MultiTileTrainer.Statistics
		summary += fmt.Sprintf("  Effective precision: %.2f bits\n", stats.EffectivePrecision)
		summary += fmt.Sprintf("  Residual magnitude: %.6f\n", stats.ResidualMagnitude)
	}

	if ists.OpenLoopUpdater != nil {
		stats := ists.OpenLoopUpdater.Statistics
		summary += fmt.Sprintf("  Total pulses: %d\n", stats.TotalPulses)
		summary += fmt.Sprintf("  Average error: %.6f\n", stats.AverageError)
	}

	return summary
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

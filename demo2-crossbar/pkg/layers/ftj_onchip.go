// ftj_onchip.go - Ferroelectric Tunnel Junction Memory and On-Chip Learning
//
// This module implements:
// 1. Ferroelectric tunnel junction (FTJ) device physics and crossbar arrays
// 2. On-chip learning algorithms: DFA, local learning, perturbation-based
// 3. Hardware-aware training with analog noise and variation
// 4. Meta-learning for rapid adaptation on neuromorphic hardware
//
// Based on research:
// - "Giant TER in HZO-based FTJ" (APL Machine Learning 2025)
// - "HfO2/ZrO2/HfO2 Superlattice FTJ" (Materials Horizons 2024)
// - "FTJ Crossbar with Annealing Optimization" (Adv. Intell. Sys. 2025)
// - "DFA for Memristive SNNs" (Frontiers 2019)
// - "Neuromorphic Processor with On-Chip Learning" (Nature Comms 2025)
//
// Key specifications:
// - FTJ TER: up to 2×10^7, 128-level MLC, 2×10^8 endurance
// - DFA: 2.7× faster training than backprop, comparable accuracy

package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// FERROELECTRIC TUNNEL JUNCTION (FTJ) DEVICE MODEL
// =============================================================================

// FTJConfig configures FTJ device parameters
type FTJConfig struct {
	// Layer structure
	FerroelectricMaterial string  // "HZO", "HZH-SL", "BFO", "PZT"
	FEThicknessNm         float64 // Ferroelectric layer thickness (1-10 nm)
	InterlayerMaterial    string  // "TiO2", "Al2O3", "none"
	InterlayerThicknessNm float64 // Interlayer thickness (0-5 nm)

	// Electrode materials
	TopElectrode          string  // "Pt", "TiN", "Ru"
	BottomElectrode       string  // "NSTO", "TiN", "ITO"

	// Electrical parameters
	PolarizationUCcm2     float64 // Remnant polarization (µC/cm²)
	CoerciveFieldMVcm     float64 // Coercive field (MV/cm)
	BarrierHeightEV       float64 // Tunneling barrier height (eV)

	// TER parameters
	TERRatio              float64 // ON/OFF resistance ratio
	NumLevels             int     // Multi-level states
	SwitchingVoltageV     float64 // Polarization switching voltage

	// Reliability
	EnduranceCycles       float64 // Programming endurance
	RetentionSeconds      float64 // State retention time
	WakeUpCycles          int     // Cycles for wake-up

	// Noise and variation
	CycleVariationPercent float64 // Cycle-to-cycle variation
	DeviceVariationPercent float64 // Device-to-device variation
}

// DefaultFTJConfig returns HZO-based FTJ configuration
func DefaultFTJConfig() *FTJConfig {
	return &FTJConfig{
		FerroelectricMaterial: "HZO",
		FEThicknessNm:         3.0,
		InterlayerMaterial:    "TiO2",
		InterlayerThicknessNm: 1.0,
		TopElectrode:          "Pt",
		BottomElectrode:       "TiN",
		PolarizationUCcm2:     25,
		CoerciveFieldMVcm:     1.5,
		BarrierHeightEV:       1.2,
		TERRatio:              1000,
		NumLevels:             128,
		SwitchingVoltageV:     2.5,
		EnduranceCycles:       2e8,
		RetentionSeconds:      3.15e8, // 10 years
		WakeUpCycles:          100,
		CycleVariationPercent: 2.75,
		DeviceVariationPercent: 5.0,
	}
}

// SuperlatticeConfig for HZH superlattice FTJ
func SuperlatticeConfig() *FTJConfig {
	config := DefaultFTJConfig()
	config.FerroelectricMaterial = "HZH-SL"
	config.FEThicknessNm = 6.0
	config.TERRatio = 1273
	config.NumLevels = 64
	config.PolarizationUCcm2 = 35 // Enhanced by superlattice
	return config
}

// FTJDevice represents a single FTJ memristor
type FTJDevice struct {
	Config           *FTJConfig
	PolarizationState float64 // -1 (down) to +1 (up)
	Resistance       float64 // Current resistance (Ohm)
	ConductanceLevel int     // Discrete level (0 to NumLevels-1)
	TotalCycles      int64   // Total programming cycles
	WakeUpComplete   bool    // Wake-up effect complete
	FatigueLevel     float64 // Fatigue degradation (0-1)
}

// NewFTJDevice creates a new FTJ device
func NewFTJDevice(config *FTJConfig) *FTJDevice {
	if config == nil {
		config = DefaultFTJConfig()
	}

	ftj := &FTJDevice{
		Config:           config,
		PolarizationState: 0,
		WakeUpComplete:   false,
		FatigueLevel:     0,
	}

	ftj.updateResistance()
	return ftj
}

// updateResistance calculates resistance from polarization state
func (ftj *FTJDevice) updateResistance() {
	// TER effect: resistance depends on polarization
	// R = R_off * exp(-α * P) where P is polarization

	// Normalized polarization (0 to 1)
	normalizedP := (ftj.PolarizationState + 1) / 2

	// Calculate resistance
	Roff := 1e8 // High resistance state (Ohm)
	Ron := Roff / ftj.Config.TERRatio

	// Exponential interpolation
	logR := math.Log(Ron) + normalizedP*(math.Log(Roff)-math.Log(Ron))
	ftj.Resistance = math.Exp(logR)

	// Apply wake-up effect (higher variation before wake-up)
	if !ftj.WakeUpComplete && ftj.TotalCycles < int64(ftj.Config.WakeUpCycles) {
		wakeUpFactor := float64(ftj.TotalCycles) / float64(ftj.Config.WakeUpCycles)
		variation := (1 - wakeUpFactor) * 0.3 // 30% variation during wake-up
		ftj.Resistance *= 1 + (rand.Float64()-0.5)*2*variation
	}

	// Apply fatigue
	ftj.Resistance *= 1 + ftj.FatigueLevel*0.5

	// Calculate discrete level
	ftj.ConductanceLevel = int(normalizedP * float64(ftj.Config.NumLevels-1))
	ftj.ConductanceLevel = max(0, min(ftj.Config.NumLevels-1, ftj.ConductanceLevel))
}

// Program sets the polarization state with voltage pulse
func (ftj *FTJDevice) Program(voltage float64, pulseDurationNs float64) {
	// Check if voltage exceeds coercive voltage
	coerciveV := ftj.Config.CoerciveFieldMVcm * ftj.Config.FEThicknessNm * 1e-7 * 10

	if math.Abs(voltage) < coerciveV*0.5 {
		return // Voltage too low for switching
	}

	// Calculate polarization change
	deltaP := 0.0
	if voltage > 0 {
		// Positive voltage: polarization increases
		maxDelta := 1.0 - ftj.PolarizationState
		efficiency := math.Tanh((voltage - coerciveV) / coerciveV)
		deltaP = maxDelta * efficiency * math.Min(1, pulseDurationNs/50)
	} else {
		// Negative voltage: polarization decreases
		maxDelta := ftj.PolarizationState + 1.0
		efficiency := math.Tanh((-voltage - coerciveV) / coerciveV)
		deltaP = -maxDelta * efficiency * math.Min(1, pulseDurationNs/50)
	}

	// Apply cycle-to-cycle variation
	variation := rand.NormFloat64() * ftj.Config.CycleVariationPercent / 100
	deltaP *= 1 + variation

	// Update polarization
	ftj.PolarizationState += deltaP
	ftj.PolarizationState = math.Max(-1, math.Min(1, ftj.PolarizationState))

	// Update cycle count
	ftj.TotalCycles++

	// Check wake-up completion
	if ftj.TotalCycles >= int64(ftj.Config.WakeUpCycles) {
		ftj.WakeUpComplete = true
	}

	// Accumulate fatigue
	ftj.FatigueLevel += 1.0 / ftj.Config.EnduranceCycles

	ftj.updateResistance()
}

// ProgramToLevel programs to a specific conductance level
func (ftj *FTJDevice) ProgramToLevel(level int) {
	if level < 0 || level >= ftj.Config.NumLevels {
		return
	}

	targetP := 2*float64(level)/float64(ftj.Config.NumLevels-1) - 1

	// Iteratively program to target
	for i := 0; i < 10; i++ {
		diff := targetP - ftj.PolarizationState
		if math.Abs(diff) < 0.02 {
			break
		}

		voltage := ftj.Config.SwitchingVoltageV * diff / math.Abs(diff)
		ftj.Program(voltage*math.Min(1, math.Abs(diff)*2), 50)
	}
}

// Read returns the current conductance
func (ftj *FTJDevice) Read() float64 {
	// Add read noise
	noise := rand.NormFloat64() * 0.01 * (1 / ftj.Resistance)
	return 1/ftj.Resistance + noise
}

// GetWeight returns normalized weight (0-1)
func (ftj *FTJDevice) GetWeight() float64 {
	return (ftj.PolarizationState + 1) / 2
}

// SetWeight sets weight (0-1)
func (ftj *FTJDevice) SetWeight(weight float64) {
	level := int(weight * float64(ftj.Config.NumLevels-1))
	ftj.ProgramToLevel(level)
}

// FTJCrossbar represents an FTJ-based crossbar array
type FTJCrossbar struct {
	Config       *FTJConfig
	Rows         int
	Cols         int
	Devices      [][]*FTJDevice
	LineResistance float64 // Parasitic line resistance (Ohm)
	SelfRectifying bool    // Self-rectifying property
}

// NewFTJCrossbar creates an FTJ crossbar array
func NewFTJCrossbar(config *FTJConfig, rows, cols int) *FTJCrossbar {
	if config == nil {
		config = DefaultFTJConfig()
	}

	xbar := &FTJCrossbar{
		Config:         config,
		Rows:           rows,
		Cols:           cols,
		Devices:        make([][]*FTJDevice, rows),
		LineResistance: 10, // 10 Ohm per segment
		SelfRectifying: true,
	}

	for i := 0; i < rows; i++ {
		xbar.Devices[i] = make([]*FTJDevice, cols)
		for j := 0; j < cols; j++ {
			ftj := NewFTJDevice(config)
			// Apply device-to-device variation
			variation := 1 + rand.NormFloat64()*config.DeviceVariationPercent/100
			ftj.Resistance *= variation
			xbar.Devices[i][j] = ftj
		}
	}

	return xbar
}

// SetWeights programs all weights
func (xbar *FTJCrossbar) SetWeights(weights [][]float64) error {
	if len(weights) != xbar.Rows {
		return fmt.Errorf("weight rows %d != xbar rows %d", len(weights), xbar.Rows)
	}

	for i := 0; i < xbar.Rows; i++ {
		if len(weights[i]) != xbar.Cols {
			return fmt.Errorf("weight cols %d != xbar cols %d", len(weights[i]), xbar.Cols)
		}
		for j := 0; j < xbar.Cols; j++ {
			xbar.Devices[i][j].SetWeight(weights[i][j])
		}
	}

	return nil
}

// Forward performs MVM with FTJ array
func (xbar *FTJCrossbar) Forward(input []float64) ([]float64, error) {
	if len(input) != xbar.Rows {
		return nil, fmt.Errorf("input size %d != rows %d", len(input), xbar.Rows)
	}

	output := make([]float64, xbar.Cols)

	for j := 0; j < xbar.Cols; j++ {
		sum := 0.0
		for i := 0; i < xbar.Rows; i++ {
			conductance := xbar.Devices[i][j].Read()

			// Apply self-rectifying behavior
			if xbar.SelfRectifying && input[i] < 0 {
				conductance *= 0.001 // Highly suppressed reverse current
			}

			sum += input[i] * conductance
		}
		output[j] = sum
	}

	return output, nil
}

// GetWeightMatrix returns the weight matrix
func (xbar *FTJCrossbar) GetWeightMatrix() [][]float64 {
	weights := make([][]float64, xbar.Rows)
	for i := 0; i < xbar.Rows; i++ {
		weights[i] = make([]float64, xbar.Cols)
		for j := 0; j < xbar.Cols; j++ {
			weights[i][j] = xbar.Devices[i][j].GetWeight()
		}
	}
	return weights
}

// GetMetrics returns FTJ crossbar metrics
func (xbar *FTJCrossbar) GetMetrics() map[string]float64 {
	totalCycles := int64(0)
	avgFatigue := 0.0
	wakeUpCount := 0

	for i := 0; i < xbar.Rows; i++ {
		for j := 0; j < xbar.Cols; j++ {
			dev := xbar.Devices[i][j]
			totalCycles += dev.TotalCycles
			avgFatigue += dev.FatigueLevel
			if dev.WakeUpComplete {
				wakeUpCount++
			}
		}
	}

	numDevices := float64(xbar.Rows * xbar.Cols)

	return map[string]float64{
		"total_devices":      numDevices,
		"total_cycles":       float64(totalCycles),
		"avg_fatigue":        avgFatigue / numDevices,
		"wakeup_complete":    float64(wakeUpCount) / numDevices,
		"ter_ratio":          xbar.Config.TERRatio,
		"num_levels":         float64(xbar.Config.NumLevels),
		"endurance":          xbar.Config.EnduranceCycles,
	}
}

// =============================================================================
// ON-CHIP LEARNING ALGORITHMS
// =============================================================================

// LearningAlgorithm defines the type of on-chip learning
type LearningAlgorithm int

const (
	AlgorithmBackprop     LearningAlgorithm = iota // Standard backpropagation
	AlgorithmDFA                                   // Direct Feedback Alignment
	AlgorithmSDFA                                  // Sparse DFA
	AlgorithmPerturbation                          // Perturbation-based
	AlgorithmLocal                                 // Local Hebbian learning
	AlgorithmMetaLearn                             // Meta-learning (L2L)
)

// String returns algorithm name
func (a LearningAlgorithm) String() string {
	names := []string{"Backprop", "DFA", "SDFA", "Perturbation", "Local", "MetaLearn"}
	if int(a) < len(names) {
		return names[a]
	}
	return "unknown"
}

// OnChipLearningConfig configures on-chip learning
type OnChipLearningConfig struct {
	Algorithm         LearningAlgorithm
	LearningRate      float64
	BatchSize         int
	NumEpochs         int

	// DFA parameters
	FeedbackSparsity  float64 // Fraction of feedback connections (SDFA)
	RandomSeed        int64

	// Perturbation parameters
	PerturbationScale float64
	PerturbationDecay float64

	// Local learning parameters
	HebbianRate       float64
	AntiHebbianRate   float64

	// Meta-learning parameters
	MetaLearningRate  float64
	NumInnerSteps     int
	TaskBatchSize     int

	// Hardware constraints
	WeightPrecision   int     // Bits for weights
	GradientNoise     float64 // Noise in gradient computation
	UpdateThreshold   float64 // Minimum gradient for update
}

// DefaultOnChipLearningConfig returns default DFA configuration
func DefaultOnChipLearningConfig() *OnChipLearningConfig {
	return &OnChipLearningConfig{
		Algorithm:         AlgorithmDFA,
		LearningRate:      0.01,
		BatchSize:         32,
		NumEpochs:         10,
		FeedbackSparsity:  1.0,
		RandomSeed:        42,
		PerturbationScale: 0.01,
		PerturbationDecay: 0.99,
		HebbianRate:       0.001,
		AntiHebbianRate:   0.0001,
		MetaLearningRate:  0.001,
		NumInnerSteps:     5,
		TaskBatchSize:     5,
		WeightPrecision:   6,
		GradientNoise:     0.05,
		UpdateThreshold:   0.001,
	}
}

// DFATrainer implements Direct Feedback Alignment
type DFATrainer struct {
	Config           *OnChipLearningConfig
	FeedbackMatrices [][][]float64 // Random feedback matrices B
	LayerSizes       []int
	NumOutputs       int
}

// NewDFATrainer creates a DFA trainer
func NewDFATrainer(config *OnChipLearningConfig, layerSizes []int, numOutputs int) *DFATrainer {
	if config == nil {
		config = DefaultOnChipLearningConfig()
	}

	rand.Seed(config.RandomSeed)

	trainer := &DFATrainer{
		Config:           config,
		LayerSizes:       layerSizes,
		NumOutputs:       numOutputs,
		FeedbackMatrices: make([][][]float64, len(layerSizes)),
	}

	// Initialize random feedback matrices
	for l := 0; l < len(layerSizes); l++ {
		trainer.FeedbackMatrices[l] = make([][]float64, layerSizes[l])
		for i := 0; i < layerSizes[l]; i++ {
			trainer.FeedbackMatrices[l][i] = make([]float64, numOutputs)
			for j := 0; j < numOutputs; j++ {
				// Apply sparsity
				if rand.Float64() < config.FeedbackSparsity {
					trainer.FeedbackMatrices[l][i][j] = rand.NormFloat64() * 0.1
				}
			}
		}
	}

	return trainer
}

// ComputeGradient computes gradient using DFA
func (t *DFATrainer) ComputeGradient(layerIdx int, error []float64, activations []float64, inputs []float64) [][]float64 {
	if layerIdx >= len(t.FeedbackMatrices) {
		return nil
	}

	layerSize := t.LayerSizes[layerIdx]
	inputSize := len(inputs)
	gradients := make([][]float64, layerSize)

	// Compute hidden layer error using random feedback
	hiddenError := make([]float64, layerSize)
	for i := 0; i < layerSize; i++ {
		for j := 0; j < len(error); j++ {
			hiddenError[i] += t.FeedbackMatrices[layerIdx][i][j] * error[j]
		}
		// Apply derivative of activation (ReLU)
		if activations[i] <= 0 {
			hiddenError[i] = 0
		}
	}

	// Compute weight gradients
	for i := 0; i < layerSize; i++ {
		gradients[i] = make([]float64, inputSize)
		for j := 0; j < inputSize; j++ {
			grad := hiddenError[i] * inputs[j]

			// Add gradient noise
			grad += rand.NormFloat64() * t.Config.GradientNoise * math.Abs(grad)

			// Apply threshold
			if math.Abs(grad) < t.Config.UpdateThreshold {
				grad = 0
			}

			gradients[i][j] = grad
		}
	}

	return gradients
}

// UpdateWeights applies gradient update to FTJ crossbar
func (t *DFATrainer) UpdateWeights(xbar *FTJCrossbar, gradients [][]float64) {
	for i := 0; i < xbar.Rows && i < len(gradients); i++ {
		for j := 0; j < xbar.Cols && j < len(gradients[i]); j++ {
			currentWeight := xbar.Devices[i][j].GetWeight()
			newWeight := currentWeight - t.Config.LearningRate*gradients[i][j]
			newWeight = math.Max(0, math.Min(1, newWeight))
			xbar.Devices[i][j].SetWeight(newWeight)
		}
	}
}

// PerturbationTrainer implements perturbation-based learning
type PerturbationTrainer struct {
	Config           *OnChipLearningConfig
	PerturbationMask [][]float64
	BaselineOutput   []float64
}

// NewPerturbationTrainer creates a perturbation trainer
func NewPerturbationTrainer(config *OnChipLearningConfig, rows, cols int) *PerturbationTrainer {
	if config == nil {
		config = DefaultOnChipLearningConfig()
	}

	trainer := &PerturbationTrainer{
		Config:           config,
		PerturbationMask: make([][]float64, rows),
	}

	for i := 0; i < rows; i++ {
		trainer.PerturbationMask[i] = make([]float64, cols)
	}

	return trainer
}

// GeneratePerturbation creates random weight perturbation
func (t *PerturbationTrainer) GeneratePerturbation() {
	for i := range t.PerturbationMask {
		for j := range t.PerturbationMask[i] {
			t.PerturbationMask[i][j] = (2*rand.Float64() - 1) * t.Config.PerturbationScale
		}
	}
}

// EstimateGradient estimates gradient from loss difference
func (t *PerturbationTrainer) EstimateGradient(lossPlus, lossMinus float64) [][]float64 {
	gradients := make([][]float64, len(t.PerturbationMask))

	for i := range t.PerturbationMask {
		gradients[i] = make([]float64, len(t.PerturbationMask[i]))
		for j := range t.PerturbationMask[i] {
			// Finite difference gradient estimate
			gradients[i][j] = (lossPlus - lossMinus) / (2 * t.PerturbationMask[i][j])

			// Add noise
			gradients[i][j] += rand.NormFloat64() * t.Config.GradientNoise
		}
	}

	return gradients
}

// LocalLearningTrainer implements local Hebbian learning
type LocalLearningTrainer struct {
	Config    *OnChipLearningConfig
	PreTrace  []float64 // Pre-synaptic activity trace
	PostTrace []float64 // Post-synaptic activity trace
}

// NewLocalLearningTrainer creates a local learning trainer
func NewLocalLearningTrainer(config *OnChipLearningConfig, preSize, postSize int) *LocalLearningTrainer {
	if config == nil {
		config = DefaultOnChipLearningConfig()
	}

	return &LocalLearningTrainer{
		Config:    config,
		PreTrace:  make([]float64, preSize),
		PostTrace: make([]float64, postSize),
	}
}

// UpdateTraces updates activity traces with new activations
func (t *LocalLearningTrainer) UpdateTraces(preAct, postAct []float64, decay float64) {
	for i := range t.PreTrace {
		t.PreTrace[i] = decay*t.PreTrace[i] + (1-decay)*preAct[i]
	}
	for i := range t.PostTrace {
		t.PostTrace[i] = decay*t.PostTrace[i] + (1-decay)*postAct[i]
	}
}

// ComputeHebbianUpdate computes Hebbian weight update
func (t *LocalLearningTrainer) ComputeHebbianUpdate() [][]float64 {
	updates := make([][]float64, len(t.PreTrace))

	for i := range t.PreTrace {
		updates[i] = make([]float64, len(t.PostTrace))
		for j := range t.PostTrace {
			// Hebbian term: strengthen if both active
			hebbian := t.PreTrace[i] * t.PostTrace[j] * t.Config.HebbianRate

			// Anti-Hebbian term: prevent runaway
			antiHebbian := t.PreTrace[i] * t.PostTrace[j] * t.PostTrace[j] * t.Config.AntiHebbianRate

			updates[i][j] = hebbian - antiHebbian
		}
	}

	return updates
}

// MetaLearner implements meta-learning (learning-to-learn)
type MetaLearner struct {
	Config          *OnChipLearningConfig
	MetaWeights     [][]float64 // Meta-learned initialization
	TaskWeights     [][]float64 // Task-specific weights
	LearningRateLR  [][]float64 // Per-weight learning rates
}

// NewMetaLearner creates a meta-learner
func NewMetaLearner(config *OnChipLearningConfig, rows, cols int) *MetaLearner {
	if config == nil {
		config = DefaultOnChipLearningConfig()
	}

	ml := &MetaLearner{
		Config:         config,
		MetaWeights:    make([][]float64, rows),
		TaskWeights:    make([][]float64, rows),
		LearningRateLR: make([][]float64, rows),
	}

	for i := 0; i < rows; i++ {
		ml.MetaWeights[i] = make([]float64, cols)
		ml.TaskWeights[i] = make([]float64, cols)
		ml.LearningRateLR[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			ml.MetaWeights[i][j] = rand.NormFloat64() * 0.1
			ml.LearningRateLR[i][j] = config.LearningRate
		}
	}

	return ml
}

// InitializeForTask copies meta-weights to task weights
func (ml *MetaLearner) InitializeForTask() {
	for i := range ml.MetaWeights {
		copy(ml.TaskWeights[i], ml.MetaWeights[i])
	}
}

// InnerUpdate performs task-specific update
func (ml *MetaLearner) InnerUpdate(gradients [][]float64) {
	for i := range ml.TaskWeights {
		for j := range ml.TaskWeights[i] {
			ml.TaskWeights[i][j] -= ml.LearningRateLR[i][j] * gradients[i][j]
		}
	}
}

// OuterUpdate updates meta-weights from multiple tasks
func (ml *MetaLearner) OuterUpdate(taskGradients [][][]float64) {
	// Average gradients across tasks
	for i := range ml.MetaWeights {
		for j := range ml.MetaWeights[i] {
			avgGrad := 0.0
			for t := range taskGradients {
				if i < len(taskGradients[t]) && j < len(taskGradients[t][i]) {
					avgGrad += taskGradients[t][i][j]
				}
			}
			avgGrad /= float64(len(taskGradients))

			ml.MetaWeights[i][j] -= ml.Config.MetaLearningRate * avgGrad
		}
	}
}

// GetTaskWeights returns current task weights
func (ml *MetaLearner) GetTaskWeights() [][]float64 {
	result := make([][]float64, len(ml.TaskWeights))
	for i := range ml.TaskWeights {
		result[i] = make([]float64, len(ml.TaskWeights[i]))
		copy(result[i], ml.TaskWeights[i])
	}
	return result
}

// =============================================================================
// INTEGRATED ON-CHIP TRAINING SYSTEM
// =============================================================================

// OnChipTrainingSystem integrates FTJ arrays with on-chip learning
type OnChipTrainingSystem struct {
	Config        *OnChipLearningConfig
	FTJConfig     *FTJConfig
	Layers        []*FTJCrossbar
	DFATrainer    *DFATrainer
	PertTrainer   *PerturbationTrainer
	LocalTrainer  *LocalLearningTrainer
	MetaLearner   *MetaLearner

	// Training statistics
	TrainingLoss   []float64
	TrainingAccuracy []float64
	TotalUpdates   int64
}

// OnChipTrainingConfig configures the training system
type OnChipTrainingConfig struct {
	LayerSizes    []int
	FTJConfig     *FTJConfig
	LearningConfig *OnChipLearningConfig
}

// NewOnChipTrainingSystem creates an on-chip training system
func NewOnChipTrainingSystem(config *OnChipTrainingConfig) *OnChipTrainingSystem {
	if config == nil {
		config = &OnChipTrainingConfig{
			LayerSizes:     []int{784, 256, 10},
			FTJConfig:      DefaultFTJConfig(),
			LearningConfig: DefaultOnChipLearningConfig(),
		}
	}

	sys := &OnChipTrainingSystem{
		Config:          config.LearningConfig,
		FTJConfig:       config.FTJConfig,
		Layers:          make([]*FTJCrossbar, len(config.LayerSizes)-1),
		TrainingLoss:    make([]float64, 0),
		TrainingAccuracy: make([]float64, 0),
	}

	// Create FTJ crossbar layers
	for l := 0; l < len(config.LayerSizes)-1; l++ {
		sys.Layers[l] = NewFTJCrossbar(
			config.FTJConfig,
			config.LayerSizes[l],
			config.LayerSizes[l+1],
		)
	}

	// Initialize trainers based on algorithm
	switch config.LearningConfig.Algorithm {
	case AlgorithmDFA, AlgorithmSDFA:
		numOutputs := config.LayerSizes[len(config.LayerSizes)-1]
		hiddenSizes := config.LayerSizes[1 : len(config.LayerSizes)-1]
		sys.DFATrainer = NewDFATrainer(config.LearningConfig, hiddenSizes, numOutputs)

	case AlgorithmPerturbation:
		sys.PertTrainer = NewPerturbationTrainer(
			config.LearningConfig,
			config.LayerSizes[0],
			config.LayerSizes[1],
		)

	case AlgorithmLocal:
		sys.LocalTrainer = NewLocalLearningTrainer(
			config.LearningConfig,
			config.LayerSizes[0],
			config.LayerSizes[1],
		)

	case AlgorithmMetaLearn:
		sys.MetaLearner = NewMetaLearner(
			config.LearningConfig,
			config.LayerSizes[0],
			config.LayerSizes[1],
		)
	}

	return sys
}

// Forward performs forward pass through all layers
func (sys *OnChipTrainingSystem) Forward(input []float64) ([][]float64, error) {
	activations := make([][]float64, len(sys.Layers)+1)
	activations[0] = input

	current := input
	for l, layer := range sys.Layers {
		output, err := layer.Forward(current)
		if err != nil {
			return nil, err
		}

		// Apply ReLU activation (except last layer)
		if l < len(sys.Layers)-1 {
			for i := range output {
				if output[i] < 0 {
					output[i] = 0
				}
			}
		}

		activations[l+1] = output
		current = output
	}

	return activations, nil
}

// TrainStep performs one training step
func (sys *OnChipTrainingSystem) TrainStep(input []float64, target []float64) (float64, error) {
	// Forward pass
	activations, err := sys.Forward(input)
	if err != nil {
		return 0, err
	}

	// Compute output error
	output := activations[len(activations)-1]
	error := make([]float64, len(output))
	loss := 0.0

	for i := range output {
		error[i] = output[i] - target[i]
		loss += error[i] * error[i]
	}
	loss /= 2

	// Backward pass based on algorithm
	switch sys.Config.Algorithm {
	case AlgorithmDFA, AlgorithmSDFA:
		sys.trainDFA(activations, error)

	case AlgorithmPerturbation:
		sys.trainPerturbation(input, target, loss)

	case AlgorithmLocal:
		sys.trainLocal(activations)

	case AlgorithmBackprop:
		sys.trainBackprop(activations, error)
	}

	sys.TotalUpdates++
	return loss, nil
}

// trainDFA performs DFA training step
func (sys *OnChipTrainingSystem) trainDFA(activations [][]float64, error []float64) {
	// Train hidden layers using random feedback
	for l := len(sys.Layers) - 2; l >= 0; l-- {
		gradients := sys.DFATrainer.ComputeGradient(
			l,
			error,
			activations[l+1],
			activations[l],
		)
		sys.DFATrainer.UpdateWeights(sys.Layers[l], gradients)
	}

	// Train output layer with direct error
	outputGradients := make([][]float64, len(activations[len(activations)-2]))
	for i := range outputGradients {
		outputGradients[i] = make([]float64, len(error))
		for j := range error {
			outputGradients[i][j] = activations[len(activations)-2][i] * error[j]
		}
	}
	sys.DFATrainer.UpdateWeights(sys.Layers[len(sys.Layers)-1], outputGradients)
}

// trainPerturbation performs perturbation-based training
func (sys *OnChipTrainingSystem) trainPerturbation(input, target []float64, baseLoss float64) {
	// Apply positive perturbation
	sys.PertTrainer.GeneratePerturbation()
	weights := sys.Layers[0].GetWeightMatrix()

	// Perturb weights positively
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] += sys.PertTrainer.PerturbationMask[i][j]
		}
	}
	sys.Layers[0].SetWeights(weights)

	activations, _ := sys.Forward(input)
	output := activations[len(activations)-1]
	lossPlus := 0.0
	for i := range output {
		diff := output[i] - target[i]
		lossPlus += diff * diff
	}
	lossPlus /= 2

	// Perturb weights negatively
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] -= 2 * sys.PertTrainer.PerturbationMask[i][j]
		}
	}
	sys.Layers[0].SetWeights(weights)

	activations, _ = sys.Forward(input)
	output = activations[len(activations)-1]
	lossMinus := 0.0
	for i := range output {
		diff := output[i] - target[i]
		lossMinus += diff * diff
	}
	lossMinus /= 2

	// Estimate gradients and update
	gradients := sys.PertTrainer.EstimateGradient(lossPlus, lossMinus)

	for i := range weights {
		for j := range weights[i] {
			weights[i][j] += sys.PertTrainer.PerturbationMask[i][j] // Restore
			weights[i][j] -= sys.Config.LearningRate * gradients[i][j]
		}
	}
	sys.Layers[0].SetWeights(weights)
}

// trainLocal performs local Hebbian training
func (sys *OnChipTrainingSystem) trainLocal(activations [][]float64) {
	for l := 0; l < len(sys.Layers); l++ {
		sys.LocalTrainer.UpdateTraces(activations[l], activations[l+1], 0.9)
		updates := sys.LocalTrainer.ComputeHebbianUpdate()

		weights := sys.Layers[l].GetWeightMatrix()
		for i := range weights {
			for j := range weights[i] {
				if i < len(updates) && j < len(updates[i]) {
					weights[i][j] += updates[i][j]
					weights[i][j] = math.Max(0, math.Min(1, weights[i][j]))
				}
			}
		}
		sys.Layers[l].SetWeights(weights)
	}
}

// trainBackprop performs standard backpropagation (for comparison)
func (sys *OnChipTrainingSystem) trainBackprop(activations [][]float64, error []float64) {
	// Simplified backprop implementation
	currentError := error

	for l := len(sys.Layers) - 1; l >= 0; l-- {
		gradients := make([][]float64, len(activations[l]))
		for i := range gradients {
			gradients[i] = make([]float64, len(currentError))
			for j := range currentError {
				gradients[i][j] = activations[l][i] * currentError[j]
			}
		}

		// Update weights
		weights := sys.Layers[l].GetWeightMatrix()
		for i := range weights {
			for j := range weights[i] {
				if i < len(gradients) && j < len(gradients[i]) {
					weights[i][j] -= sys.Config.LearningRate * gradients[i][j]
					weights[i][j] = math.Max(0, math.Min(1, weights[i][j]))
				}
			}
		}
		sys.Layers[l].SetWeights(weights)

		// Propagate error (simplified)
		if l > 0 {
			newError := make([]float64, len(activations[l]))
			for i := range newError {
				for j := range currentError {
					newError[i] += weights[i][j] * currentError[j]
				}
				// ReLU derivative
				if activations[l][i] <= 0 {
					newError[i] = 0
				}
			}
			currentError = newError
		}
	}
}

// GetMetrics returns training system metrics
func (sys *OnChipTrainingSystem) GetMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	metrics["algorithm"] = sys.Config.Algorithm.String()
	metrics["total_updates"] = sys.TotalUpdates
	metrics["num_layers"] = len(sys.Layers)

	if len(sys.TrainingLoss) > 0 {
		metrics["last_loss"] = sys.TrainingLoss[len(sys.TrainingLoss)-1]
	}

	// Aggregate FTJ metrics
	totalDevices := 0
	totalCycles := int64(0)
	avgFatigue := 0.0

	for _, layer := range sys.Layers {
		layerMetrics := layer.GetMetrics()
		totalDevices += int(layerMetrics["total_devices"])
		totalCycles += int64(layerMetrics["total_cycles"])
		avgFatigue += layerMetrics["avg_fatigue"] * layerMetrics["total_devices"]
	}

	metrics["total_ftj_devices"] = totalDevices
	metrics["total_ftj_cycles"] = totalCycles
	if totalDevices > 0 {
		metrics["avg_ftj_fatigue"] = avgFatigue / float64(totalDevices)
	}

	return metrics
}

// =============================================================================
// DEMO AND VISUALIZATION
// =============================================================================

// FTJDeviceDemo demonstrates FTJ device behavior
func FTJDeviceDemo() string {
	result := "FTJ Device Characteristics Demo\n"
	result += "================================\n\n"

	config := DefaultFTJConfig()
	ftj := NewFTJDevice(config)

	result += fmt.Sprintf("Configuration:\n")
	result += fmt.Sprintf("  Material: %s\n", config.FerroelectricMaterial)
	result += fmt.Sprintf("  FE thickness: %.1f nm\n", config.FEThicknessNm)
	result += fmt.Sprintf("  TER ratio: %.0f\n", config.TERRatio)
	result += fmt.Sprintf("  Levels: %d\n", config.NumLevels)
	result += fmt.Sprintf("  Endurance: %.0e cycles\n\n", config.EnduranceCycles)

	// Program to different levels
	result += "Multi-level Programming:\n"
	levels := []int{0, 32, 64, 96, 127}
	for _, level := range levels {
		ftj.ProgramToLevel(level)
		result += fmt.Sprintf("  Level %3d: R=%.2e Ω, G=%.2e S, Weight=%.3f\n",
			level, ftj.Resistance, 1/ftj.Resistance, ftj.GetWeight())
	}

	result += "\nWake-up Effect Simulation:\n"
	ftj2 := NewFTJDevice(config)
	for cycle := 0; cycle < 150; cycle += 25 {
		ftj2.Program(3.0, 50)
		ftj2.Program(-3.0, 50)
		result += fmt.Sprintf("  Cycle %3d: WakeUp=%v, P=%.3f\n",
			cycle, ftj2.WakeUpComplete, ftj2.PolarizationState)
	}

	return result
}

// DFAvsBackpropDemo compares DFA and backprop
func DFAvsBackpropDemo() string {
	result := "DFA vs Backprop Comparison\n"
	result += "===========================\n\n"

	// Create two identical systems
	configDFA := &OnChipTrainingConfig{
		LayerSizes:     []int{16, 8, 4},
		FTJConfig:      DefaultFTJConfig(),
		LearningConfig: DefaultOnChipLearningConfig(),
	}
	configDFA.LearningConfig.Algorithm = AlgorithmDFA

	configBP := &OnChipTrainingConfig{
		LayerSizes:     []int{16, 8, 4},
		FTJConfig:      DefaultFTJConfig(),
		LearningConfig: DefaultOnChipLearningConfig(),
	}
	configBP.LearningConfig.Algorithm = AlgorithmBackprop

	sysDFA := NewOnChipTrainingSystem(configDFA)
	sysBP := NewOnChipTrainingSystem(configBP)

	// Train on synthetic data
	result += "Training (10 steps each):\n\n"

	for step := 0; step < 10; step++ {
		// Random input and target
		input := make([]float64, 16)
		target := make([]float64, 4)
		for i := range input {
			input[i] = rand.Float64()
		}
		targetClass := rand.Intn(4)
		target[targetClass] = 1.0

		lossDFA, _ := sysDFA.TrainStep(input, target)
		lossBP, _ := sysBP.TrainStep(input, target)

		result += fmt.Sprintf("Step %2d: DFA Loss=%.4f, BP Loss=%.4f\n",
			step+1, lossDFA, lossBP)
	}

	result += "\nKey Benefits of DFA:\n"
	result += "  • No weight transport problem\n"
	result += "  • Local computation only\n"
	result += "  • 2.7× faster training than backprop\n"
	result += "  • Fixed random feedback matrices\n"
	result += "  • Hardware-friendly implementation\n"

	return result
}

// OnChipLearningDemo demonstrates different algorithms
func OnChipLearningDemo() string {
	result := "On-Chip Learning Algorithms Demo\n"
	result += "==================================\n\n"

	algorithms := []LearningAlgorithm{
		AlgorithmDFA,
		AlgorithmSDFA,
		AlgorithmPerturbation,
		AlgorithmLocal,
	}

	for _, algo := range algorithms {
		config := &OnChipTrainingConfig{
			LayerSizes:     []int{16, 8, 4},
			FTJConfig:      DefaultFTJConfig(),
			LearningConfig: DefaultOnChipLearningConfig(),
		}
		config.LearningConfig.Algorithm = algo

		if algo == AlgorithmSDFA {
			config.LearningConfig.FeedbackSparsity = 0.3 // 30% connections
		}

		sys := NewOnChipTrainingSystem(config)

		// Train 5 steps
		totalLoss := 0.0
		for step := 0; step < 5; step++ {
			input := make([]float64, 16)
			target := make([]float64, 4)
			for i := range input {
				input[i] = rand.Float64()
			}
			target[rand.Intn(4)] = 1.0

			loss, _ := sys.TrainStep(input, target)
			totalLoss += loss
		}

		metrics := sys.GetMetrics()
		result += fmt.Sprintf("%s:\n", algo.String())
		result += fmt.Sprintf("  Avg Loss: %.4f\n", totalLoss/5)
		result += fmt.Sprintf("  FTJ Cycles: %d\n", metrics["total_ftj_cycles"])
		result += fmt.Sprintf("  Devices: %d\n\n", metrics["total_ftj_devices"])
	}

	return result
}

// FTJCrossbarDemo demonstrates FTJ crossbar for inference
func FTJCrossbarDemo() string {
	result := "FTJ Crossbar Array Demo\n"
	result += "========================\n\n"

	config := SuperlatticeConfig()
	xbar := NewFTJCrossbar(config, 8, 4)

	result += fmt.Sprintf("Configuration: HZH Superlattice FTJ\n")
	result += fmt.Sprintf("Array size: %dx%d\n", xbar.Rows, xbar.Cols)
	result += fmt.Sprintf("TER ratio: %.0f\n", config.TERRatio)
	result += fmt.Sprintf("Self-rectifying: %v\n\n", xbar.SelfRectifying)

	// Set random weights
	weights := make([][]float64, 8)
	for i := range weights {
		weights[i] = make([]float64, 4)
		for j := range weights[i] {
			weights[i][j] = rand.Float64()
		}
	}
	xbar.SetWeights(weights)

	result += "Weight Matrix:\n"
	for i := 0; i < 4; i++ {
		result += fmt.Sprintf("  Row %d: ", i)
		for j := 0; j < 4; j++ {
			result += fmt.Sprintf("%.2f ", weights[i][j])
		}
		result += "...\n"
	}

	// Perform MVM
	input := make([]float64, 8)
	for i := range input {
		input[i] = rand.Float64()
	}

	output, _ := xbar.Forward(input)

	result += "\nMVM Operation:\n"
	result += fmt.Sprintf("  Input: [%.2f, %.2f, %.2f, %.2f, ...]\n",
		input[0], input[1], input[2], input[3])
	result += fmt.Sprintf("  Output: [%.4f, %.4f, %.4f, %.4f]\n",
		output[0], output[1], output[2], output[3])

	// Metrics
	metrics := xbar.GetMetrics()
	result += "\nMetrics:\n"
	result += fmt.Sprintf("  Total devices: %.0f\n", metrics["total_devices"])
	result += fmt.Sprintf("  Wake-up complete: %.1f%%\n", metrics["wakeup_complete"]*100)
	result += fmt.Sprintf("  Avg fatigue: %.6f\n", metrics["avg_fatigue"])

	return result
}

// Serialize returns JSON representation
func (sys *OnChipTrainingSystem) Serialize() ([]byte, error) {
	data := map[string]interface{}{
		"algorithm":    sys.Config.Algorithm.String(),
		"num_layers":   len(sys.Layers),
		"ftj_material": sys.FTJConfig.FerroelectricMaterial,
		"metrics":      sys.GetMetrics(),
	}
	return json.MarshalIndent(data, "", "  ")
}

// FTJComparisonTable generates comparison table
func FTJComparisonTable() string {
	return `
┌─────────────────────────────────────────────────────────────────────────────┐
│         Ferroelectric Tunnel Junction (FTJ) Technology Comparison            │
├─────────────────────┬───────────┬───────────┬───────────┬───────────────────┤
│ Parameter           │ HZO FTJ   │ HZH-SL FTJ│ BFO FTJ   │ PZT FTJ           │
├─────────────────────┼───────────┼───────────┼───────────┼───────────────────┤
│ TER Ratio           │ 10³-10⁴   │ 1.3×10³   │ 10⁷       │ 10⁵               │
│ FE Thickness        │ 2-5 nm    │ 6 nm      │ 2-4 nm    │ 3-5 nm            │
│ MLC Levels          │ 128       │ 64        │ 32        │ 16                │
│ Endurance           │ 2×10⁸     │ 10⁸       │ 10⁶       │ 10⁶               │
│ Retention           │ 10+ yrs   │ 10+ yrs   │ 10+ yrs   │ 10+ yrs           │
│ Switching Speed     │ 50 ns     │ 100 ns    │ 100 ns    │ 200 ns            │
│ CMOS Compatible     │ ✓         │ ✓         │ Partial   │ ✗                 │
│ Self-Rectifying     │ ✓         │ ✓         │ ✗         │ ✗                 │
├─────────────────────┴───────────┴───────────┴───────────┴───────────────────┤
│ IronLattice Target: HZO and HZH Superlattice (CMOS-compatible, high TER)    │
├─────────────────────────────────────────────────────────────────────────────┤
│ Key Advantages:                                                              │
│ • Non-filamentary switching (reliable MLC)                                  │
│ • Low operating current (10 nA - 1 µA)                                      │
│ • 4F² cell size (ultra-dense)                                               │
│ • Inherent self-rectifying (no selector needed)                             │
│ • Symmetric weight update (ideal for training)                              │
└─────────────────────────────────────────────────────────────────────────────┘
`
}

// OnChipLearningComparisonTable generates algorithm comparison
func OnChipLearningComparisonTable() string {
	return `
┌─────────────────────────────────────────────────────────────────────────────┐
│         On-Chip Learning Algorithm Comparison                                │
├─────────────────────┬───────────┬───────────┬───────────┬───────────────────┤
│ Algorithm           │ Backprop  │ DFA       │ SDFA      │ Perturbation      │
├─────────────────────┼───────────┼───────────┼───────────┼───────────────────┤
│ Weight Transport    │ Required  │ Not Req   │ Not Req   │ Not Req           │
│ Locality            │ Global    │ Semi-local│ Local     │ Local             │
│ Accuracy (MNIST)    │ 98.5%     │ 98.2%     │ 97.8%     │ 96.5%             │
│ Training Speed      │ 1×        │ 2.7×      │ 3.5×      │ 0.5×              │
│ Memory Overhead     │ High      │ Medium    │ Low       │ Low               │
│ Hardware Friendly   │ ✗         │ ✓         │ ✓✓        │ ✓                 │
│ Biological Plausible│ ✗         │ Partial   │ Partial   │ ✓                 │
├─────────────────────┴───────────┴───────────┴───────────┴───────────────────┤
│ DFA Key Insight: Random feedback works because neural networks learn to     │
│ align weight updates with the fixed random matrix during training.          │
├─────────────────────────────────────────────────────────────────────────────┤
│ Hardware Implementation:                                                     │
│ • DFA feedback matrices can be stored in fixed FTJ crossbars                │
│ • No need to update feedback weights → simpler hardware                     │
│ • Compatible with noisy analog computation                                   │
│ • Enables true on-chip learning without external processor                  │
└─────────────────────────────────────────────────────────────────────────────┘
`
}

// IronLatticeFTJOnChipSystem integrates all components
type IronLatticeFTJOnChipSystem struct {
	TrainingSystem *OnChipTrainingSystem
	InferenceXbar  *FTJCrossbar
	Mode           string // "training", "inference", "hybrid"
}

// NewIronLatticeFTJOnChipSystem creates integrated system
func NewIronLatticeFTJOnChipSystem(layerSizes []int) *IronLatticeFTJOnChipSystem {
	config := &OnChipTrainingConfig{
		LayerSizes:     layerSizes,
		FTJConfig:      SuperlatticeConfig(),
		LearningConfig: DefaultOnChipLearningConfig(),
	}

	return &IronLatticeFTJOnChipSystem{
		TrainingSystem: NewOnChipTrainingSystem(config),
		Mode:           "hybrid",
	}
}

// GetSystemSummary returns system summary
func (sys *IronLatticeFTJOnChipSystem) GetSystemSummary() string {
	metrics := sys.TrainingSystem.GetMetrics()

	return fmt.Sprintf(`IronLattice FTJ On-Chip Learning System
========================================
Mode: %s
Algorithm: %s
Layers: %d
Total FTJ Devices: %d
Total Cycles: %d
FTJ Material: HZH Superlattice
TER Ratio: 1273
MLC Levels: 64
Training Speed: 2.7× faster than backprop
`,
		sys.Mode,
		metrics["algorithm"],
		metrics["num_layers"],
		metrics["total_ftj_devices"],
		metrics["total_ftj_cycles"],
	)
}

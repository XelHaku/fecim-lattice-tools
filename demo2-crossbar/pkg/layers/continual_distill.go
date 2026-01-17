// Package layers provides continual learning and knowledge distillation for CIM accelerators.
//
// This module implements techniques to prevent catastrophic forgetting and enable
// model compression for edge deployment on ferroelectric compute-in-memory systems.
//
// Key features:
// - Elastic Weight Consolidation (EWC) for sequential task learning
// - Metaplasticity-based learning with probabilistic synaptic updates
// - Knowledge distillation (teacher-student networks)
// - PQK compression pipeline (Pruning, Quantization, Knowledge Distillation)
//
// References:
// - "Probabilistic metaplasticity for catastrophic forgetting" (Scientific Reports 2024)
// - "Electrochemical ohmic memristors for neuromorphic" (Nature Communications 2025)
// - "PQK: Pruning, Quantization, Knowledge Distillation" (Edge AI 2024)
// - "Lifelong learning in neuromorphic systems" (Nature Reviews 2025)
package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// Elastic Weight Consolidation (EWC)
// =============================================================================

// FisherInformation represents diagonal Fisher information matrix for a layer.
// Fisher information measures the importance of each weight for previous tasks.
type FisherInformation struct {
	Diagonal    []float64 // F_ii diagonal elements
	LayerName   string
	NumSamples  int     // samples used for estimation
	Temperature float64 // softmax temperature for estimation
}

// EWCConfig configures Elastic Weight Consolidation.
type EWCConfig struct {
	Lambda          float64 // regularization strength (100-1000 typical)
	FisherSamples   int     // samples for Fisher estimation (1000 typical)
	OnlineEWC       bool    // online EWC with running Fisher average
	GammaDecay      float64 // decay factor for online EWC (0.9 typical)
	NormalizeByTask bool    // normalize Fisher by number of tasks
}

// DefaultEWCConfig returns standard EWC configuration.
func DefaultEWCConfig() EWCConfig {
	return EWCConfig{
		Lambda:          400.0,
		FisherSamples:   1000,
		OnlineEWC:       true,
		GammaDecay:      0.9,
		NormalizeByTask: true,
	}
}

// EWCRegularizer implements Elastic Weight Consolidation for continual learning.
type EWCRegularizer struct {
	Config          EWCConfig
	FisherMatrices  map[string]*FisherInformation
	StarWeights     map[string][]float64 // optimal weights after each task
	TaskCount       int
	ConsolidatedF   map[string][]float64 // consolidated Fisher (online EWC)
}

// NewEWCRegularizer creates an EWC regularizer.
func NewEWCRegularizer(config EWCConfig) *EWCRegularizer {
	return &EWCRegularizer{
		Config:         config,
		FisherMatrices: make(map[string]*FisherInformation),
		StarWeights:    make(map[string][]float64),
		ConsolidatedF:  make(map[string][]float64),
	}
}

// EstimateFisher computes Fisher information for current task.
// Uses output gradients squared as diagonal Fisher approximation.
func (e *EWCRegularizer) EstimateFisher(
	layerName string,
	weights []float64,
	gradientFunc func(weights []float64) []float64,
) *FisherInformation {
	fisher := &FisherInformation{
		Diagonal:    make([]float64, len(weights)),
		LayerName:   layerName,
		NumSamples:  e.Config.FisherSamples,
		Temperature: 1.0,
	}

	// Monte Carlo estimation of Fisher diagonal
	for i := 0; i < e.Config.FisherSamples; i++ {
		grads := gradientFunc(weights)
		for j, g := range grads {
			fisher.Diagonal[j] += g * g
		}
	}

	// Average over samples
	for j := range fisher.Diagonal {
		fisher.Diagonal[j] /= float64(e.Config.FisherSamples)
	}

	return fisher
}

// RegisterTask records weights and Fisher after completing a task.
func (e *EWCRegularizer) RegisterTask(layerName string, weights []float64, fisher *FisherInformation) {
	// Store optimal weights
	e.StarWeights[layerName] = make([]float64, len(weights))
	copy(e.StarWeights[layerName], weights)

	if e.Config.OnlineEWC {
		// Online EWC: consolidate Fisher matrices
		if _, exists := e.ConsolidatedF[layerName]; !exists {
			e.ConsolidatedF[layerName] = make([]float64, len(weights))
		}

		for i := range weights {
			e.ConsolidatedF[layerName][i] = e.Config.GammaDecay*e.ConsolidatedF[layerName][i] +
				fisher.Diagonal[i]
		}
	} else {
		// Standard EWC: store Fisher directly
		e.FisherMatrices[layerName] = fisher
	}

	e.TaskCount++
}

// ComputePenalty returns EWC regularization loss.
func (e *EWCRegularizer) ComputePenalty(layerName string, currentWeights []float64) float64 {
	starW, exists := e.StarWeights[layerName]
	if !exists {
		return 0.0
	}

	var penalty float64
	var fisher []float64

	if e.Config.OnlineEWC {
		fisher = e.ConsolidatedF[layerName]
	} else if f, ok := e.FisherMatrices[layerName]; ok {
		fisher = f.Diagonal
	} else {
		return 0.0
	}

	for i := range currentWeights {
		diff := currentWeights[i] - starW[i]
		penalty += fisher[i] * diff * diff
	}

	lambda := e.Config.Lambda
	if e.Config.NormalizeByTask && e.TaskCount > 0 {
		lambda /= float64(e.TaskCount)
	}

	return 0.5 * lambda * penalty
}

// ComputeGradient returns gradient of EWC penalty w.r.t. weights.
func (e *EWCRegularizer) ComputeGradient(layerName string, currentWeights []float64) []float64 {
	grad := make([]float64, len(currentWeights))

	starW, exists := e.StarWeights[layerName]
	if !exists {
		return grad
	}

	var fisher []float64
	if e.Config.OnlineEWC {
		fisher = e.ConsolidatedF[layerName]
	} else if f, ok := e.FisherMatrices[layerName]; ok {
		fisher = f.Diagonal
	} else {
		return grad
	}

	lambda := e.Config.Lambda
	if e.Config.NormalizeByTask && e.TaskCount > 0 {
		lambda /= float64(e.TaskCount)
	}

	for i := range currentWeights {
		grad[i] = lambda * fisher[i] * (currentWeights[i] - starW[i])
	}

	return grad
}

// =============================================================================
// Metaplasticity-Based Learning
// =============================================================================

// MetaplasticityConfig configures metaplasticity parameters.
type MetaplasticityConfig struct {
	InitialThreshold   float64 // initial plasticity threshold
	ThresholdDecay     float64 // decay rate for threshold adaptation
	ConsolidationRate  float64 // rate at which weights consolidate
	ProbabilisticMode  bool    // use probabilistic updates
	TemperatureInit    float64 // initial sampling temperature
	TemperatureDecay   float64 // temperature annealing
	MaxConsolidation   float64 // maximum consolidation level
}

// DefaultMetaplasticityConfig returns standard metaplasticity settings.
func DefaultMetaplasticityConfig() MetaplasticityConfig {
	return MetaplasticityConfig{
		InitialThreshold:  0.5,
		ThresholdDecay:    0.99,
		ConsolidationRate: 0.01,
		ProbabilisticMode: true,
		TemperatureInit:   1.0,
		TemperatureDecay:  0.995,
		MaxConsolidation:  0.95,
	}
}

// SynapticState tracks metaplasticity state for each synapse.
type SynapticState struct {
	Weight           float64
	Consolidation    float64 // how consolidated (0=plastic, 1=fixed)
	PlasticityThresh float64 // threshold for accepting updates
	UpdateCount      int     // number of successful updates
	LastGradient     float64 // most recent gradient
	CascadeLevel     int     // cascade metaplasticity level
}

// MetaplasticNetwork implements metaplasticity-based continual learning.
type MetaplasticNetwork struct {
	Config      MetaplasticityConfig
	Synapses    map[string][]*SynapticState
	Temperature float64
	Epoch       int
	TaskID      int
	RNG         *rand.Rand
}

// NewMetaplasticNetwork creates a metaplasticity-enabled network.
func NewMetaplasticNetwork(config MetaplasticityConfig, seed int64) *MetaplasticNetwork {
	return &MetaplasticNetwork{
		Config:      config,
		Synapses:    make(map[string][]*SynapticState),
		Temperature: config.TemperatureInit,
		RNG:         rand.New(rand.NewSource(seed)),
	}
}

// InitializeLayer sets up metaplastic synapses for a layer.
func (m *MetaplasticNetwork) InitializeLayer(layerName string, weights []float64) {
	synapses := make([]*SynapticState, len(weights))
	for i, w := range weights {
		synapses[i] = &SynapticState{
			Weight:           w,
			Consolidation:    0.0,
			PlasticityThresh: m.Config.InitialThreshold,
			CascadeLevel:     0,
		}
	}
	m.Synapses[layerName] = synapses
}

// ProbabilisticUpdate applies metaplastic weight update.
func (m *MetaplasticNetwork) ProbabilisticUpdate(layerName string, gradients []float64, lr float64) {
	synapses, exists := m.Synapses[layerName]
	if !exists {
		return
	}

	for i, syn := range synapses {
		grad := gradients[i]
		syn.LastGradient = grad

		// Effective learning rate modulated by plasticity
		effectiveLR := lr * (1.0 - syn.Consolidation)

		if m.Config.ProbabilisticMode {
			// Probabilistic update based on gradient magnitude
			updateProb := m.sigmoid(math.Abs(grad)/m.Temperature - syn.PlasticityThresh)

			if m.RNG.Float64() < updateProb {
				syn.Weight -= effectiveLR * grad
				syn.UpdateCount++

				// Increase consolidation after successful update
				syn.Consolidation = math.Min(
					syn.Consolidation+m.Config.ConsolidationRate,
					m.Config.MaxConsolidation,
				)

				// Cascade metaplasticity: increase level on update
				syn.CascadeLevel = min(syn.CascadeLevel+1, 10)
			} else {
				// Failed update: decrease cascade level
				syn.CascadeLevel = max(syn.CascadeLevel-1, 0)
			}
		} else {
			// Deterministic metaplasticity
			if math.Abs(grad) > syn.PlasticityThresh {
				syn.Weight -= effectiveLR * grad
				syn.UpdateCount++
				syn.Consolidation = math.Min(
					syn.Consolidation+m.Config.ConsolidationRate,
					m.Config.MaxConsolidation,
				)
			}
		}

		// Adapt threshold based on cascade level
		levelFactor := 1.0 + float64(syn.CascadeLevel)*0.1
		syn.PlasticityThresh *= m.Config.ThresholdDecay * levelFactor
	}

	// Anneal temperature
	m.Temperature *= m.Config.TemperatureDecay
}

func (m *MetaplasticNetwork) sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// GetWeights extracts current weights from metaplastic synapses.
func (m *MetaplasticNetwork) GetWeights(layerName string) []float64 {
	synapses, exists := m.Synapses[layerName]
	if !exists {
		return nil
	}
	weights := make([]float64, len(synapses))
	for i, s := range synapses {
		weights[i] = s.Weight
	}
	return weights
}

// GetConsolidationMap returns consolidation levels for visualization.
func (m *MetaplasticNetwork) GetConsolidationMap(layerName string) []float64 {
	synapses, exists := m.Synapses[layerName]
	if !exists {
		return nil
	}
	consolidation := make([]float64, len(synapses))
	for i, s := range synapses {
		consolidation[i] = s.Consolidation
	}
	return consolidation
}

// StartNewTask prepares network for learning new task.
func (m *MetaplasticNetwork) StartNewTask() {
	m.TaskID++
	m.Temperature = m.Config.TemperatureInit

	// Reset plasticity thresholds for unconsolidated synapses
	for _, synapses := range m.Synapses {
		for _, syn := range synapses {
			if syn.Consolidation < 0.5 {
				syn.PlasticityThresh = m.Config.InitialThreshold
			}
		}
	}
}

// =============================================================================
// Memristive Metaplasticity (Hardware-Aware)
// =============================================================================

// MemristorMetaplasticConfig configures memristor-specific metaplasticity.
type MemristorMetaplasticConfig struct {
	ConductanceMin    float64 // Gmin
	ConductanceMax    float64 // Gmax
	SetVoltage        float64 // V_SET threshold
	ResetVoltage      float64 // V_RESET threshold
	FilamentGrowth    float64 // filament growth rate
	FilamentDecay     float64 // filament decay rate
	MultifilamentMode bool    // enable multi-filament stochasticity
	NumFilaments      int     // number of parallel filaments
}

// DefaultMemristorMetaplasticConfig returns standard memristor settings.
func DefaultMemristorMetaplasticConfig() MemristorMetaplasticConfig {
	return MemristorMetaplasticConfig{
		ConductanceMin:    1e-6, // 1 µS
		ConductanceMax:    1e-3, // 1 mS
		SetVoltage:        0.8,
		ResetVoltage:      -0.6,
		FilamentGrowth:    0.1,
		FilamentDecay:     0.05,
		MultifilamentMode: true,
		NumFilaments:      5,
	}
}

// MemristorSynapse models a physical memristive synapse with metaplasticity.
type MemristorSynapse struct {
	Conductance       float64
	FilamentStrength  float64   // filament formation level (0-1)
	FilamentStates    []float64 // multi-filament states
	ProgramCycles     int       // endurance tracking
	RetentionDecay    float64   // retention time constant
	LastProgramTime   float64   // for retention modeling
}

// MemristorArray implements metaplastic memristor crossbar.
type MemristorArray struct {
	Config   MemristorMetaplasticConfig
	Synapses [][]*MemristorSynapse
	Rows     int
	Cols     int
	RNG      *rand.Rand
}

// NewMemristorArray creates a metaplastic memristor array.
func NewMemristorArray(rows, cols int, config MemristorMetaplasticConfig, seed int64) *MemristorArray {
	arr := &MemristorArray{
		Config:   config,
		Rows:     rows,
		Cols:     cols,
		RNG:      rand.New(rand.NewSource(seed)),
		Synapses: make([][]*MemristorSynapse, rows),
	}

	// Initialize synapses
	for i := 0; i < rows; i++ {
		arr.Synapses[i] = make([]*MemristorSynapse, cols)
		for j := 0; j < cols; j++ {
			syn := &MemristorSynapse{
				Conductance:      (config.ConductanceMin + config.ConductanceMax) / 2,
				FilamentStrength: 0.5,
				RetentionDecay:   1e6, // 1M seconds typical
			}
			if config.MultifilamentMode {
				syn.FilamentStates = make([]float64, config.NumFilaments)
				for k := range syn.FilamentStates {
					syn.FilamentStates[k] = arr.RNG.Float64()
				}
			}
			arr.Synapses[i][j] = syn
		}
	}

	return arr
}

// ApplyVoltage programs synapse with voltage pulse.
func (m *MemristorArray) ApplyVoltage(row, col int, voltage, pulseWidth float64) {
	syn := m.Synapses[row][col]

	if voltage > m.Config.SetVoltage {
		// SET operation - increase conductance
		deltaG := m.Config.FilamentGrowth * (voltage - m.Config.SetVoltage) * pulseWidth

		if m.Config.MultifilamentMode {
			// Stochastic multi-filament update
			for k := range syn.FilamentStates {
				if m.RNG.Float64() < syn.FilamentStates[k] {
					syn.FilamentStates[k] = math.Min(syn.FilamentStates[k]+deltaG, 1.0)
				}
			}
			// Conductance = sum of filament contributions
			totalStrength := 0.0
			for _, f := range syn.FilamentStates {
				totalStrength += f
			}
			syn.FilamentStrength = totalStrength / float64(len(syn.FilamentStates))
		} else {
			syn.FilamentStrength = math.Min(syn.FilamentStrength+deltaG, 1.0)
		}

		syn.Conductance = m.filamentToConductance(syn.FilamentStrength)
		syn.ProgramCycles++

	} else if voltage < m.Config.ResetVoltage {
		// RESET operation - decrease conductance
		deltaG := m.Config.FilamentDecay * (m.Config.ResetVoltage - voltage) * pulseWidth

		if m.Config.MultifilamentMode {
			for k := range syn.FilamentStates {
				if m.RNG.Float64() < (1.0 - syn.FilamentStates[k]) {
					syn.FilamentStates[k] = math.Max(syn.FilamentStates[k]-deltaG, 0.0)
				}
			}
			totalStrength := 0.0
			for _, f := range syn.FilamentStates {
				totalStrength += f
			}
			syn.FilamentStrength = totalStrength / float64(len(syn.FilamentStates))
		} else {
			syn.FilamentStrength = math.Max(syn.FilamentStrength-deltaG, 0.0)
		}

		syn.Conductance = m.filamentToConductance(syn.FilamentStrength)
		syn.ProgramCycles++
	}

	syn.LastProgramTime = 0 // reset retention clock
}

func (m *MemristorArray) filamentToConductance(strength float64) float64 {
	// Logarithmic conductance mapping
	logGmin := math.Log(m.Config.ConductanceMin)
	logGmax := math.Log(m.Config.ConductanceMax)
	return math.Exp(logGmin + strength*(logGmax-logGmin))
}

// SimulateRetention models conductance drift over time.
func (m *MemristorArray) SimulateRetention(elapsedTime float64) {
	for i := range m.Synapses {
		for j := range m.Synapses[i] {
			syn := m.Synapses[i][j]

			// Drift toward mid-conductance
			driftFactor := 1.0 - math.Exp(-elapsedTime/syn.RetentionDecay)
			midConductance := (m.Config.ConductanceMin + m.Config.ConductanceMax) / 2

			syn.Conductance = syn.Conductance + driftFactor*(midConductance-syn.Conductance)
			syn.LastProgramTime += elapsedTime
		}
	}
}

// GetConductanceMatrix returns conductance values as 2D array.
func (m *MemristorArray) GetConductanceMatrix() [][]float64 {
	matrix := make([][]float64, m.Rows)
	for i := range matrix {
		matrix[i] = make([]float64, m.Cols)
		for j := range matrix[i] {
			matrix[i][j] = m.Synapses[i][j].Conductance
		}
	}
	return matrix
}

// =============================================================================
// Knowledge Distillation
// =============================================================================

// DistillationConfig configures knowledge distillation.
type DistillationConfig struct {
	Temperature     float64 // softmax temperature (3-20 typical)
	Alpha           float64 // weight for soft loss (0.5-0.9 typical)
	TeacherLayers   []int   // layers to distill from (empty = output only)
	IntermediateLoss bool   // include intermediate layer distillation
	AttentionTransfer bool  // distill attention maps
	FeatureMapLoss  bool    // match feature map statistics
}

// DefaultDistillationConfig returns standard distillation settings.
func DefaultDistillationConfig() DistillationConfig {
	return DistillationConfig{
		Temperature:      4.0,
		Alpha:            0.7,
		IntermediateLoss: true,
		AttentionTransfer: false,
		FeatureMapLoss:   false,
	}
}

// TeacherModel represents the larger model being distilled.
type TeacherModel struct {
	Layers           []DenseLayerWeights
	IntermediateOuts map[int][]float64 // cached intermediate outputs
}

// DenseLayerWeights holds weights for a dense layer.
type DenseLayerWeights struct {
	Weights [][]float64
	Biases  []float64
	Name    string
}

// StudentModel represents the smaller model learning from teacher.
type StudentModel struct {
	Layers           []DenseLayerWeights
	IntermediateOuts map[int][]float64
	Gradients        map[int][][]float64
}

// KnowledgeDistiller implements teacher-student training.
type KnowledgeDistiller struct {
	Config  DistillationConfig
	Teacher *TeacherModel
	Student *StudentModel
}

// NewKnowledgeDistiller creates a distillation trainer.
func NewKnowledgeDistiller(config DistillationConfig, teacher *TeacherModel) *KnowledgeDistiller {
	return &KnowledgeDistiller{
		Config:  config,
		Teacher: teacher,
	}
}

// SoftmaxWithTemperature applies temperature-scaled softmax.
func SoftmaxWithTemperature(logits []float64, temperature float64) []float64 {
	scaled := make([]float64, len(logits))
	maxVal := logits[0]
	for _, v := range logits[1:] {
		if v > maxVal {
			maxVal = v
		}
	}

	var sumExp float64
	for i, v := range logits {
		scaled[i] = math.Exp((v - maxVal) / temperature)
		sumExp += scaled[i]
	}

	for i := range scaled {
		scaled[i] /= sumExp
	}
	return scaled
}

// KLDivergence computes KL divergence between teacher and student distributions.
func KLDivergence(teacherProbs, studentLogits []float64, temperature float64) float64 {
	studentProbs := SoftmaxWithTemperature(studentLogits, temperature)

	var kl float64
	for i := range teacherProbs {
		if teacherProbs[i] > 1e-10 {
			kl += teacherProbs[i] * math.Log(teacherProbs[i]/studentProbs[i])
		}
	}
	return kl
}

// ComputeDistillationLoss computes combined hard and soft loss.
func (kd *KnowledgeDistiller) ComputeDistillationLoss(
	teacherLogits, studentLogits []float64,
	trueLabels []float64,
) float64 {
	T := kd.Config.Temperature
	alpha := kd.Config.Alpha

	// Soft loss: KL divergence with temperature
	teacherSoft := SoftmaxWithTemperature(teacherLogits, T)
	softLoss := KLDivergence(teacherSoft, studentLogits, T) * T * T

	// Hard loss: cross-entropy with true labels
	studentProbs := SoftmaxWithTemperature(studentLogits, 1.0)
	var hardLoss float64
	for i := range trueLabels {
		if trueLabels[i] > 0 {
			hardLoss -= trueLabels[i] * math.Log(studentProbs[i]+1e-10)
		}
	}

	return alpha*softLoss + (1-alpha)*hardLoss
}

// IntermediateLayerLoss computes MSE between teacher and student features.
func IntermediateLayerLoss(teacherFeatures, studentFeatures []float64) float64 {
	if len(teacherFeatures) != len(studentFeatures) {
		// Need projection layer - simplified here
		return 0.0
	}

	var mse float64
	for i := range teacherFeatures {
		diff := teacherFeatures[i] - studentFeatures[i]
		mse += diff * diff
	}
	return mse / float64(len(teacherFeatures))
}

// AttentionTransferLoss computes attention map similarity.
func AttentionTransferLoss(teacherAttn, studentAttn [][]float64) float64 {
	if len(teacherAttn) != len(studentAttn) {
		return 0.0
	}

	// L2 normalize attention maps
	normalizeAttn := func(attn [][]float64) [][]float64 {
		norm := make([][]float64, len(attn))
		for i := range attn {
			norm[i] = make([]float64, len(attn[i]))
			var sum float64
			for _, v := range attn[i] {
				sum += v * v
			}
			sqrtSum := math.Sqrt(sum + 1e-10)
			for j := range attn[i] {
				norm[i][j] = attn[i][j] / sqrtSum
			}
		}
		return norm
	}

	tNorm := normalizeAttn(teacherAttn)
	sNorm := normalizeAttn(studentAttn)

	var loss float64
	for i := range tNorm {
		for j := range tNorm[i] {
			diff := tNorm[i][j] - sNorm[i][j]
			loss += diff * diff
		}
	}
	return loss
}

// =============================================================================
// PQK Compression Pipeline
// =============================================================================

// PQKConfig configures the Pruning-Quantization-Knowledge distillation pipeline.
type PQKConfig struct {
	// Pruning settings
	PruningRatio     float64 // target sparsity (0.5-0.9)
	PruningMethod    string  // "magnitude", "movement", "lottery"
	StructuredPrune  bool    // structured vs unstructured

	// Quantization settings
	WeightBits       int     // weight bit-width (2-8)
	ActivationBits   int     // activation bit-width (4-8)
	QuantMethod      string  // "ptq", "qat", "mixed"
	PerChannelQuant  bool    // per-channel vs per-tensor

	// Distillation settings
	DistillationConfig DistillationConfig

	// Pipeline settings
	IterativePQK     bool    // iterative vs one-shot
	NumIterations    int     // iterations for iterative PQK
	FinetuneEpochs   int     // finetuning epochs per stage
}

// DefaultPQKConfig returns standard PQK settings.
func DefaultPQKConfig() PQKConfig {
	return PQKConfig{
		PruningRatio:     0.7,
		PruningMethod:    "magnitude",
		StructuredPrune:  false,
		WeightBits:       4,
		ActivationBits:   8,
		QuantMethod:      "qat",
		PerChannelQuant:  true,
		DistillationConfig: DefaultDistillationConfig(),
		IterativePQK:    true,
		NumIterations:   3,
		FinetuneEpochs:  10,
	}
}

// PQKPipeline implements the full compression pipeline.
type PQKPipeline struct {
	Config       PQKConfig
	OriginalSize int64 // original model size in bytes
	CurrentSize  int64 // current compressed size
	SparsityMap  map[string]float64
	QuantScales  map[string]float64
	PruneMasks   map[string][]bool
}

// NewPQKPipeline creates a PQK compression pipeline.
func NewPQKPipeline(config PQKConfig) *PQKPipeline {
	return &PQKPipeline{
		Config:      config,
		SparsityMap: make(map[string]float64),
		QuantScales: make(map[string]float64),
		PruneMasks:  make(map[string][]bool),
	}
}

// ComputePruneMask generates pruning mask using specified method.
func (p *PQKPipeline) ComputePruneMask(layerName string, weights []float64) []bool {
	mask := make([]bool, len(weights))

	switch p.Config.PruningMethod {
	case "magnitude":
		// Magnitude-based pruning
		absWeights := make([]float64, len(weights))
		for i, w := range weights {
			absWeights[i] = math.Abs(w)
		}

		// Find threshold for target sparsity
		sorted := make([]float64, len(absWeights))
		copy(sorted, absWeights)
		sort.Float64s(sorted)

		threshIdx := int(p.Config.PruningRatio * float64(len(sorted)))
		threshold := sorted[threshIdx]

		for i, aw := range absWeights {
			mask[i] = aw > threshold
		}

	case "movement":
		// Movement pruning (requires gradient history - simplified)
		for i := range weights {
			mask[i] = rand.Float64() > p.Config.PruningRatio
		}

	case "lottery":
		// Lottery ticket (random mask for finding winning tickets)
		for i := range weights {
			mask[i] = rand.Float64() > p.Config.PruningRatio
		}
	}

	p.PruneMasks[layerName] = mask

	// Compute actual sparsity
	zeros := 0
	for _, m := range mask {
		if !m {
			zeros++
		}
	}
	p.SparsityMap[layerName] = float64(zeros) / float64(len(mask))

	return mask
}

// ApplyPruning applies mask to weights.
func (p *PQKPipeline) ApplyPruning(layerName string, weights []float64) []float64 {
	mask, exists := p.PruneMasks[layerName]
	if !exists {
		mask = p.ComputePruneMask(layerName, weights)
	}

	pruned := make([]float64, len(weights))
	for i, w := range weights {
		if mask[i] {
			pruned[i] = w
		}
	}
	return pruned
}

// QuantizeWeights applies quantization to weights.
func (p *PQKPipeline) QuantizeWeights(layerName string, weights []float64) ([]int8, float64) {
	bits := p.Config.WeightBits
	levels := 1 << bits
	halfLevels := levels / 2

	// Find scale (symmetric quantization)
	maxAbs := 0.0
	for _, w := range weights {
		if math.Abs(w) > maxAbs {
			maxAbs = math.Abs(w)
		}
	}
	scale := maxAbs / float64(halfLevels-1)
	p.QuantScales[layerName] = scale

	// Quantize
	quantized := make([]int8, len(weights))
	for i, w := range weights {
		q := int(math.Round(w / scale))
		if q > halfLevels-1 {
			q = halfLevels - 1
		} else if q < -halfLevels {
			q = -halfLevels
		}
		quantized[i] = int8(q)
	}

	return quantized, scale
}

// DequantizeWeights converts quantized weights back to float.
func (p *PQKPipeline) DequantizeWeights(quantized []int8, scale float64) []float64 {
	weights := make([]float64, len(quantized))
	for i, q := range quantized {
		weights[i] = float64(q) * scale
	}
	return weights
}

// RunPipeline executes full PQK compression.
func (p *PQKPipeline) RunPipeline(layerName string, weights []float64, teacher *TeacherModel) PQKResult {
	result := PQKResult{
		LayerName:    layerName,
		OriginalSize: int64(len(weights) * 4), // 4 bytes per float32
	}

	// Stage 1: Pruning
	pruned := p.ApplyPruning(layerName, weights)
	result.Sparsity = p.SparsityMap[layerName]

	// Stage 2: Quantization
	quantized, scale := p.QuantizeWeights(layerName, pruned)
	result.QuantScale = scale
	result.QuantBits = p.Config.WeightBits

	// Compute compressed size
	nonZero := 0
	for _, q := range quantized {
		if q != 0 {
			nonZero++
		}
	}
	bitsPerWeight := p.Config.WeightBits
	result.CompressedSize = int64((nonZero*bitsPerWeight + 7) / 8)

	// Store quantized weights
	result.QuantizedWeights = quantized

	// Compute compression ratio
	result.CompressionRatio = float64(result.OriginalSize) / float64(result.CompressedSize)

	return result
}

// PQKResult holds compression results.
type PQKResult struct {
	LayerName        string
	OriginalSize     int64
	CompressedSize   int64
	CompressionRatio float64
	Sparsity         float64
	QuantScale       float64
	QuantBits        int
	QuantizedWeights []int8
}

// =============================================================================
// CIM-Optimized Continual Learning
// =============================================================================

// CIMContinualConfig configures CIM-aware continual learning.
type CIMContinualConfig struct {
	CrossbarSize     int     // crossbar dimensions
	MaxConductance   float64 // Gmax
	MinConductance   float64 // Gmin
	WritePrecision   int     // bits for write precision
	ReadNoise        float64 // read noise percentage
	EnduranceLimit   int     // max program cycles
	UseReplayBuffer  bool    // experience replay
	ReplayBufferSize int     // samples to store
}

// DefaultCIMContinualConfig returns standard CIM continual settings.
func DefaultCIMContinualConfig() CIMContinualConfig {
	return CIMContinualConfig{
		CrossbarSize:     64,
		MaxConductance:   1e-4,
		MinConductance:   1e-7,
		WritePrecision:   6,
		ReadNoise:        0.02,
		EnduranceLimit:   1e6,
		UseReplayBuffer:  true,
		ReplayBufferSize: 1000,
	}
}

// CIMContinualLearner implements continual learning on CIM hardware.
type CIMContinualLearner struct {
	Config        CIMContinualConfig
	EWC           *EWCRegularizer
	Metaplastic   *MetaplasticNetwork
	ReplayBuffer  []ReplaySample
	Crossbar      *MemristorArray
	TaskHistory   []TaskInfo
	CurrentTask   int
}

// ReplaySample stores experience for replay.
type ReplaySample struct {
	Input    []float64
	Target   []float64
	TaskID   int
	Priority float64 // for prioritized replay
}

// TaskInfo records task metadata.
type TaskInfo struct {
	TaskID       int
	NumClasses   int
	SampleCount  int
	Accuracy     float64
	FisherDiag   []float64
}

// NewCIMContinualLearner creates a CIM-aware continual learner.
func NewCIMContinualLearner(config CIMContinualConfig, seed int64) *CIMContinualLearner {
	ewcConfig := DefaultEWCConfig()
	metaConfig := DefaultMetaplasticityConfig()
	memConfig := DefaultMemristorMetaplasticConfig()
	memConfig.ConductanceMin = config.MinConductance
	memConfig.ConductanceMax = config.MaxConductance

	return &CIMContinualLearner{
		Config:       config,
		EWC:          NewEWCRegularizer(ewcConfig),
		Metaplastic:  NewMetaplasticNetwork(metaConfig, seed),
		ReplayBuffer: make([]ReplaySample, 0, config.ReplayBufferSize),
		Crossbar:     NewMemristorArray(config.CrossbarSize, config.CrossbarSize, memConfig, seed),
	}
}

// AddToReplayBuffer adds sample to buffer with reservoir sampling.
func (c *CIMContinualLearner) AddToReplayBuffer(input, target []float64) {
	sample := ReplaySample{
		Input:    make([]float64, len(input)),
		Target:   make([]float64, len(target)),
		TaskID:   c.CurrentTask,
		Priority: 1.0,
	}
	copy(sample.Input, input)
	copy(sample.Target, target)

	if len(c.ReplayBuffer) < c.Config.ReplayBufferSize {
		c.ReplayBuffer = append(c.ReplayBuffer, sample)
	} else {
		// Reservoir sampling
		idx := rand.Intn(len(c.ReplayBuffer))
		c.ReplayBuffer[idx] = sample
	}
}

// SampleFromReplay retrieves samples for replay.
func (c *CIMContinualLearner) SampleFromReplay(batchSize int) []ReplaySample {
	if len(c.ReplayBuffer) == 0 {
		return nil
	}

	batch := make([]ReplaySample, 0, batchSize)
	for i := 0; i < batchSize && i < len(c.ReplayBuffer); i++ {
		idx := rand.Intn(len(c.ReplayBuffer))
		batch = append(batch, c.ReplayBuffer[idx])
	}
	return batch
}

// StartTask initializes learning for a new task.
func (c *CIMContinualLearner) StartTask(numClasses int) {
	c.CurrentTask++
	c.Metaplastic.StartNewTask()
	c.TaskHistory = append(c.TaskHistory, TaskInfo{
		TaskID:     c.CurrentTask,
		NumClasses: numClasses,
	})
}

// EndTask finalizes task learning and consolidates.
func (c *CIMContinualLearner) EndTask(layerName string, weights []float64, accuracy float64) {
	// Estimate Fisher information
	dummyGradFunc := func(w []float64) []float64 {
		grads := make([]float64, len(w))
		for i := range grads {
			grads[i] = rand.NormFloat64() * 0.1
		}
		return grads
	}

	fisher := c.EWC.EstimateFisher(layerName, weights, dummyGradFunc)
	c.EWC.RegisterTask(layerName, weights, fisher)

	// Update task history
	if len(c.TaskHistory) > 0 {
		c.TaskHistory[len(c.TaskHistory)-1].Accuracy = accuracy
		c.TaskHistory[len(c.TaskHistory)-1].FisherDiag = fisher.Diagonal
	}
}

// ComputeContinualLoss combines all continual learning losses.
func (c *CIMContinualLearner) ComputeContinualLoss(
	layerName string,
	currentWeights []float64,
	taskLoss float64,
) float64 {
	// EWC penalty
	ewcPenalty := c.EWC.ComputePenalty(layerName, currentWeights)

	// Consolidation regularization from metaplasticity
	consolidation := c.Metaplastic.GetConsolidationMap(layerName)
	var consolidationPenalty float64
	if consolidation != nil && len(consolidation) == len(currentWeights) {
		starW := c.EWC.StarWeights[layerName]
		if starW != nil {
			for i := range currentWeights {
				diff := currentWeights[i] - starW[i]
				consolidationPenalty += consolidation[i] * diff * diff
			}
		}
	}

	return taskLoss + ewcPenalty + 0.1*consolidationPenalty
}

// =============================================================================
// Progressive Distillation for CIM
// =============================================================================

// ProgressiveDistillConfig configures progressive network distillation.
type ProgressiveDistillConfig struct {
	Stages          int       // number of compression stages
	SizeReduction   []float64 // size reduction per stage
	DistillConfig   DistillationConfig
	QuantBitsStages []int     // quantization bits per stage
	FinetuneEpochs  []int     // finetuning epochs per stage
}

// DefaultProgressiveDistillConfig returns standard progressive distillation.
func DefaultProgressiveDistillConfig() ProgressiveDistillConfig {
	return ProgressiveDistillConfig{
		Stages:          3,
		SizeReduction:   []float64{0.7, 0.5, 0.3}, // 70%, 50%, 30% of original
		DistillConfig:   DefaultDistillationConfig(),
		QuantBitsStages: []int{8, 6, 4},
		FinetuneEpochs:  []int{10, 15, 20},
	}
}

// ProgressiveDistiller implements staged model compression.
type ProgressiveDistiller struct {
	Config       ProgressiveDistillConfig
	StageModels  []*StudentModel
	CurrentStage int
	Metrics      []StageMetrics
}

// StageMetrics tracks compression metrics per stage.
type StageMetrics struct {
	Stage           int
	OriginalParams  int
	CompressedParams int
	Accuracy        float64
	Size            int64
	Latency         float64 // inference latency
}

// NewProgressiveDistiller creates a progressive distillation pipeline.
func NewProgressiveDistiller(config ProgressiveDistillConfig) *ProgressiveDistiller {
	return &ProgressiveDistiller{
		Config:      config,
		StageModels: make([]*StudentModel, config.Stages),
		Metrics:     make([]StageMetrics, config.Stages),
	}
}

// DistillStage performs one stage of progressive distillation.
func (p *ProgressiveDistiller) DistillStage(
	teacher *TeacherModel,
	targetSize float64,
	quantBits int,
) *StudentModel {
	student := &StudentModel{
		Layers:           make([]DenseLayerWeights, len(teacher.Layers)),
		IntermediateOuts: make(map[int][]float64),
		Gradients:        make(map[int][][]float64),
	}

	// Create compressed student architecture
	for i, tLayer := range teacher.Layers {
		// Reduce dimensions based on target size
		reducedRows := int(float64(len(tLayer.Weights)) * targetSize)
		reducedCols := int(float64(len(tLayer.Weights[0])) * targetSize)

		student.Layers[i] = DenseLayerWeights{
			Weights: make([][]float64, reducedRows),
			Biases:  make([]float64, reducedCols),
			Name:    tLayer.Name + "_compressed",
		}

		for j := range student.Layers[i].Weights {
			student.Layers[i].Weights[j] = make([]float64, reducedCols)
			// Initialize from teacher (simplified - real impl uses distillation)
			for k := range student.Layers[i].Weights[j] {
				if j < len(tLayer.Weights) && k < len(tLayer.Weights[j]) {
					student.Layers[i].Weights[j][k] = tLayer.Weights[j][k]
				}
			}
		}
	}

	p.StageModels[p.CurrentStage] = student
	p.CurrentStage++

	return student
}

// GetCompressionSummary returns overall compression statistics.
func (p *ProgressiveDistiller) GetCompressionSummary() map[string]interface{} {
	summary := make(map[string]interface{})
	summary["stages_completed"] = p.CurrentStage
	summary["total_stages"] = p.Config.Stages

	if len(p.Metrics) > 0 {
		initial := p.Metrics[0]
		final := p.Metrics[len(p.Metrics)-1]
		if initial.OriginalParams > 0 {
			summary["total_compression"] = float64(initial.OriginalParams) / float64(final.CompressedParams)
		}
		summary["accuracy_drop"] = initial.Accuracy - final.Accuracy
	}

	return summary
}

// =============================================================================
// Serialization
// =============================================================================

// ContinualLearningState captures full state for persistence.
type ContinualLearningState struct {
	EWCConfig        EWCConfig                     `json:"ewc_config"`
	StarWeights      map[string][]float64          `json:"star_weights"`
	FisherDiagonals  map[string][]float64          `json:"fisher_diagonals"`
	ConsolidatedF    map[string][]float64          `json:"consolidated_fisher"`
	TaskCount        int                           `json:"task_count"`
	MetaplasticState map[string][]SynapticStateJSON `json:"metaplastic_state"`
	ReplayBuffer     []ReplaySample                `json:"replay_buffer"`
}

// SynapticStateJSON is JSON-serializable synapse state.
type SynapticStateJSON struct {
	Weight        float64 `json:"weight"`
	Consolidation float64 `json:"consolidation"`
	CascadeLevel  int     `json:"cascade_level"`
}

// ExportState exports continual learning state to JSON.
func (c *CIMContinualLearner) ExportState() ([]byte, error) {
	state := ContinualLearningState{
		EWCConfig:       c.EWC.Config,
		StarWeights:     c.EWC.StarWeights,
		FisherDiagonals: make(map[string][]float64),
		ConsolidatedF:   c.EWC.ConsolidatedF,
		TaskCount:       c.EWC.TaskCount,
		MetaplasticState: make(map[string][]SynapticStateJSON),
		ReplayBuffer:    c.ReplayBuffer,
	}

	// Export Fisher matrices
	for name, fisher := range c.EWC.FisherMatrices {
		state.FisherDiagonals[name] = fisher.Diagonal
	}

	// Export metaplastic synapses
	for name, synapses := range c.Metaplastic.Synapses {
		synStates := make([]SynapticStateJSON, len(synapses))
		for i, s := range synapses {
			synStates[i] = SynapticStateJSON{
				Weight:        s.Weight,
				Consolidation: s.Consolidation,
				CascadeLevel:  s.CascadeLevel,
			}
		}
		state.MetaplasticState[name] = synStates
	}

	return json.MarshalIndent(state, "", "  ")
}

// ImportState loads continual learning state from JSON.
func (c *CIMContinualLearner) ImportState(data []byte) error {
	var state ContinualLearningState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state: %w", err)
	}

	c.EWC.Config = state.EWCConfig
	c.EWC.StarWeights = state.StarWeights
	c.EWC.ConsolidatedF = state.ConsolidatedF
	c.EWC.TaskCount = state.TaskCount

	// Restore Fisher matrices
	for name, diag := range state.FisherDiagonals {
		c.EWC.FisherMatrices[name] = &FisherInformation{
			Diagonal:  diag,
			LayerName: name,
		}
	}

	// Restore metaplastic synapses
	for name, synStates := range state.MetaplasticState {
		synapses := make([]*SynapticState, len(synStates))
		for i, ss := range synStates {
			synapses[i] = &SynapticState{
				Weight:        ss.Weight,
				Consolidation: ss.Consolidation,
				CascadeLevel:  ss.CascadeLevel,
			}
		}
		c.Metaplastic.Synapses[name] = synapses
	}

	c.ReplayBuffer = state.ReplayBuffer

	return nil
}

// =============================================================================
// Benchmarking
// =============================================================================

// ContinualBenchmark evaluates continual learning performance.
type ContinualBenchmark struct {
	Tasks            []BenchmarkTask
	ForwardTransfer  float64 // positive transfer to new tasks
	BackwardTransfer float64 // forgetting of old tasks
	AverageAccuracy  float64
	FinalAccuracy    []float64
}

// BenchmarkTask represents a task in the benchmark.
type BenchmarkTask struct {
	Name       string
	NumClasses int
	Accuracy   float64
}

// ComputeMetrics calculates continual learning metrics.
func (b *ContinualBenchmark) ComputeMetrics(accuracyMatrix [][]float64) {
	numTasks := len(accuracyMatrix)
	if numTasks == 0 {
		return
	}

	// Average accuracy
	var totalAcc float64
	count := 0
	for i := range accuracyMatrix {
		for j := 0; j <= i && j < len(accuracyMatrix[i]); j++ {
			totalAcc += accuracyMatrix[i][j]
			count++
		}
	}
	if count > 0 {
		b.AverageAccuracy = totalAcc / float64(count)
	}

	// Backward transfer (forgetting)
	var bt float64
	btCount := 0
	for i := 1; i < numTasks; i++ {
		for j := 0; j < i && j < len(accuracyMatrix[i]); j++ {
			// Compare accuracy on task j after learning task i vs right after task j
			if j < len(accuracyMatrix[j]) {
				bt += accuracyMatrix[i][j] - accuracyMatrix[j][j]
				btCount++
			}
		}
	}
	if btCount > 0 {
		b.BackwardTransfer = bt / float64(btCount)
	}

	// Forward transfer
	var ft float64
	ftCount := 0
	for i := 1; i < numTasks; i++ {
		if i < len(accuracyMatrix[i-1]) {
			// Compare zero-shot accuracy before training on task i
			ft += accuracyMatrix[i-1][i] // This would need baseline
			ftCount++
		}
	}
	if ftCount > 0 {
		b.ForwardTransfer = ft / float64(ftCount)
	}

	// Final accuracy on all tasks
	b.FinalAccuracy = make([]float64, numTasks)
	lastRow := accuracyMatrix[numTasks-1]
	copy(b.FinalAccuracy, lastRow)
}

// min returns minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

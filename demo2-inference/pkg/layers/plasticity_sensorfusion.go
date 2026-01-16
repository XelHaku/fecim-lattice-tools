// Package layers provides ferroelectric synaptic plasticity rules and
// neuromorphic sensor fusion for IronLattice CIM architectures.
//
// This module simulates:
// - Spike-Timing Dependent Plasticity (STDP) with FeFET synapses
// - Pair-Pulse Facilitation (PPF) and depression
// - Long-Term Potentiation/Depression (LTP/LTD)
// - Multimodal in-sensor computing (visual, tactile, auditory)
// - Sensor fusion for edge AI applications
//
// Based on research from Advanced Science 2024, Nature npj 2025,
// and ferroelectric neuromorphic device studies.
package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// ============================================================================
// SYNAPTIC PLASTICITY CONFIGURATION
// ============================================================================

// PlasticityConfig defines parameters for synaptic plasticity rules
type PlasticityConfig struct {
	// STDP parameters
	TauPlus       float64 // LTP time constant (ms)
	TauMinus      float64 // LTD time constant (ms)
	APlus         float64 // LTP amplitude
	AMinus        float64 // LTD amplitude
	WMax          float64 // Maximum synaptic weight
	WMin          float64 // Minimum synaptic weight

	// PPF parameters
	PPFTau1       float64 // Fast facilitation time constant (ms)
	PPFTau2       float64 // Slow facilitation time constant (ms)
	PPFAmplitude  float64 // Facilitation amplitude

	// FeFET device parameters
	PulseWidthUs  float64 // Programming pulse width (µs)
	PulseVoltageV float64 // Programming pulse voltage
	EnergyPerSpike float64 // Energy per spike (aJ)
	SwitchingTime float64 // Switching time (µs)

	// Learning rate
	LearningRate  float64 // Global learning rate modifier
}

// DefaultPlasticityConfig returns biologically-inspired plasticity parameters
func DefaultPlasticityConfig() *PlasticityConfig {
	return &PlasticityConfig{
		TauPlus:        20.0,  // ms
		TauMinus:       20.0,  // ms
		APlus:          0.01,
		AMinus:         0.012, // Slightly stronger LTD
		WMax:           1.0,
		WMin:           0.0,
		PPFTau1:        10.0,  // ms
		PPFTau2:        100.0, // ms
		PPFAmplitude:   0.5,
		PulseWidthUs:   1.0,
		PulseVoltageV:  3.0,
		EnergyPerSpike: 48.0, // aJ (from 2D SnS2 FeFET)
		SwitchingTime:  1.0,  // µs
		LearningRate:   1.0,
	}
}

// SensorFusionConfig defines parameters for multimodal sensor fusion
type SensorFusionConfig struct {
	// Sensor modalities
	NumVisualChannels   int
	NumTactileChannels  int
	NumAuditoryChannels int
	NumProprioceptive   int

	// In-sensor computing
	EnableInSensor      bool
	ADCBits             int
	SamplingRateHz      float64

	// Fusion parameters
	FusionMethod        string  // "early", "late", "hierarchical"
	AttentionHeads      int
	TemporalWindowMs    float64

	// Energy constraints
	PowerBudgetUW       float64 // Power budget in µW
	LatencyTargetMs     float64 // Latency target in ms
}

// DefaultSensorFusionConfig returns optimized sensor fusion configuration
func DefaultSensorFusionConfig() *SensorFusionConfig {
	return &SensorFusionConfig{
		NumVisualChannels:   64,
		NumTactileChannels:  16,
		NumAuditoryChannels: 32,
		NumProprioceptive:   8,
		EnableInSensor:      true,
		ADCBits:             8,
		SamplingRateHz:      1000,
		FusionMethod:        "hierarchical",
		AttentionHeads:      4,
		TemporalWindowMs:    100,
		PowerBudgetUW:       20, // Ultra-low power target
		LatencyTargetMs:     10,
	}
}

// ============================================================================
// FERROELECTRIC SYNAPSE WITH PLASTICITY
// ============================================================================

// FeFETSynapse represents a ferroelectric FET-based synapse with plasticity
type FeFETSynapse struct {
	Config *PlasticityConfig

	// Synaptic state
	Weight          float64 // Current synaptic weight [WMin, WMax]
	Polarization    float64 // Ferroelectric polarization state

	// Timing state
	LastPreSpike    float64 // Time of last presynaptic spike (ms)
	LastPostSpike   float64 // Time of last postsynaptic spike (ms)

	// Facilitation state
	PPFTrace1       float64 // Fast facilitation trace
	PPFTrace2       float64 // Slow facilitation trace

	// Statistics
	LTPCount        int
	LTDCount        int
	TotalSpikes     int
	TotalEnergy     float64 // aJ
}

// NewFeFETSynapse creates a new ferroelectric synapse
func NewFeFETSynapse(config *PlasticityConfig) *FeFETSynapse {
	return &FeFETSynapse{
		Config:       config,
		Weight:       0.5, // Initialize at midpoint
		Polarization: 0.5,
		LastPreSpike: -1000, // Long time ago
		LastPostSpike: -1000,
	}
}

// ProcessPreSpike handles a presynaptic spike
func (s *FeFETSynapse) ProcessPreSpike(currentTime float64) float64 {
	s.TotalSpikes++
	s.TotalEnergy += s.Config.EnergyPerSpike

	// Update PPF traces
	dt := currentTime - s.LastPreSpike
	if dt > 0 {
		s.PPFTrace1 *= math.Exp(-dt / s.Config.PPFTau1)
		s.PPFTrace2 *= math.Exp(-dt / s.Config.PPFTau2)
	}
	s.PPFTrace1 += s.Config.PPFAmplitude
	s.PPFTrace2 += s.Config.PPFAmplitude * 0.5

	// Calculate facilitated weight
	facilitatedWeight := s.Weight * (1 + s.PPFTrace1 + s.PPFTrace2)
	if facilitatedWeight > s.Config.WMax {
		facilitatedWeight = s.Config.WMax
	}

	// STDP: Check for LTD (pre after post)
	dtPost := currentTime - s.LastPostSpike
	if dtPost > 0 && dtPost < 5*s.Config.TauMinus {
		// Post-Pre: LTD
		deltaW := -s.Config.AMinus * math.Exp(-dtPost/s.Config.TauMinus) * s.Config.LearningRate
		s.applyWeightChange(deltaW)
		s.LTDCount++
	}

	s.LastPreSpike = currentTime
	return facilitatedWeight
}

// ProcessPostSpike handles a postsynaptic spike
func (s *FeFETSynapse) ProcessPostSpike(currentTime float64) {
	// STDP: Check for LTP (post after pre)
	dtPre := currentTime - s.LastPreSpike
	if dtPre > 0 && dtPre < 5*s.Config.TauPlus {
		// Pre-Post: LTP
		deltaW := s.Config.APlus * math.Exp(-dtPre/s.Config.TauPlus) * s.Config.LearningRate
		s.applyWeightChange(deltaW)
		s.LTPCount++
	}

	s.LastPostSpike = currentTime
}

// applyWeightChange applies weight change with bounds
func (s *FeFETSynapse) applyWeightChange(deltaW float64) {
	// Soft bounds (weight-dependent)
	if deltaW > 0 {
		deltaW *= (s.Config.WMax - s.Weight) / s.Config.WMax
	} else {
		deltaW *= (s.Weight - s.Config.WMin) / s.Config.WMax
	}

	s.Weight += deltaW

	// Hard bounds
	if s.Weight > s.Config.WMax {
		s.Weight = s.Config.WMax
	}
	if s.Weight < s.Config.WMin {
		s.Weight = s.Config.WMin
	}

	// Update polarization (maps weight to FE state)
	s.Polarization = s.Weight / s.Config.WMax
}

// GetEPSC returns excitatory post-synaptic current
func (s *FeFETSynapse) GetEPSC(inputVoltage float64) float64 {
	// Current proportional to weight and facilitation
	facilitation := 1 + s.PPFTrace1 + s.PPFTrace2
	return s.Weight * facilitation * inputVoltage
}

// DecayTraces decays facilitation traces over time
func (s *FeFETSynapse) DecayTraces(dt float64) {
	s.PPFTrace1 *= math.Exp(-dt / s.Config.PPFTau1)
	s.PPFTrace2 *= math.Exp(-dt / s.Config.PPFTau2)
}

// ============================================================================
// STDP LEARNING RULE
// ============================================================================

// STDPRule implements spike-timing dependent plasticity
type STDPRule struct {
	Config *PlasticityConfig

	// STDP window function parameters
	UseTriplet     bool    // Use triplet STDP rule
	TripletTauX    float64 // Pre-trace time constant
	TripletTauY    float64 // Post-trace time constant

	// Traces for triplet rule
	PreTrace       float64
	PostTrace      float64
	SlowPostTrace  float64
}

// NewSTDPRule creates a new STDP learning rule
func NewSTDPRule(config *PlasticityConfig, useTriplet bool) *STDPRule {
	return &STDPRule{
		Config:      config,
		UseTriplet:  useTriplet,
		TripletTauX: 100.0, // ms
		TripletTauY: 125.0, // ms
	}
}

// ComputeWeightChange computes weight change based on spike timing
func (r *STDPRule) ComputeWeightChange(dtPrePost float64) float64 {
	// dtPrePost = t_post - t_pre
	// Positive: pre before post (LTP)
	// Negative: post before pre (LTD)

	if dtPrePost > 0 {
		// Pre-Post: LTP
		return r.Config.APlus * math.Exp(-dtPrePost/r.Config.TauPlus)
	} else {
		// Post-Pre: LTD
		return -r.Config.AMinus * math.Exp(dtPrePost/r.Config.TauMinus)
	}
}

// UpdateTraces updates eligibility traces
func (r *STDPRule) UpdateTraces(dt float64, preSpike, postSpike bool) {
	// Decay traces
	r.PreTrace *= math.Exp(-dt / r.Config.TauPlus)
	r.PostTrace *= math.Exp(-dt / r.Config.TauMinus)
	if r.UseTriplet {
		r.SlowPostTrace *= math.Exp(-dt / r.TripletTauY)
	}

	// Update on spikes
	if preSpike {
		r.PreTrace += 1.0
	}
	if postSpike {
		r.PostTrace += 1.0
		if r.UseTriplet {
			r.SlowPostTrace += 1.0
		}
	}
}

// GetLTPComponent returns the LTP component based on traces
func (r *STDPRule) GetLTPComponent() float64 {
	if r.UseTriplet {
		return r.Config.APlus * r.PreTrace * (1 + r.SlowPostTrace)
	}
	return r.Config.APlus * r.PreTrace
}

// GetLTDComponent returns the LTD component based on traces
func (r *STDPRule) GetLTDComponent() float64 {
	return r.Config.AMinus * r.PostTrace
}

// ============================================================================
// SPIKING NEURAL NETWORK LAYER WITH STDP
// ============================================================================

// STDPLayer represents a layer of spiking neurons with STDP learning
type STDPLayer struct {
	Config    *PlasticityConfig
	Synapses  [][]*FeFETSynapse
	Neurons   []*LIFNeuron
	STDPRule  *STDPRule

	// Layer dimensions
	InputSize  int
	OutputSize int

	// Statistics
	TotalLTP   int
	TotalLTD   int
	AvgWeight  float64
}

// LIFNeuron represents a Leaky Integrate-and-Fire neuron
type LIFNeuron struct {
	Membrane     float64 // Membrane potential
	Threshold    float64 // Firing threshold
	RestPotential float64 // Resting potential
	TauMembrane  float64 // Membrane time constant (ms)
	RefractoryMs float64 // Refractory period (ms)

	LastSpike    float64 // Time of last spike
	SpikeCount   int
}

// NewLIFNeuron creates a new LIF neuron
func NewLIFNeuron() *LIFNeuron {
	return &LIFNeuron{
		Membrane:      0,
		Threshold:     1.0,
		RestPotential: 0,
		TauMembrane:   20.0,
		RefractoryMs:  2.0,
		LastSpike:     -1000,
	}
}

// Integrate integrates input current and returns true if spike
func (n *LIFNeuron) Integrate(current float64, dt float64, currentTime float64) bool {
	// Check refractory period
	if currentTime-n.LastSpike < n.RefractoryMs {
		return false
	}

	// Leaky integration
	n.Membrane *= math.Exp(-dt / n.TauMembrane)
	n.Membrane += current

	// Check for spike
	if n.Membrane >= n.Threshold {
		n.Membrane = n.RestPotential
		n.LastSpike = currentTime
		n.SpikeCount++
		return true
	}

	return false
}

// NewSTDPLayer creates a layer with STDP-enabled synapses
func NewSTDPLayer(config *PlasticityConfig, inputSize, outputSize int) *STDPLayer {
	layer := &STDPLayer{
		Config:     config,
		InputSize:  inputSize,
		OutputSize: outputSize,
		Synapses:   make([][]*FeFETSynapse, outputSize),
		Neurons:    make([]*LIFNeuron, outputSize),
		STDPRule:   NewSTDPRule(config, false),
	}

	for i := 0; i < outputSize; i++ {
		layer.Neurons[i] = NewLIFNeuron()
		layer.Synapses[i] = make([]*FeFETSynapse, inputSize)
		for j := 0; j < inputSize; j++ {
			layer.Synapses[i][j] = NewFeFETSynapse(config)
			// Random initial weights
			layer.Synapses[i][j].Weight = 0.3 + 0.4*rand.Float64()
		}
	}

	return layer
}

// Forward processes input spikes through the layer
func (l *STDPLayer) Forward(inputSpikes []bool, currentTime float64, dt float64) []bool {
	outputSpikes := make([]bool, l.OutputSize)

	for i := 0; i < l.OutputSize; i++ {
		totalCurrent := 0.0

		// Accumulate input from all synapses
		for j := 0; j < l.InputSize; j++ {
			if inputSpikes[j] {
				// Presynaptic spike
				current := l.Synapses[i][j].ProcessPreSpike(currentTime)
				totalCurrent += current
			} else {
				// Decay traces
				l.Synapses[i][j].DecayTraces(dt)
			}
		}

		// Integrate in neuron
		if l.Neurons[i].Integrate(totalCurrent, dt, currentTime) {
			outputSpikes[i] = true

			// Postsynaptic spike: update all synapses to this neuron
			for j := 0; j < l.InputSize; j++ {
				l.Synapses[i][j].ProcessPostSpike(currentTime)
			}
		}
	}

	l.updateStatistics()
	return outputSpikes
}

// updateStatistics updates layer statistics
func (l *STDPLayer) updateStatistics() {
	totalLTP := 0
	totalLTD := 0
	totalWeight := 0.0

	for i := 0; i < l.OutputSize; i++ {
		for j := 0; j < l.InputSize; j++ {
			totalLTP += l.Synapses[i][j].LTPCount
			totalLTD += l.Synapses[i][j].LTDCount
			totalWeight += l.Synapses[i][j].Weight
		}
	}

	l.TotalLTP = totalLTP
	l.TotalLTD = totalLTD
	l.AvgWeight = totalWeight / float64(l.OutputSize*l.InputSize)
}

// GetWeightMatrix returns the current weight matrix
func (l *STDPLayer) GetWeightMatrix() [][]float64 {
	weights := make([][]float64, l.OutputSize)
	for i := 0; i < l.OutputSize; i++ {
		weights[i] = make([]float64, l.InputSize)
		for j := 0; j < l.InputSize; j++ {
			weights[i][j] = l.Synapses[i][j].Weight
		}
	}
	return weights
}

// ============================================================================
// MULTIMODAL SENSOR INTERFACE
// ============================================================================

// SensorModality represents a type of sensor input
type SensorModality int

const (
	ModalityVisual SensorModality = iota
	ModalityTactile
	ModalityAuditory
	ModalityProprioceptive
)

// SensorInput represents input from a single sensor modality
type SensorInput struct {
	Modality    SensorModality
	Timestamp   float64   // ms
	RawValues   []float64 // Raw sensor values
	SpikeTrains [][]bool  // Spike-encoded values (time x channels)
}

// InSensorProcessor performs in-sensor preprocessing
type InSensorProcessor struct {
	Config    *SensorFusionConfig
	Modality  SensorModality

	// Encoding parameters
	ThresholdPositive float64
	ThresholdNegative float64

	// State for temporal difference encoding
	PreviousValues []float64

	// Energy tracking
	ProcessingEnergy float64 // pJ
}

// NewInSensorProcessor creates a processor for a specific modality
func NewInSensorProcessor(config *SensorFusionConfig, modality SensorModality) *InSensorProcessor {
	numChannels := 0
	switch modality {
	case ModalityVisual:
		numChannels = config.NumVisualChannels
	case ModalityTactile:
		numChannels = config.NumTactileChannels
	case ModalityAuditory:
		numChannels = config.NumAuditoryChannels
	case ModalityProprioceptive:
		numChannels = config.NumProprioceptive
	}

	return &InSensorProcessor{
		Config:            config,
		Modality:          modality,
		ThresholdPositive: 0.1,
		ThresholdNegative: -0.1,
		PreviousValues:    make([]float64, numChannels),
	}
}

// ProcessRaw converts raw sensor data to spike trains
func (p *InSensorProcessor) ProcessRaw(rawValues []float64, numTimeSteps int) [][]bool {
	spikeTrains := make([][]bool, numTimeSteps)
	for t := 0; t < numTimeSteps; t++ {
		spikeTrains[t] = make([]bool, len(rawValues))
	}

	// Temporal difference encoding (event-driven)
	for i, val := range rawValues {
		diff := val - p.PreviousValues[i]

		// Generate spikes based on change magnitude
		if diff > p.ThresholdPositive {
			// Positive change: ON spikes
			numSpikes := int(diff / p.ThresholdPositive)
			for t := 0; t < numSpikes && t < numTimeSteps; t++ {
				spikeTrains[t][i] = true
			}
		} else if diff < p.ThresholdNegative {
			// Negative change: encoded as later spikes (for simplicity)
			numSpikes := int(-diff / (-p.ThresholdNegative))
			for t := numTimeSteps / 2; t < numTimeSteps && t < numTimeSteps/2+numSpikes; t++ {
				spikeTrains[t][i] = true
			}
		}

		p.PreviousValues[i] = val
	}

	// Energy: ~1 fJ per spike generated
	totalSpikes := 0
	for t := range spikeTrains {
		for _, spike := range spikeTrains[t] {
			if spike {
				totalSpikes++
			}
		}
	}
	p.ProcessingEnergy += float64(totalSpikes) * 0.001 // pJ

	return spikeTrains
}

// ============================================================================
// MULTIMODAL FUSION ENGINE
// ============================================================================

// MultimodalFusionEngine fuses multiple sensor modalities
type MultimodalFusionEngine struct {
	Config     *SensorFusionConfig
	Processors map[SensorModality]*InSensorProcessor

	// Fusion layers
	EarlyFusion    *STDPLayer
	ModalityLayers map[SensorModality]*STDPLayer
	LateFusion     *STDPLayer

	// Attention mechanism
	AttentionWeights map[SensorModality]float64

	// Temporal buffer
	TemporalBuffer []map[SensorModality]*SensorInput

	// Statistics
	TotalInferences int
	TotalEnergy     float64 // pJ
	AverageLatency  float64 // ms
}

// NewMultimodalFusionEngine creates a new fusion engine
func NewMultimodalFusionEngine(sensorConfig *SensorFusionConfig, plasticityConfig *PlasticityConfig) *MultimodalFusionEngine {
	engine := &MultimodalFusionEngine{
		Config:           sensorConfig,
		Processors:       make(map[SensorModality]*InSensorProcessor),
		ModalityLayers:   make(map[SensorModality]*STDPLayer),
		AttentionWeights: make(map[SensorModality]float64),
		TemporalBuffer:   make([]map[SensorModality]*SensorInput, 0),
	}

	// Create processors for each modality
	modalities := []SensorModality{ModalityVisual, ModalityTactile, ModalityAuditory, ModalityProprioceptive}
	channelCounts := []int{
		sensorConfig.NumVisualChannels,
		sensorConfig.NumTactileChannels,
		sensorConfig.NumAuditoryChannels,
		sensorConfig.NumProprioceptive,
	}

	totalChannels := 0
	for i, modality := range modalities {
		engine.Processors[modality] = NewInSensorProcessor(sensorConfig, modality)

		// Modality-specific processing layer
		hiddenSize := channelCounts[i] / 2
		if hiddenSize < 4 {
			hiddenSize = 4
		}
		engine.ModalityLayers[modality] = NewSTDPLayer(plasticityConfig, channelCounts[i], hiddenSize)

		// Initialize attention weights uniformly
		engine.AttentionWeights[modality] = 1.0 / float64(len(modalities))

		totalChannels += channelCounts[i]
	}

	// Create fusion layers based on method
	fusionInputSize := 0
	for _, layer := range engine.ModalityLayers {
		fusionInputSize += layer.OutputSize
	}

	switch sensorConfig.FusionMethod {
	case "early":
		engine.EarlyFusion = NewSTDPLayer(plasticityConfig, totalChannels, 32)
	case "late":
		engine.LateFusion = NewSTDPLayer(plasticityConfig, fusionInputSize, 16)
	case "hierarchical":
		engine.EarlyFusion = NewSTDPLayer(plasticityConfig, totalChannels, 64)
		engine.LateFusion = NewSTDPLayer(plasticityConfig, 64, 16)
	}

	return engine
}

// ProcessInput processes a single modality input
func (e *MultimodalFusionEngine) ProcessInput(input *SensorInput, currentTime float64, dt float64) []bool {
	// Get processor for this modality
	processor := e.Processors[input.Modality]

	// Convert raw to spikes if needed
	if len(input.SpikeTrains) == 0 {
		numTimeSteps := int(e.Config.TemporalWindowMs / dt)
		input.SpikeTrains = processor.ProcessRaw(input.RawValues, numTimeSteps)
	}

	// Process through modality-specific layer
	layer := e.ModalityLayers[input.Modality]

	// Take first time step for simplicity (full implementation would iterate)
	if len(input.SpikeTrains) > 0 {
		return layer.Forward(input.SpikeTrains[0], currentTime, dt)
	}

	return make([]bool, layer.OutputSize)
}

// FuseModalities combines outputs from all modalities
func (e *MultimodalFusionEngine) FuseModalities(modalityOutputs map[SensorModality][]bool, currentTime float64, dt float64) []bool {
	switch e.Config.FusionMethod {
	case "early":
		// Concatenate all inputs
		combined := make([]bool, 0)
		for _, output := range modalityOutputs {
			combined = append(combined, output...)
		}
		// Pad if necessary
		for len(combined) < e.EarlyFusion.InputSize {
			combined = append(combined, false)
		}
		return e.EarlyFusion.Forward(combined[:e.EarlyFusion.InputSize], currentTime, dt)

	case "late":
		// Process through modality layers first, then fuse
		combined := make([]bool, 0)
		for modality, output := range modalityOutputs {
			// Apply attention weighting (convert to spike probability)
			weight := e.AttentionWeights[modality]
			for _, spike := range output {
				if spike && rand.Float64() < weight {
					combined = append(combined, true)
				} else {
					combined = append(combined, false)
				}
			}
		}
		for len(combined) < e.LateFusion.InputSize {
			combined = append(combined, false)
		}
		return e.LateFusion.Forward(combined[:e.LateFusion.InputSize], currentTime, dt)

	case "hierarchical":
		// First stage: early fusion
		combined := make([]bool, 0)
		for _, output := range modalityOutputs {
			combined = append(combined, output...)
		}
		for len(combined) < e.EarlyFusion.InputSize {
			combined = append(combined, false)
		}
		intermediate := e.EarlyFusion.Forward(combined[:e.EarlyFusion.InputSize], currentTime, dt)

		// Second stage: late fusion
		return e.LateFusion.Forward(intermediate, currentTime, dt)

	default:
		return make([]bool, 16)
	}
}

// UpdateAttention updates attention weights based on modality importance
func (e *MultimodalFusionEngine) UpdateAttention(importanceScores map[SensorModality]float64) {
	// Softmax normalization
	total := 0.0
	for _, score := range importanceScores {
		total += math.Exp(score)
	}
	for modality, score := range importanceScores {
		e.AttentionWeights[modality] = math.Exp(score) / total
	}
}

// ============================================================================
// INTEGRATED PLASTICITY + SENSOR FUSION SYSTEM
// ============================================================================

// IronLatticePlasticitySensor combines plasticity and sensor fusion
type IronLatticePlasticitySensor struct {
	PlasticityConfig *PlasticityConfig
	SensorConfig     *SensorFusionConfig

	// Components
	FusionEngine     *MultimodalFusionEngine
	OutputLayer      *STDPLayer

	// Task-specific
	TaskType         string // "classification", "detection", "tracking"
	NumClasses       int

	// Performance tracking
	InferenceCount   int
	CorrectCount     int
	TotalEnergy      float64 // pJ
	TotalLatency     float64 // ms
}

// NewIronLatticePlasticitySensor creates an integrated system
func NewIronLatticePlasticitySensor(numClasses int) *IronLatticePlasticitySensor {
	plasticityConfig := DefaultPlasticityConfig()
	sensorConfig := DefaultSensorFusionConfig()

	system := &IronLatticePlasticitySensor{
		PlasticityConfig: plasticityConfig,
		SensorConfig:     sensorConfig,
		NumClasses:       numClasses,
		TaskType:         "classification",
	}

	// Create fusion engine
	system.FusionEngine = NewMultimodalFusionEngine(sensorConfig, plasticityConfig)

	// Create output classification layer
	fusionOutputSize := 16 // From fusion engine
	system.OutputLayer = NewSTDPLayer(plasticityConfig, fusionOutputSize, numClasses)

	return system
}

// Inference performs multimodal inference with plasticity
func (s *IronLatticePlasticitySensor) Inference(inputs map[SensorModality]*SensorInput, currentTime float64) int {
	dt := 1.0 // 1 ms time step

	// Process each modality
	modalityOutputs := make(map[SensorModality][]bool)
	for modality, input := range inputs {
		modalityOutputs[modality] = s.FusionEngine.ProcessInput(input, currentTime, dt)
	}

	// Fuse modalities
	fusedOutput := s.FusionEngine.FuseModalities(modalityOutputs, currentTime+dt, dt)

	// Classify
	classOutput := s.OutputLayer.Forward(fusedOutput, currentTime+2*dt, dt)

	// Find winning class (most spikes)
	maxSpikes := 0
	winningClass := 0
	for i, spike := range classOutput {
		if spike {
			// Count as 1 spike
			if 1 > maxSpikes {
				maxSpikes = 1
				winningClass = i
			}
		}
	}

	s.InferenceCount++

	// Estimate energy
	for _, processor := range s.FusionEngine.Processors {
		s.TotalEnergy += processor.ProcessingEnergy
	}

	return winningClass
}

// Train performs online learning on labeled data
func (s *IronLatticePlasticitySensor) Train(inputs map[SensorModality]*SensorInput, label int, currentTime float64) {
	// Forward pass
	prediction := s.Inference(inputs, currentTime)

	// If correct, reinforce; if wrong, apply correction
	if prediction == label {
		s.CorrectCount++
		// Positive reinforcement: increase weights to winning class
		for j := 0; j < s.OutputLayer.InputSize; j++ {
			s.OutputLayer.Synapses[label][j].Weight *= 1.01
			if s.OutputLayer.Synapses[label][j].Weight > s.PlasticityConfig.WMax {
				s.OutputLayer.Synapses[label][j].Weight = s.PlasticityConfig.WMax
			}
		}
	} else {
		// Negative reinforcement: decrease weights to wrong class
		for j := 0; j < s.OutputLayer.InputSize; j++ {
			s.OutputLayer.Synapses[prediction][j].Weight *= 0.99
			if s.OutputLayer.Synapses[prediction][j].Weight < s.PlasticityConfig.WMin {
				s.OutputLayer.Synapses[prediction][j].Weight = s.PlasticityConfig.WMin
			}
		}
	}
}

// GetAccuracy returns classification accuracy
func (s *IronLatticePlasticitySensor) GetAccuracy() float64 {
	if s.InferenceCount == 0 {
		return 0
	}
	return float64(s.CorrectCount) / float64(s.InferenceCount)
}

// GetPerformanceReport returns detailed performance metrics
func (s *IronLatticePlasticitySensor) GetPerformanceReport() map[string]interface{} {
	// Calculate fusion layer statistics
	fusionLTP := 0
	fusionLTD := 0
	if s.FusionEngine.EarlyFusion != nil {
		fusionLTP += s.FusionEngine.EarlyFusion.TotalLTP
		fusionLTD += s.FusionEngine.EarlyFusion.TotalLTD
	}
	if s.FusionEngine.LateFusion != nil {
		fusionLTP += s.FusionEngine.LateFusion.TotalLTP
		fusionLTD += s.FusionEngine.LateFusion.TotalLTD
	}

	return map[string]interface{}{
		"inference_count":   s.InferenceCount,
		"accuracy":          s.GetAccuracy(),
		"total_energy_pJ":   s.TotalEnergy,
		"energy_per_inf_pJ": s.TotalEnergy / float64(max(s.InferenceCount, 1)),
		"fusion_LTP_events": fusionLTP,
		"fusion_LTD_events": fusionLTD,
		"output_avg_weight": s.OutputLayer.AvgWeight,
		"attention_visual":  s.FusionEngine.AttentionWeights[ModalityVisual],
		"attention_tactile": s.FusionEngine.AttentionWeights[ModalityTactile],
		"power_budget_uW":   s.SensorConfig.PowerBudgetUW,
	}
}

// max returns the larger of two ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ============================================================================
// VISUALIZATION HELPERS
// ============================================================================

// GenerateSTDPWindow creates ASCII visualization of STDP window
func GenerateSTDPWindow(config *PlasticityConfig) string {
	plot := "Spike-Timing Dependent Plasticity (STDP) Window\n"
	plot += "═══════════════════════════════════════════════════════\n\n"

	height := 10
	width := 40

	// Generate STDP curve
	dtValues := make([]float64, width)
	dwValues := make([]float64, width)
	for i := 0; i < width; i++ {
		dt := -50 + 100*float64(i)/float64(width-1) // -50 to +50 ms
		dtValues[i] = dt
		if dt > 0 {
			dwValues[i] = config.APlus * math.Exp(-dt/config.TauPlus)
		} else {
			dwValues[i] = -config.AMinus * math.Exp(dt/config.TauMinus)
		}
	}

	// Find max for scaling
	maxDW := config.APlus
	minDW := -config.AMinus

	// Draw plot
	for row := height; row >= 0; row-- {
		threshold := minDW + (maxDW-minDW)*float64(row)/float64(height)
		var label string
		if row == height {
			label = fmt.Sprintf("%+.3f", maxDW)
		} else if row == 0 {
			label = fmt.Sprintf("%+.3f", minDW)
		} else if math.Abs(threshold) < 0.001 {
			label = " 0.000"
		} else {
			label = "      "
		}
		line := fmt.Sprintf("%s │", label)

		for i := 0; i < width; i++ {
			dw := dwValues[i]
			if i == width/2 {
				line += "│" // Zero line
			} else if (dw >= threshold && threshold >= 0) || (dw <= threshold && threshold <= 0) {
				if dw > 0 {
					line += "█"
				} else {
					line += "▓"
				}
			} else {
				line += " "
			}
		}
		plot += line + "\n"
	}

	// X-axis
	plot += "       └" + repeatChar("─", width) + "\n"
	plot += "       -50ms              0              +50ms\n"
	plot += "              Δt = t_post - t_pre\n\n"
	plot += "█ LTP (pre before post)   ▓ LTD (post before pre)\n"
	plot += fmt.Sprintf("τ+ = %.0f ms, τ- = %.0f ms\n", config.TauPlus, config.TauMinus)
	plot += fmt.Sprintf("A+ = %.3f, A- = %.3f\n", config.APlus, config.AMinus)

	return plot
}

// GenerateSensorFusionDiagram creates ASCII diagram of fusion architecture
func GenerateSensorFusionDiagram(config *SensorFusionConfig) string {
	diagram := "Multimodal Neuromorphic Sensor Fusion Architecture\n"
	diagram += "═══════════════════════════════════════════════════════════════\n\n"

	diagram += "┌─────────────────────────────────────────────────────────────┐\n"
	diagram += "│                    SENSOR MODALITIES                        │\n"
	diagram += "├─────────────────────────────────────────────────────────────┤\n"
	diagram += "│                                                             │\n"
	diagram += fmt.Sprintf("│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │\n")
	diagram += fmt.Sprintf("│  │ Visual  │  │ Tactile │  │ Audio   │  │ Proprio │        │\n")
	diagram += fmt.Sprintf("│  │  (%2d)   │  │  (%2d)   │  │  (%2d)   │  │  (%2d)   │        │\n",
		config.NumVisualChannels, config.NumTactileChannels,
		config.NumAuditoryChannels, config.NumProprioceptive)
	diagram += "│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │\n"
	diagram += "│       │            │            │            │             │\n"
	diagram += "│       ↓            ↓            ↓            ↓             │\n"
	diagram += "│  ┌─────────────────────────────────────────────────────┐   │\n"
	diagram += "│  │              IN-SENSOR PREPROCESSING                │   │\n"
	diagram += "│  │         (Temporal Difference → Spike Trains)        │   │\n"
	diagram += "│  └─────────────────────────────────────────────────────┘   │\n"
	diagram += "│       │            │            │            │             │\n"
	diagram += "│       ↓            ↓            ↓            ↓             │\n"
	diagram += "│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │\n"
	diagram += "│  │ STDP    │  │ STDP    │  │ STDP    │  │ STDP    │        │\n"
	diagram += "│  │ Layer   │  │ Layer   │  │ Layer   │  │ Layer   │        │\n"
	diagram += "│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │\n"
	diagram += "│       │            │            │            │             │\n"
	diagram += "│       └────────────┴─────┬──────┴────────────┘             │\n"
	diagram += "│                          │                                  │\n"
	diagram += "│                          ↓                                  │\n"
	diagram += "│            ┌─────────────────────────────┐                 │\n"
	diagram += fmt.Sprintf("│            │     %s FUSION            │                 │\n", config.FusionMethod)
	diagram += "│            │   (Attention-weighted)      │                 │\n"
	diagram += "│            └─────────────┬───────────────┘                 │\n"
	diagram += "│                          │                                  │\n"
	diagram += "│                          ↓                                  │\n"
	diagram += "│            ┌─────────────────────────────┐                 │\n"
	diagram += "│            │     OUTPUT CLASSIFIER       │                 │\n"
	diagram += "│            │      (STDP Learning)        │                 │\n"
	diagram += "│            └─────────────────────────────┘                 │\n"
	diagram += "│                                                             │\n"
	diagram += "└─────────────────────────────────────────────────────────────┘\n"
	diagram += "\n"
	diagram += fmt.Sprintf("Power Budget: %.0f µW | Latency Target: %.0f ms | Fusion: %s\n",
		config.PowerBudgetUW, config.LatencyTargetMs, config.FusionMethod)

	return diagram
}

// GeneratePPFPlot creates ASCII visualization of pair-pulse facilitation
func GeneratePPFPlot(synapse *FeFETSynapse, numPulses int) string {
	plot := "Pair-Pulse Facilitation (PPF) Response\n"
	plot += "═══════════════════════════════════════════════════════\n\n"

	// Simulate PPF
	responses := make([]float64, numPulses)
	times := make([]float64, numPulses)
	intervalMs := 20.0

	for i := 0; i < numPulses; i++ {
		currentTime := float64(i) * intervalMs
		times[i] = currentTime
		responses[i] = synapse.ProcessPreSpike(currentTime)
	}

	// Find max for scaling
	maxResponse := responses[0]
	for _, r := range responses {
		if r > maxResponse {
			maxResponse = r
		}
	}

	// Draw plot
	height := 8
	width := numPulses
	for row := height; row >= 0; row-- {
		threshold := maxResponse * float64(row) / float64(height)
		line := fmt.Sprintf("%5.2f │", threshold)
		for i := 0; i < width; i++ {
			if responses[i] >= threshold {
				line += "█ "
			} else {
				line += "  "
			}
		}
		plot += line + "\n"
	}

	plot += "      └" + repeatChar("──", width) + "\n"
	plot += "       Pulse number (interval: " + fmt.Sprintf("%.0f", intervalMs) + " ms)\n\n"
	plot += fmt.Sprintf("Initial response: %.3f\n", responses[0])
	plot += fmt.Sprintf("Peak facilitation: %.3f (%.1f%% increase)\n",
		maxResponse, (maxResponse/responses[0]-1)*100)

	return plot
}

// ============================================================================
// EXAMPLE USAGE AND DEMO
// ============================================================================

// RunPlasticitySensorDemo demonstrates plasticity and sensor fusion
func RunPlasticitySensorDemo() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║  IronLattice Synaptic Plasticity & Sensor Fusion Demo     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 1. STDP Window
	fmt.Println("1. STDP Learning Window:")
	config := DefaultPlasticityConfig()
	fmt.Println(GenerateSTDPWindow(config))
	fmt.Println()

	// 2. Single synapse test
	fmt.Println("2. Single FeFET Synapse Test:")
	synapse := NewFeFETSynapse(config)
	fmt.Printf("   Initial weight: %.3f\n", synapse.Weight)

	// Simulate pre-post pairing (should cause LTP)
	synapse.ProcessPreSpike(0)
	synapse.ProcessPostSpike(5) // Post 5ms after pre
	fmt.Printf("   After pre→post (5ms): %.3f (LTP)\n", synapse.Weight)

	// Simulate post-pre pairing (should cause LTD)
	synapse.ProcessPostSpike(20)
	synapse.ProcessPreSpike(25) // Pre 5ms after post
	fmt.Printf("   After post→pre (5ms): %.3f (LTD)\n", synapse.Weight)
	fmt.Printf("   Total energy: %.1f aJ\n", synapse.TotalEnergy)
	fmt.Println()

	// 3. PPF demonstration
	fmt.Println("3. Pair-Pulse Facilitation:")
	ppfSynapse := NewFeFETSynapse(config)
	fmt.Println(GeneratePPFPlot(ppfSynapse, 10))
	fmt.Println()

	// 4. Sensor fusion architecture
	fmt.Println("4. Multimodal Sensor Fusion Architecture:")
	sensorConfig := DefaultSensorFusionConfig()
	fmt.Println(GenerateSensorFusionDiagram(sensorConfig))
	fmt.Println()

	// 5. Integrated system test
	fmt.Println("5. Integrated Plasticity + Sensor Fusion System:")
	system := NewIronLatticePlasticitySensor(10) // 10 classes

	// Generate synthetic multimodal data
	for i := 0; i < 100; i++ {
		inputs := make(map[SensorModality]*SensorInput)

		// Visual input
		visualData := make([]float64, sensorConfig.NumVisualChannels)
		for j := range visualData {
			visualData[j] = rand.Float64()
		}
		inputs[ModalityVisual] = &SensorInput{
			Modality:  ModalityVisual,
			Timestamp: float64(i),
			RawValues: visualData,
		}

		// Tactile input
		tactileData := make([]float64, sensorConfig.NumTactileChannels)
		for j := range tactileData {
			tactileData[j] = rand.Float64()
		}
		inputs[ModalityTactile] = &SensorInput{
			Modality:  ModalityTactile,
			Timestamp: float64(i),
			RawValues: tactileData,
		}

		// Train with random label
		label := rand.Intn(10)
		system.Train(inputs, label, float64(i)*10)
	}

	report := system.GetPerformanceReport()
	fmt.Printf("   Inferences: %d\n", report["inference_count"])
	fmt.Printf("   Accuracy: %.1f%%\n", report["accuracy"].(float64)*100)
	fmt.Printf("   Total energy: %.2f pJ\n", report["total_energy_pJ"])
	fmt.Printf("   LTP events: %d\n", report["fusion_LTP_events"])
	fmt.Printf("   LTD events: %d\n", report["fusion_LTD_events"])
	fmt.Printf("   Visual attention: %.2f\n", report["attention_visual"])
	fmt.Printf("   Tactile attention: %.2f\n", report["attention_tactile"])
	fmt.Println()

	// 6. STDP layer statistics
	fmt.Println("6. STDP Layer Learning Statistics:")
	stdpLayer := NewSTDPLayer(config, 64, 16)

	// Run some patterns through
	for epoch := 0; epoch < 50; epoch++ {
		inputSpikes := make([]bool, 64)
		for j := range inputSpikes {
			inputSpikes[j] = rand.Float64() < 0.2 // 20% firing rate
		}
		stdpLayer.Forward(inputSpikes, float64(epoch), 1.0)
	}

	fmt.Printf("   Total LTP events: %d\n", stdpLayer.TotalLTP)
	fmt.Printf("   Total LTD events: %d\n", stdpLayer.TotalLTD)
	fmt.Printf("   Average weight: %.3f\n", stdpLayer.AvgWeight)
	fmt.Printf("   LTP/LTD ratio: %.2f\n", float64(stdpLayer.TotalLTP)/float64(max(stdpLayer.TotalLTD, 1)))
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("Plasticity + Sensor Fusion simulation complete!")
}

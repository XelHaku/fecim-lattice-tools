// Package layers provides spiking neural networks and graph neural networks for CIM accelerators.
//
// This module implements bio-inspired spiking computations and graph-structured
// learning optimized for ferroelectric compute-in-memory hardware.
//
// Key features:
// - Leaky Integrate-and-Fire (LIF) neuron model
// - Spike-Timing Dependent Plasticity (STDP)
// - Surrogate gradient training
// - Graph convolutional networks (GCN)
// - Message passing neural networks
// - Sparse adjacency matrix handling
//
// References:
// - "Personalized SNN with ferroelectric synapses" (arXiv 2601.00020)
// - "Si:HfO2 ferroelectric tunnel memristor for SNN" (Nano Energy 2022)
// - "Fully memristive SNN for graph learning" (PMC 2025)
// - "Random memristor dynamic graph CNN" (npj Unconventional Computing 2024)
// - "NEM-GNN: Near-memory GNN accelerator" (ACM TACO 2024)
package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// Spiking Neural Network Core
// =============================================================================

// SNNConfig configures spiking neural network parameters.
type SNNConfig struct {
	NeuronModel     string  // "lif", "izhikevich", "hodgkin_huxley"
	TimeSteps       int     // simulation time steps
	DeltaT          float64 // time step (ms)
	Threshold       float64 // spike threshold
	ResetPotential  float64 // post-spike reset
	RestPotential   float64 // resting potential
	TauMembrane     float64 // membrane time constant (ms)
	TauSynaptic     float64 // synaptic time constant (ms)
	RefractoryPeriod int    // refractory steps
	UseSurrogate    bool    // use surrogate gradients
	SurrogateSlope  float64 // surrogate gradient slope
}

// DefaultSNNConfig returns standard SNN settings.
func DefaultSNNConfig() SNNConfig {
	return SNNConfig{
		NeuronModel:     "lif",
		TimeSteps:       100,
		DeltaT:          1.0,
		Threshold:       1.0,
		ResetPotential:  0.0,
		RestPotential:   0.0,
		TauMembrane:     20.0,
		TauSynaptic:     5.0,
		RefractoryPeriod: 2,
		UseSurrogate:    true,
		SurrogateSlope:  25.0,
	}
}

// =============================================================================
// Leaky Integrate-and-Fire Neuron
// =============================================================================

// LIFNeuron implements the Leaky Integrate-and-Fire model.
type LIFNeuron struct {
	Config          SNNConfig
	Membrane        float64   // membrane potential
	SpikeHistory    []bool    // spike train
	RefractoryCount int       // refractory counter
	InputCurrent    float64   // accumulated input
	LastSpikeTime   int       // time of last spike
}

// NewLIFNeuron creates a LIF neuron.
func NewLIFNeuron(config SNNConfig) *LIFNeuron {
	return &LIFNeuron{
		Config:       config,
		Membrane:     config.RestPotential,
		SpikeHistory: make([]bool, 0, config.TimeSteps),
	}
}

// Step performs one simulation step.
func (n *LIFNeuron) Step(input float64, t int) bool {
	// Check refractory period
	if n.RefractoryCount > 0 {
		n.RefractoryCount--
		n.Membrane = n.Config.ResetPotential
		n.SpikeHistory = append(n.SpikeHistory, false)
		return false
	}

	// Leaky integration
	decay := math.Exp(-n.Config.DeltaT / n.Config.TauMembrane)
	n.Membrane = decay*n.Membrane + (1-decay)*input

	// Check threshold
	spiked := n.Membrane >= n.Config.Threshold
	if spiked {
		n.Membrane = n.Config.ResetPotential
		n.RefractoryCount = n.Config.RefractoryPeriod
		n.LastSpikeTime = t
	}

	n.SpikeHistory = append(n.SpikeHistory, spiked)
	return spiked
}

// Reset clears neuron state.
func (n *LIFNeuron) Reset() {
	n.Membrane = n.Config.RestPotential
	n.SpikeHistory = n.SpikeHistory[:0]
	n.RefractoryCount = 0
	n.LastSpikeTime = -1
}

// SurrogateGradient computes gradient through spike function.
func (n *LIFNeuron) SurrogateGradient(membrane float64) float64 {
	if !n.Config.UseSurrogate {
		// Hard threshold (non-differentiable)
		if membrane >= n.Config.Threshold {
			return 1.0
		}
		return 0.0
	}

	// Fast sigmoid surrogate
	x := n.Config.SurrogateSlope * (membrane - n.Config.Threshold)
	sig := 1.0 / (1.0 + math.Abs(x))
	return n.Config.SurrogateSlope * sig * sig
}

// =============================================================================
// Spiking Layer
// =============================================================================

// SpikingLayer implements a layer of spiking neurons.
type SpikingLayer struct {
	Config   SNNConfig
	Neurons  []*LIFNeuron
	Weights  [][]float64
	Biases   []float64
	Size     int
	InputSize int
}

// NewSpikingLayer creates a spiking layer.
func NewSpikingLayer(inputSize, outputSize int, config SNNConfig, seed int64) *SpikingLayer {
	rng := rand.New(rand.NewSource(seed))
	layer := &SpikingLayer{
		Config:    config,
		Neurons:   make([]*LIFNeuron, outputSize),
		Weights:   make([][]float64, outputSize),
		Biases:    make([]float64, outputSize),
		Size:      outputSize,
		InputSize: inputSize,
	}

	// Initialize neurons and weights
	scale := math.Sqrt(2.0 / float64(inputSize))
	for i := 0; i < outputSize; i++ {
		layer.Neurons[i] = NewLIFNeuron(config)
		layer.Weights[i] = make([]float64, inputSize)
		for j := 0; j < inputSize; j++ {
			layer.Weights[i][j] = rng.NormFloat64() * scale
		}
	}

	return layer
}

// Forward processes input spikes for one time step.
func (l *SpikingLayer) Forward(inputSpikes []bool, t int) []bool {
	outputSpikes := make([]bool, l.Size)

	for i, neuron := range l.Neurons {
		// Compute weighted input
		var current float64
		for j, spike := range inputSpikes {
			if spike {
				current += l.Weights[i][j]
			}
		}
		current += l.Biases[i]

		outputSpikes[i] = neuron.Step(current, t)
	}

	return outputSpikes
}

// Reset clears all neuron states.
func (l *SpikingLayer) Reset() {
	for _, neuron := range l.Neurons {
		neuron.Reset()
	}
}

// GetSpikeRates computes average firing rates.
func (l *SpikingLayer) GetSpikeRates() []float64 {
	rates := make([]float64, l.Size)
	for i, neuron := range l.Neurons {
		spikeCount := 0
		for _, spike := range neuron.SpikeHistory {
			if spike {
				spikeCount++
			}
		}
		rates[i] = float64(spikeCount) / float64(len(neuron.SpikeHistory))
	}
	return rates
}

// =============================================================================
// Spike-Timing Dependent Plasticity (STDP)
// =============================================================================

// STDPConfig configures STDP learning rule.
type STDPConfig struct {
	TauPlus      float64 // LTP time constant (ms)
	TauMinus     float64 // LTD time constant (ms)
	APlus        float64 // LTP amplitude
	AMinus       float64 // LTD amplitude
	WMax         float64 // maximum weight
	WMin         float64 // minimum weight
	LearningRate float64
	Additive     bool    // additive vs multiplicative
}

// DefaultSTDPConfig returns standard STDP settings.
func DefaultSTDPConfig() STDPConfig {
	return STDPConfig{
		TauPlus:      20.0,
		TauMinus:     20.0,
		APlus:        0.01,
		AMinus:       0.012,
		WMax:         1.0,
		WMin:         0.0,
		LearningRate: 0.01,
		Additive:     false,
	}
}

// STDPLearner implements STDP-based weight updates.
type STDPLearner struct {
	Config      STDPConfig
	PreTraces   []float64 // presynaptic eligibility traces
	PostTraces  []float64 // postsynaptic eligibility traces
}

// NewSTDPLearner creates an STDP learner.
func NewSTDPLearner(preSize, postSize int, config STDPConfig) *STDPLearner {
	return &STDPLearner{
		Config:     config,
		PreTraces:  make([]float64, preSize),
		PostTraces: make([]float64, postSize),
	}
}

// UpdateTraces updates eligibility traces based on spikes.
func (s *STDPLearner) UpdateTraces(preSpikes, postSpikes []bool, dt float64) {
	// Decay traces
	decayPre := math.Exp(-dt / s.Config.TauPlus)
	decayPost := math.Exp(-dt / s.Config.TauMinus)

	for i := range s.PreTraces {
		s.PreTraces[i] *= decayPre
		if i < len(preSpikes) && preSpikes[i] {
			s.PreTraces[i] += 1.0
		}
	}

	for i := range s.PostTraces {
		s.PostTraces[i] *= decayPost
		if i < len(postSpikes) && postSpikes[i] {
			s.PostTraces[i] += 1.0
		}
	}
}

// ComputeWeightUpdate computes STDP weight changes.
func (s *STDPLearner) ComputeWeightUpdate(
	preSpikes, postSpikes []bool,
	weights [][]float64,
) [][]float64 {
	deltaW := make([][]float64, len(weights))

	for i := range weights {
		deltaW[i] = make([]float64, len(weights[i]))
		for j := range weights[i] {
			w := weights[i][j]

			// LTP: post after pre
			if i < len(postSpikes) && postSpikes[i] && j < len(s.PreTraces) {
				if s.Config.Additive {
					deltaW[i][j] += s.Config.APlus * s.PreTraces[j]
				} else {
					// Multiplicative (soft bounds)
					deltaW[i][j] += s.Config.APlus * s.PreTraces[j] * (s.Config.WMax - w)
				}
			}

			// LTD: pre after post
			if j < len(preSpikes) && preSpikes[j] && i < len(s.PostTraces) {
				if s.Config.Additive {
					deltaW[i][j] -= s.Config.AMinus * s.PostTraces[i]
				} else {
					deltaW[i][j] -= s.Config.AMinus * s.PostTraces[i] * (w - s.Config.WMin)
				}
			}
		}
	}

	return deltaW
}

// ApplyUpdate applies weight changes with bounds.
func (s *STDPLearner) ApplyUpdate(weights [][]float64, deltaW [][]float64) {
	for i := range weights {
		for j := range weights[i] {
			weights[i][j] += s.Config.LearningRate * deltaW[i][j]
			// Enforce bounds
			weights[i][j] = math.Max(s.Config.WMin, math.Min(s.Config.WMax, weights[i][j]))
		}
	}
}

// =============================================================================
// Ferroelectric Synapse Model
// =============================================================================

// FerroelectricSynapseConfig configures FeFET/FTJ synapse.
type FerroelectricSynapseConfig struct {
	PolarizationMin    float64 // minimum polarization
	PolarizationMax    float64 // maximum polarization
	CoerciveVoltage    float64 // Ec
	SaturationSlope    float64 // polarization switching slope
	RetentionTime      float64 // retention time constant (s)
	EnduranceCycles    int     // max program cycles
	D2DVariation       float64 // device-to-device variation
	ProgrammingEnergy  float64 // energy per update (fJ)
}

// DefaultFerroelectricSynapseConfig returns standard FeFET settings.
func DefaultFerroelectricSynapseConfig() FerroelectricSynapseConfig {
	return FerroelectricSynapseConfig{
		PolarizationMin:   -1.0,
		PolarizationMax:   1.0,
		CoerciveVoltage:   0.5,
		SaturationSlope:   5.0,
		RetentionTime:     1e8, // ~10 years
		EnduranceCycles:   1e12,
		D2DVariation:      0.05,
		ProgrammingEnergy: 32.65, // fJ per update
	}
}

// FerroelectricSynapse models a ferroelectric memristive synapse.
type FerroelectricSynapse struct {
	Config       FerroelectricSynapseConfig
	Polarization float64 // current polarization state
	Weight       float64 // derived synaptic weight
	Cycles       int     // programming cycles used
	D2DFactor    float64 // device-specific variation
}

// NewFerroelectricSynapse creates a ferroelectric synapse.
func NewFerroelectricSynapse(config FerroelectricSynapseConfig, seed int64) *FerroelectricSynapse {
	rng := rand.New(rand.NewSource(seed))
	return &FerroelectricSynapse{
		Config:       config,
		Polarization: 0.0,
		D2DFactor:    1.0 + rng.NormFloat64()*config.D2DVariation,
	}
}

// Program applies voltage pulse to modify polarization.
func (s *FerroelectricSynapse) Program(voltage, pulseWidth float64) {
	// Preisach-like switching
	if math.Abs(voltage) > s.Config.CoerciveVoltage {
		direction := 1.0
		if voltage < 0 {
			direction = -1.0
		}

		// Tanh switching characteristic
		deltaP := direction * math.Tanh(s.Config.SaturationSlope*(math.Abs(voltage)-s.Config.CoerciveVoltage)) * pulseWidth

		s.Polarization += deltaP
		s.Polarization = math.Max(s.Config.PolarizationMin,
			math.Min(s.Config.PolarizationMax, s.Polarization))

		s.Cycles++
	}

	// Update weight from polarization
	s.Weight = s.polarizationToWeight()
}

// polarizationToWeight maps polarization to synaptic weight.
func (s *FerroelectricSynapse) polarizationToWeight() float64 {
	// Linear mapping with D2D variation
	normalized := (s.Polarization - s.Config.PolarizationMin) /
		(s.Config.PolarizationMax - s.Config.PolarizationMin)
	return normalized * s.D2DFactor
}

// Read returns current synaptic weight with read noise.
func (s *FerroelectricSynapse) Read() float64 {
	// Add small read noise
	noise := rand.NormFloat64() * 0.01
	return s.Weight + noise
}

// =============================================================================
// Spike Encoding/Decoding
// =============================================================================

// SpikeEncoder encodes analog values to spike trains.
type SpikeEncoder struct {
	Method    string  // "rate", "temporal", "delta", "phase"
	TimeSteps int
	MaxRate   float64 // max firing rate for rate coding
	Threshold float64 // threshold for delta coding
}

// NewSpikeEncoder creates a spike encoder.
func NewSpikeEncoder(method string, timeSteps int) *SpikeEncoder {
	return &SpikeEncoder{
		Method:    method,
		TimeSteps: timeSteps,
		MaxRate:   0.9, // max 90% firing rate
		Threshold: 0.1,
	}
}

// Encode converts analog values to spike trains.
func (e *SpikeEncoder) Encode(values []float64) [][]bool {
	spikes := make([][]bool, len(values))

	switch e.Method {
	case "rate":
		// Rate coding: spike probability proportional to value
		for i, v := range values {
			spikes[i] = make([]bool, e.TimeSteps)
			rate := math.Max(0, math.Min(e.MaxRate, v))
			for t := 0; t < e.TimeSteps; t++ {
				spikes[i][t] = rand.Float64() < rate
			}
		}

	case "temporal":
		// Temporal coding: spike time encodes value (earlier = higher)
		for i, v := range values {
			spikes[i] = make([]bool, e.TimeSteps)
			// Higher values spike earlier
			spikeTime := int((1 - math.Max(0, math.Min(1, v))) * float64(e.TimeSteps-1))
			if spikeTime < e.TimeSteps {
				spikes[i][spikeTime] = true
			}
		}

	case "delta":
		// Delta modulation: spike on significant change
		for i, v := range values {
			spikes[i] = make([]bool, e.TimeSteps)
			accumulated := 0.0
			for t := 0; t < e.TimeSteps; t++ {
				accumulated += v / float64(e.TimeSteps)
				if accumulated >= e.Threshold {
					spikes[i][t] = true
					accumulated -= e.Threshold
				}
			}
		}

	case "phase":
		// Phase coding: spike phase encodes value
		for i, v := range values {
			spikes[i] = make([]bool, e.TimeSteps)
			phase := math.Max(0, math.Min(1, v)) * 2 * math.Pi
			for t := 0; t < e.TimeSteps; t++ {
				currentPhase := float64(t) / float64(e.TimeSteps) * 2 * math.Pi
				if math.Abs(math.Mod(currentPhase-phase+math.Pi, 2*math.Pi)-math.Pi) < 0.1 {
					spikes[i][t] = true
				}
			}
		}
	}

	return spikes
}

// SpikeDecoder decodes spike trains to analog values.
type SpikeDecoder struct {
	Method string // "rate", "first_spike", "weighted"
}

// NewSpikeDecoder creates a spike decoder.
func NewSpikeDecoder(method string) *SpikeDecoder {
	return &SpikeDecoder{Method: method}
}

// Decode converts spike trains to analog values.
func (d *SpikeDecoder) Decode(spikes [][]bool) []float64 {
	values := make([]float64, len(spikes))

	switch d.Method {
	case "rate":
		// Count spikes / total time
		for i, train := range spikes {
			count := 0
			for _, spike := range train {
				if spike {
					count++
				}
			}
			values[i] = float64(count) / float64(len(train))
		}

	case "first_spike":
		// Earlier spike = higher value
		for i, train := range spikes {
			firstSpike := len(train)
			for t, spike := range train {
				if spike {
					firstSpike = t
					break
				}
			}
			values[i] = 1 - float64(firstSpike)/float64(len(train))
		}

	case "weighted":
		// Time-weighted spike count
		for i, train := range spikes {
			var weighted float64
			for t, spike := range train {
				if spike {
					// Later spikes weighted less
					weighted += math.Exp(-float64(t) / float64(len(train)) * 2)
				}
			}
			values[i] = weighted
		}
	}

	return values
}

// =============================================================================
// Spiking Neural Network
// =============================================================================

// SpikingNetwork implements a multi-layer SNN.
type SpikingNetwork struct {
	Config    SNNConfig
	Layers    []*SpikingLayer
	Encoder   *SpikeEncoder
	Decoder   *SpikeDecoder
	STDPLearners []*STDPLearner
}

// NewSpikingNetwork creates a spiking neural network.
func NewSpikingNetwork(
	layerSizes []int,
	config SNNConfig,
	seed int64,
) *SpikingNetwork {
	snn := &SpikingNetwork{
		Config:  config,
		Layers:  make([]*SpikingLayer, len(layerSizes)-1),
		Encoder: NewSpikeEncoder("rate", config.TimeSteps),
		Decoder: NewSpikeDecoder("rate"),
		STDPLearners: make([]*STDPLearner, len(layerSizes)-1),
	}

	for i := 0; i < len(layerSizes)-1; i++ {
		snn.Layers[i] = NewSpikingLayer(layerSizes[i], layerSizes[i+1], config, seed+int64(i))
		snn.STDPLearners[i] = NewSTDPLearner(layerSizes[i], layerSizes[i+1], DefaultSTDPConfig())
	}

	return snn
}

// Forward runs inference for all time steps.
func (snn *SpikingNetwork) Forward(input []float64) []float64 {
	// Reset all neurons
	for _, layer := range snn.Layers {
		layer.Reset()
	}

	// Encode input to spikes
	inputSpikes := snn.Encoder.Encode(input)

	// Run simulation
	for t := 0; t < snn.Config.TimeSteps; t++ {
		// Get input spikes at time t
		currentInput := make([]bool, len(input))
		for i := range currentInput {
			if t < len(inputSpikes[i]) {
				currentInput[i] = inputSpikes[i][t]
			}
		}

		// Propagate through layers
		layerInput := currentInput
		for _, layer := range snn.Layers {
			layerInput = layer.Forward(layerInput, t)
		}
	}

	// Decode output spikes
	lastLayer := snn.Layers[len(snn.Layers)-1]
	outputSpikes := make([][]bool, lastLayer.Size)
	for i, neuron := range lastLayer.Neurons {
		outputSpikes[i] = neuron.SpikeHistory
	}

	return snn.Decoder.Decode(outputSpikes)
}

// TrainSTDP performs one STDP training iteration.
func (snn *SpikingNetwork) TrainSTDP() {
	dt := snn.Config.DeltaT

	for l, layer := range snn.Layers {
		learner := snn.STDPLearners[l]

		// Get spike histories
		var preSpikes, postSpikes []bool
		if l == 0 {
			// First layer: use encoded input
			preSpikes = make([]bool, layer.InputSize)
		} else {
			prevLayer := snn.Layers[l-1]
			preSpikes = make([]bool, prevLayer.Size)
			for i, n := range prevLayer.Neurons {
				if len(n.SpikeHistory) > 0 {
					preSpikes[i] = n.SpikeHistory[len(n.SpikeHistory)-1]
				}
			}
		}

		postSpikes = make([]bool, layer.Size)
		for i, n := range layer.Neurons {
			if len(n.SpikeHistory) > 0 {
				postSpikes[i] = n.SpikeHistory[len(n.SpikeHistory)-1]
			}
		}

		// Update traces and weights
		learner.UpdateTraces(preSpikes, postSpikes, dt)
		deltaW := learner.ComputeWeightUpdate(preSpikes, postSpikes, layer.Weights)
		learner.ApplyUpdate(layer.Weights, deltaW)
	}
}

// =============================================================================
// Graph Neural Network Core
// =============================================================================

// GNNConfig configures graph neural network parameters.
type GNNConfig struct {
	NumLayers       int     // number of GNN layers
	HiddenDim       int     // hidden dimension
	Aggregation     string  // "mean", "sum", "max"
	Activation      string  // "relu", "leaky_relu", "gelu"
	Dropout         float64 // dropout rate
	UseAttention    bool    // use graph attention
	NumHeads        int     // attention heads
	UseBatchNorm    bool    // batch normalization
	ResidualConnect bool    // residual connections
}

// DefaultGNNConfig returns standard GNN settings.
func DefaultGNNConfig() GNNConfig {
	return GNNConfig{
		NumLayers:       3,
		HiddenDim:       64,
		Aggregation:     "mean",
		Activation:      "relu",
		Dropout:         0.5,
		UseAttention:    false,
		NumHeads:        4,
		UseBatchNorm:    true,
		ResidualConnect: true,
	}
}

// =============================================================================
// Graph Data Structures
// =============================================================================

// Graph represents a graph with node features.
type Graph struct {
	NumNodes       int
	NumEdges       int
	NodeFeatures   [][]float64 // [num_nodes][feature_dim]
	EdgeIndex      [][2]int    // [num_edges][2] (source, target)
	EdgeWeights    []float64   // optional edge weights
	AdjacencyList  map[int][]int // node -> neighbors
	Labels         []int       // node labels (for classification)
}

// NewGraph creates a graph from edge list.
func NewGraph(numNodes int, edges [][2]int, features [][]float64) *Graph {
	g := &Graph{
		NumNodes:      numNodes,
		NumEdges:      len(edges),
		NodeFeatures:  features,
		EdgeIndex:     edges,
		AdjacencyList: make(map[int][]int),
	}

	// Build adjacency list
	for _, edge := range edges {
		src, dst := edge[0], edge[1]
		g.AdjacencyList[src] = append(g.AdjacencyList[src], dst)
	}

	return g
}

// GetNeighbors returns neighbors of a node.
func (g *Graph) GetNeighbors(node int) []int {
	return g.AdjacencyList[node]
}

// ToSparseAdjacency converts to sparse adjacency matrix.
func (g *Graph) ToSparseAdjacency() *SparseMatrix {
	sparse := NewSparseMatrix(g.NumNodes, g.NumNodes)
	for _, edge := range g.EdgeIndex {
		sparse.Set(edge[0], edge[1], 1.0)
	}
	return sparse
}

// SparseMatrix implements sparse matrix storage.
type SparseMatrix struct {
	Rows    int
	Cols    int
	Data    map[int]map[int]float64
	NNZ     int // number of non-zeros
}

// NewSparseMatrix creates a sparse matrix.
func NewSparseMatrix(rows, cols int) *SparseMatrix {
	return &SparseMatrix{
		Rows: rows,
		Cols: cols,
		Data: make(map[int]map[int]float64),
	}
}

// Set sets a value in the sparse matrix.
func (s *SparseMatrix) Set(row, col int, val float64) {
	if val == 0 {
		return
	}
	if s.Data[row] == nil {
		s.Data[row] = make(map[int]float64)
	}
	if _, exists := s.Data[row][col]; !exists {
		s.NNZ++
	}
	s.Data[row][col] = val
}

// Get retrieves a value from the sparse matrix.
func (s *SparseMatrix) Get(row, col int) float64 {
	if s.Data[row] == nil {
		return 0
	}
	return s.Data[row][col]
}

// SpMV performs sparse matrix-vector multiplication.
func (s *SparseMatrix) SpMV(vec []float64) []float64 {
	result := make([]float64, s.Rows)
	for row, cols := range s.Data {
		for col, val := range cols {
			if col < len(vec) {
				result[row] += val * vec[col]
			}
		}
	}
	return result
}

// =============================================================================
// Graph Convolutional Layer
// =============================================================================

// GraphConvLayer implements graph convolution.
type GraphConvLayer struct {
	Config     GNNConfig
	InputDim   int
	OutputDim  int
	Weights    [][]float64 // transformation weights
	Bias       []float64
}

// NewGraphConvLayer creates a graph convolution layer.
func NewGraphConvLayer(inputDim, outputDim int, config GNNConfig, seed int64) *GraphConvLayer {
	rng := rand.New(rand.NewSource(seed))
	layer := &GraphConvLayer{
		Config:    config,
		InputDim:  inputDim,
		OutputDim: outputDim,
		Weights:   make([][]float64, outputDim),
		Bias:      make([]float64, outputDim),
	}

	// Xavier initialization
	scale := math.Sqrt(2.0 / float64(inputDim+outputDim))
	for i := 0; i < outputDim; i++ {
		layer.Weights[i] = make([]float64, inputDim)
		for j := 0; j < inputDim; j++ {
			layer.Weights[i][j] = rng.NormFloat64() * scale
		}
	}

	return layer
}

// Forward performs graph convolution.
func (l *GraphConvLayer) Forward(graph *Graph, nodeFeatures [][]float64) [][]float64 {
	numNodes := len(nodeFeatures)
	output := make([][]float64, numNodes)

	for node := 0; node < numNodes; node++ {
		// Aggregate neighbor features
		aggregated := l.aggregateNeighbors(graph, nodeFeatures, node)

		// Transform
		output[node] = make([]float64, l.OutputDim)
		for i := 0; i < l.OutputDim; i++ {
			for j := 0; j < l.InputDim; j++ {
				output[node][i] += l.Weights[i][j] * aggregated[j]
			}
			output[node][i] += l.Bias[i]
		}

		// Activation
		for i := range output[node] {
			output[node][i] = l.activate(output[node][i])
		}
	}

	return output
}

// aggregateNeighbors aggregates neighbor features.
func (l *GraphConvLayer) aggregateNeighbors(graph *Graph, features [][]float64, node int) []float64 {
	neighbors := graph.GetNeighbors(node)
	if len(neighbors) == 0 {
		// Return self features if no neighbors
		return features[node]
	}

	aggregated := make([]float64, l.InputDim)

	switch l.Config.Aggregation {
	case "sum":
		for _, neighbor := range neighbors {
			for j := 0; j < l.InputDim && j < len(features[neighbor]); j++ {
				aggregated[j] += features[neighbor][j]
			}
		}
		// Add self
		for j := 0; j < l.InputDim && j < len(features[node]); j++ {
			aggregated[j] += features[node][j]
		}

	case "mean":
		for _, neighbor := range neighbors {
			for j := 0; j < l.InputDim && j < len(features[neighbor]); j++ {
				aggregated[j] += features[neighbor][j]
			}
		}
		// Add self
		for j := 0; j < l.InputDim && j < len(features[node]); j++ {
			aggregated[j] += features[node][j]
		}
		// Normalize
		norm := float64(len(neighbors) + 1)
		for j := range aggregated {
			aggregated[j] /= norm
		}

	case "max":
		// Initialize with negative infinity
		for j := range aggregated {
			aggregated[j] = math.Inf(-1)
		}
		for _, neighbor := range neighbors {
			for j := 0; j < l.InputDim && j < len(features[neighbor]); j++ {
				if features[neighbor][j] > aggregated[j] {
					aggregated[j] = features[neighbor][j]
				}
			}
		}
		// Compare with self
		for j := 0; j < l.InputDim && j < len(features[node]); j++ {
			if features[node][j] > aggregated[j] {
				aggregated[j] = features[node][j]
			}
		}
	}

	return aggregated
}

// activate applies activation function.
func (l *GraphConvLayer) activate(x float64) float64 {
	switch l.Config.Activation {
	case "relu":
		if x > 0 {
			return x
		}
		return 0
	case "leaky_relu":
		if x > 0 {
			return x
		}
		return 0.01 * x
	case "gelu":
		return 0.5 * x * (1 + math.Tanh(math.Sqrt(2/math.Pi)*(x+0.044715*x*x*x)))
	}
	return x
}

// =============================================================================
// Graph Attention Layer
// =============================================================================

// GraphAttentionLayer implements graph attention network (GAT).
type GraphAttentionLayer struct {
	Config      GNNConfig
	InputDim    int
	OutputDim   int
	NumHeads    int
	HeadDim     int
	QueryWeights [][]float64 // per head
	KeyWeights   [][]float64
	ValueWeights [][]float64
	OutWeights   [][]float64
}

// NewGraphAttentionLayer creates a GAT layer.
func NewGraphAttentionLayer(inputDim, outputDim int, config GNNConfig, seed int64) *GraphAttentionLayer {
	rng := rand.New(rand.NewSource(seed))
	headDim := outputDim / config.NumHeads

	layer := &GraphAttentionLayer{
		Config:       config,
		InputDim:     inputDim,
		OutputDim:    outputDim,
		NumHeads:     config.NumHeads,
		HeadDim:      headDim,
		QueryWeights: make([][]float64, config.NumHeads),
		KeyWeights:   make([][]float64, config.NumHeads),
		ValueWeights: make([][]float64, config.NumHeads),
		OutWeights:   make([][]float64, outputDim),
	}

	// Initialize weights
	scale := math.Sqrt(2.0 / float64(inputDim))
	for h := 0; h < config.NumHeads; h++ {
		layer.QueryWeights[h] = make([]float64, inputDim*headDim)
		layer.KeyWeights[h] = make([]float64, inputDim*headDim)
		layer.ValueWeights[h] = make([]float64, inputDim*headDim)
		for i := range layer.QueryWeights[h] {
			layer.QueryWeights[h][i] = rng.NormFloat64() * scale
			layer.KeyWeights[h][i] = rng.NormFloat64() * scale
			layer.ValueWeights[h][i] = rng.NormFloat64() * scale
		}
	}

	for i := 0; i < outputDim; i++ {
		layer.OutWeights[i] = make([]float64, outputDim)
		for j := 0; j < outputDim; j++ {
			layer.OutWeights[i][j] = rng.NormFloat64() * scale
		}
	}

	return layer
}

// Forward performs graph attention.
func (l *GraphAttentionLayer) Forward(graph *Graph, nodeFeatures [][]float64) [][]float64 {
	numNodes := len(nodeFeatures)
	output := make([][]float64, numNodes)

	for node := 0; node < numNodes; node++ {
		// Multi-head attention
		headOutputs := make([][]float64, l.NumHeads)

		for h := 0; h < l.NumHeads; h++ {
			headOutputs[h] = l.computeHeadAttention(graph, nodeFeatures, node, h)
		}

		// Concatenate heads
		output[node] = make([]float64, l.OutputDim)
		for h := 0; h < l.NumHeads; h++ {
			for d := 0; d < l.HeadDim && h*l.HeadDim+d < l.OutputDim; d++ {
				output[node][h*l.HeadDim+d] = headOutputs[h][d]
			}
		}
	}

	return output
}

// computeHeadAttention computes attention for one head.
func (l *GraphAttentionLayer) computeHeadAttention(
	graph *Graph,
	features [][]float64,
	node int,
	head int,
) []float64 {
	neighbors := graph.GetNeighbors(node)
	if len(neighbors) == 0 {
		// Return transformed self features
		output := make([]float64, l.HeadDim)
		for d := 0; d < l.HeadDim; d++ {
			for j := 0; j < l.InputDim && j < len(features[node]); j++ {
				output[d] += l.ValueWeights[head][d*l.InputDim+j] * features[node][j]
			}
		}
		return output
	}

	// Compute attention scores
	scores := make([]float64, len(neighbors)+1) // +1 for self

	// Query for this node
	query := make([]float64, l.HeadDim)
	for d := 0; d < l.HeadDim; d++ {
		for j := 0; j < l.InputDim && j < len(features[node]); j++ {
			query[d] += l.QueryWeights[head][d*l.InputDim+j] * features[node][j]
		}
	}

	// Compute attention with neighbors
	for i, neighbor := range neighbors {
		key := make([]float64, l.HeadDim)
		for d := 0; d < l.HeadDim; d++ {
			for j := 0; j < l.InputDim && j < len(features[neighbor]); j++ {
				key[d] += l.KeyWeights[head][d*l.InputDim+j] * features[neighbor][j]
			}
		}

		// Dot product attention
		for d := 0; d < l.HeadDim; d++ {
			scores[i] += query[d] * key[d]
		}
		scores[i] /= math.Sqrt(float64(l.HeadDim))
	}

	// Self attention
	for d := 0; d < l.HeadDim; d++ {
		for j := 0; j < l.InputDim && j < len(features[node]); j++ {
			scores[len(neighbors)] += query[d] * l.KeyWeights[head][d*l.InputDim+j] * features[node][j]
		}
	}
	scores[len(neighbors)] /= math.Sqrt(float64(l.HeadDim))

	// Softmax
	maxScore := scores[0]
	for _, s := range scores[1:] {
		if s > maxScore {
			maxScore = s
		}
	}
	var sumExp float64
	for i := range scores {
		scores[i] = math.Exp(scores[i] - maxScore)
		sumExp += scores[i]
	}
	for i := range scores {
		scores[i] /= sumExp
	}

	// Weighted sum of values
	output := make([]float64, l.HeadDim)
	for i, neighbor := range neighbors {
		for d := 0; d < l.HeadDim; d++ {
			for j := 0; j < l.InputDim && j < len(features[neighbor]); j++ {
				output[d] += scores[i] * l.ValueWeights[head][d*l.InputDim+j] * features[neighbor][j]
			}
		}
	}
	// Add self
	for d := 0; d < l.HeadDim; d++ {
		for j := 0; j < l.InputDim && j < len(features[node]); j++ {
			output[d] += scores[len(neighbors)] * l.ValueWeights[head][d*l.InputDim+j] * features[node][j]
		}
	}

	return output
}

// =============================================================================
// CIM-Optimized GNN
// =============================================================================

// CIMGNNConfig configures CIM-aware GNN.
type CIMGNNConfig struct {
	GNNConfig
	CrossbarSize    int     // max crossbar dimension
	WeightBits      int     // weight precision
	ActivationBits  int     // activation precision
	SparsityThresh  float64 // threshold for sparse computation
	TileSize        int     // tiling for large graphs
}

// DefaultCIMGNNConfig returns standard CIM-GNN settings.
func DefaultCIMGNNConfig() CIMGNNConfig {
	return CIMGNNConfig{
		GNNConfig:      DefaultGNNConfig(),
		CrossbarSize:   256,
		WeightBits:     6,
		ActivationBits: 8,
		SparsityThresh: 0.1,
		TileSize:       64,
	}
}

// CIMGraphConvLayer implements CIM-optimized graph convolution.
type CIMGraphConvLayer struct {
	Config       CIMGNNConfig
	InputDim     int
	OutputDim    int
	Weights      [][]float64
	QuantWeights [][]int8
	Scale        float64
}

// NewCIMGraphConvLayer creates a CIM-optimized GCN layer.
func NewCIMGraphConvLayer(inputDim, outputDim int, config CIMGNNConfig, seed int64) *CIMGraphConvLayer {
	rng := rand.New(rand.NewSource(seed))
	layer := &CIMGraphConvLayer{
		Config:    config,
		InputDim:  inputDim,
		OutputDim: outputDim,
		Weights:   make([][]float64, outputDim),
	}

	scale := math.Sqrt(2.0 / float64(inputDim+outputDim))
	for i := 0; i < outputDim; i++ {
		layer.Weights[i] = make([]float64, inputDim)
		for j := 0; j < inputDim; j++ {
			layer.Weights[i][j] = rng.NormFloat64() * scale
		}
	}

	// Quantize weights
	layer.QuantizeWeights()

	return layer
}

// QuantizeWeights quantizes weights for CIM.
func (l *CIMGraphConvLayer) QuantizeWeights() {
	bits := l.Config.WeightBits
	levels := 1 << bits
	halfLevels := levels / 2

	// Find scale
	maxAbs := 0.0
	for i := range l.Weights {
		for j := range l.Weights[i] {
			if math.Abs(l.Weights[i][j]) > maxAbs {
				maxAbs = math.Abs(l.Weights[i][j])
			}
		}
	}
	l.Scale = maxAbs / float64(halfLevels-1)

	// Quantize
	l.QuantWeights = make([][]int8, len(l.Weights))
	for i := range l.Weights {
		l.QuantWeights[i] = make([]int8, len(l.Weights[i]))
		for j := range l.Weights[i] {
			q := int(math.Round(l.Weights[i][j] / l.Scale))
			if q > halfLevels-1 {
				q = halfLevels - 1
			} else if q < -halfLevels {
				q = -halfLevels
			}
			l.QuantWeights[i][j] = int8(q)
		}
	}
}

// ForwardTiled performs tiled graph convolution for large graphs.
func (l *CIMGraphConvLayer) ForwardTiled(graph *Graph, nodeFeatures [][]float64) [][]float64 {
	numNodes := len(nodeFeatures)
	output := make([][]float64, numNodes)

	tileSize := l.Config.TileSize

	// Process nodes in tiles
	for startNode := 0; startNode < numNodes; startNode += tileSize {
		endNode := startNode + tileSize
		if endNode > numNodes {
			endNode = numNodes
		}

		// Process tile
		for node := startNode; node < endNode; node++ {
			aggregated := l.aggregateWithSparsity(graph, nodeFeatures, node)

			// CIM matrix-vector multiply (simulated)
			output[node] = make([]float64, l.OutputDim)
			for i := 0; i < l.OutputDim; i++ {
				for j := 0; j < l.InputDim; j++ {
					// Use quantized weights
					w := float64(l.QuantWeights[i][j]) * l.Scale
					output[node][i] += w * aggregated[j]
				}
				// ReLU
				if output[node][i] < 0 {
					output[node][i] = 0
				}
			}
		}
	}

	return output
}

// aggregateWithSparsity aggregates using sparse operations.
func (l *CIMGraphConvLayer) aggregateWithSparsity(graph *Graph, features [][]float64, node int) []float64 {
	neighbors := graph.GetNeighbors(node)
	aggregated := make([]float64, l.InputDim)

	// Check sparsity
	sparsity := float64(len(neighbors)) / float64(graph.NumNodes)
	if sparsity < l.Config.SparsityThresh {
		// Sparse aggregation
		for _, neighbor := range neighbors {
			for j := 0; j < l.InputDim && j < len(features[neighbor]); j++ {
				aggregated[j] += features[neighbor][j]
			}
		}
	} else {
		// Dense aggregation
		for _, neighbor := range neighbors {
			for j := 0; j < l.InputDim && j < len(features[neighbor]); j++ {
				aggregated[j] += features[neighbor][j]
			}
		}
	}

	// Add self and normalize
	for j := 0; j < l.InputDim && j < len(features[node]); j++ {
		aggregated[j] += features[node][j]
	}
	norm := float64(len(neighbors) + 1)
	for j := range aggregated {
		aggregated[j] /= norm
	}

	return aggregated
}

// =============================================================================
// Serialization
// =============================================================================

// SNNState captures SNN state for persistence.
type SNNState struct {
	Config      SNNConfig       `json:"config"`
	LayerSizes  []int           `json:"layer_sizes"`
	Weights     [][][]float64   `json:"weights"`
	STDPConfig  STDPConfig      `json:"stdp_config"`
}

// ExportSNNState exports SNN state.
func (snn *SpikingNetwork) ExportSNNState() ([]byte, error) {
	state := SNNState{
		Config:     snn.Config,
		Weights:    make([][][]float64, len(snn.Layers)),
		STDPConfig: DefaultSTDPConfig(),
	}

	state.LayerSizes = make([]int, len(snn.Layers)+1)
	for i, layer := range snn.Layers {
		state.Weights[i] = layer.Weights
		state.LayerSizes[i] = layer.InputSize
	}
	if len(snn.Layers) > 0 {
		state.LayerSizes[len(snn.Layers)] = snn.Layers[len(snn.Layers)-1].Size
	}

	return json.MarshalIndent(state, "", "  ")
}

// GNNState captures GNN state for persistence.
type GNNState struct {
	Config     GNNConfig       `json:"config"`
	LayerDims  []int           `json:"layer_dims"`
	Weights    [][][]float64   `json:"weights"`
}

// =============================================================================
// Benchmarking
// =============================================================================

// SNNBenchmark evaluates SNN performance.
type SNNBenchmark struct {
	Model           string
	Dataset         string
	Accuracy        float64
	EnergyPerSpike  float64 // fJ
	TotalSpikes     int
	AverageLatency  float64 // ms
	FiringRate      float64
}

// GNNBenchmark evaluates GNN performance.
type GNNBenchmark struct {
	Model          string
	Dataset        string
	NumNodes       int
	NumEdges       int
	Accuracy       float64
	Throughput     float64 // nodes/sec
	EnergyPerNode  float64 // µJ
	Sparsity       float64
}

// RunSNNBenchmark evaluates SNN on test data.
func RunSNNBenchmark(snn *SpikingNetwork, testData [][]float64, testLabels []int) SNNBenchmark {
	bench := SNNBenchmark{Model: "SNN"}

	correct := 0
	totalSpikes := 0

	for i, input := range testData {
		output := snn.Forward(input)

		// Count spikes
		for _, layer := range snn.Layers {
			for _, neuron := range layer.Neurons {
				for _, spike := range neuron.SpikeHistory {
					if spike {
						totalSpikes++
					}
				}
			}
		}

		// Get prediction
		predicted := 0
		maxVal := output[0]
		for j, v := range output[1:] {
			if v > maxVal {
				maxVal = v
				predicted = j + 1
			}
		}

		if i < len(testLabels) && predicted == testLabels[i] {
			correct++
		}
	}

	bench.Accuracy = float64(correct) / float64(len(testData))
	bench.TotalSpikes = totalSpikes
	bench.EnergyPerSpike = 1.93 // pJ (from literature)

	return bench
}

// RunGNNBenchmark evaluates GNN on graph data.
func RunGNNBenchmark(graph *Graph, labels []int, layer *GraphConvLayer) GNNBenchmark {
	bench := GNNBenchmark{
		Model:    "GCN",
		NumNodes: graph.NumNodes,
		NumEdges: graph.NumEdges,
	}

	// Compute sparsity
	maxEdges := graph.NumNodes * graph.NumNodes
	bench.Sparsity = 1 - float64(graph.NumEdges)/float64(maxEdges)

	// Forward pass
	output := layer.Forward(graph, graph.NodeFeatures)

	// Compute accuracy
	correct := 0
	for i, out := range output {
		predicted := 0
		maxVal := out[0]
		for j, v := range out[1:] {
			if v > maxVal {
				maxVal = v
				predicted = j + 1
			}
		}
		if i < len(labels) && predicted == labels[i] {
			correct++
		}
	}
	bench.Accuracy = float64(correct) / float64(len(output))

	return bench
}

// =============================================================================
// Utility Functions
// =============================================================================

// PrintSNNStats displays SNN statistics.
func PrintSNNStats(snn *SpikingNetwork) string {
	var result string
	result += fmt.Sprintf("SNN with %d layers\n", len(snn.Layers))

	for i, layer := range snn.Layers {
		rates := layer.GetSpikeRates()
		avgRate := 0.0
		for _, r := range rates {
			avgRate += r
		}
		avgRate /= float64(len(rates))
		result += fmt.Sprintf("  Layer %d: %d neurons, avg rate=%.3f\n",
			i, layer.Size, avgRate)
	}

	return result
}

// PrintGraphStats displays graph statistics.
func PrintGraphStats(graph *Graph) string {
	// Compute degree distribution
	degrees := make([]int, graph.NumNodes)
	for _, edge := range graph.EdgeIndex {
		degrees[edge[0]]++
	}

	sort.Ints(degrees)
	avgDegree := 0.0
	for _, d := range degrees {
		avgDegree += float64(d)
	}
	avgDegree /= float64(len(degrees))

	return fmt.Sprintf("Graph: %d nodes, %d edges, avg degree=%.2f, sparsity=%.4f",
		graph.NumNodes, graph.NumEdges, avgDegree,
		1-float64(graph.NumEdges)/float64(graph.NumNodes*graph.NumNodes))
}

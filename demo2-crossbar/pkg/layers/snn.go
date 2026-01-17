// Package layers provides neural network layer implementations for crossbar-based CIM.
// snn.go implements spiking neural network components for neuromorphic CIM.
//
// SNN Components:
// - Leaky Integrate-and-Fire (LIF) neurons
// - Spike-Timing Dependent Plasticity (STDP) learning
// - Spike encoding/decoding
// - Event-driven processing
//
// FeFET SNN Mapping:
// - FeFET crossbar for synaptic weights
// - DG-MPB-TFT for LIF neurons (no capacitors needed)
// - Pulse timing for STDP
//
// References:
// - Advanced Science 2024: All-Ferroelectric SNN (94.9% accuracy)
// - ACS Nano 2024: Ambipolar WSe₂ FeFET SNN
// - Frontiers 2020: Supervised Learning in FeFET SNN

package layers

import (
	"math"
	"math/rand"
)

// ============================================================================
// Spike Encoding
// ============================================================================

// SpikeEncoder encodes continuous values to spike trains
type SpikeEncoder struct {
	Method     SpikeEncodingMethod
	NumSteps   int     // Number of time steps
	MaxRate    float64 // Maximum firing rate (Hz)
	Threshold  float64 // For threshold encoding
	TimeWindow float64 // Time window (ms)
}

// SpikeEncodingMethod defines spike encoding type
type SpikeEncodingMethod int

const (
	// EncodingRate encodes value as firing rate
	EncodingRate SpikeEncodingMethod = iota
	// EncodingTemporal encodes value as spike timing
	EncodingTemporal
	// EncodingPhase encodes value as phase offset
	EncodingPhase
	// EncodingPoisson generates Poisson spike train
	EncodingPoisson
	// EncodingDelta single spike per non-zero input
	EncodingDelta
)

// DefaultSpikeEncoder returns default encoder
func DefaultSpikeEncoder() *SpikeEncoder {
	return &SpikeEncoder{
		Method:     EncodingRate,
		NumSteps:   100,
		MaxRate:    1000.0, // 1 kHz max
		Threshold:  0.5,
		TimeWindow: 100.0, // 100 ms
	}
}

// Encode converts input values to spike trains
func (e *SpikeEncoder) Encode(inputs []float64) [][]bool {
	// Output: [time_step][neuron] -> spike (bool)
	spikes := make([][]bool, e.NumSteps)
	for t := 0; t < e.NumSteps; t++ {
		spikes[t] = make([]bool, len(inputs))
	}

	switch e.Method {
	case EncodingRate:
		e.encodeRate(inputs, spikes)
	case EncodingTemporal:
		e.encodeTemporal(inputs, spikes)
	case EncodingPoisson:
		e.encodePoisson(inputs, spikes)
	case EncodingDelta:
		e.encodeDelta(inputs, spikes)
	default:
		e.encodeRate(inputs, spikes)
	}

	return spikes
}

func (e *SpikeEncoder) encodeRate(inputs []float64, spikes [][]bool) {
	for i, val := range inputs {
		// Clamp input to [0, 1]
		if val < 0 {
			val = 0
		}
		if val > 1 {
			val = 1
		}

		// Calculate expected spike count
		rate := val * e.MaxRate * (e.TimeWindow / 1000.0)
		expectedSpikes := int(rate)

		// Distribute spikes evenly
		if expectedSpikes > 0 {
			interval := e.NumSteps / expectedSpikes
			if interval < 1 {
				interval = 1
			}
			for t := 0; t < e.NumSteps && t/interval < expectedSpikes; t += interval {
				spikes[t][i] = true
			}
		}
	}
}

func (e *SpikeEncoder) encodeTemporal(inputs []float64, spikes [][]bool) {
	// Earlier spike = higher value
	for i, val := range inputs {
		if val < 0 {
			val = 0
		}
		if val > 1 {
			val = 1
		}

		// Spike time inversely proportional to value
		spikeTime := int(float64(e.NumSteps) * (1.0 - val))
		if spikeTime >= e.NumSteps {
			spikeTime = e.NumSteps - 1
		}
		if spikeTime < 0 {
			spikeTime = 0
		}
		if val > 0.01 { // Only spike for non-zero values
			spikes[spikeTime][i] = true
		}
	}
}

func (e *SpikeEncoder) encodePoisson(inputs []float64, spikes [][]bool) {
	dt := e.TimeWindow / float64(e.NumSteps) // Time per step (ms)

	for i, val := range inputs {
		if val < 0 {
			val = 0
		}
		if val > 1 {
			val = 1
		}

		// Firing probability per step
		rate := val * e.MaxRate // Hz
		prob := rate * dt / 1000.0

		for t := 0; t < e.NumSteps; t++ {
			if rand.Float64() < prob {
				spikes[t][i] = true
			}
		}
	}
}

func (e *SpikeEncoder) encodeDelta(inputs []float64, spikes [][]bool) {
	// Single spike at t=0 for non-zero inputs
	for i, val := range inputs {
		if val > e.Threshold {
			spikes[0][i] = true
		}
	}
}

// ============================================================================
// Spike Decoding
// ============================================================================

// SpikeDecoder decodes spike trains to continuous values
type SpikeDecoder struct {
	Method     SpikeDecodingMethod
	TimeWindow float64
}

// SpikeDecodingMethod defines decoding type
type SpikeDecodingMethod int

const (
	// DecodingCount counts total spikes
	DecodingCount SpikeDecodingMethod = iota
	// DecodingFirstSpike uses first spike time
	DecodingFirstSpike
	// DecodingLastSpike uses last spike time
	DecodingLastSpike
	// DecodingMembrane uses final membrane potential
	DecodingMembrane
)

// DefaultSpikeDecoder returns default decoder
func DefaultSpikeDecoder() *SpikeDecoder {
	return &SpikeDecoder{
		Method:     DecodingCount,
		TimeWindow: 100.0,
	}
}

// Decode converts spike trains to values
func (d *SpikeDecoder) Decode(spikes [][]bool) []float64 {
	if len(spikes) == 0 {
		return nil
	}

	numNeurons := len(spikes[0])
	output := make([]float64, numNeurons)

	switch d.Method {
	case DecodingCount:
		for _, timestep := range spikes {
			for i, spike := range timestep {
				if spike {
					output[i]++
				}
			}
		}
		// Normalize by time steps
		maxCount := float64(len(spikes))
		for i := range output {
			output[i] /= maxCount
		}

	case DecodingFirstSpike:
		firstSpike := make([]int, numNeurons)
		for i := range firstSpike {
			firstSpike[i] = -1
		}
		for t, timestep := range spikes {
			for i, spike := range timestep {
				if spike && firstSpike[i] < 0 {
					firstSpike[i] = t
				}
			}
		}
		for i, t := range firstSpike {
			if t >= 0 {
				output[i] = 1.0 - float64(t)/float64(len(spikes))
			}
		}
	}

	return output
}

// ============================================================================
// LIF Neuron
// ============================================================================

// LIFNeuron implements Leaky Integrate-and-Fire neuron
type LIFNeuron struct {
	// Parameters
	Tau       float64 // Membrane time constant (ms)
	Vrest     float64 // Resting potential (mV)
	Vthresh   float64 // Spike threshold (mV)
	Vreset    float64 // Reset potential after spike (mV)
	Refract   float64 // Refractory period (ms)
	Dt        float64 // Time step (ms)

	// State
	V         float64 // Current membrane potential
	Spikes    []bool  // Spike history
	LastSpike int     // Time of last spike
}

// DefaultLIFNeuron returns default LIF parameters
func DefaultLIFNeuron() *LIFNeuron {
	return &LIFNeuron{
		Tau:       20.0, // 20 ms
		Vrest:     -65.0,
		Vthresh:   -50.0,
		Vreset:    -70.0,
		Refract:   2.0, // 2 ms
		Dt:        1.0, // 1 ms
		V:         -65.0,
		LastSpike: -1000,
	}
}

// Step advances neuron by one time step
func (n *LIFNeuron) Step(current float64, t int) bool {
	// Check refractory period
	if float64(t-n.LastSpike)*n.Dt < n.Refract {
		n.V = n.Vreset
		return false
	}

	// Leaky integration
	// dV/dt = -(V - Vrest)/tau + I
	dV := (-(n.V - n.Vrest) / n.Tau + current) * n.Dt
	n.V += dV

	// Check for spike
	if n.V >= n.Vthresh {
		n.LastSpike = t
		n.V = n.Vreset
		return true
	}

	return false
}

// Reset resets neuron state
func (n *LIFNeuron) Reset() {
	n.V = n.Vrest
	n.LastSpike = -1000
	n.Spikes = nil
}

// ============================================================================
// LIF Layer
// ============================================================================

// LIFLayer implements a layer of LIF neurons
type LIFLayer struct {
	NumNeurons int
	Neurons    []*LIFNeuron
	Weights    [][]float64 // Synaptic weights
	Biases     []float64   // Baseline currents
	Config     *LIFLayerConfig
}

// LIFLayerConfig configures LIF layer
type LIFLayerConfig struct {
	NumSteps   int
	Dt         float64
	Tau        float64
	Vthresh    float64
	Vreset     float64
	Vrest      float64
	Refract    float64
	WeightScale float64 // Scale factor for weights -> current
}

// DefaultLIFLayerConfig returns default LIF layer config
func DefaultLIFLayerConfig() *LIFLayerConfig {
	return &LIFLayerConfig{
		NumSteps:    100,
		Dt:          1.0,
		Tau:         20.0,
		Vthresh:     -50.0,
		Vreset:      -70.0,
		Vrest:       -65.0,
		Refract:     2.0,
		WeightScale: 100.0, // Scale weights to current
	}
}

// NewLIFLayer creates a new LIF layer
func NewLIFLayer(inputSize, numNeurons int, config *LIFLayerConfig) *LIFLayer {
	if config == nil {
		config = DefaultLIFLayerConfig()
	}

	layer := &LIFLayer{
		NumNeurons: numNeurons,
		Neurons:    make([]*LIFNeuron, numNeurons),
		Weights:    make([][]float64, numNeurons),
		Biases:     make([]float64, numNeurons),
		Config:     config,
	}

	// Initialize neurons
	for i := 0; i < numNeurons; i++ {
		layer.Neurons[i] = &LIFNeuron{
			Tau:       config.Tau,
			Vrest:     config.Vrest,
			Vthresh:   config.Vthresh,
			Vreset:    config.Vreset,
			Refract:   config.Refract,
			Dt:        config.Dt,
			V:         config.Vrest,
			LastSpike: -1000,
		}
		layer.Weights[i] = make([]float64, inputSize)
	}

	return layer
}

// Forward processes spike train through layer
func (l *LIFLayer) Forward(inputSpikes [][]bool) [][]bool {
	numSteps := len(inputSpikes)
	outputSpikes := make([][]bool, numSteps)

	// Reset neurons
	for _, n := range l.Neurons {
		n.Reset()
	}

	// Process each time step
	for t := 0; t < numSteps; t++ {
		outputSpikes[t] = make([]bool, l.NumNeurons)

		// Compute input currents for each neuron
		for i := 0; i < l.NumNeurons; i++ {
			current := l.Biases[i]

			// Sum weighted input spikes
			for j, spike := range inputSpikes[t] {
				if spike && j < len(l.Weights[i]) {
					current += l.Weights[i][j] * l.Config.WeightScale
				}
			}

			// Update neuron
			outputSpikes[t][i] = l.Neurons[i].Step(current, t)
		}
	}

	return outputSpikes
}

// SetWeights sets synaptic weights
func (l *LIFLayer) SetWeights(weights [][]float64) {
	l.Weights = weights
}

// ============================================================================
// STDP Learning
// ============================================================================

// STDPConfig configures STDP learning
type STDPConfig struct {
	TauPlus  float64 // LTP time constant (ms)
	TauMinus float64 // LTD time constant (ms)
	APlus    float64 // LTP amplitude
	AMinus   float64 // LTD amplitude
	Wmin     float64 // Minimum weight
	Wmax     float64 // Maximum weight
}

// DefaultSTDPConfig returns default STDP parameters
func DefaultSTDPConfig() *STDPConfig {
	return &STDPConfig{
		TauPlus:  20.0,
		TauMinus: 20.0,
		APlus:    0.01,
		AMinus:   0.01,
		Wmin:     0.0,
		Wmax:     1.0,
	}
}

// STDP implements spike-timing dependent plasticity
type STDP struct {
	Config *STDPConfig
}

// NewSTDP creates STDP learner
func NewSTDP(config *STDPConfig) *STDP {
	if config == nil {
		config = DefaultSTDPConfig()
	}
	return &STDP{Config: config}
}

// ComputeWeightChange computes STDP weight change
// dt = t_post - t_pre (positive = LTP, negative = LTD)
func (s *STDP) ComputeWeightChange(dt float64) float64 {
	if dt > 0 {
		// Post after pre -> LTP
		return s.Config.APlus * math.Exp(-dt/s.Config.TauPlus)
	} else if dt < 0 {
		// Pre after post -> LTD
		return -s.Config.AMinus * math.Exp(dt/s.Config.TauMinus)
	}
	return 0
}

// UpdateWeights updates weights based on pre/post spike times
func (s *STDP) UpdateWeights(weights [][]float64, preSpikes, postSpikes [][]bool) [][]float64 {
	if len(preSpikes) == 0 || len(postSpikes) == 0 {
		return weights
	}

	numPre := len(preSpikes[0])
	numPost := len(postSpikes[0])
	numSteps := len(preSpikes)

	// Find spike times
	preTimes := make([][]int, numPre)
	postTimes := make([][]int, numPost)

	for t := 0; t < numSteps; t++ {
		for i, spike := range preSpikes[t] {
			if spike {
				preTimes[i] = append(preTimes[i], t)
			}
		}
		for i, spike := range postSpikes[t] {
			if spike {
				postTimes[i] = append(postTimes[i], t)
			}
		}
	}

	// Update weights for each synapse
	for i := 0; i < numPost && i < len(weights); i++ {
		for j := 0; j < numPre && j < len(weights[i]); j++ {
			// All pairs of pre-post spikes
			for _, tPre := range preTimes[j] {
				for _, tPost := range postTimes[i] {
					dt := float64(tPost - tPre)
					dw := s.ComputeWeightChange(dt)
					weights[i][j] += dw

					// Bound weights
					if weights[i][j] < s.Config.Wmin {
						weights[i][j] = s.Config.Wmin
					}
					if weights[i][j] > s.Config.Wmax {
						weights[i][j] = s.Config.Wmax
					}
				}
			}
		}
	}

	return weights
}

// ============================================================================
// FeFET SNN Simulation
// ============================================================================

// FeFETSNNConfig configures FeFET-based SNN
type FeFETSNNConfig struct {
	// Device parameters
	ProgramVoltage float64 // Programming voltage (V)
	ReadVoltage    float64 // Read voltage (V)
	WriteEnergy    float64 // Energy per write (fJ)
	SpikeEnergy    float64 // Energy per spike (fJ)

	// Crossbar parameters
	ArraySize   int
	NoiseLevel  float64

	// Neuron parameters (DG-MPB-TFT)
	NeuronType  string // "MPB-TFT", "standard"
}

// DefaultFeFETSNNConfig returns default FeFET SNN config
func DefaultFeFETSNNConfig() *FeFETSNNConfig {
	return &FeFETSNNConfig{
		ProgramVoltage: 4.0,
		ReadVoltage:    1.0,
		WriteEnergy:    100.0,  // 100 fJ/write
		SpikeEnergy:    2.0,    // 2 fJ/spike (from Advanced Science 2024)
		ArraySize:      64,
		NoiseLevel:     0.02,
		NeuronType:     "MPB-TFT",
	}
}

// FeFETSNN implements FeFET-based spiking neural network
type FeFETSNN struct {
	Config  *FeFETSNNConfig
	Layers  []*LIFLayer
	Encoder *SpikeEncoder
	Decoder *SpikeDecoder
	STDP    *STDP
	Stats   *SNNStats
}

// SNNStats tracks SNN statistics
type SNNStats struct {
	TotalSpikes     int64
	TotalEnergy     float64 // fJ
	SpikeRate       float64 // Average firing rate
	InferenceCount  int
	ClassAccuracy   float64
}

// NewFeFETSNN creates FeFET-based SNN
func NewFeFETSNN(layerSizes []int, config *FeFETSNNConfig) *FeFETSNN {
	if config == nil {
		config = DefaultFeFETSNNConfig()
	}

	snn := &FeFETSNN{
		Config:  config,
		Encoder: DefaultSpikeEncoder(),
		Decoder: DefaultSpikeDecoder(),
		STDP:    NewSTDP(nil),
		Stats:   &SNNStats{},
	}

	// Create layers
	for i := 1; i < len(layerSizes); i++ {
		layer := NewLIFLayer(layerSizes[i-1], layerSizes[i], nil)
		snn.Layers = append(snn.Layers, layer)
	}

	return snn
}

// Forward performs inference
func (s *FeFETSNN) Forward(input []float64) []float64 {
	// Encode input to spikes
	spikes := s.Encoder.Encode(input)

	// Process through layers
	for _, layer := range s.Layers {
		spikes = layer.Forward(spikes)

		// Track statistics
		for _, timestep := range spikes {
			for _, spike := range timestep {
				if spike {
					s.Stats.TotalSpikes++
					s.Stats.TotalEnergy += s.Config.SpikeEnergy
				}
			}
		}
	}

	// Decode output
	output := s.Decoder.Decode(spikes)
	s.Stats.InferenceCount++

	return output
}

// Predict returns predicted class
func (s *FeFETSNN) Predict(input []float64) int {
	output := s.Forward(input)

	// Argmax
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

// Train performs STDP training on sample
func (s *FeFETSNN) Train(input []float64, label int) {
	// Encode input
	inputSpikes := s.Encoder.Encode(input)

	// Forward through layers with STDP
	currentSpikes := inputSpikes
	for _, layer := range s.Layers {
		outputSpikes := layer.Forward(currentSpikes)

		// Apply STDP to weights
		layer.Weights = s.STDP.UpdateWeights(layer.Weights, currentSpikes, outputSpikes)

		currentSpikes = outputSpikes
	}
}

// Evaluate evaluates accuracy on dataset
func (s *FeFETSNN) Evaluate(inputs [][]float64, labels []int) float64 {
	correct := 0
	for i, input := range inputs {
		pred := s.Predict(input)
		if pred == labels[i] {
			correct++
		}
	}
	accuracy := float64(correct) / float64(len(labels))
	s.Stats.ClassAccuracy = accuracy
	return accuracy
}

// GetStats returns SNN statistics
func (s *FeFETSNN) GetStats() *SNNStats {
	if s.Stats.InferenceCount > 0 {
		totalNeurons := 0
		for _, layer := range s.Layers {
			totalNeurons += layer.NumNeurons
		}
		totalSteps := s.Encoder.NumSteps * s.Stats.InferenceCount
		if totalSteps > 0 {
			s.Stats.SpikeRate = float64(s.Stats.TotalSpikes) / float64(totalNeurons*totalSteps)
		}
	}
	return s.Stats
}

// ============================================================================
// Utility Functions
// ============================================================================

// CountSpikes counts total spikes in spike train
func CountSpikes(spikes [][]bool) int {
	count := 0
	for _, timestep := range spikes {
		for _, spike := range timestep {
			if spike {
				count++
			}
		}
	}
	return count
}

// ComputeSpikeRate computes average firing rate
func ComputeSpikeRate(spikes [][]bool, timeWindowMs float64) float64 {
	if len(spikes) == 0 || len(spikes[0]) == 0 {
		return 0
	}

	totalSpikes := float64(CountSpikes(spikes))
	numNeurons := float64(len(spikes[0]))

	// Rate in Hz
	return totalSpikes / numNeurons / (timeWindowMs / 1000.0)
}

// ConvertToEventDriven converts spike train to event list
type SpikeEvent struct {
	Time   int
	Neuron int
}

func ConvertToEvents(spikes [][]bool) []SpikeEvent {
	events := make([]SpikeEvent, 0)
	for t, timestep := range spikes {
		for n, spike := range timestep {
			if spike {
				events = append(events, SpikeEvent{Time: t, Neuron: n})
			}
		}
	}
	return events
}

// ConvertFromEvents converts event list to spike train
func ConvertFromEvents(events []SpikeEvent, numSteps, numNeurons int) [][]bool {
	spikes := make([][]bool, numSteps)
	for t := 0; t < numSteps; t++ {
		spikes[t] = make([]bool, numNeurons)
	}
	for _, e := range events {
		if e.Time < numSteps && e.Neuron < numNeurons {
			spikes[e.Time][e.Neuron] = true
		}
	}
	return spikes
}

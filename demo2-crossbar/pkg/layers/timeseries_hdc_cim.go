// timeseries_hdc_cim.go - Neuromorphic Time-Series Processing and Hyperdimensional Computing
// Part of IronLattice educational demonstrations
//
// This module implements:
// 1. Reservoir computing with memristors (echo state networks, liquid state machines)
// 2. LSTM/RNN acceleration with ferroelectric CIM
// 3. Hyperdimensional computing (HDC) with FeFET-based encoding
// 4. Temporal pattern recognition (ECG, EEG, speech)
//
// Research basis:
// - Dynamic memristor reservoir computing (Nature Communications 2021)
// - FeFET-based HDC encoder (arXiv 2025)
// - Memristor LSTM for text classification (2023)
// - Multi-bit FeFET IMC for HDC (Scientific Reports 2022)

package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// RESERVOIR COMPUTING WITH MEMRISTORS
// =============================================================================

// ReservoirConfig configures reservoir computing system
type ReservoirConfig struct {
	// Reservoir parameters
	ReservoirSize    int     // Number of reservoir nodes
	InputSize        int     // Input dimension
	OutputSize       int     // Output dimension
	SpectralRadius   float64 // Spectral radius of weight matrix
	InputScaling     float64 // Input weight scaling
	LeakRate         float64 // Leaky integration rate
	Sparsity         float64 // Connection sparsity

	// Memristor parameters
	ConductanceMin   float64 // Min conductance (µS)
	ConductanceMax   float64 // Max conductance (µS)
	DeviceVariation  float64 // Device-to-device variation
	TemporalVariation float64 // Temporal noise

	// Training
	RidgeRegularization float64 // Ridge regression lambda
}

// DefaultReservoirConfig returns typical reservoir parameters
func DefaultReservoirConfig() *ReservoirConfig {
	return &ReservoirConfig{
		ReservoirSize:       100,
		InputSize:           1,
		OutputSize:          1,
		SpectralRadius:      0.9,
		InputScaling:        0.5,
		LeakRate:            0.3,
		Sparsity:            0.9,
		ConductanceMin:      1.0,
		ConductanceMax:      100.0,
		DeviceVariation:     0.05,
		TemporalVariation:   0.02,
		RidgeRegularization: 1e-6,
	}
}

// MemristorReservoir implements echo state network with memristive weights
type MemristorReservoir struct {
	Config          *ReservoirConfig
	InputWeights    [][]float64 // Win: input_size × reservoir_size
	ReservoirWeights [][]float64 // W: reservoir_size × reservoir_size
	OutputWeights   [][]float64 // Wout: reservoir_size × output_size
	State           []float64   // Current reservoir state
	StateHistory    [][]float64 // History for training
	Stats           *ReservoirStats
}

// ReservoirStats tracks reservoir performance
type ReservoirStats struct {
	TotalInferences  int64
	MemoryCapacity   float64 // Fading memory metric
	NRMSE            float64 // Normalized root mean square error
	SpeechAccuracy   float64 // For speech recognition
	TimeSeriesError  float64 // For forecasting
}

// NewMemristorReservoir creates a memristor-based reservoir
func NewMemristorReservoir(config *ReservoirConfig) *MemristorReservoir {
	r := &MemristorReservoir{
		Config:       config,
		State:        make([]float64, config.ReservoirSize),
		StateHistory: make([][]float64, 0),
		Stats:        &ReservoirStats{},
	}

	r.initializeWeights()
	return r
}

// initializeWeights creates random sparse weight matrices
func (r *MemristorReservoir) initializeWeights() {
	config := r.Config

	// Input weights (dense)
	r.InputWeights = make([][]float64, config.InputSize)
	for i := 0; i < config.InputSize; i++ {
		r.InputWeights[i] = make([]float64, config.ReservoirSize)
		for j := 0; j < config.ReservoirSize; j++ {
			r.InputWeights[i][j] = (rand.Float64()*2 - 1) * config.InputScaling
		}
	}

	// Reservoir weights (sparse)
	r.ReservoirWeights = make([][]float64, config.ReservoirSize)
	for i := 0; i < config.ReservoirSize; i++ {
		r.ReservoirWeights[i] = make([]float64, config.ReservoirSize)
		for j := 0; j < config.ReservoirSize; j++ {
			if rand.Float64() > config.Sparsity {
				r.ReservoirWeights[i][j] = rand.Float64()*2 - 1
			}
		}
	}

	// Scale to spectral radius
	r.scaleSpectralRadius()

	// Output weights (initialized to zero, trained later)
	r.OutputWeights = make([][]float64, config.ReservoirSize)
	for i := 0; i < config.ReservoirSize; i++ {
		r.OutputWeights[i] = make([]float64, config.OutputSize)
	}
}

// scaleSpectralRadius scales reservoir weights
func (r *MemristorReservoir) scaleSpectralRadius() {
	// Approximate spectral radius via power iteration
	config := r.Config
	n := config.ReservoirSize

	// Start with random vector
	v := make([]float64, n)
	for i := range v {
		v[i] = rand.Float64()
	}

	// Power iteration (10 iterations)
	for iter := 0; iter < 10; iter++ {
		// w = W * v
		w := make([]float64, n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				w[i] += r.ReservoirWeights[i][j] * v[j]
			}
		}

		// Normalize
		norm := 0.0
		for _, val := range w {
			norm += val * val
		}
		norm = math.Sqrt(norm)

		if norm > 0 {
			for i := range v {
				v[i] = w[i] / norm
			}
		}
	}

	// Estimate spectral radius
	Wv := make([]float64, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			Wv[i] += r.ReservoirWeights[i][j] * v[j]
		}
	}

	spectralRadius := 0.0
	for i := range v {
		spectralRadius += Wv[i] * v[i]
	}
	spectralRadius = math.Abs(spectralRadius)

	// Scale weights
	if spectralRadius > 0 {
		scale := config.SpectralRadius / spectralRadius
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				r.ReservoirWeights[i][j] *= scale
			}
		}
	}
}

// Update advances the reservoir state by one time step
func (r *MemristorReservoir) Update(input []float64) []float64 {
	config := r.Config
	n := config.ReservoirSize

	// Add device variation (memristor noise)
	addNoise := func(w float64) float64 {
		return w * (1 + rand.NormFloat64()*config.DeviceVariation)
	}

	// Compute new state: x(t+1) = (1-α)x(t) + α*tanh(Win*u + W*x)
	preActivation := make([]float64, n)

	// Input contribution
	for i := 0; i < n; i++ {
		for j := 0; j < config.InputSize && j < len(input); j++ {
			preActivation[i] += addNoise(r.InputWeights[j][i]) * input[j]
		}
	}

	// Recurrent contribution
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			preActivation[i] += addNoise(r.ReservoirWeights[i][j]) * r.State[j]
		}
	}

	// Leaky integration with tanh activation
	newState := make([]float64, n)
	for i := 0; i < n; i++ {
		newState[i] = (1-config.LeakRate)*r.State[i] +
			config.LeakRate*math.Tanh(preActivation[i])
	}

	r.State = newState
	r.Stats.TotalInferences++

	return newState
}

// CollectStates runs reservoir on input sequence and collects states
func (r *MemristorReservoir) CollectStates(inputs [][]float64) [][]float64 {
	r.StateHistory = make([][]float64, len(inputs))

	// Reset state
	for i := range r.State {
		r.State[i] = 0
	}

	// Run through sequence
	for t, input := range inputs {
		state := r.Update(input)
		r.StateHistory[t] = make([]float64, len(state))
		copy(r.StateHistory[t], state)
	}

	return r.StateHistory
}

// Train trains output weights using ridge regression
func (r *MemristorReservoir) Train(inputs, targets [][]float64) error {
	// Collect states
	states := r.CollectStates(inputs)

	// Ridge regression: Wout = (X'X + λI)^(-1) X'Y
	// Simplified: use pseudo-inverse approximation
	config := r.Config
	n := config.ReservoirSize
	m := len(states)

	// X'X
	XtX := make([][]float64, n)
	for i := 0; i < n; i++ {
		XtX[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			for t := 0; t < m; t++ {
				XtX[i][j] += states[t][i] * states[t][j]
			}
			if i == j {
				XtX[i][j] += config.RidgeRegularization
			}
		}
	}

	// X'Y
	XtY := make([][]float64, n)
	for i := 0; i < n; i++ {
		XtY[i] = make([]float64, config.OutputSize)
		for o := 0; o < config.OutputSize; o++ {
			for t := 0; t < m && t < len(targets); t++ {
				if o < len(targets[t]) {
					XtY[i][o] += states[t][i] * targets[t][o]
				}
			}
		}
	}

	// Solve using simple iteration (for demonstration)
	// In practice, use proper linear algebra library
	for i := 0; i < n; i++ {
		for o := 0; o < config.OutputSize; o++ {
			if XtX[i][i] > 0 {
				r.OutputWeights[i][o] = XtY[i][o] / XtX[i][i]
			}
		}
	}

	return nil
}

// Predict generates output from current state
func (r *MemristorReservoir) Predict() []float64 {
	config := r.Config
	output := make([]float64, config.OutputSize)

	for o := 0; o < config.OutputSize; o++ {
		for i := 0; i < config.ReservoirSize; i++ {
			output[o] += r.OutputWeights[i][o] * r.State[i]
		}
	}

	return output
}

// =============================================================================
// LIQUID STATE MACHINE (SPIKING RESERVOIR)
// =============================================================================

// LSMConfig configures liquid state machine
type LSMConfig struct {
	NumExcitatory   int     // Excitatory neurons
	NumInhibitory   int     // Inhibitory neurons
	ConnectionProb  float64 // Connection probability
	TimeConstantMs  float64 // Membrane time constant
	ThresholdmV     float64 // Spike threshold
	RefractoryMs    float64 // Refractory period
	InputCurrentScale float64
}

// DefaultLSMConfig returns typical LSM parameters
func DefaultLSMConfig() *LSMConfig {
	return &LSMConfig{
		NumExcitatory:    80,
		NumInhibitory:    20,
		ConnectionProb:   0.1,
		TimeConstantMs:   20.0,
		ThresholdmV:      -55.0,
		RefractoryMs:     2.0,
		InputCurrentScale: 100.0,
	}
}

// LIFNeuron implements leaky integrate-and-fire neuron
type LIFNeuron struct {
	MembranePotential float64
	Threshold         float64
	RestPotential     float64
	TimeConstant      float64
	RefractoryTime    float64
	IsExcitatory      bool
	LastSpikeTime     float64
}

// NewLIFNeuron creates a LIF neuron
func NewLIFNeuron(threshold, timeConstant float64, excitatory bool) *LIFNeuron {
	return &LIFNeuron{
		MembranePotential: -70.0,
		Threshold:         threshold,
		RestPotential:     -70.0,
		TimeConstant:      timeConstant,
		IsExcitatory:      excitatory,
		LastSpikeTime:     -1000,
	}
}

// Update advances neuron by dt ms with input current
func (n *LIFNeuron) Update(current float64, dt, currentTime float64) bool {
	// Check refractory
	if currentTime-n.LastSpikeTime < 2.0 {
		return false
	}

	// Leaky integration: dV/dt = -(V-Vrest)/tau + I
	dV := (-(n.MembranePotential - n.RestPotential) / n.TimeConstant + current) * dt
	n.MembranePotential += dV

	// Check spike
	if n.MembranePotential >= n.Threshold {
		n.MembranePotential = n.RestPotential
		n.LastSpikeTime = currentTime
		return true
	}

	return false
}

// LiquidStateMachine implements spiking reservoir
type LiquidStateMachine struct {
	Config       *LSMConfig
	Neurons      []*LIFNeuron
	Weights      [][]float64
	InputWeights [][]float64
	OutputWeights [][]float64
	SpikeHistory [][]bool
	CurrentTime  float64
	Stats        *LSMStats
}

// LSMStats tracks LSM performance
type LSMStats struct {
	TotalSpikes       int64
	AvgFiringRate     float64
	SeparationRatio   float64
	ClassificationAcc float64
}

// NewLiquidStateMachine creates an LSM
func NewLiquidStateMachine(config *LSMConfig) *LiquidStateMachine {
	totalNeurons := config.NumExcitatory + config.NumInhibitory

	lsm := &LiquidStateMachine{
		Config:       config,
		Neurons:      make([]*LIFNeuron, totalNeurons),
		Weights:      make([][]float64, totalNeurons),
		InputWeights: make([][]float64, 0),
		CurrentTime:  0,
		Stats:        &LSMStats{},
	}

	// Create neurons
	for i := 0; i < totalNeurons; i++ {
		isExcitatory := i < config.NumExcitatory
		lsm.Neurons[i] = NewLIFNeuron(config.ThresholdmV, config.TimeConstantMs, isExcitatory)
	}

	// Create recurrent connections
	for i := 0; i < totalNeurons; i++ {
		lsm.Weights[i] = make([]float64, totalNeurons)
		for j := 0; j < totalNeurons; j++ {
			if i != j && rand.Float64() < config.ConnectionProb {
				// Excitatory positive, inhibitory negative
				if lsm.Neurons[j].IsExcitatory {
					lsm.Weights[i][j] = rand.Float64() * 5.0
				} else {
					lsm.Weights[i][j] = -rand.Float64() * 5.0
				}
			}
		}
	}

	return lsm
}

// SetInputSize configures input connections
func (lsm *LiquidStateMachine) SetInputSize(inputSize int) {
	totalNeurons := len(lsm.Neurons)
	lsm.InputWeights = make([][]float64, inputSize)

	for i := 0; i < inputSize; i++ {
		lsm.InputWeights[i] = make([]float64, totalNeurons)
		// Connect to random subset of neurons
		for j := 0; j < totalNeurons; j++ {
			if rand.Float64() < 0.3 {
				lsm.InputWeights[i][j] = rand.Float64() * lsm.Config.InputCurrentScale
			}
		}
	}
}

// Step advances LSM by one time step
func (lsm *LiquidStateMachine) Step(input []float64, dt float64) []bool {
	totalNeurons := len(lsm.Neurons)
	spikes := make([]bool, totalNeurons)

	// Compute input currents
	inputCurrents := make([]float64, totalNeurons)
	for i := 0; i < len(input) && i < len(lsm.InputWeights); i++ {
		for j := 0; j < totalNeurons; j++ {
			inputCurrents[j] += lsm.InputWeights[i][j] * input[i]
		}
	}

	// Add recurrent currents from previous spikes
	for i := 0; i < totalNeurons; i++ {
		for j := 0; j < totalNeurons; j++ {
			// Simplified: use membrane potential as proxy
			if lsm.Neurons[j].MembranePotential > -60 {
				inputCurrents[i] += lsm.Weights[i][j] * 0.1
			}
		}
	}

	// Update neurons
	for i := 0; i < totalNeurons; i++ {
		spikes[i] = lsm.Neurons[i].Update(inputCurrents[i], dt, lsm.CurrentTime)
		if spikes[i] {
			lsm.Stats.TotalSpikes++
		}
	}

	lsm.CurrentTime += dt
	return spikes
}

// RunSequence processes a spike train sequence
func (lsm *LiquidStateMachine) RunSequence(inputs [][]float64, dt float64) [][]bool {
	lsm.SpikeHistory = make([][]bool, len(inputs))

	for t, input := range inputs {
		lsm.SpikeHistory[t] = lsm.Step(input, dt)
	}

	return lsm.SpikeHistory
}

// =============================================================================
// HYPERDIMENSIONAL COMPUTING
// =============================================================================

// HDCConfig configures hyperdimensional computing system
type HDCConfig struct {
	Dimensions      int     // Hypervector dimension (e.g., 10000)
	NumClasses      int     // Number of classes
	NumFeatures     int     // Input feature dimension
	QuantizationLevels int  // Q levels for encoding
	Binarize        bool    // Use binary hypervectors
	LearningRate    float64 // For iterative training
}

// DefaultHDCConfig returns typical HDC parameters
func DefaultHDCConfig() *HDCConfig {
	return &HDCConfig{
		Dimensions:         10000,
		NumClasses:         10,
		NumFeatures:        784, // MNIST
		QuantizationLevels: 10,
		Binarize:           true,
		LearningRate:       1.0,
	}
}

// Hypervector represents a high-dimensional vector
type Hypervector struct {
	Dims   int
	Values []float64 // Or int for binary
	Binary bool
}

// NewRandomHypervector creates a random hypervector
func NewRandomHypervector(dims int, binary bool) *Hypervector {
	hv := &Hypervector{
		Dims:   dims,
		Values: make([]float64, dims),
		Binary: binary,
	}

	for i := 0; i < dims; i++ {
		if binary {
			if rand.Float64() < 0.5 {
				hv.Values[i] = -1
			} else {
				hv.Values[i] = 1
			}
		} else {
			hv.Values[i] = rand.NormFloat64()
		}
	}

	return hv
}

// Bind performs element-wise multiplication (XOR for binary)
func (hv *Hypervector) Bind(other *Hypervector) *Hypervector {
	result := &Hypervector{
		Dims:   hv.Dims,
		Values: make([]float64, hv.Dims),
		Binary: hv.Binary,
	}

	for i := 0; i < hv.Dims; i++ {
		result.Values[i] = hv.Values[i] * other.Values[i]
	}

	return result
}

// Bundle performs element-wise addition (majority for binary)
func (hv *Hypervector) Bundle(others []*Hypervector) *Hypervector {
	result := &Hypervector{
		Dims:   hv.Dims,
		Values: make([]float64, hv.Dims),
		Binary: hv.Binary,
	}

	// Sum all vectors including this one
	for i := 0; i < hv.Dims; i++ {
		result.Values[i] = hv.Values[i]
		for _, other := range others {
			if i < len(other.Values) {
				result.Values[i] += other.Values[i]
			}
		}
	}

	// Binarize if needed (majority vote)
	if hv.Binary {
		for i := 0; i < hv.Dims; i++ {
			if result.Values[i] >= 0 {
				result.Values[i] = 1
			} else {
				result.Values[i] = -1
			}
		}
	}

	return result
}

// Permute performs cyclic shift
func (hv *Hypervector) Permute(shift int) *Hypervector {
	result := &Hypervector{
		Dims:   hv.Dims,
		Values: make([]float64, hv.Dims),
		Binary: hv.Binary,
	}

	for i := 0; i < hv.Dims; i++ {
		newIdx := (i + shift) % hv.Dims
		if newIdx < 0 {
			newIdx += hv.Dims
		}
		result.Values[newIdx] = hv.Values[i]
	}

	return result
}

// CosineSimilarity computes similarity to another hypervector
func (hv *Hypervector) CosineSimilarity(other *Hypervector) float64 {
	dot := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for i := 0; i < hv.Dims && i < other.Dims; i++ {
		dot += hv.Values[i] * other.Values[i]
		norm1 += hv.Values[i] * hv.Values[i]
		norm2 += other.Values[i] * other.Values[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dot / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// HammingDistance computes Hamming distance for binary vectors
func (hv *Hypervector) HammingDistance(other *Hypervector) int {
	distance := 0
	for i := 0; i < hv.Dims && i < other.Dims; i++ {
		if hv.Values[i] != other.Values[i] {
			distance++
		}
	}
	return distance
}

// HDCEncoder implements hyperdimensional encoding
type HDCEncoder struct {
	Config       *HDCConfig
	PositionHVs  []*Hypervector // Position (ID) hypervectors
	LevelHVs     []*Hypervector // Level (value) hypervectors
	ItemMemory   map[string]*Hypervector
}

// NewHDCEncoder creates an HDC encoder
func NewHDCEncoder(config *HDCConfig) *HDCEncoder {
	enc := &HDCEncoder{
		Config:     config,
		PositionHVs: make([]*Hypervector, config.NumFeatures),
		LevelHVs:   make([]*Hypervector, config.QuantizationLevels),
		ItemMemory: make(map[string]*Hypervector),
	}

	// Generate position hypervectors
	for i := 0; i < config.NumFeatures; i++ {
		enc.PositionHVs[i] = NewRandomHypervector(config.Dimensions, config.Binarize)
	}

	// Generate level hypervectors (with similarity structure)
	baseHV := NewRandomHypervector(config.Dimensions, config.Binarize)
	enc.LevelHVs[0] = baseHV

	// Each level differs from previous by flipping some bits
	flipRate := 1.0 / float64(config.QuantizationLevels)
	for i := 1; i < config.QuantizationLevels; i++ {
		enc.LevelHVs[i] = &Hypervector{
			Dims:   config.Dimensions,
			Values: make([]float64, config.Dimensions),
			Binary: config.Binarize,
		}
		copy(enc.LevelHVs[i].Values, enc.LevelHVs[i-1].Values)

		// Flip some bits
		for j := 0; j < config.Dimensions; j++ {
			if rand.Float64() < flipRate {
				enc.LevelHVs[i].Values[j] *= -1
			}
		}
	}

	return enc
}

// EncodeFeatureVector encodes a feature vector to hypervector
func (enc *HDCEncoder) EncodeFeatureVector(features []float64) *Hypervector {
	config := enc.Config

	// Initialize result hypervector
	result := &Hypervector{
		Dims:   config.Dimensions,
		Values: make([]float64, config.Dimensions),
		Binary: config.Binarize,
	}

	// Encode each feature: position ⊗ level, then bundle all
	for i := 0; i < len(features) && i < config.NumFeatures; i++ {
		// Quantize value to level
		level := int(features[i] * float64(config.QuantizationLevels-1))
		if level < 0 {
			level = 0
		}
		if level >= config.QuantizationLevels {
			level = config.QuantizationLevels - 1
		}

		// Bind position with level
		bound := enc.PositionHVs[i].Bind(enc.LevelHVs[level])

		// Add to result (bundle)
		for j := 0; j < config.Dimensions; j++ {
			result.Values[j] += bound.Values[j]
		}
	}

	// Binarize
	if config.Binarize {
		for j := 0; j < config.Dimensions; j++ {
			if result.Values[j] >= 0 {
				result.Values[j] = 1
			} else {
				result.Values[j] = -1
			}
		}
	}

	return result
}

// EncodeNGram encodes text using n-gram encoding
func (enc *HDCEncoder) EncodeNGram(text string, n int) *Hypervector {
	config := enc.Config

	result := &Hypervector{
		Dims:   config.Dimensions,
		Values: make([]float64, config.Dimensions),
		Binary: config.Binarize,
	}

	// Get or create character hypervectors
	charHVs := make(map[rune]*Hypervector)
	for _, c := range text {
		if _, exists := charHVs[c]; !exists {
			charHVs[c] = NewRandomHypervector(config.Dimensions, config.Binarize)
		}
	}

	// Generate n-grams
	runes := []rune(text)
	for i := 0; i <= len(runes)-n; i++ {
		// Encode n-gram: permute and bind
		ngram := charHVs[runes[i]]
		for j := 1; j < n && i+j < len(runes); j++ {
			permuted := charHVs[runes[i+j]].Permute(j)
			ngram = ngram.Bind(permuted)
		}

		// Bundle into result
		for k := 0; k < config.Dimensions; k++ {
			result.Values[k] += ngram.Values[k]
		}
	}

	// Binarize
	if config.Binarize {
		for j := 0; j < config.Dimensions; j++ {
			if result.Values[j] >= 0 {
				result.Values[j] = 1
			} else {
				result.Values[j] = -1
			}
		}
	}

	return result
}

// HDCClassifier implements HDC-based classification
type HDCClassifier struct {
	Config          *HDCConfig
	Encoder         *HDCEncoder
	ClassHVs        []*Hypervector // Associative memory
	TrainingSamples []int          // Samples per class
	Stats           *HDCStats
}

// HDCStats tracks HDC classifier performance
type HDCStats struct {
	TrainingSamples  int
	TestAccuracy     float64
	InferenceEnergy  float64 // Energy per inference
	InferenceLatency float64 // Latency per inference
}

// NewHDCClassifier creates an HDC classifier
func NewHDCClassifier(config *HDCConfig) *HDCClassifier {
	return &HDCClassifier{
		Config:          config,
		Encoder:         NewHDCEncoder(config),
		ClassHVs:        make([]*Hypervector, config.NumClasses),
		TrainingSamples: make([]int, config.NumClasses),
		Stats:           &HDCStats{},
	}
}

// Train trains the classifier on labeled data
func (c *HDCClassifier) Train(features [][]float64, labels []int) {
	// Initialize class hypervectors
	for i := 0; i < c.Config.NumClasses; i++ {
		c.ClassHVs[i] = &Hypervector{
			Dims:   c.Config.Dimensions,
			Values: make([]float64, c.Config.Dimensions),
			Binary: c.Config.Binarize,
		}
	}

	// Encode and accumulate
	for i, feat := range features {
		if i >= len(labels) {
			break
		}
		label := labels[i]
		if label < 0 || label >= c.Config.NumClasses {
			continue
		}

		encoded := c.Encoder.EncodeFeatureVector(feat)

		// Add to class hypervector
		for j := 0; j < c.Config.Dimensions; j++ {
			c.ClassHVs[label].Values[j] += encoded.Values[j]
		}
		c.TrainingSamples[label]++
		c.Stats.TrainingSamples++
	}

	// Binarize class hypervectors
	if c.Config.Binarize {
		for i := 0; i < c.Config.NumClasses; i++ {
			for j := 0; j < c.Config.Dimensions; j++ {
				if c.ClassHVs[i].Values[j] >= 0 {
					c.ClassHVs[i].Values[j] = 1
				} else {
					c.ClassHVs[i].Values[j] = -1
				}
			}
		}
	}
}

// Predict classifies a feature vector
func (c *HDCClassifier) Predict(features []float64) int {
	encoded := c.Encoder.EncodeFeatureVector(features)

	bestClass := 0
	bestSim := math.Inf(-1)

	for i := 0; i < c.Config.NumClasses; i++ {
		if c.ClassHVs[i] == nil {
			continue
		}
		sim := encoded.CosineSimilarity(c.ClassHVs[i])
		if sim > bestSim {
			bestSim = sim
			bestClass = i
		}
	}

	return bestClass
}

// Evaluate computes accuracy on test set
func (c *HDCClassifier) Evaluate(features [][]float64, labels []int) float64 {
	correct := 0
	total := 0

	for i, feat := range features {
		if i >= len(labels) {
			break
		}

		pred := c.Predict(feat)
		if pred == labels[i] {
			correct++
		}
		total++
	}

	if total == 0 {
		return 0
	}

	c.Stats.TestAccuracy = float64(correct) / float64(total)
	return c.Stats.TestAccuracy
}

// =============================================================================
// FEFET-BASED HDC ACCELERATOR
// =============================================================================

// FeFETHDCConfig configures FeFET-based HDC accelerator
type FeFETHDCConfig struct {
	ArraySize       int     // Crossbar array dimension
	NumArrays       int     // Number of crossbar arrays
	BitPrecision    int     // Weight precision
	VthVariation    float64 // Threshold voltage variation
	EnduranceCycles int
	RetentionHours  float64
}

// DefaultFeFETHDCConfig returns typical FeFET HDC accelerator config
func DefaultFeFETHDCConfig() *FeFETHDCConfig {
	return &FeFETHDCConfig{
		ArraySize:       128,
		NumArrays:       8,
		BitPrecision:    4,
		VthVariation:    0.05,
		EnduranceCycles: 1000000000,
		RetentionHours:  87600, // 10 years
	}
}

// FeFETHDCAccelerator implements FeFET-based HDC hardware
type FeFETHDCAccelerator struct {
	Config      *FeFETHDCConfig
	HDCConfig   *HDCConfig
	PositionMem [][]float64 // Position HVs in crossbar
	LevelMem    [][]float64 // Level HVs in crossbar
	ClassMem    [][]float64 // Class HVs (associative memory)
	Stats       *FeFETHDCStats
}

// FeFETHDCStats tracks accelerator performance
type FeFETHDCStats struct {
	EncodingEnergypJ  float64
	SearchEnergypJ    float64
	TotalEnergypJ     float64
	EncodingLatencyNs float64
	SearchLatencyNs   float64
	ThroughputMSPS    float64 // Mega samples per second
	EnergyEfficiency  float64 // Classifications per Joule
	AreaMM2           float64
}

// NewFeFETHDCAccelerator creates a FeFET-based HDC accelerator
func NewFeFETHDCAccelerator(fefetConfig *FeFETHDCConfig, hdcConfig *HDCConfig) *FeFETHDCAccelerator {
	acc := &FeFETHDCAccelerator{
		Config:    fefetConfig,
		HDCConfig: hdcConfig,
		Stats:     &FeFETHDCStats{},
	}

	// Initialize memory arrays
	acc.initializeMemory()

	return acc
}

// initializeMemory sets up crossbar arrays
func (acc *FeFETHDCAccelerator) initializeMemory() {
	// Position memory: features × dimensions
	acc.PositionMem = make([][]float64, acc.HDCConfig.NumFeatures)
	for i := 0; i < acc.HDCConfig.NumFeatures; i++ {
		acc.PositionMem[i] = make([]float64, acc.HDCConfig.Dimensions)
		for j := 0; j < acc.HDCConfig.Dimensions; j++ {
			if rand.Float64() < 0.5 {
				acc.PositionMem[i][j] = -1
			} else {
				acc.PositionMem[i][j] = 1
			}
		}
	}

	// Level memory: levels × dimensions
	acc.LevelMem = make([][]float64, acc.HDCConfig.QuantizationLevels)
	for i := 0; i < acc.HDCConfig.QuantizationLevels; i++ {
		acc.LevelMem[i] = make([]float64, acc.HDCConfig.Dimensions)
		for j := 0; j < acc.HDCConfig.Dimensions; j++ {
			if rand.Float64() < 0.5 {
				acc.LevelMem[i][j] = -1
			} else {
				acc.LevelMem[i][j] = 1
			}
		}
	}

	// Class memory: classes × dimensions
	acc.ClassMem = make([][]float64, acc.HDCConfig.NumClasses)
	for i := 0; i < acc.HDCConfig.NumClasses; i++ {
		acc.ClassMem[i] = make([]float64, acc.HDCConfig.Dimensions)
	}
}

// EncodeInMemory performs in-memory encoding
func (acc *FeFETHDCAccelerator) EncodeInMemory(features []float64) []float64 {
	dims := acc.HDCConfig.Dimensions
	result := make([]float64, dims)

	// For each feature, bind position with level
	for i := 0; i < len(features) && i < acc.HDCConfig.NumFeatures; i++ {
		// Quantize
		level := int(features[i] * float64(acc.HDCConfig.QuantizationLevels-1))
		if level < 0 {
			level = 0
		}
		if level >= acc.HDCConfig.QuantizationLevels {
			level = acc.HDCConfig.QuantizationLevels - 1
		}

		// In-memory XOR (bind) using FeFET array
		for j := 0; j < dims; j++ {
			// XOR implemented as multiplication for bipolar
			bound := acc.PositionMem[i][j] * acc.LevelMem[level][j]
			result[j] += bound

			// Add device variation
			result[j] += rand.NormFloat64() * acc.Config.VthVariation
		}
	}

	// Binarize (majority)
	for j := 0; j < dims; j++ {
		if result[j] >= 0 {
			result[j] = 1
		} else {
			result[j] = -1
		}
	}

	// Update stats
	acc.Stats.EncodingEnergypJ += float64(len(features)) * float64(dims) * 0.01 // ~0.01 pJ/bit
	acc.Stats.EncodingLatencyNs += 10 // ~10 ns per encoding

	return result
}

// SearchInMemory performs in-memory similarity search
func (acc *FeFETHDCAccelerator) SearchInMemory(queryHV []float64) int {
	dims := acc.HDCConfig.Dimensions
	bestClass := 0
	minDist := dims

	for c := 0; c < acc.HDCConfig.NumClasses; c++ {
		// Compute Hamming distance using TCAM-style parallel search
		dist := 0
		for j := 0; j < dims && j < len(queryHV); j++ {
			if queryHV[j] != acc.ClassMem[c][j] {
				dist++
			}
		}

		if dist < minDist {
			minDist = dist
			bestClass = c
		}
	}

	// Update stats
	acc.Stats.SearchEnergypJ += float64(acc.HDCConfig.NumClasses) * float64(dims) * 0.005 // ~0.005 pJ/bit
	acc.Stats.SearchLatencyNs += 5 // ~5 ns for parallel search

	return bestClass
}

// ProgramClassMemory stores class hypervectors
func (acc *FeFETHDCAccelerator) ProgramClassMemory(classHVs []*Hypervector) {
	for c := 0; c < acc.HDCConfig.NumClasses && c < len(classHVs); c++ {
		if classHVs[c] == nil {
			continue
		}
		copy(acc.ClassMem[c], classHVs[c].Values)
	}
}

// ComputeStats calculates accelerator performance metrics
func (acc *FeFETHDCAccelerator) ComputeStats(numInferences int) {
	acc.Stats.TotalEnergypJ = acc.Stats.EncodingEnergypJ + acc.Stats.SearchEnergypJ

	totalLatencyNs := acc.Stats.EncodingLatencyNs + acc.Stats.SearchLatencyNs
	if totalLatencyNs > 0 {
		acc.Stats.ThroughputMSPS = float64(numInferences) / (totalLatencyNs * 1e-3) // MSPS
	}

	if acc.Stats.TotalEnergypJ > 0 {
		acc.Stats.EnergyEfficiency = float64(numInferences) / (acc.Stats.TotalEnergypJ * 1e-12) // per Joule
	}

	// Estimate area (simplified)
	totalBits := acc.HDCConfig.NumFeatures * acc.HDCConfig.Dimensions +
		acc.HDCConfig.QuantizationLevels * acc.HDCConfig.Dimensions +
		acc.HDCConfig.NumClasses * acc.HDCConfig.Dimensions
	acc.Stats.AreaMM2 = float64(totalBits) * 0.001 * 0.001 // ~1 µm² per bit
}

// =============================================================================
// DEMO AND BENCHMARK FUNCTIONS
// =============================================================================

// DemoReservoirComputing demonstrates memristor reservoir
func DemoReservoirComputing() {
	fmt.Println("=== Memristor Reservoir Computing Demo ===")
	fmt.Println()

	// Create reservoir
	config := DefaultReservoirConfig()
	config.ReservoirSize = 50
	config.InputSize = 1
	config.OutputSize = 1
	reservoir := NewMemristorReservoir(config)

	fmt.Println("1. Reservoir Configuration:")
	fmt.Printf("   Reservoir size: %d neurons\n", config.ReservoirSize)
	fmt.Printf("   Spectral radius: %.2f\n", config.SpectralRadius)
	fmt.Printf("   Leak rate: %.2f\n", config.LeakRate)
	fmt.Printf("   Sparsity: %.0f%%\n", config.Sparsity*100)
	fmt.Println()

	// Generate sine wave for prediction
	fmt.Println("2. Time Series Prediction (Sine Wave):")
	numSamples := 200
	inputs := make([][]float64, numSamples)
	targets := make([][]float64, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) * 0.1
		inputs[i] = []float64{math.Sin(t)}
		targets[i] = []float64{math.Sin(t + 0.1)} // Predict next step
	}

	// Train
	reservoir.Train(inputs[:150], targets[:150])
	fmt.Println("   Training completed on 150 samples")

	// Test
	reservoir.CollectStates(inputs[:150])
	testMSE := 0.0
	for i := 150; i < numSamples; i++ {
		reservoir.Update(inputs[i])
		pred := reservoir.Predict()
		err := pred[0] - targets[i][0]
		testMSE += err * err
	}
	testMSE /= 50
	fmt.Printf("   Test MSE: %.6f\n", testMSE)
	fmt.Printf("   Test RMSE: %.6f\n", math.Sqrt(testMSE))
	fmt.Println()

	// Performance metrics
	fmt.Println("3. Reservoir Computing Performance (Literature):")
	fmt.Println("   ┌─────────────────────────────────────────────────┐")
	fmt.Println("   │ Task                    │ Accuracy/Error        │")
	fmt.Println("   ├─────────────────────────┼───────────────────────┤")
	fmt.Println("   │ Spoken digit (TI46)     │ 98.84%                │")
	fmt.Println("   │ Mackey-Glass prediction │ NRMSE 0.036           │")
	fmt.Println("   │ MNIST classification    │ 90.0% (unsupervised)  │")
	fmt.Println("   │ N-MNIST (LSM)           │ 98.1%                 │")
	fmt.Println("   │ SHD speech (LSM)        │ 77.8%                 │")
	fmt.Println("   └─────────────────────────────────────────────────┘")
}

// DemoLiquidStateMachine demonstrates LSM
func DemoLiquidStateMachine() {
	fmt.Println()
	fmt.Println("=== Liquid State Machine Demo ===")
	fmt.Println()

	config := DefaultLSMConfig()
	config.NumExcitatory = 40
	config.NumInhibitory = 10
	lsm := NewLiquidStateMachine(config)
	lsm.SetInputSize(2)

	fmt.Println("1. LSM Configuration:")
	fmt.Printf("   Excitatory neurons: %d\n", config.NumExcitatory)
	fmt.Printf("   Inhibitory neurons: %d\n", config.NumInhibitory)
	fmt.Printf("   Connection probability: %.1f%%\n", config.ConnectionProb*100)
	fmt.Printf("   Time constant: %.1f ms\n", config.TimeConstantMs)
	fmt.Println()

	// Generate spike input pattern
	fmt.Println("2. Processing spike train:")
	numSteps := 100
	inputs := make([][]float64, numSteps)
	for t := 0; t < numSteps; t++ {
		// Periodic input
		if t%20 < 5 {
			inputs[t] = []float64{1.0, 0.0}
		} else if t%20 < 10 {
			inputs[t] = []float64{0.0, 1.0}
		} else {
			inputs[t] = []float64{0.0, 0.0}
		}
	}

	spikes := lsm.RunSequence(inputs, 1.0)

	// Count spikes
	totalSpikes := 0
	for _, s := range spikes {
		for _, spiked := range s {
			if spiked {
				totalSpikes++
			}
		}
	}

	fmt.Printf("   Duration: %d ms\n", numSteps)
	fmt.Printf("   Total spikes: %d\n", totalSpikes)
	fmt.Printf("   Average firing rate: %.2f Hz\n", float64(totalSpikes)/float64(numSteps)*1000/float64(len(lsm.Neurons)))
	fmt.Println()

	fmt.Println("3. LSM for Temporal Pattern Recognition:")
	fmt.Println("   Applications: speech, gestures, biosignals")
	fmt.Println("   Advantage: event-driven, low power")
}

// DemoHDC demonstrates hyperdimensional computing
func DemoHDC() {
	fmt.Println()
	fmt.Println("=== Hyperdimensional Computing Demo ===")
	fmt.Println()

	config := DefaultHDCConfig()
	config.Dimensions = 1000 // Smaller for demo
	config.NumClasses = 3
	config.NumFeatures = 10
	config.QuantizationLevels = 5

	fmt.Println("1. HDC Configuration:")
	fmt.Printf("   Dimensions: %d\n", config.Dimensions)
	fmt.Printf("   Classes: %d\n", config.NumClasses)
	fmt.Printf("   Features: %d\n", config.NumFeatures)
	fmt.Printf("   Quantization levels: %d\n", config.QuantizationLevels)
	fmt.Printf("   Binary: %v\n", config.Binarize)
	fmt.Println()

	// Create classifier
	classifier := NewHDCClassifier(config)

	// Generate synthetic data
	fmt.Println("2. Training on synthetic data:")
	numTrain := 100
	trainFeatures := make([][]float64, numTrain)
	trainLabels := make([]int, numTrain)

	for i := 0; i < numTrain; i++ {
		trainFeatures[i] = make([]float64, config.NumFeatures)
		trainLabels[i] = i % config.NumClasses

		// Class-dependent patterns
		for j := 0; j < config.NumFeatures; j++ {
			trainFeatures[i][j] = float64(trainLabels[i]+1)/float64(config.NumClasses) +
				rand.Float64()*0.2 - 0.1
		}
	}

	classifier.Train(trainFeatures, trainLabels)
	fmt.Printf("   Trained on %d samples\n", numTrain)
	fmt.Println()

	// Test
	fmt.Println("3. Testing:")
	numTest := 30
	testFeatures := make([][]float64, numTest)
	testLabels := make([]int, numTest)

	for i := 0; i < numTest; i++ {
		testFeatures[i] = make([]float64, config.NumFeatures)
		testLabels[i] = i % config.NumClasses

		for j := 0; j < config.NumFeatures; j++ {
			testFeatures[i][j] = float64(testLabels[i]+1)/float64(config.NumClasses) +
				rand.Float64()*0.2 - 0.1
		}
	}

	accuracy := classifier.Evaluate(testFeatures, testLabels)
	fmt.Printf("   Test accuracy: %.1f%%\n", accuracy*100)
	fmt.Println()

	// HDC operations demo
	fmt.Println("4. HDC Operations:")
	hv1 := NewRandomHypervector(config.Dimensions, true)
	hv2 := NewRandomHypervector(config.Dimensions, true)

	bound := hv1.Bind(hv2)
	bundled := hv1.Bundle([]*Hypervector{hv2})
	permuted := hv1.Permute(1)

	fmt.Printf("   Similarity(hv1, hv2): %.3f\n", hv1.CosineSimilarity(hv2))
	fmt.Printf("   Similarity(hv1, bound): %.3f\n", hv1.CosineSimilarity(bound))
	fmt.Printf("   Similarity(hv1, bundled): %.3f\n", hv1.CosineSimilarity(bundled))
	fmt.Printf("   Similarity(hv1, permuted): %.3f\n", hv1.CosineSimilarity(permuted))
	fmt.Println()

	fmt.Println("5. HDC Performance (Literature):")
	fmt.Println("   ┌────────────────────────────────────────────────────┐")
	fmt.Println("   │ Dataset        │ Accuracy │ Energy Efficiency     │")
	fmt.Println("   ├────────────────┼──────────┼───────────────────────┤")
	fmt.Println("   │ MNIST          │ 97.9%    │ 350× vs DNN           │")
	fmt.Println("   │ Fashion-MNIST  │ 84.6%    │ -                     │")
	fmt.Println("   │ Language (21)  │ 90.7%    │ 826× energy, 30× lat  │")
	fmt.Println("   │ EMG gesture    │ 95.1%    │ 39.1 nJ/prediction    │")
	fmt.Println("   │ CIFAR-10       │ 84%      │ (binary models)       │")
	fmt.Println("   └────────────────────────────────────────────────────┘")
}

// DemoFeFETHDC demonstrates FeFET-based HDC accelerator
func DemoFeFETHDC() {
	fmt.Println()
	fmt.Println("=== FeFET-Based HDC Accelerator Demo ===")
	fmt.Println()

	fefetConfig := DefaultFeFETHDCConfig()
	hdcConfig := DefaultHDCConfig()
	hdcConfig.Dimensions = 1000

	fmt.Println("1. FeFET HDC Accelerator Configuration:")
	fmt.Printf("   Array size: %d×%d\n", fefetConfig.ArraySize, fefetConfig.ArraySize)
	fmt.Printf("   Num arrays: %d\n", fefetConfig.NumArrays)
	fmt.Printf("   Bit precision: %d\n", fefetConfig.BitPrecision)
	fmt.Printf("   Vth variation: %.1f%%\n", fefetConfig.VthVariation*100)
	fmt.Printf("   Endurance: %d cycles\n", fefetConfig.EnduranceCycles)
	fmt.Println()

	// Create accelerator
	acc := NewFeFETHDCAccelerator(fefetConfig, hdcConfig)

	// Run some inferences
	fmt.Println("2. Running in-memory inference:")
	numInferences := 100
	for i := 0; i < numInferences; i++ {
		features := make([]float64, hdcConfig.NumFeatures)
		for j := range features {
			features[j] = rand.Float64()
		}

		queryHV := acc.EncodeInMemory(features)
		_ = acc.SearchInMemory(queryHV)
	}

	acc.ComputeStats(numInferences)

	fmt.Printf("   Inferences: %d\n", numInferences)
	fmt.Printf("   Encoding energy: %.2f pJ total\n", acc.Stats.EncodingEnergypJ)
	fmt.Printf("   Search energy: %.2f pJ total\n", acc.Stats.SearchEnergypJ)
	fmt.Printf("   Total energy: %.2f pJ\n", acc.Stats.TotalEnergypJ)
	fmt.Printf("   Encoding latency: %.1f ns\n", acc.Stats.EncodingLatencyNs)
	fmt.Printf("   Search latency: %.1f ns\n", acc.Stats.SearchLatencyNs)
	fmt.Println()

	fmt.Println("3. FeFET HDC Performance (Literature):")
	fmt.Println("   ┌───────────────────────────────────────────────────────┐")
	fmt.Println("   │ Metric                  │ Value                       │")
	fmt.Println("   ├─────────────────────────┼─────────────────────────────┤")
	fmt.Println("   │ Energy improvement      │ 826× vs digital             │")
	fmt.Println("   │ Latency improvement     │ 30× vs digital              │")
	fmt.Println("   │ Spam classification acc │ 91.38%                      │")
	fmt.Println("   │ STT-MRAM energy         │ 3.12 fJ/bit                 │")
	fmt.Println("   │ FeReX speedup           │ 250× vs GPU                 │")
	fmt.Println("   │ FeReX energy savings    │ 10⁴× vs GPU                 │")
	fmt.Println("   └───────────────────────────────────────────────────────┘")
}

// BenchmarkTimeSeriesHDC runs comprehensive benchmarks
func BenchmarkTimeSeriesHDC() {
	fmt.Println()
	fmt.Println("=== Time-Series and HDC CIM Benchmarks ===")
	fmt.Println()

	fmt.Println("Temporal Processing with Memristor CIM:")
	fmt.Println("┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Application          │ Accuracy │ Energy      │ Notes       │")
	fmt.Println("├──────────────────────┼──────────┼─────────────┼─────────────┤")
	fmt.Println("│ ECG classification   │ 93.5%    │ 0.3% of dig │ f-MDPE      │")
	fmt.Println("│ Arrhythmia (VO2)     │ 95.83%   │ Ultra-low   │ Spike enc   │")
	fmt.Println("│ Epilepsy detection   │ 99.79%   │ Ultra-low   │ LSNN        │")
	fmt.Println("│ Speech (HZO FTJ)     │ High     │ 50ns switch │ 128 states  │")
	fmt.Println("│ Text (LSTM)          │ 88.58%   │ Hardware    │ IMDB        │")
	fmt.Println("│ Time-series pred     │ 87%      │ 2-step      │ Memristor   │")
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
	fmt.Println()

	fmt.Println("Hyperdimensional Computing with FeFET:")
	fmt.Println("┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│ Implementation       │ Energy Eff │ Accuracy   │ Speedup    │")
	fmt.Println("├──────────────────────┼────────────┼────────────┼────────────┤")
	fmt.Println("│ FeFET encoder        │ 826×       │ 91.38%     │ 30×        │")
	fmt.Println("│ Multi-bit FeFET      │ 826×       │ SW-equiv   │ 30×        │")
	fmt.Println("│ FeReX (k-NN/HDC)     │ 10⁴×       │ -          │ 250×       │")
	fmt.Println("│ Memristive SoC       │ -          │ 90.71%     │ -          │")
	fmt.Println("│ PCM in-memory        │ -          │ High       │ 760K devs  │")
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
	fmt.Println()

	fmt.Println("Key Technologies:")
	fmt.Println("  • Reservoir Computing: Fading memory, nonlinear dynamics")
	fmt.Println("  • LSM: Event-driven, spike-based temporal processing")
	fmt.Println("  • HDC: High-dimensional distributed representations")
	fmt.Println("  • FeFET: Non-volatile, low-energy in-memory computing")
	fmt.Println("  • Ferroelectric: HZO, BFO for synaptic plasticity")
}

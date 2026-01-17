// Package layers provides neural network layer implementations for CIM simulation.
// spiking_diffusion_ecc.go implements Spiking Diffusion Models and analog
// error correction/fault tolerance for CIM accelerators.
//
// Research basis:
// - Spiking Diffusion Models: TSM mechanism, ANN-SNN conversion
// - Bit slicing: Multi-bit precision via parallel computation
// - Matrix decomposition: M = M_A × M_B for fault tolerance
// - Compensation arrays: Iterative precision improvement
// - Differential readout: Robustness to device variability
//
// Key concepts:
// - Diffusion: x_t = √αt × x_0 + √(1-αt) × ε
// - Denoising: Predict noise ε from noisy sample
// - Spike coding: Temporal coding for diffusion timesteps
// - Fault tolerance: >99.999% cosine similarity at 39% fault rate
package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// SPIKING DIFFUSION MODELS
// =============================================================================

// SpikingDiffusionConfig configures spiking diffusion model
type SpikingDiffusionConfig struct {
	// Diffusion parameters
	NumTimesteps    int     // T for diffusion process
	BetaStart       float64 // Starting noise schedule
	BetaEnd         float64 // Ending noise schedule
	NoiseSchedule   string  // "linear", "cosine", "quadratic"

	// SNN parameters
	SpikingTimesteps int     // Number of SNN timesteps
	Threshold        float64 // Spike threshold
	LeakFactor       float64 // LIF neuron leak
	MembraneDecay    float64 // Membrane potential decay

	// Architecture
	ImageSize       int // Image dimension
	NumChannels     int // Image channels
	HiddenDim       int // Hidden dimension
	NumBlocks       int // U-Net blocks

	// CIM
	CrossbarSize int
}

// DefaultSpikingDiffusionConfig returns typical configuration
func DefaultSpikingDiffusionConfig() *SpikingDiffusionConfig {
	return &SpikingDiffusionConfig{
		NumTimesteps:     1000,
		BetaStart:        0.0001,
		BetaEnd:          0.02,
		NoiseSchedule:    "linear",
		SpikingTimesteps: 4,
		Threshold:        1.0,
		LeakFactor:       0.9,
		MembraneDecay:    0.5,
		ImageSize:        32,
		NumChannels:      3,
		HiddenDim:        64,
		NumBlocks:        4,
		CrossbarSize:     64,
	}
}

// NoiseScheduler computes noise schedules for diffusion
type NoiseScheduler struct {
	config     *SpikingDiffusionConfig
	betas      []float64 // Noise levels
	alphas     []float64 // 1 - beta
	alphasCum  []float64 // Cumulative product of alphas
	sqrtAlphas []float64 // sqrt(alpha_bar)
	sqrtOneMinusAlphas []float64 // sqrt(1 - alpha_bar)
}

// NewNoiseScheduler creates a noise scheduler
func NewNoiseScheduler(config *SpikingDiffusionConfig) *NoiseScheduler {
	if config == nil {
		config = DefaultSpikingDiffusionConfig()
	}

	ns := &NoiseScheduler{
		config:     config,
		betas:      make([]float64, config.NumTimesteps),
		alphas:     make([]float64, config.NumTimesteps),
		alphasCum:  make([]float64, config.NumTimesteps),
		sqrtAlphas: make([]float64, config.NumTimesteps),
		sqrtOneMinusAlphas: make([]float64, config.NumTimesteps),
	}

	ns.computeSchedule()
	return ns
}

// computeSchedule computes noise schedule
func (ns *NoiseScheduler) computeSchedule() {
	T := ns.config.NumTimesteps

	switch ns.config.NoiseSchedule {
	case "linear":
		for t := 0; t < T; t++ {
			ns.betas[t] = ns.config.BetaStart +
				(ns.config.BetaEnd-ns.config.BetaStart)*float64(t)/float64(T-1)
		}
	case "cosine":
		for t := 0; t < T; t++ {
			s := 0.008
			f := func(x float64) float64 {
				return math.Pow(math.Cos((x/float64(T)+s)/(1+s)*math.Pi/2), 2)
			}
			alphaBar := f(float64(t)) / f(0)
			if t == 0 {
				ns.betas[t] = 1 - alphaBar
			} else {
				prevAlphaBar := f(float64(t-1)) / f(0)
				ns.betas[t] = 1 - alphaBar/prevAlphaBar
			}
			ns.betas[t] = math.Min(ns.betas[t], 0.999)
		}
	case "quadratic":
		for t := 0; t < T; t++ {
			ratio := float64(t) / float64(T-1)
			ns.betas[t] = ns.config.BetaStart +
				(ns.config.BetaEnd-ns.config.BetaStart)*ratio*ratio
		}
	}

	// Compute derived values
	cumProd := 1.0
	for t := 0; t < T; t++ {
		ns.alphas[t] = 1 - ns.betas[t]
		cumProd *= ns.alphas[t]
		ns.alphasCum[t] = cumProd
		ns.sqrtAlphas[t] = math.Sqrt(ns.alphasCum[t])
		ns.sqrtOneMinusAlphas[t] = math.Sqrt(1 - ns.alphasCum[t])
	}
}

// AddNoise adds noise to image at timestep t
func (ns *NoiseScheduler) AddNoise(x []float64, noise []float64, t int) []float64 {
	if t < 0 || t >= ns.config.NumTimesteps {
		t = 0
	}

	result := make([]float64, len(x))
	sqrtAlpha := ns.sqrtAlphas[t]
	sqrtOneMinusAlpha := ns.sqrtOneMinusAlphas[t]

	for i := range x {
		result[i] = sqrtAlpha*x[i] + sqrtOneMinusAlpha*noise[i]
	}

	return result
}

// GetSNR returns signal-to-noise ratio at timestep t
func (ns *NoiseScheduler) GetSNR(t int) float64 {
	if t < 0 || t >= ns.config.NumTimesteps {
		return 0
	}
	return ns.alphasCum[t] / (1 - ns.alphasCum[t])
}

// SpikingNeuron implements LIF neuron for diffusion
type SpikingNeuron struct {
	config     *SpikingDiffusionConfig
	membrane   []float64 // Membrane potentials
	numNeurons int
}

// NewSpikingNeuron creates a spiking neuron layer
func NewSpikingNeuron(numNeurons int, config *SpikingDiffusionConfig) *SpikingNeuron {
	if config == nil {
		config = DefaultSpikingDiffusionConfig()
	}

	return &SpikingNeuron{
		config:     config,
		membrane:   make([]float64, numNeurons),
		numNeurons: numNeurons,
	}
}

// Forward processes input and returns spikes
func (sn *SpikingNeuron) Forward(input []float64) []float64 {
	spikes := make([]float64, sn.numNeurons)

	for i := 0; i < sn.numNeurons && i < len(input); i++ {
		// Leaky integrate
		sn.membrane[i] = sn.config.LeakFactor*sn.membrane[i] + input[i]

		// Fire and reset
		if sn.membrane[i] >= sn.config.Threshold {
			spikes[i] = 1.0
			sn.membrane[i] = 0
		}
	}

	return spikes
}

// Reset clears membrane potentials
func (sn *SpikingNeuron) Reset() {
	for i := range sn.membrane {
		sn.membrane[i] = 0
	}
}

// TemporalSpikingMechanism implements TSM from SDM paper
type TemporalSpikingMechanism struct {
	config         *SpikingDiffusionConfig
	weights        [][]float64 // Temporal attention weights
	numFeatures    int
}

// NewTemporalSpikingMechanism creates a TSM
func NewTemporalSpikingMechanism(numFeatures int, config *SpikingDiffusionConfig) *TemporalSpikingMechanism {
	if config == nil {
		config = DefaultSpikingDiffusionConfig()
	}

	rng := rand.New(rand.NewSource(42))
	tsm := &TemporalSpikingMechanism{
		config:      config,
		numFeatures: numFeatures,
		weights:     make([][]float64, config.SpikingTimesteps),
	}

	// Initialize temporal attention weights
	scale := math.Sqrt(1.0 / float64(config.SpikingTimesteps))
	for t := 0; t < config.SpikingTimesteps; t++ {
		tsm.weights[t] = make([]float64, numFeatures)
		for i := 0; i < numFeatures; i++ {
			tsm.weights[t][i] = rng.NormFloat64() * scale
		}
	}

	return tsm
}

// Forward processes spike trains with temporal attention
func (tsm *TemporalSpikingMechanism) Forward(spikeTrains [][]float64) []float64 {
	output := make([]float64, tsm.numFeatures)

	// Weighted sum across timesteps
	for t := 0; t < len(spikeTrains) && t < tsm.config.SpikingTimesteps; t++ {
		for i := 0; i < tsm.numFeatures && i < len(spikeTrains[t]); i++ {
			output[i] += tsm.weights[t][i] * spikeTrains[t][i]
		}
	}

	return output
}

// SpikingDiffusionModel implements SNN-based diffusion
type SpikingDiffusionModel struct {
	config    *SpikingDiffusionConfig
	scheduler *NoiseScheduler
	neurons   []*SpikingNeuron
	tsm       *TemporalSpikingMechanism
	unetWeights [][]float64
	rng       *rand.Rand

	// Statistics
	totalSpikes  int64
	totalOps     int64
}

// NewSpikingDiffusionModel creates a spiking diffusion model
func NewSpikingDiffusionModel(config *SpikingDiffusionConfig) *SpikingDiffusionModel {
	if config == nil {
		config = DefaultSpikingDiffusionConfig()
	}

	imgSize := config.ImageSize * config.ImageSize * config.NumChannels

	sdm := &SpikingDiffusionModel{
		config:    config,
		scheduler: NewNoiseScheduler(config),
		neurons:   make([]*SpikingNeuron, config.NumBlocks),
		rng:       rand.New(rand.NewSource(42)),
	}

	// Create spiking neurons for each block
	for i := 0; i < config.NumBlocks; i++ {
		sdm.neurons[i] = NewSpikingNeuron(config.HiddenDim, config)
	}

	sdm.tsm = NewTemporalSpikingMechanism(config.HiddenDim, config)

	// Initialize U-Net weights (simplified)
	sdm.unetWeights = make([][]float64, config.NumBlocks)
	scale := math.Sqrt(2.0 / float64(imgSize+config.HiddenDim))
	for b := 0; b < config.NumBlocks; b++ {
		sdm.unetWeights[b] = make([]float64, imgSize)
		for i := 0; i < imgSize; i++ {
			sdm.unetWeights[b][i] = sdm.rng.NormFloat64() * scale
		}
	}

	return sdm
}

// Forward predicts noise for denoising
func (sdm *SpikingDiffusionModel) Forward(noisyImg []float64, timestep int) []float64 {
	// Reset neuron states
	for _, n := range sdm.neurons {
		n.Reset()
	}

	// Encode timestep into spike train
	timestepEmbedding := sdm.encodeTimestep(timestep)

	// Process through spiking U-Net
	spikeTrains := make([][]float64, sdm.config.SpikingTimesteps)

	for t := 0; t < sdm.config.SpikingTimesteps; t++ {
		// Combine input with timestep embedding
		combined := make([]float64, sdm.config.HiddenDim)
		for i := 0; i < sdm.config.HiddenDim && i < len(noisyImg); i++ {
			combined[i] = noisyImg[i] + timestepEmbedding[i%len(timestepEmbedding)]
		}

		// Pass through spiking blocks
		for _, neuron := range sdm.neurons {
			spikes := neuron.Forward(combined)
			sdm.totalSpikes += sdm.countSpikes(spikes)
			combined = spikes
		}

		spikeTrains[t] = combined
	}

	// Apply temporal spiking mechanism
	output := sdm.tsm.Forward(spikeTrains)

	// Scale to noise prediction
	noisePred := make([]float64, len(noisyImg))
	for i := range noisePred {
		noisePred[i] = output[i%len(output)]
	}

	return noisePred
}

// encodeTimestep converts timestep to embedding
func (sdm *SpikingDiffusionModel) encodeTimestep(t int) []float64 {
	dim := sdm.config.HiddenDim
	embedding := make([]float64, dim)

	// Sinusoidal embedding
	for i := 0; i < dim/2; i++ {
		freq := math.Pow(10000, -float64(i)/float64(dim/2))
		embedding[2*i] = math.Sin(float64(t) * freq)
		embedding[2*i+1] = math.Cos(float64(t) * freq)
	}

	return embedding
}

// countSpikes counts number of spikes in a vector
func (sdm *SpikingDiffusionModel) countSpikes(spikes []float64) int64 {
	count := int64(0)
	for _, s := range spikes {
		if s > 0.5 {
			count++
		}
	}
	return count
}

// Generate generates an image from noise
func (sdm *SpikingDiffusionModel) Generate() []float64 {
	imgSize := sdm.config.ImageSize * sdm.config.ImageSize * sdm.config.NumChannels

	// Start from pure noise
	x := make([]float64, imgSize)
	for i := range x {
		x[i] = sdm.rng.NormFloat64()
	}

	// Reverse diffusion
	for t := sdm.config.NumTimesteps - 1; t >= 0; t-- {
		// Predict noise
		noisePred := sdm.Forward(x, t)

		// Denoise step
		sqrtAlpha := sdm.scheduler.sqrtAlphas[t]
		sqrtOneMinusAlpha := sdm.scheduler.sqrtOneMinusAlphas[t]

		for i := range x {
			// x_0 = (x_t - sqrt(1-α) × ε) / sqrt(α)
			x[i] = (x[i] - sqrtOneMinusAlpha*noisePred[i]) / sqrtAlpha

			// Add noise for t > 0
			if t > 0 {
				sigma := math.Sqrt(sdm.scheduler.betas[t])
				x[i] += sigma * sdm.rng.NormFloat64()
			}
		}
	}

	return x
}

// GetEnergyEstimate estimates energy consumption
func (sdm *SpikingDiffusionModel) GetEnergyEstimate() float64 {
	// Energy per spike ≈ 0.9 pJ (typical for neuromorphic)
	// Energy per MAC ≈ 0.1 pJ (for CIM)
	spikeEnergy := float64(sdm.totalSpikes) * 0.9e-12
	macEnergy := float64(sdm.totalOps) * 0.1e-12
	return spikeEnergy + macEnergy
}

// =============================================================================
// ANALOG ERROR CORRECTION AND FAULT TOLERANCE
// =============================================================================

// FaultType represents types of device faults
type FaultType int

const (
	FaultNone      FaultType = iota
	FaultStuckOn             // Device stuck at high conductance
	FaultStuckOff            // Device stuck at low conductance
	FaultDrift               // Conductance drift over time
	FaultVariation           // Random variation
)

// FaultConfig configures fault injection and tolerance
type FaultConfig struct {
	FaultRate          float64 // Probability of fault per device
	StuckOnValue       float64 // Conductance for stuck-on
	StuckOffValue      float64 // Conductance for stuck-off
	DriftRate          float64 // Drift per operation
	VariationSigma     float64 // Standard deviation of variation
	EnableCompensation bool    // Enable compensation arrays
	NumCompLayers      int     // Number of compensation layers
}

// DefaultFaultConfig returns typical fault configuration
func DefaultFaultConfig() *FaultConfig {
	return &FaultConfig{
		FaultRate:          0.01, // 1% fault rate
		StuckOnValue:       1.0,
		StuckOffValue:      0.0,
		DriftRate:          0.001,
		VariationSigma:     0.05,
		EnableCompensation: true,
		NumCompLayers:      3,
	}
}

// FaultyDevice represents a device with potential faults
type FaultyDevice struct {
	nominalValue float64
	actualValue  float64
	faultType    FaultType
}

// FaultyCrossbar represents a crossbar with faulty devices
type FaultyCrossbar struct {
	config   *FaultConfig
	rows     int
	cols     int
	devices  [][]*FaultyDevice
	faultMap [][]FaultType
	rng      *rand.Rand
}

// NewFaultyCrossbar creates a crossbar with fault injection
func NewFaultyCrossbar(rows, cols int, config *FaultConfig) *FaultyCrossbar {
	if config == nil {
		config = DefaultFaultConfig()
	}

	fc := &FaultyCrossbar{
		config:   config,
		rows:     rows,
		cols:     cols,
		devices:  make([][]*FaultyDevice, rows),
		faultMap: make([][]FaultType, rows),
		rng:      rand.New(rand.NewSource(42)),
	}

	// Initialize devices
	for i := 0; i < rows; i++ {
		fc.devices[i] = make([]*FaultyDevice, cols)
		fc.faultMap[i] = make([]FaultType, cols)
		for j := 0; j < cols; j++ {
			fc.devices[i][j] = &FaultyDevice{
				nominalValue: 0,
				actualValue:  0,
				faultType:    FaultNone,
			}
		}
	}

	return fc
}

// InjectFaults randomly injects faults into the crossbar
func (fc *FaultyCrossbar) InjectFaults() int {
	faultCount := 0

	for i := 0; i < fc.rows; i++ {
		for j := 0; j < fc.cols; j++ {
			if fc.rng.Float64() < fc.config.FaultRate {
				// Randomly choose fault type
				if fc.rng.Float64() < 0.5 {
					fc.devices[i][j].faultType = FaultStuckOn
					fc.devices[i][j].actualValue = fc.config.StuckOnValue
				} else {
					fc.devices[i][j].faultType = FaultStuckOff
					fc.devices[i][j].actualValue = fc.config.StuckOffValue
				}
				fc.faultMap[i][j] = fc.devices[i][j].faultType
				faultCount++
			}
		}
	}

	return faultCount
}

// Program programs a value to a device
func (fc *FaultyCrossbar) Program(row, col int, value float64) {
	if row < 0 || row >= fc.rows || col < 0 || col >= fc.cols {
		return
	}

	device := fc.devices[row][col]
	device.nominalValue = value

	// Apply fault effects
	switch device.faultType {
	case FaultNone:
		// Apply variation
		device.actualValue = value + fc.rng.NormFloat64()*fc.config.VariationSigma
	case FaultStuckOn:
		device.actualValue = fc.config.StuckOnValue
	case FaultStuckOff:
		device.actualValue = fc.config.StuckOffValue
	case FaultDrift:
		device.actualValue = value * (1 + fc.config.DriftRate)
	case FaultVariation:
		device.actualValue = value + fc.rng.NormFloat64()*fc.config.VariationSigma
	}
}

// MVM performs matrix-vector multiplication with faults
func (fc *FaultyCrossbar) MVM(input []float64) []float64 {
	output := make([]float64, fc.cols)

	for j := 0; j < fc.cols; j++ {
		for i := 0; i < fc.rows && i < len(input); i++ {
			output[j] += input[i] * fc.devices[i][j].actualValue
		}
	}

	return output
}

// GetFaultRate returns actual fault rate
func (fc *FaultyCrossbar) GetFaultRate() float64 {
	faults := 0
	total := fc.rows * fc.cols

	for i := 0; i < fc.rows; i++ {
		for j := 0; j < fc.cols; j++ {
			if fc.devices[i][j].faultType != FaultNone {
				faults++
			}
		}
	}

	return float64(faults) / float64(total)
}

// MatrixDecomposition implements M = M_A × M_B for fault tolerance
type MatrixDecomposition struct {
	config *FaultConfig
	rows   int
	cols   int
	rank   int    // Decomposition rank
	matA   [][]float64
	matB   [][]float64
}

// NewMatrixDecomposition creates a matrix decomposition
func NewMatrixDecomposition(rows, cols, rank int, config *FaultConfig) *MatrixDecomposition {
	if config == nil {
		config = DefaultFaultConfig()
	}

	md := &MatrixDecomposition{
		config: config,
		rows:   rows,
		cols:   cols,
		rank:   rank,
		matA:   make([][]float64, rows),
		matB:   make([][]float64, rank),
	}

	// Initialize matrices
	for i := 0; i < rows; i++ {
		md.matA[i] = make([]float64, rank)
	}
	for i := 0; i < rank; i++ {
		md.matB[i] = make([]float64, cols)
	}

	return md
}

// Decompose decomposes a target matrix with fault awareness
func (md *MatrixDecomposition) Decompose(target [][]float64, faultyCrossbar *FaultyCrossbar) {
	rng := rand.New(rand.NewSource(42))

	// Initialize with random values
	scale := math.Sqrt(2.0 / float64(md.rows+md.cols))
	for i := 0; i < md.rows; i++ {
		for k := 0; k < md.rank; k++ {
			md.matA[i][k] = rng.NormFloat64() * scale
		}
	}
	for k := 0; k < md.rank; k++ {
		for j := 0; j < md.cols; j++ {
			md.matB[k][j] = rng.NormFloat64() * scale
		}
	}

	// Iterative optimization avoiding faulty devices
	lr := 0.01
	for iter := 0; iter < 1000; iter++ {
		// Compute current approximation
		approx := md.multiply()

		// Compute gradient and update
		for i := 0; i < md.rows && i < len(target); i++ {
			for j := 0; j < md.cols && j < len(target[i]); j++ {
				// Skip faulty positions
				if faultyCrossbar != nil &&
					faultyCrossbar.devices[i][j].faultType != FaultNone {
					continue
				}

				error := target[i][j] - approx[i][j]

				// Update A and B
				for k := 0; k < md.rank; k++ {
					md.matA[i][k] += lr * error * md.matB[k][j]
					md.matB[k][j] += lr * error * md.matA[i][k]
				}
			}
		}
	}
}

// multiply computes M_A × M_B
func (md *MatrixDecomposition) multiply() [][]float64 {
	result := make([][]float64, md.rows)
	for i := 0; i < md.rows; i++ {
		result[i] = make([]float64, md.cols)
		for j := 0; j < md.cols; j++ {
			for k := 0; k < md.rank; k++ {
				result[i][j] += md.matA[i][k] * md.matB[k][j]
			}
		}
	}
	return result
}

// ComputeCosineSimilarity computes cosine similarity between matrices
func (md *MatrixDecomposition) ComputeCosineSimilarity(target [][]float64) float64 {
	approx := md.multiply()

	dotProduct := 0.0
	normTarget := 0.0
	normApprox := 0.0

	for i := 0; i < md.rows && i < len(target); i++ {
		for j := 0; j < md.cols && j < len(target[i]); j++ {
			dotProduct += target[i][j] * approx[i][j]
			normTarget += target[i][j] * target[i][j]
			normApprox += approx[i][j] * approx[i][j]
		}
	}

	if normTarget == 0 || normApprox == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normTarget) * math.Sqrt(normApprox))
}

// CompensationArray implements iterative error compensation
type CompensationArray struct {
	config     *FaultConfig
	layers     []*FaultyCrossbar
	scales     []float64 // Scaling factors per layer
}

// NewCompensationArray creates a compensation array system
func NewCompensationArray(rows, cols int, config *FaultConfig) *CompensationArray {
	if config == nil {
		config = DefaultFaultConfig()
	}

	ca := &CompensationArray{
		config: config,
		layers: make([]*FaultyCrossbar, config.NumCompLayers),
		scales: make([]float64, config.NumCompLayers),
	}

	// Create compensation layers with decreasing scale
	for i := 0; i < config.NumCompLayers; i++ {
		ca.layers[i] = NewFaultyCrossbar(rows, cols, config)
		ca.scales[i] = math.Pow(0.1, float64(i+1)) // 0.1, 0.01, 0.001...
	}

	return ca
}

// Compensate programs compensation for residual errors
func (ca *CompensationArray) Compensate(target [][]float64, main *FaultyCrossbar) {
	current := target

	for layerIdx, layer := range ca.layers {
		// Compute residual error
		residual := make([][]float64, len(target))
		for i := range target {
			residual[i] = make([]float64, len(target[i]))
			for j := range target[i] {
				mainVal := main.devices[i][j].actualValue
				residual[i][j] = (target[i][j] - mainVal) / ca.scales[layerIdx]
			}
		}

		// Program compensation layer
		for i := 0; i < len(residual) && i < layer.rows; i++ {
			for j := 0; j < len(residual[i]) && j < layer.cols; j++ {
				layer.Program(i, j, residual[i][j])
			}
		}

		current = residual
	}
}

// MVM performs compensated MVM
func (ca *CompensationArray) MVM(input []float64, main *FaultyCrossbar) []float64 {
	// Main computation
	output := main.MVM(input)

	// Add scaled compensation
	for i, layer := range ca.layers {
		compOutput := layer.MVM(input)
		for j := range output {
			if j < len(compOutput) {
				output[j] += ca.scales[i] * compOutput[j]
			}
		}
	}

	return output
}

// BitSlicing implements bit-sliced computation for precision
type BitSlicing struct {
	config     *FaultConfig
	numSlices  int
	slices     []*FaultyCrossbar
	baseValues []float64
}

// NewBitSlicing creates a bit-sliced array
func NewBitSlicing(rows, cols, numBits int, config *FaultConfig) *BitSlicing {
	if config == nil {
		config = DefaultFaultConfig()
	}

	bs := &BitSlicing{
		config:     config,
		numSlices:  numBits,
		slices:     make([]*FaultyCrossbar, numBits),
		baseValues: make([]float64, numBits),
	}

	// Create slice arrays
	for i := 0; i < numBits; i++ {
		bs.slices[i] = NewFaultyCrossbar(rows, cols, config)
		bs.baseValues[i] = math.Pow(2, float64(i))
	}

	return bs
}

// ProgramSliced programs a value using bit slicing
func (bs *BitSlicing) ProgramSliced(row, col int, value float64) {
	// Quantize to integer
	quantized := int(math.Round(value * math.Pow(2, float64(bs.numSlices-1))))
	if quantized < 0 {
		quantized = 0
	}

	// Program each bit slice
	for i := 0; i < bs.numSlices; i++ {
		bit := (quantized >> i) & 1
		bs.slices[i].Program(row, col, float64(bit))
	}
}

// MVMSliced performs bit-sliced MVM
func (bs *BitSlicing) MVMSliced(input []float64) []float64 {
	// Initialize output
	var output []float64

	// Sum weighted contributions from each slice
	for i, slice := range bs.slices {
		sliceOutput := slice.MVM(input)

		if output == nil {
			output = make([]float64, len(sliceOutput))
		}

		for j := range sliceOutput {
			output[j] += bs.baseValues[i] * sliceOutput[j]
		}
	}

	// Scale back
	scaleFactor := math.Pow(2, float64(bs.numSlices-1))
	for i := range output {
		output[i] /= scaleFactor
	}

	return output
}

// DifferentialReadout implements differential scheme for robustness
type DifferentialReadout struct {
	positive *FaultyCrossbar
	negative *FaultyCrossbar
}

// NewDifferentialReadout creates a differential readout pair
func NewDifferentialReadout(rows, cols int, config *FaultConfig) *DifferentialReadout {
	return &DifferentialReadout{
		positive: NewFaultyCrossbar(rows, cols, config),
		negative: NewFaultyCrossbar(rows, cols, config),
	}
}

// ProgramDifferential programs using differential encoding
func (dr *DifferentialReadout) ProgramDifferential(row, col int, value float64) {
	// Split into positive and negative parts
	if value >= 0 {
		dr.positive.Program(row, col, value)
		dr.negative.Program(row, col, 0)
	} else {
		dr.positive.Program(row, col, 0)
		dr.negative.Program(row, col, -value)
	}
}

// MVMDifferential performs differential MVM
func (dr *DifferentialReadout) MVMDifferential(input []float64) []float64 {
	posOutput := dr.positive.MVM(input)
	negOutput := dr.negative.MVM(input)

	output := make([]float64, len(posOutput))
	for i := range output {
		output[i] = posOutput[i] - negOutput[i]
	}

	return output
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// GenerateTestImage creates a test image
func GenerateTestImageDiffusion(size, channels int, seed int64) []float64 {
	rng := rand.New(rand.NewSource(seed))
	img := make([]float64, size*size*channels)

	for i := range img {
		img[i] = rng.Float64()
	}

	return img
}

// FormatSpikingDiffusionReport generates SDM report
func FormatSpikingDiffusionReport(sdm *SpikingDiffusionModel) string {
	report := "=== Spiking Diffusion Model Report ===\n\n"

	report += fmt.Sprintf("Configuration:\n")
	report += fmt.Sprintf("  Diffusion timesteps: %d\n", sdm.config.NumTimesteps)
	report += fmt.Sprintf("  Spiking timesteps: %d\n", sdm.config.SpikingTimesteps)
	report += fmt.Sprintf("  Image size: %dx%dx%d\n",
		sdm.config.ImageSize, sdm.config.ImageSize, sdm.config.NumChannels)
	report += fmt.Sprintf("  U-Net blocks: %d\n", sdm.config.NumBlocks)

	report += fmt.Sprintf("\nSNN Parameters:\n")
	report += fmt.Sprintf("  Threshold: %.2f\n", sdm.config.Threshold)
	report += fmt.Sprintf("  Leak factor: %.2f\n", sdm.config.LeakFactor)

	report += fmt.Sprintf("\nStatistics:\n")
	report += fmt.Sprintf("  Total spikes: %d\n", sdm.totalSpikes)
	report += fmt.Sprintf("  Est. energy: %.2e J\n", sdm.GetEnergyEstimate())

	return report
}

// FormatFaultToleranceReport generates fault tolerance report
func FormatFaultToleranceReport(fc *FaultyCrossbar, md *MatrixDecomposition, target [][]float64) string {
	report := "=== Fault Tolerance Report ===\n\n"

	report += fmt.Sprintf("Crossbar Configuration:\n")
	report += fmt.Sprintf("  Size: %d × %d\n", fc.rows, fc.cols)
	report += fmt.Sprintf("  Actual fault rate: %.2f%%\n", fc.GetFaultRate()*100)

	if md != nil && target != nil {
		similarity := md.ComputeCosineSimilarity(target)
		report += fmt.Sprintf("\nMatrix Decomposition:\n")
		report += fmt.Sprintf("  Rank: %d\n", md.rank)
		report += fmt.Sprintf("  Cosine similarity: %.6f\n", similarity)
	}

	return report
}

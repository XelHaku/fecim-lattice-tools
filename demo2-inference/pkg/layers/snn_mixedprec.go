// Package layers provides spiking neural network (SNN) and mixed-precision
// quantization implementations for IronLattice ferroelectric CIM technology.
//
// This module implements:
// - Leaky Integrate-and-Fire (LIF) neurons with FeFET synapses
// - STDP learning on ferroelectric crossbar arrays
// - Sensitivity-aware mixed-precision quantization (CMQ, CIM²PQ)
// - Layer-wise precision allocation with Hessian analysis
// - Genetic Algorithm-Based Quantization (GAQ)
// - Dynamic ADC resolution optimization
//
// Based on research:
// - Sensitivity-Aware Mixed-Precision Quantization (arXiv:2512.19445)
// - CMQ: Crossbar-Aware Mixed-Precision Quantization (IEEE TCAD 2022)
// - CIM²PQ: Arraywise Hardware-Friendly MPQ (IEEE TCAD 2024)
// - Hardware-aware training for AIMC (Nature Communications 2023)
// - All-Ferroelectric SNN (Advanced Science 2024)
package layers

import (
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// SPIKING NEURAL NETWORK COMPONENTS
// =============================================================================

// LIFConfig configures leaky integrate-and-fire neuron parameters
type LIFConfig struct {
	// Membrane parameters
	VThreshold    float64 // Spike threshold voltage (mV), typically -55
	VReset        float64 // Reset voltage after spike (mV), typically -70
	VRest         float64 // Resting membrane potential (mV), typically -65
	TauMembrane   float64 // Membrane time constant (ms), typically 10-20
	TauRefactory  float64 // Refractory period (ms), typically 2
	LeakRate      float64 // Leak conductance (nS), typically 10

	// FeFET-specific parameters
	FeFETCoupling    float64 // Capacitive coupling to FeFET gate
	PolarizationBias float64 // Bias from ferroelectric polarization
	NoiseLevel       float64 // Membrane noise std dev (mV)
}

// DefaultLIFConfig returns standard LIF neuron parameters
func DefaultLIFConfig() *LIFConfig {
	return &LIFConfig{
		VThreshold:       -55.0,
		VReset:           -70.0,
		VRest:            -65.0,
		TauMembrane:      15.0,
		TauRefactory:     2.0,
		LeakRate:         10.0,
		FeFETCoupling:    0.8,
		PolarizationBias: 0.0,
		NoiseLevel:       0.5,
	}
}

// LIFNeuron implements a leaky integrate-and-fire neuron with FeFET coupling
type LIFNeuron struct {
	Config *LIFConfig

	// State variables
	VMembrane      float64   // Current membrane potential (mV)
	LastSpikeTime  float64   // Time of last spike (ms)
	RefractoryLeft float64   // Remaining refractory time (ms)
	SpikeHistory   []float64 // History of spike times

	// FeFET state
	FeFETPolarization float64 // Polarization state affecting threshold
	AdaptiveThreshold float64 // Dynamic threshold adjustment

	// Statistics
	TotalSpikes  int
	FiringRate   float64 // Hz
	ISIVariance  float64 // Inter-spike interval variance
	EnergyPerSpike float64 // fJ
}

// NewLIFNeuron creates a new LIF neuron
func NewLIFNeuron(config *LIFConfig) *LIFNeuron {
	if config == nil {
		config = DefaultLIFConfig()
	}
	return &LIFNeuron{
		Config:            config,
		VMembrane:         config.VRest,
		LastSpikeTime:     -1000.0,
		SpikeHistory:      make([]float64, 0, 1000),
		AdaptiveThreshold: config.VThreshold,
		EnergyPerSpike:    0.1, // fJ for FeFET-based neuron
	}
}

// Integrate processes input current for one timestep
func (n *LIFNeuron) Integrate(inputCurrent float64, dt float64, currentTime float64) bool {
	// Check refractory period
	if n.RefractoryLeft > 0 {
		n.RefractoryLeft -= dt
		return false
	}

	// Leaky integration with FeFET coupling
	dv := dt / n.Config.TauMembrane * (
		-(n.VMembrane - n.Config.VRest) +                              // Leak
		inputCurrent/n.Config.LeakRate +                               // Input
		n.Config.FeFETCoupling*n.FeFETPolarization +                   // FeFET bias
		n.Config.PolarizationBias)                                     // Static bias

	// Add membrane noise
	if n.Config.NoiseLevel > 0 {
		dv += rand.NormFloat64() * n.Config.NoiseLevel * math.Sqrt(dt)
	}

	n.VMembrane += dv

	// Effective threshold with FeFET modulation
	effectiveThreshold := n.AdaptiveThreshold +
		n.Config.FeFETCoupling*n.FeFETPolarization

	// Check for spike
	if n.VMembrane >= effectiveThreshold {
		n.VMembrane = n.Config.VReset
		n.RefractoryLeft = n.Config.TauRefactory
		n.LastSpikeTime = currentTime
		n.SpikeHistory = append(n.SpikeHistory, currentTime)
		n.TotalSpikes++
		return true
	}

	return false
}

// UpdateStatistics computes firing statistics
func (n *LIFNeuron) UpdateStatistics(totalTime float64) {
	if totalTime > 0 && n.TotalSpikes > 0 {
		n.FiringRate = float64(n.TotalSpikes) / (totalTime / 1000.0) // Convert to Hz
	}

	// Compute ISI variance
	if len(n.SpikeHistory) > 1 {
		var isis []float64
		for i := 1; i < len(n.SpikeHistory); i++ {
			isis = append(isis, n.SpikeHistory[i]-n.SpikeHistory[i-1])
		}
		meanISI := 0.0
		for _, isi := range isis {
			meanISI += isi
		}
		meanISI /= float64(len(isis))

		for _, isi := range isis {
			n.ISIVariance += (isi - meanISI) * (isi - meanISI)
		}
		n.ISIVariance /= float64(len(isis))
	}
}

// STDPConfig configures spike-timing dependent plasticity
type STDPConfig struct {
	// STDP time constants
	TauPlus   float64 // LTP time constant (ms), typically 20
	TauMinus  float64 // LTD time constant (ms), typically 20

	// Learning rates
	APlus     float64 // LTP amplitude, typically 0.01
	AMinus    float64 // LTD amplitude, typically 0.012 (asymmetric)

	// Weight bounds
	WMin      float64 // Minimum weight
	WMax      float64 // Maximum weight

	// FeFET-specific
	PolarizationStep   float64 // Polarization change per spike pair
	ConductanceLevels  int     // Number of discrete levels (MLC)
	RetentionDecay     float64 // Decay rate per second
	WriteEnergyFJ      float64 // Energy per weight update (fJ)
}

// DefaultSTDPConfig returns standard STDP parameters
func DefaultSTDPConfig() *STDPConfig {
	return &STDPConfig{
		TauPlus:           20.0,
		TauMinus:          20.0,
		APlus:             0.01,
		AMinus:            0.012,
		WMin:              0.0,
		WMax:              1.0,
		PolarizationStep:  0.05,
		ConductanceLevels: 16,
		RetentionDecay:    1e-7,
		WriteEnergyFJ:     50.0,
	}
}

// FeFETSynapse implements an STDP synapse using FeFET
type FeFETSynapse struct {
	Config *STDPConfig

	// Synapse state
	Weight        float64 // Current weight (0-1 normalized)
	Polarization  float64 // FeFET polarization state
	Conductance   float64 // Conductance (µS)

	// STDP traces
	PreTrace      float64 // Pre-synaptic trace
	PostTrace     float64 // Post-synaptic trace

	// Statistics
	LTPEvents     int
	LTDEvents     int
	TotalEnergy   float64 // fJ
}

// NewFeFETSynapse creates a new FeFET-based STDP synapse
func NewFeFETSynapse(config *STDPConfig, initialWeight float64) *FeFETSynapse {
	if config == nil {
		config = DefaultSTDPConfig()
	}
	s := &FeFETSynapse{
		Config:       config,
		Weight:       initialWeight,
		Polarization: 2*initialWeight - 1, // Map [0,1] to [-1,1]
	}
	s.updateConductance()
	return s
}

// updateConductance converts polarization to conductance
func (s *FeFETSynapse) updateConductance() {
	// Quantize to discrete levels
	normalizedP := (s.Polarization + 1) / 2 // Map [-1,1] to [0,1]
	level := int(normalizedP * float64(s.Config.ConductanceLevels-1))
	s.Weight = float64(level) / float64(s.Config.ConductanceLevels-1)
	s.Conductance = s.Weight * 100.0 // Max 100 µS
}

// PreSpike handles pre-synaptic spike arrival
func (s *FeFETSynapse) PreSpike(dt float64) float64 {
	// Update trace
	s.PreTrace += s.Config.APlus

	// LTD: pre after post
	if s.PostTrace > 0 {
		dw := -s.Config.AMinus * s.PostTrace
		s.applyWeightChange(dw)
		s.LTDEvents++
	}

	return s.Weight
}

// PostSpike handles post-synaptic spike
func (s *FeFETSynapse) PostSpike(dt float64) {
	// Update trace
	s.PostTrace += s.Config.AMinus

	// LTP: post after pre
	if s.PreTrace > 0 {
		dw := s.Config.APlus * s.PreTrace
		s.applyWeightChange(dw)
		s.LTPEvents++
	}
}

// applyWeightChange updates weight with FeFET dynamics
func (s *FeFETSynapse) applyWeightChange(dw float64) {
	// Update polarization
	s.Polarization += dw * s.Config.PolarizationStep

	// Clamp polarization
	if s.Polarization < -1 {
		s.Polarization = -1
	} else if s.Polarization > 1 {
		s.Polarization = 1
	}

	s.updateConductance()
	s.TotalEnergy += s.Config.WriteEnergyFJ
}

// DecayTraces decays STDP traces over time
func (s *FeFETSynapse) DecayTraces(dt float64) {
	s.PreTrace *= math.Exp(-dt / s.Config.TauPlus)
	s.PostTrace *= math.Exp(-dt / s.Config.TauMinus)
}

// SNNCrossbarConfig configures SNN crossbar array
type SNNCrossbarConfig struct {
	Rows          int          // Number of input neurons
	Cols          int          // Number of output neurons
	LIFConfig     *LIFConfig   // Neuron configuration
	STDPConfig    *STDPConfig  // Synapse configuration
	TimeStep      float64      // Simulation timestep (ms)
	EnableSTDP    bool         // Enable online learning
	InputEncoding string       // "rate", "temporal", "delta"
}

// SNNCrossbar implements a spiking neural network on FeFET crossbar
type SNNCrossbar struct {
	Config *SNNCrossbarConfig

	// Network components
	InputNeurons  []*LIFNeuron
	OutputNeurons []*LIFNeuron
	Synapses      [][]*FeFETSynapse

	// Spike buffers
	InputSpikes   []bool
	OutputSpikes  []bool

	// Statistics
	TotalEnergy     float64 // fJ
	TotalSpikes     int
	Throughput      float64 // spikes/ms
	EnergyPerSpike  float64 // fJ/spike
	EnergyVsANN     float64 // Improvement factor
}

// NewSNNCrossbar creates a new SNN crossbar array
func NewSNNCrossbar(config *SNNCrossbarConfig) *SNNCrossbar {
	snn := &SNNCrossbar{
		Config:        config,
		InputNeurons:  make([]*LIFNeuron, config.Rows),
		OutputNeurons: make([]*LIFNeuron, config.Cols),
		Synapses:      make([][]*FeFETSynapse, config.Rows),
		InputSpikes:   make([]bool, config.Rows),
		OutputSpikes:  make([]bool, config.Cols),
	}

	// Initialize neurons
	for i := 0; i < config.Rows; i++ {
		snn.InputNeurons[i] = NewLIFNeuron(config.LIFConfig)
	}
	for j := 0; j < config.Cols; j++ {
		snn.OutputNeurons[j] = NewLIFNeuron(config.LIFConfig)
	}

	// Initialize synapses with random weights
	for i := 0; i < config.Rows; i++ {
		snn.Synapses[i] = make([]*FeFETSynapse, config.Cols)
		for j := 0; j < config.Cols; j++ {
			initialWeight := rand.Float64()
			snn.Synapses[i][j] = NewFeFETSynapse(config.STDPConfig, initialWeight)
		}
	}

	return snn
}

// Step advances simulation by one timestep
func (snn *SNNCrossbar) Step(inputCurrents []float64, currentTime float64) []bool {
	dt := snn.Config.TimeStep

	// Process input neurons
	for i, current := range inputCurrents {
		snn.InputSpikes[i] = snn.InputNeurons[i].Integrate(current, dt, currentTime)
		if snn.InputSpikes[i] {
			snn.TotalSpikes++
		}
	}

	// Compute synaptic currents to output neurons
	outputCurrents := make([]float64, snn.Config.Cols)
	for i := 0; i < snn.Config.Rows; i++ {
		if snn.InputSpikes[i] {
			for j := 0; j < snn.Config.Cols; j++ {
				weight := snn.Synapses[i][j].PreSpike(dt)
				outputCurrents[j] += weight * 10.0 // Scale factor
			}
		}
	}

	// Process output neurons
	for j, current := range outputCurrents {
		snn.OutputSpikes[j] = snn.OutputNeurons[j].Integrate(current, dt, currentTime)
		if snn.OutputSpikes[j] {
			snn.TotalSpikes++
			// STDP: post-spike
			if snn.Config.EnableSTDP {
				for i := 0; i < snn.Config.Rows; i++ {
					snn.Synapses[i][j].PostSpike(dt)
				}
			}
		}
	}

	// Decay STDP traces
	if snn.Config.EnableSTDP {
		for i := 0; i < snn.Config.Rows; i++ {
			for j := 0; j < snn.Config.Cols; j++ {
				snn.Synapses[i][j].DecayTraces(dt)
			}
		}
	}

	// Update energy
	snn.TotalEnergy += float64(snn.TotalSpikes) * 0.1 // fJ per spike event

	return snn.OutputSpikes
}

// RateEncode converts analog values to spike rates
func (snn *SNNCrossbar) RateEncode(values []float64, maxRate float64, duration float64) [][]float64 {
	steps := int(duration / snn.Config.TimeStep)
	encoded := make([][]float64, steps)

	for t := 0; t < steps; t++ {
		encoded[t] = make([]float64, len(values))
		for i, v := range values {
			// Poisson spike generation
			rate := v * maxRate
			prob := rate * snn.Config.TimeStep / 1000.0
			if rand.Float64() < prob {
				encoded[t][i] = 100.0 // Spike current
			}
		}
	}

	return encoded
}

// GetWeights returns current weight matrix
func (snn *SNNCrossbar) GetWeights() [][]float64 {
	weights := make([][]float64, snn.Config.Rows)
	for i := 0; i < snn.Config.Rows; i++ {
		weights[i] = make([]float64, snn.Config.Cols)
		for j := 0; j < snn.Config.Cols; j++ {
			weights[i][j] = snn.Synapses[i][j].Weight
		}
	}
	return weights
}

// =============================================================================
// MIXED-PRECISION QUANTIZATION
// =============================================================================

// PrecisionLevel defines quantization precision
type PrecisionLevel int

const (
	Precision2Bit PrecisionLevel = 2
	Precision3Bit PrecisionLevel = 3
	Precision4Bit PrecisionLevel = 4
	Precision5Bit PrecisionLevel = 5
	Precision6Bit PrecisionLevel = 6
	Precision8Bit PrecisionLevel = 8
)

// LayerSensitivity holds sensitivity metrics for a layer
type LayerSensitivity struct {
	LayerIndex      int
	HessianTrace    float64   // Hessian trace (second-order sensitivity)
	FisherInfo      float64   // Fisher information
	GradientNorm    float64   // Gradient L2 norm
	OutputVariance  float64   // Output activation variance
	GeometricProxy  float64   // LieQ geometric feature score
	RecommendedBits int       // Recommended precision
}

// MixedPrecisionConfig configures mixed-precision quantization
type MixedPrecisionConfig struct {
	// Precision options
	MinBits       int     // Minimum precision (typically 2)
	MaxBits       int     // Maximum precision (typically 8)
	DefaultBits   int     // Default precision (typically 4)

	// Sensitivity analysis
	UseSensitivity   bool    // Enable sensitivity-aware allocation
	SensitivityAlpha float64 // Sensitivity weighting factor

	// CMQ parameters
	CrossbarSize     int     // Target crossbar size for grouping
	GroupGranularity string  // "layer", "channel", "array", "filter"

	// CIM²PQ parameters
	QuantizeInputs   bool    // Also quantize activations
	QuantizePartials bool    // Quantize partial sums
	ADCResolution    int     // ADC bits for partial sums

	// Optimization
	TargetCompression float64 // Target model size reduction
	AccuracyConstraint float64 // Minimum accuracy to maintain
}

// DefaultMixedPrecisionConfig returns standard MPQ settings
func DefaultMixedPrecisionConfig() *MixedPrecisionConfig {
	return &MixedPrecisionConfig{
		MinBits:           2,
		MaxBits:           8,
		DefaultBits:       4,
		UseSensitivity:    true,
		SensitivityAlpha:  0.5,
		CrossbarSize:      64,
		GroupGranularity:  "array",
		QuantizeInputs:    true,
		QuantizePartials:  true,
		ADCResolution:     6,
		TargetCompression: 0.5,
		AccuracyConstraint: 0.95,
	}
}

// MixedPrecisionQuantizer implements CMQ and CIM²PQ
type MixedPrecisionQuantizer struct {
	Config *MixedPrecisionConfig

	// Layer-wise precision allocation
	LayerPrecisions  []int
	LayerSensitivities []*LayerSensitivity

	// Quantization scales
	WeightScales    []float64
	ActivationScales []float64
	PartialSumScales []float64

	// Statistics
	CompressionRatio  float64
	AvgPrecision      float64
	HardwareEfficiency float64 // Relative to baseline
	EnergyReduction   float64 // Relative to 8-bit
}

// NewMixedPrecisionQuantizer creates a new MPQ instance
func NewMixedPrecisionQuantizer(config *MixedPrecisionConfig) *MixedPrecisionQuantizer {
	if config == nil {
		config = DefaultMixedPrecisionConfig()
	}
	return &MixedPrecisionQuantizer{
		Config: config,
	}
}

// ComputeHessianSensitivity estimates layer sensitivity via Hessian trace
func (mpq *MixedPrecisionQuantizer) ComputeHessianSensitivity(
	weights [][]float64,
	gradients [][]float64,
) float64 {
	// Approximate Hessian trace using gradient outer product
	trace := 0.0
	n := len(weights)
	if n == 0 {
		return 0
	}
	m := len(weights[0])

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			// Fisher information approximation
			trace += gradients[i][j] * gradients[i][j]
		}
	}

	return trace / float64(n*m)
}

// ComputeGeometricProxy calculates LieQ geometric feature score
func (mpq *MixedPrecisionQuantizer) ComputeGeometricProxy(
	weights [][]float64,
) float64 {
	// Compute singular value spectrum
	// Simplified: use Frobenius norm as proxy
	frobNorm := 0.0
	maxVal := 0.0
	n := len(weights)
	if n == 0 {
		return 0
	}
	m := len(weights[0])

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			frobNorm += weights[i][j] * weights[i][j]
			if math.Abs(weights[i][j]) > maxVal {
				maxVal = math.Abs(weights[i][j])
			}
		}
	}

	// Geometric score: ratio of nuclear norm proxy to max
	return math.Sqrt(frobNorm) / (maxVal + 1e-8) / math.Sqrt(float64(n*m))
}

// AnalyzeLayerSensitivity computes sensitivity metrics for all layers
func (mpq *MixedPrecisionQuantizer) AnalyzeLayerSensitivity(
	allWeights [][][]float64,
	allGradients [][][]float64,
) []*LayerSensitivity {
	numLayers := len(allWeights)
	sensitivities := make([]*LayerSensitivity, numLayers)

	for l := 0; l < numLayers; l++ {
		sens := &LayerSensitivity{
			LayerIndex: l,
		}

		// Hessian trace
		if allGradients != nil && len(allGradients) > l {
			sens.HessianTrace = mpq.ComputeHessianSensitivity(allWeights[l], allGradients[l])
		}

		// Geometric proxy
		sens.GeometricProxy = mpq.ComputeGeometricProxy(allWeights[l])

		// Gradient norm
		if allGradients != nil && len(allGradients) > l {
			sens.GradientNorm = computeL2Norm(allGradients[l])
		}

		// Output variance (would need activations in practice)
		sens.OutputVariance = computeVariance(allWeights[l])

		sensitivities[l] = sens
	}

	mpq.LayerSensitivities = sensitivities
	return sensitivities
}

// AllocatePrecisions assigns bit-widths to each layer
func (mpq *MixedPrecisionQuantizer) AllocatePrecisions(
	sensitivities []*LayerSensitivity,
) []int {
	numLayers := len(sensitivities)
	precisions := make([]int, numLayers)

	if !mpq.Config.UseSensitivity {
		// Uniform precision
		for l := 0; l < numLayers; l++ {
			precisions[l] = mpq.Config.DefaultBits
		}
		mpq.LayerPrecisions = precisions
		return precisions
	}

	// Normalize sensitivity scores
	maxHessian := 0.0
	maxGeometric := 0.0
	for _, s := range sensitivities {
		if s.HessianTrace > maxHessian {
			maxHessian = s.HessianTrace
		}
		if s.GeometricProxy > maxGeometric {
			maxGeometric = s.GeometricProxy
		}
	}

	// Allocate based on combined score
	for l, s := range sensitivities {
		normHessian := 0.0
		normGeometric := 0.0
		if maxHessian > 0 {
			normHessian = s.HessianTrace / maxHessian
		}
		if maxGeometric > 0 {
			normGeometric = s.GeometricProxy / maxGeometric
		}

		// Combined sensitivity score
		combinedScore := mpq.Config.SensitivityAlpha*normHessian +
			(1-mpq.Config.SensitivityAlpha)*normGeometric

		// Map to precision
		precisionRange := float64(mpq.Config.MaxBits - mpq.Config.MinBits)
		bits := mpq.Config.MinBits + int(combinedScore*precisionRange)
		if bits > mpq.Config.MaxBits {
			bits = mpq.Config.MaxBits
		}

		precisions[l] = bits
		s.RecommendedBits = bits
	}

	mpq.LayerPrecisions = precisions
	mpq.computeAveragePrecision()
	return precisions
}

// GAQSearch performs Genetic Algorithm-Based Quantization search
func (mpq *MixedPrecisionQuantizer) GAQSearch(
	numLayers int,
	fitnessFunc func([]int) float64,
	generations int,
	populationSize int,
) []int {
	// Initialize population with random precision allocations
	population := make([][]int, populationSize)
	fitness := make([]float64, populationSize)

	for i := 0; i < populationSize; i++ {
		population[i] = make([]int, numLayers)
		for l := 0; l < numLayers; l++ {
			population[i][l] = mpq.Config.MinBits +
				rand.Intn(mpq.Config.MaxBits-mpq.Config.MinBits+1)
		}
		fitness[i] = fitnessFunc(population[i])
	}

	// Evolution loop
	for gen := 0; gen < generations; gen++ {
		// Selection (tournament)
		newPop := make([][]int, populationSize)
		for i := 0; i < populationSize; i++ {
			// Tournament selection
			a := rand.Intn(populationSize)
			b := rand.Intn(populationSize)
			if fitness[a] > fitness[b] {
				newPop[i] = copyIntSlice(population[a])
			} else {
				newPop[i] = copyIntSlice(population[b])
			}
		}

		// Crossover
		for i := 0; i < populationSize-1; i += 2 {
			if rand.Float64() < 0.8 { // Crossover probability
				point := rand.Intn(numLayers)
				for l := point; l < numLayers; l++ {
					newPop[i][l], newPop[i+1][l] = newPop[i+1][l], newPop[i][l]
				}
			}
		}

		// Mutation
		for i := 0; i < populationSize; i++ {
			for l := 0; l < numLayers; l++ {
				if rand.Float64() < 0.1 { // Mutation probability
					newPop[i][l] = mpq.Config.MinBits +
						rand.Intn(mpq.Config.MaxBits-mpq.Config.MinBits+1)
				}
			}
		}

		// Evaluate new population
		population = newPop
		for i := 0; i < populationSize; i++ {
			fitness[i] = fitnessFunc(population[i])
		}
	}

	// Return best individual
	bestIdx := 0
	bestFitness := fitness[0]
	for i := 1; i < populationSize; i++ {
		if fitness[i] > bestFitness {
			bestFitness = fitness[i]
			bestIdx = i
		}
	}

	mpq.LayerPrecisions = population[bestIdx]
	mpq.computeAveragePrecision()
	return population[bestIdx]
}

// QuantizeWeights applies mixed-precision quantization to weights
func (mpq *MixedPrecisionQuantizer) QuantizeWeights(
	weights [][]float64,
	precision int,
) ([][]float64, float64) {
	numLevels := 1 << precision
	n := len(weights)
	if n == 0 {
		return weights, 1.0
	}
	m := len(weights[0])

	// Compute scale factor
	maxVal := 0.0
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			if math.Abs(weights[i][j]) > maxVal {
				maxVal = math.Abs(weights[i][j])
			}
		}
	}
	scale := maxVal / float64(numLevels/2-1)
	if scale == 0 {
		scale = 1.0
	}

	// Quantize
	quantized := make([][]float64, n)
	for i := 0; i < n; i++ {
		quantized[i] = make([]float64, m)
		for j := 0; j < m; j++ {
			// Symmetric quantization
			q := math.Round(weights[i][j] / scale)
			// Clamp
			if q > float64(numLevels/2-1) {
				q = float64(numLevels / 2 - 1)
			} else if q < -float64(numLevels/2) {
				q = -float64(numLevels / 2)
			}
			quantized[i][j] = q * scale
		}
	}

	return quantized, scale
}

// QuantizeActivations applies quantization to activations
func (mpq *MixedPrecisionQuantizer) QuantizeActivations(
	activations []float64,
	precision int,
) ([]float64, float64) {
	numLevels := 1 << precision
	n := len(activations)

	// Compute scale (assuming ReLU, so unsigned)
	maxVal := 0.0
	for i := 0; i < n; i++ {
		if activations[i] > maxVal {
			maxVal = activations[i]
		}
	}
	scale := maxVal / float64(numLevels-1)
	if scale == 0 {
		scale = 1.0
	}

	// Quantize
	quantized := make([]float64, n)
	for i := 0; i < n; i++ {
		q := math.Round(activations[i] / scale)
		if q > float64(numLevels-1) {
			q = float64(numLevels - 1)
		} else if q < 0 {
			q = 0
		}
		quantized[i] = q * scale
	}

	return quantized, scale
}

// OptimizeADCResolution finds optimal per-layer ADC bits
func (mpq *MixedPrecisionQuantizer) OptimizeADCResolution(
	partialSumDistributions [][]float64,
	targetBits int,
) []int {
	numLayers := len(partialSumDistributions)
	adcBits := make([]int, numLayers)

	for l := 0; l < numLayers; l++ {
		dist := partialSumDistributions[l]
		if len(dist) == 0 {
			adcBits[l] = targetBits
			continue
		}

		// Analyze distribution to find required bits
		maxVal := 0.0
		minVal := 0.0
		for _, v := range dist {
			if v > maxVal {
				maxVal = v
			}
			if v < minVal {
				minVal = v
			}
		}

		dynamicRange := maxVal - minVal
		if dynamicRange == 0 {
			adcBits[l] = mpq.Config.MinBits
			continue
		}

		// Compute required bits for 99th percentile
		sort.Float64s(dist)
		p99 := dist[int(float64(len(dist))*0.99)]
		requiredRange := math.Log2(p99 - minVal + 1)
		bits := int(math.Ceil(requiredRange)) + 1

		if bits < mpq.Config.MinBits {
			bits = mpq.Config.MinBits
		} else if bits > targetBits {
			bits = targetBits
		}

		adcBits[l] = bits
	}

	return adcBits
}

// computeAveragePrecision calculates average bits across layers
func (mpq *MixedPrecisionQuantizer) computeAveragePrecision() {
	if len(mpq.LayerPrecisions) == 0 {
		return
	}
	total := 0
	for _, bits := range mpq.LayerPrecisions {
		total += bits
	}
	mpq.AvgPrecision = float64(total) / float64(len(mpq.LayerPrecisions))

	// Compression vs 8-bit baseline
	mpq.CompressionRatio = mpq.AvgPrecision / 8.0

	// Hardware efficiency (quadratic with precision)
	mpq.HardwareEfficiency = 64.0 / (mpq.AvgPrecision * mpq.AvgPrecision)

	// Energy reduction (approximately linear with precision)
	mpq.EnergyReduction = 8.0 / mpq.AvgPrecision
}

// =============================================================================
// CROSSBAR-AWARE QUANTIZATION (CMQ)
// =============================================================================

// CMQConfig configures crossbar-aware mixed-precision quantization
type CMQConfig struct {
	CrossbarRows  int     // Physical crossbar rows
	CrossbarCols  int     // Physical crossbar columns
	GroupSize     int     // Quantization group size
	NoiseAware    bool    // Consider crossbar noise
	NoiseSigma    float64 // Noise standard deviation

	// Search parameters
	SearchIterations int
	LearningRate     float64
	Temperature      float64 // For differentiable search
}

// DefaultCMQConfig returns standard CMQ settings
func DefaultCMQConfig() *CMQConfig {
	return &CMQConfig{
		CrossbarRows:     64,
		CrossbarCols:     64,
		GroupSize:        16,
		NoiseAware:       true,
		NoiseSigma:       0.02,
		SearchIterations: 100,
		LearningRate:     0.01,
		Temperature:      1.0,
	}
}

// CMQQuantizer implements crossbar-aware mixed-precision quantization
type CMQQuantizer struct {
	Config *CMQConfig
	MPQ    *MixedPrecisionQuantizer

	// Group-wise precisions
	GroupPrecisions [][]int // [layer][group]

	// Noise model
	NoiseFactors [][]float64 // Per-group noise

	// Statistics
	CrossbarUtilization float64
	NoiseRobustness     float64
}

// NewCMQQuantizer creates a new CMQ instance
func NewCMQQuantizer(config *CMQConfig, mpqConfig *MixedPrecisionConfig) *CMQQuantizer {
	if config == nil {
		config = DefaultCMQConfig()
	}
	return &CMQQuantizer{
		Config: config,
		MPQ:    NewMixedPrecisionQuantizer(mpqConfig),
	}
}

// ComputeGroupPrecisions determines precision for each weight group
func (cmq *CMQQuantizer) ComputeGroupPrecisions(
	weights [][][]float64, // [layer][row][col]
	gradients [][][]float64,
) [][]int {
	numLayers := len(weights)
	cmq.GroupPrecisions = make([][]int, numLayers)

	for l := 0; l < numLayers; l++ {
		layerWeights := weights[l]
		numRows := len(layerWeights)
		if numRows == 0 {
			continue
		}
		numCols := len(layerWeights[0])

		// Number of groups based on crossbar tiling
		numRowGroups := (numRows + cmq.Config.GroupSize - 1) / cmq.Config.GroupSize
		numColGroups := (numCols + cmq.Config.GroupSize - 1) / cmq.Config.GroupSize
		numGroups := numRowGroups * numColGroups

		cmq.GroupPrecisions[l] = make([]int, numGroups)

		// Analyze each group
		groupIdx := 0
		for rg := 0; rg < numRowGroups; rg++ {
			for cg := 0; cg < numColGroups; cg++ {
				// Extract group weights
				rStart := rg * cmq.Config.GroupSize
				rEnd := rStart + cmq.Config.GroupSize
				if rEnd > numRows {
					rEnd = numRows
				}
				cStart := cg * cmq.Config.GroupSize
				cEnd := cStart + cmq.Config.GroupSize
				if cEnd > numCols {
					cEnd = numCols
				}

				groupWeights := make([][]float64, rEnd-rStart)
				for i := 0; i < rEnd-rStart; i++ {
					groupWeights[i] = make([]float64, cEnd-cStart)
					for j := 0; j < cEnd-cStart; j++ {
						groupWeights[i][j] = layerWeights[rStart+i][cStart+j]
					}
				}

				// Compute group sensitivity
				sensitivity := cmq.MPQ.ComputeGeometricProxy(groupWeights)

				// Adjust for noise if enabled
				if cmq.Config.NoiseAware {
					// Higher precision needed for noise-sensitive groups
					noiseImpact := cmq.estimateNoiseImpact(groupWeights)
					sensitivity *= (1 + noiseImpact)
				}

				// Map sensitivity to precision
				bits := cmq.sensitivityToBits(sensitivity)
				cmq.GroupPrecisions[l][groupIdx] = bits
				groupIdx++
			}
		}
	}

	cmq.computeUtilization(weights)
	return cmq.GroupPrecisions
}

// estimateNoiseImpact estimates how noise affects a weight group
func (cmq *CMQQuantizer) estimateNoiseImpact(weights [][]float64) float64 {
	// Groups with small weights are more noise-sensitive
	avgWeight := 0.0
	count := 0
	for _, row := range weights {
		for _, w := range row {
			avgWeight += math.Abs(w)
			count++
		}
	}
	if count > 0 {
		avgWeight /= float64(count)
	}

	// SNR-based impact
	snr := avgWeight / cmq.Config.NoiseSigma
	if snr < 1 {
		return 1.0 // Very noise-sensitive
	}
	return 1.0 / snr
}

// sensitivityToBits maps sensitivity score to bit precision
func (cmq *CMQQuantizer) sensitivityToBits(sensitivity float64) int {
	// Normalize sensitivity to [0, 1]
	normSens := math.Tanh(sensitivity * 2)

	// Map to bit range
	bitRange := float64(cmq.MPQ.Config.MaxBits - cmq.MPQ.Config.MinBits)
	bits := cmq.MPQ.Config.MinBits + int(normSens*bitRange)

	if bits > cmq.MPQ.Config.MaxBits {
		bits = cmq.MPQ.Config.MaxBits
	}

	return bits
}

// computeUtilization calculates crossbar array utilization
func (cmq *CMQQuantizer) computeUtilization(weights [][][]float64) {
	totalCells := 0
	usedCells := 0

	for l := 0; l < len(weights); l++ {
		rows := len(weights[l])
		if rows == 0 {
			continue
		}
		cols := len(weights[l][0])

		// Count required crossbar tiles
		tilesR := (rows + cmq.Config.CrossbarRows - 1) / cmq.Config.CrossbarRows
		tilesC := (cols + cmq.Config.CrossbarCols - 1) / cmq.Config.CrossbarCols

		totalCells += tilesR * tilesC * cmq.Config.CrossbarRows * cmq.Config.CrossbarCols
		usedCells += rows * cols
	}

	if totalCells > 0 {
		cmq.CrossbarUtilization = float64(usedCells) / float64(totalCells)
	}
}

// QuantizeWithGroups applies group-wise quantization
func (cmq *CMQQuantizer) QuantizeWithGroups(
	weights [][]float64,
	groupPrecisions []int,
) [][]float64 {
	numRows := len(weights)
	if numRows == 0 {
		return weights
	}
	numCols := len(weights[0])

	quantized := make([][]float64, numRows)
	for i := 0; i < numRows; i++ {
		quantized[i] = make([]float64, numCols)
	}

	numRowGroups := (numRows + cmq.Config.GroupSize - 1) / cmq.Config.GroupSize
	numColGroups := (numCols + cmq.Config.GroupSize - 1) / cmq.Config.GroupSize

	groupIdx := 0
	for rg := 0; rg < numRowGroups; rg++ {
		for cg := 0; cg < numColGroups; cg++ {
			rStart := rg * cmq.Config.GroupSize
			rEnd := rStart + cmq.Config.GroupSize
			if rEnd > numRows {
				rEnd = numRows
			}
			cStart := cg * cmq.Config.GroupSize
			cEnd := cStart + cmq.Config.GroupSize
			if cEnd > numCols {
				cEnd = numCols
			}

			precision := cmq.MPQ.Config.DefaultBits
			if groupIdx < len(groupPrecisions) {
				precision = groupPrecisions[groupIdx]
			}

			// Extract, quantize, and place back
			groupWeights := make([][]float64, rEnd-rStart)
			for i := 0; i < rEnd-rStart; i++ {
				groupWeights[i] = weights[rStart+i][cStart : cEnd]
			}

			qGroup, _ := cmq.MPQ.QuantizeWeights(groupWeights, precision)

			for i := 0; i < rEnd-rStart; i++ {
				for j := 0; j < cEnd-cStart; j++ {
					quantized[rStart+i][cStart+j] = qGroup[i][j]
				}
			}

			groupIdx++
		}
	}

	return quantized
}

// =============================================================================
// CIM²PQ: ARRAYWISE MIXED-PRECISION QUANTIZATION
// =============================================================================

// CIM2PQConfig configures arraywise quantization
type CIM2PQConfig struct {
	ArraySize         int     // CIM array dimension
	InputBits         int     // Input activation bits
	WeightBits        int     // Default weight bits
	PartialSumBits    int     // Partial sum bits
	ADCBits           int     // ADC resolution

	// Evolutionary algorithm parameters
	PopulationSize    int
	Generations       int
	MutationRate      float64
	CrossoverRate     float64
}

// DefaultCIM2PQConfig returns standard CIM²PQ settings
func DefaultCIM2PQConfig() *CIM2PQConfig {
	return &CIM2PQConfig{
		ArraySize:        64,
		InputBits:        8,
		WeightBits:       4,
		PartialSumBits:   12,
		ADCBits:          6,
		PopulationSize:   50,
		Generations:      100,
		MutationRate:     0.1,
		CrossoverRate:    0.8,
	}
}

// CIM2PQQuantizer implements arraywise mixed-precision quantization
type CIM2PQQuantizer struct {
	Config *CIM2PQConfig

	// Array-wise precision allocation
	ArrayInputBits     []int // Per-array input precision
	ArrayWeightBits    []int // Per-array weight precision
	ArrayPartialBits   []int // Per-array partial sum precision

	// Fitness tracking
	BestFitness     float64
	FitnessHistory  []float64

	// Metrics
	AccuracyImprovement   float64 // vs baseline (%)
	EfficiencyImprovement float64 // vs baseline (%)
}

// NewCIM2PQQuantizer creates a new CIM²PQ instance
func NewCIM2PQQuantizer(config *CIM2PQConfig) *CIM2PQQuantizer {
	if config == nil {
		config = DefaultCIM2PQConfig()
	}
	return &CIM2PQQuantizer{
		Config:         config,
		FitnessHistory: make([]float64, 0, config.Generations),
	}
}

// QuantizationGenome represents a quantization configuration
type QuantizationGenome struct {
	InputBits    []int // Per-array input bits
	WeightBits   []int // Per-array weight bits
	PartialBits  []int // Per-array partial sum bits
	Fitness      float64
}

// EvolveQuantization runs evolutionary search for optimal precision
func (cim2pq *CIM2PQQuantizer) EvolveQuantization(
	numArrays int,
	evalFunc func(*QuantizationGenome) float64,
) *QuantizationGenome {
	// Initialize population
	population := make([]*QuantizationGenome, cim2pq.Config.PopulationSize)
	for i := 0; i < cim2pq.Config.PopulationSize; i++ {
		population[i] = cim2pq.randomGenome(numArrays)
		population[i].Fitness = evalFunc(population[i])
	}

	// Evolution loop
	for gen := 0; gen < cim2pq.Config.Generations; gen++ {
		// Sort by fitness
		sort.Slice(population, func(i, j int) bool {
			return population[i].Fitness > population[j].Fitness
		})

		cim2pq.FitnessHistory = append(cim2pq.FitnessHistory, population[0].Fitness)

		// Create new population
		newPop := make([]*QuantizationGenome, cim2pq.Config.PopulationSize)

		// Elitism: keep top 10%
		eliteCount := cim2pq.Config.PopulationSize / 10
		for i := 0; i < eliteCount; i++ {
			newPop[i] = population[i]
		}

		// Crossover and mutation for rest
		for i := eliteCount; i < cim2pq.Config.PopulationSize; i++ {
			// Tournament selection
			parent1 := cim2pq.tournamentSelect(population)
			parent2 := cim2pq.tournamentSelect(population)

			// Crossover
			child := cim2pq.crossover(parent1, parent2, numArrays)

			// Mutation
			cim2pq.mutate(child)

			// Evaluate
			child.Fitness = evalFunc(child)
			newPop[i] = child
		}

		population = newPop
	}

	// Return best
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	best := population[0]
	cim2pq.BestFitness = best.Fitness
	cim2pq.ArrayInputBits = best.InputBits
	cim2pq.ArrayWeightBits = best.WeightBits
	cim2pq.ArrayPartialBits = best.PartialBits

	return best
}

// randomGenome creates a random quantization configuration
func (cim2pq *CIM2PQQuantizer) randomGenome(numArrays int) *QuantizationGenome {
	g := &QuantizationGenome{
		InputBits:   make([]int, numArrays),
		WeightBits:  make([]int, numArrays),
		PartialBits: make([]int, numArrays),
	}

	for i := 0; i < numArrays; i++ {
		g.InputBits[i] = 4 + rand.Intn(5)   // 4-8 bits
		g.WeightBits[i] = 2 + rand.Intn(7)  // 2-8 bits
		g.PartialBits[i] = 8 + rand.Intn(9) // 8-16 bits
	}

	return g
}

// tournamentSelect performs tournament selection
func (cim2pq *CIM2PQQuantizer) tournamentSelect(population []*QuantizationGenome) *QuantizationGenome {
	tournamentSize := 3
	best := population[rand.Intn(len(population))]
	for i := 1; i < tournamentSize; i++ {
		candidate := population[rand.Intn(len(population))]
		if candidate.Fitness > best.Fitness {
			best = candidate
		}
	}
	return best
}

// crossover combines two parent genomes
func (cim2pq *CIM2PQQuantizer) crossover(p1, p2 *QuantizationGenome, numArrays int) *QuantizationGenome {
	child := &QuantizationGenome{
		InputBits:   make([]int, numArrays),
		WeightBits:  make([]int, numArrays),
		PartialBits: make([]int, numArrays),
	}

	if rand.Float64() < cim2pq.Config.CrossoverRate {
		// Two-point crossover
		point1 := rand.Intn(numArrays)
		point2 := point1 + rand.Intn(numArrays-point1)

		for i := 0; i < numArrays; i++ {
			if i < point1 || i >= point2 {
				child.InputBits[i] = p1.InputBits[i]
				child.WeightBits[i] = p1.WeightBits[i]
				child.PartialBits[i] = p1.PartialBits[i]
			} else {
				child.InputBits[i] = p2.InputBits[i]
				child.WeightBits[i] = p2.WeightBits[i]
				child.PartialBits[i] = p2.PartialBits[i]
			}
		}
	} else {
		// Copy parent 1
		copy(child.InputBits, p1.InputBits)
		copy(child.WeightBits, p1.WeightBits)
		copy(child.PartialBits, p1.PartialBits)
	}

	return child
}

// mutate applies random mutations to a genome
func (cim2pq *CIM2PQQuantizer) mutate(g *QuantizationGenome) {
	for i := range g.InputBits {
		if rand.Float64() < cim2pq.Config.MutationRate {
			g.InputBits[i] = 4 + rand.Intn(5)
		}
		if rand.Float64() < cim2pq.Config.MutationRate {
			g.WeightBits[i] = 2 + rand.Intn(7)
		}
		if rand.Float64() < cim2pq.Config.MutationRate {
			g.PartialBits[i] = 8 + rand.Intn(9)
		}
	}
}

// =============================================================================
// INTEGRATED SNN + MIXED-PRECISION SYSTEM
// =============================================================================

// IronLatticeSNNMPQConfig configures integrated system
type IronLatticeSNNMPQConfig struct {
	SNNCrossbarConfig  *SNNCrossbarConfig
	MPQConfig          *MixedPrecisionConfig
	CMQConfig          *CMQConfig
	CIM2PQConfig       *CIM2PQConfig

	// Mode selection
	EnableSNN          bool
	EnableMPQ          bool
	EnableCMQ          bool
	EnableCIM2PQ       bool

	// Target metrics
	TargetAccuracy     float64 // Minimum accuracy (0-1)
	TargetCompression  float64 // Target size reduction
	TargetEnergy       float64 // Target energy (fJ/op)
}

// DefaultIronLatticeSNNMPQConfig returns standard system settings
func DefaultIronLatticeSNNMPQConfig() *IronLatticeSNNMPQConfig {
	return &IronLatticeSNNMPQConfig{
		SNNCrossbarConfig: &SNNCrossbarConfig{
			Rows:          64,
			Cols:          64,
			LIFConfig:     DefaultLIFConfig(),
			STDPConfig:    DefaultSTDPConfig(),
			TimeStep:      1.0,
			EnableSTDP:    true,
			InputEncoding: "rate",
		},
		MPQConfig:        DefaultMixedPrecisionConfig(),
		CMQConfig:        DefaultCMQConfig(),
		CIM2PQConfig:     DefaultCIM2PQConfig(),
		EnableSNN:        true,
		EnableMPQ:        true,
		EnableCMQ:        true,
		EnableCIM2PQ:     true,
		TargetAccuracy:   0.95,
		TargetCompression: 0.5,
		TargetEnergy:     10.0,
	}
}

// IronLatticeSNNMPQ implements integrated SNN + mixed-precision system
type IronLatticeSNNMPQ struct {
	Config *IronLatticeSNNMPQConfig

	// Components
	SNNCrossbar   *SNNCrossbar
	MPQuantizer   *MixedPrecisionQuantizer
	CMQuantizer   *CMQQuantizer
	CIM2PQuantizer *CIM2PQQuantizer

	// System metrics
	TotalEnergy       float64 // fJ
	CompressionRatio  float64
	Accuracy          float64
	Throughput        float64 // ops/s

	// Comparison metrics
	EnergyVsANN       float64 // SNN energy improvement
	EnergyVs8Bit      float64 // MPQ energy improvement
	OverallImprovement float64 // Combined improvement
}

// NewIronLatticeSNNMPQ creates a new integrated system
func NewIronLatticeSNNMPQ(config *IronLatticeSNNMPQConfig) *IronLatticeSNNMPQ {
	if config == nil {
		config = DefaultIronLatticeSNNMPQConfig()
	}

	system := &IronLatticeSNNMPQ{
		Config: config,
	}

	if config.EnableSNN {
		system.SNNCrossbar = NewSNNCrossbar(config.SNNCrossbarConfig)
	}
	if config.EnableMPQ {
		system.MPQuantizer = NewMixedPrecisionQuantizer(config.MPQConfig)
	}
	if config.EnableCMQ {
		system.CMQuantizer = NewCMQQuantizer(config.CMQConfig, config.MPQConfig)
	}
	if config.EnableCIM2PQ {
		system.CIM2PQuantizer = NewCIM2PQQuantizer(config.CIM2PQConfig)
	}

	return system
}

// OptimizeSystem runs full optimization pipeline
func (sys *IronLatticeSNNMPQ) OptimizeSystem(
	weights [][][]float64,
	gradients [][][]float64,
	evalFunc func(weights [][][]float64, precisions []int) float64,
) {
	numLayers := len(weights)

	// Step 1: Sensitivity analysis
	if sys.Config.EnableMPQ {
		sys.MPQuantizer.AnalyzeLayerSensitivity(weights, gradients)

		// Allocate precisions
		precisions := sys.MPQuantizer.AllocatePrecisions(sys.MPQuantizer.LayerSensitivities)

		// GAQ refinement
		fitness := func(p []int) float64 {
			return evalFunc(weights, p)
		}
		sys.MPQuantizer.GAQSearch(numLayers, fitness, 50, 30)
	}

	// Step 2: CMQ group-wise optimization
	if sys.Config.EnableCMQ {
		sys.CMQuantizer.ComputeGroupPrecisions(weights, gradients)
	}

	// Step 3: CIM²PQ arraywise optimization
	if sys.Config.EnableCIM2PQ {
		numArrays := 0
		for l := 0; l < numLayers; l++ {
			if len(weights[l]) > 0 {
				rows := len(weights[l])
				cols := len(weights[l][0])
				tilesR := (rows + sys.Config.CIM2PQConfig.ArraySize - 1) / sys.Config.CIM2PQConfig.ArraySize
				tilesC := (cols + sys.Config.CIM2PQConfig.ArraySize - 1) / sys.Config.CIM2PQConfig.ArraySize
				numArrays += tilesR * tilesC
			}
		}

		cimEval := func(g *QuantizationGenome) float64 {
			// Simplified: compute efficiency-accuracy tradeoff
			avgBits := 0.0
			for _, b := range g.WeightBits {
				avgBits += float64(b)
			}
			avgBits /= float64(len(g.WeightBits))

			efficiency := 8.0 / avgBits // Higher is better
			accuracy := 1.0 - (8.0-avgBits)*0.01 // Simplified accuracy model

			return efficiency * accuracy
		}

		sys.CIM2PQuantizer.EvolveQuantization(numArrays, cimEval)
	}

	// Compute overall metrics
	sys.computeSystemMetrics()
}

// computeSystemMetrics calculates overall system performance
func (sys *IronLatticeSNNMPQ) computeSystemMetrics() {
	// SNN energy improvement (typically 10-100x over ANN)
	if sys.Config.EnableSNN && sys.SNNCrossbar != nil {
		// Event-driven SNN is much more efficient for sparse inputs
		sys.EnergyVsANN = 50.0 // Typical improvement factor
	} else {
		sys.EnergyVsANN = 1.0
	}

	// MPQ energy improvement
	if sys.Config.EnableMPQ && sys.MPQuantizer != nil {
		sys.EnergyVs8Bit = sys.MPQuantizer.EnergyReduction
	} else {
		sys.EnergyVs8Bit = 1.0
	}

	// Combined improvement
	sys.OverallImprovement = sys.EnergyVsANN * sys.EnergyVs8Bit

	// Compression from MPQ
	if sys.MPQuantizer != nil {
		sys.CompressionRatio = sys.MPQuantizer.CompressionRatio
	}
}

// RunSNNInference performs SNN-based inference
func (sys *IronLatticeSNNMPQ) RunSNNInference(
	input []float64,
	duration float64,
) []float64 {
	if sys.SNNCrossbar == nil {
		return nil
	}

	// Rate encode input
	encoded := sys.SNNCrossbar.RateEncode(input, 100.0, duration)

	// Run simulation
	outputSpikeCounts := make([]int, sys.Config.SNNCrossbarConfig.Cols)
	for t := 0; t < len(encoded); t++ {
		currentTime := float64(t) * sys.Config.SNNCrossbarConfig.TimeStep
		spikes := sys.SNNCrossbar.Step(encoded[t], currentTime)
		for j, spike := range spikes {
			if spike {
				outputSpikeCounts[j]++
			}
		}
	}

	// Convert spike counts to rates
	output := make([]float64, len(outputSpikeCounts))
	for j, count := range outputSpikeCounts {
		output[j] = float64(count) / duration * 1000.0 // Hz
	}

	sys.TotalEnergy += sys.SNNCrossbar.TotalEnergy

	return output
}

// QuantizeModel applies mixed-precision quantization to model
func (sys *IronLatticeSNNMPQ) QuantizeModel(
	weights [][][]float64,
) [][][]float64 {
	if sys.MPQuantizer == nil {
		return weights
	}

	quantized := make([][][]float64, len(weights))
	for l := 0; l < len(weights); l++ {
		precision := sys.Config.MPQConfig.DefaultBits
		if l < len(sys.MPQuantizer.LayerPrecisions) {
			precision = sys.MPQuantizer.LayerPrecisions[l]
		}

		quantized[l], _ = sys.MPQuantizer.QuantizeWeights(weights[l], precision)
	}

	return quantized
}

// GetStatistics returns system performance statistics
func (sys *IronLatticeSNNMPQ) GetStatistics() map[string]float64 {
	stats := make(map[string]float64)

	stats["total_energy_fJ"] = sys.TotalEnergy
	stats["compression_ratio"] = sys.CompressionRatio
	stats["accuracy"] = sys.Accuracy
	stats["energy_vs_ann"] = sys.EnergyVsANN
	stats["energy_vs_8bit"] = sys.EnergyVs8Bit
	stats["overall_improvement"] = sys.OverallImprovement

	if sys.MPQuantizer != nil {
		stats["avg_precision_bits"] = sys.MPQuantizer.AvgPrecision
		stats["hardware_efficiency"] = sys.MPQuantizer.HardwareEfficiency
	}

	if sys.CMQuantizer != nil {
		stats["crossbar_utilization"] = sys.CMQuantizer.CrossbarUtilization
	}

	if sys.CIM2PQuantizer != nil {
		stats["cim2pq_best_fitness"] = sys.CIM2PQuantizer.BestFitness
	}

	if sys.SNNCrossbar != nil {
		stats["snn_total_spikes"] = float64(sys.SNNCrossbar.TotalSpikes)
	}

	return stats
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// computeL2Norm calculates L2 norm of a matrix
func computeL2Norm(matrix [][]float64) float64 {
	sum := 0.0
	for _, row := range matrix {
		for _, v := range row {
			sum += v * v
		}
	}
	return math.Sqrt(sum)
}

// computeVariance calculates variance of matrix elements
func computeVariance(matrix [][]float64) float64 {
	var values []float64
	for _, row := range matrix {
		values = append(values, row...)
	}

	if len(values) == 0 {
		return 0
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))

	return variance
}

// copyIntSlice creates a copy of an int slice
func copyIntSlice(src []int) []int {
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

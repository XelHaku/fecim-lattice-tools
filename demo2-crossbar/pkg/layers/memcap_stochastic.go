// Package layers provides ferroelectric memcapacitor and stochastic computing
// simulation for IronLattice CIM architectures.
//
// This module simulates:
// - Ferroelectric memcapacitors (FeCap) for ultra-low power CIM
// - Memcapacitor crossbar arrays for MAC operations
// - Stochastic computing with FeFET probabilistic bits (p-bits)
// - True random number generators (TRNGs) using ferroelectric stochasticity
// - Bitstream neural network inference
//
// Based on research from Nature Scientific Reports 2024, IEEE TCAS 2024,
// and ferroelectric stochasticity studies from Small 2023.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// ============================================================================
// FERROELECTRIC MEMCAPACITOR CONFIGURATION
// ============================================================================

// MemcapConfig defines parameters for ferroelectric memcapacitor
type MemcapConfig struct {
	// Physical dimensions
	AreaUm2           float64 // Device area in µm²
	FEThicknessnm     float64 // Ferroelectric layer thickness
	ILThicknessnm     float64 // Interfacial layer thickness

	// Capacitance characteristics
	CminFPerUm2       float64 // Minimum capacitance density
	CmaxFPerUm2       float64 // Maximum capacitance density
	MemoryWindowFPerUm2 float64 // Capacitance memory window

	// Switching parameters
	CoerciveVoltageV  float64 // Coercive voltage for switching
	SaturationVoltageV float64 // Saturation voltage
	PulseWidthNs      float64 // Programming pulse width

	// Performance metrics
	EnduranceCycles   float64 // Cycling endurance
	RetentionYears    float64 // Data retention
	EnergyFJ          float64 // Energy per operation
}

// DefaultMemcapConfig returns optimized HZO memcapacitor configuration
func DefaultMemcapConfig() *MemcapConfig {
	return &MemcapConfig{
		AreaUm2:            1.0,
		FEThicknessnm:      10.0,
		ILThicknessnm:      1.0,
		CminFPerUm2:        2.0,
		CmaxFPerUm2:        9.8,
		MemoryWindowFPerUm2: 7.8,
		CoerciveVoltageV:   1.0,
		SaturationVoltageV: 3.0,
		PulseWidthNs:       100,
		EnduranceCycles:    1e9,
		RetentionYears:     10,
		EnergyFJ:           31,
	}
}

// StochasticConfig defines parameters for stochastic computing
type StochasticConfig struct {
	// P-bit parameters
	NumDomains        int     // Number of ferroelectric domains
	ThermalNoiseKT    float64 // Thermal energy (kT at room temp ~26 meV)
	DomainSizeNm      float64 // Average domain size

	// Stochasticity control
	BiasVoltageV      float64 // Bias voltage for probability tuning
	SwitchingBarrierEV float64 // Energy barrier for switching

	// Bitstream parameters
	BitstreamLength   int     // Length of stochastic bitstream
	CorrelationLength int     // Correlation in bitstream

	// TRNG parameters
	EntropyTarget     float64 // Target entropy (ideal = 1.0)
	HammingTarget     float64 // Target Hamming distance (ideal = 0.5)
}

// DefaultStochasticConfig returns optimized stochastic computing configuration
func DefaultStochasticConfig() *StochasticConfig {
	return &StochasticConfig{
		NumDomains:         100,
		ThermalNoiseKT:     0.026, // eV at 300K
		DomainSizeNm:       10,
		BiasVoltageV:       0.5,
		SwitchingBarrierEV: 0.1,
		BitstreamLength:    256,
		CorrelationLength:  8,
		EntropyTarget:      1.0,
		HammingTarget:      0.5,
	}
}

// ============================================================================
// FERROELECTRIC MEMCAPACITOR DEVICE
// ============================================================================

// FerroMemcapacitor represents a single ferroelectric memcapacitor device
type FerroMemcapacitor struct {
	Config *MemcapConfig

	// State
	Capacitance      float64 // Current capacitance (fF)
	PolarizationState float64 // Polarization state (0-1)
	StoredCharge     float64 // Stored charge (fC)

	// Cycling state
	CycleCount       float64
	DegradationFactor float64

	// Noise characteristics
	ReadNoise        float64 // Read noise standard deviation
	WriteNoise       float64 // Write noise standard deviation
}

// NewFerroMemcapacitor creates a new ferroelectric memcapacitor
func NewFerroMemcapacitor(config *MemcapConfig) *FerroMemcapacitor {
	mc := &FerroMemcapacitor{
		Config:            config,
		DegradationFactor: 1.0,
		ReadNoise:         0.01,
		WriteNoise:        0.02,
	}
	mc.Reset()
	return mc
}

// Reset resets the memcapacitor to initial state
func (mc *FerroMemcapacitor) Reset() {
	mc.PolarizationState = 0.5
	mc.updateCapacitance()
	mc.StoredCharge = 0
}

// updateCapacitance updates capacitance based on polarization state
func (mc *FerroMemcapacitor) updateCapacitance() {
	// Capacitance varies linearly with polarization
	cRange := mc.Config.CmaxFPerUm2 - mc.Config.CminFPerUm2
	mc.Capacitance = (mc.Config.CminFPerUm2 + cRange*mc.PolarizationState) *
		mc.Config.AreaUm2 * mc.DegradationFactor
}

// Program programs the memcapacitor to a target capacitance level
func (mc *FerroMemcapacitor) Program(targetPolarization float64, voltage float64) float64 {
	// Add write noise
	noise := mc.WriteNoise * rand.NormFloat64()
	actualPolarization := targetPolarization + noise

	// Clamp to valid range
	if actualPolarization < 0 {
		actualPolarization = 0
	}
	if actualPolarization > 1 {
		actualPolarization = 1
	}

	// Apply voltage-dependent switching
	if math.Abs(voltage) < mc.Config.CoerciveVoltageV {
		// Partial switching below coercive voltage
		switchFraction := math.Abs(voltage) / mc.Config.CoerciveVoltageV
		actualPolarization = mc.PolarizationState +
			(actualPolarization-mc.PolarizationState)*switchFraction
	}

	mc.PolarizationState = actualPolarization
	mc.updateCapacitance()
	mc.CycleCount++

	// Apply degradation
	if mc.CycleCount > mc.Config.EnduranceCycles*0.5 {
		degradationRate := (mc.CycleCount - mc.Config.EnduranceCycles*0.5) /
			(mc.Config.EnduranceCycles * 0.5)
		mc.DegradationFactor = 1.0 - 0.1*degradationRate
	}

	return mc.Capacitance
}

// Read reads the current capacitance with noise
func (mc *FerroMemcapacitor) Read() float64 {
	noise := mc.ReadNoise * mc.Capacitance * rand.NormFloat64()
	return mc.Capacitance + noise
}

// ChargeTransfer performs charge-based computation (key advantage over memristor)
func (mc *FerroMemcapacitor) ChargeTransfer(inputVoltage float64) float64 {
	// Q = C * V (charge transfer, not current)
	mc.StoredCharge = mc.Capacitance * inputVoltage
	return mc.StoredCharge
}

// GetEnergyConsumption returns energy for an operation
func (mc *FerroMemcapacitor) GetEnergyConsumption(voltage float64) float64 {
	// E = 0.5 * C * V^2 (capacitive energy)
	return 0.5 * mc.Capacitance * voltage * voltage // fJ
}

// ============================================================================
// MEMCAPACITOR CROSSBAR ARRAY
// ============================================================================

// MemcapCrossbar represents a crossbar array of memcapacitors
type MemcapCrossbar struct {
	Config   *MemcapConfig
	Rows     int
	Cols     int
	Devices  [][]*FerroMemcapacitor

	// Performance tracking
	TotalEnergy      float64
	OperationCount   int
	AverageLinearity float64
}

// NewMemcapCrossbar creates a memcapacitor crossbar array
func NewMemcapCrossbar(config *MemcapConfig, rows, cols int) *MemcapCrossbar {
	cb := &MemcapCrossbar{
		Config:  config,
		Rows:    rows,
		Cols:    cols,
		Devices: make([][]*FerroMemcapacitor, rows),
	}

	for i := 0; i < rows; i++ {
		cb.Devices[i] = make([]*FerroMemcapacitor, cols)
		for j := 0; j < cols; j++ {
			cb.Devices[i][j] = NewFerroMemcapacitor(config)
		}
	}

	return cb
}

// ProgramWeights programs weight matrix into the crossbar
func (cb *MemcapCrossbar) ProgramWeights(weights [][]float64) {
	for i := 0; i < cb.Rows && i < len(weights); i++ {
		for j := 0; j < cb.Cols && j < len(weights[i]); j++ {
			// Normalize weight to polarization (0-1)
			polarization := (weights[i][j] + 1) / 2 // Assume weights in [-1, 1]
			cb.Devices[i][j].Program(polarization, cb.Config.SaturationVoltageV)
		}
	}
}

// ComputeMAC performs matrix-vector multiplication using charge accumulation
func (cb *MemcapCrossbar) ComputeMAC(input []float64) []float64 {
	output := make([]float64, cb.Cols)
	totalEnergy := 0.0

	// Convert input to voltages
	inputVoltages := make([]float64, len(input))
	maxInput := 0.0
	for _, v := range input {
		if math.Abs(v) > maxInput {
			maxInput = math.Abs(v)
		}
	}
	for i, v := range input {
		inputVoltages[i] = v / maxInput * cb.Config.SaturationVoltageV
	}

	// Compute charge accumulation on each column
	for j := 0; j < cb.Cols; j++ {
		chargeSum := 0.0
		for i := 0; i < cb.Rows && i < len(inputVoltages); i++ {
			charge := cb.Devices[i][j].ChargeTransfer(inputVoltages[i])
			chargeSum += charge
			totalEnergy += cb.Devices[i][j].GetEnergyConsumption(inputVoltages[i])
		}
		output[j] = chargeSum
	}

	cb.TotalEnergy += totalEnergy
	cb.OperationCount++

	return output
}

// GetLinearity measures linearity of weight updates
func (cb *MemcapCrossbar) GetLinearity() float64 {
	// Test linearity by programming a sweep
	testDevice := NewFerroMemcapacitor(cb.Config)

	targets := make([]float64, 32)
	actuals := make([]float64, 32)

	for i := 0; i < 32; i++ {
		targets[i] = float64(i) / 31.0
		testDevice.Program(targets[i], cb.Config.SaturationVoltageV)
		actuals[i] = testDevice.PolarizationState
	}

	// Calculate R² (coefficient of determination)
	meanTarget := 0.0
	for _, t := range targets {
		meanTarget += t
	}
	meanTarget /= float64(len(targets))

	ssRes := 0.0
	ssTot := 0.0
	for i := range targets {
		ssRes += math.Pow(actuals[i]-targets[i], 2)
		ssTot += math.Pow(targets[i]-meanTarget, 2)
	}

	cb.AverageLinearity = 1.0 - ssRes/ssTot
	return cb.AverageLinearity
}

// ============================================================================
// STOCHASTIC COMPUTING WITH FeFET
// ============================================================================

// FerroelectricPBit represents a probabilistic bit using ferroelectric stochasticity
type FerroelectricPBit struct {
	Config *StochasticConfig

	// Domain states
	DomainStates []bool // Individual domain polarizations
	NetMagnetization float64 // Average polarization (-1 to 1)

	// Probability control
	CurrentProbability float64 // Probability of outputting 1
	BiasField          float64 // Applied bias field

	// Statistics
	OutputHistory      []bool
	TransitionCount    int
}

// NewFerroelectricPBit creates a new p-bit based on ferroelectric FET
func NewFerroelectricPBit(config *StochasticConfig) *FerroelectricPBit {
	pb := &FerroelectricPBit{
		Config:        config,
		DomainStates:  make([]bool, config.NumDomains),
		OutputHistory: make([]bool, 0, 1000),
	}

	// Initialize domains randomly
	for i := range pb.DomainStates {
		pb.DomainStates[i] = rand.Float64() < 0.5
	}
	pb.updateNetMagnetization()
	pb.updateProbability()

	return pb
}

// updateNetMagnetization calculates net magnetization from domains
func (pb *FerroelectricPBit) updateNetMagnetization() {
	sum := 0
	for _, state := range pb.DomainStates {
		if state {
			sum++
		} else {
			sum--
		}
	}
	pb.NetMagnetization = float64(sum) / float64(len(pb.DomainStates))
}

// updateProbability calculates output probability from magnetization and bias
func (pb *FerroelectricPBit) updateProbability() {
	// Boltzmann-like probability based on energy
	// P(1) = 1 / (1 + exp(-2 * (bias + magnetization) / kT))
	effectiveField := pb.BiasField + pb.NetMagnetization
	exponent := -2 * effectiveField / pb.Config.ThermalNoiseKT
	pb.CurrentProbability = 1.0 / (1.0 + math.Exp(exponent))

	// Clamp to avoid numerical issues
	if pb.CurrentProbability < 0.001 {
		pb.CurrentProbability = 0.001
	}
	if pb.CurrentProbability > 0.999 {
		pb.CurrentProbability = 0.999
	}
}

// SetBias sets the bias field to control probability
func (pb *FerroelectricPBit) SetBias(biasVoltage float64) {
	// Convert voltage to effective field
	pb.BiasField = biasVoltage / pb.Config.BiasVoltageV
	pb.updateProbability()
}

// Sample generates a stochastic output
func (pb *FerroelectricPBit) Sample() bool {
	// Apply thermal noise to domains
	for i := range pb.DomainStates {
		// Probability of flipping depends on barrier and thermal energy
		flipProb := math.Exp(-pb.Config.SwitchingBarrierEV / pb.Config.ThermalNoiseKT) * 0.1
		if rand.Float64() < flipProb {
			pb.DomainStates[i] = !pb.DomainStates[i]
			pb.TransitionCount++
		}
	}

	pb.updateNetMagnetization()
	pb.updateProbability()

	// Generate output based on probability
	output := rand.Float64() < pb.CurrentProbability
	pb.OutputHistory = append(pb.OutputHistory, output)

	return output
}

// GenerateBitstream generates a stochastic bitstream
func (pb *FerroelectricPBit) GenerateBitstream(length int) []bool {
	bitstream := make([]bool, length)
	for i := 0; i < length; i++ {
		bitstream[i] = pb.Sample()
	}
	return bitstream
}

// ============================================================================
// TRUE RANDOM NUMBER GENERATOR
// ============================================================================

// FerroTRNG represents a true random number generator using ferroelectric stochasticity
type FerroTRNG struct {
	Config   *StochasticConfig
	PBits    []*FerroelectricPBit

	// Quality metrics
	Entropy          float64
	HammingDistance  float64
	Autocorrelation  []float64

	// Output buffer
	OutputBuffer     []byte
	BufferPosition   int
}

// NewFerroTRNG creates a TRNG using multiple p-bits
func NewFerroTRNG(config *StochasticConfig, numPBits int) *FerroTRNG {
	trng := &FerroTRNG{
		Config:          config,
		PBits:           make([]*FerroelectricPBit, numPBits),
		OutputBuffer:    make([]byte, 0, 1024),
		Autocorrelation: make([]float64, 32),
	}

	for i := range trng.PBits {
		trng.PBits[i] = NewFerroelectricPBit(config)
		// Set each p-bit with slightly different bias for decorrelation
		trng.PBits[i].SetBias(0.5 + 0.1*rand.NormFloat64())
	}

	return trng
}

// GenerateRandomByte generates a random byte
func (trng *FerroTRNG) GenerateRandomByte() byte {
	var result byte
	for i := 0; i < 8; i++ {
		// XOR outputs from multiple p-bits for better randomness
		bit := false
		for _, pb := range trng.PBits {
			bit = bit != pb.Sample() // XOR
		}
		if bit {
			result |= 1 << i
		}
	}
	return result
}

// GenerateRandomBytes generates multiple random bytes
func (trng *FerroTRNG) GenerateRandomBytes(n int) []byte {
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = trng.GenerateRandomByte()
	}
	trng.OutputBuffer = append(trng.OutputBuffer, result...)
	return result
}

// CalculateEntropy calculates Shannon entropy of output
func (trng *FerroTRNG) CalculateEntropy() float64 {
	if len(trng.OutputBuffer) < 256 {
		// Generate more data
		trng.GenerateRandomBytes(256 - len(trng.OutputBuffer))
	}

	// Count byte frequencies
	counts := make([]int, 256)
	for _, b := range trng.OutputBuffer {
		counts[b]++
	}

	// Calculate entropy
	entropy := 0.0
	total := float64(len(trng.OutputBuffer))
	for _, c := range counts {
		if c > 0 {
			p := float64(c) / total
			entropy -= p * math.Log2(p)
		}
	}

	// Normalize to bits per byte (max = 8)
	trng.Entropy = entropy / 8.0
	return trng.Entropy
}

// CalculateHammingDistance calculates average Hamming distance between consecutive outputs
func (trng *FerroTRNG) CalculateHammingDistance() float64 {
	if len(trng.OutputBuffer) < 2 {
		return 0
	}

	totalDistance := 0
	for i := 1; i < len(trng.OutputBuffer); i++ {
		// Count differing bits
		xor := trng.OutputBuffer[i] ^ trng.OutputBuffer[i-1]
		for xor > 0 {
			totalDistance += int(xor & 1)
			xor >>= 1
		}
	}

	// Normalize to 0-1 range (8 bits max difference)
	trng.HammingDistance = float64(totalDistance) / float64((len(trng.OutputBuffer)-1)*8)
	return trng.HammingDistance
}

// CalculateAutocorrelation calculates autocorrelation function
func (trng *FerroTRNG) CalculateAutocorrelation() []float64 {
	if len(trng.OutputBuffer) < 64 {
		trng.GenerateRandomBytes(64 - len(trng.OutputBuffer))
	}

	// Convert bytes to centered values
	values := make([]float64, len(trng.OutputBuffer))
	mean := 0.0
	for _, b := range trng.OutputBuffer {
		mean += float64(b)
	}
	mean /= float64(len(trng.OutputBuffer))
	for i, b := range trng.OutputBuffer {
		values[i] = float64(b) - mean
	}

	// Calculate variance
	variance := 0.0
	for _, v := range values {
		variance += v * v
	}
	variance /= float64(len(values))

	// Calculate autocorrelation for different lags
	for lag := 0; lag < len(trng.Autocorrelation) && lag < len(values)-1; lag++ {
		sum := 0.0
		for i := 0; i < len(values)-lag; i++ {
			sum += values[i] * values[i+lag]
		}
		trng.Autocorrelation[lag] = sum / (float64(len(values)-lag) * variance)
	}

	return trng.Autocorrelation
}

// GetQualityReport returns TRNG quality metrics
func (trng *FerroTRNG) GetQualityReport() map[string]float64 {
	trng.CalculateEntropy()
	trng.CalculateHammingDistance()
	trng.CalculateAutocorrelation()

	return map[string]float64{
		"entropy":           trng.Entropy,
		"entropy_ideal":     1.0,
		"hamming_distance":  trng.HammingDistance,
		"hamming_ideal":     0.5,
		"autocorr_lag1":     trng.Autocorrelation[1],
		"autocorr_ideal":    0.0,
		"bytes_generated":   float64(len(trng.OutputBuffer)),
	}
}

// ============================================================================
// STOCHASTIC NEURAL NETWORK
// ============================================================================

// StochasticNeuralNet implements a neural network using stochastic computing
type StochasticNeuralNet struct {
	Config         *StochasticConfig
	Layers         []StochasticLayer
	BitstreamLen   int

	// Accuracy tracking
	TotalInferences int
	CorrectCount    int
}

// StochasticLayer represents a layer in stochastic neural network
type StochasticLayer struct {
	Weights    [][]float64 // Probability values for weights
	Biases     []float64
	InputSize  int
	OutputSize int
	PBits      [][]*FerroelectricPBit // P-bits for weight representation
}

// NewStochasticNeuralNet creates a stochastic neural network
func NewStochasticNeuralNet(config *StochasticConfig, layerSizes []int) *StochasticNeuralNet {
	snn := &StochasticNeuralNet{
		Config:       config,
		Layers:       make([]StochasticLayer, len(layerSizes)-1),
		BitstreamLen: config.BitstreamLength,
	}

	for i := 0; i < len(layerSizes)-1; i++ {
		inputSize := layerSizes[i]
		outputSize := layerSizes[i+1]

		layer := StochasticLayer{
			Weights:    make([][]float64, outputSize),
			Biases:     make([]float64, outputSize),
			InputSize:  inputSize,
			OutputSize: outputSize,
			PBits:      make([][]*FerroelectricPBit, outputSize),
		}

		// Initialize weights and p-bits
		for j := 0; j < outputSize; j++ {
			layer.Weights[j] = make([]float64, inputSize)
			layer.PBits[j] = make([]*FerroelectricPBit, inputSize)
			for k := 0; k < inputSize; k++ {
				// Xavier initialization
				layer.Weights[j][k] = 0.5 + 0.1*rand.NormFloat64()
				layer.PBits[j][k] = NewFerroelectricPBit(config)
				layer.PBits[j][k].SetBias(layer.Weights[j][k] - 0.5)
			}
			layer.Biases[j] = 0.5
		}

		snn.Layers[i] = layer
	}

	return snn
}

// ValueToBitstream converts a value [0,1] to a stochastic bitstream
func (snn *StochasticNeuralNet) ValueToBitstream(value float64) []bool {
	bitstream := make([]bool, snn.BitstreamLen)
	for i := 0; i < snn.BitstreamLen; i++ {
		bitstream[i] = rand.Float64() < value
	}
	return bitstream
}

// BitstreamToValue converts a bitstream back to a value
func (snn *StochasticNeuralNet) BitstreamToValue(bitstream []bool) float64 {
	count := 0
	for _, b := range bitstream {
		if b {
			count++
		}
	}
	return float64(count) / float64(len(bitstream))
}

// StochasticAND performs AND operation on bitstreams (multiplication)
func StochasticAND(a, b []bool) []bool {
	result := make([]bool, len(a))
	for i := range a {
		result[i] = a[i] && b[i]
	}
	return result
}

// StochasticMUX performs MUX operation (scaled addition)
func StochasticMUX(inputs [][]bool, select_ []bool) []bool {
	if len(inputs) == 0 {
		return nil
	}
	result := make([]bool, len(inputs[0]))
	for i := range result {
		selectIdx := 0
		// Use select bits to choose which input
		for j := 0; j < len(select_) && j < int(math.Log2(float64(len(inputs)))); j++ {
			if select_[i*len(select_)/len(inputs[0])+j] {
				selectIdx |= 1 << j
			}
		}
		if selectIdx < len(inputs) {
			result[i] = inputs[selectIdx][i]
		}
	}
	return result
}

// ForwardLayer performs forward pass for one layer using stochastic computing
func (snn *StochasticNeuralNet) ForwardLayer(layer *StochasticLayer, input []float64) []float64 {
	output := make([]float64, layer.OutputSize)

	// Convert inputs to bitstreams
	inputBitstreams := make([][]bool, len(input))
	for i, v := range input {
		inputBitstreams[i] = snn.ValueToBitstream(v)
	}

	// Compute each output neuron
	for j := 0; j < layer.OutputSize; j++ {
		// Accumulate weighted inputs using stochastic multiplication (AND)
		accumulated := make([]int, snn.BitstreamLen)

		for k := 0; k < layer.InputSize && k < len(input); k++ {
			// Generate weight bitstream from p-bit
			weightBitstream := layer.PBits[j][k].GenerateBitstream(snn.BitstreamLen)

			// Stochastic multiplication
			for i := 0; i < snn.BitstreamLen; i++ {
				if inputBitstreams[k][i] && weightBitstream[i] {
					accumulated[i]++
				}
			}
		}

		// Convert accumulated count to probability (scaled sum)
		sum := 0.0
		for _, a := range accumulated {
			sum += float64(a)
		}
		output[j] = sum / float64(snn.BitstreamLen*layer.InputSize)

		// Add bias
		output[j] += layer.Biases[j] - 0.5

		// Stochastic ReLU (threshold)
		if output[j] < 0 {
			output[j] = 0
		}
		if output[j] > 1 {
			output[j] = 1
		}
	}

	return output
}

// Forward performs complete forward pass
func (snn *StochasticNeuralNet) Forward(input []float64) []float64 {
	current := input
	for i := range snn.Layers {
		current = snn.ForwardLayer(&snn.Layers[i], current)
	}
	snn.TotalInferences++
	return current
}

// GetAccuracy returns classification accuracy
func (snn *StochasticNeuralNet) GetAccuracy() float64 {
	if snn.TotalInferences == 0 {
		return 0
	}
	return float64(snn.CorrectCount) / float64(snn.TotalInferences)
}

// ============================================================================
// INTEGRATED MEMCAPACITOR + STOCHASTIC SYSTEM
// ============================================================================

// IronLatticeMemcapStochastic combines memcapacitor CIM with stochastic computing
type IronLatticeMemcapStochastic struct {
	MemcapConfig    *MemcapConfig
	StochasticConfig *StochasticConfig

	// Components
	MemcapArray     *MemcapCrossbar
	StochasticNet   *StochasticNeuralNet
	TRNG            *FerroTRNG

	// Hybrid mode
	UseStochastic   bool
	HybridThreshold float64 // Switch to stochastic for low-precision needs

	// Performance tracking
	DeterministicOps int
	StochasticOps    int
	TotalEnergyFJ    float64
}

// NewIronLatticeMemcapStochastic creates an integrated system
func NewIronLatticeMemcapStochastic(arraySize int, layerSizes []int) *IronLatticeMemcapStochastic {
	memcapCfg := DefaultMemcapConfig()
	stochasticCfg := DefaultStochasticConfig()

	system := &IronLatticeMemcapStochastic{
		MemcapConfig:     memcapCfg,
		StochasticConfig: stochasticCfg,
		HybridThreshold:  0.1, // Switch to stochastic when precision <10% matters
	}

	// Create components
	system.MemcapArray = NewMemcapCrossbar(memcapCfg, arraySize, arraySize)
	system.StochasticNet = NewStochasticNeuralNet(stochasticCfg, layerSizes)
	system.TRNG = NewFerroTRNG(stochasticCfg, 8)

	return system
}

// Inference performs inference using hybrid approach
func (sys *IronLatticeMemcapStochastic) Inference(input []float64, precisionRequired float64) []float64 {
	if precisionRequired > sys.HybridThreshold {
		// Use deterministic memcapacitor computing for high precision
		sys.DeterministicOps++
		result := sys.MemcapArray.ComputeMAC(input)
		sys.TotalEnergyFJ += sys.MemcapArray.TotalEnergy
		return result
	} else {
		// Use stochastic computing for low precision (more energy efficient)
		sys.StochasticOps++
		// Normalize input to [0, 1]
		normalizedInput := make([]float64, len(input))
		maxVal := 0.0
		for _, v := range input {
			if math.Abs(v) > maxVal {
				maxVal = math.Abs(v)
			}
		}
		if maxVal > 0 {
			for i, v := range input {
				normalizedInput[i] = (v/maxVal + 1) / 2
			}
		}
		result := sys.StochasticNet.Forward(normalizedInput)
		// Stochastic is ~10x more energy efficient for low precision
		sys.TotalEnergyFJ += float64(len(input)) * sys.MemcapConfig.EnergyFJ * 0.1
		return result
	}
}

// GenerateRandomWeights uses TRNG for weight initialization
func (sys *IronLatticeMemcapStochastic) GenerateRandomWeights(rows, cols int) [][]float64 {
	weights := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		weights[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			// Use TRNG for random initialization
			randBytes := sys.TRNG.GenerateRandomBytes(4)
			// Convert to float in [-1, 1]
			randInt := int32(randBytes[0]) | int32(randBytes[1])<<8 |
				int32(randBytes[2])<<16 | int32(randBytes[3])<<24
			weights[i][j] = float64(randInt) / float64(math.MaxInt32)
		}
	}
	return weights
}

// GetPerformanceReport returns system performance metrics
func (sys *IronLatticeMemcapStochastic) GetPerformanceReport() map[string]interface{} {
	totalOps := sys.DeterministicOps + sys.StochasticOps
	if totalOps == 0 {
		totalOps = 1
	}

	return map[string]interface{}{
		"deterministic_ops":    sys.DeterministicOps,
		"stochastic_ops":       sys.StochasticOps,
		"stochastic_ratio":     float64(sys.StochasticOps) / float64(totalOps),
		"total_energy_fJ":      sys.TotalEnergyFJ,
		"energy_per_op_fJ":     sys.TotalEnergyFJ / float64(totalOps),
		"memcap_linearity":     sys.MemcapArray.GetLinearity(),
		"trng_entropy":         sys.TRNG.CalculateEntropy(),
		"trng_hamming":         sys.TRNG.CalculateHammingDistance(),
	}
}

// ============================================================================
// VISUALIZATION HELPERS
// ============================================================================

// GenerateMemcapCharacteristics creates ASCII visualization of memcapacitor C-V curve
func GenerateMemcapCharacteristics(config *MemcapConfig) string {
	plot := "Ferroelectric Memcapacitor C-V Characteristics\n"
	plot += "═══════════════════════════════════════════════\n\n"

	height := 12
	width := 40

	// Generate hysteresis data
	voltages := make([]float64, width*2)
	capacitances := make([]float64, width*2)

	// Forward sweep
	for i := 0; i < width; i++ {
		v := -config.SaturationVoltageV + 2*config.SaturationVoltageV*float64(i)/float64(width-1)
		voltages[i] = v

		// Hysteresis model
		if v < -config.CoerciveVoltageV {
			capacitances[i] = config.CminFPerUm2
		} else if v > config.CoerciveVoltageV {
			capacitances[i] = config.CmaxFPerUm2
		} else {
			// Transition region
			fraction := (v + config.CoerciveVoltageV) / (2 * config.CoerciveVoltageV)
			capacitances[i] = config.CminFPerUm2 + fraction*(config.CmaxFPerUm2-config.CminFPerUm2)
		}
	}

	// Backward sweep (hysteresis)
	for i := 0; i < width; i++ {
		v := config.SaturationVoltageV - 2*config.SaturationVoltageV*float64(i)/float64(width-1)
		voltages[width+i] = v

		if v > config.CoerciveVoltageV {
			capacitances[width+i] = config.CmaxFPerUm2
		} else if v < -config.CoerciveVoltageV {
			capacitances[width+i] = config.CminFPerUm2
		} else {
			fraction := (v + config.CoerciveVoltageV) / (2 * config.CoerciveVoltageV)
			capacitances[width+i] = config.CminFPerUm2 + (1-fraction)*(config.CmaxFPerUm2-config.CminFPerUm2)
		}
	}

	// Draw plot
	for row := height; row >= 0; row-- {
		threshold := config.CminFPerUm2 + (config.CmaxFPerUm2-config.CminFPerUm2)*float64(row)/float64(height)
		line := fmt.Sprintf("%5.1f │", threshold)

		for i := 0; i < width; i++ {
			// Check both forward and backward sweeps
			forwardHit := capacitances[i] >= threshold && (i == 0 || capacitances[i-1] < threshold ||
				capacitances[i] == threshold)
			backwardHit := capacitances[width+i] >= threshold && (i == 0 ||
				capacitances[width+i-1] < threshold || capacitances[width+i] == threshold)

			if forwardHit && backwardHit {
				line += "█"
			} else if forwardHit {
				line += "▀"
			} else if backwardHit {
				line += "▄"
			} else if row == height/2 {
				line += "·"
			} else {
				line += " "
			}
		}
		plot += line + "\n"
	}

	// X-axis
	plot += "      └" + repeatChar("─", width) + "\n"
	plot += fmt.Sprintf("      %.1fV              0V              +%.1fV\n",
		-config.SaturationVoltageV, config.SaturationVoltageV)
	plot += "                  Voltage (V)\n\n"
	plot += fmt.Sprintf("Memory Window: %.1f fF/µm² (%.1f - %.1f)\n",
		config.MemoryWindowFPerUm2, config.CminFPerUm2, config.CmaxFPerUm2)
	plot += fmt.Sprintf("Legend: ▀ Forward sweep  ▄ Backward sweep\n")

	return plot
}

// GeneratePBitDistribution creates ASCII visualization of p-bit probability distribution
func GeneratePBitDistribution(pbit *FerroelectricPBit, numSamples int) string {
	plot := "Ferroelectric P-Bit Output Distribution\n"
	plot += "═══════════════════════════════════════════════\n\n"

	// Generate samples
	samples := pbit.GenerateBitstream(numSamples)

	// Count ones in windows
	windowSize := numSamples / 20
	histogram := make([]int, 20)

	for i := 0; i < 20; i++ {
		ones := 0
		for j := i * windowSize; j < (i+1)*windowSize && j < numSamples; j++ {
			if samples[j] {
				ones++
			}
		}
		histogram[i] = ones
	}

	// Find max for scaling
	maxCount := 0
	for _, c := range histogram {
		if c > maxCount {
			maxCount = c
		}
	}

	// Draw histogram
	height := 8
	for row := height; row > 0; row-- {
		threshold := maxCount * row / height
		line := fmt.Sprintf("%4d │", threshold)
		for _, c := range histogram {
			if c >= threshold {
				line += "██"
			} else {
				line += "  "
			}
		}
		plot += line + "\n"
	}

	plot += "     └" + repeatChar("─", 40) + "\n"
	plot += "      Sample Windows (time →)\n\n"

	// Calculate statistics
	totalOnes := 0
	for _, s := range samples {
		if s {
			totalOnes++
		}
	}
	actualProb := float64(totalOnes) / float64(numSamples)

	plot += fmt.Sprintf("Target Probability: %.3f\n", pbit.CurrentProbability)
	plot += fmt.Sprintf("Actual Probability: %.3f\n", actualProb)
	plot += fmt.Sprintf("Samples: %d, Transitions: %d\n", numSamples, pbit.TransitionCount)

	return plot
}

// GenerateStochasticComparison creates ASCII comparison of deterministic vs stochastic
func GenerateStochasticComparison() string {
	comparison := "Deterministic vs Stochastic Computing Comparison\n"
	comparison += "═══════════════════════════════════════════════════════\n\n"

	comparison += "┌─────────────────────────────────────────────────────────────┐\n"
	comparison += "│                    DETERMINISTIC (Memcapacitor)             │\n"
	comparison += "├─────────────────────────────────────────────────────────────┤\n"
	comparison += "│                                                             │\n"
	comparison += "│   Input ──→ ┌───┐     ┌───────────┐     ┌───┐ ──→ Output   │\n"
	comparison += "│   Voltage   │DAC│ ──→ │  Q = C×V  │ ──→ │ADC│    Digital   │\n"
	comparison += "│             └───┘     │  FeCap    │     └───┘              │\n"
	comparison += "│                       └───────────┘                        │\n"
	comparison += "│                                                             │\n"
	comparison += "│   Precision: 8-bit          Energy: 31 fJ/op               │\n"
	comparison += "│   Latency: 10 ns            Error: <1%                     │\n"
	comparison += "│                                                             │\n"
	comparison += "└─────────────────────────────────────────────────────────────┘\n"
	comparison += "\n"
	comparison += "┌─────────────────────────────────────────────────────────────┐\n"
	comparison += "│                    STOCHASTIC (FeFET P-Bits)                │\n"
	comparison += "├─────────────────────────────────────────────────────────────┤\n"
	comparison += "│                                                             │\n"
	comparison += "│   Value ──→ ┌─────────┐   ┌─────┐   ┌─────────┐ ──→ Value  │\n"
	comparison += "│   (0.7)     │Bitstream│──→│ AND │──→│ Counter │    (0.49)  │\n"
	comparison += "│             │Generator│   │(×)  │   │         │            │\n"
	comparison += "│             └─────────┘   └─────┘   └─────────┘            │\n"
	comparison += "│                                                             │\n"
	comparison += "│   Bitstream: 1011101010... × 0111010011... = 0011000010... │\n"
	comparison += "│   P(1)=0.7                   P(1)=0.7        P(1)=0.49     │\n"
	comparison += "│                                                             │\n"
	comparison += "│   Precision: ~1/√N bits     Energy: 3 fJ/op                │\n"
	comparison += "│   Latency: N cycles         Error: 1/√N                    │\n"
	comparison += "│                                                             │\n"
	comparison += "└─────────────────────────────────────────────────────────────┘\n"
	comparison += "\n"
	comparison += "Trade-off: Stochastic is 10× more energy efficient but needs\n"
	comparison += "longer bitstreams for high precision. Best for fault-tolerant\n"
	comparison += "applications like neural networks where ~5% error is acceptable.\n"

	return comparison
}

// ============================================================================
// EXAMPLE USAGE AND DEMO
// ============================================================================

// RunMemcapStochasticDemo demonstrates memcapacitor and stochastic computing
func RunMemcapStochasticDemo() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║  IronLattice Memcapacitor & Stochastic Computing Demo     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 1. Memcapacitor characteristics
	fmt.Println("1. Ferroelectric Memcapacitor Characteristics:")
	memcapCfg := DefaultMemcapConfig()
	fmt.Println(GenerateMemcapCharacteristics(memcapCfg))
	fmt.Println()

	// 2. Single memcapacitor test
	fmt.Println("2. Single Memcapacitor Programming Test:")
	mc := NewFerroMemcapacitor(memcapCfg)
	fmt.Printf("   Initial capacitance: %.2f fF\n", mc.Capacitance)

	levels := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	fmt.Println("   Programming levels:")
	for _, level := range levels {
		mc.Program(level, memcapCfg.SaturationVoltageV)
		fmt.Printf("   Target: %.2f → Actual: %.2f (C = %.2f fF)\n",
			level, mc.PolarizationState, mc.Capacitance)
	}
	fmt.Println()

	// 3. Memcapacitor crossbar
	fmt.Println("3. Memcapacitor Crossbar Array (8×8):")
	crossbar := NewMemcapCrossbar(memcapCfg, 8, 8)

	// Program with random weights
	weights := make([][]float64, 8)
	for i := 0; i < 8; i++ {
		weights[i] = make([]float64, 8)
		for j := 0; j < 8; j++ {
			weights[i][j] = rand.Float64()*2 - 1 // [-1, 1]
		}
	}
	crossbar.ProgramWeights(weights)

	// Test MAC
	input := make([]float64, 8)
	for i := range input {
		input[i] = rand.Float64()
	}
	output := crossbar.ComputeMAC(input)
	fmt.Printf("   Input: [%.2f, %.2f, %.2f, ...]\n", input[0], input[1], input[2])
	fmt.Printf("   Output: [%.3f, %.3f, %.3f, ...]\n", output[0], output[1], output[2])
	fmt.Printf("   Linearity R²: %.3f\n", crossbar.GetLinearity())
	fmt.Printf("   Energy consumed: %.2f fJ\n", crossbar.TotalEnergy)
	fmt.Println()

	// 4. Stochastic P-Bit
	fmt.Println("4. Ferroelectric Probabilistic Bit (P-Bit):")
	stochCfg := DefaultStochasticConfig()
	pbit := NewFerroelectricPBit(stochCfg)
	pbit.SetBias(0.3) // Bias towards 0.7 probability

	fmt.Println(GeneratePBitDistribution(pbit, 1000))
	fmt.Println()

	// 5. TRNG quality
	fmt.Println("5. True Random Number Generator Quality:")
	trng := NewFerroTRNG(stochCfg, 8)
	trng.GenerateRandomBytes(1000)
	report := trng.GetQualityReport()

	fmt.Printf("   Entropy: %.4f (ideal: 1.0)\n", report["entropy"])
	fmt.Printf("   Hamming Distance: %.4f (ideal: 0.5)\n", report["hamming_distance"])
	fmt.Printf("   Autocorrelation(1): %.4f (ideal: 0.0)\n", report["autocorr_lag1"])
	fmt.Printf("   Bytes generated: %.0f\n", report["bytes_generated"])
	fmt.Println()

	// 6. Stochastic Neural Network
	fmt.Println("6. Stochastic Neural Network (4→8→4):")
	snn := NewStochasticNeuralNet(stochCfg, []int{4, 8, 4})

	testInput := []float64{0.3, 0.7, 0.5, 0.9}
	snnOutput := snn.Forward(testInput)
	fmt.Printf("   Input: [%.2f, %.2f, %.2f, %.2f]\n",
		testInput[0], testInput[1], testInput[2], testInput[3])
	fmt.Printf("   Output: [%.3f, %.3f, %.3f, %.3f]\n",
		snnOutput[0], snnOutput[1], snnOutput[2], snnOutput[3])
	fmt.Printf("   Bitstream length: %d\n", stochCfg.BitstreamLength)
	fmt.Println()

	// 7. Comparison
	fmt.Println("7. Computing Paradigm Comparison:")
	fmt.Println(GenerateStochasticComparison())

	// 8. Integrated system
	fmt.Println("8. Integrated Hybrid System:")
	system := NewIronLatticeMemcapStochastic(16, []int{16, 32, 16})

	// Run mixed precision inferences
	for i := 0; i < 100; i++ {
		testIn := make([]float64, 16)
		for j := range testIn {
			testIn[j] = rand.Float64()
		}
		precision := rand.Float64() // Random precision requirement
		system.Inference(testIn, precision)
	}

	perfReport := system.GetPerformanceReport()
	fmt.Printf("   Deterministic ops: %d\n", perfReport["deterministic_ops"])
	fmt.Printf("   Stochastic ops: %d\n", perfReport["stochastic_ops"])
	fmt.Printf("   Stochastic ratio: %.1f%%\n", perfReport["stochastic_ratio"].(float64)*100)
	fmt.Printf("   Total energy: %.2f fJ\n", perfReport["total_energy_fJ"])
	fmt.Printf("   Energy per op: %.2f fJ\n", perfReport["energy_per_op_fJ"])
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("Memcapacitor + Stochastic computing simulation complete!")
}

// Helper for sorting float64 slices
type float64Slice []float64

func (s float64Slice) Len() int           { return len(s) }
func (s float64Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s float64Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func sortFloat64s(slice []float64) {
	sort.Sort(float64Slice(slice))
}

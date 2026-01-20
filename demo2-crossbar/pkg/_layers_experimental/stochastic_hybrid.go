// Package layers provides stochastic computing and analog-digital hybrid CIM
// simulation for ferroelectric-based neuromorphic systems.
//
// Based on research:
// - FeBiM: Ferroelectric Bayesian Inference (DAC 2024)
// - 2D Ferroelectric Hybrid CIM (Science Advances 2024)
// - Ferroelectric TRNG with near-ideal entropy
// - FELIX mixed-signal architecture (36.5 TOPS/W)
//
// References:
// - Communications Materials 2025, "Memristor noise for computation"
// - Science Advances 2024, "2D ferroelectric hybrid CIM"
// - Nature Communications 2025, "Ferroelectric NAND Bayesian NN"
// - DAC 2024, "FeBiM Bayesian inference engine"
package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// ============================================================================
// STOCHASTIC COMPUTING FUNDAMENTALS
// ============================================================================

// StochasticBitStream represents a stochastic bit stream encoding a probability
type StochasticBitStream struct {
	Length     int     // Number of bits in stream
	Value      float64 // Encoded probability [0, 1]
	Bits       []bool  // Actual bit stream
	Bipolar    bool    // Bipolar encoding [-1, 1] vs unipolar [0, 1]
}

// NewStochasticBitStream creates a stochastic bit stream encoding a value
func NewStochasticBitStream(value float64, length int, bipolar bool) *StochasticBitStream {
	sbs := &StochasticBitStream{
		Length:  length,
		Value:   value,
		Bits:    make([]bool, length),
		Bipolar: bipolar,
	}

	// Convert value to probability
	var prob float64
	if bipolar {
		// Bipolar: value in [-1, 1] -> prob in [0, 1]
		prob = (value + 1.0) / 2.0
	} else {
		prob = value
	}

	// Clamp probability
	if prob < 0 {
		prob = 0
	}
	if prob > 1 {
		prob = 1
	}

	// Generate bit stream
	for i := 0; i < length; i++ {
		sbs.Bits[i] = rand.Float64() < prob
	}

	return sbs
}

// Decode converts bit stream back to value
func (sbs *StochasticBitStream) Decode() float64 {
	count := 0
	for _, b := range sbs.Bits {
		if b {
			count++
		}
	}

	prob := float64(count) / float64(sbs.Length)

	if sbs.Bipolar {
		return 2.0*prob - 1.0
	}
	return prob
}

// StochasticMultiply performs multiplication using AND gate (unipolar)
// or XNOR gate (bipolar)
func StochasticMultiply(a, b *StochasticBitStream) *StochasticBitStream {
	if a.Length != b.Length || a.Bipolar != b.Bipolar {
		return nil
	}

	result := &StochasticBitStream{
		Length:  a.Length,
		Bits:    make([]bool, a.Length),
		Bipolar: a.Bipolar,
	}

	for i := 0; i < a.Length; i++ {
		if a.Bipolar {
			// XNOR for bipolar multiplication
			result.Bits[i] = a.Bits[i] == b.Bits[i]
		} else {
			// AND for unipolar multiplication
			result.Bits[i] = a.Bits[i] && b.Bits[i]
		}
	}

	result.Value = result.Decode()
	return result
}

// StochasticAdd performs scaled addition using MUX
func StochasticAdd(streams []*StochasticBitStream) *StochasticBitStream {
	if len(streams) == 0 {
		return nil
	}

	length := streams[0].Length
	bipolar := streams[0].Bipolar

	result := &StochasticBitStream{
		Length:  length,
		Bits:    make([]bool, length),
		Bipolar: bipolar,
	}

	// MUX-based addition: randomly select from input streams
	for i := 0; i < length; i++ {
		idx := rand.Intn(len(streams))
		result.Bits[i] = streams[idx].Bits[i]
	}

	result.Value = result.Decode()
	return result
}

// ============================================================================
// TRUE RANDOM NUMBER GENERATOR (TRNG)
// ============================================================================

// TRNGConfig configures ferroelectric TRNG
type TRNGConfig struct {
	EntropySource    string  // "ferroelectric_switching", "charge_trapping", "domain_noise"
	BitRate          float64 // bits per second
	TargetEntropy    float64 // Target entropy (1.0 ideal)
	TargetHamming    float64 // Target Hamming distance (0.5 ideal)
	TemperatureRange []float64 // Min, max operating temperature (K)
}

// DefaultTRNGConfig returns typical ferroelectric TRNG configuration
func DefaultTRNGConfig() *TRNGConfig {
	return &TRNGConfig{
		EntropySource:    "ferroelectric_switching",
		BitRate:          1e6, // 1 Mbit/s
		TargetEntropy:    1.0,
		TargetHamming:    0.5,
		TemperatureRange: []float64{250.0, 400.0}, // Wide temperature range
	}
}

// FerroelectricTRNG implements true random number generator using ferroelectric stochasticity
type FerroelectricTRNG struct {
	Config          *TRNGConfig

	// Device parameters
	SwitchingProb   float64 // Base switching probability
	NoiseLevel      float64 // Intrinsic noise level
	TrapDensity     float64 // Charge trap density affecting stochasticity

	// Quality metrics
	MeasuredEntropy float64
	HammingDistance float64
	Autocorrelation float64
	PassedNIST      bool    // NIST randomness tests
}

// NewFerroelectricTRNG creates a ferroelectric-based TRNG
func NewFerroelectricTRNG(config *TRNGConfig) *FerroelectricTRNG {
	if config == nil {
		config = DefaultTRNGConfig()
	}

	trng := &FerroelectricTRNG{
		Config:        config,
		SwitchingProb: 0.5,
		NoiseLevel:    0.05, // 5% intrinsic noise
		TrapDensity:   1e12, // traps/cm²
	}

	// Calibrate to achieve target entropy
	trng.calibrate()

	return trng
}

// calibrate adjusts parameters to achieve target randomness
func (trng *FerroelectricTRNG) calibrate() {
	// Adjust switching probability based on noise level
	// Higher noise → more stochasticity → closer to ideal 0.5 probability
	trng.SwitchingProb = 0.5 + (rand.Float64()-0.5)*trng.NoiseLevel

	// Measure entropy
	testBits := trng.GenerateBits(10000)
	trng.MeasuredEntropy = trng.calculateEntropy(testBits)
	trng.HammingDistance = trng.calculateHammingDistance(testBits)
	trng.Autocorrelation = trng.calculateAutocorrelation(testBits)

	// Check if passes quality threshold
	trng.PassedNIST = trng.MeasuredEntropy > 0.99 &&
		math.Abs(trng.HammingDistance-0.5) < 0.02 &&
		math.Abs(trng.Autocorrelation) < 0.02
}

// GenerateBits generates random bits using ferroelectric stochasticity
func (trng *FerroelectricTRNG) GenerateBits(count int) []bool {
	bits := make([]bool, count)

	for i := 0; i < count; i++ {
		// Simulate ferroelectric switching with stochasticity
		// Based on domain nucleation randomness
		noise := (rand.Float64() - 0.5) * trng.NoiseLevel * 2
		effectiveProb := trng.SwitchingProb + noise

		// Add charge trapping contribution
		trapNoise := (rand.Float64() - 0.5) * 0.01
		effectiveProb += trapNoise

		bits[i] = rand.Float64() < effectiveProb
	}

	return bits
}

// GenerateFloat generates random float in [0, 1]
func (trng *FerroelectricTRNG) GenerateFloat() float64 {
	// Use 32 bits for float generation
	bits := trng.GenerateBits(32)
	var value uint32
	for i, b := range bits {
		if b {
			value |= 1 << uint(31-i)
		}
	}
	return float64(value) / float64(math.MaxUint32)
}

// calculateEntropy computes Shannon entropy of bit sequence
func (trng *FerroelectricTRNG) calculateEntropy(bits []bool) float64 {
	ones := 0
	for _, b := range bits {
		if b {
			ones++
		}
	}

	p1 := float64(ones) / float64(len(bits))
	p0 := 1.0 - p1

	if p0 <= 0 || p1 <= 0 {
		return 0
	}

	return -p0*math.Log2(p0) - p1*math.Log2(p1)
}

// calculateHammingDistance computes average Hamming distance
func (trng *FerroelectricTRNG) calculateHammingDistance(bits []bool) float64 {
	changes := 0
	for i := 1; i < len(bits); i++ {
		if bits[i] != bits[i-1] {
			changes++
		}
	}
	return float64(changes) / float64(len(bits)-1)
}

// calculateAutocorrelation computes autocorrelation at lag 1
func (trng *FerroelectricTRNG) calculateAutocorrelation(bits []bool) float64 {
	n := len(bits)
	mean := 0.0
	for _, b := range bits {
		if b {
			mean += 1.0
		}
	}
	mean /= float64(n)

	var num, den float64
	for i := 0; i < n-1; i++ {
		x := 0.0
		if bits[i] {
			x = 1.0
		}
		y := 0.0
		if bits[i+1] {
			y = 1.0
		}
		num += (x - mean) * (y - mean)
		den += (x - mean) * (x - mean)
	}

	if den == 0 {
		return 0
	}
	return num / den
}

// GetQualityMetrics returns TRNG quality metrics
func (trng *FerroelectricTRNG) GetQualityMetrics() map[string]interface{} {
	return map[string]interface{}{
		"entropy":          trng.MeasuredEntropy,
		"hamming_distance": trng.HammingDistance,
		"autocorrelation":  trng.Autocorrelation,
		"passed_nist":      trng.PassedNIST,
		"bit_rate":         trng.Config.BitRate,
	}
}

// ============================================================================
// BAYESIAN INFERENCE WITH FERROELECTRIC IMC
// ============================================================================

// BayesianVariable represents a random variable in Bayesian network
type BayesianVariable struct {
	Name         string
	NumStates    int       // Number of discrete states
	Prior        []float64 // Prior probability distribution
	Observed     int       // Observed state (-1 if not observed)
}

// ConditionalProbTable represents P(child | parents)
type ConditionalProbTable struct {
	Child       string
	Parents     []string
	Table       []float64 // Flattened CPT
	Dimensions  []int     // Dimension sizes
}

// BayesianNetwork represents a Bayesian network for inference
type BayesianNetwork struct {
	Variables map[string]*BayesianVariable
	CPTs      map[string]*ConditionalProbTable
	Topology  []string // Topological order
}

// NewBayesianNetwork creates a new Bayesian network
func NewBayesianNetwork() *BayesianNetwork {
	return &BayesianNetwork{
		Variables: make(map[string]*BayesianVariable),
		CPTs:      make(map[string]*ConditionalProbTable),
		Topology:  make([]string, 0),
	}
}

// AddVariable adds a variable to the network
func (bn *BayesianNetwork) AddVariable(name string, numStates int, prior []float64) {
	bn.Variables[name] = &BayesianVariable{
		Name:      name,
		NumStates: numStates,
		Prior:     prior,
		Observed:  -1,
	}
	bn.Topology = append(bn.Topology, name)
}

// SetEvidence sets observed value for a variable
func (bn *BayesianNetwork) SetEvidence(name string, state int) {
	if v, exists := bn.Variables[name]; exists {
		v.Observed = state
	}
}

// FeBiMConfig configures FeFET-based Bayesian inference engine
type FeBiMConfig struct {
	BitPrecision     int     // Bits for probability encoding
	ArraySize        int     // FeFET array size
	StochasticLength int     // Stochastic bit stream length
	EnergyPerOp      float64 // fJ per operation
}

// DefaultFeBiMConfig returns default FeBiM configuration
func DefaultFeBiMConfig() *FeBiMConfig {
	return &FeBiMConfig{
		BitPrecision:     4,
		ArraySize:        64,
		StochasticLength: 256,
		EnergyPerOp:      0.1,
	}
}

// FeBiMEngine implements ferroelectric Bayesian in-memory computing
type FeBiMEngine struct {
	Config  *FeBiMConfig
	Network *BayesianNetwork
	TRNG    *FerroelectricTRNG

	// Stored CPTs in FeFET array
	StoredTables map[string][][]float64

	// Performance metrics
	InferenceCount int64
	TotalEnergy    float64 // fJ
}

// NewFeBiMEngine creates a FeBiM inference engine
func NewFeBiMEngine(config *FeBiMConfig) *FeBiMEngine {
	if config == nil {
		config = DefaultFeBiMConfig()
	}

	return &FeBiMEngine{
		Config:       config,
		Network:      NewBayesianNetwork(),
		TRNG:         NewFerroelectricTRNG(nil),
		StoredTables: make(map[string][][]float64),
	}
}

// ProgramCPT programs a conditional probability table into FeFET array
func (fe *FeBiMEngine) ProgramCPT(varName string, cpt [][]float64) {
	fe.StoredTables[varName] = cpt
}

// StochasticInference performs Bayesian inference using stochastic computing
func (fe *FeBiMEngine) StochasticInference(query string) []float64 {
	v := fe.Network.Variables[query]
	if v == nil {
		return nil
	}

	result := make([]float64, v.NumStates)

	// Generate stochastic bit streams for prior
	priorStreams := make([]*StochasticBitStream, v.NumStates)
	for i, p := range v.Prior {
		priorStreams[i] = NewStochasticBitStream(p, fe.Config.StochasticLength, false)
	}

	// Incorporate evidence using stochastic multiplication
	if cpt, exists := fe.StoredTables[query]; exists {
		for state := 0; state < v.NumStates; state++ {
			// Multiply prior with likelihood from CPT
			if state < len(cpt) {
				likelihoodStream := NewStochasticBitStream(cpt[state][0], fe.Config.StochasticLength, false)
				posteriorStream := StochasticMultiply(priorStreams[state], likelihoodStream)
				result[state] = posteriorStream.Decode()
			} else {
				result[state] = priorStreams[state].Decode()
			}
		}
	} else {
		for state := 0; state < v.NumStates; state++ {
			result[state] = priorStreams[state].Decode()
		}
	}

	// Normalize
	sum := 0.0
	for _, p := range result {
		sum += p
	}
	if sum > 0 {
		for i := range result {
			result[i] /= sum
		}
	}

	// Update metrics
	fe.InferenceCount++
	fe.TotalEnergy += float64(v.NumStates*fe.Config.StochasticLength) * fe.Config.EnergyPerOp

	return result
}

// MullerCElementInference implements Muller C-element based Bayesian inference
func (fe *FeBiMEngine) MullerCElementInference(inputs []float64) float64 {
	// Muller C-element: output changes when all inputs agree
	// Used for cascaded Bayesian inference

	// Convert inputs to stochastic streams
	streams := make([]*StochasticBitStream, len(inputs))
	for i, p := range inputs {
		streams[i] = NewStochasticBitStream(p, fe.Config.StochasticLength, false)
	}

	// C-element operation: AND of all inputs
	result := make([]bool, fe.Config.StochasticLength)
	for i := 0; i < fe.Config.StochasticLength; i++ {
		allTrue := true
		for _, s := range streams {
			if !s.Bits[i] {
				allTrue = false
				break
			}
		}
		result[i] = allTrue
	}

	// Decode result
	count := 0
	for _, b := range result {
		if b {
			count++
		}
	}
	return float64(count) / float64(fe.Config.StochasticLength)
}

// GetMetrics returns FeBiM performance metrics
func (fe *FeBiMEngine) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"inference_count":    fe.InferenceCount,
		"total_energy_fJ":    fe.TotalEnergy,
		"avg_energy_per_inf": fe.TotalEnergy / float64(fe.InferenceCount+1),
		"bit_precision":      fe.Config.BitPrecision,
		"array_size":         fe.Config.ArraySize,
		"trng_quality":       fe.TRNG.GetQualityMetrics(),
	}
}

// ============================================================================
// ANALOG-DIGITAL HYBRID CIM ARCHITECTURE
// ============================================================================

// ComputeMode defines the compute mode
type ComputeMode int

const (
	AnalogMode  ComputeMode = iota // Analog multiply-accumulate
	DigitalMode                     // Digital bit-serial computation
	HybridMode                      // Mixed analog-digital
)

// HybridCIMConfig configures hybrid analog-digital CIM
type HybridCIMConfig struct {
	AnalogArraySize   int       // Analog crossbar size
	DigitalArraySize  int       // Digital SRAM/FeFET array size
	AnalogPrecision   int       // Bits for analog computation
	DigitalPrecision  int       // Bits for digital computation
	SwitchThreshold   float64   // Precision threshold for mode switching
	C2CVariation      float64   // Cycle-to-cycle variation (%)
	D2DVariation      float64   // Device-to-device variation (%)
}

// DefaultHybridCIMConfig returns 2D ferroelectric hybrid configuration
func DefaultHybridCIMConfig() *HybridCIMConfig {
	return &HybridCIMConfig{
		AnalogArraySize:  64,
		DigitalArraySize: 64,
		AnalogPrecision:  6,
		DigitalPrecision: 8,
		SwitchThreshold:  0.01, // Switch to digital if error > 1%
		C2CVariation:     0.3,  // 0.3% (2D ferroelectric)
		D2DVariation:     0.5,  // 0.5% (2D ferroelectric)
	}
}

// AnalogComputeUnit implements analog MVM
type AnalogComputeUnit struct {
	Rows         int
	Cols         int
	Weights      [][]float64
	Precision    int
	Variation    float64
}

// NewAnalogComputeUnit creates an analog compute unit
func NewAnalogComputeUnit(rows, cols, precision int, variation float64) *AnalogComputeUnit {
	acu := &AnalogComputeUnit{
		Rows:      rows,
		Cols:      cols,
		Weights:   make([][]float64, rows),
		Precision: precision,
		Variation: variation,
	}

	for i := 0; i < rows; i++ {
		acu.Weights[i] = make([]float64, cols)
	}

	return acu
}

// Program programs weights with quantization
func (acu *AnalogComputeUnit) Program(weights [][]float64) error {
	if len(weights) != acu.Rows {
		return fmt.Errorf("row count mismatch")
	}

	levels := 1 << acu.Precision
	for i, row := range weights {
		if len(row) != acu.Cols {
			return fmt.Errorf("column count mismatch at row %d", i)
		}
		for j, w := range row {
			// Quantize
			quantized := math.Round(w*float64(levels-1)) / float64(levels-1)
			// Add D2D variation
			variation := 1.0 + (rand.Float64()-0.5)*acu.Variation/100.0*2
			acu.Weights[i][j] = quantized * variation
		}
	}

	return nil
}

// Compute performs analog MVM with noise
func (acu *AnalogComputeUnit) Compute(input []float64) ([]float64, error) {
	if len(input) != acu.Cols {
		return nil, fmt.Errorf("input size mismatch")
	}

	output := make([]float64, acu.Rows)

	for i := 0; i < acu.Rows; i++ {
		var sum float64
		for j := 0; j < acu.Cols; j++ {
			// Add C2C variation to each operation
			variation := 1.0 + (rand.Float64()-0.5)*acu.Variation/100.0*2
			sum += acu.Weights[i][j] * input[j] * variation
		}
		output[i] = sum
	}

	return output, nil
}

// DigitalComputeUnit implements digital bit-serial computation
type DigitalComputeUnit struct {
	Rows      int
	Cols      int
	Weights   [][]int64 // Integer weights
	Precision int
}

// NewDigitalComputeUnit creates a digital compute unit
func NewDigitalComputeUnit(rows, cols, precision int) *DigitalComputeUnit {
	dcu := &DigitalComputeUnit{
		Rows:      rows,
		Cols:      cols,
		Weights:   make([][]int64, rows),
		Precision: precision,
	}

	for i := 0; i < rows; i++ {
		dcu.Weights[i] = make([]int64, cols)
	}

	return dcu
}

// Program programs weights as integers
func (dcu *DigitalComputeUnit) Program(weights [][]float64) error {
	if len(weights) != dcu.Rows {
		return fmt.Errorf("row count mismatch")
	}

	maxVal := int64(1 << (dcu.Precision - 1))
	for i, row := range weights {
		if len(row) != dcu.Cols {
			return fmt.Errorf("column count mismatch at row %d", i)
		}
		for j, w := range row {
			// Quantize to integer
			dcu.Weights[i][j] = int64(math.Round(w * float64(maxVal)))
			if dcu.Weights[i][j] > maxVal-1 {
				dcu.Weights[i][j] = maxVal - 1
			}
			if dcu.Weights[i][j] < -maxVal {
				dcu.Weights[i][j] = -maxVal
			}
		}
	}

	return nil
}

// Compute performs digital MVM (exact)
func (dcu *DigitalComputeUnit) Compute(input []int64) ([]int64, error) {
	if len(input) != dcu.Cols {
		return nil, fmt.Errorf("input size mismatch")
	}

	output := make([]int64, dcu.Rows)

	for i := 0; i < dcu.Rows; i++ {
		var sum int64
		for j := 0; j < dcu.Cols; j++ {
			sum += dcu.Weights[i][j] * input[j]
		}
		output[i] = sum
	}

	return output, nil
}

// HybridCIMArray implements hybrid analog-digital CIM
type HybridCIMArray struct {
	Config       *HybridCIMConfig
	AnalogUnit   *AnalogComputeUnit
	DigitalUnit  *DigitalComputeUnit
	CurrentMode  ComputeMode

	// Boolean logic unit for digital processing
	BooleanLogic *BooleanLogicUnit

	// Performance metrics
	AnalogOps    int64
	DigitalOps   int64
	TotalEnergy  float64 // pJ
}

// BooleanLogicUnit implements digital Boolean operations
type BooleanLogicUnit struct {
	NumGates  int
	GateTypes []string // "AND", "OR", "XOR", "NOT"
}

// NewBooleanLogicUnit creates a Boolean logic unit
func NewBooleanLogicUnit(numGates int) *BooleanLogicUnit {
	return &BooleanLogicUnit{
		NumGates:  numGates,
		GateTypes: []string{"AND", "OR", "XOR", "NOT"},
	}
}

// Execute executes Boolean operation
func (blu *BooleanLogicUnit) Execute(op string, inputs []bool) bool {
	switch op {
	case "AND":
		result := true
		for _, in := range inputs {
			result = result && in
		}
		return result
	case "OR":
		result := false
		for _, in := range inputs {
			result = result || in
		}
		return result
	case "XOR":
		result := false
		for _, in := range inputs {
			result = result != in
		}
		return result
	case "NOT":
		if len(inputs) > 0 {
			return !inputs[0]
		}
		return false
	default:
		return false
	}
}

// NewHybridCIMArray creates a hybrid analog-digital CIM array
func NewHybridCIMArray(config *HybridCIMConfig) *HybridCIMArray {
	if config == nil {
		config = DefaultHybridCIMConfig()
	}

	return &HybridCIMArray{
		Config:       config,
		AnalogUnit:   NewAnalogComputeUnit(config.AnalogArraySize, config.AnalogArraySize, config.AnalogPrecision, config.D2DVariation),
		DigitalUnit:  NewDigitalComputeUnit(config.DigitalArraySize, config.DigitalArraySize, config.DigitalPrecision),
		CurrentMode:  HybridMode,
		BooleanLogic: NewBooleanLogicUnit(1000),
	}
}

// ProgramWeights programs weights to both units
func (hca *HybridCIMArray) ProgramWeights(weights [][]float64) error {
	// Resize weights if needed
	analogWeights := resizeWeights(weights, hca.Config.AnalogArraySize, hca.Config.AnalogArraySize)
	digitalWeights := resizeWeights(weights, hca.Config.DigitalArraySize, hca.Config.DigitalArraySize)

	if err := hca.AnalogUnit.Program(analogWeights); err != nil {
		return err
	}

	if err := hca.DigitalUnit.Program(digitalWeights); err != nil {
		return err
	}

	return nil
}

// resizeWeights resizes weight matrix to target dimensions
func resizeWeights(weights [][]float64, rows, cols int) [][]float64 {
	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			if i < len(weights) && j < len(weights[i]) {
				result[i][j] = weights[i][j]
			}
		}
	}
	return result
}

// HybridCompute performs hybrid analog-digital computation
func (hca *HybridCIMArray) HybridCompute(input []float64, precisionRequired float64) ([]float64, error) {
	// Decide mode based on precision requirement
	if precisionRequired > 0.99 {
		hca.CurrentMode = DigitalMode
	} else if precisionRequired < 0.9 {
		hca.CurrentMode = AnalogMode
	} else {
		hca.CurrentMode = HybridMode
	}

	switch hca.CurrentMode {
	case AnalogMode:
		// Pure analog computation
		analogInput := resizeInput(input, hca.Config.AnalogArraySize)
		output, err := hca.AnalogUnit.Compute(analogInput)
		if err != nil {
			return nil, err
		}
		hca.AnalogOps += int64(hca.Config.AnalogArraySize * hca.Config.AnalogArraySize)
		hca.TotalEnergy += 0.1 * float64(hca.Config.AnalogArraySize*hca.Config.AnalogArraySize) // 0.1 pJ/op
		return output, nil

	case DigitalMode:
		// Pure digital computation
		intInput := make([]int64, hca.Config.DigitalArraySize)
		maxVal := int64(1 << (hca.Config.DigitalPrecision - 1))
		for i := 0; i < len(intInput) && i < len(input); i++ {
			intInput[i] = int64(input[i] * float64(maxVal))
		}
		intOutput, err := hca.DigitalUnit.Compute(intInput)
		if err != nil {
			return nil, err
		}
		// Convert back to float
		output := make([]float64, len(intOutput))
		for i, v := range intOutput {
			output[i] = float64(v) / float64(maxVal*int64(hca.Config.DigitalArraySize))
		}
		hca.DigitalOps += int64(hca.Config.DigitalArraySize * hca.Config.DigitalArraySize)
		hca.TotalEnergy += 0.5 * float64(hca.Config.DigitalArraySize*hca.Config.DigitalArraySize) // 0.5 pJ/op
		return output, nil

	case HybridMode:
		// Hybrid: analog for bulk, digital for refinement
		// Step 1: Analog coarse computation
		analogInput := resizeInput(input, hca.Config.AnalogArraySize)
		coarseOutput, err := hca.AnalogUnit.Compute(analogInput)
		if err != nil {
			return nil, err
		}
		hca.AnalogOps += int64(hca.Config.AnalogArraySize * hca.Config.AnalogArraySize)

		// Step 2: Digital refinement for high-precision elements
		output := make([]float64, len(coarseOutput))
		for i := range output {
			// Check if analog result needs refinement
			if math.Abs(coarseOutput[i]) < hca.Config.SwitchThreshold {
				// Use digital for small values (more precision needed)
				intInput := make([]int64, hca.Config.DigitalArraySize)
				maxVal := int64(1 << (hca.Config.DigitalPrecision - 1))
				for j := 0; j < len(intInput) && j < len(input); j++ {
					intInput[j] = int64(input[j] * float64(maxVal))
				}
				intOutput, _ := hca.DigitalUnit.Compute(intInput)
				if i < len(intOutput) {
					output[i] = float64(intOutput[i]) / float64(maxVal*int64(hca.Config.DigitalArraySize))
				}
				hca.DigitalOps++
			} else {
				output[i] = coarseOutput[i]
			}
		}

		hca.TotalEnergy += 0.2 * float64(hca.Config.AnalogArraySize*hca.Config.AnalogArraySize)
		return output, nil
	}

	return nil, fmt.Errorf("unknown compute mode")
}

// resizeInput resizes input to target length
func resizeInput(input []float64, size int) []float64 {
	result := make([]float64, size)
	for i := 0; i < size && i < len(input); i++ {
		result[i] = input[i]
	}
	return result
}

// GetEfficiency returns TOPS/W efficiency
func (hca *HybridCIMArray) GetEfficiency() float64 {
	totalOps := float64(hca.AnalogOps + hca.DigitalOps) * 2 // MAC = 2 ops
	energyJ := hca.TotalEnergy * 1e-12
	if energyJ == 0 {
		return 0
	}
	return totalOps / (energyJ * 1e12) // TOPS/W
}

// GetStatistics returns hybrid CIM statistics
func (hca *HybridCIMArray) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"analog_ops":         hca.AnalogOps,
		"digital_ops":        hca.DigitalOps,
		"total_energy_pJ":    hca.TotalEnergy,
		"efficiency_TOPS_W":  hca.GetEfficiency(),
		"current_mode":       hca.CurrentMode,
		"analog_precision":   hca.Config.AnalogPrecision,
		"digital_precision":  hca.Config.DigitalPrecision,
		"c2c_variation":      hca.Config.C2CVariation,
		"d2d_variation":      hca.Config.D2DVariation,
	}
}

// ============================================================================
// FELIX ARCHITECTURE (MIXED-SIGNAL FEFET)
// ============================================================================

// FELIXConfig configures FELIX mixed-signal architecture
type FELIXConfig struct {
	ArraySize        int     // FeFET array dimension
	InputPrecision   int     // Input bits (typically 8)
	WeightPrecision  int     // Weight bits (typically 4)
	OutputPrecision  int     // Output bits (ADC resolution)
	TargetTOPSPerW   float64 // Target efficiency
}

// DefaultFELIXConfig returns FELIX configuration
func DefaultFELIXConfig() *FELIXConfig {
	return &FELIXConfig{
		ArraySize:       64,
		InputPrecision:  8,
		WeightPrecision: 4,
		OutputPrecision: 8,
		TargetTOPSPerW:  36.5, // From paper
	}
}

// FELIXAccelerator implements FELIX mixed-signal architecture
type FELIXAccelerator struct {
	Config       *FELIXConfig
	FeFETArray   [][]float64 // FeFET weight storage

	// Bit decomposition units
	BitDecomposer *BitDecomposer

	// Performance tracking
	TotalMACs    int64
	TotalEnergy  float64 // pJ
}

// BitDecomposer handles bit-serial MAC decomposition
type BitDecomposer struct {
	InputBits  int
	WeightBits int
}

// NewBitDecomposer creates a bit decomposer
func NewBitDecomposer(inputBits, weightBits int) *BitDecomposer {
	return &BitDecomposer{
		InputBits:  inputBits,
		WeightBits: weightBits,
	}
}

// Decompose decomposes MAC into bit-serial operations
func (bd *BitDecomposer) Decompose(input, weight float64) []float64 {
	// Number of partial products
	numPP := bd.InputBits * bd.WeightBits
	partials := make([]float64, numPP)

	// Quantize input and weight
	inputInt := int(math.Round(input * float64(1<<bd.InputBits-1)))
	weightInt := int(math.Round(weight * float64(1<<bd.WeightBits-1)))

	idx := 0
	for i := 0; i < bd.InputBits; i++ {
		inputBit := (inputInt >> i) & 1
		for j := 0; j < bd.WeightBits; j++ {
			weightBit := (weightInt >> j) & 1
			partials[idx] = float64(inputBit * weightBit * (1 << (i + j)))
			idx++
		}
	}

	return partials
}

// NewFELIXAccelerator creates a FELIX accelerator
func NewFELIXAccelerator(config *FELIXConfig) *FELIXAccelerator {
	if config == nil {
		config = DefaultFELIXConfig()
	}

	felix := &FELIXAccelerator{
		Config:        config,
		FeFETArray:    make([][]float64, config.ArraySize),
		BitDecomposer: NewBitDecomposer(config.InputPrecision, config.WeightPrecision),
	}

	for i := 0; i < config.ArraySize; i++ {
		felix.FeFETArray[i] = make([]float64, config.ArraySize)
	}

	return felix
}

// ProgramWeights programs FeFET array
func (felix *FELIXAccelerator) ProgramWeights(weights [][]float64) error {
	for i := 0; i < felix.Config.ArraySize && i < len(weights); i++ {
		for j := 0; j < felix.Config.ArraySize && j < len(weights[i]); j++ {
			// Quantize to weight precision
			levels := 1 << felix.Config.WeightPrecision
			felix.FeFETArray[i][j] = math.Round(weights[i][j]*float64(levels-1)) / float64(levels-1)
		}
	}
	return nil
}

// MixedSignalMVM performs mixed-signal matrix-vector multiplication
func (felix *FELIXAccelerator) MixedSignalMVM(input []float64) ([]float64, error) {
	output := make([]float64, felix.Config.ArraySize)

	for i := 0; i < felix.Config.ArraySize; i++ {
		var accumulator float64

		for j := 0; j < felix.Config.ArraySize && j < len(input); j++ {
			// Bit-decomposed MAC
			partials := felix.BitDecomposer.Decompose(input[j], felix.FeFETArray[i][j])

			// Sum partial products (simulating analog accumulation)
			for _, p := range partials {
				accumulator += p
			}

			felix.TotalMACs++
		}

		// Normalize and apply ADC quantization
		levels := 1 << felix.Config.OutputPrecision
		output[i] = math.Round(accumulator*float64(levels-1)) / float64(levels-1)
	}

	// Energy estimation
	felix.TotalEnergy += float64(felix.Config.ArraySize*felix.Config.ArraySize) * 0.027 // ~27 fJ/MAC

	return output, nil
}

// GetEfficiency returns FELIX efficiency
func (felix *FELIXAccelerator) GetEfficiency() float64 {
	if felix.TotalEnergy == 0 {
		return 0
	}
	ops := float64(felix.TotalMACs) * 2 // MAC = 2 ops
	energyJ := felix.TotalEnergy * 1e-12
	return ops / (energyJ * 1e12)
}

// ============================================================================
// FECIM STOCHASTIC-HYBRID INTEGRATION
// ============================================================================

// FeCIMStochasticConfig configures FeCIM stochastic-hybrid system
type FeCIMStochasticConfig struct {
	// Stochastic computing
	EnableStochastic  bool
	BitStreamLength   int

	// Bayesian inference
	EnableBayesian    bool
	BayesianPrecision int

	// Hybrid CIM
	EnableHybrid      bool
	AnalogPrecision   int
	DigitalPrecision  int

	// TRNG
	EnableTRNG        bool
	TRNGBitRate       float64
}

// DefaultFeCIMStochasticConfig returns default configuration
func DefaultFeCIMStochasticConfig() *FeCIMStochasticConfig {
	return &FeCIMStochasticConfig{
		EnableStochastic:  true,
		BitStreamLength:   256,
		EnableBayesian:    true,
		BayesianPrecision: 4,
		EnableHybrid:      true,
		AnalogPrecision:   6,
		DigitalPrecision:  8,
		EnableTRNG:        true,
		TRNGBitRate:       1e6,
	}
}

// FeCIMStochasticSystem implements complete stochastic-hybrid system
type FeCIMStochasticSystem struct {
	Config      *FeCIMStochasticConfig
	TRNG        *FerroelectricTRNG
	FeBiM       *FeBiMEngine
	HybridCIM   *HybridCIMArray
	FELIX       *FELIXAccelerator

	// Performance tracking
	InferenceCount int64
	TotalEnergy    float64 // fJ
}

// NewFeCIMStochasticSystem creates the complete system
func NewFeCIMStochasticSystem(config *FeCIMStochasticConfig) *FeCIMStochasticSystem {
	if config == nil {
		config = DefaultFeCIMStochasticConfig()
	}

	sys := &FeCIMStochasticSystem{
		Config: config,
	}

	if config.EnableTRNG {
		trngConfig := &TRNGConfig{
			BitRate:       config.TRNGBitRate,
			TargetEntropy: 1.0,
		}
		sys.TRNG = NewFerroelectricTRNG(trngConfig)
	}

	if config.EnableBayesian {
		febiMConfig := &FeBiMConfig{
			BitPrecision:     config.BayesianPrecision,
			StochasticLength: config.BitStreamLength,
		}
		sys.FeBiM = NewFeBiMEngine(febiMConfig)
	}

	if config.EnableHybrid {
		hybridConfig := &HybridCIMConfig{
			AnalogPrecision:  config.AnalogPrecision,
			DigitalPrecision: config.DigitalPrecision,
		}
		sys.HybridCIM = NewHybridCIMArray(hybridConfig)
	}

	sys.FELIX = NewFELIXAccelerator(nil)

	return sys
}

// StochasticInference performs stochastic neural network inference
func (sys *FeCIMStochasticSystem) StochasticInference(input []float64, weights [][]float64) ([]float64, error) {
	// Program weights
	if err := sys.HybridCIM.ProgramWeights(weights); err != nil {
		return nil, err
	}

	// Perform hybrid computation
	output, err := sys.HybridCIM.HybridCompute(input, 0.95)
	if err != nil {
		return nil, err
	}

	sys.InferenceCount++
	sys.TotalEnergy += sys.HybridCIM.TotalEnergy * 1000 // Convert pJ to fJ

	return output, nil
}

// BayesianInference performs Bayesian inference
func (sys *FeCIMStochasticSystem) BayesianInference(priors []float64, likelihoods [][]float64) []float64 {
	if sys.FeBiM == nil {
		return priors
	}

	// Setup network
	sys.FeBiM.Network.AddVariable("query", len(priors), priors)
	sys.FeBiM.ProgramCPT("query", likelihoods)

	// Perform inference
	posterior := sys.FeBiM.StochasticInference("query")

	sys.InferenceCount++
	sys.TotalEnergy += sys.FeBiM.TotalEnergy

	return posterior
}

// GenerateRandomBits generates random bits using TRNG
func (sys *FeCIMStochasticSystem) GenerateRandomBits(count int) []bool {
	if sys.TRNG == nil {
		// Fallback to pseudo-random
		bits := make([]bool, count)
		for i := range bits {
			bits[i] = rand.Float64() < 0.5
		}
		return bits
	}

	return sys.TRNG.GenerateBits(count)
}

// GetSystemMetrics returns comprehensive system metrics
func (sys *FeCIMStochasticSystem) GetSystemMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"config":           sys.Config,
		"inference_count":  sys.InferenceCount,
		"total_energy_fJ":  sys.TotalEnergy,
	}

	if sys.TRNG != nil {
		metrics["trng"] = sys.TRNG.GetQualityMetrics()
	}

	if sys.FeBiM != nil {
		metrics["febim"] = sys.FeBiM.GetMetrics()
	}

	if sys.HybridCIM != nil {
		metrics["hybrid_cim"] = sys.HybridCIM.GetStatistics()
	}

	if sys.FELIX != nil {
		metrics["felix_efficiency_TOPS_W"] = sys.FELIX.GetEfficiency()
	}

	return metrics
}

// ExportConfiguration exports system configuration
func (sys *FeCIMStochasticSystem) ExportConfiguration() ([]byte, error) {
	return json.MarshalIndent(sys.GetSystemMetrics(), "", "  ")
}

// ============================================================================
// BENCHMARKS
// ============================================================================

// StochasticBenchmark runs stochastic computing benchmarks
type StochasticBenchmark struct {
	Results []map[string]interface{}
}

// RunAccuracyBenchmark tests stochastic computation accuracy
func (sb *StochasticBenchmark) RunAccuracyBenchmark(streamLengths []int) {
	for _, length := range streamLengths {
		// Test multiplication accuracy
		errors := make([]float64, 100)
		for i := 0; i < 100; i++ {
			a := rand.Float64()
			b := rand.Float64()
			expected := a * b

			streamA := NewStochasticBitStream(a, length, false)
			streamB := NewStochasticBitStream(b, length, false)
			result := StochasticMultiply(streamA, streamB)

			errors[i] = math.Abs(result.Decode() - expected)
		}

		// Calculate mean error
		var meanError float64
		for _, e := range errors {
			meanError += e
		}
		meanError /= float64(len(errors))

		sb.Results = append(sb.Results, map[string]interface{}{
			"stream_length": length,
			"mean_error":    meanError,
			"accuracy":      1.0 - meanError,
		})
	}
}

// GenerateReport generates benchmark report
func (sb *StochasticBenchmark) GenerateReport() string {
	report := "# Stochastic Computing Benchmark Report\n\n"
	report += "## Accuracy vs Stream Length\n\n"
	report += "| Stream Length | Mean Error | Accuracy |\n"
	report += "|---------------|------------|----------|\n"

	for _, r := range sb.Results {
		report += fmt.Sprintf("| %d | %.4f | %.2f%% |\n",
			r["stream_length"].(int),
			r["mean_error"].(float64),
			r["accuracy"].(float64)*100)
	}

	return report
}

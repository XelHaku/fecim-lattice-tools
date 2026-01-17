// Package layers provides security primitives and hybrid digital-analog CIM
// simulation for IronLattice ferroelectric compute-in-memory systems.
//
// This module implements:
// - Physically Unclonable Functions (PUFs)
// - True Random Number Generators (TRNGs)
// - Adversarial attack detection and mitigation
// - Hybrid analog-digital CIM accelerators
// - Mixed-signal compute architectures
//
// Based on research from:
// - Nature Electronics 2024: Safe, secure CIM accelerators
// - Science Advances 2022: Concealable memristor PUF
// - ACS Nano 2025: 3D stacked memristor PUF
// - ASP-DAC 2024: HCiM ADC-less hybrid accelerator
// - TVLSI 2024: HARDSEA hybrid transformer accelerator
package layers

import (
	"crypto/sha256"
	"math"
	"math/rand"
	"sync"
)

// =============================================================================
// Physically Unclonable Function (PUF)
// =============================================================================

// PUFConfig configures PUF parameters
type PUFConfig struct {
	ArrayRows          int
	ArrayCols          int
	DeviceType         string   // "ReRAM", "MRAM", "FeFET"
	ChallengeSpaceBits int      // Challenge space size
	ResponseBits       int      // Response bit length
	BitErrorRate       float64  // Target BER
	ConcealmentEnabled bool     // Support hide/recover
}

// DefaultPUFConfig returns standard configuration
func DefaultPUFConfig() *PUFConfig {
	return &PUFConfig{
		ArrayRows:          32,
		ArrayCols:          32,
		DeviceType:         "ReRAM",
		ChallengeSpaceBits: 64,
		ResponseBits:       256,
		BitErrorRate:       0.03, // 3% BER target
		ConcealmentEnabled: true,
	}
}

// MemristorPUF implements a physically unclonable function
type MemristorPUF struct {
	Config           *PUFConfig
	DeviceVariations [][]float64  // Inherent device variations
	ChallengeHistory map[string][]byte
	IsConcealed      bool

	// Metrics
	CRPCount         int64    // Challenge-response pairs generated
	InterHammingDist float64  // Inter-chip Hamming distance (ideally ~50%)
	IntraHammingDist float64  // Intra-chip Hamming distance (ideally ~0%)
	Uniformity       float64  // Response uniformity (ideally ~50%)
	Uniqueness       float64  // Uniqueness metric

	mu sync.Mutex
}

// NewMemristorPUF creates a new memristor-based PUF
func NewMemristorPUF(config *PUFConfig) *MemristorPUF {
	if config == nil {
		config = DefaultPUFConfig()
	}

	puf := &MemristorPUF{
		Config:           config,
		DeviceVariations: make([][]float64, config.ArrayRows),
		ChallengeHistory: make(map[string][]byte),
	}

	// Initialize device variations from manufacturing process
	// These represent inherent cell-to-cell variations
	for i := 0; i < config.ArrayRows; i++ {
		puf.DeviceVariations[i] = make([]float64, config.ArrayCols)
		for j := 0; j < config.ArrayCols; j++ {
			// Log-normal distribution typical for ReRAM/memristor
			puf.DeviceVariations[i][j] = rand.NormFloat64()*0.3 + 1.0
		}
	}

	// Calculate initial metrics
	puf.calculateMetrics()

	return puf
}

func (puf *MemristorPUF) calculateMetrics() {
	// Calculate uniformity (should be ~50% ones in response)
	totalBits := puf.Config.ArrayRows * puf.Config.ArrayCols
	onesCount := 0
	for i := 0; i < puf.Config.ArrayRows; i++ {
		for j := 0; j < puf.Config.ArrayCols; j++ {
			if puf.DeviceVariations[i][j] > 1.0 {
				onesCount++
			}
		}
	}
	puf.Uniformity = float64(onesCount) / float64(totalBits) * 100

	// Inter-Hamming distance (between different PUF instances)
	puf.InterHammingDist = 49.5 + rand.Float64()*1.0 // Simulated ~50%

	// Intra-Hamming distance (same PUF, same challenge)
	puf.IntraHammingDist = puf.Config.BitErrorRate * 100
}

// GenerateResponse generates PUF response for a challenge
func (puf *MemristorPUF) GenerateResponse(challenge []byte) []byte {
	puf.mu.Lock()
	defer puf.mu.Unlock()

	if puf.IsConcealed {
		return nil // PUF is hidden
	}

	// Hash challenge to get row/column selections
	hash := sha256.Sum256(challenge)

	response := make([]byte, puf.Config.ResponseBits/8)

	for i := 0; i < puf.Config.ResponseBits; i++ {
		// Select cells based on challenge hash
		rowIdx := int(hash[i%32]) % puf.Config.ArrayRows
		colIdx := int(hash[(i+16)%32]) % puf.Config.ArrayCols

		// Compare adjacent cells (differential pair)
		row2Idx := (rowIdx + 1) % puf.Config.ArrayRows

		if puf.DeviceVariations[rowIdx][colIdx] > puf.DeviceVariations[row2Idx][colIdx] {
			response[i/8] |= 1 << (i % 8)
		}

		// Add noise based on BER
		if rand.Float64() < puf.Config.BitErrorRate {
			response[i/8] ^= 1 << (i % 8)
		}
	}

	// Store in history
	challengeKey := string(challenge)
	puf.ChallengeHistory[challengeKey] = response
	puf.CRPCount++

	return response
}

// Conceal hides the PUF (for concealable PUF)
func (puf *MemristorPUF) Conceal() {
	puf.mu.Lock()
	defer puf.mu.Unlock()

	if !puf.Config.ConcealmentEnabled {
		return
	}

	// SET operation to uniform state (hides variations)
	for i := 0; i < puf.Config.ArrayRows; i++ {
		for j := 0; j < puf.Config.ArrayCols; j++ {
			puf.DeviceVariations[i][j] = 1.0 // Uniform state
		}
	}
	puf.IsConcealed = true
}

// Recover restores concealed PUF
func (puf *MemristorPUF) Recover() {
	puf.mu.Lock()
	defer puf.mu.Unlock()

	if !puf.Config.ConcealmentEnabled {
		return
	}

	// RESET operation restores variations (correlated switching)
	for i := 0; i < puf.Config.ArrayRows; i++ {
		for j := 0; j < puf.Config.ArrayCols; j++ {
			puf.DeviceVariations[i][j] = rand.NormFloat64()*0.3 + 1.0
		}
	}
	puf.IsConcealed = false
}

// GetChallengeSpace returns the CRP space size
func (puf *MemristorPUF) GetChallengeSpace() float64 {
	// For ReRAM, scales quadratically with array size
	n := puf.Config.ArrayRows * puf.Config.ArrayCols
	// C(n, 2) pairs possible
	return float64(n * (n - 1) / 2)
}

// =============================================================================
// 3D Stacked PUF
// =============================================================================

// Stacked3DPUFConfig configures 3D stacked PUF
type Stacked3DPUFConfig struct {
	Layers     int
	ArraySize  int  // Per layer
	SelfDiff   bool // Self-differential pairing
}

// Stacked3DPUF implements 3D stacked memristor PUF
type Stacked3DPUF struct {
	Config        *Stacked3DPUFConfig
	LayerArrays   [][][]float64
	CRPSpaceSize  float64  // ~6×10^19 for 2×32×32
	Reconfigurable bool
}

// NewStacked3DPUF creates a 3D stacked PUF
func NewStacked3DPUF(config *Stacked3DPUFConfig) *Stacked3DPUF {
	if config == nil {
		config = &Stacked3DPUFConfig{
			Layers:    2,
			ArraySize: 32,
			SelfDiff:  true,
		}
	}

	puf := &Stacked3DPUF{
		Config:         config,
		LayerArrays:    make([][][]float64, config.Layers),
		Reconfigurable: true,
	}

	for l := 0; l < config.Layers; l++ {
		puf.LayerArrays[l] = make([][]float64, config.ArraySize)
		for i := 0; i < config.ArraySize; i++ {
			puf.LayerArrays[l][i] = make([]float64, config.ArraySize)
			for j := 0; j < config.ArraySize; j++ {
				puf.LayerArrays[l][i][j] = rand.NormFloat64()*0.3 + 1.0
			}
		}
	}

	// Calculate CRP space: approximately 6×10^19 for 2×32×32
	totalCells := config.Layers * config.ArraySize * config.ArraySize
	puf.CRPSpaceSize = math.Pow(2, float64(totalCells)/32) // Approximation

	return puf
}

// GenerateSelfDiffResponse uses self-differential pairing across layers
func (puf *Stacked3DPUF) GenerateSelfDiffResponse(challenge []byte) []byte {
	hash := sha256.Sum256(challenge)
	responseBits := puf.Config.ArraySize * puf.Config.ArraySize

	response := make([]byte, responseBits/8)

	for i := 0; i < responseBits; i++ {
		row := (int(hash[i%32]) + i) % puf.Config.ArraySize
		col := (int(hash[(i+8)%32]) + i/8) % puf.Config.ArraySize

		// Compare same position across layers (self-differential)
		if puf.LayerArrays[0][row][col] > puf.LayerArrays[1][row][col] {
			response[i/8] |= 1 << (i % 8)
		}
	}

	return response
}

// Reconfigure uses cycle-to-cycle variations for new identity
func (puf *Stacked3DPUF) Reconfigure() {
	for l := 0; l < puf.Config.Layers; l++ {
		for i := 0; i < puf.Config.ArraySize; i++ {
			for j := 0; j < puf.Config.ArraySize; j++ {
				// Cycle-to-cycle variation during reset
				puf.LayerArrays[l][i][j] = rand.NormFloat64()*0.3 + 1.0
			}
		}
	}
}

// =============================================================================
// True Random Number Generator (TRNG)
// =============================================================================

// TRNGConfig configures TRNG parameters
type TRNGConfig struct {
	EntropySource    string   // "RTN", "C2C", "TD", "HRS"
	BitRate          float64  // Bits per second
	PowerConsumption float64  // nJ/bit
	NISTCompliant    bool
}

// DefaultTRNGConfig returns standard configuration
func DefaultTRNGConfig() *TRNGConfig {
	return &TRNGConfig{
		EntropySource:    "RTN",  // Random Telegraph Noise
		BitRate:          32e6,   // 32 Mbps
		PowerConsumption: 0.04,   // 0.04 nJ/bit
		NISTCompliant:    true,
	}
}

// MemristorTRNG implements a true random number generator
type MemristorTRNG struct {
	Config          *TRNGConfig
	EntropyBuffer   []byte
	BufferIndex     int

	// RTN parameters
	CaptureTime     float64  // τc in seconds
	EmissionTime    float64  // τe in seconds
	BiasVoltage     float64  // Read voltage

	// Statistics
	TotalBitsGenerated int64
	NISTTestsPassed    int
}

// NewMemristorTRNG creates a new memristor TRNG
func NewMemristorTRNG(config *TRNGConfig) *MemristorTRNG {
	if config == nil {
		config = DefaultTRNGConfig()
	}

	trng := &MemristorTRNG{
		Config:        config,
		EntropyBuffer: make([]byte, 1024),
		CaptureTime:   1e-6,   // 1 µs typical
		EmissionTime:  1e-6,   // 1 µs typical
		BiasVoltage:   0.2,    // 200 mV read
	}

	// Initialize entropy buffer
	trng.RefillBuffer()

	return trng
}

// RefillBuffer generates new random bytes
func (trng *MemristorTRNG) RefillBuffer() {
	for i := range trng.EntropyBuffer {
		var b byte
		for bit := 0; bit < 8; bit++ {
			// Simulate RTN-based randomness
			// Capture/emission ratio determines bias
			ratio := trng.CaptureTime / (trng.CaptureTime + trng.EmissionTime)

			// Add temperature/voltage variation
			ratio += (rand.Float64() - 0.5) * 0.1

			if rand.Float64() > ratio {
				b |= 1 << bit
			}
		}
		trng.EntropyBuffer[i] = b
	}
	trng.BufferIndex = 0
}

// GetRandomBytes returns random bytes from TRNG
func (trng *MemristorTRNG) GetRandomBytes(n int) []byte {
	result := make([]byte, n)

	for i := 0; i < n; i++ {
		if trng.BufferIndex >= len(trng.EntropyBuffer) {
			trng.RefillBuffer()
		}
		result[i] = trng.EntropyBuffer[trng.BufferIndex]
		trng.BufferIndex++
		trng.TotalBitsGenerated += 8
	}

	return result
}

// GetRandomBit returns a single random bit
func (trng *MemristorTRNG) GetRandomBit() int {
	bytes := trng.GetRandomBytes(1)
	return int(bytes[0] & 1)
}

// RunNISTTests simulates NIST SP800-22 test suite
func (trng *MemristorTRNG) RunNISTTests() *NISTTestResults {
	// Generate test sequence
	testBits := trng.GetRandomBytes(125000) // 1 million bits

	results := &NISTTestResults{
		TotalTests: 15,
	}

	// Simulate test results (real implementation would run actual tests)
	// Good memristor TRNGs pass all 15 tests
	testNames := []string{
		"Frequency", "BlockFrequency", "Runs", "LongestRun",
		"Rank", "FFT", "NonOverlappingTemplate", "OverlappingTemplate",
		"Universal", "LinearComplexity", "Serial", "ApproximateEntropy",
		"CumulativeSums", "RandomExcursions", "RandomExcursionsVariant",
	}

	results.TestResults = make(map[string]bool)

	// Count ones for frequency test
	ones := 0
	for _, b := range testBits {
		for i := 0; i < 8; i++ {
			if (b>>i)&1 == 1 {
				ones++
			}
		}
	}
	totalBits := len(testBits) * 8
	proportion := float64(ones) / float64(totalBits)

	for _, name := range testNames {
		// Frequency test
		if name == "Frequency" {
			results.TestResults[name] = math.Abs(proportion-0.5) < 0.01
		} else {
			// Simulate other tests passing with high probability
			results.TestResults[name] = rand.Float64() > 0.05
		}

		if results.TestResults[name] {
			results.PassedTests++
		}
	}

	trng.NISTTestsPassed = results.PassedTests
	return results
}

// NISTTestResults stores NIST test outcomes
type NISTTestResults struct {
	TotalTests  int
	PassedTests int
	TestResults map[string]bool
}

// =============================================================================
// Adversarial Attack Detection and Mitigation
// =============================================================================

// AdversarialDefenseConfig configures defense mechanisms
type AdversarialDefenseConfig struct {
	DetectionEnabled   bool
	OnChipFinetuning   bool
	NoiseInjection     bool
	QuantizationDefense bool
	PerturbationThreshold float64
}

// AdversarialDefense implements CIM-specific defenses
type AdversarialDefense struct {
	Config              *AdversarialDefenseConfig
	ADCOffsetPattern    []float64  // Chip-specific ADC offsets
	WeightNoise         float64    // Injected weight noise level
	DetectedAttacks     int
	MitigatedAttacks    int
}

// NewAdversarialDefense creates defense module
func NewAdversarialDefense(config *AdversarialDefenseConfig) *AdversarialDefense {
	if config == nil {
		config = &AdversarialDefenseConfig{
			DetectionEnabled:     true,
			OnChipFinetuning:     true,
			NoiseInjection:       true,
			QuantizationDefense:  true,
			PerturbationThreshold: 0.1,
		}
	}

	defense := &AdversarialDefense{
		Config:           config,
		ADCOffsetPattern: make([]float64, 64), // 64 ADC channels
		WeightNoise:      0.05,                 // 5% noise level
	}

	// Initialize chip-specific ADC offsets
	for i := range defense.ADCOffsetPattern {
		defense.ADCOffsetPattern[i] = rand.NormFloat64() * 0.02 // 2% variation
	}

	return defense
}

// DetectAdversarial checks for adversarial perturbations
func (ad *AdversarialDefense) DetectAdversarial(input []float64) bool {
	if !ad.Config.DetectionEnabled {
		return false
	}

	// Check for abnormal input statistics
	var mean, variance float64
	for _, v := range input {
		mean += v
	}
	mean /= float64(len(input))

	for _, v := range input {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(input))

	// Adversarial inputs often have unusual variance
	if variance > 2.0 || variance < 0.1 {
		ad.DetectedAttacks++
		return true
	}

	// Check for small perturbations that could be adversarial
	maxPerturbation := 0.0
	for _, v := range input {
		if math.Abs(v-mean) > maxPerturbation {
			maxPerturbation = math.Abs(v - mean)
		}
	}

	if maxPerturbation < ad.Config.PerturbationThreshold {
		// Suspiciously uniform - could be crafted
		ad.DetectedAttacks++
		return true
	}

	return false
}

// ApplyDefense applies mitigation techniques
func (ad *AdversarialDefense) ApplyDefense(input []float64, weights [][]float64) ([]float64, [][]float64) {
	// Apply noise injection to weights
	if ad.Config.NoiseInjection {
		for i := range weights {
			for j := range weights[i] {
				weights[i][j] += rand.NormFloat64() * ad.WeightNoise
			}
		}
	}

	// Apply ADC offset (breaks transferability of attacks)
	if ad.Config.OnChipFinetuning {
		for i := range input {
			adcIdx := i % len(ad.ADCOffsetPattern)
			input[i] += ad.ADCOffsetPattern[adcIdx]
		}
	}

	// Quantization-based defense
	if ad.Config.QuantizationDefense {
		for i := range input {
			// 8-bit quantization
			input[i] = math.Round(input[i]*128) / 128
		}
	}

	ad.MitigatedAttacks++
	return input, weights
}

// OnChipFinetune adapts weights to chip-specific characteristics
func (ad *AdversarialDefense) OnChipFinetune(weights [][]float64, labels []int, inputs [][]float64) [][]float64 {
	// Simplified on-chip finetuning
	// Real implementation would use gradient descent with ADC-aware loss

	learningRate := 0.001
	epochs := 10

	for epoch := 0; epoch < epochs; epoch++ {
		for sample := 0; sample < len(inputs); sample++ {
			// Forward pass with ADC offset
			output := make([]float64, len(weights[0]))
			for j := range output {
				for i := range inputs[sample] {
					output[j] += inputs[sample][i] * weights[i][j]
				}
				// Apply ADC offset
				adcIdx := j % len(ad.ADCOffsetPattern)
				output[j] += ad.ADCOffsetPattern[adcIdx]
			}

			// Simple weight update (gradient approximation)
			for i := range weights {
				for j := range weights[i] {
					error := float64(labels[sample]) - output[j%len(output)]
					weights[i][j] += learningRate * error * inputs[sample][i%len(inputs[sample])]
				}
			}
		}
	}

	return weights
}

// =============================================================================
// Hybrid Digital-Analog CIM Accelerator
// =============================================================================

// HybridCIMConfig configures hybrid accelerator
type HybridCIMConfig struct {
	AnalogArraySize    int
	DigitalArraySize   int
	ADCBits            int
	ADCLess            bool     // HCiM mode
	AnalogPrecision    int      // Weight bits in analog
	DigitalPrecision   int      // Scale factor bits in digital
	Architecture       string   // "HCiM", "HARDSEA", "LightCIM"
}

// DefaultHybridCIMConfig returns standard configuration
func DefaultHybridCIMConfig() *HybridCIMConfig {
	return &HybridCIMConfig{
		AnalogArraySize:  64,
		DigitalArraySize: 16,
		ADCBits:          7,
		ADCLess:          true,
		AnalogPrecision:  4,
		DigitalPrecision: 8,
		Architecture:     "HCiM",
	}
}

// HybridCIMAccelerator implements hybrid analog-digital CIM
type HybridCIMAccelerator struct {
	Config            *HybridCIMConfig
	AnalogCrossbar    *AnalogCIMArray
	DigitalArray      *DigitalCIMArray
	ScaleFactors      []float64

	// Performance metrics
	EnergyReduction   float64  // vs pure analog with ADC
	AccuracyDrop      float64  // vs baseline
	EffectiveTOPSW    float64
}

// AnalogCIMArray represents analog crossbar for MVM
type AnalogCIMArray struct {
	Weights      [][]float64
	Rows         int
	Cols         int
	Precision    int
	NoiseLevel   float64
}

// DigitalCIMArray represents digital array for scale factors
type DigitalCIMArray struct {
	ScaleFactors []float64
	Size         int
	Precision    int
}

// NewHybridCIMAccelerator creates a hybrid accelerator
func NewHybridCIMAccelerator(config *HybridCIMConfig) *HybridCIMAccelerator {
	if config == nil {
		config = DefaultHybridCIMConfig()
	}

	accel := &HybridCIMAccelerator{
		Config:       config,
		ScaleFactors: make([]float64, config.DigitalArraySize),
	}

	// Initialize analog crossbar
	accel.AnalogCrossbar = &AnalogCIMArray{
		Weights:    make([][]float64, config.AnalogArraySize),
		Rows:       config.AnalogArraySize,
		Cols:       config.AnalogArraySize,
		Precision:  config.AnalogPrecision,
		NoiseLevel: 0.02, // 2% analog noise
	}

	for i := 0; i < config.AnalogArraySize; i++ {
		accel.AnalogCrossbar.Weights[i] = make([]float64, config.AnalogArraySize)
		for j := 0; j < config.AnalogArraySize; j++ {
			accel.AnalogCrossbar.Weights[i][j] = rand.Float64()*2 - 1
		}
	}

	// Initialize digital array for scale factors
	accel.DigitalArray = &DigitalCIMArray{
		ScaleFactors: make([]float64, config.DigitalArraySize),
		Size:         config.DigitalArraySize,
		Precision:    config.DigitalPrecision,
	}

	for i := range accel.DigitalArray.ScaleFactors {
		accel.DigitalArray.ScaleFactors[i] = 1.0 + rand.Float64()*0.5
	}

	// Calculate energy reduction based on architecture
	if config.ADCLess {
		// HCiM: 11-28× energy reduction vs ADC-based
		accel.EnergyReduction = 20.0 // 20× average
		accel.AccuracyDrop = 0.5     // 0.5% accuracy drop
	} else {
		accel.EnergyReduction = 5.0
		accel.AccuracyDrop = 0.1
	}

	accel.EffectiveTOPSW = 150.0 // Hybrid typically 100-200 TOPS/W

	return accel
}

// ComputeMVM performs matrix-vector multiplication
func (hca *HybridCIMAccelerator) ComputeMVM(input []float64) []float64 {
	// Analog MVM in crossbar
	analogOutput := hca.analogMVM(input)

	// Apply digital scale factors
	output := make([]float64, len(analogOutput))
	for i := range analogOutput {
		scaleIdx := i % hca.Config.DigitalArraySize
		output[i] = analogOutput[i] * hca.DigitalArray.ScaleFactors[scaleIdx]
	}

	return output
}

func (hca *HybridCIMAccelerator) analogMVM(input []float64) []float64 {
	output := make([]float64, hca.AnalogCrossbar.Cols)

	for j := 0; j < hca.AnalogCrossbar.Cols; j++ {
		var sum float64
		for i := 0; i < hca.AnalogCrossbar.Rows && i < len(input); i++ {
			sum += input[i] * hca.AnalogCrossbar.Weights[i][j]
		}

		// Add analog noise
		sum += rand.NormFloat64() * hca.AnalogCrossbar.NoiseLevel * sum

		// Quantize (if not ADC-less, would use ADC here)
		if !hca.Config.ADCLess {
			levels := 1 << hca.Config.ADCBits
			sum = math.Round(sum*float64(levels)) / float64(levels)
		}

		output[j] = sum
	}

	return output
}

// =============================================================================
// HARDSEA Architecture
// =============================================================================

// HARDSEAConfig configures HARDSEA accelerator
type HARDSEAConfig struct {
	ReRAMClusters    int     // Number of analog ReRAM clusters
	SRAMBanks        int     // Number of digital SRAM banks
	SparsityRatio    float64 // Dynamic attention sparsity
	TokenClusters    int     // Product quantization clusters
}

// HARDSEAAccelerator implements hybrid ReRAM-SRAM transformer accelerator
type HARDSEAAccelerator struct {
	Config           *HARDSEAConfig
	ReRAMArrays      []*AnalogCIMArray  // High efficiency
	SRAMArrays       []*DigitalCIMArray // High precision

	// Sparse attention support
	AttentionMask    [][]bool
	RelevanceScores  []float64
}

// NewHARDSEAAccelerator creates HARDSEA accelerator
func NewHARDSEAAccelerator(config *HARDSEAConfig) *HARDSEAAccelerator {
	if config == nil {
		config = &HARDSEAConfig{
			ReRAMClusters:  8,
			SRAMBanks:      4,
			SparsityRatio:  0.9,
			TokenClusters:  16,
		}
	}

	accel := &HARDSEAAccelerator{
		Config:      config,
		ReRAMArrays: make([]*AnalogCIMArray, config.ReRAMClusters),
		SRAMArrays:  make([]*DigitalCIMArray, config.SRAMBanks),
	}

	// Initialize ReRAM arrays (analog, high efficiency)
	for i := 0; i < config.ReRAMClusters; i++ {
		accel.ReRAMArrays[i] = &AnalogCIMArray{
			Rows:       64,
			Cols:       64,
			Precision:  4,
			NoiseLevel: 0.03,
		}
	}

	// Initialize SRAM arrays (digital, high precision)
	for i := 0; i < config.SRAMBanks; i++ {
		accel.SRAMArrays[i] = &DigitalCIMArray{
			Size:      256,
			Precision: 8,
		}
	}

	return accel
}

// ComputeSparseAttention computes attention with dynamic sparsity
func (ha *HARDSEAAccelerator) ComputeSparseAttention(queries, keys, values [][]float64) [][]float64 {
	seqLen := len(queries)

	// Step 1: Product quantization for lightweight relevance prediction
	relevance := ha.predictRelevance(queries, keys)

	// Step 2: Create sparse attention mask
	threshold := ha.getSparsityThreshold(relevance, ha.Config.SparsityRatio)
	mask := make([][]bool, seqLen)
	for i := range mask {
		mask[i] = make([]bool, seqLen)
		for j := range mask[i] {
			mask[i][j] = relevance[i*seqLen+j] > threshold
		}
	}

	// Step 3: Compute attention only for non-sparse positions
	// Use ReRAM for bulk computation, SRAM for precision-critical
	output := make([][]float64, seqLen)
	for i := range output {
		output[i] = make([]float64, len(values[0]))

		for j := 0; j < seqLen; j++ {
			if mask[i][j] {
				// Compute attention score
				score := relevance[i*seqLen+j]

				// Weighted sum of values
				for k := range output[i] {
					output[i][k] += score * values[j][k]
				}
			}
		}
	}

	return output
}

func (ha *HARDSEAAccelerator) predictRelevance(queries, keys [][]float64) []float64 {
	seqLen := len(queries)
	relevance := make([]float64, seqLen*seqLen)

	// Simplified dot-product relevance
	for i := 0; i < seqLen; i++ {
		for j := 0; j < seqLen; j++ {
			var dot float64
			for k := 0; k < len(queries[i]) && k < len(keys[j]); k++ {
				dot += queries[i][k] * keys[j][k]
			}
			relevance[i*seqLen+j] = dot / math.Sqrt(float64(len(queries[i])))
		}
	}

	return relevance
}

func (ha *HARDSEAAccelerator) getSparsityThreshold(scores []float64, sparsity float64) float64 {
	// Find threshold that achieves target sparsity
	sorted := make([]float64, len(scores))
	copy(sorted, scores)

	// Simple sorting (bubble sort for demonstration)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] < sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	keepCount := int(float64(len(sorted)) * (1 - sparsity))
	if keepCount >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	if keepCount < 1 {
		keepCount = 1
	}

	return sorted[keepCount-1]
}

// =============================================================================
// Light-CIM: Fully Analog Tiles
// =============================================================================

// LightCIMConfig configures Light-CIM accelerator
type LightCIMConfig struct {
	NumFANTs         int     // Fully Analog Tiles
	TileSize         int
	NonidealityAware bool
	ADCFree          bool
	DACFree          bool
}

// LightCIMAccelerator implements fully analog CIM
type LightCIMAccelerator struct {
	Config       *LightCIMConfig
	AnalogTiles  []*FullyAnalogTile
	EnergyPJ     float64
}

// FullyAnalogTile represents a FANT
type FullyAnalogTile struct {
	Weights       [][]float64
	Size          int
	AnalogInputs  []float64
	AnalogOutputs []float64
}

// NewLightCIMAccelerator creates Light-CIM accelerator
func NewLightCIMAccelerator(config *LightCIMConfig) *LightCIMAccelerator {
	if config == nil {
		config = &LightCIMConfig{
			NumFANTs:         16,
			TileSize:         32,
			NonidealityAware: true,
			ADCFree:          true,
			DACFree:          true,
		}
	}

	accel := &LightCIMAccelerator{
		Config:      config,
		AnalogTiles: make([]*FullyAnalogTile, config.NumFANTs),
	}

	for i := 0; i < config.NumFANTs; i++ {
		accel.AnalogTiles[i] = &FullyAnalogTile{
			Weights: make([][]float64, config.TileSize),
			Size:    config.TileSize,
		}
		for j := 0; j < config.TileSize; j++ {
			accel.AnalogTiles[i].Weights[j] = make([]float64, config.TileSize)
		}
	}

	return accel
}

// ProcessFullyAnalog performs computation without ADC/DAC
func (lca *LightCIMAccelerator) ProcessFullyAnalog(inputs []float64) []float64 {
	outputs := make([]float64, lca.Config.TileSize)

	for _, tile := range lca.AnalogTiles {
		// Fully analog MVM
		for j := 0; j < tile.Size; j++ {
			var sum float64
			for i := 0; i < tile.Size && i < len(inputs); i++ {
				sum += inputs[i] * tile.Weights[i][j]
			}
			outputs[j] += sum
		}
	}

	// Energy: much lower without ADC/DAC
	lca.EnergyPJ += float64(len(inputs)) * 0.1 // 0.1 pJ per operation

	return outputs
}

// =============================================================================
// Mixed-Signal Design Automation
// =============================================================================

// MixedSignalConfig configures mixed-signal design parameters
type MixedSignalConfig struct {
	TargetSNR         float64  // Signal-to-noise ratio target
	ADCPrecision      int      // ADC bits
	DACPrecision      int      // DAC bits
	PowerBudgetMW     float64  // Total power budget
	AreaBudgetMM2     float64  // Total area budget
	PerColumnAffine   bool     // Per-column compensation
}

// MixedSignalDesigner assists with CIM design
type MixedSignalDesigner struct {
	Config            *MixedSignalConfig
	SNRModel          *ComputeAwareSNRModel
	OptimalADCBits    int
	EnergyBreakdown   map[string]float64
}

// ComputeAwareSNRModel models SNR for CIM
type ComputeAwareSNRModel struct {
	ComputeSNR        float64  // CSNR
	DeviceVariation   float64  // σ/µ
	ADCQuantNoise     float64
	IRDropNoise       float64
}

// NewMixedSignalDesigner creates design automation tool
func NewMixedSignalDesigner(config *MixedSignalConfig) *MixedSignalDesigner {
	if config == nil {
		config = &MixedSignalConfig{
			TargetSNR:       40.0,  // 40 dB
			ADCPrecision:    8,
			DACPrecision:    8,
			PowerBudgetMW:   100,
			AreaBudgetMM2:   1.0,
			PerColumnAffine: true,
		}
	}

	designer := &MixedSignalDesigner{
		Config: config,
		SNRModel: &ComputeAwareSNRModel{
			DeviceVariation: 0.05,  // 5% variation
			ADCQuantNoise:   0.01,
			IRDropNoise:     0.02,
		},
		EnergyBreakdown: make(map[string]float64),
	}

	designer.OptimizeADC()

	return designer
}

// OptimizeADC finds optimal ADC precision
func (msd *MixedSignalDesigner) OptimizeADC() {
	// Compute-aware SNR allows ADC reduction by 3 bits
	// with 6 dB CSNR gain → 40-64× ADC energy savings

	baseADC := msd.Config.ADCPrecision
	csnrGain := 6.0 // dB

	// Calculate optimal ADC bits
	// Higher CSNR allows lower ADC precision
	msd.OptimalADCBits = baseADC - 3 // 3-bit reduction possible

	if msd.OptimalADCBits < 4 {
		msd.OptimalADCBits = 4
	}

	// Energy breakdown
	msd.EnergyBreakdown["ADC"] = 30.0 // Base %
	msd.EnergyBreakdown["DAC"] = 10.0
	msd.EnergyBreakdown["Array"] = 40.0
	msd.EnergyBreakdown["Digital"] = 20.0

	// With optimization
	adcSavings := math.Pow(2, float64(baseADC-msd.OptimalADCBits))
	msd.EnergyBreakdown["ADC_optimized"] = msd.EnergyBreakdown["ADC"] / adcSavings

	// Calculate compute SNR
	msd.SNRModel.ComputeSNR = msd.Config.TargetSNR + csnrGain
}

// EstimatePerformance estimates accelerator performance
func (msd *MixedSignalDesigner) EstimatePerformance(arraySize int) map[string]float64 {
	metrics := make(map[string]float64)

	// TOPS calculation
	opsPerCycle := float64(arraySize * arraySize * 2) // 2 ops per MAC
	frequency := 500e6 // 500 MHz typical
	tops := opsPerCycle * frequency / 1e12

	metrics["TOPS"] = tops
	metrics["Power_mW"] = msd.Config.PowerBudgetMW
	metrics["TOPS_per_W"] = tops / (msd.Config.PowerBudgetMW / 1000)
	metrics["Area_mm2"] = msd.Config.AreaBudgetMM2
	metrics["TOPS_per_mm2"] = tops / msd.Config.AreaBudgetMM2
	metrics["Optimal_ADC_bits"] = float64(msd.OptimalADCBits)
	metrics["CSNR_dB"] = msd.SNRModel.ComputeSNR

	return metrics
}

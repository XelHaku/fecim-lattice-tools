// security_multichip_cim.go - CIM Security and Multi-Chip Integration
// IronLattice Visualization Project - Iteration 144
//
// This module implements security mechanisms (PUF, watermarking, attack detection)
// and multi-chip/chiplet integration for compute-in-memory systems.
//
// Research sources:
// - Nature Communications 2025: Physical unclonable in-memory computing
// - ResearchGate 2025: MRAM-based CIM model extraction vulnerabilities
// - ResearchGate 2024: PowerGAN side-channel attacks on CIM accelerators
// - IEEE 2024: Integrated Security Mechanisms for Memristive Crossbar Arrays
// - imec: Chiplet interconnect technology
// - IEEE HIR 2025: Interconnect Technologies for Multi-Chiplet Integration

package layers

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
)

// ============================================================================
// Physically Unclonable Functions (PUF) for CIM
// ============================================================================

// PUFConfig configures PUF-based security
type PUFConfig struct {
	ArrayRows         int
	ArrayCols         int
	PUFType           string  // "reram", "sram", "arbiter", "ring_oscillator"
	ChallengeLength   int     // bits
	ResponseLength    int     // bits
	EnrollmentSamples int     // samples for enrollment
	ReliabilityTarget float64 // target BER < 3%
	UniquenessTarget  float64 // target ~50% Hamming distance
}

// ReRAMPUF implements ReRAM-based physically unclonable function
type ReRAMPUF struct {
	Config            *PUFConfig
	CellStates        [][]float64 // Conductance values
	EnrolledCRPs      []*ChallengeResponsePair
	HelperData        []byte
	Statistics        *PUFStatistics
}

// ChallengeResponsePair stores a CRP
type ChallengeResponsePair struct {
	Challenge  []byte
	Response   []byte
	Timestamp  int64
	Reliable   bool
}

// PUFStatistics tracks PUF quality metrics
type PUFStatistics struct {
	Uniqueness        float64 // Inter-device Hamming distance
	Reliability       float64 // Intra-device consistency
	Uniformity        float64 // Bias towards 0 or 1
	BitAliasing       float64 // Bit-specific bias
	BitErrorRate      float64 // BER after error correction
	EntropyPerBit     float64 // min-entropy estimation
}

// NewReRAMPUF creates a ReRAM-based PUF
func NewReRAMPUF(config *PUFConfig) *ReRAMPUF {
	puf := &ReRAMPUF{
		Config:       config,
		CellStates:   make([][]float64, config.ArrayRows),
		EnrolledCRPs: make([]*ChallengeResponsePair, 0),
		Statistics:   &PUFStatistics{},
	}

	// Initialize random conductance states (simulating device variation)
	for r := 0; r < config.ArrayRows; r++ {
		puf.CellStates[r] = make([]float64, config.ArrayCols)
		for c := 0; c < config.ArrayCols; c++ {
			// Bimodal distribution with variation
			if rand.Float64() < 0.5 {
				puf.CellStates[r][c] = 0.2 + rand.NormFloat64()*0.05 // LRS
			} else {
				puf.CellStates[r][c] = 0.8 + rand.NormFloat64()*0.05 // HRS
			}
		}
	}

	return puf
}

// GenerateResponse produces PUF response for a challenge
func (puf *ReRAMPUF) GenerateResponse(challenge []byte) []byte {
	response := make([]byte, puf.Config.ResponseLength/8)

	// Challenge selects which cells to compare
	// Quadratic challenge space: select pairs of cells
	for i := 0; i < puf.Config.ResponseLength; i++ {
		// Derive cell addresses from challenge
		idx1 := int(binary.BigEndian.Uint32(challenge[i%len(challenge):]))
		idx2 := int(binary.BigEndian.Uint32(challenge[(i+4)%len(challenge):]))

		row1 := idx1 % puf.Config.ArrayRows
		col1 := (idx1 / puf.Config.ArrayRows) % puf.Config.ArrayCols
		row2 := idx2 % puf.Config.ArrayRows
		col2 := (idx2 / puf.Config.ArrayRows) % puf.Config.ArrayCols

		// Compare conductances
		g1 := puf.CellStates[row1][col1]
		g2 := puf.CellStates[row2][col2]

		// Add read noise
		g1 += rand.NormFloat64() * 0.01
		g2 += rand.NormFloat64() * 0.01

		// Generate response bit
		byteIdx := i / 8
		bitIdx := uint(i % 8)
		if g1 > g2 {
			response[byteIdx] |= 1 << bitIdx
		}
	}

	return response
}

// Enroll generates helper data for reliable reproduction
func (puf *ReRAMPUF) Enroll(numChallenges int) {
	for i := 0; i < numChallenges; i++ {
		// Generate random challenge
		challenge := make([]byte, puf.Config.ChallengeLength/8)
		rand.Read(challenge)

		// Sample multiple responses
		responses := make([][]byte, puf.Config.EnrollmentSamples)
		for s := 0; s < puf.Config.EnrollmentSamples; s++ {
			responses[s] = puf.GenerateResponse(challenge)
		}

		// Determine reliable bits via majority voting
		goldResponse := puf.majorityVote(responses)

		// Check reliability
		reliable := true
		for s := 0; s < puf.Config.EnrollmentSamples; s++ {
			hd := puf.hammingDistance(goldResponse, responses[s])
			if float64(hd)/float64(len(goldResponse)*8) > 0.05 {
				reliable = false
				break
			}
		}

		crp := &ChallengeResponsePair{
			Challenge: challenge,
			Response:  goldResponse,
			Reliable:  reliable,
		}
		puf.EnrolledCRPs = append(puf.EnrolledCRPs, crp)
	}

	// Generate fuzzy extractor helper data
	puf.generateHelperData()
}

// majorityVote computes majority response across samples
func (puf *ReRAMPUF) majorityVote(responses [][]byte) []byte {
	if len(responses) == 0 {
		return nil
	}

	result := make([]byte, len(responses[0]))
	for byteIdx := 0; byteIdx < len(result); byteIdx++ {
		for bitIdx := 0; bitIdx < 8; bitIdx++ {
			ones := 0
			for _, resp := range responses {
				if (resp[byteIdx] >> uint(bitIdx)) & 1 == 1 {
					ones++
				}
			}
			if ones > len(responses)/2 {
				result[byteIdx] |= 1 << uint(bitIdx)
			}
		}
	}
	return result
}

// hammingDistance calculates Hamming distance between two byte slices
func (puf *ReRAMPUF) hammingDistance(a, b []byte) int {
	if len(a) != len(b) {
		return -1
	}
	dist := 0
	for i := 0; i < len(a); i++ {
		xor := a[i] ^ b[i]
		for xor != 0 {
			dist += int(xor & 1)
			xor >>= 1
		}
	}
	return dist
}

// generateHelperData creates BCH code helper data
func (puf *ReRAMPUF) generateHelperData() {
	// Simplified: concatenate enrolled responses with parity
	var helper []byte
	for _, crp := range puf.EnrolledCRPs {
		if crp.Reliable {
			helper = append(helper, crp.Response...)
		}
	}

	// Add simple parity bits (BCH simplified)
	parity := byte(0)
	for _, b := range helper {
		parity ^= b
	}
	helper = append(helper, parity)

	puf.HelperData = helper
}

// Authenticate verifies a response against enrolled data
func (puf *ReRAMPUF) Authenticate(challenge, response []byte) bool {
	// Find enrolled CRP
	for _, crp := range puf.EnrolledCRPs {
		if string(crp.Challenge) == string(challenge) {
			// Allow small Hamming distance for noise
			hd := puf.hammingDistance(crp.Response, response)
			threshold := len(response) * 8 / 10 // 10% tolerance
			return hd <= threshold
		}
	}
	return false
}

// EvaluateStatistics computes PUF quality metrics
func (puf *ReRAMPUF) EvaluateStatistics(numDevices int) *PUFStatistics {
	// Simulate multiple device instances
	devices := make([]*ReRAMPUF, numDevices)
	for d := 0; d < numDevices; d++ {
		devices[d] = NewReRAMPUF(puf.Config)
	}

	// Test challenge
	challenge := make([]byte, puf.Config.ChallengeLength/8)
	rand.Read(challenge)

	// Collect responses
	responses := make([][]byte, numDevices)
	for d := 0; d < numDevices; d++ {
		responses[d] = devices[d].GenerateResponse(challenge)
	}

	// Uniqueness: average inter-device Hamming distance
	totalHD := 0
	pairs := 0
	for i := 0; i < numDevices; i++ {
		for j := i + 1; j < numDevices; j++ {
			totalHD += puf.hammingDistance(responses[i], responses[j])
			pairs++
		}
	}
	puf.Statistics.Uniqueness = float64(totalHD) / float64(pairs) / float64(len(responses[0])*8) * 100

	// Reliability: intra-device consistency
	numSamples := 100
	totalBER := 0.0
	for s := 0; s < numSamples; s++ {
		resp := puf.GenerateResponse(challenge)
		hd := puf.hammingDistance(responses[0], resp)
		totalBER += float64(hd) / float64(len(resp)*8)
	}
	puf.Statistics.Reliability = 100 - (totalBER/float64(numSamples))*100
	puf.Statistics.BitErrorRate = totalBER / float64(numSamples)

	// Uniformity: bias towards 0 or 1
	ones := 0
	total := 0
	for _, resp := range responses {
		for _, b := range resp {
			for i := 0; i < 8; i++ {
				if (b >> uint(i)) & 1 == 1 {
					ones++
				}
				total++
			}
		}
	}
	puf.Statistics.Uniformity = float64(ones) / float64(total) * 100

	// Entropy estimation (simplified)
	puf.Statistics.EntropyPerBit = -math.Log2(math.Max(float64(ones), float64(total-ones))/float64(total))

	return puf.Statistics
}

// ============================================================================
// Model Extraction Attack Simulation
// ============================================================================

// ModelExtractionConfig configures extraction attack simulation
type ModelExtractionConfig struct {
	TargetModel       *CIMModel
	QueryBudget       int
	AttackType        string  // "query_synthesis", "knockoff", "memory_sca"
	NoiseLevel        float64 // Simulated measurement noise
	PartialKnowledge  float64 // Fraction of architecture known
}

// CIMModel represents a model deployed on CIM hardware
type CIMModel struct {
	Name            string
	Architecture    []int      // Layer sizes
	Weights         [][][]float64
	QuantizedWeights [][][]int
	QuantBits       int
	Confidential    bool
}

// ModelExtractionAttack simulates extraction attacks
type ModelExtractionAttack struct {
	Config          *ModelExtractionConfig
	QueriesUsed     int
	ExtractedWeights [][][]float64
	AccuracyMatch   float64
	FidelityScore   float64
}

// QuerySynthesisAttack performs query-based extraction
func (mea *ModelExtractionAttack) QuerySynthesisAttack() *ExtractionResult {
	result := &ExtractionResult{
		AttackType:    "query_synthesis",
		QueriesUsed:   0,
		PartialWeights: make([][][]float64, 0),
	}

	// Generate synthetic queries
	for q := 0; q < mea.Config.QueryBudget; q++ {
		// Random input
		inputSize := mea.Config.TargetModel.Architecture[0]
		input := make([]float64, inputSize)
		for i := range input {
			input[i] = rand.Float64()*2 - 1 // [-1, 1]
		}

		// Query target model (simulated)
		output := mea.queryTargetModel(input)

		// Collect input-output pairs for training substitute
		result.QueriesUsed++
		mea.QueriesUsed = result.QueriesUsed

		// Add noise to simulate real-world conditions
		for i := range output {
			output[i] += rand.NormFloat64() * mea.Config.NoiseLevel
		}
	}

	// Train substitute model (simplified simulation)
	result.FidelityEstimate = mea.estimateFidelity()
	result.Success = result.FidelityEstimate > 0.8

	return result
}

// queryTargetModel simulates querying the target CIM model
func (mea *ModelExtractionAttack) queryTargetModel(input []float64) []float64 {
	// Simple forward pass simulation
	current := input
	for l := 0; l < len(mea.Config.TargetModel.Weights); l++ {
		weights := mea.Config.TargetModel.Weights[l]
		next := make([]float64, len(weights[0]))

		for j := 0; j < len(weights[0]); j++ {
			sum := 0.0
			for i := 0; i < len(current); i++ {
				sum += current[i] * weights[i][j]
			}
			// ReLU activation
			if sum < 0 {
				sum = 0
			}
			next[j] = sum
		}
		current = next
	}
	return current
}

// estimateFidelity estimates how well extraction succeeded
func (mea *ModelExtractionAttack) estimateFidelity() float64 {
	// Simplified: based on query budget and noise
	baseFidelity := float64(mea.QueriesUsed) / float64(mea.Config.QueryBudget)
	noisePenalty := mea.Config.NoiseLevel * 2
	knowledgeBonus := mea.Config.PartialKnowledge * 0.2

	fidelity := baseFidelity - noisePenalty + knowledgeBonus
	return math.Max(0, math.Min(1, fidelity))
}

// ExtractionResult holds attack results
type ExtractionResult struct {
	AttackType       string
	QueriesUsed      int
	PartialWeights   [][][]float64
	FidelityEstimate float64
	Success          bool
	DetectedBy       []string
}

// ============================================================================
// Side-Channel Attack Simulation
// ============================================================================

// SideChannelConfig configures side-channel attack simulation
type SideChannelConfig struct {
	AttackType        string  // "power", "timing", "em"
	SamplingRate      float64 // MHz
	NumTraces         int
	SignalNoiseRatio  float64
	TargetComponent   string  // "adc", "crossbar", "accumulator"
}

// PowerSideChannelAttack simulates power analysis attacks
type PowerSideChannelAttack struct {
	Config        *SideChannelConfig
	PowerTraces   [][]float64
	ExtractedInfo *ExtractedInformation
}

// ExtractedInformation holds side-channel extracted data
type ExtractedInformation struct {
	WeightEstimates   [][]float64
	InputReconstruct  []float64
	ArchitectureInfo  map[string]int
	ConfidenceLevel   float64
}

// NewPowerSideChannelAttack creates a power SCA simulator
func NewPowerSideChannelAttack(config *SideChannelConfig) *PowerSideChannelAttack {
	return &PowerSideChannelAttack{
		Config:      config,
		PowerTraces: make([][]float64, 0),
		ExtractedInfo: &ExtractedInformation{
			ArchitectureInfo: make(map[string]int),
		},
	}
}

// CollectTraces simulates collecting power traces
func (psca *PowerSideChannelAttack) CollectTraces(model *CIMModel, inputs [][]float64) {
	for _, input := range inputs {
		trace := psca.simulatePowerTrace(model, input)
		psca.PowerTraces = append(psca.PowerTraces, trace)
	}
}

// simulatePowerTrace generates simulated power consumption
func (psca *PowerSideChannelAttack) simulatePowerTrace(model *CIMModel, input []float64) []float64 {
	trace := make([]float64, 0)

	// Simulate power for each layer
	for l := 0; l < len(model.Weights); l++ {
		weights := model.Weights[l]

		// Crossbar power: depends on input values and weights
		crossbarPower := 0.0
		for i := 0; i < len(input); i++ {
			for j := 0; j < len(weights[0]); j++ {
				// Power ∝ V² × G
				crossbarPower += input[i] * input[i] * math.Abs(weights[i][j])
			}
		}

		// ADC power: depends on output magnitude
		adcPower := float64(len(weights[0])) * 0.5 // Base ADC power

		// Add temporal samples
		samplesPerLayer := 100
		for s := 0; s < samplesPerLayer; s++ {
			power := crossbarPower + adcPower
			// Add noise
			power += rand.NormFloat64() * power / psca.Config.SignalNoiseRatio
			trace = append(trace, power)
		}
	}

	return trace
}

// PerformCPA performs Correlation Power Analysis
func (psca *PowerSideChannelAttack) PerformCPA(targetByte int) []float64 {
	if len(psca.PowerTraces) == 0 {
		return nil
	}

	traceLen := len(psca.PowerTraces[0])
	correlations := make([]float64, 256) // For each key hypothesis

	for keyGuess := 0; keyGuess < 256; keyGuess++ {
		// Compute hypothetical intermediate values
		hypotheticals := make([]float64, len(psca.PowerTraces))
		for t := 0; t < len(psca.PowerTraces); t++ {
			// Simplified: Hamming weight model
			hypotheticals[t] = float64(hammingWeight(byte(keyGuess ^ t)))
		}

		// Correlate with power traces at each time point
		maxCorr := 0.0
		for timePoint := 0; timePoint < traceLen; timePoint++ {
			measurements := make([]float64, len(psca.PowerTraces))
			for t := 0; t < len(psca.PowerTraces); t++ {
				measurements[t] = psca.PowerTraces[t][timePoint]
			}

			corr := math.Abs(pearsonCorrelation(hypotheticals, measurements))
			if corr > maxCorr {
				maxCorr = corr
			}
		}
		correlations[keyGuess] = maxCorr
	}

	return correlations
}

// hammingWeight counts set bits
func hammingWeight(b byte) int {
	count := 0
	for b != 0 {
		count += int(b & 1)
		b >>= 1
	}
	return count
}

// pearsonCorrelation computes Pearson correlation coefficient
func pearsonCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	n := float64(len(x))
	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	num := n*sumXY - sumX*sumY
	den := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if den == 0 {
		return 0
	}
	return num / den
}

// PowerGANAttack simulates GAN-based input reconstruction
type PowerGANAttack struct {
	Config       *SideChannelConfig
	Generator    *SimpleGenerator
	Discriminator *SimpleDiscriminator
	TrainedEpochs int
}

// SimpleGenerator represents a simplified generator network
type SimpleGenerator struct {
	Weights [][]float64
	Biases  []float64
}

// SimpleDiscriminator represents a simplified discriminator
type SimpleDiscriminator struct {
	Weights [][]float64
	Biases  []float64
}

// TrainGAN trains the GAN for input reconstruction
func (pga *PowerGANAttack) TrainGAN(powerTraces [][]float64, knownInputs [][]float64, epochs int) {
	// Simplified GAN training simulation
	for epoch := 0; epoch < epochs; epoch++ {
		// Train discriminator
		// Train generator
		pga.TrainedEpochs = epoch + 1
	}
}

// ReconstructInput attempts to reconstruct input from power trace
func (pga *PowerGANAttack) ReconstructInput(powerTrace []float64) []float64 {
	// Simplified: add noise to demonstrate reconstruction
	inputSize := len(powerTrace) / 10 // Estimate
	reconstructed := make([]float64, inputSize)

	for i := range reconstructed {
		// Simple linear combination of trace samples
		idx := i * 10
		if idx < len(powerTrace) {
			reconstructed[i] = powerTrace[idx] / 100 // Normalize
		}
	}

	return reconstructed
}

// ============================================================================
// IP Protection Mechanisms
// ============================================================================

// IPProtectionConfig configures IP protection
type IPProtectionConfig struct {
	Method           string  // "keyed_permutor", "watermark", "obfuscation"
	KeyLength        int     // bits
	WatermarkStrength float64 // 0-1
	RedundancyLevel  int     // Extra columns for watermark
}

// KeyedPermutor implements keyed weight permutation
type KeyedPermutor struct {
	Config       *IPProtectionConfig
	Key          []byte
	PermutationMap []int
	InverseMap   []int
}

// NewKeyedPermutor creates a keyed permutor
func NewKeyedPermutor(config *IPProtectionConfig, key []byte) *KeyedPermutor {
	kp := &KeyedPermutor{
		Config: config,
		Key:    key,
	}
	kp.generatePermutation()
	return kp
}

// generatePermutation creates key-dependent permutation
func (kp *KeyedPermutor) generatePermutation() {
	// Use key to seed PRNG for permutation
	hash := sha256.Sum256(kp.Key)
	seed := int64(binary.BigEndian.Uint64(hash[:8]))
	rng := rand.New(rand.NewSource(seed))

	// Fisher-Yates shuffle for permutation
	size := 1 << 16 // Maximum permutation size
	kp.PermutationMap = make([]int, size)
	for i := range kp.PermutationMap {
		kp.PermutationMap[i] = i
	}
	for i := size - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		kp.PermutationMap[i], kp.PermutationMap[j] = kp.PermutationMap[j], kp.PermutationMap[i]
	}

	// Create inverse map
	kp.InverseMap = make([]int, size)
	for i, p := range kp.PermutationMap {
		kp.InverseMap[p] = i
	}
}

// PermuteWeights applies permutation to weights
func (kp *KeyedPermutor) PermuteWeights(weights [][]float64) [][]float64 {
	rows := len(weights)
	cols := len(weights[0])

	permuted := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		permuted[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			// Permute based on linearized index
			origIdx := r*cols + c
			permIdx := kp.PermutationMap[origIdx%len(kp.PermutationMap)]
			newR := permIdx / cols
			newC := permIdx % cols
			if newR < rows && newC < cols {
				permuted[newR][newC] = weights[r][c]
			}
		}
	}

	return permuted
}

// UnpermuteWeights reverses permutation
func (kp *KeyedPermutor) UnpermuteWeights(permuted [][]float64) [][]float64 {
	rows := len(permuted)
	cols := len(permuted[0])

	original := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		original[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			permIdx := r*cols + c
			origIdx := kp.InverseMap[permIdx%len(kp.InverseMap)]
			origR := origIdx / cols
			origC := origIdx % cols
			if origR < rows && origC < cols {
				original[origR][origC] = permuted[r][c]
			}
		}
	}

	return original
}

// WatermarkProtection implements watermark-based IP protection
type WatermarkProtection struct {
	Config          *IPProtectionConfig
	WatermarkKey    []byte
	WatermarkBits   []int
	VerificationHash []byte
}

// NewWatermarkProtection creates watermark protection
func NewWatermarkProtection(config *IPProtectionConfig) *WatermarkProtection {
	wp := &WatermarkProtection{
		Config:       config,
		WatermarkKey: make([]byte, config.KeyLength/8),
	}
	rand.Read(wp.WatermarkKey)

	// Generate watermark bit positions
	hash := sha256.Sum256(wp.WatermarkKey)
	rng := rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(hash[:8]))))

	wp.WatermarkBits = make([]int, config.RedundancyLevel)
	for i := range wp.WatermarkBits {
		wp.WatermarkBits[i] = rng.Intn(1000000)
	}

	return wp
}

// EmbedWatermark embeds watermark in weights
func (wp *WatermarkProtection) EmbedWatermark(weights [][]float64, ownerID string) [][]float64 {
	watermarked := make([][]float64, len(weights))
	for r := range weights {
		watermarked[r] = make([]float64, len(weights[r]))
		copy(watermarked[r], weights[r])
	}

	// Embed owner ID hash into specific weight positions
	ownerHash := sha256.Sum256([]byte(ownerID))

	for i, bitPos := range wp.WatermarkBits {
		rows := len(weights)
		cols := len(weights[0])
		r := bitPos / cols % rows
		c := bitPos % cols

		// Modify LSB based on watermark
		bitValue := (ownerHash[i%32] >> uint(i%8)) & 1
		delta := wp.Config.WatermarkStrength * 0.01 * float64(bitValue*2-1)
		watermarked[r][c] += delta
	}

	// Store verification hash
	wp.VerificationHash = ownerHash[:]

	return watermarked
}

// VerifyWatermark checks if watermark is present
func (wp *WatermarkProtection) VerifyWatermark(weights [][]float64, ownerID string) bool {
	ownerHash := sha256.Sum256([]byte(ownerID))

	matches := 0
	total := len(wp.WatermarkBits)

	for i, bitPos := range wp.WatermarkBits {
		rows := len(weights)
		cols := len(weights[0])
		r := bitPos / cols % rows
		c := bitPos % cols

		expectedBit := (ownerHash[i%32] >> uint(i%8)) & 1
		// Check if weight has expected sign of perturbation
		// This is simplified; real implementation would be more robust

		if (weights[r][c] > 0) == (expectedBit == 1) {
			matches++
		}
	}

	// Require >70% match for positive verification
	return float64(matches)/float64(total) > 0.7
}

// ============================================================================
// Multi-Chip and Chiplet Integration
// ============================================================================

// ChipletConfig configures a single chiplet
type ChipletConfig struct {
	ChipletID       int
	ChipletType     string  // "compute", "memory", "io"
	NumCrossbars    int
	CrossbarSize    int
	LocalSRAMKB     float64
	TDP_mW          float64
	Technology      string  // "fefet", "reram", "sram"
}

// ChipletSystem represents multi-chiplet CIM system
type ChipletSystem struct {
	Config           *ChipletSystemConfig
	Chiplets         []*CIMChiplet
	Interposer       *Interposer
	NetworkOnChip    *ChipletNoC
	Scheduler        *MultiChipScheduler
	PowerManager     *ChipletPowerManager
}

// ChipletSystemConfig configures the multi-chiplet system
type ChipletSystemConfig struct {
	NumComputeChiplets int
	NumMemoryChiplets  int
	InterconnectType   string  // "2.5d_interposer", "3d_stacking", "emib"
	InterconnectBW     float64 // GB/s per link
	UCIeVersion        string  // "1.0", "1.1", "2.0"
	TotalPowerBudget   float64 // Watts
}

// CIMChiplet represents a single CIM chiplet
type CIMChiplet struct {
	Config         *ChipletConfig
	Crossbars      []*CrossbarUnit
	LocalBuffer    *ChipletBuffer
	State          *ChipletState
	PUF            *ReRAMPUF
}

// CrossbarUnit represents a crossbar array within chiplet
type CrossbarUnit struct {
	UnitID        int
	Rows          int
	Cols          int
	Weights       [][]float64
	Utilization   float64
	PowerState    string // "active", "idle", "sleep"
}

// ChipletBuffer represents local SRAM buffer
type ChipletBuffer struct {
	SizeKB       float64
	Partitions   []*BufferPartition
	BandwidthGBs float64
}

// BufferPartition represents a buffer partition
type BufferPartition struct {
	PartitionID  int
	Purpose      string // "input", "output", "weight_cache"
	SizeKB       float64
	Allocated    bool
}

// ChipletState tracks chiplet runtime state
type ChipletState struct {
	PowerState       string  // "active", "idle", "sleep", "off"
	Temperature      float64 // Celsius
	CurrentPower     float64 // mW
	Utilization      float64
	TasksCompleted   int
	ErrorCount       int
}

// Interposer represents the interconnect substrate
type Interposer struct {
	Type           string // "silicon", "organic", "glass"
	AreaMM2        float64
	TSVPitch       float64 // μm
	BumpPitch      float64 // μm
	Layers         int
	RoutingDensity float64 // tracks per mm
}

// ChipletNoC implements Network-on-Chip for chiplets
type ChipletNoC struct {
	Topology       string // "mesh", "ring", "crossbar", "tree"
	NumRouters     int
	LinkBandwidth  float64 // GB/s
	LatencyCycles  int
	Routers        []*NoCRouter
	Links          []*NoCLink
}

// NoCRouter represents a NoC router
type NoCRouter struct {
	RouterID      int
	ConnectedChiplets []int
	BufferDepth   int
	Arbitration   string // "round_robin", "priority", "age_based"
	QueuedPackets []*NoCPacket
}

// NoCLink represents a NoC link
type NoCLink struct {
	LinkID       int
	SourceRouter int
	DestRouter   int
	Bandwidth    float64 // GB/s
	Latency      float64 // ns
	Utilization  float64
}

// NoCPacket represents a packet in the NoC
type NoCPacket struct {
	PacketID    int
	Source      int
	Destination int
	PayloadSize int // bytes
	Priority    int
	Timestamp   int64
}

// MultiChipScheduler schedules computation across chiplets
type MultiChipScheduler struct {
	Config         *SchedulerConfig
	TaskQueue      []*ComputeTask
	ChipletAssignment map[int]int // task -> chiplet
	DataflowPolicy string
}

// SchedulerConfig configures the scheduler
type SchedulerConfig struct {
	Policy          string  // "round_robin", "load_balance", "locality"
	MaxQueueDepth   int
	PreemptionEnabled bool
	DataLocalityWeight float64
}

// ComputeTask represents a computation task
type ComputeTask struct {
	TaskID        int
	LayerID       int
	InputShape    []int
	OutputShape   []int
	WeightShape   []int
	Dependencies  []int
	AssignedChiplet int
	Status        string // "pending", "running", "completed"
	StartTime     int64
	EndTime       int64
}

// ChipletPowerManager manages power across chiplets
type ChipletPowerManager struct {
	Config           *PowerManagerConfig
	PowerBudget      float64
	CurrentTotal     float64
	ChipletPower     map[int]float64
	DVFSStates       map[int]*DVFSState
}

// PowerManagerConfig configures power management
type PowerManagerConfig struct {
	TotalBudgetW    float64
	PerChipletMaxW  float64
	DVFSEnabled     bool
	PowerGating     bool
	ThermalThreshold float64 // Celsius
}

// DVFSState tracks DVFS state for a chiplet
type DVFSState struct {
	ChipletID    int
	Voltage      float64
	Frequency    float64
	PowerLevel   int // 0=max_perf, 1=balanced, 2=low_power
}

// NewChipletSystem creates a multi-chiplet CIM system
func NewChipletSystem(config *ChipletSystemConfig) *ChipletSystem {
	system := &ChipletSystem{
		Config:   config,
		Chiplets: make([]*CIMChiplet, 0),
	}

	// Create compute chiplets
	for i := 0; i < config.NumComputeChiplets; i++ {
		chiplet := &CIMChiplet{
			Config: &ChipletConfig{
				ChipletID:    i,
				ChipletType:  "compute",
				NumCrossbars: 16,
				CrossbarSize: 128,
				LocalSRAMKB:  256,
				TDP_mW:       500,
				Technology:   "fefet",
			},
			Crossbars:   make([]*CrossbarUnit, 16),
			LocalBuffer: &ChipletBuffer{SizeKB: 256},
			State:       &ChipletState{PowerState: "idle"},
		}

		for j := 0; j < 16; j++ {
			chiplet.Crossbars[j] = &CrossbarUnit{
				UnitID:     j,
				Rows:       128,
				Cols:       128,
				PowerState: "idle",
			}
		}

		system.Chiplets = append(system.Chiplets, chiplet)
	}

	// Create interposer
	system.Interposer = &Interposer{
		Type:           "silicon",
		AreaMM2:        400,
		TSVPitch:       9.0,  // μm
		BumpPitch:      45.0, // μm
		Layers:         2,
		RoutingDensity: 1000,
	}

	// Create NoC
	system.NetworkOnChip = &ChipletNoC{
		Topology:      "mesh",
		NumRouters:    config.NumComputeChiplets,
		LinkBandwidth: config.InterconnectBW,
		LatencyCycles: 5,
		Routers:       make([]*NoCRouter, config.NumComputeChiplets),
		Links:         make([]*NoCLink, 0),
	}

	// Initialize routers
	for i := 0; i < config.NumComputeChiplets; i++ {
		system.NetworkOnChip.Routers[i] = &NoCRouter{
			RouterID:          i,
			ConnectedChiplets: []int{i},
			BufferDepth:       16,
			Arbitration:       "round_robin",
		}
	}

	// Create mesh links
	gridSize := int(math.Ceil(math.Sqrt(float64(config.NumComputeChiplets))))
	linkID := 0
	for i := 0; i < config.NumComputeChiplets; i++ {
		// Connect to right neighbor
		if (i+1)%gridSize != 0 && i+1 < config.NumComputeChiplets {
			link := &NoCLink{
				LinkID:       linkID,
				SourceRouter: i,
				DestRouter:   i + 1,
				Bandwidth:    config.InterconnectBW,
				Latency:      1.0,
			}
			system.NetworkOnChip.Links = append(system.NetworkOnChip.Links, link)
			linkID++
		}

		// Connect to bottom neighbor
		if i+gridSize < config.NumComputeChiplets {
			link := &NoCLink{
				LinkID:       linkID,
				SourceRouter: i,
				DestRouter:   i + gridSize,
				Bandwidth:    config.InterconnectBW,
				Latency:      1.0,
			}
			system.NetworkOnChip.Links = append(system.NetworkOnChip.Links, link)
			linkID++
		}
	}

	// Create scheduler
	system.Scheduler = &MultiChipScheduler{
		Config: &SchedulerConfig{
			Policy:             "load_balance",
			MaxQueueDepth:      100,
			PreemptionEnabled:  false,
			DataLocalityWeight: 0.5,
		},
		TaskQueue:         make([]*ComputeTask, 0),
		ChipletAssignment: make(map[int]int),
	}

	// Create power manager
	system.PowerManager = &ChipletPowerManager{
		Config: &PowerManagerConfig{
			TotalBudgetW:     config.TotalPowerBudget,
			PerChipletMaxW:   0.5,
			DVFSEnabled:      true,
			PowerGating:      true,
			ThermalThreshold: 85.0,
		},
		PowerBudget:  config.TotalPowerBudget,
		ChipletPower: make(map[int]float64),
		DVFSStates:   make(map[int]*DVFSState),
	}

	return system
}

// ScheduleTask assigns a task to a chiplet
func (cs *ChipletSystem) ScheduleTask(task *ComputeTask) int {
	// Find least loaded chiplet
	minLoad := math.MaxFloat64
	selectedChiplet := 0

	for i, chiplet := range cs.Chiplets {
		if chiplet.State.Utilization < minLoad {
			minLoad = chiplet.State.Utilization
			selectedChiplet = i
		}
	}

	task.AssignedChiplet = selectedChiplet
	task.Status = "pending"
	cs.Scheduler.TaskQueue = append(cs.Scheduler.TaskQueue, task)
	cs.Scheduler.ChipletAssignment[task.TaskID] = selectedChiplet

	return selectedChiplet
}

// ExecuteTask runs a task on assigned chiplet
func (cs *ChipletSystem) ExecuteTask(task *ComputeTask) error {
	chiplet := cs.Chiplets[task.AssignedChiplet]

	// Check power budget
	if !cs.PowerManager.RequestPower(task.AssignedChiplet, 0.1) {
		return fmt.Errorf("power budget exceeded")
	}

	// Update state
	task.Status = "running"
	chiplet.State.PowerState = "active"
	chiplet.State.Utilization += 0.1

	// Simulate computation (simplified)
	// In reality, this would perform MVM on crossbars

	task.Status = "completed"
	chiplet.State.TasksCompleted++

	// Release power
	cs.PowerManager.ReleasePower(task.AssignedChiplet, 0.1)

	return nil
}

// RequestPower requests power allocation
func (pm *ChipletPowerManager) RequestPower(chipletID int, powerW float64) bool {
	if pm.CurrentTotal+powerW > pm.PowerBudget {
		return false
	}

	pm.CurrentTotal += powerW
	pm.ChipletPower[chipletID] += powerW
	return true
}

// ReleasePower releases power allocation
func (pm *ChipletPowerManager) ReleasePower(chipletID int, powerW float64) {
	pm.CurrentTotal -= powerW
	pm.ChipletPower[chipletID] -= powerW
}

// RoutePacket routes a packet through the NoC
func (noc *ChipletNoC) RoutePacket(packet *NoCPacket) []int {
	// XY routing for mesh topology
	path := []int{packet.Source}

	current := packet.Source
	dest := packet.Destination

	gridSize := int(math.Ceil(math.Sqrt(float64(noc.NumRouters))))

	// Route in X direction first
	for current%gridSize != dest%gridSize {
		if current%gridSize < dest%gridSize {
			current++
		} else {
			current--
		}
		path = append(path, current)
	}

	// Then route in Y direction
	for current/gridSize != dest/gridSize {
		if current/gridSize < dest/gridSize {
			current += gridSize
		} else {
			current -= gridSize
		}
		path = append(path, current)
	}

	return path
}

// ============================================================================
// UCIe Interconnect Support
// ============================================================================

// UCIeConfig configures UCIe interconnect
type UCIeConfig struct {
	Version          string  // "1.0", "1.1", "2.0"
	Protocol         string  // "streaming", "flit"
	DataWidth        int     // bits
	Speed            float64 // GT/s
	LatencyNs        float64
	EnergyPerBit     float64 // pJ/bit
	DieToDieReach    string  // "standard", "advanced"
}

// UCIeInterface implements UCIe interface
type UCIeInterface struct {
	Config         *UCIeConfig
	TxFIFO         []*UCIeFlit
	RxFIFO         []*UCIeFlit
	LinkState      string // "active", "idle", "sleep", "off"
	ErrorCount     int
	BandwidthUsed  float64
}

// UCIeFlit represents a UCIe flit
type UCIeFlit struct {
	FlitID     int
	FlitType   string // "header", "data", "tail"
	Payload    []byte
	CRC        uint32
	Timestamp  int64
}

// NewUCIeInterface creates a UCIe interface
func NewUCIeInterface(config *UCIeConfig) *UCIeInterface {
	return &UCIeInterface{
		Config:    config,
		TxFIFO:    make([]*UCIeFlit, 0),
		RxFIFO:    make([]*UCIeFlit, 0),
		LinkState: "idle",
	}
}

// Transmit sends flits over UCIe
func (ucie *UCIeInterface) Transmit(data []byte) error {
	// Fragment into flits
	flitSize := ucie.Config.DataWidth / 8 // bytes
	numFlits := (len(data) + flitSize - 1) / flitSize

	for i := 0; i < numFlits; i++ {
		start := i * flitSize
		end := start + flitSize
		if end > len(data) {
			end = len(data)
		}

		flitType := "data"
		if i == 0 {
			flitType = "header"
		} else if i == numFlits-1 {
			flitType = "tail"
		}

		flit := &UCIeFlit{
			FlitID:   i,
			FlitType: flitType,
			Payload:  data[start:end],
		}
		ucie.TxFIFO = append(ucie.TxFIFO, flit)
	}

	ucie.LinkState = "active"
	ucie.BandwidthUsed += float64(len(data)) / 1e9 // GB

	return nil
}

// CalculateBandwidth computes effective bandwidth
func (ucie *UCIeInterface) CalculateBandwidth() float64 {
	// BW = DataWidth * Speed * (1 - overhead)
	overhead := 0.15 // 15% protocol overhead
	return float64(ucie.Config.DataWidth) * ucie.Config.Speed * (1 - overhead) / 8 // GB/s
}

// ============================================================================
// Security Attack Detection
// ============================================================================

// AttackDetector monitors for security attacks
type AttackDetector struct {
	Config            *DetectorConfig
	QueryHistory      []*QueryRecord
	PowerBaseline     []float64
	AnomalyThreshold  float64
	AlertLog          []*SecurityAlert
}

// DetectorConfig configures attack detection
type DetectorConfig struct {
	MonitorQueries    bool
	MonitorPower      bool
	MonitorTiming     bool
	WindowSize        int
	SensitivityLevel  float64 // 0-1
}

// QueryRecord logs inference queries
type QueryRecord struct {
	QueryID     int
	Timestamp   int64
	InputHash   []byte
	OutputHash  []byte
	Latency     float64
	SourceIP    string
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	AlertID     int
	AlertType   string // "extraction", "sca", "tampering", "anomaly"
	Severity    string // "low", "medium", "high", "critical"
	Timestamp   int64
	Description string
	Evidence    map[string]interface{}
}

// NewAttackDetector creates an attack detector
func NewAttackDetector(config *DetectorConfig) *AttackDetector {
	return &AttackDetector{
		Config:           config,
		QueryHistory:     make([]*QueryRecord, 0),
		AnomalyThreshold: 0.9,
		AlertLog:         make([]*SecurityAlert, 0),
	}
}

// RecordQuery logs a query for analysis
func (ad *AttackDetector) RecordQuery(query *QueryRecord) {
	ad.QueryHistory = append(ad.QueryHistory, query)

	// Keep window size
	if len(ad.QueryHistory) > ad.Config.WindowSize {
		ad.QueryHistory = ad.QueryHistory[1:]
	}

	// Check for extraction attack patterns
	ad.checkExtractionPattern()
}

// checkExtractionPattern detects model extraction attempts
func (ad *AttackDetector) checkExtractionPattern() {
	if len(ad.QueryHistory) < 100 {
		return
	}

	// Check for systematic query patterns
	// Extraction attacks often have regular spacing and coverage
	inputHashes := make(map[string]int)
	for _, q := range ad.QueryHistory {
		key := string(q.InputHash[:8])
		inputHashes[key]++
	}

	// High diversity suggests extraction
	diversity := float64(len(inputHashes)) / float64(len(ad.QueryHistory))
	if diversity > 0.9 {
		alert := &SecurityAlert{
			AlertID:     len(ad.AlertLog),
			AlertType:   "extraction",
			Severity:    "high",
			Description: "Possible model extraction attack detected",
			Evidence: map[string]interface{}{
				"query_diversity": diversity,
				"window_size":     len(ad.QueryHistory),
			},
		}
		ad.AlertLog = append(ad.AlertLog, alert)
	}
}

// CheckPowerAnomaly detects power-based attacks
func (ad *AttackDetector) CheckPowerAnomaly(currentPower []float64) bool {
	if len(ad.PowerBaseline) == 0 {
		ad.PowerBaseline = currentPower
		return false
	}

	// Compare with baseline
	diff := 0.0
	for i := range currentPower {
		if i < len(ad.PowerBaseline) {
			diff += math.Abs(currentPower[i] - ad.PowerBaseline[i])
		}
	}
	avgDiff := diff / float64(len(currentPower))

	if avgDiff > ad.AnomalyThreshold {
		alert := &SecurityAlert{
			AlertID:     len(ad.AlertLog),
			AlertType:   "sca",
			Severity:    "medium",
			Description: "Abnormal power pattern detected",
			Evidence: map[string]interface{}{
				"avg_deviation": avgDiff,
			},
		}
		ad.AlertLog = append(ad.AlertLog, alert)
		return true
	}

	return false
}

// ============================================================================
// Integrated Security + Multi-Chip Demo
// ============================================================================

// SecureChipletSystem combines security and multi-chip features
type SecureChipletSystem struct {
	ChipletSystem    *ChipletSystem
	PUFs             map[int]*ReRAMPUF
	IPProtection     *WatermarkProtection
	AttackDetector   *AttackDetector
	SecurityLevel    string // "standard", "enhanced", "maximum"
}

// NewSecureChipletSystem creates a secure multi-chiplet system
func NewSecureChipletSystem(config *ChipletSystemConfig, securityLevel string) *SecureChipletSystem {
	system := &SecureChipletSystem{
		ChipletSystem: NewChipletSystem(config),
		PUFs:          make(map[int]*ReRAMPUF),
		SecurityLevel: securityLevel,
	}

	// Initialize PUF for each chiplet
	for i := 0; i < config.NumComputeChiplets; i++ {
		pufConfig := &PUFConfig{
			ArrayRows:         64,
			ArrayCols:         64,
			PUFType:           "reram",
			ChallengeLength:   128,
			ResponseLength:    256,
			EnrollmentSamples: 10,
			ReliabilityTarget: 0.97,
		}
		system.PUFs[i] = NewReRAMPUF(pufConfig)
		system.PUFs[i].Enroll(100)
	}

	// Initialize IP protection
	system.IPProtection = NewWatermarkProtection(&IPProtectionConfig{
		Method:            "watermark",
		KeyLength:         256,
		WatermarkStrength: 0.1,
		RedundancyLevel:   64,
	})

	// Initialize attack detector
	system.AttackDetector = NewAttackDetector(&DetectorConfig{
		MonitorQueries:   true,
		MonitorPower:     true,
		MonitorTiming:    true,
		WindowSize:       1000,
		SensitivityLevel: 0.8,
	})

	return system
}

// AuthenticateChiplet verifies chiplet identity via PUF
func (scs *SecureChipletSystem) AuthenticateChiplet(chipletID int, challenge []byte) bool {
	puf, exists := scs.PUFs[chipletID]
	if !exists {
		return false
	}

	response := puf.GenerateResponse(challenge)
	return puf.Authenticate(challenge, response)
}

// SecureInference performs inference with security monitoring
func (scs *SecureChipletSystem) SecureInference(input []float64, inputHash []byte) ([]float64, error) {
	// Log query
	query := &QueryRecord{
		QueryID:   len(scs.AttackDetector.QueryHistory),
		InputHash: inputHash,
	}
	scs.AttackDetector.RecordQuery(query)

	// Check for alerts
	if len(scs.AttackDetector.AlertLog) > 0 {
		lastAlert := scs.AttackDetector.AlertLog[len(scs.AttackDetector.AlertLog)-1]
		if lastAlert.Severity == "critical" {
			return nil, fmt.Errorf("security alert: %s", lastAlert.Description)
		}
	}

	// Perform inference (simplified)
	output := make([]float64, len(input))
	copy(output, input)

	return output, nil
}

// GenerateSecurityReport creates a security status report
func (scs *SecureChipletSystem) GenerateSecurityReport() string {
	var sb strings.Builder

	sb.WriteString("╔════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║           SECURE CHIPLET SYSTEM STATUS REPORT              ║\n")
	sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")
	sb.WriteString(fmt.Sprintf("║ Security Level: %-41s ║\n", scs.SecurityLevel))
	sb.WriteString(fmt.Sprintf("║ Active Chiplets: %-40d ║\n", len(scs.ChipletSystem.Chiplets)))
	sb.WriteString(fmt.Sprintf("║ PUFs Enrolled: %-42d ║\n", len(scs.PUFs)))
	sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")

	sb.WriteString("║ PUF STATISTICS                                             ║\n")
	for id, puf := range scs.PUFs {
		stats := puf.Statistics
		sb.WriteString(fmt.Sprintf("║   Chiplet %d: Reliability=%.1f%%, BER=%.3f%%           ║\n",
			id, stats.Reliability, stats.BitErrorRate*100))
	}

	sb.WriteString("╠════════════════════════════════════════════════════════════╣\n")
	sb.WriteString("║ SECURITY ALERTS                                            ║\n")
	sb.WriteString(fmt.Sprintf("║   Total Alerts: %-41d ║\n", len(scs.AttackDetector.AlertLog)))

	alertCounts := make(map[string]int)
	for _, alert := range scs.AttackDetector.AlertLog {
		alertCounts[alert.AlertType]++
	}
	for alertType, count := range alertCounts {
		sb.WriteString(fmt.Sprintf("║   %s: %-48d ║\n", alertType, count))
	}

	sb.WriteString("╚════════════════════════════════════════════════════════════╝\n")

	return sb.String()
}

// ============================================================================
// Benchmark Suite for Security and Multi-Chip
// ============================================================================

// SecurityBenchmark runs security benchmarks
type SecurityBenchmark struct {
	Config  *SecurityBenchmarkConfig
	Results []*SecurityBenchmarkResult
}

// SecurityBenchmarkConfig configures benchmarks
type SecurityBenchmarkConfig struct {
	NumTrials          int
	AttackTypes        []string
	ChipletConfigs     []int // Number of chiplets to test
}

// SecurityBenchmarkResult holds benchmark results
type SecurityBenchmarkResult struct {
	Config           string
	AttackType       string
	DetectionRate    float64
	FalsePositiveRate float64
	Latency          float64
	OverheadPercent  float64
}

// RunSecurityBenchmarks executes all benchmarks
func RunSecurityBenchmarks(config *SecurityBenchmarkConfig) []*SecurityBenchmarkResult {
	results := make([]*SecurityBenchmarkResult, 0)

	for _, numChiplets := range config.ChipletConfigs {
		for _, attackType := range config.AttackTypes {
			result := &SecurityBenchmarkResult{
				Config:     fmt.Sprintf("%d_chiplets", numChiplets),
				AttackType: attackType,
			}

			// Simulate detection trials
			detected := 0
			falsePos := 0

			for trial := 0; trial < config.NumTrials; trial++ {
				// Simulate attack
				isAttack := rand.Float64() < 0.5
				detectedAttack := rand.Float64() < 0.85 // 85% detection rate

				if isAttack && detectedAttack {
					detected++
				} else if !isAttack && detectedAttack {
					falsePos++
				}
			}

			result.DetectionRate = float64(detected) / float64(config.NumTrials/2) * 100
			result.FalsePositiveRate = float64(falsePos) / float64(config.NumTrials/2) * 100
			result.Latency = float64(numChiplets) * 0.5 // Simplified
			result.OverheadPercent = 5.0 + float64(numChiplets)*0.5

			results = append(results, result)
		}
	}

	return results
}

// PrintBenchmarkResults formats benchmark results
func PrintBenchmarkResults(results []*SecurityBenchmarkResult) string {
	var sb strings.Builder

	sb.WriteString("\n=== Security & Multi-Chip Benchmark Results ===\n\n")
	sb.WriteString("Config      | Attack Type  | Detection | FP Rate | Latency | Overhead\n")
	sb.WriteString("------------|--------------|-----------|---------|---------|----------\n")

	// Sort by config
	sort.Slice(results, func(i, j int) bool {
		return results[i].Config < results[j].Config
	})

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%-11s | %-12s | %7.1f%% | %6.1f%% | %5.1fns | %6.1f%%\n",
			r.Config, r.AttackType, r.DetectionRate, r.FalsePositiveRate,
			r.Latency, r.OverheadPercent))
	}

	return sb.String()
}

// Package layers provides CIM security/privacy and ferroelectric tunnel junction simulation
// for compute-in-memory architectures.
//
// This module implements:
// - Side-channel attack models and defenses
// - Physical Unclonable Functions (PUF) for CIM
// - Adversarial attack detection and mitigation
// - HZO-based Ferroelectric Tunnel Junctions (FTJ)
// - FTJ synaptic devices for neuromorphic computing
//
// Based on research from Nature Electronics 2024, ACS Applied Materials 2025, and Wiley 2024-2025.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

// =============================================================================
// SIDE-CHANNEL ATTACK MODELS
// =============================================================================

// SideChannelConfig configures side-channel attack simulation
type SideChannelConfig struct {
	AttackType        string  // "power", "timing", "em", "cache"
	NumTraces         int     // Number of power traces
	SamplesPerTrace   int     // Samples per trace
	NoiseLevel        float64 // SNR degradation
	LeakageModel      string  // "hamming_weight", "hamming_distance"
}

// DefaultSideChannelConfig returns default side-channel configuration
func DefaultSideChannelConfig() *SideChannelConfig {
	return &SideChannelConfig{
		AttackType:       "power",
		NumTraces:        10000,
		SamplesPerTrace:  1000,
		NoiseLevel:       0.1,
		LeakageModel:     "hamming_weight",
	}
}

// PowerTrace represents a power consumption trace
type PowerTrace struct {
	Samples     []float64
	Timestamp   float64
	InputData   []byte
	SecretKey   []byte // For analysis only
}

// SideChannelAttacker simulates side-channel attacks on CIM
type SideChannelAttacker struct {
	Config       *SideChannelConfig
	Traces       []*PowerTrace
	LeakedBits   []int
	SuccessRate  float64
	mu           sync.Mutex
}

// NewSideChannelAttacker creates a new side-channel attacker
func NewSideChannelAttacker(config *SideChannelConfig) *SideChannelAttacker {
	return &SideChannelAttacker{
		Config:     config,
		Traces:     make([]*PowerTrace, 0),
		LeakedBits: make([]int, 0),
	}
}

// CaptureTrace captures a power trace during CIM operation
func (sca *SideChannelAttacker) CaptureTrace(input []byte, weights [][]float64) *PowerTrace {
	sca.mu.Lock()
	defer sca.mu.Unlock()

	trace := &PowerTrace{
		Samples:   make([]float64, sca.Config.SamplesPerTrace),
		InputData: make([]byte, len(input)),
	}
	copy(trace.InputData, input)

	// Simulate power consumption based on Hamming weight
	for i := 0; i < sca.Config.SamplesPerTrace; i++ {
		basePower := 10.0 // Base power consumption (mW)

		// Data-dependent leakage
		if i < len(input) && i < len(weights) {
			hw := hammingWeight(input[i])
			leakage := float64(hw) * 0.5 // 0.5 mW per bit

			// Weight-dependent computation power
			for j := 0; j < len(weights[i]) && j < 8; j++ {
				leakage += math.Abs(weights[i][j]) * 0.1
			}

			basePower += leakage
		}

		// Add noise
		noise := rand.NormFloat64() * sca.Config.NoiseLevel * basePower
		trace.Samples[i] = basePower + noise
	}

	sca.Traces = append(sca.Traces, trace)
	return trace
}

// hammingWeight calculates the Hamming weight of a byte
func hammingWeight(b byte) int {
	count := 0
	for b != 0 {
		count += int(b & 1)
		b >>= 1
	}
	return count
}

// CorrelationPowerAnalysis performs CPA attack
func (sca *SideChannelAttacker) CorrelationPowerAnalysis() []float64 {
	if len(sca.Traces) < 100 {
		return nil
	}

	numSamples := sca.Config.SamplesPerTrace
	correlations := make([]float64, 256) // For each key hypothesis

	for keyGuess := 0; keyGuess < 256; keyGuess++ {
		// Compute hypothetical power for each trace
		hypothetical := make([]float64, len(sca.Traces))
		for t, trace := range sca.Traces {
			if len(trace.InputData) > 0 {
				// S-box output hypothesis
				intermediate := trace.InputData[0] ^ byte(keyGuess)
				hypothetical[t] = float64(hammingWeight(intermediate))
			}
		}

		// Compute correlation with actual power
		maxCorr := 0.0
		for s := 0; s < numSamples; s++ {
			actual := make([]float64, len(sca.Traces))
			for t, trace := range sca.Traces {
				if s < len(trace.Samples) {
					actual[t] = trace.Samples[s]
				}
			}

			corr := pearsonCorrelation(hypothetical, actual)
			if math.Abs(corr) > math.Abs(maxCorr) {
				maxCorr = corr
			}
		}
		correlations[keyGuess] = maxCorr
	}

	return correlations
}

// pearsonCorrelation computes Pearson correlation coefficient
func pearsonCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) == 0 {
		return 0
	}

	n := float64(len(x))
	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := range x {
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

// =============================================================================
// PHYSICAL UNCLONABLE FUNCTIONS (PUF)
// =============================================================================

// PUFConfig configures a Physical Unclonable Function
type PUFConfig struct {
	Type            string  // "arbiter", "ring_oscillator", "sram", "mram"
	NumChallenges   int     // Number of challenge-response pairs
	ChallengeBits   int     // Bits per challenge
	ResponseBits    int     // Bits per response
	NoiseLevel      float64 // Response noise
	Reliability     float64 // Response reliability (0-1)
}

// DefaultPUFConfig returns default PUF configuration
func DefaultPUFConfig() *PUFConfig {
	return &PUFConfig{
		Type:          "mram",
		NumChallenges: 10000,
		ChallengeBits: 64,
		ResponseBits:  64,
		NoiseLevel:    0.05,
		Reliability:   0.95,
	}
}

// ChallengeResponsePair represents a PUF CRP
type ChallengeResponsePair struct {
	Challenge  []byte
	Response   []byte
	Reliability float64
}

// PUF implements a Physical Unclonable Function
type PUF struct {
	Config       *PUFConfig
	SecretState  []float64 // Manufacturing variations
	CRPs         []*ChallengeResponsePair
	Stats        *PUFStats
}

// PUFStats tracks PUF statistics
type PUFStats struct {
	Uniqueness       float64 // Inter-device Hamming distance
	Reliability      float64 // Intra-device stability
	Uniformity       float64 // Bit bias (ideal: 50%)
	BitAliasing      float64 // Response predictability
	ModelingAccuracy float64 // ML attack success rate
}

// NewPUF creates a new PUF
func NewPUF(config *PUFConfig) *PUF {
	puf := &PUF{
		Config:      config,
		SecretState: make([]float64, config.ChallengeBits*2),
		CRPs:        make([]*ChallengeResponsePair, 0),
		Stats:       &PUFStats{},
	}

	// Initialize with random manufacturing variations
	for i := range puf.SecretState {
		puf.SecretState[i] = rand.NormFloat64() * 0.1
	}

	return puf
}

// GenerateResponse generates a response for a challenge
func (p *PUF) GenerateResponse(challenge []byte) []byte {
	response := make([]byte, p.Config.ResponseBits/8)

	switch p.Config.Type {
	case "arbiter":
		response = p.arbiterPUFResponse(challenge)
	case "ring_oscillator":
		response = p.roPUFResponse(challenge)
	case "sram":
		response = p.sramPUFResponse(challenge)
	case "mram":
		response = p.mramPUFResponse(challenge)
	}

	return response
}

// arbiterPUFResponse simulates arbiter PUF
func (p *PUF) arbiterPUFResponse(challenge []byte) []byte {
	response := make([]byte, p.Config.ResponseBits/8)

	for bit := 0; bit < p.Config.ResponseBits; bit++ {
		delay := 0.0
		for i := 0; i < len(challenge) && i < len(p.SecretState)/2; i++ {
			challengeBit := (challenge[i/8] >> (i % 8)) & 1
			if challengeBit == 1 {
				delay += p.SecretState[i*2]
			} else {
				delay += p.SecretState[i*2+1]
			}
		}

		// Add noise
		delay += rand.NormFloat64() * p.Config.NoiseLevel

		// Convert to response bit
		if delay > 0 {
			response[bit/8] |= (1 << (bit % 8))
		}
	}

	return response
}

// roPUFResponse simulates ring oscillator PUF
func (p *PUF) roPUFResponse(challenge []byte) []byte {
	response := make([]byte, p.Config.ResponseBits/8)

	for bit := 0; bit < p.Config.ResponseBits; bit++ {
		// Select two ring oscillators based on challenge
		idx1 := int(challenge[bit%len(challenge)]) % len(p.SecretState)
		idx2 := (idx1 + 1) % len(p.SecretState)

		freq1 := 1000.0 + p.SecretState[idx1]*100 // Base frequency + variation
		freq2 := 1000.0 + p.SecretState[idx2]*100

		// Add noise
		freq1 += rand.NormFloat64() * p.Config.NoiseLevel * 10
		freq2 += rand.NormFloat64() * p.Config.NoiseLevel * 10

		// Compare frequencies
		if freq1 > freq2 {
			response[bit/8] |= (1 << (bit % 8))
		}
	}

	return response
}

// sramPUFResponse simulates SRAM PUF
func (p *PUF) sramPUFResponse(challenge []byte) []byte {
	response := make([]byte, p.Config.ResponseBits/8)

	// SRAM startup values based on manufacturing variations
	for bit := 0; bit < p.Config.ResponseBits; bit++ {
		idx := int(challenge[bit%len(challenge)]) ^ bit
		idx = idx % len(p.SecretState)

		// SRAM cell has slight bias based on transistor mismatch
		bias := p.SecretState[idx]

		// Add noise (power-up variation)
		bias += rand.NormFloat64() * p.Config.NoiseLevel

		if bias > 0 {
			response[bit/8] |= (1 << (bit % 8))
		}
	}

	return response
}

// mramPUFResponse simulates MRAM-based PUF
// Based on: Scientific Reports 2024 - MRAM PUF resistant to ML attacks
func (p *PUF) mramPUFResponse(challenge []byte) []byte {
	response := make([]byte, p.Config.ResponseBits/8)

	for bit := 0; bit < p.Config.ResponseBits; bit++ {
		// MRAM switching probability depends on TMR variation
		idx := int(challenge[bit%len(challenge)]) ^ (bit * 17)
		idx = idx % len(p.SecretState)

		// TMR variation affects switching threshold
		tmrVariation := p.SecretState[idx]

		// Stochastic switching behavior
		switchProb := 0.5 + tmrVariation*0.3
		switchProb += rand.NormFloat64() * p.Config.NoiseLevel

		if rand.Float64() < switchProb {
			response[bit/8] |= (1 << (bit % 8))
		}
	}

	return response
}

// EvaluatePUF computes PUF quality metrics
func (p *PUF) EvaluatePUF(numSamples int) {
	// Generate sample CRPs
	challenges := make([][]byte, numSamples)
	responses := make([][]byte, numSamples)

	for i := 0; i < numSamples; i++ {
		challenges[i] = make([]byte, p.Config.ChallengeBits/8)
		rand.Read(challenges[i])
		responses[i] = p.GenerateResponse(challenges[i])
	}

	// Uniformity: average Hamming weight of responses
	totalOnes := 0
	totalBits := 0
	for _, resp := range responses {
		for _, b := range resp {
			totalOnes += hammingWeight(b)
			totalBits += 8
		}
	}
	p.Stats.Uniformity = float64(totalOnes) / float64(totalBits)

	// Reliability: repeat same challenge, measure consistency
	reliableCount := 0
	testCount := 100
	for i := 0; i < testCount; i++ {
		challenge := challenges[i%len(challenges)]
		resp1 := p.GenerateResponse(challenge)
		resp2 := p.GenerateResponse(challenge)

		// Calculate Hamming distance
		hd := 0
		for j := range resp1 {
			hd += hammingWeight(resp1[j] ^ resp2[j])
		}
		if hd == 0 {
			reliableCount++
		}
	}
	p.Stats.Reliability = float64(reliableCount) / float64(testCount)

	// Uniqueness: estimate from response entropy
	p.Stats.Uniqueness = math.Abs(p.Stats.Uniformity - 0.5) * 2
}

// =============================================================================
// ADVERSARIAL ATTACK DETECTION
// =============================================================================

// AdversarialConfig configures adversarial attack detection
type AdversarialConfig struct {
	DetectionMethod   string  // "statistical", "nn_guard", "input_transform"
	PerturbationBound float64 // Max perturbation (L-inf norm)
	DetectionThreshold float64 // Anomaly threshold
	DefenseEnabled    bool
}

// DefaultAdversarialConfig returns default adversarial configuration
func DefaultAdversarialConfig() *AdversarialConfig {
	return &AdversarialConfig{
		DetectionMethod:    "statistical",
		PerturbationBound:  0.03, // 3% perturbation
		DetectionThreshold: 2.0,  // 2 standard deviations
		DefenseEnabled:     true,
	}
}

// AdversarialDetector detects adversarial inputs
type AdversarialDetector struct {
	Config           *AdversarialConfig
	InputStatistics  *InputStats
	DetectionHistory []bool
	FalsePositives   int
	TruePositives    int
}

// InputStats tracks input statistics for anomaly detection
type InputStats struct {
	Mean       []float64
	StdDev     []float64
	NumSamples int
}

// NewAdversarialDetector creates a new adversarial detector
func NewAdversarialDetector(config *AdversarialConfig, inputDim int) *AdversarialDetector {
	return &AdversarialDetector{
		Config: config,
		InputStatistics: &InputStats{
			Mean:   make([]float64, inputDim),
			StdDev: make([]float64, inputDim),
		},
		DetectionHistory: make([]bool, 0),
	}
}

// UpdateStatistics updates input statistics with new sample
func (ad *AdversarialDetector) UpdateStatistics(input []float64) {
	stats := ad.InputStatistics
	stats.NumSamples++
	n := float64(stats.NumSamples)

	for i := range input {
		if i < len(stats.Mean) {
			// Welford's online algorithm
			delta := input[i] - stats.Mean[i]
			stats.Mean[i] += delta / n
			delta2 := input[i] - stats.Mean[i]
			stats.StdDev[i] += delta * delta2
		}
	}
}

// FinalizeStatistics computes final standard deviations
func (ad *AdversarialDetector) FinalizeStatistics() {
	stats := ad.InputStatistics
	if stats.NumSamples > 1 {
		for i := range stats.StdDev {
			stats.StdDev[i] = math.Sqrt(stats.StdDev[i] / float64(stats.NumSamples-1))
		}
	}
}

// DetectAdversarial checks if input is adversarial
func (ad *AdversarialDetector) DetectAdversarial(input []float64) (bool, float64) {
	if ad.InputStatistics.NumSamples < 100 {
		// Not enough statistics
		return false, 0
	}

	anomalyScore := 0.0
	stats := ad.InputStatistics

	switch ad.Config.DetectionMethod {
	case "statistical":
		// Z-score based detection
		for i := range input {
			if i < len(stats.Mean) && stats.StdDev[i] > 1e-10 {
				z := math.Abs(input[i]-stats.Mean[i]) / stats.StdDev[i]
				if z > anomalyScore {
					anomalyScore = z
				}
			}
		}

	case "nn_guard":
		// Neural network guard (simplified)
		// Check for unusual activation patterns
		maxVal := 0.0
		minVal := 0.0
		for _, v := range input {
			if v > maxVal {
				maxVal = v
			}
			if v < minVal {
				minVal = v
			}
		}
		anomalyScore = maxVal - minVal
		if anomalyScore > 10*ad.Config.DetectionThreshold {
			anomalyScore = 10.0
		}

	case "input_transform":
		// Input transformation consistency
		// Apply small random transformation and check prediction consistency
		// (Simplified: check gradient magnitude)
		gradMag := 0.0
		for i := 1; i < len(input); i++ {
			gradMag += math.Abs(input[i] - input[i-1])
		}
		anomalyScore = gradMag / float64(len(input))
	}

	isAdversarial := anomalyScore > ad.Config.DetectionThreshold
	ad.DetectionHistory = append(ad.DetectionHistory, isAdversarial)

	return isAdversarial, anomalyScore
}

// ApplyDefense applies adversarial defense if enabled
func (ad *AdversarialDetector) ApplyDefense(input []float64) []float64 {
	if !ad.Config.DefenseEnabled {
		return input
	}

	defended := make([]float64, len(input))
	copy(defended, input)

	// Input quantization defense
	levels := 256.0
	for i := range defended {
		defended[i] = math.Round(defended[i]*levels) / levels
	}

	// Gaussian smoothing
	if len(defended) > 2 {
		smoothed := make([]float64, len(defended))
		smoothed[0] = defended[0]
		smoothed[len(smoothed)-1] = defended[len(defended)-1]
		for i := 1; i < len(defended)-1; i++ {
			smoothed[i] = 0.25*defended[i-1] + 0.5*defended[i] + 0.25*defended[i+1]
		}
		defended = smoothed
	}

	return defended
}

// =============================================================================
// FERROELECTRIC TUNNEL JUNCTION (FTJ)
// =============================================================================

// FTJConfig configures a Ferroelectric Tunnel Junction
type FTJConfig struct {
	// Material parameters (HZO)
	FerroelectricThickness float64 // nm (typically 4-10nm)
	InterlayerThickness    float64 // nm (TiO2 interlayer)
	Area                   float64 // µm²

	// Electrical parameters
	CoerciveVoltage        float64 // V
	RemanentPolarization   float64 // µC/cm²
	TERRatio               float64 // Tunnel electroresistance ON/OFF

	// Device parameters
	NumConductanceStates   int     // Multi-level states
	WriteVoltage           float64 // V
	WritePulseWidth        float64 // ns
	ReadVoltage            float64 // V
	Endurance              int64   // Write cycles
	RetentionSeconds       float64 // Data retention
}

// DefaultFTJConfig returns default HZO FTJ configuration
// Based on: ACS Applied Materials 2025, Wiley 2024-2025
func DefaultFTJConfig() *FTJConfig {
	return &FTJConfig{
		FerroelectricThickness: 5.0,    // 5 nm HZO
		InterlayerThickness:    2.0,    // 2 nm TiO2
		Area:                   0.01,   // 0.01 µm²

		CoerciveVoltage:        1.5,    // 1.5 V
		RemanentPolarization:   25.0,   // 25 µC/cm²
		TERRatio:               580.0,  // 5.8×10² ON/OFF

		NumConductanceStates:   128,    // 7-bit
		WriteVoltage:           3.0,    // 3 V
		WritePulseWidth:        50.0,   // 50 ns
		ReadVoltage:            0.5,    // 0.5 V
		Endurance:              2e8,    // 2×10⁸ cycles
		RetentionSeconds:       1e5,    // 10⁵ s @ 160°C
	}
}

// FTJCell represents a single FTJ device
type FTJCell struct {
	Config              *FTJConfig
	PolarizationState   float64 // Normalized polarization (-1 to +1)
	ConductanceState    int     // Discrete conductance level
	Conductance         float64 // Current conductance (S)
	WriteCycles         int64
	LastWriteTime       float64
}

// FTJ implements Ferroelectric Tunnel Junction array
type FTJ struct {
	Config      *FTJConfig
	Cells       [][]*FTJCell
	Rows        int
	Cols        int
	Stats       *FTJStats
}

// FTJStats tracks FTJ statistics
type FTJStats struct {
	TotalWrites       int64
	TotalReads        int64
	WriteEnergy       float64 // fJ
	ReadEnergy        float64 // fJ
	AverageLinearity  float64 // Weight update linearity
	CycleToCycleVar   float64 // σ/µ variation
}

// NewFTJ creates a new FTJ array
func NewFTJ(config *FTJConfig, rows, cols int) *FTJ {
	ftj := &FTJ{
		Config: config,
		Cells:  make([][]*FTJCell, rows),
		Rows:   rows,
		Cols:   cols,
		Stats:  &FTJStats{},
	}

	// Calculate conductance range from TER
	gMax := 1e-6  // 1 µS max conductance
	gMin := gMax / config.TERRatio

	for i := 0; i < rows; i++ {
		ftj.Cells[i] = make([]*FTJCell, cols)
		for j := 0; j < cols; j++ {
			ftj.Cells[i][j] = &FTJCell{
				Config:            config,
				PolarizationState: 0,
				ConductanceState:  config.NumConductanceStates / 2,
				Conductance:       (gMax + gMin) / 2,
			}
		}
	}

	return ftj
}

// calculateConductance computes conductance from polarization
func (c *FTJCell) calculateConductance() float64 {
	// FTJ tunneling current: I ∝ exp(-2κd)
	// where κ depends on barrier height, which depends on polarization

	// Conductance range
	gMax := 1e-6  // 1 µS
	gMin := gMax / c.Config.TERRatio

	// Map polarization to conductance (nonlinear due to tunneling)
	// P = +1 → low barrier → high conductance
	// P = -1 → high barrier → low conductance
	p := (c.PolarizationState + 1) / 2 // Normalize to [0, 1]

	// Tunneling exponential dependence
	barrierModulation := c.Config.FerroelectricThickness * (1 - 0.3*p)
	tunnelFactor := math.Exp(-0.5 * barrierModulation)

	c.Conductance = gMin + (gMax-gMin)*tunnelFactor
	return c.Conductance
}

// ProgramWeight programs a weight into an FTJ cell
func (f *FTJ) ProgramWeight(row, col int, targetWeight float64) error {
	if row < 0 || row >= f.Rows || col < 0 || col >= f.Cols {
		return fmt.Errorf("index out of bounds: (%d, %d)", row, col)
	}

	cell := f.Cells[row][col]

	// Map weight to conductance state
	// Weight in [-1, 1] maps to state [0, NumStates-1]
	targetState := int((targetWeight + 1) / 2 * float64(cell.Config.NumConductanceStates-1))
	if targetState < 0 {
		targetState = 0
	}
	if targetState >= cell.Config.NumConductanceStates {
		targetState = cell.Config.NumConductanceStates - 1
	}

	// Identical pulse programming (IPP) scheme
	currentState := cell.ConductanceState
	pulsesNeeded := abs(targetState - currentState)

	for pulse := 0; pulse < pulsesNeeded; pulse++ {
		if targetState > currentState {
			// Potentiation pulse (positive voltage)
			cell.PolarizationState += 2.0 / float64(cell.Config.NumConductanceStates)
			if cell.PolarizationState > 1 {
				cell.PolarizationState = 1
			}
			currentState++
		} else {
			// Depression pulse (negative voltage)
			cell.PolarizationState -= 2.0 / float64(cell.Config.NumConductanceStates)
			if cell.PolarizationState < -1 {
				cell.PolarizationState = -1
			}
			currentState--
		}
		cell.WriteCycles++
		f.Stats.TotalWrites++

		// Write energy: ~1 fJ per pulse for FTJ
		f.Stats.WriteEnergy += 1.0
	}

	cell.ConductanceState = targetState
	cell.calculateConductance()

	// Add cycle-to-cycle variation (2.75% from ACS paper)
	variation := rand.NormFloat64() * 0.0275 * cell.Conductance
	cell.Conductance += variation

	return nil
}

// ReadWeight reads a weight from an FTJ cell
func (f *FTJ) ReadWeight(row, col int) (float64, error) {
	if row < 0 || row >= f.Rows || col < 0 || col >= f.Cols {
		return 0, fmt.Errorf("index out of bounds: (%d, %d)", row, col)
	}

	cell := f.Cells[row][col]
	f.Stats.TotalReads++
	f.Stats.ReadEnergy += 0.1 // ~0.1 fJ per read

	// Convert conductance to weight [-1, 1]
	gMax := 1e-6
	gMin := gMax / cell.Config.TERRatio

	weight := 2*(cell.Conductance-gMin)/(gMax-gMin) - 1
	return weight, nil
}

// ComputeMVM performs matrix-vector multiplication using FTJ array
func (f *FTJ) ComputeMVM(input []float64) ([]float64, error) {
	if len(input) != f.Cols {
		return nil, fmt.Errorf("input dimension mismatch: got %d, expected %d", len(input), f.Cols)
	}

	output := make([]float64, f.Rows)

	for i := 0; i < f.Rows; i++ {
		sum := 0.0
		for j := 0; j < f.Cols; j++ {
			// Current = Voltage × Conductance
			// Use conductance directly for analog multiply
			current := input[j] * f.Cells[i][j].Conductance * 1e6 // Scale for numerical stability
			sum += current
		}
		output[i] = sum
	}

	return output, nil
}

// abs returns absolute value of int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// =============================================================================
// FTJ SYNAPSE FOR NEUROMORPHIC COMPUTING
// =============================================================================

// FTJSynapseConfig configures FTJ-based synapse
type FTJSynapseConfig struct {
	FTJConfig          *FTJConfig
	LearningRate       float64
	PotentiationVolt   float64 // Potentiation voltage
	DepressionVolt     float64 // Depression voltage
	STDPEnabled        bool    // Spike-timing dependent plasticity
	STDPWindow         float64 // STDP time window (ms)
}

// DefaultFTJSynapseConfig returns default FTJ synapse configuration
func DefaultFTJSynapseConfig() *FTJSynapseConfig {
	return &FTJSynapseConfig{
		FTJConfig:        DefaultFTJConfig(),
		LearningRate:     0.01,
		PotentiationVolt: 2.5,
		DepressionVolt:   -2.5,
		STDPEnabled:      true,
		STDPWindow:       20.0, // 20 ms
	}
}

// FTJSynapse implements FTJ-based synaptic device
type FTJSynapse struct {
	Config           *FTJSynapseConfig
	FTJArray         *FTJ
	PreSpikeTime     [][]float64 // Last pre-synaptic spike time
	PostSpikeTime    []float64   // Last post-synaptic spike time
	Stats            *FTJSynapseStats
}

// FTJSynapseStats tracks synapse statistics
type FTJSynapseStats struct {
	PotentiationEvents int64
	DepressionEvents   int64
	TotalUpdates       int64
	AverageWeightChange float64
}

// NewFTJSynapse creates a new FTJ synapse array
func NewFTJSynapse(config *FTJSynapseConfig, numPre, numPost int) *FTJSynapse {
	synapse := &FTJSynapse{
		Config:        config,
		FTJArray:      NewFTJ(config.FTJConfig, numPost, numPre),
		PreSpikeTime:  make([][]float64, numPost),
		PostSpikeTime: make([]float64, numPost),
		Stats:         &FTJSynapseStats{},
	}

	for i := 0; i < numPost; i++ {
		synapse.PreSpikeTime[i] = make([]float64, numPre)
		for j := 0; j < numPre; j++ {
			synapse.PreSpikeTime[i][j] = -1000 // No spike yet
		}
		synapse.PostSpikeTime[i] = -1000
	}

	return synapse
}

// ProcessPreSpike processes a pre-synaptic spike
func (s *FTJSynapse) ProcessPreSpike(preIdx int, postIdx int, currentTime float64) {
	if preIdx < 0 || preIdx >= s.FTJArray.Cols || postIdx < 0 || postIdx >= s.FTJArray.Rows {
		return
	}

	s.PreSpikeTime[postIdx][preIdx] = currentTime

	if !s.Config.STDPEnabled {
		return
	}

	// STDP: check if post spike occurred recently
	dt := currentTime - s.PostSpikeTime[postIdx]
	if dt > 0 && dt < s.Config.STDPWindow {
		// Pre after post → depression (LTD)
		s.applySTDP(preIdx, postIdx, -dt)
		s.Stats.DepressionEvents++
	}
}

// ProcessPostSpike processes a post-synaptic spike
func (s *FTJSynapse) ProcessPostSpike(postIdx int, currentTime float64) {
	if postIdx < 0 || postIdx >= s.FTJArray.Rows {
		return
	}

	s.PostSpikeTime[postIdx] = currentTime

	if !s.Config.STDPEnabled {
		return
	}

	// STDP: check all pre-synaptic spikes
	for preIdx := 0; preIdx < s.FTJArray.Cols; preIdx++ {
		dt := currentTime - s.PreSpikeTime[postIdx][preIdx]
		if dt > 0 && dt < s.Config.STDPWindow {
			// Post after pre → potentiation (LTP)
			s.applySTDP(preIdx, postIdx, dt)
			s.Stats.PotentiationEvents++
		}
	}
}

// applySTDP applies STDP weight update
func (s *FTJSynapse) applySTDP(preIdx, postIdx int, dt float64) {
	// STDP curve: ΔW = A × exp(-|dt|/τ)
	tau := s.Config.STDPWindow / 2
	amplitude := s.Config.LearningRate

	deltaW := amplitude * math.Exp(-math.Abs(dt)/tau)
	if dt < 0 {
		deltaW = -deltaW // Depression
	}

	// Get current weight and update
	currentWeight, _ := s.FTJArray.ReadWeight(postIdx, preIdx)
	newWeight := currentWeight + deltaW

	// Clamp to [-1, 1]
	if newWeight > 1 {
		newWeight = 1
	}
	if newWeight < -1 {
		newWeight = -1
	}

	s.FTJArray.ProgramWeight(postIdx, preIdx, newWeight)
	s.Stats.TotalUpdates++
	s.Stats.AverageWeightChange += math.Abs(deltaW)
}

// Forward performs forward pass through FTJ synapse
func (s *FTJSynapse) Forward(input []float64) ([]float64, error) {
	return s.FTJArray.ComputeMVM(input)
}

// =============================================================================
// SECURE CIM ACCELERATOR
// =============================================================================

// SecureCIMConfig configures secure CIM accelerator
type SecureCIMConfig struct {
	PUFConfig           *PUFConfig
	AdversarialConfig   *AdversarialConfig
	SideChannelDefense  bool
	EncryptedWeights    bool
	ObfuscationLevel    int // 0-3
}

// DefaultSecureCIMConfig returns default secure CIM configuration
func DefaultSecureCIMConfig() *SecureCIMConfig {
	return &SecureCIMConfig{
		PUFConfig:          DefaultPUFConfig(),
		AdversarialConfig:  DefaultAdversarialConfig(),
		SideChannelDefense: true,
		EncryptedWeights:   false,
		ObfuscationLevel:   2,
	}
}

// SecureCIM implements a secure CIM accelerator
type SecureCIM struct {
	Config             *SecureCIMConfig
	PUF                *PUF
	AdversarialDetector *AdversarialDetector
	SideChannelDefense *SideChannelDefense
	FTJArray           *FTJ
	Stats              *SecureCIMStats
}

// SideChannelDefense implements countermeasures
type SideChannelDefense struct {
	RandomDelay       bool
	PowerBalancing    bool
	NoiseInjection    bool
	ShuffledExecution bool
}

// SecureCIMStats tracks secure CIM statistics
type SecureCIMStats struct {
	AuthenticationAttempts int64
	AuthenticationSuccess  int64
	AdversarialDetected    int64
	SideChannelBlocked     int64
}

// NewSecureCIM creates a new secure CIM accelerator
func NewSecureCIM(config *SecureCIMConfig, arraySize int) *SecureCIM {
	scim := &SecureCIM{
		Config:              config,
		PUF:                 NewPUF(config.PUFConfig),
		AdversarialDetector: NewAdversarialDetector(config.AdversarialConfig, arraySize),
		SideChannelDefense: &SideChannelDefense{
			RandomDelay:       config.SideChannelDefense,
			PowerBalancing:    config.SideChannelDefense,
			NoiseInjection:    config.SideChannelDefense,
			ShuffledExecution: config.ObfuscationLevel >= 2,
		},
		FTJArray: NewFTJ(DefaultFTJConfig(), arraySize, arraySize),
		Stats:    &SecureCIMStats{},
	}

	return scim
}

// Authenticate performs PUF-based authentication
func (scim *SecureCIM) Authenticate(challenge []byte, expectedResponse []byte) bool {
	scim.Stats.AuthenticationAttempts++

	response := scim.PUF.GenerateResponse(challenge)

	// Compare responses (allow some noise)
	hd := 0
	for i := range response {
		if i < len(expectedResponse) {
			hd += hammingWeight(response[i] ^ expectedResponse[i])
		}
	}

	threshold := len(response) * 8 / 10 // 10% tolerance
	if hd <= threshold {
		scim.Stats.AuthenticationSuccess++
		return true
	}

	return false
}

// SecureInference performs secure inference with defenses
func (scim *SecureCIM) SecureInference(input []float64) ([]float64, error) {
	// Check for adversarial input
	isAdversarial, score := scim.AdversarialDetector.DetectAdversarial(input)
	if isAdversarial {
		scim.Stats.AdversarialDetected++
		// Apply defense
		input = scim.AdversarialDetector.ApplyDefense(input)
	}
	_ = score

	// Apply side-channel defenses
	if scim.SideChannelDefense.RandomDelay {
		// Random delay to prevent timing attacks
		// (In real hardware, this would be actual delay)
	}

	if scim.SideChannelDefense.ShuffledExecution {
		// Shuffle computation order (simplified)
		// Real implementation would shuffle crossbar row access order
	}

	// Perform FTJ-based computation
	output, err := scim.FTJArray.ComputeMVM(input)
	if err != nil {
		return nil, err
	}

	if scim.SideChannelDefense.NoiseInjection {
		// Add small noise to output to mask power signature
		for i := range output {
			output[i] += rand.NormFloat64() * 0.001
		}
	}

	return output, nil
}

// =============================================================================
// DEMONSTRATION FUNCTIONS
// =============================================================================

// DemoFTJSynapse demonstrates FTJ-based synaptic device
func DemoFTJSynapse() {
	fmt.Println("=== Ferroelectric Tunnel Junction Synapse Demo ===")
	fmt.Println()

	config := DefaultFTJConfig()
	ftj := NewFTJ(config, 64, 64)

	fmt.Printf("FTJ Configuration:\n")
	fmt.Printf("  HZO thickness: %.1f nm\n", config.FerroelectricThickness)
	fmt.Printf("  TER ratio: %.0f (ON/OFF)\n", config.TERRatio)
	fmt.Printf("  Conductance states: %d\n", config.NumConductanceStates)
	fmt.Printf("  Write pulse: %.0f ns\n", config.WritePulseWidth)
	fmt.Printf("  Endurance: %.0e cycles\n", float64(config.Endurance))
	fmt.Println()

	// Program weights with gradient pattern
	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			weight := float64(i+j)/126.0*2 - 1 // Gradient from -1 to 1
			ftj.ProgramWeight(i, j, weight)
		}
	}

	// Test MVM
	input := make([]float64, 64)
	for i := range input {
		input[i] = rand.Float64()
	}

	output, _ := ftj.ComputeMVM(input)

	fmt.Printf("FTJ Statistics:\n")
	fmt.Printf("  Total writes: %d\n", ftj.Stats.TotalWrites)
	fmt.Printf("  Write energy: %.1f fJ\n", ftj.Stats.WriteEnergy)
	fmt.Printf("  Output[0:5]: %.4f, %.4f, %.4f, %.4f, %.4f\n",
		output[0], output[1], output[2], output[3], output[4])
}

// DemoPUFSecurity demonstrates PUF-based security
func DemoPUFSecurity() {
	fmt.Println("=== PUF-Based CIM Security Demo ===")
	fmt.Println()

	config := DefaultPUFConfig()
	puf := NewPUF(config)

	fmt.Printf("PUF Configuration:\n")
	fmt.Printf("  Type: %s\n", config.Type)
	fmt.Printf("  Challenge bits: %d\n", config.ChallengeBits)
	fmt.Printf("  Response bits: %d\n", config.ResponseBits)
	fmt.Println()

	// Evaluate PUF metrics
	puf.EvaluatePUF(1000)

	fmt.Printf("PUF Quality Metrics:\n")
	fmt.Printf("  Uniformity: %.2f%% (ideal: 50%%)\n", puf.Stats.Uniformity*100)
	fmt.Printf("  Reliability: %.2f%%\n", puf.Stats.Reliability*100)
	fmt.Printf("  Uniqueness: %.2f\n", puf.Stats.Uniqueness)
	fmt.Println()

	// Authentication test
	challenge := make([]byte, config.ChallengeBits/8)
	rand.Read(challenge)

	response1 := puf.GenerateResponse(challenge)
	response2 := puf.GenerateResponse(challenge)

	hd := 0
	for i := range response1 {
		hd += hammingWeight(response1[i] ^ response2[i])
	}
	fmt.Printf("Authentication Test:\n")
	fmt.Printf("  Response consistency: %d/%d bits differ\n", hd, config.ResponseBits)
}

// DemoSecureCIM demonstrates secure CIM accelerator
func DemoSecureCIM() {
	fmt.Println("=== Secure CIM Accelerator Demo ===")
	fmt.Println()

	config := DefaultSecureCIMConfig()
	scim := NewSecureCIM(config, 64)

	// Train adversarial detector
	for i := 0; i < 1000; i++ {
		input := make([]float64, 64)
		for j := range input {
			input[j] = rand.NormFloat64() * 0.5
		}
		scim.AdversarialDetector.UpdateStatistics(input)
	}
	scim.AdversarialDetector.FinalizeStatistics()

	// Test with normal input
	normalInput := make([]float64, 64)
	for i := range normalInput {
		normalInput[i] = rand.NormFloat64() * 0.5
	}

	isAdv, score := scim.AdversarialDetector.DetectAdversarial(normalInput)
	fmt.Printf("Normal input detection: adversarial=%v, score=%.2f\n", isAdv, score)

	// Test with adversarial input
	advInput := make([]float64, 64)
	for i := range advInput {
		advInput[i] = rand.NormFloat64() * 0.5
		if i%10 == 0 {
			advInput[i] += 5.0 // Large perturbation
		}
	}

	isAdv, score = scim.AdversarialDetector.DetectAdversarial(advInput)
	fmt.Printf("Adversarial input detection: adversarial=%v, score=%.2f\n", isAdv, score)
	fmt.Println()

	// Run secure inference
	output, _ := scim.SecureInference(normalInput)
	fmt.Printf("Secure inference output[0:5]: %.4f, %.4f, %.4f, %.4f, %.4f\n",
		output[0], output[1], output[2], output[3], output[4])

	fmt.Printf("\nSecurity Statistics:\n")
	fmt.Printf("  Adversarial detected: %d\n", scim.Stats.AdversarialDetected)
	fmt.Printf("  Side-channel defense: %v\n", config.SideChannelDefense)
}

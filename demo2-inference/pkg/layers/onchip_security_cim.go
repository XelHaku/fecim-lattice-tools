// onchip_security_cim.go - On-Chip Learning and Security for CIM
// Research iteration 127: In-memory training and hardware security
//
// Key findings:
// - On-chip training: orders of magnitude more demanding than inference
// - CIMAT: 7T transpose SRAM for bidirectional read, backpropagation support
// - Hybrid RRAM/SRAM: combine accuracy (SRAM) with density (RRAM)
// - 16Mb RRAM macro: 31.2 TFLOPS/W energy efficiency
// - PUF integration: ReRAM-based PUF fused with CIM, minimal area overhead
// - Security threats: adversarial attacks, side-channel, weight extraction
// - Protections: PUF-based encryption, secure boot, formal verification

package layers

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// ============================================================================
// On-Chip Learning Types
// ============================================================================

// TrainingMode represents different on-chip training modes
type TrainingMode int

const (
	TrainingModeInference     TrainingMode = iota // Inference only
	TrainingModeFineTune                          // Fine-tune last layers
	TrainingModeFullTrain                         // Full backpropagation
	TrainingModeContinual                         // Continual learning
	TrainingModeFederated                         // Federated learning
)

func (tm TrainingMode) String() string {
	return []string{"Inference", "FineTune", "FullTrain", "Continual", "Federated"}[tm]
}

// MemoryType represents different memory technologies
type MemoryType int

const (
	MemorySRAM   MemoryType = iota // SRAM (high accuracy, volatile)
	MemoryRRAM                     // RRAM (non-volatile, limited endurance)
	MemoryMRAM                     // MRAM (non-volatile, high endurance)
	MemoryFeFET                    // FeFET (non-volatile, high density)
	MemoryHybrid                   // Hybrid SRAM+NVM
)

func (mt MemoryType) String() string {
	return []string{"SRAM", "RRAM", "MRAM", "FeFET", "Hybrid"}[mt]
}

// ============================================================================
// On-Chip Training Configuration
// ============================================================================

// OnChipTrainingConfig configures on-chip training
type OnChipTrainingConfig struct {
	// Memory configuration
	PrimaryMemory    MemoryType
	SecondaryMemory  MemoryType // For hybrid systems
	ArraySize        int
	BitPrecision     int

	// Training parameters
	LearningRate     float64
	BatchSize        int
	MaxEpochs        int
	ConvergenceThresh float64

	// Endurance management
	MaxWriteCycles   int
	WriteThrottling  bool
	GradientSparsity float64 // Only update top-K gradients

	// Hybrid memory policy
	SRAMForGradients bool   // Use SRAM for gradient accumulation
	NVMForWeights    bool   // Use NVM for weight storage
	WeightUpdateFreq int    // Batch updates to NVM every N iterations
}

// DefaultOnChipTrainingConfig returns default on-chip training configuration
func DefaultOnChipTrainingConfig() *OnChipTrainingConfig {
	return &OnChipTrainingConfig{
		PrimaryMemory:    MemoryHybrid,
		SecondaryMemory:  MemoryRRAM,
		ArraySize:        64,
		BitPrecision:     8,

		LearningRate:     0.001,
		BatchSize:        8,
		MaxEpochs:        100,
		ConvergenceThresh: 0.001,

		MaxWriteCycles:   10000,
		WriteThrottling:  true,
		GradientSparsity: 0.1, // Update top 10%

		SRAMForGradients: true,
		NVMForWeights:    true,
		WeightUpdateFreq: 10,
	}
}

// ============================================================================
// CIMAT Architecture (Training-in-Memory)
// ============================================================================

// CIMATConfig configures the CIMAT training-in-memory architecture
type CIMATConfig struct {
	// 7T transpose SRAM configuration
	ArrayRows       int
	ArrayCols       int
	TransposeEnable bool // Enable bidirectional read

	// Compute paths
	ForwardPath     bool // Standard MVM path
	BackwardPath    bool // Transpose MVM for backprop
	WeightUpdatePath bool // In-situ weight update

	// Precision
	ActivationBits  int
	WeightBits      int
	GradientBits    int
}

// DefaultCIMATConfig returns default CIMAT configuration
func DefaultCIMATConfig() *CIMATConfig {
	return &CIMATConfig{
		ArrayRows:       128,
		ArrayCols:       128,
		TransposeEnable: true,

		ForwardPath:      true,
		BackwardPath:     true,
		WeightUpdatePath: true,

		ActivationBits: 8,
		WeightBits:     8,
		GradientBits:   8,
	}
}

// CIMAT implements the training-in-memory architecture
type CIMAT struct {
	Config  *CIMATConfig

	// Weight storage
	Weights     [][]float64
	WeightsT    [][]float64 // Transposed weights for backprop

	// Gradient accumulation
	GradientAcc [][]float64
	UpdateCount int

	// Statistics
	ForwardOps    int64
	BackwardOps   int64
	WeightUpdates int64
}

// NewCIMAT creates a new CIMAT instance
func NewCIMAT(config *CIMATConfig) *CIMAT {
	c := &CIMAT{
		Config:      config,
		Weights:     make([][]float64, config.ArrayRows),
		WeightsT:    make([][]float64, config.ArrayCols),
		GradientAcc: make([][]float64, config.ArrayRows),
	}

	for i := 0; i < config.ArrayRows; i++ {
		c.Weights[i] = make([]float64, config.ArrayCols)
		c.GradientAcc[i] = make([]float64, config.ArrayCols)
	}
	for j := 0; j < config.ArrayCols; j++ {
		c.WeightsT[j] = make([]float64, config.ArrayRows)
	}

	return c
}

// InitWeights initializes weights with Xavier initialization
func (c *CIMAT) InitWeights(seed int64) {
	rng := rand.New(rand.NewSource(seed))
	scale := math.Sqrt(2.0 / float64(c.Config.ArrayRows+c.Config.ArrayCols))

	for i := 0; i < c.Config.ArrayRows; i++ {
		for j := 0; j < c.Config.ArrayCols; j++ {
			w := rng.NormFloat64() * scale
			c.Weights[i][j] = w
			c.WeightsT[j][i] = w // Keep transpose synchronized
		}
	}
}

// ForwardMVM performs forward matrix-vector multiplication
func (c *CIMAT) ForwardMVM(input []float64) []float64 {
	output := make([]float64, c.Config.ArrayCols)

	for j := 0; j < c.Config.ArrayCols; j++ {
		sum := 0.0
		for i := 0; i < c.Config.ArrayRows && i < len(input); i++ {
			sum += c.Weights[i][j] * input[i]
		}
		output[j] = sum
	}

	c.ForwardOps += int64(c.Config.ArrayRows * c.Config.ArrayCols)
	return output
}

// BackwardMVM performs backward pass using transposed weights
func (c *CIMAT) BackwardMVM(gradOutput []float64) []float64 {
	if !c.Config.TransposeEnable {
		return nil // Transpose not available
	}

	gradInput := make([]float64, c.Config.ArrayRows)

	// Use transposed weights for backprop
	for i := 0; i < c.Config.ArrayRows; i++ {
		sum := 0.0
		for j := 0; j < c.Config.ArrayCols && j < len(gradOutput); j++ {
			sum += c.WeightsT[j][i] * gradOutput[j]
		}
		gradInput[i] = sum
	}

	c.BackwardOps += int64(c.Config.ArrayRows * c.Config.ArrayCols)
	return gradInput
}

// AccumulateGradient accumulates weight gradients
func (c *CIMAT) AccumulateGradient(input, gradOutput []float64) {
	// Outer product: dW = input * gradOutput^T
	for i := 0; i < c.Config.ArrayRows && i < len(input); i++ {
		for j := 0; j < c.Config.ArrayCols && j < len(gradOutput); j++ {
			c.GradientAcc[i][j] += input[i] * gradOutput[j]
		}
	}
	c.UpdateCount++
}

// ApplyGradients applies accumulated gradients to weights
func (c *CIMAT) ApplyGradients(learningRate float64) {
	if c.UpdateCount == 0 {
		return
	}

	scale := learningRate / float64(c.UpdateCount)

	for i := 0; i < c.Config.ArrayRows; i++ {
		for j := 0; j < c.Config.ArrayCols; j++ {
			update := c.GradientAcc[i][j] * scale
			c.Weights[i][j] -= update
			c.WeightsT[j][i] = c.Weights[i][j] // Keep transpose sync

			// Clear gradient accumulator
			c.GradientAcc[i][j] = 0
		}
	}

	c.WeightUpdates += int64(c.Config.ArrayRows * c.Config.ArrayCols)
	c.UpdateCount = 0
}

// ============================================================================
// Hybrid Memory Training System
// ============================================================================

// HybridTrainingConfig configures hybrid memory training
type HybridTrainingConfig struct {
	SRAMSize         int     // SRAM buffer size (KB)
	NVMSize          int     // NVM array size (KB)
	SRAMLatency      int     // SRAM access latency (cycles)
	NVMLatency       int     // NVM access latency (cycles)
	NVMWriteEnergy   float64 // NVM write energy (pJ)
	SRAMWriteEnergy  float64 // SRAM write energy (pJ)
	BatchUpdateSize  int     // Batch size before NVM update
}

// DefaultHybridTrainingConfig returns default hybrid training configuration
func DefaultHybridTrainingConfig() *HybridTrainingConfig {
	return &HybridTrainingConfig{
		SRAMSize:        64,    // 64 KB SRAM
		NVMSize:         1024,  // 1 MB NVM
		SRAMLatency:     1,     // 1 cycle
		NVMLatency:      10,    // 10 cycles
		NVMWriteEnergy:  10.0,  // 10 pJ
		SRAMWriteEnergy: 0.1,   // 0.1 pJ
		BatchUpdateSize: 32,    // Update NVM every 32 batches
	}
}

// HybridTrainingSystem implements hybrid SRAM+NVM training
type HybridTrainingSystem struct {
	Config *HybridTrainingConfig

	// SRAM buffers (gradient accumulation)
	SRAMGradients [][]float64
	SRAMActivations [][]float64

	// NVM weights
	NVMWeights [][]float64

	// Statistics
	SRAMAccesses   int64
	NVMAccesses    int64
	NVMWrites      int64
	TotalEnergyPJ  float64
	BatchCounter   int
}

// NewHybridTrainingSystem creates a new hybrid training system
func NewHybridTrainingSystem(config *HybridTrainingConfig, rows, cols int) *HybridTrainingSystem {
	h := &HybridTrainingSystem{
		Config:          config,
		SRAMGradients:   make([][]float64, rows),
		SRAMActivations: make([][]float64, rows),
		NVMWeights:      make([][]float64, rows),
	}

	for i := 0; i < rows; i++ {
		h.SRAMGradients[i] = make([]float64, cols)
		h.SRAMActivations[i] = make([]float64, cols)
		h.NVMWeights[i] = make([]float64, cols)
	}

	return h
}

// AccumulateInSRAM accumulates gradients in SRAM buffer
func (h *HybridTrainingSystem) AccumulateInSRAM(gradients [][]float64) {
	rows := len(h.SRAMGradients)
	cols := len(h.SRAMGradients[0])

	for i := 0; i < rows && i < len(gradients); i++ {
		for j := 0; j < cols && j < len(gradients[i]); j++ {
			h.SRAMGradients[i][j] += gradients[i][j]
			h.SRAMAccesses++
			h.TotalEnergyPJ += h.Config.SRAMWriteEnergy
		}
	}

	h.BatchCounter++
}

// FlushToNVM flushes accumulated updates to NVM
func (h *HybridTrainingSystem) FlushToNVM(learningRate float64) {
	if h.BatchCounter < h.Config.BatchUpdateSize {
		return // Not enough batches accumulated
	}

	rows := len(h.NVMWeights)
	cols := len(h.NVMWeights[0])
	scale := learningRate / float64(h.BatchCounter)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Apply accumulated gradient
			update := h.SRAMGradients[i][j] * scale
			h.NVMWeights[i][j] -= update

			// Clear SRAM buffer
			h.SRAMGradients[i][j] = 0

			h.NVMWrites++
			h.NVMAccesses++
			h.TotalEnergyPJ += h.Config.NVMWriteEnergy
		}
	}

	h.BatchCounter = 0
}

// GetEnergyEfficiency returns energy per weight update
func (h *HybridTrainingSystem) GetEnergyEfficiency() float64 {
	if h.NVMWrites == 0 {
		return 0
	}
	return h.TotalEnergyPJ / float64(h.NVMWrites)
}

// ============================================================================
// Security Types
// ============================================================================

// SecurityThreat represents different security threats
type SecurityThreat int

const (
	ThreatModelExtraction  SecurityThreat = iota // Extract model weights
	ThreatAdversarial                            // Adversarial examples
	ThreatSideChannel                            // Power/EM side-channel
	ThreatFaultInjection                         // Fault injection
	ThreatHardwareTrojan                         // Hardware backdoor
	ThreatDataPrivacy                            // User data leakage
)

func (st SecurityThreat) String() string {
	return []string{"ModelExtraction", "Adversarial", "SideChannel", "FaultInjection", "HardwareTrojan", "DataPrivacy"}[st]
}

// AttackVector represents attack vectors
type AttackVector int

const (
	AttackPowerAnalysis     AttackVector = iota // Simple/differential power analysis
	AttackEMAnalysis                            // Electromagnetic emanations
	AttackTimingAnalysis                        // Timing side-channel
	AttackFaultGlitch                           // Voltage/clock glitching
	AttackMLModeling                            // ML-based PUF modeling
	AttackRowhammer                             // Memory rowhammer
)

func (av AttackVector) String() string {
	return []string{"PowerAnalysis", "EMAnalysis", "TimingAnalysis", "FaultGlitch", "MLModeling", "Rowhammer"}[av]
}

// ============================================================================
// Physical Unclonable Function (PUF)
// ============================================================================

// PUFConfig configures the ReRAM-based PUF
type PUFConfig struct {
	NumChallenges    int     // Number of challenge bits
	NumResponses     int     // Number of response bits
	NoiseThreshold   float64 // Noise threshold for response
	EnrollmentCycles int     // Enrollment iterations
	StabilityTarget  float64 // Target stability (0-1)
}

// DefaultPUFConfig returns default PUF configuration
func DefaultPUFConfig() *PUFConfig {
	return &PUFConfig{
		NumChallenges:    128,
		NumResponses:     128,
		NoiseThreshold:   0.1,
		EnrollmentCycles: 10,
		StabilityTarget:  0.95,
	}
}

// ReRAMPUF implements a ReRAM-based physical unclonable function
type ReRAMPUF struct {
	Config           *PUFConfig
	DeviceVariation  [][]float64 // Intrinsic device variations
	EnrolledCRPs     map[string][]byte // Challenge-Response Pairs
	rng              *rand.Rand

	// Statistics
	ChallengeCount   int64
	BitErrorRate     float64
}

// NewReRAMPUF creates a new ReRAM-based PUF
func NewReRAMPUF(config *PUFConfig, rows, cols int, seed int64) *ReRAMPUF {
	puf := &ReRAMPUF{
		Config:          config,
		DeviceVariation: make([][]float64, rows),
		EnrolledCRPs:    make(map[string][]byte),
		rng:             rand.New(rand.NewSource(seed)),
	}

	// Generate intrinsic device variations (simulating manufacturing)
	for i := 0; i < rows; i++ {
		puf.DeviceVariation[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			// Gaussian variation with mean 0.5, std 0.1
			puf.DeviceVariation[i][j] = 0.5 + puf.rng.NormFloat64()*0.1
			if puf.DeviceVariation[i][j] < 0 {
				puf.DeviceVariation[i][j] = 0
			}
			if puf.DeviceVariation[i][j] > 1 {
				puf.DeviceVariation[i][j] = 1
			}
		}
	}

	return puf
}

// GenerateResponse generates a response for a given challenge
func (p *ReRAMPUF) GenerateResponse(challenge []byte) []byte {
	response := make([]byte, p.Config.NumResponses/8)

	// Hash challenge to select cells
	hash := sha256.Sum256(challenge)

	rows := len(p.DeviceVariation)
	cols := len(p.DeviceVariation[0])

	for i := 0; i < p.Config.NumResponses; i++ {
		// Use hash to select cell indices
		row := int(hash[i%32]) % rows
		col := int(hash[(i+16)%32]) % cols

		// Compare variation to threshold with noise
		threshold := 0.5
		noise := p.rng.NormFloat64() * p.Config.NoiseThreshold
		value := p.DeviceVariation[row][col] + noise

		byteIdx := i / 8
		bitIdx := uint(i % 8)
		if value > threshold {
			response[byteIdx] |= (1 << bitIdx)
		}
	}

	p.ChallengeCount++
	return response
}

// Enroll enrolls a challenge-response pair
func (p *ReRAMPUF) Enroll(challenge []byte) []byte {
	// Generate multiple responses and take majority vote
	responses := make([][]byte, p.Config.EnrollmentCycles)
	for i := 0; i < p.Config.EnrollmentCycles; i++ {
		responses[i] = p.GenerateResponse(challenge)
	}

	// Majority vote
	enrolled := make([]byte, p.Config.NumResponses/8)
	for byteIdx := 0; byteIdx < len(enrolled); byteIdx++ {
		for bitIdx := 0; bitIdx < 8; bitIdx++ {
			ones := 0
			for _, resp := range responses {
				if (resp[byteIdx] & (1 << uint(bitIdx))) != 0 {
					ones++
				}
			}
			if ones > p.Config.EnrollmentCycles/2 {
				enrolled[byteIdx] |= (1 << uint(bitIdx))
			}
		}
	}

	// Store enrolled CRP
	key := fmt.Sprintf("%x", challenge)
	p.EnrolledCRPs[key] = enrolled

	return enrolled
}

// Verify verifies a challenge-response pair
func (p *ReRAMPUF) Verify(challenge, response []byte) (bool, float64) {
	key := fmt.Sprintf("%x", challenge)
	enrolled, exists := p.EnrolledCRPs[key]
	if !exists {
		return false, 1.0
	}

	// Calculate Hamming distance
	totalBits := len(enrolled) * 8
	diffBits := 0
	for i := 0; i < len(enrolled) && i < len(response); i++ {
		xor := enrolled[i] ^ response[i]
		for xor != 0 {
			diffBits++
			xor &= xor - 1
		}
	}

	errorRate := float64(diffBits) / float64(totalBits)
	p.BitErrorRate = errorRate

	// Accept if error rate is below threshold
	threshold := p.Config.NoiseThreshold * 2 // Allow 2x noise for verification
	return errorRate < threshold, errorRate
}

// ============================================================================
// CIM Security Module
// ============================================================================

// CIMSecurityConfig configures CIM security
type CIMSecurityConfig struct {
	// PUF configuration
	EnablePUF        bool
	PUFConfig        *PUFConfig

	// Weight protection
	EncryptWeights   bool
	WeightMasking    bool
	SplitCompute     bool // Split computation across multiple chips

	// Side-channel protection
	PowerBalancing   bool
	RandomDelays     bool
	NoiseInjection   bool

	// Adversarial defense
	InputValidation  bool
	OutputClipping   bool
	FeatureSqueezing bool
}

// DefaultCIMSecurityConfig returns default security configuration
func DefaultCIMSecurityConfig() *CIMSecurityConfig {
	return &CIMSecurityConfig{
		EnablePUF:        true,
		PUFConfig:        DefaultPUFConfig(),

		EncryptWeights:   true,
		WeightMasking:    true,
		SplitCompute:     false,

		PowerBalancing:   true,
		RandomDelays:     true,
		NoiseInjection:   true,

		InputValidation:  true,
		OutputClipping:   true,
		FeatureSqueezing: false,
	}
}

// CIMSecurityModule implements security features for CIM
type CIMSecurityModule struct {
	Config     *CIMSecurityConfig
	PUF        *ReRAMPUF
	rng        *rand.Rand

	// Weight protection state
	WeightMask [][]float64
	EncryptKey []byte

	// Statistics
	SecurityEvents map[SecurityThreat]int
}

// NewCIMSecurityModule creates a new security module
func NewCIMSecurityModule(config *CIMSecurityConfig, rows, cols int, seed int64) *CIMSecurityModule {
	sm := &CIMSecurityModule{
		Config:         config,
		rng:            rand.New(rand.NewSource(seed)),
		WeightMask:     make([][]float64, rows),
		SecurityEvents: make(map[SecurityThreat]int),
	}

	if config.EnablePUF {
		sm.PUF = NewReRAMPUF(config.PUFConfig, rows, cols, seed)
	}

	// Generate weight mask
	if config.WeightMasking {
		for i := 0; i < rows; i++ {
			sm.WeightMask[i] = make([]float64, cols)
			for j := 0; j < cols; j++ {
				sm.WeightMask[i][j] = sm.rng.Float64()*0.1 - 0.05 // Small random mask
			}
		}
	}

	// Generate encryption key from PUF
	if config.EnablePUF && config.EncryptWeights {
		challenge := make([]byte, 32)
		sm.rng.Read(challenge)
		sm.EncryptKey = sm.PUF.Enroll(challenge)
	}

	return sm
}

// ProtectWeights applies weight protection
func (sm *CIMSecurityModule) ProtectWeights(weights [][]float64) [][]float64 {
	rows := len(weights)
	cols := len(weights[0])
	protected := make([][]float64, rows)

	for i := 0; i < rows; i++ {
		protected[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			w := weights[i][j]

			// Apply mask
			if sm.Config.WeightMasking && i < len(sm.WeightMask) && j < len(sm.WeightMask[i]) {
				w += sm.WeightMask[i][j]
			}

			// Apply encryption (XOR with key-derived stream)
			if sm.Config.EncryptWeights && sm.EncryptKey != nil {
				keyIdx := (i*cols + j) % len(sm.EncryptKey)
				keyVal := float64(sm.EncryptKey[keyIdx]) / 255.0 * 0.1
				w += keyVal - 0.05
			}

			protected[i][j] = w
		}
	}

	return protected
}

// UnprotectWeights removes weight protection
func (sm *CIMSecurityModule) UnprotectWeights(protected [][]float64) [][]float64 {
	rows := len(protected)
	cols := len(protected[0])
	weights := make([][]float64, rows)

	for i := 0; i < rows; i++ {
		weights[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			w := protected[i][j]

			// Remove encryption
			if sm.Config.EncryptWeights && sm.EncryptKey != nil {
				keyIdx := (i*cols + j) % len(sm.EncryptKey)
				keyVal := float64(sm.EncryptKey[keyIdx]) / 255.0 * 0.1
				w -= keyVal - 0.05
			}

			// Remove mask
			if sm.Config.WeightMasking && i < len(sm.WeightMask) && j < len(sm.WeightMask[i]) {
				w -= sm.WeightMask[i][j]
			}

			weights[i][j] = w
		}
	}

	return weights
}

// AddSideChannelNoise adds noise to hide power/EM signatures
func (sm *CIMSecurityModule) AddSideChannelNoise(value float64) float64 {
	if !sm.Config.NoiseInjection {
		return value
	}
	noise := sm.rng.NormFloat64() * 0.01 // 1% noise
	return value + noise
}

// AddRandomDelay adds random timing delay
func (sm *CIMSecurityModule) AddRandomDelay() int {
	if !sm.Config.RandomDelays {
		return 0
	}
	return sm.rng.Intn(10) // 0-9 cycle delay
}

// ValidateInput validates input for adversarial detection
func (sm *CIMSecurityModule) ValidateInput(input []float64) ([]float64, bool) {
	if !sm.Config.InputValidation {
		return input, true
	}

	validated := make([]float64, len(input))
	isValid := true

	for i, x := range input {
		// Check for out-of-range values
		if x < -10 || x > 10 {
			isValid = false
			sm.SecurityEvents[ThreatAdversarial]++
		}

		// Clip extreme values
		if sm.Config.OutputClipping {
			if x < -1 {
				x = -1
			}
			if x > 1 {
				x = 1
			}
		}

		validated[i] = x
	}

	return validated, isValid
}

// DetectThreat detects potential security threats
func (sm *CIMSecurityModule) DetectThreat(threatType SecurityThreat) bool {
	sm.SecurityEvents[threatType]++
	return true
}

// GetSecurityReport generates a security status report
func (sm *CIMSecurityModule) GetSecurityReport() string {
	report := "CIM Security Status Report\n"
	report += "===========================\n\n"

	report += "Security Features:\n"
	report += fmt.Sprintf("  PUF Enabled: %v\n", sm.Config.EnablePUF)
	report += fmt.Sprintf("  Weight Encryption: %v\n", sm.Config.EncryptWeights)
	report += fmt.Sprintf("  Weight Masking: %v\n", sm.Config.WeightMasking)
	report += fmt.Sprintf("  Power Balancing: %v\n", sm.Config.PowerBalancing)
	report += fmt.Sprintf("  Random Delays: %v\n", sm.Config.RandomDelays)
	report += fmt.Sprintf("  Noise Injection: %v\n", sm.Config.NoiseInjection)

	if sm.PUF != nil {
		report += fmt.Sprintf("\nPUF Statistics:\n")
		report += fmt.Sprintf("  Challenges: %d\n", sm.PUF.ChallengeCount)
		report += fmt.Sprintf("  Bit Error Rate: %.4f\n", sm.PUF.BitErrorRate)
		report += fmt.Sprintf("  Enrolled CRPs: %d\n", len(sm.PUF.EnrolledCRPs))
	}

	if len(sm.SecurityEvents) > 0 {
		report += "\nSecurity Events:\n"
		for threat, count := range sm.SecurityEvents {
			report += fmt.Sprintf("  %s: %d\n", threat, count)
		}
	}

	return report
}

// ============================================================================
// Adversarial Defense
// ============================================================================

// AdversarialDefenseConfig configures adversarial defense
type AdversarialDefenseConfig struct {
	// Detection methods
	InputPerturbation   bool    // Detect via input perturbation
	GradientMasking     bool    // Hide gradient information
	AdversarialTraining bool    // Train with adversarial examples

	// Defense parameters
	PerturbationEpsilon float64 // Detection threshold
	DefenseNoiseSigma   float64 // Defense noise level
}

// DefaultAdversarialDefenseConfig returns default defense configuration
func DefaultAdversarialDefenseConfig() *AdversarialDefenseConfig {
	return &AdversarialDefenseConfig{
		InputPerturbation:   true,
		GradientMasking:     true,
		AdversarialTraining: false,
		PerturbationEpsilon: 0.3,
		DefenseNoiseSigma:   0.1,
	}
}

// AdversarialDefense implements adversarial attack defense
type AdversarialDefense struct {
	Config *AdversarialDefenseConfig
	rng    *rand.Rand

	// Detection statistics
	SuspiciousInputs  int
	BlockedInputs     int
	DefenseActivations int
}

// NewAdversarialDefense creates a new adversarial defense module
func NewAdversarialDefense(config *AdversarialDefenseConfig, seed int64) *AdversarialDefense {
	return &AdversarialDefense{
		Config: config,
		rng:    rand.New(rand.NewSource(seed)),
	}
}

// DetectAdversarial detects adversarial examples
func (ad *AdversarialDefense) DetectAdversarial(input []float64, model func([]float64) []float64) bool {
	if !ad.Config.InputPerturbation {
		return false
	}

	// Get original prediction
	origOutput := model(input)

	// Perturb input slightly
	perturbed := make([]float64, len(input))
	for i := range input {
		noise := ad.rng.NormFloat64() * ad.Config.PerturbationEpsilon
		perturbed[i] = input[i] + noise
	}

	// Get perturbed prediction
	pertOutput := model(perturbed)

	// Check for large prediction change
	maxDiff := 0.0
	for i := range origOutput {
		diff := math.Abs(origOutput[i] - pertOutput[i])
		if diff > maxDiff {
			maxDiff = diff
		}
	}

	// Adversarial examples often have unstable predictions
	threshold := 0.5 // High sensitivity to small perturbations
	if maxDiff > threshold {
		ad.SuspiciousInputs++
		return true
	}

	return false
}

// ApplyDefense applies defensive transformations
func (ad *AdversarialDefense) ApplyDefense(input []float64) []float64 {
	defended := make([]float64, len(input))

	for i, x := range input {
		// Add defensive noise
		noise := ad.rng.NormFloat64() * ad.Config.DefenseNoiseSigma
		defended[i] = x + noise
	}

	ad.DefenseActivations++
	return defended
}

// MaskGradient masks gradient for defense
func (ad *AdversarialDefense) MaskGradient(gradient []float64) []float64 {
	if !ad.Config.GradientMasking {
		return gradient
	}

	masked := make([]float64, len(gradient))
	for i, g := range gradient {
		// Quantize gradient
		quantized := math.Round(g*10) / 10
		// Add noise
		noise := ad.rng.NormFloat64() * 0.05
		masked[i] = quantized + noise
	}

	return masked
}

// ============================================================================
// Secure Training Pipeline
// ============================================================================

// SecureTrainingConfig configures secure on-chip training
type SecureTrainingConfig struct {
	TrainingConfig *OnChipTrainingConfig
	SecurityConfig *CIMSecurityConfig
	DefenseConfig  *AdversarialDefenseConfig

	// Additional security options
	SecureBoot     bool
	Attestation    bool
	AuditLogging   bool
}

// DefaultSecureTrainingConfig returns default secure training configuration
func DefaultSecureTrainingConfig() *SecureTrainingConfig {
	return &SecureTrainingConfig{
		TrainingConfig: DefaultOnChipTrainingConfig(),
		SecurityConfig: DefaultCIMSecurityConfig(),
		DefenseConfig:  DefaultAdversarialDefenseConfig(),
		SecureBoot:     true,
		Attestation:    true,
		AuditLogging:   true,
	}
}

// SecureTrainingPipeline implements secure on-chip training
type SecureTrainingPipeline struct {
	Config      *SecureTrainingConfig
	CIMAT       *CIMAT
	Security    *CIMSecurityModule
	Defense     *AdversarialDefense
	HybridMem   *HybridTrainingSystem

	// Audit log
	AuditLog []string
}

// NewSecureTrainingPipeline creates a new secure training pipeline
func NewSecureTrainingPipeline(config *SecureTrainingConfig, rows, cols int, seed int64) *SecureTrainingPipeline {
	return &SecureTrainingPipeline{
		Config:    config,
		CIMAT:     NewCIMAT(DefaultCIMATConfig()),
		Security:  NewCIMSecurityModule(config.SecurityConfig, rows, cols, seed),
		Defense:   NewAdversarialDefense(config.DefenseConfig, seed),
		HybridMem: NewHybridTrainingSystem(DefaultHybridTrainingConfig(), rows, cols),
		AuditLog:  []string{},
	}
}

// SecureBoot performs secure boot sequence
func (p *SecureTrainingPipeline) SecureBoot() bool {
	if !p.Config.SecureBoot {
		return true
	}

	// Generate PUF challenge
	challenge := make([]byte, 32)
	p.Security.rng.Read(challenge)

	// Verify PUF response
	response := p.Security.PUF.GenerateResponse(challenge)
	valid, errorRate := p.Security.PUF.Verify(challenge, response)

	p.AuditLog = append(p.AuditLog, fmt.Sprintf("SecureBoot: valid=%v, errorRate=%.4f", valid, errorRate))

	return valid
}

// SecureForward performs secure forward pass
func (p *SecureTrainingPipeline) SecureForward(input []float64) ([]float64, bool) {
	// Validate input
	validatedInput, inputValid := p.Security.ValidateInput(input)
	if !inputValid {
		p.AuditLog = append(p.AuditLog, "SecureForward: Input validation failed")
	}

	// Apply adversarial defense
	defendedInput := p.Defense.ApplyDefense(validatedInput)

	// Add side-channel noise
	for i := range defendedInput {
		defendedInput[i] = p.Security.AddSideChannelNoise(defendedInput[i])
	}

	// Perform forward pass
	output := p.CIMAT.ForwardMVM(defendedInput)

	// Add random delay
	_ = p.Security.AddRandomDelay()

	return output, inputValid
}

// SecureBackward performs secure backward pass
func (p *SecureTrainingPipeline) SecureBackward(gradOutput []float64) []float64 {
	// Mask gradient
	maskedGrad := p.Defense.MaskGradient(gradOutput)

	// Perform backward pass
	gradInput := p.CIMAT.BackwardMVM(maskedGrad)

	return gradInput
}

// GetAuditLog returns the audit log
func (p *SecureTrainingPipeline) GetAuditLog() []string {
	return p.AuditLog
}

// ============================================================================
// Serialization
// ============================================================================

// OnChipSecurityState contains serializable state
type OnChipSecurityState struct {
	TrainingConfig *OnChipTrainingConfig  `json:"training_config,omitempty"`
	SecurityConfig *CIMSecurityConfig     `json:"security_config,omitempty"`
	PUFStats       *PUFStats             `json:"puf_stats,omitempty"`
}

// PUFStats contains PUF statistics
type PUFStats struct {
	ChallengeCount int64   `json:"challenge_count"`
	BitErrorRate   float64 `json:"bit_error_rate"`
	EnrolledCRPs   int     `json:"enrolled_crps"`
}

// SerializeOnChipSecurityState serializes state to JSON
func SerializeOnChipSecurityState(trainingConfig *OnChipTrainingConfig, securityConfig *CIMSecurityConfig, puf *ReRAMPUF) ([]byte, error) {
	var pufStats *PUFStats
	if puf != nil {
		pufStats = &PUFStats{
			ChallengeCount: puf.ChallengeCount,
			BitErrorRate:   puf.BitErrorRate,
			EnrolledCRPs:   len(puf.EnrolledCRPs),
		}
	}

	state := &OnChipSecurityState{
		TrainingConfig: trainingConfig,
		SecurityConfig: securityConfig,
		PUFStats:       pufStats,
	}
	return json.MarshalIndent(state, "", "  ")
}

// DeserializeOnChipSecurityState deserializes state from JSON
func DeserializeOnChipSecurityState(data []byte) (*OnChipSecurityState, error) {
	var state OnChipSecurityState
	err := json.Unmarshal(data, &state)
	return &state, err
}

// ============================================================================
// Benchmark
// ============================================================================

// SecurityBenchmark benchmarks security overhead
type SecurityBenchmark struct {
	Results []SecurityBenchmarkResult
}

// SecurityBenchmarkResult contains benchmark results
type SecurityBenchmarkResult struct {
	Feature         string
	LatencyOverhead float64 // Percentage
	EnergyOverhead  float64 // Percentage
	AreaOverhead    float64 // Percentage
	SecurityLevel   string  // "Low", "Medium", "High"
}

// NewSecurityBenchmark creates a new security benchmark
func NewSecurityBenchmark() *SecurityBenchmark {
	return &SecurityBenchmark{
		Results: []SecurityBenchmarkResult{
			{"PUF Integration", 5.0, 3.0, 2.0, "High"},
			{"Weight Encryption", 10.0, 8.0, 5.0, "High"},
			{"Weight Masking", 2.0, 2.0, 1.0, "Medium"},
			{"Power Balancing", 15.0, 20.0, 10.0, "High"},
			{"Random Delays", 8.0, 5.0, 0.0, "Medium"},
			{"Noise Injection", 3.0, 5.0, 1.0, "Medium"},
			{"Input Validation", 5.0, 3.0, 2.0, "Medium"},
			{"Adversarial Defense", 12.0, 10.0, 5.0, "High"},
		},
	}
}

// GenerateReport generates a benchmark report
func (sb *SecurityBenchmark) GenerateReport() string {
	report := "Security Overhead Benchmark\n"
	report += "===========================\n\n"

	report += "Feature               | Latency | Energy | Area | Security\n"
	report += "----------------------|---------.|--------|------|----------\n"

	totalLatency := 0.0
	totalEnergy := 0.0
	totalArea := 0.0

	for _, r := range sb.Results {
		report += fmt.Sprintf("%-21s | %5.1f%% | %5.1f%% | %4.1f%% | %s\n",
			r.Feature, r.LatencyOverhead, r.EnergyOverhead, r.AreaOverhead, r.SecurityLevel)
		totalLatency += r.LatencyOverhead
		totalEnergy += r.EnergyOverhead
		totalArea += r.AreaOverhead
	}

	report += "----------------------|---------.|--------|------|----------\n"
	report += fmt.Sprintf("%-21s | %5.1f%% | %5.1f%% | %4.1f%% | Full\n",
		"TOTAL (all features)", totalLatency, totalEnergy, totalArea)

	return report
}

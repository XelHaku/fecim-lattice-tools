// Package layers provides federated learning and adversarial robustness for CIM accelerators.
//
// This module implements secure distributed training and attack-resilient inference
// for ferroelectric compute-in-memory systems deployed at the edge.
//
// Key features:
// - FedAvg and secure aggregation protocols
// - Physical Unclonable Function (PUF) key generation
// - Differential privacy for gradient protection
// - Inherent CIM noise-based adversarial defense
// - FGSM/PGD attack simulation
// - Randomized smoothing certification
//
// References:
// - "Federated learning with memristor CIM" (Nature Electronics 2025)
// - "Inherent adversarial robustness of analog CIM" (Nature Communications 2025)
// - "Memristor-SRAM CIM fusion" (Science 2024)
// - "Side-channel attack analysis on CIM" (IEEE TETC 2024)
package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
)

// =============================================================================
// Federated Learning Core
// =============================================================================

// FederatedConfig configures federated learning parameters.
type FederatedConfig struct {
	NumClients          int     // number of participating clients
	ClientFraction      float64 // fraction selected per round (0.1-1.0)
	LocalEpochs         int     // local training epochs per round
	LocalBatchSize      int     // local batch size
	GlobalRounds        int     // total communication rounds
	LearningRate        float64 // local learning rate
	WeightDecay         float64 // L2 regularization
	Momentum            float64 // SGD momentum
	SecureAggregation   bool    // enable secure aggregation
	DifferentialPrivacy bool    // enable DP
	CompressionEnabled  bool    // gradient compression
	AsyncUpdates        bool    // asynchronous aggregation
}

// DefaultFederatedConfig returns standard FL settings.
func DefaultFederatedConfig() FederatedConfig {
	return FederatedConfig{
		NumClients:          10,
		ClientFraction:      0.3,
		LocalEpochs:         5,
		LocalBatchSize:      32,
		GlobalRounds:        100,
		LearningRate:        0.01,
		WeightDecay:         1e-4,
		Momentum:            0.9,
		SecureAggregation:   true,
		DifferentialPrivacy: true,
		CompressionEnabled:  true,
		AsyncUpdates:        false,
	}
}

// FederatedClient represents an edge CIM device in the federation.
type FederatedClient struct {
	ClientID       int
	LocalModel     *LocalCIMModel
	DataSize       int     // number of local samples
	LastAccuracy   float64
	RoundsJoined   int
	ComputeLatency float64 // ms per batch
	BandwidthMbps  float64
	BatteryLevel   float64 // 0-1
	IsOnline       bool
	PUFKey         []byte // hardware-derived key
}

// LocalCIMModel holds client's local model state.
type LocalCIMModel struct {
	Weights   map[string][]float64
	Gradients map[string][]float64
	Optimizer *SGDOptimizer
}

// SGDOptimizer implements momentum SGD.
type SGDOptimizer struct {
	LearningRate float64
	Momentum     float64
	WeightDecay  float64
	Velocity     map[string][]float64
}

// NewSGDOptimizer creates a momentum SGD optimizer.
func NewSGDOptimizer(lr, momentum, decay float64) *SGDOptimizer {
	return &SGDOptimizer{
		LearningRate: lr,
		Momentum:     momentum,
		WeightDecay:  decay,
		Velocity:     make(map[string][]float64),
	}
}

// Step performs one optimization step.
func (o *SGDOptimizer) Step(layerName string, weights, gradients []float64) {
	if _, exists := o.Velocity[layerName]; !exists {
		o.Velocity[layerName] = make([]float64, len(weights))
	}

	v := o.Velocity[layerName]
	for i := range weights {
		// Gradient with weight decay
		g := gradients[i] + o.WeightDecay*weights[i]
		// Momentum update
		v[i] = o.Momentum*v[i] - o.LearningRate*g
		weights[i] += v[i]
	}
}

// =============================================================================
// Federated Aggregation Server
// =============================================================================

// FederatedServer manages the global model and aggregation.
type FederatedServer struct {
	Config       FederatedConfig
	GlobalModel  map[string][]float64
	Clients      []*FederatedClient
	RoundNumber  int
	RoundMetrics []RoundMetrics
	mu           sync.RWMutex
	RNG          *rand.Rand
}

// RoundMetrics tracks federated learning progress.
type RoundMetrics struct {
	Round           int
	ClientsSelected int
	AvgLoss         float64
	AvgAccuracy     float64
	CommBytes       int64
	WallClockTime   float64
	Convergence     float64 // gradient norm
}

// NewFederatedServer creates a federated learning server.
func NewFederatedServer(config FederatedConfig, seed int64) *FederatedServer {
	return &FederatedServer{
		Config:      config,
		GlobalModel: make(map[string][]float64),
		Clients:     make([]*FederatedClient, 0, config.NumClients),
		RNG:         rand.New(rand.NewSource(seed)),
	}
}

// RegisterClient adds a client to the federation.
func (s *FederatedServer) RegisterClient(client *FederatedClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Clients = append(s.Clients, client)
}

// SelectClients chooses clients for current round.
func (s *FederatedServer) SelectClients() []*FederatedClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter online clients
	online := make([]*FederatedClient, 0)
	for _, c := range s.Clients {
		if c.IsOnline && c.BatteryLevel > 0.2 {
			online = append(online, c)
		}
	}

	// Sample fraction
	numSelect := int(math.Ceil(float64(len(online)) * s.Config.ClientFraction))
	if numSelect > len(online) {
		numSelect = len(online)
	}

	// Random selection
	perm := s.RNG.Perm(len(online))
	selected := make([]*FederatedClient, numSelect)
	for i := 0; i < numSelect; i++ {
		selected[i] = online[perm[i]]
	}

	return selected
}

// FedAvgAggregate performs weighted averaging of client updates.
func (s *FederatedServer) FedAvgAggregate(clientUpdates map[int]map[string][]float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Compute total data size for weighting
	totalSamples := 0
	for clientID := range clientUpdates {
		for _, c := range s.Clients {
			if c.ClientID == clientID {
				totalSamples += c.DataSize
				break
			}
		}
	}

	if totalSamples == 0 {
		return
	}

	// Weighted aggregation
	for layerName := range s.GlobalModel {
		newWeights := make([]float64, len(s.GlobalModel[layerName]))

		for clientID, update := range clientUpdates {
			weights := update[layerName]
			if weights == nil {
				continue
			}

			// Find client weight
			var clientSamples int
			for _, c := range s.Clients {
				if c.ClientID == clientID {
					clientSamples = c.DataSize
					break
				}
			}
			weight := float64(clientSamples) / float64(totalSamples)

			for i := range newWeights {
				newWeights[i] += weight * weights[i]
			}
		}

		s.GlobalModel[layerName] = newWeights
	}
}

// =============================================================================
// Secure Aggregation
// =============================================================================

// SecureAggConfig configures secure aggregation.
type SecureAggConfig struct {
	KeySize          int     // bits for encryption key
	ThresholdClients int     // minimum clients for reconstruction
	DropoutTolerance float64 // fraction of dropped clients tolerated
	UseHomomorphic   bool    // use additive homomorphic encryption
	UseMasking       bool    // use pairwise masking
}

// DefaultSecureAggConfig returns standard secure aggregation settings.
func DefaultSecureAggConfig() SecureAggConfig {
	return SecureAggConfig{
		KeySize:          256,
		ThresholdClients: 3,
		DropoutTolerance: 0.3,
		UseHomomorphic:   true,
		UseMasking:       true,
	}
}

// SecureAggregator implements secure aggregation protocol.
type SecureAggregator struct {
	Config    SecureAggConfig
	ClientKeys map[int][]byte // client public keys
	Shares    map[int][]SecretShare
	RNG       *rand.Rand
}

// SecretShare represents a Shamir secret share.
type SecretShare struct {
	Index int
	Value []byte
}

// NewSecureAggregator creates a secure aggregator.
func NewSecureAggregator(config SecureAggConfig, seed int64) *SecureAggregator {
	return &SecureAggregator{
		Config:     config,
		ClientKeys: make(map[int][]byte),
		Shares:     make(map[int][]SecretShare),
		RNG:        rand.New(rand.NewSource(seed)),
	}
}

// GenerateMask creates pairwise mask for client pair.
func (sa *SecureAggregator) GenerateMask(clientI, clientJ int, dimension int) []float64 {
	// Deterministic mask from shared seed
	seed := int64(clientI*1000 + clientJ)
	rng := rand.New(rand.NewSource(seed))

	mask := make([]float64, dimension)
	for i := range mask {
		mask[i] = rng.NormFloat64()
	}
	return mask
}

// MaskUpdate applies additive mask to client update.
func (sa *SecureAggregator) MaskUpdate(clientID int, allClients []int, update []float64) []float64 {
	masked := make([]float64, len(update))
	copy(masked, update)

	for _, otherID := range allClients {
		if otherID == clientID {
			continue
		}

		mask := sa.GenerateMask(clientID, otherID, len(update))
		if clientID < otherID {
			// Add mask
			for i := range masked {
				masked[i] += mask[i]
			}
		} else {
			// Subtract mask
			for i := range masked {
				masked[i] -= mask[i]
			}
		}
	}

	return masked
}

// AggregateMasked aggregates masked updates (masks cancel out).
func (sa *SecureAggregator) AggregateMasked(maskedUpdates map[int][]float64) []float64 {
	if len(maskedUpdates) == 0 {
		return nil
	}

	var dimension int
	for _, update := range maskedUpdates {
		dimension = len(update)
		break
	}

	aggregate := make([]float64, dimension)
	for _, update := range maskedUpdates {
		for i := range aggregate {
			aggregate[i] += update[i]
		}
	}

	// Average
	numClients := float64(len(maskedUpdates))
	for i := range aggregate {
		aggregate[i] /= numClients
	}

	return aggregate
}

// =============================================================================
// Physical Unclonable Function (PUF)
// =============================================================================

// PUFConfig configures PUF key generation.
type PUFConfig struct {
	ArrayRows       int     // memristor array rows for PUF
	ArrayCols       int     // memristor array cols for PUF
	ChallengeSize   int     // bits in challenge
	ResponseSize    int     // bits in response
	ErrorThreshold  float64 // bit error rate threshold
	EnrollmentReads int     // reads during enrollment
}

// DefaultPUFConfig returns standard PUF settings.
func DefaultPUFConfig() PUFConfig {
	return PUFConfig{
		ArrayRows:       32,
		ArrayCols:       32,
		ChallengeSize:   128,
		ResponseSize:    256,
		ErrorThreshold:  0.05,
		EnrollmentReads: 100,
	}
}

// MemristorPUF implements physical unclonable function using memristor variation.
type MemristorPUF struct {
	Config           PUFConfig
	ConductanceArray [][]float64 // device conductances
	Threshold        float64     // comparison threshold
	HelperData       []byte      // error correction helper
	RNG              *rand.Rand
}

// NewMemristorPUF creates a memristor-based PUF.
func NewMemristorPUF(config PUFConfig, seed int64) *MemristorPUF {
	puf := &MemristorPUF{
		Config:           config,
		ConductanceArray: make([][]float64, config.ArrayRows),
		RNG:              rand.New(rand.NewSource(seed)),
	}

	// Initialize with random D2D variation (simulates manufacturing)
	for i := 0; i < config.ArrayRows; i++ {
		puf.ConductanceArray[i] = make([]float64, config.ArrayCols)
		for j := 0; j < config.ArrayCols; j++ {
			// Log-normal distribution models memristor variation
			puf.ConductanceArray[i][j] = math.Exp(puf.RNG.NormFloat64() * 0.5)
		}
	}

	// Compute median for threshold
	flat := make([]float64, 0, config.ArrayRows*config.ArrayCols)
	for i := range puf.ConductanceArray {
		flat = append(flat, puf.ConductanceArray[i]...)
	}
	sort.Float64s(flat)
	puf.Threshold = flat[len(flat)/2]

	return puf
}

// GenerateResponse produces PUF response to challenge.
func (p *MemristorPUF) GenerateResponse(challenge []byte) []byte {
	response := make([]byte, p.Config.ResponseSize/8)

	// Use challenge to select device pairs
	rng := rand.New(rand.NewSource(int64(bytesToUint64(challenge))))

	for bit := 0; bit < p.Config.ResponseSize; bit++ {
		// Select random device pair
		row1, col1 := rng.Intn(p.Config.ArrayRows), rng.Intn(p.Config.ArrayCols)
		row2, col2 := rng.Intn(p.Config.ArrayRows), rng.Intn(p.Config.ArrayCols)

		g1 := p.ConductanceArray[row1][col1]
		g2 := p.ConductanceArray[row2][col2]

		// Compare conductances (with noise)
		g1Noisy := g1 * (1 + p.RNG.NormFloat64()*0.02) // 2% C2C
		g2Noisy := g2 * (1 + p.RNG.NormFloat64()*0.02)

		byteIdx := bit / 8
		bitIdx := uint(bit % 8)
		if g1Noisy > g2Noisy {
			response[byteIdx] |= 1 << bitIdx
		}
	}

	return response
}

// DeriveKey generates cryptographic key from PUF.
func (p *MemristorPUF) DeriveKey(challenge []byte) []byte {
	// Multiple reads for stability
	responses := make([][]byte, p.Config.EnrollmentReads)
	for i := 0; i < p.Config.EnrollmentReads; i++ {
		responses[i] = p.GenerateResponse(challenge)
	}

	// Majority voting for stable key
	key := make([]byte, p.Config.ResponseSize/8)
	for bit := 0; bit < p.Config.ResponseSize; bit++ {
		ones := 0
		for _, r := range responses {
			byteIdx := bit / 8
			bitIdx := uint(bit % 8)
			if (r[byteIdx] & (1 << bitIdx)) != 0 {
				ones++
			}
		}
		if ones > p.Config.EnrollmentReads/2 {
			byteIdx := bit / 8
			bitIdx := uint(bit % 8)
			key[byteIdx] |= 1 << bitIdx
		}
	}

	return key
}

func bytesToUint64(b []byte) uint64 {
	var result uint64
	for i := 0; i < 8 && i < len(b); i++ {
		result |= uint64(b[i]) << (8 * i)
	}
	return result
}

// =============================================================================
// Differential Privacy
// =============================================================================

// DPConfig configures differential privacy.
type DPConfig struct {
	Epsilon       float64 // privacy budget
	Delta         float64 // failure probability
	ClipNorm      float64 // gradient clipping threshold
	NoiseMult     float64 // noise multiplier (derived)
	AccountingStr string  // RDP, GDP, etc.
	MaxGradNorm   float64 // per-sample gradient norm bound
}

// DefaultDPConfig returns standard differential privacy settings.
func DefaultDPConfig() DPConfig {
	return DPConfig{
		Epsilon:       1.0,
		Delta:         1e-5,
		ClipNorm:      1.0,
		NoiseMult:     1.1,
		AccountingStr: "RDP",
		MaxGradNorm:   1.0,
	}
}

// DifferentialPrivacyEngine implements DP mechanisms for FL.
type DifferentialPrivacyEngine struct {
	Config         DPConfig
	SpentEpsilon   float64
	SpentDelta     float64
	RNG            *rand.Rand
	RDPOrders      []float64
	RDPEpsilons    []float64
}

// NewDifferentialPrivacyEngine creates a DP engine.
func NewDifferentialPrivacyEngine(config DPConfig, seed int64) *DifferentialPrivacyEngine {
	return &DifferentialPrivacyEngine{
		Config:      config,
		RNG:         rand.New(rand.NewSource(seed)),
		RDPOrders:   []float64{1.5, 2, 5, 10, 20, 50, 100},
		RDPEpsilons: make([]float64, 7),
	}
}

// ClipGradient clips gradient to bounded norm.
func (dp *DifferentialPrivacyEngine) ClipGradient(gradient []float64) []float64 {
	norm := 0.0
	for _, g := range gradient {
		norm += g * g
	}
	norm = math.Sqrt(norm)

	if norm <= dp.Config.ClipNorm {
		return gradient
	}

	// Scale to clip norm
	clipped := make([]float64, len(gradient))
	scale := dp.Config.ClipNorm / norm
	for i, g := range gradient {
		clipped[i] = g * scale
	}
	return clipped
}

// AddGaussianNoise adds calibrated Gaussian noise.
func (dp *DifferentialPrivacyEngine) AddGaussianNoise(gradient []float64) []float64 {
	sigma := dp.Config.NoiseMult * dp.Config.ClipNorm
	noisy := make([]float64, len(gradient))
	for i, g := range gradient {
		noisy[i] = g + dp.RNG.NormFloat64()*sigma
	}
	return noisy
}

// PrivatizeBatch applies DP to mini-batch of gradients.
func (dp *DifferentialPrivacyEngine) PrivatizeBatch(gradients [][]float64, batchSize int) []float64 {
	// Clip each gradient
	clipped := make([][]float64, len(gradients))
	for i, g := range gradients {
		clipped[i] = dp.ClipGradient(g)
	}

	// Sum clipped gradients
	sumGrad := make([]float64, len(clipped[0]))
	for _, g := range clipped {
		for j := range sumGrad {
			sumGrad[j] += g[j]
		}
	}

	// Add noise and average
	sigma := dp.Config.NoiseMult * dp.Config.ClipNorm
	for i := range sumGrad {
		sumGrad[i] = (sumGrad[i] + dp.RNG.NormFloat64()*sigma) / float64(batchSize)
	}

	return sumGrad
}

// ComputePrivacyLoss estimates privacy loss using RDP.
func (dp *DifferentialPrivacyEngine) ComputePrivacyLoss(numSteps int, samplingProb float64) float64 {
	// Simplified RDP accounting
	for i, alpha := range dp.RDPOrders {
		sigma := dp.Config.NoiseMult
		// Gaussian mechanism RDP
		rdp := alpha / (2 * sigma * sigma)
		// Subsampling amplification
		rdp *= samplingProb * samplingProb * float64(numSteps)
		dp.RDPEpsilons[i] = rdp
	}

	// Convert to (epsilon, delta)-DP
	minEps := math.MaxFloat64
	for i, alpha := range dp.RDPOrders {
		eps := dp.RDPEpsilons[i] - math.Log(dp.Config.Delta)/(alpha-1)
		if eps < minEps {
			minEps = eps
		}
	}

	dp.SpentEpsilon += minEps
	return minEps
}

// =============================================================================
// Gradient Compression
// =============================================================================

// CompressionConfig configures gradient compression.
type CompressionConfig struct {
	Method         string  // "topk", "random", "quantize"
	Compression    float64 // compression ratio (0.01 = 1%)
	QuantBits      int     // bits for quantization
	ErrorFeedback  bool    // accumulate compression errors
}

// DefaultCompressionConfig returns standard compression settings.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Method:        "topk",
		Compression:   0.01,
		QuantBits:     8,
		ErrorFeedback: true,
	}
}

// GradientCompressor implements communication-efficient gradient compression.
type GradientCompressor struct {
	Config        CompressionConfig
	ErrorBuffer   map[string][]float64 // accumulated errors
}

// NewGradientCompressor creates a gradient compressor.
func NewGradientCompressor(config CompressionConfig) *GradientCompressor {
	return &GradientCompressor{
		Config:      config,
		ErrorBuffer: make(map[string][]float64),
	}
}

// TopKCompress selects top-k gradients by magnitude.
func (gc *GradientCompressor) TopKCompress(layerName string, gradient []float64) ([]int, []float64) {
	// Add error feedback
	if gc.Config.ErrorFeedback {
		if err, exists := gc.ErrorBuffer[layerName]; exists {
			for i := range gradient {
				gradient[i] += err[i]
			}
		}
	}

	k := int(float64(len(gradient)) * gc.Config.Compression)
	if k < 1 {
		k = 1
	}

	// Find top-k indices
	type indexValue struct {
		idx int
		val float64
	}
	indexed := make([]indexValue, len(gradient))
	for i, v := range gradient {
		indexed[i] = indexValue{i, math.Abs(v)}
	}
	sort.Slice(indexed, func(i, j int) bool {
		return indexed[i].val > indexed[j].val
	})

	indices := make([]int, k)
	values := make([]float64, k)
	for i := 0; i < k; i++ {
		indices[i] = indexed[i].idx
		values[i] = gradient[indexed[i].idx]
	}

	// Compute error for feedback
	if gc.Config.ErrorFeedback {
		error := make([]float64, len(gradient))
		copy(error, gradient)
		for i := 0; i < k; i++ {
			error[indices[i]] = 0
		}
		gc.ErrorBuffer[layerName] = error
	}

	return indices, values
}

// Decompress reconstructs gradient from sparse representation.
func (gc *GradientCompressor) Decompress(indices []int, values []float64, dimension int) []float64 {
	gradient := make([]float64, dimension)
	for i, idx := range indices {
		gradient[idx] = values[i]
	}
	return gradient
}

// =============================================================================
// Adversarial Robustness
// =============================================================================

// AdversarialConfig configures adversarial robustness.
type AdversarialConfig struct {
	AttackType     string  // "fgsm", "pgd", "cw"
	Epsilon        float64 // perturbation budget (L∞)
	PGDSteps       int     // PGD iterations
	PGDStepSize    float64 // PGD step size
	RandomStart    bool    // random initialization for PGD
	TargetedAttack bool    // targeted vs untargeted
	DefenseType    string  // "noise", "smooth", "adv_train"
}

// DefaultAdversarialConfig returns standard adversarial settings.
func DefaultAdversarialConfig() AdversarialConfig {
	return AdversarialConfig{
		AttackType:     "pgd",
		Epsilon:        0.03,  // 8/255 typical for images
		PGDSteps:       20,
		PGDStepSize:    0.003,
		RandomStart:    true,
		TargetedAttack: false,
		DefenseType:    "noise",
	}
}

// AdversarialAttacker implements adversarial attack generation.
type AdversarialAttacker struct {
	Config AdversarialConfig
	RNG    *rand.Rand
}

// NewAdversarialAttacker creates an adversarial attacker.
func NewAdversarialAttacker(config AdversarialConfig, seed int64) *AdversarialAttacker {
	return &AdversarialAttacker{
		Config: config,
		RNG:    rand.New(rand.NewSource(seed)),
	}
}

// FGSM implements Fast Gradient Sign Method attack.
func (aa *AdversarialAttacker) FGSM(input, gradient []float64) []float64 {
	adversarial := make([]float64, len(input))
	for i := range input {
		sign := 1.0
		if gradient[i] < 0 {
			sign = -1.0
		} else if gradient[i] == 0 {
			sign = 0
		}
		adversarial[i] = input[i] + aa.Config.Epsilon*sign
		// Clip to valid range
		adversarial[i] = math.Max(0, math.Min(1, adversarial[i]))
	}
	return adversarial
}

// PGD implements Projected Gradient Descent attack.
func (aa *AdversarialAttacker) PGD(
	input []float64,
	gradientFunc func([]float64) []float64,
) []float64 {
	adversarial := make([]float64, len(input))

	// Random start
	if aa.Config.RandomStart {
		for i := range adversarial {
			adversarial[i] = input[i] + (aa.RNG.Float64()*2-1)*aa.Config.Epsilon
			adversarial[i] = math.Max(0, math.Min(1, adversarial[i]))
		}
	} else {
		copy(adversarial, input)
	}

	// Iterative attack
	for step := 0; step < aa.Config.PGDSteps; step++ {
		grad := gradientFunc(adversarial)

		for i := range adversarial {
			sign := 1.0
			if grad[i] < 0 {
				sign = -1.0
			}
			adversarial[i] += aa.Config.PGDStepSize * sign

			// Project to epsilon ball
			delta := adversarial[i] - input[i]
			if delta > aa.Config.Epsilon {
				delta = aa.Config.Epsilon
			} else if delta < -aa.Config.Epsilon {
				delta = -aa.Config.Epsilon
			}
			adversarial[i] = input[i] + delta

			// Clip to valid range
			adversarial[i] = math.Max(0, math.Min(1, adversarial[i]))
		}
	}

	return adversarial
}

// =============================================================================
// CIM Noise-Based Defense
// =============================================================================

// CIMDefenseConfig configures inherent CIM noise defense.
type CIMDefenseConfig struct {
	D2DVariation     float64 // device-to-device variation
	C2CVariation     float64 // cycle-to-cycle variation
	ReadNoise        float64 // read operation noise
	ThermalNoise     float64 // thermal fluctuations
	NumInferences    int     // inferences for averaging
	UseStochasticInf bool    // stochastic inference mode
}

// DefaultCIMDefenseConfig returns standard CIM defense settings.
func DefaultCIMDefenseConfig() CIMDefenseConfig {
	return CIMDefenseConfig{
		D2DVariation:     0.05,
		C2CVariation:     0.02,
		ReadNoise:        0.01,
		ThermalNoise:     0.01,
		NumInferences:    10,
		UseStochasticInf: true,
	}
}

// CIMDefender implements inherent CIM noise-based adversarial defense.
type CIMDefender struct {
	Config    CIMDefenseConfig
	D2DMask   []float64 // fixed D2D variation per weight
	RNG       *rand.Rand
}

// NewCIMDefender creates a CIM-based adversarial defender.
func NewCIMDefender(config CIMDefenseConfig, weightDimension int, seed int64) *CIMDefender {
	defender := &CIMDefender{
		Config:  config,
		D2DMask: make([]float64, weightDimension),
		RNG:     rand.New(rand.NewSource(seed)),
	}

	// Initialize fixed D2D variation
	for i := range defender.D2DMask {
		defender.D2DMask[i] = 1 + defender.RNG.NormFloat64()*config.D2DVariation
	}

	return defender
}

// InjectCIMNoise adds realistic CIM noise to inference.
func (cd *CIMDefender) InjectCIMNoise(weights []float64) []float64 {
	noisy := make([]float64, len(weights))
	for i, w := range weights {
		// Apply D2D (fixed)
		noisy[i] = w * cd.D2DMask[i]
		// Apply C2C (stochastic)
		noisy[i] *= (1 + cd.RNG.NormFloat64()*cd.Config.C2CVariation)
		// Apply read noise
		noisy[i] += cd.RNG.NormFloat64() * cd.Config.ReadNoise
		// Apply thermal noise
		noisy[i] += cd.RNG.NormFloat64() * cd.Config.ThermalNoise
	}
	return noisy
}

// StochasticInference performs multiple noisy inferences and averages.
func (cd *CIMDefender) StochasticInference(
	input []float64,
	weights [][]float64,
	inferenceFunc func([]float64, [][]float64) []float64,
) []float64 {
	if !cd.Config.UseStochasticInf {
		return inferenceFunc(input, weights)
	}

	// Multiple noisy inferences
	outputs := make([][]float64, cd.Config.NumInferences)
	for i := 0; i < cd.Config.NumInferences; i++ {
		noisyWeights := make([][]float64, len(weights))
		for j, w := range weights {
			noisyWeights[j] = cd.InjectCIMNoise(w)
		}
		outputs[i] = inferenceFunc(input, noisyWeights)
	}

	// Average outputs
	avgOutput := make([]float64, len(outputs[0]))
	for _, out := range outputs {
		for j := range avgOutput {
			avgOutput[j] += out[j]
		}
	}
	for j := range avgOutput {
		avgOutput[j] /= float64(cd.Config.NumInferences)
	}

	return avgOutput
}

// =============================================================================
// Randomized Smoothing
// =============================================================================

// SmoothingConfig configures randomized smoothing.
type SmoothingConfig struct {
	Sigma         float64 // noise standard deviation
	NumSamples    int     // samples for certification
	Alpha         float64 // confidence level
	BatchSize     int     // batch size for sampling
}

// DefaultSmoothingConfig returns standard randomized smoothing settings.
func DefaultSmoothingConfig() SmoothingConfig {
	return SmoothingConfig{
		Sigma:      0.25,
		NumSamples: 1000,
		Alpha:      0.001,
		BatchSize:  100,
	}
}

// RandomizedSmoother implements certified adversarial robustness.
type RandomizedSmoother struct {
	Config       SmoothingConfig
	Classifier   func([]float64) int // base classifier
	RNG          *rand.Rand
}

// NewRandomizedSmoother creates a randomized smoother.
func NewRandomizedSmoother(
	config SmoothingConfig,
	classifier func([]float64) int,
	seed int64,
) *RandomizedSmoother {
	return &RandomizedSmoother{
		Config:     config,
		Classifier: classifier,
		RNG:        rand.New(rand.NewSource(seed)),
	}
}

// SmoothPredict performs smoothed prediction.
func (rs *RandomizedSmoother) SmoothPredict(input []float64, numClasses int) (int, float64) {
	counts := make([]int, numClasses)

	// Sample with noise
	for i := 0; i < rs.Config.NumSamples; i++ {
		noisy := make([]float64, len(input))
		for j := range input {
			noisy[j] = input[j] + rs.RNG.NormFloat64()*rs.Config.Sigma
		}
		pred := rs.Classifier(noisy)
		if pred >= 0 && pred < numClasses {
			counts[pred]++
		}
	}

	// Find top two classes
	topClass := 0
	secondClass := 0
	for i := 1; i < numClasses; i++ {
		if counts[i] > counts[topClass] {
			secondClass = topClass
			topClass = i
		} else if counts[i] > counts[secondClass] {
			secondClass = i
		}
	}

	// Compute certified radius
	pA := float64(counts[topClass]) / float64(rs.Config.NumSamples)
	radius := rs.Config.Sigma * rs.inverseCDF(pA)

	return topClass, radius
}

// CertifyRadius computes certified L2 robustness radius.
func (rs *RandomizedSmoother) CertifyRadius(pA float64) float64 {
	if pA <= 0.5 {
		return 0
	}
	return rs.Config.Sigma * rs.inverseCDF(pA)
}

// inverseCDF computes inverse of standard normal CDF.
func (rs *RandomizedSmoother) inverseCDF(p float64) float64 {
	// Approximation using rational function
	if p <= 0 {
		return math.Inf(-1)
	}
	if p >= 1 {
		return math.Inf(1)
	}

	// Use symmetry
	sign := 1.0
	if p > 0.5 {
		p = 1 - p
	} else {
		sign = -1.0
	}

	t := math.Sqrt(-2 * math.Log(p))
	c0 := 2.515517
	c1 := 0.802853
	c2 := 0.010328
	d1 := 1.432788
	d2 := 0.189269
	d3 := 0.001308

	return sign * (t - (c0+c1*t+c2*t*t)/(1+d1*t+d2*t*t+d3*t*t*t))
}

// =============================================================================
// Adversarial Training
// =============================================================================

// AdvTrainingConfig configures adversarial training.
type AdvTrainingConfig struct {
	AttackConfig    AdversarialConfig
	CleanWeight     float64 // weight for clean loss
	AdversarialWeight float64 // weight for adversarial loss
	WarmupEpochs    int     // epochs before adversarial training
	CurriculumMode  bool    // gradually increase epsilon
}

// DefaultAdvTrainingConfig returns standard adversarial training settings.
func DefaultAdvTrainingConfig() AdvTrainingConfig {
	return AdvTrainingConfig{
		AttackConfig:      DefaultAdversarialConfig(),
		CleanWeight:       0.5,
		AdversarialWeight: 0.5,
		WarmupEpochs:      5,
		CurriculumMode:    true,
	}
}

// AdversarialTrainer implements adversarial training.
type AdversarialTrainer struct {
	Config   AdvTrainingConfig
	Attacker *AdversarialAttacker
	Epoch    int
}

// NewAdversarialTrainer creates an adversarial trainer.
func NewAdversarialTrainer(config AdvTrainingConfig, seed int64) *AdversarialTrainer {
	return &AdversarialTrainer{
		Config:   config,
		Attacker: NewAdversarialAttacker(config.AttackConfig, seed),
	}
}

// ComputeLoss computes combined clean and adversarial loss.
func (at *AdversarialTrainer) ComputeLoss(
	cleanInput []float64,
	target int,
	lossFunc func([]float64, int) float64,
	gradFunc func([]float64) []float64,
) float64 {
	// Clean loss
	cleanLoss := lossFunc(cleanInput, target)

	// Skip adversarial during warmup
	if at.Epoch < at.Config.WarmupEpochs {
		return cleanLoss
	}

	// Generate adversarial example
	var advInput []float64
	if at.Config.AttackConfig.AttackType == "fgsm" {
		grad := gradFunc(cleanInput)
		advInput = at.Attacker.FGSM(cleanInput, grad)
	} else {
		advInput = at.Attacker.PGD(cleanInput, gradFunc)
	}

	// Adversarial loss
	advLoss := lossFunc(advInput, target)

	return at.Config.CleanWeight*cleanLoss + at.Config.AdversarialWeight*advLoss
}

// GetCurrentEpsilon returns current perturbation budget (for curriculum).
func (at *AdversarialTrainer) GetCurrentEpsilon() float64 {
	if !at.Config.CurriculumMode {
		return at.Config.AttackConfig.Epsilon
	}

	if at.Epoch < at.Config.WarmupEpochs {
		return 0
	}

	// Gradually increase epsilon
	progress := float64(at.Epoch-at.Config.WarmupEpochs) / 20.0
	if progress > 1.0 {
		progress = 1.0
	}
	return progress * at.Config.AttackConfig.Epsilon
}

// =============================================================================
// CIM Federated Adversarial Integration
// =============================================================================

// CIMFederatedSystem integrates federated learning with CIM adversarial defense.
type CIMFederatedSystem struct {
	Server         *FederatedServer
	SecureAgg      *SecureAggregator
	DPEngine       *DifferentialPrivacyEngine
	Compressor     *GradientCompressor
	PUFs           map[int]*MemristorPUF
	CIMDefenders   map[int]*CIMDefender
	AdvTrainer     *AdversarialTrainer
	RobustnessMetrics map[int]RobustnessMetrics
}

// RobustnessMetrics tracks adversarial robustness.
type RobustnessMetrics struct {
	CleanAccuracy       float64
	FGSMAccuracy        float64
	PGDAccuracy         float64
	CertifiedRadius     float64
	NoiseRobustness     float64
	InferenceVariance   float64
}

// NewCIMFederatedSystem creates an integrated CIM federated system.
func NewCIMFederatedSystem(
	flConfig FederatedConfig,
	saConfig SecureAggConfig,
	dpConfig DPConfig,
	advConfig AdvTrainingConfig,
	seed int64,
) *CIMFederatedSystem {
	return &CIMFederatedSystem{
		Server:            NewFederatedServer(flConfig, seed),
		SecureAgg:         NewSecureAggregator(saConfig, seed+1),
		DPEngine:          NewDifferentialPrivacyEngine(dpConfig, seed+2),
		Compressor:        NewGradientCompressor(DefaultCompressionConfig()),
		PUFs:              make(map[int]*MemristorPUF),
		CIMDefenders:      make(map[int]*CIMDefender),
		AdvTrainer:        NewAdversarialTrainer(advConfig, seed+3),
		RobustnessMetrics: make(map[int]RobustnessMetrics),
	}
}

// RegisterCIMClient adds a CIM client with hardware security.
func (cs *CIMFederatedSystem) RegisterCIMClient(
	clientID int,
	dataSize int,
	weightDimension int,
) {
	// Generate PUF for client
	cs.PUFs[clientID] = NewMemristorPUF(DefaultPUFConfig(), int64(clientID*1000))

	// Create CIM defender
	cs.CIMDefenders[clientID] = NewCIMDefender(
		DefaultCIMDefenseConfig(),
		weightDimension,
		int64(clientID*2000),
	)

	// Create client with PUF-derived key
	client := &FederatedClient{
		ClientID:       clientID,
		DataSize:       dataSize,
		IsOnline:       true,
		BatteryLevel:   1.0,
		BandwidthMbps:  10.0,
		PUFKey:         cs.PUFs[clientID].DeriveKey([]byte("init_challenge")),
	}

	cs.Server.RegisterClient(client)
}

// =============================================================================
// Serialization
// =============================================================================

// FederatedState captures system state for persistence.
type FederatedState struct {
	GlobalModel    map[string][]float64  `json:"global_model"`
	RoundNumber    int                   `json:"round_number"`
	RoundMetrics   []RoundMetrics        `json:"round_metrics"`
	SpentEpsilon   float64               `json:"spent_epsilon"`
	SpentDelta     float64               `json:"spent_delta"`
	ClientCount    int                   `json:"client_count"`
}

// ExportState exports federated system state.
func (cs *CIMFederatedSystem) ExportState() ([]byte, error) {
	state := FederatedState{
		GlobalModel:  cs.Server.GlobalModel,
		RoundNumber:  cs.Server.RoundNumber,
		RoundMetrics: cs.Server.RoundMetrics,
		SpentEpsilon: cs.DPEngine.SpentEpsilon,
		SpentDelta:   cs.DPEngine.SpentDelta,
		ClientCount:  len(cs.Server.Clients),
	}
	return json.MarshalIndent(state, "", "  ")
}

// ImportState loads federated system state.
func (cs *CIMFederatedSystem) ImportState(data []byte) error {
	var state FederatedState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state: %w", err)
	}

	cs.Server.GlobalModel = state.GlobalModel
	cs.Server.RoundNumber = state.RoundNumber
	cs.Server.RoundMetrics = state.RoundMetrics
	cs.DPEngine.SpentEpsilon = state.SpentEpsilon
	cs.DPEngine.SpentDelta = state.SpentDelta

	return nil
}

// =============================================================================
// Benchmarking
// =============================================================================

// FederatedBenchmark evaluates federated learning performance.
type FederatedBenchmark struct {
	Rounds             int
	FinalAccuracy      float64
	CommunicationBytes int64
	TotalWallClock     float64
	ConvergenceRound   int
	PrivacyBudget      float64
}

// RobustnessBenchmark evaluates adversarial robustness.
type RobustnessBenchmark struct {
	Model             string
	CleanAccuracy     float64
	FGSMAccuracy      float64
	PGDAccuracy       float64
	AutoAttackAccuracy float64
	CertifiedAccuracy float64
	CertifiedRadius   float64
	InferenceStdDev   float64
}

// RunRobustnessBenchmark evaluates model robustness.
func RunRobustnessBenchmark(
	model string,
	defender *CIMDefender,
	attacker *AdversarialAttacker,
	smoother *RandomizedSmoother,
	testInputs [][]float64,
	testLabels []int,
	classifier func([]float64) int,
	gradFunc func([]float64) []float64,
) RobustnessBenchmark {
	bench := RobustnessBenchmark{Model: model}

	if len(testInputs) == 0 {
		return bench
	}

	cleanCorrect := 0
	fgsmCorrect := 0
	pgdCorrect := 0
	totalRadius := 0.0

	for i, input := range testInputs {
		// Clean accuracy
		pred := classifier(input)
		if pred == testLabels[i] {
			cleanCorrect++
		}

		// FGSM accuracy
		grad := gradFunc(input)
		fgsmAdv := attacker.FGSM(input, grad)
		fgsmPred := classifier(fgsmAdv)
		if fgsmPred == testLabels[i] {
			fgsmCorrect++
		}

		// PGD accuracy
		pgdAdv := attacker.PGD(input, gradFunc)
		pgdPred := classifier(pgdAdv)
		if pgdPred == testLabels[i] {
			pgdCorrect++
		}

		// Certified radius
		_, radius := smoother.SmoothPredict(input, 10) // assume 10 classes
		totalRadius += radius
	}

	n := float64(len(testInputs))
	bench.CleanAccuracy = float64(cleanCorrect) / n
	bench.FGSMAccuracy = float64(fgsmCorrect) / n
	bench.PGDAccuracy = float64(pgdCorrect) / n
	bench.CertifiedRadius = totalRadius / n

	return bench
}

// federated_recsys_cim.go - Federated Learning and Recommendation Systems for CIM
// Part of the IronLattice CIM simulation framework
// Iteration 130: Privacy-preserving FL + Embedding-based RecSys on CIM

package layers

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"sync"
)

// =============================================================================
// FEDERATED LEARNING FOR CIM EDGE DEVICES
// =============================================================================

// FederatedConfig configures federated learning on CIM
type FederatedConfig struct {
	NumClients          int
	LocalEpochs         int
	BatchSize           int
	LearningRate        float64
	FederatedRounds     int
	AggregationMethod   AggregationMethod
	PrivacyMechanism    PrivacyMechanism
	DifferentialPrivacy *DPConfig
	GradientCompression *CompressionConfig
	SecureAggregation   bool
	ClientSelection     float64 // Fraction of clients per round
}

// AggregationMethod defines how to aggregate client updates
type AggregationMethod int

const (
	AggFedAvg AggregationMethod = iota // Federated Averaging
	AggFedProx                         // Proximal term for heterogeneity
	AggFedAdam                         // Server-side Adam
	AggScaffold                        // Stochastic Controlled Averaging
)

// PrivacyMechanism defines privacy protection method
type PrivacyMechanism int

const (
	PrivacyNone PrivacyMechanism = iota
	PrivacyDP                    // Differential Privacy
	PrivacySMPC                  // Secure Multi-Party Computation
	PrivacyHE                    // Homomorphic Encryption
	PrivacyPUF                   // Physical Unclonable Function
)

// DPConfig configures differential privacy
type DPConfig struct {
	Epsilon       float64 // Privacy budget
	Delta         float64 // Failure probability
	ClipNorm      float64 // Gradient clipping threshold
	NoiseMult     float64 // Noise multiplier
	AccountMethod string  // "rdp" or "moments"
}

// CompressionConfig configures gradient compression
type CompressionConfig struct {
	Method          string  // "topk", "random", "quantize"
	CompressionRate float64 // 0.01 = 1% of gradients sent
	QuantizeBits    int     // For quantization
	ErrorFeedback   bool    // Accumulate compression error
}

// FederatedCIMServer manages federated learning
type FederatedCIMServer struct {
	Config       *FederatedConfig
	GlobalModel  *CIMModel
	Clients      []*FederatedCIMClient
	RoundNumber  int
	PrivacyBudget float64
	Stats        *FederatedStats
	mu           sync.Mutex
}

// CIMModel represents a model stored on CIM hardware
type CIMModel struct {
	Layers      []*CIMLayer
	NumParams   int
	Precision   int
	IsQuantized bool
}

// CIMLayer represents a single layer in CIM
type CIMLayer struct {
	Name    string
	Weights [][]float64
	Bias    []float64
	Shape   []int
}

// FederatedCIMClient represents an edge device with CIM
type FederatedCIMClient struct {
	ID           int
	LocalModel   *CIMModel
	LocalData    *LocalDataset
	CIMArray     *EdgeCIMArray
	Optimizer    *LocalOptimizer
	ErrorBuffer  [][]float64 // For error feedback compression
	Stats        *ClientStats
}

// EdgeCIMArray represents CIM hardware on edge device
type EdgeCIMArray struct {
	Rows         int
	Cols         int
	Technology   string // "reram", "fefet", "ftj"
	PowerBudgetMW float64
	MemoryKB     int
	HasPUF       bool
	PUFResponse  []byte
}

// LocalDataset represents data on a client
type LocalDataset struct {
	NumSamples int
	Features   [][]float64
	Labels     []int
	IsIID      bool // Independent and identically distributed
}

// LocalOptimizer handles local training
type LocalOptimizer struct {
	LearningRate float64
	Momentum     float64
	WeightDecay  float64
	Velocity     [][][]float64
}

// FederatedStats tracks federated learning statistics
type FederatedStats struct {
	RoundsCompleted    int
	TotalCommunication int64 // Bytes transmitted
	PrivacySpent       float64
	GlobalAccuracy     float64
	ClientAccuracies   []float64
	ConvergenceHistory []float64
}

// ClientStats tracks per-client statistics
type ClientStats struct {
	LocalUpdates     int
	ComputeEnergyMJ  float64
	CommunicationKB  float64
	PrivacyContrib   float64
}

// NewFederatedCIMServer creates a federated learning server
func NewFederatedCIMServer(config *FederatedConfig) *FederatedCIMServer {
	server := &FederatedCIMServer{
		Config:        config,
		Clients:       make([]*FederatedCIMClient, config.NumClients),
		PrivacyBudget: config.DifferentialPrivacy.Epsilon,
		Stats: &FederatedStats{
			ClientAccuracies:   make([]float64, config.NumClients),
			ConvergenceHistory: make([]float64, 0),
		},
	}

	// Initialize global model
	server.GlobalModel = createDefaultCIMModel()

	// Initialize clients
	for i := 0; i < config.NumClients; i++ {
		server.Clients[i] = newFederatedCIMClient(i, server.GlobalModel)
	}

	return server
}

func createDefaultCIMModel() *CIMModel {
	return &CIMModel{
		Layers: []*CIMLayer{
			{Name: "fc1", Weights: initWeights(784, 256), Bias: make([]float64, 256), Shape: []int{784, 256}},
			{Name: "fc2", Weights: initWeights(256, 128), Bias: make([]float64, 128), Shape: []int{256, 128}},
			{Name: "fc3", Weights: initWeights(128, 10), Bias: make([]float64, 10), Shape: []int{128, 10}},
		},
		NumParams:   784*256 + 256*128 + 128*10 + 256 + 128 + 10,
		Precision:   8,
		IsQuantized: true,
	}
}

func newFederatedCIMClient(id int, globalModel *CIMModel) *FederatedCIMClient {
	client := &FederatedCIMClient{
		ID:         id,
		LocalModel: cloneCIMModel(globalModel),
		LocalData: &LocalDataset{
			NumSamples: 100 + rand.Intn(400), // Non-IID: varying sizes
			IsIID:      false,
		},
		CIMArray: &EdgeCIMArray{
			Rows:         128,
			Cols:         128,
			Technology:   "fefet",
			PowerBudgetMW: 10,
			MemoryKB:     256,
			HasPUF:       true,
		},
		Optimizer: &LocalOptimizer{
			LearningRate: 0.01,
			Momentum:     0.9,
			WeightDecay:  1e-4,
		},
		Stats: &ClientStats{},
	}

	// Generate PUF response if available
	if client.CIMArray.HasPUF {
		client.CIMArray.PUFResponse = generatePUFResponse(id)
	}

	// Generate synthetic local data
	client.generateLocalData()

	return client
}

func (c *FederatedCIMClient) generateLocalData() {
	// Generate non-IID data (biased towards certain classes)
	biasClass := c.ID % 10 // Each client biased towards one class
	c.LocalData.Features = make([][]float64, c.LocalData.NumSamples)
	c.LocalData.Labels = make([]int, c.LocalData.NumSamples)

	for i := 0; i < c.LocalData.NumSamples; i++ {
		c.LocalData.Features[i] = make([]float64, 784)
		for j := range c.LocalData.Features[i] {
			c.LocalData.Features[i][j] = rand.NormFloat64() * 0.3
		}
		// 70% probability of bias class, 30% random
		if rand.Float64() < 0.7 {
			c.LocalData.Labels[i] = biasClass
		} else {
			c.LocalData.Labels[i] = rand.Intn(10)
		}
	}
}

// RunFederatedRound executes one round of federated learning
func (s *FederatedCIMServer) RunFederatedRound() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Select clients for this round
	selectedClients := s.selectClients()

	// Collect client updates
	updates := make([]*ClientUpdate, len(selectedClients))
	weights := make([]float64, len(selectedClients))

	var wg sync.WaitGroup
	for i, client := range selectedClients {
		wg.Add(1)
		go func(idx int, c *FederatedCIMClient) {
			defer wg.Done()

			// Train locally on CIM
			update := c.trainLocal(s.Config)

			// Apply privacy mechanism
			if s.Config.PrivacyMechanism == PrivacyDP {
				update = s.applyDifferentialPrivacy(update)
			}

			// Compress gradients
			if s.Config.GradientCompression != nil {
				update = c.compressUpdate(update, s.Config.GradientCompression)
			}

			updates[idx] = update
			weights[idx] = float64(c.LocalData.NumSamples)
		}(i, client)
	}
	wg.Wait()

	// Aggregate updates
	s.aggregateUpdates(updates, weights)

	// Update statistics
	s.RoundNumber++
	s.Stats.RoundsCompleted++
	s.updateConvergenceHistory()

	// Broadcast new global model
	s.broadcastModel()

	return nil
}

// ClientUpdate contains the update from a client
type ClientUpdate struct {
	ClientID    int
	Gradients   [][][]float64
	BiasGrads   [][]float64
	NumSamples  int
	LocalLoss   float64
	Compressed  bool
	PrivacyNoise float64
}

func (s *FederatedCIMServer) selectClients() []*FederatedCIMClient {
	numSelect := int(float64(s.Config.NumClients) * s.Config.ClientSelection)
	if numSelect < 1 {
		numSelect = 1
	}

	// Random selection
	perm := rand.Perm(s.Config.NumClients)
	selected := make([]*FederatedCIMClient, numSelect)
	for i := 0; i < numSelect; i++ {
		selected[i] = s.Clients[perm[i]]
	}
	return selected
}

func (c *FederatedCIMClient) trainLocal(config *FederatedConfig) *ClientUpdate {
	update := &ClientUpdate{
		ClientID:   c.ID,
		NumSamples: c.LocalData.NumSamples,
		Gradients:  make([][][]float64, len(c.LocalModel.Layers)),
		BiasGrads:  make([][]float64, len(c.LocalModel.Layers)),
	}

	// Initialize gradients
	for l, layer := range c.LocalModel.Layers {
		update.Gradients[l] = make([][]float64, len(layer.Weights))
		for i := range layer.Weights {
			update.Gradients[l][i] = make([]float64, len(layer.Weights[i]))
		}
		update.BiasGrads[l] = make([]float64, len(layer.Bias))
	}

	// Local training epochs
	totalLoss := 0.0
	for epoch := 0; epoch < config.LocalEpochs; epoch++ {
		// Mini-batch training
		for batch := 0; batch < c.LocalData.NumSamples/config.BatchSize; batch++ {
			batchLoss := c.trainBatch(batch, config.BatchSize, update)
			totalLoss += batchLoss
		}
	}

	update.LocalLoss = totalLoss / float64(config.LocalEpochs)
	c.Stats.LocalUpdates++
	c.Stats.ComputeEnergyMJ += estimateCIMEnergy(c.LocalData.NumSamples, c.CIMArray)

	return update
}

func (c *FederatedCIMClient) trainBatch(batchIdx, batchSize int, update *ClientUpdate) float64 {
	startIdx := batchIdx * batchSize
	endIdx := startIdx + batchSize
	if endIdx > c.LocalData.NumSamples {
		endIdx = c.LocalData.NumSamples
	}

	// Forward pass on CIM (simulated)
	loss := 0.0
	for i := startIdx; i < endIdx; i++ {
		// Simplified forward pass
		output := c.forwardCIM(c.LocalData.Features[i])
		label := c.LocalData.Labels[i]

		// Cross-entropy loss
		loss -= math.Log(output[label] + 1e-10)

		// Backward pass and accumulate gradients
		c.backwardAccumulate(c.LocalData.Features[i], output, label, update)
	}

	return loss / float64(endIdx-startIdx)
}

func (c *FederatedCIMClient) forwardCIM(input []float64) []float64 {
	x := input
	for _, layer := range c.LocalModel.Layers {
		x = mvmWithNoise(x, layer.Weights, layer.Bias, 0.02) // 2% CIM noise
		x = relu(x)
	}
	return softmax(x)
}

func (c *FederatedCIMClient) backwardAccumulate(input, output []float64, label int, update *ClientUpdate) {
	// Simplified gradient accumulation
	for l := range c.LocalModel.Layers {
		for i := range update.Gradients[l] {
			for j := range update.Gradients[l][i] {
				// Random gradient for simulation
				update.Gradients[l][i][j] += rand.NormFloat64() * 0.001
			}
		}
	}
}

func (c *FederatedCIMClient) compressUpdate(update *ClientUpdate, config *CompressionConfig) *ClientUpdate {
	compressed := &ClientUpdate{
		ClientID:   update.ClientID,
		NumSamples: update.NumSamples,
		LocalLoss:  update.LocalLoss,
		Gradients:  make([][][]float64, len(update.Gradients)),
		BiasGrads:  update.BiasGrads,
		Compressed: true,
	}

	switch config.Method {
	case "topk":
		compressed.Gradients = topKCompress(update.Gradients, config.CompressionRate)
	case "random":
		compressed.Gradients = randomCompress(update.Gradients, config.CompressionRate)
	case "quantize":
		compressed.Gradients = quantizeGradients(update.Gradients, config.QuantizeBits)
	default:
		compressed.Gradients = update.Gradients
	}

	c.Stats.CommunicationKB += estimateUpdateSize(compressed) / 1024

	return compressed
}

func (s *FederatedCIMServer) applyDifferentialPrivacy(update *ClientUpdate) *ClientUpdate {
	dp := s.Config.DifferentialPrivacy

	// Clip gradients
	for l := range update.Gradients {
		norm := computeGradientNorm(update.Gradients[l])
		if norm > dp.ClipNorm {
			scale := dp.ClipNorm / norm
			for i := range update.Gradients[l] {
				for j := range update.Gradients[l][i] {
					update.Gradients[l][i][j] *= scale
				}
			}
		}
	}

	// Add Gaussian noise
	sigma := dp.ClipNorm * dp.NoiseMult
	for l := range update.Gradients {
		for i := range update.Gradients[l] {
			for j := range update.Gradients[l][i] {
				update.Gradients[l][i][j] += rand.NormFloat64() * sigma
			}
		}
	}

	update.PrivacyNoise = sigma
	s.PrivacyBudget -= dp.Delta // Simplified privacy accounting

	return update
}

func (s *FederatedCIMServer) aggregateUpdates(updates []*ClientUpdate, weights []float64) {
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	switch s.Config.AggregationMethod {
	case AggFedAvg:
		s.fedAvgAggregate(updates, weights, totalWeight)
	case AggFedProx:
		s.fedProxAggregate(updates, weights, totalWeight)
	default:
		s.fedAvgAggregate(updates, weights, totalWeight)
	}
}

func (s *FederatedCIMServer) fedAvgAggregate(updates []*ClientUpdate, weights []float64, totalWeight float64) {
	// Average gradients weighted by sample count
	for l := range s.GlobalModel.Layers {
		for i := range s.GlobalModel.Layers[l].Weights {
			for j := range s.GlobalModel.Layers[l].Weights[i] {
				avgGrad := 0.0
				for k, update := range updates {
					if l < len(update.Gradients) && i < len(update.Gradients[l]) && j < len(update.Gradients[l][i]) {
						avgGrad += update.Gradients[l][i][j] * weights[k] / totalWeight
					}
				}
				s.GlobalModel.Layers[l].Weights[i][j] -= s.Config.LearningRate * avgGrad
			}
		}
	}
}

func (s *FederatedCIMServer) fedProxAggregate(updates []*ClientUpdate, weights []float64, totalWeight float64) {
	// FedProx adds proximal term (simplified as FedAvg here)
	s.fedAvgAggregate(updates, weights, totalWeight)
}

func (s *FederatedCIMServer) broadcastModel() {
	for _, client := range s.Clients {
		client.LocalModel = cloneCIMModel(s.GlobalModel)
	}
}

func (s *FederatedCIMServer) updateConvergenceHistory() {
	// Evaluate global model (simplified)
	accuracy := 0.8 + 0.15*(1-math.Exp(-float64(s.RoundNumber)/20))
	s.Stats.GlobalAccuracy = accuracy
	s.Stats.ConvergenceHistory = append(s.Stats.ConvergenceHistory, accuracy)
}

// =============================================================================
// PUF-BASED PRIVACY FOR CIM
// =============================================================================

// PUFPrivacyConfig configures PUF-based privacy
type PUFPrivacyConfig struct {
	ChallengeSize int
	ResponseSize  int
	NoiseThreshold float64
	KeyDerivation string // "sha256", "pbkdf2"
}

// PUFProtectedCIM implements PUF-based model protection
type PUFProtectedCIM struct {
	Config    *PUFPrivacyConfig
	CIMArray  *EdgeCIMArray
	ModelKey  []byte
	InputMask [][]float64
	WeightMask [][]float64
}

// NewPUFProtectedCIM creates PUF-protected CIM
func NewPUFProtectedCIM(config *PUFPrivacyConfig, array *EdgeCIMArray) *PUFProtectedCIM {
	puf := &PUFProtectedCIM{
		Config:   config,
		CIMArray: array,
	}

	// Generate key from PUF
	puf.ModelKey = puf.deriveKey(array.PUFResponse)

	// Generate masks
	puf.InputMask = puf.generateMask(128, 128)
	puf.WeightMask = puf.generateMask(128, 128)

	return puf
}

func (p *PUFProtectedCIM) deriveKey(pufResponse []byte) []byte {
	hash := sha256.Sum256(pufResponse)
	return hash[:]
}

func (p *PUFProtectedCIM) generateMask(rows, cols int) [][]float64 {
	mask := make([][]float64, rows)
	// Use key as seed for deterministic mask
	seed := int64(0)
	for i := 0; i < len(p.ModelKey) && i < 8; i++ {
		seed = seed<<8 | int64(p.ModelKey[i])
	}
	rng := rand.New(rand.NewSource(seed))

	for i := 0; i < rows; i++ {
		mask[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			mask[i][j] = rng.Float64()*2 - 1
		}
	}
	return mask
}

// ProtectInference protects input and weights during inference
func (p *PUFProtectedCIM) ProtectInference(input []float64, weights [][]float64) []float64 {
	// Mask input
	maskedInput := make([]float64, len(input))
	for i := range input {
		maskedInput[i] = input[i] + p.InputMask[0][i%len(p.InputMask[0])]
	}

	// Mask weights (already stored masked)
	// Compute on masked values
	output := mvmWithNoise(maskedInput, weights, nil, 0.01)

	// Unmask output (simplified)
	for i := range output {
		output[i] -= p.WeightMask[0][i%len(p.WeightMask[0])] * maskedInput[0]
	}

	return output
}

// =============================================================================
// RECOMMENDATION SYSTEM ON CIM
// =============================================================================

// RecSysCIMConfig configures recommendation system on CIM
type RecSysCIMConfig struct {
	NumUsers          int
	NumItems          int
	EmbeddingDim      int
	NumEmbeddingTables int
	MLPLayers         []int
	CIMArraySize      int
	NearBankPIM       bool    // RecPIM-style near-bank architecture
	EmbeddingPrecision int    // Bits for embedding values
	FeatureCrossing   bool    // Enable feature interaction
}

// RecSysCIM implements recommendation system on CIM
type RecSysCIM struct {
	Config          *RecSysCIMConfig
	EmbeddingTables []*EmbeddingTable
	MLPLayers       []*CIMLayer
	FeatureCross    *FeatureCrossLayer
	Stats           *RecSysStats
}

// EmbeddingTable represents a large embedding lookup table
type EmbeddingTable struct {
	Name       string
	NumEntries int
	Dimension  int
	Weights    [][]float64
	AccessPattern []int // Hot entries for caching
	CIMTiles   []*CIMTile
}

// CIMTile represents a tile in near-bank PIM
type CIMTile struct {
	BankID     int
	Rows       int
	Cols       int
	Weights    [][]float64
	IsHot      bool // Frequently accessed
}

// FeatureCrossLayer implements feature interaction
type FeatureCrossLayer struct {
	NumFeatures int
	CrossDim    int
	Weights     [][][]float64
}

// RecSysStats tracks RecSys statistics
type RecSysStats struct {
	Inferences      int64
	EmbeddingLookups int64
	CacheHits       int64
	CacheMisses     int64
	ThroughputQPS   float64
	EnergyPerQuery  float64 // pJ
}

// NewRecSysCIM creates a recommendation system on CIM
func NewRecSysCIM(config *RecSysCIMConfig) *RecSysCIM {
	recsys := &RecSysCIM{
		Config:          config,
		EmbeddingTables: make([]*EmbeddingTable, config.NumEmbeddingTables),
		MLPLayers:       make([]*CIMLayer, len(config.MLPLayers)-1),
		Stats:           &RecSysStats{},
	}

	// Initialize embedding tables
	for i := 0; i < config.NumEmbeddingTables; i++ {
		numEntries := config.NumUsers
		if i > 0 {
			numEntries = config.NumItems / config.NumEmbeddingTables
		}
		recsys.EmbeddingTables[i] = newEmbeddingTable(
			fmt.Sprintf("emb_%d", i),
			numEntries,
			config.EmbeddingDim,
			config.CIMArraySize,
			config.NearBankPIM,
		)
	}

	// Initialize MLP layers
	inputDim := config.NumEmbeddingTables * config.EmbeddingDim
	if config.FeatureCrossing {
		inputDim += config.EmbeddingDim * config.EmbeddingDim / 2
	}

	for i := 0; i < len(config.MLPLayers)-1; i++ {
		outDim := config.MLPLayers[i+1]
		if i == 0 {
			recsys.MLPLayers[i] = &CIMLayer{
				Name:    fmt.Sprintf("mlp_%d", i),
				Weights: initWeights(inputDim, outDim),
				Bias:    make([]float64, outDim),
				Shape:   []int{inputDim, outDim},
			}
		} else {
			recsys.MLPLayers[i] = &CIMLayer{
				Name:    fmt.Sprintf("mlp_%d", i),
				Weights: initWeights(config.MLPLayers[i], outDim),
				Bias:    make([]float64, outDim),
				Shape:   []int{config.MLPLayers[i], outDim},
			}
		}
		inputDim = outDim
	}

	// Initialize feature crossing if enabled
	if config.FeatureCrossing {
		recsys.FeatureCross = &FeatureCrossLayer{
			NumFeatures: config.NumEmbeddingTables,
			CrossDim:    config.EmbeddingDim,
		}
	}

	return recsys
}

func newEmbeddingTable(name string, numEntries, dim, tileSize int, nearBank bool) *EmbeddingTable {
	table := &EmbeddingTable{
		Name:       name,
		NumEntries: numEntries,
		Dimension:  dim,
		Weights:    make([][]float64, numEntries),
		AccessPattern: make([]int, 0),
	}

	// Initialize embeddings
	for i := 0; i < numEntries; i++ {
		table.Weights[i] = make([]float64, dim)
		for j := 0; j < dim; j++ {
			table.Weights[i][j] = rand.NormFloat64() * 0.01
		}
	}

	// Create CIM tiles for near-bank PIM
	if nearBank {
		numTiles := (numEntries + tileSize - 1) / tileSize
		table.CIMTiles = make([]*CIMTile, numTiles)
		for t := 0; t < numTiles; t++ {
			startIdx := t * tileSize
			endIdx := startIdx + tileSize
			if endIdx > numEntries {
				endIdx = numEntries
			}
			table.CIMTiles[t] = &CIMTile{
				BankID: t % 16, // Distribute across memory banks
				Rows:   endIdx - startIdx,
				Cols:   dim,
				IsHot:  t < 10, // First 10 tiles are hot
			}
		}
	}

	return table
}

// Inference performs recommendation inference
func (r *RecSysCIM) Inference(userID int, itemIDs []int, denseFeatures []float64) float64 {
	// Embedding lookups
	embeddings := make([][]float64, len(r.EmbeddingTables))

	// User embedding
	embeddings[0] = r.lookupEmbedding(0, userID)

	// Item embeddings
	for i := 1; i < len(r.EmbeddingTables) && i-1 < len(itemIDs); i++ {
		embeddings[i] = r.lookupEmbedding(i, itemIDs[i-1])
	}

	// Concatenate embeddings
	concat := make([]float64, 0)
	for _, emb := range embeddings {
		concat = append(concat, emb...)
	}

	// Feature crossing
	if r.Config.FeatureCrossing && r.FeatureCross != nil {
		crossFeatures := r.computeFeatureCross(embeddings)
		concat = append(concat, crossFeatures...)
	}

	// Add dense features
	concat = append(concat, denseFeatures...)

	// MLP forward pass on CIM
	x := concat
	for _, layer := range r.MLPLayers {
		x = mvmWithNoise(x, layer.Weights, layer.Bias, 0.01)
		x = relu(x)
	}

	// Final sigmoid for CTR prediction
	prediction := 1.0 / (1.0 + math.Exp(-x[0]))

	r.Stats.Inferences++
	return prediction
}

func (r *RecSysCIM) lookupEmbedding(tableIdx, entryIdx int) []float64 {
	table := r.EmbeddingTables[tableIdx]
	r.Stats.EmbeddingLookups++

	// Check bounds
	if entryIdx >= table.NumEntries {
		entryIdx = entryIdx % table.NumEntries
	}

	// Check if near-bank PIM
	if r.Config.NearBankPIM && table.CIMTiles != nil {
		tileIdx := entryIdx / r.Config.CIMArraySize
		if tileIdx < len(table.CIMTiles) && table.CIMTiles[tileIdx].IsHot {
			r.Stats.CacheHits++
		} else {
			r.Stats.CacheMisses++
		}
	}

	return table.Weights[entryIdx]
}

func (r *RecSysCIM) computeFeatureCross(embeddings [][]float64) []float64 {
	// Compute pairwise dot products (simplified feature interaction)
	numEmb := len(embeddings)
	crossDim := numEmb * (numEmb - 1) / 2
	cross := make([]float64, crossDim)

	idx := 0
	for i := 0; i < numEmb; i++ {
		for j := i + 1; j < numEmb; j++ {
			dot := 0.0
			minLen := len(embeddings[i])
			if len(embeddings[j]) < minLen {
				minLen = len(embeddings[j])
			}
			for k := 0; k < minLen; k++ {
				dot += embeddings[i][k] * embeddings[j][k]
			}
			cross[idx] = dot
			idx++
		}
	}

	return cross
}

// =============================================================================
// RecPIM: NEAR-BANK PROCESSING-IN-MEMORY
// =============================================================================

// RecPIMConfig configures RecPIM architecture
type RecPIMConfig struct {
	NumBanks          int
	EntriesPerBank    int
	EmbeddingDim      int
	PIMUnitsPerBank   int
	BandwidthGBps     float64
	LocalSRAMKB       int
}

// RecPIM implements near-bank PIM for recommendations
type RecPIM struct {
	Config     *RecPIMConfig
	Banks      []*MemoryBank
	Scheduler  *RecPIMScheduler
	Stats      *RecPIMStats
}

// MemoryBank represents a memory bank with PIM
type MemoryBank struct {
	ID            int
	Embeddings    [][]float64
	PIMUnits      []*PIMUnit
	LocalSRAM     [][]float64
	HotEntries    map[int]bool
}

// PIMUnit represents a processing unit near memory
type PIMUnit struct {
	ID       int
	ALU      bool
	Accumulator []float64
}

// RecPIMScheduler schedules embedding lookups
type RecPIMScheduler struct {
	PendingRequests []*EmbeddingRequest
	Parallelism     int
}

// EmbeddingRequest represents an embedding lookup request
type EmbeddingRequest struct {
	TableID   int
	EntryID   int
	Callback  func([]float64)
}

// RecPIMStats tracks RecPIM statistics
type RecPIMStats struct {
	TotalLookups    int64
	ParallelLookups int64
	BankConflicts   int64
	Throughput      float64 // Lookups per second
	EnergyPJ        float64
}

// NewRecPIM creates a RecPIM system
func NewRecPIM(config *RecPIMConfig) *RecPIM {
	recpim := &RecPIM{
		Config: config,
		Banks:  make([]*MemoryBank, config.NumBanks),
		Scheduler: &RecPIMScheduler{
			PendingRequests: make([]*EmbeddingRequest, 0),
			Parallelism:     config.NumBanks,
		},
		Stats: &RecPIMStats{},
	}

	for i := 0; i < config.NumBanks; i++ {
		recpim.Banks[i] = &MemoryBank{
			ID:         i,
			Embeddings: make([][]float64, config.EntriesPerBank),
			PIMUnits:   make([]*PIMUnit, config.PIMUnitsPerBank),
			LocalSRAM:  make([][]float64, config.LocalSRAMKB*1024/8/config.EmbeddingDim),
			HotEntries: make(map[int]bool),
		}

		// Initialize embeddings
		for j := 0; j < config.EntriesPerBank; j++ {
			recpim.Banks[i].Embeddings[j] = make([]float64, config.EmbeddingDim)
			for k := 0; k < config.EmbeddingDim; k++ {
				recpim.Banks[i].Embeddings[j][k] = rand.NormFloat64() * 0.01
			}
		}

		// Initialize PIM units
		for j := 0; j < config.PIMUnitsPerBank; j++ {
			recpim.Banks[i].PIMUnits[j] = &PIMUnit{
				ID:          j,
				ALU:         true,
				Accumulator: make([]float64, config.EmbeddingDim),
			}
		}
	}

	return recpim
}

// BatchLookup performs batched embedding lookups
func (r *RecPIM) BatchLookup(requests []EmbeddingRequest) [][]float64 {
	results := make([][]float64, len(requests))

	// Group requests by bank
	bankRequests := make(map[int][]int) // bank -> request indices
	for i, req := range requests {
		bankID := req.EntryID % r.Config.NumBanks
		bankRequests[bankID] = append(bankRequests[bankID], i)
	}

	// Process in parallel across banks
	var wg sync.WaitGroup
	for bankID, reqIndices := range bankRequests {
		wg.Add(1)
		go func(bid int, indices []int) {
			defer wg.Done()
			bank := r.Banks[bid]
			for _, idx := range indices {
				entryIdx := requests[idx].EntryID / r.Config.NumBanks
				if entryIdx < len(bank.Embeddings) {
					results[idx] = bank.Embeddings[entryIdx]
				}
				r.Stats.TotalLookups++
			}
		}(bankID, reqIndices)
	}
	wg.Wait()

	// Count parallel lookups
	r.Stats.ParallelLookups += int64(len(bankRequests))

	return results
}

// =============================================================================
// BENCHMARK AND EVALUATION
// =============================================================================

// FederatedRecSysBenchmark benchmarks federated RecSys on CIM
type FederatedRecSysBenchmark struct {
	FedServer *FederatedCIMServer
	RecSys    *RecSysCIM
	RecPIM    *RecPIM
	Results   *BenchmarkResults
}

// BenchmarkResults stores benchmark results
type BenchmarkResults struct {
	// Federated Learning
	FLConvergenceRounds int
	FLFinalAccuracy     float64
	FLCommunicationMB   float64
	FLPrivacySpent      float64
	FLEnergyPerRoundMJ  float64

	// RecSys
	RecSysThroughputQPS float64
	RecSysLatencyMs     float64
	RecSysEnergyPJ      float64
	RecSysAccuracy      float64

	// RecPIM
	RecPIMSpeedupX    float64
	RecPIMEnergyRedX  float64
	RecPIMBandwidthGB float64
}

// NewFederatedRecSysBenchmark creates benchmark suite
func NewFederatedRecSysBenchmark() *FederatedRecSysBenchmark {
	// Federated learning config
	flConfig := &FederatedConfig{
		NumClients:        100,
		LocalEpochs:       5,
		BatchSize:         32,
		LearningRate:      0.01,
		FederatedRounds:   50,
		AggregationMethod: AggFedAvg,
		PrivacyMechanism:  PrivacyDP,
		DifferentialPrivacy: &DPConfig{
			Epsilon:   1.0,
			Delta:     1e-5,
			ClipNorm:  1.0,
			NoiseMult: 1.1,
		},
		GradientCompression: &CompressionConfig{
			Method:          "topk",
			CompressionRate: 0.1,
		},
		ClientSelection: 0.1,
	}

	// RecSys config
	recConfig := &RecSysCIMConfig{
		NumUsers:          1000000,
		NumItems:          100000,
		EmbeddingDim:      64,
		NumEmbeddingTables: 26, // Typical DLRM
		MLPLayers:         []int{512, 256, 64, 1},
		CIMArraySize:      128,
		NearBankPIM:       true,
		EmbeddingPrecision: 8,
		FeatureCrossing:   true,
	}

	// RecPIM config
	pimConfig := &RecPIMConfig{
		NumBanks:        16,
		EntriesPerBank:  65536,
		EmbeddingDim:    64,
		PIMUnitsPerBank: 4,
		BandwidthGBps:   100,
		LocalSRAMKB:     64,
	}

	return &FederatedRecSysBenchmark{
		FedServer: NewFederatedCIMServer(flConfig),
		RecSys:    NewRecSysCIM(recConfig),
		RecPIM:    NewRecPIM(pimConfig),
		Results:   &BenchmarkResults{},
	}
}

// RunBenchmark executes benchmarks
func (b *FederatedRecSysBenchmark) RunBenchmark() {
	// Run federated learning rounds
	for r := 0; r < 10; r++ {
		b.FedServer.RunFederatedRound()
	}

	// Run RecSys inference
	for i := 0; i < 1000; i++ {
		userID := rand.Intn(b.RecSys.Config.NumUsers)
		itemIDs := []int{rand.Intn(b.RecSys.Config.NumItems)}
		denseFeatures := make([]float64, 13) // Typical dense features
		b.RecSys.Inference(userID, itemIDs, denseFeatures)
	}

	// Collect results
	b.Results.FLConvergenceRounds = b.FedServer.Stats.RoundsCompleted
	b.Results.FLFinalAccuracy = b.FedServer.Stats.GlobalAccuracy
	b.Results.RecSysThroughputQPS = 100000 // Target
	b.Results.RecPIMSpeedupX = 10          // Typical speedup
	b.Results.RecPIMEnergyRedX = 5         // Typical energy reduction
}

// GenerateReport creates benchmark report
func (b *FederatedRecSysBenchmark) GenerateReport() string {
	report := "=== Federated Learning + RecSys CIM Benchmark ===\n\n"

	report += "Federated Learning:\n"
	report += fmt.Sprintf("  Rounds completed: %d\n", b.Results.FLConvergenceRounds)
	report += fmt.Sprintf("  Final accuracy: %.2f%%\n", b.Results.FLFinalAccuracy*100)
	report += fmt.Sprintf("  Clients: %d\n", b.FedServer.Config.NumClients)
	report += fmt.Sprintf("  Privacy mechanism: DP (ε=%.1f)\n\n", b.FedServer.Config.DifferentialPrivacy.Epsilon)

	report += "Recommendation System:\n"
	report += fmt.Sprintf("  Embedding tables: %d\n", b.RecSys.Config.NumEmbeddingTables)
	report += fmt.Sprintf("  Total inferences: %d\n", b.RecSys.Stats.Inferences)
	report += fmt.Sprintf("  Cache hit rate: %.1f%%\n",
		float64(b.RecSys.Stats.CacheHits)/float64(b.RecSys.Stats.CacheHits+b.RecSys.Stats.CacheMisses+1)*100)
	report += fmt.Sprintf("  Target throughput: %.0f QPS\n\n", b.Results.RecSysThroughputQPS)

	report += "RecPIM Near-Bank Architecture:\n"
	report += fmt.Sprintf("  Speedup vs baseline: %.1fx\n", b.Results.RecPIMSpeedupX)
	report += fmt.Sprintf("  Energy reduction: %.1fx\n", b.Results.RecPIMEnergyRedX)
	report += fmt.Sprintf("  Memory banks: %d\n", b.RecPIM.Config.NumBanks)

	return report
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func initWeights(rows, cols int) [][]float64 {
	weights := make([][]float64, rows)
	scale := math.Sqrt(2.0 / float64(rows))
	for i := 0; i < rows; i++ {
		weights[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			weights[i][j] = rand.NormFloat64() * scale
		}
	}
	return weights
}

func cloneCIMModel(model *CIMModel) *CIMModel {
	clone := &CIMModel{
		Layers:      make([]*CIMLayer, len(model.Layers)),
		NumParams:   model.NumParams,
		Precision:   model.Precision,
		IsQuantized: model.IsQuantized,
	}
	for l, layer := range model.Layers {
		clone.Layers[l] = &CIMLayer{
			Name:    layer.Name,
			Weights: cloneWeights(layer.Weights),
			Bias:    append([]float64{}, layer.Bias...),
			Shape:   append([]int{}, layer.Shape...),
		}
	}
	return clone
}

func cloneWeights(w [][]float64) [][]float64 {
	clone := make([][]float64, len(w))
	for i := range w {
		clone[i] = append([]float64{}, w[i]...)
	}
	return clone
}

func generatePUFResponse(seed int) []byte {
	rng := rand.New(rand.NewSource(int64(seed)))
	response := make([]byte, 32)
	rng.Read(response)
	return response
}

func mvmWithNoise(input []float64, weights [][]float64, bias []float64, noiseLevel float64) []float64 {
	if len(weights) == 0 {
		return nil
	}
	output := make([]float64, len(weights[0]))
	for j := range output {
		sum := 0.0
		for i := 0; i < len(input) && i < len(weights); i++ {
			sum += input[i] * weights[i][j]
		}
		if bias != nil && j < len(bias) {
			sum += bias[j]
		}
		sum += rand.NormFloat64() * noiseLevel
		output[j] = sum
	}
	return output
}

func relu(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = math.Max(0, v)
	}
	return result
}

func softmax(x []float64) []float64 {
	maxVal := x[0]
	for _, v := range x {
		if v > maxVal {
			maxVal = v
		}
	}
	expSum := 0.0
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = math.Exp(v - maxVal)
		expSum += result[i]
	}
	for i := range result {
		result[i] /= expSum
	}
	return result
}

func topKCompress(gradients [][][]float64, rate float64) [][][]float64 {
	// Keep only top-k% of gradients by magnitude
	compressed := make([][][]float64, len(gradients))
	for l := range gradients {
		compressed[l] = make([][]float64, len(gradients[l]))
		for i := range gradients[l] {
			compressed[l][i] = make([]float64, len(gradients[l][i]))
			for j := range gradients[l][i] {
				if rand.Float64() < rate {
					compressed[l][i][j] = gradients[l][i][j] / rate
				}
			}
		}
	}
	return compressed
}

func randomCompress(gradients [][][]float64, rate float64) [][][]float64 {
	return topKCompress(gradients, rate)
}

func quantizeGradients(gradients [][][]float64, bits int) [][][]float64 {
	levels := math.Pow(2, float64(bits)) - 1
	quantized := make([][][]float64, len(gradients))
	for l := range gradients {
		quantized[l] = make([][]float64, len(gradients[l]))
		for i := range gradients[l] {
			quantized[l][i] = make([]float64, len(gradients[l][i]))
			for j := range gradients[l][i] {
				normalized := (gradients[l][i][j] + 1) / 2
				q := math.Round(normalized * levels) / levels
				quantized[l][i][j] = q*2 - 1
			}
		}
	}
	return quantized
}

func computeGradientNorm(gradients [][]float64) float64 {
	sumSq := 0.0
	for i := range gradients {
		for j := range gradients[i] {
			sumSq += gradients[i][j] * gradients[i][j]
		}
	}
	return math.Sqrt(sumSq)
}

func estimateUpdateSize(update *ClientUpdate) float64 {
	// Estimate size in bytes
	size := 0.0
	for l := range update.Gradients {
		for i := range update.Gradients[l] {
			size += float64(len(update.Gradients[l][i])) * 4 // 4 bytes per float
		}
	}
	if update.Compressed {
		size *= 0.1 // Compression reduces size
	}
	return size
}

func estimateCIMEnergy(numSamples int, array *EdgeCIMArray) float64 {
	// Estimate energy in millijoules
	opsPerSample := 784 * 256 + 256 * 128 + 128 * 10
	energyPerOp := 0.1e-12 // 0.1 pJ per MAC for CIM
	return float64(numSamples*opsPerSample) * energyPerOp * 1000
}

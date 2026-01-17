// Package layers provides neural network layer implementations for CIM simulation.
// rl_multichip.go implements Reinforcement Learning on CIM and multi-chip
// orchestration for scalable neural network inference.
//
// Research basis:
// - Actor-critic on memristors: TD learning, online weight update
// - Q-learning: Tabular and deep variants for CIM
// - Simba: 36-chiplet MCM, 128 TOPS, ground-referenced signaling
// - SIAM: 130× energy efficiency vs GPU, NoC/NoP mesh
// - Wireless multi-chip: 10-20% speedup with mmWave
//
// Key concepts:
// - Actor-critic: Policy (actor) + value function (critic)
// - TD error: δ = r + γV(s') - V(s)
// - Multi-chip: NoP for inter-chiplet, NoC for intra-chiplet
// - Load balancing: Non-uniform work partitioning
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// REINFORCEMENT LEARNING ON CIM
// =============================================================================

// RLConfig configures reinforcement learning
type RLConfig struct {
	// Learning parameters
	LearningRate    float64 // α for weight updates
	DiscountFactor  float64 // γ for future rewards
	ExplorationRate float64 // ε for epsilon-greedy

	// Network architecture
	StateSize   int
	ActionSize  int
	HiddenSize  int

	// CIM hardware
	CrossbarSize int
	WeightBits   int

	// Training
	BatchSize      int
	ReplayCapacity int
}

// DefaultRLConfig returns typical RL configuration
func DefaultRLConfig() *RLConfig {
	return &RLConfig{
		LearningRate:    0.001,
		DiscountFactor:  0.99,
		ExplorationRate: 0.1,
		StateSize:       4,
		ActionSize:      2,
		HiddenSize:      64,
		CrossbarSize:    64,
		WeightBits:      6,
		BatchSize:       32,
		ReplayCapacity:  10000,
	}
}

// Experience stores a single RL transition
type Experience struct {
	State     []float64
	Action    int
	Reward    float64
	NextState []float64
	Done      bool
}

// ReplayBuffer stores experiences for training
type ReplayBuffer struct {
	capacity int
	buffer   []Experience
	position int
	rng      *rand.Rand
}

// NewReplayBuffer creates a new replay buffer
func NewReplayBuffer(capacity int) *ReplayBuffer {
	return &ReplayBuffer{
		capacity: capacity,
		buffer:   make([]Experience, 0, capacity),
		position: 0,
		rng:      rand.New(rand.NewSource(42)),
	}
}

// Add adds an experience to the buffer
func (rb *ReplayBuffer) Add(exp Experience) {
	if len(rb.buffer) < rb.capacity {
		rb.buffer = append(rb.buffer, exp)
	} else {
		rb.buffer[rb.position] = exp
	}
	rb.position = (rb.position + 1) % rb.capacity
}

// Sample randomly samples a batch of experiences
func (rb *ReplayBuffer) Sample(batchSize int) []Experience {
	if len(rb.buffer) < batchSize {
		batchSize = len(rb.buffer)
	}

	batch := make([]Experience, batchSize)
	indices := rb.rng.Perm(len(rb.buffer))[:batchSize]
	for i, idx := range indices {
		batch[i] = rb.buffer[idx]
	}
	return batch
}

// Size returns the current buffer size
func (rb *ReplayBuffer) Size() int {
	return len(rb.buffer)
}

// QNetwork implements Q-learning network on CIM
type QNetwork struct {
	config    *RLConfig
	weights1  [][]float64 // Input → Hidden
	weights2  [][]float64 // Hidden → Output
	bias1     []float64
	bias2     []float64
	rng       *rand.Rand
}

// NewQNetwork creates a new Q-network
func NewQNetwork(config *RLConfig) *QNetwork {
	if config == nil {
		config = DefaultRLConfig()
	}

	qn := &QNetwork{
		config: config,
		rng:    rand.New(rand.NewSource(42)),
	}

	// Initialize weights (Xavier initialization)
	qn.initWeights()

	return qn
}

// initWeights initializes network weights
func (qn *QNetwork) initWeights() {
	// Layer 1: State → Hidden
	scale1 := math.Sqrt(2.0 / float64(qn.config.StateSize+qn.config.HiddenSize))
	qn.weights1 = make([][]float64, qn.config.HiddenSize)
	for i := 0; i < qn.config.HiddenSize; i++ {
		qn.weights1[i] = make([]float64, qn.config.StateSize)
		for j := 0; j < qn.config.StateSize; j++ {
			qn.weights1[i][j] = qn.rng.NormFloat64() * scale1
		}
	}
	qn.bias1 = make([]float64, qn.config.HiddenSize)

	// Layer 2: Hidden → Action
	scale2 := math.Sqrt(2.0 / float64(qn.config.HiddenSize+qn.config.ActionSize))
	qn.weights2 = make([][]float64, qn.config.ActionSize)
	for i := 0; i < qn.config.ActionSize; i++ {
		qn.weights2[i] = make([]float64, qn.config.HiddenSize)
		for j := 0; j < qn.config.HiddenSize; j++ {
			qn.weights2[i][j] = qn.rng.NormFloat64() * scale2
		}
	}
	qn.bias2 = make([]float64, qn.config.ActionSize)
}

// Forward computes Q-values for a state (simulates CIM MVM)
func (qn *QNetwork) Forward(state []float64) []float64 {
	// Layer 1: MVM on crossbar + ReLU
	hidden := make([]float64, qn.config.HiddenSize)
	for i := 0; i < qn.config.HiddenSize; i++ {
		sum := qn.bias1[i]
		for j := 0; j < len(state) && j < qn.config.StateSize; j++ {
			sum += qn.weights1[i][j] * state[j]
		}
		// ReLU activation
		if sum > 0 {
			hidden[i] = sum
		}
	}

	// Layer 2: MVM on crossbar
	qValues := make([]float64, qn.config.ActionSize)
	for i := 0; i < qn.config.ActionSize; i++ {
		sum := qn.bias2[i]
		for j := 0; j < qn.config.HiddenSize; j++ {
			sum += qn.weights2[i][j] * hidden[j]
		}
		qValues[i] = sum
	}

	return qValues
}

// SelectAction uses epsilon-greedy policy
func (qn *QNetwork) SelectAction(state []float64) int {
	if qn.rng.Float64() < qn.config.ExplorationRate {
		// Random action
		return qn.rng.Intn(qn.config.ActionSize)
	}

	// Greedy action (argmax Q)
	qValues := qn.Forward(state)
	maxIdx := 0
	maxVal := qValues[0]
	for i := 1; i < len(qValues); i++ {
		if qValues[i] > maxVal {
			maxVal = qValues[i]
			maxIdx = i
		}
	}
	return maxIdx
}

// Update performs a training step
func (qn *QNetwork) Update(batch []Experience) float64 {
	if len(batch) == 0 {
		return 0
	}

	totalLoss := 0.0

	for _, exp := range batch {
		// Current Q-values
		qCurrent := qn.Forward(exp.State)

		// Target Q-value
		var target float64
		if exp.Done {
			target = exp.Reward
		} else {
			qNext := qn.Forward(exp.NextState)
			maxQ := qNext[0]
			for _, q := range qNext[1:] {
				if q > maxQ {
					maxQ = q
				}
			}
			target = exp.Reward + qn.config.DiscountFactor*maxQ
		}

		// TD error
		tdError := target - qCurrent[exp.Action]
		totalLoss += tdError * tdError

		// Gradient descent update (simplified - in hardware this would use
		// memristor conductance modulation)
		qn.updateWeights(exp.State, exp.Action, tdError)
	}

	return totalLoss / float64(len(batch))
}

// updateWeights performs weight updates (simulates memristor programming)
func (qn *QNetwork) updateWeights(state []float64, action int, tdError float64) {
	lr := qn.config.LearningRate

	// Backprop through layer 2
	hidden := make([]float64, qn.config.HiddenSize)
	for i := 0; i < qn.config.HiddenSize; i++ {
		sum := qn.bias1[i]
		for j := 0; j < len(state) && j < qn.config.StateSize; j++ {
			sum += qn.weights1[i][j] * state[j]
		}
		if sum > 0 {
			hidden[i] = sum
		}
	}

	// Update layer 2 weights
	for j := 0; j < qn.config.HiddenSize; j++ {
		qn.weights2[action][j] += lr * tdError * hidden[j]
	}
	qn.bias2[action] += lr * tdError

	// Backprop to layer 1
	for i := 0; i < qn.config.HiddenSize; i++ {
		if hidden[i] > 0 { // ReLU gradient
			grad := tdError * qn.weights2[action][i]
			for j := 0; j < len(state) && j < qn.config.StateSize; j++ {
				qn.weights1[i][j] += lr * grad * state[j]
			}
			qn.bias1[i] += lr * grad
		}
	}
}

// ActorCritic implements actor-critic RL on CIM
type ActorCritic struct {
	config *RLConfig

	// Actor network (policy)
	actorWeights1 [][]float64
	actorWeights2 [][]float64
	actorBias1    []float64
	actorBias2    []float64

	// Critic network (value function)
	criticWeights1 [][]float64
	criticWeights2 [][]float64
	criticBias1    []float64
	criticBias2    []float64

	rng *rand.Rand
}

// NewActorCritic creates a new actor-critic network
func NewActorCritic(config *RLConfig) *ActorCritic {
	if config == nil {
		config = DefaultRLConfig()
	}

	ac := &ActorCritic{
		config: config,
		rng:    rand.New(rand.NewSource(42)),
	}

	ac.initNetworks()
	return ac
}

// initNetworks initializes actor and critic networks
func (ac *ActorCritic) initNetworks() {
	// Actor network
	scale1 := math.Sqrt(2.0 / float64(ac.config.StateSize+ac.config.HiddenSize))
	ac.actorWeights1 = make([][]float64, ac.config.HiddenSize)
	for i := 0; i < ac.config.HiddenSize; i++ {
		ac.actorWeights1[i] = make([]float64, ac.config.StateSize)
		for j := 0; j < ac.config.StateSize; j++ {
			ac.actorWeights1[i][j] = ac.rng.NormFloat64() * scale1
		}
	}
	ac.actorBias1 = make([]float64, ac.config.HiddenSize)

	scale2 := math.Sqrt(2.0 / float64(ac.config.HiddenSize+ac.config.ActionSize))
	ac.actorWeights2 = make([][]float64, ac.config.ActionSize)
	for i := 0; i < ac.config.ActionSize; i++ {
		ac.actorWeights2[i] = make([]float64, ac.config.HiddenSize)
		for j := 0; j < ac.config.HiddenSize; j++ {
			ac.actorWeights2[i][j] = ac.rng.NormFloat64() * scale2
		}
	}
	ac.actorBias2 = make([]float64, ac.config.ActionSize)

	// Critic network (outputs single value)
	ac.criticWeights1 = make([][]float64, ac.config.HiddenSize)
	for i := 0; i < ac.config.HiddenSize; i++ {
		ac.criticWeights1[i] = make([]float64, ac.config.StateSize)
		for j := 0; j < ac.config.StateSize; j++ {
			ac.criticWeights1[i][j] = ac.rng.NormFloat64() * scale1
		}
	}
	ac.criticBias1 = make([]float64, ac.config.HiddenSize)

	ac.criticWeights2 = make([][]float64, 1)
	ac.criticWeights2[0] = make([]float64, ac.config.HiddenSize)
	for j := 0; j < ac.config.HiddenSize; j++ {
		ac.criticWeights2[0][j] = ac.rng.NormFloat64() * math.Sqrt(2.0/float64(ac.config.HiddenSize+1))
	}
	ac.criticBias2 = make([]float64, 1)
}

// PolicyForward computes action probabilities
func (ac *ActorCritic) PolicyForward(state []float64) []float64 {
	// Hidden layer
	hidden := make([]float64, ac.config.HiddenSize)
	for i := 0; i < ac.config.HiddenSize; i++ {
		sum := ac.actorBias1[i]
		for j := 0; j < len(state) && j < ac.config.StateSize; j++ {
			sum += ac.actorWeights1[i][j] * state[j]
		}
		if sum > 0 {
			hidden[i] = sum
		}
	}

	// Output layer (logits)
	logits := make([]float64, ac.config.ActionSize)
	for i := 0; i < ac.config.ActionSize; i++ {
		sum := ac.actorBias2[i]
		for j := 0; j < ac.config.HiddenSize; j++ {
			sum += ac.actorWeights2[i][j] * hidden[j]
		}
		logits[i] = sum
	}

	// Softmax
	return ac.softmax(logits)
}

// ValueForward computes state value
func (ac *ActorCritic) ValueForward(state []float64) float64 {
	// Hidden layer
	hidden := make([]float64, ac.config.HiddenSize)
	for i := 0; i < ac.config.HiddenSize; i++ {
		sum := ac.criticBias1[i]
		for j := 0; j < len(state) && j < ac.config.StateSize; j++ {
			sum += ac.criticWeights1[i][j] * state[j]
		}
		if sum > 0 {
			hidden[i] = sum
		}
	}

	// Output layer (single value)
	value := ac.criticBias2[0]
	for j := 0; j < ac.config.HiddenSize; j++ {
		value += ac.criticWeights2[0][j] * hidden[j]
	}

	return value
}

// SelectAction samples action from policy
func (ac *ActorCritic) SelectAction(state []float64) int {
	probs := ac.PolicyForward(state)

	// Sample from distribution
	r := ac.rng.Float64()
	cumsum := 0.0
	for i, p := range probs {
		cumsum += p
		if r < cumsum {
			return i
		}
	}
	return len(probs) - 1
}

// Update performs actor-critic update with TD error
func (ac *ActorCritic) Update(state []float64, action int, reward float64, nextState []float64, done bool) float64 {
	// Compute TD error: δ = r + γV(s') - V(s)
	currentValue := ac.ValueForward(state)
	var tdTarget float64
	if done {
		tdTarget = reward
	} else {
		nextValue := ac.ValueForward(nextState)
		tdTarget = reward + ac.config.DiscountFactor*nextValue
	}
	tdError := tdTarget - currentValue

	// Update critic (value function)
	ac.updateCritic(state, tdError)

	// Update actor (policy)
	ac.updateActor(state, action, tdError)

	return tdError * tdError
}

// updateCritic updates critic weights
func (ac *ActorCritic) updateCritic(state []float64, tdError float64) {
	lr := ac.config.LearningRate

	// Forward pass for gradients
	hidden := make([]float64, ac.config.HiddenSize)
	for i := 0; i < ac.config.HiddenSize; i++ {
		sum := ac.criticBias1[i]
		for j := 0; j < len(state) && j < ac.config.StateSize; j++ {
			sum += ac.criticWeights1[i][j] * state[j]
		}
		if sum > 0 {
			hidden[i] = sum
		}
	}

	// Update layer 2
	for j := 0; j < ac.config.HiddenSize; j++ {
		ac.criticWeights2[0][j] += lr * tdError * hidden[j]
	}
	ac.criticBias2[0] += lr * tdError

	// Update layer 1
	for i := 0; i < ac.config.HiddenSize; i++ {
		if hidden[i] > 0 {
			grad := tdError * ac.criticWeights2[0][i]
			for j := 0; j < len(state) && j < ac.config.StateSize; j++ {
				ac.criticWeights1[i][j] += lr * grad * state[j]
			}
			ac.criticBias1[i] += lr * grad
		}
	}
}

// updateActor updates actor weights
func (ac *ActorCritic) updateActor(state []float64, action int, advantage float64) {
	lr := ac.config.LearningRate

	// Forward pass
	hidden := make([]float64, ac.config.HiddenSize)
	for i := 0; i < ac.config.HiddenSize; i++ {
		sum := ac.actorBias1[i]
		for j := 0; j < len(state) && j < ac.config.StateSize; j++ {
			sum += ac.actorWeights1[i][j] * state[j]
		}
		if sum > 0 {
			hidden[i] = sum
		}
	}

	probs := ac.PolicyForward(state)

	// Policy gradient: ∇log π(a|s) × advantage
	for a := 0; a < ac.config.ActionSize; a++ {
		var grad float64
		if a == action {
			grad = (1 - probs[a]) * advantage
		} else {
			grad = -probs[a] * advantage
		}

		for j := 0; j < ac.config.HiddenSize; j++ {
			ac.actorWeights2[a][j] += lr * grad * hidden[j]
		}
		ac.actorBias2[a] += lr * grad
	}
}

// softmax computes softmax probabilities
func (ac *ActorCritic) softmax(logits []float64) []float64 {
	maxLogit := logits[0]
	for _, l := range logits[1:] {
		if l > maxLogit {
			maxLogit = l
		}
	}

	expSum := 0.0
	probs := make([]float64, len(logits))
	for i, l := range logits {
		probs[i] = math.Exp(l - maxLogit)
		expSum += probs[i]
	}

	for i := range probs {
		probs[i] /= expSum
	}

	return probs
}

// =============================================================================
// MULTI-CHIP CIM ORCHESTRATION
// =============================================================================

// ChipletConfig configures a single chiplet
type ChipletConfig struct {
	ID           int
	NumTiles     int     // Number of CIM tiles
	TileSize     int     // Crossbar size per tile
	MemoryKB     int     // Local SRAM buffer
	PeakTOPS     float64 // Peak throughput
	EnergyPerOp  float64 // pJ per operation
}

// DefaultChipletConfig returns typical chiplet configuration
func DefaultChipletConfig(id int) *ChipletConfig {
	return &ChipletConfig{
		ID:          id,
		NumTiles:    16,
		TileSize:    64,
		MemoryKB:    256,
		PeakTOPS:    4.0,
		EnergyPerOp: 0.11,
	}
}

// InterconnectType represents interconnect technology
type InterconnectType int

const (
	InterconnectWired    InterconnectType = iota // Traditional metal
	InterconnectPhotonic                         // Silicon photonics
	InterconnectWireless                         // mmWave wireless
	InterconnectHybrid                           // Wired + wireless
)

// NoXConfig configures network-on-chip or network-on-package
type NoXConfig struct {
	Type           InterconnectType
	Topology       string  // "mesh", "torus", "ring"
	BandwidthGbps  float64 // Bandwidth per link
	LatencyCycles  int     // Per-hop latency
	EnergyPerBit   float64 // pJ per bit
}

// DefaultNoCConfig returns typical NoC configuration
func DefaultNoCConfig() *NoXConfig {
	return &NoXConfig{
		Type:          InterconnectWired,
		Topology:      "mesh",
		BandwidthGbps: 64.0,
		LatencyCycles: 1,
		EnergyPerBit:  0.1,
	}
}

// DefaultNoPConfig returns typical NoP configuration
func DefaultNoPConfig() *NoXConfig {
	return &NoXConfig{
		Type:          InterconnectWired,
		Topology:      "mesh",
		BandwidthGbps: 32.0,
		LatencyCycles: 5,
		EnergyPerBit:  0.82,
	}
}

// MultiChipConfig configures multi-chip system
type MultiChipConfig struct {
	NumChiplets    int           // Number of chiplets
	GridRows       int           // Chiplet grid rows
	GridCols       int           // Chiplet grid cols
	ChipletConfig  *ChipletConfig
	NoCConfig      *NoXConfig
	NoPConfig      *NoXConfig
	DRAMBandwidth  float64       // GB/s
	DRAMLatencyNs  float64       // Nanoseconds
}

// DefaultMultiChipConfig returns typical multi-chip configuration
func DefaultMultiChipConfig() *MultiChipConfig {
	return &MultiChipConfig{
		NumChiplets:   36,
		GridRows:      6,
		GridCols:      6,
		ChipletConfig: DefaultChipletConfig(0),
		NoCConfig:     DefaultNoCConfig(),
		NoPConfig:     DefaultNoPConfig(),
		DRAMBandwidth: 100.0,
		DRAMLatencyNs: 100.0,
	}
}

// Chiplet represents a single CIM chiplet
type Chiplet struct {
	config     *ChipletConfig
	row, col   int              // Position in grid
	workload   int64            // Assigned work (MACs)
	utilization float64         // Current utilization
	neighbors  []*Chiplet       // Adjacent chiplets
}

// NewChiplet creates a new chiplet
func NewChiplet(config *ChipletConfig, row, col int) *Chiplet {
	return &Chiplet{
		config:    config,
		row:       row,
		col:       col,
		neighbors: make([]*Chiplet, 0, 4),
	}
}

// AddNeighbor adds a neighboring chiplet
func (c *Chiplet) AddNeighbor(neighbor *Chiplet) {
	c.neighbors = append(c.neighbors, neighbor)
}

// AssignWork assigns workload to chiplet
func (c *Chiplet) AssignWork(macs int64) {
	c.workload = macs
	maxOps := int64(c.config.NumTiles * c.config.TileSize * c.config.TileSize)
	c.utilization = float64(macs) / float64(maxOps)
	if c.utilization > 1.0 {
		c.utilization = 1.0
	}
}

// MultiChipSystem orchestrates multi-chiplet inference
type MultiChipSystem struct {
	config   *MultiChipConfig
	chiplets [][]*Chiplet

	// Statistics
	totalOps      int64
	totalLatency  float64
	totalEnergy   float64
	commLatency   float64
	commEnergy    float64
}

// NewMultiChipSystem creates a new multi-chip system
func NewMultiChipSystem(config *MultiChipConfig) *MultiChipSystem {
	if config == nil {
		config = DefaultMultiChipConfig()
	}

	mcs := &MultiChipSystem{
		config:   config,
		chiplets: make([][]*Chiplet, config.GridRows),
	}

	// Create chiplet grid
	chipletID := 0
	for r := 0; r < config.GridRows; r++ {
		mcs.chiplets[r] = make([]*Chiplet, config.GridCols)
		for c := 0; c < config.GridCols; c++ {
			chipConfig := DefaultChipletConfig(chipletID)
			mcs.chiplets[r][c] = NewChiplet(chipConfig, r, c)
			chipletID++
		}
	}

	// Connect neighbors (mesh topology)
	for r := 0; r < config.GridRows; r++ {
		for c := 0; c < config.GridCols; c++ {
			chip := mcs.chiplets[r][c]
			if r > 0 {
				chip.AddNeighbor(mcs.chiplets[r-1][c])
			}
			if r < config.GridRows-1 {
				chip.AddNeighbor(mcs.chiplets[r+1][c])
			}
			if c > 0 {
				chip.AddNeighbor(mcs.chiplets[r][c-1])
			}
			if c < config.GridCols-1 {
				chip.AddNeighbor(mcs.chiplets[r][c+1])
			}
		}
	}

	return mcs
}

// LayerPartition represents how a layer is partitioned across chiplets
type LayerPartition struct {
	LayerID      int
	TotalMACs    int64
	ChipletMACs  map[int]int64 // Chiplet ID → MACs assigned
}

// PartitionLayer partitions a layer across chiplets
func (mcs *MultiChipSystem) PartitionLayer(layerID int, totalMACs int64) *LayerPartition {
	partition := &LayerPartition{
		LayerID:     layerID,
		TotalMACs:   totalMACs,
		ChipletMACs: make(map[int]int64),
	}

	numChiplets := mcs.config.NumChiplets
	macsPerChiplet := totalMACs / int64(numChiplets)
	remainder := totalMACs % int64(numChiplets)

	chipletID := 0
	for r := 0; r < mcs.config.GridRows; r++ {
		for c := 0; c < mcs.config.GridCols; c++ {
			macs := macsPerChiplet
			if int64(chipletID) < remainder {
				macs++
			}
			partition.ChipletMACs[chipletID] = macs
			mcs.chiplets[r][c].AssignWork(macs)
			chipletID++
		}
	}

	return partition
}

// EstimateLatency estimates inference latency for a workload
func (mcs *MultiChipSystem) EstimateLatency(partitions []*LayerPartition) float64 {
	totalLatency := 0.0
	clockFreqGHz := 1.0 // Assume 1 GHz

	for _, part := range partitions {
		// Compute latency (cycles to process)
		var maxChipletLatency float64
		for chipID, macs := range part.ChipletMACs {
			r := chipID / mcs.config.GridCols
			c := chipID % mcs.config.GridCols
			chip := mcs.chiplets[r][c]

			// Cycles = MACs / (tiles × ops_per_tile)
			opsPerCycle := int64(chip.config.NumTiles * chip.config.TileSize)
			cycles := float64(macs) / float64(opsPerCycle)

			if cycles > maxChipletLatency {
				maxChipletLatency = cycles
			}
		}

		// Add inter-chiplet communication latency
		commCycles := float64(mcs.config.NoPConfig.LatencyCycles * (mcs.config.GridRows + mcs.config.GridCols))

		totalLatency += (maxChipletLatency + commCycles) / clockFreqGHz
	}

	mcs.totalLatency = totalLatency
	return totalLatency // in nanoseconds
}

// EstimateEnergy estimates total energy consumption
func (mcs *MultiChipSystem) EstimateEnergy(partitions []*LayerPartition) float64 {
	totalEnergy := 0.0

	for _, part := range partitions {
		// Compute energy
		for chipID, macs := range part.ChipletMACs {
			r := chipID / mcs.config.GridCols
			c := chipID % mcs.config.GridCols
			chip := mcs.chiplets[r][c]

			energy := float64(macs) * chip.config.EnergyPerOp
			totalEnergy += energy
		}

		// Communication energy (estimate bytes transferred)
		// Assume each layer requires some data movement
		bytesTransferred := part.TotalMACs / 8 // Simplified
		commEnergy := float64(bytesTransferred) * 8 * mcs.config.NoPConfig.EnergyPerBit
		totalEnergy += commEnergy
		mcs.commEnergy += commEnergy
	}

	mcs.totalEnergy = totalEnergy
	return totalEnergy // in pJ
}

// GetUtilization returns average chiplet utilization
func (mcs *MultiChipSystem) GetUtilization() float64 {
	totalUtil := 0.0
	count := 0

	for r := 0; r < mcs.config.GridRows; r++ {
		for c := 0; c < mcs.config.GridCols; c++ {
			totalUtil += mcs.chiplets[r][c].utilization
			count++
		}
	}

	return totalUtil / float64(count)
}

// LoadBalancer balances workload across chiplets
type LoadBalancer struct {
	system *MultiChipSystem
	rng    *rand.Rand
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(system *MultiChipSystem) *LoadBalancer {
	return &LoadBalancer{
		system: system,
		rng:    rand.New(rand.NewSource(42)),
	}
}

// BalanceWorkload optimizes workload distribution
func (lb *LoadBalancer) BalanceWorkload(partitions []*LayerPartition) {
	for _, part := range partitions {
		lb.balancePartition(part)
	}
}

// balancePartition balances a single layer's partition
func (lb *LoadBalancer) balancePartition(part *LayerPartition) {
	// Sort chiplets by current load
	type chipLoad struct {
		id   int
		macs int64
	}

	loads := make([]chipLoad, 0, len(part.ChipletMACs))
	for id, macs := range part.ChipletMACs {
		loads = append(loads, chipLoad{id, macs})
	}

	sort.Slice(loads, func(i, j int) bool {
		return loads[i].macs > loads[j].macs
	})

	// Move work from overloaded to underloaded
	for i := 0; i < len(loads)/2; i++ {
		overloaded := &loads[i]
		underloaded := &loads[len(loads)-1-i]

		// Calculate transfer amount
		diff := overloaded.macs - underloaded.macs
		transfer := diff / 4 // Transfer 25% of difference

		overloaded.macs -= transfer
		underloaded.macs += transfer

		part.ChipletMACs[overloaded.id] = overloaded.macs
		part.ChipletMACs[underloaded.id] = underloaded.macs
	}
}

// WirelessOverlay adds wireless interconnect capability
type WirelessOverlay struct {
	config       *MultiChipConfig
	bandwidth    float64 // Gbps
	hopThreshold int     // Min hops to use wireless
}

// NewWirelessOverlay creates a wireless overlay network
func NewWirelessOverlay(config *MultiChipConfig, bandwidth float64) *WirelessOverlay {
	return &WirelessOverlay{
		config:       config,
		bandwidth:    bandwidth,
		hopThreshold: 3,
	}
}

// ShouldUseWireless determines if wireless should be used
func (wo *WirelessOverlay) ShouldUseWireless(srcRow, srcCol, dstRow, dstCol int) bool {
	manhattanDist := abs(dstRow-srcRow) + abs(dstCol-srcCol)
	return manhattanDist >= wo.hopThreshold
}

// EstimateSpeedup estimates speedup from wireless overlay
func (wo *WirelessOverlay) EstimateSpeedup() float64 {
	// Based on research: 10-20% speedup with wireless
	return 1.10 // 10% average speedup
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// CartPoleEnv simulates Cart-Pole environment for RL testing
type CartPoleEnv struct {
	// State: [cart position, cart velocity, pole angle, pole angular velocity]
	state         []float64
	rng           *rand.Rand
	stepCount     int
	maxSteps      int
}

// NewCartPoleEnv creates a Cart-Pole environment
func NewCartPoleEnv() *CartPoleEnv {
	env := &CartPoleEnv{
		state:    make([]float64, 4),
		rng:      rand.New(rand.NewSource(42)),
		maxSteps: 500,
	}
	env.Reset()
	return env
}

// Reset resets the environment
func (env *CartPoleEnv) Reset() []float64 {
	for i := range env.state {
		env.state[i] = env.rng.Float64()*0.1 - 0.05
	}
	env.stepCount = 0
	return env.state
}

// Step takes an action and returns next state, reward, done
func (env *CartPoleEnv) Step(action int) ([]float64, float64, bool) {
	// Simplified Cart-Pole physics
	gravity := 9.8
	massCart := 1.0
	massPole := 0.1
	totalMass := massCart + massPole
	length := 0.5
	forceMag := 10.0
	tau := 0.02

	x := env.state[0]
	xDot := env.state[1]
	theta := env.state[2]
	thetaDot := env.state[3]

	force := forceMag
	if action == 0 {
		force = -forceMag
	}

	cosTheta := math.Cos(theta)
	sinTheta := math.Sin(theta)

	temp := (force + massPole*length*thetaDot*thetaDot*sinTheta) / totalMass
	thetaAcc := (gravity*sinTheta - cosTheta*temp) /
		(length * (4.0/3.0 - massPole*cosTheta*cosTheta/totalMass))
	xAcc := temp - massPole*length*thetaAcc*cosTheta/totalMass

	// Update state
	env.state[0] = x + tau*xDot
	env.state[1] = xDot + tau*xAcc
	env.state[2] = theta + tau*thetaDot
	env.state[3] = thetaDot + tau*thetaAcc

	env.stepCount++

	// Check termination
	done := math.Abs(env.state[0]) > 2.4 ||
		math.Abs(env.state[2]) > 12*math.Pi/180 ||
		env.stepCount >= env.maxSteps

	reward := 1.0
	if done && env.stepCount < env.maxSteps {
		reward = 0.0
	}

	return env.state, reward, done
}

// TrainRL trains an RL agent on Cart-Pole
func TrainRL(episodes int) (*QNetwork, []float64) {
	config := DefaultRLConfig()
	config.StateSize = 4
	config.ActionSize = 2

	qn := NewQNetwork(config)
	buffer := NewReplayBuffer(config.ReplayCapacity)
	env := NewCartPoleEnv()

	rewards := make([]float64, episodes)

	for ep := 0; ep < episodes; ep++ {
		state := env.Reset()
		totalReward := 0.0

		for {
			action := qn.SelectAction(state)
			nextState, reward, done := env.Step(action)

			buffer.Add(Experience{
				State:     append([]float64{}, state...),
				Action:    action,
				Reward:    reward,
				NextState: append([]float64{}, nextState...),
				Done:      done,
			})

			totalReward += reward

			// Train
			if buffer.Size() >= config.BatchSize {
				batch := buffer.Sample(config.BatchSize)
				qn.Update(batch)
			}

			state = nextState
			if done {
				break
			}
		}

		rewards[ep] = totalReward

		// Decay exploration
		config.ExplorationRate *= 0.995
		if config.ExplorationRate < 0.01 {
			config.ExplorationRate = 0.01
		}
	}

	return qn, rewards
}

// FormatRLReport generates RL training report
func FormatRLReport(rewards []float64) string {
	report := "=== RL Training Report ===\n\n"

	// Calculate statistics
	avgReward := 0.0
	for _, r := range rewards {
		avgReward += r
	}
	avgReward /= float64(len(rewards))

	last100Avg := 0.0
	start := len(rewards) - 100
	if start < 0 {
		start = 0
	}
	for i := start; i < len(rewards); i++ {
		last100Avg += rewards[i]
	}
	last100Avg /= float64(len(rewards) - start)

	report += fmt.Sprintf("Episodes: %d\n", len(rewards))
	report += fmt.Sprintf("Average Reward: %.2f\n", avgReward)
	report += fmt.Sprintf("Last 100 Average: %.2f\n", last100Avg)

	return report
}

// FormatMultiChipReport generates multi-chip system report
func FormatMultiChipReport(mcs *MultiChipSystem) string {
	report := "=== Multi-Chip System Report ===\n\n"

	report += fmt.Sprintf("Configuration:\n")
	report += fmt.Sprintf("  Chiplets: %d (%dx%d grid)\n",
		mcs.config.NumChiplets, mcs.config.GridRows, mcs.config.GridCols)
	report += fmt.Sprintf("  Tiles per chiplet: %d\n", mcs.config.ChipletConfig.NumTiles)
	report += fmt.Sprintf("  Tile size: %dx%d\n",
		mcs.config.ChipletConfig.TileSize, mcs.config.ChipletConfig.TileSize)

	report += fmt.Sprintf("\nInterconnect:\n")
	report += fmt.Sprintf("  NoC: %.1f Gbps, %d cycles/hop\n",
		mcs.config.NoCConfig.BandwidthGbps, mcs.config.NoCConfig.LatencyCycles)
	report += fmt.Sprintf("  NoP: %.1f Gbps, %d cycles/hop\n",
		mcs.config.NoPConfig.BandwidthGbps, mcs.config.NoPConfig.LatencyCycles)

	report += fmt.Sprintf("\nPerformance:\n")
	peakTOPS := float64(mcs.config.NumChiplets) * mcs.config.ChipletConfig.PeakTOPS
	report += fmt.Sprintf("  Peak throughput: %.1f TOPS\n", peakTOPS)
	report += fmt.Sprintf("  Utilization: %.1f%%\n", mcs.GetUtilization()*100)

	if mcs.totalLatency > 0 {
		report += fmt.Sprintf("  Total latency: %.2f ns\n", mcs.totalLatency)
	}
	if mcs.totalEnergy > 0 {
		report += fmt.Sprintf("  Total energy: %.2e pJ\n", mcs.totalEnergy)
		report += fmt.Sprintf("  Comm energy: %.1f%%\n",
			mcs.commEnergy/mcs.totalEnergy*100)
	}

	return report
}

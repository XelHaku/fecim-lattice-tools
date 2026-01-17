// Package layers provides hardware-aware NAS and reinforcement learning for CIM accelerators.
//
// This module implements neural architecture search optimized for memristor crossbar
// constraints and reinforcement learning algorithms suitable for analog hardware.
//
// Key features:
// - DARTS-style differentiable architecture search
// - Multi-objective optimization (accuracy, energy, latency)
// - CIM hardware constraint modeling
// - Actor-critic networks with memristor weights
// - Temporal difference learning on analog hardware
// - Q-learning with in-memory computation
//
// References:
// - "NAS for in-memory computing" (Nature Reviews EE 2024)
// - "Actor-critic with memristors" (Nature Machine Intelligence 2025)
// - "Multi-objective NAS for IMC" (arXiv 2406.06746)
// - "CMN: co-designed NAS for CIM" (Science China IS 2024)
package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// Hardware-Aware NAS Core
// =============================================================================

// NASConfig configures neural architecture search.
type NASConfig struct {
	SearchSpace      SearchSpace
	HardwareConfig   CIMHardwareConfig
	OptimizationGoal []ObjectiveWeight // multi-objective weights
	SearchMethod     string            // "darts", "evolution", "random"
	NumEpochs        int               // search epochs
	WarmupEpochs     int               // architecture warmup
	LearningRateArch float64           // architecture parameters LR
	LearningRateNet  float64           // network weights LR
	PopulationSize   int               // for evolutionary search
	NumCandidates    int               // candidates to evaluate
}

// DefaultNASConfig returns standard NAS settings.
func DefaultNASConfig() NASConfig {
	return NASConfig{
		SearchSpace:      DefaultSearchSpace(),
		HardwareConfig:   DefaultCIMHardwareConfig(),
		OptimizationGoal: []ObjectiveWeight{{Name: "accuracy", Weight: 0.6}, {Name: "energy", Weight: 0.3}, {Name: "latency", Weight: 0.1}},
		SearchMethod:     "darts",
		NumEpochs:        50,
		WarmupEpochs:     10,
		LearningRateArch: 3e-4,
		LearningRateNet:  0.025,
		PopulationSize:   50,
		NumCandidates:    100,
	}
}

// ObjectiveWeight defines multi-objective optimization weight.
type ObjectiveWeight struct {
	Name   string
	Weight float64
}

// SearchSpace defines the NAS search space.
type SearchSpace struct {
	Operations      []OperationType // available operations
	NumCells        int             // number of cells to search
	NumNodesPerCell int             // nodes per cell
	MaxChannels     int             // maximum channels
	MinChannels     int             // minimum channels
	ChannelMult     []int           // channel multipliers
	Reductions      []int           // reduction cell positions
}

// DefaultSearchSpace returns standard DARTS-like search space.
func DefaultSearchSpace() SearchSpace {
	return SearchSpace{
		Operations: []OperationType{
			OP_CONV_3X3,
			OP_CONV_5X5,
			OP_SEP_CONV_3X3,
			OP_SEP_CONV_5X5,
			OP_DIL_CONV_3X3,
			OP_MAX_POOL_3X3,
			OP_AVG_POOL_3X3,
			OP_SKIP_CONNECT,
			OP_NONE,
		},
		NumCells:        8,
		NumNodesPerCell: 4,
		MaxChannels:     512,
		MinChannels:     16,
		ChannelMult:     []int{1, 2, 4},
		Reductions:      []int{2, 5},
	}
}

// OperationType defines operation types in search space.
type OperationType int

const (
	OP_NONE OperationType = iota
	OP_SKIP_CONNECT
	OP_MAX_POOL_3X3
	OP_AVG_POOL_3X3
	OP_CONV_3X3
	OP_CONV_5X5
	OP_SEP_CONV_3X3
	OP_SEP_CONV_5X5
	OP_DIL_CONV_3X3
	OP_DIL_CONV_5X5
)

// String returns operation name.
func (op OperationType) String() string {
	names := []string{
		"none", "skip_connect", "max_pool_3x3", "avg_pool_3x3",
		"conv_3x3", "conv_5x5", "sep_conv_3x3", "sep_conv_5x5",
		"dil_conv_3x3", "dil_conv_5x5",
	}
	if int(op) < len(names) {
		return names[op]
	}
	return "unknown"
}

// CIMHardwareConfig defines hardware constraints for NAS.
type CIMHardwareConfig struct {
	CrossbarSize     int     // max crossbar dimensions
	ADCBits          int     // ADC precision
	DACBits          int     // DAC precision
	WeightBits       int     // weight precision
	MaxPower         float64 // power budget (mW)
	MaxLatency       float64 // latency budget (ms)
	ArrayEfficiency  float64 // utilization efficiency
	WireResistance   float64 // IR drop factor
	EnergyPerMAC     float64 // fJ per MAC
	OnChipMemory     int     // KB of on-chip memory
}

// DefaultCIMHardwareConfig returns standard CIM hardware settings.
func DefaultCIMHardwareConfig() CIMHardwareConfig {
	return CIMHardwareConfig{
		CrossbarSize:    256,
		ADCBits:         6,
		DACBits:         8,
		WeightBits:      6,
		MaxPower:        100.0,
		MaxLatency:      10.0,
		ArrayEfficiency: 0.85,
		WireResistance:  0.1,
		EnergyPerMAC:    50.0, // fJ
		OnChipMemory:    512,
	}
}

// =============================================================================
// DARTS-Style Differentiable NAS
// =============================================================================

// DARTSSearcher implements differentiable architecture search.
type DARTSSearcher struct {
	Config          NASConfig
	AlphaParams     [][][]float64 // architecture parameters per cell/edge/op
	NetworkWeights  map[string][]float64
	BestArch        *Architecture
	BestScore       float64
	SearchHistory   []SearchState
	RNG             *rand.Rand
}

// SearchState tracks NAS progress.
type SearchState struct {
	Epoch      int
	TrainLoss  float64
	ValLoss    float64
	Accuracy   float64
	Energy     float64
	Latency    float64
	Score      float64
	ArchGenome []int
}

// NewDARTSSearcher creates a DARTS-style NAS searcher.
func NewDARTSSearcher(config NASConfig, seed int64) *DARTSSearcher {
	searcher := &DARTSSearcher{
		Config:         config,
		NetworkWeights: make(map[string][]float64),
		RNG:            rand.New(rand.NewSource(seed)),
	}

	// Initialize architecture parameters
	numOps := len(config.SearchSpace.Operations)
	numEdges := config.SearchSpace.NumNodesPerCell * (config.SearchSpace.NumNodesPerCell + 1) / 2
	searcher.AlphaParams = make([][][]float64, config.SearchSpace.NumCells)

	for c := 0; c < config.SearchSpace.NumCells; c++ {
		searcher.AlphaParams[c] = make([][]float64, numEdges)
		for e := 0; e < numEdges; e++ {
			searcher.AlphaParams[c][e] = make([]float64, numOps)
			// Initialize uniformly
			for o := 0; o < numOps; o++ {
				searcher.AlphaParams[c][e][o] = 1.0 / float64(numOps)
			}
		}
	}

	return searcher
}

// Softmax computes softmax over architecture weights.
func Softmax(logits []float64) []float64 {
	maxVal := logits[0]
	for _, v := range logits[1:] {
		if v > maxVal {
			maxVal = v
		}
	}

	expSum := 0.0
	result := make([]float64, len(logits))
	for i, v := range logits {
		result[i] = math.Exp(v - maxVal)
		expSum += result[i]
	}
	for i := range result {
		result[i] /= expSum
	}
	return result
}

// GetMixedOp computes mixed operation output using softmax weights.
func (d *DARTSSearcher) GetMixedOp(cell, edge int) []float64 {
	return Softmax(d.AlphaParams[cell][edge])
}

// DecodeArchitecture extracts discrete architecture from alpha parameters.
func (d *DARTSSearcher) DecodeArchitecture() *Architecture {
	arch := &Architecture{
		Cells: make([]Cell, d.Config.SearchSpace.NumCells),
	}

	for c := range d.AlphaParams {
		arch.Cells[c] = Cell{
			Edges: make([]Edge, len(d.AlphaParams[c])),
		}
		for e := range d.AlphaParams[c] {
			// Select operation with highest weight
			weights := Softmax(d.AlphaParams[c][e])
			maxIdx := 0
			maxWeight := weights[0]
			for i, w := range weights[1:] {
				if w > maxWeight {
					maxWeight = w
					maxIdx = i + 1
				}
			}
			arch.Cells[c].Edges[e] = Edge{
				Operation: d.Config.SearchSpace.Operations[maxIdx],
				Weight:    maxWeight,
			}
		}
	}

	return arch
}

// UpdateArchitectureParams updates alpha using validation gradient.
func (d *DARTSSearcher) UpdateArchitectureParams(valGradient [][][]float64, lr float64) {
	for c := range d.AlphaParams {
		for e := range d.AlphaParams[c] {
			for o := range d.AlphaParams[c][e] {
				d.AlphaParams[c][e][o] -= lr * valGradient[c][e][o]
			}
		}
	}
}

// Architecture represents a decoded neural architecture.
type Architecture struct {
	Cells       []Cell
	InputShape  []int
	OutputShape []int
	NumParams   int
	FLOPs       int64
	Energy      float64
	Latency     float64
}

// Cell represents a cell in the architecture.
type Cell struct {
	Edges      []Edge
	IsReduction bool
	Channels   int
}

// Edge represents an edge/connection in a cell.
type Edge struct {
	Operation OperationType
	Weight    float64
	FromNode  int
	ToNode    int
}

// =============================================================================
// Hardware Cost Model
// =============================================================================

// HardwareCostModel estimates CIM hardware costs.
type HardwareCostModel struct {
	Config CIMHardwareConfig
}

// NewHardwareCostModel creates a hardware cost model.
func NewHardwareCostModel(config CIMHardwareConfig) *HardwareCostModel {
	return &HardwareCostModel{Config: config}
}

// EstimateEnergy estimates energy consumption for an architecture.
func (h *HardwareCostModel) EstimateEnergy(arch *Architecture) float64 {
	var totalEnergy float64

	for _, cell := range arch.Cells {
		for _, edge := range cell.Edges {
			// Energy per operation type
			opEnergy := h.getOperationEnergy(edge.Operation, cell.Channels)
			totalEnergy += opEnergy
		}
	}

	return totalEnergy
}

// EstimateLatency estimates inference latency.
func (h *HardwareCostModel) EstimateLatency(arch *Architecture) float64 {
	var totalLatency float64

	for _, cell := range arch.Cells {
		// Critical path latency
		maxEdgeLatency := 0.0
		for _, edge := range cell.Edges {
			opLatency := h.getOperationLatency(edge.Operation, cell.Channels)
			if opLatency > maxEdgeLatency {
				maxEdgeLatency = opLatency
			}
		}
		totalLatency += maxEdgeLatency
	}

	return totalLatency
}

// EstimateMemory estimates memory requirements.
func (h *HardwareCostModel) EstimateMemory(arch *Architecture) int {
	var totalMem int

	for _, cell := range arch.Cells {
		for _, edge := range cell.Edges {
			opMem := h.getOperationMemory(edge.Operation, cell.Channels)
			totalMem += opMem
		}
	}

	return totalMem
}

// getOperationEnergy returns energy for operation type.
func (h *HardwareCostModel) getOperationEnergy(op OperationType, channels int) float64 {
	baseEnergy := h.Config.EnergyPerMAC * float64(channels*channels)

	switch op {
	case OP_CONV_3X3:
		return baseEnergy * 9
	case OP_CONV_5X5:
		return baseEnergy * 25
	case OP_SEP_CONV_3X3:
		return baseEnergy * 9 * 0.3 // depthwise efficiency
	case OP_SEP_CONV_5X5:
		return baseEnergy * 25 * 0.3
	case OP_DIL_CONV_3X3:
		return baseEnergy * 9
	case OP_MAX_POOL_3X3, OP_AVG_POOL_3X3:
		return baseEnergy * 0.1 // pooling is cheap
	case OP_SKIP_CONNECT:
		return 0.01 // minimal energy
	case OP_NONE:
		return 0
	}
	return baseEnergy
}

// getOperationLatency returns latency for operation type.
func (h *HardwareCostModel) getOperationLatency(op OperationType, channels int) float64 {
	baseLatency := float64(channels) / float64(h.Config.CrossbarSize) * 0.01 // ms

	switch op {
	case OP_CONV_3X3:
		return baseLatency * 9
	case OP_CONV_5X5:
		return baseLatency * 25
	case OP_SEP_CONV_3X3:
		return baseLatency * 9 * 0.3
	case OP_SEP_CONV_5X5:
		return baseLatency * 25 * 0.3
	case OP_DIL_CONV_3X3:
		return baseLatency * 9
	case OP_MAX_POOL_3X3, OP_AVG_POOL_3X3:
		return baseLatency * 0.1
	case OP_SKIP_CONNECT:
		return 0.001
	case OP_NONE:
		return 0
	}
	return baseLatency
}

// getOperationMemory returns memory for operation type.
func (h *HardwareCostModel) getOperationMemory(op OperationType, channels int) int {
	baseBytes := channels * channels * h.Config.WeightBits / 8

	switch op {
	case OP_CONV_3X3:
		return baseBytes * 9
	case OP_CONV_5X5:
		return baseBytes * 25
	case OP_SEP_CONV_3X3:
		return (channels*9 + channels*channels) * h.Config.WeightBits / 8
	case OP_SEP_CONV_5X5:
		return (channels*25 + channels*channels) * h.Config.WeightBits / 8
	case OP_DIL_CONV_3X3:
		return baseBytes * 9
	case OP_MAX_POOL_3X3, OP_AVG_POOL_3X3, OP_SKIP_CONNECT, OP_NONE:
		return 0
	}
	return baseBytes
}

// CheckConstraints validates architecture against hardware constraints.
func (h *HardwareCostModel) CheckConstraints(arch *Architecture) []ConstraintViolation {
	var violations []ConstraintViolation

	energy := h.EstimateEnergy(arch)
	if energy > h.Config.MaxPower*1000 { // convert to fJ
		violations = append(violations, ConstraintViolation{
			Type:     "energy",
			Value:    energy,
			Limit:    h.Config.MaxPower * 1000,
			Severity: energy / (h.Config.MaxPower * 1000),
		})
	}

	latency := h.EstimateLatency(arch)
	if latency > h.Config.MaxLatency {
		violations = append(violations, ConstraintViolation{
			Type:     "latency",
			Value:    latency,
			Limit:    h.Config.MaxLatency,
			Severity: latency / h.Config.MaxLatency,
		})
	}

	memory := h.EstimateMemory(arch)
	if memory > h.Config.OnChipMemory*1024 {
		violations = append(violations, ConstraintViolation{
			Type:     "memory",
			Value:    float64(memory),
			Limit:    float64(h.Config.OnChipMemory * 1024),
			Severity: float64(memory) / float64(h.Config.OnChipMemory*1024),
		})
	}

	return violations
}

// ConstraintViolation represents a hardware constraint violation.
type ConstraintViolation struct {
	Type     string
	Value    float64
	Limit    float64
	Severity float64
}

// =============================================================================
// Multi-Objective Optimization
// =============================================================================

// MOOEvaluator implements multi-objective optimization for NAS.
type MOOEvaluator struct {
	Objectives  []ObjectiveWeight
	CostModel   *HardwareCostModel
	ParetoFront []*Architecture
}

// NewMOOEvaluator creates a multi-objective evaluator.
func NewMOOEvaluator(objectives []ObjectiveWeight, costModel *HardwareCostModel) *MOOEvaluator {
	return &MOOEvaluator{
		Objectives:  objectives,
		CostModel:   costModel,
		ParetoFront: make([]*Architecture, 0),
	}
}

// ComputeScore computes weighted multi-objective score.
func (m *MOOEvaluator) ComputeScore(arch *Architecture, accuracy float64) float64 {
	energy := m.CostModel.EstimateEnergy(arch)
	latency := m.CostModel.EstimateLatency(arch)

	var score float64
	for _, obj := range m.Objectives {
		switch obj.Name {
		case "accuracy":
			score += obj.Weight * accuracy
		case "energy":
			// Normalize energy (lower is better)
			score += obj.Weight * (1 - math.Min(energy/(m.CostModel.Config.MaxPower*1000), 1))
		case "latency":
			// Normalize latency (lower is better)
			score += obj.Weight * (1 - math.Min(latency/m.CostModel.Config.MaxLatency, 1))
		}
	}

	// Penalty for constraint violations
	violations := m.CostModel.CheckConstraints(arch)
	for _, v := range violations {
		score -= 0.5 * v.Severity
	}

	return score
}

// UpdateParetoFront updates Pareto front with new architecture.
func (m *MOOEvaluator) UpdateParetoFront(arch *Architecture, accuracy float64) bool {
	// Compute objectives
	newObjs := map[string]float64{
		"accuracy": accuracy,
		"energy":   m.CostModel.EstimateEnergy(arch),
		"latency":  m.CostModel.EstimateLatency(arch),
	}

	// Check if dominated by existing solutions
	for _, existing := range m.ParetoFront {
		existingObjs := map[string]float64{
			"accuracy": 0.9, // placeholder
			"energy":   m.CostModel.EstimateEnergy(existing),
			"latency":  m.CostModel.EstimateLatency(existing),
		}

		if dominates(existingObjs, newObjs) {
			return false
		}
	}

	// Remove dominated solutions
	newFront := make([]*Architecture, 0)
	for _, existing := range m.ParetoFront {
		existingObjs := map[string]float64{
			"accuracy": 0.9,
			"energy":   m.CostModel.EstimateEnergy(existing),
			"latency":  m.CostModel.EstimateLatency(existing),
		}
		if !dominates(newObjs, existingObjs) {
			newFront = append(newFront, existing)
		}
	}

	newFront = append(newFront, arch)
	m.ParetoFront = newFront
	return true
}

// dominates checks if obj1 dominates obj2 (Pareto dominance).
func dominates(obj1, obj2 map[string]float64) bool {
	betterInAll := true
	strictlyBetterInOne := false

	for key, v1 := range obj1 {
		v2 := obj2[key]
		if key == "accuracy" {
			// Higher accuracy is better
			if v1 < v2 {
				betterInAll = false
			}
			if v1 > v2 {
				strictlyBetterInOne = true
			}
		} else {
			// Lower energy/latency is better
			if v1 > v2 {
				betterInAll = false
			}
			if v1 < v2 {
				strictlyBetterInOne = true
			}
		}
	}

	return betterInAll && strictlyBetterInOne
}

// =============================================================================
// Evolutionary NAS
// =============================================================================

// EvolutionaryNAS implements evolutionary architecture search.
type EvolutionaryNAS struct {
	Config      NASConfig
	Population  []*Individual
	Generation  int
	BestFitness float64
	BestGenome  []int
	RNG         *rand.Rand
}

// Individual represents an architecture in the population.
type Individual struct {
	Genome     []int   // encoded architecture
	Fitness    float64
	Accuracy   float64
	Energy     float64
	Latency    float64
	Age        int
}

// NewEvolutionaryNAS creates an evolutionary NAS searcher.
func NewEvolutionaryNAS(config NASConfig, seed int64) *EvolutionaryNAS {
	enas := &EvolutionaryNAS{
		Config:     config,
		Population: make([]*Individual, config.PopulationSize),
		RNG:        rand.New(rand.NewSource(seed)),
	}

	// Initialize random population
	genomeSize := config.SearchSpace.NumCells *
		config.SearchSpace.NumNodesPerCell * (config.SearchSpace.NumNodesPerCell + 1) / 2

	for i := 0; i < config.PopulationSize; i++ {
		genome := make([]int, genomeSize)
		for j := range genome {
			genome[j] = enas.RNG.Intn(len(config.SearchSpace.Operations))
		}
		enas.Population[i] = &Individual{Genome: genome}
	}

	return enas
}

// TournamentSelect performs tournament selection.
func (e *EvolutionaryNAS) TournamentSelect(tournamentSize int) *Individual {
	best := e.Population[e.RNG.Intn(len(e.Population))]
	for i := 1; i < tournamentSize; i++ {
		candidate := e.Population[e.RNG.Intn(len(e.Population))]
		if candidate.Fitness > best.Fitness {
			best = candidate
		}
	}
	return best
}

// Crossover performs crossover between two individuals.
func (e *EvolutionaryNAS) Crossover(parent1, parent2 *Individual) *Individual {
	child := &Individual{
		Genome: make([]int, len(parent1.Genome)),
	}

	// Uniform crossover
	for i := range child.Genome {
		if e.RNG.Float64() < 0.5 {
			child.Genome[i] = parent1.Genome[i]
		} else {
			child.Genome[i] = parent2.Genome[i]
		}
	}

	return child
}

// Mutate performs mutation on an individual.
func (e *EvolutionaryNAS) Mutate(ind *Individual, mutationRate float64) {
	numOps := len(e.Config.SearchSpace.Operations)
	for i := range ind.Genome {
		if e.RNG.Float64() < mutationRate {
			ind.Genome[i] = e.RNG.Intn(numOps)
		}
	}
}

// EvolveGeneration performs one generation of evolution.
func (e *EvolutionaryNAS) EvolveGeneration(evaluate func(*Individual)) {
	// Evaluate current population
	for _, ind := range e.Population {
		if ind.Fitness == 0 {
			evaluate(ind)
		}
	}

	// Sort by fitness
	sort.Slice(e.Population, func(i, j int) bool {
		return e.Population[i].Fitness > e.Population[j].Fitness
	})

	// Track best
	if e.Population[0].Fitness > e.BestFitness {
		e.BestFitness = e.Population[0].Fitness
		e.BestGenome = make([]int, len(e.Population[0].Genome))
		copy(e.BestGenome, e.Population[0].Genome)
	}

	// Generate new population
	newPop := make([]*Individual, e.Config.PopulationSize)

	// Elitism: keep top 10%
	numElite := e.Config.PopulationSize / 10
	for i := 0; i < numElite; i++ {
		newPop[i] = e.Population[i]
		newPop[i].Age++
	}

	// Generate rest through selection, crossover, mutation
	for i := numElite; i < e.Config.PopulationSize; i++ {
		parent1 := e.TournamentSelect(5)
		parent2 := e.TournamentSelect(5)
		child := e.Crossover(parent1, parent2)
		e.Mutate(child, 0.1)
		newPop[i] = child
	}

	e.Population = newPop
	e.Generation++
}

// =============================================================================
// Reinforcement Learning Core
// =============================================================================

// RLConfig configures reinforcement learning.
type RLConfig struct {
	Algorithm      string  // "actor_critic", "dqn", "ppo"
	GammaDiscount  float64 // discount factor
	LearningRate   float64
	TauTarget      float64 // target network update rate
	ReplayBuffer   int     // replay buffer size
	BatchSize      int
	NumHiddenUnits int
	EpsilonStart   float64 // exploration rate
	EpsilonEnd     float64
	EpsilonDecay   float64
	MaxEpisodes    int
	MaxSteps       int
}

// DefaultRLConfig returns standard RL settings.
func DefaultRLConfig() RLConfig {
	return RLConfig{
		Algorithm:      "actor_critic",
		GammaDiscount:  0.99,
		LearningRate:   0.001,
		TauTarget:      0.005,
		ReplayBuffer:   10000,
		BatchSize:      64,
		NumHiddenUnits: 128,
		EpsilonStart:   1.0,
		EpsilonEnd:     0.01,
		EpsilonDecay:   0.995,
		MaxEpisodes:    1000,
		MaxSteps:       200,
	}
}

// =============================================================================
// Actor-Critic Networks with Memristors
// =============================================================================

// MemristorActorCritic implements actor-critic with analog memristor weights.
type MemristorActorCritic struct {
	Config        RLConfig
	ActorWeights  *MemristorCrossbar
	CriticWeights *MemristorCrossbar
	ActorHidden   *MemristorCrossbar
	CriticHidden  *MemristorCrossbar
	StateSize     int
	ActionSize    int
	Episode       int
	TotalReward   float64
	EpisodeRewards []float64
	RNG           *rand.Rand
}

// MemristorCrossbar simulates memristor weight array for RL.
type MemristorCrossbar struct {
	Weights       [][]float64
	Conductance   [][]float64 // physical conductance
	GMin          float64
	GMax          float64
	D2DVariation  float64
	C2CVariation  float64
	LearningRate  float64
	EligibilityTrace [][]float64 // for TD learning
}

// NewMemristorCrossbar creates a memristor crossbar for RL.
func NewMemristorCrossbar(rows, cols int, config RLConfig, seed int64) *MemristorCrossbar {
	rng := rand.New(rand.NewSource(seed))
	cb := &MemristorCrossbar{
		Weights:          make([][]float64, rows),
		Conductance:      make([][]float64, rows),
		EligibilityTrace: make([][]float64, rows),
		GMin:             1e-7,
		GMax:             1e-4,
		D2DVariation:     0.05,
		C2CVariation:     0.02,
		LearningRate:     config.LearningRate,
	}

	for i := 0; i < rows; i++ {
		cb.Weights[i] = make([]float64, cols)
		cb.Conductance[i] = make([]float64, cols)
		cb.EligibilityTrace[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			// Initialize with Xavier/Glorot
			cb.Weights[i][j] = rng.NormFloat64() * math.Sqrt(2.0/float64(rows+cols))
			// Map to conductance
			cb.Conductance[i][j] = cb.weightToConductance(cb.Weights[i][j])
		}
	}

	return cb
}

// weightToConductance maps weight to physical conductance.
func (m *MemristorCrossbar) weightToConductance(w float64) float64 {
	// Sigmoid mapping to [GMin, GMax]
	sigmoid := 1.0 / (1.0 + math.Exp(-w))
	return m.GMin + sigmoid*(m.GMax-m.GMin)
}

// conductanceToWeight maps conductance back to weight.
func (m *MemristorCrossbar) conductanceToWeight(g float64) float64 {
	normalized := (g - m.GMin) / (m.GMax - m.GMin)
	if normalized <= 0 {
		return -10
	}
	if normalized >= 1 {
		return 10
	}
	return -math.Log(1/normalized - 1)
}

// Forward performs matrix-vector multiplication with noise.
func (m *MemristorCrossbar) Forward(input []float64) []float64 {
	output := make([]float64, len(m.Weights[0]))

	for j := range output {
		for i, in := range input {
			// Apply D2D and C2C variation
			g := m.Conductance[i][j]
			g *= (1 + rand.NormFloat64()*m.D2DVariation) // D2D
			g *= (1 + rand.NormFloat64()*m.C2CVariation) // C2C
			w := m.conductanceToWeight(g)
			output[j] += in * w
		}
	}

	return output
}

// UpdateWithTD performs temporal difference update on weights.
func (m *MemristorCrossbar) UpdateWithTD(tdError float64, input []float64, decay float64) {
	// Update eligibility traces
	for i := range m.EligibilityTrace {
		for j := range m.EligibilityTrace[i] {
			// Decay trace
			m.EligibilityTrace[i][j] *= decay
			// Add current contribution
			if i < len(input) {
				m.EligibilityTrace[i][j] += input[i]
			}
		}
	}

	// Update weights based on TD error and traces
	for i := range m.Weights {
		for j := range m.Weights[i] {
			deltaW := m.LearningRate * tdError * m.EligibilityTrace[i][j]
			m.Weights[i][j] += deltaW
			// Update conductance
			m.Conductance[i][j] = m.weightToConductance(m.Weights[i][j])
		}
	}
}

// NewMemristorActorCritic creates an actor-critic agent with memristor weights.
func NewMemristorActorCritic(stateSize, actionSize int, config RLConfig, seed int64) *MemristorActorCritic {
	hiddenSize := config.NumHiddenUnits

	return &MemristorActorCritic{
		Config:         config,
		ActorHidden:   NewMemristorCrossbar(stateSize, hiddenSize, config, seed),
		ActorWeights:  NewMemristorCrossbar(hiddenSize, actionSize, config, seed+1),
		CriticHidden:  NewMemristorCrossbar(stateSize, hiddenSize, config, seed+2),
		CriticWeights: NewMemristorCrossbar(hiddenSize, 1, config, seed+3),
		StateSize:     stateSize,
		ActionSize:    actionSize,
		RNG:           rand.New(rand.NewSource(seed + 4)),
	}
}

// ReLU activation function.
func ReLU(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

// SoftmaxAction computes action probabilities.
func SoftmaxAction(logits []float64) ([]float64, int) {
	probs := Softmax(logits)

	// Sample action
	r := rand.Float64()
	cumProb := 0.0
	action := 0
	for i, p := range probs {
		cumProb += p
		if r <= cumProb {
			action = i
			break
		}
	}

	return probs, action
}

// SelectAction chooses action using actor network.
func (ac *MemristorActorCritic) SelectAction(state []float64) (int, []float64) {
	// Forward through actor
	hidden := ac.ActorHidden.Forward(state)
	for i := range hidden {
		hidden[i] = ReLU(hidden[i])
	}
	logits := ac.ActorWeights.Forward(hidden)

	probs, action := SoftmaxAction(logits)
	return action, probs
}

// EstimateValue computes value estimate using critic network.
func (ac *MemristorActorCritic) EstimateValue(state []float64) float64 {
	hidden := ac.CriticHidden.Forward(state)
	for i := range hidden {
		hidden[i] = ReLU(hidden[i])
	}
	value := ac.CriticWeights.Forward(hidden)
	return value[0]
}

// Update performs actor-critic update using TD error.
func (ac *MemristorActorCritic) Update(
	state []float64,
	action int,
	reward float64,
	nextState []float64,
	done bool,
) float64 {
	gamma := ac.Config.GammaDiscount

	// Compute TD error
	currentValue := ac.EstimateValue(state)
	var nextValue float64
	if !done {
		nextValue = ac.EstimateValue(nextState)
	}
	tdError := reward + gamma*nextValue - currentValue

	// Update critic
	ac.CriticHidden.UpdateWithTD(tdError, state, gamma)
	hiddenCritic := ac.CriticHidden.Forward(state)
	for i := range hiddenCritic {
		hiddenCritic[i] = ReLU(hiddenCritic[i])
	}
	ac.CriticWeights.UpdateWithTD(tdError, hiddenCritic, gamma)

	// Update actor (policy gradient with TD error as advantage)
	// Simplified: update in direction of chosen action
	actorGrad := make([]float64, ac.ActionSize)
	_, probs := ac.SelectAction(state)
	for i := range actorGrad {
		if i == action {
			actorGrad[i] = (1 - probs[i]) * tdError
		} else {
			actorGrad[i] = -probs[i] * tdError
		}
	}

	hiddenActor := ac.ActorHidden.Forward(state)
	for i := range hiddenActor {
		hiddenActor[i] = ReLU(hiddenActor[i])
	}
	ac.ActorHidden.UpdateWithTD(tdError, state, gamma)

	return tdError
}

// =============================================================================
// Q-Learning with Memristors
// =============================================================================

// MemristorDQN implements Deep Q-Network with memristor weights.
type MemristorDQN struct {
	Config          RLConfig
	QNetwork        *MemristorCrossbar
	TargetNetwork   *MemristorCrossbar
	Hidden          *MemristorCrossbar
	TargetHidden    *MemristorCrossbar
	ReplayBuffer    []Experience
	StateSize       int
	ActionSize      int
	Epsilon         float64
	UpdateCounter   int
	RNG             *rand.Rand
}

// Experience represents a single transition.
type Experience struct {
	State     []float64
	Action    int
	Reward    float64
	NextState []float64
	Done      bool
}

// NewMemristorDQN creates a DQN agent with memristor weights.
func NewMemristorDQN(stateSize, actionSize int, config RLConfig, seed int64) *MemristorDQN {
	hiddenSize := config.NumHiddenUnits

	dqn := &MemristorDQN{
		Config:        config,
		Hidden:        NewMemristorCrossbar(stateSize, hiddenSize, config, seed),
		QNetwork:      NewMemristorCrossbar(hiddenSize, actionSize, config, seed+1),
		TargetHidden:  NewMemristorCrossbar(stateSize, hiddenSize, config, seed+2),
		TargetNetwork: NewMemristorCrossbar(hiddenSize, actionSize, config, seed+3),
		ReplayBuffer:  make([]Experience, 0, config.ReplayBuffer),
		StateSize:     stateSize,
		ActionSize:    actionSize,
		Epsilon:       config.EpsilonStart,
		RNG:           rand.New(rand.NewSource(seed + 4)),
	}

	// Copy weights to target
	dqn.syncTargetNetwork()

	return dqn
}

// syncTargetNetwork copies weights to target network.
func (d *MemristorDQN) syncTargetNetwork() {
	for i := range d.QNetwork.Weights {
		copy(d.TargetNetwork.Weights[i], d.QNetwork.Weights[i])
		copy(d.TargetNetwork.Conductance[i], d.QNetwork.Conductance[i])
	}
	for i := range d.Hidden.Weights {
		copy(d.TargetHidden.Weights[i], d.Hidden.Weights[i])
		copy(d.TargetHidden.Conductance[i], d.Hidden.Conductance[i])
	}
}

// GetQValues computes Q-values for all actions.
func (d *MemristorDQN) GetQValues(state []float64) []float64 {
	hidden := d.Hidden.Forward(state)
	for i := range hidden {
		hidden[i] = ReLU(hidden[i])
	}
	return d.QNetwork.Forward(hidden)
}

// SelectAction chooses action with epsilon-greedy policy.
func (d *MemristorDQN) SelectAction(state []float64) int {
	if d.RNG.Float64() < d.Epsilon {
		return d.RNG.Intn(d.ActionSize)
	}

	qValues := d.GetQValues(state)
	maxAction := 0
	maxQ := qValues[0]
	for i, q := range qValues[1:] {
		if q > maxQ {
			maxQ = q
			maxAction = i + 1
		}
	}
	return maxAction
}

// StoreExperience adds experience to replay buffer.
func (d *MemristorDQN) StoreExperience(exp Experience) {
	if len(d.ReplayBuffer) >= d.Config.ReplayBuffer {
		// Remove oldest
		d.ReplayBuffer = d.ReplayBuffer[1:]
	}
	d.ReplayBuffer = append(d.ReplayBuffer, exp)
}

// SampleBatch samples random batch from replay buffer.
func (d *MemristorDQN) SampleBatch() []Experience {
	if len(d.ReplayBuffer) < d.Config.BatchSize {
		return nil
	}

	batch := make([]Experience, d.Config.BatchSize)
	perm := d.RNG.Perm(len(d.ReplayBuffer))
	for i := 0; i < d.Config.BatchSize; i++ {
		batch[i] = d.ReplayBuffer[perm[i]]
	}
	return batch
}

// Train performs one training step on batch.
func (d *MemristorDQN) Train() float64 {
	batch := d.SampleBatch()
	if batch == nil {
		return 0
	}

	var totalLoss float64
	gamma := d.Config.GammaDiscount

	for _, exp := range batch {
		// Current Q value
		qValues := d.GetQValues(exp.State)
		currentQ := qValues[exp.Action]

		// Target Q value
		var targetQ float64
		if exp.Done {
			targetQ = exp.Reward
		} else {
			// Use target network
			targetHidden := d.TargetHidden.Forward(exp.NextState)
			for i := range targetHidden {
				targetHidden[i] = ReLU(targetHidden[i])
			}
			nextQValues := d.TargetNetwork.Forward(targetHidden)
			maxNextQ := nextQValues[0]
			for _, q := range nextQValues[1:] {
				if q > maxNextQ {
					maxNextQ = q
				}
			}
			targetQ = exp.Reward + gamma*maxNextQ
		}

		// TD error
		tdError := targetQ - currentQ
		totalLoss += tdError * tdError

		// Update with TD error
		d.Hidden.UpdateWithTD(tdError, exp.State, gamma)
		hidden := d.Hidden.Forward(exp.State)
		for i := range hidden {
			hidden[i] = ReLU(hidden[i])
		}
		d.QNetwork.UpdateWithTD(tdError, hidden, gamma)
	}

	// Soft update target network
	d.UpdateCounter++
	if d.UpdateCounter%100 == 0 {
		d.softUpdateTarget()
	}

	// Decay epsilon
	d.Epsilon = math.Max(d.Config.EpsilonEnd,
		d.Epsilon*d.Config.EpsilonDecay)

	return totalLoss / float64(len(batch))
}

// softUpdateTarget performs soft update of target network.
func (d *MemristorDQN) softUpdateTarget() {
	tau := d.Config.TauTarget

	for i := range d.QNetwork.Weights {
		for j := range d.QNetwork.Weights[i] {
			d.TargetNetwork.Weights[i][j] = tau*d.QNetwork.Weights[i][j] +
				(1-tau)*d.TargetNetwork.Weights[i][j]
		}
	}
}

// =============================================================================
// Environment Interface
// =============================================================================

// Environment defines the RL environment interface.
type Environment interface {
	Reset() []float64
	Step(action int) ([]float64, float64, bool)
	GetStateSize() int
	GetActionSize() int
}

// TMazeEnvironment implements T-maze navigation task.
type TMazeEnvironment struct {
	Position     int
	Goal         int // 0=left, 1=right
	MaxSteps     int
	CurrentStep  int
	CorridorLen  int
	CueBit       int // cue presented at start
}

// NewTMazeEnvironment creates a T-maze environment.
func NewTMazeEnvironment(corridorLen int) *TMazeEnvironment {
	return &TMazeEnvironment{
		CorridorLen: corridorLen,
		MaxSteps:    corridorLen + 5,
	}
}

// Reset initializes a new episode.
func (t *TMazeEnvironment) Reset() []float64 {
	t.Position = 0
	t.CurrentStep = 0
	t.Goal = rand.Intn(2)
	t.CueBit = t.Goal

	return t.getState()
}

// Step takes an action and returns next state, reward, done.
func (t *TMazeEnvironment) Step(action int) ([]float64, float64, bool) {
	t.CurrentStep++

	// Actions: 0=left, 1=forward, 2=right
	switch action {
	case 1: // forward
		if t.Position < t.CorridorLen {
			t.Position++
		}
	case 0: // left at junction
		if t.Position == t.CorridorLen {
			if t.Goal == 0 {
				return t.getState(), 1.0, true // correct
			}
			return t.getState(), -1.0, true // wrong
		}
	case 2: // right at junction
		if t.Position == t.CorridorLen {
			if t.Goal == 1 {
				return t.getState(), 1.0, true // correct
			}
			return t.getState(), -1.0, true // wrong
		}
	}

	done := t.CurrentStep >= t.MaxSteps
	return t.getState(), 0, done
}

// getState returns current state encoding.
func (t *TMazeEnvironment) getState() []float64 {
	state := make([]float64, t.CorridorLen+3)
	// Position one-hot
	if t.Position < t.CorridorLen {
		state[t.Position] = 1.0
	} else {
		state[t.CorridorLen] = 1.0 // at junction
	}
	// Cue bit (only visible at start)
	if t.Position == 0 {
		state[t.CorridorLen+1+t.CueBit] = 1.0
	}
	return state
}

// GetStateSize returns state dimension.
func (t *TMazeEnvironment) GetStateSize() int {
	return t.CorridorLen + 3
}

// GetActionSize returns action count.
func (t *TMazeEnvironment) GetActionSize() int {
	return 3
}

// =============================================================================
// Serialization
// =============================================================================

// NASState captures NAS state for persistence.
type NASState struct {
	AlphaParams   [][][]float64   `json:"alpha_params"`
	BestArchGenome []int          `json:"best_arch_genome"`
	BestScore     float64         `json:"best_score"`
	Generation    int             `json:"generation"`
	SearchHistory []SearchState   `json:"search_history"`
}

// ExportNASState exports NAS state.
func (d *DARTSSearcher) ExportNASState() ([]byte, error) {
	state := NASState{
		AlphaParams:   d.AlphaParams,
		BestScore:     d.BestScore,
		SearchHistory: d.SearchHistory,
	}
	return json.MarshalIndent(state, "", "  ")
}

// RLState captures RL state for persistence.
type RLState struct {
	Episode        int       `json:"episode"`
	TotalReward    float64   `json:"total_reward"`
	EpisodeRewards []float64 `json:"episode_rewards"`
	Epsilon        float64   `json:"epsilon"`
}

// ExportRLState exports RL agent state.
func (ac *MemristorActorCritic) ExportRLState() ([]byte, error) {
	state := RLState{
		Episode:        ac.Episode,
		TotalReward:    ac.TotalReward,
		EpisodeRewards: ac.EpisodeRewards,
	}
	return json.MarshalIndent(state, "", "  ")
}

// =============================================================================
// Benchmarking
// =============================================================================

// NASBenchmark evaluates NAS performance.
type NASBenchmark struct {
	SearchMethod     string
	SearchTime       float64
	NumArchsEvaluated int
	BestAccuracy     float64
	BestEnergy       float64
	BestLatency      float64
	ParetoFrontSize  int
}

// RLBenchmark evaluates RL performance.
type RLBenchmark struct {
	Algorithm        string
	Environment      string
	NumEpisodes      int
	AvgReward        float64
	MaxReward        float64
	SuccessRate      float64
	ConvergenceEp    int
	InferenceEnergy  float64
}

// RunRLBenchmark evaluates RL agent on environment.
func RunRLBenchmark(
	agent *MemristorActorCritic,
	env Environment,
	numEpisodes int,
) RLBenchmark {
	bench := RLBenchmark{
		Algorithm:   agent.Config.Algorithm,
		NumEpisodes: numEpisodes,
	}

	var totalReward float64
	var successes int

	for ep := 0; ep < numEpisodes; ep++ {
		state := env.Reset()
		var episodeReward float64

		for step := 0; step < agent.Config.MaxSteps; step++ {
			action, _ := agent.SelectAction(state)
			nextState, reward, done := env.Step(action)
			agent.Update(state, action, reward, nextState, done)

			episodeReward += reward
			state = nextState

			if done {
				break
			}
		}

		totalReward += episodeReward
		if episodeReward > 0 {
			successes++
		}

		if episodeReward > bench.MaxReward {
			bench.MaxReward = episodeReward
		}
	}

	bench.AvgReward = totalReward / float64(numEpisodes)
	bench.SuccessRate = float64(successes) / float64(numEpisodes)

	return bench
}

// =============================================================================
// Utility Functions
// =============================================================================

// PrintArchitecture displays architecture summary.
func PrintArchitecture(arch *Architecture) string {
	var result string
	for i, cell := range arch.Cells {
		result += fmt.Sprintf("Cell %d:\n", i)
		for j, edge := range cell.Edges {
			result += fmt.Sprintf("  Edge %d: %s (%.3f)\n",
				j, edge.Operation.String(), edge.Weight)
		}
	}
	return result
}

// PrintRLProgress displays RL training progress.
func PrintRLProgress(episode int, reward float64, epsilon float64) string {
	return fmt.Sprintf("Episode %d: Reward=%.2f, Epsilon=%.4f",
		episode, reward, epsilon)
}

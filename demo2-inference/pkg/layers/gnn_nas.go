// Package layers provides neural network layer implementations for CIM simulation.
// gnn_nas.go implements Graph Neural Networks (GNN) on CIM and hardware-aware
// Neural Architecture Search (NAS) for ferroelectric CIM accelerators.
//
// Research basis:
// - ReGNN: 228× speedup, 305× energy reduction vs GPU
// - GNN-PIM: Scatter-gather for message passing
// - CIM²PQ: Arraywise mixed precision quantization
// - Hardware-aware NAS: Co-optimize accuracy and hardware metrics
// - Sparse graphs: CSR format, <0.1% density typical
//
// Key GNN operations:
// - Message: Transform neighbor features
// - Aggregate: Sum/mean/max of messages
// - Update: Combine with node features
// - Readout: Graph-level pooling
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// GRAPH NEURAL NETWORKS
// =============================================================================

// GraphConfig configures graph neural network
type GraphConfig struct {
	// Architecture
	NumLayers      int // Number of GNN layers
	HiddenDim      int // Hidden dimension
	OutputDim      int // Output dimension
	AggregationType AggregationType

	// Dropout/regularization
	Dropout float64

	// CIM mapping
	CrossbarSize int
}

// AggregationType represents GNN aggregation methods
type AggregationType int

const (
	AggSum  AggregationType = iota // Sum aggregation
	AggMean                        // Mean aggregation
	AggMax                         // Max aggregation
)

// DefaultGraphConfig returns typical GNN configuration
func DefaultGraphConfig() *GraphConfig {
	return &GraphConfig{
		NumLayers:       2,
		HiddenDim:       64,
		OutputDim:       10,
		AggregationType: AggMean,
		Dropout:         0.5,
		CrossbarSize:    64,
	}
}

// Graph represents a sparse graph in CSR format
type Graph struct {
	NumNodes    int
	NumEdges    int
	RowPtr      []int       // CSR row pointers
	ColIdx      []int       // CSR column indices
	EdgeWeight  []float64   // Edge weights (optional)
	NodeFeatures [][]float64 // Node feature matrix [nodes × features]
	NodeLabels   []int       // Node labels (for classification)
}

// NewGraph creates a new graph with given structure
func NewGraph(numNodes int, edges [][2]int, features [][]float64) *Graph {
	g := &Graph{
		NumNodes:     numNodes,
		NumEdges:     len(edges),
		NodeFeatures: features,
	}

	// Build CSR representation
	g.buildCSR(edges)

	return g
}

// buildCSR builds compressed sparse row format
func (g *Graph) buildCSR(edges [][2]int) {
	// Count outgoing edges per node
	outDegree := make([]int, g.NumNodes+1)
	for _, e := range edges {
		outDegree[e[0]+1]++
	}

	// Cumulative sum for row pointers
	g.RowPtr = make([]int, g.NumNodes+1)
	for i := 1; i <= g.NumNodes; i++ {
		g.RowPtr[i] = g.RowPtr[i-1] + outDegree[i]
	}

	// Fill column indices
	g.ColIdx = make([]int, g.NumEdges)
	g.EdgeWeight = make([]float64, g.NumEdges)

	// Track current position for each row
	currentPos := make([]int, g.NumNodes)
	copy(currentPos, g.RowPtr[:g.NumNodes])

	for _, e := range edges {
		src, dst := e[0], e[1]
		pos := currentPos[src]
		g.ColIdx[pos] = dst
		g.EdgeWeight[pos] = 1.0 // Default weight
		currentPos[src]++
	}
}

// GetNeighbors returns neighbors of a node
func (g *Graph) GetNeighbors(node int) []int {
	if node < 0 || node >= g.NumNodes {
		return nil
	}
	start := g.RowPtr[node]
	end := g.RowPtr[node+1]
	return g.ColIdx[start:end]
}

// GetDegree returns the degree of a node
func (g *Graph) GetDegree(node int) int {
	if node < 0 || node >= g.NumNodes {
		return 0
	}
	return g.RowPtr[node+1] - g.RowPtr[node]
}

// GetSparsity returns graph sparsity
func (g *Graph) GetSparsity() float64 {
	maxEdges := g.NumNodes * g.NumNodes
	if maxEdges == 0 {
		return 1.0
	}
	return 1.0 - float64(g.NumEdges)/float64(maxEdges)
}

// GNNLayer represents a single GNN layer
type GNNLayer struct {
	config     *GraphConfig
	inputDim   int
	outputDim  int
	weights    [][]float64 // Feature transformation weights
	bias       []float64
	rng        *rand.Rand
}

// NewGNNLayer creates a new GNN layer
func NewGNNLayer(inputDim, outputDim int, config *GraphConfig) *GNNLayer {
	if config == nil {
		config = DefaultGraphConfig()
	}

	layer := &GNNLayer{
		config:    config,
		inputDim:  inputDim,
		outputDim: outputDim,
		rng:       rand.New(rand.NewSource(42)),
	}

	// Initialize weights (Xavier initialization)
	layer.weights = make([][]float64, outputDim)
	scale := math.Sqrt(2.0 / float64(inputDim+outputDim))
	for i := 0; i < outputDim; i++ {
		layer.weights[i] = make([]float64, inputDim)
		for j := 0; j < inputDim; j++ {
			layer.weights[i][j] = layer.rng.NormFloat64() * scale
		}
	}

	layer.bias = make([]float64, outputDim)

	return layer
}

// MessagePass performs message passing on graph
func (layer *GNNLayer) MessagePass(graph *Graph, nodeFeatures [][]float64) [][]float64 {
	numNodes := graph.NumNodes
	outputDim := layer.outputDim

	// Aggregated messages for each node
	messages := make([][]float64, numNodes)
	for i := 0; i < numNodes; i++ {
		messages[i] = make([]float64, outputDim)
	}

	// Message passing: gather from neighbors
	for node := 0; node < numNodes; node++ {
		neighbors := graph.GetNeighbors(node)
		degree := float64(len(neighbors))

		if degree == 0 {
			continue
		}

		// Aggregate neighbor features
		for _, neighbor := range neighbors {
			// Transform neighbor features
			transformed := layer.transformFeatures(nodeFeatures[neighbor])

			// Aggregate
			switch layer.config.AggregationType {
			case AggSum:
				for d := 0; d < outputDim; d++ {
					messages[node][d] += transformed[d]
				}
			case AggMean:
				for d := 0; d < outputDim; d++ {
					messages[node][d] += transformed[d] / degree
				}
			case AggMax:
				for d := 0; d < outputDim; d++ {
					if transformed[d] > messages[node][d] {
						messages[node][d] = transformed[d]
					}
				}
			}
		}
	}

	// Update: combine with self features and apply nonlinearity
	output := make([][]float64, numNodes)
	for node := 0; node < numNodes; node++ {
		output[node] = make([]float64, outputDim)

		// Self transformation
		selfTransformed := layer.transformFeatures(nodeFeatures[node])

		// Combine self and neighbor messages
		for d := 0; d < outputDim; d++ {
			output[node][d] = selfTransformed[d] + messages[node][d] + layer.bias[d]
			// ReLU activation
			if output[node][d] < 0 {
				output[node][d] = 0
			}
		}

		// Dropout during training (simulated)
		if layer.config.Dropout > 0 {
			for d := 0; d < outputDim; d++ {
				if layer.rng.Float64() < layer.config.Dropout {
					output[node][d] = 0
				} else {
					output[node][d] /= (1 - layer.config.Dropout)
				}
			}
		}
	}

	return output
}

// transformFeatures applies weight transformation (MVM on crossbar)
func (layer *GNNLayer) transformFeatures(features []float64) []float64 {
	output := make([]float64, layer.outputDim)

	for i := 0; i < layer.outputDim; i++ {
		sum := 0.0
		for j := 0; j < len(features) && j < layer.inputDim; j++ {
			sum += layer.weights[i][j] * features[j]
		}
		output[i] = sum
	}

	return output
}

// GNN represents a complete Graph Neural Network
type GNN struct {
	config *GraphConfig
	layers []*GNNLayer

	// Statistics
	totalOps int64
}

// NewGNN creates a new GNN
func NewGNN(inputDim int, config *GraphConfig) *GNN {
	if config == nil {
		config = DefaultGraphConfig()
	}

	gnn := &GNN{
		config: config,
		layers: make([]*GNNLayer, config.NumLayers),
	}

	// Create layers
	prevDim := inputDim
	for i := 0; i < config.NumLayers; i++ {
		var outDim int
		if i == config.NumLayers-1 {
			outDim = config.OutputDim
		} else {
			outDim = config.HiddenDim
		}

		gnn.layers[i] = NewGNNLayer(prevDim, outDim, config)
		prevDim = outDim
	}

	return gnn
}

// Forward performs forward pass on graph
func (gnn *GNN) Forward(graph *Graph) [][]float64 {
	features := graph.NodeFeatures

	// Pass through each GNN layer
	for _, layer := range gnn.layers {
		features = layer.MessagePass(graph, features)
		gnn.totalOps += int64(graph.NumNodes * layer.inputDim * layer.outputDim)
	}

	return features
}

// NodeClassification performs node classification
func (gnn *GNN) NodeClassification(graph *Graph) []int {
	features := gnn.Forward(graph)

	predictions := make([]int, graph.NumNodes)
	for node, feat := range features {
		// Argmax
		maxIdx := 0
		maxVal := feat[0]
		for i := 1; i < len(feat); i++ {
			if feat[i] > maxVal {
				maxVal = feat[i]
				maxIdx = i
			}
		}
		predictions[node] = maxIdx
	}

	return predictions
}

// GraphPooling performs graph-level pooling
func (gnn *GNN) GraphPooling(features [][]float64, poolType AggregationType) []float64 {
	if len(features) == 0 {
		return nil
	}

	dim := len(features[0])
	pooled := make([]float64, dim)

	switch poolType {
	case AggSum:
		for _, feat := range features {
			for d := 0; d < dim; d++ {
				pooled[d] += feat[d]
			}
		}
	case AggMean:
		for _, feat := range features {
			for d := 0; d < dim; d++ {
				pooled[d] += feat[d]
			}
		}
		for d := 0; d < dim; d++ {
			pooled[d] /= float64(len(features))
		}
	case AggMax:
		copy(pooled, features[0])
		for _, feat := range features[1:] {
			for d := 0; d < dim; d++ {
				if feat[d] > pooled[d] {
					pooled[d] = feat[d]
				}
			}
		}
	}

	return pooled
}

// =============================================================================
// HARDWARE-AWARE NEURAL ARCHITECTURE SEARCH
// =============================================================================

// NASSearchSpace defines the search space for NAS
type NASSearchSpace struct {
	// Layer options
	NumLayersRange    [2]int   // Min, max number of layers
	ChannelOptions    []int    // Possible channel counts
	KernelOptions     []int    // Possible kernel sizes

	// Quantization options
	WeightBitOptions  []int    // Possible weight bit widths
	ActivBitOptions   []int    // Possible activation bit widths

	// Architecture options
	LayerTypes        []string // "conv", "dwconv", "fc", etc.
	SkipConnection    bool     // Allow skip connections
}

// DefaultNASSearchSpace returns typical search space
func DefaultNASSearchSpace() *NASSearchSpace {
	return &NASSearchSpace{
		NumLayersRange:   [2]int{4, 12},
		ChannelOptions:   []int{16, 32, 64, 128, 256},
		KernelOptions:    []int{1, 3, 5},
		WeightBitOptions: []int{2, 4, 6, 8},
		ActivBitOptions:  []int{4, 6, 8},
		LayerTypes:       []string{"conv", "dwconv", "mbconv"},
		SkipConnection:   true,
	}
}

// NASConfig configures the NAS algorithm
type NASConfig struct {
	SearchSpace     *NASSearchSpace

	// Optimization targets
	TargetAccuracy   float64 // Minimum acceptable accuracy
	TargetLatencyMs  float64 // Maximum latency
	TargetEnergyPJ   float64 // Maximum energy

	// Search parameters
	PopulationSize   int     // For evolutionary search
	NumGenerations   int     // Number of generations
	MutationRate     float64 // Mutation probability
	CrossoverRate    float64 // Crossover probability

	// Hardware constraints
	CrossbarSize     int     // CIM crossbar size
	NumCrossbars     int     // Available crossbars
}

// DefaultNASConfig returns typical NAS configuration
func DefaultNASConfig() *NASConfig {
	return &NASConfig{
		SearchSpace:     DefaultNASSearchSpace(),
		TargetAccuracy:  0.90,
		TargetLatencyMs: 10,
		TargetEnergyPJ:  1e9,
		PopulationSize:  50,
		NumGenerations:  100,
		MutationRate:    0.1,
		CrossoverRate:   0.5,
		CrossbarSize:    64,
		NumCrossbars:    16,
	}
}

// Architecture represents a neural network architecture
type Architecture struct {
	Layers       []LayerSpec
	TotalParams  int64
	TotalMACs    int64

	// Evaluated metrics
	Accuracy     float64
	LatencyMs    float64
	EnergyPJ     float64
	Fitness      float64
}

// LayerSpec specifies a single layer
type LayerSpec struct {
	Type         string  // "conv", "dwconv", "fc", etc.
	InChannels   int
	OutChannels  int
	KernelSize   int
	Stride       int
	WeightBits   int
	ActivBits    int
	HasSkip      bool
}

// NASEngine performs hardware-aware neural architecture search
type NASEngine struct {
	config     *NASConfig
	rng        *rand.Rand
	population []*Architecture
	bestArch   *Architecture

	// Statistics
	generationsRun int
	evaluations    int
}

// NewNASEngine creates a new NAS engine
func NewNASEngine(config *NASConfig) *NASEngine {
	if config == nil {
		config = DefaultNASConfig()
	}

	return &NASEngine{
		config:     config,
		rng:        rand.New(rand.NewSource(42)),
		population: make([]*Architecture, 0),
	}
}

// InitializePopulation creates initial random population
func (nas *NASEngine) InitializePopulation() {
	nas.population = make([]*Architecture, nas.config.PopulationSize)

	for i := 0; i < nas.config.PopulationSize; i++ {
		nas.population[i] = nas.randomArchitecture()
	}
}

// randomArchitecture generates a random architecture
func (nas *NASEngine) randomArchitecture() *Architecture {
	space := nas.config.SearchSpace

	// Random number of layers
	numLayers := space.NumLayersRange[0] +
		nas.rng.Intn(space.NumLayersRange[1]-space.NumLayersRange[0]+1)

	arch := &Architecture{
		Layers: make([]LayerSpec, numLayers),
	}

	inChannels := 3 // Assume RGB input
	for i := 0; i < numLayers; i++ {
		outChannels := space.ChannelOptions[nas.rng.Intn(len(space.ChannelOptions))]
		kernelSize := space.KernelOptions[nas.rng.Intn(len(space.KernelOptions))]
		layerType := space.LayerTypes[nas.rng.Intn(len(space.LayerTypes))]
		weightBits := space.WeightBitOptions[nas.rng.Intn(len(space.WeightBitOptions))]
		activBits := space.ActivBitOptions[nas.rng.Intn(len(space.ActivBitOptions))]

		arch.Layers[i] = LayerSpec{
			Type:        layerType,
			InChannels:  inChannels,
			OutChannels: outChannels,
			KernelSize:  kernelSize,
			Stride:      1,
			WeightBits:  weightBits,
			ActivBits:   activBits,
			HasSkip:     space.SkipConnection && nas.rng.Float64() > 0.5,
		}

		inChannels = outChannels
	}

	// Calculate params and MACs
	nas.calculateMetrics(arch)

	return arch
}

// calculateMetrics estimates architecture metrics
func (nas *NASEngine) calculateMetrics(arch *Architecture) {
	var totalParams, totalMACs int64
	inputSize := 32 // Assume 32x32 input

	for _, layer := range arch.Layers {
		switch layer.Type {
		case "conv":
			params := int64(layer.OutChannels * layer.InChannels * layer.KernelSize * layer.KernelSize)
			macs := params * int64(inputSize*inputSize)
			totalParams += params
			totalMACs += macs

		case "dwconv":
			// Depthwise: only kernel per channel
			params := int64(layer.InChannels * layer.KernelSize * layer.KernelSize)
			macs := params * int64(inputSize*inputSize)
			totalParams += params
			totalMACs += macs

		case "mbconv":
			// MobileNet-style inverted residual
			expand := layer.InChannels * 4
			params := int64(layer.InChannels*expand + expand*layer.KernelSize*layer.KernelSize + expand*layer.OutChannels)
			macs := params * int64(inputSize*inputSize)
			totalParams += params
			totalMACs += macs

		case "fc":
			params := int64(layer.InChannels * layer.OutChannels)
			totalParams += params
			totalMACs += params
		}

		// Update spatial size (simplified)
		if layer.Stride == 2 {
			inputSize /= 2
		}
	}

	arch.TotalParams = totalParams
	arch.TotalMACs = totalMACs
}

// EvaluateArchitecture evaluates architecture on hardware model
func (nas *NASEngine) EvaluateArchitecture(arch *Architecture) {
	nas.evaluations++

	// Estimate accuracy (simplified model)
	// Deeper and wider networks tend to be more accurate
	depthFactor := math.Min(float64(len(arch.Layers))/10.0, 1.0)
	widthFactor := 0.0
	for _, layer := range arch.Layers {
		widthFactor += float64(layer.OutChannels)
	}
	widthFactor = math.Min(widthFactor/1000.0, 1.0)

	// Quantization affects accuracy
	avgBits := 0.0
	for _, layer := range arch.Layers {
		avgBits += float64(layer.WeightBits)
	}
	avgBits /= float64(len(arch.Layers))
	quantFactor := math.Min(avgBits/8.0, 1.0)

	arch.Accuracy = 0.7 + 0.15*depthFactor + 0.10*widthFactor + 0.05*quantFactor

	// Estimate latency on CIM
	// Based on crossbar utilization and number of tiles
	crossbarOps := int64(nas.config.CrossbarSize * nas.config.CrossbarSize)
	numTiles := (arch.TotalMACs + crossbarOps - 1) / crossbarOps
	cyclesPerTile := 10.0 // Assume 10 cycles per tile operation
	clockPeriodNs := 10.0 // 100 MHz

	arch.LatencyMs = float64(numTiles) * cyclesPerTile * clockPeriodNs / 1e6

	// Estimate energy
	// Energy per MAC depends on bit width
	energyPerMAC := 0.1 // pJ baseline for 8-bit
	for _, layer := range arch.Layers {
		bitRatio := float64(layer.WeightBits) / 8.0
		arch.EnergyPJ += float64(arch.TotalMACs) * energyPerMAC * bitRatio / float64(len(arch.Layers))
	}

	// Calculate fitness (multi-objective)
	accScore := arch.Accuracy / nas.config.TargetAccuracy
	latScore := nas.config.TargetLatencyMs / (arch.LatencyMs + 0.001)
	energyScore := nas.config.TargetEnergyPJ / (arch.EnergyPJ + 1)

	// Weighted fitness
	arch.Fitness = 0.5*accScore + 0.25*math.Min(latScore, 1.0) + 0.25*math.Min(energyScore, 1.0)
}

// Evolve runs one generation of evolution
func (nas *NASEngine) Evolve() {
	// Evaluate all architectures
	for _, arch := range nas.population {
		if arch.Fitness == 0 {
			nas.EvaluateArchitecture(arch)
		}
	}

	// Sort by fitness
	sort.Slice(nas.population, func(i, j int) bool {
		return nas.population[i].Fitness > nas.population[j].Fitness
	})

	// Update best
	if nas.bestArch == nil || nas.population[0].Fitness > nas.bestArch.Fitness {
		nas.bestArch = nas.population[0]
	}

	// Selection and reproduction
	newPop := make([]*Architecture, nas.config.PopulationSize)

	// Elitism: keep top 10%
	eliteCount := nas.config.PopulationSize / 10
	for i := 0; i < eliteCount; i++ {
		newPop[i] = nas.population[i]
	}

	// Fill rest with crossover and mutation
	for i := eliteCount; i < nas.config.PopulationSize; i++ {
		// Tournament selection
		parent1 := nas.tournamentSelect()
		parent2 := nas.tournamentSelect()

		// Crossover
		child := nas.crossover(parent1, parent2)

		// Mutation
		if nas.rng.Float64() < nas.config.MutationRate {
			nas.mutate(child)
		}

		nas.calculateMetrics(child)
		newPop[i] = child
	}

	nas.population = newPop
	nas.generationsRun++
}

// tournamentSelect performs tournament selection
func (nas *NASEngine) tournamentSelect() *Architecture {
	tournamentSize := 5
	best := nas.population[nas.rng.Intn(len(nas.population))]

	for i := 1; i < tournamentSize; i++ {
		candidate := nas.population[nas.rng.Intn(len(nas.population))]
		if candidate.Fitness > best.Fitness {
			best = candidate
		}
	}

	return best
}

// crossover creates child from two parents
func (nas *NASEngine) crossover(parent1, parent2 *Architecture) *Architecture {
	if nas.rng.Float64() > nas.config.CrossoverRate {
		// No crossover, return copy of parent1
		return nas.copyArchitecture(parent1)
	}

	// Uniform crossover of layers
	maxLayers := len(parent1.Layers)
	if len(parent2.Layers) > maxLayers {
		maxLayers = len(parent2.Layers)
	}

	child := &Architecture{
		Layers: make([]LayerSpec, 0),
	}

	for i := 0; i < maxLayers; i++ {
		if nas.rng.Float64() < 0.5 && i < len(parent1.Layers) {
			child.Layers = append(child.Layers, parent1.Layers[i])
		} else if i < len(parent2.Layers) {
			child.Layers = append(child.Layers, parent2.Layers[i])
		}
	}

	// Fix channel consistency
	for i := 1; i < len(child.Layers); i++ {
		child.Layers[i].InChannels = child.Layers[i-1].OutChannels
	}

	return child
}

// copyArchitecture creates a deep copy
func (nas *NASEngine) copyArchitecture(arch *Architecture) *Architecture {
	copy := &Architecture{
		Layers: make([]LayerSpec, len(arch.Layers)),
	}
	for i, layer := range arch.Layers {
		copy.Layers[i] = layer
	}
	return copy
}

// mutate modifies an architecture
func (nas *NASEngine) mutate(arch *Architecture) {
	space := nas.config.SearchSpace

	// Random mutation type
	mutationType := nas.rng.Intn(5)

	switch mutationType {
	case 0: // Change layer type
		if len(arch.Layers) > 0 {
			idx := nas.rng.Intn(len(arch.Layers))
			arch.Layers[idx].Type = space.LayerTypes[nas.rng.Intn(len(space.LayerTypes))]
		}

	case 1: // Change channels
		if len(arch.Layers) > 0 {
			idx := nas.rng.Intn(len(arch.Layers))
			arch.Layers[idx].OutChannels = space.ChannelOptions[nas.rng.Intn(len(space.ChannelOptions))]
		}

	case 2: // Change kernel size
		if len(arch.Layers) > 0 {
			idx := nas.rng.Intn(len(arch.Layers))
			arch.Layers[idx].KernelSize = space.KernelOptions[nas.rng.Intn(len(space.KernelOptions))]
		}

	case 3: // Change quantization
		if len(arch.Layers) > 0 {
			idx := nas.rng.Intn(len(arch.Layers))
			arch.Layers[idx].WeightBits = space.WeightBitOptions[nas.rng.Intn(len(space.WeightBitOptions))]
			arch.Layers[idx].ActivBits = space.ActivBitOptions[nas.rng.Intn(len(space.ActivBitOptions))]
		}

	case 4: // Add/remove layer
		if nas.rng.Float64() < 0.5 && len(arch.Layers) < space.NumLayersRange[1] {
			// Add layer
			idx := nas.rng.Intn(len(arch.Layers) + 1)
			newLayer := LayerSpec{
				Type:        space.LayerTypes[nas.rng.Intn(len(space.LayerTypes))],
				OutChannels: space.ChannelOptions[nas.rng.Intn(len(space.ChannelOptions))],
				KernelSize:  space.KernelOptions[nas.rng.Intn(len(space.KernelOptions))],
				WeightBits:  space.WeightBitOptions[nas.rng.Intn(len(space.WeightBitOptions))],
				ActivBits:   space.ActivBitOptions[nas.rng.Intn(len(space.ActivBitOptions))],
			}
			arch.Layers = append(arch.Layers[:idx], append([]LayerSpec{newLayer}, arch.Layers[idx:]...)...)
		} else if len(arch.Layers) > space.NumLayersRange[0] {
			// Remove layer
			idx := nas.rng.Intn(len(arch.Layers))
			arch.Layers = append(arch.Layers[:idx], arch.Layers[idx+1:]...)
		}
	}

	// Fix channel consistency
	if len(arch.Layers) > 0 {
		arch.Layers[0].InChannels = 3
		for i := 1; i < len(arch.Layers); i++ {
			arch.Layers[i].InChannels = arch.Layers[i-1].OutChannels
		}
	}
}

// Search performs full NAS search
func (nas *NASEngine) Search() *Architecture {
	nas.InitializePopulation()

	for gen := 0; gen < nas.config.NumGenerations; gen++ {
		nas.Evolve()
	}

	return nas.bestArch
}

// GetSearchProgress returns current search progress
func (nas *NASEngine) GetSearchProgress() NASProgress {
	var avgFitness float64
	for _, arch := range nas.population {
		avgFitness += arch.Fitness
	}
	avgFitness /= float64(len(nas.population))

	bestFitness := 0.0
	if nas.bestArch != nil {
		bestFitness = nas.bestArch.Fitness
	}

	return NASProgress{
		Generation:    nas.generationsRun,
		Evaluations:   nas.evaluations,
		BestFitness:   bestFitness,
		AvgFitness:    avgFitness,
		BestAccuracy:  nas.bestArch.Accuracy,
		BestLatencyMs: nas.bestArch.LatencyMs,
	}
}

// NASProgress holds NAS search progress
type NASProgress struct {
	Generation    int
	Evaluations   int
	BestFitness   float64
	AvgFitness    float64
	BestAccuracy  float64
	BestLatencyMs float64
}

// =============================================================================
// MIXED PRECISION QUANTIZATION
// =============================================================================

// MPQConfig configures mixed precision quantization
type MPQConfig struct {
	// Search space
	BitOptions []int // Available bit widths

	// Constraints
	TargetModelSize int64   // Target model size in bytes
	TargetAccuracy  float64 // Minimum accuracy

	// Search parameters
	PopulationSize int
	Generations    int
}

// DefaultMPQConfig returns typical MPQ configuration
func DefaultMPQConfig() *MPQConfig {
	return &MPQConfig{
		BitOptions:      []int{2, 4, 6, 8},
		TargetModelSize: 1024 * 1024, // 1 MB
		TargetAccuracy:  0.90,
		PopulationSize:  20,
		Generations:     50,
	}
}

// MPQEngine performs mixed precision quantization search
type MPQEngine struct {
	config      *MPQConfig
	architecture *Architecture
	rng         *rand.Rand
}

// NewMPQEngine creates a new MPQ engine
func NewMPQEngine(arch *Architecture, config *MPQConfig) *MPQEngine {
	if config == nil {
		config = DefaultMPQConfig()
	}

	return &MPQEngine{
		config:       config,
		architecture: arch,
		rng:          rand.New(rand.NewSource(42)),
	}
}

// SearchOptimalQuantization finds optimal per-layer quantization
func (mpq *MPQEngine) SearchOptimalQuantization() []int {
	numLayers := len(mpq.architecture.Layers)
	bestQuant := make([]int, numLayers)
	bestScore := 0.0

	// Simple evolutionary search
	population := make([][]int, mpq.config.PopulationSize)
	for i := range population {
		population[i] = mpq.randomQuantization(numLayers)
	}

	for gen := 0; gen < mpq.config.Generations; gen++ {
		// Evaluate
		scores := make([]float64, len(population))
		for i, quant := range population {
			scores[i] = mpq.evaluateQuantization(quant)
			if scores[i] > bestScore {
				bestScore = scores[i]
				copy(bestQuant, quant)
			}
		}

		// Evolve
		newPop := make([][]int, mpq.config.PopulationSize)
		for i := range newPop {
			// Tournament selection
			p1 := population[mpq.rng.Intn(len(population))]
			p2 := population[mpq.rng.Intn(len(population))]

			// Crossover
			child := make([]int, numLayers)
			for j := 0; j < numLayers; j++ {
				if mpq.rng.Float64() < 0.5 {
					child[j] = p1[j]
				} else {
					child[j] = p2[j]
				}
			}

			// Mutation
			if mpq.rng.Float64() < 0.1 {
				idx := mpq.rng.Intn(numLayers)
				child[idx] = mpq.config.BitOptions[mpq.rng.Intn(len(mpq.config.BitOptions))]
			}

			newPop[i] = child
		}
		population = newPop
	}

	return bestQuant
}

// randomQuantization generates random per-layer quantization
func (mpq *MPQEngine) randomQuantization(numLayers int) []int {
	quant := make([]int, numLayers)
	for i := range quant {
		quant[i] = mpq.config.BitOptions[mpq.rng.Intn(len(mpq.config.BitOptions))]
	}
	return quant
}

// evaluateQuantization evaluates quantization scheme
func (mpq *MPQEngine) evaluateQuantization(quant []int) float64 {
	// Calculate model size
	var modelSize int64
	for i, layer := range mpq.architecture.Layers {
		params := int64(layer.InChannels * layer.OutChannels)
		if layer.Type == "conv" {
			params *= int64(layer.KernelSize * layer.KernelSize)
		}
		modelSize += params * int64(quant[i]) / 8
	}

	// Estimate accuracy drop from quantization
	avgBits := 0.0
	for _, b := range quant {
		avgBits += float64(b)
	}
	avgBits /= float64(len(quant))
	accFactor := math.Min(avgBits/8.0, 1.0)
	estimatedAcc := mpq.architecture.Accuracy * accFactor

	// Fitness: balance accuracy and size
	sizeScore := float64(mpq.config.TargetModelSize) / float64(modelSize+1)
	accScore := estimatedAcc / mpq.config.TargetAccuracy

	return 0.6*accScore + 0.4*math.Min(sizeScore, 1.0)
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// GenerateRandomGraph creates a random graph for testing
func GenerateRandomGraph(numNodes, avgDegree int, featureDim int) *Graph {
	rng := rand.New(rand.NewSource(42))

	// Generate random edges
	edges := make([][2]int, 0)
	for i := 0; i < numNodes; i++ {
		numEdges := avgDegree + rng.Intn(avgDegree)
		for j := 0; j < numEdges; j++ {
			dst := rng.Intn(numNodes)
			if dst != i {
				edges = append(edges, [2]int{i, dst})
			}
		}
	}

	// Generate random features
	features := make([][]float64, numNodes)
	for i := 0; i < numNodes; i++ {
		features[i] = make([]float64, featureDim)
		for j := 0; j < featureDim; j++ {
			features[i][j] = rng.NormFloat64()
		}
	}

	return NewGraph(numNodes, edges, features)
}

// FormatGNNReport generates GNN analysis report
func FormatGNNReport(gnn *GNN, graph *Graph) string {
	report := "=== GNN Analysis Report ===\n\n"
	report += fmt.Sprintf("Graph Statistics:\n")
	report += fmt.Sprintf("  Nodes: %d\n", graph.NumNodes)
	report += fmt.Sprintf("  Edges: %d\n", graph.NumEdges)
	report += fmt.Sprintf("  Sparsity: %.4f%%\n", graph.GetSparsity()*100)

	report += fmt.Sprintf("\nGNN Architecture:\n")
	report += fmt.Sprintf("  Layers: %d\n", len(gnn.layers))
	for i, layer := range gnn.layers {
		report += fmt.Sprintf("    Layer %d: %d → %d\n", i+1, layer.inputDim, layer.outputDim)
	}

	report += fmt.Sprintf("\nCIM Mapping:\n")
	report += fmt.Sprintf("  Crossbar Size: %d × %d\n", gnn.config.CrossbarSize, gnn.config.CrossbarSize)
	report += fmt.Sprintf("  Total Operations: %d\n", gnn.totalOps)

	return report
}

// FormatNASReport generates NAS search report
func FormatNASReport(nas *NASEngine) string {
	progress := nas.GetSearchProgress()

	report := "=== NAS Search Report ===\n\n"
	report += fmt.Sprintf("Search Progress:\n")
	report += fmt.Sprintf("  Generations: %d\n", progress.Generation)
	report += fmt.Sprintf("  Evaluations: %d\n", progress.Evaluations)
	report += fmt.Sprintf("  Best Fitness: %.4f\n", progress.BestFitness)
	report += fmt.Sprintf("  Avg Fitness: %.4f\n", progress.AvgFitness)

	if nas.bestArch != nil {
		report += fmt.Sprintf("\nBest Architecture:\n")
		report += fmt.Sprintf("  Layers: %d\n", len(nas.bestArch.Layers))
		report += fmt.Sprintf("  Params: %d\n", nas.bestArch.TotalParams)
		report += fmt.Sprintf("  MACs: %d\n", nas.bestArch.TotalMACs)
		report += fmt.Sprintf("  Est. Accuracy: %.2f%%\n", nas.bestArch.Accuracy*100)
		report += fmt.Sprintf("  Est. Latency: %.3f ms\n", nas.bestArch.LatencyMs)
		report += fmt.Sprintf("  Est. Energy: %.2e pJ\n", nas.bestArch.EnergyPJ)
	}

	return report
}

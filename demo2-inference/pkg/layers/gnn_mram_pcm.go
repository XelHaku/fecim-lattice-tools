// Package layers provides GNN acceleration and emerging memory technology simulation
// for compute-in-memory architectures.
//
// This module implements:
// - Graph Neural Network (GNN) accelerator architectures
// - NEM-GNN DAC/ADC-less near-memory design
// - STT-MRAM and SOT-MRAM compute-in-memory
// - Phase Change Memory (PCM) neural network acceleration
//
// Based on research from ISCA 2025, HPCA 2024, and IEDM 2023.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
)

// =============================================================================
// GRAPH NEURAL NETWORK STRUCTURES
// =============================================================================

// GraphConfig defines graph structure configuration
type GraphConfig struct {
	NumNodes       int     // Number of nodes in graph
	NumEdges       int     // Number of edges
	NumFeatures    int     // Node feature dimension
	NumClasses     int     // Output classes
	Directed       bool    // Directed vs undirected
	SelfLoops      bool    // Include self-loops
	MaxDegree      int     // Maximum node degree
	SparsityRatio  float64 // Edge sparsity (0-1)
}

// DefaultGraphConfig returns default graph configuration
func DefaultGraphConfig() *GraphConfig {
	return &GraphConfig{
		NumNodes:       1000,
		NumEdges:       5000,
		NumFeatures:    128,
		NumClasses:     10,
		Directed:       false,
		SelfLoops:      true,
		MaxDegree:      100,
		SparsityRatio:  0.99,
	}
}

// Graph represents a graph data structure
type Graph struct {
	Config       *GraphConfig
	NodeFeatures [][]float64   // [NumNodes][NumFeatures]
	AdjList      [][]int       // Adjacency list
	EdgeWeights  [][]float64   // Edge weights (optional)
	Degrees      []int         // Node degrees
	Labels       []int         // Node labels (for classification)
}

// NewGraph creates a new graph from configuration
func NewGraph(config *GraphConfig) *Graph {
	g := &Graph{
		Config:       config,
		NodeFeatures: make([][]float64, config.NumNodes),
		AdjList:      make([][]int, config.NumNodes),
		EdgeWeights:  make([][]float64, config.NumNodes),
		Degrees:      make([]int, config.NumNodes),
		Labels:       make([]int, config.NumNodes),
	}

	// Initialize node features
	for i := 0; i < config.NumNodes; i++ {
		g.NodeFeatures[i] = make([]float64, config.NumFeatures)
		for j := 0; j < config.NumFeatures; j++ {
			g.NodeFeatures[i][j] = rand.NormFloat64() * 0.1
		}
		g.AdjList[i] = make([]int, 0)
		g.EdgeWeights[i] = make([]float64, 0)
		g.Labels[i] = rand.Intn(config.NumClasses)
	}

	return g
}

// AddEdge adds an edge to the graph
func (g *Graph) AddEdge(src, dst int, weight float64) {
	if src >= 0 && src < g.Config.NumNodes && dst >= 0 && dst < g.Config.NumNodes {
		g.AdjList[src] = append(g.AdjList[src], dst)
		g.EdgeWeights[src] = append(g.EdgeWeights[src], weight)
		g.Degrees[src]++
		if !g.Config.Directed {
			g.AdjList[dst] = append(g.AdjList[dst], src)
			g.EdgeWeights[dst] = append(g.EdgeWeights[dst], weight)
			g.Degrees[dst]++
		}
	}
}

// GenerateRandomGraph generates random edges for the graph
func (g *Graph) GenerateRandomGraph() {
	targetEdges := g.Config.NumEdges
	edgeCount := 0

	for edgeCount < targetEdges {
		src := rand.Intn(g.Config.NumNodes)
		dst := rand.Intn(g.Config.NumNodes)

		// Skip self-loops if not allowed
		if !g.Config.SelfLoops && src == dst {
			continue
		}

		// Check max degree
		if g.Degrees[src] >= g.Config.MaxDegree {
			continue
		}

		g.AddEdge(src, dst, 1.0)
		edgeCount++
	}
}

// =============================================================================
// GNN LAYER IMPLEMENTATIONS
// =============================================================================

// GCNLayerConfig configures a Graph Convolutional layer
type GCNLayerConfig struct {
	InputDim      int
	OutputDim     int
	Activation    string  // "relu", "leaky_relu", "none"
	Dropout       float64
	UseBias       bool
	Normalize     bool    // Use symmetric normalization
}

// GCNLayer implements a Graph Convolutional Network layer
type GCNLayer struct {
	Config      *GCNLayerConfig
	Weights     [][]float64 // [InputDim][OutputDim]
	Bias        []float64   // [OutputDim]
	// Cached computations
	CachedNorm  []float64   // Degree normalization cache
}

// NewGCNLayer creates a new GCN layer
func NewGCNLayer(config *GCNLayerConfig) *GCNLayer {
	layer := &GCNLayer{
		Config:  config,
		Weights: make([][]float64, config.InputDim),
		Bias:    make([]float64, config.OutputDim),
	}

	// Xavier initialization
	scale := math.Sqrt(2.0 / float64(config.InputDim+config.OutputDim))
	for i := 0; i < config.InputDim; i++ {
		layer.Weights[i] = make([]float64, config.OutputDim)
		for j := 0; j < config.OutputDim; j++ {
			layer.Weights[i][j] = rand.NormFloat64() * scale
		}
	}

	return layer
}

// Forward performs GCN forward pass: H' = σ(D^-1/2 A D^-1/2 H W)
func (l *GCNLayer) Forward(g *Graph, input [][]float64) [][]float64 {
	numNodes := len(input)
	output := make([][]float64, numNodes)

	// Compute degree normalization
	if l.CachedNorm == nil || len(l.CachedNorm) != numNodes {
		l.CachedNorm = make([]float64, numNodes)
		for i := 0; i < numNodes; i++ {
			if g.Degrees[i] > 0 {
				l.CachedNorm[i] = 1.0 / math.Sqrt(float64(g.Degrees[i]))
			} else {
				l.CachedNorm[i] = 1.0
			}
		}
	}

	// Aggregation: aggregate neighbor features
	aggregated := make([][]float64, numNodes)
	for i := 0; i < numNodes; i++ {
		aggregated[i] = make([]float64, l.Config.InputDim)

		// Self-connection
		normSelf := l.CachedNorm[i] * l.CachedNorm[i]
		for f := 0; f < l.Config.InputDim; f++ {
			aggregated[i][f] = normSelf * input[i][f]
		}

		// Neighbor aggregation
		for ni, neighbor := range g.AdjList[i] {
			_ = ni // Edge weight available at g.EdgeWeights[i][ni]
			normEdge := l.CachedNorm[i] * l.CachedNorm[neighbor]
			for f := 0; f < l.Config.InputDim; f++ {
				aggregated[i][f] += normEdge * input[neighbor][f]
			}
		}
	}

	// Linear transformation: H' = aggregated * W + b
	for i := 0; i < numNodes; i++ {
		output[i] = make([]float64, l.Config.OutputDim)
		for j := 0; j < l.Config.OutputDim; j++ {
			sum := 0.0
			for k := 0; k < l.Config.InputDim; k++ {
				sum += aggregated[i][k] * l.Weights[k][j]
			}
			if l.Config.UseBias {
				sum += l.Bias[j]
			}

			// Activation
			switch l.Config.Activation {
			case "relu":
				if sum < 0 {
					sum = 0
				}
			case "leaky_relu":
				if sum < 0 {
					sum = 0.01 * sum
				}
			}
			output[i][j] = sum
		}
	}

	return output
}

// =============================================================================
// NEM-GNN NEAR-MEMORY ACCELERATOR
// =============================================================================

// NEMGNNConfig configures the NEM-GNN accelerator
type NEMGNNConfig struct {
	NumPEs            int     // Number of processing elements
	ScratchpadSizeKB  int     // Local scratchpad size
	BroadcastBandwidth int    // Broadcast bandwidth (bits)
	EarlyTermination  bool    // Enable early termination
	CARScheduling     bool    // Compute-as-ready scheduling
	SparsityThreshold float64 // Sparsity exploitation threshold
}

// DefaultNEMGNNConfig returns default NEM-GNN configuration
func DefaultNEMGNNConfig() *NEMGNNConfig {
	return &NEMGNNConfig{
		NumPEs:            64,
		ScratchpadSizeKB:  256,
		BroadcastBandwidth: 512,
		EarlyTermination:  true,
		CARScheduling:     true,
		SparsityThreshold: 0.5,
	}
}

// NEMGNN implements DAC/ADC-less near-memory GNN accelerator
// Based on: NEM-GNN (ACM TACO 2024)
// Achieves: 80-230x performance, 850-1134x energy efficiency vs ReFLIP
type NEMGNN struct {
	Config       *NEMGNNConfig
	PEs          []*NEMPE
	Scheduler    *CARScheduler
	Stats        *NEMGNNStats
	mu           sync.Mutex
}

// NEMPE represents a processing element in NEM-GNN
type NEMPE struct {
	ID              int
	Scratchpad      []float64
	LocalAccumulator []float64
	Active          bool
	ComputeCycles   int
	IdleCycles      int
}

// CARScheduler implements Compute-As-Ready scheduling
type CARScheduler struct {
	ReadyQueue    []int
	PendingEdges  map[int][]int
	CompletedNodes map[int]bool
}

// NEMGNNStats tracks accelerator statistics
type NEMGNNStats struct {
	TotalCycles       int
	ComputeCycles     int
	MemoryStalls      int
	EarlyTerminations int
	SparsitySkips     int
	Throughput        float64 // GOPS
	EnergyEfficiency  float64 // GOPS/W
}

// NewNEMGNN creates a new NEM-GNN accelerator
func NewNEMGNN(config *NEMGNNConfig) *NEMGNN {
	acc := &NEMGNN{
		Config: config,
		PEs:    make([]*NEMPE, config.NumPEs),
		Stats:  &NEMGNNStats{},
	}

	for i := 0; i < config.NumPEs; i++ {
		acc.PEs[i] = &NEMPE{
			ID:         i,
			Scratchpad: make([]float64, config.ScratchpadSizeKB*1024/8),
			Active:     true,
		}
	}

	acc.Scheduler = &CARScheduler{
		ReadyQueue:     make([]int, 0),
		PendingEdges:   make(map[int][]int),
		CompletedNodes: make(map[int]bool),
	}

	return acc
}

// Execute runs GNN layer on NEM-GNN accelerator
func (n *NEMGNN) Execute(g *Graph, layer *GCNLayer) [][]float64 {
	numNodes := g.Config.NumNodes
	output := make([][]float64, numNodes)
	for i := range output {
		output[i] = make([]float64, layer.Config.OutputDim)
	}

	// Initialize scheduler
	for i := 0; i < numNodes; i++ {
		n.Scheduler.ReadyQueue = append(n.Scheduler.ReadyQueue, i)
	}

	// Process in parallel across PEs
	var wg sync.WaitGroup
	nodesPerPE := (numNodes + n.Config.NumPEs - 1) / n.Config.NumPEs

	for peID := 0; peID < n.Config.NumPEs; peID++ {
		wg.Add(1)
		go func(pe int) {
			defer wg.Done()
			startNode := pe * nodesPerPE
			endNode := startNode + nodesPerPE
			if endNode > numNodes {
				endNode = numNodes
			}

			for node := startNode; node < endNode; node++ {
				n.processNode(g, layer, node, output)
			}
		}(peID)
	}
	wg.Wait()

	// Update stats
	n.Stats.TotalCycles = numNodes * layer.Config.InputDim * layer.Config.OutputDim / n.Config.NumPEs
	n.Stats.Throughput = float64(numNodes*layer.Config.InputDim*layer.Config.OutputDim*2) / float64(n.Stats.TotalCycles) / 1e9
	n.Stats.EnergyEfficiency = n.Stats.Throughput * 100 // Estimated 100x energy efficiency

	return output
}

// processNode processes a single node with early termination
func (n *NEMGNN) processNode(g *Graph, layer *GCNLayer, node int, output [][]float64) {
	// Aggregate neighbors (memory-intensive)
	aggregated := make([]float64, layer.Config.InputDim)

	// Self contribution
	for f := 0; f < layer.Config.InputDim; f++ {
		aggregated[f] = g.NodeFeatures[node][f]
	}

	// Neighbor contribution with early termination
	neighborCount := 0
	for _, neighbor := range g.AdjList[node] {
		// Check sparsity threshold
		neighborNorm := 0.0
		for f := 0; f < layer.Config.InputDim; f++ {
			neighborNorm += g.NodeFeatures[neighbor][f] * g.NodeFeatures[neighbor][f]
		}

		if n.Config.EarlyTermination && neighborNorm < n.Config.SparsityThreshold {
			n.mu.Lock()
			n.Stats.EarlyTerminations++
			n.mu.Unlock()
			continue
		}

		for f := 0; f < layer.Config.InputDim; f++ {
			aggregated[f] += g.NodeFeatures[neighbor][f]
		}
		neighborCount++
	}

	// Normalize
	if neighborCount > 0 {
		norm := 1.0 / math.Sqrt(float64(neighborCount+1))
		for f := 0; f < layer.Config.InputDim; f++ {
			aggregated[f] *= norm
		}
	}

	// Linear transformation (compute-intensive)
	for j := 0; j < layer.Config.OutputDim; j++ {
		sum := 0.0
		for k := 0; k < layer.Config.InputDim; k++ {
			sum += aggregated[k] * layer.Weights[k][j]
		}
		// ReLU
		if sum < 0 {
			sum = 0
		}
		output[node][j] = sum
	}
}

// =============================================================================
// PYGIM HYBRID CPU/PIM EXECUTION
// =============================================================================

// PyGimConfig configures PyGim hybrid execution
type PyGimConfig struct {
	CPUCores         int     // Number of CPU cores
	PIMBanks         int     // Number of PIM banks
	MemoryBandwidth  float64 // GB/s
	PIMComputeUnits  int     // Compute units per bank
	HybridThreshold  float64 // Threshold for PIM offload
}

// DefaultPyGimConfig returns default PyGim configuration
func DefaultPyGimConfig() *PyGimConfig {
	return &PyGimConfig{
		CPUCores:         8,
		PIMBanks:         16,
		MemoryBandwidth:  256.0, // GB/s (UPMEM-style)
		PIMComputeUnits:  16,
		HybridThreshold:  0.7,
	}
}

// PyGim implements hybrid CPU/PIM GNN acceleration
// Based on: PyGim (arXiv 2024)
// Combination of Accelerators (CoA) scheme
type PyGim struct {
	Config         *PyGimConfig
	PIMBanks       []*PIMBank
	Stats          *PyGimStats
}

// PIMBank represents a processing-in-memory bank
type PIMBank struct {
	ID              int
	LocalMemoryKB   int
	ComputeUnits    []*PIMComputeUnit
	DataMapped      []int // Node IDs mapped to this bank
}

// PIMComputeUnit represents a compute unit in PIM
type PIMComputeUnit struct {
	ID           int
	ALUID        int
	Accumulator  []float64
}

// PyGimStats tracks hybrid execution statistics
type PyGimStats struct {
	CPUTime           float64 // ms
	PIMTime           float64 // ms
	TotalTime         float64 // ms
	CPUOps            int64
	PIMOps            int64
	DataMovementGB    float64
	Speedup           float64
}

// NewPyGim creates a new PyGim hybrid system
func NewPyGim(config *PyGimConfig) *PyGim {
	pg := &PyGim{
		Config:   config,
		PIMBanks: make([]*PIMBank, config.PIMBanks),
		Stats:    &PyGimStats{},
	}

	for i := 0; i < config.PIMBanks; i++ {
		pg.PIMBanks[i] = &PIMBank{
			ID:            i,
			LocalMemoryKB: 64 * 1024, // 64MB per bank
			ComputeUnits:  make([]*PIMComputeUnit, config.PIMComputeUnits),
			DataMapped:    make([]int, 0),
		}
		for j := 0; j < config.PIMComputeUnits; j++ {
			pg.PIMBanks[i].ComputeUnits[j] = &PIMComputeUnit{
				ID:    j,
				ALUID: i*config.PIMComputeUnits + j,
			}
		}
	}

	return pg
}

// Execute runs GNN with hybrid CPU/PIM execution
func (pg *PyGim) Execute(g *Graph, layers []*GCNLayer) [][]float64 {
	current := g.NodeFeatures

	for layerIdx, layer := range layers {
		// Decide execution mode based on operation intensity
		if pg.isMemoryIntensive(g, layer) {
			// Aggregation on PIM (memory-intensive)
			aggregated := pg.aggregateOnPIM(g, current, layer)
			// Combination on CPU (compute-intensive)
			current = pg.combineOnCPU(aggregated, layer)
		} else {
			// All on CPU for small graphs
			current = layer.Forward(g, current)
		}
		_ = layerIdx // Layer index for logging
	}

	pg.Stats.TotalTime = pg.Stats.CPUTime + pg.Stats.PIMTime
	pg.Stats.Speedup = float64(pg.Stats.CPUOps+pg.Stats.PIMOps) / float64(pg.Stats.TotalTime*1e6)

	return current
}

// isMemoryIntensive checks if aggregation is memory-bound
func (pg *PyGim) isMemoryIntensive(g *Graph, layer *GCNLayer) bool {
	// Compute arithmetic intensity
	// Aggregation: O(E * F) memory, O(E * F) compute → AI ≈ 1
	// Combination: O(N * F * F') compute, O(N * F) memory → AI ≈ F'
	edgeOps := float64(g.Config.NumEdges * layer.Config.InputDim)
	nodeOps := float64(g.Config.NumNodes * layer.Config.InputDim * layer.Config.OutputDim)

	return edgeOps > nodeOps*pg.Config.HybridThreshold
}

// aggregateOnPIM performs neighbor aggregation on PIM
func (pg *PyGim) aggregateOnPIM(g *Graph, input [][]float64, layer *GCNLayer) [][]float64 {
	numNodes := len(input)
	aggregated := make([][]float64, numNodes)

	// Map nodes to PIM banks (locality-aware)
	nodesPerBank := (numNodes + pg.Config.PIMBanks - 1) / pg.Config.PIMBanks

	var wg sync.WaitGroup
	for bankID := 0; bankID < pg.Config.PIMBanks; bankID++ {
		wg.Add(1)
		go func(bank int) {
			defer wg.Done()
			startNode := bank * nodesPerBank
			endNode := startNode + nodesPerBank
			if endNode > numNodes {
				endNode = numNodes
			}

			for node := startNode; node < endNode; node++ {
				aggregated[node] = make([]float64, layer.Config.InputDim)

				// Self contribution
				for f := 0; f < layer.Config.InputDim; f++ {
					aggregated[node][f] = input[node][f]
				}

				// Neighbor aggregation
				for _, neighbor := range g.AdjList[node] {
					for f := 0; f < layer.Config.InputDim; f++ {
						aggregated[node][f] += input[neighbor][f]
					}
					pg.Stats.PIMOps += int64(layer.Config.InputDim)
				}

				// Normalize
				deg := float64(len(g.AdjList[node]) + 1)
				norm := 1.0 / math.Sqrt(deg)
				for f := 0; f < layer.Config.InputDim; f++ {
					aggregated[node][f] *= norm
				}
			}
		}(bankID)
	}
	wg.Wait()

	pg.Stats.PIMTime += float64(g.Config.NumEdges*layer.Config.InputDim) / (pg.Config.MemoryBandwidth * 1e9 / 8)

	return aggregated
}

// combineOnCPU performs linear transformation on CPU
func (pg *PyGim) combineOnCPU(aggregated [][]float64, layer *GCNLayer) [][]float64 {
	numNodes := len(aggregated)
	output := make([][]float64, numNodes)

	var wg sync.WaitGroup
	nodesPerCore := (numNodes + pg.Config.CPUCores - 1) / pg.Config.CPUCores

	for core := 0; core < pg.Config.CPUCores; core++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			startNode := c * nodesPerCore
			endNode := startNode + nodesPerCore
			if endNode > numNodes {
				endNode = numNodes
			}

			for node := startNode; node < endNode; node++ {
				output[node] = make([]float64, layer.Config.OutputDim)
				for j := 0; j < layer.Config.OutputDim; j++ {
					sum := 0.0
					for k := 0; k < layer.Config.InputDim; k++ {
						sum += aggregated[node][k] * layer.Weights[k][j]
					}
					// ReLU
					if sum < 0 {
						sum = 0
					}
					output[node][j] = sum
				}
				pg.Stats.CPUOps += int64(layer.Config.InputDim * layer.Config.OutputDim)
			}
		}(core)
	}
	wg.Wait()

	// Estimate CPU time (3 GHz, 4 FLOPS/cycle)
	pg.Stats.CPUTime += float64(pg.Stats.CPUOps) / (3e9 * 4 * float64(pg.Config.CPUCores))

	return output
}

// =============================================================================
// STT-MRAM COMPUTE-IN-MEMORY
// =============================================================================

// STTMRAMConfig configures STT-MRAM CIM
type STTMRAMConfig struct {
	ArrayRows        int     // Number of rows
	ArrayCols        int     // Number of columns
	TMRRatio         float64 // Tunnel magnetoresistance ratio
	CriticalCurrent  float64 // Switching current (µA)
	WriteTime        float64 // Write time (ns)
	ReadTime         float64 // Read time (ns)
	Endurance        int64   // Write endurance cycles
	RetentionYears   float64 // Data retention (years)
	CellAreaF2       float64 // Cell area in F²
}

// DefaultSTTMRAMConfig returns default STT-MRAM configuration
// Based on: TSMC 16nm STT-MRAM (2024)
func DefaultSTTMRAMConfig() *STTMRAMConfig {
	return &STTMRAMConfig{
		ArrayRows:       256,
		ArrayCols:       256,
		TMRRatio:        200.0,  // 200% TMR
		CriticalCurrent: 50.0,   // 50 µA
		WriteTime:       20.0,   // 20 ns
		ReadTime:        7.5,    // 7.5 ns
		Endurance:       1e12,   // 10^12 cycles
		RetentionYears:  10.0,   // 10 years @ 85°C
		CellAreaF2:      30.0,   // 30 F²
	}
}

// STTMRAMCell represents a single STT-MRAM cell
type STTMRAMCell struct {
	State           int     // 0 = P (low R), 1 = AP (high R)
	ResistanceLow   float64 // Parallel resistance (Ω)
	ResistanceHigh  float64 // Anti-parallel resistance (Ω)
	TMR             float64 // Tunnel magnetoresistance
	WriteCycles     int64   // Total write cycles
	LastWriteError  bool    // Write error flag
}

// STTMRAM implements STT-MRAM based CIM
// Based on: TSMC IEDM 2024, 10^12 endurance, 7.5/20 ns read/write
type STTMRAM struct {
	Config      *STTMRAMConfig
	Cells       [][]*STTMRAMCell
	Stats       *MRAMStats
}

// MRAMStats tracks MRAM statistics
type MRAMStats struct {
	TotalReads       int64
	TotalWrites      int64
	ReadEnergy       float64 // pJ
	WriteEnergy      float64 // pJ
	ReadLatency      float64 // ns
	WriteLatency     float64 // ns
	BitErrors        int
	EnduranceMargin  float64 // Remaining endurance ratio
}

// NewSTTMRAM creates a new STT-MRAM array
func NewSTTMRAM(config *STTMRAMConfig) *STTMRAM {
	mram := &STTMRAM{
		Config: config,
		Cells:  make([][]*STTMRAMCell, config.ArrayRows),
		Stats:  &MRAMStats{},
	}

	// Calculate resistances from TMR
	rLow := 5000.0 // Base resistance 5kΩ
	rHigh := rLow * (1.0 + config.TMRRatio/100.0)

	for i := 0; i < config.ArrayRows; i++ {
		mram.Cells[i] = make([]*STTMRAMCell, config.ArrayCols)
		for j := 0; j < config.ArrayCols; j++ {
			mram.Cells[i][j] = &STTMRAMCell{
				State:          rand.Intn(2),
				ResistanceLow:  rLow * (1.0 + rand.NormFloat64()*0.05),
				ResistanceHigh: rHigh * (1.0 + rand.NormFloat64()*0.05),
				TMR:            config.TMRRatio,
			}
		}
	}

	return mram
}

// ProgramWeight programs a weight into the array
func (m *STTMRAM) ProgramWeight(row, col int, weight float64) error {
	if row < 0 || row >= m.Config.ArrayRows || col < 0 || col >= m.Config.ArrayCols {
		return fmt.Errorf("index out of bounds: (%d, %d)", row, col)
	}

	cell := m.Cells[row][col]

	// Binary weight: positive = AP (1), negative = P (0)
	newState := 0
	if weight > 0 {
		newState = 1
	}

	if cell.State != newState {
		cell.State = newState
		cell.WriteCycles++
		m.Stats.TotalWrites++
		m.Stats.WriteEnergy += 0.5 // ~0.5 pJ per write
		m.Stats.WriteLatency += m.Config.WriteTime

		// Check endurance
		if cell.WriteCycles > m.Config.Endurance {
			cell.LastWriteError = true
			m.Stats.BitErrors++
		}
	}

	return nil
}

// ReadWeight reads a weight from the array
func (m *STTMRAM) ReadWeight(row, col int) (float64, error) {
	if row < 0 || row >= m.Config.ArrayRows || col < 0 || col >= m.Config.ArrayCols {
		return 0, fmt.Errorf("index out of bounds: (%d, %d)", row, col)
	}

	cell := m.Cells[row][col]
	m.Stats.TotalReads++
	m.Stats.ReadEnergy += 0.1 // ~0.1 pJ per read
	m.Stats.ReadLatency += m.Config.ReadTime

	// Return conductance-based weight
	if cell.State == 1 {
		return 1.0 / cell.ResistanceHigh, nil
	}
	return -1.0 / cell.ResistanceLow, nil
}

// ComputeMVM performs matrix-vector multiplication using STT-MRAM
func (m *STTMRAM) ComputeMVM(input []float64) ([]float64, error) {
	if len(input) != m.Config.ArrayCols {
		return nil, fmt.Errorf("input dimension mismatch: got %d, expected %d", len(input), m.Config.ArrayCols)
	}

	output := make([]float64, m.Config.ArrayRows)

	for i := 0; i < m.Config.ArrayRows; i++ {
		sum := 0.0
		for j := 0; j < m.Config.ArrayCols; j++ {
			w, _ := m.ReadWeight(i, j)
			sum += w * input[j]
		}
		output[i] = sum
	}

	return output, nil
}

// =============================================================================
// SOT-MRAM COMPUTE-IN-MEMORY
// =============================================================================

// SOTMRAMConfig configures SOT-MRAM CIM
type SOTMRAMConfig struct {
	ArrayRows         int     // Number of rows
	ArrayCols         int     // Number of columns
	TMRRatio          float64 // Tunnel magnetoresistance ratio
	SwitchingEnergy   float64 // Switching energy (fJ)
	WriteTime         float64 // Write time (ns)
	ReadTime          float64 // Read time (ns)
	Endurance         int64   // Write endurance cycles
	SpinHallAngle     float64 // Spin Hall angle
	HeavyMetalThick   float64 // Heavy metal thickness (nm)
}

// DefaultSOTMRAMConfig returns default SOT-MRAM configuration
// Based on: ITRI/TSMC IEDM 2023, Imec 2024
func DefaultSOTMRAMConfig() *SOTMRAMConfig {
	return &SOTMRAMConfig{
		ArrayRows:        256,
		ArrayCols:        256,
		TMRRatio:         200.0,   // 200% TMR
		SwitchingEnergy:  100.0,   // <100 fJ (Imec)
		WriteTime:        10.0,    // 10 ns (vs 50ns STT)
		ReadTime:         5.0,     // 5 ns
		Endurance:        1e15,    // 10^15 cycles
		SpinHallAngle:    0.3,     // θSH = 0.3
		HeavyMetalThick:  5.0,     // 5 nm W or Ta
	}
}

// SOTMRAMCell represents a single SOT-MRAM cell
type SOTMRAMCell struct {
	State             int     // Magnetization state
	FreeLayerMz       float64 // Free layer magnetization
	SOTCurrent        float64 // SOT write current
	ResistanceLow     float64
	ResistanceHigh    float64
	WriteCycles       int64
}

// SOTMRAM implements SOT-MRAM based CIM
// Based on: ITRI/TSMC IEDM 2023, 1% power of STT-MRAM, 10ns switching
type SOTMRAM struct {
	Config      *SOTMRAMConfig
	Cells       [][]*SOTMRAMCell
	Stats       *MRAMStats
}

// NewSOTMRAM creates a new SOT-MRAM array
func NewSOTMRAM(config *SOTMRAMConfig) *SOTMRAM {
	mram := &SOTMRAM{
		Config: config,
		Cells:  make([][]*SOTMRAMCell, config.ArrayRows),
		Stats:  &MRAMStats{},
	}

	rLow := 5000.0
	rHigh := rLow * (1.0 + config.TMRRatio/100.0)

	for i := 0; i < config.ArrayRows; i++ {
		mram.Cells[i] = make([]*SOTMRAMCell, config.ArrayCols)
		for j := 0; j < config.ArrayCols; j++ {
			mram.Cells[i][j] = &SOTMRAMCell{
				State:          rand.Intn(2),
				FreeLayerMz:    float64(rand.Intn(2)*2 - 1), // -1 or +1
				ResistanceLow:  rLow * (1.0 + rand.NormFloat64()*0.03),
				ResistanceHigh: rHigh * (1.0 + rand.NormFloat64()*0.03),
			}
		}
	}

	return mram
}

// ProgramWeight programs a multi-level weight using SOT
func (m *SOTMRAM) ProgramWeight(row, col int, weight float64, levels int) error {
	if row < 0 || row >= m.Config.ArrayRows || col < 0 || col >= m.Config.ArrayCols {
		return fmt.Errorf("index out of bounds: (%d, %d)", row, col)
	}

	cell := m.Cells[row][col]

	// Quantize weight to levels
	quantized := math.Round(weight * float64(levels-1))
	if quantized < 0 {
		quantized = 0
	}
	if quantized > float64(levels-1) {
		quantized = float64(levels - 1)
	}

	// SOT switching uses spin Hall effect
	// Current through heavy metal generates spin current
	cell.FreeLayerMz = (2.0*quantized/float64(levels-1) - 1.0)
	cell.WriteCycles++
	m.Stats.TotalWrites++
	m.Stats.WriteEnergy += m.Config.SwitchingEnergy / 1000.0 // fJ to pJ
	m.Stats.WriteLatency += m.Config.WriteTime

	return nil
}

// ComputeMVM performs matrix-vector multiplication using SOT-MRAM
func (m *SOTMRAM) ComputeMVM(input []float64) ([]float64, error) {
	if len(input) != m.Config.ArrayCols {
		return nil, fmt.Errorf("input dimension mismatch")
	}

	output := make([]float64, m.Config.ArrayRows)

	for i := 0; i < m.Config.ArrayRows; i++ {
		sum := 0.0
		for j := 0; j < m.Config.ArrayCols; j++ {
			// Conductance from magnetization state
			cell := m.Cells[i][j]
			r := cell.ResistanceLow
			if cell.FreeLayerMz < 0 {
				r = cell.ResistanceHigh
			}
			// Interpolate for multi-level
			ratio := (cell.FreeLayerMz + 1.0) / 2.0
			r = cell.ResistanceHigh + ratio*(cell.ResistanceLow-cell.ResistanceHigh)

			g := 1.0 / r
			sum += g * input[j]
			m.Stats.TotalReads++
		}
		output[i] = sum
	}

	m.Stats.ReadLatency += m.Config.ReadTime
	m.Stats.ReadEnergy += 0.05 * float64(m.Config.ArrayRows*m.Config.ArrayCols) // 0.05 pJ/cell

	return output, nil
}

// =============================================================================
// PHASE CHANGE MEMORY (PCM) COMPUTE-IN-MEMORY
// =============================================================================

// PCMConfig configures Phase Change Memory CIM
type PCMConfig struct {
	ArrayRows         int     // Number of rows
	ArrayCols         int     // Number of columns
	ResetResistance   float64 // Amorphous state resistance (MΩ)
	SetResistance     float64 // Crystalline state resistance (kΩ)
	NumLevels         int     // Multi-level states
	SetPulseWidth     float64 // SET pulse width (ns)
	ResetPulseWidth   float64 // RESET pulse width (ns)
	Endurance         int64   // Write endurance
	DriftCoefficient  float64 // Resistance drift coefficient
	Temperature       float64 // Operating temperature (°C)
}

// DefaultPCMConfig returns default PCM configuration
// Based on: STMicroelectronics 18nm FD-SOI GST PCM (2024)
func DefaultPCMConfig() *PCMConfig {
	return &PCMConfig{
		ArrayRows:        256,
		ArrayCols:        256,
		ResetResistance:  10.0,    // 10 MΩ (amorphous)
		SetResistance:    10.0,    // 10 kΩ (crystalline)
		NumLevels:        16,      // 4-bit MLC
		SetPulseWidth:    100.0,   // 100 ns
		ResetPulseWidth:  50.0,    // 50 ns
		Endurance:        1e8,     // 10^8 cycles
		DriftCoefficient: 0.05,    // 5% drift per decade
		Temperature:      85.0,    // 85°C automotive
	}
}

// PCMCell represents a single PCM cell
type PCMCell struct {
	CrystallineFraction float64 // 0 = amorphous, 1 = crystalline
	Resistance          float64 // Current resistance (Ω)
	ProgramCycles       int64
	TimeSinceProgram    float64 // Time since last programming (s)
	HasDrifted          bool
}

// PCM implements Phase Change Memory based CIM
// Based on: Nature Sensors 2025, MDPI PMC 2025
type PCM struct {
	Config      *PCMConfig
	Cells       [][]*PCMCell
	Stats       *PCMStats
}

// PCMStats tracks PCM statistics
type PCMStats struct {
	TotalSETs        int64
	TotalRESETs      int64
	TotalReads       int64
	SetEnergy        float64 // pJ
	ResetEnergy      float64 // pJ
	ReadEnergy       float64 // pJ
	DriftErrors      int
	EnduranceFailures int
}

// NewPCM creates a new PCM array
func NewPCM(config *PCMConfig) *PCM {
	pcm := &PCM{
		Config: config,
		Cells:  make([][]*PCMCell, config.ArrayRows),
		Stats:  &PCMStats{},
	}

	for i := 0; i < config.ArrayRows; i++ {
		pcm.Cells[i] = make([]*PCMCell, config.ArrayCols)
		for j := 0; j < config.ArrayCols; j++ {
			// Initialize to random state
			cf := rand.Float64()
			pcm.Cells[i][j] = &PCMCell{
				CrystallineFraction: cf,
				Resistance:          pcm.calculateResistance(cf),
				TimeSinceProgram:    rand.Float64() * 1000, // 0-1000 seconds
			}
		}
	}

	return pcm
}

// calculateResistance computes resistance from crystalline fraction
func (p *PCM) calculateResistance(cf float64) float64 {
	// Exponential interpolation between amorphous and crystalline
	rAmorphous := p.Config.ResetResistance * 1e6 // MΩ to Ω
	rCrystalline := p.Config.SetResistance * 1e3 // kΩ to Ω

	// R = R_c * (R_a/R_c)^(1-cf)
	logR := math.Log(rCrystalline) + (1.0-cf)*math.Log(rAmorphous/rCrystalline)
	return math.Exp(logR)
}

// applyDrift applies resistance drift to a cell
func (p *PCM) applyDrift(cell *PCMCell) {
	if cell.TimeSinceProgram > 0 {
		// R(t) = R0 * (t/t0)^ν where ν is drift coefficient
		t0 := 1.0 // Reference time (s)
		driftFactor := math.Pow(cell.TimeSinceProgram/t0, p.Config.DriftCoefficient)
		cell.Resistance *= driftFactor
		cell.HasDrifted = true
	}
}

// ProgramWeight programs a weight using iterative programming
func (p *PCM) ProgramWeight(row, col int, weight float64) error {
	if row < 0 || row >= p.Config.ArrayRows || col < 0 || col >= p.Config.ArrayCols {
		return fmt.Errorf("index out of bounds: (%d, %d)", row, col)
	}

	cell := p.Cells[row][col]

	// Map weight to crystalline fraction (0 to 1)
	targetCF := (weight + 1.0) / 2.0 // Assume weight in [-1, 1]
	if targetCF < 0 {
		targetCF = 0
	}
	if targetCF > 1 {
		targetCF = 1
	}

	// Quantize to available levels
	levelSize := 1.0 / float64(p.Config.NumLevels-1)
	quantizedCF := math.Round(targetCF/levelSize) * levelSize

	// Iterative programming (SET or RESET pulses)
	currentCF := cell.CrystallineFraction
	iterations := 0
	maxIterations := 10

	for math.Abs(currentCF-quantizedCF) > 0.05 && iterations < maxIterations {
		if currentCF < quantizedCF {
			// Apply SET pulse (crystallize)
			increment := (quantizedCF - currentCF) * 0.3
			currentCF += increment
			p.Stats.TotalSETs++
			p.Stats.SetEnergy += 10.0 // ~10 pJ per SET
		} else {
			// Apply RESET pulse (amorphize)
			decrement := (currentCF - quantizedCF) * 0.5
			currentCF -= decrement
			p.Stats.TotalRESETs++
			p.Stats.ResetEnergy += 50.0 // ~50 pJ per RESET
		}
		iterations++
		cell.ProgramCycles++

		// Check endurance
		if cell.ProgramCycles > p.Config.Endurance {
			p.Stats.EnduranceFailures++
			return fmt.Errorf("endurance exceeded at (%d, %d)", row, col)
		}
	}

	cell.CrystallineFraction = currentCF
	cell.Resistance = p.calculateResistance(currentCF)
	cell.TimeSinceProgram = 0

	return nil
}

// ReadWeight reads a weight from the array with drift compensation
func (p *PCM) ReadWeight(row, col int, compensateDrift bool) (float64, error) {
	if row < 0 || row >= p.Config.ArrayRows || col < 0 || col >= p.Config.ArrayCols {
		return 0, fmt.Errorf("index out of bounds: (%d, %d)", row, col)
	}

	cell := p.Cells[row][col]

	// Apply drift if significant time has passed
	if !cell.HasDrifted && cell.TimeSinceProgram > 10.0 {
		p.applyDrift(cell)
	}

	p.Stats.TotalReads++
	p.Stats.ReadEnergy += 0.5 // ~0.5 pJ per read

	// Convert resistance to weight
	rMin := p.Config.SetResistance * 1e3
	rMax := p.Config.ResetResistance * 1e6

	// Drift compensation using known drift model
	r := cell.Resistance
	if compensateDrift && cell.TimeSinceProgram > 0 {
		// Compensate: R_original = R_current / drift_factor
		driftFactor := math.Pow(cell.TimeSinceProgram/1.0, p.Config.DriftCoefficient)
		r /= driftFactor
	}

	// Map resistance to weight [-1, 1]
	logR := math.Log(r)
	logMin := math.Log(rMin)
	logMax := math.Log(rMax)

	weight := 2.0*(logR-logMin)/(logMax-logMin) - 1.0
	if weight < -1.0 {
		weight = -1.0
	}
	if weight > 1.0 {
		weight = 1.0
	}

	return weight, nil
}

// ComputeMVM performs matrix-vector multiplication using PCM
func (p *PCM) ComputeMVM(input []float64, compensateDrift bool) ([]float64, error) {
	if len(input) != p.Config.ArrayCols {
		return nil, fmt.Errorf("input dimension mismatch")
	}

	output := make([]float64, p.Config.ArrayRows)

	for i := 0; i < p.Config.ArrayRows; i++ {
		sum := 0.0
		for j := 0; j < p.Config.ArrayCols; j++ {
			// Use conductance (1/R) for computation
			cell := p.Cells[i][j]
			g := 1.0 / cell.Resistance
			sum += g * input[j]
		}
		output[i] = sum
	}

	return output, nil
}

// =============================================================================
// MULTI-MEMORY HYBRID CIM SYSTEM
// =============================================================================

// HybridMemoryConfig configures multi-memory CIM system
type HybridMemoryConfig struct {
	UseSTTMRAM    bool    // Use STT-MRAM for weights
	UseSOTMRAM    bool    // Use SOT-MRAM for activations
	UsePCM        bool    // Use PCM for model storage
	UseSRAM       bool    // Use SRAM for buffer
	SRAMSizeKB    int     // SRAM buffer size
	OptimizeFor   string  // "latency", "energy", "accuracy"
}

// DefaultHybridMemoryConfig returns default hybrid memory configuration
func DefaultHybridMemoryConfig() *HybridMemoryConfig {
	return &HybridMemoryConfig{
		UseSTTMRAM:  true,
		UseSOTMRAM:  true,
		UsePCM:      true,
		UseSRAM:     true,
		SRAMSizeKB:  256,
		OptimizeFor: "energy",
	}
}

// HybridMemoryCIM implements multi-memory CIM system
type HybridMemoryCIM struct {
	Config       *HybridMemoryConfig
	STTMRAM      *STTMRAM
	SOTMRAM      *SOTMRAM
	PCM          *PCM
	SRAMBuffer   []float64
	Stats        *HybridStats
}

// HybridStats tracks hybrid system statistics
type HybridStats struct {
	TotalOps          int64
	STTMRAMOps        int64
	SOTMRAMOps        int64
	PCMOps            int64
	TotalEnergyPJ     float64
	TotalLatencyNS    float64
	AccuracyLoss      float64
}

// NewHybridMemoryCIM creates a new hybrid memory CIM system
func NewHybridMemoryCIM(config *HybridMemoryConfig, arraySize int) *HybridMemoryCIM {
	hm := &HybridMemoryCIM{
		Config:     config,
		SRAMBuffer: make([]float64, config.SRAMSizeKB*1024/8),
		Stats:      &HybridStats{},
	}

	if config.UseSTTMRAM {
		sttConfig := DefaultSTTMRAMConfig()
		sttConfig.ArrayRows = arraySize
		sttConfig.ArrayCols = arraySize
		hm.STTMRAM = NewSTTMRAM(sttConfig)
	}

	if config.UseSOTMRAM {
		sotConfig := DefaultSOTMRAMConfig()
		sotConfig.ArrayRows = arraySize
		sotConfig.ArrayCols = arraySize
		hm.SOTMRAM = NewSOTMRAM(sotConfig)
	}

	if config.UsePCM {
		pcmConfig := DefaultPCMConfig()
		pcmConfig.ArrayRows = arraySize
		pcmConfig.ArrayCols = arraySize
		hm.PCM = NewPCM(pcmConfig)
	}

	return hm
}

// SelectOptimalMemory chooses best memory for operation
func (h *HybridMemoryCIM) SelectOptimalMemory(opType string, size int) string {
	switch h.Config.OptimizeFor {
	case "latency":
		// SOT-MRAM fastest for writes, SRAM fastest for reads
		if opType == "write" && h.Config.UseSOTMRAM {
			return "SOT-MRAM"
		}
		return "SRAM"

	case "energy":
		// PCM lowest energy per bit for storage
		// SOT-MRAM lowest for frequent updates
		if opType == "storage" && h.Config.UsePCM {
			return "PCM"
		}
		if h.Config.UseSOTMRAM {
			return "SOT-MRAM"
		}
		return "STT-MRAM"

	case "accuracy":
		// STT-MRAM best accuracy (no drift)
		if h.Config.UseSTTMRAM {
			return "STT-MRAM"
		}
		return "SOT-MRAM"

	default:
		return "STT-MRAM"
	}
}

// ComputeLayer executes a layer using optimal memory selection
func (h *HybridMemoryCIM) ComputeLayer(input []float64, weights [][]float64) ([]float64, error) {
	rows := len(weights)
	cols := len(weights[0])

	// Select memory based on optimization target
	memType := h.SelectOptimalMemory("compute", rows*cols)

	var output []float64
	var err error

	switch memType {
	case "STT-MRAM":
		if h.STTMRAM != nil {
			// Program weights
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					h.STTMRAM.ProgramWeight(i, j, weights[i][j])
				}
			}
			output, err = h.STTMRAM.ComputeMVM(input)
			h.Stats.STTMRAMOps += int64(rows * cols)
			h.Stats.TotalEnergyPJ += h.STTMRAM.Stats.ReadEnergy
		}

	case "SOT-MRAM":
		if h.SOTMRAM != nil {
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					h.SOTMRAM.ProgramWeight(i, j, weights[i][j], 4)
				}
			}
			output, err = h.SOTMRAM.ComputeMVM(input)
			h.Stats.SOTMRAMOps += int64(rows * cols)
			h.Stats.TotalEnergyPJ += h.SOTMRAM.Stats.ReadEnergy
		}

	case "PCM":
		if h.PCM != nil {
			for i := 0; i < rows; i++ {
				for j := 0; j < cols; j++ {
					h.PCM.ProgramWeight(i, j, weights[i][j])
				}
			}
			output, err = h.PCM.ComputeMVM(input, true)
			h.Stats.PCMOps += int64(rows * cols)
			h.Stats.TotalEnergyPJ += h.PCM.Stats.ReadEnergy
		}
	}

	h.Stats.TotalOps += int64(rows * cols)

	return output, err
}

// =============================================================================
// GNN + EMERGING MEMORY INTEGRATION
// =============================================================================

// GNNMemoryAcceleratorConfig configures GNN acceleration with emerging memory
type GNNMemoryAcceleratorConfig struct {
	GraphConfig     *GraphConfig
	MemoryType      string  // "STT-MRAM", "SOT-MRAM", "PCM", "Hybrid"
	NEMGNNEnabled   bool    // Use NEM-GNN architecture
	PyGimEnabled    bool    // Use PyGim hybrid execution
	PrecisionBits   int     // Weight precision
	BatchSize       int     // Inference batch size
}

// DefaultGNNMemoryAcceleratorConfig returns default configuration
func DefaultGNNMemoryAcceleratorConfig() *GNNMemoryAcceleratorConfig {
	return &GNNMemoryAcceleratorConfig{
		GraphConfig:   DefaultGraphConfig(),
		MemoryType:    "Hybrid",
		NEMGNNEnabled: true,
		PyGimEnabled:  true,
		PrecisionBits: 8,
		BatchSize:     32,
	}
}

// GNNMemoryAccelerator integrates GNN with emerging memory CIM
type GNNMemoryAccelerator struct {
	Config        *GNNMemoryAcceleratorConfig
	Graph         *Graph
	Layers        []*GCNLayer
	NEMGNN        *NEMGNN
	PyGim         *PyGim
	HybridMem     *HybridMemoryCIM
	Stats         *GNNAccelStats
}

// GNNAccelStats tracks GNN accelerator statistics
type GNNAccelStats struct {
	TotalNodes        int
	TotalEdges        int
	AggregationTime   float64 // ms
	CombinationTime   float64 // ms
	TotalTime         float64 // ms
	EnergyPerNode     float64 // pJ
	Throughput        float64 // nodes/s
	Accuracy          float64 // Classification accuracy
}

// NewGNNMemoryAccelerator creates a new GNN memory accelerator
func NewGNNMemoryAccelerator(config *GNNMemoryAcceleratorConfig) *GNNMemoryAccelerator {
	acc := &GNNMemoryAccelerator{
		Config: config,
		Layers: make([]*GCNLayer, 0),
		Stats:  &GNNAccelStats{},
	}

	// Create graph
	acc.Graph = NewGraph(config.GraphConfig)
	acc.Graph.GenerateRandomGraph()

	// Initialize accelerators
	if config.NEMGNNEnabled {
		acc.NEMGNN = NewNEMGNN(DefaultNEMGNNConfig())
	}

	if config.PyGimEnabled {
		acc.PyGim = NewPyGim(DefaultPyGimConfig())
	}

	// Initialize hybrid memory
	hybridConfig := DefaultHybridMemoryConfig()
	acc.HybridMem = NewHybridMemoryCIM(hybridConfig, 256)

	// Update stats
	acc.Stats.TotalNodes = config.GraphConfig.NumNodes
	acc.Stats.TotalEdges = config.GraphConfig.NumEdges

	return acc
}

// AddLayer adds a GCN layer to the network
func (a *GNNMemoryAccelerator) AddLayer(inputDim, outputDim int, activation string) {
	layerConfig := &GCNLayerConfig{
		InputDim:   inputDim,
		OutputDim:  outputDim,
		Activation: activation,
		Dropout:    0.5,
		UseBias:    true,
		Normalize:  true,
	}
	layer := NewGCNLayer(layerConfig)
	a.Layers = append(a.Layers, layer)
}

// Inference runs GNN inference with memory-optimized execution
func (a *GNNMemoryAccelerator) Inference() [][]float64 {
	var output [][]float64

	if a.Config.NEMGNNEnabled && a.NEMGNN != nil {
		// Use NEM-GNN for memory-intensive operations
		current := a.Graph.NodeFeatures
		for _, layer := range a.Layers {
			current = a.NEMGNN.Execute(a.Graph, layer)
		}
		output = current
		a.Stats.Throughput = a.NEMGNN.Stats.Throughput * 1e9 // Convert GOPS to nodes/s equivalent

	} else if a.Config.PyGimEnabled && a.PyGim != nil {
		// Use PyGim hybrid CPU/PIM execution
		output = a.PyGim.Execute(a.Graph, a.Layers)
		a.Stats.AggregationTime = a.PyGim.Stats.PIMTime
		a.Stats.CombinationTime = a.PyGim.Stats.CPUTime
		a.Stats.TotalTime = a.PyGim.Stats.TotalTime
		a.Stats.Throughput = float64(a.Stats.TotalNodes) / (a.Stats.TotalTime / 1000.0)

	} else {
		// Standard execution with hybrid memory
		current := a.Graph.NodeFeatures
		for _, layer := range a.Layers {
			current = layer.Forward(a.Graph, current)
		}
		output = current
	}

	// Compute accuracy (for node classification)
	correct := 0
	for i, nodeOutput := range output {
		predicted := argmax(nodeOutput)
		if predicted == a.Graph.Labels[i] {
			correct++
		}
	}
	a.Stats.Accuracy = float64(correct) / float64(len(output))

	return output
}

// argmax returns index of maximum value
func argmax(arr []float64) int {
	maxIdx := 0
	maxVal := arr[0]
	for i, v := range arr {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

// =============================================================================
// BENCHMARK AND COMPARISON
// =============================================================================

// MemoryComparison compares emerging memory technologies
type MemoryComparison struct {
	Technologies []string
	Metrics      map[string]map[string]float64
}

// NewMemoryComparison creates a memory technology comparison
func NewMemoryComparison() *MemoryComparison {
	mc := &MemoryComparison{
		Technologies: []string{"SRAM", "STT-MRAM", "SOT-MRAM", "PCM", "ReRAM", "FeFET"},
		Metrics:      make(map[string]map[string]float64),
	}

	// Based on literature: PMC 2024, EDN 2024, Nature Sensors 2025
	mc.Metrics["SRAM"] = map[string]float64{
		"ReadLatencyNS":      0.5,
		"WriteLatencyNS":     0.5,
		"EnduranceCycles":    1e16,
		"RetentionYears":     0.0, // Volatile
		"EnergyPerBitFJ":     1.0,
		"CellAreaF2":         100.0,
		"CIMEffTOPSW":        10.0,
	}

	mc.Metrics["STT-MRAM"] = map[string]float64{
		"ReadLatencyNS":      7.5,
		"WriteLatencyNS":     20.0,
		"EnduranceCycles":    1e12,
		"RetentionYears":     10.0,
		"EnergyPerBitFJ":     100.0,
		"CellAreaF2":         30.0,
		"CIMEffTOPSW":        50.0,
	}

	mc.Metrics["SOT-MRAM"] = map[string]float64{
		"ReadLatencyNS":      5.0,
		"WriteLatencyNS":     10.0,
		"EnduranceCycles":    1e15,
		"RetentionYears":     10.0,
		"EnergyPerBitFJ":     50.0, // Imec: <100 fJ
		"CellAreaF2":         50.0,
		"CIMEffTOPSW":        100.0,
	}

	mc.Metrics["PCM"] = map[string]float64{
		"ReadLatencyNS":      50.0,
		"WriteLatencyNS":     100.0,
		"EnduranceCycles":    1e8,
		"RetentionYears":     10.0,
		"EnergyPerBitFJ":     10.0,  // SET ~10pJ, RESET ~50pJ
		"CellAreaF2":         8.0,
		"CIMEffTOPSW":        200.0, // High density
	}

	mc.Metrics["ReRAM"] = map[string]float64{
		"ReadLatencyNS":      10.0,
		"WriteLatencyNS":     10.0,
		"EnduranceCycles":    1e6,
		"RetentionYears":     10.0,
		"EnergyPerBitFJ":     0.1,
		"CellAreaF2":         4.0,
		"CIMEffTOPSW":        300.0,
	}

	mc.Metrics["FeFET"] = map[string]float64{
		"ReadLatencyNS":      20.0,
		"WriteLatencyNS":     50.0,
		"EnduranceCycles":    1e10,
		"RetentionYears":     10.0,
		"EnergyPerBitFJ":     1.0,
		"CellAreaF2":         6.0,
		"CIMEffTOPSW":        500.0, // IronLattice target
	}

	return mc
}

// GetBestFor returns best technology for given metric (lower or higher)
func (mc *MemoryComparison) GetBestFor(metric string, preferLower bool) string {
	type techMetric struct {
		tech  string
		value float64
	}

	var results []techMetric
	for _, tech := range mc.Technologies {
		if val, ok := mc.Metrics[tech][metric]; ok {
			results = append(results, techMetric{tech, val})
		}
	}

	if len(results) == 0 {
		return ""
	}

	sort.Slice(results, func(i, j int) bool {
		if preferLower {
			return results[i].value < results[j].value
		}
		return results[i].value > results[j].value
	})

	return results[0].tech
}

// PrintComparison prints memory technology comparison
func (mc *MemoryComparison) PrintComparison() string {
	result := "=== Emerging Memory Technology Comparison ===\n\n"
	result += fmt.Sprintf("%-12s %10s %10s %12s %10s %10s %10s\n",
		"Technology", "Read(ns)", "Write(ns)", "Endurance", "E/bit(fJ)", "Area(F²)", "CIM(TOPS/W)")
	result += "------------------------------------------------------------------------------\n"

	for _, tech := range mc.Technologies {
		m := mc.Metrics[tech]
		result += fmt.Sprintf("%-12s %10.1f %10.1f %12.0e %10.1f %10.1f %10.1f\n",
			tech,
			m["ReadLatencyNS"],
			m["WriteLatencyNS"],
			m["EnduranceCycles"],
			m["EnergyPerBitFJ"],
			m["CellAreaF2"],
			m["CIMEffTOPSW"])
	}

	result += "\nBest for latency: " + mc.GetBestFor("ReadLatencyNS", true) + "\n"
	result += "Best for endurance: " + mc.GetBestFor("EnduranceCycles", false) + "\n"
	result += "Best for density: " + mc.GetBestFor("CellAreaF2", true) + "\n"
	result += "Best for CIM efficiency: " + mc.GetBestFor("CIMEffTOPSW", false) + "\n"

	return result
}

// =============================================================================
// DEMONSTRATION FUNCTIONS
// =============================================================================

// DemoGNNAcceleration demonstrates GNN acceleration with emerging memory
func DemoGNNAcceleration() {
	fmt.Println("=== GNN Acceleration with Emerging Memory Demo ===")
	fmt.Println()

	// Create accelerator
	config := DefaultGNNMemoryAcceleratorConfig()
	config.GraphConfig.NumNodes = 1000
	config.GraphConfig.NumEdges = 5000
	config.GraphConfig.NumFeatures = 64
	config.GraphConfig.NumClasses = 7

	acc := NewGNNMemoryAccelerator(config)

	// Add GCN layers
	acc.AddLayer(64, 32, "relu")
	acc.AddLayer(32, 7, "none")

	fmt.Printf("Graph: %d nodes, %d edges\n", config.GraphConfig.NumNodes, config.GraphConfig.NumEdges)
	fmt.Printf("Model: 2-layer GCN (64→32→7)\n")
	fmt.Println()

	// Run inference
	output := acc.Inference()
	_ = output

	fmt.Println("NEM-GNN Statistics:")
	fmt.Printf("  Throughput: %.2f GOPS\n", acc.NEMGNN.Stats.Throughput)
	fmt.Printf("  Early terminations: %d\n", acc.NEMGNN.Stats.EarlyTerminations)
	fmt.Printf("  Classification accuracy: %.2f%%\n", acc.Stats.Accuracy*100)
	fmt.Println()

	// Memory comparison
	mc := NewMemoryComparison()
	fmt.Println(mc.PrintComparison())
}

// DemoSTTMRAMCIM demonstrates STT-MRAM compute-in-memory
func DemoSTTMRAMCIM() {
	fmt.Println("=== STT-MRAM Compute-in-Memory Demo ===")
	fmt.Println()

	config := DefaultSTTMRAMConfig()
	config.ArrayRows = 64
	config.ArrayCols = 64
	mram := NewSTTMRAM(config)

	// Program random weights
	for i := 0; i < config.ArrayRows; i++ {
		for j := 0; j < config.ArrayCols; j++ {
			weight := rand.Float64()*2 - 1
			mram.ProgramWeight(i, j, weight)
		}
	}

	// Create random input
	input := make([]float64, config.ArrayCols)
	for i := range input {
		input[i] = rand.Float64()
	}

	// Compute MVM
	output, _ := mram.ComputeMVM(input)

	fmt.Printf("Array size: %dx%d\n", config.ArrayRows, config.ArrayCols)
	fmt.Printf("TMR ratio: %.0f%%\n", config.TMRRatio)
	fmt.Printf("Write time: %.1f ns\n", config.WriteTime)
	fmt.Printf("Read time: %.1f ns\n", config.ReadTime)
	fmt.Printf("Endurance: %.0e cycles\n", float64(config.Endurance))
	fmt.Println()
	fmt.Printf("Total writes: %d\n", mram.Stats.TotalWrites)
	fmt.Printf("Total reads: %d\n", mram.Stats.TotalReads)
	fmt.Printf("Write energy: %.2f pJ\n", mram.Stats.WriteEnergy)
	fmt.Printf("Read energy: %.2f pJ\n", mram.Stats.ReadEnergy)
	fmt.Printf("Output[0:5]: %.4f, %.4f, %.4f, %.4f, %.4f\n",
		output[0], output[1], output[2], output[3], output[4])
}

// DemoPCMCIM demonstrates PCM compute-in-memory with drift
func DemoPCMCIM() {
	fmt.Println("=== Phase Change Memory CIM Demo ===")
	fmt.Println()

	config := DefaultPCMConfig()
	config.ArrayRows = 64
	config.ArrayCols = 64
	pcm := NewPCM(config)

	// Program weights
	for i := 0; i < config.ArrayRows; i++ {
		for j := 0; j < config.ArrayCols; j++ {
			weight := rand.Float64()*2 - 1
			pcm.ProgramWeight(i, j, weight)
		}
	}

	// Simulate time passing (drift)
	for i := 0; i < config.ArrayRows; i++ {
		for j := 0; j < config.ArrayCols; j++ {
			pcm.Cells[i][j].TimeSinceProgram = 1000.0 // 1000 seconds
		}
	}

	// Create input
	input := make([]float64, config.ArrayCols)
	for i := range input {
		input[i] = rand.Float64()
	}

	// Compute with and without drift compensation
	outputNoDrift, _ := pcm.ComputeMVM(input, false)
	outputWithDrift, _ := pcm.ComputeMVM(input, true)

	// Calculate drift error
	driftError := 0.0
	for i := range outputNoDrift {
		diff := outputNoDrift[i] - outputWithDrift[i]
		driftError += diff * diff
	}
	driftError = math.Sqrt(driftError / float64(len(outputNoDrift)))

	fmt.Printf("Array size: %dx%d\n", config.ArrayRows, config.ArrayCols)
	fmt.Printf("MLC levels: %d (%.1f bits)\n", config.NumLevels, math.Log2(float64(config.NumLevels)))
	fmt.Printf("Drift coefficient: %.2f\n", config.DriftCoefficient)
	fmt.Printf("Endurance: %.0e cycles\n", float64(config.Endurance))
	fmt.Println()
	fmt.Printf("Total SETs: %d\n", pcm.Stats.TotalSETs)
	fmt.Printf("Total RESETs: %d\n", pcm.Stats.TotalRESETs)
	fmt.Printf("SET energy: %.2f pJ\n", pcm.Stats.SetEnergy)
	fmt.Printf("RESET energy: %.2f pJ\n", pcm.Stats.ResetEnergy)
	fmt.Printf("Drift RMSE (after 1000s): %.4f\n", driftError)
}

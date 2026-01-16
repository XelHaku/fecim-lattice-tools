// compiler_codesign_cim.go - CIM Compiler and Hardware-Software Co-Design
// Implements multi-level compilation stack, crossbar mapping, and HW-aware NAS
// Based on CIM-MLC, COMPASS, and hardware-software co-design research

package layers

import (
	"fmt"
	"math"
	"sort"
)

// ============================================================================
// CIM MULTI-LEVEL COMPILER (CIM-MLC)
// ============================================================================

// CIMMLCConfig configures the multi-level compilation stack
type CIMMLCConfig struct {
	// Frontend configuration
	FrontendOptPasses []string // e.g., "ConstFold", "DeadCodeElim", "GraphFusion"

	// Middle-end configuration
	TilingStrategy   string  // "row", "column", "2d", "block"
	PartitionMethod  string  // "greedy", "dp", "ilp", "compass"
	MaxCrossbarSize  int     // Maximum crossbar dimension (e.g., 128)
	MinUtilization   float64 // Minimum crossbar utilization threshold

	// Backend configuration
	TargetDevice     string // "fefet", "reram", "pcm", "mram"
	ADCBits          int    // ADC resolution
	DACBits          int    // DAC resolution
	WeightBits       int    // Weight precision
	ActivationBits   int    // Activation precision

	// Dataflow configuration
	DataflowMode     string // "weight_stationary", "input_stationary", "output_stationary"
	BufferSizeKB     int    // On-chip buffer size in KB
	BandwidthGBps    float64 // External memory bandwidth
}

// DefaultCIMMLCConfig returns default compiler configuration
func DefaultCIMMLCConfig() *CIMMLCConfig {
	return &CIMMLCConfig{
		FrontendOptPasses: []string{"ConstFold", "DeadCodeElim", "GraphFusion", "BatchNormFold"},
		TilingStrategy:    "2d",
		PartitionMethod:   "compass",
		MaxCrossbarSize:   128,
		MinUtilization:    0.5,
		TargetDevice:      "fefet",
		ADCBits:           6,
		DACBits:           8,
		WeightBits:        6,
		DACBits:           8,
		ActivationBits:    8,
		DataflowMode:      "weight_stationary",
		BufferSizeKB:      256,
		BandwidthGBps:     25.6,
	}
}

// GraphIR represents intermediate representation of neural network
type GraphIR struct {
	Nodes      []*IRNode
	Edges      []*IREdge
	InputNodes []int
	OutputNodes []int
	Attributes map[string]interface{}
}

// IRNode represents a node in the computation graph
type IRNode struct {
	ID           int
	OpType       string   // "conv2d", "matmul", "relu", "add", etc.
	InputShape   [][]int  // Input tensor shapes
	OutputShape  []int    // Output tensor shape
	Weights      []float64 // Weight values if any
	Attributes   map[string]interface{}
	Predecessors []int
	Successors   []int

	// CIM-specific attributes
	CrossbarTiles  []*CrossbarTile
	PartitionID    int
	MemoryLevel    string // "crossbar", "buffer", "dram"
}

// IREdge represents data flow between nodes
type IREdge struct {
	SourceID    int
	TargetID    int
	TensorShape []int
	DataType    string
	IsWeightEdge bool
}

// CrossbarTile represents a mapped crossbar array
type CrossbarTile struct {
	TileID       int
	Rows         int
	Cols         int
	RowOffset    int  // Offset in original weight matrix
	ColOffset    int
	Weights      [][]float64
	Utilization  float64
	ADCResolution int
	DACResolution int
}

// CIMMLCompiler implements multi-level CIM compilation
type CIMMLCompiler struct {
	Config     *CIMMLCConfig
	GraphIR    *GraphIR
	TilePool   []*CrossbarTile
	Partitions []*Partition
	Schedule   *ExecutionSchedule
	Stats      *CompilerStats
}

// CompilerStats tracks compilation statistics
type CompilerStats struct {
	TotalNodes        int
	TotalTiles        int
	AverageUtilization float64
	TotalWeightSize   int64
	TotalActivationSize int64
	EstimatedLatency  float64 // microseconds
	EstimatedEnergy   float64 // nJ
	MemoryAccesses    int64
	CompressionRatio  float64
}

// NewCIMMLCompiler creates a new multi-level compiler
func NewCIMMLCompiler(config *CIMMLCConfig) *CIMMLCompiler {
	return &CIMMLCompiler{
		Config:     config,
		TilePool:   make([]*CrossbarTile, 0),
		Partitions: make([]*Partition, 0),
		Stats:      &CompilerStats{},
	}
}

// Compile performs full compilation pipeline
func (c *CIMMLCompiler) Compile(graph *GraphIR) error {
	c.GraphIR = graph
	c.Stats.TotalNodes = len(graph.Nodes)

	// Frontend: Graph optimizations
	if err := c.runFrontendPasses(); err != nil {
		return fmt.Errorf("frontend optimization failed: %v", err)
	}

	// Middle-end: Tiling and partitioning
	if err := c.runMiddleEndPasses(); err != nil {
		return fmt.Errorf("middle-end processing failed: %v", err)
	}

	// Backend: Hardware mapping
	if err := c.runBackendPasses(); err != nil {
		return fmt.Errorf("backend mapping failed: %v", err)
	}

	// Generate execution schedule
	if err := c.generateSchedule(); err != nil {
		return fmt.Errorf("scheduling failed: %v", err)
	}

	return nil
}

// runFrontendPasses executes graph optimization passes
func (c *CIMMLCompiler) runFrontendPasses() error {
	for _, passName := range c.Config.FrontendOptPasses {
		switch passName {
		case "ConstFold":
			c.constantFolding()
		case "DeadCodeElim":
			c.deadCodeElimination()
		case "GraphFusion":
			c.graphFusion()
		case "BatchNormFold":
			c.batchNormFolding()
		}
	}
	return nil
}

// constantFolding folds constant operations
func (c *CIMMLCompiler) constantFolding() {
	for _, node := range c.GraphIR.Nodes {
		if node.OpType == "const_mul" || node.OpType == "const_add" {
			// Evaluate constant expressions
			if allInputsConstant(node, c.GraphIR) {
				node.OpType = "const"
				node.Attributes["folded"] = true
			}
		}
	}
}

// deadCodeElimination removes unused nodes
func (c *CIMMLCompiler) deadCodeElimination() {
	used := make(map[int]bool)
	// Mark outputs as used
	for _, outID := range c.GraphIR.OutputNodes {
		c.markUsed(outID, used)
	}
	// Remove unused nodes
	newNodes := make([]*IRNode, 0)
	for _, node := range c.GraphIR.Nodes {
		if used[node.ID] {
			newNodes = append(newNodes, node)
		}
	}
	c.GraphIR.Nodes = newNodes
}

func (c *CIMMLCompiler) markUsed(nodeID int, used map[int]bool) {
	if used[nodeID] {
		return
	}
	used[nodeID] = true
	for _, node := range c.GraphIR.Nodes {
		if node.ID == nodeID {
			for _, predID := range node.Predecessors {
				c.markUsed(predID, used)
			}
		}
	}
}

// graphFusion fuses compatible operations
func (c *CIMMLCompiler) graphFusion() {
	// Conv + BN + ReLU fusion
	for i := 0; i < len(c.GraphIR.Nodes); i++ {
		node := c.GraphIR.Nodes[i]
		if node.OpType == "conv2d" {
			// Check for fusible successors
			fusedOps := []string{node.OpType}
			currentNode := node

			for _, succID := range currentNode.Successors {
				succ := c.findNode(succID)
				if succ == nil {
					continue
				}
				if succ.OpType == "batchnorm" || succ.OpType == "relu" {
					fusedOps = append(fusedOps, succ.OpType)
					succ.Attributes["fused"] = true
				}
			}

			if len(fusedOps) > 1 {
				node.Attributes["fused_ops"] = fusedOps
			}
		}
	}
}

// batchNormFolding folds BatchNorm into preceding Conv/Linear
func (c *CIMMLCompiler) batchNormFolding() {
	for _, node := range c.GraphIR.Nodes {
		if node.OpType == "batchnorm" {
			if fused, ok := node.Attributes["fused"].(bool); ok && fused {
				continue
			}
			// Get preceding conv/linear node
			for _, predID := range node.Predecessors {
				pred := c.findNode(predID)
				if pred != nil && (pred.OpType == "conv2d" || pred.OpType == "linear") {
					// Fold BN parameters into weights
					node.Attributes["folded_into"] = predID
				}
			}
		}
	}
}

func (c *CIMMLCompiler) findNode(id int) *IRNode {
	for _, node := range c.GraphIR.Nodes {
		if node.ID == id {
			return node
		}
	}
	return nil
}

// ============================================================================
// CROSSBAR TILING AND PARTITIONING (COMPASS-INSPIRED)
// ============================================================================

// Partition represents a set of layers that fit on-chip
type Partition struct {
	PartitionID   int
	Nodes         []int   // Node IDs in this partition
	Tiles         []*CrossbarTile
	BufferUsageKB float64
	Latency       float64
	EnergyNJ      float64
}

// runMiddleEndPasses handles tiling and partitioning
func (c *CIMMLCompiler) runMiddleEndPasses() error {
	// Step 1: Tile large weight matrices
	for _, node := range c.GraphIR.Nodes {
		if node.OpType == "conv2d" || node.OpType == "linear" || node.OpType == "matmul" {
			tiles := c.tileWeightMatrix(node)
			node.CrossbarTiles = tiles
			c.TilePool = append(c.TilePool, tiles...)
		}
	}

	// Step 2: Partition graph for on-chip execution
	switch c.Config.PartitionMethod {
	case "compass":
		c.compassPartitioning()
	case "greedy":
		c.greedyPartitioning()
	case "dp":
		c.dpPartitioning()
	default:
		c.greedyPartitioning()
	}

	c.Stats.TotalTiles = len(c.TilePool)
	c.computeUtilizationStats()

	return nil
}

// tileWeightMatrix splits a weight matrix into crossbar-sized tiles
func (c *CIMMLCompiler) tileWeightMatrix(node *IRNode) []*CrossbarTile {
	tiles := make([]*CrossbarTile, 0)

	// Get weight matrix dimensions
	rows, cols := c.getWeightDimensions(node)
	maxSize := c.Config.MaxCrossbarSize

	// 2D tiling
	tileID := len(c.TilePool)
	for rowStart := 0; rowStart < rows; rowStart += maxSize {
		for colStart := 0; colStart < cols; colStart += maxSize {
			rowEnd := min(rowStart+maxSize, rows)
			colEnd := min(colStart+maxSize, cols)

			tileRows := rowEnd - rowStart
			tileCols := colEnd - colStart

			// Calculate utilization
			utilization := float64(tileRows*tileCols) / float64(maxSize*maxSize)

			if utilization >= c.Config.MinUtilization {
				tile := &CrossbarTile{
					TileID:        tileID,
					Rows:          tileRows,
					Cols:          tileCols,
					RowOffset:     rowStart,
					ColOffset:     colStart,
					Utilization:   utilization,
					ADCResolution: c.Config.ADCBits,
					DACResolution: c.Config.DACBits,
				}
				tiles = append(tiles, tile)
				tileID++
			}
		}
	}

	return tiles
}

func (c *CIMMLCompiler) getWeightDimensions(node *IRNode) (int, int) {
	if dims, ok := node.Attributes["weight_shape"].([]int); ok && len(dims) >= 2 {
		return dims[0], dims[1]
	}
	// Infer from output shape
	if len(node.OutputShape) >= 2 {
		return node.OutputShape[len(node.OutputShape)-2], node.OutputShape[len(node.OutputShape)-1]
	}
	return 128, 128 // default
}

// compassPartitioning implements COMPASS-style optimal partitioning
func (c *CIMMLCompiler) compassPartitioning() {
	// Build dependency graph
	nodeOrder := c.topologicalSort()

	// Dynamic programming for optimal partitioning
	n := len(nodeOrder)
	if n == 0 {
		return
	}

	// dp[i] = minimum cost to process nodes 0..i
	dp := make([]float64, n+1)
	parent := make([]int, n+1)

	for i := range dp {
		dp[i] = math.Inf(1)
		parent[i] = -1
	}
	dp[0] = 0

	bufferLimit := float64(c.Config.BufferSizeKB)

	for i := 1; i <= n; i++ {
		// Try all possible partition endpoints
		bufferUsed := 0.0
		for j := i - 1; j >= 0; j-- {
			nodeID := nodeOrder[j]
			node := c.findNode(nodeID)
			if node == nil {
				continue
			}

			// Estimate buffer requirement for this node
			bufferUsed += c.estimateBufferUsage(node)

			if bufferUsed > bufferLimit {
				break
			}

			// Cost = external memory accesses + latency
			partitionCost := c.estimatePartitionCost(nodeOrder[j:i])
			totalCost := dp[j] + partitionCost

			if totalCost < dp[i] {
				dp[i] = totalCost
				parent[i] = j
			}
		}
	}

	// Reconstruct partitions
	c.reconstructPartitions(nodeOrder, parent)
}

func (c *CIMMLCompiler) topologicalSort() []int {
	inDegree := make(map[int]int)
	for _, node := range c.GraphIR.Nodes {
		inDegree[node.ID] = len(node.Predecessors)
	}

	queue := make([]int, 0)
	for _, node := range c.GraphIR.Nodes {
		if inDegree[node.ID] == 0 {
			queue = append(queue, node.ID)
		}
	}

	result := make([]int, 0)
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		result = append(result, nodeID)

		node := c.findNode(nodeID)
		if node != nil {
			for _, succID := range node.Successors {
				inDegree[succID]--
				if inDegree[succID] == 0 {
					queue = append(queue, succID)
				}
			}
		}
	}

	return result
}

func (c *CIMMLCompiler) estimateBufferUsage(node *IRNode) float64 {
	// Estimate in KB
	activationBytes := 1
	for _, dim := range node.OutputShape {
		activationBytes *= dim
	}
	activationBytes *= c.Config.ActivationBits / 8
	return float64(activationBytes) / 1024.0
}

func (c *CIMMLCompiler) estimatePartitionCost(nodeIDs []int) float64 {
	cost := 0.0
	for _, nodeID := range nodeIDs {
		node := c.findNode(nodeID)
		if node == nil {
			continue
		}

		// Tile execution cost
		for _, tile := range node.CrossbarTiles {
			// MVM latency estimation
			mvmLatency := float64(tile.Rows) * 0.01 // 10ns per row
			cost += mvmLatency
		}

		// External memory cost for partition boundary
		cost += c.estimateBufferUsage(node) * 0.1 // 100ns/KB for DRAM
	}
	return cost
}

func (c *CIMMLCompiler) reconstructPartitions(nodeOrder []int, parent []int) {
	partitions := make([][]int, 0)
	i := len(nodeOrder)

	for i > 0 {
		j := parent[i]
		if j < 0 {
			break
		}
		partitions = append([][]int{nodeOrder[j:i]}, partitions...)
		i = j
	}

	for pid, nodes := range partitions {
		partition := &Partition{
			PartitionID: pid,
			Nodes:       nodes,
			Tiles:       make([]*CrossbarTile, 0),
		}

		for _, nodeID := range nodes {
			node := c.findNode(nodeID)
			if node != nil {
				node.PartitionID = pid
				partition.Tiles = append(partition.Tiles, node.CrossbarTiles...)
				partition.BufferUsageKB += c.estimateBufferUsage(node)
			}
		}

		c.Partitions = append(c.Partitions, partition)
	}
}

func (c *CIMMLCompiler) greedyPartitioning() {
	// Simple greedy: pack nodes until buffer full
	currentPartition := &Partition{
		PartitionID: 0,
		Nodes:       make([]int, 0),
		Tiles:       make([]*CrossbarTile, 0),
	}

	bufferUsed := 0.0
	bufferLimit := float64(c.Config.BufferSizeKB)

	for _, node := range c.GraphIR.Nodes {
		nodeBuffer := c.estimateBufferUsage(node)

		if bufferUsed+nodeBuffer > bufferLimit && len(currentPartition.Nodes) > 0 {
			c.Partitions = append(c.Partitions, currentPartition)
			currentPartition = &Partition{
				PartitionID: len(c.Partitions),
				Nodes:       make([]int, 0),
				Tiles:       make([]*CrossbarTile, 0),
			}
			bufferUsed = 0
		}

		currentPartition.Nodes = append(currentPartition.Nodes, node.ID)
		currentPartition.Tiles = append(currentPartition.Tiles, node.CrossbarTiles...)
		currentPartition.BufferUsageKB += nodeBuffer
		node.PartitionID = currentPartition.PartitionID
		bufferUsed += nodeBuffer
	}

	if len(currentPartition.Nodes) > 0 {
		c.Partitions = append(c.Partitions, currentPartition)
	}
}

func (c *CIMMLCompiler) dpPartitioning() {
	// DP-based partitioning (simplified version of COMPASS)
	c.compassPartitioning()
}

func (c *CIMMLCompiler) computeUtilizationStats() {
	totalUtil := 0.0
	for _, tile := range c.TilePool {
		totalUtil += tile.Utilization
	}
	if len(c.TilePool) > 0 {
		c.Stats.AverageUtilization = totalUtil / float64(len(c.TilePool))
	}
}

// ============================================================================
// BACKEND: HARDWARE MAPPING
// ============================================================================

// runBackendPasses generates hardware-specific code
func (c *CIMMLCompiler) runBackendPasses() error {
	// Map to target device characteristics
	switch c.Config.TargetDevice {
	case "fefet":
		c.mapToFeFET()
	case "reram":
		c.mapToReRAM()
	case "pcm":
		c.mapToPCM()
	case "mram":
		c.mapToMRAM()
	}

	// Apply quantization
	c.applyQuantization()

	// Optimize dataflow
	c.optimizeDataflow()

	return nil
}

func (c *CIMMLCompiler) mapToFeFET() {
	// FeFET-specific: 6-bit MLC programming
	for _, tile := range c.TilePool {
		// FeFET write characteristics
		tile.ADCResolution = min(c.Config.ADCBits, 6)
		tile.DACResolution = min(c.Config.DACBits, 8)
	}
}

func (c *CIMMLCompiler) mapToReRAM() {
	// ReRAM: higher variability, 4-bit typical
	for _, tile := range c.TilePool {
		tile.ADCResolution = min(c.Config.ADCBits, 4)
		tile.DACResolution = min(c.Config.DACBits, 6)
	}
}

func (c *CIMMLCompiler) mapToPCM() {
	// PCM: drift compensation needed
	for _, tile := range c.TilePool {
		tile.ADCResolution = min(c.Config.ADCBits, 4)
		tile.DACResolution = min(c.Config.DACBits, 4)
	}
}

func (c *CIMMLCompiler) mapToMRAM() {
	// MRAM: binary or 2-bit typical
	for _, tile := range c.TilePool {
		tile.ADCResolution = min(c.Config.ADCBits, 2)
		tile.DACResolution = min(c.Config.DACBits, 4)
	}
}

func (c *CIMMLCompiler) applyQuantization() {
	weightBits := c.Config.WeightBits
	scale := math.Pow(2, float64(weightBits)) - 1

	for _, tile := range c.TilePool {
		if tile.Weights != nil {
			for i := range tile.Weights {
				for j := range tile.Weights[i] {
					// Quantize to target precision
					tile.Weights[i][j] = math.Round(tile.Weights[i][j]*scale) / scale
				}
			}
		}
	}
}

func (c *CIMMLCompiler) optimizeDataflow() {
	switch c.Config.DataflowMode {
	case "weight_stationary":
		c.optimizeWeightStationary()
	case "input_stationary":
		c.optimizeInputStationary()
	case "output_stationary":
		c.optimizeOutputStationary()
	}
}

func (c *CIMMLCompiler) optimizeWeightStationary() {
	// Weights stay in crossbars, inputs stream through
	for _, partition := range c.Partitions {
		partition.Latency = c.estimateWeightStationaryLatency(partition)
		partition.EnergyNJ = c.estimateWeightStationaryEnergy(partition)
	}
}

func (c *CIMMLCompiler) estimateWeightStationaryLatency(p *Partition) float64 {
	latency := 0.0
	for _, tile := range p.Tiles {
		// MVM latency: proportional to ADC conversion time
		adcTime := float64(tile.Rows) * 0.01 * float64(tile.ADCResolution) // 10ns * bits
		latency += adcTime
	}
	return latency
}

func (c *CIMMLCompiler) estimateWeightStationaryEnergy(p *Partition) float64 {
	energy := 0.0
	for _, tile := range p.Tiles {
		// MVM energy: ~1fJ/MAC for FeFET
		macs := tile.Rows * tile.Cols
		energy += float64(macs) * 0.001 // fJ to nJ
	}
	return energy
}

func (c *CIMMLCompiler) optimizeInputStationary() {
	for _, partition := range c.Partitions {
		partition.Latency = c.estimateWeightStationaryLatency(partition) * 1.2
		partition.EnergyNJ = c.estimateWeightStationaryEnergy(partition) * 0.9
	}
}

func (c *CIMMLCompiler) optimizeOutputStationary() {
	for _, partition := range c.Partitions {
		partition.Latency = c.estimateWeightStationaryLatency(partition) * 1.1
		partition.EnergyNJ = c.estimateWeightStationaryEnergy(partition) * 0.95
	}
}

// ============================================================================
// EXECUTION SCHEDULE GENERATION
// ============================================================================

// ExecutionSchedule defines the order of operations
type ExecutionSchedule struct {
	Steps          []*ScheduleStep
	TotalLatencyUS float64
	TotalEnergyNJ  float64
	PeakBufferKB   float64
}

// ScheduleStep represents one execution step
type ScheduleStep struct {
	StepID       int
	PartitionID  int
	TileIDs      []int
	Operation    string // "load_weights", "mvm", "activation", "writeback"
	LatencyUS    float64
	EnergyNJ     float64
	BufferUsedKB float64
}

func (c *CIMMLCompiler) generateSchedule() error {
	c.Schedule = &ExecutionSchedule{
		Steps: make([]*ScheduleStep, 0),
	}

	stepID := 0
	peakBuffer := 0.0

	for _, partition := range c.Partitions {
		// Step 1: Load weights (if first use)
		loadStep := &ScheduleStep{
			StepID:      stepID,
			PartitionID: partition.PartitionID,
			Operation:   "load_weights",
			LatencyUS:   c.estimateWeightLoadLatency(partition),
			EnergyNJ:    c.estimateWeightLoadEnergy(partition),
		}
		c.Schedule.Steps = append(c.Schedule.Steps, loadStep)
		stepID++

		// Step 2: Execute MVM
		mvmStep := &ScheduleStep{
			StepID:       stepID,
			PartitionID:  partition.PartitionID,
			TileIDs:      c.getTileIDs(partition),
			Operation:    "mvm",
			LatencyUS:    partition.Latency,
			EnergyNJ:     partition.EnergyNJ,
			BufferUsedKB: partition.BufferUsageKB,
		}
		c.Schedule.Steps = append(c.Schedule.Steps, mvmStep)
		stepID++

		if partition.BufferUsageKB > peakBuffer {
			peakBuffer = partition.BufferUsageKB
		}

		// Step 3: Activation (if needed)
		activationStep := &ScheduleStep{
			StepID:      stepID,
			PartitionID: partition.PartitionID,
			Operation:   "activation",
			LatencyUS:   0.1, // Digital activation is fast
			EnergyNJ:    0.01 * float64(len(partition.Tiles)),
		}
		c.Schedule.Steps = append(c.Schedule.Steps, activationStep)
		stepID++

		c.Schedule.TotalLatencyUS += loadStep.LatencyUS + mvmStep.LatencyUS + activationStep.LatencyUS
		c.Schedule.TotalEnergyNJ += loadStep.EnergyNJ + mvmStep.EnergyNJ + activationStep.EnergyNJ
	}

	c.Schedule.PeakBufferKB = peakBuffer
	c.Stats.EstimatedLatency = c.Schedule.TotalLatencyUS
	c.Stats.EstimatedEnergy = c.Schedule.TotalEnergyNJ

	return nil
}

func (c *CIMMLCompiler) estimateWeightLoadLatency(p *Partition) float64 {
	// Estimate based on weight size and bandwidth
	weightBytes := 0
	for _, tile := range p.Tiles {
		weightBytes += tile.Rows * tile.Cols * c.Config.WeightBits / 8
	}
	return float64(weightBytes) / (c.Config.BandwidthGBps * 1e6) // microseconds
}

func (c *CIMMLCompiler) estimateWeightLoadEnergy(p *Partition) float64 {
	// DRAM access energy: ~20pJ/bit
	weightBits := 0
	for _, tile := range p.Tiles {
		weightBits += tile.Rows * tile.Cols * c.Config.WeightBits
	}
	return float64(weightBits) * 0.00002 // pJ to nJ
}

func (c *CIMMLCompiler) getTileIDs(p *Partition) []int {
	ids := make([]int, len(p.Tiles))
	for i, tile := range p.Tiles {
		ids[i] = tile.TileID
	}
	return ids
}

// ============================================================================
// HARDWARE-AWARE NEURAL ARCHITECTURE SEARCH (HW-NAS)
// ============================================================================

// HWNASConfig configures hardware-aware NAS
type HWNASConfig struct {
	SearchSpace       *NASSearchSpace
	TargetLatencyUS   float64
	TargetEnergyNJ    float64
	AccuracyWeight    float64 // Weight for accuracy in fitness
	LatencyWeight     float64 // Weight for latency in fitness
	EnergyWeight      float64 // Weight for energy in fitness
	PopulationSize    int
	NumGenerations    int
	MutationRate      float64
	CrossoverRate     float64
}

// NASSearchSpace defines the architecture search space
type NASSearchSpace struct {
	LayerTypes       []string   // Available layer types
	KernelSizes      []int      // Available kernel sizes for conv
	ChannelOptions   []int      // Available channel counts
	ActivationTypes  []string   // Available activations
	MaxDepth         int        // Maximum network depth
	MinDepth         int        // Minimum network depth
}

// NASArchitecture represents a candidate architecture
type NASArchitecture struct {
	Layers          []*NASLayer
	Fitness         float64
	Accuracy        float64
	LatencyUS       float64
	EnergyNJ        float64
	TotalParams     int64
	CrossbarTiles   int
}

// NASLayer represents a layer in the search space
type NASLayer struct {
	LayerType    string
	KernelSize   int
	InChannels   int
	OutChannels  int
	Activation   string
	Skip         bool // Skip connection
}

// HWNASEngine implements hardware-aware NAS
type HWNASEngine struct {
	Config      *HWNASConfig
	Compiler    *CIMMLCompiler
	Population  []*NASArchitecture
	BestArch    *NASArchitecture
	Generation  int
}

// NewHWNASEngine creates a new HW-NAS engine
func NewHWNASEngine(config *HWNASConfig, compiler *CIMMLCompiler) *HWNASEngine {
	return &HWNASEngine{
		Config:     config,
		Compiler:   compiler,
		Population: make([]*NASArchitecture, 0),
	}
}

// Search performs hardware-aware architecture search
func (nas *HWNASEngine) Search() *NASArchitecture {
	// Initialize population
	nas.initializePopulation()

	// Evolutionary search
	for gen := 0; gen < nas.Config.NumGenerations; gen++ {
		nas.Generation = gen

		// Evaluate fitness
		nas.evaluatePopulation()

		// Selection
		parents := nas.selectParents()

		// Crossover and mutation
		offspring := nas.generateOffspring(parents)

		// Replace population
		nas.Population = nas.selectSurvivors(offspring)

		// Update best
		nas.updateBest()
	}

	return nas.BestArch
}

func (nas *HWNASEngine) initializePopulation() {
	for i := 0; i < nas.Config.PopulationSize; i++ {
		arch := nas.randomArchitecture()
		nas.Population = append(nas.Population, arch)
	}
}

func (nas *HWNASEngine) randomArchitecture() *NASArchitecture {
	ss := nas.Config.SearchSpace
	depth := ss.MinDepth + int(float64(ss.MaxDepth-ss.MinDepth)*randFloat())

	arch := &NASArchitecture{
		Layers: make([]*NASLayer, depth),
	}

	inChannels := 3 // RGB input
	for i := 0; i < depth; i++ {
		layerType := ss.LayerTypes[randInt(len(ss.LayerTypes))]
		outChannels := ss.ChannelOptions[randInt(len(ss.ChannelOptions))]

		layer := &NASLayer{
			LayerType:   layerType,
			KernelSize:  ss.KernelSizes[randInt(len(ss.KernelSizes))],
			InChannels:  inChannels,
			OutChannels: outChannels,
			Activation:  ss.ActivationTypes[randInt(len(ss.ActivationTypes))],
			Skip:        randFloat() > 0.7,
		}
		arch.Layers[i] = layer
		inChannels = outChannels
	}

	return arch
}

func (nas *HWNASEngine) evaluatePopulation() {
	for _, arch := range nas.Population {
		nas.evaluateArchitecture(arch)
	}
}

func (nas *HWNASEngine) evaluateArchitecture(arch *NASArchitecture) {
	// Convert to GraphIR and compile
	graph := nas.architectureToGraph(arch)
	nas.Compiler.Compile(graph)

	// Get hardware metrics
	arch.LatencyUS = nas.Compiler.Stats.EstimatedLatency
	arch.EnergyNJ = nas.Compiler.Stats.EstimatedEnergy
	arch.CrossbarTiles = nas.Compiler.Stats.TotalTiles

	// Estimate accuracy (simplified proxy)
	arch.Accuracy = nas.estimateAccuracy(arch)

	// Calculate total parameters
	arch.TotalParams = nas.countParameters(arch)

	// Compute fitness (multi-objective)
	nas.computeFitness(arch)
}

func (nas *HWNASEngine) architectureToGraph(arch *NASArchitecture) *GraphIR {
	graph := &GraphIR{
		Nodes:       make([]*IRNode, len(arch.Layers)),
		Edges:       make([]*IREdge, 0),
		InputNodes:  []int{0},
		OutputNodes: []int{len(arch.Layers) - 1},
	}

	inputSize := []int{1, 3, 32, 32} // CIFAR-like

	for i, layer := range arch.Layers {
		node := &IRNode{
			ID:         i,
			OpType:     layer.LayerType,
			Attributes: make(map[string]interface{}),
		}

		// Set weight shape based on layer type
		switch layer.LayerType {
		case "conv2d":
			node.Attributes["weight_shape"] = []int{
				layer.OutChannels, layer.InChannels, layer.KernelSize, layer.KernelSize,
			}
		case "linear":
			inFeatures := inputSize[1] * inputSize[2] * inputSize[3]
			node.Attributes["weight_shape"] = []int{inFeatures, layer.OutChannels}
		}

		if i > 0 {
			node.Predecessors = []int{i - 1}
		}
		if i < len(arch.Layers)-1 {
			node.Successors = []int{i + 1}
		}

		graph.Nodes[i] = node
	}

	return graph
}

func (nas *HWNASEngine) estimateAccuracy(arch *NASArchitecture) float64 {
	// Simplified accuracy proxy based on model capacity
	capacity := math.Log10(float64(arch.TotalParams + 1))
	baseAccuracy := 0.7 + 0.2*(1-math.Exp(-capacity/6))

	// Penalize extreme quantization
	if nas.Compiler.Config.WeightBits < 4 {
		baseAccuracy *= 0.95
	}

	return baseAccuracy
}

func (nas *HWNASEngine) countParameters(arch *NASArchitecture) int64 {
	total := int64(0)
	for _, layer := range arch.Layers {
		switch layer.LayerType {
		case "conv2d":
			total += int64(layer.OutChannels * layer.InChannels * layer.KernelSize * layer.KernelSize)
		case "linear":
			total += int64(layer.InChannels * layer.OutChannels)
		}
	}
	return total
}

func (nas *HWNASEngine) computeFitness(arch *NASArchitecture) {
	cfg := nas.Config

	// Normalize metrics
	latencyPenalty := 0.0
	if arch.LatencyUS > cfg.TargetLatencyUS {
		latencyPenalty = (arch.LatencyUS - cfg.TargetLatencyUS) / cfg.TargetLatencyUS
	}

	energyPenalty := 0.0
	if arch.EnergyNJ > cfg.TargetEnergyNJ {
		energyPenalty = (arch.EnergyNJ - cfg.TargetEnergyNJ) / cfg.TargetEnergyNJ
	}

	// Multi-objective fitness
	arch.Fitness = cfg.AccuracyWeight*arch.Accuracy -
		cfg.LatencyWeight*latencyPenalty -
		cfg.EnergyWeight*energyPenalty
}

func (nas *HWNASEngine) selectParents() []*NASArchitecture {
	// Tournament selection
	parents := make([]*NASArchitecture, nas.Config.PopulationSize/2)
	for i := range parents {
		a := nas.Population[randInt(len(nas.Population))]
		b := nas.Population[randInt(len(nas.Population))]
		if a.Fitness > b.Fitness {
			parents[i] = a
		} else {
			parents[i] = b
		}
	}
	return parents
}

func (nas *HWNASEngine) generateOffspring(parents []*NASArchitecture) []*NASArchitecture {
	offspring := make([]*NASArchitecture, nas.Config.PopulationSize)

	for i := 0; i < nas.Config.PopulationSize; i++ {
		p1 := parents[randInt(len(parents))]
		p2 := parents[randInt(len(parents))]

		var child *NASArchitecture
		if randFloat() < nas.Config.CrossoverRate {
			child = nas.crossover(p1, p2)
		} else {
			child = nas.copyArchitecture(p1)
		}

		if randFloat() < nas.Config.MutationRate {
			nas.mutate(child)
		}

		offspring[i] = child
	}

	return offspring
}

func (nas *HWNASEngine) crossover(p1, p2 *NASArchitecture) *NASArchitecture {
	// Single-point crossover
	minLen := min(len(p1.Layers), len(p2.Layers))
	crossPoint := randInt(minLen)

	child := &NASArchitecture{
		Layers: make([]*NASLayer, 0),
	}

	for i := 0; i < crossPoint; i++ {
		child.Layers = append(child.Layers, nas.copyLayer(p1.Layers[i]))
	}
	for i := crossPoint; i < len(p2.Layers); i++ {
		child.Layers = append(child.Layers, nas.copyLayer(p2.Layers[i]))
	}

	return child
}

func (nas *HWNASEngine) mutate(arch *NASArchitecture) {
	ss := nas.Config.SearchSpace
	idx := randInt(len(arch.Layers))
	layer := arch.Layers[idx]

	// Random mutation type
	switch randInt(4) {
	case 0: // Change layer type
		layer.LayerType = ss.LayerTypes[randInt(len(ss.LayerTypes))]
	case 1: // Change kernel size
		layer.KernelSize = ss.KernelSizes[randInt(len(ss.KernelSizes))]
	case 2: // Change channels
		layer.OutChannels = ss.ChannelOptions[randInt(len(ss.ChannelOptions))]
	case 3: // Toggle skip connection
		layer.Skip = !layer.Skip
	}
}

func (nas *HWNASEngine) copyArchitecture(arch *NASArchitecture) *NASArchitecture {
	copy := &NASArchitecture{
		Layers: make([]*NASLayer, len(arch.Layers)),
	}
	for i, layer := range arch.Layers {
		copy.Layers[i] = nas.copyLayer(layer)
	}
	return copy
}

func (nas *HWNASEngine) copyLayer(layer *NASLayer) *NASLayer {
	return &NASLayer{
		LayerType:   layer.LayerType,
		KernelSize:  layer.KernelSize,
		InChannels:  layer.InChannels,
		OutChannels: layer.OutChannels,
		Activation:  layer.Activation,
		Skip:        layer.Skip,
	}
}

func (nas *HWNASEngine) selectSurvivors(offspring []*NASArchitecture) []*NASArchitecture {
	// Elitism: keep best from population + offspring
	combined := append(nas.Population, offspring...)
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Fitness > combined[j].Fitness
	})
	return combined[:nas.Config.PopulationSize]
}

func (nas *HWNASEngine) updateBest() {
	for _, arch := range nas.Population {
		if nas.BestArch == nil || arch.Fitness > nas.BestArch.Fitness {
			nas.BestArch = arch
		}
	}
}

// Helper functions
func allInputsConstant(node *IRNode, graph *GraphIR) bool {
	for _, predID := range node.Predecessors {
		for _, n := range graph.Nodes {
			if n.ID == predID && n.OpType != "const" {
				return false
			}
		}
	}
	return true
}

// Simple pseudo-random for deterministic results
var randState uint64 = 12345

func randFloat() float64 {
	randState = randState*6364136223846793005 + 1442695040888963407
	return float64(randState>>33) / float64(1<<31)
}

func randInt(n int) int {
	if n <= 0 {
		return 0
	}
	return int(randFloat() * float64(n))
}

// ============================================================================
// DATAFLOW OPTIMIZATION
// ============================================================================

// DataflowOptimizer optimizes data movement
type DataflowOptimizer struct {
	Config         *DataflowConfig
	MemoryHierarchy *MemoryHierarchy
	Stats          *DataflowStats
}

// DataflowConfig configures dataflow optimization
type DataflowConfig struct {
	Mode            string  // "weight_stationary", "input_stationary", "output_stationary", "row_stationary"
	TileScheduling  string  // "sequential", "pipelined", "parallel"
	Prefetching     bool
	DoubleBuffering bool
	ReuseFactor     int     // Target data reuse factor
}

// MemoryHierarchy models the memory system
type MemoryHierarchy struct {
	Levels      []*MemoryLevel
	TotalSizeKB int
}

// MemoryLevel represents one level of memory
type MemoryLevel struct {
	Name         string
	SizeKB       int
	BandwidthGBps float64
	LatencyNS    float64
	EnergyPerBit float64 // pJ/bit
}

// DataflowStats tracks dataflow optimization results
type DataflowStats struct {
	DataReuse        float64
	MemoryAccesses   int64
	EstimatedLatency float64
	EstimatedEnergy  float64
	BufferUtilization float64
}

// NewDataflowOptimizer creates a dataflow optimizer
func NewDataflowOptimizer(config *DataflowConfig) *DataflowOptimizer {
	return &DataflowOptimizer{
		Config: config,
		MemoryHierarchy: &MemoryHierarchy{
			Levels: []*MemoryLevel{
				{Name: "crossbar", SizeKB: 1, BandwidthGBps: 1000, LatencyNS: 1, EnergyPerBit: 0.001},
				{Name: "buffer", SizeKB: 256, BandwidthGBps: 100, LatencyNS: 5, EnergyPerBit: 0.01},
				{Name: "dram", SizeKB: 1048576, BandwidthGBps: 25.6, LatencyNS: 100, EnergyPerBit: 0.02},
			},
		},
		Stats: &DataflowStats{},
	}
}

// Optimize performs dataflow optimization
func (opt *DataflowOptimizer) Optimize(schedule *ExecutionSchedule) {
	switch opt.Config.Mode {
	case "weight_stationary":
		opt.optimizeWeightStationary(schedule)
	case "input_stationary":
		opt.optimizeInputStationary(schedule)
	case "output_stationary":
		opt.optimizeOutputStationary(schedule)
	case "row_stationary":
		opt.optimizeRowStationary(schedule)
	}

	if opt.Config.Prefetching {
		opt.enablePrefetching(schedule)
	}

	if opt.Config.DoubleBuffering {
		opt.enableDoubleBuffering(schedule)
	}
}

func (opt *DataflowOptimizer) optimizeWeightStationary(schedule *ExecutionSchedule) {
	// Weights stay in crossbar, maximize input/output reuse
	for _, step := range schedule.Steps {
		if step.Operation == "mvm" {
			// Batch multiple inputs for same weights
			step.LatencyUS *= 0.9 // 10% improvement from batching
			opt.Stats.DataReuse += float64(len(step.TileIDs))
		}
	}
}

func (opt *DataflowOptimizer) optimizeInputStationary(schedule *ExecutionSchedule) {
	// Inputs stay in buffer, weights stream through
	for _, step := range schedule.Steps {
		if step.Operation == "load_weights" {
			step.LatencyUS *= 1.2 // More weight loading
		}
		if step.Operation == "mvm" {
			step.EnergyNJ *= 0.85 // Less input movement
		}
	}
}

func (opt *DataflowOptimizer) optimizeOutputStationary(schedule *ExecutionSchedule) {
	// Partial sums accumulate in place
	for _, step := range schedule.Steps {
		if step.Operation == "mvm" {
			step.EnergyNJ *= 0.9 // Less partial sum movement
		}
	}
}

func (opt *DataflowOptimizer) optimizeRowStationary(schedule *ExecutionSchedule) {
	// NVDLA-style row-stationary dataflow
	for _, step := range schedule.Steps {
		if step.Operation == "mvm" {
			step.LatencyUS *= 0.85 // Better parallelism
			step.EnergyNJ *= 0.88 // Better data reuse
		}
	}
}

func (opt *DataflowOptimizer) enablePrefetching(schedule *ExecutionSchedule) {
	// Overlap data loading with computation
	for i := 1; i < len(schedule.Steps); i++ {
		if schedule.Steps[i].Operation == "load_weights" {
			// Overlap with previous MVM
			schedule.Steps[i].LatencyUS *= 0.5
		}
	}
}

func (opt *DataflowOptimizer) enableDoubleBuffering(schedule *ExecutionSchedule) {
	// Hide data transfer latency
	for _, step := range schedule.Steps {
		if step.Operation == "load_weights" {
			step.LatencyUS *= 0.3 // Most loading hidden
		}
	}
	// Buffer utilization increases
	opt.Stats.BufferUtilization = 0.85
}

// ============================================================================
// PERFORMANCE ESTIMATION
// ============================================================================

// PerformanceEstimator estimates hardware performance
type PerformanceEstimator struct {
	DeviceParams *DeviceParameters
	CircuitParams *CircuitParameters
}

// DeviceParameters contains device-level parameters
type DeviceParameters struct {
	DeviceType      string
	ConductanceMin  float64 // Gmin in µS
	ConductanceMax  float64 // Gmax in µS
	OnOffRatio      float64
	WriteLatencyNS  float64
	ReadLatencyNS   float64
	WriteEnergyFJ   float64
	ReadEnergyFJ    float64
	EnduranceCycles int64
	RetentionYears  float64
}

// CircuitParameters contains circuit-level parameters
type CircuitParameters struct {
	ArrayRows       int
	ArrayCols       int
	ADCResolution   int
	DACResolution   int
	ADCLatencyNS    float64
	ADCEnergyFJ     float64
	PeripheralArea  float64 // mm²
}

// NewPerformanceEstimator creates a performance estimator
func NewPerformanceEstimator(device string) *PerformanceEstimator {
	est := &PerformanceEstimator{
		DeviceParams:  getDeviceParams(device),
		CircuitParams: getDefaultCircuitParams(),
	}
	return est
}

func getDeviceParams(device string) *DeviceParameters {
	switch device {
	case "fefet":
		return &DeviceParameters{
			DeviceType:      "fefet",
			ConductanceMin:  1.0,
			ConductanceMax:  100.0,
			OnOffRatio:      100,
			WriteLatencyNS:  100,
			ReadLatencyNS:   10,
			WriteEnergyFJ:   1000,
			ReadEnergyFJ:    1,
			EnduranceCycles: 1e12,
			RetentionYears:  10,
		}
	case "reram":
		return &DeviceParameters{
			DeviceType:      "reram",
			ConductanceMin:  0.1,
			ConductanceMax:  10.0,
			OnOffRatio:      100,
			WriteLatencyNS:  50,
			ReadLatencyNS:   5,
			WriteEnergyFJ:   100,
			ReadEnergyFJ:    0.1,
			EnduranceCycles: 1e6,
			RetentionYears:  10,
		}
	default:
		return &DeviceParameters{
			DeviceType:      "generic",
			ConductanceMin:  1.0,
			ConductanceMax:  10.0,
			OnOffRatio:      10,
			WriteLatencyNS:  100,
			ReadLatencyNS:   10,
			WriteEnergyFJ:   100,
			ReadEnergyFJ:    1,
			EnduranceCycles: 1e9,
			RetentionYears:  5,
		}
	}
}

func getDefaultCircuitParams() *CircuitParameters {
	return &CircuitParameters{
		ArrayRows:      128,
		ArrayCols:      128,
		ADCResolution:  6,
		DACResolution:  8,
		ADCLatencyNS:   50,
		ADCEnergyFJ:    1000,
		PeripheralArea: 0.01,
	}
}

// EstimateTOPS estimates throughput in TOPS
func (est *PerformanceEstimator) EstimateTOPS(numArrays int) float64 {
	// MACs per cycle
	macsPerCycle := est.CircuitParams.ArrayRows * est.CircuitParams.ArrayCols * numArrays

	// Cycle time limited by ADC
	cycleTimeNS := est.CircuitParams.ADCLatencyNS

	// TOPS = MACs/cycle / cycle_time / 1e12
	tops := float64(macsPerCycle) / cycleTimeNS / 1000.0
	return tops
}

// EstimateTOPSPerWatt estimates energy efficiency
func (est *PerformanceEstimator) EstimateTOPSPerWatt(numArrays int) float64 {
	tops := est.EstimateTOPS(numArrays)

	// Power estimation
	arrayPower := est.estimateArrayPower(numArrays)
	adcPower := est.estimateADCPower(numArrays)
	totalPowerW := arrayPower + adcPower

	if totalPowerW == 0 {
		return 0
	}
	return tops / totalPowerW
}

func (est *PerformanceEstimator) estimateArrayPower(numArrays int) float64 {
	// Read energy per MAC
	readEnergyPerMAC := est.DeviceParams.ReadEnergyFJ * 1e-15 // Convert to J
	macsPerSecond := float64(est.CircuitParams.ArrayRows*est.CircuitParams.ArrayCols*numArrays) * 1e9 / est.CircuitParams.ADCLatencyNS
	return readEnergyPerMAC * macsPerSecond // Watts
}

func (est *PerformanceEstimator) estimateADCPower(numArrays int) float64 {
	// ADC energy per conversion
	adcEnergyJ := float64(est.CircuitParams.ADCEnergyFJ) * 1e-15
	conversionsPerSecond := float64(est.CircuitParams.ArrayCols*numArrays) * 1e9 / est.CircuitParams.ADCLatencyNS
	return adcEnergyJ * conversionsPerSecond // Watts
}

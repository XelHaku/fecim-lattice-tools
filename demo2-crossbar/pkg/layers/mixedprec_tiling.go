// Package layers provides neural network layer implementations for CIM simulation.
// This file implements mixed-precision quantization and weight tiling for CIM.
//
// Mixed-Precision Quantization:
// - Hessian-based sensitivity analysis (HAWQ, HMQAT)
// - Layer-wise bit allocation via Pareto optimization
// - CIM²PQ arraywise precision assignment
// - Support for W4A8, W8A16, W2A4 configurations
//
// Weight Tiling:
// - COMPASS-style crossbar partitioning
// - Im2col mapping for convolutions
// - Multi-crossbar weight distribution
// - Tile dependency scheduling
//
// References:
// - HAWQ: Hessian AWare Quantization (arXiv 1905.03696)
// - HMQAT: Hessian-based MPQ (Neural Networks 2024)
// - CIM²PQ (IEEE TCAD 2024)
// - COMPASS: CIM Compiler Framework (arXiv 2501.06780)
// - Sensitivity-Aware MPQ for ReRAM (arXiv 2512.19445)

package layers

import (
	"fmt"
	"math"
	"sort"
)

// ============================================================================
// Mixed-Precision Configuration Types
// ============================================================================

// BitConfig represents a bit-width configuration for weights and activations.
type BitConfig struct {
	WeightBits int // Weight precision (1-8)
	ActBits    int // Activation precision (4-16)
}

// Common bit configurations
var (
	ConfigW8A8  = BitConfig{8, 8}
	ConfigW4A8  = BitConfig{4, 8}
	ConfigW4A16 = BitConfig{4, 16}
	ConfigW2A8  = BitConfig{2, 8}
	ConfigW2A4  = BitConfig{2, 4}
	ConfigW1A8  = BitConfig{1, 8} // Binary
)

// BitConfigCost returns estimated cost (memory + compute) for a config.
func (bc BitConfig) Cost() float64 {
	// Memory cost proportional to weight bits
	memoryCost := float64(bc.WeightBits)
	// Compute cost proportional to activation bits (ADC resolution)
	computeCost := float64(bc.ActBits) * 0.5
	return memoryCost + computeCost
}

// ============================================================================
// Layer Sensitivity Analysis
// ============================================================================

// SensitivityMetric enumerates sensitivity computation methods.
type SensitivityMetric int

const (
	SENSITIVITY_HESSIAN_TRACE SensitivityMetric = iota // Average Hessian trace
	SENSITIVITY_FISHER_INFO                            // Fisher Information Matrix
	SENSITIVITY_GRADIENT_NORM                          // Gradient L2 norm
	SENSITIVITY_WEIGHT_RANGE                           // Weight value range
)

// LayerSensitivity stores sensitivity information for a layer.
type LayerSensitivity struct {
	LayerID       int
	Name          string
	HessianTrace  float64 // Average trace of Hessian
	FisherInfo    float64 // Fisher Information estimate
	GradientNorm  float64 // Gradient L2 norm
	WeightRange   float64 // Max - Min weight value
	ParamCount    int64   // Number of parameters
	Sensitivity   float64 // Combined sensitivity score
}

// SensitivityAnalyzer computes layer sensitivities for bit allocation.
type SensitivityAnalyzer struct {
	Metric        SensitivityMetric
	Layers        []*LayerSensitivity
	SortedIndices []int // Layers sorted by sensitivity
}

// NewSensitivityAnalyzer creates a new sensitivity analyzer.
func NewSensitivityAnalyzer(metric SensitivityMetric) *SensitivityAnalyzer {
	return &SensitivityAnalyzer{
		Metric: metric,
		Layers: make([]*LayerSensitivity, 0),
	}
}

// AddLayer adds a layer for sensitivity analysis.
func (sa *SensitivityAnalyzer) AddLayer(id int, name string, weights [][]float64) {
	layer := &LayerSensitivity{
		LayerID: id,
		Name:    name,
	}

	// Calculate statistics
	var sum, sumSq, minVal, maxVal float64
	minVal = math.MaxFloat64
	maxVal = -math.MaxFloat64
	count := 0

	for _, row := range weights {
		for _, w := range row {
			sum += w
			sumSq += w * w
			if w < minVal {
				minVal = w
			}
			if w > maxVal {
				maxVal = w
			}
			count++
		}
	}

	layer.ParamCount = int64(count)
	layer.WeightRange = maxVal - minVal

	if count > 0 {
		mean := sum / float64(count)
		variance := sumSq/float64(count) - mean*mean

		// Estimate Hessian trace (second derivative approximation)
		// Higher variance → higher sensitivity
		layer.HessianTrace = variance * float64(count)

		// Estimate gradient norm from weight magnitude
		layer.GradientNorm = math.Sqrt(sumSq)

		// Fisher information approximation
		layer.FisherInfo = variance * variance * float64(count)
	}

	// Compute combined sensitivity based on metric
	switch sa.Metric {
	case SENSITIVITY_HESSIAN_TRACE:
		layer.Sensitivity = layer.HessianTrace
	case SENSITIVITY_FISHER_INFO:
		layer.Sensitivity = layer.FisherInfo
	case SENSITIVITY_GRADIENT_NORM:
		layer.Sensitivity = layer.GradientNorm
	case SENSITIVITY_WEIGHT_RANGE:
		layer.Sensitivity = layer.WeightRange * float64(layer.ParamCount)
	}

	sa.Layers = append(sa.Layers, layer)
}

// ComputeHessianTrace computes approximate Hessian trace using gradients.
// H_ii ≈ (∂²L/∂w_i²) estimated via finite differences or curvature.
func (sa *SensitivityAnalyzer) ComputeHessianTrace(layerID int, gradients [][]float64) float64 {
	var trace float64
	for _, row := range gradients {
		for _, g := range row {
			// Squared gradient as curvature proxy
			trace += g * g
		}
	}

	// Update layer
	for _, layer := range sa.Layers {
		if layer.LayerID == layerID {
			layer.HessianTrace = trace
			layer.Sensitivity = trace // Update if using Hessian metric
			break
		}
	}

	return trace
}

// SortBySensitivity sorts layers by sensitivity (descending).
func (sa *SensitivityAnalyzer) SortBySensitivity() {
	sa.SortedIndices = make([]int, len(sa.Layers))
	for i := range sa.SortedIndices {
		sa.SortedIndices[i] = i
	}

	sort.Slice(sa.SortedIndices, func(i, j int) bool {
		return sa.Layers[sa.SortedIndices[i]].Sensitivity >
			sa.Layers[sa.SortedIndices[j]].Sensitivity
	})
}

// GetSensitivityRank returns the sensitivity rank of a layer (0 = most sensitive).
func (sa *SensitivityAnalyzer) GetSensitivityRank(layerID int) int {
	for rank, idx := range sa.SortedIndices {
		if sa.Layers[idx].LayerID == layerID {
			return rank
		}
	}
	return -1
}

// ============================================================================
// Mixed-Precision Bit Allocator
// ============================================================================

// BitAllocatorConfig configures mixed-precision bit allocation.
type BitAllocatorConfig struct {
	TargetBits     float64     // Target average bits
	MinBits        int         // Minimum bits (e.g., 2)
	MaxBits        int         // Maximum bits (e.g., 8)
	AllowedConfigs []BitConfig // Allowed configurations
	UseParetoOpt   bool        // Use Pareto optimization
}

// DefaultBitAllocatorConfig returns default configuration.
func DefaultBitAllocatorConfig() *BitAllocatorConfig {
	return &BitAllocatorConfig{
		TargetBits: 4.0,
		MinBits:    2,
		MaxBits:    8,
		AllowedConfigs: []BitConfig{
			ConfigW8A8,
			ConfigW4A8,
			ConfigW4A16,
			ConfigW2A8,
			ConfigW2A4,
		},
		UseParetoOpt: true,
	}
}

// BitAllocation represents bit allocation for a layer.
type BitAllocation struct {
	LayerID    int
	Config     BitConfig
	MemoryCost float64 // Bits for this layer
	Accuracy   float64 // Estimated accuracy impact
}

// BitAllocator allocates bit-widths to layers.
type BitAllocator struct {
	config      *BitAllocatorConfig
	analyzer    *SensitivityAnalyzer
	allocations []*BitAllocation

	// Pareto frontier
	paretoFront []paretoPoint

	// Statistics
	AverageBits float64
	TotalMemory int64
	AccuracyLoss float64
}

// paretoPoint represents a point on the Pareto frontier.
type paretoPoint struct {
	config   []BitConfig
	cost     float64
	accuracy float64
}

// NewBitAllocator creates a new bit allocator.
func NewBitAllocator(config *BitAllocatorConfig, analyzer *SensitivityAnalyzer) *BitAllocator {
	return &BitAllocator{
		config:      config,
		analyzer:    analyzer,
		allocations: make([]*BitAllocation, 0),
	}
}

// AllocateHAWQ allocates bits using HAWQ-style Hessian sensitivity.
// High sensitivity layers → more bits; Low sensitivity → fewer bits.
func (ba *BitAllocator) AllocateHAWQ() []*BitAllocation {
	ba.analyzer.SortBySensitivity()
	numLayers := len(ba.analyzer.Layers)

	ba.allocations = make([]*BitAllocation, numLayers)

	// Calculate sensitivity thresholds
	maxSens := ba.analyzer.Layers[ba.analyzer.SortedIndices[0]].Sensitivity
	minSens := ba.analyzer.Layers[ba.analyzer.SortedIndices[numLayers-1]].Sensitivity
	sensRange := maxSens - minSens
	if sensRange == 0 {
		sensRange = 1.0
	}

	// Allocate bits based on sensitivity rank
	var totalBits float64
	var totalParams int64

	for i, layer := range ba.analyzer.Layers {
		// Normalize sensitivity to [0, 1]
		normSens := (layer.Sensitivity - minSens) / sensRange

		// Map to bit range
		bits := float64(ba.config.MinBits) + normSens*float64(ba.config.MaxBits-ba.config.MinBits)
		bits = math.Round(bits)
		bits = math.Max(float64(ba.config.MinBits), math.Min(float64(ba.config.MaxBits), bits))

		// Find closest allowed config
		config := ba.findClosestConfig(int(bits))

		alloc := &BitAllocation{
			LayerID:    layer.LayerID,
			Config:     config,
			MemoryCost: float64(config.WeightBits) * float64(layer.ParamCount),
			Accuracy:   1.0 - 0.01*float64(8-config.WeightBits), // Simplified accuracy model
		}

		ba.allocations[i] = alloc
		totalBits += float64(config.WeightBits) * float64(layer.ParamCount)
		totalParams += layer.ParamCount
	}

	if totalParams > 0 {
		ba.AverageBits = totalBits / float64(totalParams)
	}
	ba.TotalMemory = int64(totalBits / 8) // Bytes

	return ba.allocations
}

// findClosestConfig finds the closest allowed configuration to target bits.
func (ba *BitAllocator) findClosestConfig(targetBits int) BitConfig {
	var best BitConfig
	bestDiff := math.MaxFloat64

	for _, cfg := range ba.config.AllowedConfigs {
		diff := math.Abs(float64(cfg.WeightBits - targetBits))
		if diff < bestDiff {
			bestDiff = diff
			best = cfg
		}
	}

	return best
}

// AllocatePareto uses Pareto optimization to find optimal bit allocation.
func (ba *BitAllocator) AllocatePareto() []*BitAllocation {
	numLayers := len(ba.analyzer.Layers)
	numConfigs := len(ba.config.AllowedConfigs)

	// For small problems, enumerate all combinations
	if numLayers <= 10 {
		return ba.enumeratePareto()
	}

	// For larger problems, use greedy Pareto approximation
	return ba.greedyPareto()
}

// enumeratePareto enumerates all configurations for small networks.
func (ba *BitAllocator) enumeratePareto() []*BitAllocation {
	numLayers := len(ba.analyzer.Layers)
	numConfigs := len(ba.config.AllowedConfigs)

	// Generate all combinations (limited for large networks)
	maxCombinations := 1
	for i := 0; i < numLayers; i++ {
		maxCombinations *= numConfigs
		if maxCombinations > 10000 {
			return ba.greedyPareto()
		}
	}

	// Find Pareto-optimal configurations
	ba.paretoFront = make([]paretoPoint, 0)

	// Generate combinations recursively
	current := make([]BitConfig, numLayers)
	ba.generateCombinations(current, 0)

	// Select configuration closest to target
	var bestConfig []BitConfig
	bestDist := math.MaxFloat64

	for _, point := range ba.paretoFront {
		// Distance to target (normalized)
		costDist := math.Abs(point.cost - ba.config.TargetBits)
		accDist := 1.0 - point.accuracy
		dist := math.Sqrt(costDist*costDist + accDist*accDist)

		if dist < bestDist {
			bestDist = dist
			bestConfig = point.config
		}
	}

	// Convert to allocations
	ba.allocations = make([]*BitAllocation, numLayers)
	for i, layer := range ba.analyzer.Layers {
		config := ba.config.AllowedConfigs[0]
		if bestConfig != nil && i < len(bestConfig) {
			config = bestConfig[i]
		}

		ba.allocations[i] = &BitAllocation{
			LayerID:    layer.LayerID,
			Config:     config,
			MemoryCost: float64(config.WeightBits) * float64(layer.ParamCount),
		}
	}

	return ba.allocations
}

// generateCombinations generates all bit configurations recursively.
func (ba *BitAllocator) generateCombinations(current []BitConfig, depth int) {
	if depth == len(current) {
		// Evaluate this configuration
		var cost, accuracy float64
		var totalParams int64

		for i, layer := range ba.analyzer.Layers {
			cost += float64(current[i].WeightBits) * float64(layer.ParamCount)
			totalParams += layer.ParamCount

			// Accuracy model: higher bits = better accuracy
			layerAcc := 1.0 - 0.02*float64(8-current[i].WeightBits)
			accuracy += layerAcc * float64(layer.ParamCount)
		}

		if totalParams > 0 {
			cost /= float64(totalParams)
			accuracy /= float64(totalParams)
		}

		// Check if Pareto-optimal
		isDominated := false
		for _, point := range ba.paretoFront {
			if point.cost <= cost && point.accuracy >= accuracy &&
				(point.cost < cost || point.accuracy > accuracy) {
				isDominated = true
				break
			}
		}

		if !isDominated {
			// Remove dominated points
			newFront := make([]paretoPoint, 0)
			for _, point := range ba.paretoFront {
				if !(cost <= point.cost && accuracy >= point.accuracy &&
					(cost < point.cost || accuracy > point.accuracy)) {
					newFront = append(newFront, point)
				}
			}

			configCopy := make([]BitConfig, len(current))
			copy(configCopy, current)
			newFront = append(newFront, paretoPoint{configCopy, cost, accuracy})
			ba.paretoFront = newFront
		}
		return
	}

	for _, cfg := range ba.config.AllowedConfigs {
		current[depth] = cfg
		ba.generateCombinations(current, depth+1)
	}
}

// greedyPareto uses greedy approximation for large networks.
func (ba *BitAllocator) greedyPareto() []*BitAllocation {
	ba.analyzer.SortBySensitivity()
	numLayers := len(ba.analyzer.Layers)

	ba.allocations = make([]*BitAllocation, numLayers)

	// Start with minimum bits
	for i, layer := range ba.analyzer.Layers {
		ba.allocations[i] = &BitAllocation{
			LayerID: layer.LayerID,
			Config:  ConfigW2A4,
		}
	}

	// Greedily increase bits for most sensitive layers
	budget := ba.config.TargetBits * float64(numLayers)
	usedBits := float64(ba.config.MinBits) * float64(numLayers)

	for _, idx := range ba.analyzer.SortedIndices {
		if usedBits >= budget {
			break
		}

		// Upgrade to next higher config
		currentBits := ba.allocations[idx].Config.WeightBits
		for _, cfg := range ba.config.AllowedConfigs {
			if cfg.WeightBits > currentBits {
				bitIncrease := float64(cfg.WeightBits - currentBits)
				if usedBits+bitIncrease <= budget {
					ba.allocations[idx].Config = cfg
					usedBits += bitIncrease
				}
				break
			}
		}
	}

	return ba.allocations
}

// ============================================================================
// CIM²PQ: Arraywise Mixed Precision
// ============================================================================

// CIM2PQConfig configures CIM²PQ arraywise precision.
type CIM2PQConfig struct {
	ArrayRows       int     // Crossbar rows
	ArrayCols       int     // Crossbar columns
	ADCResolution   int     // ADC bits (4-8)
	DACResolution   int     // DAC bits (4-8)
	EnergyWeight    float64 // Weight for energy in optimization
	AccuracyWeight  float64 // Weight for accuracy
}

// DefaultCIM2PQConfig returns default CIM²PQ configuration.
func DefaultCIM2PQConfig() *CIM2PQConfig {
	return &CIM2PQConfig{
		ArrayRows:      256,
		ArrayCols:      256,
		ADCResolution:  6,
		DACResolution:  8,
		EnergyWeight:   0.5,
		AccuracyWeight: 0.5,
	}
}

// ArrayPrecision represents precision for a crossbar array.
type ArrayPrecision struct {
	ArrayID     int
	WeightBits  int
	InputBits   int
	OutputBits  int // ADC resolution
	Energy      float64
	Utilization float64
}

// CIM2PQOptimizer optimizes arraywise precision.
type CIM2PQOptimizer struct {
	config     *CIM2PQConfig
	arrays     []*ArrayPrecision
	TotalEnergy float64
	AvgAccuracy float64
}

// NewCIM2PQOptimizer creates a new CIM²PQ optimizer.
func NewCIM2PQOptimizer(config *CIM2PQConfig) *CIM2PQOptimizer {
	return &CIM2PQOptimizer{
		config: config,
		arrays: make([]*ArrayPrecision, 0),
	}
}

// OptimizeArrays optimizes precision for each crossbar array.
func (opt *CIM2PQOptimizer) OptimizeArrays(numArrays int, layerSensitivities []float64) []*ArrayPrecision {
	opt.arrays = make([]*ArrayPrecision, numArrays)

	// Normalize sensitivities
	maxSens := 0.0
	for _, s := range layerSensitivities {
		if s > maxSens {
			maxSens = s
		}
	}
	if maxSens == 0 {
		maxSens = 1.0
	}

	var totalEnergy float64

	for i := 0; i < numArrays; i++ {
		sens := 0.0
		if i < len(layerSensitivities) {
			sens = layerSensitivities[i] / maxSens
		}

		// Higher sensitivity → more bits
		weightBits := 2 + int(sens*6) // 2-8 bits
		inputBits := opt.config.DACResolution
		outputBits := opt.config.ADCResolution

		// Energy model: E ∝ bits² × array_size
		energy := float64(weightBits*weightBits) *
			float64(opt.config.ArrayRows*opt.config.ArrayCols) * 1e-15 // fJ

		opt.arrays[i] = &ArrayPrecision{
			ArrayID:     i,
			WeightBits:  weightBits,
			InputBits:   inputBits,
			OutputBits:  outputBits,
			Energy:      energy,
			Utilization: 0.8 + 0.2*sens, // Higher precision → higher utilization
		}

		totalEnergy += energy
	}

	opt.TotalEnergy = totalEnergy
	return opt.arrays
}

// ============================================================================
// Weight Tiling for Large Models
// ============================================================================

// TileConfig configures weight tiling for crossbar arrays.
type TileConfig struct {
	CrossbarRows    int  // Rows per crossbar
	CrossbarCols    int  // Columns per crossbar
	MaxCrossbars    int  // Maximum available crossbars
	AllowDuplication bool // Allow weight duplication for input reuse
	OptimizeMemory  bool // Optimize for memory
	OptimizeLatency bool // Optimize for latency
}

// DefaultTileConfig returns default tiling configuration.
func DefaultTileConfig() *TileConfig {
	return &TileConfig{
		CrossbarRows:    256,
		CrossbarCols:    256,
		MaxCrossbars:    64,
		AllowDuplication: true,
		OptimizeMemory:  true,
		OptimizeLatency: false,
	}
}

// WeightTile represents a tile of weights mapped to a crossbar.
type WeightTile struct {
	TileID      int
	CrossbarID  int
	LayerID     int
	RowStart    int // Start row in original weight matrix
	RowEnd      int // End row (exclusive)
	ColStart    int // Start column
	ColEnd      int // End column
	IsDuplicate bool // Whether this is a duplicate for input reuse
}

// TileDependency represents dependency between tiles.
type TileDependency struct {
	SourceTile int
	TargetTile int
	Type       string // "row", "col", "accumulate"
}

// WeightTiler partitions weight matrices across crossbar arrays.
type WeightTiler struct {
	config      *TileConfig
	tiles       []*WeightTile
	dependencies []*TileDependency

	// Statistics
	NumTiles        int
	NumCrossbars    int
	Utilization     float64
	ReloadRequired  bool
}

// NewWeightTiler creates a new weight tiler.
func NewWeightTiler(config *TileConfig) *WeightTiler {
	return &WeightTiler{
		config:       config,
		tiles:        make([]*WeightTile, 0),
		dependencies: make([]*TileDependency, 0),
	}
}

// TileLayer tiles a single layer's weights.
func (wt *WeightTiler) TileLayer(layerID int, rows, cols int) []*WeightTile {
	tiles := make([]*WeightTile, 0)

	// Calculate number of tiles needed
	rowTiles := (rows + wt.config.CrossbarRows - 1) / wt.config.CrossbarRows
	colTiles := (cols + wt.config.CrossbarCols - 1) / wt.config.CrossbarCols

	tileID := len(wt.tiles)

	for rt := 0; rt < rowTiles; rt++ {
		for ct := 0; ct < colTiles; ct++ {
			rowStart := rt * wt.config.CrossbarRows
			rowEnd := min(rowStart+wt.config.CrossbarRows, rows)
			colStart := ct * wt.config.CrossbarCols
			colEnd := min(colStart+wt.config.CrossbarCols, cols)

			tile := &WeightTile{
				TileID:      tileID,
				CrossbarID:  tileID % wt.config.MaxCrossbars,
				LayerID:     layerID,
				RowStart:    rowStart,
				RowEnd:      rowEnd,
				ColStart:    colStart,
				ColEnd:      colEnd,
				IsDuplicate: false,
			}

			tiles = append(tiles, tile)
			wt.tiles = append(wt.tiles, tile)
			tileID++

			// Add column accumulation dependency
			if ct > 0 {
				wt.dependencies = append(wt.dependencies, &TileDependency{
					SourceTile: tileID - 2,
					TargetTile: tileID - 1,
					Type:       "accumulate",
				})
			}
		}
	}

	wt.NumTiles = len(wt.tiles)
	wt.NumCrossbars = min(wt.NumTiles, wt.config.MaxCrossbars)
	wt.ReloadRequired = wt.NumTiles > wt.config.MaxCrossbars

	// Calculate utilization
	totalCells := int64(wt.NumTiles) * int64(wt.config.CrossbarRows) * int64(wt.config.CrossbarCols)
	usedCells := int64(rows) * int64(cols)
	if totalCells > 0 {
		wt.Utilization = float64(usedCells) / float64(totalCells)
	}

	return tiles
}

// TileConvolution tiles convolution weights using im2col.
func (wt *WeightTiler) TileConvolution(layerID int, outChannels, inChannels, kernelH, kernelW int) []*WeightTile {
	// Im2col unrolls convolution into matrix multiplication
	// Weight matrix: outChannels × (inChannels × kernelH × kernelW)
	rows := outChannels
	cols := inChannels * kernelH * kernelW

	return wt.TileLayer(layerID, rows, cols)
}

// GetExecutionSchedule returns tile execution schedule respecting dependencies.
func (wt *WeightTiler) GetExecutionSchedule() [][]int {
	// Build dependency graph
	inDegree := make(map[int]int)
	outEdges := make(map[int][]int)

	for _, tile := range wt.tiles {
		inDegree[tile.TileID] = 0
	}

	for _, dep := range wt.dependencies {
		inDegree[dep.TargetTile]++
		outEdges[dep.SourceTile] = append(outEdges[dep.SourceTile], dep.TargetTile)
	}

	// Topological sort with level grouping
	schedule := make([][]int, 0)
	ready := make([]int, 0)

	// Find initial ready tiles
	for id, deg := range inDegree {
		if deg == 0 {
			ready = append(ready, id)
		}
	}

	for len(ready) > 0 {
		// Execute all ready tiles in parallel (up to crossbar limit)
		level := make([]int, 0)
		for i := 0; i < len(ready) && i < wt.config.MaxCrossbars; i++ {
			level = append(level, ready[i])
		}
		schedule = append(schedule, level)

		// Find next ready tiles
		nextReady := make([]int, 0)
		for _, tileID := range ready[len(level):] {
			nextReady = append(nextReady, tileID)
		}

		for _, tileID := range level {
			for _, target := range outEdges[tileID] {
				inDegree[target]--
				if inDegree[target] == 0 {
					nextReady = append(nextReady, target)
				}
			}
		}

		ready = nextReady
	}

	return schedule
}

// ============================================================================
// COMPASS-Style Network Partitioning
// ============================================================================

// PartitionConfig configures network partitioning.
type PartitionConfig struct {
	MaxWeightMemory int64 // Maximum weight memory per partition (bytes)
	MaxActivations  int64 // Maximum activation memory
	EnableReload    bool  // Enable weight reloading from external memory
	BalanceLoad     bool  // Balance load across partitions
}

// DefaultPartitionConfig returns default partition configuration.
func DefaultPartitionConfig() *PartitionConfig {
	return &PartitionConfig{
		MaxWeightMemory: 1 << 20, // 1 MB
		MaxActivations:  1 << 18, // 256 KB
		EnableReload:    true,
		BalanceLoad:     true,
	}
}

// Partition represents a partition of the network.
type Partition struct {
	ID           int
	LayerIDs     []int
	WeightMemory int64
	ActMemory    int64
	Tiles        []*WeightTile
	RequiresReload bool
}

// NetworkPartitioner partitions a network for CIM deployment.
type NetworkPartitioner struct {
	config     *PartitionConfig
	partitions []*Partition

	// Statistics
	NumPartitions   int
	TotalReloads    int
	MaxMemoryUsed   int64
	LoadImbalance   float64
}

// NewNetworkPartitioner creates a new network partitioner.
func NewNetworkPartitioner(config *PartitionConfig) *NetworkPartitioner {
	return &NetworkPartitioner{
		config:     config,
		partitions: make([]*Partition, 0),
	}
}

// LayerInfo contains information about a layer for partitioning.
type LayerInfo struct {
	ID           int
	WeightSize   int64 // Weight memory in bytes
	ActSize      int64 // Activation memory in bytes
	InputLayers  []int // Input dependencies
	OutputLayers []int // Output dependencies
}

// PartitionNetwork partitions layers into memory-constrained partitions.
func (np *NetworkPartitioner) PartitionNetwork(layers []*LayerInfo) []*Partition {
	np.partitions = make([]*Partition, 0)

	currentPartition := &Partition{
		ID:       0,
		LayerIDs: make([]int, 0),
	}

	for _, layer := range layers {
		// Check if layer fits in current partition
		newWeightMem := currentPartition.WeightMemory + layer.WeightSize
		newActMem := max(currentPartition.ActMemory, layer.ActSize)

		if newWeightMem <= np.config.MaxWeightMemory &&
			newActMem <= np.config.MaxActivations {
			// Add to current partition
			currentPartition.LayerIDs = append(currentPartition.LayerIDs, layer.ID)
			currentPartition.WeightMemory = newWeightMem
			currentPartition.ActMemory = newActMem
		} else {
			// Start new partition
			if len(currentPartition.LayerIDs) > 0 {
				np.partitions = append(np.partitions, currentPartition)
			}

			currentPartition = &Partition{
				ID:           len(np.partitions),
				LayerIDs:     []int{layer.ID},
				WeightMemory: layer.WeightSize,
				ActMemory:    layer.ActSize,
			}

			// Check if single layer exceeds memory
			if layer.WeightSize > np.config.MaxWeightMemory {
				currentPartition.RequiresReload = true
				np.TotalReloads++
			}
		}
	}

	// Add final partition
	if len(currentPartition.LayerIDs) > 0 {
		np.partitions = append(np.partitions, currentPartition)
	}

	np.NumPartitions = len(np.partitions)
	np.calculateStatistics()

	return np.partitions
}

// calculateStatistics computes partitioning statistics.
func (np *NetworkPartitioner) calculateStatistics() {
	if len(np.partitions) == 0 {
		return
	}

	var totalMem int64
	var maxMem int64

	for _, p := range np.partitions {
		mem := p.WeightMemory + p.ActMemory
		totalMem += mem
		if mem > maxMem {
			maxMem = mem
		}
	}

	np.MaxMemoryUsed = maxMem
	avgMem := totalMem / int64(len(np.partitions))
	if avgMem > 0 {
		np.LoadImbalance = float64(maxMem-avgMem) / float64(avgMem)
	}
}

// ============================================================================
// Im2Col Mapping for Convolutions
// ============================================================================

// Im2ColMapper maps convolution operations to crossbar format.
type Im2ColMapper struct {
	// Input dimensions
	InputH, InputW, InputC int
	// Kernel dimensions
	KernelH, KernelW int
	// Convolution parameters
	Stride, Padding int
	// Output dimensions
	OutputH, OutputW, OutputC int
}

// NewIm2ColMapper creates a new im2col mapper.
func NewIm2ColMapper(inputH, inputW, inputC, kernelH, kernelW, outputC, stride, padding int) *Im2ColMapper {
	outputH := (inputH + 2*padding - kernelH) / stride + 1
	outputW := (inputW + 2*padding - kernelW) / stride + 1

	return &Im2ColMapper{
		InputH:  inputH,
		InputW:  inputW,
		InputC:  inputC,
		KernelH: kernelH,
		KernelW: kernelW,
		Stride:  stride,
		Padding: padding,
		OutputH: outputH,
		OutputW: outputW,
		OutputC: outputC,
	}
}

// GetWeightMatrixDims returns dimensions of unrolled weight matrix.
func (im *Im2ColMapper) GetWeightMatrixDims() (rows, cols int) {
	// Rows = output channels
	// Cols = input channels × kernel height × kernel width
	rows = im.OutputC
	cols = im.InputC * im.KernelH * im.KernelW
	return
}

// GetInputMatrixDims returns dimensions of unrolled input matrix.
func (im *Im2ColMapper) GetInputMatrixDims() (rows, cols int) {
	// Rows = input channels × kernel height × kernel width
	// Cols = output height × output width
	rows = im.InputC * im.KernelH * im.KernelW
	cols = im.OutputH * im.OutputW
	return
}

// UnrollWeights unrolls convolution kernels to matrix format.
func (im *Im2ColMapper) UnrollWeights(weights [][][][]float64) [][]float64 {
	rows, cols := im.GetWeightMatrixDims()
	matrix := make([][]float64, rows)

	for oc := 0; oc < im.OutputC; oc++ {
		matrix[oc] = make([]float64, cols)
		col := 0
		for ic := 0; ic < im.InputC; ic++ {
			for kh := 0; kh < im.KernelH; kh++ {
				for kw := 0; kw < im.KernelW; kw++ {
					if oc < len(weights) && ic < len(weights[oc]) &&
						kh < len(weights[oc][ic]) && kw < len(weights[oc][ic][kh]) {
						matrix[oc][col] = weights[oc][ic][kh][kw]
					}
					col++
				}
			}
		}
	}

	return matrix
}

// UnrollInput applies im2col to input tensor.
func (im *Im2ColMapper) UnrollInput(input [][][]float64) [][]float64 {
	rows, cols := im.GetInputMatrixDims()
	matrix := make([][]float64, rows)

	for i := range matrix {
		matrix[i] = make([]float64, cols)
	}

	col := 0
	for oh := 0; oh < im.OutputH; oh++ {
		for ow := 0; ow < im.OutputW; ow++ {
			row := 0
			for ic := 0; ic < im.InputC; ic++ {
				for kh := 0; kh < im.KernelH; kh++ {
					for kw := 0; kw < im.KernelW; kw++ {
						ih := oh*im.Stride + kh - im.Padding
						iw := ow*im.Stride + kw - im.Padding

						if ih >= 0 && ih < im.InputH && iw >= 0 && iw < im.InputW {
							if ic < len(input) && ih < len(input[ic]) && iw < len(input[ic][ih]) {
								matrix[row][col] = input[ic][ih][iw]
							}
						}
						row++
					}
				}
			}
			col++
		}
	}

	return matrix
}

// ============================================================================
// Utility Functions
// ============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// EstimateLayerEnergy estimates energy for a layer with given precision.
func EstimateLayerEnergy(params int64, config BitConfig, arraySize int) float64 {
	// Energy model: E ∝ bits × params × switching_energy
	switchingEnergy := 1e-15 // 1 fJ per MAC
	bits := float64(config.WeightBits * config.ActBits)
	return bits * float64(params) * switchingEnergy
}

// EstimateLayerLatency estimates latency for a layer.
func EstimateLayerLatency(params int64, config BitConfig, arraySize int, clockFreq float64) float64 {
	// Latency = params / (array_throughput × clock)
	throughput := float64(arraySize * arraySize) // MACs per cycle
	cycles := float64(params) / throughput
	return cycles / clockFreq
}

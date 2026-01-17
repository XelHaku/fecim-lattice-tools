// Package layers provides CIM scheduling and diffusion model acceleration simulation.
//
// This module implements:
// - Cross-layer scheduling (CLSA-CIM) for tiled CIM architectures
// - Resource management with dual-mode array switching
// - Workload tiling and mapping optimization
// - Diffusion model accelerator simulation (AIG-CIM style)
// - Denoising step-aware scheduling
//
// Based on research from DATE 2024, DAC 2024, and ASPLOS 2024.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// CIM SCHEDULING AND RESOURCE MANAGEMENT
// =============================================================================

// TileConfig defines the configuration of a CIM tile.
type TileConfig struct {
	TileID           int
	Rows             int
	Cols             int
	PEsPerTile       int
	WeightBits       int
	ActivationBits   int
	ComputeLatencyNS float64
	EnergyPerMACPJ   float64
	Mode             TileMode
}

// TileMode represents the operating mode of a dual-mode tile.
type TileMode int

const (
	TileModeCompute TileMode = iota // In-situ MAC computation
	TileModeMemory                  // Scratchpad storage
	TileModeIdle                    // Not active
)

// NoCConfig defines the network-on-chip configuration.
type NoCConfig struct {
	TopologyType    NoCTopology
	Rows            int
	Cols            int
	LinkBandwidthGB float64 // GB/s per link
	HopLatencyNS    float64 // Latency per hop
	RouterLatencyNS float64 // Router processing latency
}

// NoCTopology defines the NoC interconnect topology.
type NoCTopology int

const (
	NoCTopologyMesh NoCTopology = iota
	NoCTopologyTorus
	NoCTopologyRing
	NoCTopologyHierarchical
)

// LayerMapping represents the mapping of a neural network layer to tiles.
type LayerMapping struct {
	LayerID         int
	LayerName       string
	LayerType       string // conv, fc, attention, etc.
	TileAssignments []int  // Which tiles hold weights for this layer
	WeightPartition []int  // Weight partition sizes per tile
	InputDependency []int  // IDs of layers this depends on
	OutputConsumers []int  // IDs of layers that consume this output
	ComputeCycles   int64
	MemoryFootprint int64
}

// ScheduleEntry represents a scheduled operation in the execution timeline.
type ScheduleEntry struct {
	LayerID     int
	TileID      int
	StartCycle  int64
	EndCycle    int64
	Operation   ScheduleOp
	PartialData bool // Whether this processes partial results
}

// ScheduleOp defines the type of scheduled operation.
type ScheduleOp int

const (
	ScheduleOpCompute ScheduleOp = iota
	ScheduleOpDataMove
	ScheduleOpSync
	ScheduleOpActivation
)

// CLSACIMScheduler implements cross-layer scheduling for CIM architectures.
type CLSACIMScheduler struct {
	Config          *CLSAConfig
	Tiles           []*TileConfig
	NoC             *NoCConfig
	LayerMappings   []*LayerMapping
	Schedule        []*ScheduleEntry
	TileUtilization map[int]float64
	PEUtilization   float64
	TotalLatency    int64
}

// CLSAConfig holds configuration for the CLSA-CIM scheduler.
type CLSAConfig struct {
	EnableCrossLayer    bool
	EnableWeightDup     bool
	DuplicationFactor   int     // Number of weight copies
	PipelineDepth       int     // Cross-layer pipeline depth
	PartialResultSize   int     // Size of partial result chunks
	MaxActiveTiles      int     // Maximum tiles active simultaneously
	UtilizationTarget   float64 // Target PE utilization
	SchedulingHeuristic SchedulingHeuristic
}

// SchedulingHeuristic defines the scheduling strategy.
type SchedulingHeuristic int

const (
	HeuristicLayerByLayer SchedulingHeuristic = iota
	HeuristicCrossLayer
	HeuristicPipelined
	HeuristicDynamic
)

// NewCLSACIMScheduler creates a new cross-layer scheduler.
func NewCLSACIMScheduler(config *CLSAConfig, numTiles int) *CLSACIMScheduler {
	tiles := make([]*TileConfig, numTiles)
	for i := 0; i < numTiles; i++ {
		tiles[i] = &TileConfig{
			TileID:           i,
			Rows:             256,
			Cols:             256,
			PEsPerTile:       64, // 64 PEs per tile
			WeightBits:       8,
			ActivationBits:   8,
			ComputeLatencyNS: 10.0,
			EnergyPerMACPJ:   0.5,
			Mode:             TileModeIdle,
		}
	}

	return &CLSACIMScheduler{
		Config:          config,
		Tiles:           tiles,
		TileUtilization: make(map[int]float64),
	}
}

// SetNoC configures the network-on-chip.
func (s *CLSACIMScheduler) SetNoC(noc *NoCConfig) {
	s.NoC = noc
}

// MapLayer maps a neural network layer to available tiles.
func (s *CLSACIMScheduler) MapLayer(layerID int, layerName, layerType string,
	weightSize int64, inputDeps, outputConsumers []int) *LayerMapping {

	// Calculate number of tiles needed
	tileCapacity := int64(s.Tiles[0].Rows * s.Tiles[0].Cols * s.Tiles[0].WeightBits / 8)
	tilesNeeded := int((weightSize + tileCapacity - 1) / tileCapacity)

	// With weight duplication, multiply tiles
	if s.Config.EnableWeightDup {
		tilesNeeded *= s.Config.DuplicationFactor
	}

	// Assign tiles
	assignments := make([]int, tilesNeeded)
	partitions := make([]int, tilesNeeded)
	remainingWeight := weightSize

	for i := 0; i < tilesNeeded; i++ {
		assignments[i] = (len(s.LayerMappings)*tilesNeeded + i) % len(s.Tiles)
		partitions[i] = int(min64(remainingWeight, tileCapacity))
		remainingWeight -= tileCapacity
		if remainingWeight < 0 {
			remainingWeight = 0
		}
	}

	mapping := &LayerMapping{
		LayerID:         layerID,
		LayerName:       layerName,
		LayerType:       layerType,
		TileAssignments: assignments,
		WeightPartition: partitions,
		InputDependency: inputDeps,
		OutputConsumers: outputConsumers,
		ComputeCycles:   weightSize / 8, // Simplified compute estimate
		MemoryFootprint: weightSize,
	}

	s.LayerMappings = append(s.LayerMappings, mapping)
	return mapping
}

// ScheduleInference generates the execution schedule for inference.
func (s *CLSACIMScheduler) ScheduleInference() []*ScheduleEntry {
	s.Schedule = nil

	switch s.Config.SchedulingHeuristic {
	case HeuristicLayerByLayer:
		s.scheduleLayerByLayer()
	case HeuristicCrossLayer:
		s.scheduleCrossLayer()
	case HeuristicPipelined:
		s.schedulePipelined()
	case HeuristicDynamic:
		s.scheduleDynamic()
	}

	s.calculateUtilization()
	return s.Schedule
}

// scheduleLayerByLayer implements traditional layer-by-layer scheduling.
func (s *CLSACIMScheduler) scheduleLayerByLayer() {
	currentCycle := int64(0)

	for _, mapping := range s.LayerMappings {
		// Schedule all tiles for this layer
		for i, tileID := range mapping.TileAssignments {
			cyclesForPartition := int64(mapping.WeightPartition[i]) / 8

			entry := &ScheduleEntry{
				LayerID:     mapping.LayerID,
				TileID:      tileID,
				StartCycle:  currentCycle,
				EndCycle:    currentCycle + cyclesForPartition,
				Operation:   ScheduleOpCompute,
				PartialData: false,
			}
			s.Schedule = append(s.Schedule, entry)
		}

		// Find max end cycle for this layer
		maxEnd := currentCycle
		for _, entry := range s.Schedule {
			if entry.LayerID == mapping.LayerID && entry.EndCycle > maxEnd {
				maxEnd = entry.EndCycle
			}
		}
		currentCycle = maxEnd
	}

	s.TotalLatency = currentCycle
}

// scheduleCrossLayer implements cross-layer scheduling (CLSA-CIM).
func (s *CLSACIMScheduler) scheduleCrossLayer() {
	// Track when each tile becomes available
	tileAvailable := make([]int64, len(s.Tiles))
	layerComplete := make(map[int]int64)

	for _, mapping := range s.LayerMappings {
		// Find earliest start time based on dependencies
		earliestStart := int64(0)
		for _, depID := range mapping.InputDependency {
			if completeTime, ok := layerComplete[depID]; ok {
				// Cross-layer: can start with partial results
				partialReadyTime := completeTime - int64(s.Config.PartialResultSize)
				if partialReadyTime > earliestStart {
					earliestStart = partialReadyTime
				}
			}
		}

		// Schedule tiles for this layer
		var layerEndTime int64
		for i, tileID := range mapping.TileAssignments {
			startTime := max64(earliestStart, tileAvailable[tileID])
			cyclesForPartition := int64(mapping.WeightPartition[i]) / 8

			entry := &ScheduleEntry{
				LayerID:     mapping.LayerID,
				TileID:      tileID,
				StartCycle:  startTime,
				EndCycle:    startTime + cyclesForPartition,
				Operation:   ScheduleOpCompute,
				PartialData: len(mapping.InputDependency) > 0 && s.Config.EnableCrossLayer,
			}
			s.Schedule = append(s.Schedule, entry)

			tileAvailable[tileID] = entry.EndCycle
			if entry.EndCycle > layerEndTime {
				layerEndTime = entry.EndCycle
			}
		}

		layerComplete[mapping.LayerID] = layerEndTime
	}

	// Calculate total latency
	for _, t := range layerComplete {
		if t > s.TotalLatency {
			s.TotalLatency = t
		}
	}
}

// schedulePipelined implements pipelined cross-layer scheduling.
func (s *CLSACIMScheduler) schedulePipelined() {
	pipelineStages := s.Config.PipelineDepth
	tileAvailable := make([]int64, len(s.Tiles))
	layerPartialReady := make(map[int][]int64) // Partial result ready times

	for _, mapping := range s.LayerMappings {
		partitionCount := len(mapping.TileAssignments)
		partialTimes := make([]int64, pipelineStages)

		// Check dependencies with pipeline awareness
		earliestStart := int64(0)
		for _, depID := range mapping.InputDependency {
			if partials, ok := layerPartialReady[depID]; ok && len(partials) > 0 {
				// Can start when first partial result is ready
				if partials[0] > earliestStart {
					earliestStart = partials[0]
				}
			}
		}

		// Schedule with pipeline stages
		var layerEndTime int64
		for i, tileID := range mapping.TileAssignments {
			pipelineStage := i % pipelineStages
			stageDelay := int64(pipelineStage * s.Config.PartialResultSize / partitionCount)

			startTime := max64(earliestStart+stageDelay, tileAvailable[tileID])
			cyclesForPartition := int64(mapping.WeightPartition[i]) / 8

			entry := &ScheduleEntry{
				LayerID:     mapping.LayerID,
				TileID:      tileID,
				StartCycle:  startTime,
				EndCycle:    startTime + cyclesForPartition,
				Operation:   ScheduleOpCompute,
				PartialData: true,
			}
			s.Schedule = append(s.Schedule, entry)

			tileAvailable[tileID] = entry.EndCycle

			// Track partial result ready times
			stageEnd := startTime + cyclesForPartition/int64(pipelineStages)
			if pipelineStage < len(partialTimes) {
				partialTimes[pipelineStage] = stageEnd
			}

			if entry.EndCycle > layerEndTime {
				layerEndTime = entry.EndCycle
			}
		}

		layerPartialReady[mapping.LayerID] = partialTimes
	}

	// Calculate total latency
	for _, entry := range s.Schedule {
		if entry.EndCycle > s.TotalLatency {
			s.TotalLatency = entry.EndCycle
		}
	}
}

// scheduleDynamic implements dynamic scheduling with runtime decisions.
func (s *CLSACIMScheduler) scheduleDynamic() {
	// Priority queue based on criticality
	type layerPriority struct {
		mapping  *LayerMapping
		priority float64
	}

	priorities := make([]layerPriority, len(s.LayerMappings))
	for i, m := range s.LayerMappings {
		// Priority based on number of consumers and compute size
		priority := float64(len(m.OutputConsumers))*100 + float64(m.ComputeCycles)/1000
		priorities[i] = layerPriority{m, priority}
	}

	// Sort by priority (descending)
	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i].priority > priorities[j].priority
	})

	// Schedule in priority order with dynamic tile selection
	tileAvailable := make([]int64, len(s.Tiles))
	layerComplete := make(map[int]int64)
	scheduled := make(map[int]bool)

	for len(scheduled) < len(s.LayerMappings) {
		for _, lp := range priorities {
			mapping := lp.mapping
			if scheduled[mapping.LayerID] {
				continue
			}

			// Check if dependencies are met
			depsReady := true
			earliestStart := int64(0)
			for _, depID := range mapping.InputDependency {
				if !scheduled[depID] {
					depsReady = false
					break
				}
				if layerComplete[depID] > earliestStart {
					earliestStart = layerComplete[depID]
				}
			}

			if !depsReady {
				continue
			}

			// Dynamically select best tiles
			var layerEndTime int64
			for i, tileID := range mapping.TileAssignments {
				// Find earliest available tile
				bestTile := tileID
				for t := 0; t < len(s.Tiles); t++ {
					if tileAvailable[t] < tileAvailable[bestTile] {
						bestTile = t
					}
				}

				startTime := max64(earliestStart, tileAvailable[bestTile])
				cyclesForPartition := int64(mapping.WeightPartition[i]) / 8

				entry := &ScheduleEntry{
					LayerID:     mapping.LayerID,
					TileID:      bestTile,
					StartCycle:  startTime,
					EndCycle:    startTime + cyclesForPartition,
					Operation:   ScheduleOpCompute,
					PartialData: false,
				}
				s.Schedule = append(s.Schedule, entry)

				tileAvailable[bestTile] = entry.EndCycle
				if entry.EndCycle > layerEndTime {
					layerEndTime = entry.EndCycle
				}
			}

			layerComplete[mapping.LayerID] = layerEndTime
			scheduled[mapping.LayerID] = true
		}
	}

	// Calculate total latency
	for _, t := range layerComplete {
		if t > s.TotalLatency {
			s.TotalLatency = t
		}
	}
}

// calculateUtilization computes tile and PE utilization metrics.
func (s *CLSACIMScheduler) calculateUtilization() {
	if s.TotalLatency == 0 {
		return
	}

	// Per-tile utilization
	tileActive := make(map[int]int64)
	for _, entry := range s.Schedule {
		if entry.Operation == ScheduleOpCompute {
			tileActive[entry.TileID] += entry.EndCycle - entry.StartCycle
		}
	}

	totalPECycles := int64(0)
	activePECycles := int64(0)

	for tileID, active := range tileActive {
		s.TileUtilization[tileID] = float64(active) / float64(s.TotalLatency)
		totalPECycles += s.TotalLatency * int64(s.Tiles[tileID].PEsPerTile)
		activePECycles += active * int64(s.Tiles[tileID].PEsPerTile)
	}

	// Add idle tiles to total
	for _, tile := range s.Tiles {
		if _, ok := tileActive[tile.TileID]; !ok {
			totalPECycles += s.TotalLatency * int64(tile.PEsPerTile)
		}
	}

	if totalPECycles > 0 {
		s.PEUtilization = float64(activePECycles) / float64(totalPECycles)
	}
}

// GetSpeedup calculates speedup compared to layer-by-layer baseline.
func (s *CLSACIMScheduler) GetSpeedup() float64 {
	// Run layer-by-layer for comparison
	baseline := &CLSACIMScheduler{
		Config: &CLSAConfig{
			EnableCrossLayer:    false,
			EnableWeightDup:     false,
			SchedulingHeuristic: HeuristicLayerByLayer,
		},
		Tiles:           s.Tiles,
		LayerMappings:   s.LayerMappings,
		TileUtilization: make(map[int]float64),
	}
	baseline.scheduleLayerByLayer()

	if s.TotalLatency == 0 {
		return 1.0
	}
	return float64(baseline.TotalLatency) / float64(s.TotalLatency)
}

// =============================================================================
// DUAL-MODE RESOURCE MANAGEMENT
// =============================================================================

// DualModeManager manages compute/memory mode switching for CIM arrays.
type DualModeManager struct {
	Config          *DualModeConfig
	Arrays          []*DualModeArray
	CurrentWorkload *WorkloadProfile
	SwitchHistory   []*ModeSwitchEvent
}

// DualModeConfig configures dual-mode operation.
type DualModeConfig struct {
	SwitchLatencyCycles int     // Cycles to switch modes
	ComputeEnergyPJ     float64 // Energy per compute operation
	MemoryEnergyPJ      float64 // Energy per memory access
	SwitchEnergyPJ      float64 // Energy for mode switch
	MaxConcurrentSwitch int     // Max arrays switching at once
}

// DualModeArray represents a switchable CIM array.
type DualModeArray struct {
	ArrayID     int
	Mode        TileMode
	Rows        int
	Cols        int
	StoredData  [][]float64 // Data when in memory mode
	Weights     [][]float64 // Weights when in compute mode
	ActiveCycle int64
	TotalCycles int64
}

// WorkloadProfile describes workload characteristics for mode optimization.
type WorkloadProfile struct {
	ComputeIntensity   float64 // MACs per byte
	MemoryFootprint    int64   // Bytes needed
	TemporalLocality   float64 // 0-1 reuse factor
	LayerSizes         []int64 // Size of each layer
	LayerComputeRatios []float64
}

// ModeSwitchEvent records a mode switch.
type ModeSwitchEvent struct {
	ArrayID   int
	FromMode  TileMode
	ToMode    TileMode
	Cycle     int64
	EnergyPJ  float64
}

// NewDualModeManager creates a new dual-mode manager.
func NewDualModeManager(config *DualModeConfig, numArrays int) *DualModeManager {
	arrays := make([]*DualModeArray, numArrays)
	for i := 0; i < numArrays; i++ {
		arrays[i] = &DualModeArray{
			ArrayID: i,
			Mode:    TileModeIdle,
			Rows:    256,
			Cols:    256,
		}
	}

	return &DualModeManager{
		Config: config,
		Arrays: arrays,
	}
}

// AnalyzeWorkload determines optimal mode allocation.
func (m *DualModeManager) AnalyzeWorkload(profile *WorkloadProfile) (computeArrays, memoryArrays int) {
	m.CurrentWorkload = profile

	totalArrays := len(m.Arrays)
	arrayCapacity := int64(m.Arrays[0].Rows * m.Arrays[0].Cols * 2) // 2 bytes per weight

	// Calculate arrays needed for compute (weights)
	totalWeights := int64(0)
	for _, size := range profile.LayerSizes {
		totalWeights += size
	}
	computeArrays = int((totalWeights + arrayCapacity - 1) / arrayCapacity)

	// Calculate arrays needed for memory (activations)
	memoryNeeded := profile.MemoryFootprint
	memoryArrays = int((memoryNeeded + arrayCapacity - 1) / arrayCapacity)

	// Balance based on compute intensity
	if profile.ComputeIntensity > 10.0 {
		// Compute-bound: prefer more compute arrays
		computeArrays = min(totalArrays*3/4, computeArrays+2)
		memoryArrays = totalArrays - computeArrays
	} else if profile.ComputeIntensity < 1.0 {
		// Memory-bound: prefer more memory arrays
		memoryArrays = min(totalArrays*3/4, memoryArrays+2)
		computeArrays = totalArrays - memoryArrays
	}

	// Clamp to available arrays
	if computeArrays+memoryArrays > totalArrays {
		ratio := float64(totalArrays) / float64(computeArrays+memoryArrays)
		computeArrays = int(float64(computeArrays) * ratio)
		memoryArrays = totalArrays - computeArrays
	}

	return computeArrays, memoryArrays
}

// AllocateMode sets array modes based on workload analysis.
func (m *DualModeManager) AllocateMode(computeCount, memoryCount int, currentCycle int64) {
	switchCount := 0

	for i, array := range m.Arrays {
		var targetMode TileMode
		if i < computeCount {
			targetMode = TileModeCompute
		} else if i < computeCount+memoryCount {
			targetMode = TileModeMemory
		} else {
			targetMode = TileModeIdle
		}

		if array.Mode != targetMode {
			if switchCount < m.Config.MaxConcurrentSwitch {
				event := &ModeSwitchEvent{
					ArrayID:  array.ArrayID,
					FromMode: array.Mode,
					ToMode:   targetMode,
					Cycle:    currentCycle,
					EnergyPJ: m.Config.SwitchEnergyPJ,
				}
				m.SwitchHistory = append(m.SwitchHistory, event)
				array.Mode = targetMode
				switchCount++
			}
		}
	}
}

// GetSwitchOverhead returns total switch overhead in cycles and energy.
func (m *DualModeManager) GetSwitchOverhead() (cycles int64, energyPJ float64) {
	cycles = int64(len(m.SwitchHistory)) * int64(m.Config.SwitchLatencyCycles)
	for _, event := range m.SwitchHistory {
		energyPJ += event.EnergyPJ
	}
	return cycles, energyPJ
}

// =============================================================================
// WORKLOAD TILING AND MAPPING
// =============================================================================

// TilingOptimizer optimizes workload tiling for CIM arrays.
type TilingOptimizer struct {
	Config       *TilingConfig
	ArraySpec    *ArraySpec
	BestTiling   *TilingStrategy
	SearchSpace  []*TilingStrategy
	Performance  map[string]*TilingPerformance
}

// TilingConfig configures the tiling optimizer.
type TilingConfig struct {
	MaxTileRows     int
	MaxTileCols     int
	MinTileRows     int
	MinTileCols     int
	SearchIterations int
	OptimizeFor     TilingObjective
}

// TilingObjective defines what to optimize for.
type TilingObjective int

const (
	TilingObjectiveLatency TilingObjective = iota
	TilingObjectiveEnergy
	TilingObjectiveUtilization
	TilingObjectiveBalanced
)

// ArraySpec describes the target CIM array.
type ArraySpec struct {
	Rows           int
	Cols           int
	BankCount      int
	WordlineGroups int
	BitlineGroups  int
}

// TilingStrategy represents a specific tiling approach.
type TilingStrategy struct {
	TileRows       int
	TileCols       int
	RowTileCount   int
	ColTileCount   int
	DataflowType   DataflowType
	WeightReuse    int
	InputReuse     int
	OutputReuse    int
}

// DataflowType defines the dataflow strategy.
type DataflowType int

const (
	DataflowOutputStationary DataflowType = iota
	DataflowWeightStationary
	DataflowInputStationary
	DataflowRowStationary
)

// TilingPerformance holds performance metrics for a tiling strategy.
type TilingPerformance struct {
	LatencyCycles   int64
	EnergyPJ        float64
	Utilization     float64
	DataMovementMB  float64
	ComputeToMemory float64
}

// NewTilingOptimizer creates a new tiling optimizer.
func NewTilingOptimizer(config *TilingConfig, arraySpec *ArraySpec) *TilingOptimizer {
	return &TilingOptimizer{
		Config:      config,
		ArraySpec:   arraySpec,
		Performance: make(map[string]*TilingPerformance),
	}
}

// GenerateSearchSpace creates candidate tiling strategies.
func (t *TilingOptimizer) GenerateSearchSpace(weightShape, inputShape []int) {
	t.SearchSpace = nil

	// Weight shape: [outChannels, inChannels] or [outChannels, inChannels, kH, kW]
	// Input shape: [batch, channels, height, width]

	outSize := weightShape[0]
	inSize := weightShape[1]
	if len(weightShape) > 2 {
		inSize *= weightShape[2] * weightShape[3] // Flatten kernel
	}

	// Generate tile sizes that divide evenly
	for tileRows := t.Config.MinTileRows; tileRows <= t.Config.MaxTileRows; tileRows *= 2 {
		for tileCols := t.Config.MinTileCols; tileCols <= t.Config.MaxTileCols; tileCols *= 2 {
			if tileRows > outSize || tileCols > inSize {
				continue
			}

			rowTiles := (outSize + tileRows - 1) / tileRows
			colTiles := (inSize + tileCols - 1) / tileCols

			// Try different dataflow types
			for df := DataflowOutputStationary; df <= DataflowRowStationary; df++ {
				var wReuse, iReuse, oReuse int
				switch df {
				case DataflowOutputStationary:
					oReuse = colTiles
					wReuse = 1
					iReuse = rowTiles
				case DataflowWeightStationary:
					wReuse = inputShape[0] // batch size
					iReuse = 1
					oReuse = 1
				case DataflowInputStationary:
					iReuse = rowTiles
					wReuse = 1
					oReuse = 1
				case DataflowRowStationary:
					wReuse = inputShape[0]
					iReuse = tileCols / t.ArraySpec.Rows
					oReuse = tileRows / t.ArraySpec.Cols
				}

				strategy := &TilingStrategy{
					TileRows:     tileRows,
					TileCols:     tileCols,
					RowTileCount: rowTiles,
					ColTileCount: colTiles,
					DataflowType: df,
					WeightReuse:  max(1, wReuse),
					InputReuse:   max(1, iReuse),
					OutputReuse:  max(1, oReuse),
				}
				t.SearchSpace = append(t.SearchSpace, strategy)
			}
		}
	}
}

// EvaluateStrategy evaluates a tiling strategy's performance.
func (t *TilingOptimizer) EvaluateStrategy(strategy *TilingStrategy, batchSize int) *TilingPerformance {
	key := fmt.Sprintf("%d_%d_%d", strategy.TileRows, strategy.TileCols, strategy.DataflowType)

	if perf, ok := t.Performance[key]; ok {
		return perf
	}

	totalMACs := int64(strategy.RowTileCount) * int64(strategy.ColTileCount) *
		int64(strategy.TileRows) * int64(strategy.TileCols) * int64(batchSize)

	// Latency calculation
	arrayCapacity := int64(t.ArraySpec.Rows * t.ArraySpec.Cols)
	macsPerCycle := arrayCapacity
	latencyCycles := totalMACs / macsPerCycle

	// Data movement calculation
	weightDataMB := float64(strategy.RowTileCount*strategy.ColTileCount*
		strategy.TileRows*strategy.TileCols*2) / (1024 * 1024)
	inputDataMB := float64(strategy.ColTileCount*strategy.TileCols*batchSize*2) / (1024 * 1024)
	outputDataMB := float64(strategy.RowTileCount*strategy.TileRows*batchSize*2) / (1024 * 1024)

	// Apply reuse factors
	totalDataMB := weightDataMB/float64(strategy.WeightReuse) +
		inputDataMB/float64(strategy.InputReuse) +
		outputDataMB/float64(strategy.OutputReuse)

	// Energy calculation (simplified)
	computeEnergy := float64(totalMACs) * 0.5  // 0.5 pJ per MAC
	memoryEnergy := totalDataMB * 1024 * 1024 * 0.1 // 0.1 pJ per byte
	totalEnergy := computeEnergy + memoryEnergy

	// Utilization
	utilization := float64(totalMACs) / float64(latencyCycles*arrayCapacity)

	perf := &TilingPerformance{
		LatencyCycles:   latencyCycles,
		EnergyPJ:        totalEnergy,
		Utilization:     utilization,
		DataMovementMB:  totalDataMB,
		ComputeToMemory: float64(totalMACs) / (totalDataMB * 1024 * 1024),
	}

	t.Performance[key] = perf
	return perf
}

// FindOptimalTiling searches for the best tiling strategy.
func (t *TilingOptimizer) FindOptimalTiling(batchSize int) *TilingStrategy {
	var bestScore float64
	var bestStrategy *TilingStrategy

	for _, strategy := range t.SearchSpace {
		perf := t.EvaluateStrategy(strategy, batchSize)

		var score float64
		switch t.Config.OptimizeFor {
		case TilingObjectiveLatency:
			score = 1.0 / float64(perf.LatencyCycles)
		case TilingObjectiveEnergy:
			score = 1.0 / perf.EnergyPJ
		case TilingObjectiveUtilization:
			score = perf.Utilization
		case TilingObjectiveBalanced:
			score = perf.Utilization / (float64(perf.LatencyCycles) * perf.EnergyPJ / 1e12)
		}

		if bestStrategy == nil || score > bestScore {
			bestScore = score
			bestStrategy = strategy
		}
	}

	t.BestTiling = bestStrategy
	return bestStrategy
}

// =============================================================================
// DIFFUSION MODEL CIM ACCELERATOR
// =============================================================================

// DiffusionCIMAccelerator simulates CIM-based diffusion model acceleration.
type DiffusionCIMAccelerator struct {
	Config           *DiffusionCIMConfig
	UNetTiles        []*CIMTileBank
	AttentionTiles   []*CIMTileBank
	DenoiseScheduler *DenoiseScheduler
	Stats            *DiffusionStats
}

// DiffusionCIMConfig configures the diffusion accelerator.
type DiffusionCIMConfig struct {
	LatentChannels   int
	LatentSize       int
	TimestepCount    int
	UNetLayers       int
	AttentionHeads   int
	HiddenDim        int
	UseHeterogeneous bool     // Tri-gear heterogeneous CIM
	QuantBits        []int    // Per-layer quantization
	EnableSimilarity bool     // Exploit denoising similarity
	SimilarityThresh float64
}

// CIMTileBank represents a bank of CIM tiles.
type CIMTileBank struct {
	BankID       int
	TileCount    int
	TileSize     int
	Precision    int    // Bits
	TileType     string // "conv", "attention", "mlp"
	Utilization  float64
	EnergyPJ     float64
}

// DenoiseScheduler manages denoising step scheduling.
type DenoiseScheduler struct {
	TotalSteps        int
	CurrentStep       int
	StepLatencies     []int64
	StepEnergies      []float64
	SkippedCompute    []float64 // Fraction skipped per step
	SimilarityScores  []float64
}

// DiffusionStats tracks accelerator statistics.
type DiffusionStats struct {
	TotalLatencyCycles int64
	TotalEnergyPJ      float64
	ThroughputTOPS     float64
	EnergyEffTOPSW     float64
	ComputeSkipped     float64
	MemoryBandwidthGB  float64
}

// NewDiffusionCIMAccelerator creates a new diffusion accelerator.
func NewDiffusionCIMAccelerator(config *DiffusionCIMConfig) *DiffusionCIMAccelerator {
	// Create UNet tile banks
	unetTiles := make([]*CIMTileBank, config.UNetLayers)
	for i := 0; i < config.UNetLayers; i++ {
		precision := 8
		if config.QuantBits != nil && i < len(config.QuantBits) {
			precision = config.QuantBits[i]
		}

		unetTiles[i] = &CIMTileBank{
			BankID:    i,
			TileCount: 16,
			TileSize:  256,
			Precision: precision,
			TileType:  "conv",
		}
	}

	// Create attention tile banks
	attentionBanks := config.AttentionHeads
	attentionTiles := make([]*CIMTileBank, attentionBanks)
	for i := 0; i < attentionBanks; i++ {
		attentionTiles[i] = &CIMTileBank{
			BankID:    i,
			TileCount: 8,
			TileSize:  256,
			Precision: 8,
			TileType:  "attention",
		}
	}

	return &DiffusionCIMAccelerator{
		Config:         config,
		UNetTiles:      unetTiles,
		AttentionTiles: attentionTiles,
		DenoiseScheduler: &DenoiseScheduler{
			TotalSteps:       config.TimestepCount,
			StepLatencies:    make([]int64, config.TimestepCount),
			StepEnergies:     make([]float64, config.TimestepCount),
			SkippedCompute:   make([]float64, config.TimestepCount),
			SimilarityScores: make([]float64, config.TimestepCount),
		},
		Stats: &DiffusionStats{},
	}
}

// RunDenoising simulates the full denoising process.
func (d *DiffusionCIMAccelerator) RunDenoising() *DiffusionStats {
	var prevLatent []float64 // Previous step's latent

	for step := 0; step < d.Config.TimestepCount; step++ {
		d.DenoiseScheduler.CurrentStep = step

		// Generate current latent (simulated)
		currentLatent := d.generateLatent()

		// Calculate similarity to previous step
		similarity := 0.0
		if prevLatent != nil {
			similarity = d.calculateSimilarity(prevLatent, currentLatent)
		}
		d.DenoiseScheduler.SimilarityScores[step] = similarity

		// Determine compute to skip based on similarity
		skipFraction := 0.0
		if d.Config.EnableSimilarity && similarity > d.Config.SimilarityThresh {
			skipFraction = (similarity - d.Config.SimilarityThresh) / (1.0 - d.Config.SimilarityThresh)
			skipFraction = math.Min(skipFraction, 0.5) // Cap at 50% skip
		}
		d.DenoiseScheduler.SkippedCompute[step] = skipFraction

		// Run UNet forward pass
		stepLatency, stepEnergy := d.runUNetStep(step, skipFraction)

		d.DenoiseScheduler.StepLatencies[step] = stepLatency
		d.DenoiseScheduler.StepEnergies[step] = stepEnergy

		d.Stats.TotalLatencyCycles += stepLatency
		d.Stats.TotalEnergyPJ += stepEnergy
		d.Stats.ComputeSkipped += skipFraction

		prevLatent = currentLatent
	}

	// Calculate final statistics
	d.calculateFinalStats()

	return d.Stats
}

// generateLatent simulates latent tensor generation.
func (d *DiffusionCIMAccelerator) generateLatent() []float64 {
	size := d.Config.LatentChannels * d.Config.LatentSize * d.Config.LatentSize
	latent := make([]float64, size)
	for i := range latent {
		latent[i] = rand.NormFloat64()
	}
	return latent
}

// calculateSimilarity computes cosine similarity between latents.
func (d *DiffusionCIMAccelerator) calculateSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// runUNetStep simulates one UNet forward pass.
func (d *DiffusionCIMAccelerator) runUNetStep(step int, skipFraction float64) (latency int64, energy float64) {
	effectiveCompute := 1.0 - skipFraction

	// Calculate total MACs for UNet
	latentElements := d.Config.LatentChannels * d.Config.LatentSize * d.Config.LatentSize
	hiddenElements := d.Config.HiddenDim * d.Config.LatentSize * d.Config.LatentSize

	// Encoder: latent -> hidden
	encoderMACs := int64(latentElements * d.Config.HiddenDim)

	// Middle blocks with attention
	attentionMACs := int64(d.Config.AttentionHeads) * int64(hiddenElements) * int64(d.Config.LatentSize*d.Config.LatentSize)

	// Decoder: hidden -> latent
	decoderMACs := int64(hiddenElements * d.Config.LatentChannels)

	totalMACs := encoderMACs + attentionMACs + decoderMACs
	effectiveMACs := int64(float64(totalMACs) * effectiveCompute)

	// Calculate latency (cycles)
	// Assume 256x256 CIM array, ~1 MAC per cycle per element
	arraysUsed := len(d.UNetTiles) + len(d.AttentionTiles)
	arrayCapacity := int64(256 * 256 * arraysUsed)
	latency = effectiveMACs / arrayCapacity

	// Calculate energy
	// Heterogeneous: different precision per layer
	if d.Config.UseHeterogeneous {
		for i, tile := range d.UNetTiles {
			layerMACs := effectiveMACs / int64(len(d.UNetTiles))
			precisionFactor := float64(tile.Precision) / 8.0
			tile.EnergyPJ = float64(layerMACs) * 0.5 * precisionFactor
			energy += tile.EnergyPJ
			tile.Utilization = float64(layerMACs) / float64(latency*int64(tile.TileCount*tile.TileSize*tile.TileSize))
			_ = i
		}
	} else {
		energy = float64(effectiveMACs) * 0.5 // 0.5 pJ per MAC
	}

	// Attention energy
	for _, tile := range d.AttentionTiles {
		attMACs := attentionMACs / int64(len(d.AttentionTiles))
		tile.EnergyPJ = float64(attMACs) * 0.6 * effectiveCompute
		energy += tile.EnergyPJ
	}

	return latency, energy
}

// calculateFinalStats computes final performance metrics.
func (d *DiffusionCIMAccelerator) calculateFinalStats() {
	// Total operations
	latentSize := d.Config.LatentChannels * d.Config.LatentSize * d.Config.LatentSize
	opsPerStep := int64(latentSize * d.Config.HiddenDim * 3) // encoder + middle + decoder
	totalOps := opsPerStep * int64(d.Config.TimestepCount)

	// Throughput in TOPS (Tera Operations Per Second)
	// Assume 1 GHz clock
	clockFreqGHz := 1.0
	latencySeconds := float64(d.Stats.TotalLatencyCycles) / (clockFreqGHz * 1e9)
	d.Stats.ThroughputTOPS = float64(totalOps) / (latencySeconds * 1e12)

	// Energy efficiency in TOPS/W
	powerW := d.Stats.TotalEnergyPJ / (latencySeconds * 1e12)
	if powerW > 0 {
		d.Stats.EnergyEffTOPSW = d.Stats.ThroughputTOPS / powerW
	}

	// Average compute skipped
	d.Stats.ComputeSkipped /= float64(d.Config.TimestepCount)

	// Memory bandwidth (simplified estimate)
	bytesPerStep := int64(latentSize*2 + d.Config.HiddenDim*latentSize*2)
	totalBytes := bytesPerStep * int64(d.Config.TimestepCount)
	d.Stats.MemoryBandwidthGB = float64(totalBytes) / (latencySeconds * 1e9)
}

// =============================================================================
// TRI-GEAR HETEROGENEOUS CIM (AIG-CIM STYLE)
// =============================================================================

// TriGearCIM implements tri-gear heterogeneous CIM for diffusion models.
type TriGearCIM struct {
	Config        *TriGearConfig
	IntGear       *GearConfig // Integer processing
	FPGear        *GearConfig // Floating-point processing
	SparseGear    *GearConfig // Sparse/pruned processing
	GearScheduler *GearScheduler
	Stats         *TriGearStats
}

// TriGearConfig configures the tri-gear CIM.
type TriGearConfig struct {
	TotalTiles       int
	IntTileRatio     float64 // Fraction for integer
	FPTileRatio      float64 // Fraction for FP
	SparseTileRatio  float64 // Fraction for sparse
	EnableBoothMul   bool    // Booth-8 multiplier
	EnableExpAccel   bool    // Exponent acceleration
	EnableRedundancy bool    // In-memory redundancy search
}

// GearConfig configures a single gear type.
type GearConfig struct {
	GearType      string
	TileCount     int
	Precision     int
	SparsityRatio float64
	EnergyPerOp   float64
}

// GearScheduler schedules operations across gears.
type GearScheduler struct {
	IntQueue    []*GearOperation
	FPQueue     []*GearOperation
	SparseQueue []*GearOperation
}

// GearOperation represents a scheduled operation.
type GearOperation struct {
	OpID        int
	LayerID     int
	GearType    string
	StartCycle  int64
	EndCycle    int64
	MACs        int64
	Sparsity    float64
}

// TriGearStats tracks tri-gear performance.
type TriGearStats struct {
	IntLatency     int64
	FPLatency      int64
	SparseLatency  int64
	TotalLatency   int64
	TotalEnergy    float64
	SparsityGain   float64
	BoothSpeedup   float64
}

// NewTriGearCIM creates a new tri-gear CIM accelerator.
func NewTriGearCIM(config *TriGearConfig) *TriGearCIM {
	intTiles := int(float64(config.TotalTiles) * config.IntTileRatio)
	fpTiles := int(float64(config.TotalTiles) * config.FPTileRatio)
	sparseTiles := config.TotalTiles - intTiles - fpTiles

	return &TriGearCIM{
		Config: config,
		IntGear: &GearConfig{
			GearType:    "INT8",
			TileCount:   intTiles,
			Precision:   8,
			EnergyPerOp: 0.3,
		},
		FPGear: &GearConfig{
			GearType:    "BF16",
			TileCount:   fpTiles,
			Precision:   16,
			EnergyPerOp: 0.8,
		},
		SparseGear: &GearConfig{
			GearType:      "Sparse",
			TileCount:     sparseTiles,
			Precision:     8,
			SparsityRatio: 0.5,
			EnergyPerOp:   0.2,
		},
		GearScheduler: &GearScheduler{},
		Stats:         &TriGearStats{},
	}
}

// ScheduleLayer schedules a layer across appropriate gears.
func (t *TriGearCIM) ScheduleLayer(layerID int, layerType string, macs int64, sparsity float64) {
	// Determine gear assignment based on layer type
	var gear *GearConfig
	var queue *[]*GearOperation

	switch layerType {
	case "attention":
		// Attention uses FP for softmax precision
		gear = t.FPGear
		queue = &t.GearScheduler.FPQueue
	case "conv", "linear":
		if sparsity > 0.3 {
			// High sparsity: use sparse gear
			gear = t.SparseGear
			queue = &t.GearScheduler.SparseQueue
		} else {
			// Dense: use INT gear with Booth multiplier
			gear = t.IntGear
			queue = &t.GearScheduler.IntQueue
		}
	default:
		gear = t.IntGear
		queue = &t.GearScheduler.IntQueue
	}

	// Calculate latency
	effectiveMACs := macs
	if gear.GearType == "Sparse" {
		effectiveMACs = int64(float64(macs) * (1 - sparsity))
	}

	tileCapacity := int64(256 * 256 * gear.TileCount)
	latency := effectiveMACs / tileCapacity
	if t.Config.EnableBoothMul && gear.GearType == "INT8" {
		latency = latency * 3 / 4 // Booth-8 gives ~25% speedup
	}

	op := &GearOperation{
		OpID:       len(*queue),
		LayerID:    layerID,
		GearType:   gear.GearType,
		StartCycle: 0, // Will be scheduled
		EndCycle:   latency,
		MACs:       macs,
		Sparsity:   sparsity,
	}

	*queue = append(*queue, op)
}

// Execute runs the scheduled operations.
func (t *TriGearCIM) Execute() *TriGearStats {
	// Execute each gear in parallel
	intLatency := t.executeGear(t.GearScheduler.IntQueue, t.IntGear)
	fpLatency := t.executeGear(t.GearScheduler.FPQueue, t.FPGear)
	sparseLatency := t.executeGear(t.GearScheduler.SparseQueue, t.SparseGear)

	t.Stats.IntLatency = intLatency
	t.Stats.FPLatency = fpLatency
	t.Stats.SparseLatency = sparseLatency
	t.Stats.TotalLatency = max64(intLatency, max64(fpLatency, sparseLatency))

	// Calculate energy
	for _, op := range t.GearScheduler.IntQueue {
		effectiveMACs := op.MACs
		if t.Config.EnableBoothMul {
			effectiveMACs = effectiveMACs * 3 / 4
		}
		t.Stats.TotalEnergy += float64(effectiveMACs) * t.IntGear.EnergyPerOp
	}

	for _, op := range t.GearScheduler.FPQueue {
		t.Stats.TotalEnergy += float64(op.MACs) * t.FPGear.EnergyPerOp
		if t.Config.EnableExpAccel {
			t.Stats.TotalEnergy *= 0.8 // Exponent acceleration saves ~20%
		}
	}

	for _, op := range t.GearScheduler.SparseQueue {
		effectiveMACs := int64(float64(op.MACs) * (1 - op.Sparsity))
		t.Stats.TotalEnergy += float64(effectiveMACs) * t.SparseGear.EnergyPerOp
	}

	// Calculate speedup metrics
	if t.Config.EnableBoothMul {
		t.Stats.BoothSpeedup = 1.25
	}

	totalSparseMACs := int64(0)
	totalDenseMACs := int64(0)
	for _, op := range t.GearScheduler.SparseQueue {
		totalSparseMACs += int64(float64(op.MACs) * op.Sparsity)
		totalDenseMACs += op.MACs
	}
	if totalDenseMACs > 0 {
		t.Stats.SparsityGain = float64(totalSparseMACs) / float64(totalDenseMACs)
	}

	return t.Stats
}

// executeGear runs operations for a single gear.
func (t *TriGearCIM) executeGear(queue []*GearOperation, gear *GearConfig) int64 {
	if len(queue) == 0 {
		return 0
	}

	// Simple sequential scheduling
	currentCycle := int64(0)
	for _, op := range queue {
		op.StartCycle = currentCycle
		op.EndCycle = currentCycle + op.EndCycle
		currentCycle = op.EndCycle
	}

	return currentCycle
}

// =============================================================================
// DENOISING STEP-AWARE SCHEDULING (DDSM)
// =============================================================================

// DDSMScheduler implements denoising diffusion step-aware model scheduling.
type DDSMScheduler struct {
	Config           *DDSMConfig
	StepCapacities   []float64 // Network capacity per step
	StepImportance   []float64 // Importance score per step
	PrunedArchs      []*PrunedArchitecture
	TotalFLOPs       float64
	ReducedFLOPs     float64
}

// DDSMConfig configures the DDSM scheduler.
type DDSMConfig struct {
	TotalSteps       int
	BaseModelParams  int64
	BaseFLOPs        float64
	TargetReduction  float64 // Target FLOPs reduction (e.g., 0.5 for 50%)
	SearchIterations int
	MinCapacity      float64 // Minimum model capacity (0-1)
}

// PrunedArchitecture represents a step-specific pruned architecture.
type PrunedArchitecture struct {
	Step          int
	CapacityRatio float64
	Channels      []int
	FLOPs         float64
	Importance    float64
}

// NewDDSMScheduler creates a new DDSM scheduler.
func NewDDSMScheduler(config *DDSMConfig) *DDSMScheduler {
	return &DDSMScheduler{
		Config:         config,
		StepCapacities: make([]float64, config.TotalSteps),
		StepImportance: make([]float64, config.TotalSteps),
		TotalFLOPs:     config.BaseFLOPs * float64(config.TotalSteps),
	}
}

// AnalyzeStepImportance determines importance of each denoising step.
func (d *DDSMScheduler) AnalyzeStepImportance() {
	// Importance typically follows a pattern:
	// - Early steps (high noise): less important, more redundant
	// - Middle steps: moderate importance
	// - Late steps (low noise, detail generation): high importance

	for step := 0; step < d.Config.TotalSteps; step++ {
		t := float64(step) / float64(d.Config.TotalSteps-1)

		// Importance curve: higher at extremes, especially late steps
		importance := 0.3 + 0.3*math.Pow(t, 2) + 0.4*math.Pow(t, 3)
		d.StepImportance[step] = importance
	}

	// Normalize
	maxImp := 0.0
	for _, imp := range d.StepImportance {
		if imp > maxImp {
			maxImp = imp
		}
	}
	for i := range d.StepImportance {
		d.StepImportance[i] /= maxImp
	}
}

// SearchOptimalCapacities finds optimal network capacity per step.
func (d *DDSMScheduler) SearchOptimalCapacities() {
	d.AnalyzeStepImportance()

	// Evolutionary search for capacity allocation
	bestCapacities := make([]float64, d.Config.TotalSteps)
	bestFLOPs := d.TotalFLOPs

	for iter := 0; iter < d.Config.SearchIterations; iter++ {
		// Generate candidate capacities based on importance
		candidates := make([]float64, d.Config.TotalSteps)
		for step := 0; step < d.Config.TotalSteps; step++ {
			// Base capacity from importance
			base := d.StepImportance[step]

			// Add noise for exploration
			noise := (rand.Float64() - 0.5) * 0.2
			capacity := math.Max(d.Config.MinCapacity, math.Min(1.0, base+noise))
			candidates[step] = capacity
		}

		// Calculate FLOPs for this allocation
		totalFLOPs := 0.0
		for step := 0; step < d.Config.TotalSteps; step++ {
			// FLOPs scale quadratically with capacity (channels squared)
			stepFLOPs := d.Config.BaseFLOPs * math.Pow(candidates[step], 2)
			totalFLOPs += stepFLOPs
		}

		// Check if this meets target and is better
		reduction := 1.0 - totalFLOPs/d.TotalFLOPs
		if reduction >= d.Config.TargetReduction && totalFLOPs < bestFLOPs {
			copy(bestCapacities, candidates)
			bestFLOPs = totalFLOPs
		}
	}

	d.StepCapacities = bestCapacities
	d.ReducedFLOPs = bestFLOPs
}

// GeneratePrunedArchitectures creates step-specific architectures.
func (d *DDSMScheduler) GeneratePrunedArchitectures(baseChannels []int) []*PrunedArchitecture {
	d.SearchOptimalCapacities()

	d.PrunedArchs = make([]*PrunedArchitecture, d.Config.TotalSteps)

	for step := 0; step < d.Config.TotalSteps; step++ {
		capacity := d.StepCapacities[step]

		// Scale channels by capacity
		prunedChannels := make([]int, len(baseChannels))
		for i, ch := range baseChannels {
			prunedChannels[i] = max(1, int(float64(ch)*capacity))
		}

		// Calculate FLOPs
		stepFLOPs := d.Config.BaseFLOPs * math.Pow(capacity, 2)

		d.PrunedArchs[step] = &PrunedArchitecture{
			Step:          step,
			CapacityRatio: capacity,
			Channels:      prunedChannels,
			FLOPs:         stepFLOPs,
			Importance:    d.StepImportance[step],
		}
	}

	return d.PrunedArchs
}

// GetFLOPsReduction returns the achieved FLOPs reduction.
func (d *DDSMScheduler) GetFLOPsReduction() float64 {
	return 1.0 - d.ReducedFLOPs/d.TotalFLOPs
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max64(a, b int64) int64 {
	if a > b {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// DEMONSTRATION AND BENCHMARKING
// =============================================================================

// RunSchedulingDemo demonstrates the scheduling capabilities.
func RunSchedulingDemo() {
	fmt.Println("=== CIM Scheduling Demo ===")
	fmt.Println()

	// Create CLSA-CIM scheduler
	config := &CLSAConfig{
		EnableCrossLayer:    true,
		EnableWeightDup:     true,
		DuplicationFactor:   2,
		PipelineDepth:       4,
		PartialResultSize:   1024,
		MaxActiveTiles:      32,
		UtilizationTarget:   0.3,
		SchedulingHeuristic: HeuristicCrossLayer,
	}

	scheduler := NewCLSACIMScheduler(config, 64)
	scheduler.SetNoC(&NoCConfig{
		TopologyType:    NoCTopologyMesh,
		Rows:            8,
		Cols:            8,
		LinkBandwidthGB: 100.0,
		HopLatencyNS:    1.0,
		RouterLatencyNS: 0.5,
	})

	// Map VGG16-style layers
	layerSizes := []int64{
		64 * 3 * 3 * 3,    // conv1
		64 * 64 * 3 * 3,   // conv2
		128 * 64 * 3 * 3,  // conv3
		128 * 128 * 3 * 3, // conv4
		256 * 128 * 3 * 3, // conv5
		256 * 256 * 3 * 3, // conv6
		512 * 256 * 3 * 3, // conv7
		512 * 512 * 3 * 3, // conv8
		4096 * 512 * 7 * 7, // fc1
		4096 * 4096,        // fc2
		1000 * 4096,        // fc3
	}

	for i, size := range layerSizes {
		deps := []int{}
		if i > 0 {
			deps = []int{i - 1}
		}
		consumers := []int{}
		if i < len(layerSizes)-1 {
			consumers = []int{i + 1}
		}
		scheduler.MapLayer(i, fmt.Sprintf("layer%d", i), "conv", size, deps, consumers)
	}

	// Run scheduling
	scheduler.ScheduleInference()

	fmt.Printf("Cross-Layer Scheduling Results:\n")
	fmt.Printf("  Total Latency: %d cycles\n", scheduler.TotalLatency)
	fmt.Printf("  PE Utilization: %.1f%%\n", scheduler.PEUtilization*100)
	fmt.Printf("  Speedup vs Layer-by-Layer: %.1fx\n", scheduler.GetSpeedup())
	fmt.Println()
}

// RunDiffusionDemo demonstrates diffusion model acceleration.
func RunDiffusionDemo() {
	fmt.Println("=== Diffusion Model CIM Demo ===")
	fmt.Println()

	// Create diffusion accelerator
	config := &DiffusionCIMConfig{
		LatentChannels:   4,
		LatentSize:       64,
		TimestepCount:    50,
		UNetLayers:       12,
		AttentionHeads:   8,
		HiddenDim:        1280,
		UseHeterogeneous: true,
		QuantBits:        []int{8, 8, 8, 6, 6, 6, 6, 6, 6, 8, 8, 8},
		EnableSimilarity: true,
		SimilarityThresh: 0.7,
	}

	accel := NewDiffusionCIMAccelerator(config)
	stats := accel.RunDenoising()

	fmt.Printf("Diffusion Inference Results:\n")
	fmt.Printf("  Total Latency: %d cycles\n", stats.TotalLatencyCycles)
	fmt.Printf("  Total Energy: %.2f mJ\n", stats.TotalEnergyPJ/1e9)
	fmt.Printf("  Throughput: %.2f TOPS\n", stats.ThroughputTOPS)
	fmt.Printf("  Energy Efficiency: %.2f TOPS/W\n", stats.EnergyEffTOPSW)
	fmt.Printf("  Compute Skipped: %.1f%%\n", stats.ComputeSkipped*100)
	fmt.Println()

	// Tri-gear CIM
	triGear := NewTriGearCIM(&TriGearConfig{
		TotalTiles:       64,
		IntTileRatio:     0.5,
		FPTileRatio:      0.3,
		SparseTileRatio:  0.2,
		EnableBoothMul:   true,
		EnableExpAccel:   true,
		EnableRedundancy: true,
	})

	// Schedule typical diffusion layers
	triGear.ScheduleLayer(0, "conv", 1000000, 0.1)
	triGear.ScheduleLayer(1, "attention", 2000000, 0.0)
	triGear.ScheduleLayer(2, "linear", 500000, 0.4)
	triGear.ScheduleLayer(3, "conv", 1000000, 0.2)

	triStats := triGear.Execute()

	fmt.Printf("Tri-Gear CIM Results:\n")
	fmt.Printf("  INT Gear Latency: %d cycles\n", triStats.IntLatency)
	fmt.Printf("  FP Gear Latency: %d cycles\n", triStats.FPLatency)
	fmt.Printf("  Sparse Gear Latency: %d cycles\n", triStats.SparseLatency)
	fmt.Printf("  Total Latency: %d cycles\n", triStats.TotalLatency)
	fmt.Printf("  Total Energy: %.2f mJ\n", triStats.TotalEnergy/1e9)
	fmt.Printf("  Booth Speedup: %.2fx\n", triStats.BoothSpeedup)
	fmt.Printf("  Sparsity Gain: %.1f%%\n", triStats.SparsityGain*100)
	fmt.Println()
}

// RunDDSMDemo demonstrates step-aware model scheduling.
func RunDDSMDemo() {
	fmt.Println("=== DDSM Step-Aware Scheduling Demo ===")
	fmt.Println()

	scheduler := NewDDSMScheduler(&DDSMConfig{
		TotalSteps:       50,
		BaseModelParams:  35750000, // ~35.75M params (CIFAR-10 UNet)
		BaseFLOPs:        12.14e9,  // 12.14T FLOPs total
		TargetReduction:  0.5,      // Target 50% reduction
		SearchIterations: 100,
		MinCapacity:      0.3,
	})

	baseChannels := []int{64, 128, 256, 512, 512, 256, 128, 64}
	archs := scheduler.GeneratePrunedArchitectures(baseChannels)

	fmt.Printf("DDSM Results:\n")
	fmt.Printf("  Original FLOPs: %.2f TFLOPs\n", scheduler.TotalFLOPs/1e12)
	fmt.Printf("  Reduced FLOPs: %.2f TFLOPs\n", scheduler.ReducedFLOPs/1e12)
	fmt.Printf("  FLOPs Reduction: %.1f%%\n", scheduler.GetFLOPsReduction()*100)
	fmt.Println()

	fmt.Println("Sample Step Architectures:")
	for _, step := range []int{0, 10, 25, 40, 49} {
		if step < len(archs) {
			arch := archs[step]
			fmt.Printf("  Step %d: Capacity=%.2f, FLOPs=%.2fG, Importance=%.2f\n",
				step, arch.CapacityRatio, arch.FLOPs/1e9, arch.Importance)
		}
	}
	fmt.Println()
}

// memory_latency.go - Memory Hierarchy and Latency Modeling for CIM
//
// This module implements:
// - Three-level memory hierarchy (DRAM, on-chip buffers, tile memory)
// - Weight-stationary and output-stationary dataflows
// - Buffer traffic estimation and optimization
// - Inference latency modeling with pipeline stages
// - Throughput calculation (FPS, TOPS)
//
// Based on research findings:
// - Buffer traffic reduction: 77-87% with optimized dataflow
// - Energy reduction: 10-18% total hierarchy
// - Latency reduction: 15-28% for convolutions
// - Tile memory utilization: 84-87%
//
// References:
// - "CIM Dataflow for Minimal Buffer Traffic" (arXiv 2508.14375)
// - NeuroSim memory hierarchy modeling
// - ISAAC crossbar architecture

package layers

import (
	"math"
)

// ================== Memory Hierarchy ==================

// MemoryConfig configures memory hierarchy parameters
type MemoryConfig struct {
	// DRAM (Level 0 - off-chip)
	DRAMBandwidthGBps  float64 // DRAM bandwidth (GB/s)
	DRAMLatencyNs      float64 // DRAM access latency (ns)
	DRAMEnergyPerBit   float64 // Energy per bit (pJ)

	// On-chip buffers (Level 1)
	InputBufferKB      float64 // Input buffer size (KB)
	OutputBufferKB     float64 // Output buffer size (KB)
	WeightBufferKB     float64 // Weight buffer size (KB)
	BufferBandwidthGBps float64 // Buffer bandwidth (GB/s)
	BufferLatencyNs    float64 // Buffer access latency (ns)
	BufferEnergyPerBit float64 // Energy per bit (fJ)

	// Tile memory/register file (Level 2)
	TileMemoryBytes    int     // Tile memory size (bytes)
	TileRegFileBytes   int     // Tile register file size (bytes)
	TileLatencyNs      float64 // Tile access latency (ns)
	TileEnergyPerBit   float64 // Energy per bit (fJ)

	// Dataflow configuration
	Dataflow           string  // "weight_stationary", "output_stationary", "row_stationary"
	WeightReuse        int     // Weight reuse factor
	InputReuse         int     // Input activation reuse factor
}

// DefaultMemoryConfig returns typical CIM memory hierarchy parameters
func DefaultMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		DRAMBandwidthGBps:  25.6,   // DDR4-3200
		DRAMLatencyNs:      100,    // ~100 ns DRAM latency
		DRAMEnergyPerBit:   20,     // 20 pJ/bit

		InputBufferKB:      64,     // 64 KB input buffer
		OutputBufferKB:     32,     // 32 KB output buffer
		WeightBufferKB:     128,    // 128 KB weight buffer
		BufferBandwidthGBps: 100,   // 100 GB/s on-chip
		BufferLatencyNs:    5,      // 5 ns buffer access
		BufferEnergyPerBit: 0.5,    // 0.5 fJ/bit

		TileMemoryBytes:    180 * 8, // 180x8 = 1440 bytes (8T-SRAM)
		TileRegFileBytes:   256,     // 256 bytes TRF
		TileLatencyNs:      1,       // 1 ns tile access
		TileEnergyPerBit:   0.1,     // 0.1 fJ/bit

		Dataflow:           "weight_stationary",
		WeightReuse:        8,       // Weights reused 8x
		InputReuse:         4,       // Inputs reused 4x
	}
}

// MemoryLevel represents a level in the memory hierarchy
type MemoryLevel struct {
	Name          string
	SizeBytes     int64
	BandwidthGBps float64
	LatencyNs     float64
	EnergyPerBit  float64 // pJ or fJ depending on level

	// Statistics
	ReadBytes     int64
	WriteBytes    int64
	ReadCount     int64
	WriteCount    int64
	TotalEnergy   float64
	TotalLatency  float64
}

// MemoryHierarchy models the complete memory system
type MemoryHierarchy struct {
	Config *MemoryConfig
	Levels []*MemoryLevel

	// Buffers
	InputBuffer  *DataBuffer
	OutputBuffer *DataBuffer
	WeightBuffer *DataBuffer

	// Statistics
	TotalTraffic     int64   // Total bytes transferred
	BufferHits       int64
	DRAMAccesses     int64
	TotalEnergy      float64 // pJ
	UtilizationPct   float64
}

// DataBuffer represents an on-chip buffer
type DataBuffer struct {
	Name       string
	SizeBytes  int64
	UsedBytes  int64
	Data       []byte
	Tags       map[int64]bool // Address tags for hit detection
	Hits       int64
	Misses     int64
}

// NewDataBuffer creates a new data buffer
func NewDataBuffer(name string, sizeKB float64) *DataBuffer {
	sizeBytes := int64(sizeKB * 1024)
	return &DataBuffer{
		Name:      name,
		SizeBytes: sizeBytes,
		UsedBytes: 0,
		Data:      make([]byte, sizeBytes),
		Tags:      make(map[int64]bool),
		Hits:      0,
		Misses:    0,
	}
}

// Access checks buffer for data and returns hit/miss
func (b *DataBuffer) Access(address int64, size int64) bool {
	// Check if address range is in buffer
	if b.Tags[address] {
		b.Hits++
		return true
	}
	b.Misses++
	return false
}

// Load loads data into buffer
func (b *DataBuffer) Load(address int64, size int64) {
	// Evict if needed
	if b.UsedBytes+size > b.SizeBytes {
		b.Evict(size)
	}
	b.Tags[address] = true
	b.UsedBytes += size
}

// Evict removes data from buffer
func (b *DataBuffer) Evict(needed int64) {
	// Simple FIFO eviction
	for addr := range b.Tags {
		delete(b.Tags, addr)
		b.UsedBytes -= 8 // Assume 8-byte entries
		if b.UsedBytes+needed <= b.SizeBytes {
			break
		}
	}
}

// GetHitRate returns the buffer hit rate
func (b *DataBuffer) GetHitRate() float64 {
	total := b.Hits + b.Misses
	if total == 0 {
		return 0
	}
	return float64(b.Hits) / float64(total)
}

// NewMemoryHierarchy creates a memory hierarchy model
func NewMemoryHierarchy(config *MemoryConfig) *MemoryHierarchy {
	levels := []*MemoryLevel{
		{
			Name:          "DRAM",
			SizeBytes:     4 * 1024 * 1024 * 1024, // 4 GB
			BandwidthGBps: config.DRAMBandwidthGBps,
			LatencyNs:     config.DRAMLatencyNs,
			EnergyPerBit:  config.DRAMEnergyPerBit,
		},
		{
			Name:          "OnChipBuffer",
			SizeBytes:     int64((config.InputBufferKB + config.OutputBufferKB + config.WeightBufferKB) * 1024),
			BandwidthGBps: config.BufferBandwidthGBps,
			LatencyNs:     config.BufferLatencyNs,
			EnergyPerBit:  config.BufferEnergyPerBit / 1000, // Convert fJ to pJ
		},
		{
			Name:          "TileMemory",
			SizeBytes:     int64(config.TileMemoryBytes + config.TileRegFileBytes),
			BandwidthGBps: 1000, // Very high on-tile bandwidth
			LatencyNs:     config.TileLatencyNs,
			EnergyPerBit:  config.TileEnergyPerBit / 1000, // Convert fJ to pJ
		},
	}

	return &MemoryHierarchy{
		Config:       config,
		Levels:       levels,
		InputBuffer:  NewDataBuffer("InputBuffer", config.InputBufferKB),
		OutputBuffer: NewDataBuffer("OutputBuffer", config.OutputBufferKB),
		WeightBuffer: NewDataBuffer("WeightBuffer", config.WeightBufferKB),
	}
}

// AccessData models a data access through the hierarchy
func (mh *MemoryHierarchy) AccessData(dataType string, address int64, sizeBytes int64, isWrite bool) *MemoryAccessResult {
	result := &MemoryAccessResult{
		DataType:   dataType,
		SizeBytes:  sizeBytes,
		IsWrite:    isWrite,
	}

	// Check appropriate buffer
	var buffer *DataBuffer
	switch dataType {
	case "input":
		buffer = mh.InputBuffer
	case "output":
		buffer = mh.OutputBuffer
	case "weight":
		buffer = mh.WeightBuffer
	default:
		buffer = mh.InputBuffer
	}

	// Check for buffer hit
	if buffer.Access(address, sizeBytes) {
		result.HitLevel = 1 // Buffer hit
		result.Latency = mh.Config.BufferLatencyNs
		result.Energy = float64(sizeBytes*8) * mh.Config.BufferEnergyPerBit / 1000 // pJ
		mh.BufferHits++
	} else {
		// Buffer miss - need DRAM access
		result.HitLevel = 0 // DRAM

		// DRAM access time
		transferTime := float64(sizeBytes) / (mh.Config.DRAMBandwidthGBps * 1e9) * 1e9 // ns
		result.Latency = mh.Config.DRAMLatencyNs + transferTime

		// DRAM energy
		result.Energy = float64(sizeBytes*8) * mh.Config.DRAMEnergyPerBit // pJ

		// Load into buffer
		buffer.Load(address, sizeBytes)
		mh.DRAMAccesses++
	}

	// Update statistics
	mh.TotalTraffic += sizeBytes
	mh.TotalEnergy += result.Energy

	level := mh.Levels[result.HitLevel]
	if isWrite {
		level.WriteBytes += sizeBytes
		level.WriteCount++
	} else {
		level.ReadBytes += sizeBytes
		level.ReadCount++
	}
	level.TotalEnergy += result.Energy
	level.TotalLatency += result.Latency

	return result
}

// MemoryAccessResult contains results of a memory access
type MemoryAccessResult struct {
	DataType   string
	SizeBytes  int64
	IsWrite    bool
	HitLevel   int     // 0=DRAM, 1=Buffer, 2=Tile
	Latency    float64 // ns
	Energy     float64 // pJ
}

// SimulateLayerTraffic simulates memory traffic for a layer
func (mh *MemoryHierarchy) SimulateLayerTraffic(inputSize, outputSize, weightSize int64, batchSize int) *LayerTrafficResult {
	cfg := mh.Config
	result := &LayerTrafficResult{}

	// Weight traffic (loaded once per batch in weight-stationary)
	if cfg.Dataflow == "weight_stationary" {
		result.WeightTraffic = weightSize // Loaded once
		result.WeightReuses = int64(batchSize * cfg.WeightReuse)
	} else {
		result.WeightTraffic = weightSize * int64(batchSize)
		result.WeightReuses = int64(cfg.WeightReuse)
	}

	// Input traffic (with reuse factor)
	inputTrafficPerSample := inputSize / int64(cfg.InputReuse)
	result.InputTraffic = inputTrafficPerSample * int64(batchSize)

	// Output traffic (always written once per sample)
	result.OutputTraffic = outputSize * int64(batchSize)

	// Total traffic
	result.TotalTraffic = result.WeightTraffic + result.InputTraffic + result.OutputTraffic

	// Energy estimation
	// Assume mix of buffer and DRAM accesses
	bufferRatio := 0.8 // 80% buffer hits with good reuse
	dramRatio := 0.2

	bufferEnergy := float64(result.TotalTraffic*8) * cfg.BufferEnergyPerBit / 1000 * bufferRatio
	dramEnergy := float64(result.TotalTraffic*8) * cfg.DRAMEnergyPerBit * dramRatio
	result.TotalEnergy = bufferEnergy + dramEnergy

	// Latency estimation
	// Overlap DRAM and compute where possible
	dramTime := float64(result.TotalTraffic*int64(dramRatio)) / (cfg.DRAMBandwidthGBps * 1e9) * 1e9
	bufferTime := float64(result.TotalTraffic*int64(bufferRatio)) / (cfg.BufferBandwidthGBps * 1e9) * 1e9
	result.TransferLatency = math.Max(dramTime, bufferTime) // Overlapped

	return result
}

// LayerTrafficResult contains traffic analysis for a layer
type LayerTrafficResult struct {
	WeightTraffic   int64
	InputTraffic    int64
	OutputTraffic   int64
	TotalTraffic    int64
	WeightReuses    int64
	TotalEnergy     float64 // pJ
	TransferLatency float64 // ns
}

// GetStatistics returns memory hierarchy statistics
func (mh *MemoryHierarchy) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["TotalTrafficBytes"] = mh.TotalTraffic
	stats["TotalEnergyPJ"] = mh.TotalEnergy
	stats["BufferHits"] = mh.BufferHits
	stats["DRAMAccesses"] = mh.DRAMAccesses

	if mh.BufferHits+mh.DRAMAccesses > 0 {
		stats["BufferHitRate"] = float64(mh.BufferHits) / float64(mh.BufferHits+mh.DRAMAccesses)
	}

	stats["InputBufferHitRate"] = mh.InputBuffer.GetHitRate()
	stats["OutputBufferHitRate"] = mh.OutputBuffer.GetHitRate()
	stats["WeightBufferHitRate"] = mh.WeightBuffer.GetHitRate()

	return stats
}

// ================== Latency Modeling ==================

// LatencyConfig configures latency model parameters
type LatencyConfig struct {
	// Crossbar parameters
	ArrayRows        int
	ArrayCols        int
	CycleTimeNs      float64 // Cycle time (ns)
	ADCLatencyNs     float64 // ADC conversion time
	DACLatencyNs     float64 // DAC setup time

	// Pipeline stages
	NumPipelineStages int
	PipelineDepth     int

	// Activation function
	ActivationLatencyNs float64

	// Shift-add for bit-serial
	ShiftAddLatencyNs float64
	BitsPerCycle      int
}

// DefaultLatencyConfig returns typical latency parameters
func DefaultLatencyConfig() *LatencyConfig {
	return &LatencyConfig{
		ArrayRows:           64,
		ArrayCols:           64,
		CycleTimeNs:         100,    // 100 ns cycle
		ADCLatencyNs:        10,     // 10 ns ADC
		DACLatencyNs:        5,      // 5 ns DAC
		NumPipelineStages:   4,
		PipelineDepth:       8,
		ActivationLatencyNs: 2,      // 2 ns activation
		ShiftAddLatencyNs:   1,      // 1 ns shift-add
		BitsPerCycle:        1,      // Bit-serial
	}
}

// LatencyModel models inference latency
type LatencyModel struct {
	Config       *LatencyConfig
	Memory       *MemoryHierarchy

	// Pipeline state
	PipelineOccupancy []int
	CurrentStage      int

	// Statistics
	TotalCycles       int64
	ComputeCycles     int64
	MemoryCycles      int64
	PipelineStalls    int64
	Throughput        float64 // Operations per second
}

// NewLatencyModel creates a latency model
func NewLatencyModel(config *LatencyConfig, memConfig *MemoryConfig) *LatencyModel {
	return &LatencyModel{
		Config:            config,
		Memory:            NewMemoryHierarchy(memConfig),
		PipelineOccupancy: make([]int, config.NumPipelineStages),
		CurrentStage:      0,
	}
}

// MVMLatency calculates latency for a single MVM operation
func (lm *LatencyModel) MVMLatency(inputPrecision, weightPrecision int) *MVMLatencyResult {
	cfg := lm.Config
	result := &MVMLatencyResult{}

	// DAC latency (input conversion)
	result.DACLatency = cfg.DACLatencyNs

	// Crossbar computation
	// For bit-serial: need multiple cycles for full precision
	numBitCycles := (inputPrecision + cfg.BitsPerCycle - 1) / cfg.BitsPerCycle
	result.CrossbarLatency = float64(numBitCycles) * cfg.CycleTimeNs

	// ADC latency (output conversion)
	result.ADCLatency = cfg.ADCLatencyNs

	// Shift-add accumulation
	result.ShiftAddLatency = float64(inputPrecision) * cfg.ShiftAddLatencyNs

	// Activation function
	result.ActivationLatency = cfg.ActivationLatencyNs

	// Total latency (pipelined vs sequential)
	if cfg.NumPipelineStages > 1 {
		// Pipelined: critical path determines latency
		result.TotalLatency = math.Max(result.CrossbarLatency,
			math.Max(result.ADCLatency+result.ShiftAddLatency, result.DACLatency))
	} else {
		// Sequential: sum of all stages
		result.TotalLatency = result.DACLatency + result.CrossbarLatency +
			result.ADCLatency + result.ShiftAddLatency + result.ActivationLatency
	}

	// Update statistics
	lm.TotalCycles++
	lm.ComputeCycles++

	return result
}

// MVMLatencyResult contains MVM latency breakdown
type MVMLatencyResult struct {
	DACLatency        float64 // ns
	CrossbarLatency   float64 // ns
	ADCLatency        float64 // ns
	ShiftAddLatency   float64 // ns
	ActivationLatency float64 // ns
	TotalLatency      float64 // ns
}

// LayerLatency calculates latency for a complete layer
func (lm *LatencyModel) LayerLatency(inputRows, inputCols, outputCols int, batchSize int) *LayerLatencyResult {
	cfg := lm.Config
	result := &LayerLatencyResult{}

	// Number of tiles needed
	tilesX := (inputCols + cfg.ArrayRows - 1) / cfg.ArrayRows
	tilesY := (outputCols + cfg.ArrayCols - 1) / cfg.ArrayCols
	totalTiles := tilesX * tilesY

	// MVM latency per tile
	mvmResult := lm.MVMLatency(8, 8) // Assume 8-bit precision

	// Memory transfer latency
	inputBytes := int64(inputRows * inputCols)
	weightBytes := int64(inputCols * outputCols)
	outputBytes := int64(inputRows * outputCols)

	memResult := lm.Memory.SimulateLayerTraffic(inputBytes, outputBytes, weightBytes, batchSize)
	result.MemoryLatency = memResult.TransferLatency

	// Compute latency
	if cfg.NumPipelineStages > 1 {
		// Pipelined across tiles
		// Latency = fill time + drain time + (tiles-1) * cycle time
		fillTime := float64(cfg.NumPipelineStages) * mvmResult.TotalLatency
		drainTime := float64(cfg.NumPipelineStages) * mvmResult.TotalLatency
		steadyTime := float64(totalTiles*batchSize-cfg.NumPipelineStages) * mvmResult.TotalLatency / float64(cfg.NumPipelineStages)
		result.ComputeLatency = fillTime + steadyTime + drainTime
	} else {
		// Sequential
		result.ComputeLatency = float64(totalTiles*batchSize) * mvmResult.TotalLatency
	}

	// Activation latency (overlapped with compute)
	result.ActivationLatency = float64(batchSize*outputCols) * cfg.ActivationLatencyNs / float64(cfg.ArrayCols)

	// Total with memory overlap
	// Memory and compute can overlap in weight-stationary
	result.TotalLatency = math.Max(result.ComputeLatency, result.MemoryLatency) + result.ActivationLatency

	// Operations count
	result.TotalOps = int64(inputRows) * int64(inputCols) * int64(outputCols) * int64(batchSize)
	result.Throughput = float64(result.TotalOps) / (result.TotalLatency * 1e-9) // OPS

	return result
}

// LayerLatencyResult contains layer latency analysis
type LayerLatencyResult struct {
	MemoryLatency     float64 // ns
	ComputeLatency    float64 // ns
	ActivationLatency float64 // ns
	TotalLatency      float64 // ns
	TotalOps          int64
	Throughput        float64 // Operations per second
}

// NetworkLatency calculates latency for complete network inference
func (lm *LatencyModel) NetworkLatency(layerDims [][]int, batchSize int) *NetworkLatencyResult {
	result := &NetworkLatencyResult{
		LayerLatencies: make([]float64, len(layerDims)),
		LayerOps:       make([]int64, len(layerDims)),
	}

	totalLatency := 0.0
	totalOps := int64(0)

	for i, dims := range layerDims {
		inputDim := dims[0]
		outputDim := dims[1]

		// Assume square input for simplicity
		inputRows := int(math.Sqrt(float64(inputDim)))
		if inputRows == 0 {
			inputRows = 1
		}
		inputCols := inputDim / inputRows

		layerResult := lm.LayerLatency(inputRows, inputCols, outputDim, batchSize)

		result.LayerLatencies[i] = layerResult.TotalLatency
		result.LayerOps[i] = layerResult.TotalOps

		// Pipeline layers if possible
		if i > 0 && lm.Config.NumPipelineStages > 1 {
			// Overlap with previous layer
			overlap := math.Min(result.LayerLatencies[i-1]*0.5, layerResult.TotalLatency*0.5)
			totalLatency -= overlap
		}
		totalLatency += layerResult.TotalLatency
		totalOps += layerResult.TotalOps
	}

	result.TotalLatency = totalLatency
	result.TotalOps = totalOps
	result.Throughput = float64(totalOps) / (totalLatency * 1e-9)
	result.FPS = float64(batchSize) / (totalLatency * 1e-9)
	result.TOPS = result.Throughput / 1e12

	return result
}

// NetworkLatencyResult contains network-level latency analysis
type NetworkLatencyResult struct {
	LayerLatencies []float64
	LayerOps       []int64
	TotalLatency   float64 // ns
	TotalOps       int64
	Throughput     float64 // OPS
	FPS            float64 // Frames per second
	TOPS           float64 // Tera operations per second
}

// ================== Dataflow Optimization ==================

// DataflowConfig configures dataflow optimization
type DataflowConfig struct {
	TileSize        int     // Tile size for tiling
	UnrollFactor    int     // Loop unroll factor
	ParallelTiles   int     // Number of parallel tiles
	PrefetchDepth   int     // Prefetch buffer depth
	KernelDuplication int   // Kernel duplication factor (ConvDK)
	ShiftParameter  int     // Shift parameter for input reuse
}

// DefaultDataflowConfig returns typical dataflow parameters
func DefaultDataflowConfig() *DataflowConfig {
	return &DataflowConfig{
		TileSize:          64,
		UnrollFactor:      4,
		ParallelTiles:     4,
		PrefetchDepth:     2,
		KernelDuplication: 2,
		ShiftParameter:    4,
	}
}

// DataflowOptimizer optimizes memory access patterns
type DataflowOptimizer struct {
	Config   *DataflowConfig
	Memory   *MemoryHierarchy

	// Traffic statistics
	BaselineTraffic   int64
	OptimizedTraffic  int64
	TrafficReduction  float64
}

// NewDataflowOptimizer creates a dataflow optimizer
func NewDataflowOptimizer(config *DataflowConfig, memConfig *MemoryConfig) *DataflowOptimizer {
	return &DataflowOptimizer{
		Config: config,
		Memory: NewMemoryHierarchy(memConfig),
	}
}

// OptimizeConvolution optimizes convolution dataflow
func (dfo *DataflowOptimizer) OptimizeConvolution(
	inputH, inputW, inputC int,
	kernelH, kernelW int,
	outputC int,
	stride int,
) *DataflowResult {
	cfg := dfo.Config
	result := &DataflowResult{}

	// Output dimensions
	outputH := (inputH - kernelH) / stride + 1
	outputW := (inputW - kernelW) / stride + 1

	// Baseline traffic (naive implementation)
	// Each output needs all kernel weights and corresponding input patch
	inputPatchSize := kernelH * kernelW * inputC
	weightSize := kernelH * kernelW * inputC * outputC
	outputSize := outputH * outputW * outputC

	// Baseline: reload weights for each output position
	baselineInputTraffic := int64(outputH * outputW * inputPatchSize)
	baselineWeightTraffic := int64(weightSize) // Assume weights fit in buffer
	baselineOutputTraffic := int64(outputSize)
	result.BaselineTraffic = baselineInputTraffic + baselineWeightTraffic + baselineOutputTraffic

	// Optimized with ConvDK (kernel duplication)
	// Duplicate kernels to enable input reuse
	tilesH := (outputH + cfg.TileSize - 1) / cfg.TileSize
	tilesW := (outputW + cfg.TileSize - 1) / cfg.TileSize

	// Input traffic with shift-based reuse
	// Load input tile once, shift to generate multiple outputs
	inputTileSize := (cfg.TileSize + kernelH - 1) * (cfg.TileSize + kernelW - 1) * inputC
	reuseFactor := cfg.ShiftParameter
	optimizedInputTraffic := int64(tilesH * tilesW * inputTileSize / reuseFactor)

	// Weight traffic with duplication
	// Weights are duplicated in tile memory, loaded once per tile group
	duplicatedWeightSize := weightSize * cfg.KernelDuplication
	optimizedWeightTraffic := int64(duplicatedWeightSize / cfg.ParallelTiles)

	// Output traffic (same as baseline)
	optimizedOutputTraffic := int64(outputSize)

	result.OptimizedTraffic = optimizedInputTraffic + optimizedWeightTraffic + optimizedOutputTraffic

	// Calculate reduction
	result.TrafficReduction = 1.0 - float64(result.OptimizedTraffic)/float64(result.BaselineTraffic)

	// Tile memory utilization
	tileMemoryUsed := cfg.TileSize * cfg.TileSize * inputC / 8 // bytes
	result.TileUtilization = float64(tileMemoryUsed) / float64(dfo.Memory.Config.TileMemoryBytes)
	if result.TileUtilization > 1.0 {
		result.TileUtilization = 1.0 // Cap at 100%
	}

	// Energy savings (proportional to traffic reduction)
	result.EnergySavings = result.TrafficReduction * 0.8 // 80% of traffic reduction

	// Latency improvement
	result.LatencyReduction = result.TrafficReduction * 0.6 // Memory-bound improvement

	// Update statistics
	dfo.BaselineTraffic += result.BaselineTraffic
	dfo.OptimizedTraffic += result.OptimizedTraffic
	dfo.TrafficReduction = 1.0 - float64(dfo.OptimizedTraffic)/float64(dfo.BaselineTraffic)

	return result
}

// DataflowResult contains dataflow optimization results
type DataflowResult struct {
	BaselineTraffic   int64
	OptimizedTraffic  int64
	TrafficReduction  float64 // 0-1
	TileUtilization   float64 // 0-1
	EnergySavings     float64 // 0-1
	LatencyReduction  float64 // 0-1
}

// ================== Throughput Calculator ==================

// ThroughputConfig configures throughput calculation
type ThroughputConfig struct {
	ArraySize       int
	NumArrays       int
	FrequencyMHz    float64
	Utilization     float64 // 0-1
	BitPrecision    int
}

// DefaultThroughputConfig returns typical throughput parameters
func DefaultThroughputConfig() *ThroughputConfig {
	return &ThroughputConfig{
		ArraySize:    64,
		NumArrays:    16,
		FrequencyMHz: 100,
		Utilization:  0.85,
		BitPrecision: 8,
	}
}

// ThroughputCalculator calculates system throughput
type ThroughputCalculator struct {
	Config *ThroughputConfig
}

// NewThroughputCalculator creates a throughput calculator
func NewThroughputCalculator(config *ThroughputConfig) *ThroughputCalculator {
	return &ThroughputCalculator{Config: config}
}

// CalculatePeakTOPS calculates peak throughput
func (tc *ThroughputCalculator) CalculatePeakTOPS() float64 {
	cfg := tc.Config

	// MACs per cycle per array
	macsPerCycle := cfg.ArraySize * cfg.ArraySize

	// Total MACs per cycle (all arrays)
	totalMACs := macsPerCycle * cfg.NumArrays

	// Peak operations per second
	peakOPS := float64(totalMACs) * cfg.FrequencyMHz * 1e6

	// Convert to TOPS
	return peakOPS / 1e12
}

// CalculateEffectiveTOPS calculates effective throughput with utilization
func (tc *ThroughputCalculator) CalculateEffectiveTOPS() float64 {
	return tc.CalculatePeakTOPS() * tc.Config.Utilization
}

// CalculateFPS calculates frames per second for a model
func (tc *ThroughputCalculator) CalculateFPS(totalMACs int64, batchSize int) float64 {
	effectiveTOPS := tc.CalculateEffectiveTOPS()
	effectiveOPS := effectiveTOPS * 1e12

	// Time for one batch
	batchTime := float64(totalMACs) / effectiveOPS // seconds

	return float64(batchSize) / batchTime
}

// CalculateLatencyMs calculates inference latency
func (tc *ThroughputCalculator) CalculateLatencyMs(totalMACs int64) float64 {
	effectiveTOPS := tc.CalculateEffectiveTOPS()
	effectiveOPS := effectiveTOPS * 1e12

	// Latency in ms
	return float64(totalMACs) / effectiveOPS * 1000
}

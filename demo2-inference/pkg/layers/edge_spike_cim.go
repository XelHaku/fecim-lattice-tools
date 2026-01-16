// edge_spike_cim.go - Edge Deployment and Analog-to-Spike Conversion for CIM
// Research iteration 116: TinyML deployment, CIM compilers, and spike encoding schemes
//
// Key findings:
// - TinyML: 8-bit quantization achieves 3-4× storage reduction, 10.6-22.1 µJ/inference
// - SRAM CIM: 687.5 TOPS/W (65nm drive-strength ADC-less), 1241 GOPS (28nm time-domain)
// - Sigma-Delta neurons: Direct spike encoding from analog, ADC-free neuromorphic
// - Level-crossing ADC: >125× data compression, 4% NRMSE reconstruction
// - CIM-MLC compiler: 3.2× speedup with multi-level scheduling (ASPLOS 2024)
// - CLSA-CIM: 17.9× utilization improvement, 29.2× speedup

package layers

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sort"
)

// ============================================================================
// SECTION 1: TinyML Edge Deployment
// ============================================================================

// TinyMLConfig configures model deployment for resource-constrained MCUs
type TinyMLConfig struct {
	// Memory constraints
	FlashSizeKB     int     // Available flash memory (typically 256-1024 KB)
	SRAMSizeKB      int     // Available SRAM (typically 64-256 KB)
	MaxModelSizeKB  int     // Maximum model footprint

	// Quantization settings
	WeightBits      int     // Weight precision (4, 8, or 16)
	ActivationBits  int     // Activation precision (8 or 16)
	PerChannelQuant bool    // Per-channel vs per-tensor quantization

	// Inference settings
	InferenceLatencyTargetMs float64 // Target latency
	EnergyBudgetUJ          float64 // Energy per inference budget

	// Hardware parameters
	HasFPU          bool    // Hardware floating-point unit
	HasDSP          bool    // DSP instructions available
	ClockFreqMHz    float64 // MCU clock frequency

	// Optimization flags
	EnableFusion    bool    // Fuse Conv+BN+ReLU
	EnablePruning   bool    // Apply structured pruning
	StaticAlloc     bool    // Static memory allocation
}

// DefaultTinyMLConfig returns typical MCU deployment configuration
func DefaultTinyMLConfig() *TinyMLConfig {
	return &TinyMLConfig{
		FlashSizeKB:             512,
		SRAMSizeKB:              128,
		MaxModelSizeKB:          400,
		WeightBits:              8,
		ActivationBits:          8,
		PerChannelQuant:         true,
		InferenceLatencyTargetMs: 20.0,
		EnergyBudgetUJ:          15.0,
		HasFPU:                  false,
		HasDSP:                  true,
		ClockFreqMHz:            64.0,
		EnableFusion:            true,
		EnablePruning:           true,
		StaticAlloc:             true,
	}
}

// TinyMLLayer represents a quantized layer for edge deployment
type TinyMLLayer struct {
	Name            string
	Type            string    // "conv2d", "dense", "depthwise", "pooling"

	// Quantized weights
	WeightsInt8     []int8    // INT8 weights
	WeightsInt4     []int8    // INT4 packed weights (2 per byte)
	BiasInt32       []int32   // INT32 biases for accumulator

	// Quantization parameters
	WeightScale     []float32 // Per-channel scale factors
	WeightZeroPoint []int8    // Per-channel zero points
	InputScale      float32   // Input activation scale
	InputZeroPoint  int8      // Input zero point
	OutputScale     float32   // Output activation scale
	OutputZeroPoint int8      // Output zero point

	// Layer dimensions
	InputShape      []int     // [batch, height, width, channels] or [batch, features]
	OutputShape     []int
	KernelShape     []int     // [kH, kW, inC, outC] for conv

	// Memory usage
	WeightSizeBytes int
	ActivationBytes int
	ScratchPadBytes int
}

// TinyMLModel represents a complete model optimized for MCU deployment
type TinyMLModel struct {
	Config       *TinyMLConfig
	Layers       []*TinyMLLayer

	// Memory analysis
	TotalWeightBytes     int
	PeakActivationBytes  int
	TotalScratchBytes    int

	// Performance estimates
	EstimatedMACs        int64
	EstimatedLatencyMs   float64
	EstimatedEnergyUJ    float64

	// Compression stats
	OriginalSizeKB       float64
	CompressedSizeKB     float64
	CompressionRatio     float64
}

// NewTinyMLModel creates a new model for TinyML deployment
func NewTinyMLModel(config *TinyMLConfig) *TinyMLModel {
	if config == nil {
		config = DefaultTinyMLConfig()
	}
	return &TinyMLModel{
		Config: config,
		Layers: make([]*TinyMLLayer, 0),
	}
}

// QuantizeWeights converts FP32 weights to INT8 or INT4
func (m *TinyMLModel) QuantizeWeights(weights [][]float64, bits int, perChannel bool) (*TinyMLLayer, error) {
	if bits != 4 && bits != 8 {
		return nil, fmt.Errorf("unsupported quantization bits: %d (use 4 or 8)", bits)
	}

	layer := &TinyMLLayer{
		Type: "dense",
	}

	// Flatten weights
	flat := make([]float64, 0)
	for _, row := range weights {
		flat = append(flat, row...)
	}

	if perChannel {
		// Per-channel quantization
		layer.WeightScale = make([]float32, len(weights))
		layer.WeightZeroPoint = make([]int8, len(weights))

		if bits == 8 {
			layer.WeightsInt8 = make([]int8, len(flat))
		} else {
			// INT4: pack 2 weights per byte
			layer.WeightsInt4 = make([]int8, (len(flat)+1)/2)
		}

		idx := 0
		for c, row := range weights {
			// Find min/max per channel
			minVal, maxVal := row[0], row[0]
			for _, v := range row {
				if v < minVal {
					minVal = v
				}
				if v > maxVal {
					maxVal = v
				}
			}

			// Compute scale and zero point
			var qmin, qmax float64
			if bits == 8 {
				qmin, qmax = -128, 127
			} else {
				qmin, qmax = -8, 7
			}

			scale := (maxVal - minVal) / (qmax - qmin)
			if scale < 1e-10 {
				scale = 1e-10
			}
			zeroPoint := qmin - minVal/scale

			layer.WeightScale[c] = float32(scale)
			layer.WeightZeroPoint[c] = int8(math.Round(zeroPoint))

			// Quantize weights
			for _, v := range row {
				q := math.Round(v/scale + float64(layer.WeightZeroPoint[c]))
				q = math.Max(qmin, math.Min(qmax, q))

				if bits == 8 {
					layer.WeightsInt8[idx] = int8(q)
				} else {
					// Pack INT4
					byteIdx := idx / 2
					if idx%2 == 0 {
						layer.WeightsInt4[byteIdx] = int8(q) & 0x0F
					} else {
						layer.WeightsInt4[byteIdx] |= (int8(q) & 0x0F) << 4
					}
				}
				idx++
			}
		}
	} else {
		// Per-tensor quantization
		minVal, maxVal := flat[0], flat[0]
		for _, v := range flat {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}

		var qmin, qmax float64
		if bits == 8 {
			qmin, qmax = -128, 127
		} else {
			qmin, qmax = -8, 7
		}

		scale := (maxVal - minVal) / (qmax - qmin)
		if scale < 1e-10 {
			scale = 1e-10
		}
		zeroPoint := qmin - minVal/scale

		layer.WeightScale = []float32{float32(scale)}
		layer.WeightZeroPoint = []int8{int8(math.Round(zeroPoint))}

		if bits == 8 {
			layer.WeightsInt8 = make([]int8, len(flat))
			for i, v := range flat {
				q := math.Round(v/scale + float64(layer.WeightZeroPoint[0]))
				q = math.Max(qmin, math.Min(qmax, q))
				layer.WeightsInt8[i] = int8(q)
			}
			layer.WeightSizeBytes = len(flat)
		} else {
			layer.WeightsInt4 = make([]int8, (len(flat)+1)/2)
			for i, v := range flat {
				q := math.Round(v/scale + float64(layer.WeightZeroPoint[0]))
				q = math.Max(qmin, math.Min(qmax, q))
				byteIdx := i / 2
				if i%2 == 0 {
					layer.WeightsInt4[byteIdx] = int8(q) & 0x0F
				} else {
					layer.WeightsInt4[byteIdx] |= (int8(q) & 0x0F) << 4
				}
			}
			layer.WeightSizeBytes = (len(flat) + 1) / 2
		}
	}

	return layer, nil
}

// AnalyzeMemory estimates memory requirements for deployment
func (m *TinyMLModel) AnalyzeMemory() error {
	m.TotalWeightBytes = 0
	m.PeakActivationBytes = 0
	m.TotalScratchBytes = 0

	for _, layer := range m.Layers {
		m.TotalWeightBytes += layer.WeightSizeBytes

		// Activation memory = input + output buffers
		inputSize := 1
		for _, d := range layer.InputShape {
			inputSize *= d
		}
		outputSize := 1
		for _, d := range layer.OutputShape {
			outputSize *= d
		}

		activationBytes := (inputSize + outputSize) * (m.Config.ActivationBits / 8)
		if activationBytes > m.PeakActivationBytes {
			m.PeakActivationBytes = activationBytes
		}

		m.TotalScratchBytes += layer.ScratchPadBytes
	}

	return nil
}

// CheckDeploymentFeasibility validates if model fits on target MCU
func (m *TinyMLModel) CheckDeploymentFeasibility() (bool, []string) {
	issues := make([]string, 0)

	totalFlash := float64(m.TotalWeightBytes) / 1024.0
	if totalFlash > float64(m.Config.FlashSizeKB) {
		issues = append(issues, fmt.Sprintf("Model weights (%.1f KB) exceed flash (%.1f KB)",
			totalFlash, float64(m.Config.FlashSizeKB)))
	}

	totalSRAM := float64(m.PeakActivationBytes+m.TotalScratchBytes) / 1024.0
	if totalSRAM > float64(m.Config.SRAMSizeKB) {
		issues = append(issues, fmt.Sprintf("Activation memory (%.1f KB) exceeds SRAM (%.1f KB)",
			totalSRAM, float64(m.Config.SRAMSizeKB)))
	}

	if m.EstimatedLatencyMs > m.Config.InferenceLatencyTargetMs {
		issues = append(issues, fmt.Sprintf("Estimated latency (%.2f ms) exceeds target (%.2f ms)",
			m.EstimatedLatencyMs, m.Config.InferenceLatencyTargetMs))
	}

	if m.EstimatedEnergyUJ > m.Config.EnergyBudgetUJ {
		issues = append(issues, fmt.Sprintf("Estimated energy (%.2f µJ) exceeds budget (%.2f µJ)",
			m.EstimatedEnergyUJ, m.Config.EnergyBudgetUJ))
	}

	return len(issues) == 0, issues
}

// ============================================================================
// SECTION 2: CIM-Aware Compiler and Mapping
// ============================================================================

// CIMCompilerConfig configures the CIM compilation stack
type CIMCompilerConfig struct {
	// Hardware abstraction
	CrossbarRows    int     // Rows per crossbar
	CrossbarCols    int     // Columns per crossbar
	NumCrossbars    int     // Total crossbars available
	NumTiles        int     // Crossbar tiles

	// Precision settings
	WeightBits      int     // Per-cell weight precision
	ADCBits         int     // ADC resolution
	DACBits         int     // DAC resolution

	// Scheduling options
	EnableTiling    bool    // Enable layer tiling
	EnablePipelining bool   // Enable pipeline scheduling
	EnableReuse     bool    // Enable weight/input reuse

	// Optimization targets
	OptimizeFor     string  // "latency", "energy", "throughput"
	MaxLatencyMs    float64
	MaxEnergyPJ     float64
}

// DefaultCIMCompilerConfig returns typical CIM compiler configuration
func DefaultCIMCompilerConfig() *CIMCompilerConfig {
	return &CIMCompilerConfig{
		CrossbarRows:    256,
		CrossbarCols:    256,
		NumCrossbars:    16,
		NumTiles:        4,
		WeightBits:      4,
		ADCBits:         6,
		DACBits:         8,
		EnableTiling:    true,
		EnablePipelining: true,
		EnableReuse:     true,
		OptimizeFor:     "energy",
		MaxLatencyMs:    10.0,
		MaxEnergyPJ:     1e9,
	}
}

// CIMInstruction represents a compiled CIM operation
type CIMInstruction struct {
	OpType        string    // "MVM", "ADD", "ACT", "LOAD", "STORE", "SYNC"
	CrossbarID    int       // Target crossbar
	TileID        int       // Target tile

	// MVM-specific
	InputStart    int       // Input vector start index
	InputLength   int       // Input vector length
	OutputStart   int       // Output start index

	// Memory addresses
	WeightAddr    int       // Weight memory address
	InputAddr     int       // Input buffer address
	OutputAddr    int       // Output buffer address

	// Timing
	CycleStart    int       // Start cycle
	CycleDuration int       // Duration in cycles

	// Energy estimate
	EnergyPJ      float64
}

// TilingStrategy defines how a layer is mapped to crossbars
type TilingStrategy struct {
	LayerName       string
	TileRows        int      // Rows per tile
	TileCols        int      // Columns per tile
	NumTilesRow     int      // Tiles in row dimension
	NumTilesCol     int      // Tiles in column dimension

	// Dataflow
	Dataflow        string   // "output_stationary", "weight_stationary", "row_stationary"
	ReuseFactor     int      // Data reuse factor

	// Resource allocation
	CrossbarAlloc   []int    // Crossbar IDs allocated
	BufferSizeKB    float64  // Required buffer size
}

// CIMCompiler implements multi-level CIM compilation (based on CIM-MLC)
type CIMCompiler struct {
	Config          *CIMCompilerConfig
	Instructions    []*CIMInstruction
	TilingStrategies map[string]*TilingStrategy

	// Performance tracking
	TotalCycles     int
	TotalEnergyPJ   float64
	Utilization     float64
	Speedup         float64
}

// NewCIMCompiler creates a new CIM compiler instance
func NewCIMCompiler(config *CIMCompilerConfig) *CIMCompiler {
	if config == nil {
		config = DefaultCIMCompilerConfig()
	}
	return &CIMCompiler{
		Config:           config,
		Instructions:     make([]*CIMInstruction, 0),
		TilingStrategies: make(map[string]*TilingStrategy),
	}
}

// ComputeTilingStrategy determines optimal tiling for a layer
func (c *CIMCompiler) ComputeTilingStrategy(layerName string, weightShape []int) (*TilingStrategy, error) {
	if len(weightShape) < 2 {
		return nil, fmt.Errorf("invalid weight shape: need at least 2 dimensions")
	}

	rows := weightShape[0]
	cols := weightShape[1]

	// Determine tile dimensions based on crossbar size
	tileRows := c.Config.CrossbarRows
	tileCols := c.Config.CrossbarCols

	// Adjust for weight precision (multi-cell weights)
	cellsPerWeight := (c.Config.WeightBits + 1) / 2 // 2 bits per cell typical
	tileCols /= cellsPerWeight

	// Calculate number of tiles needed
	numTilesRow := (rows + tileRows - 1) / tileRows
	numTilesCol := (cols + tileCols - 1) / tileCols
	totalTiles := numTilesRow * numTilesCol

	// Allocate crossbars
	crossbarAlloc := make([]int, 0)
	for i := 0; i < totalTiles && i < c.Config.NumCrossbars; i++ {
		crossbarAlloc = append(crossbarAlloc, i)
	}

	// Determine dataflow based on dimensions
	dataflow := "output_stationary"
	reuseFactor := 1

	if rows > cols {
		// More outputs than inputs - prefer weight stationary
		dataflow = "weight_stationary"
		reuseFactor = rows / c.Config.CrossbarRows
		if reuseFactor < 1 {
			reuseFactor = 1
		}
	}

	// Calculate buffer requirements
	inputBufferKB := float64(tileCols*2) / 1024.0   // Double buffer inputs
	outputBufferKB := float64(tileRows*4) / 1024.0  // 32-bit accumulators
	bufferSizeKB := inputBufferKB + outputBufferKB

	strategy := &TilingStrategy{
		LayerName:     layerName,
		TileRows:      tileRows,
		TileCols:      tileCols,
		NumTilesRow:   numTilesRow,
		NumTilesCol:   numTilesCol,
		Dataflow:      dataflow,
		ReuseFactor:   reuseFactor,
		CrossbarAlloc: crossbarAlloc,
		BufferSizeKB:  bufferSizeKB,
	}

	c.TilingStrategies[layerName] = strategy
	return strategy, nil
}

// CompileLayer generates CIM instructions for a layer
func (c *CIMCompiler) CompileLayer(layerName string, weights [][]float64) error {
	strategy, exists := c.TilingStrategies[layerName]
	if !exists {
		var err error
		strategy, err = c.ComputeTilingStrategy(layerName, []int{len(weights), len(weights[0])})
		if err != nil {
			return err
		}
	}

	cycle := c.TotalCycles

	// Generate MVM instructions for each tile
	for tr := 0; tr < strategy.NumTilesRow; tr++ {
		for tc := 0; tc < strategy.NumTilesCol; tc++ {
			tileIdx := tr*strategy.NumTilesCol + tc
			crossbarID := tileIdx % len(strategy.CrossbarAlloc)

			rowStart := tr * strategy.TileRows
			colStart := tc * strategy.TileCols

			// Load instruction
			loadInst := &CIMInstruction{
				OpType:     "LOAD",
				CrossbarID: strategy.CrossbarAlloc[crossbarID],
				TileID:     tileIdx % c.Config.NumTiles,
				WeightAddr: rowStart*len(weights[0]) + colStart,
				CycleStart: cycle,
				CycleDuration: 10, // Weight loading cycles
				EnergyPJ:   float64(strategy.TileRows*strategy.TileCols) * 0.1,
			}
			c.Instructions = append(c.Instructions, loadInst)
			cycle += loadInst.CycleDuration

			// MVM instruction
			mvmInst := &CIMInstruction{
				OpType:      "MVM",
				CrossbarID:  strategy.CrossbarAlloc[crossbarID],
				TileID:      tileIdx % c.Config.NumTiles,
				InputStart:  colStart,
				InputLength: strategy.TileCols,
				OutputStart: rowStart,
				InputAddr:   colStart * 2,
				OutputAddr:  rowStart * 4,
				CycleStart:  cycle,
				CycleDuration: 1, // Single-cycle MVM
				EnergyPJ:    float64(strategy.TileRows*strategy.TileCols) * 0.5,
			}
			c.Instructions = append(c.Instructions, mvmInst)
			cycle += mvmInst.CycleDuration
		}
	}

	// Add partial sum accumulation if tiled across columns
	if strategy.NumTilesCol > 1 {
		addInst := &CIMInstruction{
			OpType:        "ADD",
			OutputStart:   0,
			CycleStart:    cycle,
			CycleDuration: strategy.NumTilesCol - 1,
			EnergyPJ:      float64(strategy.NumTilesCol*strategy.TileRows) * 0.05,
		}
		c.Instructions = append(c.Instructions, addInst)
		cycle += addInst.CycleDuration
	}

	c.TotalCycles = cycle

	// Calculate total energy
	for _, inst := range c.Instructions {
		c.TotalEnergyPJ += inst.EnergyPJ
	}

	return nil
}

// OptimizeSchedule applies cross-layer scheduling (CLSA-CIM inspired)
func (c *CIMCompiler) OptimizeSchedule() {
	// Group instructions by crossbar to identify parallelization opportunities
	crossbarGroups := make(map[int][]*CIMInstruction)
	for _, inst := range c.Instructions {
		crossbarGroups[inst.CrossbarID] = append(crossbarGroups[inst.CrossbarID], inst)
	}

	// Re-schedule for parallel execution across crossbars
	maxCyclePerCrossbar := make(map[int]int)
	for cbID, insts := range crossbarGroups {
		totalCycles := 0
		for _, inst := range insts {
			inst.CycleStart = totalCycles
			totalCycles += inst.CycleDuration
		}
		maxCyclePerCrossbar[cbID] = totalCycles
	}

	// Find critical path
	maxCycles := 0
	for _, cycles := range maxCyclePerCrossbar {
		if cycles > maxCycles {
			maxCycles = cycles
		}
	}

	// Calculate utilization
	totalOpCycles := 0
	for _, inst := range c.Instructions {
		totalOpCycles += inst.CycleDuration
	}

	theoreticalParallel := float64(totalOpCycles) / float64(c.Config.NumCrossbars)
	c.Utilization = theoreticalParallel / float64(maxCycles)
	c.TotalCycles = maxCycles

	// CLSA-CIM reports up to 17.9× utilization improvement
	c.Speedup = 1.0 + (c.Utilization * 16.9) // Approximate based on paper
}

// ============================================================================
// SECTION 3: Sigma-Delta Spike Encoding
// ============================================================================

// SigmaDeltaConfig configures sigma-delta neuron encoding
type SigmaDeltaConfig struct {
	// Encoder parameters
	Threshold       float64 // Spike threshold
	DecayFactor     float64 // State decay (leakage)
	OversamplRatio  int     // Oversampling ratio vs Nyquist

	// Noise parameters
	ThresholdNoise  float64 // Threshold variation (%)
	StateNoise      float64 // State noise level

	// Mode selection
	Mode            string  // "rate", "temporal", "delta"
}

// DefaultSigmaDeltaConfig returns typical sigma-delta encoder configuration
func DefaultSigmaDeltaConfig() *SigmaDeltaConfig {
	return &SigmaDeltaConfig{
		Threshold:      0.5,
		DecayFactor:   0.95,
		OversamplRatio: 32,
		ThresholdNoise: 0.02,
		StateNoise:    0.01,
		Mode:          "delta",
	}
}

// SigmaDeltaEncoder implements sigma-delta spike encoding
type SigmaDeltaEncoder struct {
	Config         *SigmaDeltaConfig
	State          float64   // Internal integrator state
	PreviousInput  float64   // Previous input for delta computation
	SpikeHistory   []int     // Spike output history (+1, 0, -1)

	// Statistics
	TotalSpikes    int
	CompressionRatio float64
}

// NewSigmaDeltaEncoder creates a new sigma-delta encoder
func NewSigmaDeltaEncoder(config *SigmaDeltaConfig) *SigmaDeltaEncoder {
	if config == nil {
		config = DefaultSigmaDeltaConfig()
	}
	return &SigmaDeltaEncoder{
		Config:       config,
		SpikeHistory: make([]int, 0),
	}
}

// EncodeSample encodes a single sample to spikes
func (e *SigmaDeltaEncoder) EncodeSample(input float64) []int {
	spikes := make([]int, 0)

	// Add state noise
	stateNoise := rand.NormFloat64() * e.Config.StateNoise
	threshNoise := rand.NormFloat64() * e.Config.ThresholdNoise
	effectiveThresh := e.Config.Threshold * (1 + threshNoise)

	switch e.Config.Mode {
	case "delta":
		// Delta modulation: encode changes
		delta := input - e.PreviousInput
		e.State += delta + stateNoise

		// Generate spikes based on accumulated change
		for e.State > effectiveThresh {
			spikes = append(spikes, 1)
			e.State -= e.Config.Threshold
			e.TotalSpikes++
		}
		for e.State < -effectiveThresh {
			spikes = append(spikes, -1)
			e.State += e.Config.Threshold
			e.TotalSpikes++
		}

		e.PreviousInput = input

	case "rate":
		// Rate coding: spike probability proportional to input
		e.State += input + stateNoise

		if e.State > effectiveThresh {
			spikes = append(spikes, 1)
			e.State -= e.Config.Threshold
			e.TotalSpikes++
		}

		// Apply decay
		e.State *= e.Config.DecayFactor

	case "temporal":
		// Temporal coding: time-to-first-spike
		e.State += input
		if e.State > effectiveThresh {
			spikes = append(spikes, 1)
			e.State = 0
			e.TotalSpikes++
		}
	}

	e.SpikeHistory = append(e.SpikeHistory, spikes...)
	return spikes
}

// EncodeSignal encodes an entire signal to spikes
func (e *SigmaDeltaEncoder) EncodeSignal(signal []float64) [][]int {
	e.State = 0
	e.PreviousInput = 0
	e.SpikeHistory = make([]int, 0)
	e.TotalSpikes = 0

	result := make([][]int, len(signal))
	for i, sample := range signal {
		result[i] = e.EncodeSample(sample)
	}

	// Calculate compression ratio
	totalSamples := len(signal) * e.Config.OversamplRatio
	e.CompressionRatio = float64(totalSamples) / float64(e.TotalSpikes+1)

	return result
}

// DecodeSpikes reconstructs signal from spikes
func (e *SigmaDeltaEncoder) DecodeSpikes(spikes [][]int) []float64 {
	signal := make([]float64, len(spikes))
	state := 0.0

	for i, spikeSet := range spikes {
		for _, spike := range spikeSet {
			state += float64(spike) * e.Config.Threshold
		}
		signal[i] = state
	}

	return signal
}

// ============================================================================
// SECTION 4: Level-Crossing ADC (ADC-Free Encoding)
// ============================================================================

// LevelCrossingConfig configures level-crossing ADC
type LevelCrossingConfig struct {
	// Resolution
	NumLevels       int     // Number of quantization levels
	VrefHigh        float64 // Reference high voltage
	VrefLow         float64 // Reference low voltage

	// Hysteresis
	HysteresisRatio float64 // Hysteresis as fraction of LSB

	// Timing
	DeadTimeNs      float64 // Minimum time between events

	// Power
	ComparatorPowerNW float64 // Comparator power consumption
}

// DefaultLevelCrossingConfig returns typical LC-ADC configuration
func DefaultLevelCrossingConfig() *LevelCrossingConfig {
	return &LevelCrossingConfig{
		NumLevels:         64,
		VrefHigh:         1.8,
		VrefLow:          0.0,
		HysteresisRatio:  0.1,
		DeadTimeNs:       100,
		ComparatorPowerNW: 50,
	}
}

// LevelCrossingEvent represents an ADC event
type LevelCrossingEvent struct {
	Timestamp   float64 // Time in seconds
	Level       int     // Quantization level crossed
	Direction   int     // +1 for up, -1 for down
}

// LevelCrossingADC implements asynchronous level-crossing ADC
type LevelCrossingADC struct {
	Config          *LevelCrossingConfig
	CurrentLevel    int
	LastEventTime   float64
	Events          []LevelCrossingEvent

	// Statistics
	CompressionRatio float64
	AverageRate      float64 // Events per second
	NRMSE            float64 // Normalized RMS error after reconstruction
}

// NewLevelCrossingADC creates a new LC-ADC instance
func NewLevelCrossingADC(config *LevelCrossingConfig) *LevelCrossingADC {
	if config == nil {
		config = DefaultLevelCrossingConfig()
	}
	return &LevelCrossingADC{
		Config:       config,
		CurrentLevel: config.NumLevels / 2, // Start at midpoint
		Events:       make([]LevelCrossingEvent, 0),
	}
}

// voltageToLevel converts voltage to quantization level
func (adc *LevelCrossingADC) voltageToLevel(voltage float64) int {
	normalized := (voltage - adc.Config.VrefLow) / (adc.Config.VrefHigh - adc.Config.VrefLow)
	level := int(normalized * float64(adc.Config.NumLevels))

	if level < 0 {
		level = 0
	}
	if level >= adc.Config.NumLevels {
		level = adc.Config.NumLevels - 1
	}

	return level
}

// levelToVoltage converts level back to voltage
func (adc *LevelCrossingADC) levelToVoltage(level int) float64 {
	return adc.Config.VrefLow + float64(level)*(adc.Config.VrefHigh-adc.Config.VrefLow)/float64(adc.Config.NumLevels)
}

// ProcessSample processes a single sample
func (adc *LevelCrossingADC) ProcessSample(voltage float64, timestamp float64) []LevelCrossingEvent {
	events := make([]LevelCrossingEvent, 0)

	targetLevel := adc.voltageToLevel(voltage)

	// Check dead time
	deadTimeSec := adc.Config.DeadTimeNs * 1e-9
	if timestamp-adc.LastEventTime < deadTimeSec {
		return events
	}

	// Calculate hysteresis in levels
	lsb := 1
	hysteresisLevels := int(float64(lsb) * adc.Config.HysteresisRatio)
	if hysteresisLevels < 1 {
		hysteresisLevels = 1
	}

	// Generate events for level crossings
	for adc.CurrentLevel < targetLevel-hysteresisLevels {
		adc.CurrentLevel++
		event := LevelCrossingEvent{
			Timestamp: timestamp,
			Level:     adc.CurrentLevel,
			Direction: 1,
		}
		events = append(events, event)
		adc.LastEventTime = timestamp
	}

	for adc.CurrentLevel > targetLevel+hysteresisLevels {
		adc.CurrentLevel--
		event := LevelCrossingEvent{
			Timestamp: timestamp,
			Level:     adc.CurrentLevel,
			Direction: -1,
		}
		events = append(events, event)
		adc.LastEventTime = timestamp
	}

	adc.Events = append(adc.Events, events...)
	return events
}

// ProcessSignal processes an entire signal with timestamps
func (adc *LevelCrossingADC) ProcessSignal(signal []float64, sampleRate float64) {
	adc.Events = make([]LevelCrossingEvent, 0)
	adc.CurrentLevel = adc.Config.NumLevels / 2
	adc.LastEventTime = -1.0

	for i, sample := range signal {
		timestamp := float64(i) / sampleRate
		adc.ProcessSample(sample, timestamp)
	}

	// Calculate statistics
	duration := float64(len(signal)) / sampleRate
	adc.AverageRate = float64(len(adc.Events)) / duration

	// Nyquist samples would be len(signal) * 2 (for same bandwidth)
	// LC-ADC only generates events on changes
	adc.CompressionRatio = float64(len(signal)) / float64(len(adc.Events)+1)
}

// ReconstructSignal reconstructs the original signal from events
func (adc *LevelCrossingADC) ReconstructSignal(numSamples int, sampleRate float64) []float64 {
	signal := make([]float64, numSamples)

	// Sort events by timestamp
	sortedEvents := make([]LevelCrossingEvent, len(adc.Events))
	copy(sortedEvents, adc.Events)
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].Timestamp < sortedEvents[j].Timestamp
	})

	eventIdx := 0
	currentLevel := adc.Config.NumLevels / 2

	for i := 0; i < numSamples; i++ {
		timestamp := float64(i) / sampleRate

		// Apply all events up to this timestamp
		for eventIdx < len(sortedEvents) && sortedEvents[eventIdx].Timestamp <= timestamp {
			currentLevel = sortedEvents[eventIdx].Level
			eventIdx++
		}

		signal[i] = adc.levelToVoltage(currentLevel)
	}

	return signal
}

// CalculateNRMSE computes normalized RMS error vs original signal
func (adc *LevelCrossingADC) CalculateNRMSE(original []float64, reconstructed []float64) float64 {
	if len(original) != len(reconstructed) {
		return 1.0
	}

	// Calculate RMS error
	sumSquaredError := 0.0
	sumSquaredSignal := 0.0

	for i := range original {
		err := original[i] - reconstructed[i]
		sumSquaredError += err * err
		sumSquaredSignal += original[i] * original[i]
	}

	if sumSquaredSignal < 1e-10 {
		return 1.0
	}

	adc.NRMSE = math.Sqrt(sumSquaredError / sumSquaredSignal)
	return adc.NRMSE
}

// ============================================================================
// SECTION 5: Send-on-Delta Transmission
// ============================================================================

// SendOnDeltaConfig configures send-on-delta protocol
type SendOnDeltaConfig struct {
	// Delta threshold
	DeltaThreshold  float64 // Minimum change to trigger transmission

	// Temporal settings
	MaxIntervalMs   float64 // Maximum time between transmissions
	MinIntervalMs   float64 // Minimum time between transmissions

	// Encoding
	DeltaBits       int     // Bits for delta value
	TimestampBits   int     // Bits for timestamp

	// Power model
	TxPowerUW       float64 // Transmission power
	SleepPowerNW    float64 // Sleep mode power
}

// DefaultSendOnDeltaConfig returns typical send-on-delta configuration
func DefaultSendOnDeltaConfig() *SendOnDeltaConfig {
	return &SendOnDeltaConfig{
		DeltaThreshold: 0.01,
		MaxIntervalMs:  1000,
		MinIntervalMs:  1,
		DeltaBits:      8,
		TimestampBits:  16,
		TxPowerUW:      100,
		SleepPowerNW:   100,
	}
}

// DeltaPacket represents a send-on-delta transmission
type DeltaPacket struct {
	Timestamp      float64 // Transmission time
	DeltaValue     float64 // Change value
	QuantizedDelta int8    // Quantized delta
	PacketBits     int     // Total packet size
}

// SendOnDeltaEncoder implements send-on-delta compression
type SendOnDeltaEncoder struct {
	Config            *SendOnDeltaConfig
	LastValue         float64
	LastTxTime        float64
	Packets           []DeltaPacket

	// Statistics
	CompressionRatio  float64
	AverageDataRateBps float64
	AveragePowerUW    float64
}

// NewSendOnDeltaEncoder creates a new send-on-delta encoder
func NewSendOnDeltaEncoder(config *SendOnDeltaConfig) *SendOnDeltaEncoder {
	if config == nil {
		config = DefaultSendOnDeltaConfig()
	}
	return &SendOnDeltaEncoder{
		Config:  config,
		Packets: make([]DeltaPacket, 0),
	}
}

// ProcessSample processes a sample and decides whether to transmit
func (e *SendOnDeltaEncoder) ProcessSample(value float64, timestamp float64) *DeltaPacket {
	delta := value - e.LastValue
	timeSinceLastTx := timestamp - e.LastTxTime

	// Check if we should transmit
	shouldTransmit := false

	// Delta threshold exceeded
	if math.Abs(delta) >= e.Config.DeltaThreshold {
		shouldTransmit = true
	}

	// Maximum interval exceeded
	if timeSinceLastTx*1000 >= e.Config.MaxIntervalMs {
		shouldTransmit = true
	}

	// Minimum interval not yet passed
	if timeSinceLastTx*1000 < e.Config.MinIntervalMs {
		shouldTransmit = false
	}

	if !shouldTransmit {
		return nil
	}

	// Quantize delta
	maxDelta := e.Config.DeltaThreshold * math.Pow(2, float64(e.Config.DeltaBits-1))
	quantized := int8(math.Round(delta / e.Config.DeltaThreshold))
	if delta > maxDelta {
		quantized = int8(math.Pow(2, float64(e.Config.DeltaBits-1)) - 1)
	}
	if delta < -maxDelta {
		quantized = int8(-math.Pow(2, float64(e.Config.DeltaBits-1)))
	}

	packet := DeltaPacket{
		Timestamp:      timestamp,
		DeltaValue:     delta,
		QuantizedDelta: quantized,
		PacketBits:     e.Config.DeltaBits + e.Config.TimestampBits,
	}

	e.Packets = append(e.Packets, packet)
	e.LastValue = value
	e.LastTxTime = timestamp

	return &packet
}

// ProcessSignal processes an entire signal
func (e *SendOnDeltaEncoder) ProcessSignal(signal []float64, sampleRate float64) {
	e.Packets = make([]DeltaPacket, 0)
	e.LastValue = 0
	e.LastTxTime = 0

	for i, sample := range signal {
		timestamp := float64(i) / sampleRate
		e.ProcessSample(sample, timestamp)
	}

	// Calculate statistics
	duration := float64(len(signal)) / sampleRate
	totalBits := 0
	for _, p := range e.Packets {
		totalBits += p.PacketBits
	}

	// Original would be len(signal) * 16 bits (assuming 16-bit samples)
	originalBits := len(signal) * 16
	e.CompressionRatio = float64(originalBits) / float64(totalBits+1)
	e.AverageDataRateBps = float64(totalBits) / duration

	// Power: active during tx, sleep otherwise
	txTime := float64(len(e.Packets)) * float64(e.Config.DeltaBits+e.Config.TimestampBits) / 1e6 // Assuming 1Mbps
	sleepTime := duration - txTime
	e.AveragePowerUW = (txTime*e.Config.TxPowerUW + sleepTime*e.Config.SleepPowerNW/1000) / duration
}

// ReconstructSignal reconstructs signal from delta packets
func (e *SendOnDeltaEncoder) ReconstructSignal(numSamples int, sampleRate float64) []float64 {
	signal := make([]float64, numSamples)

	// Sort packets by timestamp
	sortedPackets := make([]DeltaPacket, len(e.Packets))
	copy(sortedPackets, e.Packets)
	sort.Slice(sortedPackets, func(i, j int) bool {
		return sortedPackets[i].Timestamp < sortedPackets[j].Timestamp
	})

	packetIdx := 0
	currentValue := 0.0

	for i := 0; i < numSamples; i++ {
		timestamp := float64(i) / sampleRate

		// Apply all packets up to this timestamp
		for packetIdx < len(sortedPackets) && sortedPackets[packetIdx].Timestamp <= timestamp {
			// Use quantized delta for reconstruction
			currentValue += float64(sortedPackets[packetIdx].QuantizedDelta) * e.Config.DeltaThreshold
			packetIdx++
		}

		signal[i] = currentValue
	}

	return signal
}

// ============================================================================
// SECTION 6: Edge Deployment Pipeline
// ============================================================================

// EdgeDeploymentConfig configures complete edge deployment pipeline
type EdgeDeploymentConfig struct {
	// Target device
	DeviceType      string  // "mcu", "fpga", "asic"

	// TinyML settings
	TinyMLConfig    *TinyMLConfig

	// CIM compiler settings
	CompilerConfig  *CIMCompilerConfig

	// Spike encoding
	SpikeEncoding   string  // "sigma_delta", "level_crossing", "send_on_delta"
	EncoderConfig   interface{}

	// Optimization
	EnableProfiling bool
	EnableBenchmark bool
}

// DeploymentReport contains deployment analysis results
type DeploymentReport struct {
	// Model info
	ModelName        string
	OriginalSizeKB   float64
	DeployedSizeKB   float64
	CompressionRatio float64

	// Memory
	FlashUsageKB     float64
	SRAMUsageKB      float64
	FlashUtilization float64
	SRAMUtilization  float64

	// Performance
	InferenceLatencyMs float64
	ThroughputInfPerS  float64

	// Energy
	EnergyPerInfUJ   float64
	PowerMW          float64

	// CIM-specific
	CrossbarUtilization float64
	MVMCycles           int
	CompilerSpeedup     float64

	// Spike encoding
	SpikeCompressionRatio float64
	AverageSpikeRate      float64

	// Feasibility
	IsFeasible       bool
	Issues           []string
}

// EdgeDeploymentPipeline manages complete edge deployment
type EdgeDeploymentPipeline struct {
	Config         *EdgeDeploymentConfig
	Model          *TinyMLModel
	Compiler       *CIMCompiler
	SpikeEncoder   interface{}
	Report         *DeploymentReport
}

// NewEdgeDeploymentPipeline creates a new deployment pipeline
func NewEdgeDeploymentPipeline(config *EdgeDeploymentConfig) *EdgeDeploymentPipeline {
	if config == nil {
		config = &EdgeDeploymentConfig{
			DeviceType:    "mcu",
			TinyMLConfig:  DefaultTinyMLConfig(),
			CompilerConfig: DefaultCIMCompilerConfig(),
			SpikeEncoding: "sigma_delta",
		}
	}

	pipeline := &EdgeDeploymentPipeline{
		Config:   config,
		Model:    NewTinyMLModel(config.TinyMLConfig),
		Compiler: NewCIMCompiler(config.CompilerConfig),
		Report:   &DeploymentReport{},
	}

	// Initialize spike encoder based on type
	switch config.SpikeEncoding {
	case "sigma_delta":
		sdConfig, ok := config.EncoderConfig.(*SigmaDeltaConfig)
		if !ok {
			sdConfig = DefaultSigmaDeltaConfig()
		}
		pipeline.SpikeEncoder = NewSigmaDeltaEncoder(sdConfig)
	case "level_crossing":
		lcConfig, ok := config.EncoderConfig.(*LevelCrossingConfig)
		if !ok {
			lcConfig = DefaultLevelCrossingConfig()
		}
		pipeline.SpikeEncoder = NewLevelCrossingADC(lcConfig)
	case "send_on_delta":
		sodConfig, ok := config.EncoderConfig.(*SendOnDeltaConfig)
		if !ok {
			sodConfig = DefaultSendOnDeltaConfig()
		}
		pipeline.SpikeEncoder = NewSendOnDeltaEncoder(sodConfig)
	}

	return pipeline
}

// DeployModel performs complete deployment analysis
func (p *EdgeDeploymentPipeline) DeployModel(modelName string, layers [][]float64) (*DeploymentReport, error) {
	p.Report.ModelName = modelName

	// Calculate original size
	originalParams := 0
	for _, layer := range layers {
		originalParams += len(layer)
	}
	p.Report.OriginalSizeKB = float64(originalParams*4) / 1024.0 // FP32

	// Quantize each layer
	for i, layer := range layers {
		// Convert 1D to 2D for quantization
		weights2D := make([][]float64, 1)
		weights2D[0] = layer

		tinyLayer, err := p.Model.QuantizeWeights(weights2D, p.Config.TinyMLConfig.WeightBits, true)
		if err != nil {
			return nil, fmt.Errorf("failed to quantize layer %d: %v", i, err)
		}
		tinyLayer.Name = fmt.Sprintf("layer_%d", i)
		tinyLayer.InputShape = []int{1, len(layer)}
		tinyLayer.OutputShape = []int{1, len(layer)}
		p.Model.Layers = append(p.Model.Layers, tinyLayer)
	}

	// Analyze memory
	p.Model.AnalyzeMemory()

	// Calculate deployed size
	p.Report.DeployedSizeKB = float64(p.Model.TotalWeightBytes) / 1024.0
	p.Report.CompressionRatio = p.Report.OriginalSizeKB / p.Report.DeployedSizeKB

	// Memory utilization
	p.Report.FlashUsageKB = p.Report.DeployedSizeKB
	p.Report.SRAMUsageKB = float64(p.Model.PeakActivationBytes+p.Model.TotalScratchBytes) / 1024.0
	p.Report.FlashUtilization = p.Report.FlashUsageKB / float64(p.Config.TinyMLConfig.FlashSizeKB)
	p.Report.SRAMUtilization = p.Report.SRAMUsageKB / float64(p.Config.TinyMLConfig.SRAMSizeKB)

	// Compile for CIM
	for i, layer := range layers {
		layerName := fmt.Sprintf("layer_%d", i)
		weights2D := make([][]float64, 1)
		weights2D[0] = layer

		err := p.Compiler.CompileLayer(layerName, weights2D)
		if err != nil {
			return nil, fmt.Errorf("failed to compile layer %d: %v", i, err)
		}
	}

	// Optimize schedule
	p.Compiler.OptimizeSchedule()

	// CIM metrics
	p.Report.CrossbarUtilization = p.Compiler.Utilization
	p.Report.MVMCycles = p.Compiler.TotalCycles
	p.Report.CompilerSpeedup = p.Compiler.Speedup

	// Estimate performance
	// Assume 1 GHz clock for CIM
	p.Report.InferenceLatencyMs = float64(p.Compiler.TotalCycles) / 1e6
	p.Report.ThroughputInfPerS = 1000.0 / p.Report.InferenceLatencyMs

	// Energy estimate
	p.Report.EnergyPerInfUJ = p.Compiler.TotalEnergyPJ / 1e6
	p.Report.PowerMW = p.Report.EnergyPerInfUJ * p.Report.ThroughputInfPerS / 1000.0

	// Check feasibility
	feasible, issues := p.Model.CheckDeploymentFeasibility()
	p.Report.IsFeasible = feasible
	p.Report.Issues = issues

	return p.Report, nil
}

// ============================================================================
// SECTION 7: Model Serialization for Edge
// ============================================================================

// EdgeModelFormat defines serialization format
type EdgeModelFormat struct {
	Version         string            `json:"version"`
	ModelName       string            `json:"model_name"`
	TargetDevice    string            `json:"target_device"`
	Quantization    string            `json:"quantization"`
	Layers          []EdgeLayerFormat `json:"layers"`
	TotalParams     int               `json:"total_params"`
	TotalSizeBytes  int               `json:"total_size_bytes"`
}

// EdgeLayerFormat defines layer serialization
type EdgeLayerFormat struct {
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	InputShape      []int     `json:"input_shape"`
	OutputShape     []int     `json:"output_shape"`
	WeightScale     []float32 `json:"weight_scale"`
	WeightZeroPoint []int8    `json:"weight_zero_point"`
	WeightsBase64   string    `json:"weights_base64,omitempty"`
}

// SerializeForEdge serializes model for edge deployment
func (p *EdgeDeploymentPipeline) SerializeForEdge(writer io.Writer, compress bool) error {
	format := EdgeModelFormat{
		Version:      "1.0",
		ModelName:    p.Report.ModelName,
		TargetDevice: p.Config.DeviceType,
		Quantization: fmt.Sprintf("INT%d", p.Config.TinyMLConfig.WeightBits),
		Layers:       make([]EdgeLayerFormat, len(p.Model.Layers)),
	}

	for i, layer := range p.Model.Layers {
		format.Layers[i] = EdgeLayerFormat{
			Name:            layer.Name,
			Type:            layer.Type,
			InputShape:      layer.InputShape,
			OutputShape:     layer.OutputShape,
			WeightScale:     layer.WeightScale,
			WeightZeroPoint: layer.WeightZeroPoint,
		}
		format.TotalParams += layer.WeightSizeBytes
	}

	format.TotalSizeBytes = format.TotalParams

	if compress {
		gzWriter := gzip.NewWriter(writer)
		defer gzWriter.Close()
		return json.NewEncoder(gzWriter).Encode(format)
	}

	return json.NewEncoder(writer).Encode(format)
}

// SerializeBinary writes binary format for embedded deployment
func (p *EdgeDeploymentPipeline) SerializeBinary(writer io.Writer) error {
	// Write header
	header := make([]byte, 16)
	copy(header[0:4], []byte("EDGE"))
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(p.Model.Layers)))
	binary.LittleEndian.PutUint32(header[8:12], uint32(p.Config.TinyMLConfig.WeightBits))
	binary.LittleEndian.PutUint32(header[12:16], uint32(p.Model.TotalWeightBytes))

	if _, err := writer.Write(header); err != nil {
		return err
	}

	// Write layer data
	for _, layer := range p.Model.Layers {
		// Layer header: type (1 byte) + dims (8 bytes)
		layerHeader := make([]byte, 9)
		switch layer.Type {
		case "dense":
			layerHeader[0] = 0x01
		case "conv2d":
			layerHeader[0] = 0x02
		case "depthwise":
			layerHeader[0] = 0x03
		}

		if len(layer.InputShape) >= 2 {
			binary.LittleEndian.PutUint32(layerHeader[1:5], uint32(layer.InputShape[1]))
		}
		if len(layer.OutputShape) >= 2 {
			binary.LittleEndian.PutUint32(layerHeader[5:9], uint32(layer.OutputShape[1]))
		}

		if _, err := writer.Write(layerHeader); err != nil {
			return err
		}

		// Write scales
		for _, s := range layer.WeightScale {
			if err := binary.Write(writer, binary.LittleEndian, s); err != nil {
				return err
			}
		}

		// Write weights
		if layer.WeightsInt8 != nil {
			if _, err := writer.Write([]byte(string(layer.WeightsInt8))); err != nil {
				return err
			}
		} else if layer.WeightsInt4 != nil {
			if _, err := writer.Write([]byte(string(layer.WeightsInt4))); err != nil {
				return err
			}
		}
	}

	return nil
}

// ============================================================================
// SECTION 8: Multi-Threshold Delta Modulation
// ============================================================================

// MultiThresholdConfig configures multi-threshold delta modulation
type MultiThresholdConfig struct {
	NumThresholds   int       // Number of threshold levels
	ThresholdLevels []float64 // Threshold values
	Polarity        bool      // True for bipolar (+/-)

	// Adaptive settings
	AdaptiveEnabled bool
	AdaptationRate  float64
}

// DefaultMultiThresholdConfig returns typical configuration
func DefaultMultiThresholdConfig() *MultiThresholdConfig {
	return &MultiThresholdConfig{
		NumThresholds:   4,
		ThresholdLevels: []float64{0.125, 0.25, 0.5, 1.0},
		Polarity:        true,
		AdaptiveEnabled: true,
		AdaptationRate:  0.01,
	}
}

// MultiThresholdSpike represents a multi-level spike
type MultiThresholdSpike struct {
	Timestamp    float64
	Level        int    // Which threshold was crossed
	Direction    int    // +1 or -1
	Magnitude    float64
}

// MultiThresholdEncoder implements multi-threshold delta modulation
type MultiThresholdEncoder struct {
	Config          *MultiThresholdConfig
	State           float64
	PreviousInput   float64
	Spikes          []MultiThresholdSpike

	// Adaptive thresholds
	CurrentThresholds []float64

	// Statistics
	SpikeCounts     []int // Spikes per level
	CompressionRatio float64
	SNR             float64
}

// NewMultiThresholdEncoder creates a new multi-threshold encoder
func NewMultiThresholdEncoder(config *MultiThresholdConfig) *MultiThresholdEncoder {
	if config == nil {
		config = DefaultMultiThresholdConfig()
	}

	encoder := &MultiThresholdEncoder{
		Config:            config,
		Spikes:            make([]MultiThresholdSpike, 0),
		CurrentThresholds: make([]float64, len(config.ThresholdLevels)),
		SpikeCounts:       make([]int, len(config.ThresholdLevels)),
	}

	copy(encoder.CurrentThresholds, config.ThresholdLevels)
	return encoder
}

// EncodeSample encodes a sample with multi-threshold delta modulation
func (e *MultiThresholdEncoder) EncodeSample(input float64, timestamp float64) []MultiThresholdSpike {
	spikes := make([]MultiThresholdSpike, 0)

	delta := input - e.PreviousInput
	e.State += delta

	// Check thresholds from largest to smallest
	for i := len(e.CurrentThresholds) - 1; i >= 0; i-- {
		thresh := e.CurrentThresholds[i]

		// Positive crossings
		for e.State >= thresh {
			spike := MultiThresholdSpike{
				Timestamp: timestamp,
				Level:     i,
				Direction: 1,
				Magnitude: thresh,
			}
			spikes = append(spikes, spike)
			e.SpikeCounts[i]++
			e.State -= thresh
		}

		// Negative crossings (if bipolar)
		if e.Config.Polarity {
			for e.State <= -thresh {
				spike := MultiThresholdSpike{
					Timestamp: timestamp,
					Level:     i,
					Direction: -1,
					Magnitude: thresh,
				}
				spikes = append(spikes, spike)
				e.SpikeCounts[i]++
				e.State += thresh
			}
		}
	}

	// Adaptive threshold adjustment
	if e.Config.AdaptiveEnabled {
		for i := range e.CurrentThresholds {
			// Increase threshold if too many spikes, decrease if too few
			targetRate := 1.0 / float64(i+1) // Higher thresholds should fire less
			actualRate := float64(e.SpikeCounts[i]) / (timestamp + 1)
			adjustment := e.Config.AdaptationRate * (actualRate - targetRate)
			e.CurrentThresholds[i] *= (1 + adjustment)

			// Clamp to reasonable range
			if e.CurrentThresholds[i] < e.Config.ThresholdLevels[i]*0.5 {
				e.CurrentThresholds[i] = e.Config.ThresholdLevels[i] * 0.5
			}
			if e.CurrentThresholds[i] > e.Config.ThresholdLevels[i]*2.0 {
				e.CurrentThresholds[i] = e.Config.ThresholdLevels[i] * 2.0
			}
		}
	}

	e.PreviousInput = input
	e.Spikes = append(e.Spikes, spikes...)

	return spikes
}

// EncodeSignal encodes an entire signal
func (e *MultiThresholdEncoder) EncodeSignal(signal []float64, sampleRate float64) {
	e.State = 0
	e.PreviousInput = 0
	e.Spikes = make([]MultiThresholdSpike, 0)
	for i := range e.SpikeCounts {
		e.SpikeCounts[i] = 0
	}

	for i, sample := range signal {
		timestamp := float64(i) / sampleRate
		e.EncodeSample(sample, timestamp)
	}

	totalSpikes := 0
	for _, count := range e.SpikeCounts {
		totalSpikes += count
	}

	// Each spike needs: timestamp + level + direction
	bitsPerSpike := 16 + int(math.Ceil(math.Log2(float64(e.Config.NumThresholds)))) + 1
	totalSpikeBits := totalSpikes * bitsPerSpike
	originalBits := len(signal) * 16

	e.CompressionRatio = float64(originalBits) / float64(totalSpikeBits+1)
}

// DecodeSignal reconstructs signal from multi-threshold spikes
func (e *MultiThresholdEncoder) DecodeSignal(numSamples int, sampleRate float64) []float64 {
	signal := make([]float64, numSamples)

	// Sort spikes by timestamp
	sortedSpikes := make([]MultiThresholdSpike, len(e.Spikes))
	copy(sortedSpikes, e.Spikes)
	sort.Slice(sortedSpikes, func(i, j int) bool {
		return sortedSpikes[i].Timestamp < sortedSpikes[j].Timestamp
	})

	spikeIdx := 0
	currentValue := 0.0

	for i := 0; i < numSamples; i++ {
		timestamp := float64(i) / sampleRate

		for spikeIdx < len(sortedSpikes) && sortedSpikes[spikeIdx].Timestamp <= timestamp {
			currentValue += float64(sortedSpikes[spikeIdx].Direction) * sortedSpikes[spikeIdx].Magnitude
			spikeIdx++
		}

		signal[i] = currentValue
	}

	return signal
}

// ============================================================================
// SECTION 9: Benchmark and Profiling
// ============================================================================

// EdgeBenchmarkResult contains benchmark results
type EdgeBenchmarkResult struct {
	// Encoding benchmarks
	EncodingMethod     string
	EncodingTimeUs     float64
	CompressionRatio   float64
	ReconstructionNRMSE float64

	// CIM compilation benchmarks
	CompilationTimeMs  float64
	InstructionCount   int
	Utilization        float64
	Speedup            float64

	// Deployment benchmarks
	QuantizationTimeMs float64
	ModelSizeKB        float64
	MemoryFit          bool

	// Energy benchmarks
	EnergyPerInfUJ     float64
	TOPSW              float64
}

// RunEdgeBenchmarks runs comprehensive edge deployment benchmarks
func RunEdgeBenchmarks(signal []float64, weights [][]float64) *EdgeBenchmarkResult {
	result := &EdgeBenchmarkResult{}

	// Benchmark sigma-delta encoding
	sdEncoder := NewSigmaDeltaEncoder(DefaultSigmaDeltaConfig())
	spikes := sdEncoder.EncodeSignal(signal, 1000.0)
	result.EncodingMethod = "sigma_delta"
	result.CompressionRatio = sdEncoder.CompressionRatio

	// Reconstruct and calculate error
	reconstructed := sdEncoder.DecodeSpikes(spikes)
	if len(reconstructed) == len(signal) {
		sumSqErr := 0.0
		sumSqSig := 0.0
		for i := range signal {
			err := signal[i] - reconstructed[i]
			sumSqErr += err * err
			sumSqSig += signal[i] * signal[i]
		}
		if sumSqSig > 0 {
			result.ReconstructionNRMSE = math.Sqrt(sumSqErr / sumSqSig)
		}
	}

	// Benchmark CIM compilation
	compiler := NewCIMCompiler(DefaultCIMCompilerConfig())
	for i, layer := range weights {
		layerName := fmt.Sprintf("layer_%d", i)
		weights2D := make([][]float64, 1)
		weights2D[0] = layer
		compiler.CompileLayer(layerName, weights2D)
	}
	compiler.OptimizeSchedule()

	result.InstructionCount = len(compiler.Instructions)
	result.Utilization = compiler.Utilization
	result.Speedup = compiler.Speedup

	// Benchmark TinyML deployment
	pipeline := NewEdgeDeploymentPipeline(&EdgeDeploymentConfig{
		DeviceType:    "mcu",
		TinyMLConfig:  DefaultTinyMLConfig(),
		CompilerConfig: DefaultCIMCompilerConfig(),
		SpikeEncoding: "sigma_delta",
	})

	flatWeights := make([][]float64, len(weights))
	for i, w := range weights {
		flatWeights[i] = w
	}

	report, _ := pipeline.DeployModel("benchmark_model", flatWeights)
	if report != nil {
		result.ModelSizeKB = report.DeployedSizeKB
		result.MemoryFit = report.IsFeasible
		result.EnergyPerInfUJ = report.EnergyPerInfUJ

		if report.EnergyPerInfUJ > 0 {
			// Calculate TOPS/W
			// Assume 1 MAC per weight per inference
			totalMACs := 0
			for _, w := range weights {
				totalMACs += len(w)
			}
			opsPerInf := float64(totalMACs * 2) // 2 ops per MAC
			energyJ := report.EnergyPerInfUJ * 1e-6
			result.TOPSW = (opsPerInf / energyJ) / 1e12
		}
	}

	return result
}

// ============================================================================
// SECTION 10: Summary ASCII Diagrams
// ============================================================================

/*
Edge Deployment Pipeline:
=========================

┌─────────────────────────────────────────────────────────────────────────────┐
│                        EDGE DEPLOYMENT PIPELINE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │  FP32 Model │───▶│ Quantization│───▶│ CIM Compile │───▶│  Deploy to  │  │
│  │  (PyTorch)  │    │  INT8/INT4  │    │   Mapping   │    │    MCU      │  │
│  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘  │
│                                                                             │
│  Compression: 4×    Precision: 8-bit   Speedup: 3.2×     Flash: <512KB     │
│  Accuracy: <1%↓     Per-channel        (CIM-MLC)         SRAM: <128KB      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘


Sigma-Delta Spike Encoding:
===========================

Input Signal    ──▶  Δ-Modulator  ──▶  Spike Train  ──▶  SNN Processing
    │                    │                  │                   │
    ▼                    ▼                  ▼                   ▼
┌────────┐         ┌────────┐         ┌────────┐          ┌────────┐
│ Analog │ ────▶   │ ∫(Δx)  │ ────▶   │ +1/0/-1│  ────▶   │ Neuro- │
│ Input  │         │ > θ ?  │         │ events │          │ morphic│
└────────┘         └────────┘         └────────┘          └────────┘
                        │
            State = State + Δ
            if State > θ: spike +1, State -= θ
            if State < -θ: spike -1, State += θ

Compression: >100× for sparse signals
Energy: ADC-free, ~10× lower power


Level-Crossing ADC (Send-on-Delta):
===================================

       ┌─────────────────────────────────────────┐
   1.8V│    ●──●      ●───●                      │
       │   /    \    /     \        ●──●         │
       │  /      \  /       \      /    \        │
       │ ●        ●●         ●────●      ●       │
       │                                         │
   0.0V└─────────────────────────────────────────┘
        Time →

Events:  ↑  ↑  ↓  ↓  ↑  ↑  ↑  ↓  ↓  ↓  ↑  ↑  ↓

- Only transmit on signal changes
- Compression: >125× for ENG signals
- NRMSE: ~4% reconstruction error
- Power: µW-level operation


CIM-MLC Multi-Level Compilation:
================================

┌────────────────────────────────────────────────────────────────┐
│                    CIM-MLC Framework                           │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  Level 1: Graph Optimization                                   │
│  ┌──────────────────────────────────────────────────┐         │
│  │ Conv ─▶ BN ─▶ ReLU  ══▶  FusedConvBNReLU        │         │
│  └──────────────────────────────────────────────────┘         │
│                                                                │
│  Level 2: Tiling Strategy                                      │
│  ┌──────────────────────────────────────────────────┐         │
│  │ 1024×1024 weight ══▶ 4×4 tiles of 256×256       │         │
│  │ Dataflow: output_stationary / weight_stationary  │         │
│  └──────────────────────────────────────────────────┘         │
│                                                                │
│  Level 3: Instruction Generation                               │
│  ┌──────────────────────────────────────────────────┐         │
│  │ LOAD → MVM → ADD → ACT → STORE (pipelined)      │         │
│  └──────────────────────────────────────────────────┘         │
│                                                                │
│  Result: 3.2× speedup vs prior CIM compilers                   │
│                                                                │
└────────────────────────────────────────────────────────────────┘


TinyML Memory Layout (MCU):
===========================

┌─────────────────────────────────────────┐
│              FLASH (512KB)              │
├─────────────────────────────────────────┤
│  ┌─────────────────────────────────┐   │
│  │    INT8 Weights (~300KB)        │   │
│  │    ─────────────────────────    │   │
│  │    Layer 1: 64KB                │   │
│  │    Layer 2: 128KB               │   │
│  │    Layer 3: 96KB                │   │
│  │    Scales: 12KB                 │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │    Code + Constants (~100KB)    │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │    Free (~112KB)                │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│              SRAM (128KB)               │
├─────────────────────────────────────────┤
│  ┌─────────────────────────────────┐   │
│  │    Input Buffer (32KB)          │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │    Output Buffer (32KB)         │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │    Scratch Pad (32KB)           │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │    Stack + Heap (32KB)          │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘


Performance Comparison:
=======================

┌────────────────────────────────────────────────────────────────┐
│              Edge CIM Performance Metrics                      │
├────────────────┬───────────────┬───────────────┬──────────────┤
│ Metric         │ Traditional   │ SRAM-CIM      │ Improvement  │
├────────────────┼───────────────┼───────────────┼──────────────┤
│ Energy/inf     │ 100 µJ        │ 15 µJ         │ 6.7×         │
│ Latency        │ 20 ms         │ 3 ms          │ 6.7×         │
│ TOPS/W         │ 1             │ 687           │ 687×         │
│ Memory BW      │ 10 GB/s       │ 1 TB/s (int)  │ 100×         │
│ Model size     │ 4 MB (FP32)   │ 0.5 MB (INT8) │ 8×           │
└────────────────┴───────────────┴───────────────┴──────────────┘

Key Findings:
- SRAM-CIM achieves 687.5 TOPS/W (65nm, drive-strength based)
- Time-domain CIM: 1241 GOPS, 37.01 TOPS/W (28nm)
- INT4 quantization: 8× model size reduction
- Level-crossing ADC: >125× data compression for sparse signals
- CIM-MLC compiler: 3.2× speedup with multi-level scheduling
- CLSA-CIM: 17.9× utilization improvement, 29.2× speedup

*/

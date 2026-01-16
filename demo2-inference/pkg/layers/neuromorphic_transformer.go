// Package layers provides neuromorphic processor simulation and transformer CIM acceleration.
//
// This module implements:
// - Loihi-style neuromorphic processor architecture with mesh interconnect
// - Asynchronous spike routing with Address-Event Representation (AER)
// - Programmable neuron models with microcode-like instructions
// - Transformer attention acceleration via analog compute-in-memory
// - KV cache optimization using gain cell and SRAM-based CIM
// - Hybrid analog-digital token pruning for energy efficiency
//
// Based on:
// - Intel Loihi 2 architecture (Open Neuromorphic, 2024)
// - Nature Computational Science 2025: AIMC attention mechanism
// - Hybrid analog-digital attention accelerator (IEEE 2024)
package layers

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
)

// =============================================================================
// Part 1: Neuromorphic Processor Architecture (Loihi-Style)
// =============================================================================

// NeuroCoreConfig configures a neuromorphic core
type NeuroCoreConfig struct {
	NumCompartments   int     // Number of neuron compartments (default: 1024)
	SynapticMemoryKB  int     // Synaptic memory in KB (default: 16)
	MicrocodeSlots    int     // Number of microcode instruction slots
	GradedSpikeWidth  int     // Graded spike payload bits (default: 32)
	TimestepNs        float64 // Minimum timestep in nanoseconds (default: 200)
	LearnEnabled      bool    // Enable on-chip learning
	ThreeFactorRule   bool    // Support three-factor learning rules
}

// DefaultNeuroCoreConfig returns default Loihi-2 style configuration
func DefaultNeuroCoreConfig() NeuroCoreConfig {
	return NeuroCoreConfig{
		NumCompartments:  1024,
		SynapticMemoryKB: 16,
		MicrocodeSlots:   32,
		GradedSpikeWidth: 32,
		TimestepNs:       200,
		LearnEnabled:     true,
		ThreeFactorRule:  true,
	}
}

// Compartment represents a programmable neuron compartment
type Compartment struct {
	ID              int
	Registers       []float64 // Multi-register state
	Voltage         float64
	Current         float64
	Threshold       float64
	RefractoryCount int
	SpikeCount      int
	LastSpikeTime   int
	Microcode       []MicrocodeInstruction
	InputBuffer     []GradedSpike
	mu              sync.Mutex
}

// NewCompartment creates a new neuron compartment
func NewCompartment(id int, numRegisters int) *Compartment {
	return &Compartment{
		ID:          id,
		Registers:   make([]float64, numRegisters),
		Voltage:     0,
		Threshold:   1.0,
		Microcode:   make([]MicrocodeInstruction, 0),
		InputBuffer: make([]GradedSpike, 0),
	}
}

// MicrocodeOp defines microcode operation types
type MicrocodeOp int

const (
	OpNOP MicrocodeOp = iota
	OpADD             // Add registers
	OpMUL             // Multiply registers
	OpMAC             // Multiply-accumulate
	OpCMP             // Compare (threshold)
	OpDECAY           // Exponential decay
	OpRESET           // Reset to value
	OpSPIKE           // Generate spike
	OpLOAD            // Load from memory
	OpSTORE           // Store to memory
	OpACT             // Activation function
	OpLEARN           // Learning update
)

// MicrocodeInstruction represents a programmable neuron operation
type MicrocodeInstruction struct {
	Op       MicrocodeOp
	DestReg  int
	SrcReg1  int
	SrcReg2  int
	Imm      float64 // Immediate value
	CondFlag bool    // Conditional execution
}

// Execute runs a microcode instruction on the compartment
func (c *Compartment) Execute(inst MicrocodeInstruction) {
	switch inst.Op {
	case OpNOP:
		// No operation
	case OpADD:
		c.Registers[inst.DestReg] = c.Registers[inst.SrcReg1] + c.Registers[inst.SrcReg2]
	case OpMUL:
		c.Registers[inst.DestReg] = c.Registers[inst.SrcReg1] * c.Registers[inst.SrcReg2]
	case OpMAC:
		c.Registers[inst.DestReg] += c.Registers[inst.SrcReg1] * c.Registers[inst.SrcReg2]
	case OpCMP:
		if c.Registers[inst.SrcReg1] >= inst.Imm {
			c.Registers[inst.DestReg] = 1.0
		} else {
			c.Registers[inst.DestReg] = 0.0
		}
	case OpDECAY:
		c.Registers[inst.DestReg] *= math.Exp(-inst.Imm)
	case OpRESET:
		c.Registers[inst.DestReg] = inst.Imm
	case OpACT:
		// Apply activation (ReLU, sigmoid, etc.)
		val := c.Registers[inst.SrcReg1]
		switch int(inst.Imm) {
		case 0: // ReLU
			if val < 0 {
				val = 0
			}
		case 1: // Sigmoid
			val = 1.0 / (1.0 + math.Exp(-val))
		case 2: // Tanh
			val = math.Tanh(val)
		}
		c.Registers[inst.DestReg] = val
	}
}

// GradedSpike represents a spike with integer payload (Loihi 2 feature)
type GradedSpike struct {
	SourceCore   int
	SourceNeuron int
	DestCore     int
	DestNeuron   int
	Payload      int32 // 32-bit graded value
	Timestamp    int
}

// ToAERPacket converts graded spike to 32-bit AER packet
func (gs GradedSpike) ToAERPacket() uint32 {
	// Pack: [source_id:10][dest_id:10][payload_msb:12]
	packet := uint32(gs.SourceNeuron&0x3FF) << 22
	packet |= uint32(gs.DestNeuron&0x3FF) << 12
	packet |= uint32(gs.Payload>>20) & 0xFFF
	return packet
}

// FromAERPacket decodes a 32-bit AER packet
func FromAERPacket(packet uint32) GradedSpike {
	return GradedSpike{
		SourceNeuron: int((packet >> 22) & 0x3FF),
		DestNeuron:   int((packet >> 12) & 0x3FF),
		Payload:      int32((packet & 0xFFF) << 20),
	}
}

// SynapticMemory represents on-core synaptic weight storage
type SynapticMemory struct {
	Weights     [][]float64 // Weight matrix
	Delays      [][]int     // Axonal delays
	PlasticMask [][]bool    // Which synapses are plastic
	mu          sync.RWMutex
}

// NewSynapticMemory creates synaptic memory for a core
func NewSynapticMemory(preNeurons, postNeurons int) *SynapticMemory {
	weights := make([][]float64, preNeurons)
	delays := make([][]int, preNeurons)
	plastic := make([][]bool, preNeurons)
	for i := range weights {
		weights[i] = make([]float64, postNeurons)
		delays[i] = make([]int, postNeurons)
		plastic[i] = make([]bool, postNeurons)
	}
	return &SynapticMemory{
		Weights:     weights,
		Delays:      delays,
		PlasticMask: plastic,
	}
}

// NeuroCore represents a single neuromorphic processing core
type NeuroCore struct {
	ID           int
	Config       NeuroCoreConfig
	Compartments []*Compartment
	Synapses     *SynapticMemory
	InputQueue   chan GradedSpike
	OutputQueue  chan GradedSpike
	RouterLinks  [4]chan GradedSpike // N, S, E, W
	CurrentTime  int
	BarrierSync  *sync.WaitGroup
	mu           sync.Mutex
}

// NewNeuroCore creates a new neuromorphic core
func NewNeuroCore(id int, config NeuroCoreConfig) *NeuroCore {
	compartments := make([]*Compartment, config.NumCompartments)
	for i := range compartments {
		compartments[i] = NewCompartment(i, 8) // 8 registers per compartment
	}

	return &NeuroCore{
		ID:           id,
		Config:       config,
		Compartments: compartments,
		Synapses:     NewSynapticMemory(config.NumCompartments, config.NumCompartments),
		InputQueue:   make(chan GradedSpike, 1024),
		OutputQueue:  make(chan GradedSpike, 1024),
		CurrentTime:  0,
	}
}

// ProcessTimestep executes one algorithmic timestep
func (nc *NeuroCore) ProcessTimestep() []GradedSpike {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	var outputSpikes []GradedSpike

	// Process all queued input spikes
	for len(nc.InputQueue) > 0 {
		select {
		case spike := <-nc.InputQueue:
			if spike.DestNeuron < len(nc.Compartments) {
				nc.Compartments[spike.DestNeuron].InputBuffer = append(
					nc.Compartments[spike.DestNeuron].InputBuffer, spike)
			}
		default:
			break
		}
	}

	// Update all compartments
	for _, comp := range nc.Compartments {
		// Process inputs
		var inputCurrent float64
		for _, spike := range comp.InputBuffer {
			// Apply synaptic weight
			weight := nc.Synapses.Weights[spike.SourceNeuron][comp.ID]
			inputCurrent += weight * float64(spike.Payload)
		}
		comp.InputBuffer = comp.InputBuffer[:0]

		// Execute microcode program
		for _, inst := range comp.Microcode {
			comp.Execute(inst)
		}

		// Default LIF if no microcode
		if len(comp.Microcode) == 0 {
			// Leaky integrate
			comp.Voltage = comp.Voltage*0.9 + inputCurrent*0.1
			comp.Current = inputCurrent

			// Fire check
			if comp.Voltage >= comp.Threshold && comp.RefractoryCount == 0 {
				// Generate graded spike
				spike := GradedSpike{
					SourceCore:   nc.ID,
					SourceNeuron: comp.ID,
					Payload:      int32(comp.Voltage * 1000), // Scale to integer
					Timestamp:    nc.CurrentTime,
				}
				outputSpikes = append(outputSpikes, spike)
				comp.Voltage = 0
				comp.RefractoryCount = 5
				comp.SpikeCount++
				comp.LastSpikeTime = nc.CurrentTime
			}

			if comp.RefractoryCount > 0 {
				comp.RefractoryCount--
			}
		}
	}

	nc.CurrentTime++
	return outputSpikes
}

// MeshRouter handles spike routing between cores
type MeshRouter struct {
	GridRows int
	GridCols int
	Cores    [][]*NeuroCore
	mu       sync.RWMutex
}

// NewMeshRouter creates a 2D mesh interconnect
func NewMeshRouter(rows, cols int, config NeuroCoreConfig) *MeshRouter {
	cores := make([][]*NeuroCore, rows)
	for i := range cores {
		cores[i] = make([]*NeuroCore, cols)
		for j := range cores[i] {
			coreID := i*cols + j
			cores[i][j] = NewNeuroCore(coreID, config)
		}
	}

	router := &MeshRouter{
		GridRows: rows,
		GridCols: cols,
		Cores:    cores,
	}

	// Initialize router links (N, S, E, W)
	for i := range cores {
		for j := range cores[i] {
			for k := 0; k < 4; k++ {
				cores[i][j].RouterLinks[k] = make(chan GradedSpike, 256)
			}
		}
	}

	return router
}

// RouteSpike uses XY routing to deliver spike to destination core
func (mr *MeshRouter) RouteSpike(spike GradedSpike, srcRow, srcCol int) {
	destCoreID := spike.DestCore
	destRow := destCoreID / mr.GridCols
	destCol := destCoreID % mr.GridCols

	// XY routing: first X (columns), then Y (rows)
	currentRow, currentCol := srcRow, srcCol

	// Route in X direction first
	for currentCol != destCol {
		if currentCol < destCol {
			currentCol++ // East
		} else {
			currentCol-- // West
		}
	}

	// Then route in Y direction
	for currentRow != destRow {
		if currentRow < destRow {
			currentRow++ // South
		} else {
			currentRow-- // North
		}
	}

	// Deliver to destination core
	if destRow < mr.GridRows && destCol < mr.GridCols {
		select {
		case mr.Cores[destRow][destCol].InputQueue <- spike:
		default:
			// Queue full, drop spike (realistic behavior)
		}
	}
}

// NeuromorphicProcessor represents a complete multi-core neuromorphic processor
type NeuromorphicProcessor struct {
	Config       NeuromorphicProcessorConfig
	Mesh         *MeshRouter
	GlobalTime   int
	PowerUw      float64 // Estimated power in microwatts
	SpikeCounter int
	mu           sync.Mutex
}

// NeuromorphicProcessorConfig configures the complete processor
type NeuromorphicProcessorConfig struct {
	MeshRows      int             // Number of core rows
	MeshCols      int             // Number of core columns
	CoreConfig    NeuroCoreConfig // Per-core configuration
	ProcessNode   int             // Process node in nm (e.g., 4 for Intel 4)
	TotalNeurons  int             // Maximum neurons
	TotalSynapses int             // Maximum synapses
}

// DefaultNeuromorphicProcessorConfig returns Loihi 2 style configuration
func DefaultNeuromorphicProcessorConfig() NeuromorphicProcessorConfig {
	return NeuromorphicProcessorConfig{
		MeshRows:      8,
		MeshCols:      16,
		CoreConfig:    DefaultNeuroCoreConfig(),
		ProcessNode:   4,
		TotalNeurons:  1000000,
		TotalSynapses: 120000000,
	}
}

// NewNeuromorphicProcessor creates a complete neuromorphic processor
func NewNeuromorphicProcessor(config NeuromorphicProcessorConfig) *NeuromorphicProcessor {
	return &NeuromorphicProcessor{
		Config:     config,
		Mesh:       NewMeshRouter(config.MeshRows, config.MeshCols, config.CoreConfig),
		GlobalTime: 0,
		PowerUw:    0,
	}
}

// RunTimestep executes one global timestep with barrier synchronization
func (np *NeuromorphicProcessor) RunTimestep() int {
	np.mu.Lock()
	defer np.mu.Unlock()

	var allSpikes []GradedSpike
	var wg sync.WaitGroup

	// Process all cores in parallel
	spikeChannels := make([]chan []GradedSpike, np.Config.MeshRows*np.Config.MeshCols)
	idx := 0
	for i := 0; i < np.Config.MeshRows; i++ {
		for j := 0; j < np.Config.MeshCols; j++ {
			spikeChannels[idx] = make(chan []GradedSpike, 1)
			wg.Add(1)
			go func(core *NeuroCore, ch chan []GradedSpike) {
				defer wg.Done()
				spikes := core.ProcessTimestep()
				ch <- spikes
			}(np.Mesh.Cores[i][j], spikeChannels[idx])
			idx++
		}
	}

	wg.Wait()

	// Collect all spikes
	for _, ch := range spikeChannels {
		spikes := <-ch
		allSpikes = append(allSpikes, spikes...)
	}

	// Route spikes to destinations
	for _, spike := range allSpikes {
		srcRow := spike.SourceCore / np.Config.MeshCols
		srcCol := spike.SourceCore % np.Config.MeshCols
		np.Mesh.RouteSpike(spike, srcRow, srcCol)
	}

	np.SpikeCounter += len(allSpikes)
	np.GlobalTime++

	// Estimate power (simplified model)
	basePower := 100.0                   // Base power in uW per core
	spikePower := float64(len(allSpikes)) * 0.5 // 0.5 uW per spike
	np.PowerUw = float64(np.Config.MeshRows*np.Config.MeshCols)*basePower + spikePower

	return len(allSpikes)
}

// =============================================================================
// Part 2: Transformer CIM Attention Acceleration
// =============================================================================

// GainCellConfig configures gain cell memory for KV cache
type GainCellConfig struct {
	Rows           int     // Number of rows
	Cols           int     // Number of columns
	BitWidth       int     // Bits per cell
	RetentionMs    float64 // Retention time in milliseconds
	WriteEnergy_fJ float64 // Write energy in femtojoules
	ReadEnergy_fJ  float64 // Read energy in femtojoules
	RefreshPeriod  int     // Refresh period in cycles
}

// DefaultGainCellConfig returns typical gain cell configuration
func DefaultGainCellConfig() GainCellConfig {
	return GainCellConfig{
		Rows:           256,
		Cols:           256,
		BitWidth:       8,
		RetentionMs:    100,
		WriteEnergy_fJ: 1.5,
		ReadEnergy_fJ:  0.3,
		RefreshPeriod:  1000,
	}
}

// GainCellArray represents volatile gain cell memory for KV cache
type GainCellArray struct {
	Config       GainCellConfig
	Data         [][]float64
	Timestamps   [][]int // When each cell was last written
	CurrentCycle int
	TotalWrites  int
	TotalReads   int
	EnergyUsed   float64 // Total energy in fJ
	mu           sync.RWMutex
}

// NewGainCellArray creates a gain cell array
func NewGainCellArray(config GainCellConfig) *GainCellArray {
	data := make([][]float64, config.Rows)
	timestamps := make([][]int, config.Rows)
	for i := range data {
		data[i] = make([]float64, config.Cols)
		timestamps[i] = make([]int, config.Cols)
	}
	return &GainCellArray{
		Config:     config,
		Data:       data,
		Timestamps: timestamps,
	}
}

// Write stores a value with timestamp tracking
func (gc *GainCellArray) Write(row, col int, value float64) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if row < gc.Config.Rows && col < gc.Config.Cols {
		gc.Data[row][col] = value
		gc.Timestamps[row][col] = gc.CurrentCycle
		gc.TotalWrites++
		gc.EnergyUsed += gc.Config.WriteEnergy_fJ
	}
}

// Read retrieves a value, accounting for retention decay
func (gc *GainCellArray) Read(row, col int) float64 {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	if row >= gc.Config.Rows || col >= gc.Config.Cols {
		return 0
	}

	// Model retention decay
	age := gc.CurrentCycle - gc.Timestamps[row][col]
	decayFactor := math.Exp(-float64(age) / (gc.Config.RetentionMs * 1000))
	gc.TotalReads++

	return gc.Data[row][col] * decayFactor
}

// MatVecMul performs analog matrix-vector multiplication
func (gc *GainCellArray) MatVecMul(vector []float64) []float64 {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	result := make([]float64, gc.Config.Rows)
	for i := 0; i < gc.Config.Rows; i++ {
		for j := 0; j < gc.Config.Cols && j < len(vector); j++ {
			result[i] += gc.Data[i][j] * vector[j]
		}
		gc.EnergyUsed += gc.Config.ReadEnergy_fJ * float64(len(vector))
	}
	return result
}

// NeedRefresh checks if the array needs refresh based on cycle count
func (gc *GainCellArray) NeedRefresh() bool {
	return gc.CurrentCycle%gc.Config.RefreshPeriod == 0
}

// AdvanceCycle increments the cycle counter
func (gc *GainCellArray) AdvanceCycle() {
	gc.mu.Lock()
	gc.CurrentCycle++
	gc.mu.Unlock()
}

// KVCacheConfig configures the KV cache system
type KVCacheConfig struct {
	NumHeads      int            // Number of attention heads
	HeadDim       int            // Dimension per head
	MaxSeqLen     int            // Maximum sequence length
	GainCellCfg   GainCellConfig // Gain cell configuration
	UseCompression bool          // Enable KV compression
	CompressionRatio float64     // Compression ratio if enabled
}

// DefaultKVCacheConfig returns typical KV cache configuration
func DefaultKVCacheConfig() KVCacheConfig {
	return KVCacheConfig{
		NumHeads:         12,
		HeadDim:          64,
		MaxSeqLen:        2048,
		GainCellCfg:      DefaultGainCellConfig(),
		UseCompression:   true,
		CompressionRatio: 0.5,
	}
}

// CIMKVCache implements compute-in-memory KV cache
type CIMKVCache struct {
	Config    KVCacheConfig
	KeyCache  []*GainCellArray // One array per head
	ValueCache []*GainCellArray
	SeqLen    int
	mu        sync.RWMutex
}

// NewCIMKVCache creates a CIM-based KV cache
func NewCIMKVCache(config KVCacheConfig) *CIMKVCache {
	keyCache := make([]*GainCellArray, config.NumHeads)
	valueCache := make([]*GainCellArray, config.NumHeads)

	for i := 0; i < config.NumHeads; i++ {
		gcConfig := config.GainCellCfg
		gcConfig.Rows = config.MaxSeqLen
		gcConfig.Cols = config.HeadDim
		keyCache[i] = NewGainCellArray(gcConfig)
		valueCache[i] = NewGainCellArray(gcConfig)
	}

	return &CIMKVCache{
		Config:     config,
		KeyCache:   keyCache,
		ValueCache: valueCache,
		SeqLen:     0,
	}
}

// AppendKV adds new key-value pairs to the cache
func (kv *CIMKVCache) AppendKV(headIdx int, keys, values []float64) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if headIdx >= kv.Config.NumHeads || kv.SeqLen >= kv.Config.MaxSeqLen {
		return
	}

	for j := 0; j < kv.Config.HeadDim && j < len(keys); j++ {
		kv.KeyCache[headIdx].Write(kv.SeqLen, j, keys[j])
		kv.ValueCache[headIdx].Write(kv.SeqLen, j, values[j])
	}

	if headIdx == kv.Config.NumHeads-1 {
		kv.SeqLen++
	}
}

// ComputeAttention performs attention using CIM operations
func (kv *CIMKVCache) ComputeAttention(headIdx int, query []float64) []float64 {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	if headIdx >= kv.Config.NumHeads {
		return nil
	}

	// Compute attention scores (Q @ K^T) using analog MAC
	scores := make([]float64, kv.SeqLen)
	scale := 1.0 / math.Sqrt(float64(kv.Config.HeadDim))

	for i := 0; i < kv.SeqLen; i++ {
		for j := 0; j < kv.Config.HeadDim && j < len(query); j++ {
			k := kv.KeyCache[headIdx].Read(i, j)
			scores[i] += query[j] * k
		}
		scores[i] *= scale
	}

	// Softmax
	maxScore := scores[0]
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
	}
	var sumExp float64
	for i := range scores {
		scores[i] = math.Exp(scores[i] - maxScore)
		sumExp += scores[i]
	}
	for i := range scores {
		scores[i] /= sumExp
	}

	// Compute output (attention @ V) using analog MAC
	output := make([]float64, kv.Config.HeadDim)
	for i := 0; i < kv.SeqLen; i++ {
		for j := 0; j < kv.Config.HeadDim; j++ {
			v := kv.ValueCache[headIdx].Read(i, j)
			output[j] += scores[i] * v
		}
	}

	return output
}

// GetEnergyUsage returns total energy consumption
func (kv *CIMKVCache) GetEnergyUsage() float64 {
	var total float64
	for i := 0; i < kv.Config.NumHeads; i++ {
		total += kv.KeyCache[i].EnergyUsed
		total += kv.ValueCache[i].EnergyUsed
	}
	return total
}

// =============================================================================
// Part 3: Hybrid Analog-Digital Token Pruning
// =============================================================================

// TokenPrunerConfig configures the hybrid pruning accelerator
type TokenPrunerConfig struct {
	PruneRatio     float64 // Target pruning ratio (e.g., 0.75)
	AnalogBits     int     // Analog comparator precision
	ScoreThreshold float64 // Attention score threshold for pruning
	EnergyPerToken float64 // Energy per token in analog domain (fJ)
}

// DefaultTokenPrunerConfig returns default pruning configuration
func DefaultTokenPrunerConfig() TokenPrunerConfig {
	return TokenPrunerConfig{
		PruneRatio:     0.75,
		AnalogBits:     4,
		ScoreThreshold: 0.01,
		EnergyPerToken: 0.1,
	}
}

// HybridTokenPruner implements analog-digital hybrid pruning
type HybridTokenPruner struct {
	Config         TokenPrunerConfig
	AnalogScores   []float64
	PruneMask      []bool
	PrunedTokens   int
	TotalTokens    int
	AnalogEnergy   float64
	DigitalEnergy  float64
}

// NewHybridTokenPruner creates a hybrid pruner
func NewHybridTokenPruner(config TokenPrunerConfig) *HybridTokenPruner {
	return &HybridTokenPruner{
		Config: config,
	}
}

// AnalogScoring performs low-precision analog scoring
func (hp *HybridTokenPruner) AnalogScoring(attentionScores []float64) []bool {
	hp.TotalTokens = len(attentionScores)
	hp.AnalogScores = make([]float64, len(attentionScores))
	hp.PruneMask = make([]bool, len(attentionScores))

	// Quantize to analog precision
	maxVal := 0.0
	for _, s := range attentionScores {
		if s > maxVal {
			maxVal = s
		}
	}

	levels := float64(1 << hp.Config.AnalogBits)
	for i, s := range attentionScores {
		// Quantize score
		normalized := s / maxVal
		quantized := math.Round(normalized*levels) / levels
		hp.AnalogScores[i] = quantized

		// Binary pruning decision via analog comparator
		hp.PruneMask[i] = quantized < hp.Config.ScoreThreshold/maxVal
		hp.AnalogEnergy += hp.Config.EnergyPerToken
	}

	// Count pruned
	hp.PrunedTokens = 0
	for _, pruned := range hp.PruneMask {
		if pruned {
			hp.PrunedTokens++
		}
	}

	return hp.PruneMask
}

// GetUnprunedIndices returns indices of tokens to process digitally
func (hp *HybridTokenPruner) GetUnprunedIndices() []int {
	var indices []int
	for i, pruned := range hp.PruneMask {
		if !pruned {
			indices = append(indices, i)
		}
	}
	return indices
}

// GetPruningEfficiency returns the achieved pruning ratio
func (hp *HybridTokenPruner) GetPruningEfficiency() float64 {
	if hp.TotalTokens == 0 {
		return 0
	}
	return float64(hp.PrunedTokens) / float64(hp.TotalTokens)
}

// =============================================================================
// Part 4: SRAM-Based CIM for Softmax (UCLM Architecture)
// =============================================================================

// UCLMConfig configures Unified Compute and Lookup Module
type UCLMConfig struct {
	ArrayRows     int     // SRAM array rows
	ArrayCols     int     // SRAM array columns
	LUTEntries    int     // Lookup table entries for exp()
	BitWidth      int     // Bit width per element
	VDD           float64 // Supply voltage
	FrequencyMHz  float64 // Operating frequency
}

// DefaultUCLMConfig returns typical UCLM configuration
func DefaultUCLMConfig() UCLMConfig {
	return UCLMConfig{
		ArrayRows:    64,
		ArrayCols:    64,
		LUTEntries:   256,
		BitWidth:     8,
		VDD:          0.9,
		FrequencyMHz: 500,
	}
}

// UCLMUnit represents a unified compute and lookup module
type UCLMUnit struct {
	Config     UCLMConfig
	SRAMArray  [][]float64
	ExpLUT     []float64 // Precomputed exp() lookup table
	Operations int
	mu         sync.RWMutex
}

// NewUCLMUnit creates a UCLM unit with precomputed LUT
func NewUCLMUnit(config UCLMConfig) *UCLMUnit {
	// Initialize SRAM array
	sram := make([][]float64, config.ArrayRows)
	for i := range sram {
		sram[i] = make([]float64, config.ArrayCols)
	}

	// Precompute exp() lookup table
	expLUT := make([]float64, config.LUTEntries)
	for i := 0; i < config.LUTEntries; i++ {
		// Map index to input range [-8, 0] for softmax
		x := -8.0 + 8.0*float64(i)/float64(config.LUTEntries-1)
		expLUT[i] = math.Exp(x)
	}

	return &UCLMUnit{
		Config:    config,
		SRAMArray: sram,
		ExpLUT:    expLUT,
	}
}

// LookupExp performs exp() via LUT interpolation
func (u *UCLMUnit) LookupExp(x float64) float64 {
	// Clamp to LUT range
	x = math.Max(-8.0, math.Min(0.0, x))

	// Map to LUT index
	idx := (x + 8.0) / 8.0 * float64(u.Config.LUTEntries-1)
	idxLow := int(idx)
	idxHigh := idxLow + 1
	if idxHigh >= u.Config.LUTEntries {
		idxHigh = u.Config.LUTEntries - 1
	}

	// Linear interpolation
	frac := idx - float64(idxLow)
	return u.ExpLUT[idxLow]*(1-frac) + u.ExpLUT[idxHigh]*frac
}

// ComputeSoftmax performs softmax using CIM operations
func (u *UCLMUnit) ComputeSoftmax(logits []float64) []float64 {
	u.mu.Lock()
	defer u.mu.Unlock()

	n := len(logits)
	result := make([]float64, n)

	// Find max for numerical stability
	maxVal := logits[0]
	for _, v := range logits {
		if v > maxVal {
			maxVal = v
		}
	}

	// Compute exp via LUT
	var sumExp float64
	for i, v := range logits {
		shifted := v - maxVal
		result[i] = u.LookupExp(shifted)
		sumExp += result[i]
		u.Operations++
	}

	// Normalize
	for i := range result {
		result[i] /= sumExp
	}

	return result
}

// ConcurrentMAC performs matrix-accumulate in CIM style
func (u *UCLMUnit) ConcurrentMAC(weights [][]float64, input []float64) []float64 {
	u.mu.Lock()
	defer u.mu.Unlock()

	rows := len(weights)
	if rows == 0 {
		return nil
	}
	cols := len(weights[0])

	// Load weights to SRAM (simulated)
	for i := 0; i < rows && i < u.Config.ArrayRows; i++ {
		for j := 0; j < cols && j < u.Config.ArrayCols; j++ {
			u.SRAMArray[i][j] = weights[i][j]
		}
	}

	// Perform parallel MAC
	output := make([]float64, rows)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols && j < len(input); j++ {
			output[i] += u.SRAMArray[i][j] * input[j]
			u.Operations++
		}
	}

	return output
}

// =============================================================================
// Part 5: Full Transformer CIM Accelerator
// =============================================================================

// TransformerCIMConfig configures the complete transformer accelerator
type TransformerCIMConfig struct {
	NumLayers      int              // Number of transformer layers
	NumHeads       int              // Number of attention heads
	HiddenDim      int              // Hidden dimension
	FFNDim         int              // Feed-forward dimension
	MaxSeqLen      int              // Maximum sequence length
	KVCacheConfig  KVCacheConfig    // KV cache configuration
	PrunerConfig   TokenPrunerConfig // Pruning configuration
	UCLMConfig     UCLMConfig       // UCLM configuration
	UsePruning     bool             // Enable hybrid pruning
}

// DefaultTransformerCIMConfig returns default configuration
func DefaultTransformerCIMConfig() TransformerCIMConfig {
	return TransformerCIMConfig{
		NumLayers:     12,
		NumHeads:      12,
		HiddenDim:     768,
		FFNDim:        3072,
		MaxSeqLen:     2048,
		KVCacheConfig: DefaultKVCacheConfig(),
		PrunerConfig:  DefaultTokenPrunerConfig(),
		UCLMConfig:    DefaultUCLMConfig(),
		UsePruning:    true,
	}
}

// TransformerCIMAccelerator implements full transformer acceleration
type TransformerCIMAccelerator struct {
	Config           TransformerCIMConfig
	KVCaches         []*CIMKVCache // Per-layer KV cache
	Pruners          []*HybridTokenPruner
	SoftmaxUnits     []*UCLMUnit
	QKVWeights       [][][]float64 // [layer][head][dim*3]
	FFNWeights1      [][][]float64 // [layer][hidden][ffn]
	FFNWeights2      [][][]float64 // [layer][ffn][hidden]
	TotalEnergy_fJ   float64
	TotalLatency_ns  float64
	TokensProcessed  int
	mu               sync.RWMutex
}

// NewTransformerCIMAccelerator creates a transformer accelerator
func NewTransformerCIMAccelerator(config TransformerCIMConfig) *TransformerCIMAccelerator {
	kvCaches := make([]*CIMKVCache, config.NumLayers)
	pruners := make([]*HybridTokenPruner, config.NumLayers)
	softmaxUnits := make([]*UCLMUnit, config.NumLayers)

	for i := 0; i < config.NumLayers; i++ {
		kvConfig := config.KVCacheConfig
		kvConfig.NumHeads = config.NumHeads
		kvConfig.HeadDim = config.HiddenDim / config.NumHeads
		kvConfig.MaxSeqLen = config.MaxSeqLen
		kvCaches[i] = NewCIMKVCache(kvConfig)
		pruners[i] = NewHybridTokenPruner(config.PrunerConfig)
		softmaxUnits[i] = NewUCLMUnit(config.UCLMConfig)
	}

	// Initialize weights (random for simulation)
	headDim := config.HiddenDim / config.NumHeads
	qkvWeights := make([][][]float64, config.NumLayers)
	ffn1 := make([][][]float64, config.NumLayers)
	ffn2 := make([][][]float64, config.NumLayers)

	rng := rand.New(rand.NewSource(42))
	for l := 0; l < config.NumLayers; l++ {
		qkvWeights[l] = make([][]float64, config.NumHeads)
		for h := 0; h < config.NumHeads; h++ {
			qkvWeights[l][h] = make([]float64, headDim*3)
			for i := range qkvWeights[l][h] {
				qkvWeights[l][h][i] = rng.NormFloat64() * 0.02
			}
		}

		ffn1[l] = make([][]float64, config.HiddenDim)
		for i := range ffn1[l] {
			ffn1[l][i] = make([]float64, config.FFNDim)
			for j := range ffn1[l][i] {
				ffn1[l][i][j] = rng.NormFloat64() * 0.02
			}
		}

		ffn2[l] = make([][]float64, config.FFNDim)
		for i := range ffn2[l] {
			ffn2[l][i] = make([]float64, config.HiddenDim)
			for j := range ffn2[l][i] {
				ffn2[l][i][j] = rng.NormFloat64() * 0.02
			}
		}
	}

	return &TransformerCIMAccelerator{
		Config:       config,
		KVCaches:     kvCaches,
		Pruners:      pruners,
		SoftmaxUnits: softmaxUnits,
		QKVWeights:   qkvWeights,
		FFNWeights1:  ffn1,
		FFNWeights2:  ffn2,
	}
}

// ProcessToken processes a single token through all layers
func (t *TransformerCIMAccelerator) ProcessToken(embedding []float64) []float64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(embedding) != t.Config.HiddenDim {
		return nil
	}

	hidden := make([]float64, len(embedding))
	copy(hidden, embedding)

	headDim := t.Config.HiddenDim / t.Config.NumHeads

	for layer := 0; layer < t.Config.NumLayers; layer++ {
		// Multi-head attention
		attnOutput := make([]float64, t.Config.HiddenDim)

		for head := 0; head < t.Config.NumHeads; head++ {
			// Extract Q, K, V for this head
			startIdx := head * headDim
			query := hidden[startIdx : startIdx+headDim]

			// Compute key and value projections (simplified)
			key := make([]float64, headDim)
			value := make([]float64, headDim)
			for i := 0; i < headDim; i++ {
				key[i] = hidden[startIdx+i] * t.QKVWeights[layer][head][i]
				value[i] = hidden[startIdx+i] * t.QKVWeights[layer][head][headDim+i]
			}

			// Update KV cache
			t.KVCaches[layer].AppendKV(head, key, value)

			// Compute attention output using CIM
			headOutput := t.KVCaches[layer].ComputeAttention(head, query)
			if headOutput != nil {
				copy(attnOutput[startIdx:startIdx+headDim], headOutput)
			}
		}

		// Residual connection
		for i := range hidden {
			hidden[i] += attnOutput[i]
		}

		// Feed-forward network (using UCLM for matrix operations)
		ffnInter := t.SoftmaxUnits[layer].ConcurrentMAC(t.FFNWeights1[layer], hidden)

		// GELU activation
		for i := range ffnInter {
			x := ffnInter[i]
			ffnInter[i] = 0.5 * x * (1 + math.Tanh(math.Sqrt(2/math.Pi)*(x+0.044715*x*x*x)))
		}

		ffnOutput := t.SoftmaxUnits[layer].ConcurrentMAC(t.FFNWeights2[layer], ffnInter)

		// Residual connection
		for i := range hidden {
			if i < len(ffnOutput) {
				hidden[i] += ffnOutput[i]
			}
		}
	}

	t.TokensProcessed++

	// Accumulate energy estimates
	for _, kv := range t.KVCaches {
		t.TotalEnergy_fJ += kv.GetEnergyUsage()
	}

	return hidden
}

// GetMetrics returns accelerator performance metrics
func (t *TransformerCIMAccelerator) GetMetrics() map[string]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	metrics := map[string]float64{
		"tokens_processed":   float64(t.TokensProcessed),
		"total_energy_fJ":    t.TotalEnergy_fJ,
		"energy_per_token":   0,
		"avg_pruning_ratio":  0,
		"kv_cache_occupancy": 0,
	}

	if t.TokensProcessed > 0 {
		metrics["energy_per_token"] = t.TotalEnergy_fJ / float64(t.TokensProcessed)
	}

	// Calculate average pruning efficiency
	var totalPruning float64
	for _, p := range t.Pruners {
		totalPruning += p.GetPruningEfficiency()
	}
	metrics["avg_pruning_ratio"] = totalPruning / float64(len(t.Pruners))

	// Calculate KV cache occupancy
	var totalOccupancy float64
	for _, kv := range t.KVCaches {
		totalOccupancy += float64(kv.SeqLen) / float64(kv.Config.MaxSeqLen)
	}
	metrics["kv_cache_occupancy"] = totalOccupancy / float64(len(t.KVCaches))

	return metrics
}

// =============================================================================
// Part 6: Benchmark and Integration
// =============================================================================

// NeuromorphicTransformerBenchmark benchmarks both architectures
type NeuromorphicTransformerBenchmark struct {
	NeuromorphicProc *NeuromorphicProcessor
	TransformerAccel *TransformerCIMAccelerator
	Results          map[string]float64
}

// NewNeuromorphicTransformerBenchmark creates a benchmark suite
func NewNeuromorphicTransformerBenchmark() *NeuromorphicTransformerBenchmark {
	return &NeuromorphicTransformerBenchmark{
		NeuromorphicProc: NewNeuromorphicProcessor(DefaultNeuromorphicProcessorConfig()),
		TransformerAccel: NewTransformerCIMAccelerator(DefaultTransformerCIMConfig()),
		Results:          make(map[string]float64),
	}
}

// RunNeuromorphicBenchmark benchmarks the neuromorphic processor
func (b *NeuromorphicTransformerBenchmark) RunNeuromorphicBenchmark(timesteps int) {
	// Initialize with random activity
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < b.NeuromorphicProc.Config.MeshRows; i++ {
		for j := 0; j < b.NeuromorphicProc.Config.MeshCols; j++ {
			core := b.NeuromorphicProc.Mesh.Cores[i][j]
			for k := 0; k < 100; k++ {
				comp := core.Compartments[k]
				comp.Voltage = rng.Float64() * 0.5
			}
			// Set random synaptic weights
			for m := 0; m < 100; m++ {
				for n := 0; n < 100; n++ {
					core.Synapses.Weights[m][n] = rng.Float64()*0.2 - 0.1
				}
			}
		}
	}

	// Run simulation
	totalSpikes := 0
	for t := 0; t < timesteps; t++ {
		spikes := b.NeuromorphicProc.RunTimestep()
		totalSpikes += spikes
	}

	b.Results["neuromorphic_timesteps"] = float64(timesteps)
	b.Results["neuromorphic_total_spikes"] = float64(totalSpikes)
	b.Results["neuromorphic_spikes_per_step"] = float64(totalSpikes) / float64(timesteps)
	b.Results["neuromorphic_power_uW"] = b.NeuromorphicProc.PowerUw
}

// RunTransformerBenchmark benchmarks the transformer accelerator
func (b *NeuromorphicTransformerBenchmark) RunTransformerBenchmark(numTokens int) {
	rng := rand.New(rand.NewSource(42))
	hiddenDim := b.TransformerAccel.Config.HiddenDim

	for i := 0; i < numTokens; i++ {
		// Generate random embedding
		embedding := make([]float64, hiddenDim)
		for j := range embedding {
			embedding[j] = rng.NormFloat64() * 0.1
		}

		b.TransformerAccel.ProcessToken(embedding)
	}

	metrics := b.TransformerAccel.GetMetrics()
	for k, v := range metrics {
		b.Results["transformer_"+k] = v
	}
}

// GetResults returns benchmark results
func (b *NeuromorphicTransformerBenchmark) GetResults() map[string]float64 {
	return b.Results
}

// PrintResults prints benchmark results
func (b *NeuromorphicTransformerBenchmark) PrintResults() string {
	var result string
	result += "=== Neuromorphic-Transformer Benchmark Results ===\n\n"

	result += "Neuromorphic Processor:\n"
	result += fmt.Sprintf("  Timesteps: %.0f\n", b.Results["neuromorphic_timesteps"])
	result += fmt.Sprintf("  Total spikes: %.0f\n", b.Results["neuromorphic_total_spikes"])
	result += fmt.Sprintf("  Spikes/step: %.2f\n", b.Results["neuromorphic_spikes_per_step"])
	result += fmt.Sprintf("  Power: %.2f µW\n", b.Results["neuromorphic_power_uW"])

	result += "\nTransformer CIM Accelerator:\n"
	result += fmt.Sprintf("  Tokens processed: %.0f\n", b.Results["transformer_tokens_processed"])
	result += fmt.Sprintf("  Total energy: %.2f fJ\n", b.Results["transformer_total_energy_fJ"])
	result += fmt.Sprintf("  Energy/token: %.2f fJ\n", b.Results["transformer_energy_per_token"])
	result += fmt.Sprintf("  KV cache occupancy: %.2f%%\n", b.Results["transformer_kv_cache_occupancy"]*100)

	return result
}

// =============================================================================
// Part 7: Advanced Spike-Based Transformer (Neuromorphic + Transformer Fusion)
// =============================================================================

// SpikeTransformerConfig configures spike-based transformer
type SpikeTransformerConfig struct {
	NumLayers     int     // Number of transformer layers
	NumHeads      int     // Number of attention heads
	HiddenDim     int     // Hidden dimension
	SpikeEncoding string  // "rate", "temporal", "phase"
	TimestepsPerToken int // Spike timesteps per token
	ThresholdVoltage float64 // Neuron threshold
}

// DefaultSpikeTransformerConfig returns default configuration
func DefaultSpikeTransformerConfig() SpikeTransformerConfig {
	return SpikeTransformerConfig{
		NumLayers:        6,
		NumHeads:         8,
		HiddenDim:        512,
		SpikeEncoding:    "rate",
		TimestepsPerToken: 16,
		ThresholdVoltage: 1.0,
	}
}

// SpikeTransformerLayer implements spike-based attention
type SpikeTransformerLayer struct {
	Config       SpikeTransformerConfig
	QueryNeurons []*LIFNeuronSimple
	KeyNeurons   []*LIFNeuronSimple
	ValueNeurons []*LIFNeuronSimple
	OutputNeurons []*LIFNeuronSimple
	SpikeCount   int
}

// LIFNeuronSimple is a simplified LIF neuron for spike transformer
type LIFNeuronSimple struct {
	Voltage      float64
	Threshold    float64
	TauMembrane  float64
	ResetVoltage float64
	Spiked       bool
}

// NewLIFNeuronSimple creates a simple LIF neuron
func NewLIFNeuronSimple(threshold float64) *LIFNeuronSimple {
	return &LIFNeuronSimple{
		Voltage:      0,
		Threshold:    threshold,
		TauMembrane:  10.0,
		ResetVoltage: 0,
	}
}

// Step advances the neuron one timestep
func (n *LIFNeuronSimple) Step(input float64, dt float64) bool {
	// Leak
	decay := math.Exp(-dt / n.TauMembrane)
	n.Voltage = n.Voltage*decay + input*(1-decay)

	// Spike check
	n.Spiked = n.Voltage >= n.Threshold
	if n.Spiked {
		n.Voltage = n.ResetVoltage
	}

	return n.Spiked
}

// NewSpikeTransformerLayer creates a spike-based transformer layer
func NewSpikeTransformerLayer(config SpikeTransformerConfig) *SpikeTransformerLayer {
	headDim := config.HiddenDim / config.NumHeads

	queryNeurons := make([]*LIFNeuronSimple, config.HiddenDim)
	keyNeurons := make([]*LIFNeuronSimple, config.HiddenDim)
	valueNeurons := make([]*LIFNeuronSimple, config.HiddenDim)
	outputNeurons := make([]*LIFNeuronSimple, config.HiddenDim)

	for i := 0; i < config.HiddenDim; i++ {
		queryNeurons[i] = NewLIFNeuronSimple(config.ThresholdVoltage)
		keyNeurons[i] = NewLIFNeuronSimple(config.ThresholdVoltage)
		valueNeurons[i] = NewLIFNeuronSimple(config.ThresholdVoltage)
		outputNeurons[i] = NewLIFNeuronSimple(config.ThresholdVoltage)
	}

	_ = headDim // Silence unused variable

	return &SpikeTransformerLayer{
		Config:        config,
		QueryNeurons:  queryNeurons,
		KeyNeurons:    keyNeurons,
		ValueNeurons:  valueNeurons,
		OutputNeurons: outputNeurons,
	}
}

// EncodeToSpikes converts continuous values to spike trains
func (st *SpikeTransformerLayer) EncodeToSpikes(values []float64) [][]bool {
	timesteps := st.Config.TimestepsPerToken
	spikes := make([][]bool, len(values))

	for i, v := range values {
		spikes[i] = make([]bool, timesteps)
		switch st.Config.SpikeEncoding {
		case "rate":
			// Rate coding: probability proportional to value
			rate := math.Max(0, math.Min(1, (v+1)/2)) // Normalize to [0,1]
			for t := 0; t < timesteps; t++ {
				if rand.Float64() < rate {
					spikes[i][t] = true
				}
			}
		case "temporal":
			// Temporal coding: first spike time proportional to value
			spikeTime := int((1 - math.Max(0, math.Min(1, (v+1)/2))) * float64(timesteps-1))
			if spikeTime < timesteps {
				spikes[i][spikeTime] = true
			}
		case "phase":
			// Phase coding: spike phase relative to oscillation
			phase := (v + 1) / 2 * 2 * math.Pi
			for t := 0; t < timesteps; t++ {
				oscPhase := 2 * math.Pi * float64(t) / float64(timesteps)
				if math.Abs(oscPhase-phase) < 0.2 {
					spikes[i][t] = true
				}
			}
		}
	}

	return spikes
}

// DecodeFromSpikes converts spike trains back to continuous values
func (st *SpikeTransformerLayer) DecodeFromSpikes(spikes [][]bool) []float64 {
	values := make([]float64, len(spikes))
	timesteps := st.Config.TimestepsPerToken

	for i, spikesTrain := range spikes {
		spikeCount := 0
		for _, s := range spikesTrain {
			if s {
				spikeCount++
			}
		}
		// Decode based on spike rate
		values[i] = float64(spikeCount)/float64(timesteps)*2 - 1 // Map back to [-1,1]
	}

	return values
}

// ProcessSpikes runs spike-based attention for one timestep
func (st *SpikeTransformerLayer) ProcessSpikes(inputSpikes []bool, t int, weights [][]float64) []bool {
	dt := 1.0
	outputSpikes := make([]bool, len(st.OutputNeurons))

	// Process Q, K, V projections through neurons
	for i := 0; i < len(st.QueryNeurons); i++ {
		var qInput, kInput, vInput float64
		for j := 0; j < len(inputSpikes); j++ {
			if inputSpikes[j] {
				if i < len(weights) && j < len(weights[i]) {
					qInput += weights[i][j]
					kInput += weights[i][j]
					vInput += weights[i][j]
				}
			}
		}

		if st.QueryNeurons[i].Step(qInput, dt) {
			st.SpikeCount++
		}
		if st.KeyNeurons[i].Step(kInput, dt) {
			st.SpikeCount++
		}
		if st.ValueNeurons[i].Step(vInput, dt) {
			st.SpikeCount++
		}
	}

	// Compute attention output (simplified spike-based attention)
	for i := range st.OutputNeurons {
		attnInput := 0.0
		for j := range st.ValueNeurons {
			if st.ValueNeurons[j].Spiked && st.QueryNeurons[i].Spiked {
				attnInput += 1.0
			}
		}
		outputSpikes[i] = st.OutputNeurons[i].Step(attnInput, dt)
		if outputSpikes[i] {
			st.SpikeCount++
		}
	}

	return outputSpikes
}

// =============================================================================
// Part 8: Hardware Cost Model
// =============================================================================

// HardwareCostModelConfig configures cost estimation
type HardwareCostModelConfig struct {
	ProcessNode_nm      int     // Process node (e.g., 7, 14, 28)
	MemoryType          string  // "SRAM", "ReRAM", "GainCell", "eDRAM"
	OperatingFreqMHz    float64 // Operating frequency
	SupplyVoltage       float64 // Supply voltage (V)
	LeakagePower_mW     float64 // Static leakage power
}

// HardwareCostModel estimates energy and area costs
type HardwareCostModel struct {
	Config                HardwareCostModelConfig
	MACEnergy_fJ          float64 // Energy per MAC operation
	MemoryAccessEnergy_fJ float64 // Energy per memory access
	ADCEnergy_fJ          float64 // Energy per ADC conversion
	DACEnergy_fJ          float64 // Energy per DAC conversion
	SynapseArea_um2       float64 // Area per synapse
	NeuronArea_um2        float64 // Area per neuron
}

// NewHardwareCostModel creates a cost model
func NewHardwareCostModel(config HardwareCostModelConfig) *HardwareCostModel {
	// Scale costs based on process node
	scaleFactor := float64(config.ProcessNode_nm) / 7.0 // Normalize to 7nm

	// Base energies at 7nm
	macEnergy := 0.5 * scaleFactor * config.SupplyVoltage * config.SupplyVoltage
	memEnergy := 2.0 * scaleFactor
	adcEnergy := 50.0 * scaleFactor // ADC is expensive
	dacEnergy := 20.0 * scaleFactor

	// Memory type adjustments
	switch config.MemoryType {
	case "ReRAM":
		memEnergy *= 0.1 // ReRAM is very efficient for reads
	case "GainCell":
		memEnergy *= 0.3 // Gain cells are efficient
	case "eDRAM":
		memEnergy *= 0.5 // eDRAM moderate efficiency
	}

	return &HardwareCostModel{
		Config:                config,
		MACEnergy_fJ:          macEnergy,
		MemoryAccessEnergy_fJ: memEnergy,
		ADCEnergy_fJ:          adcEnergy,
		DACEnergy_fJ:          dacEnergy,
		SynapseArea_um2:       0.04 * scaleFactor * scaleFactor, // 4F² baseline
		NeuronArea_um2:        10.0 * scaleFactor * scaleFactor,
	}
}

// EstimateInferenceEnergy estimates energy for inference
func (hc *HardwareCostModel) EstimateInferenceEnergy(macs int, memAccesses int, adcConversions int) float64 {
	return float64(macs)*hc.MACEnergy_fJ +
		float64(memAccesses)*hc.MemoryAccessEnergy_fJ +
		float64(adcConversions)*hc.ADCEnergy_fJ
}

// EstimateChipArea estimates total chip area
func (hc *HardwareCostModel) EstimateChipArea(neurons int, synapses int) float64 {
	return float64(neurons)*hc.NeuronArea_um2 + float64(synapses)*hc.SynapseArea_um2
}

// EstimateTOPSW estimates TOPS/W efficiency
func (hc *HardwareCostModel) EstimateTOPSW(opsPerSecond float64, powerW float64) float64 {
	if powerW <= 0 {
		return 0
	}
	return opsPerSecond / 1e12 / powerW
}

// =============================================================================
// Part 9: Integration with Existing CIM Infrastructure
// =============================================================================

// CIMNeuromorphicBridge bridges CIM crossbar arrays with neuromorphic processors
type CIMNeuromorphicBridge struct {
	CrossbarSize    int
	NeuromorphicProc *NeuromorphicProcessor
	WeightMapping   map[int]int // Maps crossbar index to core/neuron
}

// NewCIMNeuromorphicBridge creates a bridge between CIM and neuromorphic
func NewCIMNeuromorphicBridge(crossbarSize int) *CIMNeuromorphicBridge {
	npConfig := DefaultNeuromorphicProcessorConfig()
	npConfig.MeshRows = 4
	npConfig.MeshCols = 4

	return &CIMNeuromorphicBridge{
		CrossbarSize:    crossbarSize,
		NeuromorphicProc: NewNeuromorphicProcessor(npConfig),
		WeightMapping:   make(map[int]int),
	}
}

// MapWeightsToCores distributes crossbar weights to neuromorphic cores
func (bridge *CIMNeuromorphicBridge) MapWeightsToCores(weights [][]float64) {
	rows := len(weights)
	cols := 0
	if rows > 0 {
		cols = len(weights[0])
	}

	neuronsPerCore := bridge.NeuromorphicProc.Config.CoreConfig.NumCompartments
	totalCores := bridge.NeuromorphicProc.Config.MeshRows * bridge.NeuromorphicProc.Config.MeshCols

	// Distribute weights across cores
	for i := 0; i < rows; i++ {
		coreIdx := i / neuronsPerCore
		if coreIdx >= totalCores {
			break
		}
		neuronIdx := i % neuronsPerCore

		coreRow := coreIdx / bridge.NeuromorphicProc.Config.MeshCols
		coreCol := coreIdx % bridge.NeuromorphicProc.Config.MeshCols
		core := bridge.NeuromorphicProc.Mesh.Cores[coreRow][coreCol]

		// Set synaptic weights
		for j := 0; j < cols && j < neuronsPerCore; j++ {
			core.Synapses.Weights[neuronIdx][j] = weights[i][j]
		}

		bridge.WeightMapping[i] = coreIdx*neuronsPerCore + neuronIdx
	}
}

// ComputeMVM performs matrix-vector multiply using spiking neurons
func (bridge *CIMNeuromorphicBridge) ComputeMVM(input []float64, timesteps int) []float64 {
	// Encode input as spike rates
	inputSpikes := make([][]GradedSpike, len(input))
	for i, v := range input {
		// Convert to graded spikes
		rate := (v + 1) / 2 // Normalize to [0,1]
		numSpikes := int(rate * float64(timesteps))
		inputSpikes[i] = make([]GradedSpike, numSpikes)
		for j := 0; j < numSpikes; j++ {
			inputSpikes[i][j] = GradedSpike{
				SourceNeuron: i,
				Payload:      int32(v * 1000),
				Timestamp:    rand.Intn(timesteps),
			}
		}
	}

	// Inject spikes into first core
	for i, spikes := range inputSpikes {
		for _, spike := range spikes {
			spike.DestCore = 0
			spike.DestNeuron = i % bridge.NeuromorphicProc.Config.CoreConfig.NumCompartments
			bridge.NeuromorphicProc.Mesh.Cores[0][0].InputQueue <- spike
		}
	}

	// Run simulation
	for t := 0; t < timesteps; t++ {
		bridge.NeuromorphicProc.RunTimestep()
	}

	// Decode output from spike counts
	outputSize := len(input)
	output := make([]float64, outputSize)
	for i := 0; i < outputSize; i++ {
		coreIdx := i / bridge.NeuromorphicProc.Config.CoreConfig.NumCompartments
		neuronIdx := i % bridge.NeuromorphicProc.Config.CoreConfig.NumCompartments
		if coreIdx < bridge.NeuromorphicProc.Config.MeshRows*bridge.NeuromorphicProc.Config.MeshCols {
			coreRow := coreIdx / bridge.NeuromorphicProc.Config.MeshCols
			coreCol := coreIdx % bridge.NeuromorphicProc.Config.MeshCols
			if coreRow < len(bridge.NeuromorphicProc.Mesh.Cores) && coreCol < len(bridge.NeuromorphicProc.Mesh.Cores[coreRow]) {
				comp := bridge.NeuromorphicProc.Mesh.Cores[coreRow][coreCol].Compartments[neuronIdx]
				output[i] = float64(comp.SpikeCount) / float64(timesteps) * 2 - 1 // Decode rate
			}
		}
	}

	return output
}

// =============================================================================
// Part 10: Serialization and Export
// =============================================================================

// NeuromorphicModelExport exports neuromorphic model for hardware deployment
type NeuromorphicModelExport struct {
	Version         string
	NumCores        int
	NeuronsPerCore  int
	SynapseWeights  [][]float64
	NeuronMicrocode [][]MicrocodeInstruction
	TopologyType    string // "mesh", "hierarchical", "custom"
}

// ExportForDeployment creates an export structure
func (np *NeuromorphicProcessor) ExportForDeployment() *NeuromorphicModelExport {
	totalCores := np.Config.MeshRows * np.Config.MeshCols
	neuronsPerCore := np.Config.CoreConfig.NumCompartments

	export := &NeuromorphicModelExport{
		Version:        "1.0",
		NumCores:       totalCores,
		NeuronsPerCore: neuronsPerCore,
		TopologyType:   "mesh",
	}

	// Collect all synaptic weights
	export.SynapseWeights = make([][]float64, 0)
	for i := 0; i < np.Config.MeshRows; i++ {
		for j := 0; j < np.Config.MeshCols; j++ {
			core := np.Mesh.Cores[i][j]
			for m := 0; m < neuronsPerCore; m++ {
				row := make([]float64, neuronsPerCore)
				copy(row, core.Synapses.Weights[m])
				export.SynapseWeights = append(export.SynapseWeights, row)
			}
		}
	}

	return export
}

// SerializeToBytes serializes export to binary format
func (export *NeuromorphicModelExport) SerializeToBytes() []byte {
	// Simple binary serialization
	buf := make([]byte, 0, 1024*1024)

	// Header
	buf = append(buf, []byte(export.Version)...)
	buf = append(buf, 0) // null terminator

	// Metadata
	numCoresBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(numCoresBytes, uint32(export.NumCores))
	buf = append(buf, numCoresBytes...)

	neuronsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(neuronsBytes, uint32(export.NeuronsPerCore))
	buf = append(buf, neuronsBytes...)

	// Weights
	for _, row := range export.SynapseWeights {
		for _, w := range row {
			weightBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(weightBytes, math.Float32bits(float32(w)))
			buf = append(buf, weightBytes...)
		}
	}

	return buf
}

// TransformerModelExport exports transformer CIM model
type TransformerModelExport struct {
	Version      string
	NumLayers    int
	NumHeads     int
	HiddenDim    int
	QKVWeights   [][][]float64
	FFNWeights1  [][][]float64
	FFNWeights2  [][][]float64
	Quantization int // Bits for quantization
}

// ExportForDeployment creates transformer export
func (t *TransformerCIMAccelerator) ExportForDeployment(quantBits int) *TransformerModelExport {
	export := &TransformerModelExport{
		Version:      "1.0",
		NumLayers:    t.Config.NumLayers,
		NumHeads:     t.Config.NumHeads,
		HiddenDim:    t.Config.HiddenDim,
		QKVWeights:   t.QKVWeights,
		FFNWeights1:  t.FFNWeights1,
		FFNWeights2:  t.FFNWeights2,
		Quantization: quantBits,
	}

	// Quantize if requested
	if quantBits > 0 && quantBits < 32 {
		export.quantizeWeights(quantBits)
	}

	return export
}

func (export *TransformerModelExport) quantizeWeights(bits int) {
	levels := float64(1 << bits)

	// Find global min/max
	minVal, maxVal := math.MaxFloat64, -math.MaxFloat64
	for l := range export.QKVWeights {
		for h := range export.QKVWeights[l] {
			for _, w := range export.QKVWeights[l][h] {
				minVal = math.Min(minVal, w)
				maxVal = math.Max(maxVal, w)
			}
		}
	}

	scale := (maxVal - minVal) / (levels - 1)
	if scale == 0 {
		scale = 1
	}

	// Quantize
	for l := range export.QKVWeights {
		for h := range export.QKVWeights[l] {
			for i := range export.QKVWeights[l][h] {
				normalized := (export.QKVWeights[l][h][i] - minVal) / scale
				quantized := math.Round(normalized)
				export.QKVWeights[l][h][i] = quantized*scale + minVal
			}
		}
	}
}

// =============================================================================
// Part 11: Performance Comparison Utilities
// =============================================================================

// ArchitectureComparison compares different CIM architectures
type ArchitectureComparison struct {
	Architectures map[string]ArchitectureMetrics
}

// ArchitectureMetrics stores metrics for an architecture
type ArchitectureMetrics struct {
	Name            string
	EnergyEfficiency float64 // TOPS/W
	AreaEfficiency   float64 // GOPS/mm²
	Accuracy         float64 // Percentage
	Latency_us       float64 // Microseconds per inference
	ProcessNode      int     // nm
}

// NewArchitectureComparison creates a comparison suite
func NewArchitectureComparison() *ArchitectureComparison {
	return &ArchitectureComparison{
		Architectures: map[string]ArchitectureMetrics{
			"Loihi2": {
				Name:             "Intel Loihi 2",
				EnergyEfficiency: 100.0, // Estimated TOPS/W equivalent
				AreaEfficiency:   50.0,
				Accuracy:         95.0,
				Latency_us:       10.0,
				ProcessNode:      4,
			},
			"GainCell_AIMC": {
				Name:             "Gain Cell AIMC (Nature 2025)",
				EnergyEfficiency: 70000.0, // 70,000x improvement claimed
				AreaEfficiency:   1000.0,
				Accuracy:         92.0,
				Latency_us:       0.1,
				ProcessNode:      28,
			},
			"Hybrid_Analog_Digital": {
				Name:             "Hybrid A/D Attention (IEEE 2024)",
				EnergyEfficiency: 14.8,
				AreaEfficiency:   976.6,
				Accuracy:         94.0,
				Latency_us:       5.0,
				ProcessNode:      65,
			},
			"FeFET_Crossbar": {
				Name:             "Ferroelectric FeFET Crossbar",
				EnergyEfficiency: 77.64, // From federated learning paper
				AreaEfficiency:   200.0,
				Accuracy:         96.0,
				Latency_us:       1.0,
				ProcessNode:      28,
			},
		},
	}
}

// RankByMetric ranks architectures by a specific metric
func (ac *ArchitectureComparison) RankByMetric(metric string) []string {
	type archMetric struct {
		name  string
		value float64
	}

	var metrics []archMetric
	for name, arch := range ac.Architectures {
		var val float64
		switch metric {
		case "energy":
			val = arch.EnergyEfficiency
		case "area":
			val = arch.AreaEfficiency
		case "accuracy":
			val = arch.Accuracy
		case "latency":
			val = -arch.Latency_us // Negative for ascending sort
		}
		metrics = append(metrics, archMetric{name, val})
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].value > metrics[j].value
	})

	result := make([]string, len(metrics))
	for i, m := range metrics {
		result[i] = m.name
	}
	return result
}

// GenerateComparisonReport creates a formatted comparison report
func (ac *ArchitectureComparison) GenerateComparisonReport() string {
	var report string
	report += "=== CIM Architecture Comparison ===\n\n"

	// Energy efficiency ranking
	report += "Energy Efficiency (TOPS/W) Ranking:\n"
	for i, name := range ac.RankByMetric("energy") {
		arch := ac.Architectures[name]
		report += fmt.Sprintf("  %d. %s: %.2f TOPS/W\n", i+1, arch.Name, arch.EnergyEfficiency)
	}

	report += "\nArea Efficiency (GOPS/mm²) Ranking:\n"
	for i, name := range ac.RankByMetric("area") {
		arch := ac.Architectures[name]
		report += fmt.Sprintf("  %d. %s: %.2f GOPS/mm²\n", i+1, arch.Name, arch.AreaEfficiency)
	}

	report += "\nAccuracy Ranking:\n"
	for i, name := range ac.RankByMetric("accuracy") {
		arch := ac.Architectures[name]
		report += fmt.Sprintf("  %d. %s: %.2f%%\n", i+1, arch.Name, arch.Accuracy)
	}

	return report
}

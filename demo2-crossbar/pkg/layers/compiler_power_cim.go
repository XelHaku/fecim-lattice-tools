// compiler_power_cim.go - CIM Compiler/Mapping and Power Management Simulation
// IronLattice Visualization Project - Iteration 124
//
// This module implements simulation models for:
// 1. CIM compiler and neural network mapping strategies
// 2. Multi-level compilation (CIM-MLC style) with hardware abstraction
// 3. Dual-mode compilation (CMSwitch) for compute/memory mode switching
// 4. Crossbar tiling and operator mapping
// 5. Power management techniques (power gating, DVFS)
// 6. Hybrid memory hierarchies (SRAM + ReRAM + MRAM)
//
// Research basis:
// - CIM-MLC: Multi-level Compilation Stack (ASPLOS 2024)
// - CMSwitch: Dual-mode-aware DNN Compiler (ASPLOS 2025)
// - CIM-Explorer: BNN/TNN Inference Optimization (SAMOS 2025)
// - Nature 2025: Mixed-precision memristor + SRAM CIM
// - Hybrid memory architectures: 80% idle energy reduction

package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
)

// ============================================================================
// CIM HARDWARE ABSTRACTION (Abs-arch from CIM-MLC)
// ============================================================================

// AbsArchConfig defines the hardware abstraction hierarchy
type AbsArchConfig struct {
	// Chip level
	NumCores       int `json:"num_cores"`
	GlobalBufferKB int `json:"global_buffer_kb"`

	// Core level
	CrossbarsPerCore int `json:"crossbars_per_core"`
	LocalBufferKB    int `json:"local_buffer_kb"`

	// Crossbar level
	CrossbarRows     int `json:"crossbar_rows"`
	CrossbarCols     int `json:"crossbar_cols"`
	DevicePrecision  int `json:"device_precision"`  // Bits per cell
	ADCPrecision     int `json:"adc_precision"`
	DACPrecision     int `json:"dac_precision"`

	// Computing modes
	SupportedModes   []ComputeMode `json:"supported_modes"`
}

// ComputeMode represents CIM computing granularity
type ComputeMode int

const (
	ModeMemory     ComputeMode = iota // Pure memory (no compute)
	ModeMVM                           // Matrix-vector multiply
	ModeConv                          // Convolution (im2col + MVM)
	ModeAttention                     // Attention (Q*K^T, softmax, *V)
	ModeCustom                        // Custom operator
)

// HardwareSpec describes a specific CIM accelerator
type HardwareSpec struct {
	Name           string         `json:"name"`
	Config         *AbsArchConfig `json:"config"`

	// Energy model (pJ per operation)
	EnergyMVM      float64 `json:"energy_mvm"`      // Per MAC
	EnergyADC      float64 `json:"energy_adc"`      // Per conversion
	EnergyDAC      float64 `json:"energy_dac"`      // Per conversion
	EnergyBuffer   float64 `json:"energy_buffer"`   // Per byte access
	EnergyLeakage  float64 `json:"energy_leakage"`  // Per cycle idle

	// Timing model (cycles)
	LatencyMVM     int `json:"latency_mvm"`
	LatencyADC     int `json:"latency_adc"`
	LatencyDAC     int `json:"latency_dac"`
}

// NewDefaultHardwareSpec creates a typical CIM hardware spec
func NewDefaultHardwareSpec() *HardwareSpec {
	return &HardwareSpec{
		Name: "DefaultCIM",
		Config: &AbsArchConfig{
			NumCores:         4,
			GlobalBufferKB:   256,
			CrossbarsPerCore: 16,
			LocalBufferKB:    32,
			CrossbarRows:     256,
			CrossbarCols:     256,
			DevicePrecision:  4,
			ADCPrecision:     6,
			DACPrecision:     8,
			SupportedModes:   []ComputeMode{ModeMemory, ModeMVM, ModeConv},
		},
		EnergyMVM:     0.1,   // 0.1 pJ/MAC
		EnergyADC:     5.0,   // 5 pJ/conversion
		EnergyDAC:     1.0,   // 1 pJ/conversion
		EnergyBuffer:  0.5,   // 0.5 pJ/byte
		EnergyLeakage: 0.01,  // 0.01 pJ/cycle
		LatencyMVM:    1,
		LatencyADC:    8,
		LatencyDAC:    4,
	}
}

// ============================================================================
// NEURAL NETWORK INTERMEDIATE REPRESENTATION (IR)
// ============================================================================

// IROperator represents a neural network operator in IR
type IROperator struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`  // "conv", "fc", "attention", etc.
	InputShape  []int             `json:"input_shape"`
	OutputShape []int             `json:"output_shape"`
	WeightShape []int             `json:"weight_shape"`
	Attributes  map[string]interface{} `json:"attributes"`

	// Mapping info (filled by compiler)
	MappedCores     []int   `json:"mapped_cores"`
	MappedCrossbars []int   `json:"mapped_crossbars"`
	TilingStrategy  *TilingStrategy `json:"tiling_strategy"`
	ComputeMode     ComputeMode `json:"compute_mode"`
}

// IRGraph represents the complete network graph
type IRGraph struct {
	Name       string        `json:"name"`
	Operators  []*IROperator `json:"operators"`
	Edges      [][2]string   `json:"edges"` // [from_id, to_id]
	InputOps   []string      `json:"input_ops"`
	OutputOps  []string      `json:"output_ops"`
}

// TilingStrategy defines how to tile an operator across crossbars
type TilingStrategy struct {
	TileM       int   `json:"tile_m"`       // Output tile height
	TileN       int   `json:"tile_n"`       // Output tile width
	TileK       int   `json:"tile_k"`       // Reduction dimension tile
	NumTilesM   int   `json:"num_tiles_m"`
	NumTilesN   int   `json:"num_tiles_n"`
	NumTilesK   int   `json:"num_tiles_k"`
	Duplication int   `json:"duplication"`  // Weight duplication factor
}

// ============================================================================
// CIM COMPILER
// ============================================================================

// CIMCompilerConfig defines compiler configuration
type CIMCompilerConfig struct {
	OptimizationLevel int  `json:"optimization_level"` // 0-3
	EnableFusion      bool `json:"enable_fusion"`      // Operator fusion
	EnableTiling      bool `json:"enable_tiling"`      // Automatic tiling
	EnablePipelining  bool `json:"enable_pipelining"`  // Pipeline stages
	EnableDualMode    bool `json:"enable_dual_mode"`   // CMSwitch mode
	TargetLatency     int  `json:"target_latency"`     // Target latency (cycles)
	TargetEnergy      float64 `json:"target_energy"`   // Target energy (pJ)
}

// CIMCompiler implements multi-level CIM compilation
type CIMCompiler struct {
	Config       *CIMCompilerConfig
	HardwareSpec *HardwareSpec
	IR           *IRGraph
	Schedule     *ExecutionSchedule
	Stats        *CompilerStats
	mu           sync.RWMutex
}

// CompilerStats tracks compilation statistics
type CompilerStats struct {
	TotalOperators     int     `json:"total_operators"`
	FusedOperators     int     `json:"fused_operators"`
	CrossbarsUsed      int     `json:"crossbars_used"`
	TotalTiles         int     `json:"total_tiles"`
	EstimatedLatency   int     `json:"estimated_latency"`
	EstimatedEnergy    float64 `json:"estimated_energy"`
	ArrayUtilization   float64 `json:"array_utilization"`
}

// ExecutionSchedule represents the compiled schedule
type ExecutionSchedule struct {
	Stages     []*ScheduleStage `json:"stages"`
	TotalCycles int             `json:"total_cycles"`
}

// ScheduleStage represents a pipeline stage
type ScheduleStage struct {
	StageID     int           `json:"stage_id"`
	Operators   []string      `json:"operators"`
	StartCycle  int           `json:"start_cycle"`
	EndCycle    int           `json:"end_cycle"`
	CoreMapping map[string]int `json:"core_mapping"`
}

// NewCIMCompiler creates a new CIM compiler
func NewCIMCompiler(config *CIMCompilerConfig, hw *HardwareSpec) *CIMCompiler {
	return &CIMCompiler{
		Config:       config,
		HardwareSpec: hw,
		Stats:        &CompilerStats{},
	}
}

// Compile compiles an IR graph to CIM execution schedule
func (c *CIMCompiler) Compile(graph *IRGraph) (*ExecutionSchedule, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.IR = graph
	c.Stats.TotalOperators = len(graph.Operators)

	// Phase 1: Operator analysis and fusion
	if c.Config.EnableFusion {
		c.fuseOperators()
	}

	// Phase 2: Tiling and mapping
	if c.Config.EnableTiling {
		c.computeTiling()
	}

	// Phase 3: Crossbar allocation
	c.allocateCrossbars()

	// Phase 4: Schedule generation
	schedule := c.generateSchedule()

	// Phase 5: Energy estimation
	c.estimateEnergy()

	c.Schedule = schedule
	return schedule, nil
}

// fuseOperators performs operator fusion optimization
func (c *CIMCompiler) fuseOperators() {
	// Identify fusible patterns: Conv+BN+ReLU, FC+ReLU, etc.
	fusiblePatterns := [][]string{
		{"conv", "batchnorm", "relu"},
		{"conv", "relu"},
		{"fc", "relu"},
		{"matmul", "add"},
	}

	fused := 0
	for _, op := range c.IR.Operators {
		// Check if op starts a fusible pattern
		for _, pattern := range fusiblePatterns {
			if op.Type == pattern[0] {
				// Try to find subsequent ops
				if c.canFusePattern(op, pattern) {
					// Mark as fused
					op.Attributes["fused_pattern"] = pattern
					fused++
				}
			}
		}
	}
	c.Stats.FusedOperators = fused
}

// canFusePattern checks if a pattern can be fused
func (c *CIMCompiler) canFusePattern(startOp *IROperator, pattern []string) bool {
	// Simplified: check if operators are adjacent in graph
	if len(pattern) <= 1 {
		return true
	}

	// Find next operator
	for _, edge := range c.IR.Edges {
		if edge[0] == startOp.ID {
			// Found successor
			for _, op := range c.IR.Operators {
				if op.ID == edge[1] && len(pattern) > 1 && op.Type == pattern[1] {
					return true
				}
			}
		}
	}
	return false
}

// computeTiling calculates tiling strategy for each operator
func (c *CIMCompiler) computeTiling() {
	hw := c.HardwareSpec.Config

	for _, op := range c.IR.Operators {
		if len(op.WeightShape) < 2 {
			continue
		}

		M := op.WeightShape[0] // Output channels
		K := op.WeightShape[1] // Input channels
		if len(op.WeightShape) > 2 {
			// Conv: multiply by kernel size
			for i := 2; i < len(op.WeightShape); i++ {
				K *= op.WeightShape[i]
			}
		}

		// Tile to fit crossbar
		tileM := hw.CrossbarRows
		if M < tileM {
			tileM = M
		}
		tileK := hw.CrossbarCols
		if K < tileK {
			tileK = K
		}

		numTilesM := (M + tileM - 1) / tileM
		numTilesK := (K + tileK - 1) / tileK

		op.TilingStrategy = &TilingStrategy{
			TileM:       tileM,
			TileK:       tileK,
			TileN:       1, // For MVM
			NumTilesM:   numTilesM,
			NumTilesK:   numTilesK,
			NumTilesN:   1,
			Duplication: 1,
		}

		c.Stats.TotalTiles += numTilesM * numTilesK
	}
}

// allocateCrossbars maps tiles to physical crossbars
func (c *CIMCompiler) allocateCrossbars() {
	hw := c.HardwareSpec.Config
	totalCrossbars := hw.NumCores * hw.CrossbarsPerCore

	crossbarIdx := 0
	totalCells := 0
	usedCells := 0

	for _, op := range c.IR.Operators {
		if op.TilingStrategy == nil {
			continue
		}

		tilesNeeded := op.TilingStrategy.NumTilesM * op.TilingStrategy.NumTilesK

		// Allocate crossbars
		op.MappedCrossbars = make([]int, tilesNeeded)
		for i := 0; i < tilesNeeded; i++ {
			op.MappedCrossbars[i] = crossbarIdx % totalCrossbars
			crossbarIdx++
		}

		// Calculate utilization
		for _, wb := range op.WeightShape {
			usedCells += wb
		}
		totalCells += tilesNeeded * hw.CrossbarRows * hw.CrossbarCols
	}

	c.Stats.CrossbarsUsed = crossbarIdx
	if totalCells > 0 {
		c.Stats.ArrayUtilization = float64(usedCells) / float64(totalCells)
	}
}

// generateSchedule creates execution schedule
func (c *CIMCompiler) generateSchedule() *ExecutionSchedule {
	schedule := &ExecutionSchedule{
		Stages: make([]*ScheduleStage, 0),
	}

	// Topological sort for dependency ordering
	sorted := c.topologicalSort()

	if c.Config.EnablePipelining {
		// Pipeline stages
		currentStage := &ScheduleStage{
			StageID:     0,
			Operators:   make([]string, 0),
			CoreMapping: make(map[string]int),
		}
		currentCycle := 0

		for _, opID := range sorted {
			op := c.getOperator(opID)
			if op == nil {
				continue
			}

			// Estimate operator latency
			latency := c.estimateOpLatency(op)

			// Add to current stage or create new
			if len(currentStage.Operators) > 0 && currentCycle+latency > c.Config.TargetLatency {
				currentStage.EndCycle = currentCycle
				schedule.Stages = append(schedule.Stages, currentStage)
				currentStage = &ScheduleStage{
					StageID:     len(schedule.Stages),
					Operators:   make([]string, 0),
					StartCycle:  currentCycle,
					CoreMapping: make(map[string]int),
				}
			}

			currentStage.Operators = append(currentStage.Operators, opID)
			if len(op.MappedCrossbars) > 0 {
				currentStage.CoreMapping[opID] = op.MappedCrossbars[0] / c.HardwareSpec.Config.CrossbarsPerCore
			}
			currentCycle += latency
		}

		// Add final stage
		currentStage.EndCycle = currentCycle
		schedule.Stages = append(schedule.Stages, currentStage)
		schedule.TotalCycles = currentCycle
	} else {
		// Sequential execution
		stage := &ScheduleStage{
			StageID:     0,
			Operators:   sorted,
			StartCycle:  0,
			CoreMapping: make(map[string]int),
		}

		totalCycles := 0
		for _, opID := range sorted {
			op := c.getOperator(opID)
			if op != nil {
				totalCycles += c.estimateOpLatency(op)
				if len(op.MappedCrossbars) > 0 {
					stage.CoreMapping[opID] = op.MappedCrossbars[0] / c.HardwareSpec.Config.CrossbarsPerCore
				}
			}
		}
		stage.EndCycle = totalCycles
		schedule.Stages = append(schedule.Stages, stage)
		schedule.TotalCycles = totalCycles
	}

	c.Stats.EstimatedLatency = schedule.TotalCycles
	return schedule
}

// topologicalSort returns operators in dependency order
func (c *CIMCompiler) topologicalSort() []string {
	// Build adjacency list
	inDegree := make(map[string]int)
	successors := make(map[string][]string)

	for _, op := range c.IR.Operators {
		inDegree[op.ID] = 0
		successors[op.ID] = make([]string, 0)
	}

	for _, edge := range c.IR.Edges {
		from, to := edge[0], edge[1]
		inDegree[to]++
		successors[from] = append(successors[from], to)
	}

	// Kahn's algorithm
	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	sorted := make([]string, 0)
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		sorted = append(sorted, curr)

		for _, succ := range successors[curr] {
			inDegree[succ]--
			if inDegree[succ] == 0 {
				queue = append(queue, succ)
			}
		}
	}

	return sorted
}

// getOperator finds operator by ID
func (c *CIMCompiler) getOperator(id string) *IROperator {
	for _, op := range c.IR.Operators {
		if op.ID == id {
			return op
		}
	}
	return nil
}

// estimateOpLatency estimates operator latency in cycles
func (c *CIMCompiler) estimateOpLatency(op *IROperator) int {
	hw := c.HardwareSpec

	if op.TilingStrategy == nil {
		return 1
	}

	// MVM latency: DAC + compute + ADC per tile
	tilesPerMVM := op.TilingStrategy.NumTilesK
	mvmLatency := tilesPerMVM * (hw.LatencyDAC + hw.LatencyMVM + hw.LatencyADC)

	// Total: all output tiles
	totalMVMs := op.TilingStrategy.NumTilesM * op.TilingStrategy.NumTilesN

	return totalMVMs * mvmLatency
}

// estimateEnergy estimates total energy consumption
func (c *CIMCompiler) estimateEnergy() {
	hw := c.HardwareSpec
	totalEnergy := 0.0

	for _, op := range c.IR.Operators {
		if op.TilingStrategy == nil {
			continue
		}

		// Compute energy
		macs := 1
		for _, dim := range op.WeightShape {
			macs *= dim
		}
		computeEnergy := float64(macs) * hw.EnergyMVM

		// ADC/DAC energy
		numADCs := op.TilingStrategy.NumTilesM * op.TilingStrategy.NumTilesK
		adcEnergy := float64(numADCs) * hw.EnergyADC
		dacEnergy := float64(numADCs) * hw.EnergyDAC

		// Buffer energy (weight + activation)
		bufferAccesses := macs * 2 // Rough estimate
		bufferEnergy := float64(bufferAccesses) * hw.EnergyBuffer

		totalEnergy += computeEnergy + adcEnergy + dacEnergy + bufferEnergy
	}

	// Leakage during execution
	leakageEnergy := float64(c.Stats.EstimatedLatency) * hw.EnergyLeakage *
		float64(hw.Config.NumCores*hw.Config.CrossbarsPerCore)

	c.Stats.EstimatedEnergy = totalEnergy + leakageEnergy
}

// ============================================================================
// CMSWITCH: DUAL-MODE COMPILER
// ============================================================================

// CMSwitchConfig defines dual-mode compiler configuration
type CMSwitchConfig struct {
	MemoryThreshold   float64 `json:"memory_threshold"`   // Utilization threshold
	ComputeThreshold  float64 `json:"compute_threshold"`  // Compute intensity threshold
	SwitchOverhead    int     `json:"switch_overhead"`    // Mode switch cycles
	EnableDynamic     bool    `json:"enable_dynamic"`     // Dynamic mode switching
}

// CMSwitchCompiler extends CIM compiler with dual-mode support
type CMSwitchCompiler struct {
	*CIMCompiler
	CMConfig    *CMSwitchConfig
	ModeAssign  map[string]ComputeMode // Operator to mode mapping
}

// NewCMSwitchCompiler creates a dual-mode aware compiler
func NewCMSwitchCompiler(config *CIMCompilerConfig, hw *HardwareSpec, cmConfig *CMSwitchConfig) *CMSwitchCompiler {
	return &CMSwitchCompiler{
		CIMCompiler: NewCIMCompiler(config, hw),
		CMConfig:    cmConfig,
		ModeAssign:  make(map[string]ComputeMode),
	}
}

// AssignModes determines compute vs memory mode for each operator
func (c *CMSwitchCompiler) AssignModes() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, op := range c.IR.Operators {
		mode := c.selectMode(op)
		c.ModeAssign[op.ID] = mode
		op.ComputeMode = mode
	}
}

// selectMode chooses optimal mode for an operator
func (c *CMSwitchCompiler) selectMode(op *IROperator) ComputeMode {
	// Calculate compute intensity (FLOPs / bytes)
	flops := 1
	bytes := 0
	for _, dim := range op.WeightShape {
		flops *= dim
		bytes += dim * 4 // Assume float32
	}

	intensity := float64(flops) / float64(bytes)

	// High intensity → compute mode (CIM)
	// Low intensity → memory mode (cache weights, digital compute)
	if intensity > c.CMConfig.ComputeThreshold {
		return ModeMVM
	}

	// Check array utilization
	if op.TilingStrategy != nil {
		utilization := float64(op.TilingStrategy.TileM*op.TilingStrategy.TileK) /
			float64(c.HardwareSpec.Config.CrossbarRows*c.HardwareSpec.Config.CrossbarCols)
		if utilization < c.CMConfig.MemoryThreshold {
			return ModeMemory
		}
	}

	return ModeMVM
}

// ============================================================================
// POWER MANAGEMENT
// ============================================================================

// PowerManagerConfig defines power management parameters
type PowerManagerConfig struct {
	// Voltage/frequency levels
	VoltageLevel    []float64 `json:"voltage_levels"`    // Vdd options
	FrequencyLevels []float64 `json:"frequency_levels"`  // MHz options

	// Power gating
	EnablePowerGating   bool    `json:"enable_power_gating"`
	GatingOverhead      int     `json:"gating_overhead"`      // Cycles to gate/ungate
	MinIdleCycles       int     `json:"min_idle_cycles"`      // Min idle before gating

	// Leakage model
	LeakagePowerPerCore float64 `json:"leakage_power_per_core"` // mW
	LeakageVoltageExp   float64 `json:"leakage_voltage_exp"`    // V^exp dependence

	// DVFS
	EnableDVFS          bool    `json:"enable_dvfs"`
	DVFSTransitionTime  int     `json:"dvfs_transition_time"`   // Cycles
}

// PowerState represents current power configuration
type PowerState struct {
	VoltageIdx     int       `json:"voltage_idx"`
	FrequencyIdx   int       `json:"frequency_idx"`
	ActiveCores    []bool    `json:"active_cores"`
	GatedCrossbars [][]bool  `json:"gated_crossbars"` // [core][crossbar]
}

// PowerManager handles power optimization for CIM
type PowerManager struct {
	Config      *PowerManagerConfig
	Hardware    *HardwareSpec
	State       *PowerState
	Stats       *PowerStats
	mu          sync.RWMutex
}

// PowerStats tracks power consumption
type PowerStats struct {
	TotalEnergy      float64 `json:"total_energy"`       // mJ
	DynamicEnergy    float64 `json:"dynamic_energy"`     // mJ
	LeakageEnergy    float64 `json:"leakage_energy"`     // mJ
	GatingEnergy     float64 `json:"gating_energy"`      // mJ (overhead)
	DVFSTransitions  int     `json:"dvfs_transitions"`
	GatingEvents     int     `json:"gating_events"`
	ActiveCycles     int64   `json:"active_cycles"`
	IdleCycles       int64   `json:"idle_cycles"`
}

// NewPowerManager creates a new power manager
func NewPowerManager(config *PowerManagerConfig, hw *HardwareSpec) *PowerManager {
	numCores := hw.Config.NumCores
	crossbarsPerCore := hw.Config.CrossbarsPerCore

	activeCores := make([]bool, numCores)
	gatedCrossbars := make([][]bool, numCores)
	for i := range activeCores {
		activeCores[i] = true
		gatedCrossbars[i] = make([]bool, crossbarsPerCore)
	}

	return &PowerManager{
		Config:   config,
		Hardware: hw,
		State: &PowerState{
			VoltageIdx:     len(config.VoltageLevel) - 1, // Start at max
			FrequencyIdx:   len(config.FrequencyLevels) - 1,
			ActiveCores:    activeCores,
			GatedCrossbars: gatedCrossbars,
		},
		Stats: &PowerStats{},
	}
}

// SetDVFSLevel changes voltage/frequency level
func (p *PowerManager) SetDVFSLevel(vIdx, fIdx int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if vIdx != p.State.VoltageIdx || fIdx != p.State.FrequencyIdx {
		p.Stats.DVFSTransitions++
		p.State.VoltageIdx = vIdx
		p.State.FrequencyIdx = fIdx
	}
}

// GateCrossbar power-gates a specific crossbar
func (p *PowerManager) GateCrossbar(coreIdx, crossbarIdx int, gated bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.State.GatedCrossbars[coreIdx][crossbarIdx] != gated {
		p.State.GatedCrossbars[coreIdx][crossbarIdx] = gated
		p.Stats.GatingEvents++
		p.Stats.GatingEnergy += float64(p.Config.GatingOverhead) * 0.001 // Rough estimate
	}
}

// GateCore power-gates an entire core
func (p *PowerManager) GateCore(coreIdx int, active bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.State.ActiveCores[coreIdx] != active {
		p.State.ActiveCores[coreIdx] = active
		p.Stats.GatingEvents++
		// Gate/ungate all crossbars in core
		for i := range p.State.GatedCrossbars[coreIdx] {
			p.State.GatedCrossbars[coreIdx][i] = !active
		}
	}
}

// CalculatePower computes power consumption for current state
func (p *PowerManager) CalculatePower() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	voltage := p.Config.VoltageLevel[p.State.VoltageIdx]
	frequency := p.Config.FrequencyLevels[p.State.FrequencyIdx]

	// Count active elements
	activeCores := 0
	activeCrossbars := 0
	for i, active := range p.State.ActiveCores {
		if active {
			activeCores++
			for _, gated := range p.State.GatedCrossbars[i] {
				if !gated {
					activeCrossbars++
				}
			}
		}
	}

	// Dynamic power: P = C * V^2 * f
	// Simplified: proportional to V^2 * f * active_units
	dynamicPower := voltage * voltage * frequency * float64(activeCrossbars) * 0.001 // mW

	// Leakage power: P = P0 * V^exp
	leakagePower := p.Config.LeakagePowerPerCore * float64(activeCores) *
		math.Pow(voltage, p.Config.LeakageVoltageExp)

	return dynamicPower + leakagePower
}

// OptimizeDVFS selects optimal DVFS level for workload
func (p *PowerManager) OptimizeDVFS(targetLatency int, currentLatency int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.Config.EnableDVFS {
		return
	}

	// If ahead of schedule, reduce frequency
	if currentLatency < targetLatency {
		slack := float64(targetLatency-currentLatency) / float64(targetLatency)

		// Find lowest frequency that meets target
		for fIdx := 0; fIdx < len(p.Config.FrequencyLevels); fIdx++ {
			freqRatio := p.Config.FrequencyLevels[fIdx] / p.Config.FrequencyLevels[p.State.FrequencyIdx]
			scaledLatency := int(float64(currentLatency) / freqRatio)

			if scaledLatency <= targetLatency {
				// Also reduce voltage if possible (V/f scaling)
				vIdx := fIdx
				if vIdx >= len(p.Config.VoltageLevel) {
					vIdx = len(p.Config.VoltageLevel) - 1
				}
				p.State.VoltageIdx = vIdx
				p.State.FrequencyIdx = fIdx
				p.Stats.DVFSTransitions++
				break
			}
		}
	}
}

// SimulateExecution simulates power during execution
func (p *PowerManager) SimulateExecution(schedule *ExecutionSchedule) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, stage := range schedule.Stages {
		cycles := stage.EndCycle - stage.StartCycle

		// Track active vs idle
		usedCores := make(map[int]bool)
		for _, coreIdx := range stage.CoreMapping {
			usedCores[coreIdx] = true
		}

		// Power gate unused cores if enabled
		if p.Config.EnablePowerGating {
			for i := range p.State.ActiveCores {
				if _, used := usedCores[i]; !used && cycles > p.Config.MinIdleCycles {
					p.GateCore(i, false)
				} else {
					p.GateCore(i, true)
				}
			}
		}

		// Calculate energy for this stage
		power := p.CalculatePower()
		frequency := p.Config.FrequencyLevels[p.State.FrequencyIdx]
		time := float64(cycles) / (frequency * 1e6) // seconds
		energy := power * time * 1000 // mJ

		p.Stats.TotalEnergy += energy
		p.Stats.ActiveCycles += int64(cycles)
	}
}

// ============================================================================
// HYBRID MEMORY HIERARCHY
// ============================================================================

// MemoryTier represents a tier in the memory hierarchy
type MemoryTier struct {
	Type          string  `json:"type"`           // "SRAM", "MRAM", "ReRAM"
	CapacityKB    int     `json:"capacity_kb"`
	ReadLatency   int     `json:"read_latency"`   // Cycles
	WriteLatency  int     `json:"write_latency"`  // Cycles
	ReadEnergy    float64 `json:"read_energy"`    // pJ/byte
	WriteEnergy   float64 `json:"write_energy"`   // pJ/byte
	LeakagePower  float64 `json:"leakage_power"`  // mW
	Endurance     int64   `json:"endurance"`      // Write cycles
	NonVolatile   bool    `json:"non_volatile"`
	SupportsCIM   bool    `json:"supports_cim"`
}

// HybridMemoryConfig defines hybrid memory hierarchy
type HybridMemoryConfig struct {
	Tiers           []*MemoryTier `json:"tiers"`
	WearLeveling    bool          `json:"wear_leveling"`
	AdaptivePlacement bool        `json:"adaptive_placement"`
}

// HybridMemory manages multi-tier memory hierarchy
type HybridMemory struct {
	Config    *HybridMemoryConfig
	TierUsage []int64           // Bytes used per tier
	TierWear  []int64           // Write cycles per tier
	DataMap   map[string]int    // Data name -> tier index
	Stats     *HybridMemoryStats
	mu        sync.RWMutex
}

// HybridMemoryStats tracks memory statistics
type HybridMemoryStats struct {
	TotalReads      int64   `json:"total_reads"`
	TotalWrites     int64   `json:"total_writes"`
	TierHits        []int64 `json:"tier_hits"`
	TotalEnergy     float64 `json:"total_energy"`      // mJ
	IdleEnergy      float64 `json:"idle_energy"`       // mJ (SRAM leakage)
	EnduranceMargin []float64 `json:"endurance_margin"` // Remaining lifetime
}

// NewHybridMemory creates a new hybrid memory system
func NewHybridMemory(config *HybridMemoryConfig) *HybridMemory {
	numTiers := len(config.Tiers)

	return &HybridMemory{
		Config:    config,
		TierUsage: make([]int64, numTiers),
		TierWear:  make([]int64, numTiers),
		DataMap:   make(map[string]int),
		Stats: &HybridMemoryStats{
			TierHits:        make([]int64, numTiers),
			EnduranceMargin: make([]float64, numTiers),
		},
	}
}

// AllocateData places data in optimal tier
func (h *HybridMemory) AllocateData(name string, sizeBytes int64, accessPattern string) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Select tier based on access pattern and available space
	selectedTier := 0

	switch accessPattern {
	case "weights":
		// Prefer non-volatile for weights (no reload needed)
		for i, tier := range h.Config.Tiers {
			if tier.NonVolatile && tier.SupportsCIM {
				if h.TierUsage[i]+sizeBytes <= int64(tier.CapacityKB)*1024 {
					selectedTier = i
					break
				}
			}
		}
	case "activations":
		// Prefer SRAM for fast activations
		for i, tier := range h.Config.Tiers {
			if tier.Type == "SRAM" {
				if h.TierUsage[i]+sizeBytes <= int64(tier.CapacityKB)*1024 {
					selectedTier = i
					break
				}
			}
		}
	case "kv_cache":
		// MRAM for medium-term persistence
		for i, tier := range h.Config.Tiers {
			if tier.Type == "MRAM" {
				if h.TierUsage[i]+sizeBytes <= int64(tier.CapacityKB)*1024 {
					selectedTier = i
					break
				}
			}
		}
	default:
		// Find first tier with space
		for i, tier := range h.Config.Tiers {
			if h.TierUsage[i]+sizeBytes <= int64(tier.CapacityKB)*1024 {
				selectedTier = i
				break
			}
		}
	}

	h.TierUsage[selectedTier] += sizeBytes
	h.DataMap[name] = selectedTier
	return selectedTier
}

// AccessData simulates data access
func (h *HybridMemory) AccessData(name string, bytes int64, isWrite bool) float64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	tier, ok := h.DataMap[name]
	if !ok {
		tier = 0 // Default to first tier
	}

	h.Stats.TierHits[tier]++

	var energy float64
	if isWrite {
		h.Stats.TotalWrites++
		h.TierWear[tier]++
		energy = float64(bytes) * h.Config.Tiers[tier].WriteEnergy / 1e9 // pJ to mJ
	} else {
		h.Stats.TotalReads++
		energy = float64(bytes) * h.Config.Tiers[tier].ReadEnergy / 1e9
	}

	h.Stats.TotalEnergy += energy
	return energy
}

// UpdateEnduranceMargin calculates remaining endurance
func (h *HybridMemory) UpdateEnduranceMargin() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, tier := range h.Config.Tiers {
		if tier.Endurance > 0 {
			h.Stats.EnduranceMargin[i] = 1.0 - float64(h.TierWear[i])/float64(tier.Endurance)
		} else {
			h.Stats.EnduranceMargin[i] = 1.0 // Infinite endurance
		}
	}
}

// CalculateIdleEnergy computes idle/leakage energy
func (h *HybridMemory) CalculateIdleEnergy(idleTimeMs float64) float64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	idleEnergy := 0.0
	for _, tier := range h.Config.Tiers {
		idleEnergy += tier.LeakagePower * idleTimeMs / 1000 // mW * s = mJ
	}

	h.Stats.IdleEnergy += idleEnergy
	return idleEnergy
}

// CreateTypicalHybridConfig creates a typical 3-tier hierarchy
func CreateTypicalHybridConfig() *HybridMemoryConfig {
	return &HybridMemoryConfig{
		Tiers: []*MemoryTier{
			{
				Type:         "SRAM",
				CapacityKB:   256,
				ReadLatency:  1,
				WriteLatency: 1,
				ReadEnergy:   0.1,
				WriteEnergy:  0.1,
				LeakagePower: 10.0, // High leakage
				Endurance:    -1,   // Unlimited
				NonVolatile:  false,
				SupportsCIM:  true,
			},
			{
				Type:         "MRAM",
				CapacityKB:   1024,
				ReadLatency:  3,
				WriteLatency: 10,
				ReadEnergy:   0.5,
				WriteEnergy:  5.0, // Higher write energy
				LeakagePower: 0.1, // Low leakage
				Endurance:    1e12,
				NonVolatile:  true,
				SupportsCIM:  true,
			},
			{
				Type:         "ReRAM",
				CapacityKB:   4096,
				ReadLatency:  5,
				WriteLatency: 50,
				ReadEnergy:   1.0,
				WriteEnergy:  10.0,
				LeakagePower: 0.01, // Very low leakage
				Endurance:    1e6,
				NonVolatile:  true,
				SupportsCIM:  true,
			},
		},
		WearLeveling:      true,
		AdaptivePlacement: true,
	}
}

// ============================================================================
// SERIALIZATION
// ============================================================================

// CompilerPowerState holds serializable state
type CompilerPowerState struct {
	CompilerStats   *CompilerStats     `json:"compiler_stats"`
	PowerStats      *PowerStats        `json:"power_stats"`
	MemoryStats     *HybridMemoryStats `json:"memory_stats"`
	Schedule        *ExecutionSchedule `json:"schedule"`
}

// ExportState exports compiler and power management state
func ExportCompilerPowerState(compiler *CIMCompiler, power *PowerManager, memory *HybridMemory) ([]byte, error) {
	state := &CompilerPowerState{
		CompilerStats: compiler.Stats,
		Schedule:      compiler.Schedule,
	}

	if power != nil {
		state.PowerStats = power.Stats
	}
	if memory != nil {
		state.MemoryStats = memory.Stats
	}

	return json.MarshalIndent(state, "", "  ")
}

// ============================================================================
// DESIGN SPACE EXPLORATION
// ============================================================================

// DSEConfig defines design space exploration parameters
type DSEConfig struct {
	CrossbarSizes    []int     `json:"crossbar_sizes"`    // [64, 128, 256]
	ADCPrecisions    []int     `json:"adc_precisions"`    // [4, 6, 8]
	NumCoresOptions  []int     `json:"num_cores_options"` // [1, 2, 4, 8]
	OptimizeFor      string    `json:"optimize_for"`      // "energy", "latency", "area"
}

// DSEResult represents a design point evaluation
type DSEResult struct {
	Config         *AbsArchConfig `json:"config"`
	Energy         float64        `json:"energy"`
	Latency        int            `json:"latency"`
	Utilization    float64        `json:"utilization"`
	Score          float64        `json:"score"`
}

// RunDSE performs design space exploration
func RunDSE(dseConfig *DSEConfig, ir *IRGraph) []DSEResult {
	results := make([]DSEResult, 0)

	for _, crossbarSize := range dseConfig.CrossbarSizes {
		for _, adcPrec := range dseConfig.ADCPrecisions {
			for _, numCores := range dseConfig.NumCoresOptions {
				// Create hardware spec for this design point
				hw := &HardwareSpec{
					Name: fmt.Sprintf("CIM_%dx%d_ADC%d_C%d", crossbarSize, crossbarSize, adcPrec, numCores),
					Config: &AbsArchConfig{
						NumCores:         numCores,
						GlobalBufferKB:   256,
						CrossbarsPerCore: 16,
						LocalBufferKB:    32,
						CrossbarRows:     crossbarSize,
						CrossbarCols:     crossbarSize,
						DevicePrecision:  4,
						ADCPrecision:     adcPrec,
						DACPrecision:     8,
					},
					EnergyMVM:     0.1,
					EnergyADC:     float64(1<<adcPrec) * 0.5, // Scales with precision
					EnergyDAC:     1.0,
					EnergyBuffer:  0.5,
					EnergyLeakage: 0.01,
					LatencyMVM:    1,
					LatencyADC:    adcPrec,
					LatencyDAC:    4,
				}

				// Compile and evaluate
				compiler := NewCIMCompiler(&CIMCompilerConfig{
					OptimizationLevel: 2,
					EnableFusion:      true,
					EnableTiling:      true,
					EnablePipelining:  true,
				}, hw)

				compiler.Compile(ir)

				// Calculate score based on optimization target
				var score float64
				switch dseConfig.OptimizeFor {
				case "energy":
					score = 1.0 / compiler.Stats.EstimatedEnergy
				case "latency":
					score = 1.0 / float64(compiler.Stats.EstimatedLatency)
				case "area":
					area := float64(numCores * 16 * crossbarSize * crossbarSize)
					score = 1.0 / area
				default:
					// EDP (Energy-Delay Product)
					score = 1.0 / (compiler.Stats.EstimatedEnergy * float64(compiler.Stats.EstimatedLatency))
				}

				results = append(results, DSEResult{
					Config:      hw.Config,
					Energy:      compiler.Stats.EstimatedEnergy,
					Latency:     compiler.Stats.EstimatedLatency,
					Utilization: compiler.Stats.ArrayUtilization,
					Score:       score,
				})
			}
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

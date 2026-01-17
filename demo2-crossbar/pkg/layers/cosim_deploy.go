// Package layers provides hardware-software co-simulation and CIM model deployment.
//
// This module implements:
// - Non-ideality simulation (IR drop, stuck-at-fault, retention, variation)
// - Full-stack CIM system simulation with FAST-style functional array simulator
// - DNN compiler for CIM accelerators (ONNX-like model import)
// - Weight mapping and tiling strategies for crossbar arrays
// - Dataflow scheduling and optimization
// - PIMCOMP-style end-to-end deployment pipeline
//
// Based on:
// - FAST: Functional Array Simulator (Science China 2025)
// - SIMBRAIN: Nonidealities-aware SNN Simulation (ScienceDirect 2025)
// - Full-Stack CIM System (Nature Communications 2025)
// - PIMCOMP: End-to-End DNN Compiler (arXiv 2024)
// - CMSwitch: Dual-mode CIM Compiler (arXiv 2025)
package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
)

// =============================================================================
// Part 1: Non-Ideality Models for Memristor Crossbars
// =============================================================================

// NonIdealityConfig configures crossbar non-idealities
type NonIdealityConfig struct {
	// IR Drop parameters
	WireResistanceOhm    float64 // Wire resistance per cell (Ω)
	EnableIRDrop         bool    // Enable IR drop simulation

	// Stuck-at-fault parameters
	SAFRate              float64 // Probability of stuck-at-fault (0-1)
	SAF_HRS_Ratio        float64 // Ratio of stuck-at-HRS vs stuck-at-LRS

	// Retention parameters
	RetentionTau_s       float64 // Retention time constant (seconds)
	EnableRetention      bool    // Enable retention degradation

	// Variation parameters
	D2DVariation         float64 // Device-to-device variation coefficient (0-0.35)
	C2CVariation         float64 // Cycle-to-cycle variation coefficient (0-0.06)
	EnableVariation      bool    // Enable variation simulation

	// ADC/DAC precision
	ADCBits              int     // ADC bit precision
	DACBits              int     // DAC bit precision
	ADCNoise_LSB         float64 // ADC noise in LSBs

	// Temperature effects
	Temperature_C        float64 // Operating temperature
	TempCoeff            float64 // Temperature coefficient
}

// DefaultNonIdealityConfig returns typical non-ideality parameters
func DefaultNonIdealityConfig() NonIdealityConfig {
	return NonIdealityConfig{
		WireResistanceOhm: 2.0,
		EnableIRDrop:      true,
		SAFRate:           0.01,
		SAF_HRS_Ratio:     0.5,
		RetentionTau_s:    86400,    // 1 day
		EnableRetention:   true,
		D2DVariation:      0.10,     // 10%
		C2CVariation:      0.03,     // 3%
		EnableVariation:   true,
		ADCBits:           6,
		DACBits:           8,
		ADCNoise_LSB:      0.5,
		Temperature_C:     25.0,
		TempCoeff:         0.001,
	}
}

// StuckAtFaultType represents the type of stuck-at-fault
type StuckAtFaultType int

const (
	NoFault StuckAtFaultType = iota
	StuckAtHRS               // Stuck at high resistance (open circuit)
	StuckAtLRS               // Stuck at low resistance (short circuit)
)

// MemristorCell represents a single memristor with non-idealities
type MemristorCell struct {
	Row              int
	Col              int
	NominalG         float64          // Nominal conductance
	ActualG          float64          // Actual conductance with non-idealities
	FaultType        StuckAtFaultType // Stuck-at-fault state
	LastWriteTime    float64          // Timestamp of last write
	D2DFactor        float64          // Device-to-device variation factor
	WriteCount       int              // Number of write cycles
}

// NonIdealCrossbar represents a crossbar with full non-ideality simulation
type NonIdealCrossbar struct {
	Config       NonIdealityConfig
	Rows         int
	Cols         int
	Cells        [][]*MemristorCell
	GMin         float64 // Minimum conductance
	GMax         float64 // Maximum conductance
	CurrentTime  float64 // Simulation time in seconds
	rng          *rand.Rand
	mu           sync.RWMutex
}

// NewNonIdealCrossbar creates a crossbar with non-ideality simulation
func NewNonIdealCrossbar(rows, cols int, config NonIdealityConfig) *NonIdealCrossbar {
	rng := rand.New(rand.NewSource(42))

	cells := make([][]*MemristorCell, rows)
	for i := range cells {
		cells[i] = make([]*MemristorCell, cols)
		for j := range cells[i] {
			// Initialize D2D variation factor
			d2dFactor := 1.0
			if config.EnableVariation {
				d2dFactor = 1.0 + rng.NormFloat64()*config.D2DVariation
			}

			// Initialize fault state
			faultType := NoFault
			if rng.Float64() < config.SAFRate {
				if rng.Float64() < config.SAF_HRS_Ratio {
					faultType = StuckAtHRS
				} else {
					faultType = StuckAtLRS
				}
			}

			cells[i][j] = &MemristorCell{
				Row:           i,
				Col:           j,
				NominalG:      0,
				ActualG:       0,
				FaultType:     faultType,
				LastWriteTime: 0,
				D2DFactor:     d2dFactor,
				WriteCount:    0,
			}
		}
	}

	return &NonIdealCrossbar{
		Config: config,
		Rows:   rows,
		Cols:   cols,
		Cells:  cells,
		GMin:   1e-6,  // 1 µS
		GMax:   100e-6, // 100 µS
		rng:    rng,
	}
}

// WriteWeight writes a weight to a cell with variation simulation
func (nc *NonIdealCrossbar) WriteWeight(row, col int, weight float64) {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if row >= nc.Rows || col >= nc.Cols {
		return
	}

	cell := nc.Cells[row][col]

	// Map weight to conductance
	nominalG := nc.GMin + (weight+1)/2*(nc.GMax-nc.GMin) // Assume weight in [-1,1]
	cell.NominalG = nominalG

	// Apply non-idealities
	actualG := nominalG

	// 1. Device-to-device variation
	if nc.Config.EnableVariation {
		actualG *= cell.D2DFactor
	}

	// 2. Cycle-to-cycle variation
	if nc.Config.EnableVariation {
		c2cNoise := nc.rng.NormFloat64() * nc.Config.C2CVariation
		actualG *= (1 + c2cNoise)
	}

	// 3. Stuck-at-fault
	switch cell.FaultType {
	case StuckAtHRS:
		actualG = nc.GMin
	case StuckAtLRS:
		actualG = nc.GMax
	}

	// Clamp to valid range
	actualG = math.Max(nc.GMin, math.Min(nc.GMax, actualG))

	cell.ActualG = actualG
	cell.LastWriteTime = nc.CurrentTime
	cell.WriteCount++
}

// ApplyRetention updates conductances based on retention decay
func (nc *NonIdealCrossbar) ApplyRetention() {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	if !nc.Config.EnableRetention {
		return
	}

	for i := range nc.Cells {
		for j := range nc.Cells[i] {
			cell := nc.Cells[i][j]
			if cell.FaultType != NoFault {
				continue // Faulty cells don't decay
			}

			// Exponential decay toward middle conductance
			age := nc.CurrentTime - cell.LastWriteTime
			decayFactor := math.Exp(-age / nc.Config.RetentionTau_s)
			midG := (nc.GMin + nc.GMax) / 2
			cell.ActualG = midG + (cell.ActualG-midG)*decayFactor
		}
	}
}

// ComputeIRDrop calculates voltage drop across the crossbar
func (nc *NonIdealCrossbar) ComputeIRDrop(inputVoltages []float64) [][]float64 {
	if !nc.Config.EnableIRDrop {
		return nil
	}

	// Simplified IR drop model: voltage decreases along rows and columns
	voltageMatrix := make([][]float64, nc.Rows)
	for i := range voltageMatrix {
		voltageMatrix[i] = make([]float64, nc.Cols)
		for j := range voltageMatrix[i] {
			// Voltage drop proportional to distance and current
			rowDrop := float64(j) * nc.Config.WireResistanceOhm * 1e-3
			colDrop := float64(i) * nc.Config.WireResistanceOhm * 1e-3
			if j < len(inputVoltages) {
				voltageMatrix[i][j] = inputVoltages[j] * (1 - rowDrop - colDrop)
			}
		}
	}

	return voltageMatrix
}

// MatVecMulWithNonIdealities performs MVM with full non-ideality simulation
func (nc *NonIdealCrossbar) MatVecMulWithNonIdealities(input []float64) []float64 {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	// DAC quantization
	quantInput := nc.quantizeDAC(input)

	// Compute IR drop
	voltageMatrix := nc.ComputeIRDrop(quantInput)

	// Analog computation with non-idealities
	output := make([]float64, nc.Rows)
	for i := 0; i < nc.Rows; i++ {
		var current float64
		for j := 0; j < nc.Cols && j < len(quantInput); j++ {
			voltage := quantInput[j]
			if voltageMatrix != nil {
				voltage = voltageMatrix[i][j]
			}
			current += nc.Cells[i][j].ActualG * voltage
		}
		output[i] = current
	}

	// ADC quantization and noise
	quantOutput := nc.quantizeADC(output)

	return quantOutput
}

// quantizeDAC applies DAC quantization
func (nc *NonIdealCrossbar) quantizeDAC(input []float64) []float64 {
	levels := float64(1 << nc.Config.DACBits)
	output := make([]float64, len(input))
	for i, v := range input {
		// Normalize to [0, 1], quantize, denormalize
		normalized := (v + 1) / 2
		quantized := math.Round(normalized*levels) / levels
		output[i] = quantized*2 - 1
	}
	return output
}

// quantizeADC applies ADC quantization with noise
func (nc *NonIdealCrossbar) quantizeADC(input []float64) []float64 {
	levels := float64(1 << nc.Config.ADCBits)
	output := make([]float64, len(input))

	// Find range for normalization
	maxVal := 0.0
	for _, v := range input {
		if math.Abs(v) > maxVal {
			maxVal = math.Abs(v)
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	for i, v := range input {
		// Normalize, quantize, add noise
		normalized := v / maxVal
		quantized := math.Round((normalized+1)/2*levels) / levels * 2 - 1
		noise := nc.rng.NormFloat64() * nc.Config.ADCNoise_LSB / levels
		output[i] = (quantized + noise) * maxVal
	}
	return output
}

// GetFaultMap returns locations of stuck-at-faults
func (nc *NonIdealCrossbar) GetFaultMap() [][]StuckAtFaultType {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	faultMap := make([][]StuckAtFaultType, nc.Rows)
	for i := range faultMap {
		faultMap[i] = make([]StuckAtFaultType, nc.Cols)
		for j := range faultMap[i] {
			faultMap[i][j] = nc.Cells[i][j].FaultType
		}
	}
	return faultMap
}

// AdvanceTime advances simulation time
func (nc *NonIdealCrossbar) AdvanceTime(deltaSeconds float64) {
	nc.mu.Lock()
	nc.CurrentTime += deltaSeconds
	nc.mu.Unlock()
	nc.ApplyRetention()
}

// =============================================================================
// Part 2: CAFM (Comparator-based Activation Function Modulation)
// =============================================================================

// CAFMConfig configures CAFM for IR drop compensation
type CAFMConfig struct {
	CompensationFactor float64 // Compensation strength
	ThresholdShift     float64 // Threshold shift for activation
	EnableCAFM         bool    // Enable CAFM
}

// CAFMCompensator implements CAFM for IR drop recovery
type CAFMCompensator struct {
	Config           CAFMConfig
	IRDropProfile    [][]float64 // Measured IR drop profile
	CompensationLUT  []float64   // Lookup table for compensation
}

// NewCAFMCompensator creates a CAFM compensator
func NewCAFMCompensator(config CAFMConfig, crossbar *NonIdealCrossbar) *CAFMCompensator {
	// Build compensation LUT based on crossbar characteristics
	lutSize := 256
	lut := make([]float64, lutSize)
	for i := range lut {
		x := float64(i) / float64(lutSize-1) * 2 - 1 // [-1, 1]
		// Nonlinear compensation function
		lut[i] = x * (1 + config.CompensationFactor*math.Abs(x))
	}

	return &CAFMCompensator{
		Config:          config,
		CompensationLUT: lut,
	}
}

// CompensateOutput applies CAFM compensation to MVM output
func (c *CAFMCompensator) CompensateOutput(output []float64) []float64 {
	if !c.Config.EnableCAFM {
		return output
	}

	compensated := make([]float64, len(output))
	lutSize := len(c.CompensationLUT)

	for i, v := range output {
		// Normalize to LUT index
		normalized := (v + 1) / 2
		idx := int(normalized * float64(lutSize-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= lutSize {
			idx = lutSize - 1
		}
		compensated[i] = c.CompensationLUT[idx]
	}

	return compensated
}

// =============================================================================
// Part 3: Full-Stack CIM System Simulation (FAST-style)
// =============================================================================

// FASTConfig configures the Functional Array Simulator
type FASTConfig struct {
	CrossbarRows     int
	CrossbarCols     int
	NonIdealityCfg   NonIdealityConfig
	EnableSparsity   bool    // Use sparse coefficient matrix
	SparsityThreshold float64 // Threshold for sparsity
}

// FASTSimulator implements FAST-style functional array simulation
type FASTSimulator struct {
	Config           FASTConfig
	Crossbar         *NonIdealCrossbar
	CAFMComp         *CAFMCompensator
	SparseCoeffMatrix map[int]map[int]float64 // Sparse coefficient storage
	EnergyConsumed_J float64
	LatencyUs        float64
	Throughput_GOPS  float64
}

// NewFASTSimulator creates a FAST simulator
func NewFASTSimulator(config FASTConfig) *FASTSimulator {
	crossbar := NewNonIdealCrossbar(config.CrossbarRows, config.CrossbarCols, config.NonIdealityCfg)

	cafmConfig := CAFMConfig{
		CompensationFactor: 0.2,
		ThresholdShift:     0.1,
		EnableCAFM:         true,
	}
	cafm := NewCAFMCompensator(cafmConfig, crossbar)

	return &FASTSimulator{
		Config:            config,
		Crossbar:          crossbar,
		CAFMComp:          cafm,
		SparseCoeffMatrix: make(map[int]map[int]float64),
	}
}

// MapWeights maps neural network weights to the crossbar
func (fs *FASTSimulator) MapWeights(weights [][]float64) error {
	rows := len(weights)
	if rows == 0 {
		return fmt.Errorf("empty weight matrix")
	}
	cols := len(weights[0])

	if rows > fs.Config.CrossbarRows || cols > fs.Config.CrossbarCols {
		return fmt.Errorf("weight matrix too large for crossbar")
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			fs.Crossbar.WriteWeight(i, j, weights[i][j])

			// Build sparse coefficient matrix
			if fs.Config.EnableSparsity && math.Abs(weights[i][j]) > fs.Config.SparsityThreshold {
				if fs.SparseCoeffMatrix[i] == nil {
					fs.SparseCoeffMatrix[i] = make(map[int]float64)
				}
				fs.SparseCoeffMatrix[i][j] = weights[i][j]
			}
		}
	}

	return nil
}

// SimulateInference runs inference with full non-ideality simulation
func (fs *FASTSimulator) SimulateInference(input []float64) []float64 {
	// MVM with non-idealities
	output := fs.Crossbar.MatVecMulWithNonIdealities(input)

	// Apply CAFM compensation
	compensated := fs.CAFMComp.CompensateOutput(output)

	// Update metrics
	numOps := float64(fs.Config.CrossbarRows * fs.Config.CrossbarCols)
	fs.EnergyConsumed_J += numOps * 1e-12 // ~1 pJ per MAC
	fs.LatencyUs += numOps / 1e6          // ~1 us per million ops
	fs.Throughput_GOPS = numOps / (fs.LatencyUs * 1e-6) / 1e9

	return compensated
}

// EvaluateAccuracy evaluates inference accuracy under non-idealities
func (fs *FASTSimulator) EvaluateAccuracy(testInputs [][]float64, testLabels []int, idealOutputs [][]float64) float64 {
	correct := 0
	for i, input := range testInputs {
		output := fs.SimulateInference(input)

		// Find predicted class
		predicted := 0
		maxVal := output[0]
		for j, v := range output {
			if v > maxVal {
				maxVal = v
				predicted = j
			}
		}

		if predicted == testLabels[i] {
			correct++
		}
	}

	return float64(correct) / float64(len(testInputs)) * 100
}

// GetMetrics returns simulation metrics
func (fs *FASTSimulator) GetMetrics() map[string]float64 {
	faultCount := 0
	for i := range fs.Crossbar.Cells {
		for j := range fs.Crossbar.Cells[i] {
			if fs.Crossbar.Cells[i][j].FaultType != NoFault {
				faultCount++
			}
		}
	}

	return map[string]float64{
		"energy_J":         fs.EnergyConsumed_J,
		"latency_us":       fs.LatencyUs,
		"throughput_GOPS":  fs.Throughput_GOPS,
		"fault_rate":       float64(faultCount) / float64(fs.Config.CrossbarRows*fs.Config.CrossbarCols),
		"sparsity":         float64(len(fs.SparseCoeffMatrix)) / float64(fs.Config.CrossbarRows),
	}
}

// =============================================================================
// Part 4: DNN Model Representation (ONNX-like)
// =============================================================================

// OpType represents neural network operation types
type OpType string

const (
	OpConv2D     OpType = "Conv2D"
	OpMatMul     OpType = "MatMul"
	OpReLU       OpType = "ReLU"
	OpSoftmax    OpType = "Softmax"
	OpBatchNorm  OpType = "BatchNorm"
	OpMaxPool    OpType = "MaxPool"
	OpAdd        OpType = "Add"
	OpFlatten    OpType = "Flatten"
	OpAttention  OpType = "Attention"
)

// TensorShape represents tensor dimensions
type TensorShape struct {
	Batch   int
	Height  int
	Width   int
	Channel int
}

// NodeDef defines a computation node in the graph
type NodeDef struct {
	Name       string
	OpType     OpType
	Inputs     []string
	Outputs    []string
	Weights    [][]float64
	Bias       []float64
	Attributes map[string]interface{}
}

// ModelGraph represents a neural network computation graph
type ModelGraph struct {
	Name       string
	Nodes      []*NodeDef
	InputNames []string
	OutputNames []string
	NodeMap    map[string]*NodeDef
}

// NewModelGraph creates an empty model graph
func NewModelGraph(name string) *ModelGraph {
	return &ModelGraph{
		Name:    name,
		Nodes:   make([]*NodeDef, 0),
		NodeMap: make(map[string]*NodeDef),
	}
}

// AddNode adds a node to the graph
func (mg *ModelGraph) AddNode(node *NodeDef) {
	mg.Nodes = append(mg.Nodes, node)
	mg.NodeMap[node.Name] = node
}

// TopoSort returns nodes in topological order
func (mg *ModelGraph) TopoSort() []*NodeDef {
	// Build dependency graph
	inDegree := make(map[string]int)
	for _, node := range mg.Nodes {
		if _, exists := inDegree[node.Name]; !exists {
			inDegree[node.Name] = 0
		}
		for _, inputName := range node.Inputs {
			if _, exists := mg.NodeMap[inputName]; exists {
				inDegree[node.Name]++
			}
		}
	}

	// Kahn's algorithm
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	sorted := make([]*NodeDef, 0)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if node, exists := mg.NodeMap[name]; exists {
			sorted = append(sorted, node)
			for _, output := range node.Outputs {
				for _, n := range mg.Nodes {
					for _, input := range n.Inputs {
						if input == output {
							inDegree[n.Name]--
							if inDegree[n.Name] == 0 {
								queue = append(queue, n.Name)
							}
						}
					}
				}
			}
		}
	}

	return sorted
}

// =============================================================================
// Part 5: CIM Compiler (PIMCOMP-style)
// =============================================================================

// TilingStrategy represents how to tile a layer across crossbars
type TilingStrategy struct {
	RowTiles    int // Number of row tiles
	ColTiles    int // Number of column tiles
	TileRows    int // Rows per tile
	TileCols    int // Columns per tile
	Replication int // Weight replication factor
}

// MappingConfig configures weight mapping strategy
type MappingConfig struct {
	CrossbarSize     int     // Size of each crossbar
	WeightBits       int     // Bits per weight
	InputBits        int     // Bits per input
	UseWeightSharing bool    // Enable weight sharing
	UseReplication   bool    // Enable weight replication
}

// LayerMapping represents how a layer is mapped to CIM hardware
type LayerMapping struct {
	LayerName      string
	TilingStrategy TilingStrategy
	CrossbarIDs    []int           // IDs of crossbars used
	WeightTiles    [][][]float64   // Tiled weight matrices
	DataFlow       string          // "weight-stationary", "input-stationary", "output-stationary"
	EstimatedCycles int
	EstimatedEnergy float64
}

// CIMCompiler compiles DNN models for CIM accelerators
type CIMCompiler struct {
	Config         MappingConfig
	Graph          *ModelGraph
	LayerMappings  []*LayerMapping
	CrossbarCount  int
	TotalCycles    int
	TotalEnergy    float64
}

// NewCIMCompiler creates a CIM compiler
func NewCIMCompiler(config MappingConfig) *CIMCompiler {
	return &CIMCompiler{
		Config:        config,
		LayerMappings: make([]*LayerMapping, 0),
	}
}

// Compile compiles the model graph for CIM execution
func (cc *CIMCompiler) Compile(graph *ModelGraph) error {
	cc.Graph = graph
	sortedNodes := graph.TopoSort()

	for _, node := range sortedNodes {
		mapping, err := cc.compileNode(node)
		if err != nil {
			return fmt.Errorf("failed to compile node %s: %v", node.Name, err)
		}
		if mapping != nil {
			cc.LayerMappings = append(cc.LayerMappings, mapping)
			cc.TotalCycles += mapping.EstimatedCycles
			cc.TotalEnergy += mapping.EstimatedEnergy
		}
	}

	return nil
}

// compileNode compiles a single node
func (cc *CIMCompiler) compileNode(node *NodeDef) (*LayerMapping, error) {
	switch node.OpType {
	case OpMatMul, OpConv2D:
		return cc.compileMatMulNode(node)
	case OpAttention:
		return cc.compileAttentionNode(node)
	default:
		// Non-mappable operations (ReLU, Softmax, etc.) run on digital processor
		return nil, nil
	}
}

// compileMatMulNode compiles MatMul/Conv2D for CIM
func (cc *CIMCompiler) compileMatMulNode(node *NodeDef) (*LayerMapping, error) {
	if len(node.Weights) == 0 {
		return nil, nil
	}

	rows := len(node.Weights)
	cols := len(node.Weights[0])

	// Compute tiling strategy
	tiling := cc.computeTiling(rows, cols)

	// Tile the weights
	weightTiles := cc.tileWeights(node.Weights, tiling)

	// Assign crossbars
	numCrossbars := tiling.RowTiles * tiling.ColTiles
	crossbarIDs := make([]int, numCrossbars)
	for i := range crossbarIDs {
		crossbarIDs[i] = cc.CrossbarCount
		cc.CrossbarCount++
	}

	// Estimate performance
	cycles := tiling.RowTiles * tiling.ColTiles * cc.Config.CrossbarSize
	energy := float64(cycles) * 1e-12 // ~1 pJ per cycle

	return &LayerMapping{
		LayerName:       node.Name,
		TilingStrategy:  tiling,
		CrossbarIDs:     crossbarIDs,
		WeightTiles:     weightTiles,
		DataFlow:        "weight-stationary",
		EstimatedCycles: cycles,
		EstimatedEnergy: energy,
	}, nil
}

// compileAttentionNode compiles attention for CIM
func (cc *CIMCompiler) compileAttentionNode(node *NodeDef) (*LayerMapping, error) {
	// Attention requires multiple crossbars: Q, K, V projections + output
	numHeads := 1
	if heads, ok := node.Attributes["num_heads"].(int); ok {
		numHeads = heads
	}

	numCrossbars := 4 * numHeads // Q, K, V, O per head
	crossbarIDs := make([]int, numCrossbars)
	for i := range crossbarIDs {
		crossbarIDs[i] = cc.CrossbarCount
		cc.CrossbarCount++
	}

	return &LayerMapping{
		LayerName:       node.Name,
		CrossbarIDs:     crossbarIDs,
		DataFlow:        "output-stationary",
		EstimatedCycles: cc.Config.CrossbarSize * cc.Config.CrossbarSize * numHeads,
		EstimatedEnergy: float64(numCrossbars) * float64(cc.Config.CrossbarSize) * 1e-12,
	}, nil
}

// computeTiling determines optimal tiling strategy
func (cc *CIMCompiler) computeTiling(rows, cols int) TilingStrategy {
	crossbarSize := cc.Config.CrossbarSize

	rowTiles := (rows + crossbarSize - 1) / crossbarSize
	colTiles := (cols + crossbarSize - 1) / crossbarSize

	return TilingStrategy{
		RowTiles: rowTiles,
		ColTiles: colTiles,
		TileRows: crossbarSize,
		TileCols: crossbarSize,
	}
}

// tileWeights partitions weights into tiles
func (cc *CIMCompiler) tileWeights(weights [][]float64, tiling TilingStrategy) [][][]float64 {
	tiles := make([][][]float64, tiling.RowTiles*tiling.ColTiles)

	for rt := 0; rt < tiling.RowTiles; rt++ {
		for ct := 0; ct < tiling.ColTiles; ct++ {
			tileIdx := rt*tiling.ColTiles + ct
			tile := make([][]float64, tiling.TileRows)

			for i := 0; i < tiling.TileRows; i++ {
				tile[i] = make([]float64, tiling.TileCols)
				srcRow := rt*tiling.TileRows + i
				if srcRow >= len(weights) {
					continue
				}

				for j := 0; j < tiling.TileCols; j++ {
					srcCol := ct*tiling.TileCols + j
					if srcCol < len(weights[srcRow]) {
						tile[i][j] = weights[srcRow][srcCol]
					}
				}
			}
			tiles[tileIdx] = tile
		}
	}

	return tiles
}

// GenerateSchedule creates execution schedule for the compiled model
func (cc *CIMCompiler) GenerateSchedule() []ScheduleEntry {
	schedule := make([]ScheduleEntry, 0)

	for _, mapping := range cc.LayerMappings {
		for tileIdx, crossbarID := range mapping.CrossbarIDs {
			entry := ScheduleEntry{
				LayerName:   mapping.LayerName,
				TileIndex:   tileIdx,
				CrossbarID:  crossbarID,
				StartCycle:  len(schedule),
				DataFlow:    mapping.DataFlow,
			}
			schedule = append(schedule, entry)
		}
	}

	return schedule
}

// ScheduleEntry represents a single scheduled operation
type ScheduleEntry struct {
	LayerName   string
	TileIndex   int
	CrossbarID  int
	StartCycle  int
	DataFlow    string
}

// =============================================================================
// Part 6: Dataflow Scheduler
// =============================================================================

// DataflowType represents dataflow strategies
type DataflowType string

const (
	WeightStationary DataflowType = "weight-stationary"
	InputStationary  DataflowType = "input-stationary"
	OutputStationary DataflowType = "output-stationary"
)

// DataflowScheduler optimizes dataflow for CIM execution
type DataflowScheduler struct {
	NumCrossbars     int
	CrossbarSize     int
	BufferSizeKB     int
	BandwidthGBps    float64
	DataflowStrategy DataflowType
}

// NewDataflowScheduler creates a dataflow scheduler
func NewDataflowScheduler(numCrossbars, crossbarSize, bufferKB int, bandwidth float64) *DataflowScheduler {
	return &DataflowScheduler{
		NumCrossbars:     numCrossbars,
		CrossbarSize:     crossbarSize,
		BufferSizeKB:     bufferKB,
		BandwidthGBps:    bandwidth,
		DataflowStrategy: WeightStationary,
	}
}

// OptimizeDataflow selects optimal dataflow for given layer
func (ds *DataflowScheduler) OptimizeDataflow(inputSize, outputSize, weightSize int) DataflowType {
	// Calculate data reuse potential for each strategy
	weightReuse := float64(inputSize) // Each weight used for all inputs
	inputReuse := float64(outputSize)  // Each input used for all outputs
	outputReuse := float64(weightSize) // Each output accumulated from all weights

	// Select strategy with maximum reuse
	maxReuse := weightReuse
	bestStrategy := WeightStationary

	if inputReuse > maxReuse {
		maxReuse = inputReuse
		bestStrategy = InputStationary
	}
	if outputReuse > maxReuse {
		bestStrategy = OutputStationary
	}

	return bestStrategy
}

// EstimateLatency estimates latency for a layer execution
func (ds *DataflowScheduler) EstimateLatency(mapping *LayerMapping) float64 {
	// Compute cycles based on tiling
	computeCycles := float64(mapping.TilingStrategy.RowTiles * mapping.TilingStrategy.ColTiles * ds.CrossbarSize)

	// Memory access cycles
	dataSize := float64(mapping.TilingStrategy.TileRows * mapping.TilingStrategy.TileCols * 4) // 4 bytes per float
	memCycles := dataSize / (ds.BandwidthGBps * 1e9 / 1e9) // ns

	return computeCycles + memCycles
}

// =============================================================================
// Part 7: Resource Allocator (CMSwitch-style Mode Switching)
// =============================================================================

// CIMMode represents operating mode of CIM array
type CIMMode string

const (
	ComputeMode CIMMode = "compute" // In-memory computing
	MemoryMode  CIMMode = "memory"  // Standard memory access
)

// ModeSwitchController manages compute/memory mode switching
type ModeSwitchController struct {
	NumArrays        int
	ArrayModes       []CIMMode
	SwitchLatency_ns float64
	ModeSwitchCount  int
}

// NewModeSwitchController creates a mode switch controller
func NewModeSwitchController(numArrays int) *ModeSwitchController {
	modes := make([]CIMMode, numArrays)
	for i := range modes {
		modes[i] = MemoryMode // Default to memory mode
	}

	return &ModeSwitchController{
		NumArrays:        numArrays,
		ArrayModes:       modes,
		SwitchLatency_ns: 10.0, // 10 ns switch latency
	}
}

// SwitchMode switches an array to the specified mode
func (mc *ModeSwitchController) SwitchMode(arrayID int, mode CIMMode) {
	if arrayID < mc.NumArrays && mc.ArrayModes[arrayID] != mode {
		mc.ArrayModes[arrayID] = mode
		mc.ModeSwitchCount++
	}
}

// GetComputeArrays returns IDs of arrays in compute mode
func (mc *ModeSwitchController) GetComputeArrays() []int {
	var ids []int
	for i, mode := range mc.ArrayModes {
		if mode == ComputeMode {
			ids = append(ids, i)
		}
	}
	return ids
}

// GetMemoryArrays returns IDs of arrays in memory mode
func (mc *ModeSwitchController) GetMemoryArrays() []int {
	var ids []int
	for i, mode := range mc.ArrayModes {
		if mode == MemoryMode {
			ids = append(ids, i)
		}
	}
	return ids
}

// OptimizeAllocation optimizes mode allocation for a workload
func (mc *ModeSwitchController) OptimizeAllocation(computeDemand, memoryDemand int) {
	// Simple heuristic: allocate based on demand ratio
	totalDemand := computeDemand + memoryDemand
	if totalDemand == 0 {
		return
	}

	computeArrays := mc.NumArrays * computeDemand / totalDemand
	for i := 0; i < mc.NumArrays; i++ {
		if i < computeArrays {
			mc.SwitchMode(i, ComputeMode)
		} else {
			mc.SwitchMode(i, MemoryMode)
		}
	}
}

// =============================================================================
// Part 8: Performance Model
// =============================================================================

// PerformanceModel estimates CIM system performance
type PerformanceModel struct {
	ProcessNode_nm     int     // Technology node
	ArraySize          int     // Crossbar array size
	NumArrays          int     // Number of arrays
	Frequency_MHz      float64 // Operating frequency
	SupplyVoltage_V    float64 // Supply voltage
	LeakagePower_mW    float64 // Static power
}

// NewPerformanceModel creates a performance model
func NewPerformanceModel(processNode, arraySize, numArrays int) *PerformanceModel {
	// Scale parameters based on process node
	freqScale := 28.0 / float64(processNode)
	voltageScale := math.Sqrt(float64(processNode) / 28.0)

	return &PerformanceModel{
		ProcessNode_nm:  processNode,
		ArraySize:       arraySize,
		NumArrays:       numArrays,
		Frequency_MHz:   500 * freqScale,
		SupplyVoltage_V: 0.9 * voltageScale,
		LeakagePower_mW: 0.1 * float64(numArrays),
	}
}

// EstimateThroughput estimates TOPS for a given workload
func (pm *PerformanceModel) EstimateThroughput(opsPerCycle float64) float64 {
	return opsPerCycle * pm.Frequency_MHz * 1e6 / 1e12 // TOPS
}

// EstimateEnergy estimates energy per inference (Joules)
func (pm *PerformanceModel) EstimateEnergy(numOps int) float64 {
	// Energy per MAC scales with V² and process node
	energyPerMAC_fJ := 0.5 * pm.SupplyVoltage_V * pm.SupplyVoltage_V * float64(pm.ProcessNode_nm) / 28.0
	return float64(numOps) * energyPerMAC_fJ * 1e-15
}

// EstimateArea estimates chip area in mm²
func (pm *PerformanceModel) EstimateArea() float64 {
	// Area per array scales with process node squared
	areaPerArray_mm2 := 0.1 * math.Pow(float64(pm.ProcessNode_nm)/28.0, 2)
	return float64(pm.NumArrays) * areaPerArray_mm2
}

// EstimateTOPSW estimates TOPS/W efficiency
func (pm *PerformanceModel) EstimateTOPSW(throughput_TOPS, power_W float64) float64 {
	if power_W <= 0 {
		return 0
	}
	return throughput_TOPS / power_W
}

// =============================================================================
// Part 9: Deployment Pipeline
// =============================================================================

// DeploymentConfig configures the deployment pipeline
type DeploymentConfig struct {
	TargetPlatform    string // "simulation", "fpga", "asic"
	OptimizationLevel int    // 0-3
	EnableQuantization bool
	QuantizationBits  int
	EnablePruning     bool
	PruningThreshold  float64
	ValidateOutput    bool
}

// DeploymentPipeline orchestrates model deployment to CIM
type DeploymentPipeline struct {
	Config       DeploymentConfig
	Compiler     *CIMCompiler
	Simulator    *FASTSimulator
	Graph        *ModelGraph
	DeployedModel *DeployedModel
}

// NewDeploymentPipeline creates a deployment pipeline
func NewDeploymentPipeline(config DeploymentConfig) *DeploymentPipeline {
	compilerConfig := MappingConfig{
		CrossbarSize:     64,
		WeightBits:       6,
		InputBits:        8,
		UseWeightSharing: true,
	}

	fastConfig := FASTConfig{
		CrossbarRows:      64,
		CrossbarCols:      64,
		NonIdealityCfg:    DefaultNonIdealityConfig(),
		EnableSparsity:    true,
		SparsityThreshold: 0.01,
	}

	return &DeploymentPipeline{
		Config:    config,
		Compiler:  NewCIMCompiler(compilerConfig),
		Simulator: NewFASTSimulator(fastConfig),
	}
}

// Deploy deploys a model through the full pipeline
func (dp *DeploymentPipeline) Deploy(graph *ModelGraph) (*DeployedModel, error) {
	dp.Graph = graph

	// Step 1: Optimization passes
	if dp.Config.OptimizationLevel > 0 {
		dp.optimizeGraph()
	}

	// Step 2: Quantization
	if dp.Config.EnableQuantization {
		dp.quantizeWeights()
	}

	// Step 3: Pruning
	if dp.Config.EnablePruning {
		dp.pruneWeights()
	}

	// Step 4: Compile for CIM
	if err := dp.Compiler.Compile(graph); err != nil {
		return nil, fmt.Errorf("compilation failed: %v", err)
	}

	// Step 5: Map to simulator
	for _, mapping := range dp.Compiler.LayerMappings {
		for _, tile := range mapping.WeightTiles {
			dp.Simulator.MapWeights(tile)
		}
	}

	// Step 6: Validation
	if dp.Config.ValidateOutput {
		dp.validateDeployment()
	}

	// Create deployed model
	dp.DeployedModel = &DeployedModel{
		Name:          graph.Name,
		NumCrossbars:  dp.Compiler.CrossbarCount,
		TotalCycles:   dp.Compiler.TotalCycles,
		TotalEnergy:   dp.Compiler.TotalEnergy,
		LayerMappings: dp.Compiler.LayerMappings,
		Schedule:      dp.Compiler.GenerateSchedule(),
	}

	return dp.DeployedModel, nil
}

// optimizeGraph applies graph-level optimizations
func (dp *DeploymentPipeline) optimizeGraph() {
	// Operator fusion
	if dp.Config.OptimizationLevel >= 1 {
		dp.fuseOperators()
	}

	// Layer reordering
	if dp.Config.OptimizationLevel >= 2 {
		dp.reorderLayers()
	}

	// Memory optimization
	if dp.Config.OptimizationLevel >= 3 {
		dp.optimizeMemory()
	}
}

func (dp *DeploymentPipeline) fuseOperators() {
	// Fuse Conv+BatchNorm+ReLU patterns
	for i := 0; i < len(dp.Graph.Nodes)-2; i++ {
		if dp.Graph.Nodes[i].OpType == OpConv2D &&
			dp.Graph.Nodes[i+1].OpType == OpBatchNorm &&
			dp.Graph.Nodes[i+2].OpType == OpReLU {
			// Mark for fusion (simplified)
			dp.Graph.Nodes[i].Attributes["fused"] = "conv_bn_relu"
		}
	}
}

func (dp *DeploymentPipeline) reorderLayers() {
	// No-op for now (preserve topological order)
}

func (dp *DeploymentPipeline) optimizeMemory() {
	// Estimate buffer requirements
	for _, node := range dp.Graph.Nodes {
		if len(node.Weights) > 0 {
			bufferSize := len(node.Weights) * len(node.Weights[0]) * 4 // bytes
			node.Attributes["buffer_size"] = bufferSize
		}
	}
}

// quantizeWeights applies quantization to weights
func (dp *DeploymentPipeline) quantizeWeights() {
	bits := dp.Config.QuantizationBits
	levels := float64(1 << bits)

	for _, node := range dp.Graph.Nodes {
		for i := range node.Weights {
			// Find range
			minVal, maxVal := node.Weights[i][0], node.Weights[i][0]
			for _, w := range node.Weights[i] {
				if w < minVal {
					minVal = w
				}
				if w > maxVal {
					maxVal = w
				}
			}

			scale := (maxVal - minVal) / (levels - 1)
			if scale == 0 {
				scale = 1
			}

			// Quantize
			for j := range node.Weights[i] {
				normalized := (node.Weights[i][j] - minVal) / scale
				quantized := math.Round(normalized)
				node.Weights[i][j] = quantized*scale + minVal
			}
		}
	}
}

// pruneWeights removes small weights
func (dp *DeploymentPipeline) pruneWeights() {
	threshold := dp.Config.PruningThreshold

	for _, node := range dp.Graph.Nodes {
		for i := range node.Weights {
			for j := range node.Weights[i] {
				if math.Abs(node.Weights[i][j]) < threshold {
					node.Weights[i][j] = 0
				}
			}
		}
	}
}

// validateDeployment checks deployment correctness
func (dp *DeploymentPipeline) validateDeployment() {
	// Simple validation: check all layers were mapped
	for _, node := range dp.Graph.Nodes {
		if node.OpType == OpMatMul || node.OpType == OpConv2D {
			found := false
			for _, mapping := range dp.Compiler.LayerMappings {
				if mapping.LayerName == node.Name {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("Warning: Layer %s not mapped to hardware\n", node.Name)
			}
		}
	}
}

// DeployedModel represents a model deployed to CIM hardware
type DeployedModel struct {
	Name          string
	NumCrossbars  int
	TotalCycles   int
	TotalEnergy   float64
	LayerMappings []*LayerMapping
	Schedule      []ScheduleEntry
}

// ToJSON serializes the deployed model
func (dm *DeployedModel) ToJSON() ([]byte, error) {
	return json.MarshalIndent(dm, "", "  ")
}

// =============================================================================
// Part 10: Benchmark Suite
// =============================================================================

// BenchmarkSuite runs comprehensive benchmarks
type BenchmarkSuite struct {
	Results map[string]BenchmarkResult
}

// BenchmarkResult stores benchmark metrics
type BenchmarkResult struct {
	ModelName     string
	Accuracy      float64
	Throughput    float64
	EnergyPerInf  float64
	TOPSW         float64
	NonIdealDrop  float64 // Accuracy drop due to non-idealities
}

// NewBenchmarkSuite creates a benchmark suite
func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{
		Results: make(map[string]BenchmarkResult),
	}
}

// RunNonIdealityBenchmark benchmarks accuracy under non-idealities
func (bs *BenchmarkSuite) RunNonIdealityBenchmark(modelName string, weights [][]float64) BenchmarkResult {
	// Ideal accuracy
	idealSim := NewFASTSimulator(FASTConfig{
		CrossbarRows:   len(weights),
		CrossbarCols:   len(weights[0]),
		NonIdealityCfg: NonIdealityConfig{EnableIRDrop: false, EnableVariation: false, EnableRetention: false},
	})
	idealSim.MapWeights(weights)

	// Non-ideal accuracy
	nonIdealSim := NewFASTSimulator(FASTConfig{
		CrossbarRows:   len(weights),
		CrossbarCols:   len(weights[0]),
		NonIdealityCfg: DefaultNonIdealityConfig(),
	})
	nonIdealSim.MapWeights(weights)

	// Generate test data
	rng := rand.New(rand.NewSource(42))
	testInputs := make([][]float64, 100)
	for i := range testInputs {
		testInputs[i] = make([]float64, len(weights[0]))
		for j := range testInputs[i] {
			testInputs[i][j] = rng.Float64()*2 - 1
		}
	}

	// Compare outputs
	var totalError float64
	for _, input := range testInputs {
		idealOut := idealSim.SimulateInference(input)
		nonIdealOut := nonIdealSim.SimulateInference(input)

		for j := range idealOut {
			totalError += math.Abs(idealOut[j] - nonIdealOut[j])
		}
	}
	avgError := totalError / float64(len(testInputs)*len(weights))

	result := BenchmarkResult{
		ModelName:    modelName,
		Accuracy:     100 - avgError*100,
		NonIdealDrop: avgError * 100,
		Throughput:   nonIdealSim.Throughput_GOPS,
		EnergyPerInf: nonIdealSim.EnergyConsumed_J / float64(len(testInputs)),
	}

	bs.Results[modelName] = result
	return result
}

// CompareConfigurations compares different non-ideality configurations
func (bs *BenchmarkSuite) CompareConfigurations(weights [][]float64, configs []NonIdealityConfig) map[string]float64 {
	comparison := make(map[string]float64)

	for i, config := range configs {
		sim := NewFASTSimulator(FASTConfig{
			CrossbarRows:   len(weights),
			CrossbarCols:   len(weights[0]),
			NonIdealityCfg: config,
		})
		sim.MapWeights(weights)

		// Run test
		rng := rand.New(rand.NewSource(42))
		input := make([]float64, len(weights[0]))
		for j := range input {
			input[j] = rng.Float64()*2 - 1
		}

		output := sim.SimulateInference(input)
		var sumOutput float64
		for _, v := range output {
			sumOutput += math.Abs(v)
		}

		configName := fmt.Sprintf("config_%d", i)
		comparison[configName] = sumOutput / float64(len(output))
	}

	return comparison
}

// GenerateReport creates a benchmark report
func (bs *BenchmarkSuite) GenerateReport() string {
	var report string
	report += "=== CIM Co-simulation Benchmark Report ===\n\n"

	// Sort results by accuracy
	type sortEntry struct {
		name   string
		result BenchmarkResult
	}
	var entries []sortEntry
	for name, result := range bs.Results {
		entries = append(entries, sortEntry{name, result})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].result.Accuracy > entries[j].result.Accuracy
	})

	for _, entry := range entries {
		result := entry.result
		report += fmt.Sprintf("Model: %s\n", result.ModelName)
		report += fmt.Sprintf("  Accuracy: %.2f%%\n", result.Accuracy)
		report += fmt.Sprintf("  Non-ideal drop: %.2f%%\n", result.NonIdealDrop)
		report += fmt.Sprintf("  Throughput: %.2f GOPS\n", result.Throughput)
		report += fmt.Sprintf("  Energy/inf: %.2e J\n", result.EnergyPerInf)
		report += "\n"
	}

	return report
}

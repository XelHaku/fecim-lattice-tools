// gnn_endurance.go - GNN CIM Acceleration and Ferroelectric Endurance Enhancement
// IronLattice Ferroelectric CIM Educational Simulation
//
// This module implements:
// 1. Graph Neural Network (GNN) Acceleration:
//    - Sparse adjacency matrix handling (CSR format)
//    - Aggregation and combination kernels
//    - GNN-PIM architecture simulation
//    - NEM-GNN sparsity-aware approaches
//    - TCAM-GNN crossbar mapping
//
// 2. Ferroelectric Endurance Enhancement:
//    - Wake-up, fatigue, and imprint modeling
//    - Oxygen vacancy dynamics simulation
//    - Interfacial layer engineering
//    - Write-verify algorithms
//    - Defect shielding layers (DSL)
//
// References:
// - NEM-GNN (ACM TACO 2024): DAC/ADC-less GNN accelerator
// - TCAM-GNN: Crossbar-based GNN training
// - Nature Comms (2025): Fatigue-free HZO via interfacial design
// - JAP (2024): HZO superlattice oxygen vacancy dynamics
// - ACS AEM (2024): Enhanced endurance via Vo tailoring

package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// PART 1: GRAPH NEURAL NETWORK CIM ACCELERATION
// =============================================================================

// SparseFormat defines the sparse matrix storage format
type SparseFormat int

const (
	FormatCSR SparseFormat = iota // Compressed Sparse Row
	FormatCSC                     // Compressed Sparse Column
	FormatCOO                     // Coordinate list
	FormatDense                   // Dense (for small graphs)
)

// CSRMatrix represents a sparse matrix in CSR format
type CSRMatrix struct {
	NumRows    int
	NumCols    int
	NumNonzero int
	RowPtr     []int     // Length: NumRows + 1
	ColIdx     []int     // Length: NumNonzero
	Values     []float64 // Length: NumNonzero
	Sparsity   float64   // Fraction of zeros (e.g., 0.997 = 99.7%)
}

// NewCSRMatrix creates a CSR matrix from dense representation
func NewCSRMatrix(dense [][]float64) *CSRMatrix {
	numRows := len(dense)
	if numRows == 0 {
		return &CSRMatrix{}
	}
	numCols := len(dense[0])

	rowPtr := make([]int, numRows+1)
	colIdx := make([]int, 0)
	values := make([]float64, 0)

	nnz := 0
	for i := 0; i < numRows; i++ {
		rowPtr[i] = nnz
		for j := 0; j < numCols; j++ {
			if dense[i][j] != 0 {
				colIdx = append(colIdx, j)
				values = append(values, dense[i][j])
				nnz++
			}
		}
	}
	rowPtr[numRows] = nnz

	sparsity := 1.0 - float64(nnz)/float64(numRows*numCols)

	return &CSRMatrix{
		NumRows:    numRows,
		NumCols:    numCols,
		NumNonzero: nnz,
		RowPtr:     rowPtr,
		ColIdx:     colIdx,
		Values:     values,
		Sparsity:   sparsity,
	}
}

// GenerateRandomGraph creates a random sparse adjacency matrix
func GenerateRandomGraph(numNodes int, avgDegree float64) *CSRMatrix {
	edgeProb := avgDegree / float64(numNodes)

	rowPtr := make([]int, numNodes+1)
	colIdx := make([]int, 0)
	values := make([]float64, 0)

	nnz := 0
	for i := 0; i < numNodes; i++ {
		rowPtr[i] = nnz
		for j := 0; j < numNodes; j++ {
			if rand.Float64() < edgeProb {
				colIdx = append(colIdx, j)
				values = append(values, 1.0)
				nnz++
			}
		}
	}
	rowPtr[numNodes] = nnz

	sparsity := 1.0 - float64(nnz)/float64(numNodes*numNodes)

	return &CSRMatrix{
		NumRows:    numNodes,
		NumCols:    numNodes,
		NumNonzero: nnz,
		RowPtr:     rowPtr,
		ColIdx:     colIdx,
		Values:     values,
		Sparsity:   sparsity,
	}
}

// GNNLayerType defines the type of GNN layer
type GNNLayerType int

const (
	GNNLayerGCN  GNNLayerType = iota // Graph Convolutional Network
	GNNLayerGAT                      // Graph Attention Network
	GNNLayerSAGE                     // GraphSAGE
	GNNLayerGIN                      // Graph Isomorphism Network
)

// GNNLayerConfig configures a GNN layer
type GNNLayerConfig struct {
	Type          GNNLayerType
	InputDim      int
	OutputDim     int
	NumHeads      int     // For GAT
	Aggregation   string  // "mean", "sum", "max"
	Activation    string  // "relu", "leaky_relu", "elu"
	Dropout       float64
	UseNormalize  bool
}

// GNNLayer represents a single GNN layer
type GNNLayer struct {
	Config   *GNNLayerConfig
	Weights  [][]float64 // [InputDim x OutputDim]
	Bias     []float64   // [OutputDim]

	// For GAT
	AttentionWeights [][]float64
}

// NewGNNLayer creates a new GNN layer
func NewGNNLayer(config *GNNLayerConfig) *GNNLayer {
	layer := &GNNLayer{
		Config: config,
	}

	// Initialize weights (Xavier initialization)
	layer.Weights = make([][]float64, config.InputDim)
	scale := math.Sqrt(2.0 / float64(config.InputDim+config.OutputDim))
	for i := 0; i < config.InputDim; i++ {
		layer.Weights[i] = make([]float64, config.OutputDim)
		for j := 0; j < config.OutputDim; j++ {
			layer.Weights[i][j] = rand.NormFloat64() * scale
		}
	}

	layer.Bias = make([]float64, config.OutputDim)

	return layer
}

// =============================================================================
// GNN CIM Accelerator
// =============================================================================

// GNNCIMConfig configures the GNN CIM accelerator
type GNNCIMConfig struct {
	CrossbarRows       int
	CrossbarCols       int
	NumCrossbars       int
	ADCBits            int
	DACBits            int
	UseDACADCLess      bool    // NEM-GNN style
	UseTCAM            bool    // TCAM-GNN style
	SparsityThreshold  float64 // Above this, use sparse mode
	AggregationMode    string  // "parallel", "sequential", "CAR"
	BufferSizeKB       int
}

// DefaultGNNCIMConfig returns default GNN CIM settings
func DefaultGNNCIMConfig() *GNNCIMConfig {
	return &GNNCIMConfig{
		CrossbarRows:      128,
		CrossbarCols:      128,
		NumCrossbars:      16,
		ADCBits:           6,
		DACBits:           6,
		UseDACADCLess:     false,
		UseTCAM:           false,
		SparsityThreshold: 0.95,
		AggregationMode:   "CAR", // Compute-as-soon-as-ready
		BufferSizeKB:      64,
	}
}

// GNNCIMAccelerator simulates a CIM accelerator for GNNs
type GNNCIMAccelerator struct {
	Config             *GNNCIMConfig
	AdjMatrix          *CSRMatrix
	FeatureMatrix      [][]float64
	Layers             []*GNNLayer

	// Crossbar state
	CrossbarWeights    [][][]float64 // [numCrossbars][rows][cols]
	CrossbarUtilization []float64

	// Performance metrics
	AggregationCycles  int64
	CombinationCycles  int64
	TotalEnergy        float64 // pJ
	Throughput         float64 // GOPs
}

// NewGNNCIMAccelerator creates a new GNN CIM accelerator
func NewGNNCIMAccelerator(config *GNNCIMConfig) *GNNCIMAccelerator {
	if config == nil {
		config = DefaultGNNCIMConfig()
	}

	acc := &GNNCIMAccelerator{
		Config:              config,
		CrossbarWeights:     make([][][]float64, config.NumCrossbars),
		CrossbarUtilization: make([]float64, config.NumCrossbars),
	}

	// Initialize crossbars
	for i := 0; i < config.NumCrossbars; i++ {
		acc.CrossbarWeights[i] = make([][]float64, config.CrossbarRows)
		for j := 0; j < config.CrossbarRows; j++ {
			acc.CrossbarWeights[i][j] = make([]float64, config.CrossbarCols)
		}
	}

	return acc
}

// SetGraph sets the graph adjacency matrix
func (acc *GNNCIMAccelerator) SetGraph(adj *CSRMatrix) {
	acc.AdjMatrix = adj
}

// SetFeatures sets the node feature matrix
func (acc *GNNCIMAccelerator) SetFeatures(features [][]float64) {
	acc.FeatureMatrix = features
}

// AddLayer adds a GNN layer
func (acc *GNNCIMAccelerator) AddLayer(layer *GNNLayer) {
	acc.Layers = append(acc.Layers, layer)
}

// MapWeightsToCrossbar maps layer weights to crossbar arrays
func (acc *GNNCIMAccelerator) MapWeightsToCrossbar(layerIdx int) error {
	if layerIdx >= len(acc.Layers) {
		return fmt.Errorf("layer index out of range")
	}

	layer := acc.Layers[layerIdx]
	inputDim := layer.Config.InputDim
	outputDim := layer.Config.OutputDim

	xbarRows := acc.Config.CrossbarRows
	xbarCols := acc.Config.CrossbarCols

	crossbarIdx := 0
	totalCells := 0
	usedCells := 0

	for rowStart := 0; rowStart < inputDim; rowStart += xbarRows {
		rowEnd := rowStart + xbarRows
		if rowEnd > inputDim {
			rowEnd = inputDim
		}

		for colStart := 0; colStart < outputDim; colStart += xbarCols {
			colEnd := colStart + xbarCols
			if colEnd > outputDim {
				colEnd = outputDim
			}

			if crossbarIdx >= acc.Config.NumCrossbars {
				return fmt.Errorf("not enough crossbars")
			}

			// Map weights
			for i := rowStart; i < rowEnd; i++ {
				for j := colStart; j < colEnd; j++ {
					acc.CrossbarWeights[crossbarIdx][i-rowStart][j-colStart] = layer.Weights[i][j]
					usedCells++
				}
			}

			totalCells += xbarRows * xbarCols
			acc.CrossbarUtilization[crossbarIdx] = float64(rowEnd-rowStart) * float64(colEnd-colStart) / float64(xbarRows*xbarCols)
			crossbarIdx++
		}
	}

	return nil
}

// Aggregation performs sparse neighborhood aggregation
func (acc *GNNCIMAccelerator) Aggregation(features [][]float64, mode string) [][]float64 {
	numNodes := acc.AdjMatrix.NumRows
	featureDim := len(features[0])

	aggregated := make([][]float64, numNodes)
	for i := range aggregated {
		aggregated[i] = make([]float64, featureDim)
	}

	// CSR-based sparse aggregation
	for i := 0; i < numNodes; i++ {
		rowStart := acc.AdjMatrix.RowPtr[i]
		rowEnd := acc.AdjMatrix.RowPtr[i+1]
		numNeighbors := rowEnd - rowStart

		if numNeighbors == 0 {
			continue
		}

		for ptr := rowStart; ptr < rowEnd; ptr++ {
			j := acc.AdjMatrix.ColIdx[ptr]
			edgeWeight := acc.AdjMatrix.Values[ptr]

			for f := 0; f < featureDim; f++ {
				switch mode {
				case "sum":
					aggregated[i][f] += features[j][f] * edgeWeight
				case "mean":
					aggregated[i][f] += features[j][f] * edgeWeight / float64(numNeighbors)
				case "max":
					if features[j][f]*edgeWeight > aggregated[i][f] {
						aggregated[i][f] = features[j][f] * edgeWeight
					}
				}
			}
			acc.AggregationCycles++
		}
	}

	return aggregated
}

// Combination performs dense feature transformation (MVM on crossbar)
func (acc *GNNCIMAccelerator) Combination(features [][]float64, layerIdx int) [][]float64 {
	if layerIdx >= len(acc.Layers) {
		return nil
	}

	layer := acc.Layers[layerIdx]
	numNodes := len(features)
	outputDim := layer.Config.OutputDim

	output := make([][]float64, numNodes)
	for i := range output {
		output[i] = make([]float64, outputDim)
	}

	// MVM on crossbar (simulated)
	for n := 0; n < numNodes; n++ {
		for j := 0; j < outputDim; j++ {
			sum := layer.Bias[j]
			for i := 0; i < len(features[n]); i++ {
				sum += features[n][i] * layer.Weights[i][j]
			}
			output[n][j] = sum
			acc.CombinationCycles++
		}
	}

	// Apply activation
	for n := 0; n < numNodes; n++ {
		for j := 0; j < outputDim; j++ {
			switch layer.Config.Activation {
			case "relu":
				if output[n][j] < 0 {
					output[n][j] = 0
				}
			case "leaky_relu":
				if output[n][j] < 0 {
					output[n][j] *= 0.01
				}
			}
		}
	}

	return output
}

// Forward performs full GNN forward pass
func (acc *GNNCIMAccelerator) Forward() [][]float64 {
	features := acc.FeatureMatrix

	for i, layer := range acc.Layers {
		// Aggregation (memory-intensive, sparse)
		aggregated := acc.Aggregation(features, layer.Config.Aggregation)

		// Combination (compute-intensive, dense)
		features = acc.Combination(aggregated, i)
	}

	return features
}

// GetPerformanceMetrics returns performance metrics
func (acc *GNNCIMAccelerator) GetPerformanceMetrics() map[string]float64 {
	totalCycles := acc.AggregationCycles + acc.CombinationCycles
	aggFraction := float64(acc.AggregationCycles) / float64(totalCycles+1)

	// Energy model (simplified)
	energyPerAggOp := 0.5  // pJ (memory access dominated)
	energyPerCombOp := 0.1 // pJ (CIM is efficient)
	totalEnergy := float64(acc.AggregationCycles)*energyPerAggOp +
		float64(acc.CombinationCycles)*energyPerCombOp

	// Throughput
	clockFreq := 100e6 // 100 MHz
	throughput := float64(totalCycles) / (float64(totalCycles) / clockFreq)

	return map[string]float64{
		"aggregation_cycles":   float64(acc.AggregationCycles),
		"combination_cycles":   float64(acc.CombinationCycles),
		"aggregation_fraction": aggFraction * 100,
		"total_energy_pj":      totalEnergy,
		"throughput_gops":      throughput / 1e9,
	}
}

// =============================================================================
// NEM-GNN DAC/ADC-less Approach
// =============================================================================

// NEMGNNConfig configures NEM-GNN accelerator
type NEMGNNConfig struct {
	NumPIMCores        int
	NumHostCores       int
	LocalBufferKB      int
	BroadcastBandwidth float64 // GB/s
	CAREnabled         bool    // Compute-As-soon-as-Ready
	EarlyTermination   bool
}

// NEMGNNAccelerator implements NEM-GNN architecture
type NEMGNNAccelerator struct {
	Config               *NEMGNNConfig
	AggregationArray     [][]float64
	CombinationArray     [][]float64

	// Performance
	SpeedupOverReFLIP    float64
	EnergyEfficiency     float64 // vs baseline
	ComputeDensity       float64 // TOPS/mm²
}

// NewNEMGNNAccelerator creates a new NEM-GNN accelerator
func NewNEMGNNAccelerator(config *NEMGNNConfig) *NEMGNNAccelerator {
	if config == nil {
		config = &NEMGNNConfig{
			NumPIMCores:        64,
			NumHostCores:       4,
			LocalBufferKB:      32,
			BroadcastBandwidth: 100.0,
			CAREnabled:         true,
			EarlyTermination:   true,
		}
	}

	return &NEMGNNAccelerator{
		Config:            config,
		SpeedupOverReFLIP: 150.0, // 80-230× reported
		EnergyEfficiency:  1000.0, // 850-1134× reported
		ComputeDensity:    7.5,    // 7-8× reported
	}
}

// =============================================================================
// TCAM-GNN Crossbar Mapping
// =============================================================================

// TCAMGNNConfig configures TCAM-based GNN accelerator
type TCAMGNNConfig struct {
	TCAMRows           int
	TCAMCols           int
	NumTCAMArrays      int
	DynamicFixedPoint  bool
	AdaptiveDataReuse  bool
}

// TCAMGNNAccelerator implements TCAM-GNN architecture
type TCAMGNNAccelerator struct {
	Config              *TCAMGNNConfig
	TCAMArrays          [][][]int // Ternary content: -1, 0, 1

	// Performance vs baseline NN accelerator
	ComputeSpeedup      float64 // 4.25× reported
	EnergyEfficiency    float64 // 9.11× reported
}

// NewTCAMGNNAccelerator creates a new TCAM-GNN accelerator
func NewTCAMGNNAccelerator(config *TCAMGNNConfig) *TCAMGNNAccelerator {
	if config == nil {
		config = &TCAMGNNConfig{
			TCAMRows:          256,
			TCAMCols:          256,
			NumTCAMArrays:     8,
			DynamicFixedPoint: true,
			AdaptiveDataReuse: true,
		}
	}

	acc := &TCAMGNNAccelerator{
		Config:           config,
		TCAMArrays:       make([][][]int, config.NumTCAMArrays),
		ComputeSpeedup:   4.25,
		EnergyEfficiency: 9.11,
	}

	// Initialize TCAM arrays
	for i := 0; i < config.NumTCAMArrays; i++ {
		acc.TCAMArrays[i] = make([][]int, config.TCAMRows)
		for j := 0; j < config.TCAMRows; j++ {
			acc.TCAMArrays[i][j] = make([]int, config.TCAMCols)
		}
	}

	return acc
}

// =============================================================================
// PART 2: FERROELECTRIC ENDURANCE ENHANCEMENT
// =============================================================================

// FerroelectricState represents the device polarization state
type FerroelectricState int

const (
	StatePristine FerroelectricState = iota
	StateWakeUp
	StateMature
	StateFatigue
	StateFailure
)

// OxygenVacancyConfig configures oxygen vacancy dynamics
type OxygenVacancyConfig struct {
	InitialConcentration float64 // % of lattice sites
	FormationEnergy      float64 // eV
	MigrationBarrier     float64 // eV
	Temperature          float64 // K
	ElectricField        float64 // MV/cm
}

// DefaultOxygenVacancyConfig returns typical HZO parameters
func DefaultOxygenVacancyConfig() *OxygenVacancyConfig {
	return &OxygenVacancyConfig{
		InitialConcentration: 2.0,  // 2%
		FormationEnergy:      5.5,  // eV
		MigrationBarrier:     0.7,  // eV
		Temperature:          300,  // K (room temperature)
		ElectricField:        2.0,  // MV/cm
	}
}

// OxygenVacancyModel simulates Vo dynamics in HZO
type OxygenVacancyModel struct {
	Config             *OxygenVacancyConfig
	BulkConcentration  float64
	InterfaceConc      float64 // At electrode interface
	GrainBoundaryConc  float64

	// State tracking
	CycleCount         int64
	State              FerroelectricState
	PolarizationLoss   float64 // Fraction of 2Pr lost
}

// NewOxygenVacancyModel creates a new Vo model
func NewOxygenVacancyModel(config *OxygenVacancyConfig) *OxygenVacancyModel {
	if config == nil {
		config = DefaultOxygenVacancyConfig()
	}

	return &OxygenVacancyModel{
		Config:            config,
		BulkConcentration: config.InitialConcentration,
		InterfaceConc:     config.InitialConcentration * 1.5, // Higher at interface
		GrainBoundaryConc: config.InitialConcentration * 2.0, // Highest at GB
		State:             StatePristine,
	}
}

// SimulateCycle simulates one switching cycle
func (vo *OxygenVacancyModel) SimulateCycle() {
	vo.CycleCount++

	// Boltzmann factor for migration
	kB := 8.617e-5 // eV/K
	migrationProb := math.Exp(-vo.Config.MigrationBarrier / (kB * vo.Config.Temperature))

	// Field-enhanced migration
	fieldFactor := 1.0 + vo.Config.ElectricField*0.1
	effectiveMigration := migrationProb * fieldFactor

	// Update concentrations based on cycling phase
	switch vo.State {
	case StatePristine:
		// Wake-up: Vo migrate from interface to bulk
		if vo.CycleCount < 1000 {
			vo.InterfaceConc -= 0.001 * effectiveMigration
			vo.BulkConcentration += 0.0005 * effectiveMigration
		} else {
			vo.State = StateWakeUp
		}

	case StateWakeUp:
		// Transition to mature state
		if vo.CycleCount < 10000 {
			// Stabilization
			vo.InterfaceConc *= 0.9999
		} else {
			vo.State = StateMature
		}

	case StateMature:
		// Normal operation
		if vo.CycleCount > 1e8 {
			vo.State = StateFatigue
		}

	case StateFatigue:
		// Vo accumulate at interface, polarization degrades
		vo.InterfaceConc += 0.0001 * effectiveMigration
		vo.PolarizationLoss += 0.0001

		if vo.PolarizationLoss > 0.5 {
			vo.State = StateFailure
		}
	}
}

// GetEndurance returns the estimated endurance cycles
func (vo *OxygenVacancyModel) GetEndurance() int64 {
	// Based on Vo concentration
	if vo.Config.InitialConcentration < 2.0 {
		return 1e10 // >10^10 for low Vo
	} else if vo.Config.InitialConcentration < 3.0 {
		return 1e8 // 10^8 typical
	}
	return 1e6 // 10^6 for high Vo
}

// =============================================================================
// Endurance Enhancement Techniques
// =============================================================================

// EnduranceEnhancementType defines enhancement methods
type EnduranceEnhancementType int

const (
	EnhancementInterfacialLayer EnduranceEnhancementType = iota
	EnhancementDefectShielding
	EnhancementOxideChannel
	Enhancement2DChannel
	EnhancementElectrodeEngineering
	EnhancementSuperlattice
	EnhancementProgramOptimization
	EnhancementWriteVerify
)

// InterfacialLayerConfig configures IL engineering
type InterfacialLayerConfig struct {
	Material          string  // "SiO2", "Al2O3", "HfO2", "La2O3"
	Thickness         float64 // nm
	DielectricConst   float64
	FormationTemp     float64 // °C
}

// InterfacialLayerEnhancement models IL engineering for endurance
type InterfacialLayerEnhancement struct {
	Config              *InterfacialLayerConfig
	MemoryWindow        float64 // V
	EnduranceImprovement float64 // × factor
	RetentionYears      float64
}

// NewInterfacialLayerEnhancement creates an IL enhancement model
func NewInterfacialLayerEnhancement(config *InterfacialLayerConfig) *InterfacialLayerEnhancement {
	if config == nil {
		config = &InterfacialLayerConfig{
			Material:        "HfO2",
			Thickness:       2.0,
			DielectricConst: 20.0,
			FormationTemp:   300.0,
		}
	}

	enhancement := &InterfacialLayerEnhancement{
		Config: config,
	}

	// Calculate improvements based on IL material
	switch config.Material {
	case "HfO2":
		enhancement.MemoryWindow = 1.1
		enhancement.EnduranceImprovement = 1000.0 // 1000× reported for IGZO FeFET
		enhancement.RetentionYears = 10.0
	case "Al2O3":
		enhancement.MemoryWindow = 0.9
		enhancement.EnduranceImprovement = 100.0
		enhancement.RetentionYears = 10.0
	case "SiO2":
		enhancement.MemoryWindow = 0.7
		enhancement.EnduranceImprovement = 1.0 // Baseline
		enhancement.RetentionYears = 10.0
	case "La2O3":
		enhancement.MemoryWindow = 1.0
		enhancement.EnduranceImprovement = 500.0
		enhancement.RetentionYears = 10.0
	}

	return enhancement
}

// DefectShieldingLayerConfig configures DSL
type DefectShieldingLayerConfig struct {
	Enabled            bool
	LayerThickness     float64 // nm
	NumLayers          int     // 1 or 2 (both sides)
	Material           string  // "TiO2", "Al2O3"
}

// DefectShieldingEnhancement models DSL for endurance
type DefectShieldingEnhancement struct {
	Config               *DefectShieldingLayerConfig
	LeakageReduction     float64 // × factor
	EnduranceTarget      int64   // cycles
	OperatingTemp        float64 // °C
}

// NewDefectShieldingEnhancement creates a DSL enhancement model
func NewDefectShieldingEnhancement(config *DefectShieldingLayerConfig) *DefectShieldingEnhancement {
	if config == nil {
		config = &DefectShieldingLayerConfig{
			Enabled:        true,
			LayerThickness: 1.0,
			NumLayers:      2,
			Material:       "TiO2",
		}
	}

	enhancement := &DefectShieldingEnhancement{
		Config:        config,
		OperatingTemp: 125.0, // High temperature operation
	}

	if config.NumLayers == 2 {
		enhancement.EnduranceTarget = 1e13 // 10^13 cycles at 125°C
		enhancement.LeakageReduction = 100.0
	} else {
		enhancement.EnduranceTarget = 1e11
		enhancement.LeakageReduction = 10.0
	}

	return enhancement
}

// Channel2DConfig configures 2D channel material
type Channel2DConfig struct {
	Material       string  // "MoS2", "WS2", "MoSe2"
	NumLayers      int     // Monolayer, bilayer, etc.
	InterfaceLayer string  // "AlOx", "HfOx"
}

// Channel2DEnhancement models 2D channel for ultra-high endurance
type Channel2DEnhancement struct {
	Config             *Channel2DConfig
	ProjectedEndurance int64   // cycles
	CoerciveField      float64 // MV/cm
	MemoryWindow       float64 // V
}

// NewChannel2DEnhancement creates a 2D channel enhancement model
func NewChannel2DEnhancement(config *Channel2DConfig) *Channel2DEnhancement {
	if config == nil {
		config = &Channel2DConfig{
			Material:       "MoS2",
			NumLayers:      1,
			InterfaceLayer: "AlOx",
		}
	}

	enhancement := &Channel2DEnhancement{
		Config: config,
	}

	// MoS2 with AlOx IL shows extraordinary endurance
	if config.Material == "MoS2" && config.InterfaceLayer == "AlOx" {
		enhancement.ProjectedEndurance = 1e18 // >10^18 projected
		enhancement.CoerciveField = 1.5
		enhancement.MemoryWindow = 1.0
	} else {
		enhancement.ProjectedEndurance = 1e12
		enhancement.CoerciveField = 2.0
		enhancement.MemoryWindow = 0.8
	}

	return enhancement
}

// SuperlatticeConfig configures HZO superlattice
type SuperlatticeConfig struct {
	HfO2Thickness  float64 // nm per layer
	ZrO2Thickness  float64 // nm per layer
	NumPeriods     int
	TotalThickness float64 // nm
}

// SuperlatticeEnhancement models HZO superlattice for enhanced reliability
type SuperlatticeEnhancement struct {
	Config              *SuperlatticeConfig
	VoMigrationBarrier  float64 // Increased barrier in Hf layer
	LeakageReduction    float64
	EnduranceImprovement float64
	Polarization2Pr     float64 // µC/cm²
}

// NewSuperlatticeEnhancement creates a superlattice enhancement model
func NewSuperlatticeEnhancement(config *SuperlatticeConfig) *SuperlatticeEnhancement {
	if config == nil {
		config = &SuperlatticeConfig{
			HfO2Thickness:  2.0,
			ZrO2Thickness:  2.0,
			NumPeriods:     5,
			TotalThickness: 20.0,
		}
	}

	enhancement := &SuperlatticeEnhancement{
		Config:               config,
		VoMigrationBarrier:   1.0, // Higher in Hf layer
		LeakageReduction:     10.0,
		EnduranceImprovement: 10.0,
		Polarization2Pr:      30.0,
	}

	return enhancement
}

// =============================================================================
// Write-Verify Algorithm
// =============================================================================

// WriteVerifyConfig configures write-verify scheme
type WriteVerifyConfig struct {
	MaxIterations      int
	TargetWindow       float64 // V
	InitialPulseWidth  float64 // ns
	PulseWidthStep     float64 // ns increment
	VoltageStep        float64 // V increment
	VerifyThreshold    float64 // Acceptable margin
}

// DefaultWriteVerifyConfig returns default write-verify settings
func DefaultWriteVerifyConfig() *WriteVerifyConfig {
	return &WriteVerifyConfig{
		MaxIterations:     10,
		TargetWindow:      1.0,
		InitialPulseWidth: 50.0,
		PulseWidthStep:    10.0,
		VoltageStep:       0.1,
		VerifyThreshold:   0.9, // 90% of target
	}
}

// WriteVerifyController implements hybrid write-verify algorithm
type WriteVerifyController struct {
	Config            *WriteVerifyConfig
	CurrentPulseWidth float64
	CurrentVoltage    float64
	IterationCount    int
	SuccessRate       float64
	EnduranceGain     float64
}

// NewWriteVerifyController creates a new write-verify controller
func NewWriteVerifyController(config *WriteVerifyConfig) *WriteVerifyController {
	if config == nil {
		config = DefaultWriteVerifyConfig()
	}

	return &WriteVerifyController{
		Config:            config,
		CurrentPulseWidth: config.InitialPulseWidth,
		EnduranceGain:     2.0, // HWVA provides ~2× endurance
	}
}

// Write performs a write operation with verification
func (wv *WriteVerifyController) Write(targetState float64, currentState float64) (float64, bool) {
	for iter := 0; iter < wv.Config.MaxIterations; iter++ {
		wv.IterationCount++

		// Apply pulse
		newState := currentState + (targetState-currentState)*0.3

		// Verify
		if math.Abs(newState-targetState) < wv.Config.TargetWindow*(1-wv.Config.VerifyThreshold) {
			wv.SuccessRate = float64(wv.IterationCount-iter) / float64(wv.IterationCount)
			return newState, true
		}

		// Adjust pulse
		wv.CurrentPulseWidth += wv.Config.PulseWidthStep
		currentState = newState
	}

	return currentState, false
}

// =============================================================================
// Integrated Endurance Manager
// =============================================================================

// EnduranceManagerConfig configures the integrated endurance manager
type EnduranceManagerConfig struct {
	EnableIL           bool
	EnableDSL          bool
	Enable2DChannel    bool
	EnableSuperlattice bool
	EnableWriteVerify  bool
	TargetEndurance    int64
	TargetRetention    float64 // years
}

// EnduranceManager integrates multiple enhancement techniques
type EnduranceManager struct {
	Config             *EnduranceManagerConfig
	ILEnhancement      *InterfacialLayerEnhancement
	DSLEnhancement     *DefectShieldingEnhancement
	Channel2D          *Channel2DEnhancement
	Superlattice       *SuperlatticeEnhancement
	WriteVerify        *WriteVerifyController
	VoModel            *OxygenVacancyModel

	// Combined metrics
	ProjectedEndurance int64
	CombinedGain       float64
	ReliabilityScore   float64
}

// NewEnduranceManager creates a new endurance manager
func NewEnduranceManager(config *EnduranceManagerConfig) *EnduranceManager {
	if config == nil {
		config = &EnduranceManagerConfig{
			EnableIL:           true,
			EnableDSL:          true,
			Enable2DChannel:    false,
			EnableSuperlattice: true,
			EnableWriteVerify:  true,
			TargetEndurance:    1e12,
			TargetRetention:    10.0,
		}
	}

	manager := &EnduranceManager{
		Config:  config,
		VoModel: NewOxygenVacancyModel(nil),
	}

	// Initialize enabled enhancements
	if config.EnableIL {
		manager.ILEnhancement = NewInterfacialLayerEnhancement(&InterfacialLayerConfig{
			Material:  "HfO2",
			Thickness: 2.0,
		})
	}

	if config.EnableDSL {
		manager.DSLEnhancement = NewDefectShieldingEnhancement(&DefectShieldingLayerConfig{
			Enabled:   true,
			NumLayers: 2,
		})
	}

	if config.Enable2DChannel {
		manager.Channel2D = NewChannel2DEnhancement(nil)
	}

	if config.EnableSuperlattice {
		manager.Superlattice = NewSuperlatticeEnhancement(nil)
	}

	if config.EnableWriteVerify {
		manager.WriteVerify = NewWriteVerifyController(nil)
	}

	// Calculate combined improvement
	manager.calculateCombinedGain()

	return manager
}

// calculateCombinedGain computes the overall endurance improvement
func (em *EnduranceManager) calculateCombinedGain() {
	baseEndurance := int64(1e6) // 10^6 baseline
	combinedGain := 1.0

	if em.ILEnhancement != nil {
		combinedGain *= em.ILEnhancement.EnduranceImprovement
	}

	if em.DSLEnhancement != nil {
		// DSL provides absolute target, not multiplicative
		if em.DSLEnhancement.EnduranceTarget > baseEndurance*int64(combinedGain) {
			combinedGain = float64(em.DSLEnhancement.EnduranceTarget) / float64(baseEndurance)
		}
	}

	if em.Channel2D != nil {
		// 2D channel provides highest endurance
		combinedGain = math.Max(combinedGain, float64(em.Channel2D.ProjectedEndurance)/float64(baseEndurance))
	}

	if em.Superlattice != nil {
		combinedGain *= em.Superlattice.EnduranceImprovement
	}

	if em.WriteVerify != nil {
		combinedGain *= em.WriteVerify.EnduranceGain
	}

	em.CombinedGain = combinedGain
	em.ProjectedEndurance = int64(float64(baseEndurance) * combinedGain)

	// Reliability score (0-100)
	targetMet := float64(em.ProjectedEndurance) / float64(em.Config.TargetEndurance)
	if targetMet > 1.0 {
		targetMet = 1.0
	}
	em.ReliabilityScore = targetMet * 100
}

// GetEnhancementSummary returns a summary of all enhancements
func (em *EnduranceManager) GetEnhancementSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	summary["base_endurance"] = int64(1e6)
	summary["projected_endurance"] = em.ProjectedEndurance
	summary["combined_gain"] = em.CombinedGain
	summary["reliability_score"] = em.ReliabilityScore
	summary["target_met"] = em.ProjectedEndurance >= em.Config.TargetEndurance

	enhancements := make([]string, 0)
	if em.ILEnhancement != nil {
		enhancements = append(enhancements, fmt.Sprintf("IL(%s): %.0f×",
			em.ILEnhancement.Config.Material,
			em.ILEnhancement.EnduranceImprovement))
	}
	if em.DSLEnhancement != nil {
		enhancements = append(enhancements, fmt.Sprintf("DSL: 10^%d",
			int(math.Log10(float64(em.DSLEnhancement.EnduranceTarget)))))
	}
	if em.Channel2D != nil {
		enhancements = append(enhancements, fmt.Sprintf("2D(%s): 10^%d",
			em.Channel2D.Config.Material,
			int(math.Log10(float64(em.Channel2D.ProjectedEndurance)))))
	}
	if em.Superlattice != nil {
		enhancements = append(enhancements, fmt.Sprintf("SL: %.0f×",
			em.Superlattice.EnduranceImprovement))
	}
	if em.WriteVerify != nil {
		enhancements = append(enhancements, fmt.Sprintf("WV: %.0f×",
			em.WriteVerify.EnduranceGain))
	}

	summary["active_enhancements"] = enhancements

	return summary
}

// =============================================================================
// Fatigue and Wake-up Simulator
// =============================================================================

// FatigueSimConfig configures the fatigue simulation
type FatigueSimConfig struct {
	InitialPolarization float64 // µC/cm²
	CoerciveField       float64 // MV/cm
	FilmThickness       float64 // nm
	Temperature         float64 // K
	CyclingFrequency    float64 // Hz
}

// FatigueSimulator simulates ferroelectric fatigue behavior
type FatigueSimulator struct {
	Config              *FatigueSimConfig
	CurrentPolarization float64
	PolarizationHistory []float64
	CycleHistory        []int64
	Phase               string // "pristine", "wake-up", "mature", "fatigue"
	WakeUpCycles        int64
	FatigueCycles       int64
}

// NewFatigueSimulator creates a new fatigue simulator
func NewFatigueSimulator(config *FatigueSimConfig) *FatigueSimulator {
	if config == nil {
		config = &FatigueSimConfig{
			InitialPolarization: 15.0, // µC/cm²
			CoerciveField:       1.5,
			FilmThickness:       10.0,
			Temperature:         300.0,
			CyclingFrequency:    1e6,
		}
	}

	return &FatigueSimulator{
		Config:              config,
		CurrentPolarization: config.InitialPolarization * 0.5, // Starts lower due to pinned domains
		PolarizationHistory: make([]float64, 0),
		CycleHistory:        make([]int64, 0),
		Phase:               "pristine",
		WakeUpCycles:        1000,
		FatigueCycles:       1e8,
	}
}

// SimulateCycles simulates a number of switching cycles
func (fs *FatigueSimulator) SimulateCycles(numCycles int64) {
	for i := int64(0); i < numCycles; i++ {
		totalCycles := int64(len(fs.CycleHistory)) + i

		// Wake-up phase: polarization increases
		if totalCycles < fs.WakeUpCycles {
			fs.Phase = "wake-up"
			wakeUpFactor := 1.0 - math.Exp(-float64(totalCycles)/float64(fs.WakeUpCycles)*5)
			fs.CurrentPolarization = fs.Config.InitialPolarization * 0.5 *
				(1.0 + wakeUpFactor)
		} else if totalCycles < fs.FatigueCycles {
			// Mature phase: stable polarization
			fs.Phase = "mature"
			fs.CurrentPolarization = fs.Config.InitialPolarization
		} else {
			// Fatigue phase: polarization decreases
			fs.Phase = "fatigue"
			fatigueFactor := math.Exp(-float64(totalCycles-fs.FatigueCycles) / 1e9)
			fs.CurrentPolarization = fs.Config.InitialPolarization * fatigueFactor
		}

		// Record every 1000 cycles
		if (totalCycles+1)%1000 == 0 {
			fs.PolarizationHistory = append(fs.PolarizationHistory, fs.CurrentPolarization)
			fs.CycleHistory = append(fs.CycleHistory, totalCycles+1)
		}
	}
}

// GetPolarizationVsCycles returns the polarization evolution
func (fs *FatigueSimulator) GetPolarizationVsCycles() ([]int64, []float64) {
	return fs.CycleHistory, fs.PolarizationHistory
}

// PredictFailureCycles predicts when polarization drops below threshold
func (fs *FatigueSimulator) PredictFailureCycles(threshold float64) int64 {
	// Simple model: exponential decay after fatigue onset
	if fs.CurrentPolarization < threshold {
		return int64(len(fs.CycleHistory))
	}

	// Predict future failure
	targetRatio := threshold / fs.Config.InitialPolarization
	cyclesAfterFatigue := -1e9 * math.Log(targetRatio)

	return fs.FatigueCycles + int64(cyclesAfterFatigue)
}

// =============================================================================
// Helper Functions
// =============================================================================

// sortByValue sorts a map by value and returns sorted keys
func sortByValue(m map[string]float64, ascending bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		if ascending {
			return m[keys[i]] < m[keys[j]]
		}
		return m[keys[i]] > m[keys[j]]
	})

	return keys
}

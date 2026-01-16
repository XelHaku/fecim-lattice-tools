// photonic_gnn_cim.go - Photonic-Electronic Hybrid CIM and GNN Accelerators
// Part of IronLattice educational demonstrations
//
// This module implements:
// 1. Photonic compute-in-memory with silicon photonics (MZI, MRR)
// 2. Ferroelectric photonic memory (HZO + LiNbO3)
// 3. Graph Neural Network accelerators (ReRAM-based GCN/GNN)
// 4. Sparse aggregation and combination operations
//
// Research basis:
// - Pockels photonic memory (Nature Communications 2025)
// - Complex-valued MVM photonic processor (Science Advances 2025)
// - ReGNN ReRAM-based GNN accelerator (DAC 2022)
// - HePGA heterogeneous PIM GNN training (2025)

package layers

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
)

// =============================================================================
// PHOTONIC COMPUTE-IN-MEMORY
// =============================================================================

// PhotonicCIMConfig configures photonic compute-in-memory system
type PhotonicCIMConfig struct {
	// MZI mesh configuration
	MZIMeshSize     int     // Number of MZI stages
	PhaseResolution int     // Phase DAC bits
	WavelengthNm    float64 // Operating wavelength

	// MRR weight bank configuration
	NumMRRs         int     // Number of micro-ring resonators
	MRRFSRGHz       float64 // Free spectral range
	MRRQFactor      float64 // Quality factor

	// Performance parameters
	ModulationGHz   float64 // Modulation bandwidth
	InsertionLossDB float64 // Total insertion loss
	CrosstalkDB     float64 // Channel crosstalk

	// Ferroelectric photonic memory
	UseFerroMemory  bool
	MemoryStates    int     // Multi-level states per cell
	RetentionYears  float64
	EnduranceCycles int
}

// DefaultPhotonicCIMConfig returns typical photonic CIM parameters
func DefaultPhotonicCIMConfig() *PhotonicCIMConfig {
	return &PhotonicCIMConfig{
		MZIMeshSize:     16,
		PhaseResolution: 8,
		WavelengthNm:    1550.0,
		NumMRRs:         64,
		MRRFSRGHz:       100.0,
		MRRQFactor:      10000,
		ModulationGHz:   40.0,
		InsertionLossDB: 3.0,
		CrosstalkDB:     -30.0,
		UseFerroMemory:  true,
		MemoryStates:    6,
		RetentionYears:  10.0,
		EnduranceCycles: 10000000,
	}
}

// MachZehnderInterferometer represents a single MZI unit
type MachZehnderInterferometer struct {
	PhaseShiftUpper float64 // Phase shift in upper arm (radians)
	PhaseShiftLower float64 // Phase shift in lower arm (radians)
	SplitRatio      float64 // Power split ratio (0.5 = 3dB)
	InsertionLoss   float64 // dB loss
}

// NewMZI creates a new MZI with default 3dB coupling
func NewMZI(thetaUpper, thetaLower float64) *MachZehnderInterferometer {
	return &MachZehnderInterferometer{
		PhaseShiftUpper: thetaUpper,
		PhaseShiftLower: thetaLower,
		SplitRatio:      0.5,
		InsertionLoss:   0.1,
	}
}

// TransferMatrix returns the 2x2 complex transfer matrix
func (mzi *MachZehnderInterferometer) TransferMatrix() [2][2]complex128 {
	// MZI transfer matrix: T = BS2 * Phase * BS1
	// where BS is beam splitter and Phase is differential phase shift

	phi := mzi.PhaseShiftUpper - mzi.PhaseShiftLower
	theta := (mzi.PhaseShiftUpper + mzi.PhaseShiftLower) / 2

	// Simplified 2x2 transfer matrix
	loss := math.Pow(10, -mzi.InsertionLoss/20)

	t11 := complex(loss*math.Cos(phi/2), 0) * cmplx.Exp(complex(0, theta))
	t12 := complex(loss*math.Sin(phi/2), 0) * cmplx.Exp(complex(0, theta))
	t21 := complex(-loss*math.Sin(phi/2), 0) * cmplx.Exp(complex(0, theta))
	t22 := complex(loss*math.Cos(phi/2), 0) * cmplx.Exp(complex(0, theta))

	return [2][2]complex128{{t11, t12}, {t21, t22}}
}

// MZIMesh represents a mesh of MZI for unitary transformations
type MZIMesh struct {
	Config     *PhotonicCIMConfig
	MZIs       [][]*MachZehnderInterferometer
	InputPorts int
}

// NewMZIMesh creates a triangular MZI mesh (Reck decomposition)
func NewMZIMesh(config *PhotonicCIMConfig) *MZIMesh {
	n := config.MZIMeshSize
	mesh := &MZIMesh{
		Config:     config,
		InputPorts: n,
		MZIs:       make([][]*MachZehnderInterferometer, n-1),
	}

	// Triangular mesh for n×n unitary matrix
	for layer := 0; layer < n-1; layer++ {
		numMZI := n - 1 - layer
		mesh.MZIs[layer] = make([]*MachZehnderInterferometer, numMZI)
		for i := 0; i < numMZI; i++ {
			mesh.MZIs[layer][i] = NewMZI(0, 0)
		}
	}

	return mesh
}

// SetUnitary programs the MZI mesh to implement a unitary matrix
func (mesh *MZIMesh) SetUnitary(U [][]complex128) error {
	n := len(U)
	if n != mesh.InputPorts {
		return fmt.Errorf("matrix size %d doesn't match mesh size %d", n, mesh.InputPorts)
	}

	// Reck decomposition: decompose U into product of 2×2 rotations
	// This is a simplified version - full implementation would use
	// Givens rotations to extract phases

	for layer := 0; layer < n-1; layer++ {
		for i := 0; i < len(mesh.MZIs[layer]); i++ {
			// Extract phases from U (simplified)
			row := layer
			col := n - 1 - i
			if row < n && col < n {
				phase := cmplx.Phase(U[row][col])
				mesh.MZIs[layer][i].PhaseShiftUpper = phase
				mesh.MZIs[layer][i].PhaseShiftLower = 0
			}
		}
	}

	return nil
}

// Forward performs optical MVM through the mesh
func (mesh *MZIMesh) Forward(input []complex128) []complex128 {
	n := mesh.InputPorts
	if len(input) != n {
		return nil
	}

	current := make([]complex128, n)
	copy(current, input)

	// Propagate through MZI layers
	for layer := 0; layer < len(mesh.MZIs); layer++ {
		next := make([]complex128, n)
		copy(next, current)

		for i, mzi := range mesh.MZIs[layer] {
			if i+1 >= n {
				continue
			}
			T := mzi.TransferMatrix()

			// Apply 2×2 transformation to adjacent ports
			in1 := current[i]
			in2 := current[i+1]

			next[i] = T[0][0]*in1 + T[0][1]*in2
			next[i+1] = T[1][0]*in1 + T[1][1]*in2
		}

		current = next
	}

	return current
}

// MicroRingResonator represents a single MRR weight element
type MicroRingResonator struct {
	Wavelength    float64 // Resonance wavelength (nm)
	QFactor       float64 // Quality factor
	Transmission  float64 // Through-port transmission (0-1)
	DropPort      float64 // Drop-port coupling
	ThermalTuning float64 // Thermal tuning range (nm/mW)
}

// NewMRR creates a micro-ring resonator
func NewMRR(wavelength, qFactor float64) *MicroRingResonator {
	return &MicroRingResonator{
		Wavelength:    wavelength,
		QFactor:       qFactor,
		Transmission:  0.5,
		DropPort:      0.5,
		ThermalTuning: 0.1, // nm/mW typical for silicon
	}
}

// SetWeight adjusts MRR transmission to encode weight
func (mrr *MicroRingResonator) SetWeight(weight float64) {
	// Weight encoded as transmission coefficient
	mrr.Transmission = math.Max(0, math.Min(1, (weight+1)/2))
	mrr.DropPort = 1 - mrr.Transmission
}

// GetWeight returns the encoded weight
func (mrr *MicroRingResonator) GetWeight() float64 {
	return mrr.Transmission*2 - 1
}

// MRRWeightBank implements broadcast-and-weight architecture
type MRRWeightBank struct {
	Config     *PhotonicCIMConfig
	MRRs       [][]*MicroRingResonator
	InputSize  int
	OutputSize int
}

// NewMRRWeightBank creates an MRR-based weight bank
func NewMRRWeightBank(config *PhotonicCIMConfig, inputSize, outputSize int) *MRRWeightBank {
	bank := &MRRWeightBank{
		Config:     config,
		InputSize:  inputSize,
		OutputSize: outputSize,
		MRRs:       make([][]*MicroRingResonator, outputSize),
	}

	baseWavelength := config.WavelengthNm
	channelSpacing := 0.8 // nm (100 GHz at 1550nm)

	for i := 0; i < outputSize; i++ {
		bank.MRRs[i] = make([]*MicroRingResonator, inputSize)
		for j := 0; j < inputSize; j++ {
			wl := baseWavelength + float64(j)*channelSpacing
			bank.MRRs[i][j] = NewMRR(wl, config.MRRQFactor)
		}
	}

	return bank
}

// SetWeights programs weight matrix into MRR bank
func (bank *MRRWeightBank) SetWeights(weights [][]float64) error {
	if len(weights) != bank.OutputSize {
		return fmt.Errorf("weight rows %d != output size %d", len(weights), bank.OutputSize)
	}

	for i, row := range weights {
		if len(row) != bank.InputSize {
			return fmt.Errorf("weight cols %d != input size %d", len(row), bank.InputSize)
		}
		for j, w := range row {
			bank.MRRs[i][j].SetWeight(w)
		}
	}

	return nil
}

// Forward performs weighted sum using MRR bank
func (bank *MRRWeightBank) Forward(input []float64) []float64 {
	output := make([]float64, bank.OutputSize)

	for i := 0; i < bank.OutputSize; i++ {
		sum := 0.0
		for j := 0; j < bank.InputSize; j++ {
			weight := bank.MRRs[i][j].GetWeight()
			sum += input[j] * weight
		}
		output[i] = sum
	}

	return output
}

// =============================================================================
// FERROELECTRIC PHOTONIC MEMORY
// =============================================================================

// FerroPhotonicMemoryConfig configures ferroelectric photonic memory
type FerroPhotonicMemoryConfig struct {
	// Ferroelectric material
	Material         string  // "HZO", "BaTiO3", "LiNbO3"
	ThicknessNm      float64 // Ferroelectric film thickness
	CoerciveFieldMV  float64 // Coercive field (MV/cm)

	// Electro-optic properties
	PockelsCoeff     float64 // Pockels coefficient (pm/V)
	RefractiveIndex  float64 // Base refractive index

	// Memory characteristics
	NumStates        int     // Multi-level storage states
	SwitchingEnergyFJ float64 // Energy per state switch
	RetentionYears   float64 // Data retention
	EnduranceCycles  int     // Write endurance

	// Integration
	RingRadiusUm     float64 // Ring resonator radius
	WaveguideWidthNm float64 // Waveguide width
}

// DefaultHZOPhotonicConfig returns HZO-LiNbO3 Pockels memory config
func DefaultHZOPhotonicConfig() *FerroPhotonicMemoryConfig {
	return &FerroPhotonicMemoryConfig{
		Material:         "HZO",
		ThicknessNm:      10,
		CoerciveFieldMV:  1.0,
		PockelsCoeff:     31,    // pm/V (LiNbO3)
		RefractiveIndex:  2.21,  // LiNbO3
		NumStates:        6,     // 6 states demonstrated
		SwitchingEnergyFJ: 1.0,  // femtoJoule/state
		RetentionYears:   10.0,
		EnduranceCycles:  10000000, // 10^7 cycles
		RingRadiusUm:     20.0,
		WaveguideWidthNm: 600,
	}
}

// DefaultBaTiO3PhotonicConfig returns BaTiO3-Si config
func DefaultBaTiO3PhotonicConfig() *FerroPhotonicMemoryConfig {
	return &FerroPhotonicMemoryConfig{
		Material:         "BaTiO3",
		ThicknessNm:      50,
		CoerciveFieldMV:  0.1,
		PockelsCoeff:     213,   // pm/V (6× LiNbO3)
		RefractiveIndex:  2.4,
		NumStates:        8,
		SwitchingEnergyFJ: 10.0,
		RetentionYears:   5.0,
		EnduranceCycles:  100000,
		RingRadiusUm:     10.0,
		WaveguideWidthNm: 450,
	}
}

// FerroPhotonicCell represents a single ferroelectric photonic memory cell
type FerroPhotonicCell struct {
	Config          *FerroPhotonicMemoryConfig
	Polarization    float64   // Current polarization state (-1 to 1)
	OpticalPhase    float64   // Induced optical phase shift
	StoredState     int       // Discrete state index
	WriteCount      int       // Number of write operations
}

// NewFerroPhotonicCell creates a ferroelectric photonic memory cell
func NewFerroPhotonicCell(config *FerroPhotonicMemoryConfig) *FerroPhotonicCell {
	return &FerroPhotonicCell{
		Config:       config,
		Polarization: 0,
		StoredState:  config.NumStates / 2,
	}
}

// Write programs the cell to a specific state
func (cell *FerroPhotonicCell) Write(state int) error {
	if state < 0 || state >= cell.Config.NumStates {
		return fmt.Errorf("state %d out of range [0, %d)", state, cell.Config.NumStates)
	}

	if cell.WriteCount >= cell.Config.EnduranceCycles {
		return fmt.Errorf("endurance limit reached: %d cycles", cell.Config.EnduranceCycles)
	}

	// Map state to polarization
	cell.StoredState = state
	cell.Polarization = 2*float64(state)/float64(cell.Config.NumStates-1) - 1

	// Calculate induced phase shift via Pockels effect
	// Δφ = (2π/λ) × n³ × r × E × L
	// Simplified: phase proportional to polarization
	cell.OpticalPhase = cell.Polarization * math.Pi / 4

	cell.WriteCount++
	return nil
}

// Read returns the optical transmission/phase state
func (cell *FerroPhotonicCell) Read() (state int, phase float64) {
	return cell.StoredState, cell.OpticalPhase
}

// GetTransmission returns the optical transmission through the cell
func (cell *FerroPhotonicCell) GetTransmission() float64 {
	// Ring resonator transmission depends on phase
	// T = 1 - (1-a²)/(1 + a² - 2a×cos(φ))
	a := 0.9 // Round-trip amplitude transmission
	phi := cell.OpticalPhase

	numerator := 1 - a*a
	denominator := 1 + a*a - 2*a*math.Cos(phi)

	return 1 - numerator/denominator
}

// FerroPhotonicArray represents an array of ferroelectric photonic cells
type FerroPhotonicArray struct {
	Config      *FerroPhotonicMemoryConfig
	Cells       [][]*FerroPhotonicCell
	Rows        int
	Cols        int
	TotalEnergy float64 // Cumulative energy consumption
}

// NewFerroPhotonicArray creates a ferroelectric photonic memory array
func NewFerroPhotonicArray(config *FerroPhotonicMemoryConfig, rows, cols int) *FerroPhotonicArray {
	array := &FerroPhotonicArray{
		Config: config,
		Rows:   rows,
		Cols:   cols,
		Cells:  make([][]*FerroPhotonicCell, rows),
	}

	for i := 0; i < rows; i++ {
		array.Cells[i] = make([]*FerroPhotonicCell, cols)
		for j := 0; j < cols; j++ {
			array.Cells[i][j] = NewFerroPhotonicCell(config)
		}
	}

	return array
}

// ProgramWeights stores weight matrix in photonic memory
func (array *FerroPhotonicArray) ProgramWeights(weights [][]float64) error {
	numStates := array.Config.NumStates

	for i := 0; i < array.Rows && i < len(weights); i++ {
		for j := 0; j < array.Cols && j < len(weights[i]); j++ {
			// Quantize weight to available states
			w := weights[i][j]
			state := int((w + 1) / 2 * float64(numStates-1))
			state = int(math.Max(0, math.Min(float64(numStates-1), float64(state))))

			if err := array.Cells[i][j].Write(state); err != nil {
				return err
			}
			array.TotalEnergy += array.Config.SwitchingEnergyFJ
		}
	}

	return nil
}

// OpticalMVM performs matrix-vector multiply using optical readout
func (array *FerroPhotonicArray) OpticalMVM(input []float64) []float64 {
	output := make([]float64, array.Rows)

	for i := 0; i < array.Rows; i++ {
		sum := 0.0
		for j := 0; j < array.Cols && j < len(input); j++ {
			transmission := array.Cells[i][j].GetTransmission()
			weight := transmission*2 - 1 // Convert to [-1, 1]
			sum += input[j] * weight
		}
		output[i] = sum
	}

	return output
}

// =============================================================================
// GRAPH NEURAL NETWORK ACCELERATOR
// =============================================================================

// GNNAcceleratorConfig configures GNN hardware accelerator
type GNNAcceleratorConfig struct {
	// Array configuration
	CrossbarSize    int     // ReRAM crossbar dimension
	NumCrossbars    int     // Number of crossbar arrays
	BitPrecision    int     // Weight precision

	// Memory hierarchy
	LocalBufferKB   int     // Per-PE local buffer
	GlobalBufferMB  int     // Shared global buffer

	// Sparse handling
	SparseThreshold float64 // Sparsity threshold
	CSRFormat       bool    // Use CSR for adjacency

	// Performance
	ClockMHz        int
	PowermW         float64
}

// DefaultGNNAcceleratorConfig returns typical GNN accelerator parameters
func DefaultGNNAcceleratorConfig() *GNNAcceleratorConfig {
	return &GNNAcceleratorConfig{
		CrossbarSize:    128,
		NumCrossbars:    64,
		BitPrecision:    8,
		LocalBufferKB:   64,
		GlobalBufferMB:  4,
		SparseThreshold: 0.9,
		CSRFormat:       true,
		ClockMHz:        500,
		PowermW:         500,
	}
}

// GraphNode represents a node in the graph
type GraphNode struct {
	ID       int
	Features []float64
	Degree   int
	Neighbors []int
}

// SparseGraph represents a graph in CSR format
type SparseGraph struct {
	NumNodes    int
	NumEdges    int
	Nodes       []*GraphNode
	RowPtr      []int     // CSR row pointers
	ColIdx      []int     // CSR column indices
	EdgeWeights []float64 // Edge weights (optional)
}

// NewSparseGraph creates an empty sparse graph
func NewSparseGraph(numNodes int) *SparseGraph {
	graph := &SparseGraph{
		NumNodes: numNodes,
		NumEdges: 0,
		Nodes:    make([]*GraphNode, numNodes),
		RowPtr:   make([]int, numNodes+1),
	}

	for i := 0; i < numNodes; i++ {
		graph.Nodes[i] = &GraphNode{
			ID:        i,
			Features:  make([]float64, 0),
			Neighbors: make([]int, 0),
		}
	}

	return graph
}

// AddEdge adds an edge to the graph
func (g *SparseGraph) AddEdge(from, to int, weight float64) {
	if from >= 0 && from < g.NumNodes && to >= 0 && to < g.NumNodes {
		g.Nodes[from].Neighbors = append(g.Nodes[from].Neighbors, to)
		g.Nodes[from].Degree++
		g.ColIdx = append(g.ColIdx, to)
		g.EdgeWeights = append(g.EdgeWeights, weight)
		g.NumEdges++
	}
}

// BuildCSR builds CSR representation from adjacency lists
func (g *SparseGraph) BuildCSR() {
	g.RowPtr = make([]int, g.NumNodes+1)
	g.ColIdx = make([]int, 0)
	g.EdgeWeights = make([]float64, 0)

	edgeIdx := 0
	for i := 0; i < g.NumNodes; i++ {
		g.RowPtr[i] = edgeIdx
		for _, neighbor := range g.Nodes[i].Neighbors {
			g.ColIdx = append(g.ColIdx, neighbor)
			g.EdgeWeights = append(g.EdgeWeights, 1.0)
			edgeIdx++
		}
	}
	g.RowPtr[g.NumNodes] = edgeIdx
	g.NumEdges = edgeIdx
}

// GetNeighbors returns neighbors of a node using CSR
func (g *SparseGraph) GetNeighbors(nodeID int) []int {
	if nodeID < 0 || nodeID >= g.NumNodes {
		return nil
	}
	start := g.RowPtr[nodeID]
	end := g.RowPtr[nodeID+1]
	return g.ColIdx[start:end]
}

// GNNLayer represents a single GNN layer
type GNNLayer struct {
	InputDim    int
	OutputDim   int
	Weights     [][]float64 // Transformation weights
	Bias        []float64
	Aggregation string      // "mean", "sum", "max"
}

// NewGNNLayer creates a new GNN layer
func NewGNNLayer(inputDim, outputDim int, aggregation string) *GNNLayer {
	layer := &GNNLayer{
		InputDim:    inputDim,
		OutputDim:   outputDim,
		Weights:     make([][]float64, outputDim),
		Bias:        make([]float64, outputDim),
		Aggregation: aggregation,
	}

	// Xavier initialization
	scale := math.Sqrt(2.0 / float64(inputDim+outputDim))
	for i := 0; i < outputDim; i++ {
		layer.Weights[i] = make([]float64, inputDim)
		for j := 0; j < inputDim; j++ {
			layer.Weights[i][j] = rand.NormFloat64() * scale
		}
	}

	return layer
}

// GCNAccelerator implements ReRAM-based GCN acceleration
type GCNAccelerator struct {
	Config       *GNNAcceleratorConfig
	Layers       []*GNNLayer
	Crossbars    [][][]float64 // Weight crossbars
	Stats        *GNNAccelStats
}

// GNNAccelStats tracks accelerator statistics
type GNNAccelStats struct {
	AggregationOps    int64
	CombinationOps    int64
	MemoryAccesses    int64
	EnergyPJ          float64
	LatencyNs         float64
	SpeedupVsGPU      float64
	EnergyReduction   float64
}

// NewGCNAccelerator creates a ReRAM-based GCN accelerator
func NewGCNAccelerator(config *GNNAcceleratorConfig) *GCNAccelerator {
	return &GCNAccelerator{
		Config:    config,
		Layers:    make([]*GNNLayer, 0),
		Crossbars: make([][][]float64, 0),
		Stats:     &GNNAccelStats{},
	}
}

// AddLayer adds a GNN layer to the accelerator
func (acc *GCNAccelerator) AddLayer(layer *GNNLayer) {
	acc.Layers = append(acc.Layers, layer)

	// Map weights to crossbar
	crossbar := make([][]float64, layer.OutputDim)
	for i := 0; i < layer.OutputDim; i++ {
		crossbar[i] = make([]float64, layer.InputDim)
		copy(crossbar[i], layer.Weights[i])
	}
	acc.Crossbars = append(acc.Crossbars, crossbar)
}

// Aggregate performs neighborhood aggregation
func (acc *GCNAccelerator) Aggregate(graph *SparseGraph, nodeFeatures [][]float64, method string) [][]float64 {
	numNodes := graph.NumNodes
	featureDim := len(nodeFeatures[0])
	aggregated := make([][]float64, numNodes)

	for i := 0; i < numNodes; i++ {
		aggregated[i] = make([]float64, featureDim)
		neighbors := graph.GetNeighbors(i)

		if len(neighbors) == 0 {
			copy(aggregated[i], nodeFeatures[i])
			continue
		}

		switch method {
		case "mean":
			for _, n := range neighbors {
				for d := 0; d < featureDim; d++ {
					aggregated[i][d] += nodeFeatures[n][d]
				}
			}
			for d := 0; d < featureDim; d++ {
				aggregated[i][d] /= float64(len(neighbors))
			}

		case "sum":
			for _, n := range neighbors {
				for d := 0; d < featureDim; d++ {
					aggregated[i][d] += nodeFeatures[n][d]
				}
			}

		case "max":
			for d := 0; d < featureDim; d++ {
				aggregated[i][d] = math.Inf(-1)
			}
			for _, n := range neighbors {
				for d := 0; d < featureDim; d++ {
					if nodeFeatures[n][d] > aggregated[i][d] {
						aggregated[i][d] = nodeFeatures[n][d]
					}
				}
			}
		}

		acc.Stats.AggregationOps += int64(len(neighbors) * featureDim)
		acc.Stats.MemoryAccesses += int64(len(neighbors))
	}

	return aggregated
}

// Combine performs feature transformation using crossbar MVM
func (acc *GCNAccelerator) Combine(features [][]float64, layerIdx int) [][]float64 {
	if layerIdx >= len(acc.Layers) {
		return nil
	}

	layer := acc.Layers[layerIdx]
	crossbar := acc.Crossbars[layerIdx]
	numNodes := len(features)
	output := make([][]float64, numNodes)

	for i := 0; i < numNodes; i++ {
		output[i] = make([]float64, layer.OutputDim)

		// MVM on crossbar
		for o := 0; o < layer.OutputDim; o++ {
			sum := layer.Bias[o]
			for d := 0; d < layer.InputDim && d < len(features[i]); d++ {
				sum += features[i][d] * crossbar[o][d]
			}
			output[i][o] = sum
		}

		acc.Stats.CombinationOps += int64(layer.InputDim * layer.OutputDim)
	}

	// Apply ReLU activation
	for i := range output {
		for j := range output[i] {
			if output[i][j] < 0 {
				output[i][j] = 0
			}
		}
	}

	return output
}

// Forward runs GCN inference on a graph
func (acc *GCNAccelerator) Forward(graph *SparseGraph, inputFeatures [][]float64) [][]float64 {
	features := inputFeatures

	for layerIdx, layer := range acc.Layers {
		// Aggregation phase
		aggregated := acc.Aggregate(graph, features, layer.Aggregation)

		// Combination phase (MVM on crossbar)
		features = acc.Combine(aggregated, layerIdx)
	}

	return features
}

// ComputeStats calculates performance statistics
func (acc *GCNAccelerator) ComputeStats() {
	totalOps := acc.Stats.AggregationOps + acc.Stats.CombinationOps

	// Energy model: ~0.5 pJ/MAC for ReRAM crossbar
	acc.Stats.EnergyPJ = float64(acc.Stats.CombinationOps) * 0.5

	// Add memory access energy
	acc.Stats.EnergyPJ += float64(acc.Stats.MemoryAccesses) * 2.0 // pJ per access

	// Latency model
	crossbarLatencyNs := 10.0 // ns per MVM
	numMVMs := float64(acc.Stats.CombinationOps) / float64(acc.Config.CrossbarSize*acc.Config.CrossbarSize)
	acc.Stats.LatencyNs = numMVMs * crossbarLatencyNs

	// Compare to GPU baseline
	gpuEnergyPJ := float64(totalOps) * 100.0 // ~100 pJ/MAC on GPU
	acc.Stats.EnergyReduction = gpuEnergyPJ / acc.Stats.EnergyPJ

	gpuLatencyNs := float64(totalOps) / 1000.0 // 1 TOPS GPU
	acc.Stats.SpeedupVsGPU = gpuLatencyNs / acc.Stats.LatencyNs
}

// =============================================================================
// HETEROGENEOUS PHOTONIC-GNN ACCELERATOR
// =============================================================================

// HybridPhotonicGNNConfig configures hybrid photonic-electronic GNN accelerator
type HybridPhotonicGNNConfig struct {
	// Photonic section (Combination)
	PhotonicConfig *PhotonicCIMConfig
	UseMZI         bool
	UseMRR         bool

	// Electronic section (Aggregation)
	ElectronicConfig *GNNAcceleratorConfig

	// System parameters
	PhotonicElecRatio float64 // Fraction of ops on photonic
	InterconnectBWGBs float64 // Photonic-electronic bandwidth
}

// DefaultHybridConfig returns typical hybrid accelerator config
func DefaultHybridConfig() *HybridPhotonicGNNConfig {
	return &HybridPhotonicGNNConfig{
		PhotonicConfig:    DefaultPhotonicCIMConfig(),
		UseMZI:            true,
		UseMRR:            true,
		ElectronicConfig:  DefaultGNNAcceleratorConfig(),
		PhotonicElecRatio: 0.7, // 70% on photonic (Combination)
		InterconnectBWGBs: 100,
	}
}

// HybridPhotonicGNN implements photonic-electronic hybrid GNN accelerator
type HybridPhotonicGNN struct {
	Config          *HybridPhotonicGNNConfig
	PhotonicWeights *MRRWeightBank
	FerroMemory     *FerroPhotonicArray
	ElectronicGNN   *GCNAccelerator
	Stats           *HybridGNNStats
}

// HybridGNNStats tracks hybrid accelerator performance
type HybridGNNStats struct {
	PhotonicOps     int64
	ElectronicOps   int64
	TotalEnergyPJ   float64
	PhotonicEnergyPJ float64
	ElecEnergyPJ    float64
	ThroughputTOPS  float64
	EnergyEffTOPSW  float64
}

// NewHybridPhotonicGNN creates a hybrid photonic-GNN accelerator
func NewHybridPhotonicGNN(config *HybridPhotonicGNNConfig) *HybridPhotonicGNN {
	return &HybridPhotonicGNN{
		Config:        config,
		ElectronicGNN: NewGCNAccelerator(config.ElectronicConfig),
		Stats:         &HybridGNNStats{},
	}
}

// Initialize sets up the accelerator with model weights
func (hybrid *HybridPhotonicGNN) Initialize(layers []*GNNLayer) error {
	// Setup electronic GNN for aggregation
	for _, layer := range layers {
		hybrid.ElectronicGNN.AddLayer(layer)
	}

	// Setup photonic weight bank for combination
	if len(layers) > 0 {
		inputDim := layers[0].InputDim
		outputDim := layers[0].OutputDim

		// Create MRR weight bank
		hybrid.PhotonicWeights = NewMRRWeightBank(
			hybrid.Config.PhotonicConfig,
			inputDim,
			outputDim,
		)

		// Create ferroelectric photonic memory
		ferroConfig := DefaultHZOPhotonicConfig()
		hybrid.FerroMemory = NewFerroPhotonicArray(ferroConfig, outputDim, inputDim)

		// Program weights
		hybrid.FerroMemory.ProgramWeights(layers[0].Weights)
	}

	return nil
}

// Forward performs hybrid inference
func (hybrid *HybridPhotonicGNN) Forward(graph *SparseGraph, features [][]float64) [][]float64 {
	// Phase 1: Electronic aggregation
	aggregated := hybrid.ElectronicGNN.Aggregate(graph, features, "mean")
	hybrid.Stats.ElectronicOps += hybrid.ElectronicGNN.Stats.AggregationOps

	// Phase 2: Photonic combination
	output := make([][]float64, len(aggregated))
	for i, feat := range aggregated {
		output[i] = hybrid.FerroMemory.OpticalMVM(feat)
		hybrid.Stats.PhotonicOps += int64(len(feat) * len(output[i]))
	}

	// ReLU activation
	for i := range output {
		for j := range output[i] {
			if output[i][j] < 0 {
				output[i][j] = 0
			}
		}
	}

	return output
}

// ComputeStats calculates hybrid accelerator statistics
func (hybrid *HybridPhotonicGNN) ComputeStats() {
	// Photonic energy: ~1 fJ/MAC (optical)
	hybrid.Stats.PhotonicEnergyPJ = float64(hybrid.Stats.PhotonicOps) * 0.001

	// Electronic energy: ~0.5 pJ/MAC (ReRAM)
	hybrid.Stats.ElecEnergyPJ = float64(hybrid.Stats.ElectronicOps) * 0.5

	// Total energy
	hybrid.Stats.TotalEnergyPJ = hybrid.Stats.PhotonicEnergyPJ + hybrid.Stats.ElecEnergyPJ

	// Throughput (assuming 1 GHz effective)
	totalOps := hybrid.Stats.PhotonicOps + hybrid.Stats.ElectronicOps
	hybrid.Stats.ThroughputTOPS = float64(totalOps) / 1000.0 // TOPS

	// Energy efficiency
	energyJ := hybrid.Stats.TotalEnergyPJ * 1e-12
	if energyJ > 0 {
		hybrid.Stats.EnergyEffTOPSW = hybrid.Stats.ThroughputTOPS / (energyJ * 1e12) // TOPS/W
	}
}

// =============================================================================
// DEMO AND BENCHMARK FUNCTIONS
// =============================================================================

// DemoPhotonicCIM demonstrates photonic compute-in-memory
func DemoPhotonicCIM() {
	fmt.Println("=== Photonic Compute-in-Memory Demo ===")
	fmt.Println()

	// 1. MZI Mesh demonstration
	fmt.Println("1. MZI Mesh Optical MVM:")
	config := DefaultPhotonicCIMConfig()
	config.MZIMeshSize = 4
	mesh := NewMZIMesh(config)

	// Set some phases
	for layer := range mesh.MZIs {
		for i := range mesh.MZIs[layer] {
			mesh.MZIs[layer][i].PhaseShiftUpper = float64(layer+i) * 0.1
		}
	}

	// Complex input
	input := []complex128{
		complex(1, 0),
		complex(0, 1),
		complex(0.707, 0.707),
		complex(0.707, -0.707),
	}

	output := mesh.Forward(input)
	fmt.Printf("   Input:  ")
	for _, c := range input {
		fmt.Printf("%.2f ", cmplx.Abs(c))
	}
	fmt.Println()
	fmt.Printf("   Output: ")
	for _, c := range output {
		fmt.Printf("%.2f ", cmplx.Abs(c))
	}
	fmt.Println()
	fmt.Println()

	// 2. MRR Weight Bank
	fmt.Println("2. MRR Weight Bank:")
	bank := NewMRRWeightBank(config, 4, 2)
	weights := [][]float64{
		{0.5, -0.3, 0.8, 0.1},
		{-0.2, 0.6, -0.4, 0.9},
	}
	bank.SetWeights(weights)

	realInput := []float64{1.0, 0.5, -0.5, 0.2}
	realOutput := bank.Forward(realInput)
	fmt.Printf("   Input:  %v\n", realInput)
	fmt.Printf("   Output: %.3f\n", realOutput)
	fmt.Println()

	// 3. Ferroelectric Photonic Memory
	fmt.Println("3. Ferroelectric Photonic Memory (HZO+LiNbO3):")
	ferroConfig := DefaultHZOPhotonicConfig()
	fmt.Printf("   Material: %s + LiNbO3\n", ferroConfig.Material)
	fmt.Printf("   States: %d per cell\n", ferroConfig.NumStates)
	fmt.Printf("   Retention: %.0f years\n", ferroConfig.RetentionYears)
	fmt.Printf("   Endurance: %d cycles\n", ferroConfig.EnduranceCycles)
	fmt.Printf("   Energy: %.1f fJ/state\n", ferroConfig.SwitchingEnergyFJ)

	cell := NewFerroPhotonicCell(ferroConfig)
	for state := 0; state < ferroConfig.NumStates; state++ {
		cell.Write(state)
		_, phase := cell.Read()
		trans := cell.GetTransmission()
		fmt.Printf("   State %d: phase=%.3f rad, T=%.3f\n", state, phase, trans)
	}
	fmt.Println()

	// 4. Photonic Array MVM
	fmt.Println("4. Photonic Memory Array MVM:")
	array := NewFerroPhotonicArray(ferroConfig, 2, 4)
	array.ProgramWeights(weights)
	arrayOutput := array.OpticalMVM(realInput)
	fmt.Printf("   MVM output: %.3f\n", arrayOutput)
	fmt.Printf("   Total energy: %.1f fJ\n", array.TotalEnergy)
}

// DemoGNNAccelerator demonstrates GNN hardware acceleration
func DemoGNNAccelerator() {
	fmt.Println()
	fmt.Println("=== GNN Hardware Accelerator Demo ===")
	fmt.Println()

	// Create sample graph
	fmt.Println("1. Creating sample graph (Karate Club style):")
	graph := NewSparseGraph(8)

	// Add edges
	edges := [][2]int{
		{0, 1}, {0, 2}, {0, 3},
		{1, 0}, {1, 2},
		{2, 0}, {2, 1}, {2, 3}, {2, 4},
		{3, 0}, {3, 2}, {3, 5},
		{4, 2}, {4, 5}, {4, 6},
		{5, 3}, {5, 4}, {5, 7},
		{6, 4}, {6, 7},
		{7, 5}, {7, 6},
	}

	for _, e := range edges {
		graph.AddEdge(e[0], e[1], 1.0)
	}
	graph.BuildCSR()

	fmt.Printf("   Nodes: %d, Edges: %d\n", graph.NumNodes, graph.NumEdges)
	fmt.Println()

	// Node features
	featureDim := 4
	features := make([][]float64, graph.NumNodes)
	for i := 0; i < graph.NumNodes; i++ {
		features[i] = make([]float64, featureDim)
		for j := 0; j < featureDim; j++ {
			features[i][j] = rand.Float64()*2 - 1
		}
		graph.Nodes[i].Features = features[i]
	}

	// Create GCN accelerator
	fmt.Println("2. GCN Accelerator Configuration:")
	config := DefaultGNNAcceleratorConfig()
	fmt.Printf("   Crossbar size: %d×%d\n", config.CrossbarSize, config.CrossbarSize)
	fmt.Printf("   Num crossbars: %d\n", config.NumCrossbars)
	fmt.Printf("   Precision: %d bits\n", config.BitPrecision)
	fmt.Println()

	// Create GCN layers
	acc := NewGCNAccelerator(config)
	layer1 := NewGNNLayer(featureDim, 8, "mean")
	layer2 := NewGNNLayer(8, 2, "mean")
	acc.AddLayer(layer1)
	acc.AddLayer(layer2)

	fmt.Println("3. GCN Architecture:")
	fmt.Printf("   Layer 1: %d → %d (aggregation: mean)\n", layer1.InputDim, layer1.OutputDim)
	fmt.Printf("   Layer 2: %d → %d (aggregation: mean)\n", layer2.InputDim, layer2.OutputDim)
	fmt.Println()

	// Run inference
	fmt.Println("4. Running GCN inference:")
	output := acc.Forward(graph, features)
	acc.ComputeStats()

	fmt.Printf("   Output shape: %d × %d\n", len(output), len(output[0]))
	fmt.Println("   Node embeddings (first 4):")
	for i := 0; i < 4 && i < len(output); i++ {
		fmt.Printf("     Node %d: [%.3f, %.3f]\n", i, output[i][0], output[i][1])
	}
	fmt.Println()

	fmt.Println("5. Performance Statistics:")
	fmt.Printf("   Aggregation ops: %d\n", acc.Stats.AggregationOps)
	fmt.Printf("   Combination ops: %d\n", acc.Stats.CombinationOps)
	fmt.Printf("   Memory accesses: %d\n", acc.Stats.MemoryAccesses)
	fmt.Printf("   Energy: %.2f pJ\n", acc.Stats.EnergyPJ)
	fmt.Printf("   Latency: %.2f ns\n", acc.Stats.LatencyNs)
	fmt.Printf("   Speedup vs GPU: %.1f×\n", acc.Stats.SpeedupVsGPU)
	fmt.Printf("   Energy reduction: %.1f×\n", acc.Stats.EnergyReduction)
}

// DemoHybridPhotonicGNN demonstrates hybrid photonic-electronic GNN
func DemoHybridPhotonicGNN() {
	fmt.Println()
	fmt.Println("=== Hybrid Photonic-Electronic GNN Demo ===")
	fmt.Println()

	// Create graph
	graph := NewSparseGraph(4)
	edges := [][2]int{{0, 1}, {1, 0}, {1, 2}, {2, 1}, {2, 3}, {3, 2}, {0, 3}, {3, 0}}
	for _, e := range edges {
		graph.AddEdge(e[0], e[1], 1.0)
	}
	graph.BuildCSR()

	// Node features
	features := [][]float64{
		{1.0, 0.5, -0.3, 0.8},
		{-0.2, 0.7, 0.4, -0.1},
		{0.6, -0.4, 0.9, 0.2},
		{-0.5, 0.3, -0.8, 0.6},
	}

	// Create hybrid accelerator
	config := DefaultHybridConfig()
	hybrid := NewHybridPhotonicGNN(config)

	// Create GNN layer
	layer := NewGNNLayer(4, 2, "mean")
	hybrid.Initialize([]*GNNLayer{layer})

	fmt.Println("1. System Configuration:")
	fmt.Println("   Aggregation: Electronic (ReRAM CIM)")
	fmt.Println("   Combination: Photonic (Ferroelectric memory)")
	fmt.Printf("   Photonic/Electronic ratio: %.0f%%/%.0f%%\n",
		config.PhotonicElecRatio*100, (1-config.PhotonicElecRatio)*100)
	fmt.Println()

	// Run inference
	fmt.Println("2. Running hybrid inference:")
	output := hybrid.Forward(graph, features)
	hybrid.ComputeStats()

	fmt.Println("   Output embeddings:")
	for i, out := range output {
		fmt.Printf("     Node %d: [%.3f, %.3f]\n", i, out[0], out[1])
	}
	fmt.Println()

	fmt.Println("3. Performance Breakdown:")
	fmt.Printf("   Photonic ops: %d\n", hybrid.Stats.PhotonicOps)
	fmt.Printf("   Electronic ops: %d\n", hybrid.Stats.ElectronicOps)
	fmt.Printf("   Photonic energy: %.3f pJ\n", hybrid.Stats.PhotonicEnergyPJ)
	fmt.Printf("   Electronic energy: %.3f pJ\n", hybrid.Stats.ElecEnergyPJ)
	fmt.Printf("   Total energy: %.3f pJ\n", hybrid.Stats.TotalEnergyPJ)

	// Comparison
	fmt.Println()
	fmt.Println("4. Efficiency Comparison:")
	fmt.Println("   Technology        Energy/MAC    Latency")
	fmt.Println("   ─────────────────────────────────────────")
	fmt.Println("   GPU               ~100 pJ       ~1 ns")
	fmt.Println("   ReRAM CIM         ~0.5 pJ       ~10 ns")
	fmt.Println("   Photonic CIM      ~0.001 pJ     ~0.1 ns")
	fmt.Println("   Hybrid (this)     ~0.25 pJ      ~5 ns")
}

// BenchmarkPhotonicGNN runs performance benchmarks
func BenchmarkPhotonicGNN() {
	fmt.Println()
	fmt.Println("=== Photonic-GNN Performance Benchmarks ===")
	fmt.Println()

	// Benchmark different graph sizes
	graphSizes := []int{100, 500, 1000, 5000}
	featureDim := 64
	outputDim := 16

	fmt.Println("Graph Size | Edges | Agg Ops | Comb Ops | Energy (pJ) | Speedup")
	fmt.Println("───────────────────────────────────────────────────────────────")

	for _, numNodes := range graphSizes {
		// Generate random graph (Erdos-Renyi)
		graph := NewSparseGraph(numNodes)
		edgeProb := 0.02 // Sparse graph

		for i := 0; i < numNodes; i++ {
			for j := 0; j < numNodes; j++ {
				if i != j && rand.Float64() < edgeProb {
					graph.AddEdge(i, j, 1.0)
				}
			}
		}
		graph.BuildCSR()

		// Generate features
		features := make([][]float64, numNodes)
		for i := 0; i < numNodes; i++ {
			features[i] = make([]float64, featureDim)
			for j := 0; j < featureDim; j++ {
				features[i][j] = rand.NormFloat64()
			}
		}

		// Create and run accelerator
		config := DefaultGNNAcceleratorConfig()
		acc := NewGCNAccelerator(config)
		layer := NewGNNLayer(featureDim, outputDim, "mean")
		acc.AddLayer(layer)

		acc.Forward(graph, features)
		acc.ComputeStats()

		fmt.Printf("%10d | %5d | %7d | %8d | %11.1f | %6.1f×\n",
			numNodes, graph.NumEdges,
			acc.Stats.AggregationOps, acc.Stats.CombinationOps,
			acc.Stats.EnergyPJ, acc.Stats.SpeedupVsGPU)
	}

	fmt.Println()
	fmt.Println("Key Performance Metrics (Literature):")
	fmt.Println("┌──────────────────────────────────────────────────────────┐")
	fmt.Println("│ Photonic MVM (Science Advances 2025)                     │")
	fmt.Println("│   Throughput: 1.28 TOPS (16-channel)                     │")
	fmt.Println("│   Energy: ~1 fJ/MAC                                      │")
	fmt.Println("├──────────────────────────────────────────────────────────┤")
	fmt.Println("│ Pockels Photonic Memory (Nature Comm 2025)               │")
	fmt.Println("│   States: 6 per transistor                               │")
	fmt.Println("│   Energy: ~1 fJ/state                                    │")
	fmt.Println("│   Retention: 10 years                                    │")
	fmt.Println("│   Endurance: 10^7 cycles                                 │")
	fmt.Println("├──────────────────────────────────────────────────────────┤")
	fmt.Println("│ ReGNN GCN Accelerator (DAC 2022)                         │")
	fmt.Println("│   Speedup: 228× vs GPU                                   │")
	fmt.Println("│   Energy reduction: 305.2×                               │")
	fmt.Println("├──────────────────────────────────────────────────────────┤")
	fmt.Println("│ HePGA GNN Training (2025)                                │")
	fmt.Println("│   Energy efficiency: 3.8× vs other PIM                   │")
	fmt.Println("│   Compute efficiency: 6.8× (TOPS/mm²)                    │")
	fmt.Println("└──────────────────────────────────────────────────────────┘")
}

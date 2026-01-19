package multilayer

import (
	"fmt"
	"math"
)

// Layer represents a single crossbar layer in the 3D stack.
type Layer struct {
	Name       string  // Layer name (e.g., "Input", "Hidden1", "Output")
	Rows       int     // Number of rows (inputs)
	Cols       int     // Number of columns (outputs)
	Levels     int     // Number of discrete conductance levels (30 for IronLattice)
	Weights    [][]int // Weight matrix (discrete levels 0-29)
	Activation string  // Activation function ("relu", "sigmoid", "none")
}

// Stack represents a 3D stack of crossbar layers.
type Stack struct {
	Layers      []*Layer // Ordered layers from input to output
	Name        string   // Stack name
	TotalVias   int      // Total via connections between layers
	Technology  string   // "IronLattice" or "Traditional"
	CellPitch   float64  // Cell pitch in nm
	LayerHeight float64  // Layer height in nm
}

// NewLayer creates a new crossbar layer.
func NewLayer(name string, rows, cols int) *Layer {
	weights := make([][]int, rows)
	for i := range weights {
		weights[i] = make([]int, cols)
		for j := range weights[i] {
			weights[i][j] = 15 // Initialize to middle level
		}
	}

	return &Layer{
		Name:       name,
		Rows:       rows,
		Cols:       cols,
		Levels:     30,
		Weights:    weights,
		Activation: "relu",
	}
}

// NewStack creates a new 3D stack.
func NewStack(name string) *Stack {
	return &Stack{
		Name:        name,
		Layers:      make([]*Layer, 0),
		Technology:  "IronLattice",
		CellPitch:   45.0, // 45nm cell pitch
		LayerHeight: 50.0, // 50nm per layer
	}
}

// MNISTStack creates a 3-layer stack for MNIST classification.
func MNISTStack() *Stack {
	stack := NewStack("MNIST-784-128-64-10")

	// Layer 1: Input (784) -> Hidden1 (128)
	layer1 := NewLayer("Hidden1", 784, 128)
	layer1.Activation = "relu"

	// Layer 2: Hidden1 (128) -> Hidden2 (64)
	layer2 := NewLayer("Hidden2", 128, 64)
	layer2.Activation = "relu"

	// Layer 3: Hidden2 (64) -> Output (10)
	layer3 := NewLayer("Output", 64, 10)
	layer3.Activation = "none" // Softmax applied externally

	stack.AddLayer(layer1)
	stack.AddLayer(layer2)
	stack.AddLayer(layer3)

	return stack
}

// SmallStack creates a smaller demo stack for visualization.
func SmallStack() *Stack {
	stack := NewStack("Demo-16-8-4")

	layer1 := NewLayer("Hidden", 16, 8)
	layer2 := NewLayer("Output", 8, 4)

	stack.AddLayer(layer1)
	stack.AddLayer(layer2)

	return stack
}

// AddLayer adds a layer to the stack.
func (s *Stack) AddLayer(layer *Layer) error {
	// Validate connectivity
	if len(s.Layers) > 0 {
		lastLayer := s.Layers[len(s.Layers)-1]
		if lastLayer.Cols != layer.Rows {
			return fmt.Errorf("layer dimension mismatch: %d cols vs %d rows",
				lastLayer.Cols, layer.Rows)
		}
	}

	s.Layers = append(s.Layers, layer)
	s.updateViaCount()
	return nil
}

// updateViaCount calculates total via connections.
func (s *Stack) updateViaCount() {
	s.TotalVias = 0
	for i := 0; i < len(s.Layers)-1; i++ {
		// Vias = number of connections between adjacent layers
		s.TotalVias += s.Layers[i].Cols
	}
}

// TotalCells returns the total number of memory cells in the stack.
func (s *Stack) TotalCells() int {
	total := 0
	for _, layer := range s.Layers {
		total += layer.Rows * layer.Cols
	}
	return total
}

// TotalParameters returns total number of weight parameters.
func (s *Stack) TotalParameters() int {
	return s.TotalCells() // Each cell stores one weight
}

// BitsPerCell returns effective bits per cell for IronLattice.
func (s *Stack) BitsPerCell() float64 {
	if len(s.Layers) == 0 {
		return 0
	}
	// log2(30) ≈ 4.9 bits for 30 levels
	return math.Log2(float64(s.Layers[0].Levels))
}

// TotalBits returns total storage capacity in bits.
func (s *Stack) TotalBits() float64 {
	return float64(s.TotalCells()) * s.BitsPerCell()
}

// StackHeight returns the total height of the stack in nm.
func (s *Stack) StackHeight() float64 {
	return float64(len(s.Layers)) * s.LayerHeight
}

// FootprintArea returns the footprint area in µm².
func (s *Stack) FootprintArea() float64 {
	if len(s.Layers) == 0 {
		return 0
	}
	// Find max layer dimensions
	maxRows, maxCols := 0, 0
	for _, layer := range s.Layers {
		if layer.Rows > maxRows {
			maxRows = layer.Rows
		}
		if layer.Cols > maxCols {
			maxCols = layer.Cols
		}
	}
	// Area in µm² (pitch is in nm)
	width := float64(maxCols) * s.CellPitch / 1000.0
	height := float64(maxRows) * s.CellPitch / 1000.0
	return width * height
}

// VolumetricDensity returns bits per µm³.
func (s *Stack) VolumetricDensity() float64 {
	area := s.FootprintArea()
	height := s.StackHeight() / 1000.0 // Convert to µm
	volume := area * height
	if volume == 0 {
		return 0
	}
	return s.TotalBits() / volume
}

// ArealDensity returns bits per µm².
func (s *Stack) ArealDensity() float64 {
	area := s.FootprintArea()
	if area == 0 {
		return 0
	}
	return s.TotalBits() / area
}

// LayerUtilization returns the utilization of each layer relative to max size.
func (s *Stack) LayerUtilization() []float64 {
	if len(s.Layers) == 0 {
		return nil
	}

	maxCells := 0
	for _, layer := range s.Layers {
		cells := layer.Rows * layer.Cols
		if cells > maxCells {
			maxCells = cells
		}
	}

	util := make([]float64, len(s.Layers))
	for i, layer := range s.Layers {
		cells := layer.Rows * layer.Cols
		util[i] = float64(cells) / float64(maxCells)
	}
	return util
}

// Forward performs forward pass through the stack.
func (s *Stack) Forward(input []float64) ([]float64, error) {
	if len(s.Layers) == 0 {
		return nil, fmt.Errorf("empty stack")
	}

	if len(input) != s.Layers[0].Rows {
		return nil, fmt.Errorf("input size mismatch: expected %d, got %d",
			s.Layers[0].Rows, len(input))
	}

	activation := input
	for _, layer := range s.Layers {
		output := make([]float64, layer.Cols)

		// Matrix-vector multiplication
		for j := 0; j < layer.Cols; j++ {
			sum := 0.0
			for i := 0; i < layer.Rows; i++ {
				// Convert discrete level to conductance
				weight := float64(layer.Weights[i][j]-15) / 15.0 // [-1, 1]
				sum += activation[i] * weight
			}
			output[j] = sum
		}

		// Apply activation
		switch layer.Activation {
		case "relu":
			for j := range output {
				if output[j] < 0 {
					output[j] = 0
				}
			}
		case "sigmoid":
			for j := range output {
				output[j] = 1.0 / (1.0 + math.Exp(-output[j]))
			}
		}

		activation = output
	}

	return activation, nil
}

// DataFlowStats returns statistics about data flow through the stack.
type DataFlowStats struct {
	LayerName      string
	InputSize      int
	OutputSize     int
	MACOperations  int
	DataMovement   int // Bytes moved (for comparison)
	CIMAdvantage   float64
}

// AnalyzeDataFlow returns data flow statistics for each layer.
func (s *Stack) AnalyzeDataFlow() []DataFlowStats {
	stats := make([]DataFlowStats, len(s.Layers))

	prevSize := 0
	if len(s.Layers) > 0 {
		prevSize = s.Layers[0].Rows
	}

	for i, layer := range s.Layers {
		macs := layer.Rows * layer.Cols

		// Traditional: Need to move weights + inputs + outputs
		// IronLattice: Weights in-situ, only move inputs + outputs
		traditionalData := layer.Rows*layer.Cols*4 + layer.Rows*4 + layer.Cols*4
		cimData := layer.Rows*4 + layer.Cols*4

		stats[i] = DataFlowStats{
			LayerName:     layer.Name,
			InputSize:     prevSize,
			OutputSize:    layer.Cols,
			MACOperations: macs,
			DataMovement:  cimData,
			CIMAdvantage:  float64(traditionalData) / float64(cimData),
		}
		prevSize = layer.Cols
	}

	return stats
}

// EnergyEstimate returns energy consumption estimate in pJ.
type EnergyEstimate struct {
	LayerName       string
	MACEnergy       float64 // pJ for MAC operations
	DataMoveEnergy  float64 // pJ for data movement (CIM advantage)
	TotalEnergy     float64
	TraditionalComp float64 // Comparison with traditional
}

// EstimateEnergy returns energy estimates for the stack.
func (s *Stack) EstimateEnergy() []EnergyEstimate {
	estimates := make([]EnergyEstimate, len(s.Layers))

	// IronLattice: ~0.001 pJ per MAC
	// Traditional: ~1-10 pJ per MAC
	cimMACEnergy := 0.001        // pJ
	traditionalMACEnergy := 5.0  // pJ
	dataMoveCostPerByte := 0.1   // pJ (for CIM, minimal)
	traditionalDataCost := 10.0  // pJ per byte

	for i, layer := range s.Layers {
		macs := layer.Rows * layer.Cols
		dataBytes := layer.Rows*4 + layer.Cols*4

		macEnergy := float64(macs) * cimMACEnergy
		moveEnergy := float64(dataBytes) * dataMoveCostPerByte
		total := macEnergy + moveEnergy

		traditionalMAC := float64(macs) * traditionalMACEnergy
		traditionalData := float64(layer.Rows*layer.Cols*4+dataBytes) * traditionalDataCost
		traditionalTotal := traditionalMAC + traditionalData

		estimates[i] = EnergyEstimate{
			LayerName:       layer.Name,
			MACEnergy:       macEnergy,
			DataMoveEnergy:  moveEnergy,
			TotalEnergy:     total,
			TraditionalComp: traditionalTotal / total,
		}
	}

	return estimates
}

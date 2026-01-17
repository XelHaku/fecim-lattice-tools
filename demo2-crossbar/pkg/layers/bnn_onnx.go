// Package layers provides neural network layer implementations for CIM simulation.
// This file implements Binary Neural Networks (BNN) and ONNX model import.
//
// Binary Neural Networks:
// - XNOR-popcount replaces MAC operations for 32× memory savings
// - 58× computational speedup on CPUs, 196.7 TOPS/W on FeFET
// - C-FeFET XNOR synapses achieve 60F² per cell density
// - Supports binarized weights (+1/-1) and activations
//
// ONNX Model Import:
// - ONNX IR representation and parsing
// - Graph traversal and operator mapping
// - Quantization passes for CIM deployment
// - Layer fusion optimizations
//
// References:
// - Monolithically Integrated C-FeFET XNOR Synapse (ACS AMI, 2024)
// - FINN: Binary Neural Network FPGA Compiler (Xilinx)
// - ONNC: Open Neural Network Compiler
// - Neuromorphic IR (Nature Communications, 2024)

package layers

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
)

// ============================================================================
// Binary Neural Network (BNN) Types
// ============================================================================

// BinaryWeight stores packed binary weights using 64-bit words.
// Each bit represents a weight: 1 = +1, 0 = -1.
type BinaryWeight struct {
	Words    []uint64 // Packed binary weights
	NumBits  int      // Total number of weights
	Rows     int      // Number of output neurons
	Cols     int      // Number of input features
	Alpha    float64  // Scaling factor (learned)
}

// NewBinaryWeight creates a new packed binary weight matrix.
func NewBinaryWeight(rows, cols int) *BinaryWeight {
	numBits := rows * cols
	numWords := (numBits + 63) / 64
	return &BinaryWeight{
		Words:   make([]uint64, numWords),
		NumBits: numBits,
		Rows:    rows,
		Cols:    cols,
		Alpha:   1.0,
	}
}

// SetWeight sets a binary weight at position (row, col).
// value > 0 sets bit to 1 (+1), value <= 0 sets bit to 0 (-1).
func (bw *BinaryWeight) SetWeight(row, col int, value float64) {
	idx := row*bw.Cols + col
	wordIdx := idx / 64
	bitIdx := uint(idx % 64)

	if value > 0 {
		bw.Words[wordIdx] |= (1 << bitIdx) // Set bit (+1)
	} else {
		bw.Words[wordIdx] &^= (1 << bitIdx) // Clear bit (-1)
	}
}

// GetWeight retrieves the binary weight at position (row, col).
// Returns +1.0 or -1.0.
func (bw *BinaryWeight) GetWeight(row, col int) float64 {
	idx := row*bw.Cols + col
	wordIdx := idx / 64
	bitIdx := uint(idx % 64)

	if (bw.Words[wordIdx] & (1 << bitIdx)) != 0 {
		return 1.0
	}
	return -1.0
}

// BinarizeWeights converts float weights to binary representation.
// Uses sign function: w_b = sign(w).
func (bw *BinaryWeight) BinarizeWeights(floatWeights [][]float64) {
	// Calculate optimal scaling factor alpha = mean(|W|)
	var sum float64
	count := 0
	for i, row := range floatWeights {
		for j, w := range row {
			sum += math.Abs(w)
			count++
			bw.SetWeight(i, j, w)
		}
	}
	if count > 0 {
		bw.Alpha = sum / float64(count)
	}
}

// BinaryActivation stores packed binary activations.
type BinaryActivation struct {
	Words   []uint64
	NumBits int
}

// NewBinaryActivation creates a new binary activation vector.
func NewBinaryActivation(size int) *BinaryActivation {
	numWords := (size + 63) / 64
	return &BinaryActivation{
		Words:   make([]uint64, numWords),
		NumBits: size,
	}
}

// Binarize converts float activations to binary using sign function.
// a_b = sign(a) where a > 0 → 1, a <= 0 → -1.
func (ba *BinaryActivation) Binarize(floatActivations []float64) {
	for i, a := range floatActivations {
		wordIdx := i / 64
		bitIdx := uint(i % 64)
		if a > 0 {
			ba.Words[wordIdx] |= (1 << bitIdx)
		} else {
			ba.Words[wordIdx] &^= (1 << bitIdx)
		}
	}
}

// XNORPopcount performs XNOR between binary weight and activation,
// then counts matching bits (popcount).
// Result = 2 * popcount(XNOR(w, a)) - n, where n is vector length.
func XNORPopcount(weights, activations []uint64, numBits int) int {
	var popcount int
	numWords := len(weights)

	for i := 0; i < numWords; i++ {
		// XNOR = NOT(XOR)
		xnor := ^(weights[i] ^ activations[i])
		popcount += bits.OnesCount64(xnor)
	}

	// Adjust for padding bits in last word
	lastWordBits := numBits % 64
	if lastWordBits > 0 {
		paddingBits := 64 - lastWordBits
		popcount -= paddingBits // Remove false matches from padding
	}

	// Convert popcount to signed accumulation: 2*popcount - n
	return 2*popcount - numBits
}

// ============================================================================
// BNN Layer Configuration
// ============================================================================

// BNNConfig configures a binary neural network layer.
type BNNConfig struct {
	InputSize       int     // Input feature dimension
	OutputSize      int     // Output feature dimension
	UseBatchNorm    bool    // Apply batch normalization
	UseThreshold    bool    // Use threshold instead of BN for binarization
	ThresholdValue  float64 // Threshold for binary activation
	LearnedAlpha    bool    // Learn per-channel scaling factors
}

// DefaultBNNConfig returns default BNN configuration.
func DefaultBNNConfig(inputSize, outputSize int) *BNNConfig {
	return &BNNConfig{
		InputSize:      inputSize,
		OutputSize:     outputSize,
		UseBatchNorm:   true,
		UseThreshold:   false,
		ThresholdValue: 0.0,
		LearnedAlpha:   true,
	}
}

// BNNLayer implements a binary neural network layer.
type BNNLayer struct {
	config      *BNNConfig
	weights     *BinaryWeight
	biases      []float64 // Float biases (optional)

	// Batch normalization parameters
	gamma       []float64 // Scale
	beta        []float64 // Shift
	runningMean []float64
	runningVar  []float64

	// Per-channel scaling
	alphas      []float64

	// Statistics
	MACOps      int64 // Number of binary MAC operations
	Throughput  float64 // GOPS
}

// NewBNNLayer creates a new binary neural network layer.
func NewBNNLayer(config *BNNConfig) *BNNLayer {
	layer := &BNNLayer{
		config:  config,
		weights: NewBinaryWeight(config.OutputSize, config.InputSize),
		biases:  make([]float64, config.OutputSize),
	}

	if config.UseBatchNorm {
		layer.gamma = make([]float64, config.OutputSize)
		layer.beta = make([]float64, config.OutputSize)
		layer.runningMean = make([]float64, config.OutputSize)
		layer.runningVar = make([]float64, config.OutputSize)
		for i := range layer.gamma {
			layer.gamma[i] = 1.0
			layer.runningVar[i] = 1.0
		}
	}

	if config.LearnedAlpha {
		layer.alphas = make([]float64, config.OutputSize)
		for i := range layer.alphas {
			layer.alphas[i] = 1.0
		}
	}

	return layer
}

// SetWeights initializes weights from float values.
func (l *BNNLayer) SetWeights(floatWeights [][]float64) {
	l.weights.BinarizeWeights(floatWeights)
}

// Forward performs binary forward pass using XNOR-popcount.
func (l *BNNLayer) Forward(input []float64) []float64 {
	// Binarize input activations
	binInput := NewBinaryActivation(len(input))
	binInput.Binarize(input)

	output := make([]float64, l.config.OutputSize)

	// Extract row weights and compute XNOR-popcount for each output
	for i := 0; i < l.config.OutputSize; i++ {
		// Get packed weights for this row
		rowWeights := l.getRowWeights(i)

		// XNOR-popcount
		result := XNORPopcount(rowWeights, binInput.Words, l.config.InputSize)

		// Scale by alpha
		alpha := l.weights.Alpha
		if l.config.LearnedAlpha && len(l.alphas) > i {
			alpha = l.alphas[i]
		}

		output[i] = float64(result) * alpha

		// Add bias
		if len(l.biases) > i {
			output[i] += l.biases[i]
		}

		l.MACOps++
	}

	// Apply batch normalization or thresholding
	if l.config.UseBatchNorm {
		output = l.applyBatchNorm(output)
	}

	return output
}

// getRowWeights extracts packed weights for a specific output row.
func (l *BNNLayer) getRowWeights(row int) []uint64 {
	startBit := row * l.config.InputSize
	numWords := (l.config.InputSize + 63) / 64
	rowWeights := make([]uint64, numWords)

	for i := 0; i < l.config.InputSize; i++ {
		srcIdx := startBit + i
		srcWord := srcIdx / 64
		srcBit := uint(srcIdx % 64)

		dstWord := i / 64
		dstBit := uint(i % 64)

		if (l.weights.Words[srcWord] & (1 << srcBit)) != 0 {
			rowWeights[dstWord] |= (1 << dstBit)
		}
	}

	return rowWeights
}

// applyBatchNorm applies batch normalization to output.
func (l *BNNLayer) applyBatchNorm(x []float64) []float64 {
	output := make([]float64, len(x))
	eps := 1e-5

	for i, val := range x {
		normalized := (val - l.runningMean[i]) / math.Sqrt(l.runningVar[i]+eps)
		output[i] = l.gamma[i]*normalized + l.beta[i]
	}

	return output
}

// BinarizeOutput applies sign function to produce binary output.
func (l *BNNLayer) BinarizeOutput(x []float64) []float64 {
	output := make([]float64, len(x))
	threshold := l.config.ThresholdValue

	for i, val := range x {
		if val > threshold {
			output[i] = 1.0
		} else {
			output[i] = -1.0
		}
	}

	return output
}

// ============================================================================
// FeFET BNN Crossbar Mapping
// ============================================================================

// FeFETBNNConfig configures FeFET-based BNN crossbar.
type FeFETBNNConfig struct {
	ArrayRows      int     // Crossbar rows
	ArrayCols      int     // Crossbar columns
	CellArea       float64 // Cell area in F² (60F² for C-FeFET XNOR)
	ReadLatency    float64 // Read latency in ns
	WriteLatency   float64 // Write latency in ns
	EnergyPerMAC   float64 // Energy per XNOR-popcount in fJ
}

// DefaultFeFETBNNConfig returns typical FeFET BNN parameters.
func DefaultFeFETBNNConfig() *FeFETBNNConfig {
	return &FeFETBNNConfig{
		ArrayRows:    512,
		ArrayCols:    512,
		CellArea:     60.0,   // 60F² per cell (C-FeFET XNOR synapse)
		ReadLatency:  10.0,   // 10 ns read
		WriteLatency: 100.0,  // 100 ns write
		EnergyPerMAC: 5.0,    // 5 fJ per binary MAC
	}
}

// FeFETBNNArray simulates a FeFET-based BNN crossbar array.
type FeFETBNNArray struct {
	config       *FeFETBNNConfig
	cells        [][]*FeFETBNNCell

	// Performance metrics
	TotalOps     int64
	TotalEnergy  float64 // Total energy in pJ
	Throughput   float64 // GOPS
	Efficiency   float64 // TOPS/W
}

// FeFETBNNCell represents a single FeFET XNOR synapse cell.
type FeFETBNNCell struct {
	Weight    bool    // Binary weight: true = +1, false = -1
	Threshold float64 // Ferroelectric polarization threshold
	Variance  float64 // Device-to-device variation
}

// NewFeFETBNNArray creates a new FeFET BNN crossbar array.
func NewFeFETBNNArray(config *FeFETBNNConfig) *FeFETBNNArray {
	cells := make([][]*FeFETBNNCell, config.ArrayRows)
	for i := range cells {
		cells[i] = make([]*FeFETBNNCell, config.ArrayCols)
		for j := range cells[i] {
			cells[i][j] = &FeFETBNNCell{
				Weight:    true,
				Threshold: 0.0,
				Variance:  0.02, // 2% variation
			}
		}
	}

	return &FeFETBNNArray{
		config: config,
		cells:  cells,
	}
}

// ProgramWeights programs binary weights into the FeFET array.
func (arr *FeFETBNNArray) ProgramWeights(weights *BinaryWeight) error {
	if weights.Rows > arr.config.ArrayRows || weights.Cols > arr.config.ArrayCols {
		return fmt.Errorf("weight matrix (%d×%d) exceeds array size (%d×%d)",
			weights.Rows, weights.Cols, arr.config.ArrayRows, arr.config.ArrayCols)
	}

	for i := 0; i < weights.Rows; i++ {
		for j := 0; j < weights.Cols; j++ {
			arr.cells[i][j].Weight = weights.GetWeight(i, j) > 0
		}
	}

	return nil
}

// XNORCompute performs in-memory XNOR computation.
func (arr *FeFETBNNArray) XNORCompute(input *BinaryActivation, outputRows int) []int {
	results := make([]int, outputRows)

	for row := 0; row < outputRows; row++ {
		var popcount int
		for col := 0; col < input.NumBits && col < arr.config.ArrayCols; col++ {
			// Get input bit
			wordIdx := col / 64
			bitIdx := uint(col % 64)
			inputBit := (input.Words[wordIdx] & (1 << bitIdx)) != 0

			// XNOR: same = 1, different = 0
			if inputBit == arr.cells[row][col].Weight {
				popcount++
			}
		}

		// Convert to signed: 2*popcount - n
		results[row] = 2*popcount - input.NumBits
		arr.TotalOps++
	}

	// Update energy
	arr.TotalEnergy += float64(outputRows) * arr.config.EnergyPerMAC / 1000.0 // pJ

	return results
}

// EstimatePerformance calculates performance metrics.
func (arr *FeFETBNNArray) EstimatePerformance(batchSize int) {
	// Throughput: GOPS
	opsPerInference := int64(arr.config.ArrayRows) * int64(arr.config.ArrayCols)
	totalOps := opsPerInference * int64(batchSize)
	latency := float64(batchSize) * arr.config.ReadLatency * 1e-9 // seconds
	arr.Throughput = float64(totalOps) / latency / 1e9 // GOPS

	// Energy efficiency: TOPS/W
	totalEnergy := float64(totalOps) * arr.config.EnergyPerMAC * 1e-15 // Joules
	power := totalEnergy / latency // Watts
	if power > 0 {
		arr.Efficiency = float64(totalOps) / 1e12 / power // TOPS/W
	}
}

// ============================================================================
// ONNX Model Import
// ============================================================================

// ONNXDataType represents ONNX tensor data types.
type ONNXDataType int

const (
	ONNX_FLOAT ONNXDataType = iota
	ONNX_UINT8
	ONNX_INT8
	ONNX_UINT16
	ONNX_INT16
	ONNX_INT32
	ONNX_INT64
	ONNX_FLOAT16
	ONNX_DOUBLE
	ONNX_BOOL
	ONNX_BFLOAT16
)

// ONNXTensorShape represents tensor dimensions.
type ONNXTensorShape struct {
	Dims []int64
}

// NumElements returns total number of elements.
func (s *ONNXTensorShape) NumElements() int64 {
	if len(s.Dims) == 0 {
		return 0
	}
	total := int64(1)
	for _, d := range s.Dims {
		if d > 0 {
			total *= d
		}
	}
	return total
}

// ONNXTensor represents an ONNX tensor.
type ONNXTensor struct {
	Name      string
	DataType  ONNXDataType
	Shape     *ONNXTensorShape
	RawData   []byte
	FloatData []float32
	Int32Data []int32
}

// ToFloat64 converts tensor data to float64 slice.
func (t *ONNXTensor) ToFloat64() []float64 {
	if len(t.FloatData) > 0 {
		result := make([]float64, len(t.FloatData))
		for i, v := range t.FloatData {
			result[i] = float64(v)
		}
		return result
	}

	if len(t.RawData) > 0 && t.DataType == ONNX_FLOAT {
		numFloats := len(t.RawData) / 4
		result := make([]float64, numFloats)
		for i := 0; i < numFloats; i++ {
			bits := binary.LittleEndian.Uint32(t.RawData[i*4:])
			result[i] = float64(math.Float32frombits(bits))
		}
		return result
	}

	return nil
}

// ONNXAttribute represents an ONNX node attribute.
type ONNXAttribute struct {
	Name   string
	Type   string // "int", "float", "string", "ints", "floats"
	Int    int64
	Float  float32
	String string
	Ints   []int64
	Floats []float32
}

// ONNXNode represents an ONNX graph node (operator).
type ONNXNode struct {
	Name       string
	OpType     string           // Conv, MatMul, Relu, etc.
	Inputs     []string         // Input tensor names
	Outputs    []string         // Output tensor names
	Attributes []*ONNXAttribute
}

// GetAttrInt retrieves an integer attribute.
func (n *ONNXNode) GetAttrInt(name string) (int64, bool) {
	for _, attr := range n.Attributes {
		if attr.Name == name && attr.Type == "int" {
			return attr.Int, true
		}
	}
	return 0, false
}

// GetAttrInts retrieves an integer array attribute.
func (n *ONNXNode) GetAttrInts(name string) ([]int64, bool) {
	for _, attr := range n.Attributes {
		if attr.Name == name && attr.Type == "ints" {
			return attr.Ints, true
		}
	}
	return nil, false
}

// ONNXGraph represents an ONNX computation graph.
type ONNXGraph struct {
	Name         string
	Nodes        []*ONNXNode
	Inputs       []*ONNXTensor // Graph inputs (with shapes)
	Outputs      []*ONNXTensor // Graph outputs
	Initializers []*ONNXTensor // Weight tensors
}

// ONNXModel represents a complete ONNX model.
type ONNXModel struct {
	IRVersion     int64
	ProducerName  string
	ProducerVer   string
	Domain        string
	ModelVersion  int64
	Graph         *ONNXGraph
	OpsetImports  map[string]int64
}

// NewONNXModel creates an empty ONNX model.
func NewONNXModel() *ONNXModel {
	return &ONNXModel{
		IRVersion:    9,
		OpsetImports: make(map[string]int64),
		Graph: &ONNXGraph{
			Nodes:        make([]*ONNXNode, 0),
			Inputs:       make([]*ONNXTensor, 0),
			Outputs:      make([]*ONNXTensor, 0),
			Initializers: make([]*ONNXTensor, 0),
		},
	}
}

// ============================================================================
// ONNX to CIM Compiler
// ============================================================================

// CIMOpType represents CIM-compatible operations.
type CIMOpType int

const (
	CIM_MVM CIMOpType = iota // Matrix-vector multiply
	CIM_ACTIVATION           // Activation function
	CIM_POOL                 // Pooling
	CIM_NORM                 // Normalization
	CIM_BINARY               // Binary operation
	CIM_CONCAT               // Concatenation
	CIM_RESHAPE              // Reshape (no compute)
)

// CIMOp represents a CIM-mapped operation.
type CIMOp struct {
	Type         CIMOpType
	Name         string
	InputShapes  [][]int64
	OutputShape  []int64
	Weights      *ONNXTensor   // Weight tensor if applicable
	Config       map[string]interface{}

	// Quantization info
	WeightScale  float64
	WeightZero   int
	ActScale     float64
	ActZero      int
	OutputBits   int
}

// CIMCompilerConfig configures the ONNX-to-CIM compiler.
type CIMCompilerConfig struct {
	TargetBits     int     // Target weight bit-width (2-8)
	ActivationBits int     // Activation bit-width
	FuseConvBN     bool    // Fuse Conv+BatchNorm
	FuseConvRelu   bool    // Fuse Conv+ReLU
	OptimizeMemory bool    // Optimize memory layout
	BinaryMode     bool    // Use binary weights
	CrossbarRows   int     // Target crossbar rows
	CrossbarCols   int     // Target crossbar columns
}

// DefaultCIMCompilerConfig returns default compiler configuration.
func DefaultCIMCompilerConfig() *CIMCompilerConfig {
	return &CIMCompilerConfig{
		TargetBits:     6,
		ActivationBits: 8,
		FuseConvBN:     true,
		FuseConvRelu:   true,
		OptimizeMemory: true,
		BinaryMode:     false,
		CrossbarRows:   256,
		CrossbarCols:   256,
	}
}

// CIMCompiler compiles ONNX models for CIM deployment.
type CIMCompiler struct {
	config       *CIMCompilerConfig
	model        *ONNXModel
	ops          []*CIMOp
	tensorShapes map[string][]int64

	// Fusion tracking
	fusedNodes   map[string]bool

	// Statistics
	TotalMACs    int64
	TotalParams  int64
	CrossbarUtil float64 // Crossbar utilization
}

// NewCIMCompiler creates a new CIM compiler.
func NewCIMCompiler(config *CIMCompilerConfig) *CIMCompiler {
	return &CIMCompiler{
		config:       config,
		ops:          make([]*CIMOp, 0),
		tensorShapes: make(map[string][]int64),
		fusedNodes:   make(map[string]bool),
	}
}

// LoadModel loads an ONNX model for compilation.
func (c *CIMCompiler) LoadModel(model *ONNXModel) {
	c.model = model

	// Extract input shapes
	for _, input := range model.Graph.Inputs {
		if input.Shape != nil {
			c.tensorShapes[input.Name] = input.Shape.Dims
		}
	}

	// Extract initializer shapes
	for _, init := range model.Graph.Initializers {
		if init.Shape != nil {
			c.tensorShapes[init.Name] = init.Shape.Dims
		}
	}
}

// Compile compiles the ONNX model to CIM operations.
func (c *CIMCompiler) Compile() error {
	if c.model == nil {
		return fmt.Errorf("no model loaded")
	}

	// Pass 1: Layer fusion
	if c.config.FuseConvBN || c.config.FuseConvRelu {
		c.fuseOperators()
	}

	// Pass 2: Convert operators to CIM ops
	for _, node := range c.model.Graph.Nodes {
		if c.fusedNodes[node.Name] {
			continue // Skip fused nodes
		}

		cimOp, err := c.convertNode(node)
		if err != nil {
			return fmt.Errorf("converting node %s: %w", node.Name, err)
		}

		if cimOp != nil {
			c.ops = append(c.ops, cimOp)
		}
	}

	// Pass 3: Quantization
	c.quantizeWeights()

	// Pass 4: Calculate statistics
	c.calculateStats()

	return nil
}

// fuseOperators performs operator fusion passes.
func (c *CIMCompiler) fuseOperators() {
	nodeMap := make(map[string]*ONNXNode)
	for _, node := range c.model.Graph.Nodes {
		nodeMap[node.Name] = node
		for _, out := range node.Outputs {
			nodeMap[out] = node
		}
	}

	for i, node := range c.model.Graph.Nodes {
		// Conv + BatchNorm fusion
		if c.config.FuseConvBN && node.OpType == "Conv" {
			if i+1 < len(c.model.Graph.Nodes) {
				next := c.model.Graph.Nodes[i+1]
				if next.OpType == "BatchNormalization" && len(next.Inputs) > 0 {
					// Check if BN input matches Conv output
					if len(node.Outputs) > 0 && node.Outputs[0] == next.Inputs[0] {
						c.fusedNodes[next.Name] = true
						// Mark for fusion in conversion
					}
				}
			}
		}

		// Conv + ReLU fusion
		if c.config.FuseConvRelu && node.OpType == "Conv" {
			if i+1 < len(c.model.Graph.Nodes) {
				next := c.model.Graph.Nodes[i+1]
				if next.OpType == "Relu" && len(next.Inputs) > 0 {
					if len(node.Outputs) > 0 && node.Outputs[0] == next.Inputs[0] {
						c.fusedNodes[next.Name] = true
					}
				}
			}
		}
	}
}

// convertNode converts an ONNX node to a CIM operation.
func (c *CIMCompiler) convertNode(node *ONNXNode) (*CIMOp, error) {
	switch node.OpType {
	case "Conv":
		return c.convertConv(node)
	case "MatMul", "Gemm":
		return c.convertMatMul(node)
	case "Relu", "Sigmoid", "Tanh", "LeakyRelu":
		return c.convertActivation(node)
	case "MaxPool", "AveragePool", "GlobalAveragePool":
		return c.convertPool(node)
	case "BatchNormalization", "LayerNormalization":
		return c.convertNorm(node)
	case "Add", "Sub", "Mul", "Div":
		return c.convertBinary(node)
	case "Concat":
		return c.convertConcat(node)
	case "Reshape", "Flatten", "Squeeze", "Unsqueeze":
		return c.convertReshape(node)
	default:
		// Unknown op - try to pass through
		return &CIMOp{
			Type:   CIM_RESHAPE, // Treat as no-op
			Name:   node.Name,
			Config: map[string]interface{}{"op": node.OpType},
		}, nil
	}
}

// convertConv converts Conv operator to CIM MVM.
func (c *CIMCompiler) convertConv(node *ONNXNode) (*CIMOp, error) {
	// Get weight tensor
	var weights *ONNXTensor
	if len(node.Inputs) > 1 {
		for _, init := range c.model.Graph.Initializers {
			if init.Name == node.Inputs[1] {
				weights = init
				break
			}
		}
	}

	// Get convolution attributes
	kernelShape, _ := node.GetAttrInts("kernel_shape")
	strides, _ := node.GetAttrInts("strides")
	pads, _ := node.GetAttrInts("pads")
	dilations, _ := node.GetAttrInts("dilations")
	group, _ := node.GetAttrInt("group")

	if group == 0 {
		group = 1
	}

	return &CIMOp{
		Type:    CIM_MVM,
		Name:    node.Name,
		Weights: weights,
		Config: map[string]interface{}{
			"kernel_shape": kernelShape,
			"strides":      strides,
			"pads":         pads,
			"dilations":    dilations,
			"group":        group,
			"im2col":       true, // Use im2col for crossbar mapping
		},
		OutputBits: c.config.ActivationBits,
	}, nil
}

// convertMatMul converts MatMul/Gemm to CIM MVM.
func (c *CIMCompiler) convertMatMul(node *ONNXNode) (*CIMOp, error) {
	var weights *ONNXTensor
	if len(node.Inputs) > 1 {
		for _, init := range c.model.Graph.Initializers {
			if init.Name == node.Inputs[1] {
				weights = init
				break
			}
		}
	}

	return &CIMOp{
		Type:       CIM_MVM,
		Name:       node.Name,
		Weights:    weights,
		Config:     map[string]interface{}{},
		OutputBits: c.config.ActivationBits,
	}, nil
}

// convertActivation converts activation functions.
func (c *CIMCompiler) convertActivation(node *ONNXNode) (*CIMOp, error) {
	config := map[string]interface{}{
		"type": node.OpType,
	}

	// LeakyReLU alpha
	if node.OpType == "LeakyRelu" {
		for _, attr := range node.Attributes {
			if attr.Name == "alpha" {
				config["alpha"] = attr.Float
			}
		}
	}

	return &CIMOp{
		Type:   CIM_ACTIVATION,
		Name:   node.Name,
		Config: config,
	}, nil
}

// convertPool converts pooling operations.
func (c *CIMCompiler) convertPool(node *ONNXNode) (*CIMOp, error) {
	kernelShape, _ := node.GetAttrInts("kernel_shape")
	strides, _ := node.GetAttrInts("strides")
	pads, _ := node.GetAttrInts("pads")

	return &CIMOp{
		Type: CIM_POOL,
		Name: node.Name,
		Config: map[string]interface{}{
			"type":         node.OpType,
			"kernel_shape": kernelShape,
			"strides":      strides,
			"pads":         pads,
		},
	}, nil
}

// convertNorm converts normalization operations.
func (c *CIMCompiler) convertNorm(node *ONNXNode) (*CIMOp, error) {
	return &CIMOp{
		Type: CIM_NORM,
		Name: node.Name,
		Config: map[string]interface{}{
			"type": node.OpType,
		},
	}, nil
}

// convertBinary converts binary operations (Add, Mul, etc.).
func (c *CIMCompiler) convertBinary(node *ONNXNode) (*CIMOp, error) {
	return &CIMOp{
		Type: CIM_BINARY,
		Name: node.Name,
		Config: map[string]interface{}{
			"op": node.OpType,
		},
	}, nil
}

// convertConcat converts concatenation.
func (c *CIMCompiler) convertConcat(node *ONNXNode) (*CIMOp, error) {
	axis, _ := node.GetAttrInt("axis")

	return &CIMOp{
		Type: CIM_CONCAT,
		Name: node.Name,
		Config: map[string]interface{}{
			"axis": axis,
		},
	}, nil
}

// convertReshape converts reshape operations (no compute).
func (c *CIMCompiler) convertReshape(node *ONNXNode) (*CIMOp, error) {
	return &CIMOp{
		Type:   CIM_RESHAPE,
		Name:   node.Name,
		Config: map[string]interface{}{},
	}, nil
}

// quantizeWeights applies quantization to all weight tensors.
func (c *CIMCompiler) quantizeWeights() {
	for _, op := range c.ops {
		if op.Weights == nil {
			continue
		}

		weights := op.Weights.ToFloat64()
		if len(weights) == 0 {
			continue
		}

		// Find min/max
		minVal, maxVal := weights[0], weights[0]
		for _, w := range weights {
			if w < minVal {
				minVal = w
			}
			if w > maxVal {
				maxVal = w
			}
		}

		// Calculate scale and zero point
		if c.config.BinaryMode {
			op.WeightScale = 1.0
			op.WeightZero = 0
		} else {
			levels := float64(int(1) << c.config.TargetBits)
			op.WeightScale = (maxVal - minVal) / (levels - 1)
			if op.WeightScale > 0 {
				op.WeightZero = int(-minVal / op.WeightScale)
			}
		}
	}
}

// calculateStats calculates compilation statistics.
func (c *CIMCompiler) calculateStats() {
	for _, op := range c.ops {
		if op.Type == CIM_MVM && op.Weights != nil {
			// Calculate MACs
			if op.Weights.Shape != nil {
				numElements := op.Weights.Shape.NumElements()
				c.TotalMACs += numElements
				c.TotalParams += numElements
			}
		}
	}

	// Calculate crossbar utilization
	totalCells := int64(c.config.CrossbarRows) * int64(c.config.CrossbarCols)
	if totalCells > 0 && c.TotalParams > 0 {
		c.CrossbarUtil = float64(c.TotalParams) / float64(totalCells) * 100.0
	}
}

// ExportCrossbarWeights exports quantized weights for crossbar programming.
func (c *CIMCompiler) ExportCrossbarWeights() map[string][]float64 {
	result := make(map[string][]float64)

	for _, op := range c.ops {
		if op.Weights == nil {
			continue
		}

		weights := op.Weights.ToFloat64()
		if len(weights) == 0 {
			continue
		}

		// Quantize weights
		quantized := make([]float64, len(weights))
		levels := float64(int(1) << c.config.TargetBits)

		for i, w := range weights {
			if c.config.BinaryMode {
				// Binary: +1 or -1
				if w >= 0 {
					quantized[i] = 1.0
				} else {
					quantized[i] = -1.0
				}
			} else {
				// Multi-bit quantization
				q := math.Round((w/op.WeightScale + float64(op.WeightZero)))
				q = math.Max(0, math.Min(levels-1, q))
				quantized[i] = q
			}
		}

		result[op.Name] = quantized
	}

	return result
}

// ============================================================================
// BNN Training Support
// ============================================================================

// StraightThroughEstimator implements STE for gradient estimation.
type StraightThroughEstimator struct{}

// Forward applies sign function for binarization.
func (ste *StraightThroughEstimator) Forward(x float64) float64 {
	if x >= 0 {
		return 1.0
	}
	return -1.0
}

// Backward passes gradient through (STE).
// Gradient is clipped to [-1, 1] range.
func (ste *StraightThroughEstimator) Backward(grad, x float64) float64 {
	if x >= -1 && x <= 1 {
		return grad
	}
	return 0.0
}

// BNNTrainer implements BNN-specific training utilities.
type BNNTrainer struct {
	LearningRate float64
	WeightDecay  float64

	// Real-valued weights (shadow weights)
	RealWeights  [][]float64

	// Momentum
	Momentum     float64
	Velocity     [][]float64

	ste          *StraightThroughEstimator
}

// NewBNNTrainer creates a new BNN trainer.
func NewBNNTrainer(lr, wd, momentum float64) *BNNTrainer {
	return &BNNTrainer{
		LearningRate: lr,
		WeightDecay:  wd,
		Momentum:     momentum,
		ste:          &StraightThroughEstimator{},
	}
}

// InitWeights initializes real-valued shadow weights.
func (t *BNNTrainer) InitWeights(rows, cols int) {
	t.RealWeights = make([][]float64, rows)
	t.Velocity = make([][]float64, rows)

	// Xavier initialization
	scale := math.Sqrt(2.0 / float64(rows+cols))

	for i := range t.RealWeights {
		t.RealWeights[i] = make([]float64, cols)
		t.Velocity[i] = make([]float64, cols)
		for j := range t.RealWeights[i] {
			// Uniform in [-scale, scale]
			t.RealWeights[i][j] = (2.0*float64(i*cols+j)/float64(rows*cols) - 1.0) * scale
		}
	}
}

// BinarizeForward binarizes weights for forward pass.
func (t *BNNTrainer) BinarizeForward() *BinaryWeight {
	rows := len(t.RealWeights)
	cols := 0
	if rows > 0 {
		cols = len(t.RealWeights[0])
	}

	bw := NewBinaryWeight(rows, cols)
	bw.BinarizeWeights(t.RealWeights)
	return bw
}

// UpdateWeights applies gradients to real-valued weights using STE.
func (t *BNNTrainer) UpdateWeights(gradients [][]float64) {
	for i := range t.RealWeights {
		for j := range t.RealWeights[i] {
			// Apply STE
			grad := t.ste.Backward(gradients[i][j], t.RealWeights[i][j])

			// Momentum update
			t.Velocity[i][j] = t.Momentum*t.Velocity[i][j] - t.LearningRate*grad

			// Weight update with decay
			t.RealWeights[i][j] += t.Velocity[i][j] - t.WeightDecay*t.RealWeights[i][j]

			// Clamp to [-1, 1]
			t.RealWeights[i][j] = math.Max(-1.0, math.Min(1.0, t.RealWeights[i][j]))
		}
	}
}

// ============================================================================
// ONNX Model Builder (for testing/export)
// ============================================================================

// ONNXModelBuilder helps construct ONNX models programmatically.
type ONNXModelBuilder struct {
	model      *ONNXModel
	nodeCount  int
}

// NewONNXModelBuilder creates a new model builder.
func NewONNXModelBuilder(name string) *ONNXModelBuilder {
	model := NewONNXModel()
	model.Graph.Name = name
	model.OpsetImports[""] = 13 // Default opset

	return &ONNXModelBuilder{
		model: model,
	}
}

// AddInput adds an input tensor to the graph.
func (b *ONNXModelBuilder) AddInput(name string, shape []int64, dtype ONNXDataType) {
	b.model.Graph.Inputs = append(b.model.Graph.Inputs, &ONNXTensor{
		Name:     name,
		DataType: dtype,
		Shape:    &ONNXTensorShape{Dims: shape},
	})
}

// AddOutput adds an output tensor to the graph.
func (b *ONNXModelBuilder) AddOutput(name string, shape []int64, dtype ONNXDataType) {
	b.model.Graph.Outputs = append(b.model.Graph.Outputs, &ONNXTensor{
		Name:     name,
		DataType: dtype,
		Shape:    &ONNXTensorShape{Dims: shape},
	})
}

// AddInitializer adds a weight tensor.
func (b *ONNXModelBuilder) AddInitializer(name string, shape []int64, data []float32) {
	b.model.Graph.Initializers = append(b.model.Graph.Initializers, &ONNXTensor{
		Name:      name,
		DataType:  ONNX_FLOAT,
		Shape:     &ONNXTensorShape{Dims: shape},
		FloatData: data,
	})
}

// AddConv adds a Conv node.
func (b *ONNXModelBuilder) AddConv(input, weight, bias, output string, kernelShape, strides, pads []int64) {
	b.nodeCount++
	inputs := []string{input, weight}
	if bias != "" {
		inputs = append(inputs, bias)
	}

	b.model.Graph.Nodes = append(b.model.Graph.Nodes, &ONNXNode{
		Name:    fmt.Sprintf("conv_%d", b.nodeCount),
		OpType:  "Conv",
		Inputs:  inputs,
		Outputs: []string{output},
		Attributes: []*ONNXAttribute{
			{Name: "kernel_shape", Type: "ints", Ints: kernelShape},
			{Name: "strides", Type: "ints", Ints: strides},
			{Name: "pads", Type: "ints", Ints: pads},
		},
	})
}

// AddMatMul adds a MatMul node.
func (b *ONNXModelBuilder) AddMatMul(inputA, inputB, output string) {
	b.nodeCount++
	b.model.Graph.Nodes = append(b.model.Graph.Nodes, &ONNXNode{
		Name:    fmt.Sprintf("matmul_%d", b.nodeCount),
		OpType:  "MatMul",
		Inputs:  []string{inputA, inputB},
		Outputs: []string{output},
	})
}

// AddRelu adds a ReLU node.
func (b *ONNXModelBuilder) AddRelu(input, output string) {
	b.nodeCount++
	b.model.Graph.Nodes = append(b.model.Graph.Nodes, &ONNXNode{
		Name:    fmt.Sprintf("relu_%d", b.nodeCount),
		OpType:  "Relu",
		Inputs:  []string{input},
		Outputs: []string{output},
	})
}

// AddBatchNorm adds a BatchNormalization node.
func (b *ONNXModelBuilder) AddBatchNorm(input, scale, bias, mean, var_, output string) {
	b.nodeCount++
	b.model.Graph.Nodes = append(b.model.Graph.Nodes, &ONNXNode{
		Name:    fmt.Sprintf("bn_%d", b.nodeCount),
		OpType:  "BatchNormalization",
		Inputs:  []string{input, scale, bias, mean, var_},
		Outputs: []string{output},
	})
}

// Build returns the constructed ONNX model.
func (b *ONNXModelBuilder) Build() *ONNXModel {
	return b.model
}

// GetOps returns compiled CIM operations.
func (c *CIMCompiler) GetOps() []*CIMOp {
	return c.ops
}

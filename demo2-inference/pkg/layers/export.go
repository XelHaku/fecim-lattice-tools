// Package layers provides neural network layer implementations for crossbar-based CIM.
// export.go implements model export utilities for hardware deployment.
//
// Export formats:
// - ONNX-like structure for interoperability
// - Crossbar mapping files for hardware
// - Quantized binary for embedded systems
// - C header files for microcontroller deployment
// - Verilog/VHDL parameters for FPGA

package layers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// ExportFormat defines the export format type
type ExportFormat string

const (
	FormatJSON           ExportFormat = "json"
	FormatBinary         ExportFormat = "binary"
	FormatCHeader        ExportFormat = "c_header"
	FormatCrossbarMap    ExportFormat = "crossbar_map"
	FormatONNXLike       ExportFormat = "onnx_like"
	FormatVerilogParams  ExportFormat = "verilog"
	FormatQuantizedPacked ExportFormat = "quantized_packed"
)

// ModelExporter exports models for various deployment targets
type ModelExporter struct {
	Model       *ModelCheckpoint
	Config      *ExportConfig
	CrossbarCfg *CrossbarCheckpoint
}

// ExportConfig configures export behavior
type ExportConfig struct {
	Format       ExportFormat
	Quantize     bool
	QuantBits    int
	PackWeights  bool
	SplitLayers  bool
	IncludeMetadata bool
	Compress     bool
}

// DefaultExportConfig returns default export configuration
func DefaultExportConfig() *ExportConfig {
	return &ExportConfig{
		Format:       FormatJSON,
		Quantize:     true,
		QuantBits:    8,
		PackWeights:  false,
		SplitLayers:  false,
		IncludeMetadata: true,
		Compress:     false,
	}
}

// NewModelExporter creates a new model exporter
func NewModelExporter(model *ModelCheckpoint, config *ExportConfig) *ModelExporter {
	if config == nil {
		config = DefaultExportConfig()
	}
	return &ModelExporter{
		Model:  model,
		Config: config,
	}
}

// SetCrossbarConfig sets crossbar configuration for hardware export
func (me *ModelExporter) SetCrossbarConfig(cfg *CrossbarCheckpoint) {
	me.CrossbarCfg = cfg
}

// Export exports the model to the specified format
func (me *ModelExporter) Export(filename string) error {
	switch me.Config.Format {
	case FormatJSON:
		return me.exportJSON(filename)
	case FormatBinary:
		return me.exportBinary(filename)
	case FormatCHeader:
		return me.exportCHeader(filename)
	case FormatCrossbarMap:
		return me.exportCrossbarMap(filename)
	case FormatONNXLike:
		return me.exportONNXLike(filename)
	case FormatVerilogParams:
		return me.exportVerilogParams(filename)
	case FormatQuantizedPacked:
		return me.exportQuantizedPacked(filename)
	default:
		return fmt.Errorf("unsupported export format: %s", me.Config.Format)
	}
}

// ============================================================================
// JSON Export
// ============================================================================

func (me *ModelExporter) exportJSON(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(me.Model)
}

// ============================================================================
// Binary Export
// ============================================================================

func (me *ModelExporter) exportBinary(filename string) error {
	return SaveBinary(filename, me.Model, me.Config.Compress)
}

// ============================================================================
// C Header Export (for embedded deployment)
// ============================================================================

func (me *ModelExporter) exportCHeader(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header guard
	guard := strings.ToUpper(strings.ReplaceAll(filename, ".", "_"))
	guard = strings.ToUpper(strings.ReplaceAll(guard, "/", "_"))
	fmt.Fprintf(file, "#ifndef %s\n", guard)
	fmt.Fprintf(file, "#define %s\n\n", guard)

	// Write includes
	fmt.Fprintf(file, "#include <stdint.h>\n\n")

	// Write model info
	fmt.Fprintf(file, "// Model exported from IronLattice CIM Simulator\n")
	fmt.Fprintf(file, "// Layers: %d\n", len(me.Model.Layers))
	fmt.Fprintf(file, "// Quantization: %d bits\n\n", me.Config.QuantBits)

	// Export each layer
	for i, layer := range me.Model.Layers {
		layerName := strings.ToUpper(strings.ReplaceAll(layer.Name, "-", "_"))

		// Write layer info comment
		fmt.Fprintf(file, "// Layer %d: %s (%s)\n", i, layer.Name, layer.Type)
		if len(layer.Shape) >= 2 {
			fmt.Fprintf(file, "// Shape: [%d, %d]\n", layer.Shape[0], layer.Shape[1])
		}

		// Write weights
		if len(layer.Weights) > 0 {
			rows := len(layer.Weights)
			cols := 0
			if rows > 0 {
				cols = len(layer.Weights[0])
			}

			if me.Config.Quantize {
				fmt.Fprintf(file, "static const int8_t %s_WEIGHTS[%d][%d] = {\n", layerName, rows, cols)
				for r := 0; r < rows; r++ {
					fmt.Fprintf(file, "    {")
					for c := 0; c < cols; c++ {
						qval := quantizeToInt8(layer.Weights[r][c], me.Config.QuantBits)
						if c > 0 {
							fmt.Fprintf(file, ", ")
						}
						fmt.Fprintf(file, "%d", qval)
					}
					fmt.Fprintf(file, "}")
					if r < rows-1 {
						fmt.Fprintf(file, ",")
					}
					fmt.Fprintf(file, "\n")
				}
				fmt.Fprintf(file, "};\n\n")
			} else {
				fmt.Fprintf(file, "static const float %s_WEIGHTS[%d][%d] = {\n", layerName, rows, cols)
				for r := 0; r < rows; r++ {
					fmt.Fprintf(file, "    {")
					for c := 0; c < cols; c++ {
						if c > 0 {
							fmt.Fprintf(file, ", ")
						}
						fmt.Fprintf(file, "%.6ff", layer.Weights[r][c])
					}
					fmt.Fprintf(file, "}")
					if r < rows-1 {
						fmt.Fprintf(file, ",")
					}
					fmt.Fprintf(file, "\n")
				}
				fmt.Fprintf(file, "};\n\n")
			}
		}

		// Write biases
		if len(layer.Biases) > 0 {
			if me.Config.Quantize {
				fmt.Fprintf(file, "static const int16_t %s_BIASES[%d] = {", layerName, len(layer.Biases))
				for j, b := range layer.Biases {
					qval := quantizeToInt16(b, me.Config.QuantBits)
					if j > 0 {
						fmt.Fprintf(file, ", ")
					}
					fmt.Fprintf(file, "%d", qval)
				}
				fmt.Fprintf(file, "};\n\n")
			} else {
				fmt.Fprintf(file, "static const float %s_BIASES[%d] = {", layerName, len(layer.Biases))
				for j, b := range layer.Biases {
					if j > 0 {
						fmt.Fprintf(file, ", ")
					}
					fmt.Fprintf(file, "%.6ff", b)
				}
				fmt.Fprintf(file, "};\n\n")
			}
		}
	}

	// Write helper macros
	fmt.Fprintf(file, "// Model dimensions\n")
	for _, layer := range me.Model.Layers {
		layerName := strings.ToUpper(strings.ReplaceAll(layer.Name, "-", "_"))
		if len(layer.Shape) >= 2 {
			fmt.Fprintf(file, "#define %s_ROWS %d\n", layerName, layer.Shape[0])
			fmt.Fprintf(file, "#define %s_COLS %d\n", layerName, layer.Shape[1])
		}
	}

	// Close header guard
	fmt.Fprintf(file, "\n#endif // %s\n", guard)

	return nil
}

// ============================================================================
// Crossbar Mapping Export
// ============================================================================

// CrossbarMappingFile represents a crossbar hardware mapping
type CrossbarMappingFile struct {
	Version     string           `json:"version"`
	NumArrays   int              `json:"num_arrays"`
	ArraySize   int              `json:"array_size"`
	ADCBits     int              `json:"adc_bits"`
	DACBits     int              `json:"dac_bits"`
	WeightBits  int              `json:"weight_bits"`
	Mappings    []ArrayMapping   `json:"mappings"`
	Connections []ArrayConnection `json:"connections"`
}

// ArrayMapping describes how a layer maps to crossbar arrays
type ArrayMapping struct {
	LayerName    string      `json:"layer_name"`
	ArrayIndices []int       `json:"array_indices"`
	TileRow      int         `json:"tile_row"`
	TileCol      int         `json:"tile_col"`
	Conductances [][]float64 `json:"conductances"`
}

// ArrayConnection describes connections between arrays
type ArrayConnection struct {
	FromArray int    `json:"from_array"`
	ToArray   int    `json:"to_array"`
	Type      string `json:"type"` // "direct", "accumulate", "concat"
}

func (me *ModelExporter) exportCrossbarMap(filename string) error {
	if me.CrossbarCfg == nil {
		return fmt.Errorf("crossbar configuration required for crossbar mapping export")
	}

	arraySize := me.CrossbarCfg.ArraySize
	mappingFile := &CrossbarMappingFile{
		Version:    "1.0",
		ArraySize:  arraySize,
		ADCBits:    me.CrossbarCfg.ADCBits,
		DACBits:    me.CrossbarCfg.DACBits,
		WeightBits: me.CrossbarCfg.WeightBits,
		Mappings:   make([]ArrayMapping, 0),
		Connections: make([]ArrayConnection, 0),
	}

	arrayIdx := 0

	// Map each layer to crossbar arrays
	for _, layer := range me.Model.Layers {
		if len(layer.Weights) == 0 {
			continue
		}

		rows := len(layer.Weights)
		cols := len(layer.Weights[0])

		// Calculate number of tiles needed
		numRowTiles := (rows + arraySize - 1) / arraySize
		numColTiles := (cols + arraySize - 1) / arraySize

		layerArrays := make([]int, 0)

		// Create mappings for each tile
		for tr := 0; tr < numRowTiles; tr++ {
			for tc := 0; tc < numColTiles; tc++ {
				// Extract tile weights
				startRow := tr * arraySize
				endRow := startRow + arraySize
				if endRow > rows {
					endRow = rows
				}

				startCol := tc * arraySize
				endCol := startCol + arraySize
				if endCol > cols {
					endCol = cols
				}

				// Convert to conductances
				conductances := make([][]float64, arraySize)
				for i := 0; i < arraySize; i++ {
					conductances[i] = make([]float64, arraySize)
					srcRow := startRow + i
					for j := 0; j < arraySize; j++ {
						srcCol := startCol + j
						if srcRow < endRow && srcCol < endCol {
							// Map weight to conductance
							w := layer.Weights[srcRow][srcCol]
							conductances[i][j] = weightToConductance(w, me.CrossbarCfg)
						} else {
							conductances[i][j] = me.CrossbarCfg.MinCond
						}
					}
				}

				mapping := ArrayMapping{
					LayerName:    layer.Name,
					ArrayIndices: []int{arrayIdx},
					TileRow:      tr,
					TileCol:      tc,
					Conductances: conductances,
				}

				mappingFile.Mappings = append(mappingFile.Mappings, mapping)
				layerArrays = append(layerArrays, arrayIdx)
				arrayIdx++
			}
		}

		// Add connections for multi-tile layers
		if len(layerArrays) > 1 {
			for i := 1; i < len(layerArrays); i++ {
				mappingFile.Connections = append(mappingFile.Connections, ArrayConnection{
					FromArray: layerArrays[i-1],
					ToArray:   layerArrays[i],
					Type:      "accumulate",
				})
			}
		}
	}

	mappingFile.NumArrays = arrayIdx

	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(mappingFile)
}

// ============================================================================
// ONNX-like Export
// ============================================================================

// ONNXLikeModel represents an ONNX-compatible model structure
type ONNXLikeModel struct {
	IRVersion    int64          `json:"ir_version"`
	ProducerName string         `json:"producer_name"`
	Graph        ONNXLikeGraph  `json:"graph"`
}

// ONNXLikeGraph represents the computation graph
type ONNXLikeGraph struct {
	Nodes        []ONNXLikeNode   `json:"nodes"`
	Inputs       []ONNXLikeTensor `json:"inputs"`
	Outputs      []ONNXLikeTensor `json:"outputs"`
	Initializers []ONNXLikeTensor `json:"initializers"`
}

// ONNXLikeNode represents a computation node
type ONNXLikeNode struct {
	Name       string            `json:"name"`
	OpType     string            `json:"op_type"`
	Inputs     []string          `json:"inputs"`
	Outputs    []string          `json:"outputs"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// ONNXLikeTensor represents a tensor
type ONNXLikeTensor struct {
	Name     string    `json:"name"`
	DataType string    `json:"data_type"`
	Dims     []int     `json:"dims"`
	Data     []float64 `json:"data,omitempty"`
}

func (me *ModelExporter) exportONNXLike(filename string) error {
	model := &ONNXLikeModel{
		IRVersion:    8,
		ProducerName: "IronLattice-CIM-Exporter",
		Graph: ONNXLikeGraph{
			Nodes:        make([]ONNXLikeNode, 0),
			Inputs:       make([]ONNXLikeTensor, 0),
			Outputs:      make([]ONNXLikeTensor, 0),
			Initializers: make([]ONNXLikeTensor, 0),
		},
	}

	// Add input tensor
	if len(me.Model.Layers) > 0 && len(me.Model.Layers[0].Shape) >= 2 {
		inputDim := me.Model.Layers[0].Shape[1] // Input features
		model.Graph.Inputs = append(model.Graph.Inputs, ONNXLikeTensor{
			Name:     "input",
			DataType: "float32",
			Dims:     []int{1, inputDim},
		})
	}

	prevOutput := "input"

	// Add nodes for each layer
	for i, layer := range me.Model.Layers {
		nodeName := fmt.Sprintf("node_%d_%s", i, layer.Name)
		weightName := fmt.Sprintf("%s_weight", layer.Name)
		biasName := fmt.Sprintf("%s_bias", layer.Name)
		outputName := fmt.Sprintf("%s_output", layer.Name)

		// Determine op type
		opType := "MatMul"
		if strings.Contains(layer.Type, "conv") {
			opType = "Conv"
		} else if strings.Contains(layer.Type, "relu") {
			opType = "Relu"
		} else if strings.Contains(layer.Type, "softmax") {
			opType = "Softmax"
		}

		// Add weight initializer
		if len(layer.Weights) > 0 {
			flatWeights := flattenWeights(layer.Weights)
			model.Graph.Initializers = append(model.Graph.Initializers, ONNXLikeTensor{
				Name:     weightName,
				DataType: "float32",
				Dims:     layer.Shape,
				Data:     flatWeights,
			})
		}

		// Add bias initializer
		if len(layer.Biases) > 0 {
			model.Graph.Initializers = append(model.Graph.Initializers, ONNXLikeTensor{
				Name:     biasName,
				DataType: "float32",
				Dims:     []int{len(layer.Biases)},
				Data:     layer.Biases,
			})
		}

		// Add node
		inputs := []string{prevOutput}
		if len(layer.Weights) > 0 {
			inputs = append(inputs, weightName)
		}
		if len(layer.Biases) > 0 {
			inputs = append(inputs, biasName)
		}

		model.Graph.Nodes = append(model.Graph.Nodes, ONNXLikeNode{
			Name:    nodeName,
			OpType:  opType,
			Inputs:  inputs,
			Outputs: []string{outputName},
		})

		prevOutput = outputName
	}

	// Add output tensor
	model.Graph.Outputs = append(model.Graph.Outputs, ONNXLikeTensor{
		Name:     prevOutput,
		DataType: "float32",
	})

	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(model)
}

// ============================================================================
// Verilog Parameters Export
// ============================================================================

func (me *ModelExporter) exportVerilogParams(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write module header
	fmt.Fprintf(file, "// IronLattice CIM Model Parameters\n")
	fmt.Fprintf(file, "// Generated for FPGA/ASIC deployment\n\n")

	// Write package
	fmt.Fprintf(file, "package model_params;\n\n")

	// Write parameters
	fmt.Fprintf(file, "  // Quantization parameters\n")
	fmt.Fprintf(file, "  parameter WEIGHT_BITS = %d;\n", me.Config.QuantBits)
	if me.CrossbarCfg != nil {
		fmt.Fprintf(file, "  parameter ADC_BITS = %d;\n", me.CrossbarCfg.ADCBits)
		fmt.Fprintf(file, "  parameter DAC_BITS = %d;\n", me.CrossbarCfg.DACBits)
		fmt.Fprintf(file, "  parameter ARRAY_SIZE = %d;\n\n", me.CrossbarCfg.ArraySize)
	}

	// Write layer parameters
	for i, layer := range me.Model.Layers {
		layerName := strings.ToUpper(strings.ReplaceAll(layer.Name, "-", "_"))

		fmt.Fprintf(file, "  // Layer %d: %s\n", i, layer.Name)

		if len(layer.Shape) >= 2 {
			fmt.Fprintf(file, "  parameter %s_IN = %d;\n", layerName, layer.Shape[1])
			fmt.Fprintf(file, "  parameter %s_OUT = %d;\n", layerName, layer.Shape[0])
		}

		// Write weights as memory initialization
		if len(layer.Weights) > 0 {
			rows := len(layer.Weights)
			cols := len(layer.Weights[0])

			fmt.Fprintf(file, "\n  // %s weights [%d x %d]\n", layer.Name, rows, cols)
			fmt.Fprintf(file, "  logic signed [%d-1:0] %s_W [%d][%d] = '{\n",
				me.Config.QuantBits, layerName, rows, cols)

			for r := 0; r < rows; r++ {
				fmt.Fprintf(file, "    '{")
				for c := 0; c < cols; c++ {
					qval := quantizeToInt(layer.Weights[r][c], me.Config.QuantBits)
					if c > 0 {
						fmt.Fprintf(file, ", ")
					}
					fmt.Fprintf(file, "%d'd%d", me.Config.QuantBits, qval)
				}
				fmt.Fprintf(file, "}")
				if r < rows-1 {
					fmt.Fprintf(file, ",")
				}
				fmt.Fprintf(file, "\n")
			}
			fmt.Fprintf(file, "  };\n")
		}
		fmt.Fprintf(file, "\n")
	}

	fmt.Fprintf(file, "endpackage\n")

	return nil
}

// ============================================================================
// Quantized Packed Export
// ============================================================================

func (me *ModelExporter) exportQuantizedPacked(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	header := struct {
		Magic       [4]byte
		Version     uint32
		NumLayers   uint32
		QuantBits   uint8
		Padding     [3]byte
	}{
		Magic:     [4]byte{'Q', 'P', 'K', 'T'},
		Version:   1,
		NumLayers: uint32(len(me.Model.Layers)),
		QuantBits: uint8(me.Config.QuantBits),
	}

	if err := binary.Write(file, binary.LittleEndian, header); err != nil {
		return err
	}

	// Write each layer
	for _, layer := range me.Model.Layers {
		// Write layer header
		nameBytes := []byte(layer.Name)
		if err := binary.Write(file, binary.LittleEndian, uint32(len(nameBytes))); err != nil {
			return err
		}
		if _, err := file.Write(nameBytes); err != nil {
			return err
		}

		rows := uint32(0)
		cols := uint32(0)
		if len(layer.Shape) >= 1 {
			rows = uint32(layer.Shape[0])
		}
		if len(layer.Shape) >= 2 {
			cols = uint32(layer.Shape[1])
		}

		if err := binary.Write(file, binary.LittleEndian, rows); err != nil {
			return err
		}
		if err := binary.Write(file, binary.LittleEndian, cols); err != nil {
			return err
		}

		// Pack and write weights
		if len(layer.Weights) > 0 {
			packedWeights := packWeights(layer.Weights, me.Config.QuantBits)
			if err := binary.Write(file, binary.LittleEndian, uint32(len(packedWeights))); err != nil {
				return err
			}
			if _, err := file.Write(packedWeights); err != nil {
				return err
			}
		} else {
			if err := binary.Write(file, binary.LittleEndian, uint32(0)); err != nil {
				return err
			}
		}

		// Pack and write biases (always 16-bit)
		if len(layer.Biases) > 0 {
			if err := binary.Write(file, binary.LittleEndian, uint32(len(layer.Biases))); err != nil {
				return err
			}
			for _, b := range layer.Biases {
				qb := int16(quantizeToInt(b, 16))
				if err := binary.Write(file, binary.LittleEndian, qb); err != nil {
					return err
				}
			}
		} else {
			if err := binary.Write(file, binary.LittleEndian, uint32(0)); err != nil {
				return err
			}
		}
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func quantizeToInt8(val float64, bits int) int8 {
	maxVal := float64(int(1)<<(bits-1)) - 1
	q := val * maxVal
	if q > maxVal {
		q = maxVal
	}
	if q < -maxVal {
		q = -maxVal
	}
	return int8(q)
}

func quantizeToInt16(val float64, bits int) int16 {
	maxVal := float64(int(1)<<(bits-1)) - 1
	q := val * maxVal
	if q > maxVal {
		q = maxVal
	}
	if q < -maxVal {
		q = -maxVal
	}
	return int16(q)
}

func quantizeToInt(val float64, bits int) int {
	maxVal := float64(int(1)<<(bits-1)) - 1
	q := val * maxVal
	if q > maxVal {
		q = maxVal
	}
	if q < -maxVal {
		q = -maxVal
	}
	return int(math.Round(q))
}

func weightToConductance(weight float64, cfg *CrossbarCheckpoint) float64 {
	// Map weight to conductance range
	// Assume weights are normalized to [-1, 1]
	normalized := (weight + 1.0) / 2.0
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}
	return cfg.MinCond + normalized*(cfg.MaxCond-cfg.MinCond)
}

func flattenWeights(weights [][]float64) []float64 {
	if len(weights) == 0 {
		return nil
	}
	rows := len(weights)
	cols := len(weights[0])
	flat := make([]float64, rows*cols)
	idx := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			flat[idx] = weights[r][c]
			idx++
		}
	}
	return flat
}

func packWeights(weights [][]float64, bits int) []byte {
	if len(weights) == 0 || bits > 8 {
		// For bits > 8, don't pack
		var buf []byte
		for _, row := range weights {
			for _, val := range row {
				q := quantizeToInt8(val, bits)
				buf = append(buf, byte(q))
			}
		}
		return buf
	}

	// Pack multiple values per byte
	totalVals := len(weights) * len(weights[0])
	valsPerByte := 8 / bits
	numBytes := (totalVals + valsPerByte - 1) / valsPerByte

	packed := make([]byte, numBytes)
	mask := byte((1 << bits) - 1)

	idx := 0
	byteIdx := 0
	bitPos := 0

	for _, row := range weights {
		for _, val := range row {
			q := byte(quantizeToInt(val, bits) & int(mask))
			packed[byteIdx] |= q << bitPos
			bitPos += bits
			if bitPos >= 8 {
				byteIdx++
				bitPos = 0
			}
			idx++
		}
	}

	return packed
}

// ============================================================================
// Model Info Export
// ============================================================================

// ModelInfo contains model metadata for documentation
type ModelInfo struct {
	Name           string                 `json:"name"`
	Version        string                 `json:"version"`
	NumLayers      int                    `json:"num_layers"`
	NumParameters  int64                  `json:"num_parameters"`
	LayerSummary   []LayerSummary         `json:"layer_summary"`
	MemoryBytes    int64                  `json:"memory_bytes"`
	CrossbarArrays int                    `json:"crossbar_arrays,omitempty"`
	Metrics        map[string]float64     `json:"metrics,omitempty"`
}

// LayerSummary contains layer information
type LayerSummary struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Shape      []int  `json:"shape"`
	Parameters int64  `json:"parameters"`
}

// GetModelInfo returns model information
func (me *ModelExporter) GetModelInfo() *ModelInfo {
	info := &ModelInfo{
		Name:         "IronLattice-Model",
		Version:      me.Model.Version,
		NumLayers:    len(me.Model.Layers),
		LayerSummary: make([]LayerSummary, len(me.Model.Layers)),
		Metrics:      me.Model.Metrics,
	}

	var totalParams int64
	var totalBytes int64

	for i, layer := range me.Model.Layers {
		params := int64(0)
		if len(layer.Weights) > 0 {
			params += int64(len(layer.Weights) * len(layer.Weights[0]))
		}
		params += int64(len(layer.Biases))

		info.LayerSummary[i] = LayerSummary{
			Name:       layer.Name,
			Type:       layer.Type,
			Shape:      layer.Shape,
			Parameters: params,
		}

		totalParams += params
		totalBytes += params * 4 // float32
	}

	info.NumParameters = totalParams
	info.MemoryBytes = totalBytes

	// Calculate crossbar arrays needed
	if me.CrossbarCfg != nil {
		arraySize := me.CrossbarCfg.ArraySize
		arrays := 0
		for _, layer := range me.Model.Layers {
			if len(layer.Shape) >= 2 {
				rowTiles := (layer.Shape[0] + arraySize - 1) / arraySize
				colTiles := (layer.Shape[1] + arraySize - 1) / arraySize
				arrays += rowTiles * colTiles
			}
		}
		info.CrossbarArrays = arrays
	}

	return info
}

// ExportModelInfo exports model information
func (me *ModelExporter) ExportModelInfo(filename string) error {
	info := me.GetModelInfo()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

// PrintModelSummary prints a model summary to writer
func (me *ModelExporter) PrintModelSummary(w io.Writer) {
	info := me.GetModelInfo()

	fmt.Fprintf(w, "Model Summary\n")
	fmt.Fprintf(w, "=============\n")
	fmt.Fprintf(w, "Version: %s\n", info.Version)
	fmt.Fprintf(w, "Layers: %d\n", info.NumLayers)
	fmt.Fprintf(w, "Parameters: %d\n", info.NumParameters)
	fmt.Fprintf(w, "Memory: %.2f KB\n\n", float64(info.MemoryBytes)/1024)

	fmt.Fprintf(w, "Layer Details:\n")
	fmt.Fprintf(w, "%-20s %-15s %-20s %s\n", "Name", "Type", "Shape", "Parameters")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 70))

	for _, layer := range info.LayerSummary {
		shapeStr := fmt.Sprintf("%v", layer.Shape)
		fmt.Fprintf(w, "%-20s %-15s %-20s %d\n", layer.Name, layer.Type, shapeStr, layer.Parameters)
	}

	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 70))
	fmt.Fprintf(w, "%-20s %-15s %-20s %d\n", "Total", "", "", info.NumParameters)

	if info.CrossbarArrays > 0 {
		fmt.Fprintf(w, "\nCrossbar Arrays Required: %d\n", info.CrossbarArrays)
	}
}

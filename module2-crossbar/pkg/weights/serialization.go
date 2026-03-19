// Package weights provides weight management utilities for crossbar neural networks.
package weights

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
)

// ModelMetadata stores model information for serialization.
type ModelMetadata struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Architecture string            `json:"architecture"`
	NumLayers    int               `json:"num_layers"`
	TotalParams  int               `json:"total_params"`
	Quantized    bool              `json:"quantized"`
	QuantBits    int               `json:"quant_bits,omitempty"`
	Custom       map[string]string `json:"custom,omitempty"`
}

// SerializedLayer stores weights for a single layer in model serialization.
type SerializedLayer struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"` // "linear", "conv2d", "attention", etc.
	Shape   []int                  `json:"shape"`
	Weights [][]float64            `json:"weights,omitempty"`
	Biases  []float64              `json:"biases,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

// Model represents a complete neural network model for serialization.
type Model struct {
	Metadata ModelMetadata     `json:"metadata"`
	Layers   []SerializedLayer `json:"layers"`
}

// NewModel creates a new model for serialization.
func NewModel(name, architecture string) *Model {
	return &Model{
		Metadata: ModelMetadata{
			Name:         name,
			Version:      "1.0",
			Architecture: architecture,
			Custom:       make(map[string]string),
		},
		Layers: make([]SerializedLayer, 0),
	}
}

// AddLayer adds a layer to the model.
func (m *Model) AddLayer(name, layerType string, weights [][]float64, biases []float64) {
	shape := make([]int, 2)
	if len(weights) > 0 {
		shape[0] = len(weights)
		if len(weights[0]) > 0 {
			shape[1] = len(weights[0])
		}
	}

	// Count parameters
	params := shape[0] * shape[1]
	if biases != nil {
		params += len(biases)
	}
	m.Metadata.TotalParams += params
	m.Metadata.NumLayers++

	m.Layers = append(m.Layers, SerializedLayer{
		Name:    name,
		Type:    layerType,
		Shape:   shape,
		Weights: weights,
		Biases:  biases,
		Extra:   make(map[string]interface{}),
	})
}

// SaveJSON saves model to JSON file.
func (m *Model) SaveJSON(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(m)
}

// LoadModelJSON loads model from JSON file.
func LoadModelJSON(path string) (*Model, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var model Model
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&model); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &model, nil
}

// SaveBinary saves model in compact binary format.
// Format: header + metadata + layers
func (m *Model) SaveBinary(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Use gzip compression
	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Write magic number and version
	if err := binary.Write(gzWriter, binary.LittleEndian, uint32(0x4D4F444C)); err != nil { // "MODL"
		return err
	}
	if err := binary.Write(gzWriter, binary.LittleEndian, uint32(1)); err != nil { // version 1
		return err
	}

	// Write metadata as JSON header
	metaBytes, err := json.Marshal(m.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := binary.Write(gzWriter, binary.LittleEndian, uint32(len(metaBytes))); err != nil {
		return err
	}
	if _, err := gzWriter.Write(metaBytes); err != nil {
		return err
	}

	// Write number of layers
	if err := binary.Write(gzWriter, binary.LittleEndian, uint32(len(m.Layers))); err != nil {
		return err
	}

	// Write each layer
	for _, layer := range m.Layers {
		if err := writeSerializedLayerBinary(gzWriter, &layer); err != nil {
			return fmt.Errorf("failed to write layer %s: %w", layer.Name, err)
		}
	}

	return nil
}

func writeSerializedLayerBinary(w io.Writer, layer *SerializedLayer) error {
	// Write layer name
	nameBytes := []byte(layer.Name)
	if err := binary.Write(w, binary.LittleEndian, uint32(len(nameBytes))); err != nil {
		return err
	}
	if _, err := w.Write(nameBytes); err != nil {
		return err
	}

	// Write layer type
	typeBytes := []byte(layer.Type)
	if err := binary.Write(w, binary.LittleEndian, uint32(len(typeBytes))); err != nil {
		return err
	}
	if _, err := w.Write(typeBytes); err != nil {
		return err
	}

	// Write shape
	if err := binary.Write(w, binary.LittleEndian, uint32(len(layer.Shape))); err != nil {
		return err
	}
	for _, dim := range layer.Shape {
		if err := binary.Write(w, binary.LittleEndian, uint32(dim)); err != nil {
			return err
		}
	}

	// Write weights
	rows := len(layer.Weights)
	if err := binary.Write(w, binary.LittleEndian, uint32(rows)); err != nil {
		return err
	}
	if rows > 0 {
		cols := len(layer.Weights[0])
		if err := binary.Write(w, binary.LittleEndian, uint32(cols)); err != nil {
			return err
		}
		for i := range layer.Weights {
			for j := range layer.Weights[i] {
				if err := binary.Write(w, binary.LittleEndian, layer.Weights[i][j]); err != nil {
					return err
				}
			}
		}
	}

	// Write biases
	biasLen := len(layer.Biases)
	if err := binary.Write(w, binary.LittleEndian, uint32(biasLen)); err != nil {
		return err
	}
	for _, b := range layer.Biases {
		if err := binary.Write(w, binary.LittleEndian, b); err != nil {
			return err
		}
	}

	return nil
}

// LoadModelBinary loads model from binary format.
func LoadModelBinary(path string) (*Model, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Read magic number
	var magic uint32
	if err := binary.Read(gzReader, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}
	if magic != 0x4D4F444C {
		return nil, fmt.Errorf("invalid magic number: expected MODL")
	}

	// Read version
	var version uint32
	if err := binary.Read(gzReader, binary.LittleEndian, &version); err != nil {
		return nil, err
	}
	if version != 1 {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	// Read metadata
	var metaLen uint32
	if err := binary.Read(gzReader, binary.LittleEndian, &metaLen); err != nil {
		return nil, err
	}
	metaBytes := make([]byte, metaLen)
	if _, err := io.ReadFull(gzReader, metaBytes); err != nil {
		return nil, err
	}

	model := &Model{}
	if err := json.Unmarshal(metaBytes, &model.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Read number of layers
	var numLayers uint32
	if err := binary.Read(gzReader, binary.LittleEndian, &numLayers); err != nil {
		return nil, err
	}

	// Read each layer
	model.Layers = make([]SerializedLayer, numLayers)
	for i := uint32(0); i < numLayers; i++ {
		layer, err := readSerializedLayerBinary(gzReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read layer %d: %w", i, err)
		}
		model.Layers[i] = *layer
	}

	return model, nil
}

func readSerializedLayerBinary(r io.Reader) (*SerializedLayer, error) {
	layer := &SerializedLayer{
		Extra: make(map[string]interface{}),
	}

	// Read layer name
	var nameLen uint32
	if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
		return nil, err
	}
	nameBytes := make([]byte, nameLen)
	if _, err := io.ReadFull(r, nameBytes); err != nil {
		return nil, err
	}
	layer.Name = string(nameBytes)

	// Read layer type
	var typeLen uint32
	if err := binary.Read(r, binary.LittleEndian, &typeLen); err != nil {
		return nil, err
	}
	typeBytes := make([]byte, typeLen)
	if _, err := io.ReadFull(r, typeBytes); err != nil {
		return nil, err
	}
	layer.Type = string(typeBytes)

	// Read shape
	var shapeDims uint32
	if err := binary.Read(r, binary.LittleEndian, &shapeDims); err != nil {
		return nil, err
	}
	layer.Shape = make([]int, shapeDims)
	for i := uint32(0); i < shapeDims; i++ {
		var dim uint32
		if err := binary.Read(r, binary.LittleEndian, &dim); err != nil {
			return nil, err
		}
		layer.Shape[i] = int(dim)
	}

	// Read weights
	var rows uint32
	if err := binary.Read(r, binary.LittleEndian, &rows); err != nil {
		return nil, err
	}
	if rows > 0 {
		var cols uint32
		if err := binary.Read(r, binary.LittleEndian, &cols); err != nil {
			return nil, err
		}
		layer.Weights = make([][]float64, rows)
		for i := uint32(0); i < rows; i++ {
			layer.Weights[i] = make([]float64, cols)
			for j := uint32(0); j < cols; j++ {
				if err := binary.Read(r, binary.LittleEndian, &layer.Weights[i][j]); err != nil {
					return nil, err
				}
			}
		}
	}

	// Read biases
	var biasLen uint32
	if err := binary.Read(r, binary.LittleEndian, &biasLen); err != nil {
		return nil, err
	}
	layer.Biases = make([]float64, biasLen)
	for i := uint32(0); i < biasLen; i++ {
		if err := binary.Read(r, binary.LittleEndian, &layer.Biases[i]); err != nil {
			return nil, err
		}
	}

	return layer, nil
}

// QuantizedModel represents a quantized model for efficient storage.
type QuantizedModel struct {
	Metadata     ModelMetadata
	Layers       []QuantizedLayerWeights
	ScaleFactors []float64 // Per-layer scale factors
}

// QuantizedLayerWeights stores quantized weights.
type QuantizedLayerWeights struct {
	Name        string
	Type        string
	Shape       []int
	Weights     [][]int8 // Quantized to int8
	Biases      []int16  // Quantized to int16
	WeightScale float64
	BiasScale   float64
}

// QuantizeModel quantizes a model to int8 weights.
func QuantizeModel(model *Model, bits int) *QuantizedModel {
	qmodel := &QuantizedModel{
		Metadata: model.Metadata,
		Layers:   make([]QuantizedLayerWeights, len(model.Layers)),
	}
	qmodel.Metadata.Quantized = true
	qmodel.Metadata.QuantBits = bits

	maxVal := float64(int(1)<<(bits-1) - 1)

	for i, layer := range model.Layers {
		qLayer := QuantizedLayerWeights{
			Name:  layer.Name,
			Type:  layer.Type,
			Shape: layer.Shape,
		}

		// Find max absolute weight
		weightMax := 0.0
		for _, row := range layer.Weights {
			for _, w := range row {
				if abs := math.Abs(w); abs > weightMax {
					weightMax = abs
				}
			}
		}

		// Quantize weights
		if weightMax > 0 {
			qLayer.WeightScale = weightMax / maxVal
		} else {
			qLayer.WeightScale = 1.0
		}

		qLayer.Weights = make([][]int8, len(layer.Weights))
		for r := range layer.Weights {
			qLayer.Weights[r] = make([]int8, len(layer.Weights[r]))
			for c := range layer.Weights[r] {
				qLayer.Weights[r][c] = int8(layer.Weights[r][c] / qLayer.WeightScale)
			}
		}

		// Quantize biases (use 16-bit for better precision)
		if len(layer.Biases) > 0 {
			biasMax := 0.0
			for _, b := range layer.Biases {
				if abs := math.Abs(b); abs > biasMax {
					biasMax = abs
				}
			}

			biasMaxVal := float64(int(1)<<15 - 1)
			if biasMax > 0 {
				qLayer.BiasScale = biasMax / biasMaxVal
			} else {
				qLayer.BiasScale = 1.0
			}

			qLayer.Biases = make([]int16, len(layer.Biases))
			for j, b := range layer.Biases {
				qLayer.Biases[j] = int16(b / qLayer.BiasScale)
			}
		}

		qmodel.Layers[i] = qLayer
	}

	return qmodel
}

// DequantizeLayer converts quantized layer back to float64.
func (qw *QuantizedLayerWeights) Dequantize() SerializedLayer {
	layer := SerializedLayer{
		Name:    qw.Name,
		Type:    qw.Type,
		Shape:   qw.Shape,
		Weights: make([][]float64, len(qw.Weights)),
		Biases:  make([]float64, len(qw.Biases)),
		Extra:   make(map[string]interface{}),
	}

	for i := range qw.Weights {
		layer.Weights[i] = make([]float64, len(qw.Weights[i]))
		for j := range qw.Weights[i] {
			layer.Weights[i][j] = float64(qw.Weights[i][j]) * qw.WeightScale
		}
	}

	for i := range qw.Biases {
		layer.Biases[i] = float64(qw.Biases[i]) * qw.BiasScale
	}

	return layer
}

// CrossbarMapping stores information about crossbar tile mapping.
type CrossbarMapping struct {
	LayerName   string
	TileSize    [2]int // [rows, cols]
	NumTiles    int
	TileOffsets [][2]int   // Starting offset for each tile
	TileMasks   [][][]bool // Skip masks for sparse weights
}

// GenerateCrossbarMapping creates mapping for deploying weights to crossbar tiles.
func GenerateCrossbarMapping(layer *SerializedLayer, tileRows, tileCols int) *CrossbarMapping {
	mapping := &CrossbarMapping{
		LayerName: layer.Name,
		TileSize:  [2]int{tileRows, tileCols},
	}

	rows := layer.Shape[0]
	cols := layer.Shape[1]

	numTileRows := (rows + tileRows - 1) / tileRows
	numTileCols := (cols + tileCols - 1) / tileCols
	mapping.NumTiles = numTileRows * numTileCols

	mapping.TileOffsets = make([][2]int, mapping.NumTiles)
	mapping.TileMasks = make([][][]bool, mapping.NumTiles)

	tileIdx := 0
	for tr := 0; tr < numTileRows; tr++ {
		for tc := 0; tc < numTileCols; tc++ {
			mapping.TileOffsets[tileIdx] = [2]int{tr * tileRows, tc * tileCols}

			// Generate skip mask
			mask := make([][]bool, tileRows)
			for i := range mask {
				mask[i] = make([]bool, tileCols)
				srcRow := tr*tileRows + i
				for j := range mask[i] {
					srcCol := tc*tileCols + j
					if srcRow < rows && srcCol < cols {
						mask[i][j] = layer.Weights[srcRow][srcCol] == 0
					} else {
						mask[i][j] = true // Padding
					}
				}
			}
			mapping.TileMasks[tileIdx] = mask
			tileIdx++
		}
	}

	return mapping
}


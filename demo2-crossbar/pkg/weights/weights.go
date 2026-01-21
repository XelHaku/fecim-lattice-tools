// Package weights provides weight import/export utilities for neural network models.
package weights

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
)

// Format specifies the file format for weight storage.
type Format int

const (
	FormatJSON   Format = iota // Human-readable JSON
	FormatBinary               // Compact binary format
	FormatNumPy                // NumPy .npy compatible
)

// LayerWeights holds weights for a single neural network layer.
type LayerWeights struct {
	Name  string     `json:"name"`
	Shape []int      `json:"shape"` // [rows, cols] or [out_features, in_features]
	Dtype string     `json:"dtype"` // "float32" or "float64"
	Data  []float64  `json:"data"`  // Flattened weight data
	Bias  []float64  `json:"bias,omitempty"`
	Quant *QuantInfo `json:"quant,omitempty"`
}

// QuantInfo holds quantization parameters for hardware deployment.
type QuantInfo struct {
	Bits       int     `json:"bits"`        // Quantization bits
	Scale      float64 `json:"scale"`       // Scaling factor
	ZeroPoint  float64 `json:"zero_point"`  // Zero point offset
	Symmetric  bool    `json:"symmetric"`   // Symmetric quantization
	PerChannel bool    `json:"per_channel"` // Per-channel vs per-tensor
}

// ModelWeights holds all weights for a complete model.
type ModelWeights struct {
	Name      string         `json:"name"`
	Version   string         `json:"version"`
	NumLayers int            `json:"num_layers"`
	Layers    []LayerWeights `json:"layers"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// NewModelWeights creates a new model weights container.
func NewModelWeights(name string, numLayers int) *ModelWeights {
	return &ModelWeights{
		Name:      name,
		Version:   "1.0",
		NumLayers: numLayers,
		Layers:    make([]LayerWeights, 0, numLayers),
		Metadata:  make(map[string]any),
	}
}

// AddLayer adds a layer's weights to the model.
func (m *ModelWeights) AddLayer(name string, weights [][]float64, bias []float64) {
	if len(weights) == 0 {
		return
	}

	rows := len(weights)
	cols := len(weights[0])

	// Flatten weights
	data := make([]float64, rows*cols)
	for i, row := range weights {
		for j, w := range row {
			data[i*cols+j] = w
		}
	}

	layer := LayerWeights{
		Name:  name,
		Shape: []int{rows, cols},
		Dtype: "float64",
		Data:  data,
		Bias:  bias,
	}

	m.Layers = append(m.Layers, layer)
}

// GetLayer retrieves a layer's weights by name.
func (m *ModelWeights) GetLayer(name string) (*LayerWeights, error) {
	for i := range m.Layers {
		if m.Layers[i].Name == name {
			return &m.Layers[i], nil
		}
	}
	return nil, fmt.Errorf("layer not found: %s", name)
}

// GetLayerByIndex retrieves a layer's weights by index.
func (m *ModelWeights) GetLayerByIndex(idx int) (*LayerWeights, error) {
	if idx < 0 || idx >= len(m.Layers) {
		return nil, fmt.Errorf("layer index out of range: %d", idx)
	}
	return &m.Layers[idx], nil
}

// ToMatrix converts flattened layer data back to 2D matrix.
func (l *LayerWeights) ToMatrix() [][]float64 {
	if len(l.Shape) != 2 {
		return nil
	}

	rows, cols := l.Shape[0], l.Shape[1]
	matrix := make([][]float64, rows)
	for i := range matrix {
		matrix[i] = make([]float64, cols)
		for j := range matrix[i] {
			matrix[i][j] = l.Data[i*cols+j]
		}
	}
	return matrix
}

// SaveJSON saves model weights to JSON format.
func (m *ModelWeights) SaveJSON(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// LoadJSON loads model weights from JSON format.
func LoadJSON(path string) (*ModelWeights, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	var m ModelWeights
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &m, nil
}

// Binary format header
const binaryMagic uint32 = 0x46455745 // "FEWE" - Ferroelectric Weights

// SaveBinary saves model weights to compact binary format.
func (m *ModelWeights) SaveBinary(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Write header
	if err := binary.Write(f, binary.LittleEndian, binaryMagic); err != nil {
		return err
	}

	// Write number of layers
	if err := binary.Write(f, binary.LittleEndian, uint32(len(m.Layers))); err != nil {
		return err
	}

	// Write each layer
	for _, layer := range m.Layers {
		if err := writeLayerBinary(f, &layer); err != nil {
			return err
		}
	}

	return nil
}

// LoadBinary loads model weights from binary format.
func LoadBinary(path string) (*ModelWeights, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Read and verify header
	var magic uint32
	if err := binary.Read(f, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}
	if magic != binaryMagic {
		return nil, fmt.Errorf("invalid file format (magic: 0x%X)", magic)
	}

	// Read number of layers
	var numLayers uint32
	if err := binary.Read(f, binary.LittleEndian, &numLayers); err != nil {
		return nil, err
	}

	m := NewModelWeights("", int(numLayers))

	// Read each layer
	for i := uint32(0); i < numLayers; i++ {
		layer, err := readLayerBinary(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read layer %d: %w", i, err)
		}
		m.Layers = append(m.Layers, *layer)
	}

	return m, nil
}

func writeLayerBinary(w io.Writer, layer *LayerWeights) error {
	// Write name length and name
	nameBytes := []byte(layer.Name)
	if err := binary.Write(w, binary.LittleEndian, uint32(len(nameBytes))); err != nil {
		return err
	}
	if _, err := w.Write(nameBytes); err != nil {
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

	// Write data as float32 for space efficiency
	if err := binary.Write(w, binary.LittleEndian, uint32(len(layer.Data))); err != nil {
		return err
	}
	for _, v := range layer.Data {
		if err := binary.Write(w, binary.LittleEndian, float32(v)); err != nil {
			return err
		}
	}

	// Write bias
	if err := binary.Write(w, binary.LittleEndian, uint32(len(layer.Bias))); err != nil {
		return err
	}
	for _, v := range layer.Bias {
		if err := binary.Write(w, binary.LittleEndian, float32(v)); err != nil {
			return err
		}
	}

	return nil
}

func readLayerBinary(r io.Reader) (*LayerWeights, error) {
	layer := &LayerWeights{Dtype: "float32"}

	// Read name
	var nameLen uint32
	if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
		return nil, err
	}
	nameBytes := make([]byte, nameLen)
	if _, err := io.ReadFull(r, nameBytes); err != nil {
		return nil, err
	}
	layer.Name = string(nameBytes)

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

	// Read data
	var dataLen uint32
	if err := binary.Read(r, binary.LittleEndian, &dataLen); err != nil {
		return nil, err
	}
	layer.Data = make([]float64, dataLen)
	for i := uint32(0); i < dataLen; i++ {
		var v float32
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return nil, err
		}
		layer.Data[i] = float64(v)
	}

	// Read bias
	var biasLen uint32
	if err := binary.Read(r, binary.LittleEndian, &biasLen); err != nil {
		return nil, err
	}
	layer.Bias = make([]float64, biasLen)
	for i := uint32(0); i < biasLen; i++ {
		var v float32
		if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
			return nil, err
		}
		layer.Bias[i] = float64(v)
	}

	return layer, nil
}

// QuantizeWeights applies quantization to layer weights for hardware deployment.
func (l *LayerWeights) QuantizeWeights(bits int, symmetric bool) {
	if bits <= 0 || bits > 16 {
		bits = 8
	}

	// Find min/max
	minVal, maxVal := l.Data[0], l.Data[0]
	for _, v := range l.Data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	levels := float64(int(1) << uint(bits))
	var scale, zeroPoint float64

	if symmetric {
		// Symmetric quantization: zero point at 0
		absMax := math.Max(math.Abs(minVal), math.Abs(maxVal))
		scale = absMax / (levels/2 - 1)
		zeroPoint = 0
	} else {
		// Asymmetric quantization
		scale = (maxVal - minVal) / (levels - 1)
		zeroPoint = -minVal / scale
	}

	// Apply quantization
	for i := range l.Data {
		q := math.Round(l.Data[i]/scale + zeroPoint)
		q = math.Max(0, math.Min(levels-1, q))
		l.Data[i] = (q - zeroPoint) * scale
	}

	l.Quant = &QuantInfo{
		Bits:      bits,
		Scale:     scale,
		ZeroPoint: zeroPoint,
		Symmetric: symmetric,
	}
}

// GetStatistics returns statistics about layer weights.
func (l *LayerWeights) GetStatistics() (min, max, mean, std float64) {
	if len(l.Data) == 0 {
		return
	}

	min, max = l.Data[0], l.Data[0]
	var sum float64

	for _, v := range l.Data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	mean = sum / float64(len(l.Data))

	var variance float64
	for _, v := range l.Data {
		diff := v - mean
		variance += diff * diff
	}
	std = math.Sqrt(variance / float64(len(l.Data)))

	return
}

// NormalizeWeights normalizes weights to [0, 1] range for crossbar programming.
func (l *LayerWeights) NormalizeWeights() (scale, offset float64) {
	min, max, _, _ := l.GetStatistics()

	scale = max - min
	if scale == 0 {
		scale = 1
	}
	offset = min

	for i := range l.Data {
		l.Data[i] = (l.Data[i] - offset) / scale
	}

	return scale, offset
}

// DenormalizeWeights reverses normalization using stored scale and offset.
func (l *LayerWeights) DenormalizeWeights(scale, offset float64) {
	for i := range l.Data {
		l.Data[i] = l.Data[i]*scale + offset
	}
}

// PrintSummary prints a summary of model weights.
func (m *ModelWeights) PrintSummary() {
	fmt.Printf("Model: %s (v%s)\n", m.Name, m.Version)
	fmt.Printf("Layers: %d\n", len(m.Layers))
	fmt.Println("---")

	var totalParams int64
	for i, layer := range m.Layers {
		params := int64(len(layer.Data)) + int64(len(layer.Bias))
		totalParams += params
		min, max, mean, std := layer.GetStatistics()
		fmt.Printf("Layer %d: %s\n", i, layer.Name)
		fmt.Printf("  Shape: %v\n", layer.Shape)
		fmt.Printf("  Params: %d\n", params)
		fmt.Printf("  Stats: min=%.4f max=%.4f mean=%.4f std=%.4f\n", min, max, mean, std)
		if layer.Quant != nil {
			fmt.Printf("  Quant: %d-bit, scale=%.6f\n", layer.Quant.Bits, layer.Quant.Scale)
		}
	}
	fmt.Println("---")
	fmt.Printf("Total parameters: %d\n", totalParams)
}

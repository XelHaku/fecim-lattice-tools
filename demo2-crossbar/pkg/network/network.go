// Package network implements neural network inference on crossbar arrays.
package network

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/ironlattice/vis/demo2-inference/pkg/crossbar"
)

// Config contains neural network configuration.
type Config struct {
	InputSize  int // Input layer size
	HiddenSize int // Hidden layer size (matches crossbar)
	OutputSize int // Output layer size
	NumLayers  int // Number of layers (including input/output)
}

// Layer represents a single neural network layer.
type Layer struct {
	inputSize  int
	outputSize int
	array      *crossbar.Array // Crossbar array storing weights
	biases     []float64
}

// Network represents a multi-layer neural network.
type Network struct {
	config   *Config
	layers   []*Layer
	opsCount int64
}

// NewNetwork creates a new neural network mapped to crossbar arrays.
func NewNetwork(cfg *Config, baseArray *crossbar.Array) (*Network, error) {
	if cfg.NumLayers < 2 {
		return nil, fmt.Errorf("network must have at least 2 layers")
	}

	net := &Network{
		config: cfg,
		layers: make([]*Layer, cfg.NumLayers-1), // N-1 weight layers for N layers
	}

	// Create layers
	sizes := net.computeLayerSizes()
	for i := 0; i < len(sizes)-1; i++ {
		inSize := sizes[i]
		outSize := sizes[i+1]

		// Create crossbar array for this layer
		arrayCfg := &crossbar.Config{
			Rows:       outSize,
			Cols:       inSize,
			NoiseLevel: 0.02,
			ADCBits:    6,
			DACBits:    8,
		}
		array, err := crossbar.NewArray(arrayCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create array for layer %d: %v", i, err)
		}

		// Initialize random weights
		weights := make([][]float64, outSize)
		for r := range weights {
			weights[r] = make([]float64, inSize)
			for c := range weights[r] {
				// Xavier initialization scaled to [0, 1]
				weights[r][c] = (rand.Float64()*2-1)*math.Sqrt(2.0/float64(inSize+outSize))*0.5 + 0.5
			}
		}
		if err := array.ProgramWeightMatrix(weights); err != nil {
			return nil, err
		}

		// Initialize biases
		biases := make([]float64, outSize)
		for b := range biases {
			biases[b] = 0.0
		}

		net.layers[i] = &Layer{
			inputSize:  inSize,
			outputSize: outSize,
			array:      array,
			biases:     biases,
		}
	}

	return net, nil
}

// computeLayerSizes returns the size of each layer.
func (n *Network) computeLayerSizes() []int {
	sizes := make([]int, n.config.NumLayers)
	sizes[0] = n.config.InputSize
	sizes[n.config.NumLayers-1] = n.config.OutputSize

	// Hidden layers
	for i := 1; i < n.config.NumLayers-1; i++ {
		sizes[i] = n.config.HiddenSize
	}

	return sizes
}

// Forward performs forward inference through the network.
func (n *Network) Forward(input []float64) []float64 {
	activation := input

	for i, layer := range n.layers {
		// Matrix-vector multiplication on crossbar
		output, err := layer.array.MVM(activation)
		if err != nil {
			panic(fmt.Sprintf("MVM failed at layer %d: %v", i, err))
		}

		// Add biases
		for j := range output {
			output[j] += layer.biases[j]
		}

		// Apply activation function (ReLU for hidden, softmax for output)
		if i < len(n.layers)-1 {
			activation = relu(output)
		} else {
			activation = softmax(output)
		}

		// Count operations
		n.opsCount += int64(layer.inputSize * layer.outputSize)
	}

	return activation
}

// GetOpsCount returns the total MAC operations performed.
func (n *Network) GetOpsCount() int64 {
	return n.opsCount
}

// relu applies ReLU activation function.
func relu(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = math.Max(0, v)
	}
	return result
}

// softmax applies softmax activation function.
func softmax(x []float64) []float64 {
	result := make([]float64, len(x))

	// Find max for numerical stability
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}

	// Compute exp and sum
	var sum float64
	for i, v := range x {
		result[i] = math.Exp(v - max)
		sum += result[i]
	}

	// Normalize
	for i := range result {
		result[i] /= sum
	}

	return result
}

// LoadWeights loads pre-trained weights into the network.
func (n *Network) LoadWeights(layerWeights [][][]float64, layerBiases [][]float64) error {
	if len(layerWeights) != len(n.layers) {
		return fmt.Errorf("weight count mismatch: got %d, expected %d", len(layerWeights), len(n.layers))
	}

	for i, layer := range n.layers {
		if err := layer.array.ProgramWeightMatrix(layerWeights[i]); err != nil {
			return fmt.Errorf("failed to program weights for layer %d: %v", i, err)
		}
		if i < len(layerBiases) {
			copy(layer.biases, layerBiases[i])
		}
	}

	return nil
}

// GetLayerCount returns the number of weight layers.
func (n *Network) GetLayerCount() int {
	return len(n.layers)
}

// GetLayerDimensions returns dimensions of a specific layer.
func (n *Network) GetLayerDimensions(layerIdx int) (input, output int, err error) {
	if layerIdx < 0 || layerIdx >= len(n.layers) {
		return 0, 0, fmt.Errorf("layer index out of range: %d", layerIdx)
	}
	layer := n.layers[layerIdx]
	return layer.inputSize, layer.outputSize, nil
}
